package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"

	"github.com/google/uuid"
)

var ErrEmptyFilePath = errors.New("bank statement service: file path must not be empty")
var ErrUnsupportedFormat = errors.New("bank statement service: unsupported file format")
var ErrDuplicate = entity.ErrDuplicate

var normalizeTxnTextRe = regexp.MustCompile(`[^a-z0-9]+`)

type BankStatementService struct {
	parsers          map[string][]port.BankStatementParser
	repo             port.BankStatementRepository
	plannedTxService port.PlannedTransactionUseCase
	discoveryService port.DiscoveryUseCase
	aiParser         port.BankStatementAIParser
	currencyService  *CurrencyService
	Logger           *slog.Logger
}

func NewBankStatementService(repo port.BankStatementRepository, logger *slog.Logger) *BankStatementService {
	if logger == nil {
		logger = slog.Default()
	}
	return &BankStatementService{
		parsers: make(map[string][]port.BankStatementParser),
		repo:    repo,
		Logger:  logger,
	}
}

func (s *BankStatementService) WithCurrencyService(svc *CurrencyService) *BankStatementService {
	s.currencyService = svc
	return s
}

func (s *BankStatementService) WithPlannedTransactionService(svc port.PlannedTransactionUseCase) *BankStatementService {
	s.plannedTxService = svc
	return s
}

func (s *BankStatementService) WithDiscoveryService(svc port.DiscoveryUseCase) *BankStatementService {
	s.discoveryService = svc
	return s
}

// WithAIParser injects the abstract AI Parser port
func (s *BankStatementService) WithAIParser(parser port.BankStatementAIParser) *BankStatementService {
	s.aiParser = parser
	return s
}

func (s *BankStatementService) RegisterParser(ext string, parser port.BankStatementParser) {
	ext = strings.ToLower(ext)
	s.parsers[ext] = append(s.parsers[ext], parser)
}

func (s *BankStatementService) ImportFromFile(ctx context.Context, userID uuid.UUID, fileName string, fileBytes []byte, useAI bool, userStmtType entity.StatementType) (entity.BankStatement, error) {
	s.Logger.Info("Starting import of bank statement file", "file", fileName, "use_ai_parser", useAI, "user_id", userID)
	if len(fileBytes) == 0 {
		return entity.BankStatement{}, errors.New("bank statement service: file bytes must not be empty")
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	var stmt entity.BankStatement
	var parseErr error
	parsedSuccessfully := false

	mimeType := http.DetectContentType(fileBytes)
	if idx := strings.IndexByte(mimeType, ';'); idx >= 0 {
		mimeType = mimeType[:idx]
	}
	if mimeType == "application/octet-stream" && strings.HasSuffix(ext, ".pdf") {
		mimeType = "application/pdf"
	}

	if useAI {
		if s.aiParser != nil {
			s.Logger.Info("Attempting AI parser exclusively as requested", "file", fileName, "user_id", userID)
			stmt, parseErr = s.aiParser.ParseBankStatement(ctx, userID, fileName, mimeType, fileBytes)
			if parseErr == nil {
				if err := stmt.IsValid(); err == nil {
					parsedSuccessfully = true
				} else {
					parseErr = fmt.Errorf("AI parser validation failed: %w", err)
					s.Logger.Error("AI parser validation failed", "file", fileName, "user_id", userID, "error", parseErr)
				}
			} else {
				s.Logger.Error("AI parser failed", "file", fileName, "user_id", userID, "error", parseErr)
			}
		} else {
			parseErr = errors.New("bank statement service: AI parser requested but not configured")
		}
	} else {
		if parserList, ok := s.parsers[ext]; ok {
			var lastErr error
			for _, parser := range parserList {
				stmt, parseErr = parser.Parse(ctx, userID, fileBytes)
				if parseErr == nil {
					if err := stmt.IsValid(); err != nil {
						lastErr = fmt.Errorf("validation failed: %w", err)
						s.Logger.Warn("Parser succeeded but result is invalid", "parser", fmt.Sprintf("%T", parser), "error", lastErr)
						continue
					}
					parsedSuccessfully = true
					break
				}

				// If it's a format mismatch, we just move on.
				// If it's a "hard" error (like a corrupted file or unexpected structure),
				// we store it but CONTINUE the loop to see if another parser (like the AI fallback)
				// can handle it.
				lastErr = parseErr
			}

			if !parsedSuccessfully {
				if lastErr != nil {
					parseErr = lastErr
				} else {
					parseErr = errors.New("bank statement service: no suitable parser found for this document format")
				}
			}
		} else {
			parseErr = ErrUnsupportedFormat
		}
	}

	if !parsedSuccessfully {
		s.Logger.Warn("Bank statement import failed: parsing was not successful", "file", fileName, "error", parseErr, "user_id", userID)
		return entity.BankStatement{}, fmt.Errorf("bank statement service: parse %s: %w", fileName, parseErr)
	}

	stmt.UserID = userID
	if userStmtType != "" {
		stmt.StatementType = userStmtType
	} else if stmt.StatementType == "" {
		stmt.StatementType = entity.StatementTypeGiro
	}

	for i := range stmt.Transactions {
		stmt.Transactions[i].UserID = userID
		stmt.Transactions[i].StatementType = stmt.StatementType
	}

	stmt.OriginalFile = fileBytes

	if stmt.ContentHash == "" {
		stmtBase := fmt.Sprintf("%s|%s|%d|%.2f", stmt.IBAN, stmt.StatementDate.Format("2006-01-02"), stmt.StatementNo, stmt.NewBalance)
		stmtHash := sha256.Sum256([]byte(stmtBase))
		stmt.ContentHash = hex.EncodeToString(stmtHash[:])
	}

	txCounts := make(map[string]int)
	for i := range stmt.Transactions {
		if stmt.Transactions[i].ContentHash == "" {
			tx := &stmt.Transactions[i]
			baseStr := fmt.Sprintf("%s|%s|%.2f|%s|%s", stmt.IBAN, tx.BookingDate.Format("2006-01-02"), tx.Amount, tx.Description, tx.Reference)
			txCounts[baseStr]++

			uniqueStr := fmt.Sprintf("%s|%d", baseStr, txCounts[baseStr])
			hash := sha256.Sum256([]byte(uniqueStr))
			tx.ContentHash = hex.EncodeToString(hash[:])
		}
	}

	if s.repo != nil && len(stmt.Transactions) > 0 {
		var minDate, maxDate time.Time
		for i, tx := range stmt.Transactions {
			if i == 0 || tx.BookingDate.Before(minDate) {
				minDate = tx.BookingDate
			}
			if i == 0 || tx.BookingDate.After(maxDate) {
				maxDate = tx.BookingDate
			}
		}

		filter := entity.TransactionFilter{
			UserID:   userID,
			FromDate: &minDate,
			ToDate:   &maxDate,
		}

		existingTxns, err := s.repo.SearchTransactions(ctx, filter)
		if err == nil && len(existingTxns) > 0 {
			var uniqueTxns []entity.Transaction
			var skippedTxns []entity.Transaction
			var linkTxIDs []uuid.UUID

			availableTxns := make([]entity.Transaction, len(existingTxns))
			copy(availableTxns, existingTxns)

			if stmt.ID == uuid.Nil {
				stmt.ID = uuid.New()
			}

			for _, tx := range stmt.Transactions {
				dupIdx := s.findDuplicateIndex(tx, availableTxns)
				if dupIdx >= 0 {
					existingTx := availableTxns[dupIdx]
					previewLen := 30
					if len(tx.Description) < 30 {
						previewLen = len(tx.Description)
					}
					s.Logger.Info("Mapping duplicate transaction to statement",
						"date", tx.BookingDate.Format("2006-01-02"),
						"amount", tx.Amount,
						"desc_preview", tx.Description[:previewLen],
						"existing_id", existingTx.ID,
						"user_id", userID)

					skippedTxns = append(skippedTxns, tx)
					linkTxIDs = append(linkTxIDs, existingTx.ID)

					availableTxns = append(availableTxns[:dupIdx], availableTxns[dupIdx+1:]...)
				} else {
					uniqueTxns = append(uniqueTxns, tx)
				}
			}

			stmt.Transactions = uniqueTxns
			stmt.SkippedTransactions = skippedTxns

			if len(stmt.Transactions) == 0 && len(linkTxIDs) == 0 {
				s.Logger.Warn("All transactions in statement are duplicates and no new statement needed", "content_hash", stmt.ContentHash, "user_id", userID)
				return stmt, ErrDuplicate
			}

			if s.repo != nil {
				if err := s.repo.Save(ctx, stmt); err != nil {
					if errors.Is(err, ErrDuplicate) {
						s.Logger.Warn("Duplicate bank statement detected (full hash match)", "content_hash", stmt.ContentHash, "user_id", userID)
						return entity.BankStatement{}, fmt.Errorf("bank statement service: %w", ErrDuplicate)
					}
					s.Logger.Error("Failed to persist bank statement", "error", err, "user_id", userID)
					return entity.BankStatement{}, fmt.Errorf("bank statement service: persist: %w", err)
				}

				// Link transactions after statement is saved
				for _, txID := range linkTxIDs {
					if err := s.repo.LinkTransactionToStatement(ctx, txID, stmt.ID, userID); err != nil {
						s.Logger.Error("Failed to link transaction to statement", "tx_id", txID, "stmt_id", stmt.ID, "user_id", userID, "error", err)
					}
				}
			}
		} else {
			if err != nil {
				s.Logger.Warn("Could not fetch existing transactions for deduplication", "error", err, "user_id", userID)
			}
			if s.repo != nil {
				if err := s.repo.Save(ctx, stmt); err != nil {
					if errors.Is(err, ErrDuplicate) {
						s.Logger.Warn("Duplicate bank statement detected (full hash match)", "content_hash", stmt.ContentHash, "user_id", userID)
						return entity.BankStatement{}, fmt.Errorf("bank statement service: %w", ErrDuplicate)
					}
					s.Logger.Error("Failed to persist bank statement", "error", err, "user_id", userID)
					return entity.BankStatement{}, fmt.Errorf("bank statement service: persist: %w", err)
				}
			}
		}
	} else {
		if s.repo != nil {
			if err := s.repo.Save(ctx, stmt); err != nil {
				if errors.Is(err, ErrDuplicate) {
					s.Logger.Warn("Duplicate bank statement detected (full hash match)", "content_hash", stmt.ContentHash, "user_id", userID)
					return entity.BankStatement{}, fmt.Errorf("bank statement service: %w", ErrDuplicate)
				}
				s.Logger.Error("Failed to persist bank statement", "error", err, "user_id", userID)
				return entity.BankStatement{}, fmt.Errorf("bank statement service: persist: %w", err)
			}
		}
	}

	s.Logger.Info("Bank statement imported successfully", "id", stmt.ID, "content_hash", stmt.ContentHash, "user_id", userID, "imported_count", len(stmt.Transactions), "skipped_count", len(stmt.SkippedTransactions))

	if s.plannedTxService != nil && len(stmt.Transactions) > 0 {
		if err := s.plannedTxService.MatchTransactions(ctx, userID, stmt.Transactions); err != nil {
			s.Logger.Error("Failed to match planned transactions", "error", err, "user_id", userID)
		}
	}

	if s.discoveryService != nil && len(stmt.Transactions) > 0 {
		if err := s.discoveryService.MatchTransactions(ctx, userID, stmt.Transactions); err != nil {
			s.Logger.Error("Failed to match subscription transactions", "error", err, "user_id", userID)
		}
	}

	// Trigger asynchronous currency conversion
	if s.currencyService != nil {
		go func() {
			cCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			if err := s.currencyService.UpdateBaseAmountsForUser(cCtx, userID); err != nil {
				s.Logger.Error("Background currency conversion failed", "user_id", userID, "error", err)
			}
		}()
	}

	return stmt, nil
}

func (s *BankStatementService) findDuplicateIndex(newTx entity.Transaction, existingTxns []entity.Transaction) int {
	newRef := normalizeTransactionText(newTx.Reference)
	newDesc := normalizeTransactionText(newTx.Description)

	for i, ex := range existingTxns {
		y1, m1, d1 := newTx.BookingDate.Date()
		y2, m2, d2 := ex.BookingDate.Date()
		if y1 != y2 || m1 != m2 || d1 != d2 {
			continue
		}

		if math.Abs(newTx.Amount-ex.Amount) > 0.01 {
			continue
		}

		exRef := normalizeTransactionText(ex.Reference)
		exDesc := normalizeTransactionText(ex.Description)

		if transactionTextMatches(newRef, newDesc, exRef, exDesc) {
			return i
		}
	}
	return -1
}

func normalizeTransactionText(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return ""
	}
	return normalizeTxnTextRe.ReplaceAllString(v, "")
}

func transactionTextMatches(newRef, newDesc, exRef, exDesc string) bool {
	if newRef != "" && exRef != "" && (newRef == exRef || strings.Contains(newRef, exRef) || strings.Contains(exRef, newRef)) {
		return true
	}

	if newDesc != "" && exDesc != "" && (newDesc == exDesc || strings.Contains(newDesc, exDesc) || strings.Contains(exDesc, newDesc)) {
		return true
	}

	// Some parsers place the same value in reference vs description fields.
	if newRef != "" && (newRef == exDesc || strings.Contains(newRef, exDesc) || strings.Contains(exDesc, newRef)) {
		return true
	}
	if newDesc != "" && (newDesc == exRef || strings.Contains(newDesc, exRef) || strings.Contains(exRef, newDesc)) {
		return true
	}

	return false
}

func (s *BankStatementService) DeleteStatement(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if s.repo == nil {
		return errors.New("bank statement service: repository not configured")
	}
	s.Logger.Info("Deleting bank statement", "id", id, "user_id", userID)
	return s.repo.Delete(ctx, id, userID)
}

func (s *BankStatementService) ImportFromDirectory(ctx context.Context, userID uuid.UUID, dirPath string) (int, []error) {
	if dirPath == "" {
		return 0, []error{errors.New("directory path is empty")}
	}

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return 0, []error{fmt.Errorf("failed to read directory: %w", err)}
	}

	imported := 0
	var errs []error

	for _, f := range files {
		select {
		case <-ctx.Done():
			return imported, append(errs, ctx.Err())
		default:
		}

		if f.IsDir() {
			continue
		}

		filePath := filepath.Join(dirPath, f.Name())
		fileBytes, err := os.ReadFile(filePath)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read file %s: %w", f.Name(), err))
			continue
		}

		// We don't know the statement type for auto-imports, so we use empty/giro
		_, err = s.ImportFromFile(ctx, userID, f.Name(), fileBytes, false, "")
		if err != nil {
			if !errors.Is(err, ErrDuplicate) && !errors.Is(err, ErrUnsupportedFormat) {
				errs = append(errs, fmt.Errorf("file %s: %w", f.Name(), err))
			}
			continue
		}
		imported++
	}

	return imported, errs
}
