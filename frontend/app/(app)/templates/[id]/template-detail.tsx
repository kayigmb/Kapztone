"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { WifiOff } from "lucide-react";
import Link from "next/link";
import { use, useState } from "react";
import { Controller, useForm, useWatch } from "react-hook-form";

import { ChoiceCards } from "@/components/chaintrack/choice-cards";
import { DataWidget } from "@/components/chaintrack/data-widget";
import { DetailHeader } from "@/components/chaintrack/detail-header";
import { TextField } from "@/components/chaintrack/text-field";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { endpoints, type Template } from "@/lib/api";
import {
  capacityRuleTypes,
  copy,
  queryKeys,
  routes,
  screenCopy,
} from "@/lib/consts";
import { liveValidation } from "@/lib/form";
import {
  templateFormSchema,
  type TemplateForm,
  type TemplateFormInput,
} from "@/lib/schemas/template";
import { formatDate, humanize } from "@/lib/verification";

const t = screenCopy.templateDetail;

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

  const tpl = useQuery({
    queryKey: [...queryKeys.template(id)],
    queryFn: () => endpoints.template(id),
  });
  const kinds = useQuery({
    queryKey: queryKeys.facilityKinds,
    queryFn: endpoints.facilityKinds,
  });

  return (
    <DataWidget
      query={tpl}
      className="px-4 py-8 lg:px-8"
      page={{ icon: WifiOff, title: copy.errors.loadTitle }}
    >
      {(data) => (
        <DataWidget query={kinds}>
          {(activeKinds) => <EditForm tpl={data} id={id} kinds={activeKinds} />}
        </DataWidget>
      )}
    </DataWidget>
  );
}

function EditForm({
  tpl,
  id,
  kinds,
}: {
  tpl: Template;
  id: string;
  kinds: { key: string; label: string; sub: string; active: boolean }[];
}) {
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
    setValue,
    formState: { errors },
  } = form;
  const capacityRule = useWatch({ control, name: "capacityRule" });
  const ruleType = capacityRule?.type ?? "none";
  const originKinds = kinds.some((kind) => kind.key === tpl.originFacilityKind)
    ? kinds
    : [
        ...kinds,
        {
          key: tpl.originFacilityKind,
          label: `${humanize(tpl.originFacilityKind ?? "")} · ${screenCopy.templates.kindInactiveLabel}`,
          sub: "",
          active: false,
        },
      ];

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <DetailHeader crumb={tpl.name} title={tpl.name} />

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
                  items={originKinds.map((kind) => ({
                    value: kind.key,
                    label: kind.label,
                  }))}
                  value={field.value}
                  onValueChange={(v) => field.onChange(v)}
                >
                  <SelectTrigger className="min-h-tap w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {originKinds.map((kind) => (
                      <SelectItem
                        key={kind.key}
                        value={kind.key}
                        disabled={!kind.active}
                      >
                        {kind.label}
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

        <TextField
          label={t.requiredEvidence}
          help={t.requiredEvidenceHelp}
          error={errors.requiredEvidence?.message}
          {...register("requiredEvidence")}
        />

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
            href={routes.templates}
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

function fromTemplate(tpl: Template): TemplateFormInput {
  const rule = tpl.capacityRule;
  let cr: TemplateFormInput["capacityRule"] = null;
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
    requiredEvidence: tpl.requiredEvidence.join(", "),
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
    requiredEvidence: v.requiredEvidence,
  };
}
