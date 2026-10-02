// Package auth answers three questions for every request: who is calling,
// for which organisation, and what they are allowed to do.
package auth

import (
	"context"
	"errors"
	"net"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrNotFound = errors.New("not found")

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	PasswordHash string `json:"-"`
}

type Organisation struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Kind   string `json:"kind"`
	Status string `json:"status"`
}

// Scope is what a user may do in the organisation they are acting for.
// Organisation is nil for a platform-wide scope or when no organisation was named.
type Scope struct {
	Organisation *Organisation `json:"organisation"`
	Role         string        `json:"role"`
	Permissions  []string      `json:"permissions"`
}

// Store is the database access auth needs. Every method returns ErrNotFound
// when there is no matching row.
type Store interface {
	// UserByEmail returns an active user including the password hash.
	UserByEmail(ctx context.Context, email string) (User, error)
	// ActiveUser returns the user only if the account is still active.
	ActiveUser(ctx context.Context, id string) (User, error)
	// Scope returns the user's scope for the organisation with this slug,
	// falling back to their platform-wide scope if they have one.
	Scope(ctx context.Context, userID, slug string) (Scope, error)
}

type Auth struct {
	Store      Store
	Secret     []byte        // signs sign-in tokens
	TokenTTL   time.Duration // how long a sign-in lasts
	BaseDomain string        // organisations live at <slug>.<BaseDomain>
	Dev        bool          // accept ?pslug= in place of a subdomain
}

func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(h), err
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// IssueToken signs a token that says only who the user is. Roles and
// permissions are deliberately left out: they are read from the database on
// every request, so removing them takes effect immediately.
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
		jwt.WithValidMethods([]string{"HS256"}),
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

var (
	slugPattern   = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)
	reservedSlugs = []string{"www", "api", "admin", "app"}
)

// Slug returns the organisation slug the request names, or "" if none.
// It is only a claim: ResolveOrg checks the user really belongs there.
func (a *Auth) Slug(r *http.Request) string {
	slug := ""
	if a.Dev {
		slug = r.URL.Query().Get("pslug")
	}
	if slug == "" {
		host := r.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		sub, ok := strings.CutSuffix(strings.ToLower(host), "."+a.BaseDomain)
		if !ok {
			return ""
		}
		slug = sub
	}
	if !slugPattern.MatchString(slug) || slices.Contains(reservedSlugs, slug) {
		return ""
	}
	return slug
}
