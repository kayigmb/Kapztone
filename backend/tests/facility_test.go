package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/upload"
)

// geo builds a GeoJSON geometry the way a client would send it.
func geo(kind, coordinates string) string {
	return fmt.Sprintf(`{"type":%q,"coordinates":%s}`, kind, coordinates)
}

var (
	// A square of 0.001° near the equator: about 111 m a side, roughly 1.2 ha.
	plotGeo = geo(facility.GeoPolygon, `[[[30.000,-2.000],[30.001,-2.000],[30.001,-1.999],[30.000,-1.999],[30.000,-2.000]]]`)
	// A boundary that crosses itself (a bow-tie).
	bowTieGeo = geo(facility.GeoPolygon, `[[[30,-2],[30.001,-1.999],[30.001,-2],[30,-1.999],[30,-2]]]`)
	pinGeo    = geo(facility.GeoPoint, `[36.8,-1.3]`)
	// Latitude 120 does not exist.
	offMapGeo = geo(facility.GeoPoint, `[30,120]`)
)

const plotBox = "29.99,-2.01,30.01,-1.99"

// Facilities may not overlap any other live one, in any test or in the
// developer's own data, so tests that store facilities draw each one at a
// place of its own: a 0.001° square in the empty South Pacific.
var (
	plotBase = [2]float64{-125 + rand.Float64()*10, -3 + rand.Float64()*2}
	plotNext atomic.Int64
)

// uniquePlot returns a fresh plot's GeoJSON and a map window around it.
func uniquePlot() (geoJSON, box string) {
	n := float64(plotNext.Add(1))
	lon, lat := plotBase[0]+n*0.002, plotBase[1]
	geoJSON = geo(facility.GeoPolygon, fmt.Sprintf(`[[[%[1]f,%[2]f],[%[3]f,%[2]f],[%[3]f,%[4]f],[%[1]f,%[4]f],[%[1]f,%[2]f]]]`,
		lon, lat, lon+0.001, lat+0.001))
	return geoJSON, fmt.Sprintf("%f,%f,%f,%f", lon-0.0005, lat-0.0005, lon+0.0015, lat+0.0015)
}

// newPin is a point of its own, below the row of plots.
func newPin() string {
	n := float64(plotNext.Add(1))
	return geo(facility.GeoPoint, fmt.Sprintf(`[%f,%f]`, plotBase[0]+n*0.002, plotBase[1]-0.01))
}

// newPlot is uniquePlot without the window.
func newPlot() string { g, _ := uniquePlot(); return g }

// approvedFacility registers a facility and has the platform approve it.
func (e *env) approvedFacility(token, slug, root string, in facility.Input) facility.Facility {
	e.t.Helper()
	rec := post(e, token, as(slug, "/facilities"), in)
	want(e.t, "create facility", rec, http.StatusCreated)
	f := decodeTo[facility.Facility](e.t, rec)
	rec = post(e, root, "/facilities/"+f.ID+"/approve", nil)
	want(e.t, "approve facility", rec, http.StatusOK)
	return decodeTo[facility.Facility](e.t, rec)
}

func facilityInput(location string) facility.Input {
	return facility.Input{
		Name: "North plot", Kind: facility.KindFarmPlot, Location: json.RawMessage(location),
		LicenceNumber: "LIC-001", Capacity: 1_200_000, CapacityUnit: "g", CapacityPeriod: facility.PeriodSeason,
		LicenceValidTo: "2036-12-31",
	}
}

func TestFacilityValidate(t *testing.T) {
	for _, geo := range []string{plotGeo, pinGeo} {
		in := facilityInput(geo)
		if msg := in.Validate(); msg != "" {
			t.Fatalf("valid facility rejected: %s", msg)
		}
	}
	custom := facilityInput(pinGeo)
	custom.Kind = "tea_estate"
	if msg := custom.Validate(); msg != "" {
		t.Fatalf("well-formed custom kind rejected before curated-kind lookup: %s", msg)
	}

	cases := map[string]func(*facility.Input){
		"empty name":         func(in *facility.Input) { in.Name = " " },
		"malformed kind key": func(in *facility.Input) { in.Kind = "Castle Kind" },
		"no licence":         func(in *facility.Input) { in.LicenceNumber = "" },
		"zero capacity":      func(in *facility.Input) { in.Capacity = 0 },
		"negative capacity":  func(in *facility.Input) { in.Capacity = -5 },
		"no unit":            func(in *facility.Input) { in.CapacityUnit = "" },
		"long unit":          func(in *facility.Input) { in.CapacityUnit = strings.Repeat("k", facility.MaxUnitLen+1) },
		"unknown period":     func(in *facility.Input) { in.CapacityPeriod = "decade" },
		"no location":        func(in *facility.Input) { in.Location = nil },
		"unsupported shape": func(in *facility.Input) {
			in.Location = json.RawMessage(`{"type":"LineString","coordinates":[[30,-2],[31,-2]]}`)
		},
		"latitude out of range": func(in *facility.Input) {
			in.Location = json.RawMessage(offMapGeo)
		},
		"longitude out of range": func(in *facility.Input) {
			in.Location = json.RawMessage(geo(facility.GeoPoint, `[181,0]`))
		},
		"point with one number": func(in *facility.Input) {
			in.Location = json.RawMessage(geo(facility.GeoPoint, `[30]`))
		},
		"boundary too short": func(in *facility.Input) {
			in.Location = json.RawMessage(geo(facility.GeoPolygon, `[[[30,-2],[30.001,-2],[30,-2]]]`))
		},
		"boundary not closed": func(in *facility.Input) {
			in.Location = json.RawMessage(geo(facility.GeoPolygon, `[[[30,-2],[30.001,-2],[30.001,-1.999],[30,-1.999]]]`))
		},
		"coordinates of the wrong shape": func(in *facility.Input) {
			in.Location = json.RawMessage(geo(facility.GeoPolygon, `[30,-2]`))
		},
	}
	for name, breakIt := range cases {
		t.Run(name, func(t *testing.T) {
			in := facilityInput(plotGeo)
			breakIt(&in)
			if in.Validate() == "" {
				t.Fatal("accepted")
			}
		})
	}
}

func TestFacilityKindsManaged(t *testing.T) {
	e, _, _ := newEnv(t, false)
	e.addUser(tRoot, auth.RoleSuperadmin, "")
	root := e.login(tRoot)
	e.verifiedOrg(tAcme, org.KindProducer, root)
	slug, member := e.slug(tAcme), e.login(tAcme)
	kindKey := "tea_estate_" + e.suffix

	created := post(e, root, "/facility-kinds", facility.Kind{Key: kindKey, Label: "Tea estate", Sub: "Tea growing land", Active: true})
	want(t, "create kind", created, http.StatusCreated)
	if got := decodeTo[facility.Kind](t, created); got.Key != kindKey || got.Label != "Tea estate" || !got.Active {
		t.Fatalf("created kind = %+v", got)
	}
	want(t, "duplicate kind", post(e, root, "/facility-kinds", facility.Kind{Key: kindKey, Label: "Again", Active: true}), http.StatusConflict)
	want(t, "member cannot curate kinds", post(e, member, "/facility-kinds", facility.Kind{Key: "another_kind", Label: "Another", Active: true}), http.StatusForbidden)

	active := do(e.h, http.MethodGet, as(slug, "/facility-kinds"), member, "")
	want(t, "list active kinds", active, http.StatusOK)
	activeRows := decodeTo[struct {
		Data []facility.Kind `json:"data"`
	}](t, active).Data
	if !slices.ContainsFunc(activeRows, func(k facility.Kind) bool { return k.Key == kindKey }) {
		t.Fatalf("new kind absent from active list: %+v", activeRows)
	}

	in := facilityInput(pinGeo)
	in.Kind = kindKey
	facRec := post(e, member, as(slug, "/facilities"), in)
	want(t, "create facility with custom kind", facRec, http.StatusCreated)
	_ = decodeTo[facility.Facility](t, facRec)

	tpl := coffeeTemplate("tea_" + e.suffix)
	tpl.OriginFacilityKind = kindKey
	body, _ := json.Marshal(tpl)
	tplRec := do(e.h, http.MethodPost, "/templates", root, string(body))
	want(t, "create template with custom origin kind", tplRec, http.StatusCreated)

	updated := do(e.h, http.MethodPut, "/facility-kinds/"+kindKey, root,
		`{"label":"Tea farm","sub":"Leaf production","active":false}`)
	want(t, "deactivate kind", updated, http.StatusOK)
	deactivated := decodeTo[facility.Kind](t, updated)
	if deactivated.Active || deactivated.Label != "Tea farm" {
		t.Fatalf("updated kind = %+v", deactivated)
	}
	active = do(e.h, http.MethodGet, as(slug, "/facility-kinds"), member, "")
	want(t, "inactive kinds hidden from forms", active, http.StatusOK)
	activeRows = decodeTo[struct {
		Data []facility.Kind `json:"data"`
	}](t, active).Data
	if slices.ContainsFunc(activeRows, func(k facility.Kind) bool { return k.Key == kindKey }) {
		t.Fatalf("deactivated kind still offered: %+v", activeRows)
	}
	all := do(e.h, http.MethodGet, "/facility-kinds/all", root, "")
	want(t, "platform sees inactive kind", all, http.StatusOK)
	allRows := decodeTo[struct {
		Data []facility.Kind `json:"data"`
	}](t, all).Data
	if !slices.ContainsFunc(allRows, func(k facility.Kind) bool { return k.Key == kindKey && !k.Active }) {
		t.Fatalf("inactive kind absent from management list: %+v", allRows)
	}

	// Historical facilities remain filterable by a kind removed from new forms.
	filtered := do(e.h, http.MethodGet, as(slug, "/facilities?kind="+kindKey+"&size=100"), member, "")
	want(t, "filter historical inactive kind", filtered, http.StatusOK)
	if got := decodeTo[listq.Page[facility.Facility]](t, filtered).Total; got != 1 {
		t.Fatalf("custom-kind facilities filtered = %d, want 1", got)
	}

	in.Name += " 2"
	facRec = post(e, member, as(slug, "/facilities"), in)
	want(t, "inactive kind refused for new facility", facRec, http.StatusBadRequest)
	tpl.Commodity += "_draft"
	body, _ = json.Marshal(tpl)
	tplRec = do(e.h, http.MethodPost, "/templates", root, string(body))
	want(t, "inactive kind refused for new template", tplRec, http.StatusBadRequest)
	if !strings.Contains(tplRec.Body.String(), "originFacilityKind") {
		t.Fatalf("inactive-kind template error = %s", tplRec.Body)
	}

	want(t, "used kind cannot be deleted", do(e.h, http.MethodDelete, "/facility-kinds/"+kindKey, root, ""), http.StatusConflict)
	unusedKey := "unused_kind_" + e.suffix
	want(t, "create removable kind", post(e, root, "/facility-kinds", facility.Kind{Key: unusedKey, Label: "Unused", Active: true}), http.StatusCreated)
	want(t, "delete unused kind", do(e.h, http.MethodDelete, "/facility-kinds/"+unusedKey, root, ""), http.StatusOK)
}

func TestParseBBox(t *testing.T) {
	if b, ok := facility.ParseBBox(""); b != nil || !ok {
		t.Fatal("empty window should mean no window")
	}
	if b, ok := facility.ParseBBox(plotBox); !ok || b == nil || b[0] != 29.99 || b[3] != -1.99 {
		t.Fatalf("valid window rejected: %v %v", b, ok)
	}
	for _, bad := range []string{"1,2,3", "a,b,c,d", "30,-2,29,-1", "30,-2,31,-3", "0,0,200,10"} {
		if _, ok := facility.ParseBBox(bad); ok {
			t.Errorf("%q accepted", bad)
		}
	}
}

func (e *env) setOrgStatus(orgID, status string) {
	e.exec(`UPDATE organisations SET status = $2 WHERE id = $1::uuid`, orgID, status)
}

func TestFacilityFlow(t *testing.T) {
	// No cache: organisation status is set directly in the database here.
	e, _, _ := newEnv(t, false)
	acme := e.register(tAcme, org.KindProducer)
	rival := e.register(tRival, org.KindExporter)
	e.addUser(tStaff, auth.RoleMember, acme.ID)
	e.addUser(tRoot, auth.RoleSuperadmin, "")
	admin, rivalAdmin, staff, root := e.login(tAcme), e.login(tRival), e.login(tStaff), e.login(tRoot)
	acmeSlug, rivalSlug := e.slug(tAcme), e.slug(tRival)

	create := func(token, slug, location string) *httptest.ResponseRecorder {
		body, _ := json.Marshal(facilityInput(location))
		return do(e.h, http.MethodPost, as(slug, "/facilities"), token, string(body))
	}
	total := func(token, path string) int {
		t.Helper()
		rec := do(e.h, http.MethodGet, path, token, "")
		want(t, "list "+path, rec, http.StatusOK)
		return decodeTo[listq.Page[facility.Facility]](t, rec).Total
	}

	plotAt, plotWindow := uniquePlot()

	// An organisation that has not passed KYB cannot register facilities.
	want(t, "pending organisation", create(admin, acmeSlug, plotAt), http.StatusForbidden)
	e.setOrgStatus(acme.ID, org.StatusVerified)
	e.setOrgStatus(rival.ID, org.StatusVerified)

	rec := create(admin, acmeSlug, plotAt)
	want(t, "create plot", rec, http.StatusCreated)
	plot := decodeTo[facility.Facility](t, rec)
	if plot.Status != facility.StatusPending || plot.OrganisationID != acme.ID || plot.LicenceValidTo == nil ||
		*plot.LicenceValidTo != "2036-12-31" || plot.ApprovedAt != nil {
		t.Fatalf("facility recorded wrongly: %+v", plot)
	}
	if statuses, _ := e.outbox(plot.ID); len(statuses) != 0 {
		t.Fatal("a facility waiting for approval was queued for the ledger")
	}

	// The same land cannot be registered twice: the caller is told which of its
	// facilities is in the way, another organisation only that one is.
	rec = create(admin, acmeSlug, plotAt)
	want(t, "overlapping own facility", rec, http.StatusConflict)
	if got := decodeTo[struct {
		FacilityID *string `json:"facility_id"`
	}](t, rec); errorOf(t, rec) != facility.MsgOverlaps(plot.Name) || got.FacilityID == nil || *got.FacilityID != plot.ID {
		t.Fatalf("overlap answer: %s", rec.Body)
	}
	rec = create(rivalAdmin, rivalSlug, plotAt)
	want(t, "overlapping another organisation's facility", rec, http.StatusConflict)
	if strings.Contains(rec.Body.String(), plot.ID) || strings.Contains(rec.Body.String(), plot.Name) {
		t.Fatalf("overlap names another organisation's facility: %s", rec.Body)
	}

	// Waiting for approval: its status cannot be set, only the platform approves it.
	suspend := fmt.Sprintf(`{"status":%q}`, facility.StatusSuspended)
	one := "/facilities/" + plot.ID
	want(t, "suspend while pending", do(e.h, http.MethodPost, one+"/status", root, suspend), http.StatusConflict)
	want(t, "organisation approves its own", post(e, admin, as(acmeSlug, one+"/approve"), nil), http.StatusForbidden)
	rec = post(e, root, one+"/approve", nil)
	want(t, "approve", rec, http.StatusOK)
	if f := decodeTo[facility.Facility](t, rec); f.Status != facility.StatusActive || f.ApprovedAt == nil {
		t.Fatalf("approved facility: %+v", f)
	}
	if statuses, _ := e.outbox(plot.ID); len(statuses) != 1 {
		t.Fatalf("approval queued %d items for the ledger, want 1", len(statuses))
	}
	want(t, "approve twice", post(e, root, one+"/approve", nil), http.StatusConflict)
	want(t, "approve unknown", post(e, root, "/facilities/"+fakeOrgID+"/approve", nil), http.StatusNotFound)

	// The licence document: stored in the bucket, read back by the owner only.
	want(t, "no licence yet", do(e.h, http.MethodGet, as(acmeSlug, one+"/licence"), admin, ""), http.StatusNotFound)
	licence := func(token, slug, validTo string, content []byte) *httptest.ResponseRecorder {
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		if validTo != "" {
			mw.WriteField(facility.FormLicenceValidTo, validTo)
		}
		fw, _ := mw.CreateFormFile(upload.FormFile, "licence.pdf")
		fw.Write(content)
		mw.Close()
		req := httptest.NewRequest(http.MethodPost, as(slug, one+"/licence"), &body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		return send(e.h, req, token)
	}
	want(t, "licence with a bad date", licence(admin, acmeSlug, "next year", pdfBytes), http.StatusBadRequest)
	want(t, "licence of another organisation's facility", licence(rivalAdmin, rivalSlug, "", pdfBytes), http.StatusNotFound)
	rec = licence(admin, acmeSlug, "2037-06-30", pdfBytes)
	want(t, "upload licence", rec, http.StatusOK)
	if f := decodeTo[facility.Facility](t, rec); f.LicenceDocument == nil || *f.LicenceValidTo != "2037-06-30" ||
		f.LicenceDocument.SizeBytes != int64(len(pdfBytes)) || strings.Contains(rec.Body.String(), "storage_key") {
		t.Fatalf("licence recorded wrongly: %s", rec.Body)
	}
	rec = do(e.h, http.MethodGet, as(acmeSlug, one+"/licence"), admin, "")
	want(t, "download licence", rec, http.StatusOK)
	if !bytes.Equal(rec.Body.Bytes(), pdfBytes) {
		t.Fatal("licence downloaded differs from the upload")
	}
	want(t, "another organisation downloads the licence", do(e.h, http.MethodGet, as(rivalSlug, one+"/licence"), rivalAdmin, ""), http.StatusNotFound)
	want(t, "replace licence", licence(admin, acmeSlug, "", pdfBytes), http.StatusOK)
	// Area comes from the boundary, measured on the globe.
	if plot.AreaHa < 1.1 || plot.AreaHa > 1.4 {
		t.Fatalf("area = %.3f ha, want about 1.2", plot.AreaHa)
	}
	var geo struct{ Type string }
	if json.Unmarshal(plot.Location, &geo); geo.Type != facility.GeoPolygon {
		t.Fatalf("location not returned as GeoJSON: %s", plot.Location)
	}

	want(t, "self-crossing boundary", create(admin, acmeSlug, bowTieGeo), http.StatusBadRequest)
	want(t, "latitude 120", create(admin, acmeSlug, offMapGeo), http.StatusBadRequest)
	want(t, "member creates", create(staff, acmeSlug, newPlot()), http.StatusForbidden)
	noDate := facilityInput(newPlot())
	noDate.LicenceValidTo = ""
	want(t, "no licence end date", post(e, admin, as(acmeSlug, "/facilities"), noDate), http.StatusBadRequest)

	rec = create(admin, acmeSlug, newPin())
	want(t, "create pin", rec, http.StatusCreated)
	if pin := decodeTo[facility.Facility](t, rec); pin.AreaHa != 0 {
		t.Fatalf("a pin has area %.3f", pin.AreaHa)
	}

	// Lists: the map window, the gate, and who may read.
	list := as(acmeSlug, "/facilities")
	if n := total(admin, list); n != 2 {
		t.Fatalf("organisation sees %d facilities, want 2", n)
	}
	if n := total(staff, list); n != 2 {
		t.Fatalf("member sees %d facilities, want 2", n)
	}
	if n := total(admin, list+"&"+facility.ParamBBox+"="+plotWindow); n != 1 {
		t.Fatalf("map window returned %d facilities, want 1", n)
	}
	want(t, "bad map window", do(e.h, http.MethodGet, list+"&"+facility.ParamBBox+"=1,2,3", admin, ""), http.StatusBadRequest)
	want(t, "bad status filter", do(e.h, http.MethodGet, list+"&"+facility.ParamStatus+"=closed", admin, ""), http.StatusBadRequest)
	if n := total(rivalAdmin, as(rivalSlug, "/facilities")); n != 0 {
		t.Fatalf("another organisation sees %d of these facilities", n)
	}
	want(t, "own facility", do(e.h, http.MethodGet, as(acmeSlug, one), admin, ""), http.StatusOK)
	want(t, "another organisation's facility", do(e.h, http.MethodGet, as(rivalSlug, one), rivalAdmin, ""), http.StatusNotFound)

	// Only the platform can suspend; a business cannot change its own status.
	want(t, "organisation suspends itself", do(e.h, http.MethodPost, as(acmeSlug, one+"/status"), admin, suspend), http.StatusForbidden)
	want(t, "unknown status", do(e.h, http.MethodPost, one+"/status", root, `{"status":"closed"}`), http.StatusBadRequest)
	rec = do(e.h, http.MethodPost, one+"/status", root, suspend)
	want(t, "platform suspends", rec, http.StatusOK)
	if f := decodeTo[facility.Facility](t, rec); f.Status != facility.StatusSuspended {
		t.Fatalf("status = %s", f.Status)
	}
	if n := total(admin, list+"&"+facility.ParamStatus+"="+facility.StatusSuspended); n != 1 {
		t.Fatalf("suspended filter returned %d, want 1", n)
	}

	// Delete hides the facility but keeps the row.
	want(t, "another organisation deletes", do(e.h, http.MethodDelete, as(rivalSlug, one), rivalAdmin, ""), http.StatusNotFound)
	want(t, "delete", do(e.h, http.MethodDelete, as(acmeSlug, one), admin, ""), http.StatusOK)
	want(t, "deleted facility", do(e.h, http.MethodGet, as(acmeSlug, one), admin, ""), http.StatusNotFound)
	want(t, "delete twice", do(e.h, http.MethodDelete, as(acmeSlug, one), admin, ""), http.StatusNotFound)
	if n := total(admin, list); n != 1 {
		t.Fatalf("after delete the list has %d, want 1", n)
	}
	var kept bool
	e.pool.QueryRow(t.Context(), `SELECT is_deleted FROM facilities WHERE id = $1::uuid`, plot.ID).Scan(&kept)
	if !kept {
		t.Fatal("deleted facility row is gone or not marked deleted")
	}
}
