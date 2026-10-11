package auth

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/kayigmb/chaintrack/backend/internal/httpx"
)

// Two-factor sign-in uses TOTP: 6-digit codes from an authenticator app.
const (
	totpIssuer = "ChainTrack"
	totpPeriod = 30 // seconds a code is valid for
	totpSkew   = 1  // periods of clock drift tolerated either side
)

var totpOpts = totp.ValidateOpts{Period: totpPeriod, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1}

// TOTPCode returns the code an authenticator app would show for a secret at time t.
func TOTPCode(secret string, t time.Time) (string, error) {
	return totp.GenerateCodeCustom(secret, t, totpOpts)
}

// checkCode reports whether code is the user's current code and has not been
// used before. An accepted code is burned, so it cannot be replayed.
func (a *Auth) checkCode(ctx context.Context, u User, code string, now time.Time) (bool, error) {
	code = strings.TrimSpace(code)
	if code == "" || u.TOTPSecret == "" {
		return false, nil
	}
	secret, err := a.Secrets.Open(u.TOTPSecret)
	if err != nil {
		return false, err
	}
	current := now.Unix() / totpPeriod
	for step := current - totpSkew; step <= current+totpSkew; step++ {
		want, err := TOTPCode(secret, time.Unix(step*totpPeriod, 0))
		if err != nil {
			return false, err
		}
		if subtle.ConstantTimeCompare([]byte(want), []byte(code)) == 1 {
			return a.Store.UseTOTPStep(ctx, u.ID, step)
		}
	}
	return false, nil
}

// freshUser reloads the signed-in user with their 2FA state.
func (a *Auth) freshUser(w http.ResponseWriter, r *http.Request) (User, bool) {
	u, err := a.Store.UserByEmail(r.Context(), CurrentUser(r.Context()).Email)
	if err != nil {
		httpx.ServerError(w, "load user", err)
		return User{}, false
	}
	return u, true
}

type codeInput struct {
	Code string `json:"code"`
}

// SetupTwoFactor creates a new secret for the user to add to their
// authenticator app. 2FA stays off until EnableTwoFactor proves the app works.
func (a *Auth) SetupTwoFactor(w http.ResponseWriter, r *http.Request) {
	u, ok := a.freshUser(w, r)
	if !ok {
		return
	}
	if u.TOTPEnabled {
		httpx.Error(w, http.StatusConflict, MsgTwoFactorAlreadyOn)
		return
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer: totpIssuer, AccountName: u.Email,
		Period: totpPeriod, Digits: totpOpts.Digits, Algorithm: totpOpts.Algorithm,
	})
	if err != nil {
		httpx.ServerError(w, "generate 2fa secret", err)
		return
	}
	sealed, err := a.Secrets.Seal(key.Secret())
	if err != nil {
		httpx.ServerError(w, "seal 2fa secret", err)
		return
	}
	if err := a.Store.SetTOTP(r.Context(), u.ID, sealed, false); err != nil {
		httpx.ServerError(w, "store 2fa secret", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"secret": key.Secret(), "otpauth_url": key.URL()})
}

// EnableTwoFactor turns 2FA on once the user sends a valid first code.
func (a *Auth) EnableTwoFactor(w http.ResponseWriter, r *http.Request) {
	a.switchTwoFactor(w, r, true)
}

// DisableTwoFactor turns 2FA off; it needs a current code.
func (a *Auth) DisableTwoFactor(w http.ResponseWriter, r *http.Request) {
	a.switchTwoFactor(w, r, false)
}

func (a *Auth) switchTwoFactor(w http.ResponseWriter, r *http.Request, enable bool) {
	var in codeInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	u, ok := a.freshUser(w, r)
	if !ok {
		return
	}
	switch {
	case enable && u.TOTPEnabled:
		httpx.Error(w, http.StatusConflict, MsgTwoFactorAlreadyOn)
		return
	case enable && u.TOTPSecret == "":
		httpx.Error(w, http.StatusConflict, MsgTwoFactorSetupFirst)
		return
	case !enable && !u.TOTPEnabled:
		httpx.Error(w, http.StatusConflict, MsgTwoFactorNotOn)
		return
	}
	valid, err := a.checkCode(r.Context(), u, in.Code, a.now())
	if err != nil {
		httpx.ServerError(w, "check 2fa code", err)
		return
	}
	if !valid {
		httpx.Error(w, http.StatusBadRequest, MsgBadCode)
		return
	}
	sealed := u.TOTPSecret
	if !enable {
		sealed = "" // forget the secret; turning 2FA on again starts from setup
	}
	if err := a.Store.SetTOTP(r.Context(), u.ID, sealed, enable); err != nil {
		httpx.ServerError(w, "switch 2fa", err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]bool{"two_factor_enabled": enable})
}
