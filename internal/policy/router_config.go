package policy

import (
	"time"

	"github.com/diewo77/go-invoices/internal/handlers"
	"github.com/diewo77/go-invoices/internal/services"
	"gorm.io/gorm"
)

// RouterConfig holds configured handlers and middleware for the application.
// Use this as a reference for how to set up authorization in your router.
type RouterConfig struct {
	// AuthGate provides authorization checks and middleware
	AuthGate *AuthGate

	// Admin handlers
	AdminProfileHandler     *handlers.AdminProfileHandler
	AdminUserProfileHandler *handlers.AdminUserProfileHandler

	// Auth handler
	AuthHandler *handlers.AuthHandler

	// Business handlers
	ClientHandler  *handlers.ClientHandler
	ProductHandler *handlers.ProductHandler
	InvoiceHandler *handlers.InvoiceHandler
	CompanyHandler *handlers.CompanyHandler

	// Services
	InvoiceService *services.InvoiceService
}

// NewRouterConfig creates a fully configured router setup.
// This wires together the authorization gate, policies, and admin handlers.
//
// Example usage in your main.go or router setup:
//
//	cfg := policy.NewRouterConfig(db)
//
//	// Protected routes with ownership check
//	mux.Handle("GET /products", cfg.AuthGate.RequirePermission("product", gate.ActionList)(productHandler.List))
//	mux.Handle("GET /products/{id}", cfg.AuthGate.RequirePermission("product", gate.ActionView)(productHandler.Get))
//
//	// Admin-only routes
//	mux.Handle("GET /admin/profiles", cfg.AuthGate.RequireAdmin()(http.HandlerFunc(cfg.AdminProfileHandler.List)))
//	mux.Handle("POST /admin/profiles/create", cfg.AuthGate.RequireAdmin()(http.HandlerFunc(cfg.AdminProfileHandler.Create)))
func NewRouterConfig(db *gorm.DB) *RouterConfig {
	// Create authorization gate with 5-minute cache
	authGate := NewAuthGate(db, 5*time.Minute)

	// Register ownership policies for each resource type
	// These check if the user owns the specific resource they're trying to access
	ownershipPolicy := NewOwnershipPolicy()
	authGate.RegisterPolicy("product", ownershipPolicy)
	authGate.RegisterPolicy("invoice", ownershipPolicy)
	authGate.RegisterPolicy("client", ownershipPolicy)
	authGate.RegisterPolicy("company_settings", ownershipPolicy)

	// Create admin handlers with cache invalidation support
	adminProfileHandler := handlers.NewAdminProfileHandler(db, authGate.CacheResolver)
	adminUserProfileHandler := handlers.NewAdminUserProfileHandler(db, authGate.CacheResolver)

	// Create auth handler
	authHandler := handlers.NewAuthHandler(db)

	// Create business handlers
	clientHandler := handlers.NewClientHandler(db)
	productHandler := handlers.NewProductHandler(db)
	invoiceHandler := handlers.NewInvoiceHandler(db)
	companyHandler := handlers.NewCompanyHandler(db)

	// Create services
	invoiceService := services.NewInvoiceService(db)

	return &RouterConfig{
		AuthGate:                authGate,
		AdminProfileHandler:     adminProfileHandler,
		AdminUserProfileHandler: adminUserProfileHandler,
		AuthHandler:             authHandler,
		ClientHandler:           clientHandler,
		ProductHandler:          productHandler,
		InvoiceHandler:          invoiceHandler,
		CompanyHandler:          companyHandler,
		InvoiceService:          invoiceService,
	}
}

/*
Example Routes Setup (add to your router.go):

// Public routes (no auth required)
mux.HandleFunc("GET /", landingHandler)
mux.HandleFunc("GET /login", loginHandler)
mux.HandleFunc("POST /login", loginHandler)

// Authenticated routes (require logged-in user)
// Use your existing auth middleware first, then permission checks

// Products - require product permissions
mux.Handle("GET /products",
	authMiddleware(
		cfg.AuthGate.RequirePermission("product", gate.ActionList)(
			http.HandlerFunc(productHandler.List))))

mux.Handle("POST /products",
	authMiddleware(
		cfg.AuthGate.RequirePermission("product", gate.ActionCreate)(
			http.HandlerFunc(productHandler.Create))))

// For update/delete, also check ownership in the handler:
//
//	func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
//	    product := h.getProduct(id)
//	    if err := cfg.AuthGate.Authorize(r.Context(), gate.ActionUpdate, "product", product); err != nil {
//	        http.Error(w, "Forbidden", 403)
//	        return
//	    }
//	    // proceed with update...
//	}

// Admin routes - require superadmin permission
mux.Handle("GET /admin/profiles",
	authMiddleware(
		cfg.AuthGate.RequireAdmin()(
			http.HandlerFunc(cfg.AdminProfileHandler.List))))

mux.Handle("POST /admin/profiles/create",
	authMiddleware(
		cfg.AuthGate.RequireAdmin()(
			http.HandlerFunc(cfg.AdminProfileHandler.Create))))

mux.Handle("GET /admin/profiles/edit",
	authMiddleware(
		cfg.AuthGate.RequireAdmin()(
			http.HandlerFunc(cfg.AdminProfileHandler.Edit))))

mux.Handle("POST /admin/profiles/update",
	authMiddleware(
		cfg.AuthGate.RequireAdmin()(
			http.HandlerFunc(cfg.AdminProfileHandler.Update))))

mux.Handle("POST /admin/profiles/delete",
	authMiddleware(
		cfg.AuthGate.RequireAdmin()(
			http.HandlerFunc(cfg.AdminProfileHandler.Delete))))

mux.Handle("GET /admin/profiles/permissions",
	authMiddleware(
		cfg.AuthGate.RequireAdmin()(
			http.HandlerFunc(cfg.AdminProfileHandler.EditPermissions))))

mux.Handle("POST /admin/profiles/permissions/save",
	authMiddleware(
		cfg.AuthGate.RequireAdmin()(
			http.HandlerFunc(cfg.AdminProfileHandler.SavePermissions))))

mux.Handle("GET /admin/users",
	authMiddleware(
		cfg.AuthGate.RequireAdmin()(
			http.HandlerFunc(cfg.AdminUserProfileHandler.List))))

mux.Handle("POST /admin/users/assign-profile",
	authMiddleware(
		cfg.AuthGate.RequireAdmin()(
			http.HandlerFunc(cfg.AdminUserProfileHandler.AssignProfile))))
*/
