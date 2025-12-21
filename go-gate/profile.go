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
// U is the user type (e.g., uint for userID, *User for full user struct).
type ProfileResolver[U any] interface {
	Resolve(ctx context.Context, user U) (Profile, error)
}

// StaticProfile is a simple in-memory profile implementation.
// Useful for testing or static configuration.
type StaticProfile struct {
	id          uint
	name        string
	permissions map[Permission]bool
}

// NewStaticProfile creates a profile with the given permissions.
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

// Permissions returns all permissions in this profile.
func (p *StaticProfile) Permissions() []Permission {
	perms := make([]Permission, 0, len(p.permissions))
	for perm := range p.permissions {
		perms = append(perms, perm)
	}
	return perms
}

// HasPermission checks if the profile has the requested permission.
// Supports wildcard matching.
func (p *StaticProfile) HasPermission(requested Permission) bool {
	for perm := range p.permissions {
		if perm.Matches(requested) {
			return true
		}
	}
	return false
}

// StaticResolver is a simple in-memory resolver for testing.
type StaticResolver[U comparable] struct {
	profiles map[U]Profile
}

// NewStaticResolver creates a resolver with predefined user-profile mappings.
func NewStaticResolver[U comparable]() *StaticResolver[U] {
	return &StaticResolver[U]{profiles: make(map[U]Profile)}
}

// Set assigns a profile to a user.
func (r *StaticResolver[U]) Set(user U, profile Profile) {
	r.profiles[user] = profile
}

// Resolve returns the profile for the given user.
func (r *StaticResolver[U]) Resolve(_ context.Context, user U) (Profile, error) {
	if profile, ok := r.profiles[user]; ok {
		return profile, nil
	}
	return nil, nil
}
