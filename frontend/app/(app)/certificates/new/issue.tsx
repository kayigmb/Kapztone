"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Circle, CircleCheck, FileText, Search } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm, useWatch, Controller } from "react-hook-form";

import { DataWidget } from "@/components/chaintrack/data-widget";
import { FieldNote } from "@/components/chaintrack/field-note";
import { PageHeader } from "@/components/chaintrack/page-header";
import { useRecordTitle } from "@/components/chaintrack/record-title";
import { Notice } from "@/components/chaintrack/notice";

import { SelectInput } from "@/components/chaintrack/select-input";
import { SigningPanel } from "@/components/chaintrack/signing";
import { StatusChip } from "@/components/chaintrack/status";
import { TextField } from "@/components/chaintrack/text-field";
import { Button } from "@/components/ui/button";
import { endpoints, problemOf, useMe } from "@/lib/api";
import { copy, queryKeys, routes, subjectTypes } from "@/lib/consts";
import { liveValidation } from "@/lib/form";
import {
  certificateBody,
  certificateFormSchema,
  type CertificateForm,
} from "@/lib/schemas/certificate";
import { lotCodeSchema, type LotCodeForm } from "@/lib/schemas/scan";
import { cn } from "@/lib/utils";
import { formatDate, humanize } from "@/lib/verification";

import { LotScanner } from "../../scan/scan";

const t = copy.certificates.new;
const today = () => new Date().toISOString().slice(0, 10);
const inAYear = () => {
  const d = new Date();
  d.setFullYear(d.getFullYear() + 1);
  return d.toISOString().slice(0, 10);
};

export function IssueCertificate() {
  useRecordTitle(t.title);
  const me = useMe().data!;
  const router = useRouter();
  const client = useQueryClient();
  const [lotRef, setLotRef] = useState("");

  const lot = useQuery({
    queryKey: queryKeys.verify(lotRef),
    queryFn: () => endpoints.verify(lotRef),
    enabled: !!lotRef,
  });
  const orgId = me.organisation?.id ?? "";
  const accredited = useQuery({
    queryKey: queryKeys.orgAccreditations(orgId),
    queryFn: () => endpoints.accreditations(orgId),
    enabled: !!orgId,
  });
  const now = today();
  const types = (accredited.data ?? [])
    .filter(
      (a) => a.valid_from.slice(0, 10) <= now && a.valid_to.slice(0, 10) >= now,
    )
    .map((a) => a.type);

  const {
    register,
    handleSubmit,
    setValue,
    control,
    formState: { errors },
  } = useForm<CertificateForm>({
    ...liveValidation,
    resolver: zodResolver(certificateFormSchema),
    defaultValues: {
      subject_type: subjectTypes.passport,
      subject_id: "",
      type: "",
      valid_from: now,
      valid_to: inAYear(),
    },
  });
  const form = useWatch({ control });

  // The PIN signs whatever the form holds at that moment, after it validates.
  const issue = useMutation({
    mutationFn: ({ f, pin }: { f: CertificateForm; pin: string }) => {
      const { file, ...rest } = f;
      return endpoints.issueAttestation(
        certificateBody(rest, file, orgId),
        pin,
      );
    },
    meta: { inPlace: true },
    onSuccess: () => {
      client.invalidateQueries({ queryKey: queryKeys.certificates });
      router.push(routes.certificates);
    },
  });

  const v = lot.data;
  const subjects = v
    ? [
        {
          type: subjectTypes.passport,
          id: v.passport.id,
          label: `${t.attachPassport} · ${v.passport.code ?? v.passport.id}`,
        },
        ...v.origins.map((o) => ({
          type: subjectTypes.facility,
          id: o.facility.id,
          label: `${t.attachFacility} · ${o.facility.name}`,
        })),
      ]
    : [];
  const checks = [
    {
      ok: !!v,
      label: v
        ? t.checkVerdict(copy.verification.verdict[v.verdict])
        : t.checkVerdict("—"),
    },
    { ok: !!form.type && types.includes(form.type), label: t.checkAccredited },
    {
      ok:
        !!form.valid_from && !!form.valid_to && form.valid_to > form.valid_from,
      label: t.checkDates,
    },
    { ok: form.file instanceof File, label: t.checkFile },
  ];

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <PageHeader title={t.title} sub={t.sub} />

      <div className="grid gap-6 lg:grid-cols-[1.2fr_0.8fr] lg:items-start">
        <section className="grid gap-5 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
          <FindLot onFound={setLotRef} />
          {lotRef && (
            <DataWidget query={lot}>
              {(found) =>
                !found ? (
                  <Notice kind="warning">
                    {copy.verification.notFoundTitle}
                  </Notice>
                ) : null
              }
            </DataWidget>
          )}

          {v && (
            <fieldset className="grid gap-2">
              <legend className="mb-1.5 text-sm font-semibold">
                {t.attach}
              </legend>
              {subjects.map((s) => {
                const on = form.subject_id === s.id;
                return (
                  <button
                    key={s.id}
                    type="button"
                    aria-pressed={on}
                    onClick={() => {
                      setValue("subject_type", s.type);
                      setValue("subject_id", s.id, { shouldValidate: true });
                    }}
                    className={cn(
                      "min-h-tap rounded-md border px-4 text-left text-sm font-semibold",
                      on
                        ? "border-primary bg-primary-soft text-primary"
                        : "border-border text-ink-muted hover:border-primary",
                    )}
                  >
                    {s.label}
                  </button>
                );
              })}
              <FieldNote error={errors.subject_id?.message} />
            </fieldset>
          )}

          <form
            id="certificate"
            noValidate
            className="grid gap-4"
            onSubmit={(e) => e.preventDefault()}
          >
            <label className="grid gap-1.5">
              <span className="text-sm font-semibold">{t.type}</span>
              {types.length === 0 ? (
                <Notice kind="info">{t.typeEmpty}</Notice>
              ) : (
                <Controller
                  control={control}
                  name="type"
                  render={({ field }) => (
                    <SelectInput
                      label={t.type}
                      placeholder={t.typePlaceholder}
                      value={field.value ?? ""}
                      onChange={field.onChange}
                      invalid={!!errors.type}
                      options={types.map((x) => ({
                        value: x,
                        label: humanize(x),
                      }))}
                    />
                  )}
                />
              )}
              <FieldNote error={errors.type?.message} help={t.typeHelp} />
            </label>
            <div className="grid gap-4 md:grid-cols-2">
              <TextField
                label={t.validFrom}
                type="date"
                error={errors.valid_from?.message}
                {...register("valid_from")}
              />
              <TextField
                label={t.validTo}
                type="date"
                help={t.validityHelp}
                error={errors.valid_to?.message}
                {...register("valid_to")}
              />
            </div>
            <label className="grid gap-1.5">
              <span className="text-sm font-semibold">{t.document}</span>
              <span className="text-small text-ink-muted">
                {t.documentHelp}
              </span>
              <input
                type="file"
                accept="application/pdf,image/png,image/jpeg"
                className="min-h-tap text-sm"
                aria-invalid={errors.file ? true : undefined}
                onChange={(e) => {
                  const f = e.target.files?.[0];
                  if (f) setValue("file", f, { shouldValidate: true });
                }}
              />
              <FieldNote
                error={
                  errors.file?.message ??
                  (errors.file ? t.documentMissing : undefined)
                }
              />
            </label>
          </form>
        </section>

        <aside className="grid gap-4">
          <section className="grid gap-3 rounded-lg border border-border bg-surface p-4 shadow-card">
            <span className="text-xs font-bold tracking-widest text-ink-muted uppercase">
              {t.preview}
            </span>
            <div className="grid grid-cols-[auto_1fr_auto] items-start gap-3">
              <span className="grid size-10 place-items-center rounded-md bg-primary-soft text-primary">
                <FileText className="size-5" aria-hidden />
              </span>
              <div className="grid gap-0.5">
                <strong className="text-sm">
                  {form.type ? humanize(form.type) : t.type}
                </strong>
                <span className="text-small text-ink-muted">
                  {me.organisation?.name}
                </span>
                {form.valid_to && (
                  <span className="text-small text-ink-muted">
                    {t.validTo} {formatDate(form.valid_to)}
                  </span>
                )}
              </div>
              <StatusChip tone="verified" icon={CircleCheck}>
                {copy.certificates.status.valid}
              </StatusChip>
            </div>
          </section>
          <section className="grid gap-2 rounded-lg border border-border bg-surface p-4 shadow-card">
            <h2 className="text-title">{t.checks}</h2>
            <ul className="grid gap-1.5">
              {checks.map((c) => (
                <li
                  key={c.label}
                  className={cn(
                    "flex items-center gap-2 text-sm",
                    c.ok ? "text-verified-ink" : "text-ink-muted",
                  )}
                >
                  {c.ok ? (
                    <CircleCheck className="size-4 shrink-0" aria-hidden />
                  ) : (
                    <Circle className="size-4 shrink-0" aria-hidden />
                  )}
                  {c.label}
                </li>
              ))}
            </ul>
          </section>
          <SigningPanel
            meaning={t.meaning}
            button={t.sign}
            busy={issue.isPending}
            blocked={checks.every((c) => c.ok) ? undefined : t.checkMissing}
            error={issue.isError ? problemOf(issue.error) : undefined}
            onSign={(pin) => handleSubmit((f) => issue.mutate({ f, pin }))()}
          />
        </aside>
      </div>
    </main>
  );
}

function FindLot({ onFound }: { onFound: (ref: string) => void }) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LotCodeForm, unknown, { code: string }>({
    ...liveValidation,
    resolver: zodResolver(lotCodeSchema),
    defaultValues: { code: "" },
  });
  return (
    <div className="grid gap-3">
      <form
        noValidate
        className="grid grid-cols-[1fr_auto] items-end gap-2"
        onSubmit={handleSubmit((x) => onFound(x.code))}
      >
        <TextField
          label={copy.scan.codeLabel}
          help={copy.scan.codeHint}
          mono
          autoCapitalize="characters"
          error={errors.code?.message}
          {...register("code")}
        />
        <Button type="submit" variant="secondary" aria-label={copy.scan.open}>
          <Search aria-hidden />
        </Button>
      </form>
      <LotScanner onCode={onFound} />
    </div>
  );
}
