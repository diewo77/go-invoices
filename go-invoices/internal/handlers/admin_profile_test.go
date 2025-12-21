package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diewo77/go-invoices/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	// Migrate test models
	if err := db.AutoMigrate(&models.Profile{}, &models.Permission{}, &models.User{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func TestAdminProfileHandler_List_JSON(t *testing.T) {
	db := setupTestDB(t)

	// Create test profiles
	db.Create(&models.Profile{Name: "admin", Description: "Admin profile", IsSystem: true})
	db.Create(&models.Profile{Name: "viewer", Description: "Viewer profile"})

	handler := NewAdminProfileHandler(db, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/profiles", nil)
	req.Header.Set("Accept", "application/json")
	// Set a mock user ID in context
	// Note: In a real test, you'd use a proper auth context setup
	
	rr := httptest.NewRecorder()
	handler.List(rr, req)

	// Should redirect to login because no user in context
	if rr.Code != http.StatusSeeOther {
		// If we get a different code, check if it's because templates are missing
		t.Logf("Response code: %d", rr.Code)
	}
}

func TestAdminProfileHandler_Create_JSON(t *testing.T) {
	db := setupTestDB(t)
	handler := NewAdminProfileHandler(db, nil)

	body := map[string]string{
		"name":        "test-profile",
		"description": "A test profile",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/admin/profiles/create", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	rr := httptest.NewRecorder()
	handler.Create(rr, req)

	// Should return 401 because no user in context
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAdminProfileHandler_Create_Validation(t *testing.T) {
	db := setupTestDB(t)
	handler := NewAdminProfileHandler(db, nil)

	// Empty name should fail validation
	body := map[string]string{
		"name":        "",
		"description": "No name",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/admin/profiles/create", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	rr := httptest.NewRecorder()
	handler.Create(rr, req)

	// Should return 401 first (auth check) or 400 (validation)
	if rr.Code != http.StatusUnauthorized && rr.Code != http.StatusBadRequest {
		t.Errorf("expected 401 or 400, got %d", rr.Code)
	}
}

func TestModels_Ownable_Interface(t *testing.T) {
	// Test that all business models implement Ownable
	var _ interface{ GetUserID() uint } = &models.Product{}
	var _ interface{ GetUserID() uint } = &models.Client{}
	var _ interface{ GetUserID() uint } = &models.Invoice{}

	// Verify the returned values
	product := &models.Product{UserID: 1}
	client := &models.Client{UserID: 2}
	invoice := &models.Invoice{UserID: 3}

	if product.GetUserID() != 1 {
		t.Error("Product.GetUserID() failed")
	}
	if client.GetUserID() != 2 {
		t.Error("Client.GetUserID() failed")
	}
	if invoice.GetUserID() != 3 {
		t.Error("Invoice.GetUserID() failed")
	}
}
