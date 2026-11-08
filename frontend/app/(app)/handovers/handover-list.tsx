"use client";

import { useQueries, useQuery } from "@tanstack/react-query";
import { ArrowDownLeft, ArrowUpRight, CircleCheck, Plus } from "lucide-react";
import Link from "next/link";
import { parseAsStringLiteral } from "nuqs";

import { ListBar } from "@/components/chaintrack/list-controls";
import {
  column,
  ListTable,
  type ListColumn,
} from "@/components/chaintrack/list-table";
import { PageHeader } from "@/components/chaintrack/page-header";
import { PageStats } from "@/components/chaintrack/page-stats";
import { HandoverStatusChip } from "@/components/chaintrack/handover-status";
import { buttonVariants } from "@/components/ui/button";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  endpoints,
  problemOf,
  useMe,
  usePaged,
  type HandoverRow,
} from "@/lib/api";
import {
  copy,
  handoverBoxes,
  listSorts,
  permissions,
  queryKeys,
  routes,
} from "@/lib/consts";
import { useListParams } from "@/lib/list-params";
import { formatDate, humanize } from "@/lib/verification";

const t = copy.receiver;
const listFilters = {
  tab: parseAsStringLiteral(handoverBoxes).withDefault(handoverBoxes[0]),
};

export function HandoverList() {
  const me = useMe().data!;
  const state = useListParams("-created_at", listFilters);
  const tab = state.filters.tab;
  // The API sorts each handover into its box (?box=) for the acting organisation.
  const list = usePaged(
    queryKeys.handovers,
    endpoints.handoversPage,
    state.query({ box: tab }),
  );
  const counts = useQueries({
    queries: handoverBoxes.map((x) => ({
      queryKey: [...queryKeys.handovers, "count", x],
      queryFn: () =>
        endpoints
          .handoversPage({ page: 1, size: 1, filters: { box: x } })
          .then((p) => p.total),
    })),
  });
  const names = useQuery({
    queryKey: queryKeys.directory,
    queryFn: endpoints.directory,
  });
  const mine = me.organisation?.id;
  const rows = list.data?.data ?? [];
  const count = (x: (typeof handoverBoxes)[number]) =>
    counts[handoverBoxes.indexOf(x)].data;
  const nameOf = (id: string) => names.data?.get(id) ?? copy.dashboard.someOrg;
  const c = t.columns;
  // "Bob · Kitabi Washing": the person, then the organisation.
  const who = (person: string | null | undefined, org: string) =>
    person ? `${person} · ${org}` : org;
  const columns: ListColumn<HandoverRow>[] = [
    column("party", c.party, (h) => (
      <Link
        href={routes.handoverDetail(h.id)}
        className="font-semibold hover:underline"
      >
        {h.to_org === mine
          ? t.from(who(h.from_user_name, h.from_name ?? nameOf(h.from_org)))
          : h.from_org === mine
            ? t.to(who(h.to_user_name, h.to_name ?? nameOf(h.to_org)))
            : // The platform is neither party: both names.
              `${h.from_name} → ${who(h.to_user_name, h.to_name ?? "")}`}
      </Link>
    )),
    column("stage", c.stage, (h) => humanize(h.stage)),
    column("quantity", c.quantity, (h) => (
      <span className="fig">{h.quantity.toLocaleString()}</span>
    )),
    column("created_at", c.date, (h) => formatDate(h.created_at), {
      sortable: true,
    }),
    column("status", c.status, (h) => <HandoverStatusChip h={h} mine={mine} />),
  ];

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <PageHeader
        title={t.title}
        sub={t.sub}
        actions={
          me.permissions.includes(permissions.handoverCreate) && (
            <Link href={routes.handover()} className={buttonVariants()}>
              <Plus aria-hidden />
              {copy.handover.title}
            </Link>
          )
        }
      />
      <PageStats
        queryKey={queryKeys.handovers}
        fetch={endpoints.handoversPage}
        cards={[
          {
            key: "incoming",
            ...copy.stats.handovers.incoming,
            icon: ArrowDownLeft,
            whenAny: "attention",
            filters: { box: "incoming" },
          },
          {
            key: "outgoing",
            ...copy.stats.handovers.outgoing,
            icon: ArrowUpRight,
            filters: { box: "outgoing" },
          },
          {
            key: "done",
            ...copy.stats.handovers.done,
            icon: CircleCheck,
            filters: { box: "done" },
          },
        ]}
      />

      <Tabs
        value={tab}
        onValueChange={(v) => state.setFilters({ tab: v as typeof tab })}
      >
        <TabsList>
          {handoverBoxes.map((x) => (
            <TabsTrigger key={x} value={x}>
              {t.tabs[x]}
              {count(x) ? (
                <span className="ml-1.5 rounded-full bg-surface-sunken px-1.5 font-mono text-xs">
                  {count(x)}
                </span>
              ) : null}
            </TabsTrigger>
          ))}
        </TabsList>
      </Tabs>

      <ListTable
        label={t.tabs[tab]}
        columns={columns}
        rows={rows}
        total={list.data?.total ?? 0}
        list={state}
        rowKey={(h) => h.id}
        loading={list.isPending}
        error={list.isError ? problemOf(list.error) : undefined}
        empty={
          list.isPending
            ? copy.loading
            : {
                incoming: t.emptyIncoming,
                outgoing: t.emptyOutgoing,
                done: t.emptyDone,
              }[tab]
        }
        toolbar={<ListBar list={state} sorts={listSorts.date} />}
      />
    </main>
  );
}
