package memory

import (
	"context"
	"strings"
	"sync"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

const (
	maxStatements   = 100
	maxTransactions = 5000
)

type BankStatementRepository struct {
	mu           sync.RWMutex
	statements   map[uuid.UUID]entity.BankStatement
	stmtOrder    []uuid.UUID
	transactions map[string]entity.Transaction // keyed by ContentHash
	txOrder      []string                      // for standalone transactions
	categoryRepo port.CategoryRepository
}

func NewBankStatementRepository() *BankStatementRepository {
	return &BankStatementRepository{
		statements:   make(map[uuid.UUID]entity.BankStatement),
		stmtOrder:    make([]uuid.UUID, 0, maxStatements),
		transactions: make(map[string]entity.Transaction),
		txOrder:      make([]string, 0, maxTransactions),
	}
}

func (r *BankStatementRepository) WithCategoryRepository(catRepo port.CategoryRepository) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.categoryRepo = catRepo
}

func (r *BankStatementRepository) Save(ctx context.Context, stmt entity.BankStatement) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if stmt.ID == uuid.Nil {
		stmt.ID = uuid.New()
	}

	if _, exists := r.statements[stmt.ID]; !exists {
		if len(r.stmtOrder) >= maxStatements {
			// Evict oldest statement and its transactions
			oldestID := r.stmtOrder[0]
			r.deleteStatement(oldestID)
			r.stmtOrder = r.stmtOrder[1:]
		}
		r.stmtOrder = append(r.stmtOrder, stmt.ID)
	}

	r.statements[stmt.ID] = stmt
	for _, tx := range stmt.Transactions {
		tx.BankStatementID = &stmt.ID
		r.transactions[tx.ContentHash] = tx
	}
	return nil
}

func (r *BankStatementRepository) deleteStatement(id uuid.UUID) {
	delete(r.statements, id)
	for hash, tx := range r.transactions {
		if tx.BankStatementID != nil && *tx.BankStatementID == id {
			delete(r.transactions, hash)
		}
	}
}

func (r *BankStatementRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (entity.BankStatement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	stmt, ok := r.statements[id]
	if !ok || stmt.UserID != userID {
		return entity.BankStatement{}, entity.ErrBankStatementNotFound
	}
	return stmt, nil
}

func (r *BankStatementRepository) FindAll(ctx context.Context, userID uuid.UUID) ([]entity.BankStatement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var stmts []entity.BankStatement
	for _, stmt := range r.statements {
		if stmt.UserID == userID {
			stmts = append(stmts, stmt)
		}
	}
	return stmts, nil
}

func (r *BankStatementRepository) FindSummaries(ctx context.Context, userID uuid.UUID) ([]entity.BankStatementSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var summaries []entity.BankStatementSummary
	for _, stmt := range r.statements {
		if stmt.UserID == userID {
			summaries = append(summaries, entity.BankStatementSummary{
				ID:               stmt.ID,
				StatementNo:      stmt.StatementNo,
				IBAN:             stmt.IBAN,
				Currency:         stmt.Currency,
				NewBalance:       stmt.NewBalance,
				TransactionCount: len(stmt.Transactions),
				StatementType:    stmt.StatementType,
				HasOriginalFile:  len(stmt.OriginalFile) > 0,
			})
		}
	}
	return summaries, nil
}

func (r *BankStatementRepository) FindTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var txns []entity.Transaction
	for _, tx := range r.transactions {
		if tx.UserID == filter.UserID && r.matchFilter(tx, filter) {
			txns = append(txns, tx)
		}
	}
	return txns, nil
}

func (r *BankStatementRepository) SearchTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
	return r.FindTransactions(ctx, filter)
}

func (r *BankStatementRepository) GetCategorizationExamples(ctx context.Context, userID uuid.UUID, examplesPerCategory int) ([]entity.CategorizationExample, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// map[categoryID][]entity.CategorizationExample
	catExamples := make(map[uuid.UUID][]entity.CategorizationExample)

	for _, tx := range r.transactions {
		if tx.CategoryID == nil || tx.UserID != userID {
			continue
		}

		if _, ok := catExamples[*tx.CategoryID]; !ok {
			catExamples[*tx.CategoryID] = make([]entity.CategorizationExample, 0)
		}

		if len(catExamples[*tx.CategoryID]) < examplesPerCategory {
			// Check for uniqueness based on description + reference
			isDuplicate := false
			for _, existing := range catExamples[*tx.CategoryID] {
				if existing.Description == tx.Description && existing.Reference == tx.Reference {
					isDuplicate = true
					break
				}
			}
			if !isDuplicate {
				catExamples[*tx.CategoryID] = append(catExamples[*tx.CategoryID], entity.CategorizationExample{
					Description:         tx.Description,
					Reference:           tx.Reference,
					CounterpartyName:    tx.CounterpartyName,
					CounterpartyIban:    tx.CounterpartyIban,
					BankTransactionCode: tx.BankTransactionCode,
					MandateReference:    tx.MandateReference,
				})
			}
		}
	}

	var examples []entity.CategorizationExample
	for catID, exList := range catExamples {
		catName := catID.String()
		if r.categoryRepo != nil {
			if c, err := r.categoryRepo.FindByID(ctx, catID, userID); err == nil {
				catName = c.Name
			}
		}

		for _, ex := range exList {
			ex.Category = catName
			examples = append(examples, ex)
		}
	}

	return examples, nil
}

func (r *BankStatementRepository) UpdateTransactionCategory(ctx context.Context, hash string, categoryID *uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tx, ok := r.transactions[hash]
	if !ok || tx.UserID != userID {
		return entity.ErrTransactionNotFound
	}
	tx.CategoryID = categoryID
	r.transactions[hash] = tx
	return nil
}

func (r *BankStatementRepository) MarkTransactionReviewed(ctx context.Context, hash string, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tx, ok := r.transactions[hash]
	if !ok || tx.UserID != userID {
		return entity.ErrTransactionNotFound
	}
	tx.Reviewed = true
	r.transactions[hash] = tx
	return nil
}

func (r *BankStatementRepository) MarkTransactionReconciled(ctx context.Context, contentHash string, reconciliationID uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tx, ok := r.transactions[contentHash]
	if !ok || tx.UserID != userID {
		return entity.ErrTransactionNotFound
	}
	tx.ReconciliationID = &reconciliationID
	tx.IsReconciled = true
	tx.Reviewed = true
	r.transactions[contentHash] = tx
	return nil
}

func (r *BankStatementRepository) LinkTransactionToStatement(ctx context.Context, id uuid.UUID, statementID uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for hash, tx := range r.transactions {
		if tx.ID == id && tx.UserID == userID {
			tx.BankStatementID = &statementID
			r.transactions[hash] = tx
			return nil
		}
	}
	return entity.ErrTransactionNotFound
}

func (r *BankStatementRepository) CreateTransactions(ctx context.Context, txns []entity.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, tx := range txns {
		if _, exists := r.transactions[tx.ContentHash]; !exists {
			if len(r.txOrder) >= maxTransactions {
				// Evict oldest standalone transaction
				oldestHash := r.txOrder[0]
				delete(r.transactions, oldestHash)
				r.txOrder = r.txOrder[1:]
			}
			r.txOrder = append(r.txOrder, tx.ContentHash)
		}
		r.transactions[tx.ContentHash] = tx
	}
	return nil
}

func (r *BankStatementRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stmt, ok := r.statements[id]
	if !ok || stmt.UserID != userID {
		return entity.ErrBankStatementNotFound
	}

	r.deleteStatement(id)

	// Update stmtOrder
	for i, sid := range r.stmtOrder {
		if sid == id {
			r.stmtOrder = append(r.stmtOrder[:i], r.stmtOrder[i+1:]...)
			break
		}
	}

	return nil
}

func (r *BankStatementRepository) matchFilter(tx entity.Transaction, filter entity.TransactionFilter) bool {
	if filter.CategoryID != nil && (tx.CategoryID == nil || *tx.CategoryID != *filter.CategoryID) {
		return false
	}
	if filter.IsReconciled != nil && tx.IsReconciled != *filter.IsReconciled {
		return false
	}
	if filter.Reviewed != nil && tx.Reviewed != *filter.Reviewed {
		return false
	}
	if filter.Search != "" {
		search := strings.ToLower(filter.Search)
		if !strings.Contains(strings.ToLower(tx.Description), search) &&
			!strings.Contains(strings.ToLower(tx.Reference), search) {
			return false
		}
	}
	// Add more filter matches as needed (dates, amount, etc.)
	return true
}

var _ port.BankStatementRepository = (*BankStatementRepository)(nil)
