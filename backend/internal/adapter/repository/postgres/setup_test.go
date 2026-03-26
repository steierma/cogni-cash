package postgres

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"cogni-cash/migrations"
)

var globalPool *pgxpool.Pool

// TestMain runs once before any tests in the package are executed.
func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := tcpg.Run(ctx,
		"postgres:16-alpine",
		tcpg.WithDatabase("financetest"),
		tcpg.WithUsername("testuser"),
		tcpg.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(15*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	globalPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("failed to open database pool: %v", err)
	}

	// Run migrations once
	runMigrations(ctx, globalPool)

	// Run all tests
	code := m.Run()

	// Teardown
	globalPool.Close()
	if err := pgContainer.Terminate(ctx); err != nil {
		log.Fatalf("failed to terminate postgres container: %v", err)
	}

	os.Exit(code)
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) {
	files, err := migrations.FS.ReadDir(".")
	if err != nil {
		log.Fatalf("failed to read embedded migrations: %v", err)
	}

	var migrationFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") && !strings.Contains(f.Name(), ".down.") {
			migrationFiles = append(migrationFiles, f.Name())
		}
	}
	sort.Strings(migrationFiles)

	for _, file := range migrationFiles {
		content, err := migrations.FS.ReadFile(file)
		if err != nil {
			log.Fatalf("failed to read migration file %s: %v", file, err)
		}

		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			log.Fatalf("failed to execute migration %s: %v", file, err)
		}
	}
}

func setupLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// clearTables truncates all application tables to ensure a clean slate for each test.
func clearTables(ctx context.Context, t *testing.T) {
	t.Helper()
	// TRUNCATE CASCADE ensures foreign key constraints don't block the truncation
	tables := []string{
		"bank_statements",
		"transactions",
		"categories",
		"reconciliations",
		"invoices",
		"payslips",
		"payslip_bonuses", // <-- Updated table name
	}
	query := "TRUNCATE TABLE " + strings.Join(tables, ", ") + " CASCADE;"
	_, err := globalPool.Exec(ctx, query)
	if err != nil {
		t.Fatalf("failed to clear tables: %v", err)
	}
}
