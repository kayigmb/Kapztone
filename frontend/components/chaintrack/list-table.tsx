"use client";

// Every table in the app is tablecn's data table: sorting by header, rows per
// page, column visibility, a Table | Cards toggle (cards on phones by
// default), and an `actions` column always pinned to the end.
// ListTable shows one page from the API (the API pages and sorts);
// ClientTable shows a short list the screen already has (reports, rankings).

import {
  FlexRender,
  type ColumnDef,
  type RowData,
} from "@tanstack/react-table";
import { LayoutGrid, Table2 } from "lucide-react";
import Link from "next/link";
import { parseAsStringLiteral, useQueryState } from "nuqs";

import { Notice } from "@/components/chaintrack/notice";
import { DataTable } from "@/components/data-table/data-table";
import { DataTableColumnHeader } from "@/components/data-table/data-table-column-header";
import { DataTableSkeleton } from "@/components/data-table/data-table-skeleton";
import { DataTableViewOptions } from "@/components/data-table/data-table-view-options";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useDataTable } from "@/hooks/use-data-table";
import { copy, listViews, type ListView } from "@/lib/consts";
import type { DataTableFeatures } from "@/lib/data-table-features";
import { tableSort } from "@/lib/list";
import { listKeys, type ListParamsState } from "@/lib/list-params";
import { cn } from "@/lib/utils";

/** A table column. Its id is the API's sort column when it sorts. The actions column's id is `actions`. */
export type ListColumn<T extends RowData> = ColumnDef<
  DataTableFeatures,
  T,
  unknown
>;
type Table<T extends RowData> = ReturnType<typeof useDataTable<T>>["table"];

/** The id of the column that is always pinned to the end. */
export const actionsColumnId = "actions";
const pinning = { start: [], end: [actionsColumnId] };

const t = copy.list;
const viewIcon = { table: Table2, cards: LayoutGrid } satisfies Record<
  ListView,
  unknown
>;

type Common<T extends RowData> = {
  columns: ListColumn<T>[];
  rows: T[];
  rowKey: (row: T) => string;
  empty: string;
  label: string;
  /** Search and filters, left of the view controls. */
  toolbar?: React.ReactNode;
};

/** One page from the API: the URL's page, perPage and sort drive the request (lib/list-params.ts). */
export function ListTable<T extends RowData>({
  list,
  total,
  loading,
  error,
  ...p
}: Common<T> & {
  list: ListParamsState;
  total: number;
  loading?: boolean;
  error?: string;
}) {
  const { table } = useDataTable({
    data: p.rows,
    columns: p.columns,
    getRowId: p.rowKey,
    pageCount: Math.max(1, Math.ceil(total / list.size)),
    initialState: {
      sorting: tableSort(list.sort),
      pagination: { pageIndex: 0, pageSize: list.size },
      columnPinning: pinning,
    },
  });
  return (
    <Frame
      table={table}
      {...p}
      loading={loading}
      error={error}
      viewKey={listKeys.view}
    />
  );
}

/** A short list already on the page: sorted and paged in the browser. `id` keeps its address keys apart from other tables. */
export function ClientTable<T extends RowData>({
  id,
  ...p
}: Common<T> & { id: string }) {
  const { table } = useDataTable({
    mode: "client",
    data: p.rows,
    columns: p.columns,
    getRowId: p.rowKey,
    queryKeys: {
      page: `${id}Page`,
      perPage: `${id}PerPage`,
      sort: `${id}Sort`,
      joinOperator: `${id}Join`,
    },
    initialState: { columnPinning: pinning },
  });
  return <Frame table={table} {...p} viewKey={`${id}View`} />;
}

function Frame<T extends RowData>({
  table,
  columns,
  rows,
  empty,
  label,
  toolbar,
  loading,
  error,
  viewKey,
}: Common<T> & {
  table: Table<T>;
  loading?: boolean;
  error?: string;
  viewKey: string;
}) {
  const [view, setView] = useQueryState(
    viewKey,
    parseAsStringLiteral(listViews).withOptions({
      history: "replace",
      scroll: false,
    }),
  );
  const controls = (
    <div className="flex flex-wrap items-center justify-between gap-2">
      <div className="flex min-w-0 flex-1 flex-wrap items-center gap-2">
        {toolbar}
      </div>
      <div className="flex items-center gap-2">
        <div
          role="group"
          aria-label={t.viewAs}
          className="inline-flex rounded-md border border-border bg-surface p-0.5"
        >
          {listViews.map((v) => {
            const Icon = viewIcon[v];
            const on = view === v;
            return (
              <Tooltip key={v}>
                <TooltipTrigger
                  render={
                    <button
                      type="button"
                      aria-pressed={on}
                      aria-label={t.views[v]}
                      onClick={() => setView(on ? null : v)}
                      className={cn(
                        "grid size-8 place-items-center rounded-sm text-ink-muted hover:text-ink",
                        on && "bg-primary-soft text-primary",
                      )}
                    />
                  }
                >
                  <Icon className="size-4" aria-hidden />
                </TooltipTrigger>
                <TooltipContent>{t.views[v]}</TooltipContent>
              </Tooltip>
            );
          })}
        </div>
        <DataTableViewOptions table={table} />
      </div>
    </div>
  );

  if (loading) {
    return (
      <div className="grid gap-3">
        {controls}
        <DataTableSkeleton
          columnCount={columns.length}
          rowCount={10}
          withViewOptions={false}
        />
      </div>
    );
  }
  if (error)
    return (
      <div className="grid gap-3">
        {controls}
        <Notice kind="blocking">{error}</Notice>
      </div>
    );

  // No explicit choice: cards below md, the table from md up.
  const showTable =
    view === "table" ? "flex" : view === "cards" ? "hidden" : "hidden md:flex";
  const showCards =
    view === "cards" ? "grid" : view === "table" ? "hidden" : "grid md:hidden";

  return (
    <div className="grid gap-3">
      {controls}
      {rows.length === 0 ? (
        <p className="rounded-lg border border-border bg-surface p-5 text-sm text-ink-muted">
          {empty}
        </p>
      ) : (
        <>
          <DataTable table={table} aria-label={label} className={showTable} />
          <ul
            className={cn(showCards, "gap-3 sm:grid-cols-2")}
            aria-label={label}
          >
            {table.getRowModel().rows.map((row) => {
              const cells = row.getVisibleCells();
              const [title, ...rest] = cells.filter(
                (c) => c.column.id !== actionsColumnId,
              );
              const actions = cells.find(
                (c) => c.column.id === actionsColumnId,
              );
              return (
                <li
                  key={row.id}
                  className="grid gap-3 rounded-lg border border-border bg-surface p-4 shadow-card"
                >
                  {title && (
                    <div className="min-w-0 font-semibold">
                      <FlexRender cell={title} />
                    </div>
                  )}
                  <dl className="grid gap-2 text-sm">
                    {rest.map((c) => (
                      <div
                        key={c.id}
                        className="grid grid-cols-3 items-center gap-2"
                      >
                        <dt className="text-xs text-ink-muted">
                          {(
                            c.column.columnDef.meta as
                              { label?: string } | undefined
                          )?.label ?? c.column.id}
                        </dt>
                        <dd className="col-span-2 min-w-0">
                          <FlexRender cell={c} />
                        </dd>
                      </div>
                    ))}
                  </dl>
                  {actions && (
                    <div className="flex justify-end border-t border-border pt-3">
                      <FlexRender cell={actions} />
                    </div>
                  )}
                </li>
              );
            })}
          </ul>
        </>
      )}
    </div>
  );
}

/**
 * A column with tablecn's header (sort menu, hide). With `sortable`, `id` is
 * the API's sort column (ListTable); with `value`, the browser sorts on it (ClientTable).
 */
export function column<T extends RowData>(
  id: string,
  label: string,
  cell: (row: T) => React.ReactNode,
  opts: { sortable?: boolean; value?: (row: T) => string | number } = {},
): ListColumn<T> {
  return {
    id,
    accessorFn: opts.value ?? (() => null),
    header: ({ column: c }) => (
      <DataTableColumnHeader column={c} label={label} />
    ),
    cell: ({ row }) => cell(row.original),
    meta: { label },
    enableSorting: !!(opts.sortable || opts.value),
    enableHiding: true,
  };
}

/** The row's actions: always last, always pinned, never hidden. */
export function actionsColumn<T extends RowData>(
  cell: (row: T) => React.ReactNode,
): ListColumn<T> {
  return {
    id: actionsColumnId,
    header: () => <span className="sr-only">{t.actions}</span>,
    cell: ({ row }) => (
      <div className="flex items-center justify-end gap-1">
        {cell(row.original)}
      </div>
    ),
    meta: { label: t.actions },
    enableSorting: false,
    enableHiding: false,
  };
}

/** An icon-only row action with a tooltip; `danger` for suspend, revoke and the like. */
export function RowAction({
  label,
  icon: Icon,
  onClick,
  href,
  danger,
  disabled,
}: {
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  onClick?: () => void;
  href?: string;
  danger?: boolean;
  disabled?: boolean;
}) {
  const className = cn(
    "grid size-8 place-items-center rounded-md hover:bg-surface-sunken disabled:opacity-40",
    danger ? "text-danger" : "text-ink-muted hover:text-ink",
  );
  return (
    <Tooltip>
      <TooltipTrigger
        render={
          href ? (
            <Link href={href} aria-label={label} className={className} />
          ) : (
            <button
              type="button"
              aria-label={label}
              onClick={onClick}
              disabled={disabled}
              className={className}
            />
          )
        }
      >
        <Icon className="size-4" />
      </TooltipTrigger>
      <TooltipContent>{label}</TooltipContent>
    </Tooltip>
  );
}
