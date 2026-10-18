import assert from "node:assert/strict";
import { test } from "node:test";

import { inviteSchema, passwordChangeSchema, pinSchema } from "./profile.ts";

test("a signing PIN is exactly 6 digits and needs the password", () => {
  assert.equal(pinSchema.safeParse({ pin: "123456", password: "x" }).success, true);
  assert.equal(pinSchema.safeParse({ pin: "12345", password: "x" }).success, false);
  assert.equal(pinSchema.safeParse({ pin: "12345a", password: "x" }).success, false);
  assert.equal(pinSchema.safeParse({ pin: "123456", password: "" }).success, false);
});

test("a new password is 12 to 72 characters and differs from the current one", () => {
  assert.equal(passwordChangeSchema.safeParse({ current: "old-password-1", new: "new-password-12" }).success, true);
  assert.equal(passwordChangeSchema.safeParse({ current: "old", new: "short" }).success, false);
  assert.equal(passwordChangeSchema.safeParse({ current: "same-password-1", new: "same-password-1" }).success, false);
});

test("an invitation needs a work email and a known role", () => {
  assert.equal(inviteSchema.parse({ email: " Ana@Kitabi.Test ", role: "member" }).email, "ana@kitabi.test");
  assert.equal(inviteSchema.safeParse({ email: "ana@kitabi.test", role: "superadmin" }).success, false);
});
