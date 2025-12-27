package db

import (
	"github.com/diewo77/go-invoices/internal/models"
	"gorm.io/gorm"
)

// SeedPermissions creates the core permissions for the application.
// Called during initial database setup or migration.
func SeedPermissions(db *gorm.DB) error {
	// Define all resource:action pairs for the application
	permissions := []struct {
		ResourceType string
		Action       string
		Description  string
	}{
		// Superadmin wildcard
		{"*", "*", "Full system access"},
		// Product permissions
		{"product", "*", "All product actions"},
		{"product", "list", "List products"},
		{"product", "view", "View product details"},
		{"product", "create", "Create products"},
		{"product", "update", "Edit products"},
		{"product", "delete", "Delete products"},
		// Invoice permissions
		{"invoice", "*", "All invoice actions"},
		{"invoice", "list", "List invoices"},
		{"invoice", "view", "View invoice details"},
		{"invoice", "create", "Create invoices"},
		{"invoice", "update", "Edit invoices"},
		{"invoice", "delete", "Delete invoices"},
		{"invoice", "finalize", "Finalize invoices"},
		// Client permissions
		{"client", "*", "All client actions"},
		{"client", "list", "List clients"},
		{"client", "view", "View client details"},
		{"client", "create", "Create clients"},
		{"client", "update", "Edit clients"},
		{"client", "delete", "Delete clients"},
		// Company settings
		{"company", "*", "All company settings"},
		{"company", "view", "View company settings"},
		{"company", "update", "Edit company settings"},
		// User management
		{"user", "*", "All user management"},
		{"user", "list", "List users"},
		{"user", "view", "View user details"},
		{"user", "update", "Edit users"},
		// Profile management (admin only)
		{"profile", "*", "All profile management"},
		{"profile", "list", "List profiles"},
		{"profile", "view", "View profile details"},
		{"profile", "create", "Create profiles"},
		{"profile", "update", "Edit profiles"},
		{"profile", "delete", "Delete profiles"},
		// Product type management
		{"product_type", "*", "All product type actions"},
		{"product_type", "list", "List product types"},
		{"product_type", "view", "View product type details"},
		{"product_type", "create", "Create product types"},
		{"product_type", "update", "Edit product types"},
		{"product_type", "delete", "Delete product types"},
		// Unit type management
		{"unit_type", "*", "All unit type actions"},
		{"unit_type", "list", "List unit types"},
		{"unit_type", "view", "View unit type details"},
		{"unit_type", "create", "Create unit types"},
		{"unit_type", "update", "Edit unit types"},
		{"unit_type", "delete", "Delete unit types"},
	}

	for _, p := range permissions {
		perm := models.Permission{
			ResourceType: p.ResourceType,
			Action:       p.Action,
			Description:  p.Description,
		}
		// Use FirstOrCreate to avoid duplicates
		result := db.Where("resource_type = ? AND action = ?", p.ResourceType, p.Action).
			FirstOrCreate(&perm)
		if result.Error != nil {
			return result.Error
		}
	}
	return nil
}

// SeedProfiles creates the default system profiles with their permissions.
func SeedProfiles(db *gorm.DB) error {
	// First ensure permissions exist
	if err := SeedPermissions(db); err != nil {
		return err
	}

	profiles := []struct {
		Name        string
		Description string
		IsSystem    bool
		Permissions []string // "resource:action" format
	}{
		{
			Name:        "admin",
			Description: "Full system administrator with all permissions",
			IsSystem:    true,
			Permissions: []string{"*:*"},
		},
		{
			Name:        "viewer",
			Description: "Read-only access to all resources",
			IsSystem:    true,
			Permissions: []string{
				"product:list",
				"product:view",
				"invoice:list",
				"invoice:view",
				"client:list",
				"client:view",
				"company:view",
				"product_type:list",
				"product_type:view",
				"unit_type:list",
				"unit_type:view",
			},
		},
		{
			Name:        "accountant",
			Description: "Manage invoices and clients, view products",
			IsSystem:    true,
			Permissions: []string{
				"invoice:*",
				"client:*",
				"product:list",
				"product:view",
				"company:view",
			},
		},
	}

	for _, p := range profiles {
		var profile models.Profile
		result := db.Where("name = ?", p.Name).First(&profile)
		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
			return result.Error
		}

		// If profile doesn't exist, create it
		if result.Error == gorm.ErrRecordNotFound {
			profile = models.Profile{
				Name:        p.Name,
				Description: p.Description,
				IsSystem:    p.IsSystem,
			}
			if err := db.Create(&profile).Error; err != nil {
				return err
			}
		}

		// Assign permissions
		var perms []models.Permission
		for _, code := range p.Permissions {
			// Split "resource:action"
			var resource, action string
			for i := 0; i < len(code); i++ {
				if code[i] == ':' {
					resource = code[:i]
					action = code[i+1:]
					break
				}
			}
			var perm models.Permission
			if err := db.Where("resource_type = ? AND action = ?", resource, action).First(&perm).Error; err == nil {
				perms = append(perms, perm)
			}
		}
		if err := db.Model(&profile).Association("Permissions").Replace(perms); err != nil {
			return err
		}
	}
	return nil
}
