# Plan: Extract Gate Package + Internal Policies

**TL;DR:** Create a standalone reusable `gate/` package (like `pdf/`) containing the core Gate/Policy framework with zero dependencies. Keep concrete policies in `internal/policy/` importing from the new `gate` package. Re-export gate types in policy package so handlers only need one import.

---

## Steps

### 1. Create `gate/go.mod`

```go
module github.com/diewo77/billing-app/gate

go 1.25.5
```

---

### 2. Create `gate/gate.go`

```go
// Package gate provides a Laravel-inspired Gate/Policy authorization system.
// The Gate is a central registry of policies; each Policy defines authorization
// rules for a specific resource type. This package has no dependencies on
// domain models and can be reused across different web applications.
package gate

import (
	"context"
	"errors"
)

// Sentinel errors returned by Gate.Authorize.
var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrNoPolicyDefined = errors.New("no policy defined for resource")
)

// Action describes the kind of operation a user wants to perform.
type Action string

const (
	ActionView   Action = "view"
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionList   Action = "list"
)

// Policy defines authorization rules for a resource type.
// Implementations check whether userID may perform action on resource.
type Policy interface {
	// Can returns true if userID is authorized to perform action on resource.
	// For list/create, resource may be nil (context-only check).
	Can(ctx context.Context, userID uint, action Action, resource any) bool
}

// Gate is the central authorization checkpoint.
// Register policies by resource type name, then call Authorize or Can.
type Gate struct {
	policies map[string]Policy
}

// NewGate creates an empty Gate ready to register policies.
func NewGate() *Gate {
	return &Gate{policies: make(map[string]Policy)}
}

// Register adds a policy for a given resource type (e.g., "invoice").
// Overwrites any existing policy for that type.
func (g *Gate) Register(resourceType string, p Policy) {
	g.policies[resourceType] = p
}

// Authorize checks authorization and returns an error if denied.
// Returns ErrUnauthorized for missing user or denied action;
// returns ErrNoPolicyDefined if resourceType has no registered policy.
func (g *Gate) Authorize(ctx context.Context, userID uint, action Action, resourceType string, resource any) error {
	if userID == 0 {
		return ErrUnauthorized
	}
	p, ok := g.policies[resourceType]
	if !ok {
		return ErrNoPolicyDefined
	}
	if !p.Can(ctx, userID, action, resource) {
		return ErrUnauthorized
	}
	return nil
}

// Can is a convenience wrapper returning bool instead of error.
// Returns true only if Authorize returns nil.
func (g *Gate) Can(ctx context.Context, userID uint, action Action, resourceType string, resource any) bool {
	return g.Authorize(ctx, userID, action, resourceType, resource) == nil
}
```

---

### 3. Create `gate/gate_test.go`

```go
package gate_test

import (
	"context"
	"testing"

	"github.com/diewo77/billing-app/gate"
)

// mockPolicy is a simple policy for testing.
type mockPolicy struct {
	allowAll bool
}

func (p *mockPolicy) Can(_ context.Context, _ uint, _ gate.Action, _ any) bool {
	return p.allowAll
}

func TestNewGate(t *testing.T) {
	g := gate.NewGate()
	if g == nil {
		t.Fatal("expected non-nil Gate")
	}
}

func TestGate_Register(t *testing.T) {
	g := gate.NewGate()
	g.Register("test", &mockPolicy{allowAll: true})
	// No error means success
}

func TestGate_Authorize_NoUser(t *testing.T) {
	g := gate.NewGate()
	g.Register("test", &mockPolicy{allowAll: true})

	err := g.Authorize(context.Background(), 0, gate.ActionView, "test", nil)
	if err != gate.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestGate_Authorize_NoPolicy(t *testing.T) {
	g := gate.NewGate()

	err := g.Authorize(context.Background(), 1, gate.ActionView, "unknown", nil)
	if err != gate.ErrNoPolicyDefined {
		t.Errorf("expected ErrNoPolicyDefined, got %v", err)
	}
}

func TestGate_Authorize_Allowed(t *testing.T) {
	g := gate.NewGate()
	g.Register("test", &mockPolicy{allowAll: true})

	err := g.Authorize(context.Background(), 1, gate.ActionView, "test", nil)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestGate_Authorize_Denied(t *testing.T) {
	g := gate.NewGate()
	g.Register("test", &mockPolicy{allowAll: false})

	err := g.Authorize(context.Background(), 1, gate.ActionView, "test", nil)
	if err != gate.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestGate_Can(t *testing.T) {
	g := gate.NewGate()
	g.Register("test", &mockPolicy{allowAll: true})

	if !g.Can(context.Background(), 1, gate.ActionCreate, "test", nil) {
		t.Error("expected Can to return true")
	}

	g.Register("denied", &mockPolicy{allowAll: false})
	if g.Can(context.Background(), 1, gate.ActionCreate, "denied", nil) {
		t.Error("expected Can to return false")
	}
}
```

---

### 4. Update `billing-app/go.mod`

Add to `require` block:

```go
github.com/diewo77/billing-app/gate v0.0.0
```

Add to `replace` directives:

```go
replace github.com/diewo77/billing-app/gate => ../gate
```

---

### 5. Replace `internal/policy/policy.go` with re-exports

```go
// Package policy provides authorization policies for the billing app.
// It re-exports the core gate package types for convenience, so handlers
// only need to import this one package.
package policy

import (
	"github.com/diewo77/billing-app/gate"
)

// Re-export core types from gate package.
// This allows handlers to use policy.Gate, policy.Action, etc.
// without importing the gate package directly.
type (
	Gate   = gate.Gate
	Policy = gate.Policy
	Action = gate.Action
)

// Re-export Action constants.
const (
	ActionView   = gate.ActionView
	ActionCreate = gate.ActionCreate
	ActionUpdate = gate.ActionUpdate
	ActionDelete = gate.ActionDelete
	ActionList   = gate.ActionList
)

// Re-export sentinel errors.
var (
	ErrUnauthorized    = gate.ErrUnauthorized
	ErrNoPolicyDefined = gate.ErrNoPolicyDefined
)

// NewGate re-exports gate.NewGate for convenience.
func NewGate() *Gate {
	return gate.NewGate()
}
```

---

### 6. Update concrete policies

No changes needed to `product_policy.go`, `invoice_policy.go`, `client_policy.go`, `product_type_policy.go`, `unit_type_policy.go`, `role_checker.go` - they already use `policy.Action`, `policy.Policy`, etc. which will now resolve to the re-exported gate types.

---

### 7. Update `policy_test.go`

Remove the Gate/Policy tests that are now in `gate/gate_test.go`. Keep only tests that depend on models (ProductPolicy, InvoicePolicy, etc.).

---

## Files Summary

| File                             | Action                          |
| -------------------------------- | ------------------------------- |
| `gate/go.mod`                    | Create                          |
| `gate/gate.go`                   | Create                          |
| `gate/gate_test.go`              | Create                          |
| `billing-app/go.mod`             | Modify - add gate dependency    |
| `internal/policy/policy.go`      | Replace - re-exports only       |
| `internal/policy/policy_test.go` | Modify - remove core Gate tests |

---

## Benefits

1. **Reusable:** `gate` package has zero dependencies, can be used in any Go web app
2. **Single import:** Handlers only need `internal/policy`, re-exports provide full functionality
3. **Separation of concerns:** Core framework in `gate/`, app-specific policies in `internal/policy/`
4. **Consistent with codebase:** Same pattern as `pdf/` package
