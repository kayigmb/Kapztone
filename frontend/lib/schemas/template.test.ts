import assert from "node:assert/strict";
import { test } from "node:test";

import { capacityRuleSchema, stageFormSchema, templateFormSchema } from "./template.ts";

const good = {
  commodity: "coffee",
  name: "Washed Arabica — Export lot",
  unit: "kg",
  originFacilityKind: "washing_station",
  capacityRule: { type: "yield_per_area", kgPerHaPerSeason: 900 },
  stages: [
    { key: "ISSUE", events: ["ISSUE"], maxYield: undefined, requiredEvidence: ["organic"] },
    { key: "HANDOVER", events: ["HANDOVER"], maxYield: undefined, requiredEvidence: [] },
  ],
};

test("template: accepts a complete form", () => {
  const res = templateFormSchema.safeParse(good);
  assert.ok(res.success);
});

test("template: each field is required and capped", () => {
  const bad = (change: object) => {
    const res = templateFormSchema.safeParse({ ...good, ...change });
    return res.success ? [] : res.error.issues.map((i) => i.path[0]);
  };
  assert.deepEqual(bad({ commodity: "  " }), ["commodity"]);
  assert.deepEqual(bad({ name: "" }), ["name"]);
  assert.deepEqual(bad({ unit: "" }), ["unit"]);
  assert.deepEqual(bad({ originFacilityKind: "factory" }), ["originFacilityKind"]);
  assert.deepEqual(bad({ stages: [] }), ["stages"]);
});

test("a stage needs at least one event, and a name", () => {
  assert.equal(stageFormSchema.safeParse({ key: "", events: [], requiredEvidence: [] }).success, false);
  assert.equal(stageFormSchema.safeParse({ key: "ISSUE", events: [], requiredEvidence: [] }).success, false);
  assert.ok(stageFormSchema.safeParse({ key: "ISSUE", events: ["ISSUE"], requiredEvidence: ["organic"] }).success);
});

test("a per-area rule needs its kg-per-hectare figure", () => {
  assert.equal(capacityRuleSchema.safeParse({ type: "yield_per_area" }).success, false);
  assert.ok(capacityRuleSchema.safeParse({ type: "facility_capacity" }).success);
  assert.ok(capacityRuleSchema.safeParse({ type: "yield_per_area", kgPerHaPerSeason: 900 }).success);
});
