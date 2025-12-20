package models

import "time"

// Company & related
type CompanySettings struct {
	ID                      uint    `gorm:"primaryKey"`
	UserID                  uint    `gorm:"not null;index"` // FK vers User obligatoire
	User                    User    `gorm:"foreignKey:UserID"`
	RaisonSociale           string  `gorm:"not null;index"`
	NomCommercial           string  `gorm:"not null;index"`
	SIREN                   string  `gorm:"size:9;not null;index"`
	SIRET                   string  `gorm:"size:14;not null;index"`
	CodeNAF                 string  `gorm:"not null"`
	TVA                     float64 // nullable, taux de TVA
	RCS                     string  // nullable
	Greffe                  string  // nullable
	RM                      string  // nullable
	DeptRM                  string  // nullable
	Capital                 float64 `gorm:"default:0"`
	ActivitePrincipale      string
	AgrementSAP             bool `gorm:"not null"`
	DateCreation            time.Time
	TypeImposition          string `gorm:"not null"`
	TypeDeclarant           string `gorm:"not null;default:'Déclarant 1'"`
	FrequenceUrssaf         string `gorm:"not null"`
	RedevableTVA            bool   `gorm:"not null"`
	DatePremiereDeclaration time.Time
	FormeJuridique          string  `gorm:"not null"`
	RegimeFiscal            string  `gorm:"not null"`
	AddressID               uint    // clé étrangère vers Address (principale)
	Address                 Address `gorm:"foreignKey:AddressID"`
	BillingAddressID        uint    // clé étrangère vers Address (facturation)
	BillingAddress          Address `gorm:"foreignKey:BillingAddressID"`
	Telephone               string  // téléphone de contact
	Email                   string  // email de contact
	SiteWeb                 string  // site web
	IBAN                    string  // IBAN/RIB pour facturation
	LogoURL                 string  // URL ou chemin du logo
	MentionsLegales         string  // mentions légales personnalisées
	RCPro                   string  // numéro d'assurance RC Pro
	TVAIntra                string  // numéro TVA intracommunautaire
	CodeAPE                 string  // code APE (souvent identique au code NAF)
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type UserCompany struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"not null"`
	CompanyID uint `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
