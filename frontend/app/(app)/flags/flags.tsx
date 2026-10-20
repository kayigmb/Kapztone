"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  CircleCheck,
  Clock,
  Flag as FlagIcon,
  OctagonX,
  ShieldAlert,
} from "lucide-react";
import Link from "next/link";
import { parseAsString, parseAsStringLiteral } from "nuqs";
import { useState } from "react";
import { useForm } from "react-hook-form";

import { useConfirm } from "@/components/chaintrack/confirm";
import { DataWidget } from "@/components/chaintrack/data-widget";
import { ListBar, ListPager } from "@/components/chaintrack/list-controls";
import { PageHeader } from "@/components/chaintrack/page-header";
import { Notice } from "@/components/chaintrack/notice";

import { KeyValue } from "@/components/chaintrack/passport-card";
import { SigningPanel } from "@/components/chaintrack/signing";
import { StatusChip } from "@/components/chaintrack/status";
import { Button } from "@/components/ui/button";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { endpoints, problemOf, useMe, usePaged, type Flag } from "@/lib/api";
import {
  copy,
  flagStatus,
  flagTabs,
  listSorts,
  orgStatus,
  permissions,
  queryKeys,
  routes,
  type flagOutcomes,
  confirmCopy,
} from "@/lib/consts";
import { liveValidation } from "@/lib/form";
import { useListParams } from "@/lib/list-params";
import {
  resolveFlagFormSchema,
  type ResolveFlagForm,
} from "@/lib/schemas/flag";
import { cn } from "@/lib/utils";
import { formatDate } from "@/lib/verification";

const t = copy.flags;
type Outcome = (typeof flagOutcomes)[number];

function FlagChip({ status }: { status: string }) {
  if (status === flagStatus.cleared)
    return (
      <StatusChip tone="verified" icon={CircleCheck}>
        {t.status.cleared}
      </StatusChip>
    );
  if (status === flagStatus.confirmed)
    return (
      <StatusChip tone="danger" icon={OctagonX}>
        {t.status.confirmed}
      </StatusChip>
    );
  return (
    <StatusChip tone="warning" icon={Clock}>
      {t.status.open}
    </StatusChip>
  );
}

const listFilters = {
  tab: parseAsStringLiteral(flagTabs).withDefault(flagTabs[0]),
  flag: parseAsString,
};

export function FlagQueue() {
  const state = useListParams("-created_at", listFilters);
  const tab = state.filters.tab;
  const selected = state.filters.flag ?? "";
  // "resolved" is cleared or confirmed (the API's list filter).
  const list = usePaged(
    queryKeys.flags,
    endpoints.flagsPage,
    state.query({ status: tab }),
  );
  const openCount = useQuery({
    queryKey: [...queryKeys.flags, "count", flagStatus.open],
    queryFn: () =>
      endpoints
        .flagsPage({ page: 1, size: 1, filters: { status: flagStatus.open } })
        .then((p) => p.total),
  });
  const picked = useQuery({
    queryKey: queryKeys.flag(selected),
    queryFn: () => endpoints.flag(selected),
    enabled: !!selected,
  });
  const open = openCount.data ?? 0;
  const go = (q: { flag: string }) => state.setFilters(q, { keepPage: true });
  const current = picked.data;

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <PageHeader title={t.title} sub={t.sub(open)} />
      <div className="grid gap-4 lg:grid-cols-[minmax(--spacing(70),0.8fr)_1.5fr] lg:items-start">
        <section className="min-w-0 overflow-hidden rounded-lg border border-border bg-surface shadow-card">
          <div className="border-b border-border p-3">
            <Tabs
              value={tab}
              onValueChange={(v) =>
                state.setFilters({ tab: v as typeof tab, flag: null })
              }
            >
              <TabsList>
                {flagTabs.map((x) => (
                  <TabsTrigger key={x} value={x}>
                    {t.tabs[x]}
                  </TabsTrigger>
                ))}
              </TabsList>
            </Tabs>
            <ListBar className="mt-3" list={state} sorts={listSorts.date} />
          </div>
          <DataWidget query={list}>
            {(page) =>
              page.data.length === 0 ? (
                <p className="p-5 text-sm text-ink-muted">
                  {tab === "open" ? t.empty : t.emptyResolved}
                </p>
              ) : (
                <ul className="divide-y divide-border">
                  {page.data.map((f) => (
                <li key={f.id}>
                  <button
                    type="button"
                    onClick={() => go({ flag: f.id })}
                    aria-current={f.id === selected ? "true" : undefined}
                    className={cn(
                      "grid w-full grid-cols-[auto_1fr] gap-3 px-4 py-3 text-left",
                      f.id === selected
                        ? "bg-primary-soft shadow-selected"
                        : "hover:bg-surface-sunken",
                    )}
                  >
                    <span
                      className={cn(
                        "grid size-8 place-items-center rounded-md",
                        f.status === flagStatus.open
                          ? "bg-warning-soft text-warning-ink"
                          : "bg-surface-sunken text-ink-muted",
                      )}
                    >
                      <FlagIcon className="size-4" aria-hidden />
                    </span>
                    <span className="grid min-w-0 gap-0.5">
                      <span className="line-clamp-2 text-sm font-semibold">
                        {f.reason}
                      </span>
                      <span className="text-small text-ink-muted">
                        {t.source[f.raised_by ?? "user"] ?? f.raised_by} ·{" "}
                        {formatDate(f.created_at)}
                      </span>
                    </span>
                      </button>
                    </li>
                  ))}
                </ul>
              )
            }
          </DataWidget>
          <div className="border-t border-border px-4 py-2 empty:hidden">
            <ListPager list={state} total={list.data?.total ?? 0} />
          </div>
        </section>
        {current ? (
          <Investigation key={current.id} f={current} />
        ) : (
          <section className="hidden rounded-lg border border-dashed border-border p-8 text-center text-sm text-ink-muted lg:block">
            {t.selectOne}
          </section>
        )}
      </div>
    </main>
  );
}

function Investigation({ f }: { f: Flag }) {
  const me = useMe().data!;
  const client = useQueryClient();
  const [deciding, setDeciding] = useState<Outcome>();
  const lot = useQuery({
    queryKey: queryKeys.verify(f.passport_id),
    queryFn: () => endpoints.verify(f.passport_id),
  });
  // Descendants are readable to the platform; a regulator may not see another organisation's tree.
  const tree = useQuery({
    queryKey: [...queryKeys.flag(f.id), "descendants"],
    queryFn: () => endpoints.genealogy(f.passport_id, "descendants"),
    retry: false,
  });
  const affected = tree.data ? tree.data.nodes.length : undefined;
  const refresh = () => client.invalidateQueries({ queryKey: queryKeys.flags });

  const confirm = useConfirm();

  const assign = useMutation({
    mutationFn: (userId: string | null) => endpoints.assignFlag(f.id, userId),
    onSuccess: refresh,
  });
  const {
    register,
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<ResolveFlagForm>({
    ...liveValidation,
    resolver: zodResolver(resolveFlagFormSchema),
    defaultValues: { outcome: flagStatus.cleared, note: "" },
  });
  const resolve = useMutation({
    mutationFn: ({ v, pin }: { v: ResolveFlagForm; pin: string }) =>
      endpoints.resolveFlag(f.id, v, pin),
    meta: { inPlace: true },
    onSuccess: () => {
      refresh();
      setDeciding(undefined);
    },
  });
  const code = lot.data?.passport.code ?? f.passport_id;
  // Mirrors the API: the platform (flag:resolve) or a verified regulator investigates; others only read.
  const investigator =
    me.permissions.includes(permissions.flagResolve) ||
    (me.organisation?.kind === "regulator" &&
      me.organisation.status === orgStatus.verified);

  return (
    <section className="grid min-w-0 gap-5 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div className="grid gap-1">
          <FlagChip status={f.status} />
          <h2 className="text-title">{f.reason}</h2>
          <span className="text-small text-ink-muted">
            {t.raised(formatDate(f.created_at))}
          </span>
        </div>
      </header>
      <KeyValue
        facts={[
          {
            label: t.passport,
            value: (
              <Link className="text-primary" href={routes.passport(code)}>
                {code}
              </Link>
            ),
            mono: true,
          },
          {
            label: t.raisedBy,
            value: t.source[f.raised_by ?? "user"] ?? f.raised_by,
          },
          { label: t.assignee, value: f.assignee_name ?? t.unassigned },
          {
            label: t.affected,
            value: affected !== undefined ? String(affected) : "—",
          },
          ...(f.resolution_note
            ? [{ label: t.resolution, value: f.resolution_note, wide: true }]
            : []),
        ]}
      />

      {f.status === flagStatus.open && !investigator && (
        <Notice kind="info">{t.resolveNotYours}</Notice>
      )}
      {f.status === flagStatus.open && investigator && (
        <>
          <div className="flex flex-wrap gap-2">
            {f.assignee_user_id === me.user.id ? (
              <Button
                variant="secondary"
                onClick={async () => {
                  const c = confirmCopy.flagUnassign;
                  if (
                    await confirm({
                      title: c.title,
                      body: c.body,
                      confirmLabel: c.label,
                    })
                  )
                    assign.mutate(null);
                }}
                loading={assign.isPending}
              >
                {t.unassign}
              </Button>
            ) : (
              <Button
                variant="secondary"
                onClick={async () => {
                  const c = confirmCopy.flagAssign;
                  if (
                    await confirm({
                      title: c.title,
                      body: c.body,
                      confirmLabel: c.label,
                    })
                  )
                    assign.mutate(me.user.id);
                }}
                loading={assign.isPending}
              >
                {t.assignMe}
              </Button>
            )}
          </div>
          <p className="text-sm text-ink-muted">{t.resolveIntro}</p>
          {!deciding ? (
            <div className="flex flex-wrap justify-end gap-2 border-t border-border pt-4">
              <Button
                variant="secondary"
                onClick={() => {
                  setDeciding(flagStatus.cleared);
                  setValue("outcome", flagStatus.cleared);
                }}
              >
                {t.clear}
              </Button>
              <Button
                variant="danger-soft"
                onClick={() => {
                  setDeciding(flagStatus.confirmed);
                  setValue("outcome", flagStatus.confirmed);
                }}
              >
                <ShieldAlert aria-hidden />
                {t.confirmFraud}
              </Button>
            </div>
          ) : (
            <div
              className={cn(
                "grid gap-4 rounded-lg border p-4",
                deciding === flagStatus.confirmed
                  ? "border-danger bg-danger-soft"
                  : "border-border",
              )}
            >
              <h3 className="text-title">
                {deciding === flagStatus.confirmed
                  ? t.fraudTitle
                  : t.clearTitle}
              </h3>
              {deciding === flagStatus.confirmed ? (
                <>
                  <p className="text-sm">
                    {affected !== undefined
                      ? t.fraudBody(affected)
                      : t.fraudBodyNone}
                  </p>
                  <Notice kind="blocking">{t.fraudWarning}</Notice>
                </>
              ) : (
                <p className="text-sm text-ink-muted">{t.clearBody}</p>
              )}
              <label className="grid gap-1.5">
                <span className="text-sm font-semibold">
                  {deciding === flagStatus.confirmed
                    ? t.fraudReason
                    : t.clearNote}
                </span>
                <Textarea
                  {...register("note")}
                  aria-invalid={errors.note ? true : undefined}
                />
                <span
                  className={cn(
                    "text-small",
                    errors.note
                      ? "font-semibold text-danger"
                      : "text-ink-muted",
                  )}
                >
                  {errors.note
                    ? deciding === flagStatus.confirmed
                      ? t.fraudReasonEmpty
                      : t.clearNoteEmpty
                    : deciding === flagStatus.confirmed
                      ? t.fraudReasonHelp
                      : t.clearNoteHelp}
                </span>
              </label>
              <SigningPanel
                button={
                  deciding === flagStatus.confirmed
                    ? t.fraudConfirm(affected ?? 1)
                    : t.clear
                }
                busy={resolve.isPending}
                error={resolve.isError ? problemOf(resolve.error) : undefined}
                onSign={(pin) =>
                  handleSubmit((v) => resolve.mutate({ v, pin }))()
                }
              />
              <Button
                variant="secondary"
                className="w-fit"
                onClick={() => setDeciding(undefined)}
              >
                {copy.handover.back}
              </Button>
            </div>
          )}
        </>
      )}
    </section>
  );
}
