package gate

import "strings"

// Permission represents an allowed action on a resource type.
// Format: "resource:action" (e.g., "product:create", "invoice:view")
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
	// Superadmin matches everything
	if p == PermissionSuperAdmin {
		return true
	}
	// Exact match
	if p == requested {
		return true
	}
	// Check resource wildcard: "product:*" matches "product:create"
	res, act := p.Parse()
	reqRes, _ := requested.Parse()
	if res == reqRes && string(act) == WildcardAll {
		return true
	}
	return false
}
