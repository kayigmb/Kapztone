import type { LucideIcon } from "lucide-react";

import { cn } from "@/lib/utils";
import { copy } from "@/lib/consts";
import { type Verdict } from "@/lib/verification";
import { VerdictStatus } from "./status";

export type Fact = {
  label: string;
  value: React.ReactNode;
  mono?: boolean;
  wide?: boolean;
};

export function KeyValue({
  facts,
  className,
}: {
  facts: Fact[];
  className?: string;
}) {
  return (
    <dl className={cn("grid grid-cols-2 gap-4", className)}>
      {facts.map((f) => (
        <div key={f.label} className={cn("min-w-0", f.wide && "col-span-2")}>
          <dt className="label mb-0.5">{f.label}</dt>
          <dd className={cn("text-body break-words", f.mono && "fig")}>
            {f.value}
          </dd>
        </div>
      ))}
    </dl>
  );
}

/** The product summary of one lot. */
export function PassportCard({
  product,
  lotCode,
  icon: Icon,
  facts,
  footer,
  flagged,
}: {
  product: string;
  lotCode: string;
  icon: LucideIcon;
  facts: Fact[];
  footer: { left: React.ReactNode; verdict: Verdict };
  flagged?: boolean;
}) {
  return (
    <article
      className={cn(
        "overflow-hidden rounded-lg border bg-surface shadow-card",
        flagged ? "border-danger" : "border-border",
      )}
    >
      <header className="flex gap-3 border-b border-border p-4">
        <span
          className={cn(
            "grid size-11 shrink-0 place-items-center rounded-md",
            flagged
              ? "bg-danger-soft text-danger"
              : "bg-primary-soft text-primary",
          )}
        >
          <Icon className="size-5" aria-hidden />
        </span>
        <div className="min-w-0">
          <h2 className="text-title">{product}</h2>
          <p className="code break-all text-ink-muted">{lotCode}</p>
        </div>
      </header>
      <div className="p-4 md:p-6">
        <KeyValue facts={facts} />
      </div>
      <footer className="flex flex-wrap items-center justify-between gap-2 bg-surface-sunken px-4 py-3 text-small text-ink-muted">
        <span>{footer.left}</span>
        <VerdictStatus verdict={footer.verdict}>
          {copy.verification.verdict[footer.verdict]}
        </VerdictStatus>
      </footer>
    </article>
  );
}
