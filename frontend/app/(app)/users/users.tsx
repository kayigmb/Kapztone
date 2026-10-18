"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CircleCheck, Lock, OctagonX, ShieldCheck, ShieldOff, UserPlus } from "lucide-react";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";

import { Notice } from "@/components/chaintrack/notice";
import { Loading, PageState } from "@/components/chaintrack/page-state";
import { StatusChip } from "@/components/chaintrack/status";
import { TextField } from "@/components/chaintrack/text-field";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { endpoints, problemOf, useMe, type MemberRow, type PlatformUserRow } from "@/lib/api";
import { copy, memberRoles, memberStatus, permissions, queryKeys } from "@/lib/consts";
import { inviteSchema, type InviteForm } from "@/lib/schemas/profile";
import { formatDate, humanize } from "@/lib/verification";

const t = copy.users;
const roleItems = memberRoles.map((r) => ({ value: r.value, label: r.label }));
const roleLabel = (role: string) => memberRoles.find((r) => r.value === role)?.label ?? humanize(role);
const matches = (q: string, u: { name: string; email: string }) =>
  !q || `${u.name} ${u.email}`.toLowerCase().includes(q.trim().toLowerCase());

export function Users() {
  const me = useMe().data!;
  const [search, setSearch] = useState("");
  if (!me.permissions.includes(permissions.userManage)) {
    return <PageState icon={Lock} title={t.notAllowedTitle}>{t.notAllowedBody}</PageState>;
  }
  const platform = !me.organisation;
  return (
    <main className="grid content-start gap-6 p-4 lg:p-8">
      <header className="grid gap-1">
        <h1 className="text-[22px] font-semibold lg:text-page">{t.title}</h1>
        <p className="text-sm text-ink-muted">{platform ? t.subPlatform : t.subOrg}</p>
      </header>
      {!platform && <Invite />}
      <Input value={search} onChange={(e) => setSearch(e.target.value)} placeholder={t.search} aria-label={t.search} className="max-w-md" />
      {platform ? <AllUsers search={search} /> : <Members search={search} myId={me.user.id} />}
    </main>
  );
}

function StatusOf({ status }: { status: string }) {
  return status === memberStatus.active
    ? <StatusChip tone="verified" icon={CircleCheck}>{t.status.active}</StatusChip>
    : <StatusChip tone="danger" icon={OctagonX}>{t.status[status] ?? humanize(status)}</StatusChip>;
}

function List<T>({ query, rows, render }: {
  query: { isPending: boolean; isError: boolean; error: unknown }; rows: T[]; render: (row: T) => React.ReactNode;
}) {
  if (query.isPending) return <Loading />;
  if (query.isError) return <Notice kind="blocking">{problemOf(query.error)}</Notice>;
  if (rows.length === 0) return <p className="text-sm text-ink-muted">{t.empty}</p>;
  return <ul className="grid divide-y divide-border rounded-lg border border-border bg-surface shadow-card">{rows.map(render)}</ul>;
}

function AllUsers({ search }: { search: string }) {
  const q = useQuery({ queryKey: queryKeys.users, queryFn: endpoints.allUsers });
  const rows = (q.data ?? []).filter((u) => matches(search, u));
  return <List query={q} rows={rows} render={(u: PlatformUserRow) => (
    <li key={u.id} className="grid gap-2 p-4 md:grid-cols-[1fr_auto] md:items-start">
      <div className="grid min-w-0 gap-0.5">
        <span className="font-semibold">{u.name}</span>
        <span className="truncate text-sm text-ink-muted">{u.email}</span>
        <span className="text-sm">
          {u.memberships.length === 0 ? t.noMemberships
            : u.memberships.map((m) => `${m.organisation ?? copy.shell.platform} · ${roleLabel(m.role)}`).join(", ")}
        </span>
      </div>
      <div className="flex flex-wrap items-center gap-2">
        {u.two_factor !== undefined && (u.two_factor
          ? <StatusChip tone="verified" icon={ShieldCheck}>{t.twoFactorOn}</StatusChip>
          : <StatusChip tone="neutral" icon={ShieldOff}>{t.twoFactorOff}</StatusChip>)}
        <StatusOf status={u.status} />
      </div>
    </li>
  )} />;
}

function Members({ search, myId }: { search: string; myId: string }) {
  const client = useQueryClient();
  const q = useQuery({ queryKey: queryKeys.users, queryFn: endpoints.members });
  const refresh = () => client.invalidateQueries({ queryKey: queryKeys.users });
  const act = useMutation({
    mutationFn: ({ id, action, role }: { id: string; action: "role" | "suspend" | "reinstate"; role?: string }) =>
      action === "role" ? endpoints.setRole(id, role!) : action === "suspend" ? endpoints.suspend(id) : endpoints.reinstate(id),
    onSuccess: (_r, v) => {
      toast.success(v.action === "role" ? t.roleChanged : v.action === "suspend" ? t.suspended : t.reinstated);
      refresh();
    },
    onError: (err) => toast.error(problemOf(err)), // last admin, yourself: the API says why
  });
  const rows = (q.data ?? []).filter((u) => matches(search, u));

  return <List query={q} rows={rows} render={(u: MemberRow) => (
    <li key={u.id} className="grid gap-3 p-4 md:grid-cols-[1fr_auto] md:items-center">
      <div className="grid min-w-0 gap-0.5">
        <span className="font-semibold">{u.name}{u.id === myId && <span className="ml-2 text-xs text-ink-muted">({t.you})</span>}</span>
        <span className="truncate text-sm text-ink-muted">{u.email}</span>
        <span className="text-xs text-ink-subtle">{t.since(formatDate(u.since))}</span>
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <StatusOf status={u.status} />
        <Select items={roleItems} value={u.role} disabled={act.isPending || u.status !== memberStatus.active}
          onValueChange={(role) => role && role !== u.role && act.mutate({ id: u.id, action: "role", role: role as string })}>
          <SelectTrigger aria-label={t.changeRole} className="min-h-tap w-48"><SelectValue /></SelectTrigger>
          <SelectContent>{roleItems.map((r) => <SelectItem key={r.value} value={r.value}>{r.label}</SelectItem>)}</SelectContent>
        </Select>
        {u.status === memberStatus.active ? (
          <Button variant="danger-outline" disabled={act.isPending} onClick={() => act.mutate({ id: u.id, action: "suspend" })}>{t.suspend}</Button>
        ) : (
          <Button variant="secondary" disabled={act.isPending} onClick={() => act.mutate({ id: u.id, action: "reinstate" })}>{t.reinstate}</Button>
        )}
      </div>
    </li>
  )} />;
}

function Invite() {
  const client = useQueryClient();
  const [link, setLink] = useState<string>();
  const { register, control, handleSubmit, reset, formState: { errors } } = useForm<InviteForm>({
    resolver: zodResolver(inviteSchema), defaultValues: { email: "", role: "member" },
  });
  const send = useMutation({
    mutationFn: endpoints.invite,
    onSuccess: (r) => {
      toast.success(t.invited(r.invite.email));
      setLink(r.emailed ? undefined : r.link); // the API returns the link only in development
      reset();
      client.invalidateQueries({ queryKey: queryKeys.users });
    },
    onError: (err) => toast.error(problemOf(err)),
  });
  return (
    <section className="grid gap-3 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
      <h2 className="flex items-center gap-2 text-title"><UserPlus className="size-5 text-ink-muted" aria-hidden />{t.invite}</h2>
      <form onSubmit={handleSubmit((v) => send.mutate(v))} noValidate className="grid gap-3 md:grid-cols-[1fr_12rem_auto] md:items-end">
        <TextField label={t.inviteEmail} type="email" inputMode="email" autoComplete="off" error={errors.email?.message} {...register("email")} />
        <div className="grid gap-1.5">
          <span className="text-sm font-semibold">{t.inviteRole}</span>
          <Controller control={control} name="role" render={({ field }) => (
            <Select items={roleItems} value={field.value} onValueChange={(v) => field.onChange(v)}>
              <SelectTrigger aria-label={t.inviteRole} className="min-h-tap w-full"><SelectValue /></SelectTrigger>
              <SelectContent>{roleItems.map((r) => <SelectItem key={r.value} value={r.value}>{r.label}</SelectItem>)}</SelectContent>
            </Select>
          )} />
        </div>
        <Button type="submit" disabled={send.isPending}>{t.sendInvite}</Button>
      </form>
      {link && (
        <Notice kind="info">
          {t.notEmailed} <span className="font-mono break-all">{link}</span>
        </Notice>
      )}
    </section>
  );
}
