import assert from "node:assert/strict";
import { test } from "node:test";

import {
  capacityRuleSchema,
  templateFormSchema,
} from "../lib/schemas/template.ts";

const good = {
  commodity: "coffee",
  name: "Washed Arabica — Export lot",
  unit: "kg",
  originFacilityKind: "washing_station",
  capacityRule: { type: "yield_per_area", kgPerHaPerSeason: 900 },
  requiredEvidence: "plot_registration, organic",
};

test("template: accepts a complete form", () => {
  const res = templateFormSchema.safeParse(good);
  assert.ok(res.success);
  assert.deepEqual(res.data.requiredEvidence, ["plot_registration", "organic"]);
});

test("template: each field is required and capped", () => {
  const bad = (change: object) => {
    const res = templateFormSchema.safeParse({ ...good, ...change });
    return res.success ? [] : res.error.issues.map((i) => i.path[0]);
  };
  assert.deepEqual(bad({ commodity: "  " }), ["commodity"]);
  assert.deepEqual(bad({ name: "" }), ["name"]);
  assert.deepEqual(bad({ unit: "" }), ["unit"]);
  assert.deepEqual(bad({ originFacilityKind: "Has Space" }), [
    "originFacilityKind",
  ]);
  assert.deepEqual(bad({ requiredEvidence: "Plot Photo" }), [
    "requiredEvidence",
  ]);
});

test("required evidence may be empty", () => {
  const res = templateFormSchema.safeParse({ ...good, requiredEvidence: " " });
  assert.ok(res.success);
  assert.deepEqual(res.data.requiredEvidence, []);
});

test("a per-area rule needs its kg-per-hectare figure", () => {
  assert.equal(
    capacityRuleSchema.safeParse({ type: "yield_per_area" }).success,
    false,
  );
  assert.ok(
    capacityRuleSchema.safeParse({ type: "facility_capacity" }).success,
  );
  assert.ok(
    capacityRuleSchema.safeParse({
      type: "yield_per_area",
      kgPerHaPerSeason: 900,
    }).success,
  );
});

