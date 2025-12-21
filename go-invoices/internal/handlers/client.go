package handlers

import (
	"net/http"
	"strconv"

	"github.com/diewo77/go-invoices/auth"
	"github.com/diewo77/go-invoices/internal/models"
	"github.com/diewo77/go-invoices/validation"
	"github.com/diewo77/go-invoices/view"
	"gorm.io/gorm"
)

type ClientHandler struct {
	db *gorm.DB
}

func NewClientHandler(db *gorm.DB) *ClientHandler {
	return &ClientHandler{db: db}
}

func (h *ClientHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	
	query := r.URL.Query().Get("q")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	var clients []models.Client
	var total int64

	db := h.db.Where("user_id = ?", userID)
	if query != "" {
		db = db.Where("name ILIKE ? OR company ILIKE ?", "%"+query+"%", "%"+query+"%")
	}

	db.Model(&models.Client{}).Count(&total)
	db.Order("name").Limit(limit).Offset(offset).Find(&clients)

	view.Render(w, r, "clients/index.html", map[string]any{
		"Clients": clients,
		"Query":   query,
		"Page":    page,
		"Total":   total,
		"Limit":   limit,
	})
}

func (h *ClientHandler) New(w http.ResponseWriter, r *http.Request) {
	view.Render(w, r, "clients/new.html", nil)
}

func (h *ClientHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())

	client := models.Client{
		UserID:     userID,
		Name:       r.FormValue("name"),
		Email:      r.FormValue("email"),
		Phone:      r.FormValue("phone"),
		Company:    r.FormValue("company"),
		Address:    r.FormValue("address"),
		City:       r.FormValue("city"),
		PostalCode: r.FormValue("postal_code"),
		Country:    r.FormValue("country"),
		SIRET:      r.FormValue("siret"),
		VATNumber:  r.FormValue("vat_number"),
	}

	v := make(validation.Violations)
	validation.Required("name", client.Name, v)

	if !v.Empty() {
		view.Render(w, r, "clients/new.html", map[string]any{
			"Client": client,
			"Errors": v,
		})
		return
	}

	if err := h.db.Create(&client).Error; err != nil {
		view.Render(w, r, "clients/new.html", map[string]any{
			"Client": client,
			"Error":  "Failed to create client",
		})
		return
	}

	http.Redirect(w, r, "/clients", http.StatusSeeOther)
}

func (h *ClientHandler) View(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	var client models.Client
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&client).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	view.Render(w, r, "clients/view.html", map[string]any{
		"Client": client,
	})
}

func (h *ClientHandler) Edit(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	var client models.Client
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&client).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	view.Render(w, r, "clients/edit.html", map[string]any{
		"Client": client,
	})
}

func (h *ClientHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	var client models.Client
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&client).Error; err != nil {
		http.NotFound(w, r)
		return
	}

	client.Name = r.FormValue("name")
	client.Email = r.FormValue("email")
	client.Phone = r.FormValue("phone")
	client.Company = r.FormValue("company")
	client.Address = r.FormValue("address")
	client.City = r.FormValue("city")
	client.PostalCode = r.FormValue("postal_code")
	client.Country = r.FormValue("country")
	client.SIRET = r.FormValue("siret")
	client.VATNumber = r.FormValue("vat_number")

	v := make(validation.Violations)
	validation.Required("name", client.Name, v)

	if !v.Empty() {
		view.Render(w, r, "clients/edit.html", map[string]any{
			"Client": client,
			"Errors": v,
		})
		return
	}

	if err := h.db.Save(&client).Error; err != nil {
		view.Render(w, r, "clients/edit.html", map[string]any{
			"Client": client,
			"Error":  "Failed to update client",
		})
		return
	}

	http.Redirect(w, r, "/clients/"+id, http.StatusSeeOther)
}

func (h *ClientHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	if err := h.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Client{}).Error; err != nil {
		http.Error(w, "Failed to delete client", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/clients", http.StatusSeeOther)
}
