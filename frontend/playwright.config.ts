// End-to-end tests against the running stack (`make infra`, `make api`, `make web`).
// Run with `make e2e`, which loads the repo .env for the superadmin sign-in.
import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "e2e",
  timeout: 120_000,
  use: { baseURL: "http://localhost:3000", trace: "retain-on-failure" },
  projects: [{ name: "chromium", use: devices["Desktop Chrome"] }],
});
