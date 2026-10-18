"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Layers, Lock, Plus, WifiOff } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Controller, useFieldArray, useForm } from "react-hook-form";
import { toast } from "sonner";

import { ChoiceCards } from "@/components/chaintrack/choice-cards";
import { Notice } from "@/components/chaintrack/notice";
import { Loading, PageState } from "@/components/chaintrack/page-state";
import { StatusChip } from "@/components/chaintrack/status";
import { TextField } from "@/components/chaintrack/text-field";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { endpoints, problemOf, useMeGuard, type Template } from "@/lib/api";
import {
  capacityRuleTypes,
  copy,
  facilityKinds,
  permissions,
  queryKeys,
  routes,
  templateEvents,
} from "@/lib/consts";
import { templateFormSchema, type TemplateForm, type TemplateFormInput } from "@/lib/schemas/template";
import { formatDate, humanize } from "@/lib/verification";

const t = {
  title: "Product templates",
  sub: "Define how a commodity moves: stages, events, evidence and capacity rule.",
  new: "New template",
  name: "Name",
  commodity: "Commodity",
  unit: "Unit",
  originFacilityKind: "Origin facility kind",
  any: "Any",
  capacityRule: "Capacity rule",
  none: "None",
  perArea: "Per area (yield per hectare per season)",
  facilityCapacity: "Facility capacity",
  kgPerHaPerSeason: "Kg per ha per season",
  stages: "Stages",
  addStage: "Add stage",
  stageKey: "Stage key",
  events: "Events",
  maxYield: "Max yield (%)",
  requiredEvidence: "Required evidence (comma separated)",
  removeStage: "Remove",
  atLeastOneEvent: "At least one event",
  listName: "Name",
  listCommodity: "Commodity",
  listUnit: "Unit",
  listOrigin: "Origin facility",
  listStageCount: "Stage count",
  listUpdated: "Updated",
  noTemplates: "No templates yet.",
  notAllowedTitle: "You can't manage templates.",
  notAllowedBody: "Only the platform admin can manage product templates.",
  created: "Template created",
};

const originOptions = [
  { value: "", label: t.any },
  ...facilityKinds.map((k) => ({ value: k.value, label: k.label })),
];

const ruleOptions = [
  { value: "none", label: t.none },
  { value: "perArea", label: t.perArea },
  { value: "facilityCapacity", label: t.facilityCapacity },
];

export function Templates() {
  const router = useRouter();
  const client = useQueryClient();
  const [showForm, setShowForm] = useState(false);

  const me = useMeGuard((me) => (me ? undefined : routes.signin));
  const allowed = me.ready && me.data!.permissions.includes(permissions.templateRead);

  const templates = useQuery({
    queryKey: queryKeys.templates,
    queryFn: endpoints.templatesList,
    enabled: allowed,
  });

  const create = useMutation({
    mutationFn: (body: TemplateForm) => endpoints.createTemplate(toApiBody(body)),
    onSuccess: (tpl) => {
      toast.success(t.created);
      client.invalidateQueries({ queryKey: queryKeys.templates });
      router.push(routes.adminTemplate(tpl.id));
    },
    onError: (err) => toast.error(problemOf(err)),
  });

  if (me.isError) {
    return (
      <PageState icon={WifiOff} tone="danger" title={copy.errors.loadTitle} retry={() => me.refetch()}>
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


  return (
    <main className="grid content-start gap-6 px-4 py-6 lg:px-8 lg:py-8">
      <header className="flex flex-wrap items-end justify-between gap-3">
        <div className="grid gap-1">
          <h1 className="text-[22px] font-semibold lg:text-page">{t.title}</h1>
          <p className="text-sm text-ink-muted">{t.sub}</p>
        </div>
        <Button size="scan" onClick={() => setShowForm(!showForm)}>
          <Plus aria-hidden />
          {t.new}
        </Button>
      </header>

      {templates.isError && (
        <Notice kind="blocking">{problemOf(templates.error)}</Notice>
      )}

      {showForm && (
        <NewTemplateForm
          onCancel={() => setShowForm(false)}
          onSubmit={(v) => create.mutate(v)}
          busy={create.isPending}
          error={create.isError ? problemOf(create.error) : undefined}
        />
      )}

      {templates.isPending ? (
        <Loading />
      ) : (
        <div className="overflow-x-auto rounded-lg border border-border bg-surface shadow-card">
          <table className="min-w-full divide-y divide-border text-sm">
            <thead className="bg-surface-sunken">
              <tr>
                <th className="px-4 py-3 text-left font-semibold">{t.listName}</th>
                <th className="px-4 py-3 text-left font-semibold">{t.listCommodity}</th>
                <th className="px-4 py-3 text-left font-semibold">{t.listUnit}</th>
                <th className="px-4 py-3 text-left font-semibold">{t.listOrigin}</th>
                <th className="px-4 py-3 text-left font-semibold">{t.listStageCount}</th>
                <th className="px-4 py-3 text-left font-semibold">{t.listUpdated}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {(templates.data ?? []).length === 0 && (
                <tr>
                  <td colSpan={6} className="px-4 py-6 text-center text-ink-muted">
                    {t.noTemplates}
                  </td>
                </tr>
              )}
              {(templates.data ?? []).map((tpl: Template) => (
                <tr key={tpl.id} className="hover:bg-surface-sunken">
                  <td className="px-4 py-3">
                    <Link href={routes.adminTemplate(tpl.id)} className="font-semibold text-primary">
                      {tpl.name}
                    </Link>
                  </td>
                  <td className="px-4 py-3">{humanize(tpl.commodity)}</td>
                  <td className="px-4 py-3 font-mono">{tpl.unit}</td>
                  <td className="px-4 py-3">{originLabel(tpl.originFacilityKind ?? "")}</td>
                  <td className="px-4 py-3">{tpl.stages?.length ?? 0}</td>
                  <td className="px-4 py-3 text-ink-muted">{formatDate(tpl.updated_at ?? tpl.created_at ?? "")}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </main>
  );
}

function originLabel(v: string) {
  if (!v) return t.any;
  return facilityKinds.find((k) => k.value === v)?.label ?? v;
}

function toApiBody(v: TemplateForm): object {
  let rule: object | null = null;
  if (v.capacityRule) {
    if (v.capacityRule.type === "yield_per_area") {
      rule = { type: capacityRuleTypes.perArea, kgPerHaPerSeason: v.capacityRule.kgPerHaPerSeason };
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

function NewTemplateForm({
  onCancel,
  onSubmit,
  busy,
  error,
}: {
  onCancel: () => void;
  onSubmit: (v: TemplateForm) => void;
  busy: boolean;
  error?: string;
}) {
  const form = useForm<TemplateFormInput, unknown, TemplateForm>({
    resolver: zodResolver(templateFormSchema),
    defaultValues: {
      commodity: "",
      name: "",
      unit: "kg",
      originFacilityKind: "",
      capacityRule: null,
      stages: [
        { key: "ISSUE", events: ["ISSUE"], maxYield: undefined, requiredEvidence: ["organic"] },
        { key: "HANDOVER", events: ["HANDOVER"], maxYield: undefined, requiredEvidence: [] },
      ],
    },
  });

  const { control, register, handleSubmit, watch, setValue, formState: { errors } } = form;
  const { fields, append, remove } = useFieldArray({ control, name: "stages" });
  const ruleType = watch("capacityRule")?.type ?? "none";

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      noValidate
      className="grid gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6"
    >
      <h2 className="text-section">{t.new}</h2>

      {error && <Notice kind="blocking">{error}</Notice>}

      <div className="grid gap-4 md:grid-cols-2">
        <TextField label={t.name} error={errors.name?.message} {...register("name")} />
        <TextField label={t.commodity} error={errors.commodity?.message} {...register("commodity")} />
        <TextField label={t.unit} error={errors.unit?.message} {...register("unit")} />
        <div className="grid gap-1.5">
          <span className="text-sm font-semibold">{t.originFacilityKind}</span>
          <Controller
            control={control}
            name="originFacilityKind"
            render={({ field }) => (
              <Select items={originOptions} value={field.value} onValueChange={(v) => field.onChange(v)}>
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
          value={ruleType === "yield_per_area" ? "perArea" : ruleType === "facility_capacity" ? "facilityCapacity" : "none"}
          onChange={(e) => {
            const v = e.target.value;
            if (v === "none") setValue("capacityRule", null);
            else if (v === "perArea") setValue("capacityRule", { type: "yield_per_area", kgPerHaPerSeason: 0 });
            else if (v === "facilityCapacity") setValue("capacityRule", { type: "facility_capacity" });
          }}
        />
        {ruleType === "yield_per_area" && (
          <TextField
            label={t.kgPerHaPerSeason}
            type="number"
            inputMode="numeric"
            mono
            error={errors.capacityRule?.kgPerHaPerSeason?.message}
            {...register("capacityRule.kgPerHaPerSeason", { valueAsNumber: true })}
          />
        )}
      </div>

      <div className="grid gap-3">
        <div className="flex items-center justify-between">
          <span className="text-sm font-semibold">{t.stages}</span>
          <Button
            type="button"
            variant="secondary"
            onClick={() => append({ key: "", events: [], maxYield: undefined, requiredEvidence: [] })}
            disabled={fields.length >= 10}
          >
            <Plus aria-hidden />
            {t.addStage}
          </Button>
        </div>
        {fields.map((field, i) => (
          <div key={field.id} className="grid gap-3 rounded-md border border-border p-3">
            <div className="flex flex-wrap items-end justify-between gap-2">
              <TextField
                label={t.stageKey}
                className="flex-1 min-w-[120px]"
                error={errors.stages?.[i]?.key?.message}
                {...register(`stages.${i}.key` as const)}
              />
              {fields.length > 1 && (
                <Button type="button" variant="danger-outline" onClick={() => remove(i)}>
                  {t.removeStage}
                </Button>
              )}
            </div>
            <div className="grid gap-2">
              <span className="text-sm font-semibold">{t.events}</span>
              <div className="flex flex-wrap gap-3">
                {templateEvents.map((ev) => {
                  const checked = watch(`stages.${i}.events`)?.includes(ev.value) ?? false;
                  return (
                    <label key={ev.value} className="flex items-center gap-2">
                      <Checkbox
                        checked={checked}
                        onCheckedChange={(c) => {
                          const current = watch(`stages.${i}.events`) ?? [];
                          if (c) {
                            setValue(`stages.${i}.events`, [...current, ev.value]);
                          } else {
                            setValue(`stages.${i}.events`, current.filter((x) => x !== ev.value));
                          }
                        }}
                      />
                      <span className="text-sm">{ev.label}</span>
                    </label>
                  );
                })}
              </div>
              {errors.stages?.[i]?.events?.message && (
                <p className="text-small font-semibold text-danger">{errors.stages?.[i]?.events?.message}</p>
              )}
            </div>
            <div className="grid gap-4 md:grid-cols-2">
              <TextField
                label={t.maxYield}
                type="number"
                inputMode="numeric"
                mono
                error={errors.stages?.[i]?.maxYield?.message}
                {...register(`stages.${i}.maxYield` as const, { valueAsNumber: true })}
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

      <div className="flex flex-wrap justify-end gap-3">
        <Button type="button" variant="secondary" onClick={onCancel} disabled={busy}>
          Cancel
        </Button>
        <Button type="submit" disabled={busy}>
          {busy ? "Saving…" : "Save"}
        </Button>
      </div>
    </form>
  );
}
