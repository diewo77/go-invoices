# Plan: Profile-Based Authorization with Hybrid Approach

**TL;DR:** Extend `gate/` with a profile system where profiles have permission sets. Use a hybrid approach: ProfileGate for global "can user access this resource type" checks, then per-resource policies for ownership. Add caching and Admin UI for runtime management.

---

## Architecture Overview

```
Request → HybridGate
           ├── CachedResolver → Profile.HasPermission("resource:action")?
           │                    ↓ No → 403 Forbidden
           │                    ↓ Yes
           └── ResourcePolicy → Owns this specific record?
                                ↓ No → 403 Forbidden
                                ↓ Yes → Allowed
```

---

## Part 1: Gate Package Extensions

### 1.1 Create `gate/permission.go`

```go
package gate

import "strings"

// Permission represents an allowed action on a resource type.
type Permission string

// NewPermission creates a permission from resource type and action.
func NewPermission(resourceType string, action Action) Permission {
    return Permission(resourceType + ":" + string(action))
}

// Parse splits a permission into resource type and action.
func (p Permission) Parse() (resourceType string, action Action) {
    parts := strings.SplitN(string(p), ":", 2)
    if len(parts) != 2 {
        return "", ""
    }
    return parts[0], Action(parts[1])
}

// Wildcards for super permissions
const (
    WildcardAll          = "*"
    PermissionSuperAdmin Permission = "*:*"
)

// Matches checks if this permission matches a requested permission.
// Supports wildcards: "*:*" matches all, "product:*" matches all product actions.
func (p Permission) Matches(requested Permission) bool {
    if p == PermissionSuperAdmin {
        return true
    }
    if p == requested {
        return true
    }
    // Check resource wildcard: "product:*"
    res, act := p.Parse()
    reqRes, _ := requested.Parse()
    if res == reqRes && string(act) == WildcardAll {
        return true
    }
    return false
}
```

### 1.2 Create `gate/profile.go`

```go
package gate

import "context"

// Profile represents a role with a set of permissions.
type Profile interface {
    ID() uint
    Name() string
    HasPermission(permission Permission) bool
    Permissions() []Permission
}

// ProfileResolver resolves a user to their profile.
type ProfileResolver[U any] interface {
    Resolve(ctx context.Context, user U) (Profile, error)
}

// StaticProfile is a simple in-memory profile implementation.
type StaticProfile struct {
    id          uint
    name        string
    permissions map[Permission]bool
}

func NewStaticProfile(id uint, name string, permissions ...Permission) *StaticProfile {
    p := &StaticProfile{
        id:          id,
        name:        name,
        permissions: make(map[Permission]bool),
    }
    for _, perm := range permissions {
        p.permissions[perm] = true
    }
    return p
}

func (p *StaticProfile) ID() uint     { return p.id }
func (p *StaticProfile) Name() string { return p.name }

func (p *StaticProfile) Permissions() []Permission {
    perms := make([]Permission, 0, len(p.permissions))
    for perm := range p.permissions {
        perms = append(perms, perm)
    }
    return perms
}

func (p *StaticProfile) HasPermission(requested Permission) bool {
    for perm := range p.permissions {
        if perm.Matches(requested) {
            return true
        }
    }
    return false
}
```

### 1.3 Create `gate/hybrid_gate.go`

```go
package gate

import "context"

// HybridGate combines profile-based global permissions with resource-specific policies.
// First checks if user's profile has the permission, then checks the resource policy.
type HybridGate[U comparable] struct {
    resolver ProfileResolver[U]
    policies map[string]Policy[U]
}

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
// 1. User is valid (non-zero)
// 2. User's profile has permission for resource:action
// 3. If a resource policy exists, checks ownership
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

func (g *HybridGate[U]) Can(ctx context.Context, user U, action Action, resourceType string, resource any) bool {
    return g.Authorize(ctx, user, action, resourceType, resource) == nil
}

// CanProfile checks only the profile permission, without ownership check.
// Useful for UI to show/hide buttons.
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
```

### 1.4 Create `gate/cached_resolver.go`

```go
package gate

import (
    "context"
    "sync"
    "time"
)

// CachedResolver wraps a ProfileResolver with TTL-based caching.
type CachedResolver[U comparable] struct {
    inner  ProfileResolver[U]
    cache  map[U]*cacheEntry
    mu     sync.RWMutex
    ttl    time.Duration
}

type cacheEntry struct {
    profile   Profile
    expiresAt time.Time
}

func NewCachedResolver[U comparable](inner ProfileResolver[U], ttl time.Duration) *CachedResolver[U] {
    return &CachedResolver[U]{
        inner: inner,
        cache: make(map[U]*cacheEntry),
        ttl:   ttl,
    }
}

func (r *CachedResolver[U]) Resolve(ctx context.Context, user U) (Profile, error) {
    // Check cache first
    r.mu.RLock()
    entry, ok := r.cache[user]
    r.mu.RUnlock()

    if ok && time.Now().Before(entry.expiresAt) {
        return entry.profile, nil
    }

    // Cache miss or expired - fetch from inner resolver
    profile, err := r.inner.Resolve(ctx, user)
    if err != nil {
        return nil, err
    }

    // Store in cache
    r.mu.Lock()
    r.cache[user] = &cacheEntry{
        profile:   profile,
        expiresAt: time.Now().Add(r.ttl),
    }
    r.mu.Unlock()

    return profile, nil
}

// Invalidate removes a user from the cache (call when profile changes).
func (r *CachedResolver[U]) Invalidate(user U) {
    r.mu.Lock()
    delete(r.cache, user)
    r.mu.Unlock()
}

// InvalidateAll clears the entire cache.
func (r *CachedResolver[U]) InvalidateAll() {
    r.mu.Lock()
    r.cache = make(map[U]*cacheEntry)
    r.mu.Unlock()
}
```

---

## Part 2: Database Models

### 2.1 Create `internal/models/profile.go`

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

// Profile represents a role with a set of permissions.
type Profile struct {
    ID          uint           `gorm:"primaryKey" json:"id"`
    Name        string         `gorm:"uniqueIndex;size:50;not null" json:"name"`
    Description string         `gorm:"size:255" json:"description"`
    IsSystem    bool           `gorm:"default:false" json:"is_system"` // Cannot be deleted
    Permissions []Permission   `gorm:"many2many:profile_permissions;" json:"permissions"`
    Users       []User         `gorm:"foreignKey:ProfileID" json:"-"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Permission represents a single permission.
type Permission struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    ResourceType string    `gorm:"size:50;not null;uniqueIndex:idx_perm_resource_action" json:"resource_type"`
    Action       string    `gorm:"size:50;not null;uniqueIndex:idx_perm_resource_action" json:"action"`
    Description  string    `gorm:"size:255" json:"description"`
    CreatedAt    time.Time `json:"created_at"`
}

// String returns the permission in "resource:action" format.
func (p Permission) String() string {
    return p.ResourceType + ":" + p.Action
}
```

### 2.2 Update `internal/models/user.go`

Add to User struct:

```go
ProfileID *uint    `gorm:"index" json:"profile_id"`
Profile   *Profile `gorm:"foreignKey:ProfileID" json:"profile,omitempty"`
```

---

## Part 3: DB Profile Resolver with Cache

### 3.1 Create `internal/policy/db_profile_resolver.go`

```go
package policy

import (
    "context"
    "github.com/diewo77/billing-app/gate"
    "github.com/diewo77/billing-app/internal/models"
    "gorm.io/gorm"
)

// DBProfileResolver resolves user profiles from the database.
type DBProfileResolver struct {
    db *gorm.DB
}

func NewDBProfileResolver(db *gorm.DB) *DBProfileResolver {
    return &DBProfileResolver{db: db}
}

func (r *DBProfileResolver) Resolve(ctx context.Context, userID uint) (gate.Profile, error) {
    var user models.User
    err := r.db.WithContext(ctx).
        Preload("Profile.Permissions").
        First(&user, userID).Error
    if err != nil {
        return nil, err
    }
    if user.Profile == nil {
        return nil, nil // User has no profile
    }
    return &dbProfileAdapter{profile: user.Profile}, nil
}

// dbProfileAdapter adapts models.Profile to gate.Profile interface.
type dbProfileAdapter struct {
    profile *models.Profile
}

func (p *dbProfileAdapter) ID() uint     { return p.profile.ID }
func (p *dbProfileAdapter) Name() string { return p.profile.Name }

func (p *dbProfileAdapter) Permissions() []gate.Permission {
    perms := make([]gate.Permission, len(p.profile.Permissions))
    for i, perm := range p.profile.Permissions {
        perms[i] = gate.NewPermission(perm.ResourceType, gate.Action(perm.Action))
    }
    return perms
}

func (p *dbProfileAdapter) HasPermission(requested gate.Permission) bool {
    for _, perm := range p.profile.Permissions {
        gPerm := gate.NewPermission(perm.ResourceType, gate.Action(perm.Action))
        if gPerm.Matches(requested) {
            return true
        }
    }
    return false
}
```

### 3.2 Update router to use HybridGate

```go
// In internal/server/router.go

// Create DB resolver with 5-minute cache
dbResolver := policy.NewDBProfileResolver(db)
cachedResolver := gate.NewCachedResolver[uint](dbResolver, 5*time.Minute)

// Create hybrid gate
hybridGate := gate.NewHybridGate[uint](cachedResolver)

// Register ownership policies
hybridGate.Register("product", &policy.ProductPolicy{})
hybridGate.Register("invoice", &policy.InvoicePolicy{})
hybridGate.Register("client", &policy.ClientPolicy{})
```

---

## Part 4: Admin UI

### 4.1 Create `internal/handlers/admin_profile.go`

```go
// Handlers for profile management
type AdminProfileHandler struct {
    DB            *gorm.DB
    CacheResolver *gate.CachedResolver[uint] // To invalidate cache
}

// List profiles with their permissions
func (h *AdminProfileHandler) List(w http.ResponseWriter, r *http.Request)

// Create new profile
func (h *AdminProfileHandler) Create(w http.ResponseWriter, r *http.Request)

// Update profile (name, description)
func (h *AdminProfileHandler) Update(w http.ResponseWriter, r *http.Request)

// Delete profile (only if not system and no users assigned)
func (h *AdminProfileHandler) Delete(w http.ResponseWriter, r *http.Request)

// AddPermission adds a permission to a profile
func (h *AdminProfileHandler) AddPermission(w http.ResponseWriter, r *http.Request)

// RemovePermission removes a permission from a profile
func (h *AdminProfileHandler) RemovePermission(w http.ResponseWriter, r *http.Request)

// ListPermissions lists all available permissions
func (h *AdminProfileHandler) ListPermissions(w http.ResponseWriter, r *http.Request)
```

### 4.2 Create `internal/handlers/admin_user_profile.go`

```go
// Assign/change user profile
type AdminUserProfileHandler struct {
    DB            *gorm.DB
    CacheResolver *gate.CachedResolver[uint]
}

// AssignProfile assigns a profile to a user
func (h *AdminUserProfileHandler) AssignProfile(w http.ResponseWriter, r *http.Request) {
    // ... update user.ProfileID
    // Invalidate cache for this user
    h.CacheResolver.Invalidate(userID)
}
```

### 4.3 Templates

```
templates/
  admin/
    profiles/
      index.html      # List profiles with permissions
      form.html       # Create/edit profile
      permissions.html # Manage profile permissions (checkboxes)
    users/
      index.html      # List users with their profiles
      assign.html     # Assign profile to user
```

### 4.4 Create `templates/admin/profiles/index.html`

```html
{{ define "content" }}
<div class="container mx-auto p-6">
  {{ template "page-header" (dict "Title" "Gestion des profils" "ActionText"
  "Nouveau profil" "ActionLink" "/admin/profiles/new") }}

  <div class="overflow-x-auto">
    <table class="table table-zebra w-full">
      <thead>
        <tr>
          <th>Nom</th>
          <th>Description</th>
          <th>Permissions</th>
          <th>Utilisateurs</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        {{ range .Profiles }}
        <tr>
          <td>
            {{ .Name }} {{ if .IsSystem }}<span class="badge badge-info"
              >Système</span
            >{{ end }}
          </td>
          <td>{{ .Description }}</td>
          <td>
            <span class="badge badge-ghost"
              >{{ len .Permissions }} permissions</span
            >
          </td>
          <td>{{ len .Users }} utilisateurs</td>
          <td>
            <a
              href="/admin/profiles/{{ .ID }}/permissions"
              class="btn btn-sm btn-ghost"
              >Permissions</a
            >
            {{ if not .IsSystem }}
            <a
              href="/admin/profiles/{{ .ID }}/edit"
              class="btn btn-sm btn-ghost"
              >Modifier</a
            >
            <form
              method="POST"
              action="/admin/profiles/{{ .ID }}/delete"
              class="inline"
            >
              <button type="submit" class="btn btn-sm btn-error btn-ghost">
                Supprimer
              </button>
            </form>
            {{ end }}
          </td>
        </tr>
        {{ end }}
      </tbody>
    </table>
  </div>
</div>
{{ end }}
```

### 4.5 Create `templates/admin/profiles/permissions.html`

```html
{{ define "content" }}
<div class="container mx-auto p-6">
  {{ template "page-header" (dict "Title" (printf "Permissions: %s"
  .Profile.Name) "BackLink" "/admin/profiles" "BackText" "Retour") }}

  <form method="POST" action="/admin/profiles/{{ .Profile.ID }}/permissions">
    {{ range $resource, $actions := .PermissionsByResource }}
    <div class="card bg-base-100 shadow mb-4">
      <div class="card-body">
        <h3 class="card-title capitalize">{{ $resource }}</h3>
        <div class="flex flex-wrap gap-4">
          {{ range $actions }}
          <label class="label cursor-pointer gap-2">
            <input
              type="checkbox"
              name="permissions"
              value="{{ .ID }}"
              {{
              if
              $.HasPermission
              .ID
              }}checked{{
              end
              }}
              class="checkbox checkbox-primary"
            />
            <span class="label-text capitalize">{{ .Action }}</span>
          </label>
          {{ end }}
        </div>
      </div>
    </div>
    {{ end }}

    <button type="submit" class="btn btn-primary">Enregistrer</button>
  </form>
</div>
{{ end }}
```

---

## Part 5: Routes

```go
// Admin routes (require admin profile)
mux.Handle("/admin/profiles", adminAuth(aph.List))
mux.Handle("/admin/profiles/new", adminAuth(aph.New))
mux.Handle("/admin/profiles/create", adminAuth(aph.Create))
mux.Handle("/admin/profiles/{id}/edit", adminAuth(aph.Edit))
mux.Handle("/admin/profiles/{id}/update", adminAuth(aph.Update))
mux.Handle("/admin/profiles/{id}/delete", adminAuth(aph.Delete))
mux.Handle("/admin/profiles/{id}/permissions", adminAuth(aph.EditPermissions))
mux.Handle("/admin/profiles/{id}/permissions/save", adminAuth(aph.SavePermissions))

mux.Handle("/admin/users", adminAuth(auph.List))
mux.Handle("/admin/users/{id}/profile", adminAuth(auph.AssignProfile))
```

---

## Part 6: Seed Data

### 6.1 Create `internal/db/seed_permissions.go`

```go
func SeedPermissions(db *gorm.DB) error {
    resources := []string{"product", "invoice", "client", "product_type", "unit_type", "profile", "user"}
    actions := []string{"list", "view", "create", "update", "delete"}

    for _, res := range resources {
        for _, act := range actions {
            perm := models.Permission{
                ResourceType: res,
                Action:       act,
                Description:  fmt.Sprintf("Can %s %s", act, res),
            }
            db.FirstOrCreate(&perm, models.Permission{ResourceType: res, Action: act})
        }
    }

    // Superadmin wildcard
    db.FirstOrCreate(&models.Permission{
        ResourceType: "*",
        Action:       "*",
        Description:  "Full access to everything",
    }, models.Permission{ResourceType: "*", Action: "*"})

    return nil
}

func SeedProfiles(db *gorm.DB) error {
    // Admin profile with all permissions
    var allPerms []models.Permission
    db.Find(&allPerms)

    admin := models.Profile{
        Name:        "admin",
        Description: "Administrateur avec accès complet",
        IsSystem:    true,
        Permissions: allPerms,
    }
    db.FirstOrCreate(&admin, models.Profile{Name: "admin"})

    // Viewer profile with read-only
    var readPerms []models.Permission
    db.Where("action IN ?", []string{"list", "view"}).Find(&readPerms)

    viewer := models.Profile{
        Name:        "viewer",
        Description: "Accès en lecture seule",
        IsSystem:    true,
        Permissions: readPerms,
    }
    db.FirstOrCreate(&viewer, models.Profile{Name: "viewer"})

    return nil
}
```

---

## Files Summary

| File                                        | Action | Description                              |
| ------------------------------------------- | ------ | ---------------------------------------- |
| `gate/permission.go`                        | Create | Permission type with wildcard matching   |
| `gate/profile.go`                           | Create | Profile interface + StaticProfile        |
| `gate/hybrid_gate.go`                       | Create | HybridGate combining profiles + policies |
| `gate/cached_resolver.go`                   | Create | TTL-based caching wrapper                |
| `gate/permission_test.go`                   | Create | Tests for permission matching            |
| `gate/hybrid_gate_test.go`                  | Create | Tests for hybrid gate                    |
| `internal/models/profile.go`                | Create | Profile + Permission models              |
| `internal/models/user.go`                   | Modify | Add ProfileID                            |
| `internal/policy/db_profile_resolver.go`    | Create | DB-backed resolver                       |
| `internal/handlers/admin_profile.go`        | Create | Admin CRUD for profiles                  |
| `internal/handlers/admin_user_profile.go`   | Create | Assign profiles to users                 |
| `internal/db/seed_permissions.go`           | Create | Seed default data                        |
| `internal/db/migrate.go`                    | Modify | Add new models                           |
| `internal/server/router.go`                 | Modify | Use HybridGate + admin routes            |
| `templates/admin/profiles/index.html`       | Create | List profiles                            |
| `templates/admin/profiles/form.html`        | Create | Create/edit profile                      |
| `templates/admin/profiles/permissions.html` | Create | Manage permissions                       |
| `templates/admin/users/index.html`          | Create | List users                               |

---

## Cache Invalidation Strategy

- When profile permissions change → `CacheResolver.InvalidateAll()`
- When user's profile changes → `CacheResolver.Invalidate(userID)`
- Optional: Background goroutine to clean expired entries

---

## Further Considerations

1. **Rate limiting cache invalidation** to prevent abuse?
2. **Audit log** for permission changes?
3. **Permission groups** (e.g., "full access to invoices" = list+view+create+update+delete)?
