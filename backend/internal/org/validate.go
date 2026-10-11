package org

import (
	"fmt"
	"net/mail"
	"slices"
	"strings"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
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
	case len(r.Password) < auth.MinPasswordLen || len(r.Password) > auth.MaxPasswordLen:
		return fmt.Sprintf("password must be %d to %d characters", auth.MinPasswordLen, auth.MaxPasswordLen)
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
	case r.Decision == StatusRejected && r.Reason == "":
		return MsgReasonNeeded
	}
	return ""
}
