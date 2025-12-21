# Plan: Authorization System with Policies

TL;DR: Implement a Laravel-inspired Policy/Gate system for Go to centralize authorization logic, add `user_id` to Invoice model for simpler checks, and add comprehensive cross-user access tests.

## Steps

### 1. Add `UserID` field to Invoice model

**File:** `internal/models/invoice.go`

Add `UserID uint` field to simplify authorization without company join lookups.

```go
type Invoice struct {
    ID             uint
    UserID         uint          // NEW: direct user ownership
    Status         string
    Items          []InvoiceItem
    CompanyID      uint
    // ...
}
```

**Migration:** Backfill existing invoices with `user_id` from `company_settings.user_id`.

---

### 2. Create policy package

**File:** `internal/policy/policy.go`

Central authorization checkpoint with `Can(ctx, userID, action, resource) bool` pattern.

```go
package policy

import (
    "context"
    "errors"
)

var (
    ErrUnauthorized    = errors.New("unauthorized")
    ErrNoPolicyDefined = errors.New("no policy defined for resource")
)

type Action string

const (
    ActionView   Action = "view"
    ActionCreate Action = "create"
    ActionUpdate Action = "update"
    ActionDelete Action = "delete"
    ActionList   Action = "list"
)

// Policy defines authorization rules for a resource type
type Policy interface {
    Can(ctx context.Context, userID uint, action Action, resource any) bool
}

// Gate is the central authorization checkpoint
type Gate struct {
    policies map[string]Policy
}

func NewGate() *Gate {
    return &Gate{policies: make(map[string]Policy)}
}

func (g *Gate) Register(resourceType string, p Policy) {
    g.policies[resourceType] = p
}

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

func (g *Gate) Can(ctx context.Context, userID uint, action Action, resourceType string, resource any) bool {
    return g.Authorize(ctx, userID, action, resourceType, resource) == nil
}
```

---

### 3. Implement resource policies

**Files:**

- `internal/policy/invoice_policy.go`
- `internal/policy/product_policy.go`
- `internal/policy/client_policy.go`

Each encapsulates ownership rules.

```go
// internal/policy/invoice_policy.go
package policy

import (
    "context"
    "github.com/diewo77/billing-app/internal/models"
)

type InvoicePolicy struct{}

func (p *InvoicePolicy) Can(ctx context.Context, userID uint, action Action, resource any) bool {
    switch action {
    case ActionList, ActionCreate:
        return userID > 0
    case ActionView, ActionUpdate, ActionDelete:
        inv, ok := resource.(*models.Invoice)
        if !ok || inv == nil {
            return false
        }
        return inv.UserID == userID
    }
    return false
}
```

```go
// internal/policy/product_policy.go
package policy

import (
    "context"
    "github.com/diewo77/billing-app/internal/models"
)

type ProductPolicy struct{}

func (p *ProductPolicy) Can(ctx context.Context, userID uint, action Action, resource any) bool {
    switch action {
    case ActionList, ActionCreate:
        return userID > 0
    case ActionView, ActionUpdate, ActionDelete:
        prod, ok := resource.(*models.Product)
        if !ok || prod == nil {
            return false
        }
        return prod.UserID == userID
    }
    return false
}
```

```go
// internal/policy/client_policy.go
package policy

import (
    "context"
    "github.com/diewo77/billing-app/internal/models"
)

type ClientPolicy struct{}

func (p *ClientPolicy) Can(ctx context.Context, userID uint, action Action, resource any) bool {
    switch action {
    case ActionList, ActionCreate:
        return userID > 0
    case ActionView, ActionUpdate, ActionDelete:
        client, ok := resource.(*models.Client)
        if !ok || client == nil {
            return false
        }
        return client.UserID == userID
    }
    return false
}
```

---

### 4. Initialize Gate in main.go

**File:** `cmd/server/main.go`

```go
// After DB connection
gate := policy.NewGate()
gate.Register("invoice", &policy.InvoicePolicy{})
gate.Register("product", &policy.ProductPolicy{})
gate.Register("client", &policy.ClientPolicy{})

// Pass gate to handlers or make it accessible
```

---

### 5. Update handlers to use Gate

**Files:** All handlers in `internal/handlers/`

Replace inline auth checks with policy checks:

```go
// Before (inline check):
uid, _ := auth.UserIDFromContext(r.Context())
if uid == 0 {
    httpx.JSONError(w, 401, "unauthorized", nil)
    return
}
var product models.Product
h.DB.First(&product, id)
// No ownership check!

// After (policy check):
uid, _ := auth.UserIDFromContext(r.Context())
var product models.Product
h.DB.First(&product, id)
if !h.Gate.Can(r.Context(), uid, policy.ActionView, "product", &product) {
    httpx.JSONError(w, 403, "forbidden", nil)
    return
}
```

---

### 6. Add authorization tests

**File:** `internal/policy/policy_test.go`

```go
func TestProductPolicy_CrossUserAccess(t *testing.T) {
    userA := uint(1)
    userB := uint(2)

    productOwnedByA := &models.Product{UserID: userA}
    policy := &ProductPolicy{}

    // User A can view their own product
    assert.True(t, policy.Can(ctx, userA, ActionView, productOwnedByA))

    // User B cannot view User A's product
    assert.False(t, policy.Can(ctx, userB, ActionView, productOwnedByA))

    // User B cannot delete User A's product
    assert.False(t, policy.Can(ctx, userB, ActionDelete, productOwnedByA))
}
```

**File:** `internal/handlers/product_authorization_test.go`

```go
func TestProductHandler_CrossUserAccessDenied(t *testing.T) {
    db := setupTestDB(t)

    // Create two users
    userA := models.User{Email: "a@test.com", Password: "hash"}
    userB := models.User{Email: "b@test.com", Password: "hash"}
    db.Create(&userA)
    db.Create(&userB)

    // Create product owned by User A
    product := models.Product{UserID: userA.ID, Name: "A's Product", Code: "PRODA"}
    db.Create(&product)

    // User B tries to view User A's product
    req := httptest.NewRequest("GET", "/products/show?id="+strconv.Itoa(int(product.ID)), nil)
    req = req.WithContext(auth.WithUserID(req.Context(), userB.ID))
    rec := httptest.NewRecorder()

    handler.Show(rec, req)

    // Should be 403 Forbidden or redirect
    assert.NotEqual(t, 200, rec.Code)
}
```

---

## Decisions Needed

### 1. ProductType/UnitType as global or per-user?

| Option                 | Pros                              | Cons                           |
| ---------------------- | --------------------------------- | ------------------------------ |
| **A: Global (shared)** | Simple, consistent reference data | All users see same types       |
| **B: Per-user**        | Users can customize               | More complex, data duplication |

**Recommendation:** Option A (global) - these are standard reference data like "pièce", "heure", "Vente de marchandises".

### 2. Role-based permissions?

Current `Role` model has `Level` field but is unused. Could extend policies:

```go
func (p *InvoicePolicy) Can(ctx context.Context, userID uint, action Action, resource any) bool {
    // Check role level for admin override
    if p.userIsAdmin(userID) {
        return true
    }
    // Normal ownership check
    // ...
}
```

**Recommendation:** Defer role-based checks until needed. Focus on user isolation first.

---

## Migration Script

```sql
-- Add user_id to invoices
ALTER TABLE invoices ADD COLUMN user_id BIGINT REFERENCES users(id);

-- Backfill from company_settings
UPDATE invoices i
SET user_id = (
    SELECT cs.user_id
    FROM company_settings cs
    WHERE cs.id = i.company_id
);

-- Make NOT NULL after backfill
ALTER TABLE invoices ALTER COLUMN user_id SET NOT NULL;

-- Add index for performance
CREATE INDEX idx_invoices_user_id ON invoices(user_id);
```

---

## Files to Create/Modify

| File                                         | Action                     |
| -------------------------------------------- | -------------------------- |
| `internal/policy/policy.go`                  | Create - Gate & interfaces |
| `internal/policy/invoice_policy.go`          | Create                     |
| `internal/policy/product_policy.go`          | Create                     |
| `internal/policy/client_policy.go`           | Create                     |
| `internal/policy/policy_test.go`             | Create - Unit tests        |
| `internal/models/invoice.go`                 | Modify - Add UserID        |
| `cmd/server/main.go`                         | Modify - Initialize Gate   |
| `internal/handlers/product.go`               | Modify - Use Gate          |
| `internal/handlers/invoice.go`               | Modify - Use Gate          |
| `internal/handlers/client.go`                | Modify - Use Gate          |
| `internal/handlers/*_authorization_test.go`  | Create - Cross-user tests  |
| `migrations/xxx_add_user_id_to_invoices.sql` | Create                     |
