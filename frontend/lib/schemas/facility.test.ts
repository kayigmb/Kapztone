import assert from "node:assert/strict";
import { test } from "node:test";

import { facilityFormSchema, type FacilityForm } from "./facility.ts";

const nextYear = `${new Date().getFullYear() + 1}-12-31`;
const good: FacilityForm = {
  name: "Kitabi Washing Station", kind: "washing_station", licence_number: "RBS-2024-118", licence_valid_to: nextYear,
  capacity: 12000, capacity_unit: "kg", capacity_period: "season", location: { type: "Point", coordinates: [29.37, -2.95] },
};

test("registration: accepts a pin or a closed boundary", () => {
  assert.ok(facilityFormSchema.safeParse(good).success);
  const area = { type: "Polygon", coordinates: [[[29, -2], [29.1, -2], [29.1, -2.1], [29, -2]]] };
  assert.ok(facilityFormSchema.safeParse({ ...good, location: area }).success);
});

test("registration: each backend rule has a refused case, on the right field", () => {
  const bad = (change: object) => {
    const res = facilityFormSchema.safeParse({ ...good, ...change });
    return res.success ? [] : res.error.issues.map((i) => i.path[0]);
  };
  assert.deepEqual(bad({ name: "  " }), ["name"]);
  assert.deepEqual(bad({ name: "a".repeat(201) }), ["name"]);
  assert.deepEqual(bad({ kind: "mine" }), ["kind"]);
  assert.deepEqual(bad({ licence_number: "" }), ["licence_number"]);
  assert.deepEqual(bad({ licence_valid_to: "2000-01-01" }), ["licence_valid_to"]);
  assert.deepEqual(bad({ capacity: 0 }), ["capacity"]);
  assert.deepEqual(bad({ capacity: 12.5 }), ["capacity"]);
  assert.deepEqual(bad({ capacity_unit: "" }), ["capacity_unit"]);
  assert.deepEqual(bad({ capacity_period: "week" }), ["capacity_period"]);
  assert.deepEqual(bad({ location: undefined }), ["location"]);
  assert.deepEqual(bad({ location: { type: "Point", coordinates: [181, 0] } }), ["location"]);
  assert.deepEqual(bad({ location: { type: "Polygon", coordinates: [[[29, -2], [29.1, -2], [29, -2]]] } }), ["location"]);
});

test("capacity typed as text coerces to a number", () => {
  const res = facilityFormSchema.safeParse({ ...good, capacity: "400" });
  assert.ok(res.success);
  assert.equal(res.data.capacity, 400);
});
