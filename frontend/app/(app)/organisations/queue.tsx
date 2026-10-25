"use client";

import { useQueries, useQuery } from "@tanstack/react-query";
import { BadgeCheck, Building2, CircleX, Clock, FileText } from "lucide-react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { parseAsString, parseAsStringLiteral } from "nuqs";
import { useState } from "react";

import { DataWidget } from "@/components/chaintrack/data-widget";
import { ListBar, ListPager } from "@/components/chaintrack/list-controls";
import { PageHeader } from "@/components/chaintrack/page-header";
import { PageStats } from "@/components/chaintrack/page-stats";
import { Notice } from "@/components/chaintrack/notice";
import { KeyValue } from "@/components/chaintrack/passport-card";
import { OrgStatusChip } from "@/components/chaintrack/status";
import { buttonVariants } from "@/components/ui/button";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  documentFileURL,
  endpoints,
  usePaged,
  type OrgDocument,
  type Organisation,
} from "@/lib/api";
import {
  adminTabs,
  copy,
  issuerKinds,
  listSorts,
  orgKinds,
  orgStatus,
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
export const kindLabel = (kind: string) =>
  orgKinds.find((k) => k.value === kind)?.label ?? humanize(kind);
const daysSince = (iso: string) =>
  Math.max(0, Math.floor((Date.now() - new Date(iso).getTime()) / 86_400_000));

export function Queue() {
  const state = useListParams("created_at", listFilters); // oldest first: longest waiting on top
  const tab = adminTabs.find((x) => x.value === state.filters.tab)!;
  const selectedId = state.filters.id;

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
    })),
  });
  const picked = useQuery({
    queryKey: [...queryKeys.orgs, "one", selectedId],
    queryFn: () => endpoints.org(selectedId!),
    enabled: !!selectedId,
  });

  const waiting =
    counts[adminTabs.findIndex((x) => x.value === "organisations")].data;
  const oldest = waiting?.data[0];
  const selected = picked.data;

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <PageHeader
        title={t.title}
        sub={t.sub(waiting?.total ?? 0, oldest && daysSince(oldest.created_at))}
      />
      <PageStats
        queryKey={queryKeys.orgs}
        fetch={endpoints.orgsPage}
        cards={[
          {
            key: "pending",
            ...copy.stats.organisations.pending,
            icon: Building2,
            whenAny: "attention",
            filters: { status: orgStatus.pending },
          },
          {
            key: "changes",
            ...copy.stats.organisations.changes,
            icon: Clock,
            filters: { status: orgStatus.changesRequested },
          },
          {
            key: "rejected",
            ...copy.stats.organisations.rejected,
            icon: CircleX,
            filters: { status: orgStatus.rejected },
          },
          {
            key: "verified",
            ...copy.stats.organisations.verified,
            icon: BadgeCheck,
            filters: { status: orgStatus.verified },
          },
        ]}
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
    return `${routes.organisations}?${next}`;
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

/** One organisation's facts, documents and the review; `full` on its own page (no link to itself). */
export function OrgDetail({
  org,
  full,
}: {
  org: Organisation;
  full?: boolean;
}) {
  const docs = useQuery({
    queryKey: queryKeys.orgDocuments(org.id),
    queryFn: () => endpoints.documents(org.id),
  });
  const issuer =
    org.status === orgStatus.verified && issuerKinds.includes(org.kind);
  const [viewing, setViewing] = useState<OrgDocument | null>(null);

  return (
    <section className="grid min-w-0 content-start gap-6 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
      {!full && (
        <header className="flex flex-wrap items-start justify-between gap-3">
          <div className="min-w-0">
            <h2 className="text-section">{org.name}</h2>
            <p className="text-small text-ink-muted">
              {t.submitted(formatDateTime(org.created_at))}
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <OrgStatusChip status={org.status} />
            <Link
              href={routes.organisation(org.id)}
              className={buttonVariants({ variant: "secondary" })}
            >
              {t.openFull}
            </Link>
          </div>
        </header>
      )}

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
                    <button
                      type="button"
                      onClick={() => setViewing(d)}
                      className="grid h-56 place-items-center overflow-hidden rounded-sm bg-surface-sunken"
                    >
                      {upload.imageTypes.includes(d.content_type) ? (
                        // eslint-disable-next-line @next/next/no-img-element -- served with the session cookie; the image optimiser would fetch it without one
                        <img
                          src={documentFileURL(d.id)}
                          alt={d.file_name}
                          className="size-full object-contain"
                        />
                      ) : (
                        <FileText
                          className="size-16 text-ink-muted"
                          aria-hidden
                        />
                      )}
                    </button>
                    <button
                      type="button"
                      onClick={() => setViewing(d)}
                      className="flex min-h-tap items-center gap-2 text-left text-sm font-semibold text-primary underline"
                    >
                      <FileText className="size-4 shrink-0" aria-hidden />
                      <span className="min-w-0 truncate">{d.file_name}</span>
                    </button>
                    <p className="text-small text-ink-muted">
                      {copy.documentTypeLabel(d.type)} ·{" "}
                      {formatDate(d.created_at)}
                    </p>
                  </li>
                ))}
              </ul>
            )
          }
        </DataWidget>
        {viewing && (
          <DocumentViewer doc={viewing} onClose={() => setViewing(null)} />
        )}
      </div>

      {org.status !== orgStatus.verified &&
        org.status !== orgStatus.suspended && <ReviewForm orgId={org.id} />}
      {issuer && <AccreditationForm orgId={org.id} />}
    </section>
  );
}

/** A document full size in a dialog: PDFs in the browser's viewer, images fitted. */
function DocumentViewer({
  doc,
  onClose,
}: {
  doc: OrgDocument;
  onClose: () => void;
}) {
  const url = documentFileURL(doc.id);
  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="flex h-[90vh] max-w-full flex-col gap-4 rounded-lg border border-border bg-surface p-6 shadow-overlay sm:max-w-5xl">
        <DialogTitle className="grid gap-1 pr-10">
          <span className="truncate text-lg font-semibold">
            {doc.file_name}
          </span>
          <span className="text-small font-normal text-ink-muted">
            {copy.documentTypeLabel(doc.type)}
          </span>
        </DialogTitle>
        {upload.imageTypes.includes(doc.content_type) ? (
          // eslint-disable-next-line @next/next/no-img-element -- served with the session cookie; the image optimiser would fetch it without one
          <img
            src={url}
            alt={doc.file_name}
            className="min-h-0 w-full flex-1 rounded-sm bg-surface-sunken object-contain"
          />
        ) : (
          <iframe
            src={url}
            title={doc.file_name}
            className="min-h-0 w-full flex-1 rounded-sm border border-border"
          />
        )}
        <a
          href={url}
          target="_blank"
          rel="noreferrer"
          className="justify-self-start text-sm font-semibold text-primary underline"
        >
          {t.openNewTab}
        </a>
      </DialogContent>
    </Dialog>
  );
}
