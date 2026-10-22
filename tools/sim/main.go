// Command sim builds the evaluation dataset of PLAN.md §6.1 against a running
// ChainTrack (API, chain and relayer), injects the fraud cases and writes what
// ChainTrack made of each one to a CSV. Run it with `make sim`.
package main

import (
	"context"
	"embed"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"hash/fnv"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	password = "Sim-password-123456!" // upper, lower, number and symbol, as the API asks
	pin      = "246810"
)

// The chains, in order: each has its template file and a corner for its facilities' pins.
var chains = []struct {
	file   string
	origin [2]float64
}{
	{"coffee.json", [2]float64{29.40, -2.10}},
	{"mineral.json", [2]float64{29.20, -1.70}},
	{"medicine.json", [2]float64{30.06, -1.95}},
}

type stage struct {
	Key              string   `json:"key"`
	Events           []string `json:"events"`
	MaxYield         *float64 `json:"maxYield,omitempty"`
	RequiredEvidence []string `json:"requiredEvidence,omitempty"`
	Tolerance        *float64 `json:"handoverTolerance,omitempty"`
}

type tmpl struct {
	ID                 string         `json:"id,omitempty"`
	Commodity          string         `json:"commodity"`
	Name               string         `json:"name"`
	Unit               string         `json:"unit"`
	OriginFacilityKind string         `json:"originFacilityKind"`
	CapacityRule       map[string]any `json:"capacityRule"`
	Stages             []stage        `json:"stages"`
}

func (t tmpl) transforms() bool {
	return slices.ContainsFunc(t.Stages, func(s stage) bool { return slices.Contains(s.Events, "TRANSFORM") })
}

type org struct {
	ID, Slug, Name string
	c              *client
}

// row is one line of the CSV.
type row struct {
	Case, Chain, Kind, Mechanism, PassportID, Code string
	Status                                         int
	Verdict, Reasons, Outcome                      string
}

type sim struct {
	api, run string
	admin    *client
	db       *pgxpool.Pool
	cert     org
	rows     []row
	events   int
	certs    int
	// shift moves this run's pins a little off the grid: pins on one point overlap, and earlier runs' facilities are still there.
	shift float64
}

func main() {
	seed := flag.Uint64("seed", 42, "random seed: the same seed gives the same dataset")
	lots := flag.Int("lots", 100, "lots per chain")
	facilities := flag.Int("facilities", 20, "facilities per chain")
	perType := flag.Int("fraud", 5, "fraud cases per chain per type")
	out := flag.String("out", "sim-results.csv", "CSV of every case")
	run := flag.String("run", strconv.FormatInt(time.Now().Unix()%1e6, 36), "run id: keeps names unique across runs")
	seal := flag.Duration("seal-timeout", 15*time.Minute, "how long to wait for the relayer to seal everything")
	flag.Parse()

	api, dbURL := os.Getenv("API_URL"), os.Getenv("DATABASE_URL")
	if api == "" || dbURL == "" || os.Getenv("SUPERADMIN_EMAIL") == "" {
		log.Fatal("API_URL, DATABASE_URL and SUPERADMIN_EMAIL/PASSWORD must be set: run through `make sim`")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	admin, err := login(api, "", os.Getenv("SUPERADMIN_EMAIL"), os.Getenv("SUPERADMIN_PASSWORD"), "")
	if err != nil {
		log.Fatal(err)
	}

	r := rand.New(rand.NewPCG(*seed, 0))
	h := fnv.New32a()
	h.Write([]byte(*run))
	s := &sim{api: api, run: *run, admin: admin, db: db, shift: float64(h.Sum32()%997+1) * gridStep / 1000}
	var templates []tmpl
	var plans []ChainPlan
	for _, ch := range chains {
		t, err := readTemplate(ch.file)
		if err != nil {
			log.Fatal(err)
		}
		p, err := planChain(r, ch.origin, *facilities, *lots, *perType, t.transforms())
		if err != nil {
			log.Fatal(err)
		}
		templates, plans = append(templates, t), append(plans, p)
	}

	start := time.Now()
	for i := range templates {
		must(s.createTemplate(&templates[i]))
	}
	s.cert = must1(s.business("certifier", "Sim Certifier"))
	must(s.accredit(templates))

	type built struct {
		t     tmpl
		lots  []lot
		chain string
	}
	var all []built
	for i, t := range templates {
		log.Printf("%s: %d facilities, %d lots", t.Commodity, len(plans[i].Facilities), len(plans[i].Lots))
		lots, err := s.buildChain(t, plans[i])
		if err != nil {
			log.Fatalf("%s: %v", t.Commodity, err)
		}
		all = append(all, built{t, lots, strings.TrimSuffix(chains[i].file, ".json")})
	}

	log.Print("waiting for the relayer to seal every record")
	must(s.waitSealed(ctx, *seal))

	for _, b := range all {
		for i, l := range b.lots {
			id := fmt.Sprintf("%s-%03d", b.chain, i+1)
			pid, mech := l.final, "none"
			if l.plan.Fraud != "" {
				mech = viaDB
				pid = must1(s.tamper(ctx, l))
			}
			s.verify(row{Case: id, Chain: b.chain, Kind: kindOr(l.plan.Fraud, "legit"), Mechanism: mech, PassportID: pid, Code: l.code})
		}
	}

	must(s.writeCSV(*out))
	s.summary(time.Since(start), *out)
}

type lot struct {
	plan        LotPlan
	final, code string   // the passport the lot ends as (a transform makes a new one)
	passports   []string // every passport the lot was, oldest first
	certs       []string // certificates issued on its passports
}

// The commodity templates are data: adding a chain is adding a file here.
//
//go:embed templates/*.json
var templateFiles embed.FS

func readTemplate(file string) (tmpl, error) {
	var t tmpl
	b, err := templateFiles.ReadFile("templates/" + file)
	if err != nil {
		return t, err
	}
	return t, json.Unmarshal(b, &t)
}

// createTemplate posts a template file as it is. Its commodity gets the run id,
// since the platform keeps one live template per commodity.
func (s *sim) createTemplate(t *tmpl) error {
	t.Commodity += "_" + strings.ReplaceAll(s.run, "-", "_")
	t.Name += " · sim " + s.run
	return s.admin.json(http.MethodPost, "/templates", t, false, t)
}

// business registers an organisation, uploads its registration, has the
// platform verify it and sets its user's signing PIN.
func (s *sim) business(kind, name string) (org, error) {
	slug := fmt.Sprintf("sim-%s-%s", s.run, strings.ToLower(strings.ReplaceAll(name, " ", "-")))
	if len(slug) > 40 {
		slug = slug[:40]
	}
	slug = strings.TrimRight(slug, "-")
	email := slug + "@example.test"
	anon := &client{base: s.api}
	var o struct {
		ID         string
		VerifyLink string `json:"verify_link"` // dev only
	}
	if err := anon.json(http.MethodPost, "/orgs", map[string]string{
		"name": name + " " + s.run, "slug": slug, "kind": kind, "user_name": name + " manager", "email": email, "password": password,
	}, false, &o); err != nil {
		return org{}, fmt.Errorf("register %s: %w", slug, err)
	}
	// Spend the confirmation link as the inbox would (the API hands it back in dev).
	token := o.VerifyLink[strings.LastIndex(o.VerifyLink, "/")+1:]
	if err := anon.json(http.MethodPost, "/auth/email/verify", map[string]string{"token": token}, false, nil); err != nil {
		return org{}, fmt.Errorf("confirm email %s: %w", slug, err)
	}
	c, err := login(s.api, slug+".localhost", email, password, pin)
	if err != nil {
		return org{}, err
	}
	if err := c.form("/orgs/"+o.ID+"/documents", map[string]string{"type": "registration"}, pdf("registration "+slug), false, nil); err != nil {
		return org{}, fmt.Errorf("documents %s: %w", slug, err)
	}
	if err := s.admin.json(http.MethodPost, "/orgs/"+o.ID+"/verify", map[string]string{"decision": "verified", "reason": ""}, false, nil); err != nil {
		return org{}, fmt.Errorf("verify %s: %w", slug, err)
	}
	if err := c.json(http.MethodPost, "/auth/pin", map[string]string{"pin": pin, "password": password}, false, nil); err != nil {
		return org{}, fmt.Errorf("pin %s: %w", slug, err)
	}
	return org{ID: o.ID, Slug: slug, Name: name, c: c}, nil
}

// accredit lets the certifier issue every kind of evidence the templates ask for.
func (s *sim) accredit(ts []tmpl) error {
	var grants []map[string]string
	from, to := time.Now().AddDate(0, 0, -1).Format(time.DateOnly), time.Now().AddDate(3, 0, 0).Format(time.DateOnly)
	for _, t := range ts {
		for _, st := range t.Stages {
			for _, ev := range st.RequiredEvidence {
				grants = append(grants, map[string]string{"type": ev, "valid_from": from, "valid_to": to})
			}
		}
	}
	return s.admin.json(http.MethodPut, "/orgs/"+s.cert.ID+"/accreditations", grants, false, nil)
}

// certify has the certifier issue one piece of evidence about a passport or facility.
func (s *sim) certify(subjectType, subjectID, kind string) (string, error) {
	var a struct{ ID string }
	err := s.cert.c.form("/attestations", map[string]string{
		"subject_type": subjectType, "subject_id": subjectID, "type": kind,
		"valid_from": time.Now().AddDate(0, 0, -1).Format(time.DateOnly), "valid_to": time.Now().AddDate(2, 0, 0).Format(time.DateOnly), // from yesterday: dates are UTC midnights
	}, pdf(kind+" "+subjectID), true, &a)
	if err == nil {
		s.certs++
	}
	return a.ID, err
}

type facility struct {
	ID    string
	owner org
}

// buildChain creates a chain's businesses and facilities and walks every lot
// through the template, trying the API frauds on the way.
func (s *sim) buildChain(t tmpl, p ChainPlan) ([]lot, error) {
	name := strings.SplitN(t.Commodity, "_", 2)[0]
	var producers []org
	for i := range 4 {
		o, err := s.business("producer", fmt.Sprintf("%s producer %d", name, i+1))
		if err != nil {
			return nil, err
		}
		producers = append(producers, o)
	}
	holders := map[string]org{} // stage key → the business the lot is handed to there
	for _, st := range t.Stages[1:] {
		kind := "exporter"
		if slices.Contains(st.Events, "TRANSFORM") {
			kind = "processor"
		}
		o, err := s.business(kind, name+" "+strings.ReplaceAll(st.Key, "_", " "))
		if err != nil {
			return nil, err
		}
		holders[st.Key] = o
	}

	origin := t.Stages[0]
	newFacility := func(i int, fp FacilityPlan, approve bool) (facility, error) {
		owner := producers[i%len(producers)]
		var f struct{ ID string }
		err := owner.c.json(http.MethodPost, "/facilities", map[string]any{
			"name": fmt.Sprintf("%s site %d", name, i+1), "kind": t.OriginFacilityKind, "licence_number": fmt.Sprintf("SIM-%s-%d", s.run, i+1),
			"licence_valid_to": time.Now().AddDate(3, 0, 0).Format(time.DateOnly),
			"capacity":         fp.Capacity, "capacity_unit": t.Unit, "capacity_period": "year",
			"location": map[string]any{"type": "Point", "coordinates": []float64{fp.Lon + s.shift, fp.Lat + s.shift}},
		}, false, &f)
		if err != nil {
			return facility{}, fmt.Errorf("facility %d: %w", i+1, err)
		}
		if approve {
			if err := s.admin.json(http.MethodPost, "/facilities/"+f.ID+"/approve", nil, false, nil); err != nil {
				return facility{}, fmt.Errorf("approve facility %d: %w", i+1, err)
			}
		}
		for _, ev := range origin.RequiredEvidence {
			if _, err := s.certify("facility", f.ID, ev); err != nil {
				return facility{}, fmt.Errorf("facility %d evidence: %w", i+1, err)
			}
		}
		return facility{f.ID, owner}, nil
	}
	var facilities []facility
	for i, fp := range p.Facilities {
		f, err := newFacility(i, fp, true)
		if err != nil {
			return nil, err
		}
		facilities = append(facilities, f)
	}

	// F3: issuing from a facility the platform hasn't approved.
	if p.F3 > 0 {
		last := p.Facilities[len(p.Facilities)-1]
		pending, err := newFacility(len(p.Facilities), FacilityPlan{Lon: last.Lon + gridStep, Lat: last.Lat + gridStep, Capacity: facilityCapacity}, false)
		if err != nil {
			return nil, err
		}
		for i := range p.F3 {
			err := pending.owner.c.json(http.MethodPost, "/passports", map[string]any{"facility_id": pending.ID, "template_id": t.ID, "quantity": minQuantity}, true, nil)
			s.attempt(name, fmt.Sprintf("%s-f3-%d", name, i+1), F3, "", err)
		}
	}
	// F4 over capacity: more than a facility may originate in its period.
	for i := range p.F4Capacity {
		f := facilities[i%len(facilities)]
		err := f.owner.c.json(http.MethodPost, "/passports", map[string]any{"facility_id": f.ID, "template_id": t.ID, "quantity": facilityCapacity + 1}, true, nil)
		s.attempt(name, fmt.Sprintf("%s-f4cap-%d", name, i+1), F4, "", err)
	}

	lots := make([]lot, len(p.Lots))
	for i, lp := range p.Lots {
		l, err := s.walk(name, t, lp, facilities[lp.Facility], holders, producers)
		if err != nil {
			return nil, fmt.Errorf("lot %d: %w", i+1, err)
		}
		lots[i] = l
		if (i+1)%25 == 0 {
			log.Printf("  %s: %d lots", name, i+1)
		}
	}
	return lots, nil
}

type passportRes struct {
	ID       string
	Code     string
	Quantity int64
}

// walk issues one lot and moves it through every stage of its template.
func (s *sim) walk(name string, t tmpl, lp LotPlan, f facility, holders map[string]org, producers []org) (lot, error) {
	var p passportRes
	if err := f.owner.c.json(http.MethodPost, "/passports", map[string]any{"facility_id": f.ID, "template_id": t.ID, "quantity": lp.Quantity}, true, &p); err != nil {
		return lot{}, fmt.Errorf("issue: %w", err)
	}
	s.events++
	l := lot{plan: lp, passports: []string{p.ID}, code: p.Code}
	holder := f.owner
	for _, st := range t.Stages[1:] {
		for _, ev := range st.RequiredEvidence {
			id, err := s.certify("passport", p.ID, ev)
			if err != nil {
				return l, fmt.Errorf("evidence %s: %w", ev, err)
			}
			l.certs = append(l.certs, id)
		}
		to := holders[st.Key]
		var h struct{ ID string }
		if err := holder.c.json(http.MethodPost, "/passports/"+p.ID+"/handovers", map[string]any{"to_org": to.ID, "stage": st.Key}, true, &h); err != nil {
			return l, fmt.Errorf("hand over to %s: %w", st.Key, err)
		}
		if lp.Attempt == F2 && l.plan.Attempt != "" {
			// A business the lot isn't addressed to signs for it.
			third := producers[(slices.IndexFunc(producers, func(o org) bool { return o.ID == holder.ID })+1)%len(producers)]
			err := third.c.json(http.MethodPost, "/handovers/"+h.ID+"/accept", map[string]any{"received_quantity": p.Quantity}, true, nil)
			s.attempt(name, name+"-f2-"+l.code, F2, p.ID, err)
			l.plan.Attempt = ""
		}
		if err := to.c.json(http.MethodPost, "/handovers/"+h.ID+"/accept", map[string]any{"received_quantity": p.Quantity}, true, nil); err != nil {
			return l, fmt.Errorf("accept at %s: %w", st.Key, err)
		}
		s.events++
		holder = to

		if slices.Contains(st.Events, "TRANSFORM") && st.MaxYield != nil {
			transform := func(out int64) (passportRes, error) {
				var res struct{ Outputs []passportRes }
				err := holder.c.json(http.MethodPost, "/transforms", map[string]any{
					"stage": st.Key, "inputs": []map[string]any{{"passport_id": p.ID, "quantity": p.Quantity}}, "outputs": []map[string]any{{"quantity": out}},
				}, true, &res)
				if err != nil || len(res.Outputs) == 0 {
					return passportRes{}, err
				}
				return res.Outputs[0], nil
			}
			allowed := float64(p.Quantity) * *st.MaxYield
			if l.plan.Attempt == F4 {
				_, err := transform(int64(allowed*1.5) + 1)
				s.attempt(name, name+"-f4yield-"+l.code, F4, p.ID, err)
				l.plan.Attempt = ""
			}
			out, err := transform(max(1, int64(allowed*lp.YieldFactor)))
			if err != nil {
				return l, fmt.Errorf("transform at %s: %w", st.Key, err)
			}
			s.events++
			p = out
			l.passports = append(l.passports, p.ID)
			l.code = p.Code
		}
	}
	l.final = p.ID
	return l, nil
}

// attempt records an API fraud: refused (4xx) is prevented, accepted is missed.
func (s *sim) attempt(chain, id, kind, passportID string, err error) {
	r := row{Case: id, Chain: chain, Kind: kind, Mechanism: viaAPI, PassportID: passportID, Status: http.StatusOK}
	var ae *apiError
	switch {
	case errors.As(err, &ae):
		r.Status, r.Reasons = ae.Status, ae.Msg
	case err != nil:
		r.Status, r.Reasons = 0, err.Error()
	}
	r.Outcome = outcome(true, viaAPI, r.Status, "")
	s.rows = append(s.rows, r)
}

// waitSealed waits until the relayer has carried every queued record to the chain.
func (s *sim) waitSealed(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		var pending, failed int
		err := s.db.QueryRow(ctx, `SELECT count(*) FILTER (WHERE status = 'pending'), count(*) FILTER (WHERE status = 'failed')
		                           FROM chain_outbox WHERE NOT is_deleted`).Scan(&pending, &failed)
		if err != nil {
			return err
		}
		if pending == 0 {
			if failed > 0 {
				log.Printf("warning: %d queued records failed to seal (see chain_outbox.last_error)", failed)
			}
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%d records still waiting to be sealed after %s: is the relayer running?", pending, timeout)
		}
		log.Printf("  %d records waiting to be sealed", pending)
		time.Sleep(5 * time.Second)
	}
}

// verify asks the public verification about one passport and records the answer.
func (s *sim) verify(r row) {
	var v struct {
		Verdict string
		Reasons []struct{ Code string }
	}
	err := (&client{base: s.api}).json(http.MethodGet, "/verify/"+r.PassportID, nil, false, &v)
	var ae *apiError
	switch {
	case errors.As(err, &ae):
		r.Status, r.Reasons = ae.Status, ae.Msg
	case err != nil:
		r.Reasons = err.Error()
	default:
		r.Status, r.Verdict = http.StatusOK, v.Verdict
		var codes []string
		for _, x := range v.Reasons {
			if !slices.Contains(codes, x.Code) {
				codes = append(codes, x.Code)
			}
		}
		r.Reasons = strings.Join(codes, ";")
	}
	r.Outcome = outcome(r.Kind != "legit", r.Mechanism, r.Status, r.Verdict)
	s.rows = append(s.rows, r)
}

func (s *sim) writeCSV(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Write([]string{"case_id", "chain", "kind", "mechanism", "passport_id", "code", "http_status", "verdict", "reasons", "outcome", "baseline_outcome"})
	for _, r := range s.rows {
		w.Write([]string{r.Case, r.Chain, r.Kind, r.Mechanism, r.PassportID, r.Code, strconv.Itoa(r.Status), r.Verdict, r.Reasons, r.Outcome, ""})
	}
	w.Flush()
	return w.Error()
}

func (s *sim) summary(took time.Duration, out string) {
	type tally struct{ n, prevented, detected, missed, clean, falseAlarm int }
	by := map[string]*tally{}
	for _, r := range s.rows {
		t := by[r.Kind]
		if t == nil {
			t = &tally{}
			by[r.Kind] = t
		}
		t.n++
		switch r.Outcome {
		case prevented:
			t.prevented++
		case detected:
			t.detected++
		case missed:
			t.missed++
		case clean:
			t.clean++
		case falseAlarm:
			t.falseAlarm++
		}
	}
	fmt.Printf("\n%d events, %d certificates, %d cases in %s (%.0f events/hour) → %s\n\n",
		s.events, s.certs, len(s.rows), took.Round(time.Second), float64(s.events)/took.Hours(), out)
	fmt.Printf("%-6s %5s %9s %8s %6s\n", "kind", "cases", "prevented", "detected", "missed")
	for _, k := range []string{F1, F2, F3, F4, F5, F6} {
		if t := by[k]; t != nil {
			fmt.Printf("%-6s %5d %9d %8d %6d\n", k, t.n, t.prevented, t.detected, t.missed)
		}
	}
	if t := by["legit"]; t != nil {
		fmt.Printf("\nlegit  %d lots: %d verified, %d false alarms\n", t.n, t.clean, t.falseAlarm)
	}
}

func kindOr(k, fallback string) string {
	if k == "" {
		return fallback
	}
	return k
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func must1[T any](v T, err error) T {
	must(err)
	return v
}
