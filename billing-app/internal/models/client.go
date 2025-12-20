package models

import "time"

// Client entity
type Client struct {
	ID               uint    `gorm:"primaryKey"`
	UserID           uint    `gorm:"not null;index"` // FK vers User
	User             User    `gorm:"foreignKey:UserID"`
	Nom              string  `gorm:"not null;index"` // Raison sociale ou nom
	NomCommercial    string  `gorm:"index"`
	Contact          string  // Nom du contact principal
	AddressID        uint    // clé étrangère vers Address (principale)
	Address          Address `gorm:"foreignKey:AddressID"`
	BillingAddressID uint    // clé étrangère vers Address (facturation)
	BillingAddress   Address `gorm:"foreignKey:BillingAddressID"`
	Telephone        string
	Email            string
	SiteWeb          string
	SIREN            string `gorm:"index"` // France
	SIRET            string `gorm:"index"` // France
	CodeNAF          string // France
	CodeAPE          string // France
	TVAIntra         string `gorm:"index"` // Numéro TVA intracommunautaire
	RCS              string // France, nullable
	Greffe           string // France, nullable
	RM               string // France, nullable
	DeptRM           string // France, nullable
	RCPro            string // Numéro d'assurance RC Pro
	MentionsLegales  string // Mentions légales personnalisées
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
