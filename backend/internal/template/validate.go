package template

import (
	"fmt"
	"slices"
	"strings"
)

// Stage is one step of a commodity's journey.
type Stage struct {
	Key              string   `json:"key"`
	Events           []string `json:"events"`
	MaxYield         *float64 `json:"maxYield,omitempty"` // output ≤ input × MaxYield on TRANSFORM
	RequiredEvidence []string `json:"requiredEvidence,omitempty"`
	// HandoverTolerance is how far, in percent, the weight a receiver records
	// may differ from what was sent; unset means DefaultHandoverTolerance.
	HandoverTolerance *float64 `json:"handoverTolerance,omitempty"`
}

// Tolerance is the weight difference, in percent, accepted at this stage.
func (st Stage) Tolerance() float64 {
	if st.HandoverTolerance != nil {
		return *st.HandoverTolerance
	}
	return DefaultHandoverTolerance
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
	if problem := in.headerProblem(); problem != "" {
		return problem
	}
	seen := map[string]bool{}
	for i, st := range in.Stages {
		at := fmt.Sprintf("stage %d (%s): ", i+1, st.Key)
		if problem := stageProblem(i, st, seen); problem != "" {
			return at + problem
		}
		seen[st.Key] = true
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
	case len(in.Stages) == 0:
		return "at least one stage is required"
	case !slices.Contains(in.Stages[0].Events, EventIssue):
		return "the first stage must allow " + EventIssue
	}
	return ""
}

// stageProblem checks stage i; seen holds the keys of the stages before it.
func stageProblem(i int, st Stage, seen map[string]bool) string {
	switch {
	case !keyPattern.MatchString(st.Key):
		return "key must be " + keyRule
	case seen[st.Key]:
		return "key is used twice"
	case st.Key == StageSplit:
		return "key " + StageSplit + " is reserved for splitting a lot at a handover"
	case st.HandoverTolerance != nil && !slices.Contains(st.Events, EventHandover):
		return "handoverTolerance only applies to a stage that allows " + EventHandover
	case st.HandoverTolerance != nil && (*st.HandoverTolerance < 0 || *st.HandoverTolerance > 100):
		return "handoverTolerance must be a percentage from 0 to 100"
	case len(st.Events) == 0:
		return "needs at least one event"
	case i > 0 && slices.Contains(st.Events, EventIssue):
		return EventIssue + " is only allowed in the first stage"
	}
	if problem := eventsProblem(st.Events); problem != "" {
		return problem
	}
	switch {
	case st.MaxYield != nil && !slices.Contains(st.Events, EventTransform):
		return "maxYield only applies to a stage that allows " + EventTransform
	case st.MaxYield != nil && *st.MaxYield <= 0:
		return "maxYield must be above 0"
	}
	for _, ev := range st.RequiredEvidence {
		if !keyPattern.MatchString(ev) {
			return "requiredEvidence entries must be " + keyRule
		}
	}
	return ""
}

func eventsProblem(events []string) string {
	for j, ev := range events {
		if !slices.Contains(Events, ev) {
			return "events must be from: " + strings.Join(Events, ", ")
		}
		if slices.Contains(events[:j], ev) {
			return "event " + ev + " is listed twice"
		}
	}
	return ""
}
