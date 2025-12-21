package view

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/diewo77/go-invoices/auth"
	"github.com/diewo77/go-invoices/i18n"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
)

// Context key for theme
type themeKey struct{}

// WithTheme returns a new context with the given theme.
func WithTheme(ctx context.Context, theme string) context.Context {
	return context.WithValue(ctx, themeKey{}, theme)
}

// ThemeFromContext retrieves the theme from context, defaulting to "light".
func ThemeFromContext(ctx context.Context) string {
	if theme, ok := ctx.Value(themeKey{}).(string); ok {
		return theme
	}
	return "light"
}

var (
	baseDir  string
	once     sync.Once
	tplCache = struct {
		sync.RWMutex
		m map[string]*template.Template
	}{m: map[string]*template.Template{}}
	assetManifest     map[string]string
	assetManifestOnce sync.Once

	langResolver  = func(_ *http.Request) string { return "fr" }
	themeResolver = func(_ *http.Request) string { return "system" }
	// permission resolvers can be set by the host app to allow templates to check auth
	canProfileResolver func(*http.Request, string, string) bool
	isAdminResolver    func(*http.Request) bool
)

// SetAuthGate allows the host app to provide the AuthGate used for permission checks in templates.
// SetCanProfileResolver sets a callback used by templates to check profile-level permissions.
func SetCanProfileResolver(f func(*http.Request, string, string) bool) {
	if f != nil {
		canProfileResolver = f
	}
}

// SetIsAdminResolver sets a callback used by templates to determine superadmin users.
func SetIsAdminResolver(f func(*http.Request) bool) {
	if f != nil {
		isAdminResolver = f
	}
}

// layoutBase walks upward from a template path to find the directory that contains layout.html.
// If none is found, it returns the template's own directory.
func layoutBase(mainPath string) string {
	d := filepath.Dir(mainPath)
	for {
		lp := filepath.Join(d, "layout.html")
		if fi, err := os.Stat(lp); err == nil && !fi.IsDir() {
			return d
		}
		p := filepath.Dir(d)
		if p == d { // reached filesystem root
			return filepath.Dir(mainPath)
		}
		d = p
	}
}

// SetLangResolver allows the host app to provide a custom language resolver (e.g., reading from context).
func SetLangResolver(f func(*http.Request) string) {
	if f != nil {
		langResolver = f
	}
}

// SetThemeResolver allows the host app to provide a custom theme resolver.
func SetThemeResolver(f func(*http.Request) string) {
	if f != nil {
		themeResolver = f
	}
}

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
	lang := langResolver(r)
	theme := themeResolver(r)
	return template.FuncMap{
		"t":     func(code string) string { return i18n.T(lang, code) },
		"lang":  func() string { return lang },
		// can checks profile-level permission (resource, action) -> bool
		"can": func(resource string, action string) bool {
			if canProfileResolver == nil {
				return false
			}
			return canProfileResolver(r, resource, action)
		},
		// isAdmin returns true if the authenticated user has superadmin permission
		"isAdmin": func() bool {
			if isAdminResolver == nil {
				return false
			}
			return isAdminResolver(r)
		},
		"theme": func() string { return theme },
		"mul": func(a, b any) float64 {
			fa, oka := toFloat64(a)
			fb, okb := toFloat64(b)
			if !oka || !okb {
				return 0
			}
			return fa * fb
		},
		"add": func(a, b any) float64 {
			fa, oka := toFloat64(a)
			fb, okb := toFloat64(b)
			if !oka || !okb {
				return 0
			}
			return fa + fb
		},
		"year":  func() int { return time.Now().Year() },
		"asset": func(path string) string { return resolveAsset(path) },
		"avgPrice": func(items any) float64 {
			return averageUnitPrice(items)
		},
		// dict creates a map from key-value pairs for passing to sub-templates.
		// Usage: {{ template "partial" (dict "Key1" val1 "Key2" val2) }}
		"dict": func(values ...any) map[string]any {
			if len(values)%2 != 0 {
				return nil
			}
			m := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					continue
				}
				m[key] = values[i+1]
			}
			return m
		},
	}
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	default:
		return 0, false
	}
}

func averageUnitPrice(items any) float64 {
	v := reflect.ValueOf(items)
	if v.Kind() == reflect.Pointer && !v.IsNil() {
		v = v.Elem()
	}
	if v.Kind() != reflect.Slice {
		return 0
	}
	n := v.Len()
	if n == 0 {
		return 0
	}
	var sum float64
	for i := 0; i < n; i++ {
		p, ok := extractUnitPrice(v.Index(i))
		if !ok {
			return 0
		}
		sum += p
	}
	return sum / float64(n)
}

func extractUnitPrice(v reflect.Value) (float64, bool) {
	if !v.IsValid() {
		return 0, false
	}
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return 0, false
		}
		v = v.Elem()
	}
	if v.CanInterface() {
		if provider, ok := v.Interface().(interface{ GetUnitPrice() float64 }); ok {
			return provider.GetUnitPrice(), true
		}
		if provider, ok := v.Interface().(interface{ UnitPrice() float64 }); ok {
			return provider.UnitPrice(), true
		}
	}
	if v.Kind() == reflect.Struct {
		field := v.FieldByName("UnitPrice")
		if field.IsValid() && field.CanConvert(reflect.TypeOf(float64(0))) {
			return field.Convert(reflect.TypeOf(float64(0))).Float(), true
		}
	}
	return 0, false
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

// SetBaseDir overrides the template base directory (useful for tests or custom setups).
func SetBaseDir(path string) {
	if path == "" {
		return
	}
	baseDir = filepath.Clean(path)
	once = sync.Once{}
}

// ResetForTests clears caches and forces base dir detection to rerun.
// Intended for test code to avoid cross-test pollution when working directories change.
func ResetForTests() {
	tplCache.Lock()
	tplCache.m = map[string]*template.Template{}
	tplCache.Unlock()
	baseDir = ""
	once = sync.Once{}
}

// Render parses and executes a single template file with shared funcs.
// name should be the filename (e.g., "dashboard.html").
func Render(w http.ResponseWriter, r *http.Request, name string, data map[string]any) error {
	if baseDir == "" {
		once.Do(detectBase)
	}
	// Ensure data map exists and inject common defaults to avoid template errors.
	if data == nil {
		data = map[string]any{}
	}
	if _, exists := data["Year"]; !exists {
		data["Year"] = time.Now().Year()
	}
	if _, exists := data["IsLoggedIn"]; !exists {
		_, loggedIn := auth.UserIDFromContext(r.Context())
		data["IsLoggedIn"] = loggedIn
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
				mainPath = c
				break
			}
		}
		if _, err2 := os.Stat(mainPath); err2 != nil {
			return err
		}
	}
	// Align baseDir to the directory that owns layout.html (typically the templates root)
	baseDir = layoutBase(mainPath)
	layoutPath := filepath.Join(baseDir, "layout.html")
	partials := []string{
		filepath.Join(baseDir, "partials", "header.html"),
		filepath.Join(baseDir, "partials", "page-header.html"),
		filepath.Join(baseDir, "partials", "errors-alert.html"),
		filepath.Join(baseDir, "partials", "stat-card.html"),
		filepath.Join(baseDir, "partials", "search-filter.html"),
		filepath.Join(baseDir, "partials", "field-text.html"),
		filepath.Join(baseDir, "partials", "field-select.html"),
	}
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
