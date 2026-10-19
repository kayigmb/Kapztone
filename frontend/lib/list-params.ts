"use client";

// A list's page, rows per page, search and sort, plus its own filters, typed
// and kept in the address (nuqs), so a reload or a shared link shows the same
// page. The keys are tablecn's (page, perPage, sort=name.asc), so card lists
// and tables (useDataTable) read and write the same address. Any change other
// than the page goes back to page 1.

import {
  parseAsInteger,
  parseAsNumberLiteral,
  parseAsString,
  useQueryStates,
  type ParserMap,
} from "nuqs";

import { getSortingStateParser } from "@/lib/data-table-parsers";
import {
  apiSort,
  pageSize,
  pageSizes,
  tableSort,
  type ListParams,
} from "@/lib/list";

/** The address keys every list uses (tablecn's for page, size and sort). */
export const listKeys = {
  page: "page",
  perPage: "perPage",
  sort: "sort",
  search: "q",
  view: "view",
} as const;

const base = {
  [listKeys.page]: parseAsInteger.withDefault(1),
  [listKeys.perPage]: parseAsNumberLiteral(pageSizes).withDefault(pageSize),
  [listKeys.search]: parseAsString.withDefault(""),
};

type Set = (changes: Record<string, unknown>) => Promise<URLSearchParams>;

export function useListParams<F extends ParserMap = Record<never, never>>(
  defaultSort: string,
  filters?: F,
) {
  const [state, rawSet] = useQueryStates(
    {
      ...base,
      [listKeys.sort]: getSortingStateParser().withDefault(
        tableSort(defaultSort),
      ),
      ...(filters ?? ({} as F)),
    },
    { history: "replace", scroll: false },
  );
  const set = rawSet as unknown as Set;
  const values = state as unknown as Record<string, unknown>;
  const page = values[listKeys.page] as number;
  const size = values[listKeys.perPage] as number;
  const search = values[listKeys.search] as string;
  /** The API's form of the sort ("-created_at"), whatever the address holds. */
  const sort =
    apiSort(values[listKeys.sort] as { id: string; desc: boolean }[]) ||
    defaultSort;

  type Filters = { [K in keyof F]: ReturnType<F[K]["parseServerSide"]> };
  const filterValues = Object.fromEntries(
    Object.keys(filters ?? {}).map((k) => [k, values[k]]),
  ) as Filters;

  return {
    page,
    size,
    search,
    sort,
    filters: filterValues,
    setPage: (p: number) => set({ [listKeys.page]: p > 1 ? p : null }),
    setSize: (s: number) =>
      set({ [listKeys.perPage]: s, [listKeys.page]: null }),
    setSearch: (q: string) =>
      set({ [listKeys.search]: q || null, [listKeys.page]: null }),
    setSort: (s: string) =>
      set({ [listKeys.sort]: tableSort(s), [listKeys.page]: null }),
    /** Changes some filters (null clears one); back to page 1 unless it's only a selection on this page. */
    setFilters: (
      changes: Partial<{ [K in keyof F]: Filters[K] | null }>,
      opts?: { keepPage?: boolean },
    ) =>
      set({ ...changes, ...(opts?.keepPage ? {} : { [listKeys.page]: null }) }),
    /** What the API takes (lib/list.ts → listQuery); the screen adds its filters. */
    query: (extra?: ListParams["filters"]): ListParams => ({
      page,
      size,
      search,
      sort,
      filters: extra,
    }),
  };
}

export type ListParamsState = ReturnType<typeof useListParams>;
