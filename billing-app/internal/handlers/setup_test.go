package handlers

import (
	"github.com/diewo77/billing-app/auth"
	"github.com/diewo77/billing-app/internal/models"
	"github.com/diewo77/billing-app/internal/services"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSetupTestDB(t *testing.T) *gorm.DB {
	// unique in-memory DB per test name to avoid leakage via shared cache
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.Address{}, &models.User{}, &models.CompanySettings{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestSetupHandlerJSONFlow(t *testing.T) {
	db := setupSetupTestDB(t)
	svc := services.NewSetupService(db)
	h := NewSetupHandler(svc)
	mux := http.NewServeMux()
	h.Register(mux)
	wrapped := auth.Middleware(mux)

	// Seed user and session cookie
	user := models.User{Email: "tester@example.com", Password: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	sessW := httptest.NewRecorder()
	auth.CreateSession(sessW, user.ID)
	cookie := sessW.Result().Cookies()[0]
	// GET not configured (JSON)
	resp := httptest.NewRecorder()
	getReq := httptest.NewRequest(http.MethodGet, "/setup", nil)
	getReq.Header.Set("Accept", "application/json")
	getReq.AddCookie(cookie)
	wrapped.ServeHTTP(resp, getReq)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", resp.Code)
	}
	var status map[string]bool
	if err := json.Unmarshal(resp.Body.Bytes(), &status); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if status["configured"] {
		t.Fatalf("expected configured=false")
	}

	// POST success (now returns 200 since we unified create/update)
	post := httptest.NewRecorder()
	body := `{"company":"Acme","address1":"1 rue test","postal_code":"75000","city":"Paris","country":"FR","siret":"12345678900011"}`
	postReq := httptest.NewRequest(http.MethodPost, "/setup", strings.NewReader(body))
	postReq.Header.Set("Content-Type", "application/json")
	postReq.AddCookie(cookie)
	wrapped.ServeHTTP(post, postReq)
	if post.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", post.Code, post.Body.String())
	}

	// Second POST now updates not conflict
	conf := httptest.NewRecorder()
	body2 := `{"company":"Acme2","address1":"1 rue test","postal_code":"75000","city":"Paris","country":"FR","siret":"12345678900011"}`
	confReq := httptest.NewRequest(http.MethodPost, "/setup", strings.NewReader(body2))
	confReq.Header.Set("Content-Type", "application/json")
	confReq.AddCookie(cookie)
	wrapped.ServeHTTP(conf, confReq)
	if conf.Code != http.StatusOK {
		t.Fatalf("expected 200 update got %d", conf.Code)
	}
}

func TestSetupHandlerFormAndHead(t *testing.T) {
	db := setupSetupTestDB(t)
	svc := services.NewSetupService(db)
	h := NewSetupHandler(svc)
	mux := http.NewServeMux()
	h.Register(mux)
	wrapped := auth.Middleware(mux)

	// HEAD before config
	headResp := httptest.NewRecorder()
	headReq := httptest.NewRequest(http.MethodHead, "/setup", nil)
	// Seed user & session
	user := models.User{Email: "headuser@example.com", Password: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	sess := httptest.NewRecorder()
	auth.CreateSession(sess, user.ID)
	cookie := sess.Result().Cookies()[0]
	wrapped.ServeHTTP(headResp, headReq)
	if headResp.Code != http.StatusOK {
		t.Fatalf("HEAD expected 200 got %d", headResp.Code)
	}
	if v := headResp.Header().Get("X-Setup-Configured"); v != "false" {
		t.Fatalf("expected header false got %s", v)
	}

	// Form POST
	formBody := "company=FormCo&address=1+rue+test&postal_code=75000&city=Paris&country=FR&siret=12345678900011&tva=non"
	postResp := httptest.NewRecorder()
	postReq := httptest.NewRequest(http.MethodPost, "/setup", strings.NewReader(formBody))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(cookie)
	wrapped.ServeHTTP(postResp, postReq)
	if postResp.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect got %d", postResp.Code)
	}

	// HEAD after config
	head2 := httptest.NewRecorder()
	headReq2 := httptest.NewRequest(http.MethodHead, "/setup", nil)
	headReq2.Header.Set("Accept", "application/json")
	headReq2.AddCookie(cookie)
	wrapped.ServeHTTP(head2, headReq2)
	if head2.Code != http.StatusOK {
		t.Fatalf("HEAD expected 200 got %d", head2.Code)
	}
	if v := head2.Header().Get("X-Setup-Configured"); v != "true" {
		t.Fatalf("expected header true got %s", v)
	}
}

func TestValidateSetupFunction(t *testing.T) {
	req := &setupRequest{}
	errs := validateSetup(req, false)
	if len(errs) == 0 || errs["company"] == "" || errs["siret"] == "" {
		t.Fatalf("expected required field errors, got %#v", errs)
	}
	req = &setupRequest{Company: "Co", Address1: "Adr", PostalCode: "75000", City: "Paris", Country: "fr", SIRET: "12345678901234"}
	errs = validateSetup(req, false)
	if len(errs) != 0 {
		t.Fatalf("unexpected errs %#v", errs)
	}
	if req.Country != "FR" {
		t.Fatalf("expected upper country got %s", req.Country)
	}
}

func TestFormValidationInlineErrors(t *testing.T) {
	db := setupSetupTestDB(t)
	// Ensure working dir is project root (one level up from internal/handlers)
	cwd, _ := os.Getwd()
	root := filepath.Clean(filepath.Join(cwd, "../.."))
	_ = os.Chdir(root)
	svc := services.NewSetupService(db)
	h := NewSetupHandler(svc)
	mux := http.NewServeMux()
	h.Register(mux)
	wrapped := auth.Middleware(mux)
	user := models.User{Email: "inline@example.com", Password: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	sess := httptest.NewRecorder()
	auth.CreateSession(sess, user.ID)
	cookie := sess.Result().Cookies()[0]
	// Missing SIRET triggers error, expect 400 and template content
	form := url.Values{}
	form.Set("company", "TestCo")
	form.Set("address", "1 rue test")
	form.Set("postal_code", "75000")
	form.Set("city", "Paris")
	form.Set("country", "FR")
	r := httptest.NewRequest(http.MethodPost, "/setup", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(cookie)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "name=\"siret\"") {
		t.Fatalf("expected form re-render")
	}
	if !strings.Contains(strings.ToLower(w.Body.String()), "siret") {
		t.Fatalf("expected siret error inline")
	}
}

func TestFlashCookieOnCreateAndUpdate(t *testing.T) {
	db := setupSetupTestDB(t)
	// set working dir to project root for templates
	cwd, _ := os.Getwd()
	root := filepath.Clean(filepath.Join(cwd, "../.."))
	_ = os.Chdir(root)
	svc := services.NewSetupService(db)
	h := NewSetupHandler(svc)
	mux := http.NewServeMux()
	h.Register(mux)
	wrapped := auth.Middleware(mux)
	user := models.User{Email: "flash@example.com", Password: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	sess := httptest.NewRecorder()
	auth.CreateSession(sess, user.ID)
	cookie := sess.Result().Cookies()[0]
	form := url.Values{
		"company":     []string{"FlashCo"},
		"address":     []string{"1 rue"},
		"postal_code": []string{"75000"},
		"city":        []string{"Paris"},
		"country":     []string{"FR"},
		"siret":       []string{"12345678900011"},
		"tva":         []string{"non"},
	}
	r := httptest.NewRequest(http.MethodPost, "/setup", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.AddCookie(cookie)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, r)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect create got %d", w.Code)
	}
	var flashSet bool
	for _, c := range w.Result().Cookies() {
		if c.Name == "flash" {
			flashSet = true
		}
	}
	if !flashSet {
		t.Fatalf("expected flash cookie on create")
	}
	// Update
	form.Set("company", "FlashCo2")
	r2 := httptest.NewRequest(http.MethodPost, "/setup", strings.NewReader(form.Encode()))
	r2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r2.AddCookie(cookie)
	w2 := httptest.NewRecorder()
	wrapped.ServeHTTP(w2, r2)
	if w2.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect update got %d", w2.Code)
	}
	flashSet = false
	for _, c := range w2.Result().Cookies() {
		if c.Name == "flash" {
			flashSet = true
		}
	}
	if !flashSet {
		t.Fatalf("expected flash cookie on update")
	}
}

// Ensure JSON validation error returns field map
func TestJSONValidationErrors(t *testing.T) {
	db := setupSetupTestDB(t)
	svc := services.NewSetupService(db)
	h := NewSetupHandler(svc)
	mux := http.NewServeMux()
	h.Register(mux)
	wrapped := auth.Middleware(mux)
	user := models.User{Email: "jsoninline@example.com", Password: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	sess := httptest.NewRecorder()
	auth.CreateSession(sess, user.ID)
	cookie := sess.Result().Cookies()[0]
	body := `{"company":"","address1":"","postal_code":"","city":"","country":"","siret":""}`
	r := httptest.NewRequest(http.MethodPost, "/setup", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.AddCookie(cookie)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "validation_error") {
		t.Fatalf("expected validation_error body=%s", w.Body.String())
	}
}
