package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/diewo77/go-invoices/auth"
	"github.com/diewo77/go-gate"
	"github.com/diewo77/go-invoices/httpx"
	"github.com/diewo77/go-invoices/internal/models"
	"github.com/diewo77/go-invoices/view"
	"gorm.io/gorm"
)

// AdminProfileHandler handles CRUD operations for profiles.
// It allows admins to create, edit, delete profiles and manage their permissions.
type AdminProfileHandler struct {
	DB            *gorm.DB
	CacheResolver *gate.CachedResolver[uint] // To invalidate cache on changes
}

// NewAdminProfileHandler creates a new admin profile handler.
func NewAdminProfileHandler(db *gorm.DB, cacheResolver *gate.CachedResolver[uint]) *AdminProfileHandler {
	return &AdminProfileHandler{DB: db, CacheResolver: cacheResolver}
}

// List displays all profiles with their permission counts.
func (h *AdminProfileHandler) List(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var profiles []models.Profile
	if err := h.DB.Preload("Permissions").Preload("Users").Find(&profiles).Error; err != nil {
		httpx.JSONError(w, http.StatusInternalServerError, "db_error", nil)
		return
	}

	// Check Accept header for JSON response
	if strings.Contains(r.Header.Get("Accept"), "application/json") &&
		!strings.Contains(r.Header.Get("Accept"), "text/html") {
		httpx.JSON(w, http.StatusOK, map[string]any{
			"profiles": profiles,
		})
		return
	}

	view.Render(w, r, "admin/profiles/index.html", map[string]any{
		"Profiles": profiles,
	})
}

// New displays the form to create a new profile.
func (h *AdminProfileHandler) New(w http.ResponseWriter, r *http.Request) {
	view.Render(w, r, "admin/profiles/form.html", map[string]any{
		"IsEdit": false,
	})
}

// Create handles POST to create a new profile.
func (h *AdminProfileHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok || uid == 0 {
		httpx.JSONError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	var profile models.Profile

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
			httpx.JSONError(w, http.StatusBadRequest, "invalid_json", nil)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			httpx.JSONError(w, http.StatusBadRequest, "invalid_form", nil)
			return
		}
		profile.Name = strings.TrimSpace(r.FormValue("name"))
		profile.Description = strings.TrimSpace(r.FormValue("description"))
	}

	// Validation: name is required
	if profile.Name == "" {
		if strings.HasPrefix(contentType, "application/json") {
			httpx.JSONError(w, http.StatusBadRequest, "validation_failed", map[string]string{"name": "required"})
		} else {
			view.Render(w, r, "admin/profiles/form.html", map[string]any{
				"IsEdit":  false,
				"Profile": profile,
				"Errors":  map[string]string{"name": "Le nom est requis"},
			})
		}
		return
	}

	if err := h.DB.Create(&profile).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			if strings.HasPrefix(contentType, "application/json") {
				httpx.JSONError(w, http.StatusConflict, "name_already_exists", nil)
			} else {
				view.Render(w, r, "admin/profiles/form.html", map[string]any{
					"IsEdit":  false,
					"Profile": profile,
					"Errors":  map[string]string{"name": "Ce nom existe déjà"},
				})
			}
			return
		}
		httpx.JSONError(w, http.StatusInternalServerError, "db_error", nil)
		return
	}

	if strings.HasPrefix(contentType, "application/json") {
		httpx.JSON(w, http.StatusCreated, profile)
		return
	}
	http.Redirect(w, r, "/admin/profiles", http.StatusSeeOther)
}

// Edit displays the form to edit an existing profile.
func (h *AdminProfileHandler) Edit(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Redirect(w, r, "/admin/profiles", http.StatusSeeOther)
		return
	}

	var profile models.Profile
	if err := h.DB.First(&profile, id).Error; err != nil {
		http.Redirect(w, r, "/admin/profiles", http.StatusSeeOther)
		return
	}

	view.Render(w, r, "admin/profiles/form.html", map[string]any{
		"IsEdit":  true,
		"Profile": profile,
	})
}

// Update handles POST to update an existing profile.
func (h *AdminProfileHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok || uid == 0 {
		httpx.JSONError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		httpx.JSONError(w, http.StatusBadRequest, "invalid_id", nil)
		return
	}

	var profile models.Profile
	if err := h.DB.First(&profile, id).Error; err != nil {
		httpx.JSONError(w, http.StatusNotFound, "not_found", nil)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
			httpx.JSONError(w, http.StatusBadRequest, "invalid_json", nil)
			return
		}
	} else {
		if err := r.ParseForm(); err != nil {
			httpx.JSONError(w, http.StatusBadRequest, "invalid_form", nil)
			return
		}
		profile.Name = strings.TrimSpace(r.FormValue("name"))
		profile.Description = strings.TrimSpace(r.FormValue("description"))
	}

	if err := h.DB.Save(&profile).Error; err != nil {
		httpx.JSONError(w, http.StatusInternalServerError, "db_error", nil)
		return
	}

	// Invalidate all cache since profile may affect multiple users
	if h.CacheResolver != nil {
		h.CacheResolver.InvalidateAll()
	}

	if strings.HasPrefix(contentType, "application/json") {
		httpx.JSON(w, http.StatusOK, profile)
		return
	}
	http.Redirect(w, r, "/admin/profiles", http.StatusSeeOther)
}

// Delete handles POST to delete a profile.
func (h *AdminProfileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok || uid == 0 {
		httpx.JSONError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		idStr = r.FormValue("id")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		httpx.JSONError(w, http.StatusBadRequest, "invalid_id", nil)
		return
	}

	var profile models.Profile
	if err := h.DB.Preload("Users").First(&profile, id).Error; err != nil {
		httpx.JSONError(w, http.StatusNotFound, "not_found", nil)
		return
	}

	// Cannot delete system profiles (admin, viewer, accountant)
	if profile.IsSystem {
		httpx.JSONError(w, http.StatusForbidden, "cannot_delete_system_profile", nil)
		return
	}

	// Cannot delete if users are assigned to this profile
	if len(profile.Users) > 0 {
		httpx.JSONError(w, http.StatusConflict, "profile_has_users", nil)
		return
	}

	// Delete profile (soft delete via GORM)
	if err := h.DB.Delete(&profile).Error; err != nil {
		httpx.JSONError(w, http.StatusInternalServerError, "db_error", nil)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") ||
		(strings.Contains(r.Header.Get("Accept"), "application/json") &&
			!strings.Contains(r.Header.Get("Accept"), "text/html")) {
		httpx.JSON(w, http.StatusOK, map[string]any{"deleted": id})
		return
	}
	http.Redirect(w, r, "/admin/profiles", http.StatusSeeOther)
}

// EditPermissions displays the permission management page for a profile.
func (h *AdminProfileHandler) EditPermissions(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Redirect(w, r, "/admin/profiles", http.StatusSeeOther)
		return
	}

	var profile models.Profile
	if err := h.DB.Preload("Permissions").First(&profile, id).Error; err != nil {
		http.Redirect(w, r, "/admin/profiles", http.StatusSeeOther)
		return
	}

	var allPermissions []models.Permission
	h.DB.Order("resource_type, action").Find(&allPermissions)

	// Group permissions by resource type for better UI organization
	permsByResource := make(map[string][]models.Permission)
	for _, p := range allPermissions {
		permsByResource[p.ResourceType] = append(permsByResource[p.ResourceType], p)
	}

	// Create a set of current profile permission IDs for easy lookup in template
	currentPermIDs := make(map[uint]bool)
	for _, p := range profile.Permissions {
		currentPermIDs[p.ID] = true
	}

	view.Render(w, r, "admin/profiles/permissions.html", map[string]any{
		"Profile":               profile,
		"PermissionsByResource": permsByResource,
		"CurrentPermissionIDs":  currentPermIDs,
	})
}

// SavePermissions handles POST to save permission changes for a profile.
func (h *AdminProfileHandler) SavePermissions(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok || uid == 0 {
		httpx.JSONError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		httpx.JSONError(w, http.StatusBadRequest, "invalid_id", nil)
		return
	}

	var profile models.Profile
	if err := h.DB.First(&profile, id).Error; err != nil {
		httpx.JSONError(w, http.StatusNotFound, "not_found", nil)
		return
	}

	if err := r.ParseForm(); err != nil {
		httpx.JSONError(w, http.StatusBadRequest, "invalid_form", nil)
		return
	}

	// Get selected permission IDs from form checkboxes
	permissionIDStrs := r.Form["permissions"]
	var permissionIDs []uint
	for _, s := range permissionIDStrs {
		if pid, err := strconv.Atoi(s); err == nil && pid > 0 {
			permissionIDs = append(permissionIDs, uint(pid))
		}
	}

	// Fetch the permissions from DB
	var permissions []models.Permission
	if len(permissionIDs) > 0 {
		h.DB.Where("id IN ?", permissionIDs).Find(&permissions)
	}

	// Replace the profile's permissions (GORM handles the many2many table)
	if err := h.DB.Model(&profile).Association("Permissions").Replace(permissions); err != nil {
		httpx.JSONError(w, http.StatusInternalServerError, "db_error", nil)
		return
	}

	// Invalidate all cache since this profile may affect multiple users
	if h.CacheResolver != nil {
		h.CacheResolver.InvalidateAll()
	}

	http.Redirect(w, r, "/admin/profiles/permissions?id="+strconv.Itoa(id), http.StatusSeeOther)
}

// ListPermissions returns all available permissions (for API use).
func (h *AdminProfileHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	var permissions []models.Permission
	h.DB.Order("resource_type, action").Find(&permissions)
	httpx.JSON(w, http.StatusOK, permissions)
}
