package db

import (
	"github.com/diewo77/billing-app/internal/models"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSeedIdempotent(t *testing.T) {
	d, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := d.AutoMigrate(&models.ProductType{}, &models.UnitType{}); err != nil {
		t.Fatal(err)
	}
	seed(d)
	seed(d)
	var ptCount, utCount int64
	d.Model(&models.ProductType{}).Count(&ptCount)
	d.Model(&models.UnitType{}).Count(&utCount)
	if ptCount < 2 {
		t.Fatalf("expected at least 2 product types got %d", ptCount)
	}
	if utCount < 2 {
		t.Fatalf("expected at least 2 unit types got %d", utCount)
	}
	// Ensure baseline entries exist exactly once (idempotency)
	var c1, c2 int64
	d.Model(&models.ProductType{}).Where("name = ?", "Vente de marchandises").Count(&c1)
	d.Model(&models.ProductType{}).Where("name = ?", "Prestation de services").Count(&c2)
	if c1 != 1 || c2 != 1 {
		t.Fatalf("baseline product types duplicated or missing: VM=%d PS=%d", c1, c2)
	}
}
