import assert from "node:assert/strict";
import { test } from "node:test";

import { toView, type ApiVerification } from "../lib/verification.ts";

const now = new Date("2026-10-12T08:00:00Z");
const passport = {
  id: "11111111-1111-1111-1111-111111111111",
  organisation_id: "org-mill",
  facility_id: null,
  transform_id: "t1",
  commodity: "coffee",
  unit: "kg",
  quantity: 1200,
  remaining: 1200,
  status: "active",
  data_hash: "0xabc",
  tx_hash: "0xdef",
  created_at: "2026-10-01T07:00:00Z",
  events: [
    {
      id: "e1",
      kind: "TRANSFORM",
      from_org: null,
      to_org: "org-mill",
      quantity: 1200,
      tx_hash: "0x1",
      created_at: "2026-10-01T07:00:00Z",
    },
    {
      id: "e2",
      kind: "HANDOVER",
      from_org: "org-mill",
      to_org: "org-export",
      quantity: 1200,
      tx_hash: null,
      created_at: "2026-10-02T07:14:00Z",
    },
  ],
};
const certificate = (
  status: "valid" | "expired" | "revoked",
  validTo: string,
) => ({
  id: "c-" + status + validTo,
  issuer_org_id: "org-lab",
  subject_id: passport.id,
  type: "export_permit",
  valid_from: "2026-01-01T00:00:00Z",
  valid_to: validTo,
  doc_hash: "0xfeed",
  status,
  revoke_reason: status === "revoked" ? "Issued in error" : null,
  tx_hash: null,
});
const base: ApiVerification = {
  verdict: "verified",
  reasons: [],
  checks: [
    "record_intact",
    "anchored",
    "origin",
    "capacity",
    "mass_balance",
    "certificates",
    "status",
  ].map((name) => ({ name, result: "verified" as const })),
  passport,
  origins: [
    {
      facility: {
        id: "f1",
        name: "North plot",
        kind: "farm_plot",
        location: { type: "Point", coordinates: [29.37, -2.95] },
      },
      organisation: { id: "org-farm", name: "Rukundo Cooperative" },
    },
  ],
  organisations: [
    { id: "org-mill", name: "Kitabi Washing Station" },
    { id: "org-export", name: "Akagera Exports" },
    { id: "org-lab", name: "Rwanda Standards Board" },
  ],
  genealogy: {
    nodes: [
      passport,
      {
        ...passport,
        id: "origin-lot",
        facility_id: "f1",
        transform_id: null,
        quantity: 6000,
        events: undefined,
      },
    ],
    edges: [],
  },
  certificates: [],
};

test("a clean lot: four passing reasons, names and formats", () => {
  const v = toView(base, now);
  assert.equal(v.reasons.length, 4);
  assert.ok(v.reasons.every((r) => r.result === "pass"));
  assert.match(v.summary, /North plot/);
  assert.equal(
    v.product.facts.find((f) => f.label === "Quantity")?.value,
    "1,200 kg",
  );
  assert.equal(
    v.product.facts.find((f) => f.label === "Current holder")?.value,
    "Kitabi Washing Station",
  );
  assert.deepEqual(
    v.journey.map((s) => [s.title, s.state, s.signedByBoth]),
    [
      ["Processed", "done", false],
      ["Handed over", "current", true],
    ],
  );
  assert.equal(v.journey[1].date, "2 Oct 2026, 09:14 CAT");
  assert.equal(v.journey[1].parties.to, "Akagera Exports");
  assert.deepEqual(
    v.madeFrom.map((m) => [m.lotCode, m.origin, m.quantity]),
    [["origin-lot", true, "6,000 kg"]],
  );
  assert.equal(
    v.meta,
    "Checked 12 Oct 2026, 10:00 CAT · 2 records · 3 organisations",
  );
  assert.deepEqual(v.origins[0].geometry, {
    type: "Point",
    coordinates: [29.37, -2.95],
  });
});

test("failures come first, repeats are said once, never more than four", () => {
  const warn = {
    level: "warning" as const,
    check: "certificates",
    code: "certificate_expired",
    message: "A certificate has expired.",
    subject_id: "c1",
  };
  const fail = {
    level: "failed" as const,
    check: "record_intact",
    code: "record_altered",
    message: "This passport's record was changed after it was written.",
    subject_id: "p1",
  };
  const v = toView(
    {
      ...base,
      verdict: "failed",
      reasons: [warn, warn, fail],
      checks: base.checks.map((c) =>
        c.name === "record_intact"
          ? { ...c, result: "failed" as const }
          : c.name === "certificates"
            ? { ...c, result: "warning" as const }
            : c,
      ),
    },
    now,
  );
  assert.deepEqual(
    v.reasons.map((r) => r.result),
    ["fail", "warn", "pass", "pass"],
  );
  assert.equal(
    v.checks.find((c) => c.name === "certificates")?.explanation,
    "A certificate has expired.",
  );
  assert.match(
    v.checks.find((c) => c.name === "capacity")!.explanation,
    /declared capacity/,
  );
});

test("certificate states: revoked is red with its reason, expiring within 30 days, expired says what to do", () => {
  const v = toView(
    {
      ...base,
      certificates: [
        certificate("revoked", "2030-01-01T00:00:00Z"),
        certificate("valid", "2026-10-30T00:00:00Z"),
        certificate("valid", "2027-06-01T00:00:00Z"),
        certificate("expired", "2026-09-08T00:00:00Z"),
      ],
    },
    now,
  );
  assert.deepEqual(
    v.certificates.map((c) => c.state),
    ["revoked", "expiring", "valid", "expired"],
  );
  assert.match(v.certificates[0].note!, /Issued in error/);
  assert.match(v.certificates[3].note!, /Expired 8 Sep 2026/);
  assert.equal(v.certificates[0].issuer, "Rwanda Standards Board");
  assert.equal(v.certificates[0].type, "Export permit");
});

test("the lot code is shown when the API gives it, and each earlier lot says how it went in", () => {
  const api: ApiVerification = {
    ...base,
    passport: { ...passport, code: "MILL-2026-0007" },
    genealogy: {
      nodes: [
        passport,
        {
          ...passport,
          id: "origin-lot",
          code: "FARM-2026-0001",
          facility_id: "f1",
          transform_id: null,
          quantity: 6000,
          events: undefined,
        },
      ],
      edges: [
        {
          parent_id: "origin-lot",
          child_id: passport.id,
          stage: "washing_station",
          quantity_used: 6000,
        },
      ],
    },
  };
  const v = toView(api, now);
  assert.equal(v.product.lotCode, "MILL-2026-0007");
  assert.equal(v.product.holder, "Kitabi Washing Station");
  assert.deepEqual(
    v.madeFrom.map((m) => [m.lotCode, m.via]),
    [["FARM-2026-0001", "Washing station · 6,000 kg used"]],
  );
  assert.equal(v.records, "2 records");
});
