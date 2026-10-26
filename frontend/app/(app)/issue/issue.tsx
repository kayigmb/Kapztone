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
import { CapacityMeter } from "@/components/chaintrack/capacity-meter";
import { ChecksList } from "@/components/chaintrack/checks-list";
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
  type Template,
} from "@/lib/api";
import {
  certificateStatus,
  copy,
  facilityStatus,
  passportStatus,
  queryKeys,
  routes,
  type PassportStatus,
  screenCopy,
} from "@/lib/consts";
import { liveValidation } from "@/lib/form";
import {
  capacityLeft,
  issueBlock,
  issueFormSchema,
  type IssueForm,
  type IssueFormInput,
} from "@/lib/schemas/issue";
import { cn } from "@/lib/utils";
import { formatQuantity, humanize, type Verdict } from "@/lib/verification";

const t = copy.issue;

const local = screenCopy.issue;

const kindIcon: Record<string, LucideIcon> = {
  farm_plot: Sprout,
  mine_site: Pickaxe,
  washing_station: Factory,
  processing_plant: Factory,
  warehouse: Warehouse,
};
const periodOf = (f: Facility) => f.capacity_period;

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
    trigger,
    control,
    formState: { errors },
  } = useForm<IssueFormInput, unknown, IssueForm>({
    ...liveValidation,
    resolver: zodResolver(issueFormSchema),
    defaultValues: { facility_id: "", template_id: "", quantity: undefined },
  });

  // Both pickers are searched and paged by the API; what's picked is read on its own, whichever page it was on.
  const [tq, setTq] = useState<Picking>({ page: 1, search: "" });
  const [fq, setFq] = useState<Picking>({ page: 1, search: "" });
  const templates = usePaged(
    [...queryKeys.templates, "picker"],
    endpoints.templatesPage,
    { ...tq, size: pickerSize, sort: "name" },
  );
  const facilities = usePaged(
    [...queryKeys.facilities, "picker"],
    endpoints.facilitiesPage,
    { ...fq, size: pickerSize, sort: "name" },
  );

  const facilityId = useWatch({ control, name: "facility_id" });
  const templateId = useWatch({ control, name: "template_id" });
  const quantity = Number(useWatch({ control, name: "quantity" })) || 0;
  const facility = useQuery({
    queryKey: [...queryKeys.facilities, facilityId],
    queryFn: () => endpoints.facility(facilityId),
    enabled: !!facilityId,
  }).data;
  const template = useQuery({
    queryKey: queryKeys.template(templateId),
    queryFn: () => endpoints.template(templateId),
    enabled: !!templateId,
  }).data;

  const evidence = useQuery({
    queryKey: [
      ...queryKeys.certificates,
      "facility",
      facilityId,
      certificateStatus.valid,
    ],
    queryFn: () =>
      endpoints.attestationsList(facilityId, certificateStatus.valid),
    enabled: !!facilityId,
  });

  const required =
    template?.requiredEvidence ?? [];
  const onFile = new Set((evidence.data ?? []).map((c) => c.type));
  const missing = required.filter((type) => !onFile.has(type));
  const checks = required.map((type) => ({
    name: type,
    label: humanize(type),
    result: (onFile.has(type) ? "verified" : "failed") as Verdict,
    explanation: onFile.has(type) ? t.evidenceOnFile : local.evidenceRowMissing,
  }));

  const unit = template?.unit ?? facility?.capacity_unit ?? "";
  const block = facility ? issueBlock(facility, quantity) : null;
  const cantSign = !!block || missing.length > 0;

  const issue = useMutation({
    mutationFn: ({ pin, v }: { pin: string; v: IssueForm }) =>
      endpoints.issuePassport(
        {
          facility_id: v.facility_id,
          template_id: v.template_id,
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

  const facts: Fact[] =
    template && facility
      ? [
          { label: t.facts.product, value: humanize(template.commodity) },
          { label: t.facts.template, value: template.name },
          { label: t.facts.facility, value: facility.name },
          {
            label: t.facts.quantity,
            value: formatQuantity(quantity, template.unit),
            mono: true,
          },
          {
            label: t.facts.evidence,
            value: local.evidenceCount(
              required.length - missing.length,
              required.length,
            ),
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
    if (await trigger(["template_id", "facility_id"])) setStep(1);
  };
  const toSign = async () => {
    if (await trigger()) setStep(2);
  };

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <PageHeader title={t.title}>
        <Stepper
          steps={t.steps}
          current={step}
          error={step === 1 && cantSign}
        />
      </PageHeader>

      <DataWidget query={facilities}>
        {() => (
          <DataWidget query={templates}>
            {() => (
              <>
                {step === 0 && (
                  <section className="grid gap-6">
                    <div className="grid gap-3">
                      <h2 className="text-2xl font-semibold">{t.where}</h2>
                      <TemplatePicker
                        page={templates.data!}
                        picking={tq}
                        onPicking={setTq}
                        error={errors.template_id?.message}
                        {...register("template_id")}
                      />
                      <p className="text-small text-ink-subtle">
                        {t.templateHelp}
                      </p>
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

                {step === 1 && facility && template && (
                  <section className="grid gap-6">
                    <div className="grid gap-3 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
                      <h2 className="text-section">{t.howMuch}</h2>
                      <TextField
                        label={`${t.quantity} (${unit})`}
                        mono
                        type="number"
                        inputMode="numeric"
                        min={1}
                        step={1}
                        help={`${humanize(template.commodity)} · ${facility.name}`}
                        error={
                          errors.quantity?.message ??
                          (block
                            ? t.blockedField(
                                block.left,
                                facility.capacity_unit,
                                periodOf(facility),
                              )
                            : undefined)
                        }
                        {...register("quantity")}
                      />
                      <CapacityMeter
                        title={facility.name}
                        used={facility.capacity_used}
                        max={facility.capacity}
                        unit={facility.capacity_unit}
                      />
                      {quantity > 0 && !block && (
                        <p className="text-small font-semibold text-ink-muted">
                          {local.afterPassport(
                            facility.capacity_used + quantity,
                            facility.capacity_unit,
                          )}
                        </p>
                      )}
                      {block && (
                        <Notice kind="blocking">
                          {t.blockedNotice(block.left, facility.capacity_unit)}
                        </Notice>
                      )}
                    </div>

                    <div className="grid gap-3 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
                      <h2 className="text-section">{t.evidence}</h2>
                      <p className="text-small text-ink-subtle">
                        {t.evidenceHelp}
                      </p>
                      {!facilityId ? (
                        required.length === 0 ? (
                          <p className="text-sm text-ink-muted">
                            {t.evidenceNone}
                          </p>
                        ) : (
                          <ChecksList checks={checks} />
                        )
                      ) : (
                        <DataWidget query={evidence}>
                          {() =>
                            required.length === 0 ? (
                              <p className="text-sm text-ink-muted">
                                {t.evidenceNone}
                              </p>
                            ) : (
                              <ChecksList checks={checks} />
                            )
                          }
                        </DataWidget>
                      )}
                      {required.length > 0 && missing.length > 0 && (
                        <Notice kind="blocking">
                          {t.evidenceMissing(missing.map(humanize).join(", "))}
                        </Notice>
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
                      {cantSign && (
                        <p className="text-small font-semibold text-danger">
                          {t.blockedReason}
                        </p>
                      )}
                      <Button onClick={toSign} disabled={cantSign}>
                        {t.review}
                      </Button>
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
                      blocked={block ? t.blockedReason : undefined}
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

/** Radio rows for the commodity template: name, commodity and unit. */
function TemplatePicker({
  page,
  picking,
  onPicking,
  error,
  ...radio
}: React.ComponentProps<"input"> & {
  page: Paged<Template>;
  picking: Picking;
  onPicking: (p: Picking) => void;
  error?: string;
}) {
  if (page.total === 0 && !picking.search)
    return <Notice kind="info">{t.templateEmpty}</Notice>;
  return (
    <PickerFrame
      legend={t.template}
      picking={picking}
      onPicking={onPicking}
      total={page.total}
      error={error}
    >
      {page.data.length === 0 ? (
        <p className="text-sm text-ink-muted">{copy.list.noMatch}</p>
      ) : (
        <div className="grid gap-2 md:grid-cols-2">
          {page.data.map((m) => (
            <label key={m.id} className={radioRow}>
              <span className="grid min-w-0 gap-1">
                <span className="truncate text-sm font-semibold">{m.name}</span>
                <span className="truncate text-xs text-ink-muted">
                  {humanize(m.commodity)} · {m.unit}
                </span>
              </span>
              <input
                type="radio"
                value={m.id}
                className="peer sr-only"
                {...radio}
              />
              <span className={radioDot} aria-hidden />
            </label>
          ))}
        </div>
      )}
    </PickerFrame>
  );
}

/** The organisation's own facilities: active ones selectable with what's left, pending or suspended ones disabled with the reason. */
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
            const free = capacityLeft(f);
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
                        : free > 0
                          ? t.facilityLeft(free, f.capacity_unit, periodOf(f))
                          : t.facilityNone}
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
