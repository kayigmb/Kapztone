package auth

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Permissions are "resource:action". This file is their single reference
// point: routes use these names with Require, and SyncRoles writes them to the
// database. To add one, declare it, list it in AllPermissions, and give it to
// the roles that need it.
const (
	OrgRead           = "org:read"
	OrgManage         = "org:manage"
	OrgVerify         = "org:verify"
	OrgSearch         = "org:search" // find other verified organisations (handover receivers)
	DocumentRead      = "document:read"
	UserManage        = "user:manage"
	PassportIssue     = "passport:issue"
	PassportRead      = "passport:read"
	PassportRetire    = "passport:retire"
	HandoverCreate    = "handover:create"
	HandoverAccept    = "handover:accept"
	TransformCreate   = "transform:create"
	CertificateIssue  = "certificate:issue"
	CertificateRevoke = "certificate:revoke"
	FlagCreate        = "flag:create"
	FlagResolve       = "flag:resolve"
	TemplateRead      = "template:read"
	TemplateManage    = "template:manage"
	FacilityRead      = "facility:read"
	FacilityManage    = "facility:manage"
	FacilitySuspend   = "facility:suspend"
)

// Audit log and reports.
const (
	AuditRead   = "audit:read"
	AuditVerify = "audit:verify" // walk the whole chain: platform only
	ReportRead  = "report:read"
)

var AllPermissions = []string{
	OrgRead, OrgManage, OrgVerify, OrgSearch, DocumentRead, UserManage,
	PassportIssue, PassportRead, PassportRetire, HandoverCreate, HandoverAccept, TransformCreate,
	CertificateIssue, CertificateRevoke, FlagCreate, FlagResolve, TemplateRead, TemplateManage,
	FacilityRead, FacilityManage, FacilitySuspend,
	AuditRead, AuditVerify, ReportRead,
}

// DefaultRoles is what each built-in role may do.
var DefaultRoles = map[string][]string{
	RoleSuperadmin: AllPermissions,
	RoleOrgAdmin: {
		OrgRead, OrgManage, OrgSearch, DocumentRead, UserManage,
		PassportIssue, PassportRead, PassportRetire, HandoverCreate, HandoverAccept, TransformCreate,
		CertificateIssue, CertificateRevoke, FlagCreate,
		TemplateRead, FacilityRead, FacilityManage,
		AuditRead, ReportRead,
	},
	RoleMember: {
		OrgSearch, PassportIssue, PassportRead, PassportRetire, HandoverCreate, HandoverAccept, TransformCreate, FlagCreate,
		TemplateRead, FacilityRead,
	},
}

// SyncRoles makes the database match this file: it adds missing permissions
// and default roles, and marks as deleted any permission or default-role link
// that is no longer declared. Roles that exist only in the database are left
// alone. Safe to run on every start.
func SyncRoles(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, s := range syncSteps() {
		if _, err := tx.Exec(ctx, s.sql, s.args...); err != nil {
			return fmt.Errorf("sync roles: %w", err)
		}
	}
	return tx.Commit(ctx)
}

type syncStep struct {
	sql  string
	args []any
}

// syncSteps adds what the code names and soft-deletes what it no longer does.
func syncSteps() []syncStep {
	steps := []syncStep{
		{`INSERT INTO permissions (name) SELECT unnest($1::text[])
		  ON CONFLICT (name) DO UPDATE SET is_deleted = false WHERE permissions.is_deleted`,
			[]any{AllPermissions}},
		{`UPDATE permissions SET is_deleted = true WHERE name <> ALL($1::text[]) AND NOT is_deleted`,
			[]any{AllPermissions}},
		{`INSERT INTO roles (name) SELECT unnest($1::text[])
		  ON CONFLICT (name) DO UPDATE SET is_deleted = false WHERE roles.is_deleted`,
			[]any{slices.Collect(maps.Keys(DefaultRoles))}},
	}
	for role, perms := range DefaultRoles {
		steps = append(steps,
			syncStep{`INSERT INTO role_permissions (role_id, permission_id)
			      SELECT r.id, p.id FROM roles r, permissions p
			      WHERE r.name = $1 AND p.name = ANY($2::text[])
			      ON CONFLICT (role_id, permission_id) DO UPDATE SET is_deleted = false
			      WHERE role_permissions.is_deleted`, []any{role, perms}},
			syncStep{`UPDATE role_permissions rp SET is_deleted = true
			      FROM roles r, permissions p
			      WHERE rp.role_id = r.id AND rp.permission_id = p.id
			        AND r.name = $1 AND p.name <> ALL($2::text[]) AND NOT rp.is_deleted`, []any{role, perms}},
		)
	}
	return steps
}
