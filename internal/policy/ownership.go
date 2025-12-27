package policy

import (
	"context"

	"github.com/diewo77/go-gate"
)

// Ownable is an interface for resources that have an owner.
// Implement this on your models to enable ownership-based authorization.
type Ownable interface {
	GetUserID() uint
}

// OwnershipPolicy is a generic policy that checks if the user owns the resource.
// Works with any model that implements the Ownable interface.
type OwnershipPolicy struct{}

// NewOwnershipPolicy creates a new ownership policy.
func NewOwnershipPolicy() *OwnershipPolicy {
	return &OwnershipPolicy{}
}

// Can checks if the user owns the resource.
// For list/create actions (resource is nil), it always returns true
// since profile permissions already control access.
func (p *OwnershipPolicy) Can(_ context.Context, userID uint, action gate.Action, resource any) bool {
	// For list/create, there's no specific resource to check ownership
	if resource == nil {
		return true
	}

	// Check if resource implements Ownable
	ownable, ok := resource.(Ownable)
	if !ok {
		// If resource doesn't implement Ownable, deny by default
		// This prevents accidental access to resources without ownership checks
		return false
	}

	// User owns the resource if their ID matches
	return ownable.GetUserID() == userID
}

// AdminBypassPolicy wraps another policy and always allows access for admins.
// This is useful when you want admins to access any resource regardless of ownership.
type AdminBypassPolicy struct {
	inner       gate.Policy[uint]
	isAdminFunc func(ctx context.Context, userID uint) bool
}

// NewAdminBypassPolicy creates a policy that bypasses ownership for admins.
func NewAdminBypassPolicy(inner gate.Policy[uint], isAdminFunc func(ctx context.Context, userID uint) bool) *AdminBypassPolicy {
	return &AdminBypassPolicy{
		inner:       inner,
		isAdminFunc: isAdminFunc,
	}
}

// Can checks if user is admin (bypass) or falls back to inner policy.
func (p *AdminBypassPolicy) Can(ctx context.Context, userID uint, action gate.Action, resource any) bool {
	// Admins can access everything
	if p.isAdminFunc(ctx, userID) {
		return true
	}
	// Otherwise, check inner policy (ownership)
	return p.inner.Can(ctx, userID, action, resource)
}
