import assert from "node:assert/strict";
import { test } from "node:test";

import { upload } from "../lib/consts.ts";
import {
  accreditationSchema,
  fileProblem,
  registrationSchema,
  reviewSchema,
  suggestSlug,
  toAccreditations,
} from "../lib/schemas/org.ts";

const maxUploadBytes = upload.maxBytes;

const good = {
  user_name: "Jeanne Mukamana",
  email: "Jeanne@Rukundo.Test ",
  password: "correct-password",
  confirm_password: "correct-password",
  consent: true,
  name: "Rukundo Farmers' Cooperative",
  slug: "rukundo",
  kind: "producer",
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
  assert.deepEqual(bad({ password: "short", confirm_password: "short" }), [
    "password",
  ]);
  assert.deepEqual(bad({ confirm_password: "another-password" }), [
    "confirm_password",
  ]);
  assert.deepEqual(
    bad({ password: "x".repeat(73), confirm_password: "x".repeat(73) }),
    ["password"],
  );
  assert.deepEqual(bad({ consent: false }), ["consent"]);
  assert.deepEqual(bad({ name: "" }), ["name"]);
  assert.deepEqual(bad({ name: "a".repeat(201) }), ["name"]);
  for (const slug of [
    "-acme",
    "acme-",
    "Acme Ltd",
    "a_b",
    "",
    "api",
    "www",
    "a".repeat(64),
  ]) {
    assert.deepEqual(bad({ slug }), ["slug"], `slug ${JSON.stringify(slug)}`);
  }
  assert.deepEqual(bad({ kind: "bank" }), ["kind"]);
});

test("suggested address is always one the form accepts", () => {
  assert.equal(
    suggestSlug("Kitabi Washing Station Ltd."),
    "kitabi-washing-station-ltd",
  );
  assert.equal(
    suggestSlug("  Société Minière — Rutongo  "),
    "societe-miniere-rutongo",
  );
  assert.deepEqual(bad({ slug: suggestSlug("A".repeat(70) + " !!") }), []);
});

test("documents: PDF, PNG or JPEG, 10 MB at most", () => {
  assert.equal(fileProblem({ size: 1000, type: "application/pdf" }), undefined);
  assert.equal(
    fileProblem({ size: maxUploadBytes, type: "image/jpeg" }),
    undefined,
  );
  assert.match(
    fileProblem({ size: maxUploadBytes + 1, type: "image/png" })!,
    /10 MB/,
  );
  assert.match(
    fileProblem({ size: 10, type: "application/x-msdownload" })!,
    /PDF, PNG or JPEG/,
  );
});

test("review: a reason is required to reject or ask for changes, not to approve", () => {
  assert.ok(
    reviewSchema.safeParse({ decision: "verified", reason: "" }).success,
  );
  for (const decision of ["rejected", "changes_requested"]) {
    const res = reviewSchema.safeParse({ decision, reason: "   " });
    assert.ok(!res.success, decision);
    assert.deepEqual(
      res.error.issues.map((i) => i.path[0]),
      ["reason"],
    );
    assert.ok(
      reviewSchema.safeParse({ decision, reason: "Licence is unreadable" })
        .success,
    );
  }
  assert.ok(
    !reviewSchema.safeParse({ decision: "pending", reason: "x" }).success,
  );
});

test("accreditation: at least one type, dates in order, sent as one window per type", () => {
  const good = {
    types: ["assay_report", "export_permit"],
    valid_from: "2026-11-01",
    valid_to: "2027-10-31",
  };
  const res = accreditationSchema.safeParse(good);
  assert.ok(res.success);
  assert.deepEqual(toAccreditations(res.data)[1], {
    type: "export_permit",
    valid_from: "2026-11-01T00:00:00Z",
    valid_to: "2027-10-31T23:59:59Z",
  });
  const bad = (change: object) => {
    const r = accreditationSchema.safeParse({ ...good, ...change });
    return r.success ? [] : r.error.issues.map((i) => i.path[0]);
  };
  assert.deepEqual(bad({ types: [] }), ["types"]);
  assert.deepEqual(bad({ valid_to: "2026-11-01" }), ["valid_to"]);
  assert.deepEqual(bad({ valid_from: "" }), ["valid_from"]);
});

test("registration: a mismatched password shows on step 1, while the organisation fields are still empty", () => {
  const res = registrationSchema.safeParse({
    ...good,
    name: "",
    slug: "",
    kind: undefined,
    confirm_password: "another-password",
  });
  assert.ok(!res.success);
  assert.ok(res.error.issues.some((i) => i.path[0] === "confirm_password"));
});
