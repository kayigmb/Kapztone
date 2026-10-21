package org

import (
	"fmt"
	"net/mail"
	"slices"
	"strings"
	"time"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/template"
)

// Registration is what a new business submits: itself and its first user.
type Registration struct {
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Kind     string `json:"kind"`
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate normalises the input and returns the first problem as a message
// that is safe to show, or "".
func (r *Registration) Validate() string {
	r.Name = strings.TrimSpace(r.Name)
	r.Slug = strings.TrimSpace(r.Slug)
	r.UserName = strings.TrimSpace(r.UserName)
	r.Email = strings.TrimSpace(r.Email)

	switch {
	case r.Name == "" || len(r.Name) > MaxNameLen:
		return fmt.Sprintf("name is required, up to %d characters", MaxNameLen)
	case !auth.ValidSlug(r.Slug):
		return fmt.Sprintf("slug must be lowercase letters, digits or hyphens, and not one of: %s",
			strings.Join(auth.ReservedSlugs, ", "))
	case !slices.Contains(Kinds, r.Kind):
		return "kind must be one of: " + strings.Join(Kinds, ", ")
	case r.UserName == "" || len(r.UserName) > MaxNameLen:
		return fmt.Sprintf("user_name is required, up to %d characters", MaxNameLen)
	case !validEmail(r.Email):
		return "email is not a valid address"
	}
	if problem := auth.PasswordFormatProblem(r.Password); problem != "" {
		return problem
	}
	return ""
}

func validEmail(s string) bool {
	a, err := mail.ParseAddress(s)
	return err == nil && a.Address == s && len(s) <= 254
}

// Review is a reviewer's decision on an organisation.
type Review struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

func (r *Review) Validate() string {
	r.Reason = strings.TrimSpace(r.Reason)
	switch {
	case !slices.Contains(Decisions, r.Decision):
		return "decision must be one of: " + strings.Join(Decisions, ", ")
	case slices.Contains(Reopened, r.Decision) && r.Reason == "":
		return MsgReasonNeeded
	}
	return ""
}

// AccreditationInput is one certificate type an organisation may issue, as sent.
type AccreditationInput struct {
	Type      string `json:"type"`
	ValidFrom string `json:"valid_from"`
	ValidTo   string `json:"valid_to"`
}

// Grant is a checked AccreditationInput, with times in whole seconds as the chain holds them.
type Grant struct {
	Type      string
	ValidFrom time.Time
	ValidTo   time.Time
}

// ParseGrants checks the full set of accreditations sent for an organisation.
// It returns the first problem as a message that is safe to show, or "".
// Whether each type is one a template asks for is checked by the store.
func ParseGrants(in []AccreditationInput) ([]Grant, string) {
	grants := make([]Grant, 0, len(in))
	for _, a := range in {
		a.Type = strings.TrimSpace(a.Type)
		from, okFrom := httpx.ParseTime(a.ValidFrom)
		to, okTo := httpx.ParseTime(a.ValidTo)
		switch {
		case !template.ValidKey(a.Type):
			return nil, "type must be " + template.KeyRule
		case slices.ContainsFunc(grants, func(g Grant) bool { return g.Type == a.Type }):
			return nil, MsgDuplicateType + ": " + a.Type
		case !okFrom || !okTo || !to.After(from) || from.Unix() < 0: // the chain holds times as unsigned seconds
			return nil, MsgAccreditationDates
		}
		grants = append(grants, Grant{Type: a.Type, ValidFrom: from.Truncate(time.Second), ValidTo: to.Truncate(time.Second)})
	}
	return grants, ""
}
