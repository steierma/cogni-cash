package memory

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

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
	r := &BankStatementRepository{
		statements:   make(map[uuid.UUID]entity.BankStatement),
		stmtOrder:    make([]uuid.UUID, 0, maxStatements),
		transactions: make(map[string]entity.Transaction),
		txOrder:      make([]string, 0, maxTransactions),
	}
	r.seedData()
	return r
}

func (r *BankStatementRepository) seedData() {
	userID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
	currentBalance := 5000.00
	stmtNo := 1

	for year := 2021; year <= 2024; year++ {
		// Calculate the yearly increase (matches the payslip logic)
		yearMultiplier := math.Pow(1.02, float64(year-2021))
		baseGross := 4500.00 * yearMultiplier
		baseNet := 2900.00 * yearMultiplier

		for month := 1; month <= 12; month++ {
			statementID := uuid.New()
			stmtDate := time.Date(year, time.Month(month), 28, 23, 59, 59, 0, time.UTC)

			stmt := entity.BankStatement{
				ID:            statementID,
				UserID:        userID,
				AccountHolder: "John Doe",
				IBAN:          "DE89370400440532013000",
				StatementDate: stmtDate,
				StatementNo:   stmtNo,
				OldBalance:    currentBalance,
				Currency:      "EUR",
				StatementType: entity.StatementTypeGiro,
				ImportedAt:    time.Now().Add(-time.Duration((2025-year)*8760) * time.Hour),
			}

			var txns []entity.Transaction

			// 1. Salary Payment
			payout := baseNet
			desc := "Salary Acme Corp"
			if month == 6 {
				payout += (baseGross * 0.5 * 0.55) // Holiday Bonus
				desc = "Salary Acme Corp + Holiday Bonus"
			} else if month == 11 {
				payout += (baseGross * 0.8 * 0.55) // Christmas Bonus
				desc = "Salary Acme Corp + Christmas Bonus"
			}

			txns = append(txns, r.createTx(statementID, userID, time.Date(year, time.Month(month), 26, 8, 0, 0, 0, time.UTC), desc, payout, entity.TransactionTypeCredit))
			currentBalance += payout

			// 2. Rent
			txns = append(txns, r.createTx(statementID, userID, time.Date(year, time.Month(month), 1, 9, 0, 0, 0, time.UTC), "Rent Payment", 1200.00, entity.TransactionTypeDebit))
			currentBalance -= 1200.00

			// 3. Groceries
			txns = append(txns, r.createTx(statementID, userID, time.Date(year, time.Month(month), 5, 14, 0, 0, 0, time.UTC), "REWE Supermarket", 150.00, entity.TransactionTypeDebit))
			txns = append(txns, r.createTx(statementID, userID, time.Date(year, time.Month(month), 15, 16, 0, 0, 0, time.UTC), "ALDI Nord", 120.00, entity.TransactionTypeDebit))
			txns = append(txns, r.createTx(statementID, userID, time.Date(year, time.Month(month), 22, 10, 0, 0, 0, time.UTC), "EDEKA Supermarket", 180.00, entity.TransactionTypeDebit))
			currentBalance -= 450.00

			// 4. Utilities & Internet
			txns = append(txns, r.createTx(statementID, userID, time.Date(year, time.Month(month), 3, 10, 0, 0, 0, time.UTC), "Stadtwerke Utilities", 150.00, entity.TransactionTypeDebit))
			txns = append(txns, r.createTx(statementID, userID, time.Date(year, time.Month(month), 4, 10, 0, 0, 0, time.UTC), "Telekom Internet", 45.00, entity.TransactionTypeDebit))
			currentBalance -= 195.00

			// 5. Entertainment / Dining Out
			txns = append(txns, r.createTx(statementID, userID, time.Date(year, time.Month(month), 12, 20, 0, 0, 0, time.UTC), "Restaurant Bella Italia", 85.00, entity.TransactionTypeDebit))
			txns = append(txns, r.createTx(statementID, userID, time.Date(year, time.Month(month), 20, 19, 0, 0, 0, time.UTC), "CinemaxX", 35.00, entity.TransactionTypeDebit))
			currentBalance -= 120.00

			// 6. Subscriptions
			txns = append(txns, r.createTx(statementID, userID, time.Date(year, time.Month(month), 10, 8, 0, 0, 0, time.UTC), "Netflix", 17.99, entity.TransactionTypeDebit))
			txns = append(txns, r.createTx(statementID, userID, time.Date(year, time.Month(month), 11, 8, 0, 0, 0, time.UTC), "Spotify", 10.99, entity.TransactionTypeDebit))
			currentBalance -= 28.98

			stmt.NewBalance = currentBalance
			stmt.Transactions = txns
			stmt.ContentHash = fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d-%d", year, month))))

			r.statements[statementID] = stmt
			r.stmtOrder = append(r.stmtOrder, statementID)
			for _, tx := range txns {
				r.transactions[tx.ContentHash] = tx
			}

			stmtNo++
		}
	}
}

func (r *BankStatementRepository) createTx(stmtID, userID uuid.UUID, date time.Time, desc string, amount float64, txType entity.TransactionType) entity.Transaction {
	id := uuid.New()
	hashStr := fmt.Sprintf("%s-%s-%f", date.Format(time.RFC3339), desc, amount)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(hashStr)))

	return entity.Transaction{
		ID:                id,
		UserID:            userID,
		BankStatementID:   &stmtID,
		BookingDate:       date,
		ValutaDate:        date,
		Description:       desc,
		CounterpartyName:  strings.Split(desc, " ")[0], // Simple heuristic for mock
		Amount:            amount,
		Currency:          "EUR",
		BaseAmount:        amount,
		BaseCurrency:      "EUR",
		Type:              txType,
		Reference:         "REF-" + id.String()[:8],
		ContentHash:       hash,
		IsReconciled:      true,
		Reviewed:          true,
		StatementType:     entity.StatementTypeGiro,
		IsPayslipVerified: txType == entity.TransactionTypeCredit && strings.Contains(desc, "Salary"),
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

	// Simple sort by booking date DESC to match Postgres behavior
	// (Note: this is a basic stable sort for memory repo)
	for i := 0; i < len(txns); i++ {
		for j := i + 1; j < len(txns); j++ {
			if txns[i].BookingDate.Before(txns[j].BookingDate) {
				txns[i], txns[j] = txns[j], txns[i]
			}
		}
	}

	if filter.Offset >= len(txns) {
		return []entity.Transaction{}, nil
	}

	end := len(txns)
	if filter.Limit > 0 && filter.Offset+filter.Limit < end {
		end = filter.Offset + filter.Limit
	}

	return txns[filter.Offset:end], nil
}

func (r *BankStatementRepository) SearchTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
	return r.FindTransactions(ctx, filter)
}

func (r *BankStatementRepository) GetCategorizationExamples(ctx context.Context, userID uuid.UUID, examplesCount int) ([]entity.CategorizationExample, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// map[categoryID][]entity.CategorizationExample
	catExamples := make(map[uuid.UUID][]entity.CategorizationExample)
	totalCount := 0

	for _, tx := range r.transactions {
		if tx.CategoryID == nil || tx.UserID != userID {
			continue
		}

		if totalCount >= examplesCount {
			break
		}

		if _, ok := catExamples[*tx.CategoryID]; !ok {
			catExamples[*tx.CategoryID] = make([]entity.CategorizationExample, 0)
		}

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
			totalCount++
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

func (r *BankStatementRepository) FindMatchingCategory(ctx context.Context, userID uuid.UUID, txn port.TransactionToCategorize) (*uuid.UUID, error) {
	return nil, nil // Not implemented for memory repository
}

func (r *BankStatementRepository) UpdateTransactionCategory(ctx context.Context, hash string, categoryID *uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tx, ok := r.transactions[hash]
	if !ok || tx.UserID != userID {
		return errors.New("transaction not found")
	}
	tx.CategoryID = categoryID
	tx.Reviewed = true
	r.transactions[hash] = tx
	return nil
}

func (r *BankStatementRepository) UpdateTransactionCategoriesBulk(ctx context.Context, hashes []string, categoryID *uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, hash := range hashes {
		tx, ok := r.transactions[hash]
		if ok && tx.UserID == userID {
			tx.CategoryID = categoryID
			tx.Reviewed = true
			r.transactions[hash] = tx
		}
	}
	return nil
}

func (r *BankStatementRepository) UpdateTransactionSubscription(ctx context.Context, hash string, subID *uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tx, ok := r.transactions[hash]
	if !ok || tx.UserID != userID {
		return errors.New("transaction not found")
	}
	tx.SubscriptionID = subID
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

func (r *BankStatementRepository) MarkTransactionsReviewedBulk(ctx context.Context, hashes []string, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, hash := range hashes {
		if tx, ok := r.transactions[hash]; ok && tx.UserID == userID {
			tx.Reviewed = true
			r.transactions[hash] = tx
		}
	}
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


func (r *BankStatementRepository) UpdateTransactionBaseAmount(ctx context.Context, hash string, baseAmount float64, baseCurrency string, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tx, ok := r.transactions[hash]
	if !ok || tx.UserID != userID {
		return nil
	}
	tx.BaseAmount = baseAmount
	tx.BaseCurrency = baseCurrency
	r.transactions[hash] = tx
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
	if filter.SubscriptionID != nil && (tx.SubscriptionID == nil || *tx.SubscriptionID != *filter.SubscriptionID) {
		return false
	}
	if filter.Search != "" {
		search := strings.ToLower(filter.Search)
		if !strings.Contains(strings.ToLower(tx.Description), search) &&
			!strings.Contains(strings.ToLower(tx.Reference), search) {
			return false
		}
	}
	return true
}

var _ port.BankStatementRepository = (*BankStatementRepository)(nil)

func (r *BankStatementRepository) UpdateStatementAccount(ctx context.Context, statementID uuid.UUID, bankAccountID *uuid.UUID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stmt, ok := r.statements[statementID]
	if !ok || stmt.UserID != userID {
		return entity.ErrBankStatementNotFound
	}
	stmt.BankAccountID = bankAccountID
	r.statements[statementID] = stmt
	
	// Cascade to transactions
	for id, t := range r.transactions {
		if t.BankStatementID == &statementID {
			t.BankAccountID = bankAccountID
			r.transactions[id] = t
		}
	}
	return nil
}

func (r *BankStatementRepository) GetTransactionsByAccountID(ctx context.Context, bankAccountID uuid.UUID, userID uuid.UUID) ([]entity.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var txns []entity.Transaction
	for _, t := range r.transactions {
		if t.BankAccountID != nil && *t.BankAccountID == bankAccountID && t.UserID == userID {
			txns = append(txns, t)
		}
	}
	return txns, nil
}
