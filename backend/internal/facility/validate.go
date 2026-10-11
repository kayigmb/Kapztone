package facility

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// Input is what a business submits to register a facility.
type Input struct {
	Name           string          `json:"name"`
	Kind           string          `json:"kind"`
	Location       json.RawMessage `json:"location"` // GeoJSON geometry
	LicenceNumber  string          `json:"licence_number"`
	Capacity       int64           `json:"capacity"`
	CapacityUnit   string          `json:"capacity_unit"`
	CapacityPeriod string          `json:"capacity_period"`
}

// Validate normalises the input and returns the first problem as a message
// that is safe to show, or "".
func (in *Input) Validate() string {
	in.Name = strings.TrimSpace(in.Name)
	in.LicenceNumber = strings.TrimSpace(in.LicenceNumber)
	in.CapacityUnit = strings.TrimSpace(in.CapacityUnit)

	switch {
	case in.Name == "" || len(in.Name) > MaxNameLen:
		return fmt.Sprintf("name is required, up to %d characters", MaxNameLen)
	case !slices.Contains(Kinds, in.Kind):
		return "kind must be one of: " + strings.Join(Kinds, ", ")
	case in.LicenceNumber == "" || len(in.LicenceNumber) > MaxNameLen:
		return fmt.Sprintf("licence_number is required, up to %d characters", MaxNameLen)
	case in.Capacity <= 0:
		return "capacity must be a whole number above 0"
	case in.CapacityUnit == "" || len(in.CapacityUnit) > MaxUnitLen:
		return fmt.Sprintf("capacity_unit is required, up to %d characters", MaxUnitLen)
	case !slices.Contains(Periods, in.CapacityPeriod):
		return "capacity_period must be one of: " + strings.Join(Periods, ", ")
	}
	return validLocation(in.Location)
}

type position []float64

func (p position) problem() string {
	if len(p) < 2 || p[0] < minLon || p[0] > maxLon || p[1] < minLat || p[1] > maxLat {
		return fmt.Sprintf("each coordinate must be [longitude, latitude] within %g…%g and %g…%g",
			minLon, maxLon, minLat, maxLat)
	}
	return ""
}

// validLocation checks the shape of the GeoJSON and the coordinate ranges.
// Whether a boundary crosses itself is checked by PostGIS when it is stored.
func validLocation(raw json.RawMessage) string {
	var g struct {
		Type        string          `json:"type"`
		Coordinates json.RawMessage `json:"coordinates"`
	}
	if json.Unmarshal(raw, &g) != nil || !slices.Contains(GeoTypes, g.Type) {
		return "location must be GeoJSON of type: " + strings.Join(GeoTypes, ", ")
	}
	bad := "location coordinates do not match type " + g.Type

	switch g.Type {
	case GeoPoint:
		var p position
		if json.Unmarshal(g.Coordinates, &p) != nil {
			return bad
		}
		return p.problem()
	case GeoPolygon:
		var rings [][]position
		if json.Unmarshal(g.Coordinates, &rings) != nil || len(rings) == 0 {
			return bad
		}
		for _, ring := range rings {
			if len(ring) < 4 {
				return "a boundary needs at least 4 points, the last equal to the first"
			}
			for _, p := range ring {
				if msg := p.problem(); msg != "" {
					return msg
				}
			}
			if first, last := ring[0], ring[len(ring)-1]; first[0] != last[0] || first[1] != last[1] {
				return "a boundary must end at the point it starts from"
			}
		}
	}
	return ""
}

// BBox is a map window: minLon, minLat, maxLon, maxLat.
type BBox [4]float64

// ParseBBox reads "minLon,minLat,maxLon,maxLat"; nil means no window was given.
func ParseBBox(s string) (*BBox, bool) {
	if s == "" {
		return nil, true
	}
	parts := strings.Split(s, ",")
	if len(parts) != len(BBox{}) {
		return nil, false
	}
	var b BBox
	for i, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return nil, false
		}
		b[i] = v
	}
	if (position{b[0], b[1]}).problem() != "" || (position{b[2], b[3]}).problem() != "" || b[0] >= b[2] || b[1] >= b[3] {
		return nil, false
	}
	return &b, true
}

// StatusInput changes a facility's status.
type StatusInput struct {
	Status string `json:"status"`
}

func (in StatusInput) Validate() string {
	if !slices.Contains(Statuses, in.Status) {
		return "status must be one of: " + strings.Join(Statuses, ", ")
	}
	return ""
}
