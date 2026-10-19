"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft, Lock, Plus, WifiOff } from "lucide-react";
import Link from "next/link";
import { use, useState } from "react";
import { Controller, useFieldArray, useForm } from "react-hook-form";

import { ChoiceCards } from "@/components/chaintrack/choice-cards";
import { Loading, PageState } from "@/components/chaintrack/page-state";
import { TextField } from "@/components/chaintrack/text-field";
import { Button, buttonVariants } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { endpoints, problemOf, useMeGuard, type Template } from "@/lib/api";
import {
  capacityRuleTypes,
  copy,
  facilityKinds,
  permissions,
  queryKeys,
  routes,
  templateEvents,
  screenCopy,
} from "@/lib/consts";
import { liveValidation } from "@/lib/form";
import {
  templateFormSchema,
  type TemplateForm,
  type TemplateFormInput,
} from "@/lib/schemas/template";
import { formatDate } from "@/lib/verification";

const t = screenCopy.templateDetail;

const originOptions = [
  { value: "", label: t.any },
  ...facilityKinds.map((k) => ({ value: k.value, label: k.label })),
];

const ruleOptions = [
  { value: "none", label: t.none },
  { value: "perArea", label: t.perArea },
  { value: "facilityCapacity", label: t.facilityCapacity },
];

export function TemplateDetail({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = use(params);

  const me = useMeGuard((me) => (me ? undefined : routes.signin));
  const allowed =
    me.ready && me.data!.permissions.includes(permissions.templateManage);

  const tpl = useQuery({
    queryKey: [...queryKeys.template(id)],
    queryFn: () => endpoints.template(id),
    enabled: allowed,
  });

  if (me.isError) {
    return (
      <PageState
        icon={WifiOff}
        tone="danger"
        title={copy.errors.loadTitle}
        retry={() => me.refetch()}
      >
        {problemOf(me.error)}
      </PageState>
    );
  }
  if (!me.ready) return <Loading className="px-4 py-8 lg:px-8" />;
  if (!allowed) {
    return (
      <PageState icon={Lock} title={t.notAllowedTitle}>
        {t.notAllowedBody}
      </PageState>
    );
  }
  if (tpl.isPending) return <Loading className="px-4 py-8 lg:px-8" />;
  if (tpl.isError) {
    return (
      <PageState
        icon={WifiOff}
        tone="danger"
        title={copy.errors.loadTitle}
        retry={() => tpl.refetch()}
      >
        {problemOf(tpl.error)}
      </PageState>
    );
  }

  return <EditForm tpl={tpl.data} id={id} />;
}

function EditForm({ tpl, id }: { tpl: Template; id: string }) {
  const client = useQueryClient();
  const [busy, setBusy] = useState(false);
  const form = useForm<TemplateFormInput, unknown, TemplateForm>({
    ...liveValidation,
    resolver: zodResolver(templateFormSchema),
    defaultValues: fromTemplate(tpl),
  });

  const save = useMutation({
    mutationFn: (v: TemplateForm) => endpoints.updateTemplate(id, toApiBody(v)),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: queryKeys.templates });
      client.invalidateQueries({ queryKey: [...queryKeys.template(id)] });
      setBusy(false);
    },
    onError: () => setBusy(false),
  });

  const {
    control,
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = form;
  const { fields, append, remove } = useFieldArray({ control, name: "stages" });
  const ruleType = watch("capacityRule")?.type ?? "none";

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <header className="grid gap-1">
        <Link
          href={routes.adminTemplates}
          className="inline-flex items-center gap-2 text-sm font-semibold text-primary"
        >
          <ArrowLeft className="size-4" aria-hidden />
          {t.back}
        </Link>
        <h1 className="text-2xl font-semibold lg:text-page">{t.title}</h1>
        <p className="text-sm text-ink-muted">{tpl.name}</p>
      </header>

      <form
        onSubmit={handleSubmit((v) => {
          setBusy(true);
          save.mutate(v);
        })}
        noValidate
        className="grid gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6"
      >
        <div className="grid gap-4 md:grid-cols-2">
          <TextField
            label={t.name}
            error={errors.name?.message}
            {...register("name")}
          />
          <TextField
            label={t.commodity}
            error={errors.commodity?.message}
            {...register("commodity")}
          />
          <TextField
            label={t.unit}
            error={errors.unit?.message}
            {...register("unit")}
          />
          <div className="grid gap-1.5">
            <span className="text-sm font-semibold">
              {t.originFacilityKind}
            </span>
            <Controller
              control={control}
              name="originFacilityKind"
              render={({ field }) => (
                <Select
                  items={originOptions}
                  value={field.value}
                  onValueChange={(v) => field.onChange(v)}
                >
                  <SelectTrigger className="min-h-tap w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {originOptions.map((o) => (
                      <SelectItem key={o.value} value={o.value}>
                        {o.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            />
          </div>
        </div>

        <div className="grid gap-3">
          <span className="text-sm font-semibold">{t.capacityRule}</span>
          <ChoiceCards
            legend={t.capacityRule}
            options={ruleOptions}
            value={
              ruleType === "yield_per_area"
                ? "perArea"
                : ruleType === "facility_capacity"
                  ? "facilityCapacity"
                  : "none"
            }
            onChange={(e) => {
              const v = e.target.value;
              if (v === "none") setValue("capacityRule", null);
              else if (v === "perArea")
                setValue("capacityRule", {
                  type: "yield_per_area",
                  kgPerHaPerSeason: 0,
                });
              else if (v === "facilityCapacity")
                setValue("capacityRule", { type: "facility_capacity" });
            }}
          />
          {ruleType === "yield_per_area" && (
            <TextField
              label={t.kgPerHaPerSeason}
              type="number"
              inputMode="numeric"
              mono
              error={errors.capacityRule?.kgPerHaPerSeason?.message}
              {...register("capacityRule.kgPerHaPerSeason", {
                valueAsNumber: true,
              })}
            />
          )}
        </div>

        <div className="grid gap-3">
          <div className="flex items-center justify-between">
            <span className="text-sm font-semibold">{t.stages}</span>
            <Button
              type="button"
              variant="secondary"
              onClick={() =>
                append({
                  key: "",
                  events: [],
                  maxYield: undefined,
                  requiredEvidence: [],
                })
              }
              disabled={fields.length >= 10}
            >
              <Plus aria-hidden />
              {t.addStage}
            </Button>
          </div>
          {fields.map((field, i) => (
            <div
              key={field.id}
              className="grid gap-3 rounded-md border border-border p-3"
            >
              <div className="flex flex-wrap items-end justify-between gap-2">
                <TextField
                  label={t.stageKey}
                  className="flex-1 min-w-30"
                  error={errors.stages?.[i]?.key?.message}
                  {...register(`stages.${i}.key` as const)}
                />
                {fields.length > 1 && (
                  <Button
                    type="button"
                    variant="danger-outline"
                    onClick={() => remove(i)}
                  >
                    {t.removeStage}
                  </Button>
                )}
              </div>
              <div className="grid gap-2">
                <span className="text-sm font-semibold">{t.events}</span>
                <div className="flex flex-wrap gap-3">
                  {templateEvents.map((ev) => {
                    const checked =
                      watch(`stages.${i}.events`)?.includes(ev.value) ?? false;
                    return (
                      <label key={ev.value} className="flex items-center gap-2">
                        <Checkbox
                          checked={checked}
                          onCheckedChange={(c) => {
                            const current = watch(`stages.${i}.events`) ?? [];
                            if (c)
                              setValue(`stages.${i}.events`, [
                                ...current,
                                ev.value,
                              ]);
                            else
                              setValue(
                                `stages.${i}.events`,
                                current.filter((x) => x !== ev.value),
                              );
                          }}
                        />
                        <span className="text-sm">{ev.label}</span>
                      </label>
                    );
                  })}
                </div>
                {errors.stages?.[i]?.events?.message && (
                  <p className="text-small font-semibold text-danger">
                    {errors.stages?.[i]?.events?.message}
                  </p>
                )}
              </div>
              <div className="grid gap-4 md:grid-cols-2">
                <TextField
                  label={t.maxYield}
                  type="number"
                  inputMode="numeric"
                  mono
                  error={errors.stages?.[i]?.maxYield?.message}
                  {...register(`stages.${i}.maxYield` as const, {
                    valueAsNumber: true,
                  })}
                />
                <TextField
                  label={t.requiredEvidence}
                  error={errors.stages?.[i]?.requiredEvidence?.message}
                  {...register(`stages.${i}.requiredEvidence` as const)}
                />
              </div>
            </div>
          ))}
        </div>

        <div className="grid gap-2 text-small text-ink-muted">
          <div>
            {t.id}: <span className="font-mono">{tpl.id}</span>
          </div>
          {tpl.created_at && (
            <div>
              {t.createdAt}: {formatDate(tpl.created_at)}
            </div>
          )}
          {tpl.updated_at && (
            <div>
              {t.updatedAt}: {formatDate(tpl.updated_at)}
            </div>
          )}
        </div>

        <div className="flex flex-wrap justify-end gap-3">
          <Link
            href={routes.adminTemplates}
            className={buttonVariants({ variant: "secondary" })}
          >
            {t.back}
          </Link>
          <Button type="submit" disabled={busy}>
            {busy ? t.saving : t.save}
          </Button>
        </div>
      </form>
    </main>
  );
}

function fromTemplate(tpl: Template): TemplateForm {
  const rule = tpl.capacityRule;
  let cr: TemplateForm["capacityRule"] = null;
  if (rule?.type === "yield_per_area") {
    cr = {
      type: "yield_per_area",
      kgPerHaPerSeason: rule.kgPerHaPerSeason ?? 0,
    };
  } else if (rule?.type === "facility_capacity") {
    cr = { type: "facility_capacity" };
  }
  return {
    commodity: tpl.commodity,
    name: tpl.name,
    unit: tpl.unit,
    originFacilityKind: tpl.originFacilityKind ?? "",
    capacityRule: cr,
    stages: (tpl.stages ?? []).map((s) => ({
      key: s.key,
      events: s.events ?? [],
      maxYield: s.maxYield ?? undefined,
      requiredEvidence: s.requiredEvidence ?? [],
    })),
  };
}

function toApiBody(v: TemplateForm) {
  let rule: object | null = null;
  if (v.capacityRule) {
    if (v.capacityRule.type === "yield_per_area") {
      rule = {
        type: capacityRuleTypes.perArea,
        kgPerHaPerSeason: v.capacityRule.kgPerHaPerSeason,
      };
    } else if (v.capacityRule.type === "facility_capacity") {
      rule = { type: capacityRuleTypes.facility, kgPerHaPerSeason: null };
    }
  }
  return {
    commodity: v.commodity,
    name: v.name,
    unit: v.unit,
    originFacilityKind: v.originFacilityKind === "" ? "" : v.originFacilityKind,
    capacityRule: rule,
    stages: v.stages.map((s) => ({
      key: s.key,
      events: s.events,
      maxYield: s.maxYield ?? null,
      requiredEvidence: s.requiredEvidence.filter((e) => e.trim().length > 0),
    })),
  };
}
