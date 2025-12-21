package models

import (
	"testing"
)

func TestProduct_GetUserID(t *testing.T) {
	product := &Product{UserID: 42}
	if got := product.GetUserID(); got != 42 {
		t.Errorf("GetUserID() = %d, want 42", got)
	}
}

func TestProduct_PriceWithVAT(t *testing.T) {
	tests := []struct {
		name      string
		unitPrice float64
		vatRate   float64
		want      float64
	}{
		{"20% VAT on €100", 100, 0.20, 120},
		{"10% VAT on €50", 50, 0.10, 55},
		{"0% VAT", 100, 0, 100},
		{"5.5% VAT", 100, 0.055, 105.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Product{UnitPrice: tt.unitPrice, VATRate: tt.vatRate}
			got := p.PriceWithVAT()
			// Use a small epsilon for floating point comparison
			if diff := got - tt.want; diff > 0.001 || diff < -0.001 {
				t.Errorf("PriceWithVAT() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestProduct_VATAmount(t *testing.T) {
	p := &Product{UnitPrice: 100, VATRate: 0.20}
	if got := p.VATAmount(); got != 20 {
		t.Errorf("VATAmount() = %f, want 20", got)
	}
}

func TestProduct_VATRatePercent(t *testing.T) {
	p := &Product{VATRate: 0.20}
	if got := p.VATRatePercent(); got != 20 {
		t.Errorf("VATRatePercent() = %f, want 20", got)
	}
}

func TestClient_GetUserID(t *testing.T) {
	client := &Client{UserID: 123}
	if got := client.GetUserID(); got != 123 {
		t.Errorf("GetUserID() = %d, want 123", got)
	}
}

func TestClient_FullAddress(t *testing.T) {
	tests := []struct {
		name    string
		client  Client
		want    string
	}{
		{
			name: "full address",
			client: Client{
				Address:    "123 Main St",
				PostalCode: "75001",
				City:       "Paris",
				Country:    "France",
			},
			want: "123 Main St\n75001 Paris\nFrance",
		},
		{
			name: "only city",
			client: Client{
				City: "Paris",
			},
			want: "Paris",
		},
		{
			name: "address and city",
			client: Client{
				Address: "123 Main St",
				City:    "Paris",
			},
			want: "123 Main St\nParis",
		},
		{
			name:   "empty",
			client: Client{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.client.FullAddress(); got != tt.want {
				t.Errorf("FullAddress() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInvoice_GetUserID(t *testing.T) {
	invoice := &Invoice{UserID: 456}
	if got := invoice.GetUserID(); got != 456 {
		t.Errorf("GetUserID() = %d, want 456", got)
	}
}

func TestInvoice_Status(t *testing.T) {
	tests := []struct {
		name     string
		status   InvoiceStatus
		isDraft  bool
		isFinal  bool
		canEdit  bool
	}{
		{"draft", InvoiceStatusDraft, true, false, true},
		{"final", InvoiceStatusFinal, false, true, false},
		{"paid", InvoiceStatusPaid, false, true, false},
		{"cancelled", InvoiceStatusCancelled, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &Invoice{Status: tt.status}
			if got := inv.IsDraft(); got != tt.isDraft {
				t.Errorf("IsDraft() = %v, want %v", got, tt.isDraft)
			}
			if got := inv.IsFinal(); got != tt.isFinal {
				t.Errorf("IsFinal() = %v, want %v", got, tt.isFinal)
			}
			if got := inv.CanEdit(); got != tt.canEdit {
				t.Errorf("CanEdit() = %v, want %v", got, tt.canEdit)
			}
		})
	}
}

func TestInvoice_Totals(t *testing.T) {
	invoice := &Invoice{
		Items: []InvoiceItem{
			{Quantity: 2, UnitPrice: 100, VATRate: 0.20},   // HT: 200, VAT: 40
			{Quantity: 1, UnitPrice: 50, VATRate: 0.10},    // HT: 50, VAT: 5
			{Quantity: 3, UnitPrice: 10, VATRate: 0.055},   // HT: 30, VAT: 1.65
		},
	}

	// Total HT should be 200 + 50 + 30 = 280
	if got := invoice.TotalHT(); got != 280 {
		t.Errorf("TotalHT() = %f, want 280", got)
	}

	// Total VAT should be 40 + 5 + 1.65 = 46.65
	expectedVAT := 46.65
	if got := invoice.TotalVAT(); got != expectedVAT {
		t.Errorf("TotalVAT() = %f, want %f", got, expectedVAT)
	}

	// Total TTC should be HT + VAT = 280 + 46.65 = 326.65
	expectedTTC := 326.65
	if got := invoice.TotalTTC(); got != expectedTTC {
		t.Errorf("TotalTTC() = %f, want %f", got, expectedTTC)
	}
}

func TestInvoiceItem_Totals(t *testing.T) {
	item := &InvoiceItem{
		Quantity:  5,
		UnitPrice: 20,
		VATRate:   0.20,
	}

	// HT = 5 * 20 = 100
	if got := item.TotalHT(); got != 100 {
		t.Errorf("TotalHT() = %f, want 100", got)
	}

	// VAT = 100 * 0.20 = 20
	if got := item.TotalVAT(); got != 20 {
		t.Errorf("TotalVAT() = %f, want 20", got)
	}

	// TTC = 100 + 20 = 120
	if got := item.TotalTTC(); got != 120 {
		t.Errorf("TotalTTC() = %f, want 120", got)
	}
}
