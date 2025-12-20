package middleware

import (
	"github.com/diewo77/billing-app/i18n"
	"context"
	"net/http"
	"net/url"
)

type ctxKey string

const (
	ctxLang  ctxKey = "pref_lang"
	ctxTheme ctxKey = "pref_theme"
)

// Prefs extracts language/theme preferences (cookie > query > header) and stores them in context.
// It also normalizes values and persists query-provided prefs in cookies for ~30 days.
func Prefs(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := "fr"
		if c, err := r.Cookie("lang"); err == nil && c.Value != "" {
			lang = c.Value
		}
		if ql := r.URL.Query().Get("lang"); ql != "" {
			lang = ql
			http.SetCookie(w, &http.Cookie{Name: "lang", Value: lang, Path: "/", MaxAge: 86400 * 30})
		}
		if lang != "fr" && lang != "en" {
			lang = i18n.DetectLanguage(r.Header.Get("Accept-Language"))
		}
		if lang != "fr" && lang != "en" {
			lang = "fr"
		}
		theme := "system"
		if c, err := r.Cookie("theme"); err == nil && c.Value != "" {
			theme = c.Value
		}
		if qt := r.URL.Query().Get("theme"); qt != "" {
			theme = qt
			http.SetCookie(w, &http.Cookie{Name: "theme", Value: theme, Path: "/", MaxAge: 86400 * 30})
		}
		ctx := context.WithValue(r.Context(), ctxLang, lang)
		ctx = context.WithValue(ctx, ctxTheme, theme)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LangFrom returns language preference from context or fallback.
func LangFrom(r *http.Request) string {
	if v, ok := r.Context().Value(ctxLang).(string); ok && v != "" {
		return v
	}
	return "fr"
}

// ThemeFrom returns theme preference from context or fallback.
func ThemeFrom(r *http.Request) string {
	if v, ok := r.Context().Value(ctxTheme).(string); ok && v != "" {
		return v
	}
	return "system"
}

// Flash sets a translated flash message cookie using translation code (or literal if missing).
func Flash(w http.ResponseWriter, r *http.Request, code string) {
	lang := LangFrom(r)
	msg := i18n.T(lang, code)
	http.SetCookie(w, &http.Cookie{Name: "flash", Value: url.QueryEscape(msg), Path: "/"})
}
