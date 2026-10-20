"use client";

import { useQuery } from "@tanstack/react-query";
import {
  CheckCircle2,
  Download,
  Lock,
  Search,
  ShieldX,
  WifiOff,
} from "lucide-react";
import { parseAsString } from "nuqs";
import { useState } from "react";

import { ListBar } from "@/components/chaintrack/list-controls";
import {
  ClientTable,
  ListTable,
  column,
  type ListColumn,
} from "@/components/chaintrack/list-table";
import { PageHeader } from "@/components/chaintrack/page-header";
import { Loading, PageState } from "@/components/chaintrack/page-state";
import { StatusChip } from "@/components/chaintrack/status";
import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  downloadURL,
  endpoints,
  problemOf,
  useMeGuard,
  usePaged,
  type AuditEntry,
  type OutcomeRow,
  type SavedReport,
  type VolumeRow,
} from "@/lib/api";
import {
  apiPaths,
  copy,
  exportFormats,
  permissions,
  queryKeys,
  reportsCopy,
  routes,
} from "@/lib/consts";
import { useListParams } from "@/lib/list-params";
import { formatDateTime } from "@/lib/verification";

const t = reportsCopy;
const c = t.columns;

const ALL_ROUTES = "";
const routeOptions = [
  { value: ALL_ROUTES, label: t.allRoutes },
  ...["/passports", "/handovers", "/flags", "/attestations", "/orgs"].map(
    (r) => ({ value: r, label: r }),
  ),
];
const auditFilters = {
  start_date: parseAsString,
  end_date: parseAsString,
  route: parseAsString,
};

const volumeColumns: ListColumn<VolumeRow>[] = [
  column(
    "month",
    c.month,
    (v) => <span className="font-mono">{v.month}</span>,
    { value: (v) => v.month },
  ),
  column("commodity", c.commodity, (v) => v.commodity, {
    value: (v) => v.commodity,
  }),
  column("unit", c.unit, (v) => <span className="font-mono">{v.unit}</span>),
  column(
    "quantity",
    c.quantity,
    (v) => <span className="fig">{v.quantity.toLocaleString()}</span>,
    { value: (v) => v.quantity },
  ),
];

const outcomeColumns: ListColumn<OutcomeRow>[] = [
  column("period", "Period", (o) => <span className="font-mono">{o.period}</span>, {
    value: (o) => o.period,
  }),
  column("verified", "Verified", (o) => <span className="fig">{o.verified}</span>, {
    value: (o) => o.verified,
  }),
  column("warning", "Warning", (o) => <span className="fig">{o.warning}</span>, {
    value: (o) => o.warning,
  }),
  column("failed", "Failed", (o) => <span className="fig">{o.failed}</span>, {
    value: (o) => o.failed,
  }),
  column("total", "Total", (o) => <span className="fig">{o.total}</span>, {
    value: (o) => o.total,
  }),
  column("pct_verified", "% verified", (o) => <span className="fig">{o.pct_verified}</span>, {
    value: (o) => o.pct_verified,
  }),
];

const savedColumns: ListColumn<SavedReport>[] = [
  column("name", "Name", (s) => s.name, { value: (s) => s.name }),
  column("kind", "Kind", (s) => <span className="font-mono">{s.kind}</span>, {
    value: (s) => s.kind,
  }),
  column("created_at", "Created", (s) => (
    <span className="text-ink-muted">{formatDateTime(s.created_at)}</span>
  )),
];

const auditColumns: ListColumn<AuditEntry>[] = [
  column("seq", c.seq, (a) => <span className="font-mono">{a.seq}</span>, {
    sortable: true,
  }),
  column(
    "created_at",
    c.created,
    (a) => (
      <span className="text-ink-muted">{formatDateTime(a.created_at)}</span>
    ),
    { sortable: true },
  ),
  column("actor", c.actor, (a) => (
    <span className="block max-w-48 truncate">
      {a.actor_email || a.actor_user_id || "–"}
    </span>
  )),
  column(
    "route",
    c.action,
    (a) => (
      <span className="font-mono text-xs">
        {a.method} {a.route}
      </span>
    ),
    { sortable: true },
  ),
  column("subject", c.subject, (a) => (
    <span
      className="block max-w-32 truncate font-mono text-xs"
      title={a.subject_id ?? ""}
    >
      {a.subject_id ?? "–"}
    </span>
  )),
  column("status", c.status, (a) => (
    <span className="font-mono">{a.status}</span>
  )),
  column("hash", c.hash, (a) => (
    <span
      className="block max-w-28 truncate font-mono text-xs text-ink-muted"
      title={a.hash}
    >
      {a.hash}
    </span>
  )),
];

export function Reports() {
  const me = useMeGuard((me) => (me ? undefined : routes.signin));
  const allowed =
    me.ready &&
    (me.data!.permissions.includes(permissions.auditRead) ||
      me.data!.permissions.includes(permissions.reportRead));

  const [volumeStart, setVolumeStart] = useState("");
  const [volumeEnd, setVolumeEnd] = useState("");
  const volume = useQuery({
    queryKey: [...queryKeys.volume, volumeStart, volumeEnd],
    queryFn: () =>
      endpoints.volumeReport({
        startDate: volumeStart || undefined,
        endDate: volumeEnd || undefined,
      }),
    enabled: allowed,
  });

  const [outStart, setOutStart] = useState("");
  const [outEnd, setOutEnd] = useState("");
  const outcomes = useQuery({
    queryKey: [...queryKeys.outcomes, outStart, outEnd],
    queryFn: () =>
      endpoints.outcomesReport({
        startDate: outStart || undefined,
        endDate: outEnd || undefined,
      }),
    enabled: allowed,
  });
  const saved = useQuery({
    queryKey: queryKeys.saved,
    queryFn: endpoints.savedReports,
    enabled: allowed,
  });
  const [savedName, setSavedName] = useState("");
  const [savedKind, setSavedKind] = useState("volume");

  const list = useListParams("-created_at", auditFilters);
  const f = list.filters;
  const auditParams = {
    start_date: f.start_date ?? undefined,
    end_date: f.end_date ?? undefined,
    route: f.route ?? undefined,
  };
  const audit = usePaged(
    queryKeys.audit,
    endpoints.auditPage,
    list.query(auditParams),
  );

  const chain = useQuery({
    queryKey: [...queryKeys.audit, "verify"],
    queryFn: endpoints.auditVerify,
    enabled: false,
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
  if (!me.ready) return <Loading className="p-4 lg:p-8" />;
  if (!allowed)
    return (
      <PageState icon={Lock} title={t.notAllowedTitle}>
        {t.notAllowedBody}
      </PageState>
    );

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <PageHeader title={t.title} sub={t.sub} />

      <Section title={t.volume}>
        <div className="flex flex-wrap items-end gap-3">
          <DateField
            label={t.startDate}
            value={volumeStart}
            onChange={setVolumeStart}
          />
          <DateField
            label={t.endDate}
            value={volumeEnd}
            onChange={setVolumeEnd}
          />
          <Downloads
            path={apiPaths.reportsVolume}
            params={{ start_date: volumeStart, end_date: volumeEnd }}
          />
        </div>
        {volume.isPending ? (
          <Loading />
        ) : volume.isError ? (
          <p role="alert" className="text-sm text-danger-ink">
            {problemOf(volume.error)}
          </p>
        ) : (
          <ClientTable
            id="volume"
            label={t.volume}
            columns={volumeColumns}
            rows={volume.data}
            rowKey={(v) => `${v.month}|${v.commodity}|${v.unit}`}
            empty={t.noData}
          />
        )}
      </Section>

      <Section title={t.outcomes}>
        <div className="flex flex-wrap items-end gap-3">
          <DateField label={t.startDate} value={outStart} onChange={setOutStart} />
          <DateField label={t.endDate} value={outEnd} onChange={setOutEnd} />
          <Downloads
            path={apiPaths.reportsOutcomes}
            params={{ start_date: outStart, end_date: outEnd }}
          />
        </div>
        {outcomes.isPending ? (
          <Loading />
        ) : outcomes.isError ? (
          <p role="alert" className="text-sm text-danger-ink">
            {problemOf(outcomes.error)}
          </p>
        ) : (
          <ClientTable
            id="outcomes"
            label={t.outcomes}
            columns={outcomeColumns}
            rows={outcomes.data}
            rowKey={(o) => o.period}
            empty={t.noData}
          />
        )}
      </Section>

      <Section title={t.saved}>
        <div className="flex flex-wrap items-end gap-3">
          <label className="grid gap-1.5">
            <span className="text-sm font-semibold">{t.savedName}</span>
            <Input
              value={savedName}
              onChange={(e) => setSavedName(e.target.value)}
              className="w-56"
              placeholder="Volume by commodity · Monthly"
            />
          </label>
          <div className="grid gap-1.5">
            <span className="text-sm font-semibold">{t.savedKind}</span>
            <Select
              items={[
                { value: "volume", label: "volume" },
                { value: "blocked", label: "blocked" },
                { value: "verification-outcomes", label: "verification-outcomes" },
                { value: "audit", label: "audit" },
              ]}
              value={savedKind}
              onValueChange={(v) => setSavedKind(v as string)}
            >
              <SelectTrigger className="w-56">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {["volume", "blocked", "verification-outcomes", "audit"].map((k) => (
                  <SelectItem key={k} value={k}>
                    {k}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <Button
            disabled={!savedName.trim()}
            onClick={async () => {
              await endpoints.createSaved({
                name: savedName.trim(),
                kind: savedKind,
                params: {
                  start_date: volumeStart,
                  end_date: volumeEnd,
                },
              });
              setSavedName("");
              saved.refetch();
            }}
          >
            {t.save}
          </Button>
        </div>
        {saved.isPending ? (
          <Loading />
        ) : saved.isError ? (
          <p role="alert" className="text-sm text-danger-ink">
            {problemOf(saved.error)}
          </p>
        ) : (
          <ClientTable
            id="saved"
            label={t.saved}
            columns={savedColumns}
            rows={saved.data}
            rowKey={(s) => s.id}
            empty={t.noData}
          />
        )}
        <p className="text-sm text-ink-muted">
          Running a saved report keeps its CSV/PDF in the bucket; open it from its files.
        </p>
      </Section>

      <Section title={t.auditTrail}>
        <ListTable
          label={t.auditTrail}
          columns={auditColumns}
          rows={audit.data?.data ?? []}
          total={audit.data?.total ?? 0}
          list={list}
          rowKey={(a) => a.id}
          loading={audit.isPending}
          error={audit.isError ? problemOf(audit.error) : undefined}
          empty={t.noData}
          toolbar={
            <div className="flex flex-wrap items-end gap-3">
              <ListBar list={list} />
              <DateField
                label={t.startDate}
                value={f.start_date ?? ""}
                onChange={(v) => list.setFilters({ start_date: v || null })}
              />
              <DateField
                label={t.endDate}
                value={f.end_date ?? ""}
                onChange={(v) => list.setFilters({ end_date: v || null })}
              />
              <div className="grid gap-1.5">
                <span className="text-sm font-semibold">{t.route}</span>
                <Select
                  items={routeOptions}
                  value={f.route ?? ALL_ROUTES}
                  onValueChange={(v) =>
                    list.setFilters({ route: (v as string) || null })
                  }
                >
                  <SelectTrigger className="w-44">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {routeOptions.map((o) => (
                      <SelectItem key={o.value} value={o.value}>
                        {o.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <Downloads
                path={apiPaths.audit}
                params={{
                  start_date: f.start_date ?? "",
                  end_date: f.end_date ?? "",
                  route: f.route ?? "",
                }}
              />
            </div>
          }
        />
      </Section>

      <Section title={t.chainCheck}>
        <div className="flex flex-wrap items-center gap-3">
          <Button onClick={() => chain.refetch()} disabled={chain.isFetching}>
            <Search aria-hidden />
            {chain.isFetching ? t.running : t.runCheck}
          </Button>
          {chain.data?.ok && (
            <StatusChip tone="verified" icon={CheckCircle2}>
              {t.verified} · {t.checked(chain.data.checked)}
            </StatusChip>
          )}
          {chain.data && !chain.data.ok && (
            <StatusChip tone="danger" icon={ShieldX}>
              {t.brokenAt(chain.data.broken_at)} ·{" "}
              {t.checked(chain.data.checked)}
            </StatusChip>
          )}
        </div>
      </Section>
    </main>
  );
}

function Section({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section className="grid min-w-0 gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
      <h2 className="text-section">{title}</h2>
      {children}
    </section>
  );
}

function DateField({
  label,
  value,
  onChange,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
}) {
  return (
    <label className="grid gap-1.5">
      <span className="text-sm font-semibold">{label}</span>
      <Input
        type="date"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-44"
      />
    </label>
  );
}

/** The same report as a CSV or PDF file (the API's ?format=), with the filters on screen. */
function Downloads({
  path,
  params,
}: {
  path: string;
  params: Record<string, string>;
}) {
  return (
    <div className="flex flex-wrap gap-2">
      {exportFormats.map((format, i) => (
        <a
          key={format}
          href={downloadURL(path, format, params)}
          className={buttonVariants({
            variant: i === 0 ? "primary" : "secondary",
          })}
        >
          <Download aria-hidden />
          {i === 0 ? t.downloadCsv : t.downloadPdf}
        </a>
      ))}
    </div>
  );
}
