package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

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
	categorySvc           port.CategoryUseCase
	settingsSvc           port.SettingsUseCase
	payslipSvc            port.PayslipUseCase
	bankSvc               port.BankUseCase
	forecastingSvc        port.ForecastingUseCase
	userSvc               port.UserUseCase
	plannedTransactionSvc port.PlannedTransactionUseCase
	sharingSvc            port.SharingUseCase
	discoverySvc          port.DiscoveryUseCase
	notificationSvc       port.NotificationUseCase
	bridgeTokenSvc        port.BridgeAccessTokenUseCase
	documentSvc           port.DocumentUseCase
	payslipRepo           port.PayslipRepository
	bankStmtRepo          port.BankStatementRepository
	reconciliationRepo    port.ReconciliationRepository
	Logger                *slog.Logger
	LogLevel              *slog.LevelVar
	AppCtx                context.Context
	WaitGroup             *sync.WaitGroup
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
	appCtx context.Context,
	wg *sync.WaitGroup,
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
		AppCtx:           appCtx,
		WaitGroup:        wg,
		storageMode:      storageMode,
		dbHost:           dbHost,
		dbPinger:         dbPinger,
	}
}

func (h *Handler) WithLogLevel(v *slog.LevelVar) *Handler {
	h.LogLevel = v
	return h
}

func (h *Handler) WithBankService(svc port.BankUseCase) *Handler {
	h.bankSvc = svc
	return h
}

func (h *Handler) WithInvoiceService(svc port.InvoiceUseCase) *Handler {
	h.invoiceSvc = svc
	return h
}

func (h *Handler) WithDocumentService(svc port.DocumentUseCase) *Handler {
	h.documentSvc = svc
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

func (h *Handler) WithCategoryService(svc port.CategoryUseCase) *Handler {
	h.categorySvc = svc
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

func (h *Handler) WithSharingService(svc port.SharingUseCase) *Handler {
	h.sharingSvc = svc
	return h
}

func (h *Handler) WithDiscoveryService(svc port.DiscoveryUseCase) *Handler {
	h.discoverySvc = svc
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

			r.With(h.adminMiddleware).Get("/system/info/", h.getSystemInfo)
			r.With(h.adminMiddleware).Get("/system/log-level/", h.getLogLevel)
			r.With(h.adminMiddleware).Put("/system/log-level/", h.updateLogLevel)

			r.Route("/users/", func(r chi.Router) {
				r.Use(h.adminMiddleware) // ONLY ADMINS CAN ACCESS THESE
				r.Get("/", h.listUsers)
				r.Get("/{id}/", h.getUser)
				r.Post("/", h.createUser)
				r.Put("/{id}/", h.updateUser)
				r.Delete("/{id}/", h.deleteUser)
			})

			r.Route("/settings/", func(r chi.Router) {
				r.Get("/", h.getSettings)
				r.Patch("/", h.updateSettings)

				r.Group(func(r chi.Router) {
					r.Use(h.adminMiddleware)
					r.Post("/test-email/", h.sendTestEmail)
				})
			})

			r.Route("/invoices/", func(r chi.Router) {
				r.Get("/", h.listInvoices)
				r.Post("/", h.importManual)
				r.Post("/import/", h.importInvoice)
				r.Get("/{id}/", h.getInvoice)
				r.Put("/{id}/", h.updateInvoice)
				r.Delete("/{id}/", h.deleteInvoice)
				r.Get("/{id}/download/", h.downloadInvoiceFile)

				// Sharing
				r.Post("/{id}/share/", h.shareInvoice)
				r.Delete("/{id}/share/{user_id}/", h.revokeInvoiceShare)
				r.Get("/{id}/shares/", h.listInvoiceShares)
			})

			r.Route("/sharing/", func(r chi.Router) {
				r.Get("/dashboard/", h.getSharingDashboard)
			})

			r.Route("/subscriptions/", func(r chi.Router) {
				r.Get("/", h.ListSubscriptions)
				r.Get("/suggested/", h.GetSuggestedSubscriptions)
				r.Get("/feedback/", h.GetDiscoveryFeedback)
				r.Post("/approve/", h.ApproveSubscription)
				r.Post("/decline/", h.DeclineSubscription)
				r.Post("/remove-feedback/", h.RemoveDiscoveryFeedback)
				r.Get("/{id}/", h.GetSubscription)
				r.Put("/{id}/", h.UpdateSubscription)
				r.Delete("/{id}/", h.DeleteSubscription)
				r.Post("/{id}/enrich/", h.EnrichSubscription)
				r.Post("/{id}/preview-cancellation/", h.PreviewCancellation)
				r.Post("/{id}/cancel/", h.CancelSubscription)
				r.Get("/{id}/events/", h.GetSubscriptionEvents)
				r.Post("/{id}/transactions/link/", h.LinkTransactions)
				r.Post("/{id}/transactions/{hash}/link/", h.LinkTransaction)
				r.Delete("/{id}/transactions/{hash}/unlink/", h.UnlinkTransaction)
				r.Post("/from-transaction/", h.CreateSubscriptionFromTransaction)
			})

			r.Route("/bank-statements/", func(r chi.Router) {
				r.Get("/{id}/download/", h.downloadBankStatementFile)
				r.Get("/", h.listBankStatements)
				r.Get("/{id}/", h.getBankStatement)
				r.Post("/import/", h.importBankStatement)
				r.Patch("/{id}/", h.updateBankStatement)
				r.Delete("/{id}/", h.deleteBankStatement)
			})

			r.Route("/transactions/", func(r chi.Router) {
				r.Get("/analytics/", h.getTransactionAnalytics)
				r.Get("/forecast/", h.getForecast)

				r.Get("/", h.listTransactions)
				r.Patch("/bulk-review/", h.markTransactionsReviewedBulk)
				r.Patch("/{hash}/category/", h.updateTransactionCategory)
				r.Patch("/{hash}/review/", h.markTransactionReviewed)
				r.Post("/auto-categorize/start/", h.startAutoCategorize)
				r.Get("/auto-categorize/status/", h.getAutoCategorizeStatus)
				r.Post("/auto-categorize/cancel/", h.cancelAutoCategorize)
			})

			r.Route("/categories/", func(r chi.Router) {
				r.Get("/", h.listCategories)
				r.Post("/", h.createCategory)

				r.Route("/{id}/", func(r chi.Router) {
					r.Put("/", h.updateCategory)
					r.Delete("/", h.deleteCategory)
					r.Post("/restore/", h.restoreCategory)
					r.Get("/average/", h.getCategoryAverage)

					// Sharing
					r.Post("/share/", h.shareCategory)
					r.Delete("/share/{user_id}/", h.revokeCategoryShare)
					r.Get("/shares/", h.listCategoryShares)
				})
			})

			reconciliationRoutes := func(r chi.Router) {
				r.Get("/suggestions/", h.getReconciliationSuggestions)
				r.Get("/", h.listReconciliations)
				r.Post("/", h.createReconciliation)
				r.Delete("/{id}/", h.deleteReconciliation)
			}
			r.Route("/reconciliations/", reconciliationRoutes)
			r.Route("/reconciliation/", reconciliationRoutes) // Singular alias for robustness

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
					r.Post("/virtual/", h.createVirtualBankAccount)
					r.Put("/{id}/type/", h.updateBankAccountType)

					// Sharing
					r.Post("/{id}/share/", h.shareBankAccount)
					r.Delete("/{id}/share/{user_id}/", h.revokeBankAccountShare)
					r.Get("/{id}/shares/", h.listBankAccountShares)
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

			r.Route("/documents/", func(r chi.Router) {
				r.Get("/", h.listDocuments)
				r.Post("/upload/", h.uploadDocument)
				r.Get("/tax-summary/{year}/", h.getTaxYearSummary)
				r.Get("/{id}/", h.getDocument)
				r.Put("/{id}/", h.updateDocument) // <-- ADD THIS
				r.Delete("/{id}/", h.deleteDocument)
				r.Get("/{id}/download/", h.downloadDocument)
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

func (h *Handler) getClientIP(r *http.Request) string {
	// 1. Check X-Forwarded-For (standard for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	// 2. Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 3. Fallback to RemoteAddr
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
