package models

import "time"

type Product struct {
	ID        uint    `gorm:"primaryKey"`
	Name      string  `gorm:"size:255;not null"`
	UnitPrice float64 `gorm:"not null"`
	VATRate   float64 `gorm:"not null"`
}

type Invoice struct {
	ID         uint          `gorm:"primaryKey"`
	ClientName string        `gorm:"size:255;not null"`
	IssueDate  time.Time     `gorm:"not null"`
	DueDate    time.Time     `gorm:"not null"`
	Status     string        `gorm:"size:50;not null"` // "draft" ou "final"
	TotalHT    float64       `gorm:"not null"`
	TotalTVA   float64       `gorm:"not null"`
	TotalTTC   float64       `gorm:"not null"`
	Items      []InvoiceItem `gorm:"constraint:OnDelete:CASCADE"`
}

type InvoiceItem struct {
	ID        uint    `gorm:"primaryKey"`
	InvoiceID uint    `gorm:"index;not null"`
	ProductID uint    `gorm:"not null"`
	Quantity  int     `gorm:"not null"`
	UnitPrice float64 `gorm:"not null"`
	AmountHT  float64 `gorm:"not null"`
	AmountVAT float64 `gorm:"not null"`
}
