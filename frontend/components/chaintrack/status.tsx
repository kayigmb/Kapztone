import {
  Ban,
  CircleCheck,
  Clock,
  OctagonX,
  TriangleAlert,
  type LucideIcon,
} from "lucide-react";

import { copy, orgStatus, type OrgStatus } from "@/lib/consts";
import { cn } from "@/lib/utils";
import type { Verdict } from "@/lib/verification";
import { tone, verdictTone, type Tone } from "./tone";

// A status is always an icon and a word, never colour alone.
export const verdictIcon: Record<Verdict, LucideIcon> = {
  verified: CircleCheck,
  warning: TriangleAlert,
  failed: OctagonX,
};

/** Inline status for card footers, table cells and key-value values. */
export function Status({
  tone: t,
  icon: Icon,
  children,
  className,
}: {
  tone: Tone;
  icon: LucideIcon;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 text-sm font-semibold",
        tone[t].fg,
        className,
      )}
    >
      <Icon className="size-4 shrink-0" aria-hidden />
      {children}
    </span>
  );
}

export function VerdictStatus({
  verdict,
  children,
}: {
  verdict: Verdict;
  children: React.ReactNode;
}) {
  return (
    <Status tone={verdictTone[verdict]} icon={verdictIcon[verdict]}>
      {children}
    </Status>
  );
}

/** Pill for where a record stands: "Signed by both parties", "Expired". */
export function StatusChip({
  tone: t,
  icon: Icon,
  children,
}: {
  tone: Tone;
  icon: LucideIcon;
  children: React.ReactNode;
}) {
  return (
    <span
      className={cn(
        "inline-flex min-h-6.5 w-fit items-center gap-1 rounded-full border py-0.75 pr-2.5 pl-2 text-xs font-bold whitespace-nowrap",
        tone[t].soft,
        tone[t].chip,
      )}
    >
      <Icon className="size-3.5 shrink-0" aria-hidden />
      {children}
    </span>
  );
}

// Where an organisation's verification stands.
const orgStatusLook: Record<OrgStatus, { tone: Tone; icon: LucideIcon }> = {
  [orgStatus.pending]: { tone: "pending", icon: Clock },
  [orgStatus.verified]: { tone: "verified", icon: CircleCheck },
  [orgStatus.rejected]: { tone: "danger", icon: OctagonX },
  [orgStatus.changesRequested]: { tone: "warning", icon: TriangleAlert },
  [orgStatus.suspended]: { tone: "danger", icon: Ban },
};

export function OrgStatusChip({ status }: { status: string }) {
  const look = orgStatusLook[status as OrgStatus] ?? orgStatusLook.pending;
  return (
    <StatusChip tone={look.tone} icon={look.icon}>
      {copy.orgStatus[status as OrgStatus] ?? status}
    </StatusChip>
  );
}
