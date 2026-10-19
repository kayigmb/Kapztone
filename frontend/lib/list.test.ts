import assert from "node:assert/strict";
import { test } from "node:test";

import { apiSort, listQuery, pageRange, tableSort } from "./list.ts";

test("the query string carries page, size, search, sort and only the filters that are set", () => {
  assert.equal(listQuery({ page: 2 }), "?page=2&size=10");
  assert.equal(
    listQuery({
      page: 1,
      size: 20,
      search: "  kigoma ",
      sort: "-created_at",
      filters: { kind: "farm_plot", status: "" },
    }),
    "?page=1&size=20&search=kigoma&sort=-created_at&kind=farm_plot",
  );
  assert.equal(listQuery({ page: 0 }), "?page=1&size=10");
});

test("the pager's range", () => {
  assert.deepEqual(pageRange(1, 10, 86), {
    from: 1,
    to: 10,
    pages: 9,
    prev: false,
    next: true,
  });
  assert.deepEqual(pageRange(9, 10, 86), {
    from: 81,
    to: 86,
    pages: 9,
    prev: true,
    next: false,
  });
  assert.deepEqual(pageRange(1, 10, 0), {
    from: 0,
    to: 0,
    pages: 1,
    prev: false,
    next: false,
  });
});

test("table sorting and the API's sort parameter map both ways", () => {
  assert.equal(
    apiSort([
      { id: "created_at", desc: true },
      { id: "name", desc: false },
    ]),
    "-created_at",
  );
  assert.equal(apiSort([{ id: "name", desc: false }]), "name");
  assert.equal(apiSort([]), "");
  assert.deepEqual(tableSort("-created_at"), [
    { id: "created_at", desc: true },
  ]);
  assert.deepEqual(tableSort(""), []);
});
