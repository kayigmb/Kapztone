// Package server builds the HTTP router for the API.
package server

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/kayigmb/chaintrack/backend/internal/attestation"
	"github.com/kayigmb/chaintrack/backend/internal/audit"
	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/facility"
	"github.com/kayigmb/chaintrack/backend/internal/flag"
	"github.com/kayigmb/chaintrack/backend/internal/httpx"
	"github.com/kayigmb/chaintrack/backend/internal/logx"
	"github.com/kayigmb/chaintrack/backend/internal/notify"
	"github.com/kayigmb/chaintrack/backend/internal/org"
	"github.com/kayigmb/chaintrack/backend/internal/passport"
	"github.com/kayigmb/chaintrack/backend/internal/stats"
	"github.com/kayigmb/chaintrack/backend/internal/template"
	"github.com/kayigmb/chaintrack/backend/internal/user"
	"github.com/kayigmb/chaintrack/backend/internal/verify"
)

// Health check response values.
const (
	HealthOK     = "ok"
	HealthDBDown = "db unavailable"
	ModeDev      = "dev"
	ModeProd     = "prod"
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
	Stats *stats.Handlers
}

const idPath = "/{" + httpx.ParamID + "}"

func Routes(d Deps) []Route {
	a, o, f, t, p := d.Auth, d.Org, d.Facility, d.Template, d.Passport
	c, fl, n, v, u := d.Attestation, d.Flag, d.Notify, d.Verify, d.User
	const (
		get, post, put, del = http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete
	)
	return []Route{
		{get, "/healthz", Public, "", healthz(d.Ping, Mode(a.Dev))},
		{post, "/auth/login", Public, "", a.Login},
		{post, "/auth/logout", Public, "", a.Logout},
		{post, "/auth/password/forgot", Public, "", a.LimitIP(auth.ForgotRate, a.ForgotPassword)},
		{post, "/auth/password/reset", Public, "", a.LimitIP(auth.RegisterRate, a.ResetPassword)},
		{get, "/auth/email/availability", Public, "", a.LimitIP(auth.RegisterRate, o.EmailAvailability)},
		{post, "/auth/lock", Public, "", a.LimitIP(auth.RegisterRate, a.LockAccount)},
		{post, "/auth/email/verify", Public, "", a.LimitIP(auth.RegisterRate, a.VerifyEmail)},
		{post, "/auth/handoff/accept", Public, "", a.LimitIP(auth.RegisterRate, a.AcceptHandoff)},
		{post, "/orgs", Public, "", a.LimitIP(auth.RegisterRate, o.Register)},
		{post, "/invites/{" + user.ParamToken + "}/accept", Public, "", a.LimitIP(auth.RegisterRate, u.Accept)},
		{get, "/verify" + idPath, Public, "", a.LimitIP(auth.VerifyRate, v.Verify)},
		// Caddy asks before getting a certificate; not rate limited, all its calls share one address.
		{get, "/tls/allowed", Public, "", o.TLSAllowed},

		{get, "/me", SignedIn, "", a.Me},
		{put, "/me", SignedIn, "", u.SetName},
		{put, "/me/notifications", SignedIn, "", u.SetEmailNotifications},
		{get, "/me/organisations", SignedIn, "", u.MyOrganisations},
		{post, "/auth/password", SignedIn, "", a.ChangePassword},
		{post, "/auth/sessions/revoke", SignedIn, "", a.RevokeSessions},
		{post, "/auth/handoff", SignedIn, "", a.StartHandoff},
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
		{get, "/orgs" + idPath + "/members", Permission, auth.OrgRead, u.OrgMembers},
		{post, "/orgs" + idPath + "/suspend", Permission, auth.OrgVerify, o.Suspend},
		{post, "/orgs" + idPath + "/reactivate", Permission, auth.OrgVerify, o.Reactivate},
		{put, "/orgs" + idPath + "/accreditations", Permission, auth.OrgVerify, o.SetAccreditations},
		{get, "/orgs" + idPath + "/documents", Permission, auth.DocumentRead, o.ListDocuments},
		{post, "/orgs" + idPath + "/documents", Permission, auth.OrgManage, o.Upload},
		{get, "/documents" + idPath + "/file", Permission, auth.DocumentRead, o.Download},

		{post, "/facilities", VerifiedPermission, auth.FacilityManage, f.Create},
		{get, "/facilities", Permission, auth.FacilityRead, f.List},
		{get, "/facilities" + idPath, Permission, auth.FacilityRead, f.Get},
		{post, "/facilities" + idPath + "/status", Permission, auth.FacilitySuspend, f.SetStatus},
		{post, "/facilities" + idPath + "/approve", Permission, auth.FacilitySuspend, f.Approve},
		{post, "/facilities" + idPath + "/licence", Permission, auth.FacilityManage, f.UploadLicence},
		{get, "/facilities" + idPath + "/licence", Permission, auth.FacilityRead, f.Licence},
		{del, "/facilities" + idPath, Permission, auth.FacilityManage, f.Delete},

		{get, "/facility-kinds", Permission, auth.FacilityRead, f.Kinds},
		{get, "/facility-kinds/all", Permission, auth.TemplateManage, f.AllKinds},
		{post, "/facility-kinds", Permission, auth.TemplateManage, f.CreateKind},
		{put, "/facility-kinds/{key}", Permission, auth.TemplateManage, f.UpdateKind},
		{del, "/facility-kinds/{key}", Permission, auth.TemplateManage, f.DeleteKind},

		{post, "/templates", Permission, auth.TemplateManage, t.Create},
		{get, "/templates", Permission, auth.TemplateRead, t.List},
		{get, "/templates" + idPath, Permission, auth.TemplateRead, t.Get},
		{put, "/templates" + idPath, Permission, auth.TemplateManage, t.Update},
		{post, "/templates" + idPath + "/publish", Permission, auth.TemplateManage, t.Publish},
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
		{get, "/handovers" + idPath + "/photos/{" + passport.ParamPhotoID + "}", Permission, auth.PassportRead, p.HandoverPhoto},
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
		// Likewise the handler decides who assigns: an investigator.
		{put, "/flags" + idPath + "/assignee", Permission, auth.FlagCreate, fl.Assign},

		{get, "/users", Permission, auth.UserManage, u.List},
		{post, "/invites", Permission, auth.UserManage, u.Invite},
		{put, "/users" + idPath + "/role", Permission, auth.UserManage, u.SetRole},
		{post, "/users" + idPath + "/suspend", Permission, auth.UserManage, u.Suspend},
		{post, "/users" + idPath + "/reinstate", Permission, auth.UserManage, u.Reinstate},
		// The platform admin only; the handlers refuse an organisation's admin.
		{get, "/users" + idPath, Permission, auth.UserManage, u.Get},
		{post, "/users" + idPath + "/deactivate", Permission, auth.UserManage, u.Deactivate},
		{post, "/users" + idPath + "/activate", Permission, auth.UserManage, u.Activate},

		{get, "/audit", Permission, auth.AuditRead, d.Audit.List},
		{get, "/audit/verify", Permission, auth.AuditVerify, d.Audit.Verify},
		{get, "/reports/volume", Permission, auth.ReportRead, d.Audit.Volume},
		{get, "/reports/blocked", Permission, auth.ReportRead, d.Audit.Blocked},
		{get, "/reports/verification-outcomes", Permission, auth.ReportRead, d.Audit.Outcomes},
		{get, "/reports/saved", Permission, auth.ReportRead, d.Audit.ListSaved},
		{post, "/reports/saved", Permission, auth.ReportRead, d.Audit.CreateSaved},
		{post, "/reports/saved" + idPath + "/run", Permission, auth.ReportRead, d.Audit.RunSaved},
		{get, "/reports/saved" + idPath + "/files", Permission, auth.ReportRead, d.Audit.ListFiles},
		{get, "/reports/files" + idPath, Permission, auth.ReportRead, d.Audit.DownloadFile},
		{get, "/stats", Permission, auth.AuditVerify, d.Stats.Get},
	}
}

func NewRouter(d Deps) http.Handler {
	a := d.Auth
	r := chi.NewRouter()
	r.Use(logx.Middleware(auth.ClientIP))

	for _, rt := range Routes(d) {
		var chain chi.Middlewares
		switch rt.Access {
		case Public:
			if rt.Method != http.MethodGet {
				chain = chi.Middlewares{auth.SameOrigin}
			}
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
// Mode names how the API runs: "dev" (no rate limits, dev-only conveniences) or "prod".
func Mode(dev bool) string {
	if dev {
		return ModeDev
	}
	return ModeProd
}

func healthz(ping func(context.Context) error, mode string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
		defer cancel()
		if err := ping(ctx); err != nil {
			logx.From(req.Context()).Error("health check failed", logx.KeyError, err)
			httpx.OK(w, http.StatusServiceUnavailable, "", map[string]string{"status": HealthDBDown, "mode": mode})
			return
		}
		httpx.OK(w, http.StatusOK, "", map[string]string{"status": HealthOK, "mode": mode})
	}
}
