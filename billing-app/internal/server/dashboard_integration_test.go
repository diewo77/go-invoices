package server_test

import (
	"github.com/diewo77/billing-app/auth"
	"github.com/diewo77/billing-app/internal/models"
	srv "github.com/diewo77/billing-app/internal/server"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupFullTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbi, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := dbi.AutoMigrate(&models.Address{}, &models.Role{}, &models.User{}, &models.CompanySettings{}, &models.Product{}, &models.Invoice{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return dbi
}

func extractCookie(rr *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range rr.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// NOTE: /dashboard is currently implemented only in main.go, not in server.New router.
// This test is a placeholder documenting that limitation.
func TestDashboardRouteNotInRouter(t *testing.T) {
	dbi := setupFullTestDB(t)
	root := srv.New(dbi)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	root.ServeHTTP(rr, req)
	// Document current behavior: router does not serve /dashboard; expect 404 or 200 plain text root message depending on trailing slash normalization.
	if rr.Code == http.StatusOK && rr.Body.Len() == 0 {
		t.Fatalf("/dashboard returned 200 with empty body; unexpected")
	}
}

func TestSetupProtectedRedirectOrUnauthorized(t *testing.T) {
	dbi := setupFullTestDB(t)
	root := srv.New(dbi)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/setup", nil)
	root.ServeHTTP(rr, req)
	if rr.Code == http.StatusOK {
		t.Fatalf("expected redirect/unauthorized for /setup without auth, got 200")
	}
}

func TestSessionCookieFormat(t *testing.T) {
	rr := httptest.NewRecorder()
	auth.CreateSession(rr, 7)
	c := extractCookie(rr, "session")
	if c == nil {
		t.Fatalf("missing session cookie")
	}
	if !regexp.MustCompile(`^[0-9]+\.[A-Za-z0-9_-]+$`).MatchString(c.Value) {
		t.Fatalf("bad cookie format: %s", c.Value)
	}
}
