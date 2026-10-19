import { cn } from "@/lib/utils";
import { copy } from "@/lib/consts";
import { type View } from "@/lib/verification";
import { verdictIcon } from "./status";
import { tone, verdictTone } from "./tone";

/** "What we checked" (Figma .checks-list): one row per rule, an icon and a word for what was found. */
export function ChecksList({ checks }: { checks: View["checks"] }) {
  return (
    <ul>
      {checks.map((c) => {
        const Icon = verdictIcon[c.result];
        const look = tone[verdictTone[c.result]];
        return (
          <li
            key={c.name}
            className="grid grid-cols-[auto_1fr] gap-x-3 border-t border-border py-3 first:border-0 first:pt-0"
          >
            <span
              className={cn(
                "grid size-7 place-items-center rounded-full",
                look.solid,
              )}
            >
              <Icon className="size-4" aria-hidden />
            </span>
            <div className="grid min-w-0 gap-1">
              <div className="flex flex-wrap items-baseline justify-between gap-x-2">
                <span className="text-sm font-semibold">{c.label}</span>
                <span className={cn("text-xs font-bold", look.ink)}>
                  {copy.verification.result[c.result]}
                </span>
              </div>
              <p className="text-xs leading-[18px] text-ink-muted">
                {c.explanation}
              </p>
            </div>
          </li>
        );
      })}
    </ul>
  );
}
