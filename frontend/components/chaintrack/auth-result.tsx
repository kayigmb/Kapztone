import type { LucideIcon } from "lucide-react";

import { cn } from "@/lib/utils";

const tones = {
  verified: "bg-verified-soft text-verified",
  warning: "bg-warning-soft text-warning-ink",
  danger: "bg-danger-soft text-danger",
  neutral: "bg-primary-soft text-primary",
} as const;

/** The outcome of an access step (drafts 2e, 2g, 2h): a round mark, a title, what happened, what to do next. */
export function AuthResult({
  icon: Icon,
  tone = "neutral",
  title,
  children,
  actions,
}: {
  icon: LucideIcon;
  tone?: keyof typeof tones;
  title: string;
  children: React.ReactNode;
  actions?: React.ReactNode;
}) {
  return (
    <main
      className="mx-auto grid w-full max-w-100 gap-5"
      role={tone === "danger" || tone === "warning" ? "alert" : "status"}
    >
      <span
        className={cn(
          "grid size-14 place-items-center rounded-full",
          tones[tone],
        )}
      >
        <Icon className="size-7" aria-hidden />
      </span>
      <h1 className="text-page">{title}</h1>
      <div className="grid gap-3 text-sm text-ink-muted">{children}</div>
      {actions && <div className="grid gap-3">{actions}</div>}
    </main>
  );
}
