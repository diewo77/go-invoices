package services

import (
	"github.com/diewo77/billing-app/internal/models"
)

// InvoiceService encapsulates invoice-related business logic.
// Keep DB access in handlers or add a *gorm.DB here later when needed.
type InvoiceService struct{}

func NewInvoiceService() *InvoiceService { return &InvoiceService{} }

// ComputeTotals computes HT, TVA, and TTC amounts for an invoice based on its items.
// Assumes each InvoiceItem has Product preloaded with UnitPrice and VATRate (0..1).
func (s *InvoiceService) ComputeTotals(inv *models.Invoice) (ht, tva, ttc float64) {
	if inv == nil {
		return 0, 0, 0
	}
	for _, it := range inv.Items {
		qty := float64(it.Quantity)
		price := it.Product.UnitPrice
		lineHT := qty * price
		// Apply optional item-level discount (Remise) if set (>0) as absolute amount capped at line amount.
		if it.Remise > 0 && it.Remise < lineHT {
			lineHT -= it.Remise
		}
		ht += lineHT
		rate := it.Product.VATRate
		if rate < 0 {
			rate = 0
		}
		tva += lineHT * rate
	}
	// Apply invoice-level adjustments
	if inv.Remise > 0 && inv.Remise < ht {
		ht -= inv.Remise
	}
	if inv.Avoir > 0 && inv.Avoir < ht {
		ht -= inv.Avoir
	}
	// Acompte is a prepayment and does not change TTC; tracked for display/payment status.
	ttc = ht + tva
	return ht, tva, ttc
}
