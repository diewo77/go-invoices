package gate

import "context"

// HybridGate combines profile-based global permissions with resource-specific policies.
// Authorization flow:
//  1. Check if user is valid (non-zero)
//  2. Check if user's profile has the required permission (resource:action)
//  3. If a resource policy exists and resource is provided, check ownership
type HybridGate[U comparable] struct {
	resolver ProfileResolver[U]
	policies map[string]Policy[U]
}

// NewHybridGate creates a hybrid gate with the given profile resolver.
func NewHybridGate[U comparable](resolver ProfileResolver[U]) *HybridGate[U] {
	return &HybridGate[U]{
		resolver: resolver,
		policies: make(map[string]Policy[U]),
	}
}

// Register adds a resource-specific policy for ownership checks.
func (g *HybridGate[U]) Register(resourceType string, p Policy[U]) {
	g.policies[resourceType] = p
}

// Authorize checks:
//  1. User is valid (non-zero)
//  2. User's profile has permission for resource:action
//  3. If a resource policy exists and resource is provided, checks ownership
func (g *HybridGate[U]) Authorize(ctx context.Context, user U, action Action, resourceType string, resource any) error {
	var zero U
	if user == zero {
		return ErrUnauthorized
	}

	// Step 1: Check profile permission
	profile, err := g.resolver.Resolve(ctx, user)
	if err != nil || profile == nil {
		return ErrUnauthorized
	}

	perm := NewPermission(resourceType, action)
	if !profile.HasPermission(perm) {
		return ErrUnauthorized
	}

	// Step 2: Check resource policy (ownership) if exists and resource is provided
	if resource != nil {
		if policy, ok := g.policies[resourceType]; ok {
			if !policy.Can(ctx, user, action, resource) {
				return ErrUnauthorized
			}
		}
	}

	return nil
}

// Can is a convenience wrapper returning bool instead of error.
func (g *HybridGate[U]) Can(ctx context.Context, user U, action Action, resourceType string, resource any) bool {
	return g.Authorize(ctx, user, action, resourceType, resource) == nil
}

// CanProfile checks only the profile permission, without ownership check.
// Useful for UI to show/hide buttons before a specific resource is loaded.
func (g *HybridGate[U]) CanProfile(ctx context.Context, user U, action Action, resourceType string) bool {
	var zero U
	if user == zero {
		return false
	}
	profile, err := g.resolver.Resolve(ctx, user)
	if err != nil || profile == nil {
		return false
	}
	return profile.HasPermission(NewPermission(resourceType, action))
}
