// Package auth answers three questions for every request: who is calling,
// for which organisation, and what they are allowed to do.
package auth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/kayigmb/chaintrack/backend/internal/secret"
)

var ErrNotFound = errors.New("not found")

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	PasswordHash string `json:"-"`

	// Two-factor state; filled by UserByEmail only.
	TOTPSecret  string `json:"-"` // sealed with secret.Box
	TOTPEnabled bool   `json:"-"`
}

type Organisation struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Kind   string `json:"kind"`
	Status string `json:"status"`
}

// Scope is what a user may do in the organisation they are acting for.
type Scope struct {
	// Organisation is nil for a platform-wide scope or when none was named.
	Organisation *Organisation `json:"organisation"`
	// Platform is true for a scope that is not tied to one organisation.
	Platform    bool     `json:"platform"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

// Gate is the one rule for which rows a scope may see. Every query that
// returns organisation-owned data filters by it.
//
//	ok == false           no data access at all
//	orgID == "" (ok)      platform-wide: all organisations
//	orgID != ""           only rows belonging to that organisation
func (s Scope) Gate() (orgID string, ok bool) {
	switch {
	case s.Platform:
		return "", true
	case s.Organisation != nil:
		return s.Organisation.ID, true
	}
	return "", false
}

// GateSQL is the SQL form of Gate: a condition on an organisation id column,
// with the gate's orgID passed as argument number n. An empty argument means
// "all organisations".
func GateSQL(column string, n int) string {
	return fmt.Sprintf("($%d = '' OR %s = NULLIF($%d, '')::uuid)", n, column, n)
}

// Store is the database access auth needs. Every method returns ErrNotFound
// when there is no matching live row.
type Store interface {
	// UserByEmail returns an active user including the password hash.
	UserByEmail(ctx context.Context, email string) (User, error)
	// ActiveUser returns the user only if the account is still active.
	ActiveUser(ctx context.Context, id string) (User, error)
	// Scope returns the user's scope for the organisation with this slug,
	// falling back to their platform-wide scope if they have one.
	Scope(ctx context.Context, userID, slug string) (Scope, error)
	// SetTOTP stores the user's sealed 2FA secret and whether 2FA is on.
	SetTOTP(ctx context.Context, userID, sealedSecret string, enabled bool) error
	// UseTOTPStep burns a 2FA time-step; false means it was already used.
	UseTOTPStep(ctx context.Context, userID string, step int64) (bool, error)
	// PINHash returns the user's signing PIN hash, "" when none is set.
	PINHash(ctx context.Context, userID string) (string, error)
	SetPIN(ctx context.Context, userID, hash string) error
	// SetRecoveryCodes replaces the user's recovery codes (SHA-256 hex each); nil removes them.
	SetRecoveryCodes(ctx context.Context, userID string, hashes []string) error
	// UseRecoveryCode burns an unused code; false means no such unused code.
	UseRecoveryCode(ctx context.Context, userID, hash string) (bool, error)
}

type Auth struct {
	Store    Store
	Secret   []byte           // signs sign-in tokens
	TokenTTL time.Duration    // how long a sign-in lasts
	Dev      bool             // accept ?pslug= in place of the frontend subdomain
	Secrets  *secret.Box      // seals 2FA secrets
	Limiter  *Limiter         // nil disables rate limiting
	Now      func() time.Time // the clock; nil means time.Now (tests set it)
}

func (a *Auth) now() time.Time {
	if a.Now != nil {
		return a.Now()
	}
	return time.Now()
}

func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(h), err
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// IssueToken signs a token that says only who the user is. Roles and
// permissions are deliberately left out: they are looked up on every request,
// so removing them takes effect immediately.
func (a *Auth) IssueToken(userID string, now time.Time) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(a.TokenTTL)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(a.Secret)
}

// ParseToken returns the user id from a valid, unexpired token.
func (a *Auth) ParseToken(token string) (string, error) {
	var claims jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(token, &claims,
		func(*jwt.Token) (any, error) { return a.Secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return "", err
	}
	if claims.Subject == "" {
		return "", errors.New("token has no subject")
	}
	return claims.Subject, nil
}

// Frontend is where the user's browser is: the site that made the request.
type Frontend struct {
	Origin string // scheme://host[:port], usable as the base of links back to the site
	Host   string // lowercase host without port, e.g. "acme.chaintrack.app"
}

// siteHost is the host (with port) of the site the browser is on: the one
// the web app forwarded, or the request's own when called directly.
func siteHost(r *http.Request) string {
	if h := r.Header.Get(HeaderForwardedHost); h != "" {
		return h
	}
	return r.Host
}

// frontendOf works out the frontend from the request. A browser names the
// calling site in the Origin header; without one (a page load, curl) it is
// the host the web app forwarded, or the request's own. Either way it is
// only a claim.
func frontendOf(r *http.Request) Frontend {
	if u, err := url.Parse(r.Header.Get(HeaderOrigin)); err == nil && u.Host != "" {
		return Frontend{Origin: u.Scheme + "://" + u.Host, Host: strings.ToLower(u.Hostname())}
	}
	scheme := "http"
	if r.TLS != nil || r.Header.Get(HeaderForwardedProto) == "https" {
		scheme = "https"
	}
	host := siteHost(r)
	full := host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return Frontend{Origin: scheme + "://" + full, Host: strings.ToLower(host)}
}

// Slug returns the organisation slug the request names, or "" if none: the
// first label of the frontend's host ("acme" in acme.chaintrack.app). In dev,
// ?pslug= names it instead.
//
// No domain is configured, so this cannot tell an organisation's subdomain
// from the platform's own name; it does not need to. The slug is only a
// candidate: ResolveOrg asks the database whether the user belongs to an
// organisation with that slug, and anything else means "no organisation".
func (a *Auth) Slug(r *http.Request) string {
	slug := ""
	if a.Dev {
		slug = r.URL.Query().Get(ParamOrgSlug)
	}
	if slug == "" {
		host := frontendOf(r).Host
		label, _, hasMore := strings.Cut(host, ".")
		if !hasMore || net.ParseIP(host) != nil {
			return "" // "localhost" or an IP address has no subdomain
		}
		slug = label
	}
	if !ValidSlug(slug) {
		return ""
	}
	return slug
}
