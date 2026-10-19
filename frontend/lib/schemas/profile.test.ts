import assert from "node:assert/strict";
import { test } from "node:test";

import { inviteSchema, passwordChangeSchema, pinSchema } from "./profile.ts";

test("a signing PIN is exactly 6 digits and needs the password", () => {
  assert.equal(
    pinSchema.safeParse({ pin: "123456", password: "x" }).success,
    true,
  );
  assert.equal(
    pinSchema.safeParse({ pin: "12345", password: "x" }).success,
    false,
  );
  assert.equal(
    pinSchema.safeParse({ pin: "12345a", password: "x" }).success,
    false,
  );
  assert.equal(
    pinSchema.safeParse({ pin: "123456", password: "" }).success,
    false,
  );
});

test("a new password is 8 to 72 characters, typed twice, and differs from the current one", () => {
  const ok = (current: string, next: string, confirm = next) =>
    passwordChangeSchema.safeParse({ current, new: next, confirm }).success;
  assert.equal(ok("old-password-1", "newpass8"), true);
  assert.equal(ok("old", "short"), false);
  assert.equal(ok("same-password-1", "same-password-1"), false);
  assert.equal(
    ok("old-password-1", "new-password-12", "new-password-21"),
    false,
  );
});

test("an invitation needs a Email and a known role", () => {
  assert.equal(
    inviteSchema.parse({ email: " Ana@Kitabi.Test ", role: "member" }).email,
    "ana@kitabi.test",
  );
  assert.equal(
    inviteSchema.safeParse({ email: "ana@kitabi.test", role: "superadmin" })
      .success,
    false,
  );
});
