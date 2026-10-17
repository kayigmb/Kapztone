import { ArrowLeftRight, CheckCheck, Flag, OctagonX, Package, RefreshCcw, Sprout, type LucideIcon } from "lucide-react";

import { cn } from "@/lib/utils";
import { copy, type View } from "@/lib/verification";
import { StatusChip } from "./status";

const icons: Record<string, LucideIcon> = {
  ISSUE: Sprout, HANDOVER: ArrowLeftRight, TRANSFORM: RefreshCcw, FLAG: Flag, REVOKE: OctagonX,
};

const marker: Record<View["journey"][number]["state"], string> = {
  done: "border-verified text-verified",
  current: "bg-primary border-primary text-on-primary",
  warning: "border-warning text-warning",
  failed: "bg-danger border-danger text-on-danger",
};

/** The ordered history of a lot, origin first. */
export function JourneyTimeline({ steps }: { steps: View["journey"] }) {
  return (
    <ol>
      {steps.map((s, i) => {
        const Icon = icons[s.kind] ?? Package;
        const failed = s.state === "failed";
        return (
          <li key={s.id} className="relative grid grid-cols-[40px_1fr] gap-x-3 pb-6 last:pb-0">
            {i < steps.length - 1 && <span className="absolute top-10 bottom-0 left-[19px] w-0.5 bg-border-strong" aria-hidden />}
            <span className={cn("relative z-10 grid size-10 place-items-center rounded-full border-2 bg-surface", marker[s.state])}>
              <Icon className="size-5" aria-hidden />
            </span>
            <div className={cn("min-w-0", failed && "rounded-md bg-danger-soft p-3")}>
              <div className="flex flex-wrap justify-between gap-x-3">
                <h3 className={cn("text-title", failed && "text-danger-ink")}>{s.title}</h3>
                <time className="text-small text-ink-subtle tabular-nums">{s.date}</time>
              </div>
              <p className="text-body">
                {s.parties.from && s.parties.to ? (
                  <>{s.parties.from} → <span className="font-semibold">{s.parties.to}</span></>
                ) : (
                  (s.parties.to ?? s.parties.from)
                )}
              </p>
              <div className="mt-2 flex flex-wrap gap-2">
                <span className="fig rounded-sm bg-surface-sunken px-2 py-0.5 text-sm">{s.quantity}</span>
                {s.signedByBoth && <StatusChip tone="verified" icon={CheckCheck}>{copy.signedByBoth}</StatusChip>}
              </div>
            </div>
          </li>
        );
      })}
    </ol>
  );
}
