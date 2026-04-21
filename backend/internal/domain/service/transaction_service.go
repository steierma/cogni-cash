package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
	"log/slog"

	"github.com/google/uuid"
)

var _ port.TransactionUseCase = (*TransactionService)(nil)

type TransactionService struct {
	repo         port.BankStatementRepository
	categoryRepo port.CategoryRepository
	settingsRepo port.SettingsRepository
	llm          port.TransactionCategorizer
	Logger       *slog.Logger
	JobTracker   port.JobTracker
}

func NewTransactionService(repo port.BankStatementRepository, catRepo port.CategoryRepository, settingsRepo port.SettingsRepository, llm port.TransactionCategorizer, logger *slog.Logger) *TransactionService {
	if logger == nil {
		logger = slog.Default()
	}
	return &TransactionService{
		repo:         repo,
		categoryRepo: catRepo,
		settingsRepo: settingsRepo,
		llm:          llm,
		Logger:       logger,
		JobTracker:   NewJobManager(),
	}
}

func (s *TransactionService) ListTransactions(ctx context.Context, filter entity.TransactionFilter) ([]entity.Transaction, error) {
	return s.repo.FindTransactions(ctx, filter)
}

func (s *TransactionService) UpdateCategory(ctx context.Context, hash string, categoryID *uuid.UUID, userID uuid.UUID) error {
	return s.repo.UpdateTransactionCategory(ctx, hash, categoryID, userID)
}

func (s *TransactionService) MarkAsReviewed(ctx context.Context, hash string, userID uuid.UUID) error {
	return s.repo.MarkTransactionReviewed(ctx, hash, userID)
}

func (s *TransactionService) ToggleSkipForecasting(ctx context.Context, hash string, skip bool, userID uuid.UUID) error {
	return s.repo.UpdateTransactionSkipForecasting(ctx, hash, skip, userID)
}

func (s *TransactionService) GetTransactionAnalytics(ctx context.Context, filter entity.TransactionFilter) (entity.TransactionAnalytics, error) {
	if s.repo == nil {
		return entity.TransactionAnalytics{}, errors.New("repository not configured")
	}

	txns, err := s.repo.FindTransactions(ctx, filter)
	if err != nil {
		return entity.TransactionAnalytics{}, fmt.Errorf("failed to fetch transactions: %w", err)
	}

	var cats []entity.Category
	if s.categoryRepo != nil {
		cats, _ = s.categoryRepo.FindAll(ctx, filter.UserID)
	}
	colorMap := make(map[string]string)
	nameMap := make(map[string]string)
	for _, c := range cats {
		idStr := c.ID.String()
		colorMap[idStr] = c.Color
		nameMap[idStr] = c.Name
	}

	result := entity.TransactionAnalytics{
		CategoryTotals: make([]entity.CategoryTotal, 0),
		TimeSeries:     make([]entity.TimeSeriesPoint, 0),
		TopMerchants:   make([]entity.MerchantTotal, 0),
	}

	if len(txns) == 0 {
		return result, nil
	}

	timeFormat := "2006-01"

	catNetMap := make(map[string]float64)
	merchantNetMap := make(map[string]float64)
	timeSeriesMap := make(map[string]*entity.TimeSeriesPoint)

	for _, tx := range txns {
		// Net-Worth Isolation: Shared transactions from other users don't count towards my net worth
		// unless we explicitly want to see collaborative analytics.
		if !filter.IncludeShared && tx.UserID != filter.UserID {
			continue
		}

		dateStr := tx.BookingDate.Format(timeFormat)
		if _, ok := timeSeriesMap[dateStr]; !ok {
			timeSeriesMap[dateStr] = &entity.TimeSeriesPoint{
				Date:            dateStr,
				CategoryAmounts: make(map[string]float64),
			}
		}

		catID := "uncategorized"
		if tx.CategoryID != nil {
			idStr := tx.CategoryID.String()
			// If the category is not in our active list (e.g. soft-deleted), treat it as uncategorized
			if _, ok := nameMap[idStr]; ok {
				catID = idStr
			}
		}

		if tx.BaseAmount >= 0 {
			result.TotalIncome += tx.BaseAmount
			timeSeriesMap[dateStr].Income += tx.BaseAmount
		} else {
			absAmount := math.Abs(tx.BaseAmount)
			result.TotalExpense += absAmount
			timeSeriesMap[dateStr].Expense += absAmount
		}

		// Always accumulate net amount for category and merchant
		catNetMap[catID] += tx.BaseAmount
		timeSeriesMap[dateStr].CategoryAmounts[catID] += tx.BaseAmount

		merchant := strings.TrimSpace(tx.Description)
		if merchant == "" {
			merchant = strings.TrimSpace(tx.Reference)
		}
		if merchant == "" {
			merchant = "Unknown"
		} else if len(merchant) > 40 {
			merchant = merchant[:37] + "..."
		}
		merchantNetMap[merchant] += tx.BaseAmount
	}

	result.NetSavings = result.TotalIncome - result.TotalExpense

	// Convert net maps to DTOs
	for id, netAmount := range catNetMap {
		color := colorMap[id]
		name := nameMap[id]
		if id == "uncategorized" {
			color = "#9ca3af"
			name = "Uncategorized"
		}

		// We classify as expense if net is negative, or income if net is positive
		if netAmount < 0 {
			result.CategoryTotals = append(result.CategoryTotals, entity.CategoryTotal{
				CategoryID: id,
				Category:   name,
				Amount:     math.Abs(netAmount),
				Type:       "expense",
				Color:      color,
			})
		} else if netAmount > 0 {
			result.CategoryTotals = append(result.CategoryTotals, entity.CategoryTotal{
				CategoryID: id,
				Category:   name,
				Amount:     netAmount,
				Type:       "income",
				Color:      color,
			})
		}
	}

	sort.Slice(result.CategoryTotals, func(i, j int) bool {
		return result.CategoryTotals[i].Amount > result.CategoryTotals[j].Amount
	})

	for _, ts := range timeSeriesMap {
		result.TimeSeries = append(result.TimeSeries, *ts)
	}

	sort.Slice(result.TimeSeries, func(i, j int) bool {
		return result.TimeSeries[i].Date < result.TimeSeries[j].Date
	})

	for merchant, netAmount := range merchantNetMap {
		// Only include if it's a net expense for top merchants list
		if netAmount < 0 {
			result.TopMerchants = append(result.TopMerchants, entity.MerchantTotal{
				Merchant: merchant,
				Amount:   math.Abs(netAmount),
			})
		}
	}

	sort.Slice(result.TopMerchants, func(i, j int) bool {
		return result.TopMerchants[i].Amount > result.TopMerchants[j].Amount
	})
	if len(result.TopMerchants) > 5 {
		result.TopMerchants = result.TopMerchants[:5]
	}

	return result, nil
}

func (s *TransactionService) StartAutoCategorizeAsync(ctx context.Context, userID uuid.UUID, batchSize int) error {
	if s.llm == nil {
		return errors.New("transaction service: LLM categorizer not configured")
	}

	allTxns, err := s.repo.SearchTransactions(ctx, entity.TransactionFilter{UserID: userID})
	if err != nil {
		return fmt.Errorf("fetch pending transactions: %w", err)
	}

	var toCategorize []port.TransactionToCategorize
	for _, tx := range allTxns {
		if tx.CategoryID == nil {
			toCategorize = append(toCategorize, port.TransactionToCategorize{
				Hash:                tx.ContentHash,
				Description:         tx.Description,
				Reference:           tx.Reference,
				CounterpartyName:    tx.CounterpartyName,
				CounterpartyIban:    tx.CounterpartyIban,
				BankTransactionCode: tx.BankTransactionCode,
				MandateReference:    tx.MandateReference,
			})
		}
	}

	if len(toCategorize) == 0 {
		return ErrNothingToCategorize
	}

	cats, err := s.categoryRepo.FindAll(ctx, userID)
	if err != nil {
		return fmt.Errorf("fetch categories: %w", err)
	}

	// Fetch historical examples for few-shot learning
	examplesCount := 20
	if s.settingsRepo != nil {
		if val, err := s.settingsRepo.Get(ctx, "auto_categorization_examples_per_category", userID); err == nil && val != "" {
			fmt.Sscanf(val, "%d", &examplesCount)
		}
	}

	examples, err := s.repo.GetCategorizationExamples(ctx, userID, examplesCount)
	if err != nil {
		s.Logger.Warn("Failed to fetch categorization examples, proceeding without them", "error", err)
	}

	jobCtx, cancelFunc := context.WithCancel(context.Background())

	if err := s.JobTracker.Start(len(toCategorize), cancelFunc); err != nil {
		cancelFunc()
		return err
	}

	go s.runCategorizeLoop(jobCtx, userID, toCategorize, cats, examples, batchSize)

	return nil
}

func (s *TransactionService) runCategorizeLoop(ctx context.Context, userID uuid.UUID, txns []port.TransactionToCategorize, categories []entity.Category, examples []entity.CategorizationExample, batchSize int) {
	s.Logger.Info("Starting auto-categorization job", "total_transactions", len(txns), "examples", len(examples), "user_id", userID)
	defer func() {
		if s.JobTracker.GetState().Status == "running" {
			s.JobTracker.Finish("completed")
		}
	}()

	catNames := make([]string, len(categories))
	for i, c := range categories {
		catNames[i] = c.Name
	}

	if batchSize <= 0 {
		batchSize = 10
	}

	for i := 0; i < len(txns); i += batchSize {
		select {
		case <-ctx.Done():
			s.JobTracker.Finish("cancelled")
			return
		default:
		}

		end := i + batchSize
		if end > len(txns) {
			end = len(txns)
		}
		batch := txns[i:end]

		var toCategorizeViaLLM []port.TransactionToCategorize
		var successfulResults []port.CategorizedTransaction

		// 1. Try to find matches in database first
		for _, tx := range batch {
			if matchedID, err := s.repo.FindMatchingCategory(ctx, userID, tx); err == nil && matchedID != nil {
				// Found a high-confidence match in DB
				if err := s.repo.UpdateTransactionCategory(ctx, tx.Hash, matchedID, userID); err == nil {
					// We need to find the category name for the job tracker
					catName := "Categorized"
					for _, c := range categories {
						if c.ID == *matchedID {
							catName = c.Name
							break
						}
					}
					successfulResults = append(successfulResults, port.CategorizedTransaction{
						Hash:     tx.Hash,
						Category: catName,
					})
					continue
				}
			}
			toCategorizeViaLLM = append(toCategorizeViaLLM, tx)
		}

		// 2. Only call LLM for transactions that weren't matched in DB
		if len(toCategorizeViaLLM) > 0 {
			results, err := s.llm.CategorizeTransactionsBatch(ctx, userID, toCategorizeViaLLM, catNames, examples)
			if err != nil {
				if ctx.Err() != nil {
					s.JobTracker.Finish("cancelled")
				} else {
					s.Logger.Error("Failed to categorize batch via LLM", "error", err, "batch_start", i)
					// We still report the DB-matched results
					s.JobTracker.AddResults(len(batch), successfulResults)
					continue
				}
				return
			}

			for _, res := range results {
				var validCategoryID *uuid.UUID
				for _, knownCat := range categories {
					if res.Category == knownCat.Name {
						validCategoryID = &knownCat.ID
						break
					}
				}

				if validCategoryID != nil {
					if err := s.repo.UpdateTransactionCategory(ctx, res.Hash, validCategoryID, userID); err != nil {
						s.Logger.Error("Failed to update transaction category", "hash", res.Hash, "error", err)
					} else {
						successfulResults = append(successfulResults, res)
					}
				}
			}
		}

		s.JobTracker.AddResults(len(batch), successfulResults)
	}
}

func (s *TransactionService) GetJobStatus() port.JobState {
	return s.JobTracker.GetState()
}

func (s *TransactionService) CancelJob() {
	s.JobTracker.Cancel()
}
