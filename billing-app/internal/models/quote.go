package models

import "time"

// Quote / estimate models
type Quote struct {
	ID                   uint        `gorm:"primaryKey"`
	Status               string      `gorm:"not null"` // draft, sent, accepted, rejected, converted
	CompanyID            uint        `gorm:"not null"`
	ClientID             uint        `gorm:"not null"`
	Items                []QuoteItem `gorm:"foreignKey:QuoteID"`
	TotalHT              float64
	TotalTVA             float64
	TotalTTC             float64
	ConvertedToInvoiceID uint // si le devis est converti en facture
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type QuoteItem struct {
	ID        uint    `gorm:"primaryKey"`
	QuoteID   uint    `gorm:"not null"`
	ProductID uint    `gorm:"not null"`
	Quantity  int     `gorm:"not null"`
	Product   Product `gorm:"foreignKey:ProductID"`
}
