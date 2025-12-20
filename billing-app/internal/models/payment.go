package models

import "time"

// Payment tied to invoices
type Payment struct {
	ID          uint      `gorm:"primaryKey"`
	InvoiceID   uint      `gorm:"not null"` // FK vers Invoice
	Date        time.Time `gorm:"not null"`
	Montant     float64   `gorm:"not null"`
	Mode        string    `gorm:"not null"` // ex: virement, CB, chèque, espèces
	Statut      string    `gorm:"not null"` // ex: pending, paid, failed, refunded
	Commentaire string    // optionnel
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
