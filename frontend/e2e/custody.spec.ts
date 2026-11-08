import { expect, test, type BrowserContext } from "@playwright/test";

import { copy, routes } from "../lib/consts";
import { business, ok, pin, run, sign, signedIn, tinyPDF } from "./helpers";

const certType = `organic_${run}`;
const signing = { headers: { "x-signing-pin": pin } };
const day = (offset: number) =>
  new Date(Date.now() + offset * 86_400_000).toISOString().slice(0, 10);

type Biz = Awaited<ReturnType<typeof business>>;
let farm: Biz, wash: Biz, certifier: Biz, regulator: Biz;
let admin: BrowserContext;
let facilityID = "";

/** Issues a fresh lot from the farm; returns its id and lot code. */
async function lot() {
  const p = await ok(
    await farm.context.request.post("/api/passports", {
      data: {
        facility_id: facilityID,
        commodity: `e2e_custody_${run}`,
        unit: "kg",
        quantity: 100,
      },
      ...signing,
    }),
  );
  return p as unknown as { id: string; code: string };
}

/** The farm hands a lot to the washing station; returns the handover id. */
async function handOver(lotID: string) {
  // To the washing station's manager: a handover goes to a person.
  const people = (await (
    await farm.context.request.get(`/api/orgs/${wash.org.id}/people`)
  ).json()) as { data: { id: string }[] };
  const h = await ok(
    await farm.context.request.post(`/api/passports/${lotID}/handovers`, {
      data: {
        to_org: wash.org.id,
        to_user: people.data[0].id,
        stage: "washing_station",
      },
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
  admin = await signedIn(browser, adminEmail!, adminPassword!);
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
  await expect(page.getByText(h.cancelled).first()).toBeVisible();
});

test("certifier issues and revokes a certificate", async () => {
  const { code } = await lot();
  const c = copy.certificates;

  const page = await certifier.context.newPage();
  await page.goto(routes.certificateNew);
  await page.getByLabel(copy.scan.codeLabel).fill(code);
  await page.getByLabel(copy.scan.codeLabel).press("Enter");
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

test("producer raises a flag, the platform assigns it, the regulator's admin clears it", async () => {
  const { id } = await lot();
  const f = copy.flags;

  const page = await farm.context.newPage();
  await page.goto(routes.passportDetail(id));
  await page.getByRole("link", { name: f.raise }).click();
  await page.getByLabel(f.raiseReason).fill("Bags weighed twice at the gate");
  await page.getByRole("button", { name: f.raiseConfirm }).click();
  await expect(page).toHaveURL(/flag=/);
  const flagID = new URL(page.url()).searchParams.get("flag")!;

  // A regulator only sees the flags given to it: the platform assigns this one.
  const me = await ok(await regulator.context.request.get("/api/me"));
  await ok(
    await admin.request.put(`/api/flags/${flagID}/assignee`, {
      data: { user_id: me.user!.id },
    }),
  );
  const reg = await regulator.context.newPage();
  await reg.goto(routes.flag(flagID));
  await reg.getByRole("button", { name: f.clear }).click();
  await reg.getByLabel(f.clearNote).fill("Scales recalibrated");
  await sign(reg, f.clear);
  await expect(reg.getByText(f.status.cleared).first()).toBeVisible();
});

test("a regulator inspects but never trades, and registers only its own sites", async () => {
  const nav = copy.shell.nav;
  const page = await regulator.context.newPage();
  await page.goto(routes.home);
  const sidebar = page.getByRole("navigation").first();
  for (const shown of [
    nav.passports,
    nav.certificates,
    nav.flags,
    nav.facilities,
  ])
    await expect(sidebar.getByRole("link", { name: shown })).toBeVisible();
  for (const hidden of [nav.handovers, nav.transform])
    await expect(sidebar.getByRole("link", { name: hidden })).toHaveCount(0);
  // Its Passports is the lookup page, not a list of everyone's lots.
  await sidebar.getByRole("link", { name: nav.passports }).click();
  await expect(page).toHaveURL(/\/scan$/);

  // Typed by hand, a trading screen sends the regulator back to its dashboard.
  for (const blocked of [routes.issue, routes.transform, routes.handovers]) {
    await page.goto(blocked);
    await expect(page).toHaveURL(/\/dashboard$/);
  }

  // It looks any organisation's lot up and never lists them all; it registers its own site but approves no one's.
  const { id } = await lot();
  expect(
    (await regulator.context.request.get(`/api/passports/${id}`)).ok(),
  ).toBe(true);
  const listed = await ok(
    await regulator.context.request.get("/api/passports?size=100"),
  );
  expect(listed.data.some((x) => x.id === id)).toBe(false);
  const at = () => 29 + Math.random();
  const pending = await ok(
    await farm.context.request.post("/api/facilities", {
      data: {
        name: `Regulated plot ${run}`,
        kind: "farm_plot",
        location: { type: "Point", coordinates: [at(), -at() / 15] },
        licence_valid_to: day(1000),
      },
    }),
  );
  const approve = await regulator.context.request.post(
    `/api/facilities/${pending.id}/approve`,
  );
  expect(approve.status()).toBe(403);
  await ok(
    await regulator.context.request.post("/api/facilities", {
      data: {
        name: `Regulator office ${run}`,
        kind: "warehouse",
        location: { type: "Point", coordinates: [at(), -at() / 15] },
        licence_valid_to: day(1000),
      },
    }),
  );
  const handover = await regulator.context.request.post(
    `/api/passports/${id}/handovers`,
    { data: { to_org: wash.org.id }, ...signing },
  );
  expect(handover.status()).toBe(403);
});
