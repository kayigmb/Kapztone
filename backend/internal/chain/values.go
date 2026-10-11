// Package chain turns backend values into what the contracts expect, signs
// handovers, and carries queued changes to the chain.
package chain

import "time"

// Every value in this file is declared here once and referenced everywhere
// else, including SQL arguments and tests.

// The signing domain. It must equal what CustodyManager was deployed with;
// the shared vector in contracts/test/vectors/handover.json pins both sides.
const (
	DomainName    = "ChainTrack"
	DomainVersion = "1"
)

// What a queued item asks the relayer to do.
const (
	KindOrgRegister       = "org_register"
	KindOrgVerified       = "org_set_verified"
	KindFacilityRegister  = "facility_register"
	KindFacilitySuspended = "facility_set_suspended"
	KindPassportIssue     = "passport_issue"
	KindHandover          = "handover"
)

// Where a queued item stands.
const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
	// StatusFailed: the contract refused it. It will not be retried; a person must look.
	StatusFailed = "failed"
)

const (
	// How long the relayer waits for a transaction to be included.
	minedTimeout = 2 * time.Minute
	// How often the relayer looks for new work when the queue is empty.
	PollInterval = 2 * time.Second
)
