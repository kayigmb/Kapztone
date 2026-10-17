// Package audit keeps the audit log: one sealed row for every successful
// change made through the API, and the reports read from the platform's data.
package audit

// Every value in this file is declared here once and referenced everywhere
// else, including SQL arguments and tests.

// URL parameter names.
const (
	ParamActor  = "actor"  // user id
	ParamRoute  = "route"  // route pattern, e.g. /passports/{id}/handovers
	ParamFormat = "format" // empty for JSON, FormatCSV for a download
)

const FormatCSV = "csv"

// File names of CSV downloads.
const (
	FileAudit  = "audit.csv"
	FileVolume = "volume.csv"
)

// Messages shown to callers.
const (
	MsgBadFormat = ParamFormat + " must be " + FormatCSV + " or left out"
	MsgBadActor  = ParamActor + " must be a user id"
)

// chainLock is the advisory lock that keeps rows in chain order. Arbitrary;
// distinct from the superadmin seed's lock.
const chainLock = 7203115

// CSV column headings.
var (
	AuditColumns  = []string{"seq", "created_at", "actor_user_id", "actor_email", "organisation_id", "method", "route", "subject_id", "status", "hash"}
	VolumeColumns = []string{"month", "commodity", "unit", "quantity"}
)
