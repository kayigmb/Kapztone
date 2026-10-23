"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Ban,
  CircleCheck,
  Clock,
  Lock,
  RotateCcw,
  ShieldCheck,
  ShieldOff,
} from "lucide-react";
import Link from "next/link";
import { use } from "react";

import { useConfirm } from "@/components/chaintrack/confirm";
import { DataWidget } from "@/components/chaintrack/data-widget";
import {
  ClientTable,
  RowAction,
  actionsColumn,
  column,
  type ListColumn,
} from "@/components/chaintrack/list-table";
import { DetailHeader } from "@/components/chaintrack/detail-header";
import { PageState } from "@/components/chaintrack/page-state";
import { KeyValue } from "@/components/chaintrack/passport-card";
import { StatusChip } from "@/components/chaintrack/status";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  endpoints,
  useMe,
  type AuditEntry,
  type PlatformUserRow,
} from "@/lib/api";
import {
  confirmCopy,
  copy,
  memberRoles,
  memberStatus,
  queryKeys,
  routes,
} from "@/lib/consts";
import { formatDate, formatDateTime } from "@/lib/verification";
import { Person, StatusOf, roleLabel, useSetActive } from "../users";

const t = copy.users;
type Membership = PlatformUserRow["memberships"][number];
// The latest page of activity only; link to the full audit report if more is needed.
const activitySize = 50;

export function UserDetail({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const me = useMe().data!;
  const user = useQuery({
    queryKey: queryKeys.user(id),
    queryFn: () => endpoints.user(id),
    enabled: !me.organisation,
  });
  if (me.organisation) {
    return (
      <PageState icon={Lock} title={t.platformOnlyTitle}>
        {t.platformOnlyBody}
      </PageState>
    );
  }
  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <DataWidget query={user}>
        {(u) => <Account user={u} myId={me.user.id} />}
      </DataWidget>
    </main>
  );
}

function Account({ user: u, myId }: { user: PlatformUserRow; myId: string }) {
  const { ask, pending } = useSetActive();
  const active = u.status === memberStatus.active;
  return (
    <>
      <DetailHeader
        crumb={u.name}
        title={<Person name={u.name} email={u.email} />}
        actions={
          u.id !== myId &&
          (active ? (
            <Button
              variant="danger-outline"
              onClick={() => ask(u, false)}
              loading={pending}
            >
              {t.deactivate}
            </Button>
          ) : (
            <Button
              variant="secondary"
              onClick={() => ask(u, true)}
              loading={pending}
            >
              {t.activate}
            </Button>
          ))
        }
      />
      <section className="grid gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
        <h2 className="text-title">{t.account}</h2>
        <KeyValue
          className="md:grid-cols-4"
          facts={[
            { label: t.facts.email, value: u.email, wide: true },
            { label: t.facts.status, value: <StatusOf status={u.status} /> },
            {
              label: t.facts.emailVerified,
              value: u.email_verified ? (
                <StatusChip tone="verified" icon={CircleCheck}>
                  {t.emailConfirmed}
                </StatusChip>
              ) : (
                <StatusChip tone="pending" icon={Clock}>
                  {t.emailUnconfirmed}
                </StatusChip>
              ),
            },
            {
              label: t.facts.twoFactor,
              value: u.two_factor ? (
                <StatusChip tone="verified" icon={ShieldCheck}>
                  {t.twoFactorOn}
                </StatusChip>
              ) : (
                <StatusChip tone="neutral" icon={ShieldOff}>
                  {t.twoFactorOff}
                </StatusChip>
              ),
            },
            ...(u.created_at
              ? [{ label: t.facts.created, value: formatDate(u.created_at) }]
              : []),
          ]}
        />
      </section>
      <Memberships user={u} />
      <Activity userId={u.id} />
    </>
  );
}

function Memberships({ user: u }: { user: PlatformUserRow }) {
  const client = useQueryClient();
  const confirm = useConfirm();
  const act = useMutation({
    mutationFn: ({
      orgId,
      action,
      role,
    }: {
      orgId: string;
      action: "role" | "suspend" | "reinstate";
      role?: string;
    }) =>
      action === "role"
        ? endpoints.setRole(u.id, role!, orgId)
        : action === "suspend"
          ? endpoints.suspend(u.id, orgId)
          : endpoints.reinstate(u.id, orgId),
    onSuccess: () => client.invalidateQueries({ queryKey: queryKeys.users }),
  });
  const ask = async (
    c: { title: string; body: string; label: string },
    run: () => void,
    danger = false,
  ) => {
    if (
      await confirm({
        title: c.title,
        body: c.body,
        confirmLabel: c.label,
        danger,
      })
    )
      run();
  };
  // The platform scope itself is not managed here.
  const rows = u.memberships.filter((m) => m.organisation_id);
  const columns: ListColumn<Membership>[] = [
    column("organisation", t.membershipColumns.organisation, (m) => (
      <Link
        href={routes.organisation(m.organisation_id!)}
        className="font-semibold hover:underline"
      >
        {m.organisation}
      </Link>
    )),
    column("role", t.membershipColumns.role, (m) => (
      <Select
        items={memberRoles.map((r) => ({ value: r.value, label: r.label }))}
        value={m.role}
        disabled={act.isPending || m.status !== memberStatus.active}
        onValueChange={(role) =>
          role &&
          role !== m.role &&
          ask(confirmCopy.userRole(u.name, roleLabel(role as string)), () =>
            act.mutate({
              orgId: m.organisation_id!,
              action: "role",
              role: role as string,
            }),
          )
        }
      >
        <SelectTrigger aria-label={t.changeRole} className="h-9 w-40">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {memberRoles.map((r) => (
            <SelectItem key={r.value} value={r.value}>
              {r.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    )),
    column("status", t.membershipColumns.status, (m) => (
      <StatusOf status={m.status} />
    )),
    actionsColumn((m) =>
      m.status === memberStatus.active ? (
        <RowAction
          label={t.suspend}
          icon={Ban}
          danger
          disabled={act.isPending}
          onClick={() =>
            ask(
              confirmCopy.userScopeSuspend(u.name, m.organisation ?? ""),
              () =>
                act.mutate({ orgId: m.organisation_id!, action: "suspend" }),
              true,
            )
          }
        />
      ) : (
        <RowAction
          label={t.reinstate}
          icon={RotateCcw}
          disabled={act.isPending}
          onClick={() =>
            ask(
              confirmCopy.userScopeReinstate(u.name, m.organisation ?? ""),
              () =>
                act.mutate({ orgId: m.organisation_id!, action: "reinstate" }),
            )
          }
        />
      ),
    ),
  ];
  return (
    <section className="grid gap-3">
      <h2 className="text-title">{t.memberships}</h2>
      <ClientTable
        id="memberships"
        label={t.memberships}
        columns={columns}
        rows={rows}
        rowKey={(m) => m.organisation_id!}
        empty={t.noMemberships}
      />
    </section>
  );
}

function Activity({ userId }: { userId: string }) {
  const q = useQuery({
    queryKey: [...queryKeys.user(userId), "activity"],
    queryFn: () =>
      endpoints.auditPage({
        page: 1,
        size: activitySize,
        sort: "-created_at",
        filters: { actor: userId },
      }),
  });
  const c = t.activityColumns;
  const columns: ListColumn<AuditEntry>[] = [
    column("created_at", c.when, (a) => (
      <span className="text-ink-muted">{formatDateTime(a.created_at)}</span>
    )),
    column("route", c.action, (a) => (
      <span className="font-mono text-xs">
        {a.method} {a.route}
      </span>
    )),
    column("subject", c.subject, (a) => (
      <span
        className="block max-w-40 truncate font-mono text-xs"
        title={a.subject_id ?? ""}
      >
        {a.subject_id ?? "–"}
      </span>
    )),
  ];
  return (
    <section className="grid gap-3">
      <div>
        <h2 className="text-title">{t.activity}</h2>
        <p className="text-small text-ink-muted">{t.activitySub}</p>
      </div>
      <DataWidget query={q}>
        {(p) => (
          <ClientTable
            id="activity"
            label={t.activity}
            columns={columns}
            rows={p.data}
            rowKey={(a) => a.id}
            empty={t.noActivity}
          />
        )}
      </DataWidget>
    </section>
  );
}
