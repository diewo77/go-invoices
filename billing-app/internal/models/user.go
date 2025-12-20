package models

import "time"

// User & auth related models
type User struct {
	ID          uint    `gorm:"primaryKey"`
	Email       string  `gorm:"unique;not null;index"`
	Password    string  `gorm:"not null"` // hashé
	Nom         string  `gorm:"index"`
	Prenom      string  `gorm:"index"`
	AddressID   uint    // clé étrangère vers Address
	Address     Address `gorm:"foreignKey:AddressID"`
	RoleID      uint    // clé étrangère vers Role
	Role        Role    `gorm:"foreignKey:RoleID"`
	Permissions string  // permissions personnalisées (ex: JSON ou CSV)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Role struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"unique;not null"` // admin, manager, user
	Description string // optionnel
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Notification struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`          // destinataire
	User      User      `gorm:"foreignKey:UserID"` // relation vers User
	Type      string    // ex: "mail", "dashboard", "sms"
	Title     string    // titre ou sujet
	Message   string    // contenu
	Read      bool      // notification lue ou non
	SentAt    time.Time // date d'envoi
	CreatedAt time.Time
	UpdatedAt time.Time
}
