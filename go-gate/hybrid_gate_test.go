package gate_test

import (
	"context"
	"testing"

	"github.com/diewo77/go-gate"
)

// mockOwnerPolicy checks if resource.OwnerID == userID
type mockOwnerPolicy struct{}

type mockResource struct {
	OwnerID uint
}

func (p *mockOwnerPolicy) Can(_ context.Context, userID uint, _ gate.Action, resource any) bool {
	if r, ok := resource.(*mockResource); ok {
		return r.OwnerID == userID
	}
	return false
}

func TestHybridGate_ProfileOnly(t *testing.T) {
	resolver := gate.NewStaticResolver[uint]()
	profile := gate.NewStaticProfile(1, "editor",
		gate.NewPermission("product", gate.ActionCreate),
		gate.NewPermission("product", gate.ActionView),
	)
	resolver.Set(1, profile)

	g := gate.NewHybridGate[uint](resolver)

	// User 1 can create product (profile permission, no resource)
	if !g.Can(context.Background(), 1, gate.ActionCreate, "product", nil) {
		t.Error("user with permission should be allowed")
	}

	// User 1 cannot delete product (no permission)
	if g.Can(context.Background(), 1, gate.ActionDelete, "product", nil) {
		t.Error("user without permission should be denied")
	}

	// User 2 has no profile
	if g.Can(context.Background(), 2, gate.ActionView, "product", nil) {
		t.Error("user without profile should be denied")
	}

	// Zero user is denied
	if g.Can(context.Background(), 0, gate.ActionView, "product", nil) {
		t.Error("zero user should be denied")
	}
}

func TestHybridGate_WithOwnershipPolicy(t *testing.T) {
	resolver := gate.NewStaticResolver[uint]()
	profile := gate.NewStaticProfile(1, "editor",
		gate.NewPermission("product", gate.ActionView),
		gate.NewPermission("product", gate.ActionUpdate),
	)
	resolver.Set(1, profile)
	resolver.Set(2, profile) // User 2 has same profile

	g := gate.NewHybridGate[uint](resolver)
	g.Register("product", &mockOwnerPolicy{})

	resource := &mockResource{OwnerID: 1}

	// User 1 owns the resource - allowed
	if !g.Can(context.Background(), 1, gate.ActionUpdate, "product", resource) {
		t.Error("owner should be allowed")
	}

	// User 2 has permission but doesn't own - denied
	if g.Can(context.Background(), 2, gate.ActionUpdate, "product", resource) {
		t.Error("non-owner should be denied even with profile permission")
	}
}

func TestHybridGate_CanProfile(t *testing.T) {
	resolver := gate.NewStaticResolver[uint]()
	profile := gate.NewStaticProfile(1, "editor",
		gate.NewPermission("product", gate.ActionView),
	)
	resolver.Set(1, profile)

	g := gate.NewHybridGate[uint](resolver)
	g.Register("product", &mockOwnerPolicy{})

	// CanProfile ignores ownership - just checks profile
	if !g.CanProfile(context.Background(), 1, gate.ActionView, "product") {
		t.Error("CanProfile should return true for user with permission")
	}

	if g.CanProfile(context.Background(), 1, gate.ActionDelete, "product") {
		t.Error("CanProfile should return false for missing permission")
	}
}
