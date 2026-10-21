import assert from "node:assert/strict";
import { test } from "node:test";

import {
  acceptSchema,
  inviteSchema,
  passwordChangeSchema,
  pinSchema,
} from "../lib/schemas/profile.ts";

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
  assert.equal(ok("old-password-1", "Newpass8!"), true);
  assert.equal(ok("old", "short"), false);
  assert.equal(ok("Strong-pass1!", "Strong-pass1!"), false);
  assert.equal(
    ok("old-password-1", "New-password-12!", "New-password-21!"),
    false,
  );
  assert.equal(ok("old-password-1", "lowercase-pass1!"), false);
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

test("invite accepts a strong password for new accounts but permits current passwords", () => {
  assert.equal(
    acceptSchema.safeParse({ name: "Nia", password: "Strong-pass1!" }).success,
    true,
  );
  assert.equal(
    acceptSchema.safeParse({ name: "Nia", password: "lowercase-pass1" })
      .success,
    false,
  );
  assert.equal(
    acceptSchema.safeParse({ name: "", password: "legacy-password" }).success,
    true,
  );
});
