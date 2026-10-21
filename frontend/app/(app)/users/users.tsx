"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Ban,
  CircleCheck,
  Eye,
  Lock,
  OctagonX,
  Power,
  PowerOff,
  RotateCcw,
  ShieldCheck,
  ShieldOff,
  UserPlus,
} from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";

import { useConfirm } from "@/components/chaintrack/confirm";
import {
  ListTable,
  RowAction,
  actionsColumn,
  column,
  type ListColumn,
} from "@/components/chaintrack/list-table";
import { ListBar } from "@/components/chaintrack/list-controls";
import { Notice } from "@/components/chaintrack/notice";
import { PageHeader } from "@/components/chaintrack/page-header";
import { PageState } from "@/components/chaintrack/page-state";
import { StatusChip } from "@/components/chaintrack/status";
import { TextField } from "@/components/chaintrack/text-field";
import { Button } from "@/components/ui/button";
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
  type MemberRow,
  type PlatformUserRow,
} from "@/lib/api";
import {
  copy,
  memberRoles,
  memberStatus,
  permissions,
  queryKeys,
  confirmCopy,
  routes,
} from "@/lib/consts";
import { liveValidation } from "@/lib/form";
import { useListParams, type ListParamsState } from "@/lib/list-params";
import { inviteSchema, type InviteForm } from "@/lib/schemas/profile";
import { formatDate, humanize } from "@/lib/verification";

const t = copy.users;
const roleItems = memberRoles.map((r) => ({ value: r.value, label: r.label }));
export const roleLabel = (role: string) =>
  memberRoles.find((r) => r.value === role)?.label ?? humanize(role);

export function Users() {
  const me = useMe().data!;
  const list = useListParams("name");
  if (!me.permissions.includes(permissions.userManage)) {
    return (
      <PageState icon={Lock} title={t.notAllowedTitle}>
        {t.notAllowedBody}
      </PageState>
    );
  }
  const platform = !me.organisation;
  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <PageHeader title={t.title} sub={platform ? t.subPlatform : t.subOrg} />
      {!platform && <Invite />}
      {platform ? (
        <AllUsers list={list} myId={me.user.id} />
      ) : (
        <Members list={list} myId={me.user.id} />
      )}
    </main>
  );
}

export function StatusOf({ status }: { status: string }) {
  return status === memberStatus.active ? (
    <StatusChip tone="verified" icon={CircleCheck}>
      {t.status.active}
    </StatusChip>
  ) : (
    <StatusChip tone="danger" icon={OctagonX}>
      {t.status[status] ?? humanize(status)}
    </StatusChip>
  );
}

export function Person({
  name,
  email,
  you,
}: {
  name: string;
  email: string;
  you?: boolean;
}) {
  return (
    <span className="grid min-w-0">
      <span className="truncate font-semibold">
        {name}
        {you && <span className="ml-2 text-xs text-ink-muted">({t.you})</span>}
      </span>
      <span className="truncate text-xs text-ink-muted">{email}</span>
    </span>
  );
}

const accountColumns: ListColumn<PlatformUserRow>[] = [
  column(
    "name",
    t.columns.person,
    (u) => (
      <Link href={routes.userDetail(u.id)} className="hover:underline">
        <Person name={u.name} email={u.email} />
      </Link>
    ),
    { sortable: true },
  ),
  column("memberships", t.memberships, (u) =>
    u.memberships.length === 0
      ? t.noMemberships
      : u.memberships
          .map(
            (m) =>
              `${m.organisation ?? copy.shell.platform} · ${roleLabel(m.role)}`,
          )
          .join(", "),
  ),
  column("twoFactor", t.columns.twoFactor, (u) =>
    u.two_factor ? (
      <StatusChip tone="verified" icon={ShieldCheck}>
        {t.twoFactorOn}
      </StatusChip>
    ) : (
      <StatusChip tone="neutral" icon={ShieldOff}>
        {t.twoFactorOff}
      </StatusChip>
    ),
  ),
  column("status", t.columns.status, (u) => <StatusOf status={u.status} />, {
    sortable: true,
  }),
];

function AllUsers({ list, myId }: { list: ListParamsState; myId: string }) {
  const q = usePaged(queryKeys.users, endpoints.allUsers, list.query());
  const setActive = useSetActive();
  const columns = [
    ...accountColumns,
    actionsColumn((u: PlatformUserRow) => (
      <>
        <RowAction label={t.view} icon={Eye} href={routes.userDetail(u.id)} />
        <ActiveAction user={u} disabled={u.id === myId} {...setActive} />
      </>
    )),
  ];
  return (
    <ListTable
      label={t.title}
      columns={columns}
      rows={q.data?.data ?? []}
      total={q.data?.total ?? 0}
      list={list}
      rowKey={(u) => u.id}
      loading={q.isPending}
      error={q.isError ? problemOf(q.error) : undefined}
      empty={list.search ? copy.list.noMatch : t.empty}
      toolbar={<ListBar list={list} />}
    />
  );
}

/** Turning a whole account off or on (platform admin), asked first. */
export function useSetActive() {
  const client = useQueryClient();
  const confirm = useConfirm();
  const act = useMutation({
    mutationFn: ({ id, active }: { id: string; active: boolean }) =>
      active ? endpoints.activate(id) : endpoints.deactivate(id),
    onSuccess: () => client.invalidateQueries({ queryKey: queryKeys.users }),
  });
  const ask = async (u: PlatformUserRow, active: boolean) => {
    const c = active
      ? confirmCopy.userActivate(u.name)
      : confirmCopy.userDeactivate(u.name);
    if (
      await confirm({
        title: c.title,
        body: c.body,
        confirmLabel: c.label,
        danger: !active,
      })
    )
      act.mutate({ id: u.id, active });
  };
  return { ask, pending: act.isPending };
}

export function ActiveAction({
  user,
  disabled,
  ask,
  pending,
}: {
  user: PlatformUserRow;
  disabled?: boolean;
  ask: (u: PlatformUserRow, active: boolean) => void;
  pending: boolean;
}) {
  return user.status === memberStatus.active ? (
    <RowAction
      label={t.deactivate}
      icon={PowerOff}
      danger
      disabled={pending || disabled}
      onClick={() => ask(user, false)}
    />
  ) : (
    <RowAction
      label={t.activate}
      icon={Power}
      disabled={pending}
      onClick={() => ask(user, true)}
    />
  );
}

type MemberAction = {
  id: string;
  action: "role" | "suspend" | "reinstate";
  role?: string;
};

function Members({ list, myId }: { list: ListParamsState; myId: string }) {
  const client = useQueryClient();
  const q = usePaged(queryKeys.users, endpoints.members, list.query());
  const act = useMutation({
    mutationFn: ({ id, action, role }: MemberAction) =>
      action === "role"
        ? endpoints.setRole(id, role!)
        : action === "suspend"
          ? endpoints.suspend(id)
          : endpoints.reinstate(id),
    onSuccess: () => client.invalidateQueries({ queryKey: queryKeys.users }), // refusals (last admin, yourself): the API says why
  });
  const confirm = useConfirm();
  const ask = async (
    c: { title: string; body: string; label: string },
    a: MemberAction,
    danger = false,
  ) => {
    if (
      await confirm({
        title: c.title,
        body: c.body,
        confirmLabel: c.label,
        danger,
      })
    )
      act.mutate(a);
  };
  const columns: ListColumn<MemberRow>[] = [
    column(
      "name",
      t.columns.person,
      (u) => <Person name={u.name} email={u.email} you={u.id === myId} />,
      { sortable: true },
    ),
    column(
      "role",
      t.columns.role,
      (u) => (
        <Select
          items={roleItems}
          value={u.role}
          disabled={
            act.isPending || u.status !== memberStatus.active || u.id === myId
          }
          onValueChange={(role) =>
            role &&
            role !== u.role &&
            ask(confirmCopy.userRole(u.name, roleLabel(role as string)), {
              id: u.id,
              action: "role",
              role: role as string,
            })
          }
        >
          <SelectTrigger aria-label={t.changeRole} className="h-9 w-40">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {roleItems.map((r) => (
              <SelectItem key={r.value} value={r.value}>
                {r.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      ),
      { sortable: true },
    ),
    column("status", t.columns.status, (u) => <StatusOf status={u.status} />, {
      sortable: true,
    }),
    column("created_at", t.columns.joined, (u) => formatDate(u.since), {
      sortable: true,
    }),
    actionsColumn((u) =>
      u.status === memberStatus.active ? (
        <RowAction
          label={t.suspend}
          icon={Ban}
          danger
          disabled={act.isPending || u.id === myId}
          onClick={() =>
            ask(
              confirmCopy.userSuspend(u.name),
              { id: u.id, action: "suspend" },
              true,
            )
          }
        />
      ) : (
        <RowAction
          label={t.reinstate}
          icon={RotateCcw}
          disabled={act.isPending}
          onClick={() =>
            ask(confirmCopy.userReinstate(u.name), {
              id: u.id,
              action: "reinstate",
            })
          }
        />
      ),
    ),
  ];
  return (
    <ListTable
      label={t.title}
      columns={columns}
      rows={q.data?.data ?? []}
      total={q.data?.total ?? 0}
      list={list}
      rowKey={(u) => u.id}
      loading={q.isPending}
      error={q.isError ? problemOf(q.error) : undefined}
      empty={list.search ? copy.list.noMatch : t.empty}
      toolbar={<ListBar list={list} />}
    />
  );
}

function Invite() {
  const client = useQueryClient();
  const [link, setLink] = useState<string>();
  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<InviteForm>({
    ...liveValidation,
    resolver: zodResolver(inviteSchema),
    defaultValues: { email: "", role: "member" },
  });
  const send = useMutation({
    mutationFn: endpoints.invite,
    onSuccess: ({ data }) => {
      setLink(data.emailed ? undefined : data.link); // the API returns the link only in development
      reset();
      client.invalidateQueries({ queryKey: queryKeys.users });
    },
  });
  return (
    <section className="grid gap-3 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
      <h2 className="flex items-center gap-2 text-title">
        <UserPlus className="size-5 text-ink-muted" aria-hidden />
        {t.invite}
      </h2>
      <form
        onSubmit={handleSubmit((v) => send.mutate(v))}
        noValidate
        className="grid gap-3 md:grid-cols-[1fr_12rem_auto] md:items-end"
      >
        <TextField
          label={t.inviteEmail}
          type="email"
          inputMode="email"
          autoComplete="off"
          error={errors.email?.message}
          {...register("email")}
        />
        <div className="grid gap-1.5">
          <span className="text-sm font-semibold">{t.inviteRole}</span>
          <Controller
            control={control}
            name="role"
            render={({ field }) => (
              <Select
                items={roleItems}
                value={field.value}
                onValueChange={(v) => field.onChange(v)}
              >
                <SelectTrigger
                  aria-label={t.inviteRole}
                  className="min-h-tap w-full"
                >
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {roleItems.map((r) => (
                    <SelectItem key={r.value} value={r.value}>
                      {r.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          />
        </div>
        <Button type="submit" loading={send.isPending}>
          {t.sendInvite}
        </Button>
      </form>
      {link && (
        <Notice kind="info">
          {t.notEmailed} <span className="font-mono break-all">{link}</span>
        </Notice>
      )}
    </section>
  );
}
