package handlers

import (
	"github.com/diewo77/billing-app/auth"
	"github.com/diewo77/billing-app/internal/models"
	"github.com/diewo77/billing-app/internal/services"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupInvoiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.Role{}, &models.User{}, &models.Address{}, &models.CompanySettings{}, &models.ProductType{}, &models.UnitType{}, &models.Product{}, &models.Client{}, &models.Invoice{}, &models.InvoiceItem{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// seed minimal user/company/client/product for invoices
func seedInvoiceFixtures(t *testing.T, db *gorm.DB) (user models.User, company models.CompanySettings, client models.Client, product models.Product) {
	t.Helper()
	addr := models.Address{Ligne1: "1 rue", CodePostal: "75000", Ville: "Paris", Pays: "FR", Type: "principale"}
	if err := db.Create(&addr).Error; err != nil {
		t.Fatalf("addr: %v", err)
	}
	role := models.Role{Name: "user"}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("role: %v", err)
	}
	user = models.User{Email: "inv@test", Password: "x", Prenom: "I", Nom: "User", AddressID: addr.ID, RoleID: role.ID}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("user: %v", err)
	}
	company = models.CompanySettings{UserID: user.ID, RaisonSociale: "RS", NomCommercial: "NC", SIREN: "123456789", SIRET: "12345678900011", CodeNAF: "6201Z", TypeImposition: "IS", FrequenceUrssaf: "mensuelle", RedevableTVA: true, FormeJuridique: "SAS", RegimeFiscal: "Réel", TVA: 0.2, DateCreation: time.Now()}
	if err := db.Create(&company).Error; err != nil {
		t.Fatalf("company: %v", err)
	}
	client = models.Client{UserID: user.ID, Nom: "ClientCo"}
	if err := db.Create(&client).Error; err != nil {
		t.Fatalf("client: %v", err)
	}
	product = models.Product{CompanyID: company.ID, UserID: user.ID, Code: "SKU1", Name: "Service", UnitPrice: 100, VATRate: 0.2, Currency: "EUR"}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("product: %v", err)
	}
	return
}

func TestInvoiceCreateAndListJSON(t *testing.T) {
	db := setupInvoiceTestDB(t)
	user, _, client, product := seedInvoiceFixtures(t, db)
	h := NewInvoiceHandler(db, services.NewInvoiceService())

	// Create JSON
	body := `{"client_id":` + strconv.Itoa(int(client.ID)) + `,"items":[{"product_id":` + strconv.Itoa(int(product.ID)) + `,"quantity":2}]}`
	req := httptest.NewRequest(http.MethodPost, "/invoices", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// attach auth context
	req = req.WithContext(auth.WithUserID(req.Context(), user.ID))
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d body=%s", w.Code, w.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if created["id"] == nil {
		t.Fatalf("missing id in response: %#v", created)
	}

	// List JSON
	listReq := httptest.NewRequest(http.MethodGet, "/invoices", nil)
	listReq.Header.Set("Accept", "application/json")
	listW := httptest.NewRecorder()
	h.List(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", listW.Code)
	}
	var list struct {
		Items  []models.Invoice `json:"items"`
		Total  int64            `json:"total"`
		Limit  int              `json:"limit"`
		Offset int              `json:"offset"`
	}
	if err := json.Unmarshal(listW.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list.Items) != 1 || list.Total < 1 {
		t.Fatalf("unexpected list: %#v", list)
	}
}

func TestInvoiceFinalizeAndPDF(t *testing.T) {
	db := setupInvoiceTestDB(t)
	user, company, client, product := seedInvoiceFixtures(t, db)
	h := NewInvoiceHandler(db, services.NewInvoiceService())

	// Create invoice
	body := `{"client_id":` + strconv.Itoa(int(client.ID)) + `,"items":[{"product_id":` + strconv.Itoa(int(product.ID)) + `,"quantity":1}]}`
	req := httptest.NewRequest(http.MethodPost, "/invoices", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(auth.WithUserID(req.Context(), user.ID))
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create expected 201 got %d", w.Code)
	}
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	id := int(created["id"].(float64))

	// Finalize
	finReq := httptest.NewRequest(http.MethodPost, "/invoices/finalize?id="+strconv.Itoa(id), nil)
	finW := httptest.NewRecorder()
	h.Finalize(finW, finReq)
	if finW.Code != http.StatusOK {
		t.Fatalf("finalize expected 200 got %d", finW.Code)
	}

	// PDF
	pdfReq := httptest.NewRequest(http.MethodGet, "/invoices/pdf?id="+strconv.Itoa(id), nil)
	pdfW := httptest.NewRecorder()
	h.PDF(pdfW, pdfReq)
	if pdfW.Code != http.StatusOK {
		t.Fatalf("pdf expected 200 got %d", pdfW.Code)
	}
	if ct := pdfW.Header().Get("Content-Type"); !strings.Contains(ct, "application/pdf") {
		t.Fatalf("expected pdf content-type got %s", ct)
	}

	// Finalize should fail for empty invoice
	empty := models.Invoice{Status: "draft", CompanyID: company.ID, ClientID: client.ID}
	if err := db.Create(&empty).Error; err != nil {
		t.Fatalf("insert empty inv: %v", err)
	}
	badReq := httptest.NewRequest(http.MethodPost, "/invoices/finalize?id="+strconv.Itoa(int(empty.ID)), nil)
	badW := httptest.NewRecorder()
	h.Finalize(badW, badReq)
	if badW.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty invoice got %d", badW.Code)
	}
}

func TestInvoiceFinalizeBlockedIfProductSoftDeleted(t *testing.T) {
	db := setupInvoiceTestDB(t)
	user, _, client, product := seedInvoiceFixtures(t, db)
	h := NewInvoiceHandler(db, services.NewInvoiceService())

	// Create invoice
	body := `{"client_id":` + strconv.Itoa(int(client.ID)) + `,"items":[{"product_id":` + strconv.Itoa(int(product.ID)) + `,"quantity":1}]}`
	req := httptest.NewRequest(http.MethodPost, "/invoices", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(auth.WithUserID(req.Context(), user.ID))
	w := httptest.NewRecorder()
	h.Create(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create got %d body=%s", w.Code, w.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &created)
	id := int(created["id"].(float64))

	// Soft delete the product
	if err := db.Where("id = ?", product.ID).Delete(&models.Product{}).Error; err != nil {
		t.Fatalf("soft delete product: %v", err)
	}
	// Try finalize
	finReq := httptest.NewRequest(http.MethodPost, "/invoices/finalize?id="+strconv.Itoa(id), nil)
	finW := httptest.NewRecorder()
	h.Finalize(finW, finReq)
	if finW.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when product soft-deleted, got %d", finW.Code)
	}
}
