package passport

import (
	"fmt"
	"slices"
	"strings"

	"github.com/kayigmb/chaintrack/backend/internal/httpx"
)

// IssueInput is what a producer submits to issue an origin passport.
type IssueInput struct {
	FacilityID string `json:"facility_id"`
	TemplateID string `json:"template_id"`
	Quantity   int64  `json:"quantity"` // whole number, in the template's unit
}

// Validate returns the first problem as a message that is safe to show, or "".
func (in IssueInput) Validate() string {
	switch {
	case !httpx.IsUUID(in.FacilityID):
		return "facility_id is required"
	case !httpx.IsUUID(in.TemplateID):
		return "template_id is required"
	case in.Quantity <= 0:
		return "quantity must be a whole number above 0"
	}
	return ""
}

// HandoverInput is what the holder submits to hand a passport over, as JSON
// or as a multipart form when photos come with it.
type HandoverInput struct {
	ToOrganisationID string `json:"to_organisation_id"`
	ToOrg            string `json:"to_org"` // the same, under its shorter name
	// Stage is the template stage the lot enters with the receiver.
	Stage string `json:"stage"`
	// Quantity is how much is handed over; unset means all that remains.
	// Less than that splits the part off into its own passport first.
	Quantity *int64   `json:"quantity"`
	Note     string   `json:"note"` // where and how the goods were handed over
	Lat      *float64 `json:"lat"`
	Lon      *float64 `json:"lon"`
	SealNumber string `json:"seal_number"`
	Checklist  map[string]bool `json:"checklist"`
}

func (in *HandoverInput) Validate() string {
	in.Note, in.Stage = strings.TrimSpace(in.Note), strings.TrimSpace(in.Stage)
	in.SealNumber = strings.TrimSpace(in.SealNumber)
	if in.ToOrganisationID == "" {
		in.ToOrganisationID = in.ToOrg
	}
	switch {
	case !httpx.IsUUID(in.ToOrganisationID):
		return "to_org is required"
	case in.Stage == "":
		return "stage is required"
	case in.Quantity != nil && *in.Quantity <= 0:
		return "quantity must be a whole number above 0"
	case len(in.Note) > MaxNoteLen:
		return fmt.Sprintf("note may be up to %d characters", MaxNoteLen)
	case len(in.SealNumber) > 64:
		return "seal_number may be up to 64 characters"
	case len(in.Checklist) > 32:
		return "checklist may hold up to 32 items"
	}
	return locationProblem(in.Lat, in.Lon)
}

// locationProblem checks an optional point: both coordinates or neither, in range.
func locationProblem(lat, lon *float64) string {
	if (lat == nil) != (lon == nil) || (lat != nil && (*lat < -90 || *lat > 90 || *lon < -180 || *lon > 180)) {
		return MsgLocation
	}
	return ""
}

// AcceptInput is the receiver's half: what they weighed, and where.
type AcceptInput struct {
	ReceivedQuantity *int64   `json:"received_quantity"`
	Lat              *float64 `json:"lat"`
	Lon              *float64 `json:"lon"`
}

func (in *AcceptInput) Validate() string {
	if in.ReceivedQuantity == nil || *in.ReceivedQuantity < 0 {
		return MsgReceivedNeeded
	}
	return locationProblem(in.Lat, in.Lon)
}

// RejectInput is why the receiver refuses a handover.
type RejectInput struct {
	Reasons []string `json:"reasons"` // from RejectReasons
	Note    string   `json:"note"`
	Reason  string   `json:"reason"` // older name for the note
	// AlsoFlag raises a flag about the passport with the rejection.
	AlsoFlag bool `json:"also_flag"`
}

func (in *RejectInput) Validate() string {
	if in.Note == "" {
		in.Note = in.Reason
	}
	in.Note = strings.TrimSpace(in.Note)
	if len(in.Reasons) == 0 {
		return MsgReasonsNeeded
	}
	for i, r := range in.Reasons {
		if !slices.Contains(RejectReasons, r) || slices.Contains(in.Reasons[:i], r) {
			return "reasons must be distinct values from: " + strings.Join(RejectReasons, ", ")
		}
	}
	if len(in.Note) > MaxNoteLen {
		return fmt.Sprintf("note may be up to %d characters", MaxNoteLen)
	}
	return ""
}
