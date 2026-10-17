// Package passport handles digital product passports: issuing them at an
// origin facility and handing them from one organisation to the next.
package passport

import (
	"time"

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
	// EventRetire closes the journey: the lot reached its final buyer.
	EventRetire = "RETIRE"
)

// CodeFormat writes a lot code: organisation address in capitals, year, and
// the organisation's count for that year, at least four digits (ACME-2026-0412).
const CodeFormat = "%s-%d-%04d"

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
	MsgHandoverStage     = "stage must be a stage of this commodity that allows " + EventHandover

	MsgNotifyHandoverStarted  = "A passport is being handed over to your organisation and awaits your answer."
	MsgNotifyHandoverAccepted = "The receiving organisation accepted your handover."
	MsgNotifyHandoverRejected = "The receiving organisation rejected your handover."
	MsgNotifyPassportRevoked  = "A passport your organisation holds was revoked."
)
