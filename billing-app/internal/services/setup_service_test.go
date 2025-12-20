package services

import (
	"github.com/diewo77/billing-app/internal/models"
	"errors"
	"fmt"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T, name string) *gorm.DB {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.Address{}, &models.User{}, &models.CompanySettings{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestSetupServiceRunSameBilling(t *testing.T) {
	db := setupTestDB(t, t.Name())
	svc := NewSetupService(db)
	// Seed user
	u := models.User{Email: "user@example.com", Password: "hash"}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	out, err := svc.Run(SetupInput{Company: "Acme", Address1: "1 rue", PostalCode: "75000", City: "Paris", Country: "FR", SIRET: "12345678901234", UserID: u.ID})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.AddressID == 0 || out.BillingAddressID == 0 {
		t.Fatalf("expected address IDs set")
	}
	if out.AddressID != out.BillingAddressID {
		t.Fatalf("expected same billing address when not separate")
	}
	var addrCount int64
	if err := db.Model(&models.Address{}).Count(&addrCount).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if addrCount != 1 {
		t.Fatalf("expected 1 address got %d", addrCount)
	}
}

func TestSetupServiceRunSeparateBilling(t *testing.T) {
	db := setupTestDB(t, t.Name())
	svc := NewSetupService(db)
	u := models.User{Email: "user2@example.com", Password: "hash"}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	out, err := svc.Run(SetupInput{Company: "Acme", Address1: "1 rue", PostalCode: "75000", City: "Paris", Country: "FR", SIRET: "12345678901234", UserID: u.ID, BillingAddress1: "2 rue", BillingPostalCode: "69000", BillingCity: "Lyon", BillingCountry: "FR"})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if out.AddressID == out.BillingAddressID {
		t.Fatalf("expected different billing address IDs")
	}
	var addrCount int64
	if err := db.Model(&models.Address{}).Count(&addrCount).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if addrCount != 2 {
		t.Fatalf("expected 2 addresses got %d", addrCount)
	}
}

func TestSetupServiceDuplicateAndIsConfigured(t *testing.T) {
	db := setupTestDB(t, t.Name())
	svc := NewSetupService(db)
	u := models.User{Email: "user3@example.com", Password: "hash"}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	configured, err := svc.IsConfigured()
	if err != nil || configured {
		t.Fatalf("expected not configured, err=%v", err)
	}
	if _, err := svc.Run(SetupInput{Company: "Acme", Address1: "1 rue", PostalCode: "75000", City: "Paris", Country: "FR", SIRET: "12345678901234", UserID: u.ID}); err != nil {
		t.Fatalf("first run err: %v", err)
	}
	configured, err = svc.IsConfigured()
	if err != nil || !configured {
		t.Fatalf("expected configured, err=%v", err)
	}
	if _, err := svc.Run(SetupInput{Company: "Acme2", Address1: "1 rue", PostalCode: "75000", City: "Paris", Country: "FR", SIRET: "12345678901235", UserID: u.ID}); !errors.Is(err, ErrAlreadyConfigured) {
		t.Fatalf("expected ErrAlreadyConfigured got %v", err)
	}
}

func TestSetupServiceMissingUser(t *testing.T) {
	db := setupTestDB(t, t.Name())
	svc := NewSetupService(db)
	if _, err := svc.Run(SetupInput{Company: "Acme", Address1: "1 rue", PostalCode: "75000", City: "Paris", Country: "FR", SIRET: "12345678901234"}); err == nil || err.Error() != "missing_user_id" {
		t.Fatalf("expected missing_user_id err got %v", err)
	}
}
