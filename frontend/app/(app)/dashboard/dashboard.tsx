"use client";

import { useQuery, type UseQueryResult } from "@tanstack/react-query";
import {
  ArrowLeftRight,
  Building2,
  ChartColumn,
  ChevronRight,
  CircleCheck,
  ClipboardCheck,
  Clock,
  FileText,
  Flag,
  Layers,
  MapPin,
  OctagonX,
  Package,
  PackagePlus,
  QrCode,
  RefreshCcw,
  Send,
  Stamp,
  TriangleAlert,
  Users,
  type LucideIcon,
} from "lucide-react";
import Link from "next/link";

import { CapacityMeter } from "@/components/chaintrack/capacity-meter";
import { ListCard } from "@/components/chaintrack/list-card";
import { ScanButton } from "@/components/chaintrack/scan-button";
import { StatTile } from "@/components/chaintrack/stat-tile";
import { OrgStatusChip, StatusChip } from "@/components/chaintrack/status";
import {
  endpoints,
  problemOf,
  useMe,
  type HandoverRow,
  type Me,
} from "@/lib/api";
import {
  copy,
  orgKinds,
  orgStatus,
  passportStatus,
  queryKeys,
  routes,
  type ActionKey,
  type PanelKey,
  type TileKey,
} from "@/lib/consts";
import { actionsFor, layoutFor, toneOf } from "@/lib/dashboard";
import { cn } from "@/lib/utils";
import { formatDate, humanize } from "@/lib/verification";
import { PlatformOverview } from "./platform";
import { WorkspacePicker } from "./workspace-picker";

const t = copy.dashboard;
const tileIcon: Record<TileKey, LucideIcon> = {
  passportsHeld: Package,
  incoming: ArrowLeftRight,
  outgoing: Send,
  openFlags: Flag,
  certValid: FileText,
  certExpired: TriangleAlert,
  certRevoked: OctagonX,
  orgsPending: Building2,
  orgsChanges: Clock,
  orgsRejected: OctagonX,
};
const actionIcon: Record<ActionKey, LucideIcon> = {
  scan: QrCode,
  issue: PackagePlus,
  registerFacility: MapPin,
  handover: ArrowLeftRight,
  transform: RefreshCcw,
  reviewFlags: Flag,
  issueCertificate: Stamp,
  reviewOrgs: ClipboardCheck,
  templates: Layers,
  users: Users,
  reports: ChartColumn,
};
const tileHref: Partial<Record<TileKey, string>> = {
  orgsPending: routes.adminOrgs,
  orgsChanges: routes.adminOrg("rejected"),
  orgsRejected: routes.adminOrg("rejected"),
};
const kindLabel = (kind: string) =>
  orgKinds.find((k) => k.value === kind)?.label ?? humanize(kind);
const qty = (n: number, unit: string) => `${n.toLocaleString()} ${unit}`;

export function Dashboard() {
  const me = useMe().data!; // the layout only renders once it is loaded
  const layout = layoutFor(me);
  // Incoming and outgoing tiles both come from the one list of open handovers.
  const handovers = useQuery({
    queryKey: queryKeys.panel("handovers"),
    queryFn: endpoints.awaitingHandovers,
    enabled:
      !!layout &&
      !!me.organisation &&
      layout.tiles.some((x) => x === "incoming" || x === "outgoing"),
  });

  if (!layout) return <WorkspacePicker />; // a member on the bare address

  const mine = me.organisation?.id;
  const incoming = handovers.data?.filter((h) => h.to_org === mine);
  const outgoing = handovers.data?.filter((h) => h.from_org === mine);

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <header className="grid gap-1.5 pb-1">
        <h1 className="text-2xl leading-7 font-semibold tracking-tight lg:text-2xl lg:leading-8">
          {t.greeting(me.user.name)}
        </h1>
        {me.organisation ? (
          <p className="flex flex-wrap items-center gap-2 text-sm text-ink-muted">
            <OrgStatusChip status={me.organisation.status} />
            {me.organisation.name} · {kindLabel(me.organisation.kind)}
          </p>
        ) : (
          <p className="text-sm text-ink-muted">{t.platformLine}</p>
        )}
      </header>

      <section className="grid grid-cols-2 gap-2 sm:gap-3 lg:grid-cols-4">
        {layout.tiles.map((tile) =>
          tile === "incoming" || tile === "outgoing" ? (
            <Tile
              key={tile}
              tile={tile}
              value={(tile === "incoming" ? incoming : outgoing)?.length}
              failed={handovers.isError ? t.loadFailed : undefined}
            />
          ) : (
            <CountTile key={tile} tile={tile} />
          ),
        )}
      </section>

      {!me.organisation && <PlatformOverview />}

      <QuickActions me={me} />

      <div
        className={cn(
          "grid gap-4 lg:gap-5",
          layout.side.length > 0 && "lg:grid-cols-[1.6fr_1fr]",
        )}
      >
        <div className="grid min-w-0 content-start gap-4 lg:gap-5">
          {layout.main.map((p) => (
            <Panel
              key={p}
              panel={p}
              incoming={incoming}
              handovers={handovers}
            />
          ))}
        </div>
        {layout.side.length > 0 && (
          <div className="grid min-w-0 content-start gap-4 lg:gap-5">
            {layout.side.map((p) => (
              <Panel
                key={p}
                panel={p}
                incoming={incoming}
                handovers={handovers}
              />
            ))}
          </div>
        )}
      </div>
    </main>
  );
}

function PassportChip({ status }: { status: string }) {
  if (status === passportStatus.revoked)
    return (
      <StatusChip tone="danger" icon={OctagonX}>
        {humanize(status)}
      </StatusChip>
    );
  if (status === passportStatus.active)
    return (
      <StatusChip tone="verified" icon={CircleCheck}>
        {humanize(status)}
      </StatusChip>
    );
  return (
    <StatusChip tone="neutral" icon={Package}>
      {humanize(status)}
    </StatusChip>
  );
}

function Tile({
  tile,
  value,
  failed,
}: {
  tile: TileKey;
  value: number | undefined;
  failed?: string;
}) {
  return (
    <StatTile
      label={t.tiles[tile].label}
      note={t.tiles[tile].note}
      icon={tileIcon[tile]}
      value={value}
      tone={toneOf(tile, value)}
      href={tileHref[tile]}
      failed={failed}
    />
  );
}

// Each count is its own query, so a slow list never holds up the others.
function CountTile({
  tile,
}: {
  tile: Exclude<TileKey, "incoming" | "outgoing">;
}) {
  const count = useQuery({
    queryKey: queryKeys.count(tile),
    queryFn: () => endpoints.count(tile),
  });
  return (
    <Tile
      tile={tile}
      value={count.data}
      failed={count.isError ? t.loadFailed : undefined}
    />
  );
}

const actionCard =
  "grid grid-cols-[auto_1fr_auto] py-2.5 items-center gap-3 rounded-lg border border-border bg-surface px-3 text-left text-sm font-semibold text-ink transition-colors hover:border-primary";

function QuickActions({ me }: { me: Me }) {
  const actions = actionsFor(me);
  if (actions.length === 0) return null;
  const face = (key: ActionKey) => {
    const Icon = actionIcon[key];
    return (
      <>
        <span className="grid size-8 place-items-center rounded-md bg-primary-soft text-primary">
          <Icon className="size-4.5" aria-hidden />
        </span>
        <span className="min-w-0 truncate">{t.actions[key]}</span>
        <ChevronRight className="size-4 text-ink-muted" aria-hidden />
      </>
    );
  };
  return (
    <section
      aria-label={t.quickActions}
      className="grid grid-cols-2 gap-2 lg:grid-cols-4"
    >
      {actions.map((a) =>
        a.key === "scan" ? (
          <ScanButton
            key={a.key}
            trigger={(open) => (
              <button type="button" onClick={open} className={actionCard}>
                {face(a.key)}
              </button>
            )}
          />
        ) : (
          <Link key={a.key} href={a.href} className={actionCard}>
            {face(a.key)}
          </Link>
        ),
      )}
    </section>
  );
}

type HandoverQuery = UseQueryResult<HandoverRow[]>;

function Panel({
  panel,
  incoming,
  handovers,
}: {
  panel: PanelKey;
  incoming: HandoverRow[] | undefined;
  handovers: HandoverQuery;
}) {
  switch (panel) {
    case "incoming":
      return <IncomingPanel rows={incoming} query={handovers} />;
    case "passports":
      return (
        <QueryPanel
          panel={panel}
          fn={endpoints.passports}
          render={(x) => (
            <li
              key={x.id}
              className="grid grid-cols-[1fr_auto] items-center gap-x-4 gap-y-1 px-5 py-3 sm:grid-cols-[1.4fr_1fr_1fr_auto]"
            >
              <Link
                href={routes.passportDetail(x.id)}
                className="code min-h-tap content-center text-sm font-semibold text-primary hover:underline sm:min-h-0"
              >
                {x.code}
              </Link>
              <span className="text-sm font-semibold sm:order-none">
                {humanize(x.commodity)}
              </span>
              <span className="fig text-sm">{qty(x.remaining, x.unit)}</span>
              <span className="justify-self-end">
                <PassportChip status={x.status} />
              </span>
            </li>
          )}
        />
      );
    case "capacity":
      return (
        <QueryPanel
          panel={panel}
          fn={endpoints.facilities}
          render={(f) => (
            <li key={f.id} className="px-5 py-4">
              <CapacityMeter
                title={f.name}
                used={f.capacity_used}
                max={f.capacity}
                unit={f.capacity_unit}
              />
            </li>
          )}
        />
      );
    case "certificates":
      return (
        <QueryPanel
          panel={panel}
          fn={endpoints.certificates}
          render={(c) => (
            <li
              key={c.id}
              className="flex min-h-14 flex-wrap items-center justify-between gap-2 px-5 py-3"
            >
              <span className="text-sm font-semibold">{humanize(c.type)}</span>
              <span className="text-sm text-ink-muted">
                {humanize(c.status)} · {formatDate(c.valid_to)}
              </span>
            </li>
          )}
        />
      );
    case "flags":
      return (
        <QueryPanel
          panel={panel}
          fn={endpoints.openFlags}
          render={(f) => (
            <li key={f.id} className="grid gap-1 px-5 py-3">
              <span className="text-sm">{f.reason}</span>
              <span className="text-xs text-ink-muted">
                {formatDate(f.created_at)}
              </span>
            </li>
          )}
        />
      );
    case "orgsPending":
      return (
        <QueryPanel
          panel={panel}
          href={routes.adminOrgs}
          fn={() =>
            endpoints.orgs(orgStatus.pending).then((o) => o.slice(0, 5))
          }
          render={(o) => (
            <li
              key={o.id}
              className="flex min-h-14 flex-wrap items-center justify-between gap-2 px-5 py-3"
            >
              <Link
                href={routes.adminOrg("organisations", o.id)}
                className="text-sm font-semibold text-primary"
              >
                {o.name}
              </Link>
              <span className="text-sm text-ink-muted">
                {kindLabel(o.kind)} · {formatDate(o.created_at)}
              </span>
            </li>
          )}
        />
      );
  }
}

function QueryPanel<T>({
  panel,
  fn,
  render,
  href,
}: {
  panel: PanelKey;
  fn: () => Promise<T[]>;
  render: (row: T) => React.ReactNode;
  href?: string;
}) {
  const q = useQuery({ queryKey: queryKeys.panel(panel), queryFn: fn });
  const p = t.panels[panel];
  const state = q.isPending
    ? "loading"
    : q.isError
      ? "failed"
      : q.data.length === 0
        ? "empty"
        : "ready";
  return (
    <ListCard
      title={p.title}
      sub={p.sub}
      empty={p.empty}
      href={href}
      state={state}
      error={q.isError ? problemOf(q.error) : undefined}
    >
      {q.data?.map(render)}
    </ListCard>
  );
}

function IncomingPanel({
  rows,
  query,
}: {
  rows: HandoverRow[] | undefined;
  query: HandoverQuery;
}) {
  const names = useQuery({
    queryKey: queryKeys.directory,
    queryFn: endpoints.directory,
    enabled: !!rows?.length,
  });
  const p = t.panels.incoming;
  const state = query.isPending
    ? "loading"
    : query.isError
      ? "failed"
      : !rows?.length
        ? "empty"
        : "ready";
  return (
    <ListCard
      title={p.title}
      sub={p.sub}
      empty={p.empty}
      state={state}
      error={query.isError ? problemOf(query.error) : undefined}
    >
      {rows?.map((h) => (
        <li
          key={h.id}
          className="grid grid-cols-[auto_1fr] items-center gap-3 px-5 py-3 sm:grid-cols-[auto_1fr_auto]"
        >
          <span className="grid size-9 place-items-center rounded-md bg-primary-soft text-primary">
            <Package className="size-4.5" aria-hidden />
          </span>
          <span className="grid min-w-0 gap-0.5">
            <span className="truncate text-sm font-semibold">
              {t.from(names.data?.get(h.from_org) ?? t.someOrg)}
            </span>
            <span className="text-xs text-ink-muted">
              {t.stage(humanize(h.stage))} · {formatDate(h.created_at)}
            </span>
          </span>
          <span className="col-start-2 flex flex-wrap items-center gap-2 sm:col-start-3 sm:justify-end">
            <span className="fig text-sm">{h.quantity.toLocaleString()}</span>
            <StatusChip tone="pending" icon={Clock}>
              {t.tiles.incoming.note}
            </StatusChip>
          </span>
        </li>
      ))}
    </ListCard>
  );
}
