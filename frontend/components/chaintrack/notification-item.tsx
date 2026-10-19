"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Check } from "lucide-react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";

import { endpoints, problemOf, type NotificationRow } from "@/lib/api";
import { copy, queryKeys, routes, subjectTypes } from "@/lib/consts";
import { certificateKinds, linkOf } from "@/lib/notifications";
import { cn } from "@/lib/utils";
import { formatDateTime } from "@/lib/verification";

const t = copy.notifications;

/** Opening a notification marks it read and goes to what it is about. */
export function useOpenNotification(onOpened?: () => void) {
  const router = useRouter();
  const client = useQueryClient();
  const read = useMutation({
    mutationFn: endpoints.markRead,
    onSuccess: () =>
      client.invalidateQueries({ queryKey: queryKeys.notifications }), // also refreshes the unread count
  });
  const open = async (n: NotificationRow) => {
    if (!n.read_at) read.mutate(n.id);
    let href = linkOf(n.kind, n.subject_id);
    if (!href && certificateKinds.includes(n.kind)) {
      try {
        const c = await endpoints.attestation(n.subject_id);
        href =
          c.subject_type === subjectTypes.passport
            ? routes.passportDetail(c.subject_id)
            : routes.facility(c.subject_id);
      } catch (err) {
        toast.error(problemOf(err));
      }
    }
    onOpened?.();
    if (href) router.push(href);
  };
  return {
    open,
    markRead: (id: string) => read.mutate(id),
    busy: read.isPending,
  };
}

/** One notification: title by kind, the API's message, when; unread ones tinted. */
export function NotificationItem({
  n,
  onOpen,
  onMarkRead,
  busy,
}: {
  n: NotificationRow;
  onOpen: (n: NotificationRow) => void;
  onMarkRead: (id: string) => void;
  busy?: boolean;
}) {
  const title = t.kinds[n.kind];
  return (
    <li
      className={cn(
        "flex items-start gap-2 px-4 py-3",
        !n.read_at && "bg-primary-soft/60",
      )}
    >
      <button
        type="button"
        onClick={() => onOpen(n)}
        className="grid min-w-0 flex-1 gap-0.5 text-left"
      >
        <span className="flex items-center gap-2">
          {!n.read_at && (
            <span
              className="size-2 shrink-0 rounded-full bg-primary"
              aria-label={t.unread}
            />
          )}
          {title && <span className="text-sm font-semibold">{title}</span>}
        </span>
        <span className={cn("text-sm", title && "text-ink-muted")}>
          {n.message}
        </span>
        <span className="text-xs text-ink-subtle">
          {formatDateTime(n.created_at)}
        </span>
      </button>
      {n.read_at ? (
        <span className="flex h-8 items-center gap-1 text-xs text-ink-subtle">
          <Check className="size-4" aria-hidden />
          {t.read}
        </span>
      ) : (
        <button
          type="button"
          onClick={() => onMarkRead(n.id)}
          disabled={busy}
          className="h-8 shrink-0 rounded-md px-2 text-xs font-semibold text-primary hover:bg-surface-sunken"
        >
          {t.markRead}
        </button>
      )}
    </li>
  );
}
