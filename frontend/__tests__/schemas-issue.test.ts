import assert from "node:assert/strict";
import { test } from "vitest";

import { issueFormSchema } from "../lib/schemas/issue.ts";

const good = {
  facility_id: "f-1",
  commodity: "Green coffee",
  unit: "kg",
  quantity: "2500",
};

test("issue: the producer names the product and its unit", () => {
  const res = issueFormSchema.safeParse(good);
  assert.ok(res.success);
  assert.equal(res.data.quantity, 2500);
  assert.equal(res.data.commodity, "Green coffee");
});

test("issue: each field is required, and the quantity is a whole number above zero", () => {
  const bad = (change: object) => {
    const res = issueFormSchema.safeParse({ ...good, ...change });
    return res.success ? [] : res.error.issues.map((i) => i.path[0]);
  };
  assert.deepEqual(bad({ facility_id: "" }), ["facility_id"]);
  assert.deepEqual(bad({ commodity: " " }), ["commodity"]);
  assert.deepEqual(bad({ commodity: "9 coffee" }), ["commodity"]); // the API's key starts with a letter
  assert.deepEqual(bad({ commodity: "café" }), ["commodity"]);
  assert.deepEqual(bad({ unit: "crates" }), []); // a unit of one's own
  assert.deepEqual(bad({ unit: " " }), ["unit"]);
  assert.deepEqual(bad({ unit: "x".repeat(17) }), ["unit"]);
  assert.deepEqual(bad({ quantity: 0 }), ["quantity"]);
  assert.deepEqual(bad({ quantity: -3 }), ["quantity"]);
  assert.deepEqual(bad({ quantity: 2.5 }), ["quantity"]);
});
