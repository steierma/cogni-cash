package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	httpAdapter "cogni-cash/internal/adapter/http"
	llmadapter "cogni-cash/internal/adapter/ollama"
	aiparser "cogni-cash/internal/adapter/parser/bank_statement/ai"
	amazonvisaparser "cogni-cash/internal/adapter/parser/bank_statement/amazonvisa"
	ingparser "cogni-cash/internal/adapter/parser/bank_statement/ing"
	ingcsvparser "cogni-cash/internal/adapter/parser/bank_statement/ingcsv"
	vwparser "cogni-cash/internal/adapter/parser/bank_statement/vw"
	payslipaiparser "cogni-cash/internal/adapter/parser/payslip/ai"
	cariadparser "cogni-cash/internal/adapter/parser/payslip/cariad"
	pgrepo "cogni-cash/internal/adapter/repository/postgres"
	"cogni-cash/internal/domain/port"
	"cogni-cash/internal/domain/service"
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
	dbUser := envOrDefault("POSTGRES_USER", "")
	dbPassword := envOrDefault("POSTGRES_PASSWORD", "")
	dbHost := envOrDefault("DATABASE_HOST", "localhost")
	dbPort := envOrDefault("DATABASE_PORT", "5432")
	dbName := envOrDefault("POSTGRES_DB", "")

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	logger.Info("Attempting to connect to PostgreSQL...", "host", dbHost, "db", dbName)
	pool, err := pgrepo.NewPool(appCtx, databaseURL)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL. A database connection is strictly required. Exiting.", "error", err)
		os.Exit(1)
	}
	logger.Info("Successfully connected to PostgreSQL")

	storageMode := "postgres"
	dbPinger := func(ctx context.Context) error { return pool.Ping(ctx) }

	// --- Instantiate Repositories ---
	invoiceRepo := pgrepo.NewInvoiceRepository(pool, logger)
	bankStmtRepo := pgrepo.NewBankStatementRepository(pool, logger)
	categoryRepo := pgrepo.NewCategoryRepository(pool, logger)
	userRepo := pgrepo.NewUserRepository(pool, logger)
	reconciliationRepo := pgrepo.NewReconciliationRepository(pool, bankStmtRepo, logger)
	settingsRepo := pgrepo.NewSettingsRepository(pool, logger)
	payslipRepo := pgrepo.NewPayslipRepository(pool)

	ensureDefaultSettings(appCtx, settingsRepo, logger)

	// --- Instantiate Services ---
	authSvc := service.NewAuthService(userRepo, jwtSecret, logger)
	if err := authSvc.EnsureAdminUser(appCtx, adminUsername, adminPassword); err != nil {
		logger.Error("Failed to seed default admin user", "error", err)
	}

	userSvc := service.NewUserService(userRepo)
	llmClient := llmadapter.NewAdapter(settingsRepo, logger)
	settingsSvc := service.NewSettingsService(settingsRepo, logger)
	categorizationSvc := service.NewCategorizationService(llmClient, invoiceRepo, nil, logger)

	bankStatementSvc := service.NewBankStatementService(bankStmtRepo, logger)
	transactionSvc := service.NewTransactionService(bankStmtRepo, categoryRepo, llmClient, logger)
	reconciliationSvc := service.NewReconciliationService(bankStmtRepo, reconciliationRepo, logger)

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

	// --- HTTP Handler Setup ---
	handler := httpAdapter.NewHandler(authSvc, categorizationSvc, bankStatementSvc, settingsSvc, invoiceRepo, logger, storageMode, dbHost, dbPinger)

	handler.WithUserService(userSvc)
	handler.WithTransactionService(transactionSvc)
	handler.WithReconciliationService(reconciliationSvc)
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

	// 1. Cancel root context to stop background workers
	appCancel()
	logger.Info("Signalled background workers to stop")

	// 2. Stop accepting new HTTP requests and finish active ones
	shutCtx, srvCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer srvCancel()

	logger.Info("Shutting down HTTP server...")
	if err := srv.Shutdown(shutCtx); err != nil {
		logger.Error("HTTP server shutdown forced or timed out", "error", err)
	} else {
		logger.Info("HTTP server shut down cleanly")
	}

	// 3. Close database connections synchronously
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
		"payslip_import_json_path":       envOrDefault("PAYSLIP_IMPORT_JSON_PATH", ""),
		"payslip_import_interval":        envOrDefault("PAYSLIP_IMPORT_INTERVAL", "1h"),
	}

	// Keys that are always synced from their env var to the DB — even when the
	// env var is empty. This ensures that switching environments (e.g. Docker →
	// local) never leaves a stale path in the DB from a previous run.
	envOverrideKeys := map[string]string{
		"payslip_import_json_path": "PAYSLIP_IMPORT_JSON_PATH",
		"payslip_import_interval":  "PAYSLIP_IMPORT_INTERVAL",
		"import_dir":               "IMPORT_DIR",
		"import_interval":          "IMPORT_INTERVAL",
	}

	for k, v := range defaults {
		existing, _ := repo.Get(ctx, k)

		// If this key has a dedicated env var, always write it so the DB stays
		// in sync with the runtime environment (even when the value is "").
		if envKey, ok := envOverrideKeys[k]; ok {
			envVal, envSet := os.LookupEnv(envKey)
			if envSet && envVal != existing {
				if err := repo.Set(ctx, k, envVal); err != nil {
					logger.Warn("Failed to sync env setting to DB", "key", k, "error", err)
				}
			}
			continue // skip the "only write when empty" logic below
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
