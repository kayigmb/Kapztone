import type { LucideIcon } from "lucide-react";
import Link from "next/link";

import { cn } from "@/lib/utils";

// Figma .metric-card: a quiet card; only an alert turns its border and number red.
const look = {
  neutral: {
    card: "border-border",
    value: "",
    note: "text-ink-muted",
    icon: "bg-primary-soft text-primary",
  },
  attention: {
    card: "border-pending/40",
    value: "",
    note: "text-pending-ink",
    icon: "bg-pending-soft text-pending-ink",
  }, // something for you to do
  alert: {
    card: "border-danger/40",
    value: "text-danger",
    note: "text-danger-ink",
    icon: "bg-danger-soft text-danger",
  }, // e.g. open flags
} as const;

/** A dashboard number (components.md → StatTile / metric card). `value` undefined: still loading. */
export function StatTile({
  label,
  note,
  value,
  icon: Icon,
  tone = "neutral",
  href,
  failed,
}: {
  label: string;
  note: string;
  value: number | undefined;
  icon: LucideIcon;
  tone?: keyof typeof look;
  href?: string;
  failed?: string;
}) {
  const l = look[tone];
  const body = (
    <>
      <span className="flex items-center justify-between gap-2 text-small font-semibold text-ink-muted">
        {label}
        <span
          className={cn(
            "grid size-8 shrink-0 place-items-center rounded-md",
            l.icon,
          )}
        >
          <Icon className="size-4" aria-hidden />
        </span>
      </span>
      {value === undefined ? (
        failed ? (
          <span className="text-sm text-ink-muted">{failed}</span>
        ) : (
          <span
            className="h-8 w-12 animate-pulse rounded-sm bg-surface-sunken motion-reduce:animate-none"
            aria-busy="true"
          />
        )
      ) : (
        <span
          className={cn(
            "text-3xl leading-8 font-semibold tracking-tight tabular-nums",
            l.value,
          )}
        >
          {value.toLocaleString()}
        </span>
      )}
      <span className={cn("text-xs font-medium", l.note)}>{note}</span>
    </>
  );
  const cls = cn(
    "grid min-h-28 content-start gap-2 rounded-lg border bg-surface p-4 shadow-card lg:p-5",
    l.card,
  );
  return href ? (
    <Link
      href={href}
      className={cn(cls, "transition-colors hover:border-primary")}
    >
      {body}
    </Link>
  ) : (
    <div className={cls}>{body}</div>
  );
}
