// The custody paths beside the happy one, through the real UI: a receiver
// rejects a handover, a sender cancels one, a certifier issues and revokes a
// certificate, and a regulator's admin clears a flag a producer raised.
// Businesses, the template, a facility and the lots are set up through the API.
import { expect, test, type BrowserContext } from "@playwright/test";

import { copy, routes } from "../lib/consts";
import { business, ok, pin, run, sign, signedIn, tinyPDF } from "./helpers";

const certType = `organic_${run}`;
const signing = { headers: { "x-signing-pin": pin } };
const day = (offset: number) =>
  new Date(Date.now() + offset * 86_400_000).toISOString().slice(0, 10);

type Biz = Awaited<ReturnType<typeof business>>;
let farm: Biz, wash: Biz, certifier: Biz, regulator: Biz;
let templateID = "";
let facilityID = "";

/** Issues a fresh lot from the farm; returns its id and lot code. */
async function lot() {
  const p = await ok(
    await farm.context.request.post("/api/passports", {
      data: { facility_id: facilityID, template_id: templateID, quantity: 100 },
      ...signing,
    }),
  );
  return p as unknown as { id: string; code: string };
}

/** The farm hands a lot to the washing station; returns the handover id. */
async function handOver(lotID: string) {
  const h = await ok(
    await farm.context.request.post(`/api/passports/${lotID}/handovers`, {
      data: { to_org: wash.org.id, stage: "washing_station" },
      ...signing,
    }),
  );
  return h.id;
}

test.describe.configure({ mode: "serial" });

test.beforeAll(async ({ browser }) => {
  const adminEmail = process.env.SUPERADMIN_EMAIL;
  const adminPassword = process.env.SUPERADMIN_PASSWORD;
  test.skip(
    !adminEmail || !adminPassword,
    "run through `make e2e` so SUPERADMIN_EMAIL / SUPERADMIN_PASSWORD come from .env",
  );
  const admin: BrowserContext = await signedIn(
    browser,
    adminEmail!,
    adminPassword!,
  );
  const template = await ok(
    await admin.request.post("/api/templates", {
      data: {
        commodity: `e2e_custody_${run}`,
        name: `E2E custody ${run}`,
        unit: "kg",
        originFacilityKind: "farm_plot",
        capacityRule: { type: "facility_capacity" },
        stages: [
          { key: "origin", events: ["ISSUE"] },
          {
            key: "washing_station",
            events: ["HANDOVER", "TRANSFORM"],
            maxYield: 0.2,
          },
          // Never entered here: it only makes the certificate type exist.
          { key: "export", events: ["HANDOVER"], requiredEvidence: [certType] },
        ],
      },
    }),
  );
  templateID = template.id;

  farm = await business(browser, admin, "producer", `E2E Custody Farm ${run}`);
  wash = await business(browser, admin, "processor", `E2E Custody Wash ${run}`);
  certifier = await business(
    browser,
    admin,
    "certifier",
    `E2E Certifier ${run}`,
  );
  regulator = await business(
    browser,
    admin,
    "regulator",
    `E2E Regulator ${run}`,
  );
  await ok(
    await admin.request.put(`/api/orgs/${certifier.org.id}/accreditations`, {
      data: [{ type: certType, valid_from: day(-1), valid_to: day(365) }],
    }),
  );

  const at = () => 29 + Math.random(); // a fresh point each run: facilities may not overlap
  const facility = await ok(
    await farm.context.request.post("/api/facilities", {
      data: {
        name: `Custody plot ${run}`,
        kind: "farm_plot",
        location: { type: "Point", coordinates: [at(), -at() / 15] },
        licence_number: `LIC-${run}`,
        capacity: 100_000,
        capacity_unit: "kg",
        capacity_period: "season",
        licence_valid_to: day(1000),
      },
    }),
  );
  facilityID = facility.id;
  await ok(await admin.request.post(`/api/facilities/${facilityID}/approve`));
});

test("receiver rejects a handover and loses sight of the lot", async () => {
  const { id } = await lot();
  const handover = await handOver(id);
  const r = copy.receiver;

  const page = await wash.context.newPage();
  await page.goto(routes.handoverDetail(handover));
  await page.getByRole("tab", { name: r.reject }).click();
  await page.getByRole("button", { name: r.rejectChips.not_expected }).click();
  // A receiver who also flags stays party to the investigation; without a flag it is out.
  await page.getByLabel(r.rejectFlag).uncheck();
  await sign(page, r.rejectOnly);
  // Out of the lot's custody, the page falls back to the handover's own record.
  await expect(
    page.getByText(copy.handover.rejectedByReceiver).first(),
  ).toBeVisible();

  const after = await wash.context.request.get(`/api/passports/${id}`);
  expect(after.status()).toBe(404);
});

test("sender cancels a handover", async () => {
  const { id } = await lot();
  const handover = await handOver(id);
  const h = copy.handover;

  const page = await farm.context.newPage();
  await page.goto(routes.handoverDetail(handover));
  await page.getByRole("button", { name: h.cancel }).click();
  await page.getByRole("button", { name: h.cancelConfirm }).click();
  await expect(page.getByText(h.cancelled)).toBeVisible();
});

test("certifier issues and revokes a certificate", async () => {
  const { code } = await lot();
  const c = copy.certificates;

  const page = await certifier.context.newPage();
  await page.goto(routes.certificateNew);
  await page.getByLabel(copy.scan.codeLabel).fill(code);
  await page.getByLabel(copy.scan.codeLabel).press("Enter");
  await page
    .getByRole("button", { name: `${c.new.attachPassport} · ${code}` })
    .click();
  await page.getByRole("combobox", { name: c.new.type }).click();
  await page.getByRole("option", { name: new RegExp(run, "i") }).click();
  await page.getByLabel(c.new.document).setInputFiles({
    name: "certificate.pdf",
    mimeType: "application/pdf",
    buffer: tinyPDF,
  });
  await sign(page, c.new.sign);
  await expect(page.getByRole("heading", { name: c.title })).toBeVisible();

  await page
    .getByRole("button", { name: c.revoke, exact: true })
    .first()
    .click();
  await page.getByLabel(c.revokeReason).fill("Issued in error");
  await sign(page, c.revokeSign);
  await expect(page.getByText(c.status.revoked).first()).toBeVisible();
});

test("producer raises a flag and the regulator's admin clears it", async () => {
  const { id } = await lot();
  const f = copy.flags;

  const page = await farm.context.newPage();
  await page.goto(routes.passportDetail(id));
  await page.getByRole("link", { name: f.raise }).click();
  await page.getByLabel(f.raiseReason).fill("Bags weighed twice at the gate");
  await page.getByRole("button", { name: f.raiseConfirm }).click();
  await expect(page).toHaveURL(/flag=/);
  const flagID = new URL(page.url()).searchParams.get("flag")!;

  const reg = await regulator.context.newPage();
  await reg.goto(routes.flag(flagID));
  await reg.getByRole("button", { name: f.clear }).click();
  await reg.getByLabel(f.clearNote).fill("Scales recalibrated");
  await sign(reg, f.clear);
  await expect(reg.getByText(f.status.cleared).first()).toBeVisible();
});
