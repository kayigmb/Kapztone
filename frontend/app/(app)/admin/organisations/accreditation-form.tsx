"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { BadgeCheck, CircleDashed, OctagonX } from "lucide-react";
import { Controller, useForm } from "react-hook-form";

import { useConfirm } from "@/components/chaintrack/confirm";
import { Notice } from "@/components/chaintrack/notice";
import { Loading } from "@/components/chaintrack/page-state";
import { StatusChip } from "@/components/chaintrack/status";
import { TextField } from "@/components/chaintrack/text-field";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { endpoints, problemOf, type AccreditationRow } from "@/lib/api";
import { copy, queryKeys, confirmCopy } from "@/lib/consts";
import { liveValidation } from "@/lib/form";
import {
  accreditationSchema,
  toAccreditations,
  type Accreditation,
} from "@/lib/schemas/org";
import { humanize } from "@/lib/verification";

const t = copy.admin.accreditation;

/** Which certificate types a verified certifier or regulator may issue, and for how long. */
export function AccreditationForm({ orgId }: { orgId: string }) {
  const types = useQuery({
    queryKey: queryKeys.templates,
    queryFn: endpoints.certificateTypes,
  });
  const current = useQuery({
    queryKey: queryKeys.orgAccreditations(orgId),
    queryFn: () => endpoints.accreditations(orgId),
  });

  return (
    <section className="grid gap-3 border-t border-border pt-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h3 className="text-title">{t.title}</h3>
          <p className="text-small text-ink-muted">{t.sub}</p>
        </div>
        {current.data && (
          <StatusChip
            tone={current.data.length ? "verified" : "neutral"}
            icon={current.data.length ? BadgeCheck : CircleDashed}
          >
            {t.current(current.data.length)}
          </StatusChip>
        )}
      </header>
      {types.isPending || current.isPending ? (
        <Loading />
      ) : types.isError || current.isError ? (
        <Notice kind="blocking">
          {problemOf(types.error ?? current.error)}
        </Notice>
      ) : (
        // Mounted once the current accreditation is known, so the form starts from it.
        <Form orgId={orgId} types={types.data} current={current.data} />
      )}
    </section>
  );
}

function Form({
  orgId,
  types,
  current,
}: {
  orgId: string;
  types: string[];
  current: AccreditationRow[];
}) {
  const client = useQueryClient();
  // Types accredited earlier stay choosable even if no template asks for them any more.
  const choices = [
    ...new Set([...types, ...current.map((a) => a.type)]),
  ].sort();
  const {
    control,
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<Accreditation>({
    ...liveValidation,
    resolver: zodResolver(accreditationSchema),
    defaultValues: {
      types: current.map((a) => a.type),
      valid_from:
        current[0]?.valid_from.slice(0, 10) ??
        new Date().toISOString().slice(0, 10),
      valid_to: current[0]?.valid_to.slice(0, 10) ?? "",
    },
  });

  const confirm = useConfirm();

  const save = useMutation({
    mutationFn: (rows: AccreditationRow[]) =>
      endpoints.setAccreditations(orgId, rows),
    onSuccess: () =>
      client.invalidateQueries({
        queryKey: queryKeys.orgAccreditations(orgId),
      }),
  });

  if (choices.length === 0) return <Notice kind="info">{t.noTypes}</Notice>;

  return (
    <form
      onSubmit={handleSubmit(async (a) => {
        const c = confirmCopy.accredit;
        if (
          await confirm({ title: c.title, body: c.body, confirmLabel: c.label })
        )
          save.mutate(toAccreditations(a));
      })}
      noValidate
      className="grid gap-4"
    >
      <fieldset
        className="grid gap-1.5"
        aria-invalid={errors.types ? true : undefined}
      >
        <legend className="mb-1.5 text-sm font-semibold">{t.types}</legend>
        <Controller
          control={control}
          name="types"
          render={({ field }) => (
            <div className="grid gap-1 md:grid-cols-2">
              {choices.map((type) => (
                <label
                  key={type}
                  className="flex min-h-tap cursor-pointer items-center gap-3 rounded-md px-2 hover:bg-surface-sunken"
                >
                  <Checkbox
                    checked={field.value.includes(type)}
                    onCheckedChange={(on) =>
                      field.onChange(
                        on
                          ? [...field.value, type]
                          : field.value.filter((x) => x !== type),
                      )
                    }
                  />
                  <span className="text-sm">{humanize(type)}</span>
                </label>
              ))}
            </div>
          )}
        />
        {errors.types && (
          <p className="flex gap-1.5 text-small font-semibold text-danger">
            <OctagonX className="mt-px size-4 shrink-0" aria-hidden />
            {errors.types.message}
          </p>
        )}
      </fieldset>
      <div className="grid gap-4 md:grid-cols-2">
        <TextField
          label={t.from}
          type="date"
          error={errors.valid_from?.message}
          {...register("valid_from")}
        />
        <TextField
          label={t.until}
          type="date"
          error={errors.valid_to?.message}
          {...register("valid_to")}
        />
      </div>
      <Notice kind="info">{t.notice}</Notice>
      <div className="flex flex-wrap justify-end gap-2">
        <Button
          type="button"
          variant="secondary"
          disabled={save.isPending || current.length === 0}
          onClick={async () => {
            const c = confirmCopy.decline;
            if (
              await confirm({
                title: c.title,
                body: c.body,
                confirmLabel: c.label,
                danger: true,
              })
            )
              save.mutate([]);
          }}
        >
          {t.decline}
        </Button>
        <Button type="submit" disabled={save.isPending}>
          {save.isPending ? t.saving : t.accredit}
        </Button>
      </div>
    </form>
  );
}
