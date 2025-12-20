package db

import (
	"net/url"
	"os"
	"regexp"
	"strings"
)

var kvPairRegex = regexp.MustCompile(`(?i)\b(host|user|password|dbname|port|sslmode)=`)

// NormalizeDSN accepts either a URL style DSN (postgres://...) or a lib/pq key=value list.
// It trims quotes and whitespace and, if given key=value form, returns it cleaned.
// If given only partial info, it supplements with sensible defaults when possible.
func NormalizeDSN(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.Trim(s, "\"'")
	if s == "" {
		return s
	}
	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "postgres://") || strings.HasPrefix(lower, "postgresql://") {
		return s
	}
	// key=value list expected
	// If it does not look like key=value pairs, return unchanged (driver will error)
	if !kvPairRegex.MatchString(s) {
		return s
	}
	// Collapse multiple spaces
	fields := strings.Fields(s)
	cleaned := strings.Join(fields, " ")
	// Ensure sslmode present (default disable if missing)
	if !strings.Contains(strings.ToLower(cleaned), "sslmode=") {
		cleaned += " sslmode=disable"
	}
	return cleaned
}

// Helper to build a URL style DSN from key=value if URL form is preferred elsewhere.
func ToURLDSN(kvDSN string) string {
	if kvDSN == "" {
		return kvDSN
	}
	if strings.HasPrefix(strings.ToLower(kvDSN), "postgres://") {
		return kvDSN
	}
	// parse minimal parts
	m := map[string]string{}
	for _, part := range strings.Fields(kvDSN) {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			m[strings.ToLower(kv[0])] = kv[1]
		}
	}
	host := m["host"]
	port := m["port"]
	user := m["user"]
	pass := m["password"]
	dbname := m["dbname"]
	if host == "" || user == "" || dbname == "" {
		return kvDSN
	}
	u := &url.URL{Scheme: "postgres", Host: host}
	if port != "" {
		u.Host = host + ":" + port
	}
	if user != "" {
		if pass != "" {
			u.User = url.UserPassword(user, pass)
		} else {
			u.User = url.User(user)
		}
	}
	u.Path = "/" + dbname
	q := url.Values{}
	if sslm, ok := m["sslmode"]; ok {
		q.Set("sslmode", sslm)
	}
	if len(q) > 0 {
		u.RawQuery = q.Encode()
	}
	return u.String()
}

// GetNormalizedDSN fetches DATABASE_DSN env var and normalizes it.
func GetNormalizedDSN() string { return NormalizeDSN(os.Getenv("DATABASE_DSN")) }
