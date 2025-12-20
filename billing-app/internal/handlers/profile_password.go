package handlers

import (
	"github.com/diewo77/billing-app/auth"
	"github.com/diewo77/billing-app/i18n"
	"github.com/diewo77/billing-app/internal/middleware"
	"github.com/diewo77/billing-app/internal/models"
	"net/http"
	"net/url"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ProfileHandler struct {
	DB *gorm.DB
}

func NewProfileHandler(db *gorm.DB) *ProfileHandler { return &ProfileHandler{DB: db} }

// ChangePassword handles POST /profile/password
func (h *ProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	lang := middleware.LangFrom(r)
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok || uid == 0 {
		if parsed, ok2 := auth.ParseSession(r); ok2 {
			uid = parsed
		}
	}
	if uid == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		h.flash(w, i18n.T(lang, "flash_form_invalid"))
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}
	current := r.FormValue("current")
	newPass := r.FormValue("new")
	confirm := r.FormValue("confirm")
	var user models.User
	if err := h.DB.First(&user, uid).Error; err != nil {
		h.flash(w, i18n.T(lang, "flash_user_not_found"))
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(current)) != nil {
		h.flash(w, i18n.T(lang, "flash_password_current_bad"))
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}
	if len(newPass) < 8 || newPass != confirm {
		h.flash(w, i18n.T(lang, "flash_password_mismatch"))
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(newPass), bcrypt.DefaultCost)
	if err := h.DB.Model(&user).Update("password", string(hash)).Error; err != nil {
		h.flash(w, i18n.T(lang, "flash_password_save_error"))
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}
	h.flash(w, i18n.T(lang, "flash_password_saved"))
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (h *ProfileHandler) flash(w http.ResponseWriter, msg string) {
	http.SetCookie(w, &http.Cookie{Name: "flash", Value: url.QueryEscape(msg), Path: "/"})
}
