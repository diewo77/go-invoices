package main

import (
	"github.com/diewo77/billing-app/auth"
	"github.com/diewo77/billing-app/internal/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupE2EDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbi, err := gorm.Open(sqlite.Open("file:e2e_"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := dbi.AutoMigrate(&models.Address{}, &models.Role{}, &models.User{}, &models.CompanySettings{}, &models.Product{}, &models.Invoice{}, &models.Client{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return dbi
}

func TestDashboardRenderingE2E(t *testing.T) {
	dbi := setupE2EDB(t)
	u := models.User{Email: "e2e@example.com", Password: "hash"}
	if err := dbi.Create(&u).Error; err != nil {
		t.Fatalf("user: %v", err)
	}
	addr := models.Address{Ligne1: "1 rue E2E", CodePostal: "75000", Ville: "Paris", Pays: "FR"}
	if err := dbi.Create(&addr).Error; err != nil {
		t.Fatalf("addr: %v", err)
	}
	cs := models.CompanySettings{RaisonSociale: "E2E Corp", NomCommercial: "E2E Corp", SIREN: "123456789", SIRET: "12345678900011", CodeNAF: "6201Z", TypeImposition: "IS", FrequenceUrssaf: "mensuelle", RedevableTVA: true, FormeJuridique: "SAS", RegimeFiscal: "Réel", AddressID: addr.ID}
	if err := dbi.Create(&cs).Error; err != nil {
		t.Fatalf("cs: %v", err)
	}
	if err := dbi.Create(&models.Product{Name: "ProdX", UnitPrice: 15.5, VATRate: 0.2}).Error; err != nil {
		t.Fatalf("prod: %v", err)
	}
	if err := dbi.Create(&models.Invoice{Status: "draft"}).Error; err != nil {
		t.Fatalf("inv: %v", err)
	}

	app := NewApp(dbi)
	recSess := httptest.NewRecorder()
	auth.CreateSession(recSess, u.ID)
	var sess *http.Cookie
	for _, c := range recSess.Result().Cookies() {
		if c.Name == "session" {
			sess = c
			break
		}
	}
	if sess == nil {
		t.Fatalf("no session cookie")
	}

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req.AddCookie(sess)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		body := rr.Body.String()
		t.Fatalf("expected 200 got %d body=%s", rr.Code, body)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Bienvenue") {
		t.Fatalf("missing welcome text (FR): %s", body)
	}
	if !strings.Contains(body, "E2E Corp") {
		t.Fatalf("company name not rendered: %s", body)
	}
	if !strings.Contains(body, "Factures:") {
		t.Fatalf("stats block missing body=%s", body)
	}
}
