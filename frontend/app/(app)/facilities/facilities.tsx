"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  CircleCheck,
  Clock,
  Factory,
  FileText,
  Inbox,
  MapPin,
  OctagonX,
  Pickaxe,
  Plus,
  Sprout,
  Trash2,
  TriangleAlert,
  Upload,
  Warehouse,
  WifiOff,
  X,
  type LucideIcon,
} from "lucide-react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { parseAsString } from "nuqs";

import { useConfirm } from "@/components/chaintrack/confirm";
import { CapacityMeter } from "@/components/chaintrack/capacity-meter";
import {
  Choice,
  ListBar,
  ListPager,
} from "@/components/chaintrack/list-controls";
import { PageHeader } from "@/components/chaintrack/page-header";
import { MapView, type Shape } from "@/components/chaintrack/map";
import { Notice } from "@/components/chaintrack/notice";
import { DataWidget } from "@/components/chaintrack/data-widget";
import { KeyValue } from "@/components/chaintrack/passport-card";
import { StatusChip } from "@/components/chaintrack/status";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  endpoints,
  facilityLicenceURL,
  useMe,
  usePaged,
  type Facility,
} from "@/lib/api";
import {
  capacityPeriods,
  copy,
  facilityStatus,
  listSorts,
  permissions,
  queryKeys,
  routes,
  upload,
  type FacilityKind,
  screenCopy,
  confirmCopy,
} from "@/lib/consts";
import { useListParams } from "@/lib/list-params";
import { facilityState, type FacilityState } from "@/lib/facility";
import type { Geometry } from "@/lib/geo";
import { cn } from "@/lib/utils";
import { formatDate, humanize } from "@/lib/verification";

const t = copy.facilities;
const local = screenCopy.facilities;

const kindIcon: Record<string, LucideIcon> = {
  farm_plot: Sprout,
  mine_site: Pickaxe,
  washing_station: Factory,
  processing_plant: Factory,
  warehouse: Warehouse,
};
const periodLabel = (period: string) =>
  capacityPeriods.find((p) => p.value === period)?.label ?? humanize(period);

const stateChip: Record<
  FacilityState,
  {
    tone: "verified" | "pending" | "warning" | "danger";
    icon: LucideIcon;
    label: string;
  }
> = {
  active: { tone: "verified", icon: CircleCheck, label: t.active },
  pending: { tone: "pending", icon: Clock, label: t.pending },
  suspended: { tone: "warning", icon: TriangleAlert, label: t.suspended },
  expired: { tone: "danger", icon: OctagonX, label: t.expired },
};
const stateTone: Record<FacilityState, Shape["tone"]> = {
  active: "primary",
  pending: "pending",
  suspended: "warning",
  expired: "danger",
};
const stateNote: Partial<Record<FacilityState, string>> = {
  pending: t.pendingNote,
  suspended: t.suspendedNote,
  expired: t.expiredNote,
};

export function FacilityChip({ f }: { f: Facility }) {
  const c = stateChip[facilityState(f)];
  return (
    <StatusChip tone={c.tone} icon={c.icon}>
      {c.label}
    </StatusChip>
  );
}

const shapeOf = (f: Facility): Shape[] =>
  f.location
    ? [
        {
          id: f.id,
          geometry: f.location as Geometry,
          tone: stateTone[facilityState(f)],
        },
      ]
    : [];

/** The API stores a point as GeoJSON [longitude, latitude]; show it latitude first. */
function coordsOf(f: Facility): [number, number] | null {
  const c = f.location?.coordinates;
  if (
    f.location?.type !== "Point" ||
    !Array.isArray(c) ||
    typeof c[0] !== "number" ||
    typeof c[1] !== "number"
  )
    return null;
  return [c[0], c[1]];
}

const listFilters = {
  kind: parseAsString,
  facility: parseAsString,
};

export function Facilities() {
  const me = useMe().data!; // the layout only renders once /me has loaded
  const kinds = useQuery({
    queryKey: queryKeys.facilityKinds,
    queryFn: endpoints.facilityKinds,
  });
  const list = useListParams("name", listFilters);
  const kind = list.filters.kind ?? "";
  const q = usePaged(
    queryKeys.facilities,
    endpoints.facilitiesPage,
    list.query({ kind }),
  );
  const selectedId = list.filters.facility ?? "";
  const picked = useQuery({
    queryKey: [...queryKeys.facilities, selectedId],
    queryFn: () => endpoints.facility(selectedId),
    enabled: !!selectedId,
  });

  const selected = picked.data;
  const mayManage = me.permissions.includes(permissions.facilityManage);
  const maySuspend = me.permissions.includes(permissions.facilitySuspend);
  const org = me.organisation?.name ?? copy.shell.platform;
  const filtered = !!(list.search || kind);
  const kindLabel = (value: string) =>
    kinds.data?.find((k) => k.key === value)?.label ?? humanize(value);
  const kindOptions = [
    { value: "", label: t.filterAll },
    ...(kinds.data ?? []).map((k) => ({ value: k.key, label: k.label })),
  ];

  return (
    <DataWidget
      query={q}
      className="px-4 py-8 lg:px-8"
      page={{ icon: WifiOff, title: copy.errors.loadTitle }}
    >
      {({ data: rows, total }) => (
        <main className="grid w-full content-start gap-6 p-4 lg:p-8">
          <PageHeader
            title={t.title}
            sub={t.sub(total, org)}
            actions={
              mayManage && (
                <Link href={routes.facilityNew} className={buttonVariants()}>
                  <Plus aria-hidden />
                  {t.register}
                </Link>
              )
            }
          />

          {total === 0 && !filtered ? (
            <div className="grid justify-items-center gap-2 px-4 py-12 text-center">
              <span className="grid size-14 place-items-center rounded-full bg-surface-sunken text-ink-muted">
                <Inbox className="size-7" aria-hidden />
              </span>
              <h2 className="text-title">{local.emptyTitle}</h2>
              <p className="max-w-[48ch] text-ink-muted">{t.empty}</p>
              {mayManage && (
                <Link
                  href={routes.facilityNew}
                  className={cn(buttonVariants(), "mt-2")}
                >
                  <Plus aria-hidden />
                  {t.register}
                </Link>
              )}
            </div>
          ) : (
            <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.2fr)] lg:gap-6">
              <div className="grid min-w-0 content-start gap-3">
                <ListBar list={list} sorts={listSorts.name}>
                  <Choice
                    label={t.filterLabel}
                    options={kindOptions}
                    value={kind}
                    onChange={(v) =>
                      list.setFilters({
                        kind: (v || null) as FacilityKind | null,
                        facility: null,
                      })
                    }
                  />
                </ListBar>
                {rows.length === 0 ? (
                  <p className="rounded-lg border border-border bg-surface p-4 text-ink-muted">
                    {filtered ? copy.list.noMatch : t.emptyType}
                  </p>
                ) : (
                  <ul className="grid divide-y divide-border overflow-hidden rounded-lg border border-border bg-surface shadow-card">
                    {rows.map((f) => (
                      <Row
                        key={f.id}
                        f={f}
                        selected={f.id === selectedId}
                        kindLabel={kindLabel}
                      />
                    ))}
                  </ul>
                )}
                <ListPager list={list} total={total} />
              </div>
              {selected ? (
                <Detail
                  f={selected}
                  mayManage={mayManage}
                  maySuspend={maySuspend}
                  kindLabel={kindLabel}
                />
              ) : (
                <section className="grid min-w-0 content-start gap-3 rounded-lg border border-border bg-surface p-4 shadow-card">
                  <MapView
                    label={t.mapLabel}
                    shapes={rows.flatMap(shapeOf)}
                    className="h-72 overflow-hidden rounded-md lg:h-96"
                  />
                  <ul className="flex flex-wrap gap-x-4 gap-y-1 text-small text-ink-muted">
                    <li className="flex items-center gap-1.5">
                      <CircleCheck
                        className="size-4 text-primary"
                        aria-hidden
                      />
                      {t.legend.active}
                    </li>
                    <li className="flex items-center gap-1.5">
                      <Clock className="size-4 text-pending" aria-hidden />
                      {t.legend.pending}
                    </li>
                    <li className="flex items-center gap-1.5">
                      <OctagonX className="size-4 text-danger" aria-hidden />
                      {t.legend.inactive}
                    </li>
                  </ul>
                  <p className="text-small text-ink-subtle">
                    {local.selectPrompt}
                  </p>
                </section>
              )}
            </div>
          )}
        </main>
      )}
    </DataWidget>
  );
}

/** One facility: a compact row with what's left this period; actions live in the detail. */
function Row({
  f,
  selected,
  kindLabel,
}: {
  f: Facility;
  selected: boolean;
  kindLabel: (value: string) => string;
}) {
  const params = useSearchParams();
  const Icon = kindIcon[f.kind] ?? MapPin;
  const pct =
    f.capacity > 0
      ? Math.min(100, Math.round((f.capacity_used / f.capacity) * 100))
      : 0;
  const next = new URLSearchParams(params);
  next.set("facility", f.id);
  return (
    <li>
      <Link
        href={`${routes.facilities}?${next}`}
        aria-current={selected ? "true" : undefined}
        scroll={false}
        className={cn(
          "grid grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-3 px-4 py-3 hover:bg-surface-sunken",
          selected && "bg-primary-soft/60 shadow-selected",
        )}
      >
        <span className="grid size-9 place-items-center rounded-md bg-primary-soft text-primary">
          <Icon className="size-4" aria-hidden />
        </span>
        <span className="grid min-w-0 gap-1">
          <span className="truncate text-sm font-semibold">{f.name}</span>
          <span className="flex min-w-0 items-center gap-2 text-xs text-ink-muted">
            <span className="truncate">
              {kindLabel(f.kind)} ·{" "}
              <span className="font-mono">{f.licence_number}</span>
            </span>
            <span
              className="hidden h-1.5 w-16 shrink-0 overflow-hidden rounded-full bg-surface-sunken sm:block"
              aria-hidden
            >
              <span
                className={cn(
                  "block h-full",
                  pct >= 90 ? "bg-warning" : "bg-primary",
                )}
                style={{ width: `${pct}%` }}
              />
            </span>
            <span className="hidden font-mono sm:inline">{pct}%</span>
          </span>
        </span>
        <FacilityChip f={f} />
      </Link>
    </li>
  );
}

function Detail({
  f,
  mayManage,
  maySuspend,
  kindLabel,
}: {
  f: Facility;
  mayManage: boolean;
  maySuspend: boolean;
  kindLabel: (value: string) => string;
}) {
  const router = useRouter();
  const client = useQueryClient();
  const confirm = useConfirm();
  const ask = async (
    o: { title: string; body: string; label: string },
    run: () => void,
    danger = false,
  ) => {
    if (
      await confirm({
        title: o.title,
        body: o.body,
        confirmLabel: o.label,
        danger,
      })
    )
      run();
  };
  const refresh = () =>
    client.invalidateQueries({ queryKey: queryKeys.facilities });
  const set = useMutation({
    mutationFn: (status: string) => endpoints.setFacilityStatus(f.id, status),
    onSuccess: refresh,
  });
  const remove = useMutation({
    mutationFn: () => endpoints.deleteFacility(f.id),
    onSuccess: () => {
      refresh();
      router.replace(routes.facilities);
    },
  });
  const approve = useMutation({
    mutationFn: () => endpoints.approveFacility(f.id),
    onSuccess: refresh,
  });
  const licence = useMutation({
    mutationFn: (file: File) => endpoints.uploadLicence(f.id, file),
    onSuccess: refresh,
  });
  const Icon = kindIcon[f.kind] ?? MapPin;
  const active = f.status === facilityStatus.active;
  const pending = f.status === facilityStatus.pending;
  const coords = coordsOf(f);
  const note = stateNote[facilityState(f)];

  return (
    <section
      aria-label={f.name}
      className="grid min-w-0 content-start gap-6 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6"
    >
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex min-w-0 items-center gap-3">
          <span className="grid size-11 shrink-0 place-items-center rounded-md bg-primary-soft text-primary">
            <Icon className="size-5" aria-hidden />
          </span>
          <div className="min-w-0">
            <h2 className="text-section break-words">{f.name}</h2>
            <p className="text-small text-ink-muted">{kindLabel(f.kind)}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <FacilityChip f={f} />
          <Link
            href={routes.facilities}
            aria-label={local.closeDetail}
            className="grid size-11 place-items-center rounded-md text-ink-muted hover:bg-surface-sunken"
          >
            <X className="size-5" aria-hidden />
          </Link>
        </div>
      </header>

      <KeyValue
        facts={[
          { label: t.new.kind, value: kindLabel(f.kind) },
          { label: t.detail.licence, value: f.licence_number, mono: true },
          {
            label: t.detail.licenceValid,
            value: f.licence_valid_to
              ? formatDate(f.licence_valid_to)
              : t.detail.licenceUnknown,
          },
          { label: local.areaLabel, value: t.detail.area(f.area_ha) },
          {
            label: t.detail.declared,
            value: t.detail.declaredOf(
              f.capacity,
              f.capacity_unit,
              periodLabel(f.capacity_period),
            ),
          },
          {
            label: local.issuedLabel,
            value: t.detail.issued(
              f.capacity_used,
              f.capacity_unit,
              periodLabel(f.capacity_period),
            ),
          },
          coords
            ? {
                label: t.detail.coordinates,
                value: `${coords[1]}, ${coords[0]}`,
                mono: true,
              }
            : { label: t.detail.coordinates, value: local.noPoint },
          { label: local.registeredLabel, value: formatDate(f.created_at) },
        ]}
      />

      {f.location && (
        <MapView
          label={t.detail.mapLabel(f.name)}
          shapes={shapeOf(f)}
          selected={f.id}
          className="h-64 overflow-hidden rounded-md border border-border"
        />
      )}

      <CapacityMeter
        title={t.detail.declared}
        used={f.capacity_used}
        max={f.capacity}
        unit={f.capacity_unit}
      />
      {note && <Notice kind={pending ? "info" : "warning"}>{note}</Notice>}

      <div className="flex flex-wrap items-center justify-between gap-3">
        <span className="grid gap-0.5">
          <span className="text-sm font-semibold">{t.detail.licenceDoc}</span>
          {f.licence_document ? (
            <a
              href={facilityLicenceURL(f.id)}
              target="_blank"
              rel="noreferrer"
              className="flex min-h-tap items-center gap-1.5 text-sm font-semibold text-primary underline"
            >
              <FileText className="size-4" aria-hidden />
              {t.detail.licenceOpen}
            </a>
          ) : (
            <span className="text-small text-ink-muted">
              {t.detail.licenceNone}
            </span>
          )}
        </span>
        {mayManage && (
          <label
            className={cn(
              buttonVariants({ variant: "secondary" }),
              "cursor-pointer",
              licence.isPending && "pointer-events-none opacity-50",
            )}
          >
            <Upload aria-hidden />
            {t.detail.licenceUpload}
            <input
              type="file"
              accept={upload.types.join(",")}
              className="sr-only"
              onChange={(e) => {
                const file = e.target.files?.[0];
                if (file) licence.mutate(file);
                e.target.value = "";
              }}
            />
          </label>
        )}
      </div>

      {(mayManage || maySuspend) && (
        <div className="flex flex-wrap justify-end gap-2 border-t border-border pt-4">
          {mayManage && (
            <Button
              variant="danger-outline"
              loading={remove.isPending}
              onClick={() =>
                ask(
                  {
                    title: t.detail.deleteTitle(f.name),
                    body: t.detail.deleteBody,
                    label: t.detail.deleteConfirm,
                  },
                  () => remove.mutate(),
                  true,
                )
              }
            >
              <Trash2 aria-hidden />
              {t.detail.delete}
            </Button>
          )}
          {maySuspend && pending && (
            <Button
              loading={approve.isPending}
              onClick={() =>
                ask(confirmCopy.facilityApprove(f.name), () => approve.mutate())
              }
            >
              <CircleCheck aria-hidden />
              {t.approve}
            </Button>
          )}
          {maySuspend &&
            !pending &&
            (active ? (
              <Button
                variant="secondary"
                loading={set.isPending}
                onClick={() =>
                  ask(
                    {
                      title: t.suspendTitle(f.name),
                      body: t.suspendBody,
                      label: t.suspendConfirm,
                    },
                    () => set.mutate(facilityStatus.suspended),
                    true,
                  )
                }
              >
                {t.suspend}
              </Button>
            ) : (
              <Button
                variant="secondary"
                loading={set.isPending}
                onClick={() =>
                  ask(confirmCopy.facilityReinstate(f.name), () =>
                    set.mutate(facilityStatus.active),
                  )
                }
              >
                {t.reinstate}
              </Button>
            ))}
        </div>
      )}
    </section>
  );
}
