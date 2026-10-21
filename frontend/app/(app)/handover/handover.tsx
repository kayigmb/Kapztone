"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CircleCheck, Clock, Search } from "lucide-react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";
import { useForm, useWatch, Controller } from "react-hook-form";

import { DataWidget } from "@/components/chaintrack/data-widget";
import { FieldNote } from "@/components/chaintrack/field-note";
import { PageHeader } from "@/components/chaintrack/page-header";
import { Notice } from "@/components/chaintrack/notice";

import {
  PlaceAndPhotos,
  type Place,
} from "@/components/chaintrack/place-and-photos";
import { SelectInput } from "@/components/chaintrack/select-input";
import { SigningPanel } from "@/components/chaintrack/signing";
import { StatusChip } from "@/components/chaintrack/status";
import { Stepper } from "@/components/chaintrack/stepper";
import { TextField } from "@/components/chaintrack/text-field";
import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { endpoints, problemOf, useMe, type PassportDetail } from "@/lib/api";
import { copy, passportStatus, queryKeys, routes } from "@/lib/consts";
import { liveValidation } from "@/lib/form";
import { handoverStages, isPartial } from "@/lib/handover";
import {
  handoverFormSchema,
  handoverQuantity,
  type HandoverForm,
  type HandoverFormInput,
} from "@/lib/schemas/handover";
import { cn } from "@/lib/utils";
import { formatQuantity, humanize } from "@/lib/verification";

import { LotScanner } from "../scan/scan";

const t = copy.handover;
type Receiver = { id: string; name: string; kind: string };

export function HandoverFlow() {
  const params = useSearchParams();
  const router = useRouter();
  const me = useMe().data!;
  const [ref, setRef] = useState(params.get("passport") ?? "");
  const [step, setStep] = useState(ref ? 1 : 0);
  const [details, setDetails] = useState<HandoverForm>();
  const [place, setPlace] = useState<Place>();
  const [photos, setPhotos] = useState<File[]>([]);
  const [receiver, setReceiver] = useState<Receiver>();

  const passport = useQuery({
    queryKey: queryKeys.passport(ref),
    queryFn: () => endpoints.passport(ref),
    enabled: !!ref,
  });

  const pick = (id: string) => {
    setRef(id);
    setStep(1);
    router.replace(routes.handover(id));
  };

  if (step === 0 || !ref) return <PickLot onPick={pick} />;

  return (
    <DataWidget
      query={passport}
      className="p-4 lg:p-8"
      page={{ icon: Search, title: copy.errors.loadTitle }}
    >
      {(p) =>
        p.organisation_id !== me.organisation?.id ||
        p.status !== passportStatus.active ? (
          <main className="grid w-full content-start gap-6 p-4 lg:p-8">
            <Notice kind="blocking">{t.notHolder}</Notice>
          </main>
        ) : (
          <main className="grid w-full content-start gap-6 p-4 lg:p-8">
            <PageHeader title={t.title}>
              <Stepper steps={[...t.steps]} current={step} />
            </PageHeader>
            <LotSummary p={p} />
            {step === 1 && (
              <Details
                p={p}
                initial={details}
                place={place}
                onPlace={setPlace}
                photos={photos}
                onPhotos={setPhotos}
                onNext={(d) => {
                  setDetails(d);
                  setStep(2);
                }}
              />
            )}
            {step === 2 && (
              <PickReceiver
                selected={receiver}
                onPick={setReceiver}
                onBack={() => setStep(1)}
                onNext={() => setStep(3)}
              />
            )}
            {step === 3 && details && receiver && (
              <Sign
                p={p}
                details={details}
                receiver={receiver}
                place={place}
                photos={photos}
                onBack={() => setStep(2)}
              />
            )}
          </main>
        )
      }
    </DataWidget>
  );
}

function LotSummary({ p }: { p: PassportDetail }) {
  return (
    <section className="grid grid-cols-[1fr_auto] items-center gap-3 rounded-lg border border-border bg-surface-sunken p-4">
      <div className="grid min-w-0 gap-1">
        <span className="font-mono text-sm font-semibold">{p.code}</span>
        <span className="text-sm text-ink-muted">
          {humanize(p.commodity)} · {formatQuantity(p.remaining, p.unit)}
        </span>
      </div>
      <StatusChip tone="verified" icon={CircleCheck}>
        {t.youHold}
      </StatusChip>
    </section>
  );
}

function PickLot({ onPick }: { onPick: (id: string) => void }) {
  const held = useQuery({
    queryKey: [...queryKeys.passports, "active"],
    queryFn: () => endpoints.passportsList(passportStatus.active),
  });
  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <header className="grid gap-1">
        <h1 className="text-2xl font-semibold lg:text-page">{t.pickTitle}</h1>
        <p className="text-sm text-ink-muted">{t.pickHint}</p>
      </header>
      <LotScanner onCode={onPick} />
      <DataWidget query={held}>
        {(lots) =>
          lots.length ? (
            <ul className="grid gap-2">
              {lots.map((x) => (
                <li key={x.id}>
                  <button
                    type="button"
                    onClick={() => onPick(x.id)}
                    className="grid w-full grid-cols-[1fr_auto] items-center gap-3 rounded-lg border border-border bg-surface px-4 py-3 text-left hover:border-primary"
                  >
                    <span className="grid gap-0.5">
                      <span className="font-mono text-sm font-semibold">
                        {x.code}
                      </span>
                      <span className="text-small text-ink-muted">
                        {humanize(x.commodity)}
                      </span>
                    </span>
                    <span className="font-mono text-sm">
                      {formatQuantity(x.remaining, x.unit)}
                    </span>
                  </button>
                </li>
              ))}
            </ul>
          ) : (
            <Notice kind="info">{t.pickEmpty}</Notice>
          )
        }
      </DataWidget>
    </main>
  );
}

function Details({
  p,
  initial,
  place,
  onPlace,
  photos,
  onPhotos,
  onNext,
}: {
  p: PassportDetail;
  initial?: HandoverForm;
  place: Place;
  onPlace: (p: Place) => void;
  photos: File[];
  onPhotos: (f: File[]) => void;
  onNext: (d: HandoverForm) => void;
}) {
  const template = useQuery({
    queryKey: queryKeys.template(p.template_id ?? ""),
    queryFn: () => endpoints.template(p.template_id!),
    enabled: !!p.template_id,
  });
  const stages = handoverStages(template.data?.stages ?? []);
  const {
    register,
    handleSubmit,
    control,
    setValue,
    formState: { errors },
  } = useForm<HandoverFormInput, unknown, HandoverForm>({
    ...liveValidation,
    resolver: zodResolver(handoverFormSchema),
    defaultValues: initial ?? {
      all: true,
      quantity: undefined,
      stage: "",
      note: "",
    },
  });
  const [all, qty] = useWatch({ control, name: ["all", "quantity"] });
  const sending = all ? p.remaining : Number(qty) || 0;
  const over = sending > p.remaining;

  return (
    <form
      noValidate
      className="grid gap-5"
      onSubmit={handleSubmit((d) => !over && onNext(d))}
    >
      <h2 className="text-title">{t.details}</h2>
      <fieldset className="grid gap-2">
        <legend className="mb-1.5 text-sm font-semibold">{t.quantity}</legend>
        <div className="grid grid-cols-2 gap-2">
          {[true, false].map((v) => (
            <button
              key={String(v)}
              type="button"
              onClick={() => setValue("all", v)}
              aria-pressed={all === v}
              className={cn(
                "min-h-tap rounded-md border px-3 text-sm font-semibold",
                all === v
                  ? "border-primary bg-primary-soft text-primary"
                  : "border-border text-ink-muted",
              )}
            >
              {v ? t.allOfIt : t.partOfIt}
            </button>
          ))}
        </div>
        {!all && (
          <TextField
            label={t.quantity}
            mono
            inputMode="numeric"
            help={t.holds(p.remaining, p.unit)}
            error={
              errors.quantity?.message ??
              (over ? t.holds(p.remaining, p.unit) : undefined)
            }
            {...register("quantity")}
          />
        )}
        {isPartial(sending, p.remaining) && (
          <p className="text-small text-ink-muted">
            {t.partial(formatQuantity(p.remaining - sending, p.unit))}
          </p>
        )}
      </fieldset>

      <label className="grid gap-1.5">
        <span className="text-sm font-semibold">{t.stage}</span>
        {!p.template_id ? (
          <Notice kind="warning">{t.noStages}</Notice>
        ) : (
          <DataWidget query={template}>
            {(tpl) => {
              const pickStages = handoverStages(tpl.stages ?? []);
              return pickStages.length === 0 ? (
                <Notice kind="warning">{t.noStages}</Notice>
              ) : (
                <Controller
                  control={control}
                  name="stage"
                  render={({ field }) => (
                    <SelectInput
                      label={t.stage}
                      placeholder={t.stagePlaceholder}
                      value={field.value ?? ""}
                      onChange={field.onChange}
                      invalid={!!errors.stage}
                      options={pickStages.map((st) => ({
                        value: st.key,
                        label: humanize(st.key),
                      }))}
                    />
                  )}
                />
              );
            }}
          </DataWidget>
        )}
        <FieldNote error={errors.stage?.message} help={t.stageHelp} />
      </label>

      <TextField
        label={t.note}
        help={t.noteHelp}
        error={errors.note?.message}
        {...register("note")}
      />
      <PlaceAndPhotos
        place={place}
        onPlace={onPlace}
        photos={photos}
        onPhotos={onPhotos}
        copy={t}
      />
      <Button type="submit" disabled={stages.length === 0}>
        {t.next}
      </Button>
    </form>
  );
}

function PickReceiver({
  selected,
  onPick,
  onBack,
  onNext,
}: {
  selected?: Receiver;
  onPick: (r: Receiver) => void;
  onBack: () => void;
  onNext: () => void;
}) {
  const [search, setSearch] = useState("");
  const found = useQuery({
    queryKey: [...queryKeys.directory, "search", search],
    queryFn: () => endpoints.directorySearch(search),
  });
  return (
    <section className="grid gap-4">
      <h2 className="text-title">{t.receiver}</h2>
      <label className="grid gap-1.5">
        <span className="text-sm font-semibold">{t.receiverSearch}</span>
        <Input
          type="search"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </label>
      <DataWidget query={found}>
        {(orgs) =>
          orgs.length === 0 ? (
            <Notice kind="info">
              {search ? t.receiverEmpty : t.receiverEmptyAll}
            </Notice>
          ) : (
            <ul
              className="grid gap-2"
              role="radiogroup"
              aria-label={t.receiver}
            >
              {orgs.map((o) => (
                <li key={o.id}>
                  <button
                    type="button"
                    role="radio"
                    aria-checked={selected?.id === o.id}
                    onClick={() => onPick(o)}
                    className={cn(
                      "grid w-full grid-cols-[1fr_auto] items-center gap-3 rounded-lg border px-4 py-3 text-left",
                      selected?.id === o.id
                        ? "border-primary bg-primary-soft"
                        : "border-border bg-surface hover:border-primary",
                    )}
                  >
                    <span className="grid gap-0.5">
                      <span className="text-sm font-semibold">{o.name}</span>
                      <span className="text-small text-ink-muted">
                        {humanize(o.kind)}
                      </span>
                    </span>
                    <StatusChip tone="verified" icon={CircleCheck}>
                      {t.verified}
                    </StatusChip>
                  </button>
                </li>
              ))}
            </ul>
          )
        }
      </DataWidget>
      <div className="flex flex-wrap justify-end gap-2">
        <Button variant="secondary" onClick={onBack}>
          {t.back}
        </Button>
        <Button onClick={onNext} disabled={!selected}>
          {t.review}
        </Button>
      </div>
    </section>
  );
}

function Sign({
  p,
  details,
  receiver,
  place,
  photos,
  onBack,
}: {
  p: PassportDetail;
  details: HandoverForm;
  receiver: Receiver;
  place: Place;
  photos: File[];
  onBack: () => void;
}) {
  const client = useQueryClient();
  const quantity = handoverQuantity(details, p.remaining);
  const send = useMutation({
    mutationFn: (pin: string) =>
      endpoints.startHandover(
        p.id,
        {
          to_org: receiver.id,
          stage: details.stage,
          quantity,
          note: details.note,
          lat: place?.lat,
          lon: place?.lon,
        },
        photos,
        pin,
      ),
    meta: { inPlace: true },
    onSuccess: () => {
      for (const key of [
        queryKeys.passports,
        queryKeys.handovers,
        queryKeys.panel("handovers"),
      ])
        client.invalidateQueries({ queryKey: key });
    },
  });

  if (send.isSuccess) {
    return (
      <section className="grid gap-4" role="status">
        <span className="grid size-14 place-items-center rounded-full bg-pending-soft text-pending">
          <Clock className="size-7" aria-hidden />
        </span>
        <h2 className="text-page">{t.sentTitle}</h2>
        <p className="text-sm text-ink-muted">
          {t.sentBody(formatQuantity(quantity, p.unit), receiver.name)}
        </p>
        <ol className="grid gap-1 text-sm">
          {t.stepsNext.map((s) => (
            <li key={s}>{s}</li>
          ))}
        </ol>
        <div className="flex flex-wrap gap-2">
          <Link
            href={routes.handoverDetail(send.data.data.id)}
            className={buttonVariants({ variant: "secondary" })}
          >
            {copy.receiver.open}
          </Link>
          <Link
            href={routes.handover()}
            className={buttonVariants({ variant: "ghost" })}
          >
            {t.another}
          </Link>
        </div>
      </section>
    );
  }

  return (
    <section className="grid gap-4">
      <Notice kind="info">{t.stayHolder}</Notice>
      <SigningPanel
        facts={[
          { label: t.facts.lot, value: p.code, mono: true },
          { label: t.facts.product, value: humanize(p.commodity) },
          {
            label: t.facts.quantity,
            value: formatQuantity(quantity, p.unit),
            mono: true,
          },
          { label: t.facts.to, value: receiver.name },
          { label: t.stage, value: humanize(details.stage) },
          ...(details.note
            ? [{ label: t.facts.where, value: details.note, wide: true }]
            : []),
        ]}
        meaning={t.meaning(receiver.name)}
        button={t.sign}
        busy={send.isPending}
        error={send.isError ? problemOf(send.error) : undefined}
        onSign={(pin) => send.mutate(pin)}
      />
      <Button
        variant="secondary"
        className="w-fit"
        onClick={onBack}
        loading={send.isPending}
      >
        {t.back}
      </Button>
    </section>
  );
}
