package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/server"
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	port := envOr("PORT", "8080")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		log.Fatal("JWT_SECRET is required and must be at least 32 characters")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// The pool connects lazily, so the API starts even if the database is briefly down.
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("database config: %v", err)
	}
	defer pool.Close()

	store := auth.PgStore{Pool: pool}
	if email := os.Getenv("SUPERADMIN_EMAIL"); email != "" {
		created, err := store.EnsureSuperadmin(ctx, email, os.Getenv("SUPERADMIN_PASSWORD"))
		if err != nil {
			log.Fatalf("seed superadmin: %v", err)
		}
		if created {
			log.Printf("created superadmin %s", email)
		}
	}

	a := &auth.Auth{
		Store:      store,
		Secret:     []byte(jwtSecret),
		TokenTTL:   12 * time.Hour,
		BaseDomain: envOr("BASE_DOMAIN", "localhost"),
		Dev:        os.Getenv("APP_ENV") == "dev",
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           server.NewRouter(pool.Ping, a),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Printf("api listening on :%s", port)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
