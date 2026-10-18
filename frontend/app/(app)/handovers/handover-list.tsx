"use client";

import { useQuery } from "@tanstack/react-query";
import { ArrowLeftRight, ChevronRight, Plus } from "lucide-react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";

import { ListCard } from "@/components/chaintrack/list-card";
import { HandoverStatusChip } from "@/components/chaintrack/handover-status";
import { buttonVariants } from "@/components/ui/button";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { endpoints, problemOf, useMe, type HandoverDetail } from "@/lib/api";
import { copy, handoverStatus, permissions, queryKeys, routes } from "@/lib/consts";
import { formatDate, humanize } from "@/lib/verification";

const t = copy.receiver;
const tabs = ["incoming", "outgoing", "done"] as const;
type Tab = (typeof tabs)[number];

/** Which tab a handover belongs on for the acting organisation. */
const tabOf = (h: HandoverDetail, mine: string | undefined): Tab =>
  h.status !== handoverStatus.awaiting ? "done" : h.to_org === mine ? "incoming" : "outgoing";

export function HandoverList() {
  const params = useSearchParams();
  const router = useRouter();
  const me = useMe().data!;
  const tab: Tab = tabs.find((x) => x === params.get("tab")) ?? "incoming";
  const list = useQuery({ queryKey: queryKeys.handovers, queryFn: () => endpoints.handoversList() });
  const names = useQuery({ queryKey: queryKeys.directory, queryFn: endpoints.directory });
  const mine = me.organisation?.id;
  const rows = list.data?.filter((h) => tabOf(h, mine) === tab) ?? [];
  const count = (x: Tab) => list.data?.filter((h) => tabOf(h, mine) === x).length;
  const nameOf = (id: string) => names.data?.get(id) ?? copy.dashboard.someOrg;

  return (
    <main className="grid content-start gap-6 p-4 lg:p-8">
      <header className="flex flex-wrap items-end justify-between gap-3">
        <div className="grid gap-1">
          <h1 className="text-[22px] font-semibold lg:text-page">{t.title}</h1>
          <p className="text-sm text-ink-muted">{t.sub}</p>
        </div>
        {me.permissions.includes(permissions.handoverCreate) && (
          <Link href={routes.handover()} className={buttonVariants()}><Plus aria-hidden />{copy.handover.title}</Link>
        )}
      </header>

      <Tabs value={tab} onValueChange={(v) => router.replace(`${routes.handovers}?tab=${v}`)}>
        <TabsList>
          {tabs.map((x) => (
            <TabsTrigger key={x} value={x}>
              {t.tabs[x]}{count(x) ? <span className="ml-1.5 rounded-full bg-surface-sunken px-1.5 font-mono text-xs">{count(x)}</span> : null}
            </TabsTrigger>
          ))}
        </TabsList>
      </Tabs>

      <ListCard title={t.tabs[tab]} empty={{ incoming: t.emptyIncoming, outgoing: t.emptyOutgoing, done: t.emptyDone }[tab]}
        state={list.isPending ? "loading" : list.isError ? "failed" : rows.length ? "ready" : "empty"}
        error={list.isError ? problemOf(list.error) : undefined}>
        {rows.map((h) => (
          <li key={h.id}>
            <Link href={routes.handoverDetail(h.id)}
              className="grid min-h-tap-scan grid-cols-[36px_1fr_auto_16px] items-center gap-3 px-5 py-3 hover:bg-surface-sunken">
              <span className="grid size-9 place-items-center rounded-md bg-primary-soft text-primary"><ArrowLeftRight className="size-4" aria-hidden /></span>
              <span className="grid min-w-0 gap-0.5">
                <span className="truncate text-sm font-semibold">{h.to_org === mine ? t.from(nameOf(h.from_org)) : t.to(nameOf(h.to_org))}</span>
                <span className="text-small text-ink-muted">
                  {humanize(h.stage)} · <span className="font-mono">{h.quantity.toLocaleString()}</span> · {formatDate(h.created_at)}
                </span>
              </span>
              <HandoverStatusChip h={h} mine={mine} />
              <ChevronRight className="size-4 text-ink-muted" aria-hidden />
            </Link>
          </li>
        ))}
      </ListCard>
    </main>
  );
}
