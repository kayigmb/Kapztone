// Package verify answers the public question a QR code asks: can this
// passport be trusted, back to where the goods came from?
package verify

// Every value in this file is declared here once and referenced everywhere
// else, including tests.

// Verdicts, which are also the levels of a reason. The verdict is the worst
// level among the reasons.
const (
	Verified = "verified"
	Warning  = "warning"
	Failed   = "failed"
)

// The checks, in the order they are reported.
const (
	CheckRecord       = "record_intact"
	CheckAnchored     = "anchored"
	CheckOrigin       = "origin"
	CheckCapacity     = "capacity"
	CheckMassBalance  = "mass_balance"
	CheckCertificates = "certificates"
	CheckStatus       = "status"
)

var Checks = []string{CheckRecord, CheckAnchored, CheckOrigin, CheckCapacity, CheckMassBalance, CheckCertificates, CheckStatus}

// Reason codes and the message shown for each.
const (
	RecordAltered       = "record_altered"
	NotAnchored         = "not_anchored"
	ChainUnreachable    = "chain_unreachable"
	ChainMismatch       = "chain_mismatch"
	ChainMissing        = "chain_missing"
	CustodyPending      = "custody_pending"
	OriginMissing       = "origin_missing"
	FacilityAltered     = "facility_altered"
	FacilitySuspended   = "facility_suspended"
	OwnerNotVerified    = "owner_not_verified"
	CapacityExceeded    = "capacity_exceeded"
	MassBalanceBroken   = "mass_balance_broken"
	CertificateAltered  = "certificate_altered"
	CertificateExpired  = "certificate_expired"
	CertificateRevoked  = "certificate_revoked"
	CertificateFileGone = "certificate_file_unavailable"
	IssuerNotAccredited = "issuer_not_accredited"
	PassportRevoked     = "passport_revoked"
	FlagOpen            = "flag_open"
)

// CacheKeyPrefix starts the cache key of every stored answer.
const CacheKeyPrefix = "verify:"

// The messages are shown on the public page as they are, so they follow its
// rule: say what was checked, never how. ForbiddenWords may not appear in them.
var ForbiddenWords = []string{"blockchain", "crypto", "wallet", "token", "on-chain", "smart contract", "hash"}

var messages = map[string]string{
	RecordAltered:       "This passport's record was changed after it was written.",
	NotAnchored:         "This record has not been sealed yet. Check again in a few minutes.",
	ChainUnreachable:    "The ledger could not be reached, so records could not be compared with their sealed copies. Check again later.",
	ChainMismatch:       "This record differs from its sealed copy.",
	ChainMissing:        "This record is marked as sealed, but no sealed copy exists.",
	CustodyPending:      "The latest change of holder has not been sealed yet. Check again in a few minutes.",
	OriginMissing:       "This lot cannot be traced back to an origin facility.",
	FacilityAltered:     "The origin facility's record differs from its sealed copy.",
	FacilitySuspended:   "The origin facility is suspended.",
	OwnerNotVerified:    "The organisation that owns the origin facility is not verified.",
	CapacityExceeded:    "More was issued from the origin facility in the period than its capacity allows.",
	MassBalanceBroken:   "More came out of a processing step than its inputs could yield.",
	CertificateAltered:  "A certificate does not match the copy sealed when it was issued.",
	CertificateExpired:  "A certificate has expired.",
	CertificateRevoked:  "A certificate was revoked by its issuer.",
	CertificateFileGone: "A certificate's file could not be read to check it.",
	IssuerNotAccredited: "A certificate's issuer is not accredited for that type of certificate.",
	PassportRevoked:     "This passport, or one it was made from, has been revoked.",
	FlagOpen:            "A concern about this passport, or one it was made from, is being looked into.",
}
