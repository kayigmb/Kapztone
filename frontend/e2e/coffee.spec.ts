import { expect, test } from "@playwright/test";

import { copy } from "../lib/consts";
import { business, ok, origin, run, sign, signedIn } from "./helpers";

const commodity = `e2e_coffee_${run}`;

test("coffee: farm plot → issue → handover → accept → transform → public page", async ({
  browser,
}) => {
  const adminEmail = process.env.SUPERADMIN_EMAIL;
  const adminPassword = process.env.SUPERADMIN_PASSWORD;
  test.skip(
    !adminEmail || !adminPassword,
    "run through `make e2e` so SUPERADMIN_EMAIL / SUPERADMIN_PASSWORD come from .env",
  );
  const admin = await signedIn(browser, adminEmail!, adminPassword!);

  const template = await ok(
    await admin.request.post("/api/templates", {
      data: {
        commodity,
        name: `E2E coffee ${run}`,
        unit: "kg",
        originFacilityKind: "farm_plot",
        capacityRule: { type: "facility_capacity" },
        stages: [
          { key: "origin", events: ["ISSUE"] },
          {
            key: "washing_station",
            events: ["HANDOVER", "TRANSFORM"],
            maxYield: 0.2,
          }, // the API holds a fraction
        ],
      },
    }),
  );
  const farm = await business(browser, admin, "producer", `E2E Farm ${run}`);
  const wash = await business(
    browser,
    admin,
    "processor",
    `E2E Washing ${run}`,
  );

  const p = await farm.context.newPage();
  const f = copy.facilities.new;
  const plot = `Plot ${run}`;
  await p.goto("/facilities/new");
  // A different spot each run: two facilities on one point overlap, and the API refuses that.
  const spot = () => 40 + Math.floor(Math.random() * 240);
  await p
    .getByRole("region", { name: f.mapLabel })
    .click({ position: { x: spot(), y: spot() } });
  await p.getByLabel(f.name).fill(plot);
  await p.getByText("Farm plot", { exact: true }).click();
  await p.getByLabel(f.licence).fill(`LIC-${run}`);
  await p.getByLabel(f.licenceValid).fill("2030-12-31");
  await p.getByLabel(f.capacity).fill("5000");
  await p.getByRole("button", { name: f.save }).click();
  await expect(p).toHaveURL(/\/facilities$/);
  await expect(p.getByText(plot, { exact: true })).toBeVisible();

  const facilities = await ok(
    await farm.context.request.get("/api/facilities?size=100"),
  );
  const facility = facilities.data.find((x) => x.name === plot)!;
  await ok(await admin.request.post(`/api/facilities/${facility.id}/approve`));

  const i = copy.issue;
  await p.goto("/issue");
  // The pickers are paged by the API: search for this run's template and plot.
  const searches = p.getByRole("searchbox", { name: copy.list.search });
  await searches.nth(0).fill(template.name);
  await searches.nth(1).fill(plot);
  await p.getByText(template.name, { exact: true }).click();
  await p.getByText(plot, { exact: true }).click();
  await p.getByRole("button", { name: "Continue" }).click();
  await p.getByLabel(new RegExp(`^${i.quantity}`)).fill("1000");
  await p.getByRole("button", { name: i.review }).click();
  await sign(p, i.sign);
  await expect(p.getByRole("heading", { name: i.createdTitle })).toBeVisible();
  const code = (await p.locator(".code").first().textContent())!.trim();
  expect(code).not.toBe("");

  const h = copy.handover;
  await p.getByRole("link", { name: i.handOver }).click();
  await p.getByRole("combobox", { name: h.stage }).click();
  await p.getByRole("option", { name: "Washing station" }).click();
  await p.getByRole("button", { name: h.next }).click();
  await p.getByLabel(h.receiverSearch).fill(wash.name);
  await p.getByRole("radio", { name: new RegExp(wash.name) }).click();
  await p.getByRole("button", { name: h.review }).click();
  await sign(p, h.sign);
  await expect(p.getByRole("heading", { name: h.sentTitle })).toBeVisible();
  const handoverURL = await p
    .getByRole("link", { name: copy.receiver.open })
    .getAttribute("href");

  const w = await wash.context.newPage();
  const r = copy.receiver;
  await w.goto(handoverURL!);
  await w.getByLabel(r.received).fill("1000");
  await sign(w, r.accept);
  await expect(w.getByRole("heading", { name: r.completeTitle })).toBeVisible();

  const t = copy.transform;
  await w.goto("/transform");
  await w.getByLabel(copy.scan.codeLabel).fill(code);
  await w.getByRole("button", { name: "Add lot" }).click();
  await w.getByLabel(t.quantity, { exact: true }).fill("1000");
  await w.getByText("Washing station", { exact: true }).click();
  await w.getByLabel(t.output(1)).fill("180");
  await sign(w, t.sign(1));
  await expect(
    w.getByRole("heading", { name: t.createdTitle(1) }),
  ).toBeVisible();
  const child = (await w.locator("ul .code").first().textContent())!.trim();

  const visitor = await browser.newContext();
  const v = await visitor.newPage();
  await v.goto(`${origin()}/p/${encodeURIComponent(child)}`);
  const verdict = v.getByRole("status").first();
  await expect(verdict).toBeVisible();
  // Verified needs the chain and the relayer running; without them the records wait to be sealed (Warning).
  await expect(verdict).not.toContainText("Failed");
  await expect(v.getByText(child).first()).toBeVisible();
});
