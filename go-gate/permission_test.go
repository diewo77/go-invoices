package gate_test

import (
	"testing"

	"github.com/diewo77/go-gate"
)

func TestPermission_NewPermission(t *testing.T) {
	perm := gate.NewPermission("product", gate.ActionCreate)
	if perm != "product:create" {
		t.Errorf("expected 'product:create', got '%s'", perm)
	}
}

func TestPermission_Parse(t *testing.T) {
	perm := gate.Permission("invoice:view")
	res, act := perm.Parse()
	if res != "invoice" {
		t.Errorf("expected resource 'invoice', got '%s'", res)
	}
	if act != gate.ActionView {
		t.Errorf("expected action 'view', got '%s'", act)
	}
}

func TestPermission_Parse_Invalid(t *testing.T) {
	perm := gate.Permission("invalid")
	res, act := perm.Parse()
	if res != "" || act != "" {
		t.Errorf("expected empty strings, got '%s' and '%s'", res, act)
	}
}

func TestPermission_Matches_Exact(t *testing.T) {
	perm := gate.Permission("product:create")
	if !perm.Matches("product:create") {
		t.Error("expected exact match to succeed")
	}
	if perm.Matches("product:delete") {
		t.Error("expected different action to fail")
	}
	if perm.Matches("invoice:create") {
		t.Error("expected different resource to fail")
	}
}

func TestPermission_Matches_SuperAdmin(t *testing.T) {
	perm := gate.PermissionSuperAdmin
	if !perm.Matches("product:create") {
		t.Error("superadmin should match any permission")
	}
	if !perm.Matches("invoice:delete") {
		t.Error("superadmin should match any permission")
	}
}

func TestPermission_Matches_ResourceWildcard(t *testing.T) {
	perm := gate.Permission("product:*")
	if !perm.Matches("product:create") {
		t.Error("product:* should match product:create")
	}
	if !perm.Matches("product:delete") {
		t.Error("product:* should match product:delete")
	}
	if perm.Matches("invoice:create") {
		t.Error("product:* should not match invoice:create")
	}
}
