package models

import "time"

// Audit logging
type AuditLog struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      // qui a fait la modification
	EntityType string    // ex: "Product", "Invoice", "Client", etc.
	EntityID   uint      // ID de l'entité modifiée
	Action     string    // ex: "create", "update", "delete"
	Field      string    // champ modifié (optionnel)
	OldValue   string    // ancienne valeur (optionnel)
	NewValue   string    // nouvelle valeur (optionnel)
	CreatedAt  time.Time // quand
}
