package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

const maxDocuments = 500

type DocumentRepository struct {
	mu        sync.RWMutex
	documents map[uuid.UUID]entity.Document
	order     []uuid.UUID
}

func NewDocumentRepository() *DocumentRepository {
	r := &DocumentRepository{
		documents: make(map[uuid.UUID]entity.Document),
		order:     make([]uuid.UUID, 0, maxDocuments),
	}
	r.seedData()
	return r
}

func (r *DocumentRepository) seedData() {
	userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")

	// Employment contract
	contractID := uuid.New()
	contract := entity.Document{
		ID:               contractID,
		UserID:           userID,
		Type:             entity.DocTypeContract,
		OriginalFileName: "Employment_Contract_AcmeCorp.pdf",
		MimeType:         "application/pdf",
		ExtractedText:    "Employment Contract Acme Corp Full Time position starting 2021...",
		CreatedAt:        time.Date(2020, 12, 1, 10, 0, 0, 0, time.UTC),
	}
	r.documents[contractID] = contract
	r.order = append(r.order, contractID)

	// Yearly Tax Certificates matching payslips
	for year := 2021; year <= 2023; year++ {
		docID := uuid.New()
		doc := entity.Document{
			ID:               docID,
			UserID:           userID,
			Type:             entity.DocTypeTaxCertificate,
			OriginalFileName: fmt.Sprintf("Tax_Certificate_%d.pdf", year),
			MimeType:         "application/pdf",
			ExtractedText:    fmt.Sprintf("Annual Income Tax Certificate for the year %d", year),
			CreatedAt:        time.Date(year+1, 2, 15, 10, 0, 0, 0, time.UTC),
		}
		r.documents[docID] = doc
		r.order = append(r.order, docID)
	}
}

func (r *DocumentRepository) Save(ctx context.Context, doc entity.Document) (entity.Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Enforce unique hash per user
	for _, existing := range r.documents {
		if existing.UserID == doc.UserID && existing.ContentHash == doc.ContentHash {
			return entity.Document{}, entity.ErrDocumentDuplicate
		}
	}

	if doc.ID == uuid.Nil {
		doc.ID = uuid.New()
	}
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now()
	}

	if _, exists := r.documents[doc.ID]; !exists {
		if len(r.order) >= maxDocuments {
			// Evict oldest
			oldestID := r.order[0]
			delete(r.documents, oldestID)
			r.order = r.order[1:]
		}
		r.order = append(r.order, doc.ID)
	}

	r.documents[doc.ID] = doc
	return doc, nil
}

func (r *DocumentRepository) FindAll(ctx context.Context, filter entity.DocumentFilter) ([]entity.Document, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []entity.Document
	for _, id := range r.order {
		doc := r.documents[id]
		if doc.UserID != filter.UserID {
			continue
		}

		// In Postgres, we only return these types implicitly
		validType := doc.Type == entity.DocTypeTaxCertificate ||
			doc.Type == entity.DocTypeReceipt ||
			doc.Type == entity.DocTypeContract ||
			doc.Type == entity.DocTypeOther
		if !validType {
			continue
		}

		if filter.Type != "" && doc.Type != filter.Type {
			continue
		}

		if filter.Search != "" {
			searchLower := strings.ToLower(filter.Search)
			nameMatch := strings.Contains(strings.ToLower(doc.OriginalFileName), searchLower)
			textMatch := strings.Contains(strings.ToLower(doc.ExtractedText), searchLower)
			if !nameMatch && !textMatch {
				continue
			}
		}

		// Ensure we don't return the raw binary content in list views for memory parity
		cleanDoc := doc
		cleanDoc.OriginalFileContent = nil
		result = append(result, cleanDoc)
	}

	// Reverse to simulate ORDER BY created_at DESC
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result, nil
}

func (r *DocumentRepository) FindByID(ctx context.Context, id, userID uuid.UUID) (entity.Document, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	doc, ok := r.documents[id]
	if !ok || doc.UserID != userID {
		return entity.Document{}, entity.ErrDocumentNotFound
	}
	return doc, nil
}

func (r *DocumentRepository) Update(ctx context.Context, doc entity.Document) (entity.Document, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	old, ok := r.documents[doc.ID]
	if !ok || old.UserID != doc.UserID {
		return entity.Document{}, entity.ErrDocumentNotFound
	}

	// Retain immutable fields
	doc.ContentHash = old.ContentHash
	doc.MimeType = old.MimeType
	doc.ExtractedText = old.ExtractedText
	doc.OriginalFileContent = old.OriginalFileContent
	doc.CreatedAt = old.CreatedAt

	r.documents[doc.ID] = doc
	return doc, nil
}

func (r *DocumentRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	doc, ok := r.documents[id]
	if !ok || doc.UserID != userID {
		return entity.ErrDocumentNotFound
	}

	delete(r.documents, id)

	// Remove from order slice
	for i, orderedID := range r.order {
		if orderedID == id {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}

	return nil
}

func (r *DocumentRepository) ExistsByHash(ctx context.Context, userID uuid.UUID, contentHash string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, doc := range r.documents {
		if doc.UserID == userID && doc.ContentHash == contentHash {
			return true, nil
		}
	}
	return false, nil
}

var _ port.DocumentRepository = (*DocumentRepository)(nil)
