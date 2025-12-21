package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// DB-backed check for user existence
// We import models and gorm only here to keep the rest of the package lightweight.
// This allows us to verify that the session refers to a real user on each request.

type ctxKey string

const (
	sessionCookieName = "session"
	userIDCtxKey      = ctxKey("userID")
)

// UserVerifier is an optional callback to validate that a session's user still exists/is allowed.
// Set it during app bootstrap via SetUserVerifier. If nil, no extra verification is performed.
type UserVerifier func(ctx context.Context, uid uint) bool

var verifier UserVerifier

// SetUserVerifier configures the global verifier used by RequireAuth.
func SetUserVerifier(v UserVerifier) { verifier = v }

// Secret returns SESSION_SECRET or default dev value.
func Secret() string {
	if s := os.Getenv("SESSION_SECRET"); s != "" {
		return s
	}
	return "devsessionsecret"
}

// CreateSession sets a signed cookie with the user id.
func CreateSession(w http.ResponseWriter, userID uint) {
	uidStr := strconv.FormatUint(uint64(userID), 10)
	mac := hmac.New(sha256.New, []byte(Secret()))
	mac.Write([]byte(uidStr))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	value := uidStr + "." + sig
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(14 * 24 * time.Hour),
	})
}

// ClearSession deletes the session cookie.
func ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: "", Path: "/", Expires: time.Unix(0, 0), HttpOnly: true, SameSite: http.SameSiteLaxMode})
}

// ParseSession validates cookie and returns user id.
func ParseSession(r *http.Request) (uint, bool) {
	c, err := r.Cookie(sessionCookieName)
	if err != nil || c.Value == "" {
		return 0, false
	}
	parts := strings.Split(c.Value, ".")
	if len(parts) != 2 {
		return 0, false
	}
	uidStr, sig := parts[0], parts[1]
	mac := hmac.New(sha256.New, []byte(Secret()))
	mac.Write([]byte(uidStr))
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return 0, false
	}
	id64, err := strconv.ParseUint(uidStr, 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(id64), true
}

// WithUserID stores user id in context.
func WithUserID(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, userIDCtxKey, userID)
}

// UserIDFromContext extracts user id.
func UserIDFromContext(ctx context.Context) (uint, bool) {
	v := ctx.Value(userIDCtxKey)
	if v == nil {
		return 0, false
	}
	id, ok := v.(uint)
	return id, ok
}

// Middleware attaches user id to request context if present.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if uid, ok := ParseSession(r); ok {
			r = r.WithContext(WithUserID(r.Context(), uid))
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAuth redirects to /login if not authenticated (HTML) or returns 401 JSON.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if uid, ok := UserIDFromContext(r.Context()); !ok {
			accept := r.Header.Get("Accept")
			if strings.Contains(accept, "application/json") && !strings.Contains(accept, "text/html") {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"error":"unauthorized"}`)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		} else if verifier != nil && !verifier(r.Context(), uid) {
			// Session refers to a non-existing/disabled user: clear and treat as unauthorized.
			ClearSession(w)
			accept := r.Header.Get("Accept")
			if strings.Contains(accept, "application/json") && !strings.Contains(accept, "text/html") {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"error":"unauthorized"}`)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
