package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/diewo77/go-invoices/auth"
	"github.com/diewo77/go-gate"
	"github.com/diewo77/go-invoices/i18n"
	"github.com/diewo77/go-invoices/internal/models"
	"github.com/diewo77/go-invoices/internal/policy"
	"github.com/diewo77/go-invoices/view"
	"gorm.io/gorm"
)

// App is the main application handler that sets up all routes.
type App struct {
	mux       *http.ServeMux
	db        *gorm.DB
	routerCfg *policy.RouterConfig
}

// NewApp creates a new application with all routes configured.
func NewApp(db *gorm.DB, routerCfg *policy.RouterConfig) *App {
	app := &App{
		mux:       http.NewServeMux(),
		db:        db,
		routerCfg: routerCfg,
	}
	// Expose minimal permission resolvers to the view layer so templates can show/hide UI based on permissions
	// Use resolver callbacks to avoid importing policy types into the view package.
	view.SetCanProfileResolver(func(r *http.Request, resource, action string) bool {
		// In dev mode, if DEV=1 we may allow permissive rendering for ease of testing.
		if os.Getenv("DEV") == "1" {
			return true
		}
		if routerCfg == nil || routerCfg.AuthGate == nil {
			return false
		}
		return routerCfg.AuthGate.CanProfile(r.Context(), gate.Action(action), resource)
	})
	view.SetIsAdminResolver(func(r *http.Request) bool {
		if os.Getenv("DEV") == "1" {
			return true
		}
		if routerCfg == nil || routerCfg.AuthGate == nil {
			return false
		}
		uid, ok := auth.UserIDFromContext(r.Context())
		if !ok {
			return false
		}
		prof, err := routerCfg.AuthGate.CacheResolver.Resolve(r.Context(), uid)
		if err != nil || prof == nil {
			return false
		}
		return prof.HasPermission(gate.PermissionSuperAdmin)
	})
	app.setupRoutes()
	return app
}

// ServeHTTP implements http.Handler.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Apply global middleware: auth context + preferences (language, theme)
	handler := auth.Middleware(withPreferences(a.mux))
	handler.ServeHTTP(w, r)
}

// setupRoutes configures all application routes.
func (a *App) setupRoutes() {
	// ─────────────────────────────────────────────────────────────────────────
	// Public routes (no auth required)
	// ─────────────────────────────────────────────────────────────────────────
	ah := a.routerCfg.AuthHandler

	a.mux.HandleFunc("GET /", a.landingPage)
	a.mux.HandleFunc("GET /login", ah.Login)
	a.mux.HandleFunc("POST /login", ah.Login)
	a.mux.HandleFunc("GET /signup", ah.Signup)
	a.mux.HandleFunc("POST /signup", ah.Signup)
	a.mux.HandleFunc("GET /logout", ah.Logout)
	a.mux.HandleFunc("POST /logout", ah.Logout)

	// ─────────────────────────────────────────────────────────────────────────
	// Authenticated routes (require logged-in user)
	// ─────────────────────────────────────────────────────────────────────────
	a.mux.Handle("GET /dashboard", a.requireAuth(http.HandlerFunc(a.dashboard)))

	// ─────────────────────────────────────────────────────────────────────────
	// Protected resource routes (require auth + specific permissions)
	// ─────────────────────────────────────────────────────────────────────────
	ph := a.routerCfg.ProductHandler
	ch := a.routerCfg.ClientHandler
	ih := a.routerCfg.InvoiceHandler

	// Products - require product:list, product:create, etc.
	a.mux.Handle("GET /products",
		a.requireAuth(a.requirePermission("product", gate.ActionList)(http.HandlerFunc(ph.List))))
	a.mux.Handle("GET /products/new",
		a.requireAuth(a.requirePermission("product", gate.ActionCreate)(http.HandlerFunc(ph.New))))
	a.mux.Handle("POST /products",
		a.requireAuth(a.requirePermission("product", gate.ActionCreate)(http.HandlerFunc(ph.Create))))
	a.mux.Handle("GET /products/{id}",
		a.requireAuth(a.requirePermission("product", gate.ActionView)(http.HandlerFunc(ph.View))))
	a.mux.Handle("GET /products/{id}/edit",
		a.requireAuth(a.requirePermission("product", gate.ActionUpdate)(http.HandlerFunc(ph.Edit))))
	a.mux.Handle("POST /products/{id}",
		a.requireAuth(a.requirePermission("product", gate.ActionUpdate)(http.HandlerFunc(ph.Update))))
	a.mux.Handle("POST /products/{id}/delete",
		a.requireAuth(a.requirePermission("product", gate.ActionDelete)(http.HandlerFunc(ph.Delete))))

	// Clients - require client:list, client:create, etc.
	a.mux.Handle("GET /clients",
		a.requireAuth(a.requirePermission("client", gate.ActionList)(http.HandlerFunc(ch.List))))
	a.mux.Handle("GET /clients/new",
		a.requireAuth(a.requirePermission("client", gate.ActionCreate)(http.HandlerFunc(ch.New))))
	a.mux.Handle("POST /clients",
		a.requireAuth(a.requirePermission("client", gate.ActionCreate)(http.HandlerFunc(ch.Create))))
	a.mux.Handle("GET /clients/{id}",
		a.requireAuth(a.requirePermission("client", gate.ActionView)(http.HandlerFunc(ch.View))))
	a.mux.Handle("GET /clients/{id}/edit",
		a.requireAuth(a.requirePermission("client", gate.ActionUpdate)(http.HandlerFunc(ch.Edit))))
	a.mux.Handle("POST /clients/{id}",
		a.requireAuth(a.requirePermission("client", gate.ActionUpdate)(http.HandlerFunc(ch.Update))))
	a.mux.Handle("POST /clients/{id}/delete",
		a.requireAuth(a.requirePermission("client", gate.ActionDelete)(http.HandlerFunc(ch.Delete))))

	// Invoices - require invoice:list, invoice:create, etc.
	a.mux.Handle("GET /invoices",
		a.requireAuth(a.requirePermission("invoice", gate.ActionList)(http.HandlerFunc(ih.List))))
	a.mux.Handle("GET /invoices/new",
		a.requireAuth(a.requirePermission("invoice", gate.ActionCreate)(http.HandlerFunc(ih.New))))
	a.mux.Handle("POST /invoices",
		a.requireAuth(a.requirePermission("invoice", gate.ActionCreate)(http.HandlerFunc(ih.Create))))
	a.mux.Handle("GET /invoices/{id}",
		a.requireAuth(a.requirePermission("invoice", gate.ActionView)(http.HandlerFunc(ih.View))))
	a.mux.Handle("GET /invoices/{id}/edit",
		a.requireAuth(a.requirePermission("invoice", gate.ActionUpdate)(http.HandlerFunc(ih.Edit))))
	a.mux.Handle("POST /invoices/{id}",
		a.requireAuth(a.requirePermission("invoice", gate.ActionUpdate)(http.HandlerFunc(ih.Update))))
	a.mux.Handle("POST /invoices/{id}/delete",
		a.requireAuth(a.requirePermission("invoice", gate.ActionDelete)(http.HandlerFunc(ih.Delete))))
	a.mux.Handle("POST /invoices/{id}/finalize",
		a.requireAuth(a.requirePermission("invoice", "finalize")(http.HandlerFunc(ih.Finalize))))
	a.mux.Handle("GET /invoices/{id}/pdf",
		a.requireAuth(a.requirePermission("invoice", gate.ActionView)(http.HandlerFunc(ih.PDF))))

	// Invoice Items
	a.mux.Handle("POST /invoices/{id}/items",
		a.requireAuth(a.requirePermission("invoice", gate.ActionUpdate)(http.HandlerFunc(ih.AddItem))))
	a.mux.Handle("POST /invoices/{id}/items/{item_id}/delete",
		a.requireAuth(a.requirePermission("invoice", gate.ActionUpdate)(http.HandlerFunc(ih.RemoveItem))))

	// Company Settings
	sh := a.routerCfg.CompanyHandler
	a.mux.Handle("GET /settings",
		a.requireAuth(http.HandlerFunc(sh.Edit)))
	a.mux.Handle("POST /settings",
		a.requireAuth(http.HandlerFunc(sh.Update)))
	a.mux.HandleFunc("GET /setup", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/settings", http.StatusMovedPermanently)
	})

	// ─────────────────────────────────────────────────────────────────────────
	// Admin routes (require admin profile with profile:* permission)
	// ─────────────────────────────────────────────────────────────────────────
	aph := a.routerCfg.AdminProfileHandler
	auph := a.routerCfg.AdminUserProfileHandler

	// Profile management
	a.mux.Handle("GET /admin/profiles",
		a.requireAdmin(http.HandlerFunc(aph.List)))
	a.mux.Handle("GET /admin/profiles/new",
		a.requireAdmin(http.HandlerFunc(aph.New)))
	a.mux.Handle("POST /admin/profiles/create",
		a.requireAdmin(http.HandlerFunc(aph.Create)))
	a.mux.Handle("GET /admin/profiles/{id}/edit",
		a.requireAdmin(http.HandlerFunc(aph.Edit)))
	a.mux.Handle("POST /admin/profiles/{id}/update",
		a.requireAdmin(http.HandlerFunc(aph.Update)))
	a.mux.Handle("POST /admin/profiles/{id}/delete",
		a.requireAdmin(http.HandlerFunc(aph.Delete)))
	a.mux.Handle("GET /admin/profiles/{id}/permissions",
		a.requireAdmin(http.HandlerFunc(aph.EditPermissions)))
	a.mux.Handle("POST /admin/profiles/{id}/permissions",
		a.requireAdmin(http.HandlerFunc(aph.SavePermissions)))

	// User profile assignment
	a.mux.Handle("GET /admin/users",
		a.requireAdmin(http.HandlerFunc(auph.List)))
	a.mux.Handle("POST /admin/users/{id}/profile",
		a.requireAdmin(http.HandlerFunc(auph.AssignProfile)))

	// ─────────────────────────────────────────────────────────────────────────
	// Static files
	// ─────────────────────────────────────────────────────────────────────────
	a.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}

// ─────────────────────────────────────────────────────────────────────────────
// Middleware
// ─────────────────────────────────────────────────────────────────────────────

// requireAuth wraps a handler to require authentication.
func (a *App) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse session and get user ID
		userID, ok := auth.UserIDFromContext(r.Context())
		if !ok || userID == 0 {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireAdmin wraps a handler to require admin permissions.
// Uses the AuthGate to check for profile:* or *:* permission.
func (a *App) requireAdmin(next http.Handler) http.Handler {
	return a.routerCfg.AuthGate.RequireAdmin()(next)
}

// requirePermission wraps a handler to require specific resource permission.
func (a *App) requirePermission(resourceType string, action gate.Action) func(http.Handler) http.Handler {
	return a.routerCfg.AuthGate.RequirePermission(resourceType, action)
}

// withPreferences injects language and theme preferences from cookies/query.
func withPreferences(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get language preference (from cookie or query)
		lang := "fr" // default
		if c, err := r.Cookie("lang"); err == nil && c.Value != "" {
			lang = c.Value
		}
		if q := r.URL.Query().Get("lang"); q != "" {
			lang = q
			http.SetCookie(w, &http.Cookie{
				Name:     "lang",
				Value:    lang,
				Path:     "/",
				MaxAge:   86400 * 365,
				HttpOnly: true,
			})
		}
		ctx = i18n.WithLang(ctx, lang)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Page handlers (stubs - implement as needed)
// ─────────────────────────────────────────────────────────────────────────────

func (a *App) landingPage(w http.ResponseWriter, r *http.Request) {
	userID, loggedIn := auth.UserIDFromContext(r.Context())
	data := map[string]any{
		"IsLoggedIn": loggedIn,
		"UserID":     userID,
	}
	if err := view.Render(w, r, "index.html", data); err != nil {
		http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) dashboard(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	// Get user with profile
	var user models.User
	a.db.Preload("Profile").First(&user, userID)

	// Get stats
	var productCount, clientCount, invoiceCount int64
	a.db.Model(&models.Product{}).Where("user_id = ?", userID).Count(&productCount)
	a.db.Model(&models.Client{}).Where("user_id = ?", userID).Count(&clientCount)
	a.db.Model(&models.Invoice{}).Where("user_id = ?", userID).Count(&invoiceCount)

	// Get recent products (last 5)
	var recentProducts []models.Product
	a.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(5).Find(&recentProducts)

	// Get recent invoices (last 5)
	var recentInvoices []models.Invoice
	a.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(5).Find(&recentInvoices)

	// Get revenue
	revenue, _ := a.routerCfg.InvoiceService.GetRevenue(userID)

	view.Render(w, r, "dashboard.html", map[string]any{
		"User": user,
		"Stats": map[string]any{
			"Products": productCount,
			"Clients":  clientCount,
			"Invoices": invoiceCount,
			"Revenue":  fmt.Sprintf("€%.2f", revenue),
		},
		"RecentProducts": recentProducts,
		"RecentInvoices": recentInvoices,
	})
}
