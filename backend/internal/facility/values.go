// Package facility handles the places where goods originate or are handled:
// farm plots, mine sites, washing stations, plants and warehouses.
package facility

import (
	"regexp"
	"strings"
	"time"
)

// keyPattern is the shape of a facility kind. It mirrors
// template.ValidKey on purpose (that package imports this one, so it cannot
// be shared); kinds_test.go asserts the two stay in lockstep.
var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,49}$`)

const keyRule = "lowercase letters, digits and underscores, starting with a letter"

// Every value in this file is declared here once and referenced everywhere
// else, including SQL arguments and tests.

const (
	KindFarmPlot        = "farm_plot"
	KindMineSite        = "mine_site"
	KindWashingStation  = "washing_station"
	KindProcessingPlant = "processing_plant"
	KindWarehouse       = "warehouse"
)

// Kind is one curated facility kind. New kinds are added by the platform
// (template:manage); facilities and templates name kinds, never invent them.
type Kind struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Sub    string `json:"sub"`
	Active bool   `json:"active"`
}

const (
	// StatusPending: registered, waiting for the platform to approve it. It
	// cannot originate goods and is not on the ledger yet.
	StatusPending   = "pending_approval"
	StatusActive    = "active"
	StatusSuspended = "suspended"
)

var Statuses = []string{StatusPending, StatusActive, StatusSuspended}

// SettableStatuses are the ones POST /facilities/{id}/status may set; a
// pending facility becomes active only by approval.
var SettableStatuses = []string{StatusActive, StatusSuspended}

// The period a facility's capacity applies to.
const (
	PeriodSeason = "season"
	PeriodYear   = "year"
	PeriodMonth  = "month"
)

var Periods = []string{PeriodSeason, PeriodYear, PeriodMonth}

// GeoJSON geometry types accepted for a location: a pin or a drawn boundary.
const (
	GeoPoint   = "Point"
	GeoPolygon = "Polygon"
)

var GeoTypes = []string{GeoPoint, GeoPolygon}

const (
	MaxNameLen = 200
	MaxUnitLen = 16

	// SRID is WGS 84, the longitude/latitude system GeoJSON uses.
	SRID = 4326

	// A season is half a year: season 1 is January to June, season 2 July to December.
	firstSeasonLastMonth = time.June

	minLon, maxLon = -180.0, 180.0
	minLat, maxLat = -90.0, 90.0

	// DateLayout is how licence dates are written: 2027-12-31.
	DateLayout = "2006-01-02"

	// createLock serialises facility registration, so two overlapping
	// facilities cannot both pass the overlap check. Arbitrary.
	createLock = 7203116
)

// Multipart field beside a licence document: its (new) end date.
const FormLicenceValidTo = "licence_valid_to"

// URL parameter names.
const (
	ParamStatus  = "status"
	ParamKind    = "kind"
	ParamKindKey = "key"  // a facility kind in /facility-kinds/{key}
	ParamBBox    = "bbox" // minLon,minLat,maxLon,maxLat
)

// Messages shown to callers.
const (
	MsgBadKind       = "kind must be " + keyRule
	MsgBadKindKey    = "kind key must be " + keyRule
	MsgBadKindName   = "kind needs a label up to 120 characters"
	MsgBadKindSub    = "kind description may be up to 500 characters"
	MsgKindInUse     = "this kind is used by facilities or templates and cannot be deleted"
	MsgBadGeometry   = "location is not a valid shape (a boundary must not cross itself)"
	MsgBadBBox       = ParamBBox + " must be minLon,minLat,maxLon,maxLat"
	MsgLicenceDate   = "licence_valid_to must be a date, e.g. 2027-12-31"
	MsgNotPending    = "this facility is not waiting for approval"
	MsgPending       = "this facility is waiting for approval; its status changes once it is approved"
	MsgNoLicenceFile = "no licence document has been uploaded for this facility"
	// MsgOverlapsOther names nothing: another organisation's facility is not the caller's to see.
	MsgOverlapsOther = "overlaps a facility registered by another organisation"
)

// MsgOverlaps names the caller's own facility the new one overlaps.
func MsgOverlaps(name string) string { return "overlaps " + name }

// MsgUnknownKind names the active kinds a facility or template may use.
func MsgUnknownKind(active []string) string {
	return "kind must be one of: " + strings.Join(active, ", ")
}

// Confirmations the app shows as they are.
const (
	MsgApproved     = "Facility approved. It can issue passports now."
	MsgLicenceSaved = "Licence uploaded."
	MsgDeleted      = "Facility deleted. Passports already issued there keep their records."
	MsgKindCreated  = "Facility kind added. Forms offer it now."
	MsgKindUpdated  = "Facility kind updated."
	MsgKindDeleted  = "Facility kind deleted."
	MsgKindTaken    = "a facility kind with this key already exists"
)

// statusMessage confirms a status change.
var statusMessage = map[string]string{
	StatusSuspended: "Facility suspended. It can't issue passports until it's reinstated.",
	StatusActive:    "Facility reinstated.",
}

// MsgRegistered confirms a new facility, which waits for approval.
func MsgRegistered(name string) string {
	return name + " is registered. It can issue passports once the platform approves it."
}
