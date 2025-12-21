package gate_test

import (
	"context"
	"testing"

	"github.com/diewo77/go-gate"
)

func TestStaticProfile_HasPermission(t *testing.T) {
	profile := gate.NewStaticProfile(1, "editor",
		gate.NewPermission("product", gate.ActionCreate),
		gate.NewPermission("product", gate.ActionUpdate),
	)

	if !profile.HasPermission(gate.NewPermission("product", gate.ActionCreate)) {
		t.Error("should have product:create permission")
	}
	if profile.HasPermission(gate.NewPermission("product", gate.ActionDelete)) {
		t.Error("should not have product:delete permission")
	}
}

func TestStaticProfile_HasPermission_Wildcard(t *testing.T) {
	profile := gate.NewStaticProfile(1, "admin", gate.PermissionSuperAdmin)

	if !profile.HasPermission(gate.NewPermission("product", gate.ActionCreate)) {
		t.Error("superadmin should have any permission")
	}
	if !profile.HasPermission(gate.NewPermission("invoice", gate.ActionDelete)) {
		t.Error("superadmin should have any permission")
	}
}

func TestStaticResolver(t *testing.T) {
	resolver := gate.NewStaticResolver[uint]()
	profile := gate.NewStaticProfile(1, "viewer", gate.NewPermission("product", gate.ActionView))
	resolver.Set(1, profile)

	resolved, err := resolver.Resolve(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved == nil {
		t.Fatal("expected profile, got nil")
	}
	if resolved.Name() != "viewer" {
		t.Errorf("expected 'viewer', got '%s'", resolved.Name())
	}

	// Unknown user returns nil
	unknown, err := resolver.Resolve(context.Background(), 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if unknown != nil {
		t.Error("expected nil for unknown user")
	}
}
