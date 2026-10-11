// Package facility handles the places where goods originate or are handled:
// farm plots, mine sites, washing stations, plants and warehouses.
package facility

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
	StatusActive    = "active"
	StatusSuspended = "suspended"
)

var Statuses = []string{StatusActive, StatusSuspended}

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

	minLon, maxLon = -180.0, 180.0
	minLat, maxLat = -90.0, 90.0
)

// URL parameter names.
const (
	ParamStatus = "status"
	ParamBBox   = "bbox" // minLon,minLat,maxLon,maxLat
)

// Messages shown to callers.
const (
	MsgBadGeometry = "location is not a valid shape (a boundary must not cross itself)"
	MsgBadBBox     = ParamBBox + " must be minLon,minLat,maxLon,maxLat"
)
