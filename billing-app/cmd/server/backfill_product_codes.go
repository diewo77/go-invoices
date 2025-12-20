package main

// Helper: go run ./cmd/server -backfill-product-codes
// Adds codes for existing products where Code is NULL/empty.

import (
	"github.com/diewo77/billing-app/internal/db"
	"github.com/diewo77/billing-app/internal/models"
	"flag"
	"fmt"
	"log"
)

func runBackfillProductCodes() {
	conn, err := db.ConnectAndMigrate()
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	var products []models.Product
	if err := conn.Where("code = '' OR code IS NULL").Find(&products).Error; err != nil {
		log.Fatalf("list products: %v", err)
	}
	updated := 0
	for _, p := range products {
		code := fmt.Sprintf("P%06d", p.ID)
		if err := conn.Model(&models.Product{}).Where("id = ?", p.ID).Update("code", code).Error; err == nil {
			updated++
		}
	}
	log.Printf("Backfill done: %d updated", updated)
}

var backfillFlag = flag.Bool("backfill-product-codes", false, "Backfill missing product codes and exit")

func init() {
	// Nothing else; main() in main.go will parse flags.
}
