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

// Upload rules.
const (
	MaxUploadBytes = 10 << 20
	MaxNameLen     = 200

	FormFile = "file" // multipart field holding the document
	FormType = "type" // multipart field holding the document type
)

// File types accepted for KYB documents, detected from the file's content.
var AllowedContentTypes = []string{"application/pdf", "image/png", "image/jpeg"}

// URL parameter names.
const ParamStatus = "status"

// Messages shown to callers.
const (
	MsgSlugTaken    = "this organisation address is already taken"
	MsgEmailTaken   = "an account with this email already exists"
	MsgNoDocuments  = "this organisation has no documents to review; it cannot be verified"
	MsgReasonNeeded = "give a reason for rejecting"
	MsgFileNeeded   = "attach the document in the 'file' field"
	MsgFileTooLarge = "file is larger than 10 MB"
	MsgFileType     = "file must be a PDF, PNG or JPEG"
)
