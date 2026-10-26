// Package user lets an organisation's admin manage who works for it: invite
// people by e-mail, change their role, suspend and reinstate them. Every
// change is limited to the acting organisation.
package user

import (
	"time"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
)

// Every value in this file is declared here once and referenced everywhere
// else, including SQL arguments and tests.

// Roles an organisation admin may give. The platform role is not among them.
var AssignableRoles = []string{auth.RoleOrgAdmin, auth.RoleMember}

// A member's status in one organisation. Suspending removes access to that
// organisation only; the person's account and other organisations stay.
const (
	StatusActive    = "active"
	StatusSuspended = "suspended"
)

// ?status= on GET /users: a member's place in the organisation, or for the
// platform's list of every account, the account's own status.
const ParamStatus = "status"

var (
	MemberStatuses  = []string{StatusActive, StatusSuspended}
	AccountStatuses = []string{auth.UserActive, auth.UserDisabled}
)

const (
	InviteTTL   = 24 * time.Hour
	tokenBytes  = 32
	ParamToken  = "token"
	MaxEmailLen = 254
	MaxNameLen  = 200

	// InvitePath is the frontend page an invitation link opens: InvitePath + token.
	InvitePath    = "/invite/"
	InviteSubject = "You are invited to ChainTrack"
)

// Messages shown to callers.
const (
	MsgEmail                      = "enter a valid e-mail address"
	MsgName                       = "enter your name"
	MsgInviteInvalid              = "this invitation is not valid, was already used, or has expired"
	MsgAlreadyMember              = "this person already belongs to the organisation"
	MsgLastAdmin                  = "the organisation must keep at least one admin"
	MsgSelf                       = "you cannot suspend yourself"
	MsgSelfRole                   = "you cannot change your own role"
	MsgNotMember                  = "this person does not belong to the organisation"
	MsgNotSuspended               = "this person is not suspended"
	MsgInviteBodyFmt              = "%s has invited you to join %s on ChainTrack as %s.\n\nAccept the invitation here:\n\n%s"
	MsgPasswordFormat             = "password must be %d to %d characters"
	MsgNameLength                 = "name must be 1 to 200 characters"
	MsgEmailNotificationsRequired = "email must be true or false"
)

// Confirmations the app shows as they are.
const (
	MsgNameSaved = "Name saved."

	MsgEmailNotificationsOn  = "Notifications will be e-mailed to you."
	MsgEmailNotificationsOff = "Notifications will no longer be e-mailed to you."
	MsgInviteLink            = "Invitation made. E-mail isn't set up, so copy the link and send it."
	MsgInviteAccepted        = "Invitation accepted. Sign in at the organisation's address."
	MsgRoleChanged           = "Role changed."
	MsgSuspended             = "Suspended. They can't act for this organisation until reinstated."
	MsgReinstated            = "Reinstated."
)

// MsgInviteSent confirms an invitation e-mailed to email.
func MsgInviteSent(email string) string { return "Invitation sent to " + email + "." }
