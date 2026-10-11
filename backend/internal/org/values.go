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
	StatusPending  = auth.OrgPending
	StatusVerified = auth.OrgVerified
	StatusRejected = auth.OrgRejected
)

var Statuses = []string{StatusPending, StatusVerified, StatusRejected}

// Decisions a reviewer can make are the two final statuses.
var Decisions = []string{StatusVerified, StatusRejected}

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
const ParamStatus = "status"

// Messages shown to callers.
const (
	MsgSlugTaken    = "this organisation address is already taken"
	MsgEmailTaken   = "an account with this email already exists"
	MsgNoDocuments  = "this organisation has no documents to review; it cannot be verified"
	MsgReasonNeeded = "give a reason for rejecting"
	MsgNotIssuer    = "only a verified certifier or regulator can be accredited"
)
