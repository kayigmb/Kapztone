"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  CircleCheck, Clock, Factory, FileText, Inbox, MapPin, OctagonX, Pickaxe, Plus, Sprout, Trash2, TriangleAlert, Upload, Warehouse, WifiOff, X,
  type LucideIcon,
} from "lucide-react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";
import { toast } from "sonner";

import { CapacityMeter } from "@/components/chaintrack/capacity-meter";
import { MapView, type Shape } from "@/components/chaintrack/map";
import { Notice } from "@/components/chaintrack/notice";
import { Loading, PageState } from "@/components/chaintrack/page-state";
import { KeyValue } from "@/components/chaintrack/passport-card";
import { StatusChip } from "@/components/chaintrack/status";
import { Button, buttonVariants } from "@/components/ui/button";
import { Dialog, DialogClose, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { endpoints, facilityLicenceURL, problemOf, useMe, type Facility } from "@/lib/api";
import { capacityPeriods, copy, facilityKinds, facilityStatus, permissions, queryKeys, routes, upload } from "@/lib/consts";
import { facilityState, type FacilityState } from "@/lib/facility";
import type { Geometry } from "@/lib/geo";
import { cn } from "@/lib/utils";
import { formatDate, humanize } from "@/lib/verification";

const t = copy.facilities;
// Local copy: labels and prompts that lib/consts.ts doesn't hold yet.
const cancel = "Cancel";
const closeDetail = "Close facility details";
const emptyTitle = "No facilities yet";
const selectPrompt = "Choose a facility to see its details.";
const noPoint = "No point recorded yet.";
const areaLabel = "Area";
const registeredLabel = "Registered";
const issuedLabel = "Issued";

const kindIcon: Record<string, LucideIcon> = {
  farm_plot: Sprout, mine_site: Pickaxe, washing_station: Factory, processing_plant: Factory, warehouse: Warehouse,
};
const kindLabel = (kind: string) => facilityKinds.find((k) => k.value === kind)?.label ?? humanize(kind);
const periodLabel = (period: string) => capacityPeriods.find((p) => p.value === period)?.label ?? humanize(period);

const stateChip: Record<FacilityState, { tone: "verified" | "pending" | "warning" | "danger"; icon: LucideIcon; label: string }> = {
  active: { tone: "verified", icon: CircleCheck, label: t.active },
  pending: { tone: "pending", icon: Clock, label: t.pending },
  suspended: { tone: "warning", icon: TriangleAlert, label: t.suspended },
  expired: { tone: "danger", icon: OctagonX, label: t.expired },
};
const stateTone: Record<FacilityState, Shape["tone"]> = { active: "primary", pending: "pending", suspended: "warning", expired: "danger" };
const stateNote: Partial<Record<FacilityState, string>> = { pending: t.pendingNote, suspended: t.suspendedNote, expired: t.expiredNote };

function FacilityChip({ f }: { f: Facility }) {
  const c = stateChip[facilityState(f)];
  return <StatusChip tone={c.tone} icon={c.icon}>{c.label}</StatusChip>;
}

const shapeOf = (f: Facility): Shape[] =>
  f.location ? [{ id: f.id, geometry: f.location as Geometry, tone: stateTone[facilityState(f)] }] : [];

/** The API stores a point as GeoJSON [longitude, latitude]; show it latitude first. */
function coordsOf(f: Facility): [number, number] | null {
  const c = f.location?.coordinates;
  if (f.location?.type !== "Point" || !Array.isArray(c) || typeof c[0] !== "number" || typeof c[1] !== "number") return null;
  return [c[0], c[1]];
}

export function Facilities() {
  const me = useMe().data!; // the layout only renders once /me has loaded
  const params = useSearchParams();
  const [filter, setFilter] = useState("");
  const q = useQuery({ queryKey: queryKeys.facilities, queryFn: endpoints.facilitiesList });

  if (q.isPending) return <Loading className="px-4 py-8 lg:px-8" />;
  if (q.isError) {
    return (
      <PageState icon={WifiOff} tone="danger" title={copy.errors.loadTitle} retry={() => q.refetch()}>
        {problemOf(q.error)}
      </PageState>
    );
  }

  const rows = q.data;
  const shown = filter ? rows.filter((f) => f.kind === filter) : rows;
  const selectedId = params.get("facility");
  const selected = rows.find((f) => f.id === selectedId);
  const mayManage = me.permissions.includes(permissions.facilityManage);
  const maySuspend = me.permissions.includes(permissions.facilitySuspend);
  const org = me.organisation?.name ?? copy.shell.platform;

  return (
    <main className="grid content-start gap-6 p-4 lg:p-8">
      <header className="flex flex-wrap items-end justify-between gap-3">
        <div className="grid gap-1">
          <h1 className="text-[22px] font-semibold lg:text-page">{t.title}</h1>
          <p className="text-sm text-ink-muted">{t.sub(rows.length, org)}</p>
        </div>
        {mayManage && (
          <Link href={routes.facilityNew} className={buttonVariants()}>
            <Plus aria-hidden />
            {t.register}
          </Link>
        )}
      </header>

      <nav aria-label={t.filterLabel} className="flex flex-wrap gap-2">
        {[{ value: "", label: t.filterAll }, ...facilityKinds].map((k) => (
          <button
            key={k.value || "all"} type="button" onClick={() => setFilter(k.value)} aria-pressed={filter === k.value}
            className={cn("min-h-tap rounded-full border px-3 text-sm font-semibold",
              filter === k.value
                ? "border-primary bg-primary-soft text-primary"
                : "border-border-strong bg-surface text-ink-muted hover:bg-surface-sunken")}
          >
            {k.label}
          </button>
        ))}
      </nav>

      {rows.length === 0 ? (
        <div className="grid justify-items-center gap-2 px-4 py-12 text-center">
          <span className="grid size-14 place-items-center rounded-full bg-surface-sunken text-ink-muted">
            <Inbox className="size-7" aria-hidden />
          </span>
          <h2 className="text-title">{emptyTitle}</h2>
          <p className="max-w-[48ch] text-ink-muted">{t.empty}</p>
          {mayManage && (
            <Link href={routes.facilityNew} className={cn(buttonVariants(), "mt-2")}>
              <Plus aria-hidden />
              {t.register}
            </Link>
          )}
        </div>
      ) : (
        <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.4fr)] lg:gap-6">
          <div className="grid min-w-0 content-start gap-3">
            {shown.length === 0 ? (
              <p className="rounded-lg border border-border bg-surface p-4 text-ink-muted">{t.emptyType}</p>
            ) : (
              <ul className="grid content-start gap-3">
                {shown.map((f) => (
                  <Row key={f.id} f={f} selected={f.id === selectedId} maySuspend={maySuspend} />
                ))}
              </ul>
            )}
          </div>
          {selected ? (
            <Detail f={selected} mayManage={mayManage} maySuspend={maySuspend} />
          ) : (
            <section className="grid min-w-0 content-start gap-3 rounded-lg border border-border bg-surface p-4 shadow-card">
              <MapView label={t.mapLabel} shapes={shown.flatMap(shapeOf)} className="h-80 overflow-hidden rounded-md lg:h-[28rem]" />
              <ul className="flex flex-wrap gap-x-4 gap-y-1 text-small text-ink-muted">
                <li className="flex items-center gap-1.5"><CircleCheck className="size-4 text-primary" aria-hidden />{t.legend.active}</li>
                <li className="flex items-center gap-1.5"><Clock className="size-4 text-pending" aria-hidden />{t.legend.pending}</li>
                <li className="flex items-center gap-1.5"><OctagonX className="size-4 text-danger" aria-hidden />{t.legend.inactive}</li>
              </ul>
              <p className="text-small text-ink-subtle">{selectPrompt}</p>
            </section>
          )}
        </div>
      )}
    </main>
  );
}

function Row({ f, selected, maySuspend }: { f: Facility; selected: boolean; maySuspend: boolean }) {
  const client = useQueryClient();
  const [confirming, setConfirming] = useState(false);
  const set = useMutation({
    mutationFn: (status: string) => endpoints.setFacilityStatus(f.id, status),
    onSuccess: (_r, status) => {
      toast.success(status === facilityStatus.suspended ? t.suspendedToast : t.reinstatedToast);
      client.invalidateQueries({ queryKey: queryKeys.facilities });
    },
    onError: (err) => toast.error(problemOf(err)),
    onSettled: () => setConfirming(false),
  });
  const Icon = kindIcon[f.kind] ?? MapPin;
  const active = f.status === facilityStatus.active;
  const pending = f.status === facilityStatus.pending;

  return (
    <li className="grid gap-3 rounded-lg border border-border bg-surface p-4 shadow-card">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <Link
          href={routes.facility(f.id)} aria-current={selected ? "true" : undefined}
          className={cn("grid min-h-tap min-w-0 flex-1 grid-cols-[40px_1fr] items-center gap-3 rounded-md border-2 p-3 hover:bg-surface-sunken",
            selected ? "border-primary" : "border-transparent")}
        >
          <span className="grid size-10 place-items-center rounded-md bg-primary-soft text-primary">
            <Icon className="size-5" aria-hidden />
          </span>
          <span className="grid min-w-0 gap-0.5">
            <span className="truncate text-[15px] font-semibold">{f.name}</span>
            <span className="truncate text-small text-ink-muted">
              {kindLabel(f.kind)} · <span className="font-mono text-xs">{f.licence_number}</span>
            </span>
          </span>
        </Link>
        <div className="flex flex-wrap items-center gap-2">
          <FacilityChip f={f} />
          {maySuspend && !pending && (active ? (
            <Button variant="secondary" disabled={set.isPending} onClick={() => setConfirming(true)}>{t.suspend}</Button>
          ) : (
            <Button variant="secondary" disabled={set.isPending} onClick={() => set.mutate(facilityStatus.active)}>{t.reinstate}</Button>
          ))}
        </div>
      </div>
      {stateNote[facilityState(f)] && <p className="text-small text-ink-subtle">{stateNote[facilityState(f)]}</p>}
      <CapacityMeter title={t.detail.declared} used={f.capacity_used} max={f.capacity} unit={f.capacity_unit} />
      <Confirm
        open={confirming} onOpenChange={setConfirming} icon={TriangleAlert} iconClass="bg-warning-soft text-warning-ink"
        title={t.suspendTitle(f.name)} body={t.suspendBody} confirmLabel={t.suspendConfirm} pending={set.isPending}
        onConfirm={() => set.mutate(facilityStatus.suspended)}
      />
    </li>
  );
}

function Detail({ f, mayManage, maySuspend }: { f: Facility; mayManage: boolean; maySuspend: boolean }) {
  const router = useRouter();
  const client = useQueryClient();
  const [confirming, setConfirming] = useState<"suspend" | "delete" | null>(null);
  const set = useMutation({
    mutationFn: (status: string) => endpoints.setFacilityStatus(f.id, status),
    onSuccess: (_r, status) => {
      toast.success(status === facilityStatus.suspended ? t.suspendedToast : t.reinstatedToast);
      client.invalidateQueries({ queryKey: queryKeys.facilities });
    },
    onError: (err) => toast.error(problemOf(err)),
    onSettled: () => setConfirming(null),
  });
  const remove = useMutation({
    mutationFn: () => endpoints.deleteFacility(f.id),
    onSuccess: () => {
      toast.success(t.detail.deleted);
      client.invalidateQueries({ queryKey: queryKeys.facilities });
      router.replace(routes.facilities);
    },
    onError: (err) => toast.error(problemOf(err)),
    onSettled: () => setConfirming(null),
  });
  const approve = useMutation({
    mutationFn: () => endpoints.approveFacility(f.id),
    onSuccess: () => { toast.success(t.approved(f.name)); client.invalidateQueries({ queryKey: queryKeys.facilities }); },
    onError: (err) => toast.error(problemOf(err)),
  });
  const licence = useMutation({
    mutationFn: (file: File) => endpoints.uploadLicence(f.id, file),
    onSuccess: () => { toast.success(t.detail.licenceUploaded); client.invalidateQueries({ queryKey: queryKeys.facilities }); },
    onError: (err) => toast.error(problemOf(err)),
  });
  const Icon = kindIcon[f.kind] ?? MapPin;
  const active = f.status === facilityStatus.active;
  const pending = f.status === facilityStatus.pending;
  const coords = coordsOf(f);
  const note = stateNote[facilityState(f)];

  return (
    <section aria-label={f.name} className="grid min-w-0 content-start gap-6 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
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
            href={routes.facilities} aria-label={closeDetail}
            className="grid size-11 place-items-center rounded-md text-ink-muted hover:bg-surface-sunken"
          >
            <X className="size-5" aria-hidden />
          </Link>
        </div>
      </header>

      <KeyValue facts={[
        { label: t.new.kind, value: kindLabel(f.kind) },
        { label: t.detail.licence, value: f.licence_number, mono: true },
        { label: t.detail.licenceValid, value: f.licence_valid_to ? formatDate(f.licence_valid_to) : t.detail.licenceUnknown },
        { label: areaLabel, value: t.detail.area(f.area_ha) },
        { label: t.detail.declared, value: t.detail.declaredOf(f.capacity, f.capacity_unit, periodLabel(f.capacity_period)) },
        { label: issuedLabel, value: t.detail.issued(f.capacity_used, f.capacity_unit, periodLabel(f.capacity_period)) },
        coords
          ? { label: t.detail.coordinates, value: `${coords[1]}, ${coords[0]}`, mono: true }
          : { label: t.detail.coordinates, value: noPoint },
        { label: registeredLabel, value: formatDate(f.created_at) },
      ]} />

      {f.location && (
        <MapView label={t.detail.mapLabel(f.name)} shapes={shapeOf(f)} selected={f.id} className="h-64 overflow-hidden rounded-md border border-border" />
      )}

      <CapacityMeter title={t.detail.declared} used={f.capacity_used} max={f.capacity} unit={f.capacity_unit} />
      {note && <Notice kind={pending ? "info" : "warning"}>{note}</Notice>}

      <div className="flex flex-wrap items-center justify-between gap-3">
        <span className="grid gap-0.5">
          <span className="text-sm font-semibold">{t.detail.licenceDoc}</span>
          {f.licence_document
            ? <a href={facilityLicenceURL(f.id)} target="_blank" rel="noreferrer" className="flex min-h-tap items-center gap-1.5 text-sm font-semibold text-primary underline">
                <FileText className="size-4" aria-hidden />{t.detail.licenceOpen}
              </a>
            : <span className="text-small text-ink-muted">{t.detail.licenceNone}</span>}
        </span>
        {mayManage && (
          <label className={cn(buttonVariants({ variant: "secondary" }), "cursor-pointer", licence.isPending && "pointer-events-none opacity-50")}>
            <Upload aria-hidden />
            {t.detail.licenceUpload}
            <input type="file" accept={upload.types.join(",")} className="sr-only"
              onChange={(e) => { const file = e.target.files?.[0]; if (file) licence.mutate(file); e.target.value = ""; }} />
          </label>
        )}
      </div>

      {(mayManage || maySuspend) && (
        <div className="flex flex-wrap justify-end gap-2 border-t border-border pt-4">
          {mayManage && (
            <Button variant="danger-outline" disabled={remove.isPending} onClick={() => setConfirming("delete")}>
              <Trash2 aria-hidden />
              {t.detail.delete}
            </Button>
          )}
          {maySuspend && pending && (
            <Button disabled={approve.isPending} onClick={() => approve.mutate()}>
              <CircleCheck aria-hidden />
              {t.approve}
            </Button>
          )}
          {maySuspend && !pending && (active ? (
            <Button variant="secondary" disabled={set.isPending} onClick={() => setConfirming("suspend")}>{t.suspend}</Button>
          ) : (
            <Button variant="secondary" disabled={set.isPending} onClick={() => set.mutate(facilityStatus.active)}>{t.reinstate}</Button>
          ))}
        </div>
      )}

      <Confirm
        open={confirming === "suspend"} onOpenChange={(open) => { if (!open) setConfirming(null); }}
        icon={TriangleAlert} iconClass="bg-warning-soft text-warning-ink"
        title={t.suspendTitle(f.name)} body={t.suspendBody} confirmLabel={t.suspendConfirm} pending={set.isPending}
        onConfirm={() => set.mutate(facilityStatus.suspended)}
      />
      <Confirm
        danger open={confirming === "delete"} onOpenChange={(open) => { if (!open) setConfirming(null); }}
        icon={Trash2} iconClass="bg-danger-soft text-danger"
        title={t.detail.deleteTitle(f.name)} body={t.detail.deleteBody} confirmLabel={t.detail.deleteConfirm} pending={remove.isPending}
        onConfirm={() => remove.mutate()}
      />
    </section>
  );
}

/** The confirmation a suspend or delete always asks for (patterns.md → Destructive actions). */
function Confirm({ open, onOpenChange, icon: Icon, iconClass, title, body, confirmLabel, pending, danger, onConfirm }: {
  open: boolean; onOpenChange: (open: boolean) => void; icon: LucideIcon; iconClass: string;
  title: string; body: string; confirmLabel: string; pending: boolean; danger?: boolean; onConfirm: () => void;
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent showCloseButton={false} className="gap-0 overflow-hidden rounded-lg border border-border bg-surface p-0 shadow-overlay">
        <div className="flex gap-3 p-6 pb-0">
          <span className={cn("grid size-10 shrink-0 place-items-center rounded-md", iconClass)}>
            <Icon className="size-5" aria-hidden />
          </span>
          <DialogTitle className="text-[19px] leading-snug font-semibold">{title}</DialogTitle>
        </div>
        <p className="p-6 text-sm text-ink-muted">{body}</p>
        <div className="flex flex-col-reverse gap-2 border-t bg-surface-sunken p-4 sm:flex-row sm:justify-end">
          <DialogClose render={<Button variant="secondary" />}>{cancel}</DialogClose>
          <Button variant={danger ? "danger" : "primary"} disabled={pending} onClick={onConfirm}>{confirmLabel}</Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
