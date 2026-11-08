import assert from "node:assert/strict";
import { test } from "vitest";

import {
  facilityFormSchema,
  type FacilityForm,
} from "../lib/schemas/facility.ts";

const nextYear = `${new Date().getFullYear() + 1}-12-31`;
const good: FacilityForm = {
  name: "Kitabi Washing Station",
  kind: "washing_station",
  licence_valid_to: nextYear,
  location: { type: "Point", coordinates: [29.37, -2.95] },
  licence: new File(["%PDF-1.4"], "licence.pdf", { type: "application/pdf" }),
};

test("registration: accepts a pin or a closed boundary", () => {
  assert.ok(facilityFormSchema.safeParse(good).success);
  const area = {
    type: "Polygon",
    coordinates: [
      [
        [29, -2],
        [29.1, -2],
        [29.1, -2.1],
        [29, -2],
      ],
    ],
  };
  assert.ok(facilityFormSchema.safeParse({ ...good, location: area }).success);
});

test("registration: each backend rule has a refused case, on the right field", () => {
  const bad = (change: object) => {
    const res = facilityFormSchema.safeParse({ ...good, ...change });
    return res.success ? [] : res.error.issues.map((i) => i.path[0]);
  };
  assert.deepEqual(bad({ name: "  " }), ["name"]);
  assert.deepEqual(bad({ name: "a".repeat(201) }), ["name"]);
  assert.deepEqual(bad({ kind: "has space" }), ["kind"]);
  assert.deepEqual(bad({ kind: "inspection_office" }), []); // a type of one's own
  assert.deepEqual(bad({ licence: undefined }), ["licence"]); // the document is the proof
  assert.deepEqual(
    bad({
      licence: new File(["x"], "a.exe", { type: "application/x-msdownload" }),
    }),
    ["licence"],
  );
  assert.deepEqual(bad({ licence_valid_to: "2000-01-01" }), [
    "licence_valid_to",
  ]);
  assert.deepEqual(bad({ location: undefined }), ["location"]);
  assert.deepEqual(
    bad({ location: { type: "Point", coordinates: [181, 0] } }),
    ["location"],
  );
  assert.deepEqual(
    bad({
      location: {
        type: "Polygon",
        coordinates: [
          [
            [29, -2],
            [29.1, -2],
            [29, -2],
          ],
        ],
      },
    }),
    ["location"],
  );
});
