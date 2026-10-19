// Package facility handles the places where goods originate or are handled:
// farm plots, mine sites, washing stations, plants and warehouses.
package facility

import "time"

// Every value in this file is declared here once and referenced everywhere
// else, including SQL arguments and tests.

const (
	KindFarmPlot        = "farm_plot"
	KindMineSite        = "mine_site"
	KindWashingStation  = "washing_station"
	KindProcessingPlant = "processing_plant"
	KindWarehouse       = "warehouse"
)

var Kinds = []string{KindFarmPlot, KindMineSite, KindWashingStation, KindProcessingPlant, KindWarehouse}

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
	ParamStatus = "status"
	ParamKind   = "kind"
	ParamBBox   = "bbox" // minLon,minLat,maxLon,maxLat
)

// Messages shown to callers.
const (
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

// Confirmations the app shows as they are.
const (
	MsgApproved     = "Facility approved. It can issue passports now."
	MsgLicenceSaved = "Licence uploaded."
	MsgDeleted      = "Facility deleted. Passports already issued there keep their records."
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
