"use client";

import { useQuery } from "@tanstack/react-query";
import { CircleCheck, Layers, Package, Plus, XCircle } from "lucide-react";
import Link from "next/link";

import { ListBar } from "@/components/chaintrack/list-controls";
import {
  column,
  ListTable,
  type ListColumn,
} from "@/components/chaintrack/list-table";
import { PageHeader } from "@/components/chaintrack/page-header";
import { PageStats } from "@/components/chaintrack/page-stats";
import { StatusChip } from "@/components/chaintrack/status";
import { buttonVariants } from "@/components/ui/button";
import {
  endpoints,
  problemOf,
  useMe,
  usePaged,
  type PassportRow,
} from "@/lib/api";
import {
  copy,
  passportStatus,
  permissions,
  queryKeys,
  routes,
} from "@/lib/consts";
import { useListParams } from "@/lib/list-params";
import { formatDate, humanize } from "@/lib/verification";
import { statusLook } from "./[id]/passport-detail";

const t = copy.passportsList;
const s = copy.stats.passports;

/** Every passport this person may see: their organisation's lots, or for the platform all of them. */
export function Passports() {
  const me = useMe().data!; // the layout only renders once it is loaded
  const platform = !me.organisation;
  const list = useListParams("-created_at");
  const q = usePaged(
    queryKeys.passports,
    endpoints.passportsPage,
    list.query(),
  );

  const columns: ListColumn<PassportRow>[] = [
    column("code", t.columns.code, (p) => (
      <Link
        href={routes.passportDetail(p.id)}
        className="font-mono font-semibold hover:underline"
      >
        {p.code}
      </Link>
    )),
    column("commodity", t.columns.commodity, (p) => humanize(p.commodity), {
      sortable: true,
    }),
    column("remaining", t.columns.remaining, (p) => (
      <span className="fig">
        {p.remaining.toLocaleString()} {p.unit}
      </span>
    )),
    ...(platform
      ? [
          column("holder", t.columns.holder, (p: PassportRow) => (
            <Holder id={p.organisation_id} />
          )),
        ]
      : []),
    column(
      "status",
      t.columns.status,
      (p) => {
        const look = statusLook[p.status] ?? statusLook.active;
        return (
          <StatusChip tone={look.tone} icon={look.icon}>
            {look.label}
          </StatusChip>
        );
      },
      { sortable: true },
    ),
    column("created_at", t.columns.issued, (p) => formatDate(p.created_at), {
      sortable: true,
    }),
  ];

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <PageHeader
        title={t.title}
        sub={platform ? t.subPlatform : t.sub}
        actions={
          me.organisation &&
          me.permissions.includes(permissions.passportIssue) && (
            <Link href={routes.issue} className={buttonVariants()}>
              <Plus aria-hidden />
              {t.issue}
            </Link>
          )
        }
      />
      <PageStats
        queryKey={queryKeys.passports}
        fetch={endpoints.passportsPage}
        cards={[
          { key: "total", ...s.total, icon: Package },
          {
            key: "active",
            ...s.active,
            icon: CircleCheck,
            filters: { status: passportStatus.active },
          },
          {
            key: "consumed",
            ...s.consumed,
            icon: Layers,
            filters: { status: passportStatus.consumed },
          },
          {
            key: "revoked",
            ...s.revoked,
            icon: XCircle,
            filters: { status: passportStatus.revoked },
          },
        ]}
      />
      <ListTable
        label={t.title}
        columns={columns}
        rows={q.data?.data ?? []}
        total={q.data?.total ?? 0}
        list={list}
        rowKey={(p) => p.id}
        loading={q.isPending}
        error={q.isError ? problemOf(q.error) : undefined}
        empty={list.search ? copy.list.noMatch : t.empty}
        toolbar={<ListBar list={list} />}
      />
    </main>
  );
}

/** An organisation's name; reads the cached directory itself, so the table's columns never go stale. */
function Holder({ id }: { id: string }) {
  const names = useQuery({
    queryKey: queryKeys.directory,
    queryFn: endpoints.directory,
  });
  return names.data?.get(id) ?? (names.isPending ? copy.loading : "—");
}
