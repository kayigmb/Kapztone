"use client";

import { useQuery } from "@tanstack/react-query";
import { parseAsStringLiteral } from "nuqs";

import { ListCard } from "@/components/chaintrack/list-card";
import { ListBar, ListPager } from "@/components/chaintrack/list-controls";
import {
  NotificationItem,
  useOpenNotification,
} from "@/components/chaintrack/notification-item";
import { PageHeader } from "@/components/chaintrack/page-header";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { endpoints, problemOf, usePaged } from "@/lib/api";
import { copy, listSorts, notificationTabs, queryKeys } from "@/lib/consts";
import { useListParams } from "@/lib/list-params";

const t = copy.notifications;
const listFilters = {
  tab: parseAsStringLiteral(notificationTabs).withDefault(notificationTabs[0]),
};

export function Notifications() {
  const state = useListParams("-created_at", listFilters);
  const tab = state.filters.tab;
  const unreadOnly = tab === notificationTabs[1];
  const list = usePaged(
    queryKeys.notifications,
    endpoints.notificationsPage,
    state.query({ unread: unreadOnly ? "true" : undefined }),
  );
  const unread = useQuery({
    queryKey: queryKeys.unread,
    queryFn: endpoints.unreadCount,
  });
  const item = useOpenNotification();
  const rows = list.data?.data ?? [];

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <PageHeader title={t.title} sub={t.sub} />

      <Tabs
        value={tab}
        onValueChange={(v) => state.setFilters({ tab: v as typeof tab })}
      >
        <TabsList>
          {notificationTabs.map((x) => (
            <TabsTrigger key={x} value={x}>
              {t.tabs[x]}
              {x === notificationTabs[1] && !!unread.data && (
                <span className="ml-1.5 rounded-full bg-surface-sunken px-1.5 font-mono text-xs">
                  {unread.data}
                </span>
              )}
            </TabsTrigger>
          ))}
        </TabsList>
      </Tabs>

      <ListBar list={state} sorts={listSorts.date} />

      <ListCard
        title={t.tabs[tab]}
        empty={
          state.search
            ? copy.list.noMatch
            : unreadOnly
              ? t.emptyUnread
              : t.empty
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
        {rows.map((n) => (
          <NotificationItem
            key={n.id}
            n={n}
            onOpen={item.open}
            onMarkRead={item.markRead}
            busy={item.busy}
          />
        ))}
      </ListCard>
      <ListPager list={state} total={list.data?.total ?? 0} />
    </main>
  );
}
