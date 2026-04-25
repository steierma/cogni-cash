package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run vault_check.go <DOCUMENT_ID> <VAULT_KEY>")
		os.Exit(1)
	}

	docID := os.Args[1]
	key := os.Args[2]

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Try to construct from parts if URL is missing
		host := os.Getenv("DATABASE_HOST")
		user := os.Getenv("POSTGRES_USER")
		pass := os.Getenv("POSTGRES_PASSWORD")
		dbName := os.Getenv("POSTGRES_DB")
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:5432/%s", user, pass, host, dbName)
	}

	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	var rawContent []byte
	err = conn.QueryRow(context.Background(), "SELECT original_file_content FROM documents WHERE id = $1", docID).Scan(&rawContent)
	if err != nil {
		log.Fatalf("Failed to fetch document: %v", err)
	}

	fmt.Printf("Document found. Encrypted size: %d bytes\n", len(rawContent))

	var decrypted []byte
	err = conn.QueryRow(context.Background(), "SELECT pgp_sym_decrypt_bytea($1, $2)", rawContent, key).Scan(&decrypted)
	if err != nil {
		fmt.Printf("\n❌ DECRYPTION FAILED!\nError: %v\n", err)
		fmt.Println("This key is NOT the one used to encrypt this document.")
	} else {
		fmt.Printf("\n✅ DECRYPTION SUCCESSFUL!\nDecrypted size: %d bytes\n", len(decrypted))
		if len(decrypted) > 4 && string(decrypted[:4]) == "%PDF" {
			fmt.Println("Format: Valid PDF")
		}
	}
}
