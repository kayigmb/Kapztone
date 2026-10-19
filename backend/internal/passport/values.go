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

// Why a receiver refuses a handover; at least one is given.
const (
	RejectWeightShort  = "weight_short"
	RejectDamaged      = "damaged"
	RejectWrongProduct = "wrong_product"
	RejectSealBroken   = "seal_broken"
	RejectNotExpected  = "not_expected"
)

var RejectReasons = []string{RejectWeightShort, RejectDamaged, RejectWrongProduct, RejectSealBroken, RejectNotExpected}

// The rules raise a flag about an action they allow but find close to a limit.
const (
	// RuleYieldShare: a transform's outputs at this share of the maximum yield or more.
	RuleYieldShare = 0.98
	// RuleFlagShare: an accepted handover whose weight differs by more than this share of the tolerance.
	RuleFlagShare = 0.5
)

// Who raised a flag (flag.Sources); the passport package raises through RaiseFlag.
const (
	SourceUser  = "user"
	SourceRules = "rules"
)

// Who took a handover photo.
const (
	PhotoBySender   = "sender"
	PhotoByReceiver = "receiver"
)

const (
	// MaxPhotos is how many photos one side of a handover may attach.
	MaxPhotos = 6
	// FormPhotos is the multipart field holding them (repeated).
	FormPhotos = "photos"
)

// PhotoTypes are the file types a handover photo may be, decided from its content.
var PhotoTypes = []string{"image/png", "image/jpeg"}

// Fields of a handover sent as a multipart form (with photos); the same names as in JSON.
const (
	FieldToOrg            = "to_org"
	FieldStage            = "stage"
	FieldQuantity         = "quantity"
	FieldNote             = "note"
	FieldLat              = "lat"
	FieldLon              = "lon"
	FieldReceivedQuantity = "received_quantity"
)

// URL parameter names.
const (
	ParamStatus    = "status"
	ParamBox       = "box"
	ParamDirection = "direction"
	ParamPhotoID   = "photoId"
)

// Messages shown to callers.
const (
	MsgFacilityNotFound  = "facility not found in your organisation"
	MsgFacilitySuspended = "this facility is suspended and cannot originate goods"
	MsgFacilityPending   = "this facility is waiting for approval and cannot originate goods yet"
	MsgLicenceExpired    = "this facility's licence has expired; upload the renewed licence before issuing"
	MsgTemplateNotFound  = "commodity template not found"
	MsgKindMismatch      = "this commodity cannot originate at this kind of facility"
	MsgUnitMismatch      = "the facility's capacity unit differs from the commodity's unit"
	MsgNotHolder         = "your organisation does not hold this passport"
	MsgNotActive         = "this passport is no longer active"
	MsgOpenHandover      = "this passport already has a handover awaiting the receiver"
	MsgReceiverInvalid   = "the receiving organisation must be a different, verified organisation"
	MsgHandoverClosed    = "this handover is no longer awaiting an answer"
	MsgHandoverStale     = "the passport changed since this handover was created; start a new one"
	MsgInputNotHeld      = "your organisation does not hold every input passport"
	MsgMixedInputs       = "all input passports must be of the same commodity and unit"
	MsgStageInvalid      = "this commodity has no such stage that allows " + EventTransform + " with a maximum yield"
	MsgHandoverStage     = "stage must be a stage of this commodity that allows " + EventHandover
	MsgReasonsNeeded     = "choose at least one reason for rejecting"
	MsgLocation          = "lat and lon must be given together: lat from -90 to 90, lon from -180 to 180"
	MsgPhotoType         = "photos must be PNG or JPEG"
	MsgBadForm           = "send JSON, or a multipart form with the photos"
	MsgNumbers           = "quantity, received_quantity, lat and lon must be numbers"
	MsgReceivedNeeded    = "received_quantity is required: the weight you recorded on arrival"
	// MsgFlagFromReject starts the reason of a flag raised with a rejection.
	MsgFlagFromReject = "Handover rejected"
	// Reasons of flags the rules raise.
	MsgRuleYieldFmt     = "Rules: the outputs are %.1f%% of the most stage %s may yield"
	MsgRuleToleranceFmt = "Rules: the weight received differs by %.2f%% from what was sent; the stage accepts %.2f%%"

	MsgNotifyHandoverStarted  = "A passport is being handed over to your organisation and awaits your answer."
	MsgNotifyHandoverAccepted = "The receiving organisation accepted your handover."
	MsgNotifyHandoverRejected = "The receiving organisation rejected your handover."
	MsgNotifyPassportRevoked  = "A passport your organisation holds was revoked."
)

// Handover boxes (?box=), as the acting organisation sees them: waiting to be
// signed by it, waiting on the other side, or finished one way or another.
const (
	BoxIncoming = "incoming"
	BoxOutgoing = "outgoing"
	BoxDone     = "done"
)

var Boxes = []string{BoxIncoming, BoxOutgoing, BoxDone}

// Confirmations the app shows as they are.
const (
	MsgIssued            = "Passport issued. Print the QR label for the lot."
	MsgRetired           = "Passport retired. Its history stays verifiable."
	MsgHandoverStarted   = "Handover sent. The lot stays yours until the receiver signs."
	MsgHandoverAccepted  = "Handover complete. The lot is yours now."
	MsgHandoverRejected  = "Handover rejected. The lot stays with the sender."
	MsgHandoverCancelled = "Handover cancelled. The lot stays with you."
	MsgTransformed       = "Transform recorded. The new passports are ready."
)
