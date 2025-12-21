package policy

import (
	"context"
	"net/http"
	"time"

	"github.com/diewo77/go-invoices/auth"
	"github.com/diewo77/go-gate"
	"gorm.io/gorm"
)

// AuthGate holds the configured HybridGate with caching.
// Use this as a central authorization point in your application.
type AuthGate struct {
	Gate          *gate.HybridGate[uint]
	CacheResolver *gate.CachedResolver[uint]
}

// NewAuthGate creates a fully configured authorization gate.
// - db: GORM database connection for profile lookups
// - cacheTTL: how long to cache user profiles (e.g., 5*time.Minute)
func NewAuthGate(db *gorm.DB, cacheTTL time.Duration) *AuthGate {
	// Create DB resolver that fetches profiles from database
	dbResolver := NewDBProfileResolver(db)

	// Wrap with caching to avoid DB queries on every request
	cachedResolver := gate.NewCachedResolver[uint](dbResolver, cacheTTL)

	// Create hybrid gate that combines profile permissions with ownership policies
	hybridGate := gate.NewHybridGate[uint](cachedResolver)

	return &AuthGate{
		Gate:          hybridGate,
		CacheResolver: cachedResolver,
	}
}

// RegisterPolicy adds an ownership policy for a resource type.
// Example: authGate.RegisterPolicy("product", policy.NewOwnershipPolicy())
func (ag *AuthGate) RegisterPolicy(resourceType string, p gate.Policy[uint]) {
	ag.Gate.Register(resourceType, p)
}

// Authorize checks if the current user can perform an action on a resource.
// Returns nil if authorized, gate.ErrUnauthorized otherwise.
func (ag *AuthGate) Authorize(ctx context.Context, action gate.Action, resourceType string, resource any) error {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return gate.ErrUnauthorized
	}
	return ag.Gate.Authorize(ctx, userID, action, resourceType, resource)
}

// Can is a convenience method that returns bool instead of error.
func (ag *AuthGate) Can(ctx context.Context, action gate.Action, resourceType string, resource any) bool {
	return ag.Authorize(ctx, action, resourceType, resource) == nil
}

// CanProfile checks only profile permissions (no ownership check).
// Useful for UI to show/hide buttons before a specific resource is loaded.
func (ag *AuthGate) CanProfile(ctx context.Context, action gate.Action, resourceType string) bool {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return false
	}
	return ag.Gate.CanProfile(ctx, userID, action, resourceType)
}

// InvalidateUser clears the cache for a specific user.
// Call this when a user's profile is changed.
func (ag *AuthGate) InvalidateUser(userID uint) {
	ag.CacheResolver.Invalidate(userID)
}

// InvalidateAll clears the entire profile cache.
// Call this when profile permissions are modified.
func (ag *AuthGate) InvalidateAll() {
	ag.CacheResolver.InvalidateAll()
}

// RequirePermission returns middleware that checks profile permission.
// Blocks access if user doesn't have the required permission.
func (ag *AuthGate) RequirePermission(resourceType string, action gate.Action) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !ag.CanProfile(r.Context(), action, resourceType) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin returns middleware that only allows users with admin profile.
// Uses the "*:*" superadmin permission check.
func (ag *AuthGate) RequireAdmin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := auth.UserIDFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check for superadmin permission
			profile, err := ag.CacheResolver.Resolve(r.Context(), userID)
			if err != nil || profile == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Check if profile has superadmin permission
			if !profile.HasPermission(gate.PermissionSuperAdmin) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
