package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents an authenticated user in the system.
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Email     string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Name      string         `gorm:"size:255" json:"name,omitempty"`
	Password  string         `gorm:"size:255;not null" json:"-"` // Hashed, never exposed in JSON
	// ProfileID links the user to an authorization profile.
	// A nil value means the user has no profile assigned (limited access).
	ProfileID *uint    `gorm:"index" json:"profile_id,omitempty"`
	Profile   *Profile `gorm:"foreignKey:ProfileID" json:"profile,omitempty"`
}
