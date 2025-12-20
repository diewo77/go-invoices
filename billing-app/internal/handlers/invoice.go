package handlers

import (
	"github.com/diewo77/billing-app/auth"
	"github.com/diewo77/billing-app/httpx"
	"github.com/diewo77/billing-app/internal/models"
	pdfgen "github.com/diewo77/billing-app/pdf"
	"github.com/diewo77/billing-app/internal/services"
	"github.com/diewo77/billing-app/view"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// InvoiceHandler mirrors the dual-format pattern used elsewhere.
type InvoiceHandler struct {
	DB  *gorm.DB
	Svc *services.InvoiceService
}

func NewInvoiceHandler(db *gorm.DB, svc *services.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{DB: db, Svc: svc}
}

// List: GET /invoices – HTML or JSON
func (h *InvoiceHandler) List(w http.ResponseWriter, r *http.Request) {
	// Scope to first company for now (until multi-company user context is wired)
	var company models.CompanySettings
	if err := h.DB.Select("id").First(&company).Error; err != nil {
		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "application/json") && !strings.Contains(accept, "text/html") {
			httpx.JSON(w, http.StatusOK, map[string]any{"items": []models.Invoice{}, "total": 0, "limit": 50, "offset": 0})
			return
		}
		_ = view.Render(w, r, "invoices.html", map[string]any{"Invoices": []models.Invoice{}, "NoCompany": true})
		return
	}
	// Pagination
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	offset := 0
	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 1 {
			offset = (n - 1) * limit
		}
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	dbq := h.DB.Where("company_id = ?", company.ID)
	if q != "" {
		safe := regexp.MustCompile(`[^a-zA-Z0-9 \-_]`).ReplaceAllString(q, "")
		like := "%" + strings.ToLower(safe) + "%"
		dbq = dbq.Where("lower(status) LIKE ?", like)
	}
	var total int64
	dbq.Model(&models.Invoice{}).Count(&total)
	var invs []models.Invoice
	if err := dbq.Preload("Items.Product").Order("id desc").Limit(limit).Offset(offset).Find(&invs).Error; err != nil {
		httpx.JSONError(w, http.StatusInternalServerError, "failed_to_list_invoices", nil)
		return
	}
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") && !strings.Contains(accept, "text/html") {
		httpx.JSON(w, http.StatusOK, map[string]any{"items": invs, "total": total, "limit": limit, "offset": offset})
		return
	}
	_ = view.Render(w, r, "invoices.html", map[string]any{"Invoices": invs, "Total": total, "PageSize": limit, "Query": q})
}

// Create: POST /invoices – JSON or form
func (h *InvoiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserIDFromContext(r.Context()); !ok {
		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "application/json") && !strings.Contains(accept, "text/html") {
			httpx.JSONError(w, http.StatusUnauthorized, "unauthorized", nil)
			return
		}
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	// Company scope
	var company models.CompanySettings
	if err := h.DB.Select("id").First(&company).Error; err != nil {
		httpx.JSONError(w, http.StatusBadRequest, "company_not_configured", nil)
		return
	}
	type itemReq struct {
		ProductID uint `json:"product_id"`
		Quantity  int  `json:"quantity"`
	}
	type createReq struct {
		ClientID uint      `json:"client_id"`
		Items    []itemReq `json:"items"`
	}
	var req createReq
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.JSONError(w, http.StatusBadRequest, "invalid_json", nil)
			return
		}
	} else {
		// Basic form parsing: items[].product_id, items[].quantity not supported in depth here; accept single item
		if err := r.ParseForm(); err == nil {
			if v := r.Form.Get("client_id"); v != "" {
				if id, err := strconv.Atoi(v); err == nil {
					req.ClientID = uint(id)
				}
			}
			if v := r.Form.Get("product_id"); v != "" {
				pid, _ := strconv.Atoi(v)
				qty := 1
				if qv := r.Form.Get("quantity"); qv != "" {
					if n, err := strconv.Atoi(qv); err == nil {
						qty = n
					}
				}
				req.Items = []itemReq{{ProductID: uint(pid), Quantity: qty}}
			}
		}
	}
	if req.ClientID == 0 || len(req.Items) == 0 {
		httpx.JSONError(w, http.StatusBadRequest, "validation_failed", map[string]string{"client_id": "required", "items": "required"})
		return
	}
	// Validate items and load products
	var products []models.Product
	productIDs := make([]uint, 0, len(req.Items))
	for _, it := range req.Items {
		if it.ProductID == 0 || it.Quantity <= 0 {
			httpx.JSONError(w, http.StatusBadRequest, "validation_failed", map[string]string{"items": "invalid_product_or_quantity"})
			return
		}
		productIDs = append(productIDs, it.ProductID)
	}
	if err := h.DB.Where("id IN ? AND company_id = ? AND deleted_at IS NULL", productIDs, company.ID).Find(&products).Error; err != nil {
		httpx.JSONError(w, http.StatusBadRequest, "invalid_products", nil)
		return
	}
	if len(products) != len(productIDs) {
		httpx.JSONError(w, http.StatusBadRequest, "invalid_products", nil)
		return
	}
	// Build invoice + items
	inv := models.Invoice{Status: "draft", CompanyID: company.ID, ClientID: req.ClientID}
	items := make([]models.InvoiceItem, 0, len(req.Items))
	// map products by id
	prodByID := map[uint]models.Product{}
	for _, p := range products {
		prodByID[p.ID] = p
	}
	for _, it := range req.Items {
		p := prodByID[it.ProductID]
		items = append(items, models.InvoiceItem{ProductID: p.ID, Quantity: it.Quantity})
	}
	// Persist in transaction
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&inv).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].InvoiceID = inv.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		httpx.JSONError(w, http.StatusInternalServerError, "failed_to_create_invoice", nil)
		return
	}
	// Reload with products for totals
	if err := h.DB.Preload("Items.Product").First(&inv, inv.ID).Error; err != nil {
		httpx.JSONError(w, http.StatusInternalServerError, "failed_to_load_invoice", nil)
		return
	}
	ht, tva, ttc := h.Svc.ComputeTotals(&inv)
	httpx.JSON(w, http.StatusCreated, map[string]any{"id": inv.ID, "status": inv.Status, "ht": ht, "tva": tva, "ttc": ttc})
}

// Finalize: POST /invoices/finalize?id=...
func (h *InvoiceHandler) Finalize(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		httpx.JSONError(w, http.StatusBadRequest, "missing_id", nil)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		httpx.JSONError(w, http.StatusBadRequest, "invalid_id", nil)
		return
	}
	var inv models.Invoice
	if err := h.DB.Preload("Items.Product").First(&inv, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpx.JSONError(w, http.StatusNotFound, "not_found", nil)
			return
		}
		httpx.JSONError(w, http.StatusInternalServerError, "failed_to_load_invoice", nil)
		return
	}
	if len(inv.Items) == 0 {
		httpx.JSONError(w, http.StatusBadRequest, "empty_invoice", nil)
		return
	}
	// Prevent finalize if any product soft-deleted now
	var cnt int64
	if err := h.DB.Table("invoice_items ii").Joins("JOIN products p ON p.id = ii.product_id").Where("ii.invoice_id = ? AND p.deleted_at IS NOT NULL", inv.ID).Count(&cnt).Error; err == nil && cnt > 0 {
		httpx.JSONError(w, http.StatusBadRequest, "contains_deleted_products", nil)
		return
	}
	if inv.Status != "final" {
		if err := h.DB.Model(&inv).Update("status", "final").Error; err != nil {
			httpx.JSONError(w, http.StatusInternalServerError, "failed_to_finalize", nil)
			return
		}
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"status": "finalized"})
}

// PDF: GET /invoices/pdf?id=...
func (h *InvoiceHandler) PDF(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		httpx.JSONError(w, http.StatusBadRequest, "invalid_id", nil)
		return
	}
	var inv models.Invoice
	if err := h.DB.Preload("Items.Product").First(&inv, id).Error; err != nil {
		httpx.JSONError(w, http.StatusNotFound, "not_found", nil)
		return
	}
	var company models.CompanySettings
	if err := h.DB.First(&company, inv.CompanyID).Error; err != nil {
		httpx.JSONError(w, http.StatusInternalServerError, "failed_to_load_company", nil)
		return
	}
	var client models.Client
	if err := h.DB.First(&client, inv.ClientID).Error; err != nil {
		httpx.JSONError(w, http.StatusInternalServerError, "failed_to_load_client", nil)
		return
	}

	var items []pdfgen.InvoiceItem
	var totalHT, totalVAT float64

	for _, item := range inv.Items {
		lineTotal := float64(item.Quantity) * item.Product.UnitPrice
		items = append(items, pdfgen.InvoiceItem{
			Description: item.Product.Name,
			Quantity:    item.Quantity,
			UnitPrice:   item.Product.UnitPrice,
			Total:       lineTotal,
		})
		totalHT += lineTotal
		totalVAT += lineTotal * item.Product.VATRate
	}

	pdfData := pdfgen.InvoiceData{
		InvoiceNumber: strconv.Itoa(int(inv.ID)),
		Date:          inv.CreatedAt.Format("2006-01-02"),
		DueDate:       inv.CreatedAt.AddDate(0, 1, 0).Format("2006-01-02"),
		Items:         items,
		Total:         totalHT,
		VAT:           totalVAT,
		GrandTotal:    totalHT + totalVAT,
		Client: pdfgen.ClientData{
			Name:    client.Nom,
			Address: "", // TODO: Format address
			Email:   client.Email,
		},
		Company: pdfgen.CompanyData{
			Name:    company.RaisonSociale,
			Address: "", // TODO: Format address
			LogoURL: "",
		},
	}

	data, genErr := pdfgen.InvoicePDF(pdfData)
	if genErr != nil {
		httpx.JSONError(w, http.StatusInternalServerError, "pdf_generation_failed", nil)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"invoice-"+strconv.Itoa(int(inv.ID))+".pdf\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
