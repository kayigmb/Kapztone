package verify

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kayigmb/chaintrack/backend/internal/attestation"
	"github.com/kayigmb/chaintrack/backend/internal/cache"
	"github.com/kayigmb/chaintrack/backend/internal/chain"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
	"github.com/kayigmb/chaintrack/backend/internal/upload"
)

// Reason is one thing a check found.
type Reason struct {
	Level     string `json:"level"`
	Check     string `json:"check"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	SubjectID string `json:"subject_id"` // the passport, facility or certificate concerned
}

// Check is one line of "what was checked".
type Check struct {
	Name   string `json:"name"`
	Result string `json:"result"` // verified, warning or failed
}

// Named is the public face of an organisation or facility: no licence
// numbers, no people.
type Named struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Origin struct {
	Facility struct {
		Named
		Kind     string          `json:"kind"`
		Location json.RawMessage `json:"location"` // GeoJSON
	} `json:"facility"`
	Organisation Named `json:"organisation"`
}

// RoutePoint is one facility on the map: points only, never an invented leg.
// Old handovers without GPS still show their facilities.
type RoutePoint struct {
	FacilityID string          `json:"facility_id"`
	Name       string          `json:"name"`
	Kind       string          `json:"kind"`
	Location   json.RawMessage `json:"location"` // GeoJSON
	Seq        int             `json:"seq"`
}

// Result is everything the public verification page shows.
type Result struct {
	Verdict  string            `json:"verdict"`
	Reasons  []Reason          `json:"reasons"`
	Checks   []Check           `json:"checks"`
	Passport passport.Passport `json:"passport"`
	Origins  []Origin          `json:"origins"`
	// Route is the facility points in origin order for the map.
	Route []RoutePoint `json:"route"`
	// Organisations names every organisation the answer refers to by id.
	Organisations []Named                   `json:"organisations"`
	Genealogy     passport.Genealogy        `json:"genealogy"`
	Certificates  []attestation.Attestation `json:"certificates"`
}

type Handlers struct {
	Pool         *pgxpool.Pool
	Passports    passport.Store
	Certificates attestation.Store
	// Contracts are read, never written, here.
	Contracts chain.Contracts
	// Files holds certificate files, re-hashed on every check.
	Files *upload.Bucket
	// Cache holds each answer briefly; nil means every request is worked out afresh.
	Cache *cache.Cache
}

// Verify is public: no sign-in. It takes the passport's id or its lot code.
// Each answer is cached, so a revocation can take up to the cache's lifetime to show.
func (h *Handlers) Verify(w http.ResponseWriter, r *http.Request) {
	id, err := h.Passports.Resolve(r.Context(), chi.URLParam(r, httpx.ParamID))
	if errors.Is(err, passport.ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return
	}
	if err != nil {
		httpx.ServerError(w, "find passport", err)
		return
	}
	key := CacheKeyPrefix + id
	var res Result
	if h.Cache != nil && h.Cache.Get(r.Context(), key, &res) {
		httpx.OK(w, http.StatusOK, "", res)
		return
	}
	res, err = h.Run(r.Context(), id)
	switch {
	case errors.Is(err, passport.ErrNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.MsgNotFound)
		return
	case err != nil:
		httpx.ServerError(w, "verify", err)
		return
	}
	if h.Cache != nil {
		h.Cache.Set(r.Context(), key, res)
	}
	httpx.OK(w, http.StatusOK, "", res)
}

// run carries the state of one verification.
type run struct {
	h    *Handlers
	ctx  context.Context
	opts *bind.CallOpts
	res  *Result
	// chainDown is set the first time the chain cannot be reached; no
	// further chain reads are tried.
	chainDown bool
}

func (v *run) add(level, check, code, subjectID string) {
	v.res.Reasons = append(v.res.Reasons, Reason{Level: level, Check: check, Code: code, Message: messages[code], SubjectID: subjectID})
}

// onChain interprets the error of a chain read. found is false when the
// chain does not hold the thing, or could not be asked.
func (v *run) onChain(err error) (found bool) {
	switch {
	case err == nil:
		return true
	case chain.Refused(err):
		return false
	}
	if !v.chainDown {
		v.chainDown = true
		v.add(Warning, CheckAnchored, ChainUnreachable, v.res.Passport.ID)
	}
	return false
}

// absent reports a record the chain does not hold: a warning while it waits
// for the relayer, a failure if the database claims it was anchored.
func (v *run) absent(check, subjectID string, anchored bool) {
	if v.chainDown {
		return
	}
	if anchored {
		v.add(Failed, check, ChainMissing, subjectID)
	} else {
		v.add(Warning, check, NotAnchored, subjectID)
	}
}

// Run verifies one passport back to its origins (PLAN.md §4.6).
//
// Chain reads happen one after another. If a lot with many
// ancestors is too slow on a real network, run them concurrently.
func (h *Handlers) Run(ctx context.Context, id string) (Result, error) {
	res := Result{Reasons: []Reason{}, Origins: []Origin{}, Route: []RoutePoint{}}
	v := &run{h: h, ctx: ctx, opts: &bind.CallOpts{Context: ctx}, res: &res}

	var err error
	if res.Passport, err = h.Passports.Get(ctx, id, ""); err != nil {
		return res, err
	}
	if res.Genealogy, err = h.Passports.Genealogy(ctx, id, "", passport.DirectionAncestors); err != nil {
		return res, err
	}
	if err := v.all(res.Genealogy.Nodes); err != nil {
		return res, err
	}
	res.Verdict, res.Checks = verdict(res.Reasons)
	return res, nil
}

// all runs every check over the lot and its ancestors.
func (v *run) all(nodes []passport.Passport) error {
	ids := make([]string, len(nodes))
	for i, p := range nodes {
		ids[i] = p.ID
		if err := v.passport(p); err != nil {
			return err
		}
	}
	facilityIDs, err := v.origins(nodes)
	if err != nil {
		return err
	}
	if err := v.massBalance(nodes); err != nil {
		return err
	}
	if err := v.certificates(append(ids, facilityIDs...)); err != nil {
		return err
	}
	if err := v.flags(ids); err != nil {
		return err
	}
	return v.names()
}

// verdict is the worst level found; each check reports its own worst.
func verdict(reasons []Reason) (string, []Check) {
	worst := map[string]string{}
	rank := map[string]int{Verified: 0, Warning: 1, Failed: 2}
	overall := Verified
	for _, r := range reasons {
		if rank[r.Level] > rank[worst[r.Check]] {
			worst[r.Check] = r.Level
		}
		if rank[r.Level] > rank[overall] {
			overall = r.Level
		}
	}
	var checks []Check
	for _, name := range Checks {
		result := Verified
		if w, ok := worst[name]; ok {
			result = w
		}
		checks = append(checks, Check{Name: name, Result: result})
	}
	return overall, checks
}
