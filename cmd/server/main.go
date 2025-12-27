package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/diewo77/go-invoices/auth"
	"github.com/diewo77/go-invoices/internal/config"
	"github.com/diewo77/go-invoices/internal/db"
	"github.com/diewo77/go-invoices/internal/models"
	"github.com/diewo77/go-invoices/internal/policy"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	migrateOnlyFlag = flag.Bool("migrate-only", false, "Run DB migrations and exit")
	seedOnlyFlag    = flag.Bool("seed-only", false, "Run DB seed and exit")
)

func main() {
	flag.Parse()

	// Load environment variables from .env file
	_ = godotenv.Load()

	// Load configuration from environment
	cfg := config.Load()

	// Connect to database using config struct
	dbConn, err := connectDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Handle migrate-only flag
	if *migrateOnlyFlag {
		if err := db.Migrate(dbConn); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migrations completed successfully")
		return
	}

	// Handle seed-only flag
	if *seedOnlyFlag {
		if err := db.Seed(dbConn); err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}
		log.Println("Seeding completed successfully")
		return
	}

	// Run migrations on startup if enabled
	if cfg.App.Migrations {
		if err := db.Migrate(dbConn); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migrations completed")
	}

	// Seed default data (profiles, permissions)
	if err := db.Seed(dbConn); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	// Configure auth verifier to check if user exists in DB
	auth.SetUserVerifier(func(ctx context.Context, uid uint) bool {
		var count int64
		dbConn.Model(&models.User{}).Where("id = ?", uid).Count(&count)
		return count > 0
	})

	// Create router config with authorization
	routerCfg := policy.NewRouterConfig(dbConn)

	// Create application handler
	appHandler := NewApp(dbConn, routerCfg)

	// Create server with config timeouts
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      withLogging(appHandler),
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s (dev=%v)", cfg.Server.Port, cfg.App.Dev)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
	log.Println("Server stopped gracefully")
}

// connectDB establishes a connection to the PostgreSQL database using config.
func connectDB(dbCfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := dbCfg.DSN()
	log.Printf("Connecting to database: host=%s port=%d dbname=%s user=%s",
		dbCfg.Host, dbCfg.Port, dbCfg.DBName, dbCfg.User)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// withLogging adds request logging middleware.
func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
