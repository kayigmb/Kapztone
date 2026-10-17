import { ShieldAlert, ShieldCheck, ShieldX } from "lucide-react";

import { cn } from "@/lib/utils";
import { copy, type Result, type Verdict } from "@/lib/verification";
import { verdictIcon } from "./status";
import { tone, verdictTone } from "./tone";

const mark = { verified: ShieldCheck, warning: ShieldAlert, failed: ShieldX };
const ground: Record<Verdict, string> = {
  verified: "bg-verified-soft border-verified/35",
  warning: "bg-warning-soft border-warning/35",
  failed: "bg-danger-soft border-danger/35",
};
const reasonVerdict: Record<Result, Verdict> = { pass: "verified", warn: "warning", fail: "failed" };

/** The result of a passport check, shown once at the top of the public page. */
export function VerdictBadge({ verdict, summary, reasons, meta }: {
  verdict: Verdict; summary: string; reasons: { result: Result; text: string }[]; meta?: React.ReactNode;
}) {
  const t = tone[verdictTone[verdict]];
  const Mark = mark[verdict];
  return (
    <section role="status" className={cn("rounded-lg border p-4 pt-6 md:p-6 md:pt-8", ground[verdict])}>
      <div className="flex items-center gap-4">
        <span className={cn("grid size-14 shrink-0 place-items-center rounded-full md:size-16", t.solid)}>
          <Mark className="size-8" aria-hidden />
        </span>
        <div>
          <p className={cn("label", t.ink)}>{copy.kicker}</p>
          <p className={cn("text-verdict md:text-verdict-lg", t.ink)}>{copy.verdict[verdict]}</p>
        </div>
      </div>
      <p className="mt-4 max-w-[60ch] text-[16px] leading-6 text-ink">{summary}</p>
      <ul className="mt-4 grid gap-2 md:grid-cols-2 md:gap-x-6">
        {reasons.map((r) => {
          const Icon = verdictIcon[reasonVerdict[r.result]];
          return (
            <li key={r.text} className="flex items-start gap-2">
              <Icon className={cn("mt-0.5 size-5 shrink-0", tone[verdictTone[reasonVerdict[r.result]]].fg)} aria-hidden />
              <span className={cn(r.result !== "pass" && "font-semibold")}>{r.text}</span>
            </li>
          );
        })}
      </ul>
      {meta && (
        <p className="mt-4 flex flex-wrap gap-x-3 gap-y-1 border-t border-current/20 pt-3 text-small text-ink-muted">{meta}</p>
      )}
    </section>
  );
}
