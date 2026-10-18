import assert from "node:assert/strict";
import { test } from "node:test";

import { centreOf, toGeometry } from "./geo.ts";

test("a pin is the last point clicked", () => {
  assert.equal(toGeometry("pin", []), undefined);
  assert.deepEqual(toGeometry("pin", [[30, -2], [30.1, -2.1]]), { type: "Point", coordinates: [30.1, -2.1] });
});

test("a boundary needs three corners and comes back closed", () => {
  assert.equal(toGeometry("area", [[30, -2], [30.1, -2]]), undefined);
  assert.deepEqual(toGeometry("area", [[30, -2], [30.1, -2], [30.1, -2.1]]), {
    type: "Polygon", coordinates: [[[30, -2], [30.1, -2], [30.1, -2.1], [30, -2]]],
  });
});

test("the centre of some places is the middle of their corners", () => {
  assert.deepEqual(centreOf([{ type: "Point", coordinates: [30, -2] }, { type: "Point", coordinates: [32, -4] }]), [31, -3]);
  assert.equal(centreOf([]), undefined);
});
