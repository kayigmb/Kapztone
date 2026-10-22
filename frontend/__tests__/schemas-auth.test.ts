import assert from "node:assert/strict";
import { test } from "node:test";

import { resetSchema, signInSchema } from "../lib/schemas/auth.ts";

test("sign-in: accepts what the API accepts, tidied the way the API tidies it", () => {
  const ok = signInSchema.safeParse({
    email: "  Ada@Acme.Test ",
    password: "x",
    code: "",
  });
  assert.ok(ok.success);
  assert.equal(ok.data.email, "ada@acme.test"); // the API trims and lowercases too
  assert.ok(
    signInSchema.safeParse({
      email: "ada@acme.test",
      password: "correct-password",
      code: " 123456 ",
    }).success,
  );
});

test("sign-in: refuses a missing email, password or a code that is not 6 digits, and says what to enter", () => {
  for (const [field, input] of Object.entries({
    email: { email: "ada", password: "x", code: "" },
    password: { email: "ada@acme.test", password: "", code: "" },
    code: { email: "ada@acme.test", password: "x", code: "12345" },
  })) {
    const res = signInSchema.safeParse(input);
    assert.ok(!res.success, field);
    assert.deepEqual(
      res.error.issues.map((i) => i.path[0]),
      [field],
    );
    assert.match(res.error.issues[0].message, /^Enter /);
  }
});

test("a reset password meets the strength policy and is typed the same twice", () => {
  assert.equal(
    resetSchema.safeParse({
      new: "Harbour-lantern-41!",
      confirm: "Harbour-lantern-41!",
    }).success,
    true,
  );
  assert.equal(
    resetSchema.safeParse({ new: "short", confirm: "short" }).success,
    false,
  );
  const overBcryptLimit = `Aa1!${"😀".repeat(18)}`;
  assert.equal(
    resetSchema.safeParse({
      new: overBcryptLimit,
      confirm: overBcryptLimit,
    }).success,
    false,
  );
  const differ = resetSchema.safeParse({
    new: "Harbour-lantern-41!",
    confirm: "Harbour-lantern-42!",
  });
  assert.equal(differ.success, false);
  assert.deepEqual(differ.error?.issues[0].path, ["confirm"]);
});
