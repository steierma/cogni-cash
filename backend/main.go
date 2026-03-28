package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	dynamicbank "cogni-cash/internal/adapter/bank/dynamic"
	mockbank "cogni-cash/internal/adapter/bank/mock"
	httpAdapter "cogni-cash/internal/adapter/http"
	llmadapter "cogni-cash/internal/adapter/ollama"
	aiparser "cogni-cash/internal/adapter/parser/bank_statement/ai"
	amazonvisaparser "cogni-cash/internal/adapter/parser/bank_statement/amazonvisa"
	ingparser "cogni-cash/internal/adapter/parser/bank_statement/ing"
	ingcsvparser "cogni-cash/internal/adapter/parser/bank_statement/ingcsv"
	vwparser "cogni-cash/internal/adapter/parser/bank_statement/vw"
	invoiceparser "cogni-cash/internal/adapter/parser/invoice"
	payslipaiparser "cogni-cash/internal/adapter/parser/payslip/ai"
	cariadparser "cogni-cash/internal/adapter/parser/payslip/cariad"
	memrepo "cogni-cash/internal/adapter/repository/memory"
	pgrepo "cogni-cash/internal/adapter/repository/postgres"
	"cogni-cash/internal/domain/port"
	"cogni-cash/internal/domain/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))
	addr := envOrDefault("SERVER_ADDR", ":8080")

	jwtSecret := envOrDefault("JWT_SECRET", "super-secret-default-key-change-me")
	adminUsername := envOrDefault("ADMIN_USERNAME", "admin")
	adminPassword := envOrDefault("ADMIN_PASSWORD", "")

	// Root context for background workers
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	// --- Database Configuration & Connection ---
	storageMode := envOrDefault("DB_TYPE", "postgres")
	var pool *pgxpool.Pool
	var dbPinger func(context.Context) error
	var dbHost string

	var invoiceRepo port.InvoiceRepository
	var bankStmtRepo port.BankStatementRepository
	var bankRepo port.BankRepository
	var categoryRepo port.CategoryRepository
	var userRepo port.UserRepository
	var reconciliationRepo port.ReconciliationRepository
	var settingsRepo port.SettingsRepository
	var payslipRepo port.PayslipRepository

	if storageMode == "memory" {
		logger.Info("Using in-memory storage mode")
		invoiceRepo = memrepo.NewInvoiceRepository()
		bankStmtRepo = memrepo.NewBankStatementRepository()
		bankRepo = memrepo.NewBankRepository()
		categoryRepo = memrepo.NewCategoryRepository()
		userRepo = memrepo.NewUserRepository()
		reconciliationRepo = memrepo.NewReconciliationRepository()
		settingsRepo = memrepo.NewSettingsRepository()
		payslipRepo = memrepo.NewPayslipRepository()

		dbPinger = func(ctx context.Context) error { return nil }
		dbHost = "memory"

		// Seed default data for in-memory mode
		seedInMemoryData(appCtx, userRepo, categoryRepo, logger, bankRepo, bankStmtRepo, invoiceRepo, payslipRepo)
	} else {
		dbUser := envOrDefault("POSTGRES_USER", "")
		dbPassword := envOrDefault("POSTGRES_PASSWORD", "")
		dbHost = envOrDefault("DATABASE_HOST", "localhost")
		dbPort := envOrDefault("DATABASE_PORT", "5432")
		dbName := envOrDefault("POSTGRES_DB", "")

		databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			dbUser, dbPassword, dbHost, dbPort, dbName)

		logger.Info("Attempting to connect to PostgreSQL...", "host", dbHost, "db", dbName)
		var err error
		pool, err = pgrepo.NewPool(appCtx, databaseURL)
		if err != nil {
			logger.Error("Failed to connect to PostgreSQL. A database connection is strictly required. Exiting.", "error", err)
			os.Exit(1)
		}
		logger.Info("Successfully connected to PostgreSQL")

		dbPinger = func(ctx context.Context) error { return pool.Ping(ctx) }

		// --- Instantiate Repositories ---
		invoiceRepo = pgrepo.NewInvoiceRepository(pool, logger)
		bankStmtRepo = pgrepo.NewBankStatementRepository(pool, logger)
		bankRepo = pgrepo.NewBankRepository(pool, logger)
		categoryRepo = pgrepo.NewCategoryRepository(pool, logger)
		userRepo = pgrepo.NewUserRepository(pool, logger)
		reconciliationRepo = pgrepo.NewReconciliationRepository(pool, bankStmtRepo, logger)
		settingsRepo = pgrepo.NewSettingsRepository(pool, logger)
		payslipRepo = pgrepo.NewPayslipRepository(pool)
	}

	ensureDefaultSettings(appCtx, settingsRepo, logger)

	// --- Lade Enable Banking Key aus Environment ---
	ebPrivateKey, err := loadEnableBankingKey(logger)
	if err != nil {
		logger.Warn("Enable Banking private key not loaded or invalid (EB provider will be disabled/fail)", "reason", err)
	}

	// --- Instantiate Services ---
	authSvc := service.NewAuthService(userRepo, jwtSecret, logger)
	if err := authSvc.EnsureAdminUser(appCtx, adminUsername, adminPassword); err != nil {
		logger.Error("Failed to seed default admin user", "error", err)
	}

	userSvc := service.NewUserService(userRepo, logger)
	llmClient := llmadapter.NewAdapter(settingsRepo, logger)
	settingsSvc := service.NewSettingsService(settingsRepo, logger)

	// Load categories once at startup (used by InvoiceService for LLM prompts).
	initialCategories, _ := categoryRepo.FindAll(appCtx)

	invoiceFileParser := invoiceparser.NewParser()
	invoiceSvc := service.NewInvoiceService(invoiceRepo, invoiceFileParser, llmClient, initialCategories, logger)

	bankStatementSvc := service.NewBankStatementService(bankStmtRepo, logger)
	transactionSvc := service.NewTransactionService(bankStmtRepo, categoryRepo, settingsRepo, llmClient, logger)
	reconciliationSvc := service.NewReconciliationService(bankStmtRepo, reconciliationRepo, logger)

	var bankProvider port.BankProvider
	if storageMode == "memory" {
		bankProvider = mockbank.NewMockBankProvider()
	} else {
		bankProvider = dynamicbank.NewAdapter(settingsRepo, ebPrivateKey, logger)
	}

	bankSvc := service.NewBankService(bankRepo, bankStmtRepo, settingsRepo, bankProvider, logger)

	// If using dynamic adapter (Postgres mode), decide if we allow mock interception
	if storageMode != "memory" {
		if dyn, ok := bankProvider.(*dynamicbank.Adapter); ok {
			dyn.AllowMocks = envOrDefault("DEMO_MODE", "false") == "true"
		}
	}

	aiFallbackParser := aiparser.NewParser(llmClient, logger)
	bankStatementSvc.WithFallbackParser(aiFallbackParser)

	// --- Parser Registration ---
	bankStatementSvc.RegisterParser(".pdf", vwparser.NewParser())
	bankStatementSvc.RegisterParser(".pdf", ingparser.NewParser(logger))
	bankStatementSvc.RegisterParser(".csv", ingcsvparser.NewParser(logger))
	bankStatementSvc.RegisterParser(".xls", amazonvisaparser.NewParser(logger))

	// --- Payslip Service ---
	cariadParser := cariadparser.NewParser(logger)
	aiPayslipParser := payslipaiparser.NewPayslipParser(llmClient, logger)
	payslipSvc := service.NewPayslipService(payslipRepo, cariadParser, aiPayslipParser, logger)

	// --- Background Workers ---
	go func() {
		logger.Info("Background worker: Auto-import started")
		for {
			dir, _ := settingsRepo.Get(appCtx, "import_dir")
			intervalStr, _ := settingsRepo.Get(appCtx, "import_interval")

			interval, err := time.ParseDuration(intervalStr)
			if err != nil || interval < time.Minute {
				interval = time.Hour
			}

			if dir != "" {
				runImport(appCtx, bankStatementSvc, dir, logger)
			}

			select {
			case <-appCtx.Done():
				logger.Info("Background worker: Auto-import shutting down")
				return
			case <-time.After(interval):
			}
		}
	}()

	go func() {
		logger.Info("Background worker: Auto-categorization started")
		for {
			enabledStr, _ := settingsRepo.Get(appCtx, "auto_categorization_enabled")
			intervalStr, _ := settingsRepo.Get(appCtx, "auto_categorization_interval")

			interval, err := time.ParseDuration(intervalStr)
			if err != nil || interval < time.Minute {
				interval = 5 * time.Minute
			}

			if enabledStr == "true" {
				batchSizeStr, _ := settingsRepo.Get(appCtx, "auto_categorization_batch_size")
				batchSize := 10
				if bs, err := strconv.Atoi(batchSizeStr); err == nil && bs > 0 {
					batchSize = bs
				}

				if err := transactionSvc.StartAutoCategorizeAsync(appCtx, batchSize); err != nil {
					if !errors.Is(err, service.ErrJobAlreadyRunning) && !errors.Is(err, service.ErrNothingToCategorize) {
						logger.Error("Autocategorization tick error", "error", err)
					}
				} else {
					logger.Info("Auto-categorization job triggered via schedule", "batch_size", batchSize)
				}
			}

			select {
			case <-appCtx.Done():
				logger.Info("Background worker: Auto-categorization shutting down")
				return
			case <-time.After(interval):
			}
		}
	}()

	go func() {
		logger.Info("Background worker: Payslip JSON import started")
		for {
			jsonPath, _ := settingsRepo.Get(appCtx, "payslip_import_json_path")
			intervalStr, _ := settingsRepo.Get(appCtx, "payslip_import_interval")

			interval, err := time.ParseDuration(intervalStr)
			if err != nil || interval < time.Minute {
				interval = time.Hour
			}

			if jsonPath != "" {
				runPayslipJSONImport(appCtx, payslipSvc, jsonPath, logger)
			}

			select {
			case <-appCtx.Done():
				logger.Info("Background worker: Payslip JSON import shutting down")
				return
			case <-time.After(interval):
			}
		}
	}()

	go func() {
		logger.Info("Background worker: Smart Bank Sync started")
		for {
			enabledStr, _ := settingsRepo.Get(appCtx, "bank_sync_enabled")
			if enabledStr != "true" {
				select {
				case <-appCtx.Done():
					return
				case <-time.After(1 * time.Hour):
					continue
				}
			}

			nextSyncStr, _ := settingsRepo.Get(appCtx, "bank_sync_next_run")
			var nextSync time.Time
			if nextSyncStr != "" {
				nextSync, _ = time.Parse(time.RFC3339, nextSyncStr)
			}

			now := time.Now()
			if nextSync.IsZero() || now.After(nextSync) {
				logger.Info("Starting scheduled smart bank sync for all users")
				users, err := userRepo.FindAll(appCtx, "")
				if err != nil {
					logger.Error("Smart bank sync: failed to fetch users", "error", err)
				} else {
					for _, u := range users {
						if err := bankSvc.SyncAllAccounts(appCtx, u.ID); err != nil {
							logger.Error("Smart bank sync: failed for user", "user_id", u.ID, "error", err)
						}
					}
				}

				// Schedule next run: random time between 8:00 and 20:00, 1 day from now
				daysToAdd := 1
				nextDate := now.AddDate(0, 0, daysToAdd)

				// Random hour between 11 and 13 (exclusive of 11, so up to 12:59)
				randomHour := 11 + rand.Intn(2)
				randomMinute := rand.Intn(60)

				nextSync = time.Date(nextDate.Year(), nextDate.Month(), nextDate.Day(), randomHour, randomMinute, 0, 0, now.Location())

				_ = settingsRepo.Set(appCtx, "bank_sync_next_run", nextSync.Format(time.RFC3339))
				logger.Info("Smart bank sync finished. Next sync scheduled.", "at", nextSync.Format(time.RFC3339))
			}

			// Check every hour if it's time to sync
			select {
			case <-appCtx.Done():
				logger.Info("Background worker: Smart Bank Sync shutting down")
				return
			case <-time.After(1 * time.Hour):
			}
		}
	}()

	// --- HTTP Handler Setup ---
	handler := httpAdapter.NewHandler(authSvc, invoiceSvc, bankStatementSvc, settingsSvc, bankSvc, logger, storageMode, dbHost, dbPinger)

	handler.WithUserService(userSvc)
	handler.WithTransactionService(transactionSvc)
	handler.WithReconciliationService(reconciliationSvc)
	handler.WithBankService(bankSvc)
	handler.WithBankStatementRepository(bankStmtRepo)
	handler.WithCategoryRepository(categoryRepo)
	handler.WithReconciliationRepository(reconciliationRepo)
	handler.WithPayslipService(payslipSvc)
	handler.WithPayslipRepository(payslipRepo)

	srv := httpAdapter.NewServer(addr, handler)

	// --- Server Start ---
	go func() {
		logger.Info("Starting HTTP server", "address", addr)
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server failed to start or crashed", "error", err)
			os.Exit(1)
		}
	}()

	// --- Graceful Shutdown Sequence ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutdown signal received. Initiating graceful shutdown...")

	appCancel()
	logger.Info("Signalled background workers to stop")

	shutCtx, srvCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer srvCancel()

	logger.Info("Shutting down HTTP server...")
	if err := srv.Shutdown(shutCtx); err != nil {
		logger.Error("HTTP server shutdown forced or timed out", "error", err)
	} else {
		logger.Info("HTTP server shut down cleanly")
	}

	if pool != nil {
		logger.Info("Closing PostgreSQL connection pool...")
		pool.Close()
		logger.Info("PostgreSQL connection pool closed")
	}

	logger.Info("Application shutdown complete. Goodbye!")
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func ensureDefaultSettings(ctx context.Context, repo port.SettingsRepository, logger *slog.Logger) {
	defaults := map[string]string{
		"llm_api_url":                    envOrDefault("OLLAMA_URL", "http://localhost:11434"),
		"llm_api_token":                  "",
		"llm_model":                      "deepseek-r1",
		"import_dir":                     envOrDefault("IMPORT_DIR", ""),
		"import_interval":                envOrDefault("IMPORT_INTERVAL", "1h"),
		"theme":                          "system",
		"default_currency":               "EUR",
		"auto_categorization_enabled":    "false",
		"auto_categorization_interval":   "5m",
		"auto_categorization_batch_size": "10",
		"auto_categorization_examples_per_category": "20",
		"payslip_import_json_path":                  envOrDefault("PAYSLIP_IMPORT_JSON_PATH", ""),
		"payslip_import_interval":                   envOrDefault("PAYSLIP_IMPORT_INTERVAL", "1h"),
		"bank_provider":                             envOrDefault("BANK_PROVIDER", "enablebanking"),
		"bank_sync_enabled":                         "true",
		"bank_sync_interval":                        "1h",
		"bank_sync_next_run":                        "",
		"gocardless_secret_id":                      envOrDefault("GOCARDLESS_SECRET_ID", ""),
		"gocardless_secret_key":                     envOrDefault("GOCARDLESS_SECRET_KEY", ""),
		"enablebanking_app_id":                      envOrDefault("ENABLEBANKING_APP_ID", ""),
	}

	envOverrideKeys := map[string]string{
		"payslip_import_json_path": "PAYSLIP_IMPORT_JSON_PATH",
		"payslip_import_interval":  "PAYSLIP_IMPORT_INTERVAL",
		"import_dir":               "IMPORT_DIR",
		"import_interval":          "IMPORT_INTERVAL",
	}

	for k, v := range defaults {
		existing, _ := repo.Get(ctx, k)

		if envKey, ok := envOverrideKeys[k]; ok {
			envVal, envSet := os.LookupEnv(envKey)
			if envSet && envVal != existing {
				if err := repo.Set(ctx, k, envVal); err != nil {
					logger.Warn("Failed to sync env setting to DB", "key", k, "error", err)
				}
			}
			continue
		}

		if existing == "" {
			if err := repo.Set(ctx, k, v); err != nil {
				logger.Warn("Failed to set default setting", "key", k, "error", err)
			}
		}
	}
}

func runImport(ctx context.Context, svc *service.BankStatementService, dir string, logger *slog.Logger) {
	count, errs := svc.ImportFromDirectory(ctx, dir)
	if len(errs) > 0 {
		logger.Error("Directory import completed with errors", "imported_count", count, "errors_count", len(errs))
		for _, err := range errs {
			logger.Error("Import error", "error", err)
		}
	} else if count > 0 {
		logger.Info("Directory import completed successfully", "imported_count", count)
	}
}

func runPayslipJSONImport(ctx context.Context, svc *service.PayslipService, jsonPath string, logger *slog.Logger) {
	imported, skipped, errs, fatalErr := svc.ImportFromJSONFile(ctx, jsonPath)
	if fatalErr != nil {
		logger.Error("Payslip JSON import: fatal error", "file", jsonPath, "error", fatalErr)
		return
	}
	if len(errs) > 0 {
		logger.Error("Payslip JSON import: completed with per-record errors",
			"imported", imported, "skipped", skipped, "errors_count", len(errs))
		for _, e := range errs {
			logger.Error("Payslip JSON import error", "error", e)
		}
		return
	}
	if imported > 0 || skipped > 0 {
		logger.Info("Payslip JSON import: completed", "imported", imported, "skipped", skipped)
	}
}

func loadEnableBankingKey(logger *slog.Logger) (*rsa.PrivateKey, error) {
	keyString := os.Getenv("ENABLEBANKING_PRIVATE_KEY")
	if keyString != "" {
		// Repariert escappte Zeilenumbrüche, die in manchen .env Setups entstehen
		keyString = strings.ReplaceAll(keyString, "\\n", "\n")
		return parseRSAPem([]byte(keyString))
	}

	keyPath := os.Getenv("ENABLEBANKING_PRIVATE_KEY_PATH")
	if keyPath != "" {
		keyBytes, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file at %s: %w", keyPath, err)
		}
		return parseRSAPem(keyBytes)
	}

	return nil, errors.New("neither ENABLEBANKING_PRIVATE_KEY nor ENABLEBANKING_PRIVATE_KEY_PATH is set")
}

func parseRSAPem(keyBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block - check your key format")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key as PKCS1 or PKCS8: %w", err)
	}

	key, ok := k.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("key is not a valid RSA private key")
	}

	return key, nil
}

func seedInMemoryData(ctx context.Context, userRepo port.UserRepository, catRepo port.CategoryRepository, logger *slog.Logger, bankRepo port.BankRepository, bankStmtRepo port.BankStatementRepository, invoiceRepo port.InvoiceRepository, payslipRepo port.PayslipRepository) {
	logger.Info("Seeding in-memory default data...")

	// 1. Seed Users (admin/admin and test/test)
	var adminID uuid.UUID
	users := []struct {
		username string
		password string
		role     string
	}{
		{"admin", "admin", "admin"},
		{"test", "test", "manager"},
	}

	for _, u := range users {
		hash, _ := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		id := uuid.New()
		if u.username == "admin" {
			adminID = id
		}
		user := entity.User{
			ID:           id,
			Username:     u.username,
			PasswordHash: string(hash),
			Email:        u.username + "@localhost",
			FullName:     u.username + " User",
			Role:         u.role,
		}
		_ = userRepo.Upsert(ctx, user)
	}

	// 2. Seed default categories
	categories := []struct {
		name  string
		color string
	}{
		{"Salary", "#4caf50"},
		{"Rent", "#f44336"},
		{"Groceries", "#ff9800"},
		{"Insurance", "#2196f3"},
		{"Leisure", "#9c27b0"},
		{"Internal Transfer", "#607d8b"},
		{"Tech & Software", "#3b82f6"},
	}

	catMap := make(map[string]uuid.UUID)
	for _, c := range categories {
		cat, _ := catRepo.Save(ctx, entity.Category{
			ID:    uuid.New(),
			Name:  c.name,
			Color: c.color,
		})
		catMap[c.name] = cat.ID
	}

	// 3. Seed Bank Connections & Accounts
	conn := entity.BankConnection{
		ID:              uuid.New(),
		UserID:          adminID,
		InstitutionID:   "SANDBOX_ID",
		InstitutionName: "Sandbox Bank",
		Provider:        "gocardless",
		Status:          entity.StatusLinked,
	}
	_ = bankRepo.CreateConnection(ctx, &conn)

	acc := entity.BankAccount{
		ID:                uuid.New(),
		ConnectionID:      conn.ID,
		ProviderAccountID: "dummy_acc_id",
		IBAN:              "DE12345678901234567890",
		Name:              "Main Giro",
		Currency:          "EUR",
		Balance:           5240.50,
		LastSyncedAt:      time.Now(),
		AccountType:       entity.StatementTypeGiro,
	}
	_ = bankRepo.UpsertAccounts(ctx, []entity.BankAccount{acc})

	// 4. Seed Bank Statements and Transactions (last 3 months)
	now := time.Now()
	for i := 0; i < 3; i++ {
		monthDate := now.AddDate(0, -i, 0)
		stmtDate := time.Date(monthDate.Year(), monthDate.Month(), 28, 0, 0, 0, 0, time.UTC)

		stmt := entity.BankStatement{
			ID:            uuid.New(),
			AccountHolder: "Max Mustermann",
			IBAN:          acc.IBAN,
			StatementDate: stmtDate,
			StatementNo:   100 - i,
			Currency:      "EUR",
			StatementType: entity.StatementTypeGiro,
			ContentHash:   uuid.New().String(),
			ImportedAt:    time.Now(),
		}

		salaryCat := catMap["Salary"]
		rentCat := catMap["Rent"]
		groceriesCat := catMap["Groceries"]
		techCat := catMap["Tech & Software"]

		stmt.Transactions = []entity.Transaction{
			{
				ID:            uuid.New(),
				BankAccountID: &acc.ID,
				BookingDate:   stmtDate.AddDate(0, 0, -27),
				Description:   "Salary Mustermann GmbH",
				Amount:        3500.00,
				Currency:      "EUR",
				Type:          entity.TransactionTypeCredit,
				CategoryID:    &salaryCat,
				ContentHash:   uuid.New().String(),
				StatementType: entity.StatementTypeGiro,
				Reviewed:      true,
			},
			{
				ID:            uuid.New(),
				BankAccountID: &acc.ID,
				BookingDate:   stmtDate.AddDate(0, 0, -25),
				Description:   "Rent Payment",
				Amount:        -1200.00,
				Currency:      "EUR",
				Type:          entity.TransactionTypeDebit,
				CategoryID:    &rentCat,
				ContentHash:   uuid.New().String(),
				StatementType: entity.StatementTypeGiro,
				Reviewed:      true,
			},
			{
				ID:            uuid.New(),
				BankAccountID: &acc.ID,
				BookingDate:   stmtDate.AddDate(0, 0, -15),
				Description:   "REWE Supermarket",
				Amount:        -85.40,
				Currency:      "EUR",
				Type:          entity.TransactionTypeDebit,
				CategoryID:    &groceriesCat,
				ContentHash:   uuid.New().String(),
				StatementType: entity.StatementTypeGiro,
				Reviewed:      true,
			},
			{
				ID:            uuid.New(),
				BankAccountID: &acc.ID,
				BookingDate:   stmtDate.AddDate(0, 0, -10),
				Description:   "Hetzner Online GmbH",
				Amount:        -42.15,
				Currency:      "EUR",
				Type:          entity.TransactionTypeDebit,
				CategoryID:    &techCat,
				ContentHash:   uuid.New().String(),
				StatementType: entity.StatementTypeGiro,
				Reviewed:      true,
			},
		}

		_ = bankStmtRepo.Save(ctx, stmt)

		// 5. Seed Invoices
		_ = invoiceRepo.Save(ctx, entity.Invoice{
			ID:          uuid.New(),
			Vendor:      entity.Vendor{Name: "Hetzner Online GmbH"},
			Amount:      42.15,
			Currency:    "EUR",
			IssuedAt:    stmtDate.AddDate(0, 0, -10),
			Description: "Cloud Server Invoice",
			CategoryID:  &techCat,
			ContentHash: uuid.New().String(),
		})

		// 6. Seed Payslips
		_ = payslipRepo.Save(ctx, &entity.Payslip{
			ID:               uuid.New().String(),
			OriginalFileName: fmt.Sprintf("Payslip_%d_%02d.pdf", stmtDate.Year(), stmtDate.Month()),
			PeriodMonthNum:   int(stmtDate.Month()),
			PeriodYear:       stmtDate.Year(),
			EmployeeName:     "Max Mustermann",
			GrossPay:         5500.00,
			NetPay:           3500.00,
			PayoutAmount:     3500.00,
			ContentHash:      uuid.New().String(),
		})
	}
}
