import {
  ArrowLeftRight,
  CheckCheck,
  Flag,
  OctagonX,
  Package,
  RefreshCcw,
  Sprout,
  type LucideIcon,
} from "lucide-react";

import { cn } from "@/lib/utils";
import { copy } from "@/lib/consts";
import { type View } from "@/lib/verification";
import { StatusChip } from "./status";

const icons: Record<string, LucideIcon> = {
  ISSUE: Sprout,
  HANDOVER: ArrowLeftRight,
  TRANSFORM: RefreshCcw,
  FLAG: Flag,
  REVOKE: OctagonX,
};

// Figma .timeline-dot: a soft teal circle; a concern is amber, a revocation red.
const marker: Record<View["journey"][number]["state"], string> = {
  done: "border-primary/25 bg-primary-soft text-primary",
  current: "border-primary bg-primary text-on-primary",
  warning: "border-warning/40 bg-warning-soft text-warning-ink",
  failed: "border-danger bg-danger text-on-danger",
};

/** The ordered history of a lot, origin first (Figma .timeline). */
export function JourneyTimeline({ steps }: { steps: View["journey"] }) {
  return (
    <ol>
      {steps.map((s, i) => {
        const Icon = icons[s.kind] ?? Package;
        const failed = s.state === "failed";
        return (
          <li
            key={s.id}
            className="relative grid grid-cols-[auto_1fr] gap-x-3 pb-6 last:pb-0"
          >
            {i < steps.length - 1 && (
              <span
                className="absolute top-9 bottom-1 left-[15.5px] w-px bg-border"
                aria-hidden
              />
            )}
            <span
              className={cn(
                "relative z-10 grid size-8 place-items-center rounded-full border",
                marker[s.state],
              )}
            >
              <Icon className="size-4" aria-hidden />
            </span>
            <div
              className={cn(
                "grid min-w-0 gap-1",
                failed && "rounded-md bg-danger-soft p-3",
              )}
            >
              <time className="text-xs font-semibold tracking-[0.06em] text-ink-muted uppercase tabular-nums">
                {s.date}
              </time>
              <h3
                className={cn(
                  "text-base leading-5 font-semibold",
                  failed && "text-danger-ink",
                )}
              >
                {s.title}
              </h3>
              <p className="text-sm text-ink-muted">
                {s.parties.from && s.parties.to ? (
                  <>
                    {s.parties.from} →{" "}
                    <span className="font-semibold text-ink">
                      {s.parties.to}
                    </span>
                  </>
                ) : (
                  (s.parties.to ?? s.parties.from)
                )}
              </p>
              <div className="mt-1 flex flex-wrap items-center gap-2">
                <span className="fig rounded-sm border border-border bg-bg px-2 py-0.5 text-sm">
                  {s.quantity}
                </span>
                {s.signedByBoth && (
                  <StatusChip tone="verified" icon={CheckCheck}>
                    {copy.verification.signedByBoth}
                  </StatusChip>
                )}
              </div>
            </div>
          </li>
        );
      })}
    </ol>
  );
}
