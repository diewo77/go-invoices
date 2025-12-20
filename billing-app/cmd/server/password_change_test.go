package main

import (
	"github.com/diewo77/billing-app/auth"
	"github.com/diewo77/billing-app/internal/models"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPwdDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:pwd_"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	if err := db.AutoMigrate(&models.Address{}, &models.Role{}, &models.User{}, &models.CompanySettings{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestPasswordChangeSuccess(t *testing.T) {
	db := setupPwdDB(t)
	hash, _ := bcrypt.GenerateFromPassword([]byte("OldPass123"), bcrypt.DefaultCost)
	u := models.User{Email: "pwd@example.com", Password: string(hash)}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("user: %v", err)
	}
	app := NewApp(db)
	rec := httptest.NewRecorder()
	auth.CreateSession(rec, u.ID)
	var sess *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "session" {
			sess = c
			break
		}
	}
	if sess == nil {
		t.Fatalf("no session cookie")
	}
	form := url.Values{}
	form.Set("current", "OldPass123")
	form.Set("new", "NewPass456")
	form.Set("confirm", "NewPass456")
	req := httptest.NewRequest(http.MethodPost, "/profile/password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sess)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect got %d", rr.Code)
	}
	var updated models.User
	if err := db.First(&updated, u.ID).Error; err != nil {
		t.Fatalf("load: %v", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(updated.Password), []byte("NewPass456")) != nil {
		t.Fatalf("password not updated")
	}
}

func TestPasswordChangeWrongCurrent(t *testing.T) {
	db := setupPwdDB(t)
	hash, _ := bcrypt.GenerateFromPassword([]byte("OldPass123"), bcrypt.DefaultCost)
	u := models.User{Email: "pwd2@example.com", Password: string(hash)}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("user: %v", err)
	}
	app := NewApp(db)
	rec := httptest.NewRecorder()
	auth.CreateSession(rec, u.ID)
	var sess *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "session" {
			sess = c
			break
		}
	}
	if sess == nil {
		t.Fatalf("no session cookie")
	}
	form := url.Values{}
	form.Set("current", "WrongPass")
	form.Set("new", "NewPass456")
	form.Set("confirm", "NewPass456")
	req := httptest.NewRequest(http.MethodPost, "/profile/password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(sess)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect got %d", rr.Code)
	}
	var updated models.User
	if err := db.First(&updated, u.ID).Error; err != nil {
		t.Fatalf("load: %v", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(updated.Password), []byte("OldPass123")) != nil {
		t.Fatalf("original password changed unexpectedly")
	}
}
