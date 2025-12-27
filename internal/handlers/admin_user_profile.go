package handlers

import (
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

// AdminUserProfileHandler handles user profile assignment.
// It allows admins to view users and assign them to profiles.
type AdminUserProfileHandler struct {
	DB            *gorm.DB
	CacheResolver *gate.CachedResolver[uint] // To invalidate cache on changes
}

// NewAdminUserProfileHandler creates a new admin user profile handler.
func NewAdminUserProfileHandler(db *gorm.DB, cacheResolver *gate.CachedResolver[uint]) *AdminUserProfileHandler {
	return &AdminUserProfileHandler{DB: db, CacheResolver: cacheResolver}
}

// List displays all users with their profile assignments.
func (h *AdminUserProfileHandler) List(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var users []models.User
	if err := h.DB.Preload("Profile").Find(&users).Error; err != nil {
		httpx.JSONError(w, http.StatusInternalServerError, "db_error", nil)
		return
	}

	var profiles []models.Profile
	h.DB.Find(&profiles)

	// Check Accept header for JSON response
	if strings.Contains(r.Header.Get("Accept"), "application/json") &&
		!strings.Contains(r.Header.Get("Accept"), "text/html") {
		httpx.JSON(w, http.StatusOK, map[string]any{
			"users":    users,
			"profiles": profiles,
		})
		return
	}

	view.Render(w, r, "admin/users/index.html", map[string]any{
		"Users":    users,
		"Profiles": profiles,
	})
}

// AssignProfile handles POST to assign a profile to a user.
func (h *AdminUserProfileHandler) AssignProfile(w http.ResponseWriter, r *http.Request) {
	currentUID, ok := auth.UserIDFromContext(r.Context())
	if !ok || currentUID == 0 {
		httpx.JSONError(w, http.StatusUnauthorized, "unauthorized", nil)
		return
	}

	if err := r.ParseForm(); err != nil {
		httpx.JSONError(w, http.StatusBadRequest, "invalid_form", nil)
		return
	}

	userIDStr := r.FormValue("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		httpx.JSONError(w, http.StatusBadRequest, "invalid_user_id", nil)
		return
	}

	profileIDStr := r.FormValue("profile_id")
	var profileID *uint
	if profileIDStr != "" && profileIDStr != "0" {
		pid, err := strconv.Atoi(profileIDStr)
		if err != nil || pid <= 0 {
			httpx.JSONError(w, http.StatusBadRequest, "invalid_profile_id", nil)
			return
		}
		pidUint := uint(pid)
		profileID = &pidUint

		// Verify profile exists
		var profile models.Profile
		if err := h.DB.First(&profile, pid).Error; err != nil {
			httpx.JSONError(w, http.StatusNotFound, "profile_not_found", nil)
			return
		}
	}

	// Update the user's profile
	if err := h.DB.Model(&models.User{}).Where("id = ?", userID).Update("profile_id", profileID).Error; err != nil {
		httpx.JSONError(w, http.StatusInternalServerError, "db_error", nil)
		return
	}

	// Invalidate cache for this specific user
	if h.CacheResolver != nil {
		h.CacheResolver.Invalidate(uint(userID))
	}

	// Check Accept header for JSON response
	if strings.Contains(r.Header.Get("Accept"), "application/json") &&
		!strings.Contains(r.Header.Get("Accept"), "text/html") {
		httpx.JSON(w, http.StatusOK, map[string]any{
			"user_id":    userID,
			"profile_id": profileID,
		})
		return
	}
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}
