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

	// Standard middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(120 * time.Second))

	// CORS Configuration
	allowedOrigins := []string{}
	allowCredentials := true
	var allowOriginFunc func(r *http.Request, origin string) bool

	// Default development policy: allow localhost and local network ranges with reflection
	allowOriginFunc = func(r *http.Request, origin string) bool {
		if strings.HasPrefix(origin, "http://localhost:") ||
			strings.HasPrefix(origin, "https://localhost") ||
			strings.HasPrefix(origin, "http://127.0.0.1:") ||
			strings.HasPrefix(origin, "https://127.0.0.1") ||
			strings.HasPrefix(origin, "http://192.168.") ||
			strings.HasPrefix(origin, "http://10.") ||
			strings.HasPrefix(origin, "http://172.") {
			return true
		}
		return false
	}

	if raw := os.Getenv("ALLOWED_ORIGINS"); raw != "" {
		if raw == "*" {
			// In development, reflect the origin to allow credentials (Authorization header)
			allowOriginFunc = func(r *http.Request, origin string) bool {
				return true
			}
			handler.Logger.Warn("CORS: Wildcard origin enabled via reflection. This should only be used in development.")
		} else {
			for _, o := range strings.Split(raw, ",") {
				if trimmed := strings.TrimSpace(o); trimmed != "" {
					allowedOrigins = append(allowedOrigins, trimmed)
				}
			}
			// If specific origins are provided, we disable the dynamic localhost matching
			allowOriginFunc = nil
		}
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowOriginFunc:  allowOriginFunc,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-Id", "X-Bridge-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: allowCredentials,
		MaxAge:           300,
	}))

	// Security headers
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("X-XSS-Protection", "0")
			next.ServeHTTP(w, r)
		})
	})

	handler.RegisterRoutes(r)
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      r,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 120 * time.Second,
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
