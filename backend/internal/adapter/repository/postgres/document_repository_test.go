package postgres

import (
	"context"
	"testing"

	"cogni-cash/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocumentRepository(t *testing.T) {
	ctx := context.Background()
	clearTables(ctx, t)

	// Setup: Need a user
	userID := uuid.New()
	_, err := globalPool.Exec(ctx, `INSERT INTO users (id, username, email, password_hash) VALUES ($1, 'testuser', 'test@example.com', 'hash')`, userID)
	require.NoError(t, err)

	encryptionKey := "my-secret-key"
	repo := NewDocumentRepository(globalPool, encryptionKey)

	t.Run("Save and FindByID with binary content", func(t *testing.T) {
		// Binary content that contains non-UTF8 bytes (e.g., 0xbf, 0xff)
		binaryContent := []byte{0x25, 0x50, 0x44, 0x46, 0x2d, 0x31, 0x2e, 0x34, 0x0a, 0xbf, 0xff, 0x12, 0x34}

		doc := entity.Document{
			UserID:              userID,
			Type:                entity.DocTypeTaxCertificate,
			OriginalFileName:    "test.pdf",
			OriginalFileContent: binaryContent,
			ContentHash:         "hash123",
			MimeType:            "application/pdf",
			Metadata:            map[string]interface{}{"year": 2024.0},
			ExtractedText:       "Some extracted text",
		}

		// Test Save
		savedDoc, err := repo.Save(ctx, doc)
		require.NoError(t, err, "Save should handle binary content without UTF8 encoding error")
		assert.NotEqual(t, uuid.Nil, savedDoc.ID)

		// Test FindByID (verify decryption)
		foundDoc, err := repo.FindByID(ctx, savedDoc.ID, userID)
		require.NoError(t, err)
		assert.Equal(t, binaryContent, foundDoc.OriginalFileContent, "Decrypted content should match original binary content")
		assert.Equal(t, doc.OriginalFileName, foundDoc.OriginalFileName)
		assert.Equal(t, doc.Type, foundDoc.Type)
		assert.Equal(t, doc.Metadata["year"], foundDoc.Metadata["year"])
	})

	t.Run("FindAll", func(t *testing.T) {
		docs, err := repo.FindAll(ctx, entity.DocumentFilter{UserID: userID})
		require.NoError(t, err)
		assert.Len(t, docs, 1)
		// FindAll doesn't return file content for performance
		assert.Nil(t, docs[0].OriginalFileContent)
	})

	t.Run("Update", func(t *testing.T) {
		docs, _ := repo.FindAll(ctx, entity.DocumentFilter{UserID: userID})
		doc := docs[0]

		newName := "updated.pdf"
		newType := entity.DocTypeReceipt
		doc.OriginalFileName = newName
		doc.Type = newType
		doc.Metadata["notes"] = "some notes"

		updatedDoc, err := repo.Update(ctx, doc)
		require.NoError(t, err)
		assert.Equal(t, newName, updatedDoc.OriginalFileName)
		assert.Equal(t, newType, updatedDoc.Type)

		foundDoc, _ := repo.FindByID(ctx, doc.ID, userID)
		assert.Equal(t, "some notes", foundDoc.Metadata["notes"])
	})

	t.Run("ExistsByHash", func(t *testing.T) {
		exists, err := repo.ExistsByHash(ctx, userID, "hash123")
		require.NoError(t, err)
		assert.True(t, exists)

		exists, err = repo.ExistsByHash(ctx, userID, "non-existent")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Save - Duplicate Hash", func(t *testing.T) {
		// Since we deleted hash123 in the Delete test, let's re-save it first if needed,
		// or use a fresh one.
		// Actually, let's just save one and then try to save again.

		docFresh := entity.Document{
			UserID:           userID,
			Type:             entity.DocTypeOther,
			OriginalFileName: "fresh.pdf",
			ContentHash:      "fresh_hash",
			Metadata:         map[string]interface{}{},
		}
		_, err := repo.Save(ctx, docFresh)
		require.NoError(t, err)

		_, err = repo.Save(ctx, docFresh)
		assert.ErrorIs(t, err, entity.ErrDocumentDuplicate)
	})

	t.Run("FindAll with search", func(t *testing.T) {
		// Save doc for search
		docS := entity.Document{
			UserID:           userID,
			Type:             entity.DocTypeContract,
			OriginalFileName: "contract-search.pdf",
			ContentHash:      "hash-search-1",
			Metadata:         map[string]interface{}{},
		}
		_, err := repo.Save(ctx, docS)
		require.NoError(t, err)

		docs, err := repo.FindAll(ctx, entity.DocumentFilter{UserID: userID, Search: "contract-search"})
		require.NoError(t, err)
		assert.Len(t, docs, 1)
		assert.Equal(t, "contract-search.pdf", docs[0].OriginalFileName)

		// Search in extracted text
		doc3 := entity.Document{
			UserID:           userID,
			Type:             entity.DocTypeOther,
			OriginalFileName: "searchable.pdf",
			ContentHash:      "hash789",
			ExtractedText:    "The quick brown fox",
			Metadata:         map[string]interface{}{},
		}
		_, err = repo.Save(ctx, doc3)
		require.NoError(t, err)

		docs, err = repo.FindAll(ctx, entity.DocumentFilter{UserID: userID, Search: "brown fox"})
		require.NoError(t, err)
		assert.Len(t, docs, 1)
		assert.Equal(t, "searchable.pdf", docs[0].OriginalFileName)
	})

	t.Run("FindByID - Not Found", func(t *testing.T) {
		_, err := repo.FindByID(ctx, uuid.New(), userID)
		assert.ErrorIs(t, err, entity.ErrDocumentNotFound)
	})

	t.Run("FindAll with filters", func(t *testing.T) {
		newUserID := uuid.New()
		_, err := globalPool.Exec(ctx, `INSERT INTO users (id, username, email, password_hash) VALUES ($1, 'testuser2', 'test2@example.com', 'hash')`, newUserID)
		require.NoError(t, err)

		// Save docs with different types for the new user
		docC := entity.Document{
			UserID:           newUserID,
			Type:             entity.DocTypeContract,
			OriginalFileName: "contract.pdf",
			ContentHash:      "hash-c",
			Metadata:         map[string]interface{}{},
		}
		_, err = repo.Save(ctx, docC)
		require.NoError(t, err)

		docR := entity.Document{
			UserID:           newUserID,
			Type:             entity.DocTypeReceipt,
			OriginalFileName: "receipt.pdf",
			ContentHash:      "hash-r",
			Metadata:         map[string]interface{}{},
		}
		_, err = repo.Save(ctx, docR)
		require.NoError(t, err)

		// Filter by contract
		docs, err := repo.FindAll(ctx, entity.DocumentFilter{UserID: newUserID, Type: entity.DocTypeContract})
		require.NoError(t, err)
		assert.Len(t, docs, 1)
		assert.Equal(t, entity.DocTypeContract, docs[0].Type)
	})

	t.Run("Update - Not Found", func(t *testing.T) {
		doc := entity.Document{ID: uuid.New(), UserID: userID, Type: entity.DocTypeOther}
		_, err := repo.Update(ctx, doc)
		assert.ErrorIs(t, err, entity.ErrDocumentNotFound)
	})

	t.Run("Delete", func(t *testing.T) {
		docs, _ := repo.FindAll(ctx, entity.DocumentFilter{UserID: userID})
		// Find the first one to delete
		var docToDelete entity.Document
		for _, d := range docs {
			if d.ContentHash == "hash123" {
				docToDelete = d
				break
			}
		}
		err := repo.Delete(ctx, docToDelete.ID, userID)
		require.NoError(t, err)

		exists, _ := repo.ExistsByHash(ctx, userID, "hash123")
		assert.False(t, exists)
	})

	t.Run("Delete - Not Found", func(t *testing.T) {
		err := repo.Delete(ctx, uuid.New(), userID)
		assert.ErrorIs(t, err, entity.ErrDocumentNotFound)
	})
}
