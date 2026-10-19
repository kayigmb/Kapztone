// Package template handles commodity templates: the platform-wide rules for
// one commodity (unit, stages, allowed events, yields, capacity rule).
package template

import "regexp"

// Every value in this file is declared here once and referenced everywhere
// else, including SQL arguments and tests.

// Events a stage may allow.
const (
	EventIssue     = "ISSUE"
	EventHandover  = "HANDOVER"
	EventTransform = "TRANSFORM"
	EventExport    = "EXPORT"
	EventImport    = "IMPORT"
)

var Events = []string{EventIssue, EventHandover, EventTransform, EventExport, EventImport}

// How the amount a facility may originate per period is worked out.
const (
	// CapacityYieldPerArea: facility area × a rate per hectare per season.
	CapacityYieldPerArea = "yield_per_area"
	// CapacityFacility: the capacity recorded on the facility itself.
	CapacityFacility = "facility_capacity"
)

var CapacityTypes = []string{CapacityYieldPerArea, CapacityFacility}

const (
	MaxNameLen = 200
	MaxUnitLen = 16
)

// StageSplit is the stage at which part of a lot is cut into its own passport
// before being handed over. It is the platform's, at a yield of 1, so no
// template may use the key for a stage of its own.
const StageSplit = "split"

// DefaultHandoverTolerance is the weight difference, in percent, a receiver
// may accept when the stage sets no handoverTolerance of its own.
const DefaultHandoverTolerance = 1.0

// Keys (commodity, stage, evidence) are lowercase identifiers such as "washing_station".
var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,49}$`)

const keyRule = "lowercase letters, digits and underscores, starting with a letter"

// KeyRule and ValidKey let other packages hold their keys to the same rule:
// a certificate's type is the evidence key a stage asks for.
const KeyRule = keyRule

func ValidKey(s string) bool { return keyPattern.MatchString(s) }

// Messages shown to callers.
const (
	MsgCommodityTaken = "a template for this commodity already exists"
	MsgInUse          = "passports already use this template; its commodity and unit can no longer change"
	MsgNoDraft        = "this template has no draft to publish; save the changes first"
)

// Confirmations the app shows as they are.
const (
	MsgCreated    = "Template created and live."
	MsgDraftSaved = "Draft saved. Publish it to make it the live version."
	MsgPublished  = "Template published. New passports use this version."
	MsgDeleted    = "Template deleted."
)
