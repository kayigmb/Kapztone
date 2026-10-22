import type { LucideIcon } from "lucide-react";

import { Notice } from "@/components/chaintrack/notice";
import { Loading, PageState } from "@/components/chaintrack/page-state";
import { problemOf } from "@/lib/api";

/** The query half every section needs: pending, failed or answered. */
export type QueryState<T> = {
  isPending: boolean;
  isError: boolean;
  error: unknown;
  data: T | undefined;
  refetch?: () => void;
};

/** One loading / error / data branch for every section. */
export function DataWidget<T>({
  query,
  retryLabel,
  className,
  page,
  children,
}: {
  query: QueryState<T>;
  retryLabel?: string;
  /** Padding / layout for the Loading state; pages pass the page frame. */
  className?: string;
  /** Render failures as a whole-page outcome instead of an inline Notice. */
  page?: { icon: LucideIcon; title: string };
  children: (data: T) => React.ReactNode;
}) {
  if (query.isPending) return <Loading className={className} />;
  if (query.isError || query.data == null) {
    const message =
      query.error === undefined ? undefined : problemOf(query.error);
    if (page) {
      return (
        <PageState
          icon={page.icon}
          tone="danger"
          title={page.title}
          retry={query.refetch ? () => query.refetch?.() : undefined}
        >
          {message}
        </PageState>
      );
    }
    return (
      <Notice kind="blocking">
        {message !== undefined && <p>{message}</p>}
        {query.refetch && (
          <button
            type="button"
            onClick={() => query.refetch?.()}
            className="font-semibold underline underline-offset-2"
          >
            {retryLabel ?? "Try again"}
          </button>
        )}
      </Notice>
    );
  }
  return <>{children(query.data)}</>;
}
