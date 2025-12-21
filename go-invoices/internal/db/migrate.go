package db

import (
	"github.com/diewo77/go-invoices/internal/models"
	"gorm.io/gorm"
)

// Migrate runs AutoMigrate for all models.
// Call this at application startup or as part of a migration step.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// Auth & Authorization
		&models.User{},
		&models.Profile{},
		&models.Permission{},
		// Business entities
		&models.CompanySettings{},
		&models.Client{},
		&models.Product{},
		&models.Invoice{},
		&models.InvoiceItem{},
	)
}

// Seed initializes the database with required seed data.
// Should be called after Migrate.
func Seed(db *gorm.DB) error {
	return SeedProfiles(db)
}
