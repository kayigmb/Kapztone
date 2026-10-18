import { CircleCheck, Info, OctagonX, TriangleAlert } from "lucide-react";

import { cn } from "@/lib/utils";

// Figma .notice: a tinted box with a hairline border in the same tone.
const kinds = {
  info: { icon: Info, classes: "border-pending/25 bg-pending-soft text-pending-ink" },
  success: { icon: CircleCheck, classes: "border-verified/25 bg-verified-soft text-verified-ink" },
  warning: { icon: TriangleAlert, classes: "border-warning/30 bg-warning-soft text-warning-ink" },
  blocking: { icon: OctagonX, classes: "border-danger bg-danger-soft text-danger-ink" },
};

/** Inline notice. A blocking notice leads with a bold sentence and is announced. */
export function Notice({ kind, children }: { kind: keyof typeof kinds; children: React.ReactNode }) {
  const { icon: Icon, classes } = kinds[kind];
  return (
    <div role={kind === "blocking" ? "alert" : undefined} className={cn("flex items-start gap-3 rounded-md border px-4 py-3 text-sm", classes)}>
      <Icon className="mt-0.5 size-5 shrink-0" aria-hidden />
      <div className="grid gap-1 [&_strong]:font-semibold">{children}</div>
    </div>
  );
}
