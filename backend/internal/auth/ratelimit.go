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
)

// Limiter counts events in the cache. If the cache is unreachable it allows
// the request: a Redis outage must not lock everyone out.
type Limiter struct {
	Cache *cache.Cache
}

func (rt Rate) key(parts ...string) string {
	return cacheRatePrefix + rt.Name + ":" + strings.Join(parts, ":")
}

// clientIP is the address the request came from.
// ponytail: uses the direct peer address. Behind a reverse proxy every user
// shares the proxy's address; when nginx is added, take the address from its
// forwarded header (and only from that proxy).
func clientIP(r *http.Request) string {
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
