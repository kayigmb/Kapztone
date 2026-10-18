import assert from "node:assert/strict";
import { test } from "node:test";

import { baseHost, orgAddress } from "./address.ts";

const dev = { protocol: "http:", host: "acme.localhost:3000" };
const platform = { protocol: "http:", host: "localhost:3000" };
const prod = { protocol: "https:", host: "acme.chaintrack.example" };

test("switching organisation keeps the site and swaps the first label", () => {
  assert.equal(orgAddress("kitabi", dev, "acme"), "http://kitabi.localhost:3000");
  assert.equal(orgAddress("kitabi", prod, "acme", "/signin"), "https://kitabi.chaintrack.example/signin");
  assert.equal(orgAddress("acme", platform), "http://acme.localhost:3000");
});

test("the platform is the bare host", () => {
  assert.equal(orgAddress(null, dev, "acme"), "http://localhost:3000");
  assert.equal(baseHost(platform), "localhost:3000");
  // a host that merely starts with similar letters is not the organisation's label
  assert.equal(baseHost({ protocol: "http:", host: "acmecorp.localhost:3000" }, "acme"), "acmecorp.localhost:3000");
});
