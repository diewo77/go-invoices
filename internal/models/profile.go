package models

import (
	"time"

	"gorm.io/gorm"
)

// Profile represents a user authorization profile that groups permissions.
// A user is assigned to one profile, inheriting all its permissions.
type Profile struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Name        string         `gorm:"uniqueIndex;size:100;not null" json:"name"`
	Description string         `gorm:"size:500" json:"description,omitempty"`
	IsSystem    bool           `gorm:"default:false" json:"is_system"`
	// Permissions holds the set of permissions this profile grants.
	// Many-to-many relationship via profile_permissions join table.
	Permissions []Permission `gorm:"many2many:profile_permissions;" json:"permissions,omitempty"`
	// Users that have this profile assigned.
	Users []User `gorm:"foreignKey:ProfileID" json:"users,omitempty"`
}

// Permission represents a single action allowed on a resource type.
// Format: "resource:action" (e.g., "product:create", "invoice:read").
type Permission struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	ResourceType string         `gorm:"size:50;not null;index:idx_perm_resource_action" json:"resource_type"`
	Action       string         `gorm:"size:50;not null;index:idx_perm_resource_action" json:"action"`
	Description  string         `gorm:"size:200" json:"description,omitempty"`
}

// Code returns the permission in "resource:action" format for matching.
func (p Permission) Code() string {
	return p.ResourceType + ":" + p.Action
}
