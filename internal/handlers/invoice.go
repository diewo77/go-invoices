package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/diewo77/go-invoices/auth"
	"github.com/diewo77/go-invoices/internal/models"
	"github.com/diewo77/go-invoices/validation"
	"github.com/diewo77/go-invoices/view"
	"github.com/diewo77/go-pdf"
	"gorm.io/gorm"
)

type InvoiceHandler struct {
	db *gorm.DB
}

func NewInvoiceHandler(db *gorm.DB) *InvoiceHandler {
	return &InvoiceHandler{db: db}
}

func (h *InvoiceHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	query := r.URL.Query().Get("q")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	var invoices []models.Invoice
	var total int64

	db := h.db.Where("user_id = ?", userID).Preload("Client")
	if query != "" {
		db = db.Where("number ILIKE ? OR reference ILIKE ?", "%"+query+"%", "%"+query+"%")
	}

	db.Model(&models.Invoice{}).Count(&total)
	db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&invoices)

	view.Render(w, r, "invoices/index.html", map[string]any{
		"Invoices": invoices,
		"Query":    query,
		"Page":     page,
		"Total":    total,
		"Limit":    limit,
	})
}

func (h *InvoiceHandler) New(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	var clients []models.Client
	h.db.Where("user_id = ?", userID).Order("name").Find(&clients)

	var products []models.Product
	h.db.Where("user_id = ?", userID).Order("name").Find(&products)

	view.Render(w, r, "invoices/new.html", map[string]any{
		"Clients":  clients,
		"Products": products,
	})
}

func (h *InvoiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	clientID, _ := strconv.ParseUint(r.FormValue("client_id"), 10, 32)
	issueDate, _ := time.Parse("2006-01-02", r.FormValue("issue_date"))
	dueDate, _ := time.Parse("2006-01-02", r.FormValue("due_date"))

	invoice := models.Invoice{
		UserID:       userID,
		ClientID:     uint(clientID),
		IssueDate:    issueDate,
		DueDate:      dueDate,
		Reference:    r.FormValue("reference"),
		Notes:        r.FormValue("notes"),
		PaymentTerms: r.FormValue("payment_terms"),
		Status:       models.InvoiceStatusDraft,
	}

	// Generate a temporary number if empty
	if invoice.Number == "" {
		invoice.Number = "DRAFT-" + time.Now().Format("20060102-150405")
	}

	v := make(validation.Violations)
	if invoice.ClientID == 0 {
		v["client_id"] = "required"
	}

	if !v.Empty() {
		h.New(w, r) // Re-render with errors (simplified)
		return
	}

	if err := h.db.Create(&invoice).Error; err != nil {
		http.Error(w, "Failed to create invoice", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/invoices/"+strconv.Itoa(int(invoice.ID))+"/edit", http.StatusSeeOther)
}

func (h *InvoiceHandler) View(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	var invoice models.Invoice
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).Preload("Client").Preload("Items.Product").First(&invoice).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	view.Render(w, r, "invoices/view.html", map[string]any{
		"Invoice": invoice,
	})
}

func (h *InvoiceHandler) Edit(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	var invoice models.Invoice
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).Preload("Client").Preload("Items.Product").First(&invoice).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	if !invoice.CanEdit() {
		http.Redirect(w, r, "/invoices/"+id, http.StatusSeeOther)
		return
	}

	var clients []models.Client
	h.db.Where("user_id = ?", userID).Order("name").Find(&clients)

	var products []models.Product
	h.db.Where("user_id = ?", userID).Order("name").Find(&products)

	view.Render(w, r, "invoices/edit.html", map[string]any{
		"Invoice":  invoice,
		"Clients":  clients,
		"Products": products,
	})
}

func (h *InvoiceHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	var invoice models.Invoice
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&invoice).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	if !invoice.CanEdit() {
		http.Error(w, "Cannot edit finalized invoice", http.StatusForbidden)
		return
	}

	clientID, _ := strconv.ParseUint(r.FormValue("client_id"), 10, 32)
	issueDate, _ := time.Parse("2006-01-02", r.FormValue("issue_date"))
	dueDate, _ := time.Parse("2006-01-02", r.FormValue("due_date"))

	invoice.ClientID = uint(clientID)
	invoice.IssueDate = issueDate
	invoice.DueDate = dueDate
	invoice.Reference = r.FormValue("reference")
	invoice.Notes = r.FormValue("notes")
	invoice.PaymentTerms = r.FormValue("payment_terms")

	if err := h.db.Save(&invoice).Error; err != nil {
		http.Error(w, "Failed to update invoice", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/invoices/"+id, http.StatusSeeOther)
}

func (h *InvoiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	var invoice models.Invoice
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&invoice).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	if !invoice.CanEdit() {
		http.Error(w, "Cannot delete finalized invoice", http.StatusForbidden)
		return
	}

	if err := h.db.Delete(&invoice).Error; err != nil {
		http.Error(w, "Failed to delete invoice", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/invoices", http.StatusSeeOther)
}

func (h *InvoiceHandler) Finalize(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	var invoice models.Invoice
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).Preload("Items").First(&invoice).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	if !invoice.IsDraft() {
		http.Redirect(w, r, "/invoices/"+id, http.StatusSeeOther)
		return
	}

	if len(invoice.Items) == 0 {
		http.Error(w, "Cannot finalize invoice with no items", http.StatusBadRequest)
		return
	}

	// Generate final number
	var count int64
	h.db.Model(&models.Invoice{}).Where("user_id = ? AND status != ?", userID, models.InvoiceStatusDraft).Count(&count)
	invoice.Number = time.Now().Format("2006") + "-" + strconv.FormatInt(count+1, 10)
	invoice.Status = models.InvoiceStatusFinal

	if err := h.db.Save(&invoice).Error; err != nil {
		http.Error(w, "Failed to finalize invoice", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/invoices/"+id, http.StatusSeeOther)
}

func (h *InvoiceHandler) PDF(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	var invoice models.Invoice
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).
		Preload("Client").
		Preload("Items.Product").
		First(&invoice).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	var company models.CompanySettings
	if err := h.db.Where("user_id = ?", userID).First(&company).Error; err != nil {
		// Fallback or error if company settings not found
		company.Name = "My Company" // Minimal fallback
	}

	// Map to PDF data
	pdfData := pdf.InvoiceData{
		InvoiceNumber: invoice.Number,
		Date:          invoice.IssueDate.Format("02/01/2006"),
		DueDate:       invoice.DueDate.Format("02/01/2006"),
		Total:         invoice.TotalHT(),
		VAT:           invoice.TotalVAT(),
		GrandTotal:    invoice.TotalTTC(),
		Client: pdf.ClientData{
			Name:    invoice.Client.Name,
			Address: invoice.Client.FullAddress(),
			Email:   invoice.Client.Email,
		},
		Company: pdf.CompanyData{
			Name:    company.Name,
			Address: company.Address + "\n" + company.PostalCode + " " + company.City,
			LogoURL: company.LogoURL,
		},
	}

	for _, item := range invoice.Items {
		pdfData.Items = append(pdfData.Items, pdf.InvoiceItem{
			Description: item.Description,
			Quantity:    int(item.Quantity),
			UnitPrice:   item.UnitPrice,
			Total:       item.TotalHT(),
		})
	}

	pdfBytes, err := pdf.InvoicePDF(pdfData)
	if err != nil {
		http.Error(w, "Failed to generate PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"invoice-%s.pdf\"", invoice.Number))
	w.Write(pdfBytes)
}

func (h *InvoiceHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	var invoice models.Invoice
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&invoice).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	if !invoice.CanEdit() {
		http.Error(w, "Cannot edit finalized invoice", http.StatusForbidden)
		return
	}

	productID, _ := strconv.ParseUint(r.FormValue("product_id"), 10, 32)
	quantity, _ := strconv.ParseFloat(r.FormValue("quantity"), 64)

	var product models.Product
	if err := h.db.Where("id = ? AND user_id = ?", productID, userID).First(&product).Error; err != nil {
		http.Error(w, "Product not found", http.StatusBadRequest)
		return
	}

	item := models.InvoiceItem{
		InvoiceID:   invoice.ID,
		ProductID:   &product.ID,
		Description: product.Name,
		Quantity:    quantity,
		UnitPrice:   product.UnitPrice,
		VATRate:     product.VATRate,
	}

	if err := h.db.Create(&item).Error; err != nil {
		http.Error(w, "Failed to add item", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/invoices/"+id+"/edit", http.StatusSeeOther)
}

func (h *InvoiceHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")
	itemID := r.PathValue("item_id")

	var invoice models.Invoice
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&invoice).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	if !invoice.CanEdit() {
		http.Error(w, "Cannot edit finalized invoice", http.StatusForbidden)
		return
	}

	if err := h.db.Where("id = ? AND invoice_id = ?", itemID, invoice.ID).Delete(&models.InvoiceItem{}).Error; err != nil {
		http.Error(w, "Failed to remove item", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/invoices/"+id+"/edit", http.StatusSeeOther)
}
