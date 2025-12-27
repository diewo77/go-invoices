package policy

import (
	"context"

	"github.com/diewo77/go-gate"
	"github.com/diewo77/go-invoices/internal/models"
	"gorm.io/gorm"
)

// DBProfileResolver fetches user profiles from the database.
// It implements the gate.ProfileResolver interface for uint user IDs.
type DBProfileResolver struct {
	DB *gorm.DB
}

// NewDBProfileResolver creates a new database-backed profile resolver.
func NewDBProfileResolver(db *gorm.DB) *DBProfileResolver {
	return &DBProfileResolver{DB: db}
}

// Resolve looks up the user's profile from the database, preloading permissions.
// Returns nil if user has no profile assigned or user not found.
func (r *DBProfileResolver) Resolve(ctx context.Context, userID uint) (gate.Profile, error) {
	var user models.User
	err := r.DB.WithContext(ctx).Preload("Profile.Permissions").First(&user, userID).Error
	if err != nil {
		return nil, err
	}
	if user.Profile == nil {
		return nil, nil // User has no profile assigned
	}
	return &dbProfileAdapter{profile: user.Profile}, nil
}

// dbProfileAdapter wraps a models.Profile to implement gate.Profile interface.
type dbProfileAdapter struct {
	profile *models.Profile
}

// ID returns the profile's unique identifier.
func (a *dbProfileAdapter) ID() uint {
	return a.profile.ID
}

// Name returns the profile's display name.
func (a *dbProfileAdapter) Name() string {
	return a.profile.Name
}

// HasPermission checks if the profile has the requested permission.
// Supports wildcards: "*:*" (superadmin) and "resource:*" (all actions on resource).
func (a *dbProfileAdapter) HasPermission(perm gate.Permission) bool {
	for _, p := range a.profile.Permissions {
		dbPerm := gate.NewPermission(p.ResourceType, gate.Action(p.Action))
		// Check if db permission matches or is a wildcard that covers the requested
		if dbPerm.Matches(perm) {
			return true
		}
	}
	return false
}

// Permissions returns all permissions as gate.Permission slice.
func (a *dbProfileAdapter) Permissions() []gate.Permission {
	result := make([]gate.Permission, len(a.profile.Permissions))
	for i, p := range a.profile.Permissions {
		result[i] = gate.NewPermission(p.ResourceType, gate.Action(p.Action))
	}
	return result
}
