import assert from "node:assert/strict";
import { test } from "node:test";

import { fill } from "../lib/format.ts";

test("fill replaces every placeholder", () => {
  assert.equal(
    fill("Page {page} of {pages}", { page: 2, pages: 9 }),
    "Page 2 of 9",
  );
});

test("fill leaves unknown placeholders visible", () => {
  assert.equal(fill("Hello {name}", {}), "Hello {name}");
});

test("fill with no placeholders returns the string", () => {
  assert.equal(fill("Try again"), "Try again");
});
