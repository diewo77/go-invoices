package models

import (
	"time"

	"gorm.io/gorm"
)

// Product domain models
type ProductType struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null;unique"` // ex: Vente de marchandises, Prestations de services, etc.
	Code      string // BIC, BNC, etc.
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UnitType struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null;unique"` // ex: pièce, heure, kg, etc.
	Symbol    string // ex: h, kg, m, etc.
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Product struct {
	ID        uint            `gorm:"primaryKey"`
	CompanyID uint            `gorm:"not null;index"` // FK vers CompanySettings
	Company   CompanySettings `gorm:"foreignKey:CompanyID"`
	// Code produit unique par utilisateur (créateur). Permet un identifiant lisible.
	Code          string      `gorm:"size:40;not null;index:idx_user_code,unique,priority:2"`
	UserID        uint        `gorm:"not null;index:idx_user_code,priority:1"` // propriétaire/créateur
	Name          string      `gorm:"not null"`
	UnitPrice     float64     `gorm:"not null"`
	VATRate       float64     `gorm:"not null"` // e.g. 0.20 for 20%
	ProductTypeID uint        // clé étrangère vers ProductType
	ProductType   ProductType `gorm:"foreignKey:ProductTypeID"`
	UnitTypeID    uint        // clé étrangère vers UnitType
	UnitType      UnitType    `gorm:"foreignKey:UnitTypeID"`
	Currency      string      `gorm:"not null;default:'EUR'"` // devise du produit
	// Contrainte d'unicité composite (UserID, Code) gérée via les index ci-dessus
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
