import assert from "node:assert/strict";
import { test } from "node:test";

import {
  defaultTolerancePct,
  handoverStages,
  isPartial,
  toleranceOf,
  weigh,
} from "../lib/handover.ts";

test("a weight inside the tolerance can be accepted; outside it can't", () => {
  assert.deepEqual(weigh(1480, 1476, 1), {
    difference: -4,
    pct: 0.3,
    within: true,
  });
  assert.deepEqual(weigh(640, 588, 1), {
    difference: -52,
    pct: 8.1,
    within: false,
  });
  assert.equal(weigh(1000, 1010, 1).within, true); // exactly at the limit
  assert.equal(weigh(1000, 1011, 1).within, false); // more than sent counts too
});

test("only stages allowing HANDOVER can receive a lot, with their own tolerance or the default", () => {
  const stages = [
    { key: "origin", events: ["ISSUE"] },
    { key: "export", events: ["HANDOVER", "EXPORT"], handoverTolerance: 0.5 },
    { key: "import", events: ["HANDOVER"] },
  ];
  assert.deepEqual(
    handoverStages(stages).map((s) => s.key),
    ["export", "import"],
  );
  assert.equal(toleranceOf(stages, "export"), 0.5);
  assert.equal(toleranceOf(stages, "import"), defaultTolerancePct);
});

test("sending less than the whole passport is a partial handover", () => {
  assert.equal(isPartial(400, 1000), true);
  assert.equal(isPartial(1000, 1000), false);
  assert.equal(isPartial(0, 1000), false);
});
