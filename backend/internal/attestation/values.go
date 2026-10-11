// Package attestation handles certificates: documents an accredited
// organisation issues about a passport or a facility. The file stays off
// chain; its hash is anchored, so an altered file no longer matches.
package attestation

// Every value in this file is declared here once and referenced everywhere
// else, including SQL arguments and tests.

// What a certificate is about.
const (
	SubjectPassport = "passport"
	SubjectFacility = "facility"
)

var SubjectTypes = []string{SubjectPassport, SubjectFacility}

// subjectTable is where each kind of subject lives. Table names come from
// here, never from the request.
var subjectTable = map[string]string{
	SubjectPassport: "passports",
	SubjectFacility: "facilities",
}

// Certificate statuses.
const (
	StatusValid   = "valid"
	StatusRevoked = "revoked"
	// StatusExpired is never stored: a valid certificate past valid_to reads as expired.
	StatusExpired = "expired"
)

var Statuses = []string{StatusValid, StatusRevoked, StatusExpired}

// Multipart fields beside the file.
const (
	FormSubjectType = "subject_type"
	FormSubjectID   = "subject_id"
	FormType        = "type"
	FormValidFrom   = "valid_from"
	FormValidTo     = "valid_to"
)

// URL parameter names.
const (
	ParamStatus    = "status"
	ParamSubjectID = "subject_id"
)

const MaxReasonLen = 1000

// Messages shown to callers.
const (
	MsgNotAccredited   = "your organisation is not accredited to issue certificates"
	MsgSubjectNotFound = "the passport or facility this certificate is about was not found"
	MsgAlreadyRevoked  = "this certificate is already revoked"
	MsgReasonNeeded    = "give a reason for revoking"
	MsgDates           = "valid_from and valid_to must be dates (2026-01-31) or RFC 3339 timestamps, with valid_to after valid_from"

	MsgNotifyIssued  = "A certificate was issued for something your organisation holds."
	MsgNotifyRevoked = "A certificate about something your organisation holds was revoked."
)
