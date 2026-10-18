import assert from "node:assert/strict";
import { test } from "node:test";

import { handoverFormSchema, handoverQuantity, receiveFormSchema, rejectFormSchema } from "./handover.ts";

const base = { stage: "export", note: "" };

test("handover: a whole-lot handover needs no quantity; a partial one does; a stage is required", () => {
  assert.ok(handoverFormSchema.safeParse({ ...base, all: true }).success);
  assert.ok(handoverFormSchema.safeParse({ ...base, all: false, quantity: 100 }).success);
  assert.equal(handoverFormSchema.safeParse({ ...base, all: false, quantity: undefined }).success, false);
  assert.equal(handoverFormSchema.safeParse({ ...base, all: true, stage: "" }).success, false);
});

test("handover: the note may be long, but only up to the limit", () => {
  assert.ok(handoverFormSchema.safeParse({ ...base, all: true, note: "a".repeat(1000) }).success);
  assert.equal(handoverFormSchema.safeParse({ ...base, all: true, note: "a".repeat(1001) }).success, false);
});

test("the quantity sent is the whole remainder when `all` is ticked", () => {
  assert.equal(handoverQuantity({ ...base, all: true, quantity: undefined }, 50), 50);
  assert.equal(handoverQuantity({ ...base, all: false, quantity: 20 }, 50), 20);
});

test("receiving: the weighed quantity is a whole number above zero", () => {
  assert.ok(receiveFormSchema.safeParse({ received_quantity: "1476" }).success);
  assert.equal(receiveFormSchema.safeParse({ received_quantity: "0" }).success, false);
  assert.equal(receiveFormSchema.safeParse({ received_quantity: "14.5" }).success, false);
});

test("rejecting: at least one listed reason; the note is capped", () => {
  assert.ok(rejectFormSchema.safeParse({ reasons: ["weight_short"], note: "", also_flag: true }).success);
  assert.equal(rejectFormSchema.safeParse({ reasons: [], note: "", also_flag: false }).success, false);
  assert.equal(rejectFormSchema.safeParse({ reasons: ["lost"], note: "", also_flag: false }).success, false);
  assert.equal(rejectFormSchema.safeParse({ reasons: ["damaged"], note: "a".repeat(1001), also_flag: false }).success, false);
});
