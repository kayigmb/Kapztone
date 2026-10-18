"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ArrowLeft, CircleCheck, Clock, Factory, Layers, Lock, MapPin, OctagonX, Pickaxe, Sprout, Warehouse, type LucideIcon,
} from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { useForm, useWatch } from "react-hook-form";

import { CapacityMeter } from "@/components/chaintrack/capacity-meter";
import { ChecksList } from "@/components/chaintrack/checks-list";
import { ChoiceCards } from "@/components/chaintrack/choice-cards";
import { Notice } from "@/components/chaintrack/notice";
import { Loading, PageState } from "@/components/chaintrack/page-state";
import { KeyValue, type Fact } from "@/components/chaintrack/passport-card";
import { SigningPanel } from "@/components/chaintrack/signing";
import { StatusChip } from "@/components/chaintrack/status";
import { Stepper } from "@/components/chaintrack/stepper";
import { TextField } from "@/components/chaintrack/text-field";
import { Button, buttonVariants } from "@/components/ui/button";
import { endpoints, problemOf, useMe, type Facility, type PassportDetail, type Template } from "@/lib/api";
import {
  certificateStatus, copy, facilityStatus, passportStatus, permissions, queryKeys, routes, templateEvents,
  type PassportStatus,
} from "@/lib/consts";
import { capacityLeft, issueBlock, issueFormSchema, templateStages, type IssueForm, type IssueFormInput } from "@/lib/schemas/issue";
import { formatQuantity, humanize, type Verdict } from "@/lib/verification";

const t = copy.issue;

// Copy that lib/consts.ts doesn't have yet.
const back = "Back";
const next = "Continue";
const whatNext = "What happens next";
const notAllowedTitle = "You can't issue passports.";
const notAllowedBody = "Ask an admin of your organisation for the issue-passports permission.";
const afterPassport = (used: number, unit: string) => `With this passport: ${used.toLocaleString()} ${unit} issued.`;
const evidenceCount = (have: number, total: number) => `${have} of ${total} on file`;
const evidenceRowMissing = "Not on file yet.";
const nextSteps = [
  "Print a QR label for each bag or crate.",
  "Collectors and inspectors add their records against this lot.",
  "Hand the lot over when it leaves this site.",
];

const kindIcon: Record<string, LucideIcon> = {
  farm_plot: Sprout, mine_site: Pickaxe, washing_station: Factory, processing_plant: Factory, warehouse: Warehouse,
};
const stageLabel = (key: string) => templateEvents.find((e) => e.value === key)?.label ?? humanize(key);
const periodOf = (f: Facility) => f.capacity_period;

export function Issue() {
  const me = useMe().data!; // the layout only renders once it is loaded
  if (!me.permissions.includes(permissions.passportIssue)) {
    return <PageState icon={Lock} title={notAllowedTitle}>{notAllowedBody}</PageState>;
  }
  return <Flow />;
}

function Flow() {
  const client = useQueryClient();
  const [step, setStep] = useState(0); // 0 facility · 1 quantity · 2 sign · 3 done
  const [done, setDone] = useState<{ passport: PassportDetail; facilityName: string }>();
  const { register, getValues, trigger, control, formState: { errors } } = useForm<IssueFormInput, unknown, IssueForm>({
    resolver: zodResolver(issueFormSchema),
    defaultValues: { facility_id: "", template_id: "", quantity: undefined },
  });

  const facilities = useQuery({ queryKey: queryKeys.facilities, queryFn: endpoints.facilitiesList });
  const templates = useQuery({ queryKey: queryKeys.templates, queryFn: endpoints.templatesList });

  const facilityId = useWatch({ control, name: "facility_id" });
  const templateId = useWatch({ control, name: "template_id" });
  const quantity = Number(useWatch({ control, name: "quantity" })) || 0;
  const facility = facilities.data?.find((f) => f.id === facilityId);
  const template = templates.data?.find((m) => m.id === templateId);

  const evidence = useQuery({
    queryKey: [...queryKeys.certificates, "facility", facilityId, certificateStatus.valid],
    queryFn: () => endpoints.attestationsList(facilityId, certificateStatus.valid),
    enabled: !!facilityId,
  });

  const required = template?.stages.find((s) => s.key === "ISSUE")?.requiredEvidence ?? [];
  const onFile = new Set((evidence.data ?? []).map((c) => c.type));
  const missing = required.filter((type) => !onFile.has(type));
  const checks = required.map((type) => ({
    name: type, label: humanize(type),
    result: (onFile.has(type) ? "verified" : "failed") as Verdict,
    explanation: onFile.has(type) ? t.evidenceOnFile : evidenceRowMissing,
  }));

  const unit = template?.unit ?? facility?.capacity_unit ?? "";
  const block = facility ? issueBlock(facility, quantity) : null;
  const cantSign = !!block || missing.length > 0;

  const issue = useMutation({
    mutationFn: ({ pin, v }: { pin: string; v: IssueForm }) =>
      endpoints.issuePassport({ facility_id: v.facility_id, template_id: v.template_id, quantity: Number(v.quantity) }, pin),
    onSuccess: (passport, { v }) => {
      setDone({ passport, facilityName: facilities.data?.find((f) => f.id === v.facility_id)?.name ?? "" });
      setStep(3);
      client.invalidateQueries({ queryKey: queryKeys.facilities });
      client.invalidateQueries({ queryKey: queryKeys.passports });
    },
  });

  const facts: Fact[] = template && facility ? [
    { label: t.facts.product, value: humanize(template.commodity) },
    { label: t.facts.template, value: template.name },
    { label: t.facts.facility, value: facility.name },
    { label: t.facts.quantity, value: formatQuantity(quantity, template.unit), mono: true },
    { label: t.facts.evidence, value: evidenceCount(required.length - missing.length, required.length) },
  ] : [];

  if (step === 3 && done) return <Created done={done} onAgain={() => { setDone(undefined); setStep(0); }} />;

  const toQuantity = async () => {
    if (await trigger(["template_id", "facility_id"])) setStep(1);
  };
  const toSign = async () => {
    if (await trigger()) setStep(2);
  };

  return (
    <main className="mx-auto grid w-full max-w-content content-start gap-6 p-4 lg:p-8">
      <header className="grid gap-3">
        <h1 className="text-[22px] font-semibold lg:text-page">{t.title}</h1>
        <Stepper steps={t.steps} current={step} error={step === 1 && cantSign} />
      </header>

      {facilities.isError && <Notice kind="blocking">{problemOf(facilities.error)}</Notice>}
      {templates.isError && <Notice kind="blocking">{problemOf(templates.error)}</Notice>}
      {facilities.isPending || templates.isPending ? <Loading /> : (
        <>
          {step === 0 && (
            <section className="grid gap-6">
              <div className="grid gap-3">
                <h2 className="text-[22px] font-semibold">{t.where}</h2>
                <TemplatePicker templates={templates.data ?? []} error={errors.template_id?.message} {...register("template_id")} />
                <p className="text-small text-ink-subtle">{t.templateHelp}</p>
              </div>
              <FacilityPicker facilities={facilities.data ?? []} error={errors.facility_id?.message} {...register("facility_id")} />
              <div className="flex flex-wrap items-center justify-end gap-3">
                <Button size="scan" onClick={toQuantity}>{next}</Button>
              </div>
            </section>
          )}

          {step === 1 && facility && template && (
            <section className="grid gap-6">
              <div className="grid gap-3 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
                <h2 className="text-section">{t.howMuch}</h2>
                <TextField
                  label={`${t.quantity} (${unit})`} mono type="number" inputMode="numeric" min={1} step={1}
                  help={`${humanize(template.commodity)} · ${facility.name}`}
                  error={errors.quantity?.message
                    ?? (block ? t.blockedField(block.left, facility.capacity_unit, periodOf(facility)) : undefined)}
                  {...register("quantity")}
                />
                <CapacityMeter title={facility.name} used={facility.capacity_used} max={facility.capacity}
                  unit={facility.capacity_unit} />
                {quantity > 0 && !block && (
                  <p className="text-small font-semibold text-ink-muted">
                    {afterPassport(facility.capacity_used + quantity, facility.capacity_unit)}
                  </p>
                )}
                {block && <Notice kind="blocking">{t.blockedNotice(block.left, facility.capacity_unit)}</Notice>}
              </div>

              <div className="grid gap-3 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
                <h2 className="text-section">{t.evidence}</h2>
                <p className="text-small text-ink-subtle">{t.evidenceHelp}</p>
                {evidence.isPending ? <Loading />
                  : evidence.isError ? <Notice kind="blocking">{problemOf(evidence.error)}</Notice>
                  : required.length === 0 ? <p className="text-sm text-ink-muted">{t.evidenceNone}</p>
                  : <ChecksList checks={checks} />}
                {required.length > 0 && missing.length > 0 && (
                  <Notice kind="blocking">{t.evidenceMissing(missing.map(humanize).join(", "))}</Notice>
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
                  {back}
                </Button>
                {cantSign && <p className="text-small font-semibold text-danger">{t.blockedReason}</p>}
                <Button size="scan" onClick={toSign} disabled={cantSign}>{t.review}</Button>
              </div>
            </section>
          )}

          {step === 2 && (
            <section className="grid gap-4">
              <Button variant="ghost" onClick={() => setStep(1)} className="justify-self-start">
                <ArrowLeft aria-hidden />
                {back}
              </Button>
              <SigningPanel
                facts={facts} meaning={t.meaning} button={t.sign} busy={issue.isPending}
                error={issue.isError ? problemOf(issue.error) : undefined} blocked={block ? t.blockedReason : undefined}
                onSign={(pin) => issue.mutate({ pin, v: issueFormSchema.parse(getValues()) })}
              />
            </section>
          )}
        </>
      )}
    </main>
  );
}

/** Radio cards for the commodity template: name, commodity, unit and the stages it moves through. */
function TemplatePicker({ templates, error, ...radio }: React.ComponentProps<"input"> & {
  templates: Template[]; error?: string;
}) {
  if (templates.length === 0) return <Notice kind="info">{t.templateEmpty}</Notice>;
  return (
    <fieldset className="grid gap-1.5" aria-invalid={error ? true : undefined}>
      <legend className="mb-1.5 text-sm font-semibold">{t.template}</legend>
      <div className="grid gap-2 md:grid-cols-2">
        {templates.map((m) => (
          <label key={m.id} className="grid min-h-tap-scan cursor-pointer grid-cols-[1fr_auto] items-center gap-x-3 rounded-md border border-border-strong bg-surface p-3 has-checked:border-2 has-checked:border-primary has-checked:bg-primary-soft has-focus-visible:ring-3 has-focus-visible:ring-ring/50">
            <span className="grid min-w-0 gap-1">
              <span className="text-[15px] font-semibold">{m.name}</span>
              <span className="text-small text-ink-muted">{humanize(m.commodity)} · {m.unit}</span>
              <span className="flex flex-wrap gap-1.5">
                {templateStages(m).map((key) => (
                  <span key={key} className="inline-flex items-center gap-1 rounded-full border border-border bg-surface-sunken px-2 py-0.5 text-xs font-semibold text-ink-muted">
                    <Layers className="size-3.5" aria-hidden />
                    {stageLabel(key)}
                  </span>
                ))}
              </span>
            </span>
            <input type="radio" value={m.id} className="peer sr-only" {...radio} />
            <span className="size-[22px] rounded-full border-2 border-border-strong peer-checked:border-[7px] peer-checked:border-primary" aria-hidden />
          </label>
        ))}
      </div>
      {error && (
        <p className="flex gap-1.5 text-small font-semibold text-danger">
          <OctagonX className="mt-px size-4 shrink-0" aria-hidden />
          {error}
        </p>
      )}
    </fieldset>
  );
}

/** The organisation's own facilities: active ones selectable with what's left, pending or suspended ones disabled. */
function FacilityPicker({ facilities, error, ...radio }: React.ComponentProps<"input"> & {
  facilities: Facility[]; error?: string;
}) {
  const usable = facilities.filter((f) => f.status === facilityStatus.active);
  const suspended = facilities.filter((f) => f.status !== facilityStatus.active);
  if (facilities.length === 0) {
    return (
      <div className="grid gap-1.5">
        <Notice kind="info">
          {t.facilityEmpty}{" "}
          <Link href={routes.facilityNew} className="font-semibold underline">{copy.facilities.register}</Link>
        </Notice>
        {error && (
          <p className="flex gap-1.5 text-small font-semibold text-danger">
            <OctagonX className="mt-px size-4 shrink-0" aria-hidden />
            {error}
          </p>
        )}
      </div>
    );
  }
  return (
    <div className="grid gap-3">
      <ChoiceCards
        legend={t.facility} error={error}
        options={usable.map((f) => {
          const free = capacityLeft(f);
          return {
            value: f.id,
            label: f.name,
            icon: kindIcon[f.kind] ?? MapPin,
            sub: free > 0 ? t.facilityLeft(free, f.capacity_unit, periodOf(f)) : t.facilityNone,
          };
        })}
        {...radio}
      />
      {suspended.length > 0 && (
        <div className="grid gap-2 md:grid-cols-2" aria-label={t.facilitySuspended}>
          {suspended.map((f) => (
            <div key={f.id} aria-disabled className="grid min-h-tap-scan grid-cols-[1fr_auto] items-center gap-x-3 rounded-md border border-dashed border-border-strong bg-surface-sunken p-3">
              <span>
                <span className="block text-[15px] font-semibold text-ink-muted">{f.name}</span>
                <span className="block text-small text-ink-muted">{f.status === facilityStatus.pending ? t.facilityPending : t.facilitySuspended}</span>
              </span>
              <OctagonX className="size-5 text-ink-subtle" aria-hidden />
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

/** The signed record, with what happens next. */
function Created({ done, onAgain }: { done: { passport: PassportDetail; facilityName: string }; onAgain: () => void }) {
  const p = done.passport;
  const active = p.status === passportStatus.active;
  return (
    <main className="mx-auto grid w-full max-w-content content-start gap-6 p-4 lg:p-8">
      <section className="grid justify-items-center gap-3 py-6 text-center">
        <span className="grid size-18 place-items-center rounded-full bg-verified-soft text-verified">
          <CircleCheck className="size-8" aria-hidden />
        </span>
        <h1 className="text-[22px] font-semibold lg:text-page">{t.createdTitle}</h1>
        <p className="text-ink-muted">{t.createdBody(formatQuantity(p.quantity, p.unit), done.facilityName)}</p>
        <p className="code text-ink">{p.code}</p>
        <StatusChip tone={active ? "verified" : "pending"} icon={active ? CircleCheck : Clock}>
          {copy.passportView.status[p.status as PassportStatus] ?? humanize(p.status)}
        </StatusChip>
      </section>

      <section className="grid gap-3 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
        <h2 className="text-section">{whatNext}</h2>
        <ol className="grid gap-2">
          {nextSteps.map((s) => (
            <li key={s} className="flex gap-2 text-sm text-ink-muted">
              <CircleCheck className="mt-0.5 size-4 shrink-0 text-verified" aria-hidden />
              {s}
            </li>
          ))}
        </ol>
        <p className="text-small text-ink-subtle">{t.labelsNote}</p>
      </section>

      <div className="flex flex-wrap justify-center gap-2">
        <Link href={routes.passportDetail(p.id)} className={buttonVariants({ size: "scan" })}>{t.viewPassport}</Link>
        <Link href={routes.handover(p.id)} className={buttonVariants({ variant: "secondary", size: "scan" })}>{t.handOver}</Link>
        <Button variant="secondary" size="scan" onClick={onAgain}>{t.another}</Button>
      </div>
    </main>
  );
}
