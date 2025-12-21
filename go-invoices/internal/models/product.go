package models

import (
	"time"

	"gorm.io/gorm"
)

// Product represents a product or service in the billing system.
// Implements the Ownable interface for ownership-based authorization.
type Product struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// UserID is the owner of this product (for multi-tenant isolation)
	UserID uint `gorm:"index;not null" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"-"`

	// Product information
	Code        string  `gorm:"size:50;not null;uniqueIndex:idx_product_user_code" json:"code"`
	Name        string  `gorm:"size:255;not null" json:"name"`
	Description string  `gorm:"type:text" json:"description,omitempty"`
	UnitPrice   float64 `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	Unit        string  `gorm:"size:50;default:'unit'" json:"unit"` // unit, hour, day, kg, etc.

	// VAT rate stored as decimal (0.20 = 20%)
	VATRate float64 `gorm:"type:decimal(5,4);default:0.20" json:"vat_rate"`

	// Optional categorization
	Category string `gorm:"size:100" json:"category,omitempty"`
	IsActive bool   `gorm:"default:true" json:"is_active"`
}

// GetUserID implements the Ownable interface for authorization.
func (p *Product) GetUserID() uint {
	return p.UserID
}

// PriceWithVAT returns the unit price including VAT.
func (p *Product) PriceWithVAT() float64 {
	return p.UnitPrice * (1 + p.VATRate)
}

// VATAmount returns the VAT amount for one unit.
func (p *Product) VATAmount() float64 {
	return p.UnitPrice * p.VATRate
}

// VATRatePercent returns the VAT rate as a percentage (e.g., 20 for 20%).
func (p *Product) VATRatePercent() float64 {
	return p.VATRate * 100
}
