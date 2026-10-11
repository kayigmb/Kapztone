package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/listq"
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/template"
)

const washingYield = 0.2

// coffeeTemplate is the example from PLAN.md §4.2.
func coffeeTemplate(commodity string) template.Input {
	yield := washingYield
	return template.Input{
		Commodity:          commodity,
		Unit:               "kg",
		OriginFacilityKind: facility.KindFarmPlot,
		CapacityRule:       template.CapacityRule{Type: template.CapacityYieldPerArea, KgPerHaPerSeason: 1200},
		Stages: []template.Stage{
			{Key: "origin", Events: []string{template.EventIssue}, RequiredEvidence: []string{"plot_geolocation"}},
			{Key: "washing_station", Events: []string{template.EventHandover, template.EventTransform}, MaxYield: &yield},
			{Key: "export", Events: []string{template.EventHandover, template.EventExport},
				RequiredEvidence: []string{"export_permit", "certificate_of_origin"}},
			{Key: "import", Events: []string{template.EventHandover, template.EventImport}},
		},
	}
}

func TestTemplateValidate(t *testing.T) {
	ok := coffeeTemplate("coffee")
	if msg := ok.Validate(); msg != "" {
		t.Fatalf("PLAN.md example rejected: %s", msg)
	}
	if ok.Name != ok.Commodity {
		t.Fatalf("name should default to the commodity, got %q", ok.Name)
	}
	fixed := coffeeTemplate("gold")
	fixed.CapacityRule = template.CapacityRule{Type: template.CapacityFacility}
	if msg := fixed.Validate(); msg != "" {
		t.Fatalf("facility-capacity template rejected: %s", msg)
	}

	zero, negative := 0.0, -0.5
	cases := map[string]func(*template.Input){
		"commodity with spaces":      func(in *template.Input) { in.Commodity = "Green Coffee" },
		"no unit":                    func(in *template.Input) { in.Unit = "" },
		"unknown facility kind":      func(in *template.Input) { in.OriginFacilityKind = "castle" },
		"unknown capacity rule":      func(in *template.Input) { in.CapacityRule.Type = "guess" },
		"yield rule without a rate":  func(in *template.Input) { in.CapacityRule.KgPerHaPerSeason = 0 },
		"rate on the wrong rule":     func(in *template.Input) { in.CapacityRule.Type = template.CapacityFacility },
		"no stages":                  func(in *template.Input) { in.Stages = nil },
		"first stage cannot issue":   func(in *template.Input) { in.Stages[0].Events = []string{template.EventHandover} },
		"issue in a later stage":     func(in *template.Input) { in.Stages[2].Events = append(in.Stages[2].Events, template.EventIssue) },
		"stage key twice":            func(in *template.Input) { in.Stages[3].Key = in.Stages[1].Key },
		"bad stage key":              func(in *template.Input) { in.Stages[1].Key = "Washing Station" },
		"stage without events":       func(in *template.Input) { in.Stages[3].Events = nil },
		"unknown event":              func(in *template.Input) { in.Stages[3].Events = []string{"TELEPORT"} },
		"event listed twice":         func(in *template.Input) { in.Stages[3].Events = []string{template.EventImport, template.EventImport} },
		"zero yield":                 func(in *template.Input) { in.Stages[1].MaxYield = &zero },
		"negative yield":             func(in *template.Input) { in.Stages[1].MaxYield = &negative },
		"yield without a transform":  func(in *template.Input) { in.Stages[3].MaxYield = in.Stages[1].MaxYield },
		"evidence that is not a key": func(in *template.Input) { in.Stages[0].RequiredEvidence = []string{"Plot Photo"} },
	}
	for name, breakIt := range cases {
		t.Run(name, func(t *testing.T) {
			in := coffeeTemplate("coffee")
			breakIt(&in)
			if in.Validate() == "" {
				t.Fatal("accepted")
			}
		})
	}
}

func TestTemplateCRUD(t *testing.T) {
	e, _, _ := newEnv(t, false)
	acme := e.register(tAcme, org.KindProducer)
	e.addUser(tStaff, auth.RoleMember, acme.ID)
	e.addUser(tRoot, auth.RoleSuperadmin, "")
	staff, root := e.login(tStaff), e.login(tRoot)
	acmeSlug := e.slug(tAcme)

	in := coffeeTemplate("coffee_" + e.suffix)
	in.Name = "Coffee " + e.suffix // the name carries the suffix so cleanup can find it
	body, _ := json.Marshal(in)

	rec := do(e.h, http.MethodPost, "/templates", root, string(body))
	want(t, "create", rec, http.StatusCreated)
	tpl := decodeTo[template.Template](t, rec)
	if len(tpl.Stages) != len(in.Stages) || tpl.Stages[1].MaxYield == nil || *tpl.Stages[1].MaxYield != washingYield ||
		tpl.CapacityRule != in.CapacityRule || tpl.OriginFacilityKind != in.OriginFacilityKind {
		t.Fatalf("template did not round-trip: %s", rec.Body)
	}

	want(t, "duplicate commodity", do(e.h, http.MethodPost, "/templates", root, string(body)), http.StatusConflict)
	want(t, "invalid template", do(e.h, http.MethodPost, "/templates", root, `{"commodity":"x"}`), http.StatusBadRequest)
	want(t, "member creates", do(e.h, http.MethodPost, as(acmeSlug, "/templates"), staff, string(body)), http.StatusForbidden)

	// Templates are platform-wide: any organisation's member can read them.
	list := do(e.h, http.MethodGet, as(acmeSlug, "/templates?"+listq.ParamSearch+"="+e.suffix), staff, "")
	want(t, "member lists", list, http.StatusOK)
	if p := decodeTo[listq.Page[template.Template]](t, list); p.Total != 1 || p.Data[0].ID != tpl.ID {
		t.Fatalf("member list wrong: %s", list.Body)
	}
	one := "/templates/" + tpl.ID
	want(t, "member reads one", do(e.h, http.MethodGet, as(acmeSlug, one), staff, ""), http.StatusOK)

	in.Unit = "g"
	body, _ = json.Marshal(in)
	want(t, "member updates", do(e.h, http.MethodPut, as(acmeSlug, one), staff, string(body)), http.StatusForbidden)
	rec = do(e.h, http.MethodPut, one, root, string(body))
	want(t, "update", rec, http.StatusOK)
	if got := decodeTo[template.Template](t, rec); got.Unit != in.Unit || !got.UpdatedAt.After(got.CreatedAt) {
		t.Fatalf("update not applied: %s", rec.Body)
	}

	want(t, "member deletes", do(e.h, http.MethodDelete, as(acmeSlug, one), staff, ""), http.StatusForbidden)
	want(t, "delete", do(e.h, http.MethodDelete, one, root, ""), http.StatusNoContent)
	want(t, "deleted template", do(e.h, http.MethodGet, one, root, ""), http.StatusNotFound)
	want(t, "update deleted", do(e.h, http.MethodPut, one, root, string(body)), http.StatusNotFound)
	want(t, "delete twice", do(e.h, http.MethodDelete, one, root, ""), http.StatusNotFound)
	// A deleted template frees its commodity key.
	want(t, "recreate", do(e.h, http.MethodPost, "/templates", root, string(body)), http.StatusCreated)
}
