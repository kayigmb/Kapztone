package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/org"
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

func facilityInput(location string) facility.Input {
	return facility.Input{
		Name: "North plot", Kind: facility.KindFarmPlot, Location: json.RawMessage(location),
		LicenceNumber: "LIC-001", Capacity: 1_200_000, CapacityUnit: "g", CapacityPeriod: facility.PeriodSeason,
	}
}

func TestFacilityValidate(t *testing.T) {
	for _, geo := range []string{plotGeo, pinGeo} {
		in := facilityInput(geo)
		if msg := in.Validate(); msg != "" {
			t.Fatalf("valid facility rejected: %s", msg)
		}
	}

	cases := map[string]func(*facility.Input){
		"empty name":        func(in *facility.Input) { in.Name = " " },
		"unknown kind":      func(in *facility.Input) { in.Kind = "castle" },
		"no licence":        func(in *facility.Input) { in.LicenceNumber = "" },
		"zero capacity":     func(in *facility.Input) { in.Capacity = 0 },
		"negative capacity": func(in *facility.Input) { in.Capacity = -5 },
		"no unit":           func(in *facility.Input) { in.CapacityUnit = "" },
		"long unit":         func(in *facility.Input) { in.CapacityUnit = strings.Repeat("k", facility.MaxUnitLen+1) },
		"unknown period":    func(in *facility.Input) { in.CapacityPeriod = "decade" },
		"no location":       func(in *facility.Input) { in.Location = nil },
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

	// An organisation that has not passed KYB cannot register facilities.
	want(t, "pending organisation", create(admin, acmeSlug, plotGeo), http.StatusForbidden)
	e.setOrgStatus(acme.ID, org.StatusVerified)
	e.setOrgStatus(rival.ID, org.StatusVerified)

	rec := create(admin, acmeSlug, plotGeo)
	want(t, "create plot", rec, http.StatusCreated)
	plot := decodeTo[facility.Facility](t, rec)
	if plot.Status != facility.StatusActive || plot.OrganisationID != acme.ID {
		t.Fatalf("facility recorded wrongly: %+v", plot)
	}
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
	want(t, "member creates", create(staff, acmeSlug, plotGeo), http.StatusForbidden)

	rec = create(admin, acmeSlug, pinGeo)
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
	if n := total(admin, list+"&"+facility.ParamBBox+"="+plotBox); n != 1 {
		t.Fatalf("map window returned %d facilities, want 1", n)
	}
	want(t, "bad map window", do(e.h, http.MethodGet, list+"&"+facility.ParamBBox+"=1,2,3", admin, ""), http.StatusBadRequest)
	want(t, "bad status filter", do(e.h, http.MethodGet, list+"&"+facility.ParamStatus+"=closed", admin, ""), http.StatusBadRequest)
	if n := total(rivalAdmin, as(rivalSlug, "/facilities")); n != 0 {
		t.Fatalf("another organisation sees %d of these facilities", n)
	}
	one := "/facilities/" + plot.ID
	want(t, "own facility", do(e.h, http.MethodGet, as(acmeSlug, one), admin, ""), http.StatusOK)
	want(t, "another organisation's facility", do(e.h, http.MethodGet, as(rivalSlug, one), rivalAdmin, ""), http.StatusNotFound)

	// Only the platform can suspend; a business cannot change its own status.
	suspend := fmt.Sprintf(`{"status":%q}`, facility.StatusSuspended)
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
	want(t, "delete", do(e.h, http.MethodDelete, as(acmeSlug, one), admin, ""), http.StatusNoContent)
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
