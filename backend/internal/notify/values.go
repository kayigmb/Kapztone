// Package notify tells organisations about things that concern them: in the
// app, and by e-mail when a mail server is configured.
package notify

import "time"

// Every value in this file is declared here once and referenced everywhere
// else, including SQL arguments and tests.

// What a notification is about. subject_id is the handover, passport,
// certificate or flag concerned.
const (
	KindHandoverStarted    = "handover_started"
	KindHandoverAccepted   = "handover_accepted"
	KindHandoverRejected   = "handover_rejected"
	KindCertificateIssued  = "certificate_issued"
	KindCertificateRevoked = "certificate_revoked"
	KindFlagRaised         = "flag_raised"
	KindFlagResolved       = "flag_resolved"
	KindPassportRevoked    = "passport_revoked"
)

// URL parameter names.
const ParamUnread = "unread"

const (
	// How often the mailer looks for notifications to send.
	mailInterval = 10 * time.Second
	// Notifications older than this are never e-mailed: turning mail on must
	// not send out a backlog.
	mailMaxAge = 24 * time.Hour
	mailBatch  = 20

	mailSubject = "ChainTrack notification"
)
