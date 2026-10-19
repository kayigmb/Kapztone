import assert from "node:assert/strict";
import { test } from "node:test";

import { raiseFlagFormSchema, resolveFlagFormSchema } from "./flag.ts";

test("raising a flag: a lot and a reason, both required", () => {
  assert.ok(
    raiseFlagFormSchema.safeParse({
      passport_id: "ACME-2026-0412",
      reason: "Seals were broken",
    }).success,
  );
  assert.equal(
    raiseFlagFormSchema.safeParse({ passport_id: "", reason: "x" }).success,
    false,
  );
  assert.equal(
    raiseFlagFormSchema.safeParse({
      passport_id: "ACME-2026-0412",
      reason: "  ",
    }).success,
    false,
  );
  assert.equal(
    raiseFlagFormSchema.safeParse({
      passport_id: "x",
      reason: "a".repeat(1001),
    }).success,
    false,
  );
});

test("resolving a flag: an outcome and a note are both required", () => {
  for (const outcome of ["cleared", "confirmed"]) {
    assert.ok(
      resolveFlagFormSchema.safeParse({
        outcome,
        note: "Checked against the manifest",
      }).success,
      outcome,
    );
  }
  assert.equal(
    resolveFlagFormSchema.safeParse({ outcome: "closed", note: "x" }).success,
    false,
  );
  assert.equal(
    resolveFlagFormSchema.safeParse({ outcome: "cleared", note: "" }).success,
    false,
  );
  assert.equal(
    resolveFlagFormSchema.safeParse({ outcome: "cleared", note: "  " }).success,
    false,
  );
});
