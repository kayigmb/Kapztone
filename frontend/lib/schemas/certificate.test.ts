import assert from "node:assert/strict";
import { test } from "node:test";

import { certificateBody, certificateFormSchema, revokeFormSchema } from "./certificate.ts";

const file = new File(["%PDF"], "organic.pdf", { type: "application/pdf" });
const good = {
  subject_type: "passport", subject_id: "ACME-2026-0412", type: "organic",
  valid_from: "2026-01-01", valid_to: "2026-12-31", file,
};

test("certificate: accepts a complete form", () => {
  const res = certificateFormSchema.safeParse(good);
  assert.ok(res.success);
});

test("certificate: the end date is after the start date, and the file is required", () => {
  const res = certificateFormSchema.safeParse({ ...good, valid_to: "2025-12-31" });
  assert.equal(res.success, false);
  assert.equal(certificateFormSchema.safeParse({ ...good, subject_id: "" }).success, false);
});

test("certificate: only a PDF, PNG or JPEG of at most 10 MB is accepted", () => {
  const badFile = new File(["x"], "scan.gif", { type: "image/gif" });
  const bigFile = new File([new ArrayBuffer(10 * 1024 * 1024 + 1)], "scan.pdf", { type: "application/pdf" });
  assert.equal(certificateFormSchema.safeParse({ ...good, file: badFile }).success, false);
  assert.equal(certificateFormSchema.safeParse({ ...good, file: bigFile }).success, false);
});

test("the multipart body carries the form's fields and the file itself", () => {
  const { file: f, ...rest } = good;
  const body = certificateBody(rest, f, "org-1");
  assert.equal(body.get("issuer_org_id"), "org-1");
  assert.equal(body.get("subject_id"), "ACME-2026-0412");
  assert.equal(body.get("file"), f);
});

test("revoking: a reason is required and capped", () => {
  assert.ok(revokeFormSchema.safeParse({ reason: "Duplicate of an earlier certificate" }).success);
  assert.equal(revokeFormSchema.safeParse({ reason: "  " }).success, false);
  assert.equal(revokeFormSchema.safeParse({ reason: "a".repeat(1001) }).success, false);
});
