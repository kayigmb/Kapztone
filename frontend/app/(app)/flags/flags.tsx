"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CircleCheck, Clock, Flag as FlagIcon, OctagonX, ShieldAlert } from "lucide-react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { Notice } from "@/components/chaintrack/notice";
import { Loading } from "@/components/chaintrack/page-state";
import { KeyValue } from "@/components/chaintrack/passport-card";
import { SigningPanel } from "@/components/chaintrack/signing";
import { StatusChip } from "@/components/chaintrack/status";
import { Button } from "@/components/ui/button";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { endpoints, problemOf, useMe, type Flag } from "@/lib/api";
import { copy, flagStatus, orgStatus, permissions, queryKeys, routes, type flagOutcomes } from "@/lib/consts";
import { resolveFlagFormSchema, type ResolveFlagForm } from "@/lib/schemas/flag";
import { cn } from "@/lib/utils";
import { formatDate } from "@/lib/verification";

const t = copy.flags;
type Outcome = (typeof flagOutcomes)[number];

function FlagChip({ status }: { status: string }) {
  if (status === flagStatus.cleared) return <StatusChip tone="verified" icon={CircleCheck}>{t.status.cleared}</StatusChip>;
  if (status === flagStatus.confirmed) return <StatusChip tone="danger" icon={OctagonX}>{t.status.confirmed}</StatusChip>;
  return <StatusChip tone="warning" icon={Clock}>{t.status.open}</StatusChip>;
}

export function FlagQueue() {
  const params = useSearchParams();
  const router = useRouter();
  const tab = params.get("tab") === "resolved" ? "resolved" : "open";
  const selected = params.get("flag");
  const list = useQuery({ queryKey: queryKeys.flags, queryFn: () => endpoints.flagsList() });
  const rows = (list.data ?? []).filter((f) => (tab === "open") === (f.status === flagStatus.open));
  const open = list.data?.filter((f) => f.status === flagStatus.open).length ?? 0;
  const go = (q: Record<string, string>) => router.replace(`${routes.flags}?${new URLSearchParams({ tab, ...q })}`);
  const current = list.data?.find((f) => f.id === selected);

  return (
    <main className="grid content-start gap-6 p-4 lg:p-8">
      <header className="grid gap-1">
        <h1 className="text-[22px] font-semibold lg:text-page">{t.title}</h1>
        <p className="text-sm text-ink-muted">{t.sub(open)}</p>
      </header>
      <div className="grid gap-4 lg:grid-cols-[minmax(280px,0.8fr)_1.5fr] lg:items-start">
        <section className="min-w-0 overflow-hidden rounded-lg border border-border bg-surface shadow-card">
          <div className="border-b border-border p-3">
            <Tabs value={tab} onValueChange={(v) => router.replace(`${routes.flags}?tab=${v}`)}>
              <TabsList>
                <TabsTrigger value="open">{t.tabs.open}</TabsTrigger>
                <TabsTrigger value="resolved">{t.tabs.resolved}</TabsTrigger>
              </TabsList>
            </Tabs>
          </div>
          {list.isPending ? <Loading className="p-5" /> : list.isError ? <p role="alert" className="p-5 text-sm text-danger-ink">{problemOf(list.error)}</p>
            : rows.length === 0 ? <p className="p-5 text-sm text-ink-muted">{tab === "open" ? t.empty : t.emptyResolved}</p> : (
              <ul className="divide-y divide-border">
                {rows.map((f) => (
                  <li key={f.id}>
                    <button type="button" onClick={() => go({ flag: f.id })} aria-current={f.id === selected ? "true" : undefined}
                      className={cn("grid min-h-tap-scan w-full grid-cols-[32px_1fr] gap-3 px-4 py-3 text-left",
                        f.id === selected ? "bg-primary-soft shadow-[inset_3px_0_var(--primary)]" : "hover:bg-surface-sunken")}>
                      <span className={cn("grid size-8 place-items-center rounded-md",
                        f.status === flagStatus.open ? "bg-warning-soft text-warning-ink" : "bg-surface-sunken text-ink-muted")}>
                        <FlagIcon className="size-4" aria-hidden />
                      </span>
                      <span className="grid min-w-0 gap-0.5">
                        <span className="line-clamp-2 text-sm font-semibold">{f.reason}</span>
                        <span className="text-small text-ink-muted">
                          {t.source[f.raised_by ?? "user"] ?? f.raised_by} · {formatDate(f.created_at)}
                        </span>
                      </span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
        </section>
        {current ? <Investigation key={current.id} f={current} /> : (
          <section className="hidden rounded-lg border border-dashed border-border p-8 text-center text-sm text-ink-muted lg:block">{t.selectOne}</section>
        )}
      </div>
    </main>
  );
}

function Investigation({ f }: { f: Flag }) {
  const me = useMe().data!;
  const client = useQueryClient();
  const [deciding, setDeciding] = useState<Outcome>();
  const lot = useQuery({ queryKey: queryKeys.verify(f.passport_id), queryFn: () => endpoints.verify(f.passport_id) });
  // Descendants are readable to the platform; a regulator may not see another organisation's tree.
  const tree = useQuery({ queryKey: [...queryKeys.flag(f.id), "descendants"], queryFn: () => endpoints.genealogy(f.passport_id, "descendants"), retry: false });
  const affected = tree.data ? tree.data.nodes.length : undefined;
  const refresh = () => client.invalidateQueries({ queryKey: queryKeys.flags });

  const assign = useMutation({
    mutationFn: (userId: string | null) => endpoints.assignFlag(f.id, userId),
    onSuccess: () => { refresh(); toast.success(t.assigned); },
    onError: (err) => toast.error(problemOf(err)),
  });
  const { register, handleSubmit, setValue, formState: { errors } } = useForm<ResolveFlagForm>({
    resolver: zodResolver(resolveFlagFormSchema), defaultValues: { outcome: flagStatus.cleared, note: "" },
  });
  const resolve = useMutation({
    mutationFn: ({ v, pin }: { v: ResolveFlagForm; pin: string }) => endpoints.resolveFlag(f.id, v, pin),
    onSuccess: (_, { v }) => {
      refresh();
      toast.success(v.outcome === flagStatus.confirmed ? t.confirmed(affected ?? 1) : t.cleared);
      setDeciding(undefined);
    },
  });
  const code = lot.data?.passport.code ?? f.passport_id;
  // Mirrors the API: the platform (flag:resolve) or a verified regulator investigates; others only read.
  const investigator = me.permissions.includes(permissions.flagResolve)
    || (me.organisation?.kind === "regulator" && me.organisation.status === orgStatus.verified);

  return (
    <section className="grid min-w-0 gap-5 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div className="grid gap-1">
          <FlagChip status={f.status} />
          <h2 className="text-title">{f.reason}</h2>
          <span className="text-small text-ink-muted">{t.raised(formatDate(f.created_at))}</span>
        </div>
      </header>
      <KeyValue facts={[
        { label: t.passport, value: <Link className="text-primary" href={routes.passport(code)}>{code}</Link>, mono: true },
        { label: t.raisedBy, value: t.source[f.raised_by ?? "user"] ?? f.raised_by },
        { label: t.assignee, value: f.assignee_name ?? t.unassigned },
        { label: t.affected, value: affected !== undefined ? String(affected) : "—" },
        ...(f.resolution_note ? [{ label: t.resolution, value: f.resolution_note, wide: true }] : []),
      ]} />

      {f.status === flagStatus.open && !investigator && <Notice kind="info">{t.resolveNotYours}</Notice>}
      {f.status === flagStatus.open && investigator && (
        <>
          <div className="flex flex-wrap gap-2">
            {f.assignee_user_id === me.user.id
              ? <Button variant="secondary" onClick={() => assign.mutate(null)} disabled={assign.isPending}>{t.unassign}</Button>
              : <Button variant="secondary" onClick={() => assign.mutate(me.user.id)} disabled={assign.isPending}>{t.assignMe}</Button>}
          </div>
          <p className="text-sm text-ink-muted">{t.resolveIntro}</p>
          {!deciding ? (
            <div className="flex flex-wrap justify-end gap-2 border-t border-border pt-4">
              <Button variant="secondary" onClick={() => { setDeciding(flagStatus.cleared); setValue("outcome", flagStatus.cleared); }}>{t.clear}</Button>
              <Button variant="danger-soft" onClick={() => { setDeciding(flagStatus.confirmed); setValue("outcome", flagStatus.confirmed); }}>
                <ShieldAlert aria-hidden />{t.confirmFraud}
              </Button>
            </div>
          ) : (
            <div className={cn("grid gap-4 rounded-lg border p-4", deciding === flagStatus.confirmed ? "border-danger bg-danger-soft" : "border-border")}>
              <h3 className="text-title">{deciding === flagStatus.confirmed ? t.fraudTitle : t.clearTitle}</h3>
              {deciding === flagStatus.confirmed ? (
                <>
                  <p className="text-sm">{affected !== undefined ? t.fraudBody(affected) : t.fraudBodyNone}</p>
                  <Notice kind="blocking">{t.fraudWarning}</Notice>
                </>
              ) : <p className="text-sm text-ink-muted">{t.clearBody}</p>}
              <label className="grid gap-1.5">
                <span className="text-sm font-semibold">{deciding === flagStatus.confirmed ? t.fraudReason : t.clearNote}</span>
                <Textarea {...register("note")} aria-invalid={errors.note ? true : undefined} />
                <span className={cn("text-small", errors.note ? "font-semibold text-danger" : "text-ink-muted")}>
                  {errors.note ? (deciding === flagStatus.confirmed ? t.fraudReasonEmpty : t.clearNoteEmpty)
                    : deciding === flagStatus.confirmed ? t.fraudReasonHelp : t.clearNoteHelp}
                </span>
              </label>
              <SigningPanel
                button={deciding === flagStatus.confirmed ? t.fraudConfirm(affected ?? 1) : t.clear}
                busy={resolve.isPending}
                error={resolve.isError ? problemOf(resolve.error) : undefined}
                onSign={(pin) => handleSubmit((v) => resolve.mutate({ v, pin }))()}
              />
              <Button variant="secondary" className="w-fit" onClick={() => setDeciding(undefined)}>{copy.handover.back}</Button>
            </div>
          )}
        </>
      )}
    </section>
  );
}
