"use client";

import { useQueries, useQuery } from "@tanstack/react-query";
import { Building2, FileText, Lock, WifiOff } from "lucide-react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";

import { Notice } from "@/components/chaintrack/notice";
import { Loading, PageState } from "@/components/chaintrack/page-state";
import { KeyValue } from "@/components/chaintrack/passport-card";
import { OrgStatusChip } from "@/components/chaintrack/status";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { documentFileURL, endpoints, problemOf, useMeGuard, type Organisation } from "@/lib/api";
import {
  adminTabs, copy, issuerKinds, orgKinds, orgStatus, permissions, queryKeys, routes, upload, type AdminTab, type OrgStatus,
} from "@/lib/consts";
import { cn } from "@/lib/utils";
import { formatDate, formatDateTime, humanize } from "@/lib/verification";
import { AccreditationForm } from "./accreditation-form";
import { ReviewForm } from "./review-form";

const t = copy.admin;
const statuses = Object.values(orgStatus);
const kindLabel = (kind: string) => orgKinds.find((k) => k.value === kind)?.label ?? humanize(kind);
const daysSince = (iso: string) => Math.max(0, Math.floor((Date.now() - new Date(iso).getTime()) / 86_400_000));

export function Queue() {
  const router = useRouter();
  const params = useSearchParams();
  const tab = (adminTabs.find((x) => x.value === params.get("tab")) ?? adminTabs[0]);
  const selectedId = params.get("id");

  const me = useMeGuard((me) => (me ? undefined : routes.signin));
  const allowed = me.ready && me.data!.permissions.includes(permissions.orgVerify);

  // ponytail: every status is fetched (one page each) so every tab can show its count; paging when a list outgrows a page.
  const lists = useQueries({
    queries: statuses.map((s) => ({ queryKey: queryKeys.orgList(s), queryFn: () => endpoints.orgs(s), enabled: allowed })),
  });
  const byStatus = Object.fromEntries(statuses.map((s, i) => [s, lists[i].data ?? []])) as Record<OrgStatus, Organisation[]>;
  const listOf = (x: (typeof adminTabs)[number]) =>
    x.statuses.flatMap((s) => byStatus[s]).filter((o) => !("kinds" in x) || x.kinds.includes(o.kind))
      .sort((a, b) => a.created_at.localeCompare(b.created_at)); // oldest first: longest waiting on top

  if (me.isError) {
    return <PageState icon={WifiOff} tone="danger" title={copy.errors.loadTitle} retry={() => me.refetch()}>{problemOf(me.error)}</PageState>;
  }
  if (!me.ready) return <Loading className="px-4 py-8 lg:px-8" />;
  if (!allowed) return <PageState icon={Lock} title={t.notAllowedTitle}>{t.notAllowedBody}</PageState>;

  const failed = lists.find((l) => l.isError);
  const loading = lists.some((l) => l.isPending);
  const pending = listOf(adminTabs[0]);
  const selected = selectedId ? statuses.flatMap((s) => byStatus[s]).find((o) => o.id === selectedId) : undefined;

  return (
    <main className="grid content-start gap-6 px-4 py-6 lg:px-8 lg:py-8">
      <header className="flex flex-wrap items-end justify-between gap-3">
        <div className="grid gap-1">
          <h1 className="text-[22px] font-semibold lg:text-page">{t.title}</h1>
          <p className="text-sm text-ink-muted">{t.sub(pending.length, pending[0] && daysSince(pending[0].created_at))}</p>
        </div>
      </header>

      <Tabs value={tab.value} onValueChange={(v) => router.replace(routes.adminOrg(v as AdminTab))}>
        <TabsList className="h-auto w-full flex-wrap justify-start lg:w-fit">
          {adminTabs.map((x) => (
            <TabsTrigger key={x.value} value={x.value} className="min-h-tap flex-none px-3">
              {loading ? t.tabs[x.value] : t.tabCount(t.tabs[x.value], listOf(x).length)}
            </TabsTrigger>
          ))}
        </TabsList>

        {adminTabs.map((x) => (
          <TabsContent key={x.value} value={x.value} className="mt-4">
            {failed ? (
              <PageState icon={WifiOff} tone="danger" title={copy.errors.loadTitle}
                retry={() => lists.forEach((l) => l.refetch())}>{problemOf(failed.error)}</PageState>
            ) : loading ? (
              <Loading />
            ) : (
              <div className="grid gap-4 lg:grid-cols-[380px_1fr] lg:gap-6">
                <OrgList tab={x.value} orgs={listOf(x)} selectedId={selectedId} />
                {selected ? (
                  <OrgDetail key={selected.id} org={selected} />
                ) : (
                  <p className="hidden rounded-lg border border-dashed border-border-strong p-6 text-ink-muted lg:block">{t.selectOne}</p>
                )}
              </div>
            )}
          </TabsContent>
        ))}
      </Tabs>
    </main>
  );
}

function OrgList({ tab, orgs, selectedId }: { tab: AdminTab; orgs: Organisation[]; selectedId: string | null }) {
  if (orgs.length === 0) return <p className="rounded-lg border border-border bg-surface p-4 text-ink-muted">{t.empty[tab]}</p>;
  return (
    <ul aria-label={t.queueLabel(t.tabs[tab])} className="grid content-start gap-2">
      {orgs.map((o) => (
        <li key={o.id}>
          <Link href={routes.adminOrg(tab, o.id)} aria-current={o.id === selectedId ? "true" : undefined}
            className={cn("grid min-h-tap grid-cols-[40px_1fr_auto] items-center gap-3 rounded-md border bg-surface p-3 hover:bg-surface-sunken",
              o.id === selectedId ? "border-2 border-primary" : "border-border")}>
            <span className="grid size-10 place-items-center rounded-md bg-primary-soft text-primary">
              <Building2 className="size-5" aria-hidden />
            </span>
            <span className="min-w-0">
              <span className="block truncate font-semibold">{o.name}</span>
              <span className="block text-small text-ink-muted">{kindLabel(o.kind)} · {formatDate(o.created_at)}</span>
            </span>
            {o.status === orgStatus.pending ? (
              <span className="text-small text-ink-muted">{t.age(daysSince(o.created_at))}</span>
            ) : (
              <OrgStatusChip status={o.status} />
            )}
          </Link>
        </li>
      ))}
    </ul>
  );
}

function OrgDetail({ org }: { org: Organisation }) {
  const docs = useQuery({ queryKey: queryKeys.orgDocuments(org.id), queryFn: () => endpoints.documents(org.id) });
  const issuer = org.status === orgStatus.verified && issuerKinds.includes(org.kind);

  return (
    <section className="grid min-w-0 content-start gap-6 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h2 className="text-section">{org.name}</h2>
          <p className="text-small text-ink-muted">{t.submitted(formatDateTime(org.created_at))}</p>
        </div>
        <OrgStatusChip status={org.status} />
      </header>

      <KeyValue facts={[
        { label: t.facts.name, value: org.name, wide: true },
        { label: t.facts.slug, value: org.slug, mono: true },
        { label: t.facts.kind, value: kindLabel(org.kind) },
        { label: t.facts.submitted, value: formatDate(org.created_at) },
        { label: t.facts.waiting, value: t.age(daysSince(org.created_at)) },
        ...(org.review_reason ? [{ label: t.facts.reason, value: org.review_reason, wide: true }] : []),
      ]} />

      <div className="grid gap-3">
        <h3 className="text-title">{t.documents}</h3>
        {docs.isPending ? <Loading /> : docs.isError ? (
          <Notice kind="blocking">{problemOf(docs.error)}</Notice>
        ) : docs.data.length === 0 ? (
          <Notice kind="warning">{t.noDocuments}</Notice>
        ) : (
          <ul className="grid gap-3 md:grid-cols-2">
            {docs.data.map((d) => (
              <li key={d.id} className="grid gap-2 rounded-md border border-border p-3">
                {upload.imageTypes.includes(d.content_type) && (
                  // Opens full size in a new tab to zoom.
                  <a href={documentFileURL(d.id)} target="_blank" rel="noreferrer">
                    {/* eslint-disable-next-line @next/next/no-img-element -- served with the session cookie; the image optimiser would fetch it without one */}
                    <img src={documentFileURL(d.id)} alt={d.file_name} className="max-h-56 w-full rounded-sm bg-surface-sunken object-contain" />
                  </a>
                )}
                <a href={documentFileURL(d.id)} target="_blank" rel="noreferrer"
                  className="flex min-h-tap items-center gap-2 text-sm font-semibold text-primary underline">
                  <FileText className="size-4 shrink-0" aria-hidden />
                  <span className="min-w-0 truncate">{d.file_name}</span>
                </a>
                <p className="text-small text-ink-muted">
                  {copy.documentTypeLabel(d.type)} · {formatDate(d.created_at)}
                </p>
              </li>
            ))}
          </ul>
        )}
      </div>

      {org.status !== orgStatus.verified && <ReviewForm orgId={org.id} />}
      {issuer && <AccreditationForm orgId={org.id} />}
    </section>
  );
}
