package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// A trusted device skips the 2FA code at sign-in for TrustTTL. The cookie
// names the user, the token version and the expiry, signed with the token
// secret; any token-version bump (password change, sign out everywhere,
// lock) voids it. Host-only like the session cookie: trusted per address.

func (a *Auth) trustMAC(payload string) []byte {
	m := hmac.New(sha256.New, a.Secret)
	m.Write([]byte("trust:" + payload))
	return m.Sum(nil)
}

func (a *Auth) trustCookie(u User, now time.Time) *http.Cookie {
	payload := u.ID + "|" + strconv.Itoa(u.TokenVersion) + "|" + strconv.FormatInt(now.Add(TrustTTL).Unix(), 10)
	value := base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." +
		base64.RawURLEncoding.EncodeToString(a.trustMAC(payload))
	return &http.Cookie{
		Name: TrustCookie, Value: value, Path: "/", MaxAge: int(TrustTTL.Seconds()),
		HttpOnly: true, Secure: !a.Dev, SameSite: http.SameSiteLaxMode,
	}
}

// trusted reports whether the request carries a valid trusted-device cookie for u.
func (a *Auth) trusted(r *http.Request, u User, now time.Time) bool {
	c, err := r.Cookie(TrustCookie)
	if err != nil {
		return false
	}
	p64, m64, ok := strings.Cut(c.Value, ".")
	payload, err1 := base64.RawURLEncoding.DecodeString(p64)
	mac, err2 := base64.RawURLEncoding.DecodeString(m64)
	if !ok || err1 != nil || err2 != nil || !hmac.Equal(mac, a.trustMAC(string(payload))) {
		return false
	}
	parts := strings.Split(string(payload), "|")
	if len(parts) != 3 {
		return false
	}
	ver, err1 := strconv.Atoi(parts[1])
	exp, err2 := strconv.ParseInt(parts[2], 10, 64)
	return err1 == nil && err2 == nil && parts[0] == u.ID && ver == u.TokenVersion && now.Unix() < exp
}
