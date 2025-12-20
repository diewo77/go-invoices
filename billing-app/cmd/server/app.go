package main

import (
	"github.com/diewo77/billing-app/auth"
	"github.com/diewo77/billing-app/internal/handlers"
	"github.com/diewo77/billing-app/internal/middleware"
	"github.com/diewo77/billing-app/internal/models"
	"github.com/diewo77/billing-app/internal/server"
	"github.com/diewo77/billing-app/internal/services"
	"github.com/diewo77/billing-app/view"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

// NewApp bundles landing, dashboard, and API routes for end-to-end tests.
var templateBase string

func init() {
	// Detect templates directory whether running from repo root or subdir (e.g., cmd/server).
	candidates := []string{"templates", "../templates", "../../templates"}
	for _, c := range candidates {
		if fi, err := os.Stat(filepath.Clean(c)); err == nil && fi.IsDir() {
			templateBase = filepath.Clean(c)
			break
		}
	}
	if templateBase == "" { // fallback to relative; parsing will error clearly
		templateBase = "templates"
	}

	// Inject language/theme resolvers into the shared view package so it stays decoupled
	// from the middleware package while still reflecting user preferences.
	view.SetLangResolver(middleware.LangFrom)
	view.SetThemeResolver(middleware.ThemeFrom)
}

// resolve language/theme from context (middleware injects) fallback to defaults
func prefsFrom(r *http.Request) (string, string) {
	return middleware.LangFrom(r), middleware.ThemeFrom(r)
}

func NewApp(dbConn *gorm.DB) http.Handler {
	rootAPI := auth.Middleware(server.New(dbConn))

	// serve static assets (CSS, JS) under /static/
	fs := http.FileServer(http.Dir("static"))
	staticHandler := http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path
		ext := filepath.Ext(name)
		// open file manually to compute ETag
		f, err := os.Open(filepath.Join("static", name))
		if err == nil {
			defer f.Close()
			h := sha1.New()
			// small files only; large could be optimized with stat modtime
			if _, cerr := io.Copy(h, f); cerr == nil {
				etag := fmt.Sprintf("\"%x\"", h.Sum(nil)[:8])
				w.Header().Set("ETag", etag)
				if match := r.Header.Get("If-None-Match"); match == etag {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
			// rewind for file server by reopening
			f.Close()
		}
		if ext == ".css" {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		} else if ext == ".js" {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		}
		if os.Getenv("DEV") != "1" {
			// Long cache for versioned assets (1 year)
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		}
		fs.ServeHTTP(w, r)
	}))
	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/static/" || len(r.URL.Path) > 8 && r.URL.Path[:8] == "/static/" {
			// Add sensible cache headers for production; disable in DEV
			if os.Getenv("DEV") != "1" {
				w.Header().Set("Cache-Control", "public, max-age=86400")
			} else {
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
			}
			staticHandler.ServeHTTP(w, r)
			return
		}
		_, _ = prefsFrom(r) // ensure prefs middleware executed (values used within view funcs)

		if r.URL.Path == "/logout" {
			auth.ClearSession(w)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if r.URL.Path == "/dashboard" {
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
			data := map[string]any{"Year": time.Now().Year()}
			if c, err := r.Cookie("flash"); err == nil {
				if dec, derr := url.QueryUnescape(c.Value); derr == nil {
					data["Flash"] = dec
				} else {
					data["Flash"] = c.Value
				}
				http.SetCookie(w, &http.Cookie{Name: "flash", Value: "", Path: "/", Expires: time.Unix(0, 0), MaxAge: -1})
			}
			svc := services.NewSetupService(dbConn)
			if cs, err := svc.Get(); err == nil && cs != nil {
				data["Company"] = cs
			}
			var user models.User
			if err := dbConn.First(&user, uid).Error; err == nil {
				data["User"] = user
			}
			var invoiceCount, productCount, clientCount int64
			dbConn.Model(&models.Invoice{}).Count(&invoiceCount)
			dbConn.Model(&models.Product{}).Count(&productCount)
			dbConn.Model(&models.Client{}).Count(&clientCount)
			data["Stats"] = map[string]any{"InvoiceCount": invoiceCount, "ProductCount": productCount, "ClientCount": clientCount}
			var recentProducts []models.Product
			dbConn.Order("created_at desc").Limit(5).Find(&recentProducts)
			data["RecentProducts"] = recentProducts
			var recentInvoices []models.Invoice
			dbConn.Order("created_at desc").Limit(5).Find(&recentInvoices)
			data["RecentInvoices"] = recentInvoices
			if err := view.Render(w, r, "dashboard.html", data); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "template render error: %v", err)
			}
			return
		}
		if r.URL.Path == "/" {
			data := map[string]any{}
			if c, err := r.Cookie("flash"); err == nil {
				if dec, derr := url.QueryUnescape(c.Value); derr == nil {
					data["Flash"] = dec
				} else {
					data["Flash"] = c.Value
				}
				http.SetCookie(w, &http.Cookie{Name: "flash", Value: "", Path: "/", Expires: time.Unix(0, 0), MaxAge: -1})
			}
			uid, ok := auth.UserIDFromContext(r.Context())
			if !ok || uid == 0 {
				if parsed, ok2 := auth.ParseSession(r); ok2 {
					uid = parsed
				}
			}
			if uid != 0 {
				var user models.User
				if err := dbConn.First(&user, uid).Error; err == nil {
					data["User"] = user
				}
				svc := services.NewSetupService(dbConn)
				if cs, err := svc.Get(); err == nil && cs != nil {
					data["Company"] = cs
				}
			}
			if err := view.Render(w, r, "index.html", data); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				if _, werr := w.Write([]byte("render error")); werr != nil {
					_ = werr
				}
			}
			return
		}
		if r.URL.Path == "/profile" {
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
			var user models.User
			if err := dbConn.First(&user, uid).Error; err != nil {
				w.WriteHeader(http.StatusNotFound)
				if _, werr := w.Write([]byte("utilisateur introuvable")); werr != nil {
					_ = werr
				}
				return
			}
			data := map[string]any{"User": user, "Year": time.Now().Year()}
			if c, err := r.Cookie("flash"); err == nil {
				data["Flash"] = c.Value
				http.SetCookie(w, &http.Cookie{Name: "flash", Value: "", Path: "/", Expires: time.Unix(0, 0), MaxAge: -1})
			}
			if err := view.Render(w, r, "profile.html", data); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				if _, werr := w.Write([]byte("render error")); werr != nil {
					_ = werr
				}
			}
			return
		}
		if r.URL.Path == "/profile/password" && r.Method == http.MethodPost {
			ph := handlers.NewProfileHandler(dbConn)
			ph.ChangePassword(w, r)
			return
		}
		// Friendly alias: /settings -> /setup (historical naming mismatch)
		if r.URL.Path == "/settings" {
			http.Redirect(w, r, "/setup", http.StatusSeeOther)
			return
		}
		rootAPI.ServeHTTP(w, r)
	})
	return middleware.Prefs(baseHandler)
}
