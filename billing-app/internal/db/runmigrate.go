package db

import (
	"log"
	"os"
)

// RunMigrations is a lightweight entry point you can invoke from tests or a small main.
// It respects the MIGRATIONS env var just like ConnectAndMigrate.
func RunMigrations() error {
	dsn := GetNormalizedDSN()
	if dsn == "" {
		return nil
	}
	if v := os.Getenv("MIGRATIONS"); v == "" {
		log.Println("MIGRATIONS env not set; skipping sql migrations (AutoMigrate path used at app start).")
		return nil
	}
	log.Println("Running explicit SQL migrations...")
	return runSQLMigrations(dsn)
}
