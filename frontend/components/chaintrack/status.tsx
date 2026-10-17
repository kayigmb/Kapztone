import { CircleCheck, OctagonX, TriangleAlert, type LucideIcon } from "lucide-react";

import { cn } from "@/lib/utils";
import type { Verdict } from "@/lib/verification";
import { tone, verdictTone, type Tone } from "./tone";

// A status is always an icon and a word, never colour alone.
export const verdictIcon: Record<Verdict, LucideIcon> = { verified: CircleCheck, warning: TriangleAlert, failed: OctagonX };

/** Inline status for card footers, table cells and key-value values. */
export function Status({ tone: t, icon: Icon, children, className }: {
  tone: Tone; icon: LucideIcon; children: React.ReactNode; className?: string;
}) {
  return (
    <span className={cn("inline-flex items-center gap-1.5 text-sm font-semibold", tone[t].fg, className)}>
      <Icon className="size-4 shrink-0" aria-hidden />
      {children}
    </span>
  );
}

export function VerdictStatus({ verdict, children }: { verdict: Verdict; children: React.ReactNode }) {
  return <Status tone={verdictTone[verdict]} icon={verdictIcon[verdict]}>{children}</Status>;
}

/** Pill for where a record stands: "Signed by both parties", "Expired". */
export function StatusChip({ tone: t, icon: Icon, children }: { tone: Tone; icon: LucideIcon; children: React.ReactNode }) {
  return (
    <span className={cn(
      "inline-flex min-h-[26px] items-center gap-1 rounded-full border py-[3px] pr-2.5 pl-2 text-small font-semibold whitespace-nowrap",
      tone[t].soft, tone[t].chip,
    )}>
      <Icon className="size-4 shrink-0" aria-hidden />
      {children}
    </span>
  );
}
