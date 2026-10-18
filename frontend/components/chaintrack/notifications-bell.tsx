"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Bell, Check } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { toast } from "sonner";

import { Loading } from "@/components/chaintrack/page-state";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { endpoints, problemOf, type NotificationRow } from "@/lib/api";
import { copy, notificationRefetchMs, queryKeys, routes } from "@/lib/consts";
import { cn } from "@/lib/utils";
import { formatDateTime } from "@/lib/verification";

const t = copy.notifications;

// Where each kind of notification leads, for the screens that exist.
const linkOf = (n: NotificationRow): string | undefined => ({
  org_registered: routes.adminOrg("organisations", n.subject_id),
  documents_uploaded: routes.adminOrg("organisations", n.subject_id),
  passport_revoked: routes.passport(n.subject_id),
} as Record<string, string>)[n.kind];

/** Header bell: unread count, and the latest notifications with mark-as-read. */
export function NotificationsBell() {
  const [open, setOpen] = useState(false);
  const client = useQueryClient();
  const unread = useQuery({ queryKey: queryKeys.unread, queryFn: endpoints.unreadCount, refetchInterval: notificationRefetchMs });
  const list = useQuery({ queryKey: queryKeys.notifications, queryFn: endpoints.notifications, enabled: open });
  const markRead = useMutation({
    mutationFn: endpoints.markRead,
    onSuccess: () => client.invalidateQueries({ queryKey: queryKeys.notifications }), // also refreshes the unread count
    onError: (err) => toast.error(problemOf(err)),
  });
  const count = unread.data ?? 0;

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        aria-label={t.open(count)}
        className="relative grid size-tap place-items-center rounded-md border border-border text-ink-muted hover:bg-surface-sunken hover:text-ink"
      >
        <Bell className="size-5" aria-hidden />
        {count > 0 && (
          <span className="absolute -top-1.5 -right-1.5 grid min-w-5 place-items-center rounded-full border-2 border-surface bg-danger px-1 font-mono text-xs font-semibold text-on-danger">
            {count > 99 ? "99+" : count}
          </span>
        )}
      </PopoverTrigger>
      <PopoverContent align="end" className="w-[min(24rem,calc(100vw-2rem))] gap-0 p-0">
        <h2 className="border-b border-border px-4 py-3 text-title">{t.title}</h2>
        <div className="max-h-[60vh] overflow-y-auto">
          {list.isPending ? <Loading className="p-4" /> : list.isError ? (
            <p role="alert" className="p-4 text-sm text-danger-ink">{problemOf(list.error)}</p>
          ) : list.data.length === 0 ? (
            <p className="p-4 text-sm text-ink-muted">{t.empty}</p>
          ) : (
            <ul className="divide-y divide-border">
              {list.data.map((n) => {
                const href = linkOf(n);
                const title = t.kinds[n.kind];
                const body = (
                  <>
                    {title && <span className="text-sm font-semibold">{title}</span>}
                    <span className={cn("text-sm", title && "text-ink-muted")}>{n.message}</span>
                    <span className="text-xs text-ink-subtle">{formatDateTime(n.created_at)}</span>
                  </>
                );
                return (
                  <li key={n.id} className={cn("flex items-start gap-2 px-4 py-3", !n.read_at && "bg-primary-soft")}>
                    {href ? (
                      <Link href={href} onClick={() => setOpen(false)} className="grid min-w-0 flex-1 gap-0.5">{body}</Link>
                    ) : (
                      <div className="grid min-w-0 flex-1 gap-0.5">{body}</div>
                    )}
                    {n.read_at ? (
                      <span className="flex min-h-tap items-center gap-1 text-xs text-ink-subtle"><Check className="size-4" aria-hidden />{t.read}</span>
                    ) : (
                      <button type="button" onClick={() => markRead.mutate(n.id)} disabled={markRead.isPending}
                        className="min-h-tap shrink-0 rounded-md px-2 text-xs font-semibold text-primary hover:bg-surface-sunken">
                        {t.markRead}
                      </button>
                    )}
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}
