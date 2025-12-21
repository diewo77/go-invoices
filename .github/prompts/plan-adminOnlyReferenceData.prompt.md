# Plan: Admin-Only Restrictions for ProductType/UnitType

TL;DR: Add admin-only restrictions for create/update/delete on ProductType and UnitType while keeping read-only access for regular users.

## Steps

### 1. Create `internal/policy/role_checker.go`

```go
package policy

import (
	"context"

	"github.com/diewo77/billing-app/internal/models"
	"gorm.io/gorm"
)

// RoleChecker provides role lookup for policies that need admin checks.
type RoleChecker interface {
	IsAdmin(ctx context.Context, userID uint) bool
}

// DBRoleChecker implements RoleChecker using GORM.
type DBRoleChecker struct {
	DB *gorm.DB
}

// NewDBRoleChecker creates a RoleChecker backed by the database.
func NewDBRoleChecker(db *gorm.DB) *DBRoleChecker {
	return &DBRoleChecker{DB: db}
}

// IsAdmin returns true if the user has the "admin" role.
func (c *DBRoleChecker) IsAdmin(ctx context.Context, userID uint) bool {
	if userID == 0 {
		return false
	}
	var user models.User
	if err := c.DB.Preload("Role").First(&user, userID).Error; err != nil {
		return false
	}
	return user.Role.Name == "admin"
}
```

---

### 2. Create `internal/policy/product_type_policy.go`

```go
package policy

import (
	"context"
)

// ProductTypePolicy defines authorization rules for ProductType (global reference data).
// List/View: any authenticated user.
// Create/Update/Delete: admin only.
type ProductTypePolicy struct {
	RoleChecker RoleChecker
}

// NewProductTypePolicy creates a ProductTypePolicy with the given role checker.
func NewProductTypePolicy(rc RoleChecker) *ProductTypePolicy {
	return &ProductTypePolicy{RoleChecker: rc}
}

// Can checks whether userID may perform action on ProductType.
func (p *ProductTypePolicy) Can(ctx context.Context, userID uint, action Action, resource any) bool {
	if userID == 0 {
		return false
	}

	switch action {
	case ActionList, ActionView:
		return true // Any authenticated user can list/view
	case ActionCreate, ActionUpdate, ActionDelete:
		return p.RoleChecker.IsAdmin(ctx, userID) // Admin only
	}
	return false
}
```

---

### 3. Create `internal/policy/unit_type_policy.go`

```go
package policy

import (
	"context"
)

// UnitTypePolicy defines authorization rules for UnitType (global reference data).
// List/View: any authenticated user.
// Create/Update/Delete: admin only.
type UnitTypePolicy struct {
	RoleChecker RoleChecker
}

// NewUnitTypePolicy creates a UnitTypePolicy with the given role checker.
func NewUnitTypePolicy(rc RoleChecker) *UnitTypePolicy {
	return &UnitTypePolicy{RoleChecker: rc}
}

// Can checks whether userID may perform action on UnitType.
func (p *UnitTypePolicy) Can(ctx context.Context, userID uint, action Action, resource any) bool {
	if userID == 0 {
		return false
	}

	switch action {
	case ActionList, ActionView:
		return true // Any authenticated user can list/view
	case ActionCreate, ActionUpdate, ActionDelete:
		return p.RoleChecker.IsAdmin(ctx, userID) // Admin only
	}
	return false
}
```

---

### 4. Update `internal/server/router.go`

Add after the existing gate registrations:

```go
// Role checker for admin-restricted policies
roleChecker := policy.NewDBRoleChecker(db)

// Register policies for global reference data (admin-only write access)
gate.Register("product_type", policy.NewProductTypePolicy(roleChecker))
gate.Register("unit_type", policy.NewUnitTypePolicy(roleChecker))
```

---

### 5. Update `internal/handlers/product_type.go`

Add Gate to struct and check admin for write operations:

```go
type ProductTypeHandler struct {
	DB   *gorm.DB
	Gate *policy.Gate
}

func NewProductTypeHandler(db *gorm.DB, gate *policy.Gate) *ProductTypeHandler {
	return &ProductTypeHandler{DB: db, Gate: gate}
}
```

In Create/Update/Delete methods, add at the start:

```go
if !h.Gate.Can(r.Context(), uid, policy.ActionCreate, "product_type", nil) {
	httpx.JSONError(w, http.StatusForbidden, "forbidden", nil)
	return
}
```

---

### 6. Update `internal/handlers/unit_type.go`

Same pattern as ProductTypeHandler:

```go
type UnitTypeHandler struct {
	DB   *gorm.DB
	Gate *policy.Gate
}

func NewUnitTypeHandler(db *gorm.DB, gate *policy.Gate) *UnitTypeHandler {
	return &UnitTypeHandler{DB: db, Gate: gate}
}
```

Add policy checks in Create/Update/Delete methods.

---

### 7. Update router handler instantiation

```go
pth := handlers.NewProductTypeHandler(db, gate)
uth := handlers.NewUnitTypeHandler(db, gate)
```

---

### 8. Add tests

Create `internal/policy/role_checker_test.go` and update policy tests:

```go
func TestProductTypePolicy_AdminOnly(t *testing.T) {
	mockChecker := &mockRoleChecker{isAdmin: false}
	policy := NewProductTypePolicy(mockChecker)
	ctx := context.Background()

	// Regular user can list/view
	assert.True(t, policy.Can(ctx, 1, ActionList, nil))
	assert.True(t, policy.Can(ctx, 1, ActionView, nil))

	// Regular user cannot create/update/delete
	assert.False(t, policy.Can(ctx, 1, ActionCreate, nil))
	assert.False(t, policy.Can(ctx, 1, ActionUpdate, nil))
	assert.False(t, policy.Can(ctx, 1, ActionDelete, nil))

	// Admin can do everything
	mockChecker.isAdmin = true
	assert.True(t, policy.Can(ctx, 1, ActionCreate, nil))
	assert.True(t, policy.Can(ctx, 1, ActionUpdate, nil))
	assert.True(t, policy.Can(ctx, 1, ActionDelete, nil))
}

type mockRoleChecker struct {
	isAdmin bool
}

func (m *mockRoleChecker) IsAdmin(ctx context.Context, userID uint) bool {
	return m.isAdmin
}
```

---

## Files to Create/Modify

| File                                     | Action                                     |
| ---------------------------------------- | ------------------------------------------ |
| `internal/policy/role_checker.go`        | Create                                     |
| `internal/policy/product_type_policy.go` | Create                                     |
| `internal/policy/unit_type_policy.go`    | Create                                     |
| `internal/policy/policy_test.go`         | Add tests for new policies                 |
| `internal/handlers/product_type.go`      | Modify - Add Gate + policy checks          |
| `internal/handlers/unit_type.go`         | Modify - Add Gate + policy checks          |
| `internal/server/router.go`              | Modify - Register new policies + pass gate |

---

## Authorization Matrix

| Resource    | Action | Regular User | Admin |
| ----------- | ------ | ------------ | ----- |
| ProductType | List   | ✅           | ✅    |
| ProductType | View   | ✅           | ✅    |
| ProductType | Create | ❌ 403       | ✅    |
| ProductType | Update | ❌ 403       | ✅    |
| ProductType | Delete | ❌ 403       | ✅    |
| UnitType    | List   | ✅           | ✅    |
| UnitType    | View   | ✅           | ✅    |
| UnitType    | Create | ❌ 403       | ✅    |
| UnitType    | Update | ❌ 403       | ✅    |
| UnitType    | Delete | ❌ 403       | ✅    |
