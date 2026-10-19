// Lists are paged, searched and sorted by the API (listq). This builds the
// query string and the "1–10 of 86" range. Pure: `node --test` runs it.

/** Rows-per-page choices on every list; the first is the default. */
export const pageSizes = [10, 20, 50] as const;
export type PageSize = (typeof pageSizes)[number];
export const pageSize: PageSize = pageSizes[0];

export type ListParams = {
  page: number;
  size?: number;
  search?: string;
  sort?: string; // a column, "-" in front for newest / largest first
  filters?: Record<string, string | undefined>; // status=, kind=, box=…: empty ones are left out
};

export function listQuery(p: ListParams): string {
  const q = new URLSearchParams({
    page: String(Math.max(1, p.page)),
    size: String(p.size ?? pageSize),
  });
  if (p.search?.trim()) q.set("search", p.search.trim());
  if (p.sort) q.set("sort", p.sort);
  for (const [k, v] of Object.entries(p.filters ?? {})) if (v) q.set(k, v);
  return `?${q}`;
}

/** What the pager says, and whether there's a page either side. */
export function pageRange(page: number, size: number, total: number) {
  const pages = Math.max(1, Math.ceil(total / size));
  const from = total === 0 ? 0 : (page - 1) * size + 1;
  return {
    from,
    to: Math.min(page * size, total),
    pages,
    prev: page > 1,
    next: page < pages,
  };
}

/** A table's sorting (tablecn / TanStack: [{ id, desc }]) as the API's ?sort= ("-created_at"): the first column only. */
export function apiSort(
  sorting: readonly { id: string; desc: boolean }[],
): string {
  const first = sorting[0];
  return first ? `${first.desc ? "-" : ""}${first.id}` : "";
}

/** The API's ?sort= as a table's sorting. */
export function tableSort(sort: string): { id: string; desc: boolean }[] {
  return sort
    ? [{ id: sort.replace(/^-/, ""), desc: sort.startsWith("-") }]
    : [];
}
