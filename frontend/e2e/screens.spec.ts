// Screenshots of the built screens at phone and desktop width, light and dark,
// for a person to review (CHANGELOG: visual check). Asserts nothing.
// Run: SCREENS=1 pnpm exec playwright test e2e/screens.spec.ts (stack running).
// Images land in screens/ (gitignored; test-results/ is wiped on every run).
import { test } from "@playwright/test";

import { routes } from "../lib/consts";
import { business, ok, origin, pin, run, signedIn } from "./helpers";

test.skip(!process.env.SCREENS, "set SCREENS=1 to take the screenshots");

const widths = { phone: 390, desktop: 1280 } as const;
const schemes = ["light", "dark"] as const;

test("screens", async ({ browser }) => {
  test.setTimeout(600_000);
  const admin = await signedIn(
    browser,
    process.env.SUPERADMIN_EMAIL!,
    process.env.SUPERADMIN_PASSWORD!,
  );
  const template = await ok(
    await admin.request.post("/api/templates", {
      data: {
        commodity: `e2e_screens_${run}`,
        name: `E2E screens ${run}`,
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
        ],
      },
    }),
  );
  const farm = await business(
    browser,
    admin,
    "producer",
    `E2E Screens Farm ${run}`,
  );
  const facility = await ok(
    await farm.context.request.post("/api/facilities", {
      data: {
        name: `Screens plot ${run}`,
        kind: "farm_plot",
        location: { type: "Point", coordinates: [29 + Math.random(), -1.9] },
        licence_number: `LIC-${run}`,
        capacity: 100_000,
        capacity_unit: "kg",
        capacity_period: "season",
        licence_valid_to: "2030-12-31",
      },
    }),
  );
  await ok(await admin.request.post(`/api/facilities/${facility.id}/approve`));
  const lot = (await ok(
    await farm.context.request.post("/api/passports", {
      data: {
        facility_id: facility.id,
        template_id: template.id,
        quantity: 500,
      },
      headers: { "x-signing-pin": pin },
    }),
  )) as unknown as { id: string; code: string };

  const app = [
    routes.home,
    routes.facilities,
    routes.facilityNew,
    routes.issue,
    routes.passportDetail(lot.id),
    routes.handovers,
    routes.transform,
    routes.scan,
    routes.certificates,
    routes.flags,
    routes.adminReports,
    routes.users,
    routes.notifications,
    routes.profile,
  ];
  const open = [
    routes.landing,
    routes.signin,
    routes.signup,
    routes.passport(lot.code),
  ];

  for (const [size, width] of Object.entries(widths)) {
    for (const scheme of schemes) {
      const shoot = async (
        page: import("@playwright/test").Page,
        path: string,
      ) => {
        await page.goto(path);
        // Fixed settle time; some screens poll, so networkidle never comes.
        await page.waitForTimeout(1500);
        const name =
          path.replace(/[^a-z0-9]+/gi, "-").replace(/^-|-$/g, "") || "home";
        await page.screenshot({
          path: `screens/${size}-${scheme}/${name}.png`,
          fullPage: true,
        });
      };
      const viewport = { width, height: 844 };
      const signedPage = await farm.context.newPage();
      await signedPage.setViewportSize(viewport);
      await signedPage.emulateMedia({ colorScheme: scheme });
      for (const path of app) await shoot(signedPage, path);
      await signedPage.close();

      const visitor = await browser.newContext({
        baseURL: origin(),
        viewport,
        colorScheme: scheme,
      });
      const page = await visitor.newPage();
      for (const path of open) await shoot(page, path);
      await visitor.close();
    }
  }
});
