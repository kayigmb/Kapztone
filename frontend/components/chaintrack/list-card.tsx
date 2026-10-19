import Link from "next/link";

import { Loading } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";

/**
 * A titled card holding one list, with its empty, loading and failed states
 * (Figma .app-card + .app-card-head). Rows are `<li>`s; give them `px-5`.
 */
export function ListCard({
  title,
  sub,
  empty,
  href,
  state,
  error,
  children,
}: {
  title: string;
  sub?: string;
  empty: string;
  href?: string;
  state: "loading" | "failed" | "empty" | "ready";
  error?: string;
  children?: React.ReactNode;
}) {
  return (
    <section className="min-w-0 overflow-hidden rounded-lg border border-border bg-surface shadow-card">
      <header className="flex min-h-16 flex-wrap items-center justify-between gap-2 border-b border-border px-5 py-3">
        <div className="grid gap-0.5">
          <h2 className="text-base leading-5 font-semibold">{title}</h2>
          {sub && <p className="text-xs text-ink-muted">{sub}</p>}
        </div>
        {href && state === "ready" && (
          <Link
            href={href}
            className="inline-flex min-h-tap items-center text-sm font-semibold text-primary hover:underline"
          >
            {copy.dashboard.all}
          </Link>
        )}
      </header>
      {state === "loading" && <Loading className="px-5 py-5" />}
      {state === "failed" && (
        <p role="alert" className="px-5 py-5 text-sm text-danger-ink">
          {error}
        </p>
      )}
      {state === "empty" && (
        <p className="px-5 py-5 text-sm text-ink-muted">{empty}</p>
      )}
      {state === "ready" && (
        <ul className="grid divide-y divide-border">{children}</ul>
      )}
    </section>
  );
}
