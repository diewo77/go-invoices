package view

import (
	"github.com/diewo77/billing-app/i18n"
	"github.com/diewo77/billing-app/internal/middleware"
	"github.com/diewo77/billing-app/internal/models"
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	baseDir  string
	once     sync.Once
	tplCache = struct {
		sync.RWMutex
		m map[string]*template.Template
	}{m: map[string]*template.Template{}}
	assetManifest     map[string]string
	assetManifestOnce sync.Once
)

func detectBase() {
	candidates := []string{"templates", "../templates", "../../templates"}
	for _, c := range candidates {
		if fi, err := os.Stat(filepath.Clean(c)); err == nil && fi.IsDir() {
			baseDir = filepath.Clean(c)
			return
		}
	}
	baseDir = "templates"
}

// Funcs returns the standard func map including i18n and simple helpers.
func Funcs(r *http.Request) template.FuncMap {
	lang := middleware.LangFrom(r)
	theme := middleware.ThemeFrom(r)
	return template.FuncMap{
		"t":     func(code string) string { return i18n.T(lang, code) },
		"lang":  func() string { return lang },
		"theme": func() string { return theme },
		"mul":   func(a, b float64) float64 { return a * b },
		"year":  func() int { return time.Now().Year() },
		"asset": func(path string) string { return resolveAsset(path) },
		"avgPrice": func(ps []models.Product) float64 {
			if len(ps) == 0 {
				return 0
			}
			var sum float64
			for _, p := range ps {
				sum += p.UnitPrice
			}
			return sum / float64(len(ps))
		},
	}
}

// versionedAsset returns /static/<name>?v=<hash> (or hashed filename) for cache busting.
// Current implementation uses query param; could be adapted to write hashed copy.
func versionedAsset(rel string) string {
	// If absolute URL or starts with http, return as-is
	if strings.HasPrefix(rel, "http://") || strings.HasPrefix(rel, "https://") || strings.HasPrefix(rel, "//") {
		return rel
	}
	p := filepath.Join("static", rel)
	b, err := os.ReadFile(p)
	if err != nil {
		return "/static/" + rel
	}
	h := sha1.Sum(b)
	return "/static/" + rel + "?v=" + fmt.Sprintf("%x", h[:8])
}

// resolveAsset prefers a hashed filename from manifest.json then falls back to query param versioning.
func resolveAsset(rel string) string {
	dev := os.Getenv("DEV") == "1"
	if dev {
		parseManifest() // reload each request in dev
	} else {
		assetManifestOnce.Do(parseManifest)
	}
	if assetManifest != nil {
		if h, ok := assetManifest[rel]; ok {
			return "/static/" + h
		}
	}
	return versionedAsset(rel)
}

func parseManifest() {
	mf := filepath.Join("static", "manifest.json")
	b, err := os.ReadFile(mf)
	if err != nil {
		return
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return
	}
	assetManifest = m
}

// Render parses and executes a single template file with shared funcs.
// name should be the filename (e.g., "dashboard.html").
func Render(w http.ResponseWriter, r *http.Request, name string, data map[string]any) error {
	once.Do(detectBase)
	// Ensure data map exists and inject common defaults to avoid template errors.
	if data == nil {
		data = map[string]any{}
	}
	if _, exists := data["Year"]; !exists {
		data["Year"] = time.Now().Year()
	}
	key := name
	devMode := os.Getenv("DEV") == "1"
	if !devMode {
		tplCache.RLock()
		t, ok := tplCache.m[key]
		tplCache.RUnlock()
		if ok && t != nil {
			return t.Execute(w, data)
		}
	}

	var t *template.Template
	mainPath := filepath.Join(baseDir, name)
	if _, err := os.Stat(mainPath); err != nil {
		// Attempt dynamic fallback search across relative parent levels
		candidates := []string{
			filepath.Join("templates", name),
			filepath.Join("../templates", name),
			filepath.Join("../../templates", name),
			filepath.Join("../../../templates", name),
		}
		for _, c := range candidates {
			if fi, e2 := os.Stat(c); e2 == nil && !fi.IsDir() {
				// Update baseDir for subsequent renders (thread-safe since once.Do already ran)
				baseDir = filepath.Dir(c)
				mainPath = c
				break
			}
		}
		if _, err2 := os.Stat(mainPath); err2 != nil {
			return err
		}
	}
	layoutPath := filepath.Join(baseDir, "layout.html")
	partials := []string{filepath.Join(baseDir, "partials", "header.html")}
	funcMap := Funcs(r)
	contentBytes, _ := os.ReadFile(mainPath)
	useLayout := true
	if bytes.Contains(bytes.ToLower(contentBytes), []byte("<!doctype")) {
		// Full document provided; skip layout wrapping.
		useLayout = false
	}
	if useLayout {
		if fi, err := os.Stat(layoutPath); err == nil && !fi.IsDir() {
			files := []string{layoutPath, mainPath}
			// append existing partials if they exist
			for _, p := range partials {
				if pf, err2 := os.Stat(p); err2 == nil && !pf.IsDir() {
					files = append(files, p)
				}
			}
			parsed, err := template.New("layout.html").Funcs(funcMap).ParseFiles(files...)
			if err != nil {
				return err
			}
			t = parsed
		} else {
			useLayout = false
		}
	}
	if !useLayout {
		parsed, err := template.New(name).Funcs(funcMap).ParseFiles(mainPath)
		if err != nil {
			return err
		}
		t = parsed
	}
	if !devMode {
		tplCache.Lock()
		tplCache.m[key] = t
		tplCache.Unlock()
	}
	if t == nil {
		return errors.New("template not cached")
	}
	return t.Execute(w, data)
}
