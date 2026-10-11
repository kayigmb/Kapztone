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
	"github.com/kayigmb/chaintrack/backend/internal/cache"
	"github.com/kayigmb/chaintrack/backend/internal/config"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
	"github.com/kayigmb/chaintrack/backend/internal/server"
	"github.com/kayigmb/chaintrack/backend/internal/template"
)

// How long access lookups may live in the cache if nothing invalidates them.
const cacheTTL = 60 * time.Second

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(cfg.UploadDir, 0o700); err != nil {
		log.Fatalf("upload directory: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// The pool connects lazily, so the API starts even if the database is briefly down.
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database config: %v", err)
	}
	defer pool.Close()

	c, err := cache.New(cfg.RedisURL, cacheTTL)
	if err != nil {
		log.Fatalf("redis config: %v", err)
	}
	defer c.Close()
	if err := c.Ping(ctx); err != nil {
		log.Printf("redis unavailable, serving from the database: %v", err)
	}

	pg := auth.PgStore{Pool: pool}
	store := auth.CachedStore{Store: pg, Cache: c}

	// Code is the reference for roles and permissions; make the database match.
	if err := auth.SyncRoles(ctx, pool); err != nil {
		log.Fatal(err)
	}
	store.InvalidateAllScopes(ctx)

	created, err := pg.EnsureSuperadmin(ctx, cfg.SuperadminEmail, cfg.SuperadminPassword)
	if err != nil {
		log.Fatalf("seed superadmin: %v", err)
	}
	if created {
		log.Printf("created superadmin %s", cfg.SuperadminEmail)
	}

	a := &auth.Auth{
		Store:    store,
		Secret:   cfg.JWTSecret,
		TokenTTL: cfg.TokenTTL,
		Dev:      cfg.Dev,
		Secrets:  cfg.Secrets,
		Limiter:  &auth.Limiter{Cache: c},
	}
	o := &org.Handlers{
		Store:         org.Store{Pool: pool},
		UploadDir:     cfg.UploadDir,
		InvalidateOrg: store.InvalidateOrg,
		Secrets:       cfg.Secrets,
	}

	srv := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: server.NewRouter(server.Deps{
			Ping:     pool.Ping,
			Auth:     a,
			Org:      o,
			Facility: &facility.Handlers{Store: facility.Store{Pool: pool}},
			Template: &template.Handlers{Store: template.Store{Pool: pool}},
			Passport: &passport.Handlers{Store: passport.Store{Pool: pool, Secrets: cfg.Secrets, Domain: cfg.Domain()}},
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Printf("api listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
