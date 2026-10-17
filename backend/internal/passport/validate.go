package passport

import (
	"fmt"
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

// HandoverInput is what the holder submits to hand a passport over.
type HandoverInput struct {
	ToOrganisationID string `json:"to_organisation_id"`
	// Stage is the template stage the lot enters with the receiver.
	Stage string `json:"stage"`
	Note  string `json:"note"` // where and how the goods were handed over
}

func (in *HandoverInput) Validate() string {
	in.Note, in.Stage = strings.TrimSpace(in.Note), strings.TrimSpace(in.Stage)
	switch {
	case !httpx.IsUUID(in.ToOrganisationID):
		return "to_organisation_id is required"
	case in.Stage == "":
		return "stage is required"
	case len(in.Note) > MaxNoteLen:
		return fmt.Sprintf("note may be up to %d characters", MaxNoteLen)
	}
	return ""
}

// RejectInput is the receiver's reason for refusing a handover.
type RejectInput struct {
	Reason string `json:"reason"`
}

func (in *RejectInput) Validate() string {
	in.Reason = strings.TrimSpace(in.Reason)
	switch {
	case in.Reason == "":
		return MsgReasonNeeded
	case len(in.Reason) > MaxNoteLen:
		return fmt.Sprintf("reason may be up to %d characters", MaxNoteLen)
	}
	return ""
}
