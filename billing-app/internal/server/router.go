package server

import (
	"github.com/diewo77/billing-app/auth"
	"github.com/diewo77/billing-app/internal/handlers"
	"github.com/diewo77/billing-app/httpx"
	"github.com/diewo77/billing-app/internal/middleware"
	"github.com/diewo77/billing-app/internal/models"
	"github.com/diewo77/billing-app/internal/services"
	"context"
	"net/http"
	"time"

	"gorm.io/gorm"
)

// New constructs the root http.Handler with all routes and middlewares applied.
func New(db *gorm.DB) http.Handler {
	mux := http.NewServeMux()

	// Configure a user verifier so RequireAuth can ensure the user still exists.
	auth.SetUserVerifier(func(_ context.Context, uid uint) bool {
		var count int64
		if err := db.Model(&models.User{}).Where("id = ?", uid).Limit(1).Count(&count).Error; err != nil {
			return false
		}
		return count > 0
	})

	// --- Health endpoints ---
	//revive:disable:unused-parameter simple handlers intentionally ignore *http.Request
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		// Perform a lightweight DB check (SELECT 1) – ignore detailed errors in body
		if err := db.Exec("SELECT 1").Error; err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Header().Set("Content-Type", "application/json")
			if _, werr := w.Write([]byte(`{"status":"degraded"}`)); werr != nil {
				_ = werr
			}
			return
		}
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Auth endpoints
	authHandler := handlers.NewAuthHandler(db)
	authHandler.Register(mux)

	// Setup endpoint (fully requires auth)
	setupSvc := services.NewSetupService(db)
	setupHandler := handlers.NewSetupHandler(setupSvc)
	mux.Handle("/setup", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { setupHandler.Handle(w, r) }))))

	// Product endpoints (CRUD-like subset). List/Create via /products. Update/Delete via /products/update & /products/delete for simplicity.
	ph := handlers.NewProductHandler(db)
	mux.Handle("/products", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			ph.List(w, r)
			return
		}
		if r.Method == http.MethodPost {
			ph.Create(w, r)
			return
		}
		w.Header().Set("Allow", "GET,POST")
		httpx.JSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", nil)
	}))))
	mux.Handle("/products/delete", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ph.Delete(w, r) }))))
	mux.Handle("/products/update", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ph.Update(w, r) }))))

	// Invoice endpoints
	invSvc := services.NewInvoiceService()
	ih := handlers.NewInvoiceHandler(db, invSvc)
	mux.Handle("/invoices", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ih.List(w, r)
		case http.MethodPost:
			ih.Create(w, r)
		default:
			w.Header().Set("Allow", "GET,POST")
			httpx.JSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", nil)
		}
	}))))
	mux.Handle("/invoices/finalize", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ih.Finalize(w, r) }))))
	mux.Handle("/invoices/pdf", auth.Middleware(auth.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ih.PDF(w, r) }))))

	// OpenAPI spec
	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		http.ServeFile(w, r, "openapi.yaml")
	})

	// Root placeholder (could be replaced by template rendering in main)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, werr := w.Write([]byte("Billing App API - see /openapi.yaml")); werr != nil {
			_ = werr
		}
	})
	//revive:enable:unused-parameter

	return middleware.Prefs(withRecover(withLogging(mux)))
}

// Simple middleware logging & recovery kept private to this package to avoid duplication.
func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		// Basic stdout log – replace by structured logger if needed
		// (Avoiding import cycle with existing main.go logger)
		_ = start // placeholder if switched to structured logging later
	})
}

func withRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				httpx.JSONError(w, http.StatusInternalServerError, "internal_error", nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
