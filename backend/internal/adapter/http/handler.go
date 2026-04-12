package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"cogni-cash/internal/domain/port"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "userID"

type Handler struct {
	authSvc               port.AuthUseCase
	invoiceSvc            port.InvoiceUseCase
	bankStatementSvc      port.BankStatementUseCase
	transactionSvc        port.TransactionUseCase
	reconciliationSvc     port.ReconciliationUseCase
	settingsSvc           port.SettingsUseCase
	payslipSvc            port.PayslipUseCase
	bankSvc               port.BankUseCase
	forecastingSvc        port.ForecastingUseCase
	userSvc               port.UserUseCase
	plannedTransactionSvc port.PlannedTransactionUseCase
	notificationSvc       port.NotificationUseCase
	bridgeTokenSvc        port.BridgeAccessTokenUseCase
	payslipRepo           port.PayslipRepository
	bankStmtRepo          port.BankStatementRepository
	categoryRepo          port.CategoryRepository
	reconciliationRepo    port.ReconciliationRepository
	Logger                *slog.Logger
	storageMode           string
	dbHost                string
	dbPinger              func(context.Context) error
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

func (h *Handler) WithForecastingService(svc port.ForecastingUseCase) *Handler {
	h.forecastingSvc = svc
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

func (h *Handler) WithPlannedTransactionService(svc port.PlannedTransactionUseCase) *Handler {
	h.plannedTransactionSvc = svc
	return h
}

func (h *Handler) WithBridgeTokenService(svc port.BridgeAccessTokenUseCase) *Handler {
	h.bridgeTokenSvc = svc
	return h
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.healthCheck)
	r.Post("/api/v1/login/", h.login)
	r.Post("/api/v1/logout/", h.logout)
	r.Post("/api/v1/auth/refresh/", h.refresh)
	r.Post("/api/v1/auth/forgot-password/", h.forgotPassword)
	r.Get("/api/v1/auth/reset-password/validate/", h.validateResetToken)
	r.Post("/api/v1/auth/reset-password/confirm/", h.confirmPasswordReset)

	r.Group(func(r chi.Router) {
		r.Use(h.authMiddleware)

		r.Route("/api/v1", func(r chi.Router) {
			r.Post("/auth/change-password/", h.changePassword)
			r.Get("/auth/me/", h.getMe) // MUST BE OUTSIDE ADMIN MIDDLEWARE

			r.Get("/system/info/", h.getSystemInfo)

			r.Route("/users/", func(r chi.Router) {
				r.Use(h.adminMiddleware) // ONLY ADMINS CAN ACCESS THESE
				r.Get("/", h.listUsers)
				r.Get("/{id}/", h.getUser)
				r.Post("/", h.createUser)
				r.Put("/{id}/", h.updateUser)
				r.Delete("/{id}/", h.deleteUser)
			})

			r.Group(func(r chi.Router) {
				r.Use(h.adminMiddleware)
				r.Route("/settings/", func(r chi.Router) {
					r.Get("/", h.getSettings)
					r.Patch("/", h.updateSettings)
					r.Post("/test-email/", h.sendTestEmail)
				})
			})

			r.Route("/invoices/", func(r chi.Router) {
				r.Get("/", h.listInvoices)
				r.Post("/import/", h.importInvoice)
				r.Get("/{id}/", h.getInvoice)
				r.Put("/{id}/", h.updateInvoice)
				r.Delete("/{id}/", h.deleteInvoice)
				r.Get("/{id}/download/", h.downloadInvoiceFile)
			})

			r.Route("/bank-statements/", func(r chi.Router) {
				r.Get("/{id}/download/", h.downloadBankStatementFile)
				r.Get("/", h.listBankStatements)
				r.Get("/{id}/", h.getBankStatement)
				r.Post("/import/", h.importBankStatement)
				r.Delete("/{id}/", h.deleteBankStatement)
			})

			r.Route("/transactions/", func(r chi.Router) {
				r.Get("/analytics/", h.getTransactionAnalytics)
				r.Get("/forecast/", h.getForecast)
				r.Post("/forecast/exclude/{id}/", h.excludeForecast)
				r.Post("/forecast/include/{id}/", h.includeForecast)

				r.Get("/forecast/patterns/exclusions/", h.listPatternExclusions)
				r.Post("/forecast/patterns/exclude/", h.excludePattern)
				r.Post("/forecast/patterns/include/", h.includePattern)

				r.Get("/", h.listTransactions)
				r.Patch("/{hash}/category/", h.updateTransactionCategory)
				r.Patch("/{hash}/review/", h.markTransactionReviewed)
				r.Patch("/{hash}/skip-forecasting/", h.toggleTransactionSkipForecasting)
				r.Post("/auto-categorize/start/", h.startAutoCategorize)
				r.Get("/auto-categorize/status/", h.getAutoCategorizeStatus)
				r.Post("/auto-categorize/cancel/", h.cancelAutoCategorize)
			})

			r.Route("/categories/", func(r chi.Router) {
				r.Get("/", h.listCategories)
				r.Post("/", h.createCategory)
				r.Put("/{id}/", h.updateCategory)
				r.Delete("/{id}/", h.deleteCategory)
				r.Post("/{id}/restore/", h.restoreCategory)
			})

			r.Route("/reconciliations/", func(r chi.Router) {
				r.Get("/suggestions/", h.getReconciliationSuggestions)
				r.Get("/", h.listReconciliations)
				r.Post("/", h.createReconciliation)
				r.Delete("/{id}/", h.deleteReconciliation)
			})

			r.Route("/planned-transactions/", func(r chi.Router) {
				r.Get("/", h.listPlannedTransactions)
				r.Post("/", h.createPlannedTransaction)
				r.Put("/{id}/", h.updatePlannedTransaction)
				r.Delete("/{id}/", h.deletePlannedTransaction)
			})

			r.Route("/payslips/", func(r chi.Router) {
				r.Get("/summary/", h.getPayslipSummary)
				r.Get("/", h.listPayslips)
				r.Get("/{id}/", h.getPayslip)
				r.Post("/import/", h.importPayslip)
				r.Post("/import/batch/", h.importPayslipsBatch)
				r.Put("/{id}/", h.updatePayslip)
				r.Patch("/{id}/", h.updatePayslip)
				r.Delete("/{id}/", h.deletePayslip)
				r.Get("/{id}/download/", h.downloadPayslipFile)
			})

			r.Route("/bank/", func(r chi.Router) {
				r.Get("/institutions/", h.listBankInstitutions)
				r.Post("/sync/", h.syncAllBankAccounts)

				r.Route("/accounts/", func(r chi.Router) {
					r.Put("/{id}/type/", h.updateBankAccountType)
				})

				r.Route("/connections/", func(r chi.Router) {
					r.Get("/", h.listBankConnections)
					r.Post("/", h.createBankConnection)
					r.Post("/finish/", h.finishBankConnection)
					r.Delete("/{id}/", h.deleteBankConnection)
				})
			})

			r.Route("/bridge-tokens/", func(r chi.Router) {
				r.Get("/", h.listBridgeTokens)
				r.Post("/", h.createBridgeToken)
				r.Delete("/{id}/", h.revokeBridgeToken)
			})
		})
	})
}

// mimeToExt returns the canonical file extension (including dot) for a MIME type.
func mimeToExt(mimeType string) string {
	switch mimeType {
	case "application/pdf":
		return ".pdf"
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "text/csv":
		return ".csv"
	case "application/vnd.ms-excel":
		return ".xls"
	}
	return ""
}

// isImageMIME returns true when the MIME type is a supported image format.
func isImageMIME(mimeType string) bool {
	switch mimeType {
	case "image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp":
		return true
	}
	return false
}

// resolveMIME returns the canonical, lower-cased MIME type for an uploaded file.
// It prefers the Content-Type header but falls back to extension-based detection
// when the header is absent or the generic "application/octet-stream".
// If nothing can be determined it returns "application/octet-stream".
func resolveMIME(contentTypeHeader, filename string) string {
	mt := contentTypeHeader
	if mt == "" || mt == "application/octet-stream" {
		ext := strings.ToLower(filepath.Ext(filename))
		if mapped, ok := extToMIME[ext]; ok {
			mt = mapped
		} else if detected := mime.TypeByExtension(ext); detected != "" {
			mt = detected
		}
	}
	// Strip parameters (e.g. "image/jpeg; charset=utf-8")
	if idx := strings.IndexByte(mt, ';'); idx != -1 {
		mt = strings.TrimSpace(mt[:idx])
	}
	mt = strings.ToLower(mt)
	if mt == "" {
		mt = "application/octet-stream"
	}
	return mt
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
