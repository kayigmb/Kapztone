import { CircleCheck, OctagonX, TriangleAlert } from "lucide-react";

import { capacityNearShare, copy } from "@/lib/consts";
import { cn } from "@/lib/utils";

const t = copy.dashboard;

/** How much of a facility's declared capacity this period has used (components.md → Meter, compact; Figma .meter-block). */
export function CapacityMeter({ title, used, max, unit }: { title: string; used: number; max: number; unit: string }) {
  const share = max > 0 ? used / max : 1;
  const state = used > max ? "over" : share >= capacityNearShare ? "near" : "within";
  const fill = { within: "bg-verified", near: "bg-warning", over: "bg-danger" }[state];
  const fmt = (n: number) => `${n.toLocaleString()} ${unit}`;
  const Icon = { within: CircleCheck, near: TriangleAlert, over: OctagonX }[state];
  const message = {
    within: t.capacityLeft(fmt(max - used)),
    near: `${t.capacityNear} · ${t.capacityLeft(fmt(Math.max(0, max - used)))}`,
    over: t.capacityOver,
  }[state];

  return (
    <div className="grid gap-2">
      <span className="min-w-0 text-sm font-semibold">{title}</span>
      <div className="flex flex-wrap justify-between gap-2 text-xs text-ink-muted">
        <span className="font-mono">{t.capacityOf(used.toLocaleString(), fmt(max))}</span>
        <span className="font-mono font-semibold text-ink">{Math.round(share * 100)}%</span>
      </div>
      <div
        role="meter" aria-label={title} aria-valuenow={used} aria-valuemin={0} aria-valuemax={max}
        className="relative h-2 overflow-hidden rounded-full bg-surface-sunken"
      >
        <span className={cn("absolute inset-y-0 left-0 rounded-full", fill)} style={{ width: `${Math.min(100, share * 100)}%` }} />
      </div>
      <p className={cn("flex items-center gap-1.5 rounded-md px-2.5 py-2 text-xs font-semibold",
        { within: "bg-verified-soft text-verified-ink", near: "bg-warning-soft text-warning-ink", over: "bg-danger-soft text-danger-ink" }[state])}>
        <Icon className="size-4 shrink-0" aria-hidden />
        {message}
      </p>
    </div>
  );
}
