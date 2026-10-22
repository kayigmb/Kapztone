// Shared by the end-to-end specs: sign-in, businesses set up through the API,
// and the signing step. Run them through `make e2e` (needs the API and the web app).
import {
  expect,
  type Browser,
  type BrowserContext,
  type Page,
} from "@playwright/test";

import { copy } from "../lib/consts";

export const run = Date.now().toString(36);
// Meets the password rule: upper and lower case, a number and a symbol.
export const password = "E2e-password-123456!";
export const pin = "135790";
export const tinyPDF = Buffer.from(
  "%PDF-1.4\n1 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n",
);

export const origin = (slug?: string) =>
  `http://${slug ? `${slug}.` : ""}localhost:3000`;

/** A browser context on an organisation's address, signed in through the API. */
export async function signedIn(
  browser: Browser,
  email: string,
  pass: string,
  slug?: string,
) {
  const context = await browser.newContext({
    baseURL: origin(slug),
    extraHTTPHeaders: { Origin: origin(slug) },
  });
  const res = await context.request.post("/api/auth/login", {
    data: { email, password: pass },
  });
  expect(res.ok(), await res.text()).toBeTruthy();
  return context;
}

/** The few fields this test reads back from the API. */
export type Json = {
  id: string;
  name: string;
  data: Json[];
  verify_link?: string;
};

export async function ok(
  res: Awaited<ReturnType<BrowserContext["request"]["post"]>>,
): Promise<Json> {
  expect(
    res.ok(),
    `${res.url()} → ${res.status()} ${await res.text()}`,
  ).toBeTruthy();
  return ((await res.json()) as { data: Json }).data ?? ({} as Json); // {status, message, data}
}

/** Registers a business, uploads its registration, has the platform verify it, sets the user's PIN. */
export async function business(
  browser: Browser,
  admin: BrowserContext,
  kind: string,
  name: string,
) {
  const slug = `e2e-${kind.slice(0, 4)}-${run}`;
  const email = `${slug}@example.test`;
  const anon = await browser.newContext({
    baseURL: origin(),
    extraHTTPHeaders: {
      Origin: origin(),
      // ponytail: the pass-through forwards this as the visitor's address, which dodges the
      // 10 sign-ups/hour limit across runs. Stops working behind Caddy, which sets the real address.
      "X-Forwarded-For": `10.${run.length}.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}`,
    },
  });
  const org = await ok(
    await anon.request.post("/api/orgs", {
      data: { name, slug, kind, user_name: `${name} manager`, email, password },
    }),
  );
  // Dev hands back the email confirmation link; spend it as the inbox would.
  await ok(
    await anon.request.post("/api/auth/email/verify", {
      data: { token: org.verify_link!.split("/").pop() },
    }),
  );
  await anon.close();

  const context = await signedIn(browser, email, password, slug);
  await ok(
    await context.request.post(`/api/orgs/${org.id}/documents`, {
      multipart: {
        type: "registration",
        file: {
          name: "registration.pdf",
          mimeType: "application/pdf",
          buffer: tinyPDF,
        },
      },
    }),
  );
  await ok(
    await admin.request.post(`/api/orgs/${org.id}/verify`, {
      data: { decision: "verified", reason: "" },
    }),
  );
  await ok(
    await context.request.post("/api/auth/pin", { data: { pin, password } }),
  );
  return { context, name, org };
}

export async function sign(page: Page, button: string) {
  await page.getByLabel(`${copy.signing.label} 1`).click();
  await page.keyboard.type(pin);
  await page.getByRole("button", { name: button }).click();
}
