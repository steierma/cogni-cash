package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"cogni-cash/internal/domain/entity"
	"cogni-cash/internal/domain/port"
)

type SharingService struct {
	categoryRepo port.CategoryRepository
	invoiceRepo  port.InvoiceRepository
	sharingRepo  port.SharingRepository
	txnRepo      port.BankStatementRepository
	userRepo     port.UserRepository
	logger       *slog.Logger
}

func NewSharingService(
	categoryRepo port.CategoryRepository,
	invoiceRepo port.InvoiceRepository,
	sharingRepo port.SharingRepository,
	txnRepo port.BankStatementRepository,
	userRepo port.UserRepository,
	logger *slog.Logger,
) *SharingService {
	if logger == nil {
		logger = slog.Default()
	}
	return &SharingService{
		categoryRepo: categoryRepo,
		invoiceRepo:  invoiceRepo,
		sharingRepo:  sharingRepo,
		txnRepo:      txnRepo,
		userRepo:     userRepo,
		logger:       logger,
	}
}

func (s *SharingService) GetDashboard(ctx context.Context, userID uuid.UUID) (entity.SharingDashboard, error) {
	s.logger.Info("Generating sharing dashboard", "user_id", userID)

	// 1. Fetch categories owned or shared with user
	cats, err := s.categoryRepo.FindAll(ctx, userID)
	if err != nil {
		return entity.SharingDashboard{}, fmt.Errorf("dashboard: fetch categories: %w", err)
	}

	sharedCats := make([]entity.SharedCategorySummary, 0)
	for _, c := range cats {
		if c.IsShared || c.UserID != userID {
			perm := "view"
			if c.UserID == userID {
				perm = "owner"
			}
			// In a real implementation, we'd fetch the actual permission level from shared_categories table
			sharedCats = append(sharedCats, entity.SharedCategorySummary{
				Category:    c,
				Permissions: perm,
			})
		}
	}

	// 2. Fetch invoices owned or shared with user
	invoices, err := s.invoiceRepo.FindAll(ctx, entity.InvoiceFilter{
		UserID:        userID,
		IncludeShared: true,
	})
	if err != nil {
		return entity.SharingDashboard{}, fmt.Errorf("dashboard: fetch invoices: %w", err)
	}

	sharedInvoices := make([]entity.Invoice, 0)
	for _, inv := range invoices {
		isSharedCat := false
		if inv.CategoryID != nil {
			for _, sc := range sharedCats {
				if sc.Category.ID == *inv.CategoryID {
					isSharedCat = true
					break
				}
			}
		}

		// An invoice is shared if it belongs to someone else (shared with me)
		// OR if it belongs to me but I have explicitly shared it (len(SharedWith) > 0)
		// OR if it belongs to a shared category
		if inv.UserID != userID || len(inv.SharedWith) > 0 || isSharedCat {
			sharedInvoices = append(sharedInvoices, inv)
		}
	}

	// 3. Calculate Balances for shared categories
	balances := make([]entity.CategoryBalance, 0)
	for _, c := range cats {
		if !c.IsShared && c.UserID == userID {
			continue // skip private categories
		}

		// Find transactions for this category (from anyone)
		filter := entity.TransactionFilter{
			UserID:        userID,
			CategoryID:    &c.ID,
			IncludeShared: true,
		}
		txns, err := s.txnRepo.FindTransactions(ctx, filter)
		if err != nil {
			s.logger.Warn("Failed to fetch transactions for sharing balance", "category_id", c.ID, "error", err)
			continue
		}

		if len(txns) == 0 {
			continue
		}

		balance := entity.CategoryBalance{
			CategoryID:    c.ID,
			CategoryName:  c.Name,
			UserBreakdown: make([]entity.UserSpending, 0),
		}

		userSpendingMap := make(map[uuid.UUID]float64)
		for _, tx := range txns {
			userSpendingMap[tx.UserID] += tx.Amount
			balance.TotalSpent += tx.Amount
		}

		for uid, amt := range userSpendingMap {
			username := uid.String()
			if user, err := s.userRepo.FindByID(ctx, uid); err == nil {
				username = user.Username
			}
			balance.UserBreakdown = append(balance.UserBreakdown, entity.UserSpending{
				UserID:   uid,
				Username: username,
				Amount:   amt,
			})
		}
		balances = append(balances, balance)
	}

	return entity.SharingDashboard{
		SharedCategories: sharedCats,
		SharedInvoices:   sharedInvoices,
		Balances:         balances,
	}, nil
}
