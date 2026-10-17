import { Check } from "lucide-react";

import { cn } from "@/lib/utils";

/** Where someone is in a focused flow. `current` counts from 0; `error` marks the current step red. */
export function Stepper({ steps, current, error }: { steps: string[]; current: number; error?: boolean }) {
  return (
    <ol className="flex gap-1" aria-label={`Step ${current + 1} of ${steps.length}`}>
      {steps.map((name, i) => {
        const done = i < current;
        const now = i === current;
        return (
          <li key={name} className="grid flex-1 gap-1.5" aria-current={now ? "step" : undefined}>
            <span className={cn("h-1 rounded-sm",
              done ? "bg-verified" : now ? (error ? "bg-danger" : "bg-primary") : "bg-border")} />
            <span className={cn("flex items-center gap-1 truncate text-xs",
              done ? "text-ink-muted" : now ? (error ? "font-semibold text-danger" : "font-semibold text-ink") : "text-ink-subtle")}>
              {done && <Check className="size-3 shrink-0" aria-hidden />}
              {name}
            </span>
          </li>
        );
      })}
    </ol>
  );
}
