"use client";

import { useQuery } from "@tanstack/react-query";
import { Building2, Eye, Flag, Package, Users } from "lucide-react";

import dynamic from "next/dynamic";

// Recharts is heavy and needs the browser: it ships in its own chunk behind a skeleton instead of weighing down the dashboard route.
const Bars = dynamic(
  () => import("@/components/chaintrack/charts").then((m) => m.Bars),
  { ssr: false, loading: () => <ChartSkeleton /> },
);
const LinesOverTime = dynamic(
  () => import("@/components/chaintrack/charts").then((m) => m.LinesOverTime),
  { ssr: false, loading: () => <ChartSkeleton /> },
);

function ChartSkeleton() {
  return (
    <div
      aria-busy="true"
      className="grid h-48 place-items-center rounded-md bg-surface-sunken"
    >
      <span className="text-sm text-ink-muted">{t.chartsLoading}</span>
    </div>
  );
}
import {
  ClientTable,
  RowAction,
  actionsColumn,
  column,
  type ListColumn,
} from "@/components/chaintrack/list-table";
import { Notice } from "@/components/chaintrack/notice";
import { StatTile } from "@/components/chaintrack/stat-tile";
import { endpoints, problemOf, type PlatformStats } from "@/lib/api";
import { flagStatus, platformCopy as t, queryKeys, routes } from "@/lib/consts";
import { formatDateTime, humanize } from "@/lib/verification";

const countOf = (rows: { key: string; count: number }[], key: string) =>
  rows.find((r) => r.key === key)?.count ?? 0;
const total = (rows: { count: number }[]) =>
  rows.reduce((n, r) => n + r.count, 0);
const toBars = (rows: { key: string; count: number }[]) =>
  rows.map((r) => ({ name: humanize(r.key), value: r.count }));
type TopOrg = PlatformStats["top_organisations"][number];
const topOrgColumns: ListColumn<TopOrg>[] = [
  column("name", t.topOrgs.name, (o) => o.name, { value: (o) => o.name }),
  column(
    "lots",
    t.topOrgs.lots,
    (o) => <span className="fig">{o.lots.toLocaleString()}</span>,
    { value: (o) => o.lots },
  ),
  column(
    "quantity",
    t.topOrgs.quantity,
    (o) => <span className="fig">{o.quantity.toLocaleString()}</span>,
    { value: (o) => o.quantity },
  ),
  actionsColumn((o) => (
    <RowAction
      label={t.topOrgs.open}
      icon={Eye}
      href={routes.organisationsTab("approved", o.id)}
    />
  )),
];
const seriesKeys = [
  "passports",
  "handovers",
  "transforms",
  "flags",
  "signups",
] as const;

/** What the platform admin sees first: totals, the last 30 days, volume, and whether the ledger keeps up. */
export function PlatformOverview() {
  const q = useQuery({
    queryKey: queryKeys.stats,
    queryFn: endpoints.platformStats,
    refetchInterval: 60_000,
  });
  if (q.isError)
    return (
      <Notice kind="blocking">
        {t.loadFailed} {problemOf(q.error)}
      </Notice>
    );
  const s = q.data;

  return (
    <div className="grid gap-4 lg:gap-5">
      <section className="grid grid-cols-2 gap-2 sm:gap-3 lg:grid-cols-4">
        <StatTile
          icon={Building2}
          label={t.tiles.organisations.label}
          href={routes.organisations}
          value={s && total(s.organisations)}
          note={t.tiles.organisations.note}
        />
        <StatTile
          icon={Users}
          label={t.tiles.users.label}
          href={routes.users}
          value={s?.users}
          note={t.tiles.users.note}
        />
        <StatTile
          icon={Package}
          label={t.tiles.passports.label}
          value={s && total(s.passports)}
          href={routes.issue}
          note={t.tiles.passports.note}
        />
        <StatTile
          icon={Flag}
          label={t.tiles.flags.label}
          href={routes.flags}
          note={t.tiles.flags.note}
          value={s && countOf(s.flags, flagStatus.open)}
          tone={
            s && countOf(s.flags, flagStatus.open) > 0 ? "alert" : "neutral"
          }
        />
      </section>

      <Card
        title={t.activity.title}
        sub={
          s
            ? `${t.activity.sub} · ${t.updated(formatDateTime(s.generated_at))}`
            : t.activity.sub
        }
      >
        {s && (
          <LinesOverTime
            rows={s.activity}
            x="date"
            label={t.activity.title}
            series={seriesKeys.map((key) => ({ key, label: t.series[key] }))}
          />
        )}
      </Card>

      <div className="grid gap-4 lg:grid-cols-2 lg:gap-5">
        <Card title={t.volume.title} sub={t.volume.sub}>
          {s &&
            (s.volume.length ? (
              <Bars
                label={t.volume.title}
                valueLabel={t.volume.lots}
                rows={s.volume.map((v) => ({
                  name: `${humanize(v.commodity)} (${v.unit})`,
                  value: v.lots,
                }))}
              />
            ) : (
              <Empty />
            ))}
        </Card>
        <Outbox s={s} />
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:gap-5">
        {(Object.keys(t.breakdowns) as (keyof typeof t.breakdowns)[]).map(
          (k) => (
            <Card key={k} title={t.breakdowns[k]}>
              {s &&
                (s[k].length ? (
                  <Bars
                    label={t.breakdowns[k]}
                    valueLabel={t.value}
                    rows={toBars(s[k])}
                  />
                ) : (
                  <Empty />
                ))}
            </Card>
          ),
        )}
      </div>

      <Card title={t.topOrgs.title}>
        {s && (
          <ClientTable
            id="topOrgs"
            label={t.topOrgs.title}
            columns={topOrgColumns}
            rows={s.top_organisations}
            rowKey={(o) => o.id}
            empty={t.none}
          />
        )}
      </Card>
    </div>
  );
}

function Card({
  title,
  sub,
  children,
}: {
  title: string;
  sub?: string;
  children?: React.ReactNode;
}) {
  return (
    <section className="grid min-w-0 content-start gap-3 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-5">
      <header className="grid gap-0.5">
        <h2 className="text-base font-semibold">{title}</h2>
        {sub && <p className="text-xs text-ink-muted">{sub}</p>}
      </header>
      {children ?? (
        <div
          className="h-40 animate-pulse rounded-md bg-surface-sunken motion-reduce:animate-none"
          aria-busy="true"
        />
      )}
    </section>
  );
}

function Empty() {
  return <p className="text-sm text-ink-muted">{t.none}</p>;
}

function Outbox({ s }: { s?: PlatformStats }) {
  const o = s?.outbox;
  return (
    <Card title={t.outbox.title}>
      {o &&
        (o.pending === 0 && o.failed === 0 ? (
          <p className="text-sm text-ink-muted">{t.outbox.clear}</p>
        ) : (
          <dl className="grid grid-cols-2 gap-3 text-sm">
            <div className="grid gap-0.5">
              <dt className="text-xs text-ink-muted">{t.outbox.pending}</dt>
              <dd className="fig text-2xl">{o.pending.toLocaleString()}</dd>
            </div>
            <div className="grid gap-0.5">
              <dt className="text-xs text-ink-muted">{t.outbox.failed}</dt>
              <dd
                className={
                  o.failed ? "fig text-2xl text-danger" : "fig text-2xl"
                }
              >
                {o.failed.toLocaleString()}
              </dd>
            </div>
            {o.oldest_pending_seconds != null && (
              <p className="col-span-2 text-xs text-ink-muted">
                {t.outbox.oldest(Math.round(o.oldest_pending_seconds / 60))}
              </p>
            )}
          </dl>
        ))}
    </Card>
  );
}
