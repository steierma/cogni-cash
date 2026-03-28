package http

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// Server wraps the http.Server and the chi router.
type Server struct {
	httpServer *http.Server
	Logger     *slog.Logger
}

// NewServer creates a configured HTTP server.
func NewServer(addr string, handler *Handler) *Server {
	r := chi.NewRouter()

	// CORS — origins are read from ALLOWED_ORIGINS (comma-separated).
	//
	// ⚠ go-chi/cors treats an *empty* AllowedOrigins slice as "allow all".
	//   We therefore default to a deny-all sentinel ("null") so that when
	//   ALLOWED_ORIGINS is unset the backend blocks all cross-origin requests.
	//   Set ALLOWED_ORIGINS=http://localhost:3000 (or comma-separated list)
	//   in backend/.env to allow your frontend origin.
	allowedOrigins := []string{"null"} // deny-all default
	if raw := os.Getenv("ALLOWED_ORIGINS"); raw != "" {
		parsed := []string{}
		for _, o := range strings.Split(raw, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				parsed = append(parsed, trimmed)
			}
		}
		if len(parsed) > 0 {
			allowedOrigins = parsed
		}
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-Id"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Security headers — applied to every response from the backend.
	// (The Nginx reverse-proxy adds the same headers for the SPA, but the
	// backend also exposes :8080 directly in development, so we harden here
	// too.)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("X-XSS-Protection", "0") // modern browsers ignore this; CSP is the real guard
			next.ServeHTTP(w, r)
		})
	})

	// Standard middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	handler.RegisterRoutes(r)

	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      r,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		Logger: handler.Logger,
	}
}

// Start begins listening and blocks until the server stops.
func (s *Server) Start() error {
	s.Logger.Info("server listening", "addr", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.Logger.Info("server shutting down")
	return s.httpServer.Shutdown(ctx)
}
