package models

import "time"

// Address model
type Address struct {
	ID         uint   `gorm:"primaryKey"`
	Ligne1     string `gorm:"not null"` // Rue, numéro, etc.
	Ligne2     string // Complément
	CodePostal string `gorm:"not null"`
	Ville      string `gorm:"not null"`
	Pays       string `gorm:"not null"`
	Type       string // ex: "principale", "facturation", "livraison", etc.
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
