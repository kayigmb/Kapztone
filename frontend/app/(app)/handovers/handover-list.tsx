"use client";

import { useQueries, useQuery } from "@tanstack/react-query";
import {
  ArrowDownLeft,
  ArrowLeftRight,
  ArrowUpRight,
  ChevronRight,
  CircleCheck,
  Plus,
} from "lucide-react";
import Link from "next/link";
import { parseAsStringLiteral } from "nuqs";

import { ListCard } from "@/components/chaintrack/list-card";
import { ListBar, ListPager } from "@/components/chaintrack/list-controls";
import { PageHeader } from "@/components/chaintrack/page-header";
import { PageStats } from "@/components/chaintrack/page-stats";
import { HandoverStatusChip } from "@/components/chaintrack/handover-status";
import { buttonVariants } from "@/components/ui/button";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { endpoints, problemOf, useMe, usePaged } from "@/lib/api";
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

      <ListBar list={state} sorts={listSorts.date} />

      <ListCard
        title={t.tabs[tab]}
        empty={
          {
            incoming: t.emptyIncoming,
            outgoing: t.emptyOutgoing,
            done: t.emptyDone,
          }[tab]
        }
        state={
          list.isPending
            ? "loading"
            : list.isError
              ? "failed"
              : rows.length
                ? "ready"
                : "empty"
        }
        error={list.isError ? problemOf(list.error) : undefined}
      >
        {rows.map((h) => (
          <li key={h.id}>
            <Link
              href={routes.handoverDetail(h.id)}
              className="grid grid-cols-[auto_1fr_auto_auto] items-center gap-3 px-5 py-3 hover:bg-surface-sunken"
            >
              <span className="grid size-9 place-items-center rounded-md bg-primary-soft text-primary">
                <ArrowLeftRight className="size-4" aria-hidden />
              </span>
              <span className="grid min-w-0 gap-0.5">
                <span className="truncate text-sm font-semibold">
                  {h.to_org === mine
                    ? t.from(nameOf(h.from_org))
                    : t.to(nameOf(h.to_org))}
                </span>
                <span className="text-small text-ink-muted">
                  {humanize(h.stage)} ·{" "}
                  <span className="font-mono">
                    {h.quantity.toLocaleString()}
                  </span>{" "}
                  · {formatDate(h.created_at)}
                </span>
              </span>
              <HandoverStatusChip h={h} mine={mine} />
              <ChevronRight className="size-4 text-ink-muted" aria-hidden />
            </Link>
          </li>
        ))}
      </ListCard>
      <ListPager list={state} total={list.data?.total ?? 0} />
    </main>
  );
}
