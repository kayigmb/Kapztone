import assert from "node:assert/strict";
import { test } from "node:test";

import {
  capacityLeft,
  issueBlock,
  issueFormSchema,
  templateStages,
} from "../lib/schemas/issue.ts";

const facility = { capacity: 12000, capacity_used: 9500 };

test("issue: accepts a whole quantity above zero", () => {
  const res = issueFormSchema.safeParse({
    facility_id: "f-1",
    template_id: "t-1",
    quantity: "2500",
  });
  assert.ok(res.success);
  assert.equal(res.data.quantity, 2500);
});

test("issue: each field is required, and the quantity is a whole number above zero", () => {
  const bad = (change: object) => {
    const res = issueFormSchema.safeParse({
      facility_id: "f-1",
      template_id: "t-1",
      quantity: 10,
      ...change,
    });
    return res.success ? [] : res.error.issues.map((i) => i.path[0]);
  };
  assert.deepEqual(bad({ facility_id: "" }), ["facility_id"]);
  assert.deepEqual(bad({ template_id: "" }), ["template_id"]);
  assert.deepEqual(bad({ quantity: 0 }), ["quantity"]);
  assert.deepEqual(bad({ quantity: -3 }), ["quantity"]);
  assert.deepEqual(bad({ quantity: 2.5 }), ["quantity"]);
});

test("capacity left is the remainder, never below zero", () => {
  assert.equal(capacityLeft(facility), 2500);
  assert.equal(capacityLeft({ capacity: 100, capacity_used: 150 }), 0);
});

test("issuing more than is left is blocked, with the numbers to say why", () => {
  assert.equal(issueBlock(facility, 2500), null);
  assert.deepEqual(issueBlock(facility, 2501), {
    over: true,
    left: 2500,
    capacity: 12000,
    capacityUsed: 9500,
  });
});

test("template stages come in the template's own order", () => {
  assert.deepEqual(
    templateStages({ stages: [{ key: "ISSUE" }, { key: "HANDOVER" }] }),
    ["ISSUE", "HANDOVER"],
  );
});
