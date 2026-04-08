// cmd/resetpw/main.go
// Standalone helper: bcrypt-hash a password and UPDATE it in the users table.
//
// Usage (via Makefile):
//   make db-reset-password USER=admin PASSWORD=newpass
//
// Direct usage:
//   cd backend && go run ./cmd/resetpw -user admin -password newpass
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx/v5"
)

func main() {
	user := flag.String("user", "", "Username to update (required)")
	pass := flag.String("password", "", "New plain-text password (required)")
	flag.Parse()

	if *user == "" || *pass == "" {
		fmt.Fprintln(os.Stderr, "Usage: resetpw -user <username> -password <newpassword>")
		os.Exit(1)
	}

	// ── bcrypt hash ──────────────────────────────────────────────────────────
	hash, err := bcrypt.GenerateFromPassword([]byte(*pass), 12)
	if err != nil {
		log.Fatalf("bcrypt: %v", err)
	}

	// ── DB connection (same env vars as the main app) ─────────────────────────
	host := envOr("DATABASE_HOST", "127.0.0.1")
	port := envOr("DATABASE_PORT", "5432")
	dbUser := envOr("POSTGRES_USER", "")
	dbPass := envOr("POSTGRES_PASSWORD", "")
	dbName := envOr("POSTGRES_DB", "")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPass, host, port, dbName)

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer conn.Close(context.Background())

	// ── UPDATE ────────────────────────────────────────────────────────────────
	tag, err := conn.Exec(context.Background(),
		`UPDATE users SET password_hash = $1 WHERE username = $2`,
		string(hash), *user,
	)
	if err != nil {
		log.Fatalf("update: %v", err)
	}

	if tag.RowsAffected() == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: user %q not found in the database\n", *user)
		os.Exit(1)
	}

	fmt.Printf("✅  Password for %q updated successfully.\n", *user)
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

