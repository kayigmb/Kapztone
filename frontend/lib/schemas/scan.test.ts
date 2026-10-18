import assert from "node:assert/strict";
import { test } from "node:test";

import { lotCodeSchema, passportRef } from "./scan.ts";

test("a scanned label opens the passport it names; a typed code is used as it is", () => {
  assert.equal(passportRef("https://chaintrack.example/p/ACME-2026-0412"), "ACME-2026-0412");
  assert.equal(passportRef("http://localhost:3000/p/ACME-2026-0412?lang=en"), "ACME-2026-0412");
  assert.equal(passportRef("  ACME-2026-0412 "), "ACME-2026-0412");
  assert.equal(passportRef("https://example.com/somewhere-else"), "https://example.com/somewhere-else");
});

test("an empty lot code is refused", () => {
  assert.equal(lotCodeSchema.safeParse({ code: "   " }).success, false);
  assert.equal(lotCodeSchema.parse({ code: "acme-2026-0001" }).code, "acme-2026-0001");
});
