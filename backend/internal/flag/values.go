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

// URL parameter names.
const ParamStatus = "status"

// Messages shown to callers.
const (
	MsgReasonNeeded = "say what is wrong with this passport"
	MsgResolved     = "this flag has already been resolved"

	MsgNotifyRaised   = "A concern was raised about a passport your organisation holds."
	MsgNotifyResolved = "A concern your organisation raised has been resolved."
)
