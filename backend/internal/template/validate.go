package template

import (
	"fmt"
	"slices"
	"strings"

	"github.com/kayigmb/chaintrack/backend/internal/facility"
)

// Stage is one step of a commodity's journey.
type Stage struct {
	Key              string   `json:"key"`
	Events           []string `json:"events"`
	MaxYield         *float64 `json:"maxYield,omitempty"` // output ≤ input × MaxYield on TRANSFORM
	RequiredEvidence []string `json:"requiredEvidence,omitempty"`
}

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
	Stages             []Stage      `json:"stages"`
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

	switch {
	case !keyPattern.MatchString(in.Commodity):
		return "commodity must be " + keyRule
	case len(in.Name) > MaxNameLen:
		return fmt.Sprintf("name may be up to %d characters", MaxNameLen)
	case in.Unit == "" || len(in.Unit) > MaxUnitLen:
		return fmt.Sprintf("unit is required, up to %d characters", MaxUnitLen)
	case !slices.Contains(facility.Kinds, in.OriginFacilityKind):
		return "originFacilityKind must be one of: " + strings.Join(facility.Kinds, ", ")
	case !slices.Contains(CapacityTypes, in.CapacityRule.Type):
		return "capacityRule.type must be one of: " + strings.Join(CapacityTypes, ", ")
	case in.CapacityRule.Type == CapacityYieldPerArea && in.CapacityRule.KgPerHaPerSeason <= 0:
		return "capacityRule.kgPerHaPerSeason must be above 0 for " + CapacityYieldPerArea
	case in.CapacityRule.Type != CapacityYieldPerArea && in.CapacityRule.KgPerHaPerSeason != 0:
		return "capacityRule.kgPerHaPerSeason only applies to " + CapacityYieldPerArea
	case len(in.Stages) == 0:
		return "at least one stage is required"
	case !slices.Contains(in.Stages[0].Events, EventIssue):
		return "the first stage must allow " + EventIssue
	}

	seen := map[string]bool{}
	for i, st := range in.Stages {
		at := fmt.Sprintf("stage %d (%s): ", i+1, st.Key)
		switch {
		case !keyPattern.MatchString(st.Key):
			return at + "key must be " + keyRule
		case seen[st.Key]:
			return at + "key is used twice"
		case len(st.Events) == 0:
			return at + "needs at least one event"
		case i > 0 && slices.Contains(st.Events, EventIssue):
			return at + EventIssue + " is only allowed in the first stage"
		}
		seen[st.Key] = true

		for j, ev := range st.Events {
			if !slices.Contains(Events, ev) {
				return at + "events must be from: " + strings.Join(Events, ", ")
			}
			if slices.Contains(st.Events[:j], ev) {
				return at + "event " + ev + " is listed twice"
			}
		}
		if st.MaxYield != nil {
			switch {
			case !slices.Contains(st.Events, EventTransform):
				return at + "maxYield only applies to a stage that allows " + EventTransform
			case *st.MaxYield <= 0:
				return at + "maxYield must be above 0"
			}
		}
		for _, ev := range st.RequiredEvidence {
			if !keyPattern.MatchString(ev) {
				return at + "requiredEvidence entries must be " + keyRule
			}
		}
	}
	return ""
}
