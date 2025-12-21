package models

import (
	"time"

	"gorm.io/gorm"
)

// CompanySettings represents the user's company information for invoices.
type CompanySettings struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// UserID is the owner of these settings
	UserID uint `gorm:"uniqueIndex;not null" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"-"`

	// Company information
	Name    string `gorm:"size:255;not null" json:"name"`
	Email   string `gorm:"size:255" json:"email,omitempty"`
	Phone   string `gorm:"size:50" json:"phone,omitempty"`
	Website string `gorm:"size:255" json:"website,omitempty"`

	// Address
	Address    string `gorm:"size:500" json:"address,omitempty"`
	City       string `gorm:"size:100" json:"city,omitempty"`
	PostalCode string `gorm:"size:20" json:"postal_code,omitempty"`
	Country    string `gorm:"size:100" json:"country,omitempty"`

	// Tax & Legal information
	SIRET     string `gorm:"size:14" json:"siret,omitempty"`
	VATNumber string `gorm:"size:20" json:"vat_number,omitempty"`
	RCS       string `gorm:"size:100" json:"rcs,omitempty"`
	Capital   string `gorm:"size:100" json:"capital,omitempty"`

	// Branding
	LogoURL string `gorm:"size:500" json:"logo_url,omitempty"`
}

// GetUserID implements the Ownable interface.
func (c *CompanySettings) GetUserID() uint {
	return c.UserID
}
