package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"cogni-cash/internal/domain/port"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "userID"

type Handler struct {
	authSvc            port.AuthUseCase
	invoiceSvc         port.InvoiceUseCase
	bankStatementSvc   port.BankStatementUseCase
	transactionSvc     port.TransactionUseCase
	reconciliationSvc  port.ReconciliationUseCase
	settingsSvc        port.SettingsUseCase
	payslipSvc         port.PayslipUseCase
	bankSvc            port.BankUseCase
	userSvc            port.UserUseCase
	notificationSvc    port.NotificationUseCase
	payslipRepo        port.PayslipRepository
	bankStmtRepo       port.BankStatementRepository
	categoryRepo       port.CategoryRepository
	reconciliationRepo port.ReconciliationRepository
	Logger             *slog.Logger
	storageMode        string
	dbHost             string
	dbPinger           func(context.Context) error
}

func (h *Handler) getUserID(ctx context.Context) uuid.UUID {
	val := ctx.Value(userIDKey)
	if val == nil {
		return uuid.Nil
	}

	if id, ok := val.(uuid.UUID); ok {
		return id
	}

	if idStr, ok := val.(string); ok {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return uuid.Nil
		}
		return id
	}

	return uuid.Nil
}

func NewHandler(
	authSvc port.AuthUseCase,
	invoiceSvc port.InvoiceUseCase,
	bankStatementSvc port.BankStatementUseCase,
	settingsSvc port.SettingsUseCase,
	bankSvc port.BankUseCase,
	logger *slog.Logger,
	storageMode string,
	dbHost string,
	dbPinger func(context.Context) error,
) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		authSvc:          authSvc,
		invoiceSvc:       invoiceSvc,
		bankStatementSvc: bankStatementSvc,
		settingsSvc:      settingsSvc,
		bankSvc:          bankSvc,
		Logger:           logger,
		storageMode:      storageMode,
		dbHost:           dbHost,
		dbPinger:         dbPinger,
	}
}

func (h *Handler) WithBankService(svc port.BankUseCase) *Handler {
	h.bankSvc = svc
	return h
}

func (h *Handler) WithUserService(svc port.UserUseCase) *Handler {
	h.userSvc = svc
	return h
}

func (h *Handler) WithTransactionService(svc port.TransactionUseCase) *Handler {
	h.transactionSvc = svc
	return h
}

func (h *Handler) WithReconciliationService(svc port.ReconciliationUseCase) *Handler {
	h.reconciliationSvc = svc
	return h
}

func (h *Handler) WithBankStatementRepository(repo port.BankStatementRepository) *Handler {
	h.bankStmtRepo = repo
	return h
}

func (h *Handler) WithCategoryRepository(repo port.CategoryRepository) *Handler {
	h.categoryRepo = repo
	return h
}

func (h *Handler) WithReconciliationRepository(repo port.ReconciliationRepository) *Handler {
	h.reconciliationRepo = repo
	return h
}

func (h *Handler) WithPayslipService(svc port.PayslipUseCase) *Handler {
	h.payslipSvc = svc
	return h
}

func (h *Handler) WithPayslipRepository(repo port.PayslipRepository) *Handler {
	h.payslipRepo = repo
	return h
}

func (h *Handler) WithNotificationService(svc port.NotificationUseCase) *Handler {
	h.notificationSvc = svc
	return h
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.healthCheck)
	r.Post("/api/v1/login", h.login)
	r.Post("/api/v1/logout", h.logout)
	r.Post("/api/v1/auth/forgot-password", h.forgotPassword)
	r.Get("/api/v1/auth/reset-password/validate", h.validateResetToken)
	r.Post("/api/v1/auth/reset-password/confirm", h.confirmPasswordReset)

	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware)

		r.Route("/api/v1", func(r chi.Router) {
			r.Post("/auth/change-password", h.changePassword)
			r.Get("/auth/me", h.getMe) // MUST BE OUTSIDE ADMIN MIDDLEWARE

			r.Get("/system/info", h.getSystemInfo)

			r.Route("/users", func(r chi.Router) {
				r.Use(h.adminMiddleware) // ONLY ADMINS CAN ACCESS THESE
				r.Get("/", h.listUsers)
				r.Get("/{id}", h.getUser)
				r.Post("/", h.createUser)
				r.Put("/{id}", h.updateUser)
				r.Delete("/{id}", h.deleteUser)
			})

			r.Group(func(r chi.Router) {
				r.Use(h.adminMiddleware)
				r.Route("/settings", func(r chi.Router) {
					r.Get("/", h.getSettings)
					r.Patch("/", h.updateSettings)
					r.Post("/test-email", h.sendTestEmail)
				})
			})

			r.Route("/invoices", func(r chi.Router) {
				r.Get("/", h.listInvoices)
				r.Post("/import", h.importInvoice)
				r.Get("/{id}", h.getInvoice)
				r.Put("/{id}", h.updateInvoice)
				r.Delete("/{id}", h.deleteInvoice)
				r.Get("/{id}/download", h.downloadInvoiceFile)
			})

			r.Route("/bank-statements", func(r chi.Router) {
				r.Get("/{id}/download", h.downloadBankStatementFile)
				r.Get("/", h.listBankStatements)
				r.Get("/{id}", h.getBankStatement)
				r.Post("/import", h.importBankStatement)
				r.Delete("/{id}", h.deleteBankStatement)
			})

			r.Route("/transactions", func(r chi.Router) {
				r.Get("/analytics", h.getTransactionAnalytics)
				r.Get("/", h.listTransactions)
				r.Patch("/{hash}/category", h.updateTransactionCategory)
				r.Patch("/{hash}/review", h.markTransactionReviewed)
				r.Post("/auto-categorize/start", h.startAutoCategorize)
				r.Get("/auto-categorize/status", h.getAutoCategorizeStatus)
				r.Post("/auto-categorize/cancel", h.cancelAutoCategorize)
			})

			r.Route("/categories", func(r chi.Router) {
				r.Get("/", h.listCategories)
				r.Post("/", h.createCategory)
				r.Put("/{id}", h.updateCategory)
				r.Delete("/{id}", h.deleteCategory)
			})

			r.Route("/reconciliations", func(r chi.Router) {
				r.Get("/suggestions", h.getReconciliationSuggestions)
				r.Get("/", h.listReconciliations)
				r.Post("/", h.createReconciliation)
				r.Delete("/{id}", h.deleteReconciliation)
			})

			r.Route("/payslips", func(r chi.Router) {
				r.Get("/", h.listPayslips)
				r.Get("/{id}", h.getPayslip)
				r.Post("/import", h.importPayslip)
				r.Post("/import/batch", h.importPayslipsBatch)
				r.Put("/{id}", h.updatePayslip)
				r.Patch("/{id}", h.updatePayslip)
				r.Delete("/{id}", h.deletePayslip)
				r.Get("/{id}/download", h.downloadPayslipFile)
			})

			r.Route("/bank", func(r chi.Router) {
				r.Get("/institutions", h.listBankInstitutions)
				r.Post("/sync", h.syncAllBankAccounts)

				r.Route("/accounts", func(r chi.Router) {
					r.Put("/{id}/type", h.updateBankAccountType)
				})

				r.Route("/connections", func(r chi.Router) {
					r.Get("/", h.listBankConnections)
					r.Post("/", h.createBankConnection)
					r.Post("/finish", h.finishBankConnection)
					r.Delete("/{id}", h.deleteBankConnection)
				})
			})
		})
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
