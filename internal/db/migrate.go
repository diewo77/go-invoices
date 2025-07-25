package db

import (
	"fmt"
	"time"

	"github.com/diewo77/billing-app/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectAndMigrate essaie de se connecter et d’appliquer les migrations GORM.
func ConnectAndMigrate(dsn string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	// Retry simple pour laisser le temps à Postgres de démarrer
	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		fmt.Printf("Tentative %d/5 échouée, retry...\n", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("connexion BDD échouée : %w", err)
	}

	// Migrations
	if err := db.AutoMigrate(
		&models.Product{},
		&models.Invoice{},
		&models.InvoiceItem{},
	); err != nil {
		return nil, fmt.Errorf("migrations échouées : %w", err)
	}

	return db, nil
}
