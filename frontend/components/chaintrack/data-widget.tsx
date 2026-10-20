import type { LucideIcon } from "lucide-react";

import { Notice } from "@/components/chaintrack/notice";
import { Loading, PageState } from "@/components/chaintrack/page-state";
import { problemOf } from "@/lib/api";

/** The query half every section needs: pending, failed or answered. TanStack's
 *  useQuery and usePaged both return this shape, so one component fits all. */
export type QueryState<T> = {
  isPending: boolean;
  isError: boolean;
  error: unknown;
  data: T | undefined;
  refetch?: () => void;
};

/**
 * One loading / error / data branch for every section (patterns.md → Lists).
 * Pending shows Loading, failure shows the API's message in a blocking
 * Notice with a retry when the query offers one, and answers render through
 * the render fn. Empty lists stay the table's own `empty` state. A query
 * that settles with no data is a load failure, not an empty page.
 *
 *   <DataWidget query={volume}>{(rows) => <ClientTable rows={rows} ... />}</DataWidget>
 *
 * Whole pages pass `page`: pending keeps the page frame (`className`), and
 * failure renders a `PageState` instead of an inline Notice:
 *
 *   <DataWidget query={passport} className="p-4 lg:p-8" page={{ icon: Search, title: copy.errors.loadTitle }}>
 *     {(p) => <main>…</main>}
 *   </DataWidget>
 */
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
