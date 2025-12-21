package policy_test

import (
	"context"
	"testing"

	"github.com/diewo77/go-gate"
	"github.com/diewo77/go-invoices/internal/policy"
)

// mockOwnable is a test resource that implements Ownable.
type mockOwnable struct {
	userID uint
}

func (m *mockOwnable) GetUserID() uint {
	return m.userID
}

// mockNonOwnable is a test resource that does NOT implement Ownable.
type mockNonOwnable struct {
	ID uint
}

func TestOwnershipPolicy_NilResource(t *testing.T) {
	p := policy.NewOwnershipPolicy()
	ctx := context.Background()

	// For nil resource (list/create), should return true
	if !p.Can(ctx, 1, gate.ActionList, nil) {
		t.Error("Expected Can to return true for nil resource")
	}
	if !p.Can(ctx, 1, gate.ActionCreate, nil) {
		t.Error("Expected Can to return true for nil resource on create")
	}
}

func TestOwnershipPolicy_OwnerCanAccess(t *testing.T) {
	p := policy.NewOwnershipPolicy()
	ctx := context.Background()
	resource := &mockOwnable{userID: 42}

	// Owner (userID 42) should have access
	if !p.Can(ctx, 42, gate.ActionView, resource) {
		t.Error("Expected owner to have access")
	}
	if !p.Can(ctx, 42, gate.ActionUpdate, resource) {
		t.Error("Expected owner to have access for update")
	}
	if !p.Can(ctx, 42, gate.ActionDelete, resource) {
		t.Error("Expected owner to have access for delete")
	}
}

func TestOwnershipPolicy_NonOwnerDenied(t *testing.T) {
	p := policy.NewOwnershipPolicy()
	ctx := context.Background()
	resource := &mockOwnable{userID: 42}

	// Non-owner (userID 99) should be denied
	if p.Can(ctx, 99, gate.ActionView, resource) {
		t.Error("Expected non-owner to be denied")
	}
	if p.Can(ctx, 99, gate.ActionUpdate, resource) {
		t.Error("Expected non-owner to be denied for update")
	}
	if p.Can(ctx, 99, gate.ActionDelete, resource) {
		t.Error("Expected non-owner to be denied for delete")
	}
}

func TestOwnershipPolicy_NonOwnableResource(t *testing.T) {
	p := policy.NewOwnershipPolicy()
	ctx := context.Background()
	resource := &mockNonOwnable{ID: 1}

	// Resource that doesn't implement Ownable should be denied
	if p.Can(ctx, 1, gate.ActionView, resource) {
		t.Error("Expected non-Ownable resource to be denied")
	}
}

func TestAdminBypassPolicy_AdminAllowed(t *testing.T) {
	inner := policy.NewOwnershipPolicy()
	isAdmin := func(_ context.Context, userID uint) bool {
		return userID == 1 // User 1 is admin
	}
	p := policy.NewAdminBypassPolicy(inner, isAdmin)
	ctx := context.Background()
	resource := &mockOwnable{userID: 42}

	// Admin (userID 1) should bypass ownership
	if !p.Can(ctx, 1, gate.ActionView, resource) {
		t.Error("Expected admin to bypass ownership check")
	}
	if !p.Can(ctx, 1, gate.ActionDelete, resource) {
		t.Error("Expected admin to bypass ownership for delete")
	}
}

func TestAdminBypassPolicy_NonAdminChecksOwnership(t *testing.T) {
	inner := policy.NewOwnershipPolicy()
	isAdmin := func(_ context.Context, userID uint) bool {
		return userID == 1 // Only user 1 is admin
	}
	p := policy.NewAdminBypassPolicy(inner, isAdmin)
	ctx := context.Background()
	resource := &mockOwnable{userID: 42}

	// Non-admin owner should have access
	if !p.Can(ctx, 42, gate.ActionView, resource) {
		t.Error("Expected owner to have access")
	}

	// Non-admin non-owner should be denied
	if p.Can(ctx, 99, gate.ActionView, resource) {
		t.Error("Expected non-owner non-admin to be denied")
	}
}
