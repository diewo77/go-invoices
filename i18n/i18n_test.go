package i18n

import "testing"

func TestDetectLanguage(t *testing.T) {
	if DetectLanguage("en-US,en;q=0.9") != "en" {
		t.Fatalf("expected en")
	}
	if DetectLanguage("EN-gb") != "en" {
		t.Fatalf("expected en for EN-gb")
	}
	if DetectLanguage("fr-FR,fr;q=0.8") != "fr" {
		t.Fatalf("expected fr fallback")
	}
	if DetectLanguage("") != "fr" {
		t.Fatalf("expected default fr")
	}
}

func TestTranslations(t *testing.T) {
	if T("en", "required") != "Required" {
		t.Fatalf("expected Required")
	}
	if T("fr", "required") != "Requis" {
		t.Fatalf("expected Requis")
	}
	// unknown code -> fallback to code
	if T("en", "__nope__") != "__nope__" {
		t.Fatalf("expected fallback to code")
	}
	// unknown language -> fallback to fr translation if exists
	if T("es", "required") != "Requis" {
		t.Fatalf("expected fr fallback for es lang")
	}
}
