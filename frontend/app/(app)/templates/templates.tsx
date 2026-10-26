"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus, Eye } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Controller, useForm, useWatch } from "react-hook-form";

import { ChoiceCards } from "@/components/chaintrack/choice-cards";
import { DataWidget } from "@/components/chaintrack/data-widget";
import { FacilityKindsManager } from "@/components/chaintrack/facility-kinds-manager";
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
import { TextField } from "@/components/chaintrack/text-field";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
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
  useMe,
  usePaged,
  type Template,
} from "@/lib/api";
import {
  capacityRuleTypes,
  copy,
  permissions,
  queryKeys,
  routes,
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

const ruleOptions = [
  { value: "none", label: t.none },
  { value: "perArea", label: t.perArea },
  { value: "facilityCapacity", label: t.facilityCapacity },
];

export function Templates() {
  const router = useRouter();
  const client = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [tab, setTab] = useState("templates");

  const me = useMe().data!; // the layout only renders once it is loaded
  const canCreate = me.permissions.includes(permissions.templateManage);

  const list = useListParams("name");
  const templates = usePaged(
    queryKeys.templates,
    endpoints.templatesPage,
    list.query(),
  );
  const kinds = useQuery({
    queryKey: queryKeys.facilityKinds,
    queryFn: endpoints.facilityKinds,
    enabled: canCreate,
  });

  const create = useMutation({
    mutationFn: (body: TemplateForm) =>
      endpoints.createTemplate(toApiBody(body)),
    meta: { inPlace: true },
    onSuccess: ({ data }) => {
      client.invalidateQueries({ queryKey: queryKeys.templates });
      router.push(routes.template(data.id));
    },
  });

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

      <Tabs value={tab} onValueChange={(v) => setTab(v)}>
        <TabsList>
          <TabsTrigger value="templates">{t.tabTemplates}</TabsTrigger>
          {canCreate && (
            <TabsTrigger value="facility-kinds">
              {t.tabFacilityKinds}
            </TabsTrigger>
          )}
        </TabsList>
        <TabsContent value="templates" className="grid gap-6">
          {showForm && (
            <DataWidget query={kinds}>
              {(rows) => (
                <NewTemplateForm
                  kinds={rows}
                  onCancel={() => setShowForm(false)}
                  onSubmit={(v) => create.mutate(v)}
                  busy={create.isPending}
                  error={create.isError ? problemOf(create.error) : undefined}
                />
              )}
            </DataWidget>
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
        </TabsContent>
        {canCreate && (
          <TabsContent value="facility-kinds">
            <FacilityKindsManager />
          </TabsContent>
        )}
      </Tabs>
    </main>
  );
}

const columns: ListColumn<Template>[] = [
  column(
    "name",
    t.listName,
    (x) => (
      <Link href={routes.template(x.id)} className="font-semibold text-primary">
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
  column("evidence", t.listEvidence, (x) => x.requiredEvidence.length),
  column(
    "created_at",
    t.listUpdated,
    (x) => formatDate(x.updated_at ?? x.created_at ?? ""),
    { sortable: true },
  ),
  actionsColumn((x) => (
    <RowAction label={t.open} icon={Eye} href={routes.template(x.id)} />
  )),
];

function originLabel(v: string) {
  return humanize(v);
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

function NewTemplateForm({
  kinds,
  onCancel,
  onSubmit,
  busy,
  error,
}: {
  kinds: { key: string; label: string; sub: string }[];
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
      originFacilityKind: kinds[0]?.key ?? "",
      capacityRule: null,
      requiredEvidence: "",
    },
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
                items={kinds.map((k) => ({ value: k.key, label: k.label }))}
                value={field.value}
                onValueChange={(v) => field.onChange(v)}
              >
                <SelectTrigger className="min-h-tap w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {kinds.map((kind) => (
                    <SelectItem key={kind.key} value={kind.key}>
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
