package services

import (
	"github.com/diewo77/billing-app/internal/models"
	"errors"

	"gorm.io/gorm"
)

type SetupInput struct {
	Company           string
	Address1          string
	Address2          string
	PostalCode        string
	City              string
	Country           string
	SIRET             string
	VATEnabled        bool
	VATRate           float64
	UserID            uint // required: owner user performing setup
	BillingAddress1   string
	BillingAddress2   string
	BillingPostalCode string
	BillingCity       string
	BillingCountry    string
}

type SetupService struct{ DB *gorm.DB }

func NewSetupService(db *gorm.DB) *SetupService { return &SetupService{DB: db} }

var ErrAlreadyConfigured = errors.New("company_already_configured")

func (s *SetupService) IsConfigured() (bool, error) {
	var count int64
	if err := s.DB.Model(&models.CompanySettings{}).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *SetupService) Run(in SetupInput) (*models.CompanySettings, error) {
	configured, err := s.IsConfigured()
	if err != nil {
		return nil, err
	}
	if configured {
		return nil, ErrAlreadyConfigured
	}
	if in.UserID == 0 {
		return nil, errors.New("missing_user_id")
	}
	addr := models.Address{Ligne1: in.Address1, Ligne2: in.Address2, CodePostal: in.PostalCode, Ville: in.City, Pays: in.Country, Type: "principale"}
	if err := s.DB.Create(&addr).Error; err != nil {
		return nil, err
	}

	billingAddrID := addr.ID
	// If billing differs from primary, create separate
	if in.BillingAddress1 != "" && (in.BillingAddress1 != in.Address1 || in.BillingPostalCode != in.PostalCode || in.BillingCity != in.City || in.BillingCountry != in.Country) {
		bAddr := models.Address{Ligne1: in.BillingAddress1, Ligne2: in.BillingAddress2, CodePostal: in.BillingPostalCode, Ville: in.BillingCity, Pays: in.BillingCountry, Type: "facturation"}
		if err := s.DB.Create(&bAddr).Error; err != nil {
			return nil, err
		}
		billingAddrID = bAddr.ID
	}

	var siren string
	if len(in.SIRET) >= 9 {
		siren = in.SIRET[:9]
	}
	cs := models.CompanySettings{UserID: in.UserID, RaisonSociale: in.Company, NomCommercial: in.Company, SIREN: siren, SIRET: in.SIRET, CodeNAF: "0000Z", TVA: in.VATRate, RedevableTVA: in.VATEnabled, FormeJuridique: "SARL", RegimeFiscal: "Réel", TypeImposition: "BIC", FrequenceUrssaf: "mensuelle", AddressID: addr.ID, BillingAddressID: billingAddrID}
	if err := s.DB.Create(&cs).Error; err != nil {
		return nil, err
	}
	return &cs, nil
}

// Get returns the single company settings record if present (with addresses), otherwise nil.
func (s *SetupService) Get() (*models.CompanySettings, error) {
	var cs models.CompanySettings
	err := s.DB.Preload("Address").Preload("BillingAddress").First(&cs).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// Update modifies existing company settings (single company app) with new input values.
// Creates or updates billing address depending on whether it differs from main address.
func (s *SetupService) Update(in SetupInput) (*models.CompanySettings, error) {
	var cs models.CompanySettings
	if err := s.DB.Preload("Address").Preload("BillingAddress").First(&cs).Error; err != nil {
		return nil, err
	}
	// Update main fields
	cs.RaisonSociale = in.Company
	cs.NomCommercial = in.Company
	if len(in.SIRET) >= 9 {
		cs.SIREN = in.SIRET[:9]
	}
	cs.SIRET = in.SIRET
	cs.RedevableTVA = in.VATEnabled
	cs.TVA = in.VATRate

	// Update main address
	if err := s.DB.Model(&models.Address{}).Where("id = ?", cs.AddressID).Updates(models.Address{Ligne1: in.Address1, Ligne2: in.Address2, CodePostal: in.PostalCode, Ville: in.City, Pays: in.Country}).Error; err != nil {
		return nil, err
	}

	separate := in.BillingAddress1 != "" && (in.BillingAddress1 != in.Address1 || in.BillingPostalCode != in.PostalCode || in.BillingCity != in.City || in.BillingCountry != in.Country)
	if separate {
		if cs.BillingAddressID == cs.AddressID || cs.BillingAddressID == 0 { // need new billing address
			bAddr := models.Address{Ligne1: in.BillingAddress1, Ligne2: in.BillingAddress2, CodePostal: in.BillingPostalCode, Ville: in.BillingCity, Pays: in.BillingCountry, Type: "facturation"}
			if err := s.DB.Create(&bAddr).Error; err != nil {
				return nil, err
			}
			cs.BillingAddressID = bAddr.ID
		} else { // update existing billing address
			if err := s.DB.Model(&models.Address{}).Where("id = ?", cs.BillingAddressID).Updates(models.Address{Ligne1: in.BillingAddress1, Ligne2: in.BillingAddress2, CodePostal: in.BillingPostalCode, Ville: in.BillingCity, Pays: in.BillingCountry}).Error; err != nil {
				return nil, err
			}
		}
	} else {
		// unify billing to main
		cs.BillingAddressID = cs.AddressID
	}

	if err := s.DB.Save(&cs).Error; err != nil {
		return nil, err
	}
	// reload associations
	if err := s.DB.Preload("Address").Preload("BillingAddress").First(&cs, cs.ID).Error; err != nil {
		return nil, err
	}
	return &cs, nil
}
