package handlers

import (
	"net/http"

	"github.com/diewo77/go-invoices/auth"
	"github.com/diewo77/go-invoices/internal/models"
	"github.com/diewo77/go-invoices/view"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db *gorm.DB
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		view.Render(w, r, "login.html", nil)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	var user models.User
	if err := h.db.Where("email = ?", email).First(&user).Error; err != nil {
		view.Render(w, r, "login.html", map[string]any{"Error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		view.Render(w, r, "login.html", map[string]any{"Error": "Invalid email or password"})
		return
	}

	auth.CreateSession(w, user.ID)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		view.Render(w, r, "signup.html", nil)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	name := r.FormValue("name")

	if email == "" || password == "" {
		view.Render(w, r, "signup.html", map[string]any{"Error": "Email and password are required"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		view.Render(w, r, "signup.html", map[string]any{"Error": "Internal server error"})
		return
	}

	user := models.User{
		Email:    email,
		Password: string(hashedPassword),
		Name:     name,
	}

	if err := h.db.Create(&user).Error; err != nil {
		view.Render(w, r, "signup.html", map[string]any{"Error": "Email already exists"})
		return
	}

	auth.CreateSession(w, user.ID)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	auth.ClearSession(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
