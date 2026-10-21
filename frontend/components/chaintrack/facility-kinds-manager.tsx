"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Pencil, Plus, Trash2 } from "lucide-react";
import { useState } from "react";
import { useForm, useWatch } from "react-hook-form";

import { useConfirm } from "@/components/chaintrack/confirm";
import { DataWidget } from "@/components/chaintrack/data-widget";
import {
  ClientTable,
  actionsColumn,
  column,
  RowAction,
  type ListColumn,
} from "@/components/chaintrack/list-table";
import { TextField } from "@/components/chaintrack/text-field";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { endpoints, problemOf, type FacilityKindRow } from "@/lib/api";
import { confirmCopy, queryKeys, screenCopy } from "@/lib/consts";
import { fill } from "@/lib/format";
import { liveValidation } from "@/lib/form";
import {
  facilityKindSchema,
  type FacilityKindForm,
} from "@/lib/schemas/facility-kind";

const t = screenCopy.templates;

/** Curates the key, label, description and active state offered by both forms. */
export function FacilityKindsManager() {
  const client = useQueryClient();
  const confirm = useConfirm();
  const [editing, setEditing] = useState<FacilityKindRow | null>(null);
  const kinds = useQuery({
    queryKey: [...queryKeys.facilityKinds, "all"],
    queryFn: endpoints.facilityKindsAll,
  });
  const create = useMutation({
    mutationFn: endpoints.createFacilityKind,
    onSuccess: () => {
      client.invalidateQueries({ queryKey: queryKeys.facilityKinds });
      client.invalidateQueries({
        queryKey: [...queryKeys.facilityKinds, "all"],
      });
      setEditing(null);
    },
  });
  const update = useMutation({
    mutationFn: ({
      key,
      body,
    }: {
      key: string;
      body: { label: string; sub: string; active: boolean };
    }) => endpoints.updateFacilityKind(key, body),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: queryKeys.facilityKinds });
      client.invalidateQueries({
        queryKey: [...queryKeys.facilityKinds, "all"],
      });
      setEditing(null);
    },
  });
  const remove = useMutation({
    mutationFn: endpoints.deleteFacilityKind,
    onSuccess: () => {
      client.invalidateQueries({ queryKey: queryKeys.facilityKinds });
      client.invalidateQueries({
        queryKey: [...queryKeys.facilityKinds, "all"],
      });
    },
  });

  const columns: ListColumn<FacilityKindRow>[] = [
    column("label", t.kindLabel, (k) => (
      <span className="font-semibold">{k.label}</span>
    )),
    column("key", t.kindKey, (k) => <span className="font-mono">{k.key}</span>),
    column("sub", t.kindDescription, (k) => k.sub || "—"),
    column("active", t.kindActiveLabel, (k) =>
      k.active ? t.kindActiveLabel : t.kindInactiveLabel,
    ),
    actionsColumn((k) => (
      <div className="flex gap-1">
        <RowAction
          label={t.kindEditAction}
          icon={Pencil}
          onClick={() => setEditing(k)}
        />
        <RowAction
          label={t.kindDeleteAction}
          icon={Trash2}
          danger
          disabled={remove.isPending}
          onClick={async () => {
            const yes = await confirm({
              title: fill(t.kindDeleteTitle, { name: k.label }),
              body: t.kindDeleteBody,
              confirmLabel: t.kindDeleteConfirm,
              danger: true,
            });
            if (yes) remove.mutate(k.key);
          }}
        />
      </div>
    )),
  ];

  return (
    <div className="grid gap-5">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div className="grid gap-1">
          <h2 className="text-title">{t.kindsTitle}</h2>
          <p className="text-sm text-ink-muted">{t.kindsSub}</p>
        </div>
        {!editing && (
          <Button onClick={() => setEditing(EMPTY_KIND)}>
            <Plus aria-hidden />
            {t.kindAdd}
          </Button>
        )}
      </header>

      {editing && (
        <KindEditor
          key={editing.key || "new"}
          initial={editing}
          busy={create.isPending || update.isPending}
          error={create.error || update.error}
          onCancel={() => setEditing(null)}
          onSubmit={(form) => {
            if (editing.key) {
              update.mutate({
                key: editing.key,
                body: { label: form.label, sub: form.sub, active: form.active },
              });
            } else {
              create.mutate({
                key: form.key,
                label: form.label,
                sub: form.sub,
              });
            }
          }}
        />
      )}

      <DataWidget query={kinds}>
        {(rows) => (
          <ClientTable
            id="facility-kinds"
            label={t.kindsTitle}
            columns={columns}
            rows={rows}
            rowKey={(k) => k.key}
            empty={t.kindEmpty}
          />
        )}
      </DataWidget>
    </div>
  );
}

const EMPTY_KIND: FacilityKindRow = {
  key: "",
  label: "",
  sub: "",
  active: true,
};

function KindEditor({
  initial,
  busy,
  error,
  onCancel,
  onSubmit,
}: {
  initial: FacilityKindRow;
  busy: boolean;
  error: unknown;
  onCancel: () => void;
  onSubmit: (value: FacilityKindForm) => void;
}) {
  const form = useForm<FacilityKindForm>({
    ...liveValidation,
    resolver: zodResolver(facilityKindSchema),
    defaultValues: initial,
  });
  const active = useWatch({ control: form.control, name: "active" });
  const t2 = screenCopy.templates;
  return (
    <form
      noValidate
      className="grid gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6"
      onSubmit={form.handleSubmit(onSubmit)}
    >
      {error != null && (
        <p role="alert" className="text-sm text-danger-ink">
          {problemOf(error)}
        </p>
      )}
      <div className="grid gap-4 md:grid-cols-3">
        <TextField
          label={t2.kindKey}
          help={t2.kindKeyHelp}
          disabled={!!initial.key}
          autoCapitalize="none"
          autoComplete="off"
          error={form.formState.errors.key?.message}
          {...form.register("key")}
        />
        <TextField
          label={t2.kindLabel}
          error={form.formState.errors.label?.message}
          {...form.register("label")}
        />
        <TextField
          label={t2.kindDescription}
          error={form.formState.errors.sub?.message}
          {...form.register("sub")}
        />
      </div>
      {initial.key && (
        <label className="flex min-h-tap items-center gap-3 text-sm">
          <Checkbox
            checked={active}
            onCheckedChange={(checked) =>
              form.setValue("active", checked === true)
            }
          />
          {t2.kindActive}
        </label>
      )}
      <div className="flex justify-end gap-2">
        <Button type="button" variant="secondary" onClick={onCancel}>
          {confirmCopy.cancel}
        </Button>
        <Button type="submit" loading={busy} loadingText={t2.kindSaving}>
          {t2.kindSave}
        </Button>
      </div>
    </form>
  );
}
