// cmd/migrate/main.go
//
// Usage:
//
//	POSTGRES_USER=x POSTGRES_PASSWORD=x DATABASE_HOST=x DATABASE_PORT=5432 POSTGRES_DB=x go run ./cmd/migrate
//
// The runner reads every *.sql file from the migrations/ directory in
// lexicographic order and applies only those whose version is not yet
// recorded in schema_migrations (or whose content hash has changed).
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dir := flag.String("dir", "migrations", "directory containing *.sql migration files")
	flag.Parse()

	dbUser := envOrDefault("POSTGRES_USER", "")
	dbPassword := envOrDefault("POSTGRES_PASSWORD", "")
	dbHost := envOrDefault("DATABASE_HOST", "localhost")
	dbPort := envOrDefault("DATABASE_PORT", "5432")
	dbName := envOrDefault("POSTGRES_DB", "")

	if dbUser == "" || dbPassword == "" || dbName == "" {
		log.Fatal("migrate: POSTGRES_USER, POSTGRES_PASSWORD and POSTGRES_DB must all be set")
	}

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("migrate: connect: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("migrate: ping: %v", err)
	}
	log.Printf("Connected to database.")

	// Ensure the tracking table exists before we query it.
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version      TEXT        PRIMARY KEY,
			applied_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			content_hash TEXT        NOT NULL DEFAULT ''
		)`)
	if err != nil {
		log.Fatalf("migrate: ensure schema_migrations: %v", err)
	}

	// Add content_hash to existing installations that pre-date this column.
	_, err = pool.Exec(ctx, `
		ALTER TABLE schema_migrations
			ADD COLUMN IF NOT EXISTS content_hash TEXT NOT NULL DEFAULT ''`)
	if err != nil {
		log.Fatalf("migrate: alter schema_migrations: %v", err)
	}

	// Load already-applied versions and their stored content hashes.
	applied, err := loadApplied(ctx, pool)
	if err != nil {
		log.Fatalf("migrate: load applied: %v", err)
	}

	// Discover migration files.
	files, err := filepath.Glob(filepath.Join(*dir, "*.sql"))
	if err != nil {
		log.Fatalf("migrate: glob: %v", err)
	}
	sort.Strings(files)

	if len(files) == 0 {
		log.Printf("No migration files found in %q.", *dir)
		return
	}

	ran := 0
	for _, f := range files {
		version := versionFromFile(f)

		sql, err := os.ReadFile(f)
		if err != nil {
			log.Fatalf("migrate: read %s: %v", f, err)
		}

		currentHash := sha256Hex(sql)

		if storedHash, seen := applied[version]; seen && storedHash == currentHash {
			log.Printf("  skip  %s (already applied, hash unchanged)", version)
			continue
		} else if seen {
			log.Printf("  rerun %s (content changed — re-applying idempotent migration)", version)
		} else {
			log.Printf("  apply %s ...", version)
		}

		if err := runMigration(ctx, pool, version, string(sql), currentHash); err != nil {
			log.Fatalf("migrate: apply %s: %v", version, err)
		}
		log.Printf("  ✓     %s applied", version)
		ran++
	}

	if ran == 0 {
		log.Printf("Database schema is up to date.")
	}
}

// runMigration executes the SQL and records the version + content hash inside a single transaction.
func runMigration(ctx context.Context, pool *pgxpool.Pool, version, sql, contentHash string) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, sql); err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	// Upsert: insert on first run, update hash+timestamp on re-run.
	_, err = tx.Exec(ctx,
		`INSERT INTO schema_migrations (version, content_hash)
		 VALUES ($1, $2)
		 ON CONFLICT (version) DO UPDATE
		   SET content_hash = EXCLUDED.content_hash,
		       applied_at   = NOW()`,
		version, contentHash,
	)
	if err != nil {
		return fmt.Errorf("record: %w", err)
	}

	return tx.Commit(ctx)
}

// loadApplied returns a map of version → content_hash for all already-applied migrations.
func loadApplied(ctx context.Context, pool *pgxpool.Pool) (map[string]string, error) {
	rows, err := pool.Query(ctx, `SELECT version, content_hash FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var version, hash string
		if err := rows.Scan(&version, &hash); err != nil {
			return nil, err
		}
		result[version] = hash
	}
	return result, rows.Err()
}

// sha256Hex returns the hex-encoded SHA-256 digest of data.
func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// versionFromFile derives the migration version key from the filename,
// e.g. "migrations/001_initial_schema.sql" → "001_initial_schema".
func versionFromFile(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
