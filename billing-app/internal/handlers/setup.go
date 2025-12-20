package handlers

import (
	"github.com/diewo77/billing-app/auth"
	"github.com/diewo77/billing-app/httpx"
	"github.com/diewo77/billing-app/i18n"
	"github.com/diewo77/billing-app/internal/services"
	"github.com/diewo77/billing-app/view"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

type SetupHandler struct {
	Service *services.SetupService
	mu      sync.RWMutex // kept if future caching needed
}

func NewSetupHandler(s *services.SetupService) *SetupHandler { return &SetupHandler{Service: s} }

func (h *SetupHandler) Register(mux *http.ServeMux) { mux.HandleFunc("/setup", h.handle) }

// Handle exported wrapper for integration when composing custom middleware chains.
func (h *SetupHandler) Handle(w http.ResponseWriter, r *http.Request) { h.handle(w, r) }

type setupRequest struct {
	Company           string  `json:"company"`
	Address1          string  `json:"address1"`
	Address2          string  `json:"address2"`
	PostalCode        string  `json:"postal_code"`
	City              string  `json:"city"`
	Country           string  `json:"country"`
	SIRET             string  `json:"siret"`
	VATEnabled        bool    `json:"vat_enabled"`
	VATRate           float64 `json:"vat_rate"`
	BillingAddress1   string  `json:"billing_address1"`
	BillingAddress2   string  `json:"billing_address2"`
	BillingPostalCode string  `json:"billing_postal_code"`
	BillingCity       string  `json:"billing_city"`
	BillingCountry    string  `json:"billing_country"`
}

// validateSetup normalises request values and returns field -> error code.
// Codes: required, siret_length, siret_digits, tva_rate_invalid
func validateSetup(req *setupRequest, separate bool) map[string]string {
	fe := make(map[string]string)
	req.Company = strings.TrimSpace(req.Company)
	req.Address1 = strings.TrimSpace(req.Address1)
	req.Address2 = strings.TrimSpace(req.Address2)
	req.PostalCode = strings.TrimSpace(req.PostalCode)
	req.City = strings.TrimSpace(req.City)
	req.Country = strings.ToUpper(strings.TrimSpace(req.Country))
	req.SIRET = strings.TrimSpace(req.SIRET)
	req.BillingAddress1 = strings.TrimSpace(req.BillingAddress1)
	req.BillingAddress2 = strings.TrimSpace(req.BillingAddress2)
	req.BillingPostalCode = strings.TrimSpace(req.BillingPostalCode)
	req.BillingCity = strings.TrimSpace(req.BillingCity)
	req.BillingCountry = strings.ToUpper(strings.TrimSpace(req.BillingCountry))

	if req.Company == "" {
		fe["company"] = "required"
	}
	if req.Address1 == "" {
		fe["address"] = "required"
	}
	if req.PostalCode == "" {
		fe["postal_code"] = "required"
	}
	if req.City == "" {
		fe["city"] = "required"
	}
	if req.Country == "" {
		fe["country"] = "required"
	}
	if req.SIRET == "" {
		fe["siret"] = "required"
	} else {
		if len(req.SIRET) != 14 {
			fe["siret"] = "siret_length"
		} else {
			for _, r := range req.SIRET {
				if r < '0' || r > '9' {
					fe["siret"] = "siret_digits"
					break
				}
			}
		}
	}

	if separate {
		if req.BillingAddress1 == "" {
			fe["billing_address1"] = "required"
		}
		if req.BillingPostalCode == "" {
			fe["billing_postal_code"] = "required"
		}
		if req.BillingCity == "" {
			fe["billing_city"] = "required"
		}
		if req.BillingCountry == "" {
			fe["billing_country"] = "required"
		}
	} else {
		req.BillingAddress1 = req.Address1
		req.BillingAddress2 = req.Address2
		req.BillingPostalCode = req.PostalCode
		req.BillingCity = req.City
		req.BillingCountry = req.Country
	}

	if req.VATEnabled {
		if req.VATRate <= 0 || req.VATRate > 1 {
			fe["tva_rate"] = "tva_rate_invalid"
		}
	} else {
		req.VATRate = 0
	}

	return fe
}

// renderSetup renders the setup template with provided data using cached template.
func (h *SetupHandler) renderSetup(w http.ResponseWriter, r *http.Request, data map[string]any, status int) {
	if status != 0 {
		w.WriteHeader(status)
	}
	if err := view.Render(w, r, "setup.html", data); err != nil {
		log.Println("setup render error:", err)
		http.Error(w, "template error", http.StatusInternalServerError)
	}
}

// handle main setup logic.
func (h *SetupHandler) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		configured, err := h.Service.IsConfigured()
		if err != nil {
			httpx.JSONError(w, http.StatusInternalServerError, "db_error", err.Error())
			return
		}
		w.Header().Set("X-Setup-Configured", strconv.FormatBool(configured))
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		// JSON preference
		if strings.Contains(r.Header.Get("Accept"), "application/json") && !strings.Contains(r.Header.Get("Accept"), "text/html") {
			httpx.JSON(w, http.StatusOK, map[string]bool{"configured": configured})
			return
		}
		uid, _ := auth.UserIDFromContext(r.Context())
		data := map[string]any{"Configured": configured, "UserID": uid}
		if configured {
			if cs, err := h.Service.Get(); err == nil && cs != nil {
				values := map[string]string{
					"company":     cs.RaisonSociale,
					"address":     cs.Address.Ligne1,
					"address2":    cs.Address.Ligne2,
					"postal_code": cs.Address.CodePostal,
					"city":        cs.Address.Ville,
					"country":     cs.Address.Pays,
					"siret":       cs.SIRET,
					"tva_rate": func() string {
						if cs.RedevableTVA && cs.TVA > 0 {
							return strconv.FormatFloat(cs.TVA*100, 'f', 2, 64)
						}
						return ""
					}(),
				}
				billingSeparate := cs.AddressID != cs.BillingAddressID
				if billingSeparate {
					values["billing_address1"] = cs.BillingAddress.Ligne1
					values["billing_address2"] = cs.BillingAddress.Ligne2
					values["billing_postal_code"] = cs.BillingAddress.CodePostal
					values["billing_city"] = cs.BillingAddress.Ville
					values["billing_country"] = cs.BillingAddress.Pays
				}
				data["Values"] = values
				data["VATEnabled"] = cs.RedevableTVA
				data["BillingSeparate"] = billingSeparate
			}
		}
		h.renderSetup(w, r, data, 0)
		return
	}
	if r.Method != http.MethodPost {
		httpx.JSONError(w, http.StatusMethodNotAllowed, "method_not_allowed", "")
		return
	}

	ct := r.Header.Get("Content-Type")
	var req setupRequest
	if strings.HasPrefix(ct, "application/json") || ct == "" {
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			if strings.HasPrefix(ct, "application/json") {
				httpx.JSONError(w, http.StatusBadRequest, "invalid_json", err.Error())
				return
			}
		}
	}
	if req.Company == "" && (strings.HasPrefix(ct, "application/x-www-form-urlencoded") || strings.HasPrefix(ct, "multipart/form-data") || ct == "") {
		_ = r.ParseForm()
		req.Company = r.FormValue("company")
		req.Address1 = r.FormValue("address")
		req.Address2 = r.FormValue("address2")
		req.PostalCode = r.FormValue("postal_code")
		req.City = r.FormValue("city")
		req.Country = r.FormValue("country")
		req.SIRET = r.FormValue("siret")
		req.BillingAddress1 = r.FormValue("billing_address1")
		req.BillingAddress2 = r.FormValue("billing_address2")
		req.BillingPostalCode = r.FormValue("billing_postal_code")
		req.BillingCity = r.FormValue("billing_city")
		req.BillingCountry = r.FormValue("billing_country")
		if v := r.FormValue("tva"); v != "" {
			req.VATEnabled = strings.ToLower(v) == "oui" || strings.ToLower(v) == "yes" || strings.ToLower(v) == "true"
		}
		if rate := r.FormValue("tva_rate"); rate != "" {
			if f, err := strconv.ParseFloat(rate, 64); err == nil {
				req.VATRate = f / 100.0
			}
		}
	}

	// Determine separate billing for JSON (if any billing fields and different from main)
	separate := r.FormValue("billing_separate") == "1"
	if !separate && (req.BillingAddress1 != "" || req.BillingPostalCode != "" || req.BillingCity != "" || req.BillingCountry != "") {
		if req.BillingAddress1 != req.Address1 || req.BillingPostalCode != req.PostalCode || req.BillingCity != req.City || !strings.EqualFold(req.BillingCountry, req.Country) {
			separate = true
		}
	}

	lang := i18n.DetectLanguage(r.Header.Get("Accept-Language"))
	fieldErrors := validateSetup(&req, separate)
	isForm := strings.HasPrefix(ct, "application/x-www-form-urlencoded") || strings.HasPrefix(ct, "multipart/form-data")
	if len(fieldErrors) > 0 {
		localized := make(map[string]string, len(fieldErrors))
		for k, v := range fieldErrors {
			localized[k] = i18n.T(lang, v)
		}
		if isForm {
			configured, _ := h.Service.IsConfigured()
			values := map[string]string{
				"company":             req.Company,
				"address":             req.Address1,
				"address2":            req.Address2,
				"postal_code":         req.PostalCode,
				"city":                req.City,
				"country":             req.Country,
				"siret":               req.SIRET,
				"billing_address1":    req.BillingAddress1,
				"billing_address2":    req.BillingAddress2,
				"billing_postal_code": req.BillingPostalCode,
				"billing_city":        req.BillingCity,
				"billing_country":     req.BillingCountry,
				"tva_rate": func() string {
					if req.VATRate > 0 {
						return strconv.FormatFloat(req.VATRate*100, 'f', 2, 64)
					}
					return ""
				}(),
			}
			h.renderSetup(w, r, map[string]any{"Configured": configured, "Values": values, "FieldErrors": localized, "VATEnabled": req.VATEnabled, "BillingSeparate": separate}, http.StatusBadRequest)
			return
		}
		httpx.JSON(w, http.StatusBadRequest, map[string]any{"error": "validation_error", "fields": localized, "lang": lang})
		return
	}

	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		if isForm {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		httpx.JSONError(w, http.StatusUnauthorized, "unauthorized", "login required")
		return
	}

	input := services.SetupInput{Company: req.Company, Address1: req.Address1, Address2: req.Address2, PostalCode: req.PostalCode, City: req.City, Country: req.Country, SIRET: req.SIRET, VATEnabled: req.VATEnabled, VATRate: req.VATRate, UserID: uid, BillingAddress1: req.BillingAddress1, BillingAddress2: req.BillingAddress2, BillingPostalCode: req.BillingPostalCode, BillingCity: req.BillingCity, BillingCountry: req.BillingCountry}
	configured, _ := h.Service.IsConfigured()
	var err error
	var id uint
	if configured {
		updated, uerr := h.Service.Update(input)
		err = uerr
		if updated != nil {
			id = updated.ID
		}
	} else {
		created, cerr := h.Service.Run(input)
		err = cerr
		if created != nil {
			id = created.ID
		}
	}
	if err != nil {
		if err == services.ErrAlreadyConfigured { // should not happen now
			if isForm {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			httpx.JSONError(w, http.StatusConflict, "already_configured", "setup already completed")
			return
		}
		httpx.JSONError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	if isForm {
		msg := "Configuration mise à jour"
		if !configured {
			msg = "Configuration initiale terminée"
		}
		http.SetCookie(w, &http.Cookie{Name: "flash", Value: url.QueryEscape(msg), Path: "/", MaxAge: 15})
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"ok": true, "configured": true, "id": id})
}
