// Package server builds the HTTP router for the API.
package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/kayigmb/chaintrack/backend/internal/attestation"
	"github.com/kayigmb/chaintrack/backend/internal/audit"
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/flag"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/notify"
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
	"github.com/kayigmb/chaintrack/backend/internal/template"
	"github.com/kayigmb/chaintrack/backend/internal/user"
	"github.com/kayigmb/chaintrack/backend/internal/verify"
)

// Health check response values.
const (
	HealthOK     = "ok"
	HealthDBDown = "db unavailable"
)

// Access says who may call a route.
type Access int

const (
	Public             Access = iota // no sign-in
	SignedIn                         // any signed-in user
	Permission                       // signed in, acting for an organisation, holding Route.Permission
	VerifiedPermission               // as Permission, and the acting organisation has passed KYB
)

// Route is one endpoint. The whole API is the list returned by Routes, so
// every endpoint's access rule is visible in one place.
type Route struct {
	Method     string
	Path       string
	Access     Access
	Permission string // set when Access is Permission or VerifiedPermission
	Handler    http.HandlerFunc
}

// Deps is everything the routes need.
type Deps struct {
	Ping        func(context.Context) error // database health; a func so it can be faked
	Auth        *auth.Auth
	Org         *org.Handlers
	Facility    *facility.Handlers
	Template    *template.Handlers
	Passport    *passport.Handlers
	Attestation *attestation.Handlers
	Flag        *flag.Handlers
	Notify      *notify.Handlers
	Verify      *verify.Handlers
	User        *user.Handlers
	// Audit records every successful change; nil turns the audit log off (unit tests).
	Audit *audit.Handlers
}

const idPath = "/{" + httpx.ParamID + "}"

func Routes(d Deps) []Route {
	a, o, f, t, p := d.Auth, d.Org, d.Facility, d.Template, d.Passport
	c, fl, n, v, u := d.Attestation, d.Flag, d.Notify, d.Verify, d.User
	const (
		get, post, put, del = http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete
	)
	return []Route{
		{get, "/healthz", Public, "", healthz(d.Ping)},
		{post, "/auth/login", Public, "", a.Login},
		{post, "/auth/logout", Public, "", a.Logout},
		{post, "/orgs", Public, "", a.LimitIP(auth.RegisterRate, o.Register)},
		{post, "/invites/{" + user.ParamToken + "}/accept", Public, "", a.LimitIP(auth.RegisterRate, u.Accept)},
		{get, "/verify" + idPath, Public, "", a.LimitIP(auth.VerifyRate, v.Verify)},

		{get, "/me", SignedIn, "", a.Me},
		{post, "/auth/2fa/setup", SignedIn, "", a.SetupTwoFactor},
		{post, "/auth/2fa/enable", SignedIn, "", a.EnableTwoFactor},
		{post, "/auth/2fa/disable", SignedIn, "", a.DisableTwoFactor},
		{post, "/auth/2fa/recovery", SignedIn, "", a.RegenerateRecoveryCodes},
		{post, "/auth/pin", SignedIn, "", a.SetPIN},
		{get, "/notifications", SignedIn, "", n.List},
		{post, "/notifications" + idPath + "/read", SignedIn, "", n.Read},

		{get, "/orgs/directory", VerifiedPermission, auth.OrgSearch, o.Directory},
		{get, "/orgs", Permission, auth.OrgRead, o.List},
		{get, "/orgs" + idPath, Permission, auth.OrgRead, o.Get},
		{post, "/orgs" + idPath + "/verify", Permission, auth.OrgVerify, o.Review},
		{get, "/orgs" + idPath + "/accreditations", Permission, auth.OrgRead, o.Accreditations},
		{put, "/orgs" + idPath + "/accreditations", Permission, auth.OrgVerify, o.SetAccreditations},
		{get, "/orgs" + idPath + "/documents", Permission, auth.DocumentRead, o.ListDocuments},
		{post, "/orgs" + idPath + "/documents", Permission, auth.OrgManage, o.Upload},
		{get, "/documents" + idPath + "/file", Permission, auth.DocumentRead, o.Download},

		{post, "/facilities", VerifiedPermission, auth.FacilityManage, f.Create},
		{get, "/facilities", Permission, auth.FacilityRead, f.List},
		{get, "/facilities" + idPath, Permission, auth.FacilityRead, f.Get},
		{post, "/facilities" + idPath + "/status", Permission, auth.FacilitySuspend, f.SetStatus},
		{del, "/facilities" + idPath, Permission, auth.FacilityManage, f.Delete},

		{post, "/templates", Permission, auth.TemplateManage, t.Create},
		{get, "/templates", Permission, auth.TemplateRead, t.List},
		{get, "/templates" + idPath, Permission, auth.TemplateRead, t.Get},
		{put, "/templates" + idPath, Permission, auth.TemplateManage, t.Update},
		{del, "/templates" + idPath, Permission, auth.TemplateManage, t.Delete},

		{post, "/passports", VerifiedPermission, auth.PassportIssue, p.Issue},
		{get, "/passports", Permission, auth.PassportRead, p.List},
		{get, "/passports" + idPath, Permission, auth.PassportRead, p.Get},
		{get, "/passports" + idPath + "/genealogy", Permission, auth.PassportRead, p.Genealogy},
		{post, "/passports" + idPath + "/retire", VerifiedPermission, auth.PassportRetire, p.Retire},
		{post, "/transforms", VerifiedPermission, auth.TransformCreate, p.Transform},
		{post, "/passports" + idPath + "/handovers", VerifiedPermission, auth.HandoverCreate, p.StartHandover},
		{get, "/handovers", Permission, auth.PassportRead, p.ListHandovers},
		{get, "/handovers" + idPath, Permission, auth.PassportRead, p.GetHandover},
		{post, "/handovers" + idPath + "/accept", VerifiedPermission, auth.HandoverAccept, p.Accept},
		{post, "/handovers" + idPath + "/reject", VerifiedPermission, auth.HandoverAccept, p.Reject},
		{post, "/handovers" + idPath + "/cancel", VerifiedPermission, auth.HandoverCreate, p.Cancel},

		{post, "/attestations", VerifiedPermission, auth.CertificateIssue, c.Issue},
		{get, "/attestations", Permission, auth.PassportRead, c.List},
		{get, "/attestations" + idPath, Permission, auth.PassportRead, c.Get},
		{get, "/attestations" + idPath + "/file", Permission, auth.PassportRead, c.Download},
		{post, "/attestations" + idPath + "/revoke", Permission, auth.CertificateRevoke, c.Revoke},

		{post, "/flags", VerifiedPermission, auth.FlagCreate, fl.Raise},
		{get, "/flags", Permission, auth.FlagCreate, fl.List},
		{get, "/flags" + idPath, Permission, auth.FlagCreate, fl.Get},
		// The handler decides who resolves: the superadmin or a verified regulator.
		{post, "/flags" + idPath + "/resolve", Permission, auth.FlagCreate, fl.Resolve},

		{get, "/users", Permission, auth.UserManage, u.List},
		{post, "/invites", Permission, auth.UserManage, u.Invite},
		{put, "/users" + idPath + "/role", Permission, auth.UserManage, u.SetRole},
		{post, "/users" + idPath + "/suspend", Permission, auth.UserManage, u.Suspend},
		{post, "/users" + idPath + "/reinstate", Permission, auth.UserManage, u.Reinstate},

		{get, "/audit", Permission, auth.AuditRead, d.Audit.List},
		{get, "/audit/verify", Permission, auth.AuditVerify, d.Audit.Verify},
		{get, "/reports/volume", Permission, auth.ReportRead, d.Audit.Volume},
	}
}

func NewRouter(d Deps) http.Handler {
	a := d.Auth
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer, auth.WithFrontend)

	for _, rt := range Routes(d) {
		var chain chi.Middlewares
		switch rt.Access {
		case SignedIn:
			chain = chi.Middlewares{a.Authenticate, a.ResolveOrg}
		case Permission:
			chain = chi.Middlewares{a.Authenticate, a.ResolveOrg, a.Require(rt.Permission)}
		case VerifiedPermission:
			chain = chi.Middlewares{a.Authenticate, a.ResolveOrg, a.Require(rt.Permission), a.RequireVerified}
		}
		// Inside the sign-in steps, so the log knows who acted and for whom.
		if d.Audit != nil && rt.Method != http.MethodGet {
			chain = append(chain, d.Audit.Record(rt.Method, rt.Path))
		}
		r.With(chain...).Method(rt.Method, rt.Path, rt.Handler)
	}
	return r
}

// healthz takes ping so the check can be tested without a database.
func healthz(ping func(context.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
		defer cancel()
		if err := ping(ctx); err != nil {
			log.Printf("healthz: %v", err)
			httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": HealthDBDown})
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": HealthOK})
	}
}
