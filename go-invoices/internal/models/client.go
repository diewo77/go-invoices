package models

import (
	"time"

	"gorm.io/gorm"
)

// Client represents a customer/client in the billing system.
// Implements the Ownable interface for ownership-based authorization.
type Client struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// UserID is the owner of this client (for multi-tenant isolation)
	UserID uint `gorm:"index;not null" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"-"`

	// Client information
	Name    string `gorm:"size:255;not null" json:"name"`
	Email   string `gorm:"size:255" json:"email,omitempty"`
	Phone   string `gorm:"size:50" json:"phone,omitempty"`
	Company string `gorm:"size:255" json:"company,omitempty"`

	// Address
	Address    string `gorm:"size:500" json:"address,omitempty"`
	City       string `gorm:"size:100" json:"city,omitempty"`
	PostalCode string `gorm:"size:20" json:"postal_code,omitempty"`
	Country    string `gorm:"size:100" json:"country,omitempty"`

	// Tax information
	SIRET    string `gorm:"size:14" json:"siret,omitempty"`
	VATNumber string `gorm:"size:20" json:"vat_number,omitempty"`

	// Relations
	Invoices []Invoice `gorm:"foreignKey:ClientID" json:"invoices,omitempty"`
}

// GetUserID implements the Ownable interface for authorization.
func (c *Client) GetUserID() uint {
	return c.UserID
}

// FullAddress returns the formatted full address.
func (c *Client) FullAddress() string {
	addr := c.Address
	if c.PostalCode != "" || c.City != "" {
		if addr != "" {
			addr += "\n"
		}
		addr += c.PostalCode
		if c.PostalCode != "" && c.City != "" {
			addr += " "
		}
		addr += c.City
	}
	if c.Country != "" {
		if addr != "" {
			addr += "\n"
		}
		addr += c.Country
	}
	return addr
}
