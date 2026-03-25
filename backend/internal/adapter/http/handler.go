package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"cogni-cash/internal/domain/port"
	"cogni-cash/internal/domain/service"

	"github.com/go-chi/chi/v5"
)

type contextKey string

const userIDKey contextKey = "userID"

type Handler struct {
	authSvc            *service.AuthService
	categorizationSvc  *service.CategorizationService
	bankStatementSvc   *service.BankStatementService
	transactionSvc     *service.TransactionService
	reconciliationSvc  *service.ReconciliationService
	settingsSvc        *service.SettingsService
	payslipSvc         *service.PayslipService
	userSvc            *service.UserService
	payslipRepo        port.PayslipRepository
	invoiceRepo        port.InvoiceRepository
	bankStmtRepo       port.BankStatementRepository
	categoryRepo       port.CategoryRepository
	reconciliationRepo port.ReconciliationRepository
	Logger             *slog.Logger
	storageMode        string
	dbHost             string
	dbPinger           func(context.Context) error
}

func NewHandler(authSvc *service.AuthService, svc *service.CategorizationService, bankSvc *service.BankStatementService, settingsSvc *service.SettingsService, repo port.InvoiceRepository, logger *slog.Logger, storageMode string, dbHost string, dbPinger func(context.Context) error) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		authSvc:           authSvc,
		categorizationSvc: svc,
		bankStatementSvc:  bankSvc,
		settingsSvc:       settingsSvc,
		invoiceRepo:       repo,
		Logger:            logger,
		storageMode:       storageMode,
		dbHost:            dbHost,
		dbPinger:          dbPinger,
	}
}

func (h *Handler) WithUserService(svc *service.UserService) *Handler {
	h.userSvc = svc
	return h
}

func (h *Handler) WithTransactionService(svc *service.TransactionService) *Handler {
	h.transactionSvc = svc
	return h
}

func (h *Handler) WithReconciliationService(svc *service.ReconciliationService) *Handler {
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

func (h *Handler) WithPayslipService(svc *service.PayslipService) *Handler {
	h.payslipSvc = svc
	return h
}

func (h *Handler) WithPayslipRepository(repo port.PayslipRepository) *Handler {
	h.payslipRepo = repo
	return h
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.healthCheck)
	r.Post("/api/v1/login", h.login)

	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware)

		r.Route("/api/v1", func(r chi.Router) {
			r.Post("/auth/change-password", h.changePassword)
			r.Get("/auth/me", h.getMe) // MUST BE OUTSIDE ADMIN MIDDLEWARE

			r.Get("/system/info", h.getSystemInfo)
			r.Route("/settings", func(r chi.Router) {
				r.Get("/", h.getSettings)
				r.Patch("/", h.updateSettings)
			})

			r.Route("/users", func(r chi.Router) {
				r.Use(h.adminMiddleware) // ONLY ADMINS CAN ACCESS THESE
				r.Get("/", h.listUsers)
				r.Get("/{id}", h.getUser)
				r.Post("/", h.createUser)
				r.Put("/{id}", h.updateUser)
				r.Delete("/{id}", h.deleteUser)
			})

			r.Route("/invoices", func(r chi.Router) {
				r.Get("/", h.listInvoices)
				r.Get("/{id}", h.getInvoice)
				r.Post("/categorize", h.categorizeDocument)
				r.Delete("/{id}", h.deleteInvoice)
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
