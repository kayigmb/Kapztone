// Package org handles organisation registration and KYB review.
package org

import "github.com/kayigmb/chaintrack/backend/internal/auth"

// Every value in this file is declared here once and referenced everywhere
// else, including SQL arguments and tests.

// What kind of business an organisation is.
const (
	KindProducer    = "producer"
	KindProcessor   = "processor"
	KindTransporter = "transporter"
	KindExporter    = "exporter"
	KindImporter    = "importer"
	KindCertifier   = "certifier"
	KindRegulator   = "regulator"
)

var Kinds = []string{
	KindProducer, KindProcessor, KindTransporter, KindExporter, KindImporter, KindCertifier, KindRegulator,
}

// IssuerKinds are the kinds of organisation that can be accredited to issue certificates.
var IssuerKinds = []string{KindCertifier, KindRegulator}

// KYB statuses are declared in auth, which checks them on every request.
const (
	StatusPending          = auth.OrgPending
	StatusVerified         = auth.OrgVerified
	StatusRejected         = auth.OrgRejected
	StatusChangesRequested = auth.OrgChangesRequested
)

var Statuses = []string{StatusPending, StatusVerified, StatusRejected, StatusChangesRequested}

// Decisions a reviewer can make: every status but pending.
var Decisions = []string{StatusVerified, StatusRejected, StatusChangesRequested}

// Reopened are the decisions that a new document upload sends back to pending.
// Both need a reason, which the organisation is shown.
var Reopened = []string{StatusRejected, StatusChangesRequested}

// KYB document types.
const (
	DocRegistration = "registration"
	DocLicence      = "licence"
	DocOther        = "other"
)

var DocTypes = []string{DocRegistration, DocLicence, DocOther}

const (
	MaxNameLen = 200

	FormType = "type" // multipart field holding the document type
)

// URL parameter names.
const (
	ParamStatus = "status"
	ParamKind   = "kind"
)

// Messages shown to callers.
const (
	MsgSlugTaken    = "this organisation address is already taken"
	MsgEmailTaken   = "an account with this email already exists"
	MsgNoDocuments  = "this organisation has no documents to review; it cannot be verified"
	MsgReasonNeeded = "give a reason for rejecting or asking for changes"
	MsgNotIssuer    = "only a verified certifier or regulator can be accredited"

	MsgNotifyRegistered = "A new organisation registered and is waiting for verification."
	MsgNotifyDocuments  = "An organisation uploaded documents for review."

	MsgUnknownCertType    = "each type must be evidence that a product template asks for"
	MsgDuplicateType      = "a type is listed more than once"
	MsgAccreditationDates = "valid_from and valid_to must be dates (2026-01-31) or RFC 3339 timestamps, with valid_to after valid_from"
)

// Confirmations the app shows as they are.
const (
	MsgRegistered            = "Account created. Upload your documents so we can verify the organisation."
	MsgAccreditationsSaved   = "Accreditations saved."
	MsgAccreditationsRemoved = "Accreditations removed."
	MsgDocumentUploaded      = "Document uploaded."
)

// reviewMessage confirms a review decision.
var reviewMessage = map[string]string{
	StatusVerified:         "Organisation approved. It can trade now.",
	StatusRejected:         "Organisation rejected. It sees your note.",
	StatusChangesRequested: "Changes requested. The organisation sees your note.",
}
