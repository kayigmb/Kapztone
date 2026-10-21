"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft, Lock } from "lucide-react";
import Link from "next/link";
import { use, useState } from "react";

import { CertificateTile } from "@/components/chaintrack/certificate-tile";
import { useConfirm } from "@/components/chaintrack/confirm";
import { DataWidget } from "@/components/chaintrack/data-widget";
import {
  ClientTable,
  column,
  type ListColumn,
} from "@/components/chaintrack/list-table";
import { Notice } from "@/components/chaintrack/notice";
import { LoadingWithLogo, PageState } from "@/components/chaintrack/page-state";
import { StatusChip } from "@/components/chaintrack/status";
import { TextField } from "@/components/chaintrack/text-field";
import { Button, buttonVariants } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  endpoints,
  useMeGuard,
  type Facility,
  type MemberRow,
  type Organisation,
  type PassportRow,
} from "@/lib/api";
import type { ListParams } from "@/lib/list";
import {
  confirmCopy,
  copy,
  orgStatus,
  permissions,
  queryKeys,
  routes,
} from "@/lib/consts";
import { formatDate, humanize } from "@/lib/verification";
import { certificateState } from "../../../certificates/certificates";
import { FacilityChip } from "../../../facilities/facilities";
import { statusLook } from "../../../passports/[id]/passport-detail";
import { Person, StatusOf, roleLabel } from "../../../users/users";
import { OrgDetail } from "../queue";

const t = copy.admin;
// ponytail: each section shows the newest 50; open the full list screen if an organisation outgrows it.
const sectionSize = 50;
const sections = [
  "members",
  "facilities",
  "passports",
  "certificates",
] as const;

export function OrgPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const me = useMeGuard((me) => (me ? undefined : routes.signin));
  const allowed =
    me.ready && me.data!.permissions.includes(permissions.orgVerify);
  const org = useQuery({
    queryKey: queryKeys.org(id),
    queryFn: () => endpoints.org(id),
    enabled: allowed,
  });

  if (!me.ready) return <LoadingWithLogo />;
  if (!allowed)
    return (
      <PageState icon={Lock} title={t.notAllowedTitle}>
        {t.notAllowedBody}
      </PageState>
    );
  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <Link
        href={routes.adminOrgs}
        className={buttonVariants({ variant: "ghost", className: "w-fit" })}
      >
        <ArrowLeft aria-hidden />
        {t.back}
      </Link>
      <DataWidget query={org}>
        {(o) => (
          <>
            <OrgDetail org={o} full />
            <StatusPanel org={o} />
            <Records orgId={o.id} />
          </>
        )}
      </DataWidget>
    </main>
  );
}

function StatusPanel({ org }: { org: Organisation }) {
  const client = useQueryClient();
  const confirm = useConfirm();
  const [reason, setReason] = useState("");
  const change = useMutation({
    mutationFn: (suspend: boolean) =>
      suspend
        ? endpoints.orgSuspend(org.id, reason.trim())
        : endpoints.orgReactivate(org.id),
    onSuccess: () => {
      setReason("");
      client.invalidateQueries({ queryKey: queryKeys.orgs });
    },
  });
  const ask = async (suspend: boolean) => {
    const c = suspend
      ? confirmCopy.orgSuspend(org.name)
      : confirmCopy.orgReactivate(org.name);
    if (
      await confirm({
        title: c.title,
        body: c.body,
        confirmLabel: c.label,
        danger: suspend,
      })
    )
      change.mutate(suspend);
  };
  const s = t.status;
  return (
    <section className="grid gap-3 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
      <h2 className="text-title">{s.title}</h2>
      {org.status === orgStatus.verified ? (
        <form
          noValidate
          className="grid gap-3 md:grid-cols-[1fr_auto] md:items-end"
          onSubmit={(e) => {
            e.preventDefault();
            ask(true);
          }}
        >
          <TextField
            label={s.suspendNote}
            help={s.suspendHelp}
            value={reason}
            onChange={(e) => setReason(e.target.value)}
          />
          <Button
            type="submit"
            variant="danger-outline"
            disabled={!reason.trim()}
            loading={change.isPending}
          >
            {s.suspend}
          </Button>
        </form>
      ) : org.status === orgStatus.suspended ? (
        <div className="grid gap-3">
          <Notice kind="warning">{s.suspendedBody}</Notice>
          <Button
            className="w-fit"
            variant="secondary"
            onClick={() => ask(false)}
            loading={change.isPending}
          >
            {s.reactivate}
          </Button>
        </div>
      ) : (
        <p className="text-small text-ink-muted">{s.pendingBody}</p>
      )}
    </section>
  );
}

function Records({ orgId }: { orgId: string }) {
  return (
    <Tabs defaultValue={sections[0]}>
      <TabsList className="h-auto w-full flex-wrap justify-start lg:w-fit">
        {sections.map((x) => (
          <TabsTrigger key={x} value={x}>
            {t.sections[x]}
          </TabsTrigger>
        ))}
      </TabsList>
      <TabsContent value="members" className="mt-4">
        <Members orgId={orgId} />
      </TabsContent>
      <TabsContent value="facilities" className="mt-4">
        <Facilities orgId={orgId} />
      </TabsContent>
      <TabsContent value="passports" className="mt-4">
        <Passports orgId={orgId} />
      </TabsContent>
      <TabsContent value="certificates" className="mt-4">
        <Certificates orgId={orgId} />
      </TabsContent>
    </Tabs>
  );
}

/** The newest page of one of the organisation's lists. */
function useSection<T>(
  name: string,
  orgId: string,
  fetch: (p: ListParams) => Promise<{ data: T[] }>,
  filters: Record<string, string> = {},
) {
  return useQuery({
    queryKey: [...queryKeys.org(orgId), name],
    queryFn: () =>
      fetch({
        page: 1,
        size: sectionSize,
        sort: "-created_at",
        filters: { organisation_id: orgId, ...filters },
      }).then((p) => p.data),
  });
}

function Members({ orgId }: { orgId: string }) {
  const q = useSection<MemberRow>(
    "members",
    orgId,
    endpoints.orgMembers(orgId),
  );
  const c = copy.users.columns;
  const columns: ListColumn<MemberRow>[] = [
    column("name", c.person, (m) => (
      <Link href={routes.userDetail(m.id)} className="hover:underline">
        <Person name={m.name} email={m.email} />
      </Link>
    )),
    column("role", c.role, (m) => roleLabel(m.role)),
    column("status", c.status, (m) => <StatusOf status={m.status} />),
    column("since", c.joined, (m) => formatDate(m.since)),
  ];
  return (
    <DataWidget query={q}>
      {(rows) => (
        <ClientTable
          id="members"
          label={t.sections.members}
          columns={columns}
          rows={rows}
          rowKey={(m) => m.id}
          empty={t.emptySection.members}
        />
      )}
    </DataWidget>
  );
}

function Facilities({ orgId }: { orgId: string }) {
  const q = useSection<Facility>("facilities", orgId, endpoints.facilitiesPage);
  const c = t.recordColumns;
  const columns: ListColumn<Facility>[] = [
    column("name", c.name, (f) => (
      <Link
        href={routes.facility(f.id)}
        className="font-semibold hover:underline"
      >
        {f.name}
      </Link>
    )),
    column("kind", c.kind, (f) => humanize(f.kind)),
    column("status", c.status, (f) => <FacilityChip f={f} />),
  ];
  return (
    <DataWidget query={q}>
      {(rows) => (
        <ClientTable
          id="facilities"
          label={t.sections.facilities}
          columns={columns}
          rows={rows}
          rowKey={(f) => f.id}
          empty={t.emptySection.facilities}
        />
      )}
    </DataWidget>
  );
}

function Passports({ orgId }: { orgId: string }) {
  const q = useSection<PassportRow>(
    "passports",
    orgId,
    endpoints.passportsPage,
  );
  const c = t.recordColumns;
  const columns: ListColumn<PassportRow>[] = [
    column("code", c.code, (p) => (
      <Link
        href={routes.passportDetail(p.id)}
        className="font-mono font-semibold hover:underline"
      >
        {p.code}
      </Link>
    )),
    column("commodity", c.commodity, (p) => humanize(p.commodity)),
    column("remaining", c.remaining, (p) => (
      <span className="fig">
        {p.remaining.toLocaleString()} {p.unit}
      </span>
    )),
    column("status", c.status, (p) => {
      const look = statusLook[p.status] ?? statusLook.active;
      return (
        <StatusChip tone={look.tone} icon={look.icon}>
          {look.label}
        </StatusChip>
      );
    }),
  ];
  return (
    <DataWidget query={q}>
      {(rows) => (
        <ClientTable
          id="passports"
          label={t.sections.passports}
          columns={columns}
          rows={rows}
          rowKey={(p) => p.id}
          empty={t.emptySection.passports}
        />
      )}
    </DataWidget>
  );
}

function Certificates({ orgId }: { orgId: string }) {
  // What it issued, not what the narrowed gate would see about its lots.
  const q = useSection("certificates", orgId, endpoints.attestationsPage, {
    organisation_id: "",
    issuer_id: orgId,
  });
  return (
    <DataWidget query={q}>
      {(rows) =>
        rows.length === 0 ? (
          <p className="rounded-lg border border-border bg-surface p-4 text-ink-muted">
            {t.emptySection.certificates}
          </p>
        ) : (
          <ul className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
            {rows.map((c) => (
              <li key={c.id}>
                <CertificateTile
                  certificate={{
                    id: c.id,
                    type: humanize(c.type),
                    issuer: "",
                    state: certificateState(c),
                    validFrom: formatDate(c.valid_from),
                    validTo: formatDate(c.valid_to),
                  }}
                />
              </li>
            ))}
          </ul>
        )
      }
    </DataWidget>
  );
}
