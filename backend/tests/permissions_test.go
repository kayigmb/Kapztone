package tests

import (
	"slices"
	"testing"

	"github.com/kayigmb/chaintrack/backend/internal/auth"
	"github.com/kayigmb/chaintrack/backend/internal/org"
)

// What each organisation kind may do, whatever the role (the kind table in
// README.md). The platform-only permissions (org:verify, audit:verify,
// facility:kinds, facility:suspend) appear in no kind.
var kindTable = map[string][]string{
	org.KindProducer: {auth.PassportIssue, auth.PassportRetire, auth.TransformCreate,
		auth.HandoverCreate, auth.HandoverAccept},
	org.KindProcessor:   {auth.PassportRetire, auth.TransformCreate, auth.HandoverCreate, auth.HandoverAccept},
	org.KindTransporter: {auth.HandoverCreate, auth.HandoverAccept},
	org.KindExporter:    {auth.PassportRetire, auth.TransformCreate, auth.HandoverCreate, auth.HandoverAccept},
	org.KindImporter:    {auth.PassportRetire, auth.TransformCreate, auth.HandoverCreate, auth.HandoverAccept},
	org.KindCertifier:   {auth.CertificateIssue, auth.CertificateRevoke},
	org.KindRegulator:   {auth.CertificateIssue, auth.CertificateRevoke, auth.FlagResolve},
}

// Every kind may also do these.
var everyKind = []string{auth.OrgRead, auth.OrgManage, auth.OrgSearch, auth.DocumentRead, auth.UserManage,
	auth.PassportRead, auth.FlagCreate, auth.FacilityRead, auth.FacilityManage, auth.AuditRead, auth.ReportRead}

// What a member may never do, whatever the kind.
var adminOnly = []string{auth.OrgRead, auth.OrgManage, auth.DocumentRead, auth.UserManage, auth.AuditRead,
	auth.CertificateRevoke, auth.FlagResolve, auth.FacilityManage, auth.FacilitySuspend}

func TestEffectivePermissions(t *testing.T) {
	for kind, own := range kindTable {
		allowed := append(slices.Clone(everyKind), own...)
		for _, role := range []string{auth.RoleOrgAdmin, auth.RoleMember} {
			var wantPerms []string
			for _, p := range auth.AllPermissions {
				if slices.Contains(allowed, p) && slices.Contains(auth.DefaultRoles[role], p) {
					wantPerms = append(wantPerms, p)
				}
			}
			got := auth.Effective(auth.DefaultRoles[role], kind)
			slices.Sort(got)
			slices.Sort(wantPerms)
			if !slices.Equal(got, wantPerms) {
				t.Errorf("%s %s = %v, want %v", kind, role, got, wantPerms)
			}
			if role == auth.RoleOrgAdmin && !slices.Equal(got, sorted(allowed)) {
				t.Errorf("%s admin = %v, want everything the kind allows %v", kind, got, sorted(allowed))
			}
			if role == auth.RoleMember {
				for _, p := range adminOnly {
					if slices.Contains(got, p) {
						t.Errorf("%s member holds admin-only %s", kind, p)
					}
				}
			}
		}
	}
	if got := auth.Effective(auth.DefaultRoles[auth.RoleOrgAdmin], "castle"); len(got) != 0 {
		t.Errorf("unknown kind allows %v", got)
	}
	// The ones who check the chain never trade in it.
	for _, kind := range auth.InspectorKinds {
		for _, p := range []string{auth.PassportIssue, auth.HandoverCreate, auth.HandoverAccept, auth.TransformCreate} {
			if slices.Contains(auth.Effective(auth.AllPermissions, kind), p) {
				t.Errorf("%s may %s", kind, p)
			}
		}
	}
	if len(kindTable) != len(org.Kinds) {
		t.Errorf("kind table covers %d kinds, there are %d", len(kindTable), len(org.Kinds))
	}
}

func sorted(s []string) []string {
	s = slices.Clone(s)
	slices.Sort(s)
	return s
}
