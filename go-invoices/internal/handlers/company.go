package handlers

import (
	"net/http"

	"github.com/diewo77/go-invoices/auth"
	"github.com/diewo77/go-invoices/internal/models"
	"github.com/diewo77/go-invoices/view"
	"gorm.io/gorm"
)

type CompanyHandler struct {
	db *gorm.DB
}

func NewCompanyHandler(db *gorm.DB) *CompanyHandler {
	return &CompanyHandler{db: db}
}

// Edit shows the company settings form.
func (h *CompanyHandler) Edit(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	if userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var settings models.CompanySettings
	err := h.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// If not found, we'll just show an empty form (or default values)
	if err == gorm.ErrRecordNotFound {
		settings.UserID = userID
	}

	view.Render(w, r, "company/edit.html", map[string]any{
		"Settings": settings,
	})
}

// Update saves the company settings.
func (h *CompanyHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	if userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var settings models.CompanySettings
	err := h.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err == gorm.ErrRecordNotFound {
		settings.UserID = userID
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	settings.Name = r.FormValue("name")
	settings.Email = r.FormValue("email")
	settings.Phone = r.FormValue("phone")
	settings.Website = r.FormValue("website")
	settings.Address = r.FormValue("address")
	settings.City = r.FormValue("city")
	settings.PostalCode = r.FormValue("postal_code")
	settings.Country = r.FormValue("country")
	settings.SIRET = r.FormValue("siret")
	settings.VATNumber = r.FormValue("vat_number")
	settings.RCS = r.FormValue("rcs")
	settings.Capital = r.FormValue("capital")

	if err := h.db.Save(&settings).Error; err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set flash message (if we had a helper, but let's just redirect for now)
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}
