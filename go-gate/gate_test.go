package gate_test

import (
	"context"
	"testing"

	"github.com/diewo77/go-gate"
)

// mockPolicy is a simple policy for testing with uint user type.
type mockPolicy struct {
	allowAll bool
}

func (p *mockPolicy) Can(_ context.Context, _ uint, _ gate.Action, _ any) bool {
	return p.allowAll
}

func TestNewGate(t *testing.T) {
	g := gate.NewGate[uint]()
	if g == nil {
		t.Fatal("expected non-nil Gate")
	}
}

func TestGate_Register(t *testing.T) {
	g := gate.NewGate[uint]()
	g.Register("test", &mockPolicy{allowAll: true})
	// No error means success
}

func TestGate_Authorize_NoUser(t *testing.T) {
	g := gate.NewGate[uint]()
	g.Register("test", &mockPolicy{allowAll: true})

	err := g.Authorize(context.Background(), 0, gate.ActionView, "test", nil)
	if err != gate.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestGate_Authorize_NoPolicy(t *testing.T) {
	g := gate.NewGate[uint]()

	err := g.Authorize(context.Background(), 1, gate.ActionView, "unknown", nil)
	if err != gate.ErrNoPolicyDefined {
		t.Errorf("expected ErrNoPolicyDefined, got %v", err)
	}
}

func TestGate_Authorize_Allowed(t *testing.T) {
	g := gate.NewGate[uint]()
	g.Register("test", &mockPolicy{allowAll: true})

	err := g.Authorize(context.Background(), 1, gate.ActionView, "test", nil)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestGate_Authorize_Denied(t *testing.T) {
	g := gate.NewGate[uint]()
	g.Register("test", &mockPolicy{allowAll: false})

	err := g.Authorize(context.Background(), 1, gate.ActionView, "test", nil)
	if err != gate.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestGate_Can(t *testing.T) {
	g := gate.NewGate[uint]()
	g.Register("test", &mockPolicy{allowAll: true})

	if !g.Can(context.Background(), 1, gate.ActionCreate, "test", nil) {
		t.Error("expected Can to return true")
	}

	g.Register("denied", &mockPolicy{allowAll: false})
	if g.Can(context.Background(), 1, gate.ActionCreate, "denied", nil) {
		t.Error("expected Can to return false")
	}
}

// Test with a custom user type to verify generics work
type testUser struct {
	ID   int
	Role string
}

type userPolicy struct{}

func (p *userPolicy) Can(_ context.Context, user *testUser, action gate.Action, _ any) bool {
	if user == nil {
		return false
	}
	// Admin can do anything
	if user.Role == "admin" {
		return true
	}
	// Regular users can only view
	return action == gate.ActionView
}

func TestGate_WithCustomUserType(t *testing.T) {
	g := gate.NewGate[*testUser]()
	g.Register("resource", &userPolicy{})

	admin := &testUser{ID: 1, Role: "admin"}
	regular := &testUser{ID: 2, Role: "user"}

	// Admin can create
	if !g.Can(context.Background(), admin, gate.ActionCreate, "resource", nil) {
		t.Error("admin should be able to create")
	}

	// Regular user cannot create
	if g.Can(context.Background(), regular, gate.ActionCreate, "resource", nil) {
		t.Error("regular user should not be able to create")
	}

	// Regular user can view
	if !g.Can(context.Background(), regular, gate.ActionView, "resource", nil) {
		t.Error("regular user should be able to view")
	}

	// Nil user is unauthorized
	err := g.Authorize(context.Background(), nil, gate.ActionView, "resource", nil)
	if err != gate.ErrUnauthorized {
		t.Errorf("nil user should be unauthorized, got %v", err)
	}
}
