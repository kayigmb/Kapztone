package auth

import (
	"fmt"
	"regexp"
	"slices"
	"time"
)

// Every value in this file is declared here once and referenced everywhere
// else, including SQL arguments and tests.

// Default role names.
const (
	RoleSuperadmin = "superadmin"
	RoleOrgAdmin   = "org_admin"
	RoleMember     = "member"
)

// User account statuses.
const (
	UserActive   = "active"
	UserDisabled = "disabled"
)

// KYB status of an organisation.
const (
	OrgPending  = "pending"
	OrgVerified = "verified"
	OrgRejected = "rejected"
	// OrgChangesRequested: the reviewer wants corrected documents. Like a
	// rejection, a new upload sends the organisation back to pending.
	OrgChangesRequested = "changes_requested"
)

const (
	// ParamOrgSlug names the acting organisation in dev, in place of a subdomain.
	ParamOrgSlug = "pslug"

	MinPasswordLen = 8
	MaxPasswordLen = 72 // bcrypt ignores anything past 72 bytes

	bearerPrefix = "Bearer "
	// HeaderOrigin is the header a browser sets to the address of the calling site.
	HeaderOrigin = "Origin"
	// HeaderForwardedHost carries the site's own host when the web app passes
	// a browser's request through to the API.
	HeaderForwardedHost  = "X-Forwarded-Host"
	HeaderForwardedProto = "X-Forwarded-Proto"
	// HeaderForwardedFor carries the visitor's address from the web app.
	HeaderForwardedFor = "X-Forwarded-For"
	// SessionCookie holds the sign-in token in a browser. The API sets and
	// clears it; scripts on the page cannot read it.
	SessionCookie = "ct_session"
	// HeaderSigningPIN carries the user's signing PIN on every signing action.
	HeaderSigningPIN = "X-Signing-PIN"
	PINLen           = 6

	// 2FA recovery codes: shown once, each works once in place of an app code.
	RecoveryCodeCount = 10
	RecoveryCodeLen   = 10
	superadminName    = "Platform admin"

	// TrustCookie lets a device skip the 2FA code: HMAC-signed user, token
	// version and expiry. A token-version bump (password change, sign out
	// everywhere, lock) voids it.
	TrustCookie = "ct_trust"
	// PasswordHistoryLen: a new password may not repeat one of the last this many.
	PasswordHistoryLen = 5

	// Emailed one-use links (password_resets.kind).
	LinkReset = "reset"
	LinkLock  = "lock"
	// LinkHandoff carries a session to another address (handoff.go).
	LinkHandoff = "handoff"
	// Pages of the frontend those links open: <origin><path><token>.
	PathReset = "/reset/"
	PathLock  = "/lock/"
)

// How long things last.
const (
	TrustTTL     = 30 * 24 * time.Hour
	ResetLinkTTL = 30 * time.Minute
	LockLinkTTL  = 7 * 24 * time.Hour
	HandoffTTL   = time.Minute
)

// Messages shown to callers.
const (
	MsgSignInRequired      = "sign in required"
	MsgBadCredentials      = "Email or password is incorrect. Check both and try again."
	MsgNotAllowed          = "you are not allowed to do this"
	MsgForeignOrigin       = "this request did not come from the site you are signed in to"
	MsgNoOrgSelected       = "choose an organisation to act for"
	MsgOrgNotVerified      = "your organisation must be verified before it can do this"
	MsgTooManyAttempts     = "Too many attempts. Wait a few minutes, then try again."
	MsgTwoFactorCode       = "Enter the code from your authenticator app."
	MsgBadCode             = "That code is incorrect or already used. Enter the code your app shows now."
	MsgTwoFactorAlreadyOn  = "two-factor sign-in is already on"
	MsgTwoFactorNotOn      = "two-factor sign-in is not on"
	MsgTwoFactorSetupFirst = "set up two-factor sign-in first"
	MsgPasswordLength      = "new password must be 8 to 72 characters"
	MsgPasswordComplexity  = "new password must include an uppercase letter, a lowercase letter, a number and a symbol"
	MsgPINNotSet           = "set a signing PIN first"
	MsgPINNeeded           = "enter your signing PIN"
	MsgBadPIN              = "signing PIN is incorrect"
	MsgPINFormat           = "signing PIN must be exactly 6 digits"
	MsgPINLocked           = "too many wrong PINs; signing is locked for 15 minutes"
	MsgLinkInvalid         = "this link has expired or was already used"
	MsgNotMember           = "you are not a member of that organisation"
	MsgHandoffInvalid      = "this switch has expired or was already used; sign in here instead"
	// Confirmations the app shows as they are.
	MsgSignedIn        = "Signed in."
	MsgSignedOut       = "Signed out."
	MsgPasswordChanged = "Password changed. Other devices are signed out; we've e-mailed you a link to lock the account if this wasn't you."
	MsgPasswordReset   = "Password changed. Sign in with the new one."
	MsgResetSent       = "If that email has an account, a reset link is on its way."
	MsgAccountLocked   = "Account locked. Everyone signed in is signed out."
	MsgSessionsRevoked = "Signed out of every other device."
	MsgPINSet          = "Signing PIN saved."
	MsgRecoveryCodes   = "New recovery codes made. The old ones no longer work."
	MsgTwoFactorOn     = "Two-factor sign-in is on. Keep your recovery codes somewhere safe."
	MsgTwoFactorOff    = "Two-factor sign-in is off."
	MsgPasswordReused  = "choose a password you haven't used for your last 5 passwords"
)

// MsgAttemptsLeft follows a failed sign-in while rate limits are on.
func MsgAttemptsLeft(n int64) string {
	if n == 1 {
		return "1 attempt left before sign-in is paused for 15 minutes."
	}
	return fmt.Sprintf("%d attempts left before sign-in is paused for 15 minutes.", n)
}

// Emails the account flows send.
const (
	MailResetSubject = "Reset your ChainTrack password"
	MailResetBody    = "Someone asked to reset the password of your ChainTrack account.\r\n\r\n" +
		"Choose a new password here (the link works once, for 30 minutes):\r\n%s\r\n\r\n" +
		"If this wasn't you, ignore this email. Your password has not changed."
	MailChangedSubject = "Your ChainTrack password was changed"
	MailChangedBody    = "The password of your ChainTrack account was just changed, and every other device was signed out.\r\n\r\n" +
		"If this wasn't you, lock your account now (the link works once, for 7 days):\r\n%s"
)

// Cache keys.
const (
	cacheUserPrefix  = "auth:user:"
	cacheScopePrefix = "auth:scope:"
	cacheRatePrefix  = "auth:rate:"
)

// An organisation's slug is its subdomain, so it must be a valid hostname
// label and must not collide with a system hostname.
var (
	slugPattern   = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)
	ReservedSlugs = []string{"www", "api", "admin", "app"}
)

func ValidSlug(s string) bool {
	return slugPattern.MatchString(s) && !slices.Contains(ReservedSlugs, s)
}
