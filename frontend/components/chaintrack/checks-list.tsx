import { cn } from "@/lib/utils";
import { copy, type View } from "@/lib/verification";
import { verdictIcon, VerdictStatus } from "./status";
import { tone, verdictTone } from "./tone";

/** "What we checked": one row per rule, with what was found. */
export function ChecksList({ checks }: { checks: View["checks"] }) {
  return (
    <ul>
      {checks.map((c) => {
        const Icon = verdictIcon[c.result];
        return (
          <li key={c.name} className="grid grid-cols-[24px_1fr_auto] gap-x-3 border-t border-border py-3 first:border-0 first:pt-0">
            <Icon className={cn("size-5", tone[verdictTone[c.result]].fg)} aria-hidden />
            <span className="font-semibold">{c.label}</span>
            <VerdictStatus verdict={c.result}>{copy.result[c.result]}</VerdictStatus>
            <p className="col-span-2 col-start-2 text-small text-ink-muted">{c.explanation}</p>
          </li>
        );
      })}
    </ul>
  );
}
