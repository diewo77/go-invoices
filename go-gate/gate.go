// Package gate provides a Laravel-inspired Gate/Policy authorization system.
// The Gate is a central registry of policies; each Policy defines authorization
// rules for a specific resource type. This package has no dependencies on
// domain models and can be reused across different web applications.
//
// The package uses generics to allow any user/subject type:
//   - Gate[uint] for simple user ID based auth
//   - Gate[*User] for full user struct based auth
//   - Gate[*Claims] for JWT claims based auth
package gate

import "context"

// Gate is the central authorization checkpoint.
// U is the user/subject type (must be comparable for zero-value check).
// Register policies by resource type name, then call Authorize or Can.
type Gate[U comparable] struct {
	policies map[string]Policy[U]
}

// NewGate creates an empty Gate ready to register policies.
// Example: gate.NewGate[uint]() for userID-based authorization.
func NewGate[U comparable]() *Gate[U] {
	return &Gate[U]{policies: make(map[string]Policy[U])}
}

// Register adds a policy for a given resource type (e.g., "invoice").
// Overwrites any existing policy for that type.
func (g *Gate[U]) Register(resourceType string, p Policy[U]) {
	g.policies[resourceType] = p
}

// Authorize checks authorization and returns an error if denied.
// Returns ErrUnauthorized for zero-value user or denied action;
// returns ErrNoPolicyDefined if resourceType has no registered policy.
func (g *Gate[U]) Authorize(ctx context.Context, user U, action Action, resourceType string, resource any) error {
	var zero U
	if user == zero {
		return ErrUnauthorized
	}
	p, ok := g.policies[resourceType]
	if !ok {
		return ErrNoPolicyDefined
	}
	if !p.Can(ctx, user, action, resource) {
		return ErrUnauthorized
	}
	return nil
}

// Can is a convenience wrapper returning bool instead of error.
// Returns true only if Authorize returns nil.
func (g *Gate[U]) Can(ctx context.Context, user U, action Action, resourceType string, resource any) bool {
	return g.Authorize(ctx, user, action, resourceType, resource) == nil
}
