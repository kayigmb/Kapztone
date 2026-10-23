"use client";

import { useQueries } from "@tanstack/react-query";
import type { LucideIcon } from "lucide-react";

import { StatTile } from "@/components/chaintrack/stat-tile";
import { copy } from "@/lib/consts";
import type { ListParams } from "@/lib/list";
import { cn } from "@/lib/utils";

export type StatCard = {
  key: string;
  label: string;
  note: string;
  icon: LucideIcon;
  /** The tone the card takes once its count is above zero. */
  whenAny?: "attention" | "alert";
  /** The page's own list, narrowed by these filters; only its total is read. */
  filters?: ListParams["filters"];
};

const cols: Record<number, string> = {
  2: "lg:grid-cols-2",
  3: "lg:grid-cols-3",
  4: "lg:grid-cols-4",
};

/**
 * A screen's breakdown: one count per card, each from the screen's own list
 * endpoint (size 1, read the total), so the counts follow the same access
 * rules as the list. ponytail: one request per card; a /counts endpoint if
 * this ever gets slow.
 */
export function PageStats({
  queryKey,
  fetch,
  cards,
}: {
  queryKey: readonly unknown[];
  fetch: (p: ListParams) => Promise<{ total: number }>;
  cards: readonly StatCard[];
}) {
  const counts = useQueries({
    queries: cards.map((c) => ({
      queryKey: [...queryKey, "stats", c.key, c.filters],
      queryFn: () => fetch({ page: 1, size: 1, filters: c.filters }),
    })),
  });
  return (
    <section
      aria-label={copy.stats.label}
      className={cn("grid grid-cols-2 gap-2 sm:gap-3", cols[cards.length])}
    >
      {cards.map((c, i) => {
        const total = counts[i].data?.total;
        return (
          <StatTile
            key={c.key}
            label={c.label}
            note={c.note}
            icon={c.icon}
            value={total}
            tone={total && c.whenAny ? c.whenAny : "neutral"}
            failed={counts[i].isError ? copy.stats.failed : undefined}
          />
        );
      })}
    </section>
  );
}
