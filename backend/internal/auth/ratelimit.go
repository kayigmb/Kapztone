package auth

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/kayigmb/chaintrack/backend/internal/cache"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
)

// Rate is "at most Max events per Window" for one named action.
type Rate struct {
	Name   string
	Max    int64
	Window time.Duration
}

var (
	// LoginRate counts failed sign-ins per email and address. Without it a
	// 6-digit 2FA code or a weak password could simply be guessed.
	LoginRate = Rate{Name: "login", Max: 5, Window: 15 * time.Minute}
	// RegisterRate counts sign-ups per address.
	RegisterRate = Rate{Name: "register", Max: 10, Window: time.Hour}
	// VerifyRate counts public verifications per address: each one reads the chain.
	VerifyRate = Rate{Name: "verify", Max: 60, Window: time.Minute}
	// PINRate counts wrong signing PINs per user; a 6-digit PIN must not be guessable.
	PINRate = Rate{Name: "pin", Max: 5, Window: 15 * time.Minute}
)

// Limiter counts events in the cache. If the cache is unreachable it allows
// the request: a Redis outage must not lock everyone out.
type Limiter struct {
	Cache *cache.Cache
}

func (rt Rate) key(parts ...string) string {
	return cacheRatePrefix + rt.Name + ":" + strings.Join(parts, ":")
}

// clientIP is the address the request came from. Through the web app it is
// the last X-Forwarded-For entry, which the web app sets to its visitor's
// address. The header is trusted because the API is never public: only the
// web app can reach it (see docs/operations.md). Called directly, it is the
// peer address.
// ponytail: last entry only; with nginx in front of the web app, the web app
// must pass on nginx's visitor address instead of its own peer.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get(HeaderForwardedFor); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[len(parts)-1]); ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (l *Limiter) blocked(ctx context.Context, rt Rate, parts ...string) bool {
	return l != nil && l.Cache.Count(ctx, rt.key(parts...)) >= rt.Max
}

func (l *Limiter) record(ctx context.Context, rt Rate, parts ...string) {
	if l != nil {
		l.Cache.Incr(ctx, rt.key(parts...), rt.Window)
	}
}

func (l *Limiter) reset(ctx context.Context, rt Rate, parts ...string) {
	if l != nil {
		l.Cache.Del(ctx, rt.key(parts...))
	}
}

// LimitIP wraps a handler so each address may call it at most rt.Max times per window.
func (a *Auth) LimitIP(rt Rate, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if a.Limiter.blocked(r.Context(), rt, ip) {
			httpx.Error(w, http.StatusTooManyRequests, MsgTooManyAttempts)
			return
		}
		a.Limiter.record(r.Context(), rt, ip)
		next(w, r)
	}
}
