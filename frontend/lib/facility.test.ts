import assert from "node:assert/strict";
import { test } from "node:test";

import { facilityState } from "./facility.ts";

test("a facility's state: approval and suspension first, then the licence date", () => {
  const today = "2026-10-18";
  assert.equal(
    facilityState(
      { status: "pending_approval", licence_valid_to: "2020-01-01" },
      today,
    ),
    "pending",
  );
  assert.equal(facilityState({ status: "suspended" }, today), "suspended");
  assert.equal(
    facilityState({ status: "active", licence_valid_to: "2026-10-17" }, today),
    "expired",
  );
  assert.equal(
    facilityState({ status: "active", licence_valid_to: "2026-10-18" }, today),
    "active",
  );
  assert.equal(
    facilityState({ status: "active", licence_valid_to: null }, today),
    "active",
  );
});
