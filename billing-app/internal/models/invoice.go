package models

import "time"

// Invoicing models
type Invoice struct {
	ID             uint          `gorm:"primaryKey"`
	Status         string        `gorm:"not null"` // draft, final
	Items          []InvoiceItem `gorm:"foreignKey:InvoiceID"`
	CompanyID      uint          `gorm:"not null"`
	ClientID       uint          `gorm:"not null"`
	TVAIntraClient string        // numéro TVA intracommunautaire du client
	Remise         float64       // montant ou pourcentage de remise sur la facture
	Acompte        float64       // montant d'acompte versé
	Avoir          float64       // montant d'avoir appliqué
	Currency       string        `gorm:"not null;default:'EUR'"` // devise de la facture
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type InvoiceItem struct {
	ID        uint    `gorm:"primaryKey"`
	InvoiceID uint    `gorm:"not null"`
	ProductID uint    `gorm:"not null"`
	Quantity  int     `gorm:"not null"`
	Product   Product `gorm:"foreignKey:ProductID"`
	Remise    float64 // remise spécifique à la ligne
}
