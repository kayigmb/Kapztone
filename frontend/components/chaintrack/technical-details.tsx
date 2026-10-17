import { ChevronDown, Lock } from "lucide-react";

import { copy, type View } from "@/lib/verification";

/** The only place fingerprints and ledger words appear. Open by default when the verdict is Failed. */
export function TechnicalDetails({ rows, open }: { rows: View["technical"]; open?: boolean }) {
  return (
    <details open={open} className="group rounded-lg border border-border bg-surface shadow-card">
      <summary className="flex min-h-tap cursor-pointer list-none items-center gap-2 px-4 py-3 font-semibold [&::-webkit-details-marker]:hidden">
        <Lock className="size-5 text-ink-muted" aria-hidden />
        {copy.sections.technical}
        <ChevronDown className="ml-auto size-5 text-ink-muted transition-transform group-open:rotate-180" aria-hidden />
      </summary>
      <div className="grid gap-3 border-t border-border p-4">
        <p className="text-small text-ink-muted">{copy.technicalIntro}</p>
        <dl className="grid gap-3">
          {rows.map((r, i) => (
            <div key={i}>
              <dt className="text-small font-semibold">{r.label}</dt>
              <dd className="font-mono text-hash break-all text-ink-muted">{r.value ?? copy.notSealed}</dd>
            </div>
          ))}
        </dl>
      </div>
    </details>
  );
}
