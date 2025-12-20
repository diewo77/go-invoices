package handlers

import (
	"github.com/diewo77/billing-app/auth"
	"github.com/diewo77/billing-app/internal/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	// Use a unique in-memory database per test to avoid cross-test collisions.
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.Role{}, &models.User{}, &models.Address{}, &models.CompanySettings{}, &models.ProductType{}, &models.UnitType{}, &models.Product{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestProductCreateAndList(t *testing.T) {
	db := setupTestDB(t)
	h := NewProductHandler(db)

	// Seed minimal company (requires user + address per model constraints)
	addr := models.Address{Ligne1: "1 rue", CodePostal: "75000", Ville: "Paris", Pays: "FR", Type: "principale"}
	if err := db.Create(&addr).Error; err != nil {
		t.Fatalf("addr: %v", err)
	}
	role := models.Role{Name: "user"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("role: %v", err)
	}
	user := models.User{Email: "u@test", Password: "x", Prenom: "U", Nom: "Test", AddressID: addr.ID, RoleID: role.ID}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("user: %v", err)
	}
	company := models.CompanySettings{UserID: user.ID, RaisonSociale: "RS", NomCommercial: "NC", SIREN: "123456789", SIRET: "12345678900011", CodeNAF: "6201Z", TypeImposition: "IS", FrequenceUrssaf: "mensuelle", RedevableTVA: true, FormeJuridique: "SAS", RegimeFiscal: "Réel", TVA: 0.2, DateCreation: time.Now()}
	if err := db.Create(&company).Error; err != nil {
		t.Fatalf("company: %v", err)
	}

	// Create (JSON path)
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(`{"code":"SKU1","name":"Test","unit_price":12.5,"vat_rate":0.2}`))
	req.Header.Set("Content-Type", "application/json")
	// Inject auth user into context
	req = req.WithContext(auth.WithUserID(req.Context(), user.ID))
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d", w.Code)
	}

	// List JSON
	req2 := httptest.NewRequest(http.MethodGet, "/products", nil)
	req2.Header.Set("Accept", "application/json")
	w2 := httptest.NewRecorder()
	h.List(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w2.Code)
	}
	var payload struct {
		Items  []models.Product `json:"items"`
		Total  int64            `json:"total"`
		Limit  int              `json:"limit"`
		Offset int              `json:"offset"`
	}
	if err := json.Unmarshal(w2.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 product got %d", len(payload.Items))
	}
	if payload.Items[0].Name != "Test" {
		t.Fatalf("unexpected product name: %s", payload.Items[0].Name)
	}
}

func TestProductListHTML_NoCompany(t *testing.T) {
	db := setupTestDB(t)
	h := NewProductHandler(db)
	// No company seeded
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	// Browser-like accept
	req.Header.Set("Accept", "text/html")
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Aucune société configurée") {
		t.Fatalf("expected no company notice in body: %s", body)
	}
}

func TestProductAvgPriceFunc(t *testing.T) {
	db := setupTestDB(t)
	h := NewProductHandler(db)
	// Seed company + products
	addr := models.Address{Ligne1: "1 rue", CodePostal: "75000", Ville: "Paris", Pays: "FR", Type: "principale"}
	_ = db.Create(&addr).Error
	role := models.Role{Name: "user"}
	db.Create(&role)
	user := models.User{Email: "a@b", Password: "x", Prenom: "A", Nom: "B", AddressID: addr.ID, RoleID: role.ID}
	_ = db.Create(&user).Error
	company := models.CompanySettings{UserID: user.ID, RaisonSociale: "RS", NomCommercial: "NC", SIREN: "123456789", SIRET: "12345678900011", CodeNAF: "6201Z", TypeImposition: "IS", FrequenceUrssaf: "mensuelle", RedevableTVA: true, FormeJuridique: "SAS", RegimeFiscal: "Réel"}
	_ = db.Create(&company).Error
	db.Create(&models.Product{CompanyID: company.ID, UserID: user.ID, Code: "P1", Name: "P1", UnitPrice: 10, VATRate: 0.2, Currency: "EUR"})
	db.Create(&models.Product{CompanyID: company.ID, UserID: user.ID, Code: "P2", Name: "P2", UnitPrice: 30, VATRate: 0.2, Currency: "EUR"})
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	req.Header.Set("Accept", "text/html")
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "20.00€") {
		t.Fatalf("expected avg 20.00€ in body, got: %s", w.Body.String())
	}
}
