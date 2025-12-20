package db

import (
	"github.com/diewo77/billing-app/internal/models"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	migrate "github.com/golang-migrate/migrate/v4"
	// The following blank imports register the postgres driver and file source for golang-migrate.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectAndMigrate() (*gorm.DB, error) {
	dsn := GetNormalizedDSN()
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_DSN est vide, vérifiez la configuration de l'environnement")
	}
	var db *gorm.DB
	var err error
	logLevel := logger.Silent
	if os.Getenv("DB_DEBUG") == "1" {
		logLevel = logger.Info
	}
	cfg := &gorm.Config{Logger: logger.Default.LogMode(logLevel)}
	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dsn), cfg)
		if err == nil {
			break
		}
		fmt.Println("Retrying DB connection...", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect database after retries: %w", err)
	}

	// Basic connectivity test
	if pingErr := db.Exec("SELECT 1").Error; pingErr != nil {
		return nil, fmt.Errorf("db ping failed: %w", pingErr)
	}

	// Always print masked DSN once for diagnostics (before migrations for visibility)
	masked := dsn
	if strings.Contains(masked, "password=") {
		re := regexp.MustCompile(`(password=)([^\s]+)`)
		masked = re.ReplaceAllString(masked, `${1}***`)
	}
	fmt.Println("[DB] Using DSN:", masked)
	// If MIGRATIONS=1 (or true) we run sql migrations via golang-migrate; otherwise keep old AutoMigrate fallback (dev convenience)
	if v := strings.ToLower(os.Getenv("MIGRATIONS")); v == "1" || v == "true" || v == "yes" {
		if err := runSQLMigrations(dsn); err != nil {
			return nil, fmt.Errorf("sql migrations failed: %w", err)
		}
	} else {
		modelsToMigrate := []interface{}{
			&models.Role{}, &models.User{}, &models.Address{}, &models.CompanySettings{}, &models.UserCompany{}, &models.ProductType{}, &models.UnitType{}, &models.Product{}, &models.Client{}, &models.Invoice{}, &models.InvoiceItem{}, &models.Quote{}, &models.QuoteItem{}, &models.Payment{}, &models.Document{}, &models.Notification{}, &models.AuditLog{}, &models.Template{},
		}
		for _, m := range modelsToMigrate {
			if migErr := db.AutoMigrate(m); migErr != nil {
				fmt.Printf("[DB] AutoMigrate detailed error model=%T type=%T value=%#v\n", m, migErr, migErr)
				return nil, fmt.Errorf("automigrate %T: %w", m, migErr)
			}
		}
	}

	// sanity check: ensure required core tables exist
	for _, table := range []string{"roles", "users", "company_settings"} {
		if !db.Migrator().HasTable(table) {
			return nil, errors.New("missing table after migration: " + table)
		}
	}
	// Seeding only when explicitly requested (e.g. development) via DB_SEED=1|true
	if v := strings.ToLower(os.Getenv("DB_SEED")); v == "1" || v == "true" || v == "yes" {
		seed(db)
	}
	return db, nil
}

func seed(db *gorm.DB) {
	// Product Types
	baseProductTypes := []models.ProductType{
		{Name: "Vente de marchandises", Code: "VM"},
		{Name: "Prestation de services", Code: "PS"},
		{Name: "Abonnement", Code: "SUB"},
	}
	for _, pt := range baseProductTypes {
		var existing models.ProductType
		if err := db.Where("name = ?", pt.Name).First(&existing).Error; err == gorm.ErrRecordNotFound {
			db.Create(&pt)
		}
	}
	// Unit Types
	baseUnitTypes := []models.UnitType{
		{Name: "pièce", Symbol: "pc"},
		{Name: "heure", Symbol: "h"},
		{Name: "kilogramme", Symbol: "kg"},
		{Name: "mètre", Symbol: "m"},
	}
	for _, ut := range baseUnitTypes {
		var existing models.UnitType
		if err := db.Where("name = ?", ut.Name).First(&existing).Error; err == gorm.ErrRecordNotFound {
			db.Create(&ut)
		}
	}
}

// runSQLMigrations executes migrations in ./migrations using golang-migrate file source.
func runSQLMigrations(dsn string) error {
	// golang-migrate expects DSN without gorm specific extras; reuse as-is (URL form supported)
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return err
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
