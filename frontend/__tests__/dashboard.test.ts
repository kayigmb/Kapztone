import assert from "node:assert/strict";
import { test } from "vitest";

import { navGroups, permissions } from "../lib/consts.ts";
import {
  actionsFor,
  canOpen,
  dashboardFor,
  layoutFor,
  hasWorkspace,
  navFor,
  toneOf,
} from "../lib/dashboard.ts";

const producer = {
  organisation: { kind: "producer" },
  permissions: ["passport:issue"],
};
// The person reviewing organisations also runs users and the audit reports.
const admin = {
  organisation: null,
  permissions: [
    permissions.orgVerify,
    permissions.userManage,
    permissions.auditRead,
  ],
};

// What GET /me gives an organisation's admin: its kind's permissions (backend KindPermissions ∩ org_admin).
const everyKind = [
  "org:read",
  "org:manage",
  "org:search",
  "document:read",
  "user:manage",
  "passport:read",
  "flag:create",
  "facility:read",
  "audit:read",
  "report:read",
];
const regulator = {
  organisation: { kind: "regulator" },
  permissions: [
    ...everyKind,
    "certificate:issue",
    "certificate:revoke",
    "flag:resolve",
    "facility:manage",
  ],
};
const certifier = {
  organisation: { kind: "certifier" },
  permissions: [
    ...everyKind,
    "certificate:issue",
    "certificate:revoke",
    "facility:manage",
  ],
};
const producerAdmin = {
  organisation: { kind: "producer" },
  permissions: [
    ...everyKind,
    "facility:manage",
    "passport:issue",
    "passport:retire",
    "transform:create",
    "handover:create",
    "handover:accept",
  ],
};

test("each kind of organisation gets its dashboard; the platform admin gets the platform one", () => {
  assert.equal(dashboardFor(producer), "producer");
  assert.equal(
    dashboardFor({ organisation: { kind: "exporter" }, permissions: [] }),
    "trader",
  );
  assert.equal(
    dashboardFor({ organisation: { kind: "regulator" }, permissions: [] }),
    "regulator",
  );
  assert.equal(dashboardFor(admin), "platform");
  assert.equal(
    dashboardFor({ organisation: null, permissions: [] }),
    undefined,
  ); // signed in, acting for nobody
  assert.deepEqual(layoutFor(producer)?.side, []);
});

test("nav and actions need their permission", () => {
  assert.deepEqual(
    navFor(producer).map((n) => n.key),
    ["home", "profile"],
  ); // issuing alone isn't reading: the passports list needs passport:read
  assert.deepEqual(
    navFor({ ...producer, permissions: [permissions.passportRead] }).map(
      (n) => n.key,
    ),
    ["home", "passports", "certificates", "profile"],
  );
  assert.deepEqual(
    navFor(admin).map((n) => n.key),
    ["home", "organisations", "reports", "users", "profile"],
  );
  const orgAdmin = {
    organisation: { kind: "producer" },
    permissions: [permissions.userManage],
  };
  assert.deepEqual(
    navFor(orgAdmin).map((n) => n.key),
    ["home", "users", "profile"],
  );
  // The producer dashboard asks for issue · scan · handover · register facility, of which this person may use two.
  assert.deepEqual(
    actionsFor(producer).map((a) => a.key),
    ["issue", "scan"],
  );
  assert.deepEqual(
    actionsFor(admin).map((a) => a.key),
    ["reviewOrgs", "users", "reports"],
  );
});

test("the kind decides: inspectors never see the trading screens, traders never certify", () => {
  assert.deepEqual(
    navFor(regulator).map((n) => n.key),
    [
      "home",
      "facilities",
      "passports",
      "certificates",
      "flags",
      "reports",
      "users",
      "profile",
    ],
  );
  assert.deepEqual(
    navFor(certifier).map((n) => n.key),
    navFor(regulator).map((n) => n.key),
  );
  // Inspectors look a lot up (code or QR) instead of listing everyone's.
  const passportsOf = (v: typeof regulator) =>
    navFor(v).find((n) => n.key === "passports")?.href;
  assert.equal(passportsOf(regulator), "/scan");
  assert.equal(passportsOf(producerAdmin), "/passports");
  assert.deepEqual(
    navFor(producerAdmin).map((n) => n.key),
    [
      "home",
      "facilities",
      "passports",
      "handovers",
      "transform",
      "certificates",
      "flags",
      "reports",
      "users",
      "profile",
    ],
  );
  assert.deepEqual(
    actionsFor(regulator).map((a) => a.key),
    ["scan", "issueCertificate", "registerFacility", "reports"],
  );
  assert.deepEqual(
    actionsFor(producerAdmin).map((a) => a.key),
    ["issue", "scan", "handover", "registerFacility"],
  );
  for (const path of ["/issue", "/transform", "/handover", "/handovers"])
    assert.equal(canOpen(path, regulator), false, path);
  // Inspectors register their own sites, to issue certificates at.
  assert.equal(canOpen("/facilities/new", regulator), true);
  // A handover or facility stays readable from a passport's page.
  assert.equal(canOpen("/handovers/h1", regulator), true);
  assert.equal(canOpen("/facilities", regulator), true);
  assert.equal(canOpen("/certificates/new", regulator), true);
  assert.equal(canOpen("/certificates/new", producerAdmin), false);
  // Templates are gone for everyone, the platform included.
  const all = { organisation: null, permissions: Object.values(permissions) };
  assert.ok(!navFor(all).some((n) => (n.key as string) === "templates"));
});

test("a member on the bare address only picks a workspace", () => {
  const member = { organisation: null, permissions: [] };
  assert.equal(hasWorkspace(member), false);
  assert.equal(hasWorkspace(producer), true);
  assert.equal(hasWorkspace(admin), true); // the platform is a workspace
  assert.deepEqual(
    navFor(member).map((n) => n.key),
    ["workspaces", "profile"],
  );
});

test("a tile asks for attention only when there is something in it", () => {
  assert.equal(toneOf("incoming", 2), "attention");
  assert.equal(toneOf("incoming", 0), "neutral");
  assert.equal(toneOf("openFlags", 1), "alert");
  assert.equal(toneOf("passportsHeld", 9), "neutral");
});

test("every nav item but Profile sits in a sidebar group, groups in their order", () => {
  const all = { organisation: null, permissions: Object.values(permissions) };
  const items = navFor(all);
  assert.ok(items.every((n) => "group" in n && navGroups.includes(n.group)));
  const order = items.flatMap((n) =>
    "group" in n ? [navGroups.indexOf(n.group)] : [],
  );
  assert.deepEqual(
    order,
    [...order].sort((a, b) => a - b),
  );
});

test("screens open only to the permissions they need; the longest prefix decides", () => {
  assert.equal(canOpen("/organisations", admin), true);
  assert.equal(canOpen("/organisations/o1", admin), true);
  assert.equal(canOpen("/organisations", producer), false);
  assert.equal(canOpen("/reports", producer), false);
  assert.equal(canOpen("/issue", producer), true);
  assert.equal(canOpen("/handover", producer), false);
  assert.equal(
    canOpen("/handovers", { permissions: ["passport:read"] }),
    false,
  );
  assert.equal(
    canOpen("/handovers", { permissions: ["handover:create"] }),
    true,
  );
  assert.equal(canOpen("/dashboard", producer), true);
  assert.equal(canOpen("/profile", { permissions: [] }), true);
});
