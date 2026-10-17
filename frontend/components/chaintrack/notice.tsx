import { CircleCheck, Info, OctagonX, TriangleAlert } from "lucide-react";

import { cn } from "@/lib/utils";

const kinds = {
  info: { icon: Info, classes: "bg-pending-soft text-pending-ink" },
  success: { icon: CircleCheck, classes: "bg-verified-soft text-verified-ink" },
  warning: { icon: TriangleAlert, classes: "bg-warning-soft text-warning-ink" },
  blocking: { icon: OctagonX, classes: "bg-danger-soft text-danger-ink border border-danger" },
};

/** Inline notice. A blocking notice leads with a bold sentence and is announced. */
export function Notice({ kind, children }: { kind: keyof typeof kinds; children: React.ReactNode }) {
  const { icon: Icon, classes } = kinds[kind];
  return (
    <div role={kind === "blocking" ? "alert" : undefined} className={cn("flex items-start gap-3 rounded-md px-4 py-3 text-sm", classes)}>
      <Icon className="mt-0.5 size-5 shrink-0" aria-hidden />
      <div>{children}</div>
    </div>
  );
}
