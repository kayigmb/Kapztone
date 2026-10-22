import { Check } from "lucide-react";

import { copy } from "@/lib/consts";
import { cn } from "@/lib/utils";

/** Where someone is in a focused flow: numbered circles joined by a line. */
export function Stepper({
  steps,
  current,
  error,
}: {
  steps: string[];
  current: number;
  error?: boolean;
}) {
  return (
    <ol
      className="flex items-center gap-2"
      aria-label={copy.stepOf(current + 1, steps.length)}
    >
      {steps.map((name, i) => {
        const done = i < current;
        const now = i === current;
        return (
          <li
            key={name}
            className={cn(
              "flex min-w-0 items-center gap-2",
              i < steps.length - 1 && "flex-1",
            )}
            aria-current={now ? "step" : undefined}
          >
            <span
              className={cn(
                "grid size-7 shrink-0 place-items-center rounded-full border text-xs font-semibold",
                done
                  ? "border-primary/25 bg-primary-soft text-primary"
                  : now
                    ? error
                      ? "border-danger bg-danger text-on-danger"
                      : "border-primary bg-primary text-on-primary"
                    : "border-border bg-surface text-ink-muted",
              )}
            >
              {done ? <Check className="size-3.5" aria-hidden /> : i + 1}
            </span>
            <span
              className={cn(
                "truncate text-xs",
                now
                  ? error
                    ? "font-semibold text-danger"
                    : "font-semibold text-ink"
                  : "hidden text-ink-muted sm:inline",
                done && "text-ink-muted",
              )}
            >
              {name}
            </span>
            {i < steps.length - 1 && (
              <span className="h-px min-w-3 flex-1 bg-border" aria-hidden />
            )}
          </li>
        );
      })}
    </ol>
  );
}
