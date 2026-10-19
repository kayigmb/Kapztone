// Package flag handles concerns raised about a passport, and what follows
// when one is confirmed: the passport and everything made from it is revoked.
package flag

// Every value in this file is declared here once and referenced everywhere
// else, including SQL arguments and tests.

// Flag statuses.
const (
	StatusOpen      = "open"
	StatusCleared   = "cleared"   // looked into; nothing wrong
	StatusConfirmed = "confirmed" // upheld; the passport tree is revoked
)

var Statuses = []string{StatusOpen, StatusCleared, StatusConfirmed}

// Outcomes a resolver can choose are the two final statuses.
var Outcomes = []string{StatusCleared, StatusConfirmed}

const MaxTextLen = 2000

// Where a flag came from (flags.source, shown as raised_by).
const (
	SourceUser   = "user"   // a person raised it
	SourceRules  = "rules"  // the platform's rules raised it (see passport.RaiseFlag)
	SourcePublic = "public" // from the public verification page
)

var Sources = []string{SourceUser, SourceRules, SourcePublic}

// URL parameter names.
const ParamStatus = "status"

// StatusResolved is a list filter only (?status=resolved): cleared or confirmed.
const StatusResolved = "resolved"

// Messages shown to callers.
const (
	MsgReasonNeeded = "say what is wrong with this passport"
	MsgResolved     = "this flag has already been resolved"
	MsgAssignee     = "the assignee must be an active platform admin or a member of a verified regulator"

	MsgNotifyRaised         = "A concern was raised about a passport your organisation holds."
	MsgNotifyResolved       = "A concern your organisation raised has been resolved."
	MsgNotifyRaisedPlatform = "A concern was raised about a passport and is waiting for review."
)

// Confirmations the app shows as they are.
const (
	MsgRaised   = "Flag raised. An investigator will look into it."
	MsgAssigned = "Investigator updated."
)

// resolvedMessage confirms a resolution, by outcome.
var resolvedMessage = map[string]string{
	StatusCleared:   "Flag cleared. The lot is back to normal.",
	StatusConfirmed: "Fraud confirmed. The lot and every lot made from it are revoked.",
}
