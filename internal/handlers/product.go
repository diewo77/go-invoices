package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/diewo77/go-invoices/auth"
	"github.com/diewo77/go-invoices/internal/models"
	"github.com/diewo77/go-invoices/validation"
	"github.com/diewo77/go-invoices/view"
	"gorm.io/gorm"
)

type ProductHandler struct {
	db *gorm.DB
}

func NewProductHandler(db *gorm.DB) *ProductHandler {
	return &ProductHandler{db: db}
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	query := r.URL.Query().Get("q")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	var products []models.Product
	var total int64

	db := h.db.Where("user_id = ?", userID)
	if query != "" {
		db = db.Where("name ILIKE ? OR code ILIKE ?", "%"+query+"%", "%"+query+"%")
	}

	db.Model(&models.Product{}).Count(&total)
	db.Order("name").Limit(limit).Offset(offset).Find(&products)

	view.Render(w, r, "products/index.html", map[string]any{
		"Products": products,
		"Query":    query,
		"Page":     page,
		"Total":    total,
		"Limit":    limit,
	})
}

func (h *ProductHandler) New(w http.ResponseWriter, r *http.Request) {
	view.Render(w, r, "products/new.html", nil)
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	unitPrice, _ := strconv.ParseFloat(r.FormValue("unit_price"), 64)
	vatRate, _ := strconv.ParseFloat(r.FormValue("vat_rate"), 64)

	// Convert VAT rate to decimal if it's > 1 (e.g. 20 -> 0.20)
	if vatRate > 1 {
		vatRate = vatRate / 100
	}

	product := models.Product{
		UserID:      userID,
		Code:        strings.ToUpper(r.FormValue("code")),
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		UnitPrice:   unitPrice,
		Unit:        r.FormValue("unit"),
		VATRate:     vatRate,
		Category:    r.FormValue("category"),
		IsActive:    r.FormValue("is_active") == "on",
	}

	v := make(validation.Violations)
	validation.Required("code", product.Code, v)
	validation.Required("name", product.Name, v)
	validation.PositiveFloat("unit_price", product.UnitPrice, v)

	if !v.Empty() {
		view.Render(w, r, "products/new.html", map[string]any{
			"Product": product,
			"Errors":  v,
		})
		return
	}

	if err := h.db.Create(&product).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			v["code"] = "code_already_exists"
			view.Render(w, r, "products/new.html", map[string]any{
				"Product": product,
				"Errors":  v,
			})
			return
		}
		view.Render(w, r, "products/new.html", map[string]any{
			"Product": product,
			"Error":   "Failed to create product",
		})
		return
	}

	http.Redirect(w, r, "/products", http.StatusSeeOther)
}

func (h *ProductHandler) View(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	var product models.Product
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&product).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	view.Render(w, r, "products/view.html", map[string]any{
		"Product": product,
	})
}

func (h *ProductHandler) Edit(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	var product models.Product
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&product).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	view.Render(w, r, "products/edit.html", map[string]any{
		"Product": product,
	})
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	var product models.Product
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&product).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	unitPrice, _ := strconv.ParseFloat(r.FormValue("unit_price"), 64)
	vatRate, _ := strconv.ParseFloat(r.FormValue("vat_rate"), 64)

	if vatRate > 1 {
		vatRate = vatRate / 100
	}

	product.Name = r.FormValue("name")
	product.Description = r.FormValue("description")
	product.UnitPrice = unitPrice
	product.Unit = r.FormValue("unit")
	product.VATRate = vatRate
	product.Category = r.FormValue("category")
	product.IsActive = r.FormValue("is_active") == "on"

	v := make(validation.Violations)
	validation.Required("name", product.Name, v)
	validation.PositiveFloat("unit_price", product.UnitPrice, v)

	if !v.Empty() {
		view.Render(w, r, "products/edit.html", map[string]any{
			"Product": product,
			"Errors":  v,
		})
		return
	}

	if err := h.db.Save(&product).Error; err != nil {
		view.Render(w, r, "products/edit.html", map[string]any{
			"Product": product,
			"Error":   "Failed to update product",
		})
		return
	}

	http.Redirect(w, r, "/products/"+id, http.StatusSeeOther)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	if err := h.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Product{}).Error; err != nil {
		http.Error(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/products", http.StatusSeeOther)
}
