package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/hex"
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

// newRecoveryCodes returns RecoveryCodeCount fresh codes and the hashes to store.
func newRecoveryCodes() (codes, hashes []string, err error) {
	for range RecoveryCodeCount {
		raw := make([]byte, 7) // 56 random bits; 50 of them reach the 10 characters
		if _, err := rand.Read(raw); err != nil {
			return nil, nil, err
		}
		code := base32.StdEncoding.EncodeToString(raw)[:RecoveryCodeLen]
		codes = append(codes, code)
		hashes = append(hashes, recoveryHash(code))
	}
	return codes, hashes, nil
}

// recoveryHash is how a recovery code is stored and looked up. Case, spaces
// and dashes are ignored, so a code copied from a printout still works.
func recoveryHash(code string) string {
	clean := strings.NewReplacer(" ", "", "-", "").Replace(strings.ToUpper(strings.TrimSpace(code)))
	sum := sha256.Sum256([]byte(clean))
	return hex.EncodeToString(sum[:])
}

// issueRecoveryCodes replaces the user's codes and returns the new ones, shown once.
func (a *Auth) issueRecoveryCodes(ctx context.Context, userID string) ([]string, error) {
	codes, hashes, err := newRecoveryCodes()
	if err != nil {
		return nil, err
	}
	return codes, a.Store.SetRecoveryCodes(ctx, userID, hashes)
}

// RegenerateRecoveryCodes replaces the user's recovery codes; it needs a
// current authenticator code. The old codes stop working.
func (a *Auth) RegenerateRecoveryCodes(w http.ResponseWriter, r *http.Request) {
	var in codeInput
	if !httpx.DecodeJSON(w, r, &in) {
		return
	}
	u, ok := a.freshUser(w, r)
	if !ok {
		return
	}
	if !u.TOTPEnabled {
		httpx.Error(w, http.StatusConflict, MsgTwoFactorNotOn)
		return
	}
	if !a.signedInCode(w, r, u, in.Code) {
		return
	}
	codes, err := a.issueRecoveryCodes(r.Context(), u.ID)
	if err != nil {
		httpx.ServerError(w, "issue recovery codes", err)
		return
	}
	httpx.OK(w, http.StatusOK, MsgRecoveryCodes, map[string]any{"recovery_codes": codes})
}

// signedInCode checks an app code sent by a signed-in user, or writes the
// refusal. Wrong codes are counted per user, so a stolen session cannot guess
// its way to turning 2FA off.
func (a *Auth) signedInCode(w http.ResponseWriter, r *http.Request, u User, code string) bool {
	ctx := r.Context()
	if a.Limiter.blocked(ctx, CodeRate, u.ID) {
		httpx.Error(w, http.StatusTooManyRequests, MsgTooManyAttempts)
		return false
	}
	valid, err := a.checkCode(ctx, u, code, a.now())
	if err != nil {
		httpx.ServerError(w, "check 2fa code", err)
		return false
	}
	if !valid {
		a.Limiter.record(ctx, CodeRate, u.ID)
		httpx.Error(w, http.StatusBadRequest, MsgBadCode)
		return false
	}
	a.Limiter.reset(ctx, CodeRate, u.ID)
	return true
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
	httpx.OK(w, http.StatusOK, "", map[string]string{"secret": key.Secret(), "otpauth_url": key.URL()})
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
	if !a.signedInCode(w, r, u, in.Code) {
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
	out := map[string]any{"two_factor_enabled": enable}
	if enable {
		// Shown once: only their hashes are kept.
		codes, err := a.issueRecoveryCodes(r.Context(), u.ID)
		if err != nil {
			httpx.ServerError(w, "issue recovery codes", err)
			return
		}
		out["recovery_codes"] = codes
	} else if err := a.Store.SetRecoveryCodes(r.Context(), u.ID, nil); err != nil {
		httpx.ServerError(w, "remove recovery codes", err)
		return
	}
	msg := MsgTwoFactorOff
	if enable {
		msg = MsgTwoFactorOn
	}
	httpx.OK(w, http.StatusOK, msg, out)
}
