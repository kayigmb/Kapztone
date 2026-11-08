"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ArrowLeft,
  CircleCheck,
  Clock,
  Factory,
  MapPin,
  OctagonX,
  Pickaxe,
  Sprout,
  Warehouse,
  type LucideIcon,
} from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { useForm, useWatch } from "react-hook-form";

import { DataWidget } from "@/components/chaintrack/data-widget";
import { FieldNote } from "@/components/chaintrack/field-note";
import { PageHeader } from "@/components/chaintrack/page-header";
import { SelectInput } from "@/components/chaintrack/select-input";
import { ListToolbar, Pager } from "@/components/chaintrack/list-controls";
import { Notice } from "@/components/chaintrack/notice";
import { KeyValue, type Fact } from "@/components/chaintrack/passport-card";
import { SigningPanel } from "@/components/chaintrack/signing";
import { StatusChip } from "@/components/chaintrack/status";
import { Stepper } from "@/components/chaintrack/stepper";
import { TextField } from "@/components/chaintrack/text-field";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  endpoints,
  problemOf,
  usePaged,
  type Facility,
  type Paged,
  type PassportDetail,
} from "@/lib/api";
import {
  copy,
  facilityStatus,
  passportStatus,
  passportUnits,
  queryKeys,
  routes,
  type PassportStatus,
  screenCopy,
} from "@/lib/consts";
import { liveValidation } from "@/lib/form";
import {
  issueFormSchema,
  type IssueForm,
  type IssueFormInput,
} from "@/lib/schemas/issue";
import { cn } from "@/lib/utils";
import { formatQuantity, humanize } from "@/lib/verification";

const t = copy.issue;
/** The unit select's choice for a unit of one's own. */
const otherUnit = "other";

const local = screenCopy.issue;

const kindIcon: Record<string, LucideIcon> = {
  farm_plot: Sprout,
  mine_site: Pickaxe,
  washing_station: Factory,
  processing_plant: Factory,
  warehouse: Warehouse,
};

export function Issue() {
  return <Flow />;
}

function Flow() {
  const client = useQueryClient();
  const [step, setStep] = useState(0); // 0 facility · 1 quantity · 2 sign · 3 done
  const [done, setDone] = useState<{
    passport: PassportDetail;
    facilityName: string;
  }>();
  const {
    register,
    getValues,
    setValue,
    trigger,
    control,
    formState: { errors },
  } = useForm<IssueFormInput, unknown, IssueForm>({
    ...liveValidation,
    resolver: zodResolver(issueFormSchema),
    defaultValues: {
      facility_id: "",
      commodity: "",
      unit: "kg",
      quantity: undefined,
    },
  });

  // The picker is searched and paged by the API; what's picked is read on its own, whichever page it was on.
  const [fq, setFq] = useState<Picking>({ page: 1, search: "" });
  const facilities = usePaged(
    [...queryKeys.facilities, "picker"],
    endpoints.facilitiesPage,
    { ...fq, size: pickerSize, sort: "name" },
  );

  const facilityId = useWatch({ control, name: "facility_id" });
  const commodity = useWatch({ control, name: "commodity" }) ?? "";
  const unit = useWatch({ control, name: "unit" }) ?? "";
  // The unit select: a common unit, or "other" with the producer's own typed in.
  const [unitChoice, setUnitChoice] = useState<string>(passportUnits[0]);
  const [ownUnit, setOwnUnit] = useState("");
  const pickUnit = (choice: string, typed = ownUnit) => {
    setUnitChoice(choice);
    setOwnUnit(typed);
    setValue("unit", choice === otherUnit ? typed : choice, {
      shouldValidate: choice !== otherUnit || typed !== "",
    });
  };
  const quantity = Number(useWatch({ control, name: "quantity" })) || 0;
  const facility = useQuery({
    queryKey: [...queryKeys.facilities, facilityId],
    queryFn: () => endpoints.facility(facilityId),
    enabled: !!facilityId,
  }).data;

  const issue = useMutation({
    mutationFn: ({ pin, v }: { pin: string; v: IssueForm }) =>
      endpoints.issuePassport(
        {
          facility_id: v.facility_id,
          commodity: v.commodity,
          unit: v.unit,
          quantity: Number(v.quantity),
        },
        pin,
      ),
    meta: { inPlace: true },
    onSuccess: ({ data: passport }, { v }) => {
      setDone({
        passport,
        facilityName: facility?.id === v.facility_id ? facility.name : "",
      });
      setStep(3);
      client.invalidateQueries({ queryKey: queryKeys.facilities });
      client.invalidateQueries({ queryKey: queryKeys.passports });
    },
  });

  const facts: Fact[] = facility
    ? [
        { label: t.facts.product, value: commodity.trim() },
        { label: t.facts.facility, value: facility.name },
        {
          label: t.facts.quantity,
          value: formatQuantity(quantity, unit),
          mono: true,
        },
      ]
    : [];

  if (step === 3 && done)
    return (
      <Created
        done={done}
        onAgain={() => {
          setDone(undefined);
          setStep(0);
        }}
      />
    );

  const toQuantity = async () => {
    if (await trigger(["commodity", "facility_id"])) setStep(1);
  };
  const toSign = async () => {
    if (await trigger()) setStep(2);
  };

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <PageHeader title={t.title}>
        <Stepper steps={t.steps} current={step} />
      </PageHeader>

      <DataWidget query={facilities}>
        {() => (
          <>
            {step === 0 && (
              <section className="grid gap-6">
                <div className="grid gap-3">
                  <h2 className="text-2xl font-semibold">{t.where}</h2>
                  <TextField
                    label={t.commodity}
                    help={t.commodityHelp}
                    error={errors.commodity?.message}
                    maxLength={50}
                    {...register("commodity")}
                  />
                </div>
                <FacilityPicker
                  page={facilities.data!}
                  picking={fq}
                  onPicking={setFq}
                  error={errors.facility_id?.message}
                  {...register("facility_id")}
                />
                <div className="flex flex-wrap items-center justify-end gap-3">
                  <Button onClick={toQuantity}>{local.next}</Button>
                </div>
              </section>
            )}

            {step === 1 && facility && (
              <section className="grid gap-6">
                <div className="grid gap-3 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
                  <h2 className="text-section">{t.howMuch}</h2>
                  <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_10rem]">
                    <TextField
                      label={t.quantity}
                      mono
                      type="number"
                      inputMode="numeric"
                      min={1}
                      step={1}
                      help={`${commodity.trim()} · ${facility.name}`}
                      error={errors.quantity?.message}
                      {...register("quantity")}
                    />
                    <div className="grid content-start gap-1.5">
                      <span className="text-sm font-semibold">{t.unit}</span>
                      <SelectInput
                        label={t.unit}
                        value={unitChoice}
                        onChange={(v) => pickUnit(v)}
                        options={[
                          ...passportUnits.map((u) => ({ value: u, label: u })),
                          { value: otherUnit, label: t.unitOther },
                        ]}
                      />
                      {unitChoice !== otherUnit && (
                        <FieldNote error={errors.unit?.message} />
                      )}
                    </div>
                  </div>
                  {unitChoice === otherUnit && (
                    <TextField
                      label={t.unitOwn}
                      help={t.unitOwnHelp}
                      error={errors.unit?.message}
                      value={ownUnit}
                      maxLength={16}
                      onChange={(e) => pickUnit(otherUnit, e.target.value)}
                    />
                  )}
                </div>

                <div className="grid gap-3 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
                  <h2 className="text-section">{t.review}</h2>
                  <KeyValue facts={facts} />
                  <p className="text-sm text-ink-muted">{t.meaning}</p>
                </div>

                <div className="flex flex-wrap items-center justify-end gap-3">
                  <Button variant="secondary" onClick={() => setStep(0)}>
                    <ArrowLeft aria-hidden />
                    {local.back}
                  </Button>
                  <Button onClick={toSign}>{t.review}</Button>
                </div>
              </section>
            )}

            {step === 2 && (
              <section className="grid gap-4">
                <Button
                  variant="ghost"
                  onClick={() => setStep(1)}
                  className="justify-self-start"
                >
                  <ArrowLeft aria-hidden />
                  {local.back}
                </Button>
                <SigningPanel
                  facts={facts}
                  meaning={t.meaning}
                  button={t.sign}
                  busy={issue.isPending}
                  error={issue.isError ? problemOf(issue.error) : undefined}
                  onSign={(pin) =>
                    issue.mutate({
                      pin,
                      v: issueFormSchema.parse(getValues()),
                    })
                  }
                />
              </section>
            )}
          </>
        )}
      </DataWidget>
    </main>
  );
}

type Picking = { page: number; search: string };
const pickerSize = 6;
const radioRow =
  "grid cursor-pointer grid-cols-[1fr_auto] items-center gap-x-3 rounded-md border border-border-strong bg-surface px-3 py-2.5 has-checked:border-primary has-checked:bg-primary-soft has-focus-visible:ring-3 has-focus-visible:ring-ring/50";
const radioDot =
  "size-4.5 rounded-full border-2 border-border-strong peer-checked:border-6 peer-checked:border-primary";

/** A picker's search box and pager (the API does the searching). */
function PickerFrame({
  legend,
  picking,
  onPicking,
  total,
  error,
  children,
}: {
  legend: string;
  picking: Picking;
  onPicking: (p: Picking) => void;
  total: number;
  error?: string;
  children: React.ReactNode;
}) {
  return (
    <fieldset className="grid gap-2" aria-invalid={error ? true : undefined}>
      <legend className="mb-1 text-sm font-semibold">{legend}</legend>
      <ListToolbar
        search={picking.search}
        onSearch={(search) => onPicking({ page: 1, search })}
      />
      {children}
      <Pager
        page={picking.page}
        size={pickerSize}
        total={total}
        onPage={(page) => onPicking({ ...picking, page })}
      />
      <FieldNote error={error} />
    </fieldset>
  );
}

/** The organisation's own facilities: active ones selectable, pending or suspended ones disabled with the reason. */
function FacilityPicker({
  page,
  picking,
  onPicking,
  error,
  ...radio
}: React.ComponentProps<"input"> & {
  page: Paged<Facility>;
  picking: Picking;
  onPicking: (p: Picking) => void;
  error?: string;
}) {
  if (page.total === 0 && !picking.search) {
    return (
      <div className="grid gap-1.5">
        <Notice kind="info">
          {t.facilityEmpty}{" "}
          <Link href={routes.facilityNew} className="font-semibold underline">
            {copy.facilities.register}
          </Link>
        </Notice>
        <FieldNote error={error} />
      </div>
    );
  }
  return (
    <PickerFrame
      legend={t.facility}
      picking={picking}
      onPicking={onPicking}
      total={page.total}
      error={error}
    >
      {page.data.length === 0 ? (
        <p className="text-sm text-ink-muted">{copy.list.noMatch}</p>
      ) : (
        <div className="grid gap-2 md:grid-cols-2">
          {page.data.map((f) => {
            const Icon = kindIcon[f.kind] ?? MapPin;
            const active = f.status === facilityStatus.active;
            return (
              <label
                key={f.id}
                className={cn(
                  radioRow,
                  !active &&
                    "cursor-not-allowed border-dashed bg-surface-sunken",
                )}
              >
                <span className="flex min-w-0 items-center gap-3">
                  <span className="grid size-8 shrink-0 place-items-center rounded-md bg-primary-soft text-primary">
                    <Icon className="size-4" aria-hidden />
                  </span>
                  <span className="grid min-w-0 gap-0.5">
                    <span
                      className={cn(
                        "truncate text-sm font-semibold",
                        !active && "text-ink-muted",
                      )}
                    >
                      {f.name}
                    </span>
                    <span className="truncate text-xs text-ink-muted">
                      {!active
                        ? f.status === facilityStatus.pending
                          ? t.facilityPending
                          : t.facilitySuspended
                        : t.facilityReady}
                    </span>
                  </span>
                </span>
                <input
                  type="radio"
                  value={f.id}
                  disabled={!active}
                  className="peer sr-only"
                  {...radio}
                />
                {active ? (
                  <span className={radioDot} aria-hidden />
                ) : (
                  <OctagonX className="size-4 text-ink-subtle" aria-hidden />
                )}
              </label>
            );
          })}
        </div>
      )}
    </PickerFrame>
  );
}

/** The signed record, with what happens local.next. */
function Created({
  done,
  onAgain,
}: {
  done: { passport: PassportDetail; facilityName: string };
  onAgain: () => void;
}) {
  const p = done.passport;
  const active = p.status === passportStatus.active;
  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <section className="grid justify-items-center gap-3 py-6 text-center">
        <span className="grid size-18 place-items-center rounded-full bg-verified-soft text-verified">
          <CircleCheck className="size-8" aria-hidden />
        </span>
        <h1 className="text-2xl font-semibold lg:text-page">
          {t.createdTitle}
        </h1>
        <p className="text-ink-muted">
          {t.createdBody(formatQuantity(p.quantity, p.unit), done.facilityName)}
        </p>
        <p className="code text-ink">{p.code}</p>
        <StatusChip
          tone={active ? "verified" : "pending"}
          icon={active ? CircleCheck : Clock}
        >
          {copy.passportView.status[p.status as PassportStatus] ??
            humanize(p.status)}
        </StatusChip>
      </section>

      <section className="grid gap-3 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
        <h2 className="text-section">{local.whatNext}</h2>
        <ol className="grid gap-2">
          {local.nextSteps.map((s) => (
            <li key={s} className="flex gap-2 text-sm text-ink-muted">
              <CircleCheck
                className="mt-0.5 size-4 shrink-0 text-verified"
                aria-hidden
              />
              {s}
            </li>
          ))}
        </ol>
        <p className="text-small text-ink-subtle">{t.labelsNote}</p>
      </section>

      <div className="flex flex-wrap justify-center gap-2">
        <Link href={routes.passportDetail(p.id)} className={buttonVariants()}>
          {t.viewPassport}
        </Link>
        <Link
          href={routes.handover(p.id)}
          className={buttonVariants({ variant: "secondary" })}
        >
          {t.handOver}
        </Link>
        <Button variant="secondary" onClick={onAgain}>
          {t.another}
        </Button>
      </div>
    </main>
  );
}
