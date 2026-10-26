package template

import (
	"fmt"
	"slices"
	"strings"
)

type CapacityRule struct {
	Type             string  `json:"type"`
	KgPerHaPerSeason float64 `json:"kgPerHaPerSeason,omitempty"`
}

// Input is a commodity template as written in PLAN.md §4.2.
type Input struct {
	Commodity          string       `json:"commodity"`
	Name               string       `json:"name"`
	Unit               string       `json:"unit"`
	OriginFacilityKind string       `json:"originFacilityKind"`
	CapacityRule       CapacityRule `json:"capacityRule"`
	// RequiredEvidence are the certificate types the origin facility must
	// hold before a passport is issued under the template.
	RequiredEvidence []string `json:"requiredEvidence"`
}

// Validate normalises the input and returns the first problem as a message
// that is safe to show, or "". A wrong template would silently break mass
// balance and capacity checks later, so everything is checked here.
func (in *Input) Validate() string {
	in.Name = strings.TrimSpace(in.Name)
	in.Unit = strings.TrimSpace(in.Unit)
	if in.Name == "" {
		in.Name = in.Commodity
	}
	if problem := in.headerProblem(); problem != "" {
		return problem
	}
	if in.RequiredEvidence == nil {
		in.RequiredEvidence = []string{}
	}
	for i, ev := range in.RequiredEvidence {
		switch {
		case !keyPattern.MatchString(ev):
			return "requiredEvidence entries must be " + keyRule
		case slices.Contains(in.RequiredEvidence[:i], ev):
			return "requiredEvidence " + ev + " is listed twice"
		}
	}
	return ""
}

func (in *Input) headerProblem() string {
	switch {
	case !keyPattern.MatchString(in.Commodity):
		return "commodity must be " + keyRule
	case len(in.Name) > MaxNameLen:
		return fmt.Sprintf("name may be up to %d characters", MaxNameLen)
	case in.Unit == "" || len(in.Unit) > MaxUnitLen:
		return fmt.Sprintf("unit is required, up to %d characters", MaxUnitLen)
	case in.OriginFacilityKind == "":
		return "originFacilityKind is required"
	case !keyPattern.MatchString(in.OriginFacilityKind):
		return "originFacilityKind must be " + keyRule
	case !slices.Contains(CapacityTypes, in.CapacityRule.Type):
		return "capacityRule.type must be one of: " + strings.Join(CapacityTypes, ", ")
	case in.CapacityRule.Type == CapacityYieldPerArea && in.CapacityRule.KgPerHaPerSeason <= 0:
		return "capacityRule.kgPerHaPerSeason must be above 0 for " + CapacityYieldPerArea
	case in.CapacityRule.Type != CapacityYieldPerArea && in.CapacityRule.KgPerHaPerSeason != 0:
		return "capacityRule.kgPerHaPerSeason only applies to " + CapacityYieldPerArea
	}
	return ""
}
