package http

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"cogni-cash/internal/domain/port"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Router struct {
	Auth               *AuthHandler
	Invoice            *InvoiceHandler
	BankStatement      *BankStatementHandler
	Category           *CategoryHandler
	Payslip            *PayslipHandler
	Bank               *BankHandler
	Forecasting        *ForecastingHandler
	User               *UserHandler
	PlannedTransaction *PlannedTransactionHandler
	Sharing            *SharingHandler
	Discovery          *DiscoveryHandler
	BridgeToken        *BridgeTokenHandler
	Document           *DocumentHandler
	System             *SystemHandler
	Reconciliation     *ReconciliationHandler

	authSvc        port.AuthUseCase
	userSvc        port.UserUseCase
	bridgeTokenSvc port.BridgeAccessTokenUseCase

	Logger    *slog.Logger
	AppCtx    context.Context
	WaitGroup *sync.WaitGroup
}

func NewRouter(
	logger *slog.Logger,
	appCtx context.Context,
	wg *sync.WaitGroup,
	authSvc port.AuthUseCase,
	userSvc port.UserUseCase,
	bridgeTokenSvc port.BridgeAccessTokenUseCase,
) *Router {
	if logger == nil {
		logger = slog.Default()
	}
	return &Router{
		Logger:         logger,
		AppCtx:         appCtx,
		WaitGroup:      wg,
		authSvc:        authSvc,
		userSvc:        userSvc,
		bridgeTokenSvc: bridgeTokenSvc,
	}
}

func (router *Router) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Check for Bridge Access Token (BAT)
		bridgeToken := r.Header.Get("X-Bridge-Token")
		if bridgeToken != "" && router.bridgeTokenSvc != nil {
			userID, err := router.bridgeTokenSvc.ValidateToken(r.Context(), bridgeToken)
			if err == nil {
				ctx := context.WithValue(r.Context(), userIDKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			router.Logger.Warn("Invalid bridge token attempted", "token_prefix", bridgeToken[:4])
		}

		// 2. Fallback to standard JWT (Bearer or Cookie)
		var tokenStr string
		authHeader := r.Header.Get("Authorization")

		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// Fallback to cookie if Authorization header is missing
			cookie, err := r.Cookie(authTokenCookieName)
			if err == nil {
				tokenStr = cookie.Value
			}
		}

		if tokenStr == "" {
			writeError(w, http.StatusUnauthorized, "missing or invalid authorization")
			return
		}

		userID, err := router.authSvc.ValidateToken(tokenStr)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (router *Router) adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserID(r.Context())
		if userID == uuid.Nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		user, err := router.userSvc.GetUser(r.Context(), userID.String())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to verify user permissions")
			return
		}

		if user.Role != "admin" {
			router.Logger.Warn("Forbidden action attempted by non-admin", "user_id", userID.String(), "path", r.URL.Path)
			writeError(w, http.StatusForbidden, "forbidden: administrator access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (router *Router) RegisterRoutes(r chi.Router) {
	r.Get("/health", router.System.healthCheck)
	r.Post("/api/v1/login/", router.Auth.login)
	r.Post("/api/v1/logout/", router.Auth.logout)
	r.Post("/api/v1/auth/refresh/", router.Auth.refresh)
	r.Post("/api/v1/auth/forgot-password/", router.Auth.forgotPassword)
	r.Get("/api/v1/auth/reset-password/validate/", router.Auth.validateResetToken)
	r.Post("/api/v1/auth/reset-password/confirm/", router.Auth.confirmPasswordReset)

	r.Group(func(r chi.Router) {
		r.Use(router.authMiddleware)

		r.Route("/api/v1", func(r chi.Router) {
			r.Post("/auth/change-password/", router.Auth.changePassword)
			r.Get("/auth/me/", router.User.getMe) // MUST BE OUTSIDE ADMIN MIDDLEWARE

			r.With(router.adminMiddleware).Get("/system/info/", router.System.getSystemInfo)
			r.With(router.adminMiddleware).Get("/system/log-level/", router.System.getLogLevel)
			r.With(router.adminMiddleware).Put("/system/log-level/", router.System.updateLogLevel)

			r.Route("/users/", func(r chi.Router) {
				r.Use(router.adminMiddleware) // ONLY ADMINS CAN ACCESS THESE
				r.Get("/", router.User.listUsers)
				r.Get("/{id}/", router.User.getUser)
				r.Post("/", router.User.createUser)
				r.Put("/{id}/", router.User.updateUser)
				r.Delete("/{id}/", router.User.deleteUser)
			})

			r.Route("/settings/", func(r chi.Router) {
				r.Get("/", router.System.getSettings)
				r.Patch("/", router.System.updateSettings)

				r.Group(func(r chi.Router) {
					r.Use(router.adminMiddleware)
					r.Post("/test-email/", router.System.sendTestEmail)
				})
			})

			r.Route("/invoices/", func(r chi.Router) {
				r.Get("/", router.Invoice.listInvoices)
				r.Post("/", router.Invoice.importManual)
				r.Post("/import/", router.Invoice.importInvoice)
				r.Get("/{id}/", router.Invoice.getInvoice)
				r.Put("/{id}/", router.Invoice.updateInvoice)
				r.Delete("/{id}/", router.Invoice.deleteInvoice)
				r.Get("/{id}/download/", router.Invoice.downloadInvoiceFile)

				// Sharing
				r.Post("/{id}/share/", router.Invoice.shareInvoice)
				r.Delete("/{id}/share/{user_id}/", router.Invoice.revokeInvoiceShare)
				r.Get("/{id}/shares/", router.Invoice.listInvoiceShares)
			})

			r.Route("/sharing/", func(r chi.Router) {
				r.Get("/dashboard/", router.Sharing.getSharingDashboard)
			})

			r.Route("/subscriptions/", func(r chi.Router) {
				r.Get("/", router.Discovery.ListSubscriptions)
				r.Get("/suggested/", router.Discovery.GetSuggestedSubscriptions)
				r.Get("/feedback/", router.Discovery.GetDiscoveryFeedback)
				r.Post("/approve/", router.Discovery.ApproveSubscription)
				r.Post("/decline/", router.Discovery.DeclineSubscription)
				r.Post("/remove-feedback/", router.Discovery.RemoveDiscoveryFeedback)
				r.Get("/{id}/", router.Discovery.GetSubscription)
				r.Put("/{id}/", router.Discovery.UpdateSubscription)
				r.Delete("/{id}/", router.Discovery.DeleteSubscription)
				r.Post("/{id}/enrich/", router.Discovery.EnrichSubscription)
				r.Post("/{id}/preview-cancellation/", router.Discovery.PreviewCancellation)
				r.Post("/{id}/cancel/", router.Discovery.CancelSubscription)
				r.Get("/{id}/events/", router.Discovery.GetSubscriptionEvents)
				r.Post("/{id}/transactions/link/", router.Discovery.LinkTransactions)
				r.Post("/{id}/transactions/{hash}/link/", router.Discovery.LinkTransaction)
				r.Delete("/{id}/transactions/{hash}/unlink/", router.Discovery.UnlinkTransaction)
				r.Post("/from-transaction/", router.Discovery.CreateSubscriptionFromTransaction)
			})

			r.Route("/bank-statements/", func(r chi.Router) {
				r.Get("/{id}/download/", router.BankStatement.downloadBankStatementFile)
				r.Get("/", router.BankStatement.listBankStatements)
				r.Get("/{id}/", router.BankStatement.getBankStatement)
				r.Post("/import/", router.BankStatement.importBankStatement)
				r.Patch("/{id}/", router.BankStatement.updateBankStatement)
				r.Delete("/{id}/", router.BankStatement.deleteBankStatement)
			})

			r.Route("/transactions/", func(r chi.Router) {
				r.Get("/analytics/", router.BankStatement.getTransactionAnalytics)
				r.Get("/forecast/", router.Forecasting.getForecast)

				r.Get("/", router.BankStatement.listTransactions)
				r.Patch("/bulk-review/", router.BankStatement.markTransactionsReviewedBulk)
				r.Patch("/{hash}/category/", router.BankStatement.updateTransactionCategory)
				r.Patch("/{hash}/review/", router.BankStatement.markTransactionReviewed)
				r.Post("/auto-categorize/start/", router.BankStatement.startAutoCategorize)
				r.Get("/auto-categorize/status/", router.BankStatement.getAutoCategorizeStatus)
				r.Post("/auto-categorize/cancel/", router.BankStatement.cancelAutoCategorize)
			})

			r.Route("/categories/", func(r chi.Router) {
				r.Get("/", router.Category.listCategories)
				r.Post("/", router.Category.createCategory)

				r.Route("/{id}/", func(r chi.Router) {
					r.Put("/", router.Category.updateCategory)
					r.Delete("/", router.Category.deleteCategory)
					r.Post("/restore/", router.Category.restoreCategory)
					r.Get("/average/", router.Category.getCategoryAverage)

					// Sharing
					r.Post("/share/", router.Category.shareCategory)
					r.Delete("/share/{user_id}/", router.Category.revokeCategoryShare)
					r.Get("/shares/", router.Category.listCategoryShares)
				})
			})

			reconciliationRoutes := func(r chi.Router) {
				r.Get("/suggestions/", router.Reconciliation.getReconciliationSuggestions)
				r.Get("/", router.Reconciliation.listReconciliations)
				r.Post("/", router.Reconciliation.createReconciliation)
				r.Delete("/{id}/", router.Reconciliation.deleteReconciliation)
			}
			r.Route("/reconciliations/", reconciliationRoutes)
			r.Route("/reconciliation/", reconciliationRoutes) // Singular alias for robustness

			r.Route("/planned-transactions/", func(r chi.Router) {
				r.Get("/", router.PlannedTransaction.listPlannedTransactions)
				r.Post("/", router.PlannedTransaction.createPlannedTransaction)
				r.Put("/{id}/", router.PlannedTransaction.updatePlannedTransaction)
				r.Delete("/{id}/", router.PlannedTransaction.deletePlannedTransaction)
			})

			r.Route("/payslips/", func(r chi.Router) {
				r.Get("/summary/", router.Payslip.getPayslipSummary)
				r.Get("/", router.Payslip.listPayslips)
				r.Get("/{id}/", router.Payslip.getPayslip)
				r.Post("/import/", router.Payslip.importPayslip)
				r.Post("/import/batch/", router.Payslip.importPayslipsBatch)
				r.Put("/{id}/", router.Payslip.updatePayslip)
				r.Patch("/{id}/", router.Payslip.updatePayslip)
				r.Delete("/{id}/", router.Payslip.deletePayslip)
				r.Get("/{id}/download/", router.Payslip.downloadPayslipFile)
			})

			r.Route("/bank/", func(r chi.Router) {
				r.Get("/institutions/", router.Bank.listBankInstitutions)
				r.Post("/sync/", router.Bank.syncAllBankAccounts)

				r.Route("/accounts/", func(r chi.Router) {
					r.Post("/virtual/", router.Bank.createVirtualBankAccount)
					r.Put("/{id}/type/", router.Bank.updateBankAccountType)

					// Sharing
					r.Post("/{id}/share/", router.Bank.shareBankAccount)
					r.Delete("/{id}/share/{user_id}/", router.Bank.revokeBankAccountShare)
					r.Get("/{id}/shares/", router.Bank.listBankAccountShares)
				})

				r.Route("/connections/", func(r chi.Router) {
					r.Get("/", router.Bank.listBankConnections)
					r.Post("/", router.Bank.createBankConnection)
					r.Post("/finish/", router.Bank.finishBankConnection)
					r.Delete("/{id}/", router.Bank.deleteBankConnection)
				})
			})

			r.Route("/bridge-tokens/", func(r chi.Router) {
				r.Get("/", router.BridgeToken.listBridgeTokens)
				r.Post("/", router.BridgeToken.createBridgeToken)
				r.Delete("/{id}/", router.BridgeToken.revokeBridgeToken)
			})

			r.Route("/documents/", func(r chi.Router) {
				r.Get("/", router.Document.listDocuments)
				r.Post("/upload/", router.Document.uploadDocument)
				r.Get("/tax-summary/{year}/", router.Document.getTaxYearSummary)
				r.Get("/{id}/", router.Document.getDocument)
				r.Put("/{id}/", router.Document.updateDocument) // <-- ADD THIS
				r.Delete("/{id}/", router.Document.deleteDocument)
				r.Get("/{id}/download/", router.Document.downloadDocument)
			})
		})
	})
}

func (r *Router) WithLogLevel(v *slog.LevelVar) *Router {
	r.System = NewSystemHandler(v, r.Logger, "localhost", nil, nil, nil, "memory", r.userSvc)
	return r
}

func (r *Router) WithUserService(svc port.UserUseCase) *Router {
	r.User = NewUserHandler(r.AppCtx, r.Logger, r.WaitGroup, nil, svc)
	// Must update the router's reference to userSvc for middlewares
	r.userSvc = svc
	r.Auth = NewAuthHandler(r.Logger, r.authSvc, r.bridgeTokenSvc, svc)
	if r.System != nil {
		r.System = NewSystemHandler(r.System.LogLevel, r.Logger, r.System.dbHost, r.System.dbPinger, r.System.notificationSvc, r.System.settingsSvc, r.System.storageMode, svc)
	}
	return r
}

func (r *Router) WithTransactionService(svc port.TransactionUseCase) *Router {
	if r.BankStatement == nil {
		r.BankStatement = NewBankStatementHandler(r.Logger, nil, nil, nil, nil, svc)
	} else {
		r.BankStatement = NewBankStatementHandler(r.Logger, nil, nil, nil, nil, svc)
	}
	return r
}

func (r *Router) WithCategoryService(svc port.CategoryUseCase) *Router {
	r.Category = NewCategoryHandler(svc, nil)
	return r
}

func (r *Router) WithSharingService(svc port.SharingUseCase) *Router {
	r.Sharing = NewSharingHandler(r.Logger, svc)
	return r
}

func (r *Router) WithNotificationService(svc port.NotificationUseCase) *Router {
	// Notification Service is used by SystemHandler
	r.System = NewSystemHandler(&slog.LevelVar{}, r.Logger, "localhost", nil, svc, nil, "memory", r.userSvc)
	return r
}

func (r *Router) WithInvoiceService(svc port.InvoiceUseCase) *Router {
	r.Invoice = NewInvoiceHandler(r.Logger, svc)
	return r
}

func (r *Router) WithReconciliationService(svc port.ReconciliationUseCase) *Router {
	r.Reconciliation = NewReconciliationHandler(r.Logger, nil, svc)
	return r
}

func (r *Router) WithForecastingService(svc port.ForecastingUseCase) *Router {
	r.Forecasting = NewForecastingHandler(r.Logger, svc)
	return r
}

func (r *Router) WithPlannedTransactionService(svc port.PlannedTransactionUseCase) *Router {
	r.PlannedTransaction = NewPlannedTransactionHandler(r.Logger, svc)
	return r
}

func (r *Router) WithBridgeTokenService(svc port.BridgeAccessTokenUseCase) *Router {
	r.BridgeToken = NewBridgeTokenHandler(r.Logger, svc)
	return r
}

func (r *Router) WithDocumentService(svc port.DocumentUseCase) *Router {
	r.Document = NewDocumentHandler(r.Logger, svc)
	return r
}

func (r *Router) WithDiscoveryService(svc port.DiscoveryUseCase) *Router {
	r.Discovery = NewDiscoveryHandler(r.AppCtx, r.Logger, r.WaitGroup, svc)
	return r
}

func (r *Router) WithBankService(svc port.BankUseCase) *Router {
	r.Bank = NewBankHandler(r.AppCtx, r.Logger, r.WaitGroup, svc)
	return r
}

func (r *Router) WithBankStatementRepository(repo port.BankStatementRepository) *Router {
	if r.BankStatement == nil {
		r.BankStatement = NewBankStatementHandler(r.Logger, nil, repo, nil, nil, nil)
	} else {
		r.BankStatement = NewBankStatementHandler(r.Logger, r.BankStatement.bankStatementSvc, repo, r.BankStatement.forecastingSvc, r.BankStatement.settingsSvc, r.BankStatement.transactionSvc)
	}
	return r
}

func (r *Router) WithReconciliationRepository(repo port.ReconciliationRepository) *Router {
	if r.Reconciliation == nil {
		r.Reconciliation = NewReconciliationHandler(r.Logger, repo, nil)
	} else {
		r.Reconciliation = NewReconciliationHandler(r.Logger, repo, nil)
	}
	return r
}

func (r *Router) WithPayslipService(svc port.PayslipUseCase) *Router {
	var repo port.PayslipRepository
	if r.Payslip != nil {
		repo = r.Payslip.payslipRepo
	}
	r.Payslip = NewPayslipHandler(r.Logger, repo, svc)
	return r
}

func (r *Router) WithPayslipRepository(repo port.PayslipRepository) *Router {
	if r.Payslip == nil {
		r.Payslip = NewPayslipHandler(r.Logger, repo, nil)
	} else {
		r.Payslip = NewPayslipHandler(r.Logger, repo, r.Payslip.payslipSvc)
	}
	return r
}
