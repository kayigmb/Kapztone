import assert from "node:assert/strict";
import { test } from "node:test";

import { fileProblem, maxUploadBytes, registrationSchema, suggestSlug } from "./org.ts";

const good = {
  user_name: "Jeanne Mukamana", email: "Jeanne@Rukundo.Test ", password: "correct-password",
  consent: true, name: "Rukundo Farmers' Cooperative", slug: "rukundo", kind: "producer",
};
const bad = (change: object) => {
  const res = registrationSchema.safeParse({ ...good, ...change });
  return res.success ? [] : res.error.issues.map((i) => i.path[0]);
};

test("registration: accepts a complete form and tidies the email", () => {
  const res = registrationSchema.safeParse(good);
  assert.ok(res.success);
  assert.equal(res.data.email, "jeanne@rukundo.test");
});

test("registration: each backend rule has a refused case, on the right field", () => {
  assert.deepEqual(bad({ user_name: "  " }), ["user_name"]);
  assert.deepEqual(bad({ email: "jeanne" }), ["email"]);
  assert.deepEqual(bad({ password: "short" }), ["password"]);
  assert.deepEqual(bad({ password: "x".repeat(73) }), ["password"]);
  assert.deepEqual(bad({ consent: false }), ["consent"]);
  assert.deepEqual(bad({ name: "" }), ["name"]);
  assert.deepEqual(bad({ name: "a".repeat(201) }), ["name"]);
  for (const slug of ["-acme", "acme-", "Acme Ltd", "a_b", "", "api", "www", "a".repeat(64)]) {
    assert.deepEqual(bad({ slug }), ["slug"], `slug ${JSON.stringify(slug)}`);
  }
  assert.deepEqual(bad({ kind: "bank" }), ["kind"]);
});

test("suggested address is always one the form accepts", () => {
  assert.equal(suggestSlug("Kitabi Washing Station Ltd."), "kitabi-washing-station-ltd");
  assert.equal(suggestSlug("  Société Minière — Rutongo  "), "societe-miniere-rutongo");
  assert.deepEqual(bad({ slug: suggestSlug("A".repeat(70) + " !!") }), []);
});

test("documents: PDF, PNG or JPEG, 10 MB at most", () => {
  assert.equal(fileProblem({ size: 1000, type: "application/pdf" }), undefined);
  assert.equal(fileProblem({ size: maxUploadBytes, type: "image/jpeg" }), undefined);
  assert.match(fileProblem({ size: maxUploadBytes + 1, type: "image/png" })!, /10 MB/);
  assert.match(fileProblem({ size: 10, type: "application/x-msdownload" })!, /PDF, PNG or JPEG/);
});
