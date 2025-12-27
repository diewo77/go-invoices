package services

import (
	"github.com/diewo77/go-invoices/internal/models"
	"gorm.io/gorm"
)

type InvoiceService struct {
	db *gorm.DB
}

func NewInvoiceService(db *gorm.DB) *InvoiceService {
	return &InvoiceService{db: db}
}

// ComputeTotals calculates HT, TVA, and TTC for an invoice.
func (s *InvoiceService) ComputeTotals(inv *models.Invoice) (ht, tva, ttc float64) {
	for _, item := range inv.Items {
		ht += item.TotalHT()
		tva += item.TotalVAT()
	}
	ttc = ht + tva
	return
}

// GetRevenue calculates the total revenue from paid invoices for a user.
func (s *InvoiceService) GetRevenue(userID uint) (float64, error) {
	var invoices []models.Invoice
	err := s.db.Where("user_id = ? AND status = ?", userID, models.InvoiceStatusPaid).
		Preload("Items").
		Find(&invoices).Error
	if err != nil {
		return 0, err
	}

	var total float64
	for _, inv := range invoices {
		_, _, ttc := s.ComputeTotals(&inv)
		total += ttc
	}
	return total, nil
}
