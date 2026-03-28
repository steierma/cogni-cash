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

func (s *TransactionService) UpdateCategory(ctx context.Context, hash string, categoryID *uuid.UUID) error {
	return s.repo.UpdateTransactionCategory(ctx, hash, categoryID)
}

func (s *TransactionService) MarkAsReviewed(ctx context.Context, hash string) error {
	return s.repo.MarkTransactionReviewed(ctx, hash)
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
		cats, _ = s.categoryRepo.FindAll(ctx)
	}
	colorMap := make(map[string]string)
	for _, c := range cats {
		colorMap[c.Name] = c.Color
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

	catExpenseMap := make(map[string]float64)
	catIncomeMap := make(map[string]float64)
	merchantExpenseMap := make(map[string]float64)
	timeSeriesMap := make(map[string]*entity.TimeSeriesPoint)

	for _, tx := range txns {
		dateStr := tx.BookingDate.Format(timeFormat)
		if _, ok := timeSeriesMap[dateStr]; !ok {
			timeSeriesMap[dateStr] = &entity.TimeSeriesPoint{Date: dateStr}
		}

		catName := "Uncategorized"
		if tx.CategoryID != nil {
			for _, c := range cats {
				if c.ID == *tx.CategoryID {
					catName = c.Name
					break
				}
			}
		}

		if tx.Amount >= 0 {
			result.TotalIncome += tx.Amount
			timeSeriesMap[dateStr].Income += tx.Amount
			catIncomeMap[catName] += tx.Amount
		} else {
			absAmount := math.Abs(tx.Amount)
			result.TotalExpense += absAmount
			timeSeriesMap[dateStr].Expense += absAmount
			catExpenseMap[catName] += absAmount

			merchant := strings.TrimSpace(tx.Description)
			if merchant == "" {
				merchant = strings.TrimSpace(tx.Reference)
			}
			if merchant == "" {
				merchant = "Unknown"
			} else if len(merchant) > 40 {
				merchant = merchant[:37] + "..."
			}
			merchantExpenseMap[merchant] += absAmount
		}
	}

	result.NetSavings = result.TotalIncome - result.TotalExpense

	for cat, amount := range catExpenseMap {
		color := colorMap[cat]
		if color == "" {
			color = "#9ca3af"
		}
		result.CategoryTotals = append(result.CategoryTotals, entity.CategoryTotal{
			Category: cat,
			Amount:   amount,
			Type:     "expense",
			Color:    color,
		})
	}

	for cat, amount := range catIncomeMap {
		color := colorMap[cat]
		if color == "" {
			color = "#9ca3af"
		}
		result.CategoryTotals = append(result.CategoryTotals, entity.CategoryTotal{
			Category: cat,
			Amount:   amount,
			Type:     "income",
			Color:    color,
		})
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

	for merchant, amount := range merchantExpenseMap {
		result.TopMerchants = append(result.TopMerchants, entity.MerchantTotal{
			Merchant: merchant,
			Amount:   amount,
		})
	}

	sort.Slice(result.TopMerchants, func(i, j int) bool {
		return result.TopMerchants[i].Amount > result.TopMerchants[j].Amount
	})
	if len(result.TopMerchants) > 5 {
		result.TopMerchants = result.TopMerchants[:5]
	}

	return result, nil
}

func (s *TransactionService) StartAutoCategorizeAsync(ctx context.Context, batchSize int) error {
	if s.llm == nil {
		return errors.New("transaction service: LLM categorizer not configured")
	}

	allTxns, err := s.repo.SearchTransactions(ctx, entity.TransactionFilter{})
	if err != nil {
		return fmt.Errorf("fetch pending transactions: %w", err)
	}

	var toCategorize []port.TransactionToCategorize
	for _, tx := range allTxns {
		if tx.CategoryID == nil {
			toCategorize = append(toCategorize, port.TransactionToCategorize{
				Hash:        tx.ContentHash,
				Description: tx.Description,
				Reference:   tx.Reference,
			})
		}
	}

	if len(toCategorize) == 0 {
		return ErrNothingToCategorize
	}

	cats, err := s.categoryRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("fetch categories: %w", err)
	}

	// Fetch historical examples for few-shot learning
	examplesCount := 20
	if s.settingsRepo != nil {
		if val, err := s.settingsRepo.Get(ctx, "auto_categorization_examples_per_category"); err == nil && val != "" {
			fmt.Sscanf(val, "%d", &examplesCount)
		}
	}

	examples, err := s.repo.GetCategorizationExamples(ctx, examplesCount)
	if err != nil {
		s.Logger.Warn("Failed to fetch categorization examples, proceeding without them", "error", err)
	}

	jobCtx, cancelFunc := context.WithCancel(context.Background())

	if err := s.JobTracker.Start(len(toCategorize), cancelFunc); err != nil {
		cancelFunc()
		return err
	}

	go s.runCategorizeLoop(jobCtx, toCategorize, cats, examples, batchSize)

	return nil
}

func (s *TransactionService) runCategorizeLoop(ctx context.Context, txns []port.TransactionToCategorize, categories []entity.Category, examples []entity.CategorizationExample, batchSize int) {
	s.Logger.Info("Starting auto-categorization job", "total_transactions", len(txns), "examples", len(examples))
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

		results, err := s.llm.CategorizeBatch(ctx, batch, catNames, examples)
		if err != nil {
			if ctx.Err() != nil {
				s.JobTracker.Finish("cancelled")
			} else {
				s.Logger.Error("Failed to categorize batch", "error", err, "batch_start", i)
				continue
			}
			return
		}

		var successfulResults []port.CategorizedTransaction
		for _, res := range results {
			var validCategoryID *uuid.UUID
			for _, knownCat := range categories {
				if res.Category == knownCat.Name {
					validCategoryID = &knownCat.ID
					break
				}
			}

			if validCategoryID != nil {
				if err := s.repo.UpdateTransactionCategory(ctx, res.Hash, validCategoryID); err != nil {
					s.Logger.Error("Failed to update transaction category", "hash", res.Hash, "error", err)
				} else {
					successfulResults = append(successfulResults, res)
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
