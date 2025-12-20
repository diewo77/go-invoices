package models

import "time"

// Documents & templates
type Document struct {
	ID         uint   `gorm:"primaryKey"`
	OwnerType  string // ex: "Invoice", "Quote", "Client", "CompanySettings", etc.
	OwnerID    uint   // ID de l'entité liée
	Type       string // ex: "pdf", "justificatif", "contrat", etc.
	Name       string // nom du fichier
	Path       string // chemin ou URL du fichier
	MimeType   string // type MIME
	UploadedBy uint   // UserID de l'uploader
	User       User   `gorm:"foreignKey:UploadedBy"` // relation vers User
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Template struct {
	ID        uint            `gorm:"primaryKey"`
	CompanyID uint            `gorm:"not null"` // FK vers CompanySettings
	Company   CompanySettings `gorm:"foreignKey:CompanyID"`
	Type      string          // ex: "invoice", "quote", "credit_note"
	Name      string          // nom du template
	Content   string          // contenu HTML ou texte du template
	IsDefault bool            // template par défaut pour l'entreprise
	CreatedAt time.Time
	UpdatedAt time.Time
}
