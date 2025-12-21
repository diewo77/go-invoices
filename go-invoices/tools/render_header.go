package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"html/template"

	"github.com/diewo77/go-invoices/view"
)

func main() {
	// Set permissive resolvers for local test
	view.SetCanProfileResolver(func(r *http.Request, resource, action string) bool { return true })
	view.SetIsAdminResolver(func(r *http.Request) bool { return true })

	// Create a dummy request with context
	req := &http.Request{Header: http.Header{}, URL: &url.URL{Path: "/"}}

	// Parse the header partial using view.Funcs
	tplPath := "templates/partials/header.html"
	funcMap := view.Funcs(req)
	parsed, err := template.New("header.html").Funcs(funcMap).ParseFiles(tplPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(2)
	}

	var buf bytes.Buffer
	data := map[string]any{"IsLoggedIn": true}
	if err := parsed.Execute(&buf, data); err != nil {
		fmt.Fprintf(os.Stderr, "exec error: %v\n", err)
		os.Exit(3)
	}

	fmt.Print(buf.String())
}
