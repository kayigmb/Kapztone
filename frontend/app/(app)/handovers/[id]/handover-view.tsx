"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ArrowLeft,
  CircleCheck,
  Clock,
  ImageIcon,
  MapPin,
  OctagonX,
  TriangleAlert,
  WifiOff,
} from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { useForm, useWatch } from "react-hook-form";

import { HandoverStatusChip } from "@/components/chaintrack/handover-status";
import { FieldNote } from "@/components/chaintrack/field-note";
import { Notice } from "@/components/chaintrack/notice";
import { Loading, PageState } from "@/components/chaintrack/page-state";
import { KeyValue } from "@/components/chaintrack/passport-card";
import {
  PlaceAndPhotos,
  type Place,
} from "@/components/chaintrack/place-and-photos";
import { SigningPanel } from "@/components/chaintrack/signing";
import { StatusChip } from "@/components/chaintrack/status";
import { TextField } from "@/components/chaintrack/text-field";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  endpoints,
  handoverPhotoURL,
  problemOf,
  useMe,
  type HandoverDetail,
} from "@/lib/api";
import {
  copy,
  handoverStatus,
  queryKeys,
  rejectReasons,
  routes,
  type RejectReason,
} from "@/lib/consts";
import { liveValidation } from "@/lib/form";
import { toleranceOf, weigh } from "@/lib/handover";
import {
  receiveFormSchema,
  rejectFormSchema,
  type ReceiveForm,
  type ReceiveFormInput,
  type RejectForm,
} from "@/lib/schemas/handover";
import { cn } from "@/lib/utils";
import { formatDate, formatQuantity, humanize } from "@/lib/verification";

const t = copy.receiver;
const h = copy.handover;

export function HandoverView({ id }: { id: string }) {
  const me = useMe().data!;
  const handover = useQuery({
    queryKey: queryKeys.handover(id),
    queryFn: () => endpoints.handover(id),
  });
  const hv = handover.data;
  // The receiver isn't the holder yet, so the lot's code and unit come from its public page.
  const lot = useQuery({
    queryKey: queryKeys.verify(hv?.passport_id ?? ""),
    queryFn: () => endpoints.verify(hv!.passport_id),
    enabled: !!hv,
  });
  const templates = useQuery({
    queryKey: queryKeys.templates,
    queryFn: endpoints.templatesList,
    enabled: !!hv,
  });
  const names = useQuery({
    queryKey: queryKeys.directory,
    queryFn: endpoints.directory,
  });
  const [mode, setMode] = useState<"accept" | "reject">("accept");

  if (handover.isError) {
    return (
      <PageState
        icon={WifiOff}
        tone="danger"
        title={copy.errors.loadTitle}
        retry={() => handover.refetch()}
      >
        {problemOf(handover.error)}
      </PageState>
    );
  }
  if (!hv) return <Loading className="p-4 lg:p-8" />;

  const p = lot.data?.passport;
  const unit = p?.unit ?? "";
  const stages =
    templates.data?.find((x) => x.commodity === p?.commodity)?.stages ?? [];
  const tolerance = toleranceOf(stages, hv.stage);
  const mine = me.organisation?.id;
  const sender = names.data?.get(hv.from_org) ?? copy.dashboard.someOrg;
  const receiver = names.data?.get(hv.to_org) ?? copy.dashboard.someOrg;
  const open = hv.status === handoverStatus.awaiting;
  const iReceive = open && hv.to_org === mine;
  const iSent = open && hv.from_org === mine;

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <Link
        href={routes.handovers}
        className="inline-flex min-h-tap w-fit items-center gap-1.5 text-sm font-semibold text-ink-muted hover:text-ink"
      >
        <ArrowLeft className="size-4" aria-hidden />
        {t.backToList}
      </Link>
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div className="grid gap-1">
          <h1 className="text-2xl font-semibold lg:text-page">
            {iReceive ? copy.handover.title : t.title}
          </h1>
          <p className="text-sm text-ink-muted">
            {hv.to_org === mine ? t.from(sender) : t.to(receiver)}
          </p>
        </div>
        <HandoverStatusChip h={hv} mine={mine} />
      </header>

      <section className="grid gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
        <div className="flex flex-wrap gap-2">
          <StatusChip tone="verified" icon={CircleCheck}>
            {h.signedBySender}
          </StatusChip>
          {open && (
            <StatusChip tone="pending" icon={Clock}>
              {copy.dashboard.tiles.incoming.note}
            </StatusChip>
          )}
        </div>
        <KeyValue
          facts={[
            {
              label: h.facts.lot,
              value: p ? (
                <Link
                  className="text-primary"
                  href={routes.passport(p.code ?? p.id)}
                >
                  {p.code ?? p.id}
                </Link>
              ) : (
                "…"
              ),
              mono: true,
            },
            { label: h.facts.product, value: p ? humanize(p.commodity) : "…" },
            {
              label: t.sentQty,
              value: formatQuantity(hv.quantity, unit),
              mono: true,
            },
            { label: h.stage, value: humanize(hv.stage) },
            ...(hv.received_quantity != null
              ? [
                  {
                    label: t.received,
                    value: formatQuantity(hv.received_quantity, unit),
                    mono: true,
                  },
                ]
              : []),
            { label: t.sentOn, value: formatDate(hv.created_at) },
            ...(hv.note
              ? [{ label: h.facts.where, value: hv.note, wide: true }]
              : []),
          ]}
        />
        {hv.split_from_passport_id && (
          <p className="text-small text-ink-muted">{t.partOf(p?.code ?? "")}</p>
        )}
        {hv.location && (
          <p className="flex items-center gap-1.5 text-small text-ink-muted">
            <MapPin className="size-4" aria-hidden />
            {hv.location.lat.toFixed(5)}, {hv.location.lon.toFixed(5)}
          </p>
        )}
        {hv.photos.length > 0 && (
          <ul className="flex flex-wrap gap-2">
            {hv.photos.map((ph) => (
              <li key={ph.id}>
                <a
                  href={handoverPhotoURL(hv.id, ph.id)}
                  target="_blank"
                  rel="noreferrer"
                  className="flex min-h-tap items-center gap-1.5 rounded-md bg-surface-sunken px-3 text-small hover:text-primary"
                >
                  <ImageIcon className="size-4" aria-hidden />
                  {ph.file_name}
                </a>
              </li>
            ))}
          </ul>
        )}
        {hv.reject_reasons?.length ? (
          <Notice kind="blocking">
            <strong>{h.rejectedByReceiver}.</strong>{" "}
            {hv.reject_reasons
              .map((r) => t.rejectChips[r as RejectReason] ?? r)
              .join(", ")}
            {hv.reject_reason ? ` — ${hv.reject_reason}` : ""}
          </Notice>
        ) : null}
      </section>

      {iReceive && (
        <>
          <div className="grid grid-cols-2 gap-2" role="tablist">
            {(["accept", "reject"] as const).map((m) => (
              <button
                key={m}
                type="button"
                role="tab"
                aria-selected={mode === m}
                onClick={() => setMode(m)}
                className={cn(
                  "min-h-tap rounded-md border px-3 text-sm font-semibold",
                  mode === m
                    ? m === "accept"
                      ? "border-primary bg-primary-soft text-primary"
                      : "border-danger bg-danger-soft text-danger-ink"
                    : "border-border text-ink-muted",
                )}
              >
                {m === "accept" ? t.accept : t.reject}
              </button>
            ))}
          </div>
          {mode === "accept" ? (
            <Accept hv={hv} unit={unit} tolerance={tolerance} sender={sender} />
          ) : (
            <Reject hv={hv} sender={sender} />
          )}
        </>
      )}
      {hv.status === handoverStatus.completed && hv.to_org === mine && (
        <Complete hv={hv} unit={unit} />
      )}
      {iSent && <Cancel hv={hv} />}
    </main>
  );
}

/** The receiver's confirmation. Shown from the handover's status too: accepting refreshes it, which unmounts the form. */
function Complete({ hv, unit }: { hv: HandoverDetail; unit: string }) {
  return (
    <section className="grid gap-3" role="status">
      <span className="grid size-14 place-items-center rounded-full bg-verified-soft text-verified">
        <CircleCheck className="size-7" aria-hidden />
      </span>
      <h2 className="text-page">{t.completeTitle}</h2>
      <p className="text-sm text-ink-muted">
        {t.completeBody(
          formatQuantity(hv.received_quantity ?? hv.quantity, unit),
        )}
      </p>
      <Link
        href={routes.passportDetail(hv.child_passport_id ?? hv.passport_id)}
        className={buttonVariants({ className: "w-fit" })}
      >
        {t.viewPassport}
      </Link>
    </section>
  );
}

function useDone(hv: HandoverDetail) {
  const client = useQueryClient();
  return () => {
    for (const key of [
      queryKeys.handover(hv.id),
      queryKeys.handovers,
      queryKeys.passports,
      queryKeys.panel("handovers"),
    ]) {
      client.invalidateQueries({ queryKey: key });
    }
  };
}

function Accept({
  hv,
  unit,
  tolerance,
  sender,
}: {
  hv: HandoverDetail;
  unit: string;
  tolerance: number;
  sender: string;
}) {
  const done = useDone(hv);
  const [place, setPlace] = useState<Place>();
  const [photos, setPhotos] = useState<File[]>([]);
  const {
    register,
    control,
    trigger,
    getValues,
    formState: { errors },
  } = useForm<ReceiveFormInput, unknown, ReceiveForm>({
    ...liveValidation,
    resolver: zodResolver(receiveFormSchema),
    defaultValues: { received_quantity: hv.quantity },
  });
  const received =
    Number(useWatch({ control, name: "received_quantity" })) || 0;
  const w = weigh(hv.quantity, received, tolerance);
  const accept = useMutation({
    mutationFn: (pin: string) =>
      endpoints.acceptHandover(
        hv.id,
        {
          received_quantity: Number(getValues("received_quantity")),
          lat: place?.lat,
          lon: place?.lon,
        },
        photos,
        pin,
      ),
    meta: { inPlace: true },
    onSuccess: done,
  });
  if (accept.isSuccess)
    return <Complete hv={{ ...hv, received_quantity: received }} unit={unit} />;
  const fmt = (n: number) => Math.abs(n).toLocaleString();
  return (
    <section className="grid gap-5">
      <TextField
        label={t.received}
        mono
        inputMode="numeric"
        help={t.receivedHelp}
        error={errors.received_quantity?.message}
        {...register("received_quantity", {
          onChange: () => trigger("received_quantity"),
        })}
      />
      {received > 0 && (
        <Notice kind={w.within ? "success" : "blocking"}>
          {w.within
            ? t.within(
                fmt(w.difference),
                unit,
                String(w.pct),
                String(tolerance),
              )
            : t.outside(
                fmt(w.difference),
                unit,
                String(w.pct),
                String(tolerance),
              )}
        </Notice>
      )}
      <PlaceAndPhotos
        place={place}
        onPlace={setPlace}
        photos={photos}
        onPhotos={setPhotos}
        copy={t}
      />
      <SigningPanel
        meaning={t.meaning(sender)}
        button={t.accept}
        blocked={
          !w.within
            ? t.outside(
                fmt(w.difference),
                unit,
                String(w.pct),
                String(tolerance),
              )
            : errors.received_quantity?.message
        }
        busy={accept.isPending}
        error={accept.isError ? problemOf(accept.error) : undefined}
        onSign={(pin) => accept.mutate(pin)}
      />
    </section>
  );
}

function Reject({ hv, sender }: { hv: HandoverDetail; sender: string }) {
  const done = useDone(hv);
  const {
    register,
    control,
    setValue,
    trigger,
    getValues,
    formState: { errors },
  } = useForm<RejectForm>({
    ...liveValidation,
    resolver: zodResolver(rejectFormSchema),
    defaultValues: { reasons: [], note: "", also_flag: true },
  });
  const picked = useWatch({ control, name: "reasons" });
  const alsoFlag = useWatch({ control, name: "also_flag" });
  const reject = useMutation({
    mutationFn: (pin: string) =>
      endpoints.rejectHandover(hv.id, getValues(), pin),
    meta: { inPlace: true },
    onSuccess: done,
  });
  if (reject.isSuccess) {
    return (
      <section className="grid gap-3" role="status">
        <span className="grid size-14 place-items-center rounded-full bg-danger-soft text-danger">
          <OctagonX className="size-7" aria-hidden />
        </span>
        <h2 className="text-page">{t.rejectedTitle}</h2>
        <p className="text-sm text-ink-muted">{t.rejectedBody(sender)}</p>
      </section>
    );
  }
  const toggle = (r: RejectReason) => {
    setValue(
      "reasons",
      picked.includes(r) ? picked.filter((x) => x !== r) : [...picked, r],
    );
    void trigger("reasons");
  };
  return (
    <section className="grid gap-5">
      <fieldset className="grid gap-2">
        <legend className="mb-1.5 text-sm font-semibold">
          {t.rejectReason}
        </legend>
        <div className="flex flex-wrap gap-2">
          {rejectReasons.map((r) => (
            <button
              key={r}
              type="button"
              aria-pressed={picked.includes(r)}
              onClick={() => toggle(r)}
              className={cn(
                "min-h-tap rounded-full border px-4 text-sm font-semibold",
                picked.includes(r)
                  ? "border-danger bg-danger-soft text-danger-ink"
                  : "border-border text-ink-muted",
              )}
            >
              {t.rejectChips[r]}
            </button>
          ))}
        </div>
        <FieldNote error={errors.reasons?.message} />
      </fieldset>
      <TextField
        label={t.rejectNote}
        help={t.rejectNoteHelp}
        error={errors.note?.message}
        {...register("note")}
      />
      <label className="flex min-h-tap w-fit items-center gap-2.5 text-sm">
        <input
          type="checkbox"
          className="size-5 accent-primary"
          {...register("also_flag")}
        />
        {t.rejectFlag}
      </label>
      <Notice kind="info">{t.rejectStays(sender)}</Notice>
      <SigningPanel
        button={alsoFlag ? t.rejectConfirm : t.rejectOnly}
        blocked={picked.length === 0 ? t.rejectPick : undefined}
        busy={reject.isPending}
        error={reject.isError ? problemOf(reject.error) : undefined}
        onSign={(pin) => reject.mutate(pin)}
      />
    </section>
  );
}

function Cancel({ hv }: { hv: HandoverDetail }) {
  const done = useDone(hv);
  const [confirm, setConfirm] = useState(false);
  const cancel = useMutation({
    mutationFn: () => endpoints.cancelHandover(hv.id),
    onSuccess: done,
  });
  if (!confirm) {
    return (
      <Button
        variant="danger-outline"
        className="w-fit"
        onClick={() => setConfirm(true)}
      >
        {h.cancel}
      </Button>
    );
  }
  return (
    <section
      className="grid gap-3 rounded-lg border border-danger bg-danger-soft p-4"
      role="alert"
    >
      <h2 className="flex items-center gap-2 text-title text-danger-ink">
        <TriangleAlert className="size-5" aria-hidden />
        {h.cancelTitle}
      </h2>
      <p className="text-sm">{h.cancelBody}</p>
      <div className="flex flex-wrap gap-2">
        <Button
          variant="danger"
          onClick={() => cancel.mutate()}
          loading={cancel.isPending}
        >
          {h.cancelConfirm}
        </Button>
        <Button variant="secondary" onClick={() => setConfirm(false)}>
          {copy.scan.close}
        </Button>
      </div>
    </section>
  );
}
