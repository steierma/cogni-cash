package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"log/slog"

	"github.com/google/uuid"
)

var ErrEmptyFilePath = errors.New("bank statement service: file path must not be empty")
var ErrUnsupportedFormat = errors.New("bank statement service: unsupported file format")
var ErrDuplicate = entity.ErrDuplicate

var normalizeTxnTextRe = regexp.MustCompile(`[^a-z0-9]+`)

type BankStatementService struct {
	parsers        map[string][]port.BankStatementParser
	fallbackParser port.BankStatementParser
	repo           port.BankStatementRepository
	Logger         *slog.Logger
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

func (s *BankStatementService) WithFallbackParser(parser port.BankStatementParser) *BankStatementService {
	s.fallbackParser = parser
	return s
}

func (s *BankStatementService) RegisterParser(ext string, parser port.BankStatementParser) {
	ext = strings.ToLower(ext)
	s.parsers[ext] = append(s.parsers[ext], parser)
}

func (s *BankStatementService) ImportFromDirectory(ctx context.Context, dirPath string) (int, []error) {
	if dirPath == "" {
		return 0, []error{ErrEmptyFilePath}
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return 0, []error{fmt.Errorf("bank statement service: read dir %s: %w", dirPath, err)}
	}

	var importedCount int
	var errs []error

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if _, ok := s.parsers[ext]; !ok {
			continue
		}

		fullPath := filepath.Join(dirPath, entry.Name())
		_, err := s.ImportFromFile(ctx, fullPath, false, "")
		if err != nil {
			if errors.Is(err, ErrDuplicate) {
				s.Logger.Debug("Skipped duplicate file in directory", "file", entry.Name())
				continue
			}
			errs = append(errs, fmt.Errorf("file %s: %w", entry.Name(), err))
			continue
		}
		importedCount++
	}

	return importedCount, errs
}

func (s *BankStatementService) ImportFromFile(ctx context.Context, filePath string, useAI bool, userStmtType entity.StatementType) (entity.BankStatement, error) {
	s.Logger.Info("Starting import of bank statement file", "file", filePath, "use_ai_parser", useAI, "fallback_parser", s.fallbackParser)
	if filePath == "" {
		return entity.BankStatement{}, ErrEmptyFilePath
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	var stmt entity.BankStatement
	var parseErr error
	parsedSuccessfully := false

	if useAI {
		if s.fallbackParser != nil {
			s.Logger.Info("Attempting AI parser exclusively as requested", "file", filePath)
			stmt, parseErr = s.fallbackParser.Parse(ctx, filePath)
			if parseErr == nil {
				if err := stmt.IsValid(); err == nil {
					parsedSuccessfully = true
				} else {
					parseErr = fmt.Errorf("AI parser validation failed: %w", err)
					s.Logger.Error("AI parser validation failed", "file", filePath, "error", parseErr)
				}
			} else {
				s.Logger.Error("AI parser failed", "file", filePath, "error", parseErr)
			}
		} else {
			parseErr = errors.New("bank statement service: AI parser requested but not configured")
		}
	} else {
		if parserList, ok := s.parsers[ext]; ok {
			var lastErr error
			for _, parser := range parserList {
				stmt, parseErr = parser.Parse(ctx, filePath)
				if parseErr == nil {
					if err := stmt.IsValid(); err != nil {
						lastErr = fmt.Errorf("validation failed: %w", err)
						continue
					}
					parsedSuccessfully = true
					break
				}

				if strings.Contains(strings.ToLower(parseErr.Error()), "format mismatch") || strings.Contains(strings.ToLower(parseErr.Error()), "does not match format") {
					continue
				}

				lastErr = parseErr
				break
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
		return entity.BankStatement{}, fmt.Errorf("bank statement service: parse %s: %w", filePath, parseErr)
	}

	if userStmtType != "" {
		stmt.StatementType = userStmtType
		for i := range stmt.Transactions {
			stmt.Transactions[i].StatementType = userStmtType
		}
	} else if stmt.StatementType == "" {
		stmt.StatementType = entity.StatementTypeGiro
		for i := range stmt.Transactions {
			stmt.Transactions[i].StatementType = entity.StatementTypeGiro
		}
	}

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return entity.BankStatement{}, fmt.Errorf("bank statement service: read file %s: %w", filePath, err)
	}
	stmt.OriginalFile = fileBytes

	if stmt.SourceFile == "" {
		stmt.SourceFile = filepath.Base(filePath)
	}

	if stmt.ContentHash == "" {
		stmtBase := fmt.Sprintf("%s|%s|%d|%.2f", stmt.IBAN, stmt.StatementDate.Format("2006-01-02"), stmt.StatementNo, stmt.NewBalance)
		stmtHash := sha256.Sum256([]byte(stmtBase))
		stmt.ContentHash = hex.EncodeToString(stmtHash[:])
	}

	txCounts := make(map[string]int)
	for i := range stmt.Transactions {
		if stmt.Transactions[i].ContentHash == "" {
			tx := &stmt.Transactions[i]
			baseStr := fmt.Sprintf("%s|%.2f|%s|%s", tx.BookingDate.Format("2006-01-02"), tx.Amount, tx.Description, tx.Reference)
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
			FromDate: &minDate,
			ToDate:   &maxDate,
		}

		existingTxns, err := s.repo.SearchTransactions(ctx, filter)
		if err == nil && len(existingTxns) > 0 {
			var uniqueTxns []entity.Transaction
			var skippedTxns []entity.Transaction

			availableTxns := make([]entity.Transaction, len(existingTxns))
			copy(availableTxns, existingTxns)

			for _, tx := range stmt.Transactions {
				dupIdx := s.findDuplicateIndex(tx, availableTxns)
				if dupIdx >= 0 {
					previewLen := 30
					if len(tx.Description) < 30 {
						previewLen = len(tx.Description)
					}
					s.Logger.Info("Skipping duplicate transaction", "date", tx.BookingDate.Format("2006-01-02"), "amount", tx.Amount, "desc_preview", tx.Description[:previewLen])
					skippedTxns = append(skippedTxns, tx)

					availableTxns = append(availableTxns[:dupIdx], availableTxns[dupIdx+1:]...)
				} else {
					uniqueTxns = append(uniqueTxns, tx)
				}
			}

			stmt.Transactions = uniqueTxns
			stmt.SkippedTransactions = skippedTxns

			if len(stmt.Transactions) == 0 {
				s.Logger.Warn("All transactions in statement are duplicates", "content_hash", stmt.ContentHash)
				return stmt, ErrDuplicate
			}
		} else if err != nil {
			s.Logger.Warn("Could not fetch existing transactions for deduplication", "error", err)
		}
	}

	if s.repo != nil {
		if err := s.repo.Save(ctx, stmt); err != nil {
			if errors.Is(err, ErrDuplicate) {
				s.Logger.Warn("Duplicate bank statement detected (full hash match)", "content_hash", stmt.ContentHash)
				return entity.BankStatement{}, fmt.Errorf("bank statement service: %w", ErrDuplicate)
			}
			s.Logger.Error("Failed to persist bank statement", "error", err)
			return entity.BankStatement{}, fmt.Errorf("bank statement service: persist: %w", err)
		}
	}

	s.Logger.Info("Bank statement imported successfully", "id", stmt.ID, "content_hash", stmt.ContentHash, "imported_count", len(stmt.Transactions), "skipped_count", len(stmt.SkippedTransactions))
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
	if newRef != "" && exRef != "" && newRef == exRef {
		return true
	}

	if newDesc != "" && exDesc != "" && newDesc == exDesc {
		return true
	}

	// Some parsers place the same value in reference vs description fields.
	if newRef != "" && newRef == exDesc {
		return true
	}
	if newDesc != "" && newDesc == exRef {
		return true
	}

	return false
}

func (s *BankStatementService) DeleteStatement(ctx context.Context, id uuid.UUID) error {
	if s.repo == nil {
		return errors.New("bank statement service: repository not configured")
	}
	return s.repo.Delete(ctx, id)
}
