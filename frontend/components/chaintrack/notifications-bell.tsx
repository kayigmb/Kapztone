"use client";

import { useQuery } from "@tanstack/react-query";
import { Bell } from "lucide-react";
import Link from "next/link";
import { useState } from "react";

import {
  NotificationItem,
  useOpenNotification,
} from "@/components/chaintrack/notification-item";
import { Loading } from "@/components/chaintrack/page-state";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { endpoints, problemOf } from "@/lib/api";
import { copy, notificationRefetchMs, queryKeys, routes } from "@/lib/consts";

const t = copy.notifications;

/** Header bell: unread count, and the latest notifications; opening one marks it read and goes to it. */
export function NotificationsBell() {
  const [open, setOpen] = useState(false);
  const unread = useQuery({
    queryKey: queryKeys.unread,
    queryFn: endpoints.unreadCount,
    refetchInterval: notificationRefetchMs,
  });
  const list = useQuery({
    queryKey: queryKeys.notifications,
    queryFn: endpoints.notifications,
    enabled: open,
  });
  const item = useOpenNotification(() => setOpen(false));
  const count = unread.data ?? 0;

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        aria-label={t.open(count)}
        className="relative grid size-9 place-items-center rounded-md border border-border text-ink-muted hover:bg-surface-sunken hover:text-ink"
      >
        <Bell className="size-4.5" aria-hidden />
        {count > 0 && (
          <span className="absolute -top-1.5 -right-1.5 grid min-w-5 place-items-center rounded-full border-2 border-surface bg-danger px-1 font-mono text-xs font-semibold text-on-danger">
            {count > 99 ? "99+" : count}
          </span>
        )}
      </PopoverTrigger>
      <PopoverContent
        align="end"
        className="w-[min(24rem,calc(100vw-2rem))] gap-0 p-0"
      >
        <h2 className="border-b border-border px-4 py-3 text-title">
          {t.title}
        </h2>
        <div className="max-h-[60vh] overflow-y-auto">
          {list.isPending ? (
            <Loading className="p-4" />
          ) : list.isError ? (
            <p role="alert" className="p-4 text-sm text-danger-ink">
              {problemOf(list.error)}
            </p>
          ) : list.data.length === 0 ? (
            <p className="p-4 text-sm text-ink-muted">{t.empty}</p>
          ) : (
            <ul className="divide-y divide-border">
              {list.data.slice(0, 6).map((n) => (
                <NotificationItem
                  key={n.id}
                  n={n}
                  onOpen={item.open}
                  onMarkRead={item.markRead}
                  busy={item.busy}
                />
              ))}
            </ul>
          )}
        </div>
        <Link
          href={routes.notifications}
          onClick={() => setOpen(false)}
          className="block border-t border-border px-4 py-3 text-center text-sm font-semibold text-primary hover:bg-surface-sunken"
        >
          {t.viewAll}
        </Link>
      </PopoverContent>
    </Popover>
  );
}
