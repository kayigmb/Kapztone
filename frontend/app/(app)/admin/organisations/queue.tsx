"use client";

import { useQueries, useQuery } from "@tanstack/react-query";
import { Building2, FileText, Lock, WifiOff } from "lucide-react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { parseAsString, parseAsStringLiteral } from "nuqs";

import { DataWidget } from "@/components/chaintrack/data-widget";
import { ListBar, ListPager } from "@/components/chaintrack/list-controls";
import { PageHeader } from "@/components/chaintrack/page-header";
import { Notice } from "@/components/chaintrack/notice";
import { Loading, PageState } from "@/components/chaintrack/page-state";
import { KeyValue } from "@/components/chaintrack/passport-card";
import { OrgStatusChip } from "@/components/chaintrack/status";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  documentFileURL,
  endpoints,
  problemOf,
  useMeGuard,
  usePaged,
  type Organisation,
} from "@/lib/api";
import {
  adminTabs,
  copy,
  issuerKinds,
  listSorts,
  orgKinds,
  orgStatus,
  permissions,
  queryKeys,
  routes,
  upload,
  type AdminTab,
} from "@/lib/consts";
import { useListParams } from "@/lib/list-params";
import { cn } from "@/lib/utils";
import { formatDate, formatDateTime, humanize } from "@/lib/verification";
import { AccreditationForm } from "./accreditation-form";
import { ReviewForm } from "./review-form";

const t = copy.admin;
/** A tab's filters, as the API takes them (several values comma-separated). */
const filtersOf = (x: (typeof adminTabs)[number]) => ({
  status: x.statuses.join(","),
  kind: "kinds" in x ? x.kinds.join(",") : undefined,
});
const tabValues = adminTabs.map((x) => x.value);
const listFilters = {
  tab: parseAsStringLiteral(tabValues).withDefault(tabValues[0]),
  id: parseAsString,
};
const kindLabel = (kind: string) =>
  orgKinds.find((k) => k.value === kind)?.label ?? humanize(kind);
const daysSince = (iso: string) =>
  Math.max(0, Math.floor((Date.now() - new Date(iso).getTime()) / 86_400_000));

export function Queue() {
  const state = useListParams("created_at", listFilters); // oldest first: longest waiting on top
  const tab = adminTabs.find((x) => x.value === state.filters.tab)!;
  const selectedId = state.filters.id;

  const me = useMeGuard((me) => (me ? undefined : routes.signin));
  const allowed =
    me.ready && me.data!.permissions.includes(permissions.orgVerify);

  const list = usePaged([...queryKeys.orgs, tab.value], endpoints.orgsPage, {
    ...state.query(filtersOf(tab)),
  });
  // Each tab's count, and the oldest waiting registration for the header.
  const counts = useQueries({
    queries: adminTabs.map((x) => ({
      queryKey: [...queryKeys.orgs, "count", x.value],
      queryFn: () =>
        endpoints.orgsPage({
          page: 1,
          size: 1,
          sort: "created_at",
          filters: filtersOf(x),
        }),
      enabled: allowed,
    })),
  });
  const picked = useQuery({
    queryKey: [...queryKeys.orgs, "one", selectedId],
    queryFn: () => endpoints.org(selectedId!),
    enabled: allowed && !!selectedId,
  });

  if (me.isError) {
    return (
      <PageState
        icon={WifiOff}
        tone="danger"
        title={copy.errors.loadTitle}
        retry={() => me.refetch()}
      >
        {problemOf(me.error)}
      </PageState>
    );
  }
  if (!me.ready) return <Loading className="px-4 py-8 lg:px-8" />;
  if (!allowed)
    return (
      <PageState icon={Lock} title={t.notAllowedTitle}>
        {t.notAllowedBody}
      </PageState>
    );

  const waiting = counts[0].data;
  const oldest = waiting?.data[0];
  const selected = picked.data;

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <PageHeader
        title={t.title}
        sub={t.sub(waiting?.total ?? 0, oldest && daysSince(oldest.created_at))}
      />

      <Tabs
        value={tab.value}
        onValueChange={(v) =>
          state.setFilters({ tab: v as AdminTab, id: null })
        }
      >
        <TabsList className="h-auto w-full flex-wrap justify-start lg:w-fit">
          {adminTabs.map((x) => (
            <TabsTrigger
              key={x.value}
              value={x.value}
              className="flex-none px-3"
            >
              {counts[adminTabs.indexOf(x)].data === undefined
                ? t.tabs[x.value]
                : t.tabCount(
                    t.tabs[x.value],
                    counts[adminTabs.indexOf(x)].data!.total,
                  )}
            </TabsTrigger>
          ))}
        </TabsList>
        <ListBar className="mt-4" list={state} sorts={listSorts.waiting} />

        {adminTabs.map((x) => (
          <TabsContent key={x.value} value={x.value} className="mt-4">
            <DataWidget query={list} retryLabel={copy.errors.tryAgain}>
              {(page) => (
                <div className="grid gap-4 lg:grid-cols-[--spacing(95)_1fr] lg:gap-6">
                  <div className="grid content-start gap-3">
                    <OrgList
                      tab={x.value}
                      orgs={page.data}
                      selectedId={selectedId}
                    />
                    <ListPager list={state} total={page.total} />
                  </div>
                  {selected ? (
                    <OrgDetail key={selected.id} org={selected} />
                  ) : (
                    <p className="hidden rounded-lg border border-dashed border-border-strong p-6 text-ink-muted lg:block">
                      {t.selectOne}
                    </p>
                  )}
                </div>
              )}
            </DataWidget>
          </TabsContent>
        ))}
      </Tabs>
    </main>
  );
}

function OrgList({
  tab,
  orgs,
  selectedId,
}: {
  tab: AdminTab;
  orgs: Organisation[];
  selectedId: string | null;
}) {
  const params = useSearchParams();
  const hrefOf = (id: string) => {
    // keeps the page, search and sort
    const next = new URLSearchParams(params);
    next.set("tab", tab);
    next.set("id", id);
    return `${routes.adminOrgs}?${next}`;
  };
  if (orgs.length === 0)
    return (
      <p className="rounded-lg border border-border bg-surface p-4 text-ink-muted">
        {t.empty[tab]}
      </p>
    );
  return (
    <ul
      aria-label={t.queueLabel(t.tabs[tab])}
      className="grid content-start divide-y divide-border overflow-hidden rounded-lg border border-border bg-surface shadow-card"
    >
      {orgs.map((o) => (
        <li key={o.id}>
          <Link
            href={hrefOf(o.id)}
            scroll={false}
            aria-current={o.id === selectedId ? "true" : undefined}
            className={cn(
              "grid grid-cols-[auto_1fr_auto] items-center gap-3 px-4 py-2.5 hover:bg-surface-sunken",
              o.id === selectedId && "bg-primary-soft/60 shadow-selected",
            )}
          >
            <span className="grid size-9 place-items-center rounded-md bg-primary-soft text-primary">
              <Building2 className="size-4" aria-hidden />
            </span>
            <span className="min-w-0">
              <span className="block truncate text-sm font-semibold">
                {o.name}
              </span>
              <span className="block text-xs text-ink-muted">
                {kindLabel(o.kind)} · {formatDate(o.created_at)}
              </span>
            </span>
            {o.status === orgStatus.pending ? (
              <span className="text-small text-ink-muted">
                {t.age(daysSince(o.created_at))}
              </span>
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
  const docs = useQuery({
    queryKey: queryKeys.orgDocuments(org.id),
    queryFn: () => endpoints.documents(org.id),
  });
  const issuer =
    org.status === orgStatus.verified && issuerKinds.includes(org.kind);

  return (
    <section className="grid min-w-0 content-start gap-6 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <h2 className="text-section">{org.name}</h2>
          <p className="text-small text-ink-muted">
            {t.submitted(formatDateTime(org.created_at))}
          </p>
        </div>
        <OrgStatusChip status={org.status} />
      </header>

      <KeyValue
        facts={[
          { label: t.facts.name, value: org.name, wide: true },
          { label: t.facts.slug, value: org.slug, mono: true },
          { label: t.facts.kind, value: kindLabel(org.kind) },
          { label: t.facts.submitted, value: formatDate(org.created_at) },
          { label: t.facts.waiting, value: t.age(daysSince(org.created_at)) },
          ...(org.review_reason
            ? [{ label: t.facts.reason, value: org.review_reason, wide: true }]
            : []),
        ]}
      />

      <div className="grid gap-3">
        <h3 className="text-title">{t.documents}</h3>
        <DataWidget query={docs}>
          {(rows) =>
            rows.length === 0 ? (
              <Notice kind="warning">{t.noDocuments}</Notice>
            ) : (
              <ul className="grid gap-3 md:grid-cols-2">
                {rows.map((d) => (
                  <li
                    key={d.id}
                    className="grid gap-2 rounded-md border border-border p-3"
                  >
                {upload.imageTypes.includes(d.content_type) && (
                  // Opens full size in a new tab to zoom.
                  <a
                    href={documentFileURL(d.id)}
                    target="_blank"
                    rel="noreferrer"
                  >
                    {/* eslint-disable-next-line @next/next/no-img-element -- served with the session cookie; the image optimiser would fetch it without one */}
                    <img
                      src={documentFileURL(d.id)}
                      alt={d.file_name}
                      className="max-h-56 w-full rounded-sm bg-surface-sunken object-contain"
                    />
                  </a>
                )}
                <a
                  href={documentFileURL(d.id)}
                  target="_blank"
                  rel="noreferrer"
                  className="flex min-h-tap items-center gap-2 text-sm font-semibold text-primary underline"
                >
                  <FileText className="size-4 shrink-0" aria-hidden />
                  <span className="min-w-0 truncate">{d.file_name}</span>
                </a>
                    <p className="text-small text-ink-muted">
                      {copy.documentTypeLabel(d.type)} · {formatDate(d.created_at)}
                    </p>
                  </li>
                ))}
              </ul>
            )
          }
        </DataWidget>
      </div>

      {org.status !== orgStatus.verified && <ReviewForm orgId={org.id} />}
      {issuer && <AccreditationForm orgId={org.id} />}
    </section>
  );
}
