package postgres

import (
	"context"
	"errors"
	"fmt"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DocumentRepository struct {
	db  *pgxpool.Pool
	key string
}

func NewDocumentRepository(db *pgxpool.Pool, key string) *DocumentRepository {
	return &DocumentRepository{db: db, key: key}
}

func (r *DocumentRepository) Save(ctx context.Context, doc entity.Document) (entity.Document, error) {
	query := `
		INSERT INTO documents (
			user_id, document_type, original_file_name, original_file_content,
			content_hash, mime_type, extracted_text, metadata
		) VALUES (
			$1, $2, $3, pgp_sym_encrypt_bytea($4, $5), $6, $7, $8, $9
		) RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		doc.UserID, doc.Type, doc.OriginalFileName, doc.OriginalFileContent, r.key,
		doc.ContentHash, doc.MimeType, doc.ExtractedText, doc.Metadata,
	).Scan(&doc.ID, &doc.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return entity.Document{}, entity.ErrDocumentDuplicate
		}
		return entity.Document{}, fmt.Errorf("document repo save: %w", err)
	}

	return doc, nil
}

func (r *DocumentRepository) FindAll(ctx context.Context, filter entity.DocumentFilter) ([]entity.Document, error) {
	query := `
		SELECT id, user_id, document_type, original_file_name,
		       content_hash, mime_type, metadata, created_at
		FROM documents
		WHERE user_id = $1
		AND document_type IN ('tax_certificate', 'receipt', 'contract', 'other')
	`
	args := []interface{}{filter.UserID}

	if filter.Type != "" {
		args = append(args, filter.Type)
		query += fmt.Sprintf(" AND document_type = $%d", len(args))
	}

	if filter.Search != "" {
		args = append(args, filter.Search)
		// Use gin_trgm_ops index for extracted_text and also search file name
		query += fmt.Sprintf(" AND (extracted_text ILIKE '%%' || $%d || '%%' OR original_file_name ILIKE '%%' || $%d || '%%')", len(args), len(args))
	}

	query += " ORDER BY COALESCE((metadata->>'date')::date, created_at::date) DESC, created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("document repo find all: %w", err)
	}
	defer rows.Close()

	var documents []entity.Document
	for rows.Next() {
		var doc entity.Document
		err := rows.Scan(
			&doc.ID, &doc.UserID, &doc.Type, &doc.OriginalFileName,
			&doc.ContentHash, &doc.MimeType, &doc.Metadata, &doc.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("document repo find all scan: %w", err)
		}
		documents = append(documents, doc)
	}
	return documents, nil
}

func (r *DocumentRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (entity.Document, error) {
	query := `
		SELECT id, user_id, document_type, original_file_name, original_file_content,
		       content_hash, mime_type, metadata, extracted_text, created_at
		FROM documents
		WHERE id = $1 AND user_id = $2
	`
	var doc entity.Document
	var rawContent []byte
	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&doc.ID, &doc.UserID, &doc.Type, &doc.OriginalFileName, &rawContent,
		&doc.ContentHash, &doc.MimeType, &doc.Metadata, &doc.ExtractedText, &doc.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Document{}, entity.ErrDocumentNotFound
		}
		return entity.Document{}, fmt.Errorf("document repo find by id: %w", err)
	}

	if len(rawContent) == 0 {
		return doc, nil
	}

	// Try decryption
	var decrypted []byte
	decryptQuery := "SELECT pgp_sym_decrypt_bytea($1, $2)"
	err = r.db.QueryRow(ctx, decryptQuery, rawContent, r.key).Scan(&decrypted)
	if err != nil {
		// Fallback: If decryption fails, it might be legacy plain text or corrupt.
		// Return raw content but log a loud warning.
		// We use fmt.Printf because this is specifically for stdout/stderr visibility in logs.
		fmt.Printf("WARNING: DOCUMENT DECRYPTION FAILED for id %s. Error: %v. This usually means the DOCUMENT_VAULT_KEY has changed or is incorrect.\n", id, err)
		doc.OriginalFileContent = rawContent
	} else {
		doc.OriginalFileContent = decrypted
	}

	return doc, nil
}

func (r *DocumentRepository) Update(ctx context.Context, doc entity.Document) (entity.Document, error) {
	query := `
		UPDATE documents
		SET document_type = $1,
		    original_file_name = $2,
		    metadata = $3,
		    original_file_content = COALESCE(pgp_sym_encrypt_bytea($4, $5), original_file_content),
		    content_hash = COALESCE($6, content_hash),
		    mime_type = COALESCE($7, mime_type),
		    extracted_text = COALESCE($8, extracted_text)
		WHERE id = $9 AND user_id = $10
		RETURNING id, created_at
	`

	var fileContent []byte
	if len(doc.OriginalFileContent) > 0 {
		fileContent = doc.OriginalFileContent
	}

	var contentHash, mimeType, extractedText *string
	if doc.ContentHash != "" {
		contentHash = &doc.ContentHash
	}
	if doc.MimeType != "" {
		mimeType = &doc.MimeType
	}
	if doc.ExtractedText != "" {
		extractedText = &doc.ExtractedText
	}

	err := r.db.QueryRow(ctx, query,
		doc.Type, doc.OriginalFileName, doc.Metadata,
		fileContent, r.key,
		contentHash, mimeType, extractedText,
		doc.ID, doc.UserID,
	).Scan(&doc.ID, &doc.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Document{}, entity.ErrDocumentNotFound
		}
		return entity.Document{}, fmt.Errorf("document repo update: %w", err)
	}

	return doc, nil
}

func (r *DocumentRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	cmd, err := r.db.Exec(ctx, `DELETE FROM documents WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("document repo delete: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrDocumentNotFound
	}
	return nil
}

func (r *DocumentRepository) ExistsByHash(ctx context.Context, userID uuid.UUID, contentHash string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM documents WHERE user_id = $1 AND content_hash = $2)`
	err := r.db.QueryRow(ctx, query, userID, contentHash).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("document repo exists by hash: %w", err)
	}
	return exists, nil
}
