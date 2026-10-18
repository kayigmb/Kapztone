import assert from "node:assert/strict";
import { test } from "node:test";

import { navGroups, permissions } from "./consts.ts";
import { actionsFor, dashboardFor, layoutFor, navFor, toneOf } from "./dashboard.ts";

const producer = { organisation: { kind: "producer" }, permissions: ["passport:issue"] };
// The person reviewing organisations also runs templates, users and the audit reports.
const admin = {
  organisation: null,
  permissions: [permissions.orgVerify, permissions.templateRead, permissions.userManage, permissions.auditRead],
};

test("each kind of organisation gets its dashboard; the platform admin gets the platform one", () => {
  assert.equal(dashboardFor(producer), "producer");
  assert.equal(dashboardFor({ organisation: { kind: "exporter" }, permissions: [] }), "trader");
  assert.equal(dashboardFor({ organisation: { kind: "regulator" }, permissions: [] }), "regulator");
  assert.equal(dashboardFor(admin), "platform");
  assert.equal(dashboardFor({ organisation: null, permissions: [] }), undefined); // signed in, acting for nobody
  assert.deepEqual(layoutFor(producer)?.side, ["capacity"]);
});

test("nav and actions need their permission", () => {
  assert.deepEqual(navFor(producer).map((n) => n.key), ["home", "issue", "profile"]);
  assert.deepEqual(navFor(admin).map((n) => n.key), ["home", "adminOrgs", "templates", "reports", "users", "profile"]);
  const orgAdmin = { organisation: { kind: "producer" }, permissions: [permissions.userManage] };
  assert.deepEqual(navFor(orgAdmin).map((n) => n.key), ["home", "users", "profile"]);
  // The producer dashboard asks for issue · scan · handover · register facility, of which this person may use two.
  assert.deepEqual(actionsFor(producer).map((a) => a.key), ["issue", "scan"]);
  assert.deepEqual(actionsFor(admin).map((a) => a.key), ["reviewOrgs", "templates", "users", "reports"]);
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
  assert.ok(items.every((n) => n.key === "profile" || ("group" in n && navGroups.includes(n.group))));
  const order = items.flatMap((n) => ("group" in n ? [navGroups.indexOf(n.group)] : []));
  assert.deepEqual(order, [...order].sort((a, b) => a - b));
});
