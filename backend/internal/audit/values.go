// Package audit keeps the audit log: one sealed row for every successful
// change made through the API, and the reports read from the platform's data.
package audit

// Every value in this file is declared here once and referenced everywhere
// else, including SQL arguments and tests.

// URL parameter names.
const (
	ParamActor  = "actor"  // user id
	ParamRoute  = "route"  // route pattern, e.g. /passports/{id}/handovers
	ParamFormat = "format" // empty for JSON; FormatCSV or FormatPDF for a download
)

// Download formats (?format=); empty means JSON.
const (
	FormatCSV = "csv"
	FormatPDF = "pdf"
)

var Formats = []string{FormatCSV, FormatPDF}

// PDF titles.
const (
	TitleAudit     = "Audit log"
	TitleVolume    = "Volume issued at origin"
	TitleBlocked   = "Actions the rules refused"
	TitleOutcomes  = "Verification outcomes"
	TitleSaved     = "Saved report"
)

// File names of downloads, without extension.
const (
	FileAudit    = "audit"
	FileVolume   = "volume"
	FileBlocked  = "blocked"
	FileOutcomes = "verification-outcomes"
	FileSaved    = "saved-report"
)

// Saved report kinds.
const (
	SavedVolume   = "volume"
	SavedBlocked  = "blocked"
	SavedOutcomes = "verification-outcomes"
	SavedAudit    = "audit"
)

var SavedKinds = []string{SavedVolume, SavedBlocked, SavedOutcomes, SavedAudit}

// Messages for saved reports.
const (
	MsgBadSavedKind = "kind must be volume, blocked, verification-outcomes or audit"
	MsgBadSavedName = "name must be 1 to 120 characters"
)

// Rules whose refusals are counted (rule_events.rule), declared once.
const (
	RuleCapacity         = "capacity"     // more issued than a facility may originate
	RuleMassBalance      = "mass_balance" // a transform's output above input × max yield
	RuleEvidence         = "evidence"     // a stage's required certificate missing
	RuleTolerance        = "tolerance"    // received weight outside the handover tolerance
	RuleHandoverRejected = "handover_rejected"
)

var Rules = []string{RuleCapacity, RuleMassBalance, RuleEvidence, RuleTolerance, RuleHandoverRejected}

// handoverRejected is the passport package's rejected handover status. Not
// imported: passport records rule events here, so audit cannot import it.
const (
	handoverRejected  = "rejected"
	handoverCompleted = "completed"
)

// Messages shown to callers.
const (
	MsgBadFormat = ParamFormat + " must be " + FormatCSV + " or " + FormatPDF + ", or left out"
	MsgBadActor  = ParamActor + " must be a user id"
)

// chainLock is the advisory lock that keeps rows in chain order. Arbitrary;
// distinct from the superadmin seed's lock.
const chainLock = 7203115

// CSV column headings.
var (
	AuditColumns    = []string{"seq", "created_at", "actor_user_id", "actor_email", "organisation_id", "method", "route", "subject_id", "status", "hash"}
	VolumeColumns   = []string{"month", "commodity", "unit", "quantity", "exported"}
	BlockedColumns  = []string{"rule", "count"}
	OutcomesColumns = []string{"period", "verified", "warning", "failed", "total", "pct_verified"}
)
