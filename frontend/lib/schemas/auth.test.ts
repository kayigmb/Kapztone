// Run: node --test "lib/**/*.test.ts"
import assert from "node:assert/strict";
import { test } from "node:test";

import { signInSchema } from "./auth.ts";

test("sign-in: accepts what the API accepts, tidied the way the API tidies it", () => {
  const ok = signInSchema.safeParse({ email: "  Ada@Acme.Test ", password: "x", code: "" });
  assert.ok(ok.success);
  assert.equal(ok.data.email, "ada@acme.test"); // the API trims and lowercases too
  assert.ok(signInSchema.safeParse({ email: "ada@acme.test", password: "correct-password", code: " 123456 " }).success);
});

test("sign-in: refuses a missing email, password or a code that is not 6 digits, and says what to enter", () => {
  for (const [field, input] of Object.entries({
    email: { email: "ada", password: "x", code: "" },
    password: { email: "ada@acme.test", password: "", code: "" },
    code: { email: "ada@acme.test", password: "x", code: "12345" },
  })) {
    const res = signInSchema.safeParse(input);
    assert.ok(!res.success, field);
    assert.deepEqual(res.error.issues.map((i) => i.path[0]), [field]);
    assert.match(res.error.issues[0].message, /^Enter /);
  }
});
