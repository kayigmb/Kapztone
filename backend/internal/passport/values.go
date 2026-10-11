// Package passport handles digital product passports: issuing them at an
// origin facility and handing them from one organisation to the next.
package passport

import (
	"time"

	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/template"
)

// Every value in this file is declared here once and referenced everywhere
// else, including SQL arguments and tests.

// Passport statuses. They mirror PassportRegistry.Status on-chain.
const (
	StatusActive   = "active"
	StatusConsumed = "consumed"
	StatusRetired  = "retired"
	StatusRevoked  = "revoked"
)

var Statuses = []string{StatusActive, StatusConsumed, StatusRetired, StatusRevoked}

// Handover statuses.
const (
	// HandoverAwaiting: the sender has signed; the receiver has not answered.
	HandoverAwaiting = "awaiting_receiver"
	// HandoverCompleted: both have signed; the receiver now holds the passport.
	HandoverCompleted = "completed"
	HandoverRejected  = "rejected"
	HandoverCancelled = "cancelled"
	// HandoverExpired is never stored: an awaiting handover past its deadline reads as expired.
	HandoverExpired = "expired"
)

var HandoverStatuses = []string{HandoverAwaiting, HandoverCompleted, HandoverRejected, HandoverCancelled, HandoverExpired}

// Journey event kinds: the template's event names, plus the two things that
// can happen to a passport at any stage.
const (
	EventIssue     = template.EventIssue
	EventHandover  = template.EventHandover
	EventTransform = template.EventTransform
	EventFlag      = "FLAG"
	EventRevoke    = "REVOKE"
)

// Genealogy directions.
const (
	DirectionAncestors   = "ancestors"
	DirectionDescendants = "descendants"
)

var Directions = []string{DirectionAncestors, DirectionDescendants}

const (
	// HandoverTTL is how long a receiver has to answer before the sender's signature lapses.
	HandoverTTL = 7 * 24 * time.Hour

	MaxNoteLen = 1000

	// MaxTransformItems caps the inputs, and the outputs, of one transform:
	// the whole transform must fit in one chain transaction.
	MaxTransformItems = 50

	// A season is half a year: season 1 is January to June, season 2 July to December.
	firstSeasonLastMonth = time.June
)

// URL parameter names.
const (
	ParamStatus    = "status"
	ParamDirection = "direction"
)

// Messages shown to callers.
const (
	MsgFacilityNotFound  = "facility not found in your organisation"
	MsgFacilitySuspended = "this facility is suspended and cannot originate goods"
	MsgTemplateNotFound  = "commodity template not found"
	MsgKindMismatch      = "this commodity cannot originate at this kind of facility"
	MsgUnitMismatch      = "the facility's capacity unit differs from the commodity's unit"
	MsgNotHolder         = "your organisation does not hold this passport"
	MsgNotActive         = "this passport is no longer active"
	MsgOpenHandover      = "this passport already has a handover awaiting the receiver"
	MsgReceiverInvalid   = "the receiving organisation must be a different, verified organisation"
	MsgHandoverClosed    = "this handover is no longer awaiting an answer"
	MsgHandoverStale     = "the passport changed since this handover was created; start a new one"
	MsgReasonNeeded      = "give a reason for rejecting"
	MsgInputNotHeld      = "your organisation does not hold every input passport"
	MsgMixedInputs       = "all input passports must be of the same commodity and unit"
	MsgStageInvalid      = "this commodity has no such stage that allows " + EventTransform + " with a maximum yield"

	MsgNotifyHandoverStarted  = "A passport is being handed over to your organisation and awaits your answer."
	MsgNotifyHandoverAccepted = "The receiving organisation accepted your handover."
	MsgNotifyHandoverRejected = "The receiving organisation rejected your handover."
	MsgNotifyPassportRevoked  = "A passport your organisation holds was revoked."
)

// PeriodFor returns the capacity period a date falls in, as one number:
// 2026 for a year, 202610 for a month, 20261 or 20262 for a season.
func PeriodFor(capacityPeriod string, t time.Time) int64 {
	t = t.UTC()
	year := int64(t.Year())
	switch capacityPeriod {
	case facility.PeriodMonth:
		return year*100 + int64(t.Month())
	case facility.PeriodSeason:
		if t.Month() <= firstSeasonLastMonth {
			return year*10 + 1
		}
		return year*10 + 2
	}
	return year
}
