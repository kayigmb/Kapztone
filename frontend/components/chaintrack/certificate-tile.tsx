import { CircleCheck, Clock, FileText, OctagonX, TriangleAlert, type LucideIcon } from "lucide-react";

import { cn } from "@/lib/utils";
import { copy, type CertificateState, type View } from "@/lib/verification";
import { tone, type Tone } from "./tone";

const states: Record<CertificateState, { tone: Tone; icon: LucideIcon; border: string }> = {
  valid: { tone: "verified", icon: CircleCheck, border: "border-border" },
  expiring: { tone: "pending", icon: Clock, border: "border-border" },
  expired: { tone: "warning", icon: TriangleAlert, border: "border-warning" },
  revoked: { tone: "danger", icon: OctagonX, border: "border-danger" },
};

export function CertificateTile({ certificate: c }: { certificate: View["certificates"][number] }) {
  const s = states[c.state];
  const Badge = s.icon;
  return (
    <article className={cn("grid grid-cols-[40px_1fr] gap-3 rounded-md border bg-surface p-4", s.border)}>
      <span className="grid size-10 place-items-center rounded-md bg-surface-sunken text-ink-muted">
        <FileText className="size-5" aria-hidden />
      </span>
      <div className="min-w-0">
        <h3 className="text-[15px] font-semibold">{c.type}</h3>
        <p className="text-small text-ink-muted">{c.issuer}</p>
        <p className="font-mono text-xs break-all text-ink-subtle">{c.id}</p>
        <div className="mt-3 flex flex-wrap justify-between gap-2 border-t border-border pt-3">
          <p className="text-small text-ink-muted">
            {c.validFrom} to <span className="font-semibold text-ink">{c.validTo}</span>
          </p>
          <span className={cn("inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-small font-semibold", tone[s.tone].soft)}>
            <Badge className="size-4" aria-hidden />
            {copy.certificate[c.state]}
          </span>
        </div>
      </div>
      {c.note && <p className={cn("col-span-2 rounded-sm px-3 py-2 text-small", tone[s.tone].soft)}>{c.note}</p>}
    </article>
  );
}
