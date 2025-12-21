package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// InvoiceStatus represents the status of an invoice.
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusFinal     InvoiceStatus = "final"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
)

// Invoice represents a billing invoice.
// Implements the Ownable interface for ownership-based authorization.
type Invoice struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// UserID is the owner of this invoice (for multi-tenant isolation)
	UserID uint `gorm:"index;not null" json:"user_id"`
	User   User `gorm:"foreignKey:UserID" json:"-"`

	// Invoice identification
	Number    string `gorm:"size:50;uniqueIndex" json:"number"`
	Reference string `gorm:"size:100" json:"reference,omitempty"`

	// Client relationship
	ClientID uint    `gorm:"index;not null" json:"client_id"`
	Client   *Client `gorm:"foreignKey:ClientID" json:"client,omitempty"`

	// Invoice dates
	IssueDate time.Time  `gorm:"not null" json:"issue_date"`
	DueDate   time.Time  `gorm:"not null" json:"due_date"`
	PaidDate  *time.Time `json:"paid_date,omitempty"`

	// Status
	Status InvoiceStatus `gorm:"size:20;default:'draft'" json:"status"`

	// Notes and terms
	Notes          string `gorm:"type:text" json:"notes,omitempty"`
	PaymentTerms   string `gorm:"size:500" json:"payment_terms,omitempty"`
	FooterText     string `gorm:"type:text" json:"footer_text,omitempty"`

	// Invoice items
	Items []InvoiceItem `gorm:"foreignKey:InvoiceID" json:"items,omitempty"`
}

// GetUserID implements the Ownable interface for authorization.
func (i *Invoice) GetUserID() uint {
	return i.UserID
}

// IsDraft returns true if the invoice is in draft status.
func (i *Invoice) IsDraft() bool {
	return i.Status == InvoiceStatusDraft
}

// IsFinal returns true if the invoice has been finalized.
func (i *Invoice) IsFinal() bool {
	return i.Status == InvoiceStatusFinal || i.Status == InvoiceStatusPaid
}

// CanEdit returns true if the invoice can still be edited.
func (i *Invoice) CanEdit() bool {
	return i.Status == InvoiceStatusDraft
}

// TotalHT calculates the total excluding VAT.
func (i *Invoice) TotalHT() float64 {
	var total float64
	for _, item := range i.Items {
		total += item.TotalHT()
	}
	return total
}

// TotalVAT calculates the total VAT amount.
func (i *Invoice) TotalVAT() float64 {
	var total float64
	for _, item := range i.Items {
		total += item.TotalVAT()
	}
	return total
}

// TotalTTC calculates the total including VAT.
func (i *Invoice) TotalTTC() float64 {
	return i.TotalHT() + i.TotalVAT()
}

// InvoiceItem represents a line item on an invoice.
type InvoiceItem struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Parent invoice
	InvoiceID uint     `gorm:"index;not null" json:"invoice_id"`
	Invoice   *Invoice `gorm:"foreignKey:InvoiceID" json:"-"`

	// Optional product reference (can be null for custom items)
	ProductID *uint    `gorm:"index" json:"product_id,omitempty"`
	Product   *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`

	// Item details (copied from product or custom)
	Description string  `gorm:"size:500;not null" json:"description"`
	Quantity    float64 `gorm:"type:decimal(10,3);not null;default:1" json:"quantity"`
	UnitPrice   float64 `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	Unit        string  `gorm:"size:50;default:'unit'" json:"unit"`
	VATRate     float64 `gorm:"type:decimal(5,4);not null" json:"vat_rate"`

	// Position for ordering
	Position int `gorm:"default:0" json:"position"`
}

// TotalHT calculates the line total excluding VAT.
func (item *InvoiceItem) TotalHT() float64 {
	return item.Quantity * item.UnitPrice
}

// TotalVAT calculates the VAT amount for this line.
func (item *InvoiceItem) TotalVAT() float64 {
	return item.TotalHT() * item.VATRate
}

// TotalTTC calculates the line total including VAT.
func (item *InvoiceItem) TotalTTC() float64 {
	return item.TotalHT() + item.TotalVAT()
}

// GenerateInvoiceNumber generates a unique invoice number.
// Format: INV-YYYY-NNNN (e.g., INV-2025-0001)
func GenerateInvoiceNumber(db *gorm.DB, userID uint, year int) (string, error) {
	var count int64
	err := db.Model(&Invoice{}).
		Where("user_id = ? AND EXTRACT(YEAR FROM issue_date) = ?", userID, year).
		Count(&count).Error
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("INV-%d-%04d", year, count+1), nil
}
