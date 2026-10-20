"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Lock, Plus, WifiOff, Eye } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Controller, useFieldArray, useForm } from "react-hook-form";

import { ChoiceCards } from "@/components/chaintrack/choice-cards";
import {
  ListTable,
  RowAction,
  actionsColumn,
  column,
  type ListColumn,
} from "@/components/chaintrack/list-table";
import { ListBar } from "@/components/chaintrack/list-controls";
import { PageHeader } from "@/components/chaintrack/page-header";
import { Notice } from "@/components/chaintrack/notice";
import { Loading, PageState } from "@/components/chaintrack/page-state";
import { TextField } from "@/components/chaintrack/text-field";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  endpoints,
  problemOf,
  useMeGuard,
  usePaged,
  type Template,
} from "@/lib/api";
import {
  capacityRuleTypes,
  copy,
  facilityKinds,
  permissions,
  queryKeys,
  routes,
  templateEvent,
  templateEvents,
  screenCopy,
} from "@/lib/consts";
import { liveValidation } from "@/lib/form";
import { useListParams } from "@/lib/list-params";
import {
  templateFormSchema,
  type TemplateForm,
  type TemplateFormInput,
} from "@/lib/schemas/template";
import { formatDate, humanize } from "@/lib/verification";

const t = screenCopy.templates;

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
  const allowed =
    me.ready && me.data!.permissions.includes(permissions.templateRead);
  const canCreate =
    me.ready && me.data!.permissions.includes(permissions.templateManage);

  const list = useListParams("name");
  const templates = usePaged(
    queryKeys.templates,
    endpoints.templatesPage,
    list.query(),
  );

  const create = useMutation({
    mutationFn: (body: TemplateForm) =>
      endpoints.createTemplate(toApiBody(body)),
    meta: { inPlace: true },
    onSuccess: ({ data }) => {
      client.invalidateQueries({ queryKey: queryKeys.templates });
      router.push(routes.adminTemplate(data.id));
    },
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

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <PageHeader
        title={t.title}
        sub={t.sub}
        actions={
          canCreate && (
            <Button onClick={() => setShowForm(!showForm)}>
              <Plus aria-hidden />
              {t.new}
            </Button>
          )
        }
      />

      {showForm && (
        <NewTemplateForm
          onCancel={() => setShowForm(false)}
          onSubmit={(v) => create.mutate(v)}
          busy={create.isPending}
          error={create.isError ? problemOf(create.error) : undefined}
        />
      )}

      <ListTable
        label={t.title}
        columns={columns}
        rows={templates.data?.data ?? []}
        total={templates.data?.total ?? 0}
        list={list}
        rowKey={(x) => x.id}
        loading={templates.isPending}
        error={templates.isError ? problemOf(templates.error) : undefined}
        empty={list.search ? copy.list.noMatch : t.noTemplates}
        toolbar={<ListBar list={list} />}
      />
    </main>
  );
}

const columns: ListColumn<Template>[] = [
  column(
    "name",
    t.listName,
    (x) => (
      <Link
        href={routes.adminTemplate(x.id)}
        className="font-semibold text-primary"
      >
        {x.name}
      </Link>
    ),
    { sortable: true },
  ),
  column("commodity", t.listCommodity, (x) => humanize(x.commodity), {
    sortable: true,
  }),
  column("unit", t.listUnit, (x) => (
    <span className="font-mono">{x.unit}</span>
  )),
  column("origin", t.listOrigin, (x) =>
    originLabel(x.originFacilityKind ?? ""),
  ),
  column("stages", t.listStageCount, (x) => x.stages?.length ?? 0),
  column(
    "created_at",
    t.listUpdated,
    (x) => formatDate(x.updated_at ?? x.created_at ?? ""),
    { sortable: true },
  ),
  actionsColumn((x) => (
    <RowAction label={t.open} icon={Eye} href={routes.adminTemplate(x.id)} />
  )),
];

function originLabel(v: string) {
  if (!v) return t.any;
  return facilityKinds.find((k) => k.value === v)?.label ?? v;
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
    ...liveValidation,
    resolver: zodResolver(templateFormSchema),
    defaultValues: {
      commodity: "",
      name: "",
      unit: "kg",
      originFacilityKind: "",
      capacityRule: null,
      stages: [
        {
          key: "origin",
          events: [templateEvent.issue],
          maxYield: undefined,
          requiredEvidence: [],
        },
        {
          key: "transport",
          events: [templateEvent.handover],
          maxYield: undefined,
          requiredEvidence: [],
        },
      ],
    },
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
    <form
      onSubmit={handleSubmit(onSubmit)}
      noValidate
      className="grid gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6"
    >
      <h2 className="text-section">{t.new}</h2>

      {error && <Notice kind="blocking">{error}</Notice>}

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
          <span className="text-sm font-semibold">{t.originFacilityKind}</span>
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
                          if (c) {
                            setValue(`stages.${i}.events`, [
                              ...current,
                              ev.value,
                            ]);
                          } else {
                            setValue(
                              `stages.${i}.events`,
                              current.filter((x) => x !== ev.value),
                            );
                          }
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

      <div className="flex flex-wrap justify-end gap-3">
        <Button
          type="button"
          variant="secondary"
          onClick={onCancel}
          disabled={busy}
        >
          Cancel
        </Button>
        <Button type="submit" disabled={busy}>
          {busy ? "Saving…" : "Save"}
        </Button>
      </div>
    </form>
  );
}
