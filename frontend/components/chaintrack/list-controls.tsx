"use client";

import { ChevronLeft, ChevronRight, Search } from "lucide-react";
import { useEffect, useState } from "react";

import { SelectInput } from "@/components/chaintrack/select-input";
import { Input } from "@/components/ui/input";
import { copy } from "@/lib/consts";
import { pageRange, pageSize, pageSizes } from "@/lib/list";
import type { ListParamsState } from "@/lib/list-params";
import { cn } from "@/lib/utils";

const t = copy.list;

export type Option = { value: string; label: string };

/** Search box (sent to the API 300 ms after typing stops), sort select and any filters, in one bar. */
export function ListToolbar({
  search,
  onSearch,
  sorts,
  sort,
  onSort,
  children,
  className,
}: {
  search: string;
  onSearch: (q: string) => void;
  sorts?: Option[];
  sort?: string;
  onSort?: (s: string) => void;
  children?: React.ReactNode;
  className?: string;
}) {
  const [typed, setTyped] = useState(search);
  useEffect(() => {
    if (typed === search) return;
    const id = setTimeout(() => onSearch(typed), 300);
    return () => clearTimeout(id);
  }, [typed]); // eslint-disable-line react-hooks/exhaustive-deps -- only typing starts a search

  return (
    <div className={cn("flex flex-wrap items-center gap-2", className)}>
      <label className="relative min-w-48 flex-1 sm:max-w-80">
        <span className="sr-only">{t.search}</span>
        <Search
          className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-ink-subtle"
          aria-hidden
        />
        <Input
          type="search"
          value={typed}
          onChange={(e) => setTyped(e.target.value)}
          placeholder={t.search}
          className="pl-9"
        />
      </label>
      {children}
      {sorts && onSort && (
        <Choice
          label={t.sort}
          options={sorts}
          value={sort ?? ""}
          onChange={onSort}
        />
      )}
    </div>
  );
}

/** A small labelled select for a sort or a filter. */
export function Choice({
  label,
  options,
  value,
  onChange,
}: {
  label: string;
  options: Option[];
  value: string;
  onChange: (v: string) => void;
}) {
  return (
    <div className="flex items-center gap-2 text-sm text-ink-muted">
      <span className="hidden sm:inline" aria-hidden>
        {label}
      </span>
      <SelectInput
        label={label}
        options={options}
        value={value}
        onChange={onChange}
        className="w-auto"
      />
    </div>
  );
}

/** "1–10 of 86", rows per page, previous / next. */
export function Pager({
  page,
  total,
  size = pageSize,
  onPage,
  onSize,
}: {
  page: number;
  total: number;
  size?: number;
  onPage: (p: number) => void;
  onSize?: (s: number) => void;
}) {
  const r = pageRange(page, size, total);
  if (total <= pageSizes[0] && page === 1) return null;
  const arrow =
    "grid size-9 place-items-center rounded-md border border-border text-ink-muted hover:bg-surface-sunken hover:text-ink disabled:opacity-40 disabled:hover:bg-transparent";
  return (
    <nav
      aria-label={t.page(page, r.pages)}
      className="flex flex-wrap items-center justify-between gap-3 text-sm text-ink-muted"
    >
      <span className="font-mono text-xs">{t.range(r.from, r.to, total)}</span>
      <span className="flex items-center gap-2">
        {onSize && (
          <Choice
            label={t.perPage}
            value={String(size)}
            onChange={(v) => onSize(Number(v))}
            options={pageSizes.map((n) => ({
              value: String(n),
              label: String(n),
            }))}
          />
        )}
        <span className="hidden text-xs sm:inline">
          {t.page(page, r.pages)}
        </span>
        <button
          type="button"
          className={arrow}
          disabled={!r.prev}
          onClick={() => onPage(page - 1)}
          aria-label={t.prev}
        >
          <ChevronLeft className="size-4" aria-hidden />
        </button>
        <button
          type="button"
          className={arrow}
          disabled={!r.next}
          onClick={() => onPage(page + 1)}
          aria-label={t.next}
        >
          <ChevronRight className="size-4" aria-hidden />
        </button>
      </span>
    </nav>
  );
}

/** A Pager wired to a list's URL state. */
export function ListPager({
  list,
  total,
}: {
  list: ListParamsState;
  total: number;
}) {
  return (
    <Pager
      page={list.page}
      size={list.size}
      total={total}
      onPage={list.setPage}
      onSize={list.setSize}
    />
  );
}

/** A ListToolbar wired to a list's URL state; filters go in as children. */
export function ListBar({
  list,
  sorts,
  children,
  className,
}: {
  list: ListParamsState;
  sorts?: Option[];
  children?: React.ReactNode;
  className?: string;
}) {
  return (
    <ListToolbar
      search={list.search}
      onSearch={list.setSearch}
      sorts={sorts}
      sort={list.sort}
      onSort={list.setSort}
      className={className}
    >
      {children}
    </ListToolbar>
  );
}
