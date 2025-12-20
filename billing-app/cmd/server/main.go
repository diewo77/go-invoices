package main

import (
	"github.com/diewo77/billing-app/internal/config"
	"github.com/diewo77/billing-app/internal/db"
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

// simple middleware chain
func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

var migrateOnlyFlag = flag.Bool("migrate-only", false, "Run DB migrations and exit")

func main() {
	flag.Parse()
	if migrateOnlyFlag != nil && *migrateOnlyFlag {
		_ = godotenv.Load()
		if _, err := db.ConnectAndMigrate(); err != nil {
			log.Fatalf("migrate-only failed: %v", err)
		}
		log.Println("migrations completed; exiting as requested")
		return
	}
	if backfillFlag != nil && *backfillFlag {
		runBackfillProductCodes()
		return
	}
	_ = godotenv.Load()
	cfg := config.Load()
	dbConn, err := db.ConnectAndMigrate()
	if err != nil {
		log.Fatalf("Erreur connexion DB: %v", err)
	}
	log.Printf("Starting server env=%s port=%s", cfg.Env, cfg.Port)

	// Use the consolidated application handler (includes prefs + view rendering)
	appHandler := NewApp(dbConn)
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: withLogging(appHandler)}

	go func() {
		log.Printf("Server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
	log.Println("Server gracefully stopped")
}

// legacy choose helper kept for compatibility with handlers referencing it (if any)
func choose(v, def string) string {
	if v != "" {
		return v
	}
	return def
}
