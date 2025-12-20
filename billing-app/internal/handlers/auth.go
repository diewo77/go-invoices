package handlers

import (
	"github.com/diewo77/billing-app/auth"
	"github.com/diewo77/billing-app/httpx"
	"github.com/diewo77/billing-app/internal/models"
	"github.com/diewo77/billing-app/view"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ensureDefaultRole fetches or creates the base "user" role.
func ensureDefaultRole(db *gorm.DB) (*models.Role, error) {
	var role models.Role
	if err := db.Where("name = ?", "user").First(&role).Error; err == nil {
		return &role, nil
	}
	role = models.Role{Name: "user", Description: "Default user role"}
	if err := db.Create(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

type AuthHandler struct{ DB *gorm.DB }

// Explicit constant for 303 See Other (Post/Redirect/Get)
const statusSeeOther = 303

func NewAuthHandler(db *gorm.DB) *AuthHandler { return &AuthHandler{DB: db} }

func (h *AuthHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/signup", h.signup)
	mux.HandleFunc("/login", h.login)
	mux.HandleFunc("/logout", h.logout)
}

// render uses the shared view.Render to ensure layout, partials, funcs, and caching.
func renderTemplate(w http.ResponseWriter, r *http.Request, name string, data map[string]any) {
	if err := view.Render(w, r, name+".html", data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, werr := w.Write([]byte("template error")); werr != nil {
			_ = werr
		}
	}
}

func (h *AuthHandler) signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, r, "signup", nil)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET,POST")
		httpx.JSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", nil)
		return
	}
	if err := r.ParseForm(); err != nil {
		httpx.JSONError(w, http.StatusBadRequest, "invalid_form", nil)
		return
	}
	email := strings.TrimSpace(r.FormValue("email"))
	pass := r.FormValue("password")
	prenom := strings.TrimSpace(r.FormValue("prenom"))
	nom := strings.TrimSpace(r.FormValue("nom"))
	addr1 := strings.TrimSpace(r.FormValue("address1"))
	addr2 := strings.TrimSpace(r.FormValue("address2"))
	postal := strings.TrimSpace(r.FormValue("postal_code"))
	city := strings.TrimSpace(r.FormValue("city"))
	country := strings.TrimSpace(r.FormValue("country"))
	if email == "" || pass == "" {
		renderTemplate(w, r, "signup", map[string]any{"Error": "email and password required"})
		return
	}
	if addr1 == "" || postal == "" || city == "" || country == "" {
		renderTemplate(w, r, "signup", map[string]any{"Error": "missing required address fields"})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	address := models.Address{Ligne1: addr1, Ligne2: addr2, CodePostal: postal, Ville: city, Pays: country, Type: "principale"}
	if err := h.DB.Create(&address).Error; err != nil {
		renderTemplate(w, r, "signup", map[string]any{"Error": "could not save address"})
		return
	}
	// Assign default role (create if missing)
	role, err := ensureDefaultRole(h.DB)
	if err != nil {
		renderTemplate(w, r, "signup", map[string]any{"Error": "could not ensure role"})
		return
	}
	user := models.User{Email: email, Password: string(hash), Prenom: prenom, Nom: nom, AddressID: address.ID, RoleID: role.ID}
	if err := h.DB.Create(&user).Error; err != nil {
		renderTemplate(w, r, "signup", map[string]any{"Error": "could not create user"})
		return
	}
	auth.CreateSession(w, user.ID)
	// PRG redirect (303)
	http.Redirect(w, r, "/dashboard", statusSeeOther)
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// If already logged in, verify user still exists, then redirect to dashboard.
		if uid, ok := auth.UserIDFromContext(r.Context()); ok && uid != 0 {
			var count int64
			if err := h.DB.Model(&models.User{}).Where("id = ?", uid).Limit(1).Count(&count).Error; err == nil && count > 0 {
				http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
				return
			}
			// Stale session: clear and continue to render login
			auth.ClearSession(w)
		} else if parsed, ok2 := auth.ParseSession(r); ok2 && parsed != 0 {
			var count int64
			if err := h.DB.Model(&models.User{}).Where("id = ?", parsed).Limit(1).Count(&count).Error; err == nil && count > 0 {
				http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
				return
			}
			auth.ClearSession(w)
		}
		renderTemplate(w, r, "login", nil)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET,POST")
		httpx.JSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", nil)
		return
	}
	if err := r.ParseForm(); err != nil {
		httpx.JSONError(w, http.StatusBadRequest, "invalid_form", nil)
		return
	}
	email := strings.TrimSpace(r.FormValue("email"))
	pass := r.FormValue("password")
	if email == "" || pass == "" {
		renderTemplate(w, r, "login", map[string]any{"Error": "email and password required"})
		return
	}
	var user models.User
	if err := h.DB.Where("email = ?", email).First(&user).Error; err != nil {
		renderTemplate(w, r, "login", map[string]any{"Error": "invalid credentials"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass)) != nil {
		renderTemplate(w, r, "login", map[string]any{"Error": "invalid credentials"})
		return
	}
	auth.CreateSession(w, user.ID)
	http.Redirect(w, r, "/dashboard", statusSeeOther)
}

func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	auth.ClearSession(w)
	http.Redirect(w, r, "/login", statusSeeOther)
}
