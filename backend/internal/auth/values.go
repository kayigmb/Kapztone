package auth

import (
	"regexp"
	"slices"
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
)

const (
	// ParamOrgSlug names the acting organisation in dev, in place of a subdomain.
	ParamOrgSlug = "pslug"

	MinPasswordLen = 12
	MaxPasswordLen = 72 // bcrypt ignores anything past 72 bytes

	bearerPrefix = "Bearer "
	// HeaderOrigin is the header a browser sets to the address of the calling site.
	HeaderOrigin = "Origin"
	// HeaderForwardedHost carries the site's own host when the web app passes
	// a browser's request through to the API.
	HeaderForwardedHost  = "X-Forwarded-Host"
	HeaderForwardedProto = "X-Forwarded-Proto"
	// SessionCookie holds the sign-in token in a browser. The API sets and
	// clears it; scripts on the page cannot read it.
	SessionCookie  = "ct_session"
	superadminName = "Platform admin"
)

// Messages shown to callers.
const (
	MsgSignInRequired      = "sign in required"
	MsgBadCredentials      = "email or password is incorrect"
	MsgNotAllowed          = "you are not allowed to do this"
	MsgForeignOrigin       = "this request did not come from the site you are signed in to"
	MsgNoOrgSelected       = "choose an organisation to act for"
	MsgOrgNotVerified      = "your organisation must be verified before it can do this"
	MsgTooManyAttempts     = "too many attempts; try again later"
	MsgTwoFactorCode       = "enter the code from your authenticator app"
	MsgBadCode             = "code is incorrect or already used"
	MsgTwoFactorAlreadyOn  = "two-factor sign-in is already on"
	MsgTwoFactorNotOn      = "two-factor sign-in is not on"
	MsgTwoFactorSetupFirst = "set up two-factor sign-in first"
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
