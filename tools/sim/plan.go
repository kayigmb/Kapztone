package main

import (
	"fmt"
	"math/rand/v2"
)

// The plan is every random choice of a run, drawn from the seed alone and
// before anything is sent: the same seed gives the same dataset and the same
// fraud targets, whatever ids the API hands out.

// Fraud kinds (PLAN.md §6.1).
const (
	F1 = "F1" // record alteration
	F2 = "F2" // unilateral or forged handover
	F3 = "F3" // unverified origin
	F4 = "F4" // quantity inflation
	F5 = "F5" // cloned passport
	F6 = "F6" // invalid certificate
)

// How a fraud is attempted.
const (
	viaAPI = "api" // a request a cheating business could send; ChainTrack should refuse it
	viaDB  = "db"  // an insider changing the database; verification should catch it
)

type FacilityPlan struct {
	Lon, Lat float64
	Capacity int64
}

type LotPlan struct {
	Facility    int
	Quantity    int64
	YieldFactor float64 // share of the stage's maximum yield a legitimate transform uses
	Fraud       string  // a database fraud done to this lot after sealing, or ""
	Attempt     string  // an API fraud tried while this lot moves (F2: accept by a third party; F4: over-yield transform), or ""
}

type ChainPlan struct {
	Facilities []FacilityPlan
	Lots       []LotPlan
	F3, F4Capacity int // API attempts not tied to a lot: issue from a pending facility, issue over capacity
}

const (
	facilityCapacity = 1_000_000
	minQuantity      = 500
	maxQuantity      = 2_000
	gridStep         = 0.01 // degrees between facility pins: no two overlap
)

// planChain draws one chain's facilities, lots and fraud targets. perType
// frauds of each kind; F2 and F4 alternate between their two forms, and F4
// is always over capacity where the template has no transform.
func planChain(r *rand.Rand, origin [2]float64, facilities, lots, perType int, transforms bool) (ChainPlan, error) {
	var p ChainPlan
	for i := range facilities {
		p.Facilities = append(p.Facilities, FacilityPlan{
			Lon: origin[0] + float64(i%5)*gridStep, Lat: origin[1] + float64(i/5)*gridStep, Capacity: facilityCapacity,
		})
	}
	for range lots {
		p.Lots = append(p.Lots, LotPlan{
			Facility:    r.IntN(facilities),
			Quantity:    minQuantity + r.Int64N(maxQuantity-minQuantity+1),
			YieldFactor: 0.8 + r.Float64()*0.2,
		})
	}

	var db, api []string
	for i := range perType {
		db = append(db, F1, F5, F6)
		if i%2 == 0 {
			api = append(api, F2)
		} else {
			db = append(db, F2)
		}
		if i%2 == 1 && transforms {
			api = append(api, F4)
		} else {
			p.F4Capacity++
		}
		p.F3++
	}
	if len(db)+len(api) > lots {
		return p, fmt.Errorf("%d lots can't carry %d frauds: raise -lots or lower -fraud", lots, len(db)+len(api))
	}
	order := r.Perm(lots)
	for i, kind := range db {
		p.Lots[order[i]].Fraud = kind
	}
	for i, kind := range api {
		p.Lots[order[len(db)+i]].Attempt = kind
	}
	return p, nil
}

// Outcomes, per CSV row.
const (
	prevented  = "prevented"   // fraud refused by the API
	detected   = "detected"    // fraud recorded, verification not Verified
	missed     = "missed"      // fraud recorded and verified clean, or accepted by the API
	clean      = "clean"       // legitimate lot verified
	falseAlarm = "false_alarm" // legitimate lot not verified
)

// outcome classifies one row. For an API attempt status is the HTTP status;
// for a verified lot verdict is the public verdict.
func outcome(fraud bool, mechanism string, status int, verdict string) string {
	switch {
	case fraud && mechanism == viaAPI && status >= 400 && status < 500:
		return prevented
	case fraud && mechanism == viaAPI:
		return missed
	case fraud && verdict != "verified":
		return detected
	case fraud:
		return missed
	case verdict == "verified":
		return clean
	default:
		return falseAlarm
	}
}
