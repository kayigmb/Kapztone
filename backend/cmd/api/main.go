package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/attestation"
	"github.com/kayigmb/chaintrack/backend/internal/audit"
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/cache"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/config"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/flag"
	"github.com/kayigmb/chaintrack/backend/internal/logx"
	"github.com/kayigmb/chaintrack/backend/internal/notify"
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
	"github.com/kayigmb/chaintrack/backend/internal/server"
	"github.com/kayigmb/chaintrack/backend/internal/stats"
	"github.com/kayigmb/chaintrack/backend/internal/template"
	"github.com/kayigmb/chaintrack/backend/internal/upload"
	"github.com/kayigmb/chaintrack/backend/internal/user"
	"github.com/kayigmb/chaintrack/backend/internal/verify"
)

// How long access lookups may live in the cache if nothing invalidates them.
const cacheTTL = 60 * time.Second

func main() {
	format, level := config.Logging()
	if _, err := logx.Setup(os.Stderr, format, level); err != nil {
		logx.Fatal("log settings", err)
	}
	cfg, err := config.Load()
	if err != nil {
		logx.Fatal("settings", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Uploaded files live in a bucket. In development a missing one is created.
	files, err := upload.New(ctx, cfg.Storage, cfg.Dev)
	if err != nil {
		logx.Fatal("start-up", err)
	}
	// The pool connects lazily, so the API starts even if the database is briefly down.
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logx.Fatal("database config", err)
	}
	defer pool.Close()
	c := openCache(ctx, cfg.RedisURL)
	defer c.Close()
	client, contracts := dialChain(ctx, cfg)
	defer client.Close()

	pg := auth.PgStore{Pool: pool}
	store := auth.CachedStore{Store: pg, Cache: c}
	seed(ctx, pool, pg, store, cfg)
	a := newAuth(cfg, pg, store, c)
	deps := handlers(ctx, cfg, pool, files, c, a, store, contracts)
	serve(ctx, cfg, server.NewRouter(deps))
}

func openCache(ctx context.Context, redisURL string) *cache.Cache {
	c, err := cache.New(redisURL, cacheTTL)
	if err != nil {
		logx.Fatal("redis config", err)
	}
	if err := c.Ping(ctx); err != nil {
		slog.Warn("redis unavailable, serving from the database", logx.KeyError, err)
	}
	return c
}

// dialChain binds the contracts public verification reads. Dial does not need
// the chain to be up; while it is down, verification answers with a warning.
func dialChain(ctx context.Context, cfg config.Config) (*ethclient.Client, chain.Contracts) {
	client, err := ethclient.DialContext(ctx, cfg.ChainRPCURL)
	if err != nil {
		logx.Fatal("chain config", err)
	}
	contracts, err := chain.Bind(cfg.Contracts, client)
	if err != nil {
		logx.Fatal("contracts", err)
	}
	return client, contracts
}

// seed makes the database's roles match the code, which is their reference,
// and creates the first superadmin if there is none.
func seed(ctx context.Context, pool *pgxpool.Pool, pg auth.PgStore, store auth.CachedStore, cfg config.Config) {
	if err := auth.SyncRoles(ctx, pool); err != nil {
		logx.Fatal("start-up", err)
	}
	store.InvalidateAllScopes(ctx)
	created, err := pg.EnsureSuperadmin(ctx, cfg.SuperadminEmail, cfg.SuperadminPassword)
	if err != nil {
		logx.Fatal("seed superadmin", err)
	}
	if created {
		slog.Info("created superadmin", "email", cfg.SuperadminEmail)
	}
}

func newAuth(cfg config.Config, pg auth.PgStore, store auth.CachedStore, c *cache.Cache) *auth.Auth {
	return &auth.Auth{
		Store:       store,
		Secret:      cfg.JWTSecret,
		TokenTTL:    cfg.TokenTTL,
		RememberTTL: cfg.RememberTTL,
		AppURL:      cfg.AppURL,
		Accounts:    pg,
		Dev:         cfg.Dev,
		Secrets:     cfg.Secrets,
		Limiter:     limiter(cfg.Dev, c),
		LogoURL:     cfg.MailLogoURL,
	}
}

func handlers(ctx context.Context, cfg config.Config, pool *pgxpool.Pool, files *upload.Bucket, c *cache.Cache,
	a *auth.Auth, store auth.CachedStore, contracts chain.Contracts) server.Deps {
	passports := passport.Store{Pool: pool, Secrets: cfg.Secrets, Domain: cfg.Domain()}
	certificates := certificateHandlers(ctx, cfg, pool, files, a)
	users := &user.Handlers{Store: user.Store{Pool: pool}, InvalidateUser: store.InvalidateUser, AppURL: cfg.AppURL, Dev: cfg.Dev, LogoURL: cfg.MailLogoURL}
	startMail(ctx, cfg, pool, a, users)
	return server.Deps{
		Ping: pool.Ping,
		Auth: a,
		Org: &org.Handlers{Store: org.Store{Pool: pool}, Files: files, InvalidateOrg: store.InvalidateOrg,
			Secrets: cfg.Secrets, SendVerify: a.SendVerifyTo, AppURL: cfg.AppURL},
		Facility: &facility.Handlers{Store: facility.Store{Pool: pool}, Files: files},
		Template: &template.Handlers{Store: template.Store{Pool: pool}},
		Passport: &passport.Handlers{Store: passports, PIN: a.CheckPIN, Files: files,
			// Rejected handovers and the rules flag passports, in the same transaction.
			RaiseFlag: flag.RaiseFlag},
		Attestation: certificates,
		Flag:        &flag.Handlers{Store: flag.Store{Pool: pool}, Passports: passports, Auth: a},
		User:        users,
		Notify:      &notify.Handlers{Pool: pool},
		Verify: &verify.Handlers{Pool: pool, Passports: passports, Certificates: certificates.Store,
			Contracts: contracts, Files: files, Cache: c},
		Audit: &audit.Handlers{Store: audit.Store{Pool: pool}, Files: files},
		Stats: &stats.Handlers{Store: stats.Store{Pool: pool}},
	}
}

// certificateHandlers also starts the IPFS retry loop when IPFS copies are on.
func certificateHandlers(ctx context.Context, cfg config.Config, pool *pgxpool.Pool, files *upload.Bucket, a *auth.Auth) *attestation.Handlers {
	h := &attestation.Handlers{Store: attestation.Store{Pool: pool}, Files: files, PIN: a.CheckPIN}
	if cfg.PinataJWT != "" {
		h.Pin = attestation.Pinata{JWT: cfg.PinataJWT, URL: attestation.PinataURL}.Pin
		retrier := &attestation.Retrier{Store: h.Store, Files: files, Pin: h.Pin}
		go retrier.Run(ctx)
	}
	return h
}

func startMail(ctx context.Context, cfg config.Config, pool *pgxpool.Pool, a *auth.Auth, users *user.Handlers) {
	if cfg.SMTP == nil {
		return
	}
	mailer := notify.NewMailer(pool, *cfg.SMTP)
	mailer.LogoURL = cfg.MailLogoURL
	users.Mail = mailer.SendOne
	a.Mail = mailer.SendOne
	go mailer.Run(ctx)
}

// serve runs until ctx ends, then gives open requests 10 seconds to finish.
func serve(ctx context.Context, cfg config.Config, h http.Handler) {
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		// Generous for photo uploads on slow links. No WriteTimeout: PDF and
		// CSV exports may take a while.
		ReadTimeout: 2 * time.Minute,
		IdleTimeout: 2 * time.Minute,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Warn("shutdown", logx.KeyError, err)
		}
	}()
	slog.Info("api listening", "port", cfg.Port, "mode", server.Mode(cfg.Dev), "rate_limits", !cfg.Dev, "chain_id", cfg.ChainID.String(),
		"bucket", cfg.Storage.Bucket, "ipfs", cfg.PinataJWT != "", "smtp", cfg.SMTP != nil)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logx.Fatal("start-up", err)
	}
}

// limiter counts sign-in and sign-up attempts in production; development has
// no rate limits, so testing by hand is never paused.
func limiter(dev bool, c *cache.Cache) *auth.Limiter {
	if dev {
		return nil
	}
	return &auth.Limiter{Cache: c}
}
