"use client";

// Turning lots into new lots: pick the input lots and how much of each goes
// in, choose the stage the new lots enter, say how much comes out, then sign.
// The mass balance is checked here before the PIN is asked for.

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ArrowLeftRight,
  ArrowRight,
  CircleCheck,
  Layers,
  Lock,
  Plus,
  RefreshCcw,
  Ship,
  Trash2,
  Warehouse,
} from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { useFieldArray, useForm, useWatch } from "react-hook-form";
import { toast } from "sonner";

import { FieldNote } from "@/components/chaintrack/field-note";
import { PageHeader } from "@/components/chaintrack/page-header";
import { ChoiceCards } from "@/components/chaintrack/choice-cards";
import { Notice } from "@/components/chaintrack/notice";
import { Loading, PageState } from "@/components/chaintrack/page-state";
import type { Fact } from "@/components/chaintrack/passport-card";
import { SigningPanel } from "@/components/chaintrack/signing";
import { TextField } from "@/components/chaintrack/text-field";
import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { endpoints, problemOf, useMe, type PassportDetail } from "@/lib/api";
import {
  copy,
  limits,
  permissions,
  queryKeys,
  routes,
  templateEvent,
  screenCopy,
} from "@/lib/consts";
import { liveValidation } from "@/lib/form";
import { lotCodeSchema, type LotCodeForm } from "@/lib/schemas/scan";
import {
  transformFormSchema,
  yieldOf,
  type TransformForm,
  type TransformFormInput,
} from "@/lib/schemas/transform";
import { cn } from "@/lib/utils";
import { formatQuantity, humanize } from "@/lib/verification";

import { LotScanner } from "../scan/scan";

const t = copy.transform;
const s = copy.scan;
const v = copy.passportView;

// Copy this screen needs that consts doesn't hold yet.
const local = screenCopy.transform;

// Stage keys are a template's own; the label is worked out from the key here.
const stageIcons: Record<string, typeof Layers> = {
  HANDOVER: ArrowLeftRight,
  TRANSFORM: RefreshCcw,
  EXPORT: Ship,
  IMPORT: Warehouse,
};
const stageLabel = humanize;

type ResultState = {
  rows: PassportDetail[];
  into: string;
  out: string;
  pct: string;
  max: string | null;
};

export function Transform() {
  const me = useMe().data!; // the layout only renders once it is loaded
  const client = useQueryClient();
  const [known, setKnown] = useState<Record<string, PassportDetail>>({});
  const [result, setResult] = useState<ResultState | null>(null);

  const {
    control,
    register,
    handleSubmit,
    reset,
    formState: { errors, isValid },
  } = useForm<TransformFormInput, unknown, TransformForm>({
    ...liveValidation,
    resolver: zodResolver(transformFormSchema),
    mode: "onChange", // the sign button follows isValid while typing
    defaultValues: {
      stage: "",
      inputs: [],
      outputs: [{ quantity: "" as unknown as number }],
    },
  });
  const inputs = useFieldArray({ control, name: "inputs" });
  const outputs = useFieldArray({ control, name: "outputs" });
  const addForm = useForm<LotCodeForm, unknown, LotCodeForm>({
    ...liveValidation,
    resolver: zodResolver(lotCodeSchema),
    defaultValues: { code: "" },
  });

  const form = useWatch({ control }) as TransformFormInput;
  const rows = form.inputs;
  const first = rows[0] ? known[rows[0].passport_id] : undefined;
  const unit = first?.unit ?? "";
  const templateId = first?.template_id;
  const template = useQuery({
    queryKey: queryKeys.template(templateId ?? ""),
    queryFn: () => endpoints.template(templateId!),
    enabled: !!templateId,
  });
  const stages = (template.data?.stages ?? []).filter((st) =>
    st.events?.includes(templateEvent.transform),
  );
  const stageKey = form.stage;
  const maxYield = stages.find((st) => st.key === stageKey)?.maxYield ?? null;

  const y = yieldOf(
    rows.map((r) => Number(r.quantity) || 0),
    (form.outputs ?? []).map((o) => Number(o.quantity) || 0),
    maxYield,
  );
  const allowed = maxYield === null ? null : (y.totalIn * maxYield) / 100;
  const share = allowed && allowed > 0 ? y.totalOut / allowed : 0;

  const atCap = rows.length >= limits.maxItems;

  const add = useMutation({
    mutationFn: (code: string) => endpoints.passport(code),
    onSuccess: (p) => {
      setKnown((k) => ({ ...k, [p.id]: p }));
      if (rows.some((r) => r.passport_id === p.id)) {
        toast.error(local.alreadyAdded);
        return;
      }
      if (atCap) {
        toast.error(local.tooManyRows(limits.maxItems));
        return;
      }
      inputs.append({ passport_id: p.id, quantity: p.remaining });
      addForm.reset();
    },
  });

  const sign = useMutation({
    mutationFn: ({ values, pin }: { values: TransformForm; pin: string }) =>
      endpoints.transform(
        { stage: values.stage, inputs: values.inputs, outputs: values.outputs },
        pin,
      ),
    meta: { inPlace: true },
    onSuccess: ({ data: created }, vars) => {
      const into = vars.values.inputs.reduce((a, r) => a + r.quantity, 0);
      const out = vars.values.outputs.reduce((a, r) => a + r.quantity, 0);
      setResult({
        rows: created,
        into: formatQuantity(into, unit),
        out: formatQuantity(out, unit),
        pct: (into === 0 ? 0 : (out / into) * 100).toFixed(1),
        max: maxYield === null ? null : String(maxYield),
      });
      vars.values.inputs.forEach((r) =>
        client.invalidateQueries({
          queryKey: queryKeys.passport(r.passport_id),
        }),
      );
      client.invalidateQueries({ queryKey: queryKeys.passports });
      client.invalidateQueries({ queryKey: queryKeys.count("passportsHeld") });
    },
  });

  if (!me.permissions.includes(permissions.transformCreate)) {
    return (
      <PageState icon={Lock} title={local.notAllowedTitle}>
        {local.notAllowedBody}
      </PageState>
    );
  }
  if (result) {
    return (
      <Done
        result={result}
        onAgain={() => {
          setResult(null);
          setKnown({});
          reset();
        }}
      />
    );
  }

  const quantityHelp = (p: PassportDetail | undefined) =>
    p
      ? v.facts.remaining + ": " + formatQuantity(p.remaining, p.unit)
      : undefined;
  const validateQty = (p: PassportDetail | undefined) => (value: string) => {
    const n = Number(value);
    if (!Number.isInteger(n) || n <= 0) return copy.validation.quantityPositive;
    if (p && n > p.remaining) return local.overRemaining(p.remaining, p.unit);
    return true;
  };
  const blocked = y.over
    ? t.blocked
    : !isValid
      ? local.readyBlocked
      : undefined;
  const signFacts: Fact[] = [
    {
      label: t.inputs,
      wide: true,
      value: (
        <ul className="grid gap-1">
          {rows.map((r) => {
            const p = known[r.passport_id];
            return (
              <li
                key={r.passport_id}
                className="flex flex-wrap items-baseline justify-between gap-x-3"
              >
                <Link
                  href={routes.passportDetail(r.passport_id)}
                  className="code break-all text-primary"
                >
                  {p?.code ?? r.passport_id}
                </Link>
                <span className="fig">
                  {formatQuantity(Number(r.quantity) || 0, p?.unit ?? unit)}
                </span>
              </li>
            );
          })}
        </ul>
      ),
    },
    { label: t.outputs, value: formatQuantity(y.totalOut, unit), mono: true },
    { label: local.newLots, value: String(outputs.fields.length) },
    { label: t.stage, value: stageLabel(stageKey) },
    {
      label: local.yieldLabel,
      value:
        maxYield === null
          ? local.noYieldShort
          : t.yieldLine(y.pct.toFixed(1), String(maxYield)),
    },
  ];

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <PageHeader title={t.title} sub={t.sub} />

      <div className="grid items-start gap-4 lg:grid-cols-2 lg:gap-6">
        <section className="grid min-w-0 gap-3 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
          <header className="flex flex-wrap items-baseline justify-between gap-2">
            <h2 className="text-section">{t.inputs}</h2>
            {rows.length > 0 && (
              <p className="text-small text-ink-muted">
                {t.selected(rows.length, formatQuantity(y.totalIn, unit))}
              </p>
            )}
          </header>

          {rows.length === 0 ? (
            <p className="text-sm text-ink-muted">{t.inputEmpty}</p>
          ) : (
            <ul className="grid gap-2">
              {inputs.fields.map((field, i) => {
                const row = rows[i];
                const p = known[row.passport_id];
                const id = `in-${field.id}`;
                return (
                  <li
                    key={field.id}
                    className="grid gap-3 rounded-md border border-border bg-surface p-3"
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="min-w-0">
                        <Link
                          href={routes.passportDetail(row.passport_id)}
                          className="code break-all text-primary"
                        >
                          {p?.code ?? row.passport_id}
                        </Link>
                        <p className="text-small text-ink-muted">
                          {p ? humanize(p.commodity) : copy.loading}
                        </p>
                      </div>
                      <Button
                        variant="ghost"
                        aria-label={t.remove}
                        onClick={() => inputs.remove(i)}
                      >
                        <Trash2 aria-hidden />
                      </Button>
                    </div>
                    <QuantityField
                      id={id}
                      label={t.quantity}
                      unit={p?.unit ?? unit}
                      help={quantityHelp(p)}
                      error={errors.inputs?.[i]?.quantity?.message}
                      {...register(`inputs.${i}.quantity`, {
                        validate: (v) => validateQty(p)(String(v ?? "")),
                      })}
                    />
                  </li>
                );
              })}
            </ul>
          )}

          <div className="grid gap-2">
            <LotScanner onCode={(code) => add.mutate(code)} disabled={atCap} />
            <form
              onSubmit={addForm.handleSubmit((val) => add.mutate(val.code))}
              noValidate
              className="flex flex-wrap items-end gap-2"
            >
              <TextField
                label={s.codeLabel}
                help={s.codeHint}
                error={addForm.formState.errors.code?.message}
                mono
                className="min-w-56 flex-1"
                autoComplete="off"
                autoCapitalize="characters"
                disabled={atCap}
                {...addForm.register("code")}
              />
              <Button
                type="submit"
                variant="secondary"
                disabled={atCap}
                loading={add.isPending}
              >
                <Plus aria-hidden />
                {local.addLot}
              </Button>
            </form>
            {atCap && (
              <Notice kind="info">{local.tooManyRows(limits.maxItems)}</Notice>
            )}
          </div>
        </section>

        <div className="grid min-w-0 content-start gap-4 lg:gap-6">
          {first && (
            <section className="grid min-w-0 gap-3 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
              <header className="grid gap-1">
                <h2 className="text-section">{t.stage}</h2>
                <p className="text-small text-ink-muted">{t.stageHelp}</p>
                {template.data && (
                  <p className="text-small text-ink-muted">
                    {local.templateLine(template.data.name)}
                  </p>
                )}
              </header>
              {template.isPending && <Loading />}
              {template.isError && (
                <Notice kind="blocking">{problemOf(template.error)}</Notice>
              )}
              {template.isSuccess && stages.length === 0 && (
                <Notice kind="info">{local.noTemplate}</Notice>
              )}
              {stages.length > 0 && (
                <ChoiceCards
                  legend={t.stage}
                  error={errors.stage?.message}
                  options={stages.map((st) => ({
                    value: st.key,
                    label: stageLabel(st.key),
                    sub:
                      st.maxYield == null
                        ? local.noYieldShort
                        : local.allowedYield(st.maxYield),
                    icon: stageIcons[st.key] ?? Layers,
                  }))}
                  {...register("stage")}
                />
              )}
              {stageKey !== "" && maxYield === null && (
                <Notice kind="info">{t.noYield}</Notice>
              )}
            </section>
          )}

          <section
            className={cn(
              "grid min-w-0 gap-3 rounded-lg border bg-surface p-4 shadow-card lg:p-6",
              y.over ? "border-2 border-danger" : "border-border",
            )}
          >
            <header className="flex flex-wrap items-baseline justify-between gap-2">
              <h2 className="text-section">{t.outputs}</h2>
              <p className="fig text-sm">{formatQuantity(y.totalOut, unit)}</p>
            </header>
            <ul className="grid gap-3">
              {outputs.fields.map((field, i) => {
                const id = `out-${field.id}`;
                return (
                  <li key={field.id} className="flex items-end gap-2">
                    <QuantityField
                      id={id}
                      label={t.output(i + 1)}
                      unit={unit}
                      className="min-w-0 flex-1"
                      error={
                        errors.outputs?.[i]?.quantity?.message ??
                        (y.over && i === 0 ? local.totalTooHigh : undefined)
                      }
                      {...register(`outputs.${i}.quantity`)}
                    />
                    <Button
                      variant="ghost"
                      aria-label={t.remove}
                      onClick={() => outputs.remove(i)}
                    >
                      <Trash2 aria-hidden />
                    </Button>
                  </li>
                );
              })}
            </ul>
            <Button
              variant="secondary"
              onClick={() =>
                outputs.append({ quantity: "" as unknown as number })
              }
              disabled={outputs.fields.length >= limits.maxItems}
            >
              <Plus aria-hidden />
              {t.addOutput}
            </Button>
            {outputs.fields.length >= limits.maxItems && (
              <Notice kind="info">{local.tooManyRows(limits.maxItems)}</Notice>
            )}
          </section>

          <section
            className={cn(
              "grid min-w-0 gap-3 rounded-lg border bg-surface p-4 shadow-card lg:p-6",
              y.over ? "border-2 border-danger" : "border-border",
            )}
          >
            <header className="flex flex-wrap items-baseline justify-between gap-2">
              <h2 className="text-section">{local.massBalance}</h2>
              <p className="fig text-sm">
                {t.inOut(
                  formatQuantity(y.totalIn, unit),
                  formatQuantity(y.totalOut, unit),
                )}
              </p>
            </header>
            {stageKey === "" ? (
              <Notice kind="info">{local.pickStage}</Notice>
            ) : maxYield === null || allowed === null ? (
              <Notice kind="info">{t.noYield}</Notice>
            ) : (
              <>
                <div
                  role="meter"
                  aria-label={local.massBalance}
                  aria-valuenow={y.totalOut}
                  aria-valuemin={0}
                  aria-valuemax={allowed}
                  className="relative h-3.5 overflow-hidden rounded-sm bg-surface-sunken ring-1 ring-border-strong ring-inset"
                >
                  <span
                    className={cn(
                      "absolute inset-y-0 left-0",
                      y.over ? "bg-danger" : "bg-verified",
                    )}
                    style={{ width: `${Math.min(100, share * 100)}%` }}
                  />
                </div>
                {y.over ? (
                  <Notice kind="blocking">
                    {t.meterOver(formatQuantity(y.totalOut - allowed, unit))}{" "}
                    {local.overSaid(
                      formatQuantity(y.totalOut, unit),
                      formatQuantity(allowed, unit),
                    )}
                  </Notice>
                ) : (
                  <>
                    <p className="flex items-center gap-1.5 text-small font-semibold text-verified">
                      <CircleCheck className="size-4 shrink-0" aria-hidden />
                      {t.meterWithin(
                        formatQuantity(allowed - y.totalOut, unit),
                      )}
                    </p>
                    <p className="text-small text-ink-muted">
                      {t.yieldLine(y.pct.toFixed(1), String(maxYield))}
                    </p>
                    <p className="text-small text-ink-muted">
                      {local.formula(
                        formatQuantity(y.totalIn, unit),
                        String(maxYield),
                        formatQuantity(allowed, unit),
                      )}
                    </p>
                  </>
                )}
              </>
            )}
          </section>

          <SigningPanel
            facts={signFacts}
            meaning={t.meaning}
            button={t.sign(outputs.fields.length)}
            blocked={blocked}
            error={sign.isError ? problemOf(sign.error) : undefined}
            onSign={(pin) =>
              handleSubmit((values) => sign.mutate({ values, pin }))()
            }
          />
        </div>
      </div>
    </main>
  );
}

/** One quantity input with its unit, help and error (components.md → Form fields). */
function QuantityField({
  label,
  unit,
  help,
  error,
  className,
  ...props
}: React.ComponentProps<"input"> & {
  label: string;
  unit: string;
  help?: string;
  error?: string;
}) {
  const noteId = `${props.id}-note`;
  return (
    <div className={cn("grid gap-1.5", className)}>
      <label htmlFor={props.id} className="text-sm font-semibold">
        {label}
      </label>
      <div className="flex">
        <Input
          aria-invalid={error ? true : undefined}
          aria-describedby={error || help ? noteId : undefined}
          className={cn("font-mono text-base", unit && "rounded-r-none")}
          inputMode="numeric"
          {...props}
        />
        {unit && (
          <span className="grid place-items-center rounded-r-md border border-l-0 border-border-strong bg-surface-sunken px-3 font-mono text-sm text-ink-muted">
            {unit}
          </span>
        )}
      </div>
      <FieldNote id={noteId} error={error} help={help} />
    </div>
  );
}

/** What was created: the new lot codes, and where to go from here. */
function Done({
  result,
  onAgain,
}: {
  result: ResultState;
  onAgain: () => void;
}) {
  return (
    <main className="grid w-full justify-items-center gap-4 p-4 text-center lg:gap-6 lg:p-8">
      <span className="mt-4 grid size-18 place-items-center rounded-full bg-verified text-on-verified">
        <CircleCheck className="size-8" aria-hidden />
      </span>
      <h1 className="text-2xl font-semibold lg:text-page">
        {t.createdTitle(result.rows.length)}
      </h1>
      <p className="max-w-[48ch] text-ink-muted">
        {result.max === null
          ? t.inOut(result.into, result.out)
          : t.createdBody(result.into, result.out, result.pct, result.max)}
      </p>
      <Notice kind="success">{t.inputsClosed}</Notice>
      <ul className="grid w-full max-w-md gap-2">
        {result.rows.map((p) => (
          <li
            key={p.id}
            className="flex flex-wrap items-center justify-between gap-2 rounded-md border border-border bg-surface px-3 py-2"
          >
            <Link
              href={routes.passportDetail(p.id)}
              className="code break-all text-primary"
            >
              {p.code}
            </Link>
            <span className="fig text-sm">
              {formatQuantity(p.quantity, p.unit)}
            </span>
          </li>
        ))}
      </ul>
      <div className="flex flex-wrap justify-center gap-2">
        <Button onClick={onAgain}>
          <RefreshCcw aria-hidden />
          {t.again}
        </Button>
        <Link
          href={routes.handover(result.rows[0]?.id)}
          className={buttonVariants({ variant: "secondary" })}
        >
          <ArrowRight aria-hidden />
          {copy.issue.handOver}
        </Link>
      </div>
    </main>
  );
}
