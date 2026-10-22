import assert from "node:assert/strict";
import { test } from "node:test";

import {
  confirmCopy,
  copy,
  marketing,
  platformCopy,
  reportsCopy,
  screenCopy,
} from "../lib/consts.ts";

test("static copy survived the move to JSON", () => {
  assert.equal(copy.brand, "ChainTrack");
  assert.equal(copy.errors.tryAgain, "Try again");
  assert.equal(copy.shell.nav.home, "Dashboard");
  assert.equal(reportsCopy.title, "Reports and audit");
  assert.equal(typeof copy.transform.output, "function");
});

test("template wrappers fill their placeholders", () => {
  assert.equal(
    copy.errors.unexpectedRef("abc"),
    "Something went wrong on our side. Try again; if it keeps happening, quote reference abc.",
  );
  assert.equal(copy.list.page(2, 9), "Page 2 of 9");
  assert.equal(copy.removeFile("a.pdf"), "Remove a.pdf");
  assert.equal(reportsCopy.checked(7), "7 entries checked");
  assert.equal(reportsCopy.brokenAt(null), "Broken");
  assert.equal(reportsCopy.brokenAt(12), "Broken at seq 12");
  assert.equal(copy.transform.output(1), "Output 1");
  assert.equal(copy.verification.steps(1), "1 step");
  assert.equal(copy.verification.steps(4), "4 steps");
  assert.equal(screenCopy.passportDetail.held(10, 100, "kg"), "10 of 100 kg");
});

test("logic copy kept its branches", () => {
  assert.equal(copy.flags.sub(0), "Nothing open");
  assert.equal(copy.flags.sub(2), "2 flags open");
  assert.equal(copy.dashboard.waitingDays(0), "Today");
  assert.equal(
    copy.dashboard.greeting("Ada Lovelace").startsWith("Good "),
    true,
  );
});

test("marketing covers every role with title, body and two steps", () => {
  const guides = (
    marketing as unknown as {
      roles: {
        guides: Record<
          string,
          { title: string; body: string; steps: [string, string][] }
        >;
      };
    }
  ).roles.guides;
  for (const role of [
    "producer",
    "logistics",
    "certifier",
    "regulator",
  ] as const) {
    const entry = guides[role];
    assert.ok(entry.title, role);
    assert.ok(entry.body, role);
    assert.ok(entry.steps.length >= 2, role);
  }
});

test("confirm builders still return title, body and label", () => {
  const c = confirmCopy.facilityApprove("Acme");
  assert.equal(c.title, "Approve Acme?");
  assert.ok(c.body && c.label);
  assert.equal(platformCopy.tiles.passports.note(3), "3 active");
});
