import assert from "node:assert/strict";
import { test } from "node:test";

import { transformFormSchema, yieldOf } from "./transform.ts";

const good = {
  stage: "PROCESS",
  inputs: [{ passport_id: "p-1", quantity: 1000 }, { passport_id: "p-2", quantity: 500 }],
  outputs: [{ quantity: 1200 }],
};

test("transform: accepts inputs and outputs with whole quantities", () => {
  const res = transformFormSchema.safeParse(good);
  assert.ok(res.success);
  assert.equal(res.data.inputs.length, 2);
});

test("transform: at least one input and one output, each a whole number above zero", () => {
  const bad = (change: object) => {
    const res = transformFormSchema.safeParse({ ...good, ...change });
    return res.success ? [] : res.error.issues;
  };
  assert.ok(bad({ inputs: [] }).length > 0);
  assert.ok(bad({ outputs: [] }).length > 0);
  assert.ok(bad({ stage: "  " }).length > 0);
  assert.ok(bad({ inputs: [{ passport_id: "p-1", quantity: 0 }] }).length > 0);
  assert.ok(bad({ inputs: [{ passport_id: "p-1", quantity: 1.5 }] }).length > 0);
});

test("yield: output over input, in percent", () => {
  assert.deepEqual(yieldOf([1000, 500], [1200], null), { totalIn: 1500, totalOut: 1200, pct: 80, max: null, over: false });
});

test("yield: over the template's maximum is flagged", () => {
  assert.equal(yieldOf([1000], [850], 80).over, true);
  assert.equal(yieldOf([1000], [800], 80).over, false);
  assert.equal(yieldOf([1000], [1100], undefined).over, false);
});
