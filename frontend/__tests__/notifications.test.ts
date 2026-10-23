import assert from "node:assert/strict";
import { test } from "node:test";

import { certificateKinds, linkOf } from "../lib/notifications.ts";

test("each notification leads to the thing it is about", () => {
  assert.equal(linkOf("handover_started", "h1"), "/handovers/h1");
  assert.equal(linkOf("handover_rejected", "h1"), "/handovers/h1");
  assert.equal(linkOf("flag_raised", "f1"), "/flags?flag=f1");
  assert.equal(linkOf("passport_revoked", "p1"), "/passports/p1");
  assert.equal(
    linkOf("documents_uploaded", "o1"),
    "/organisations?tab=organisations&id=o1",
  );
  assert.equal(linkOf("something_new", "x"), undefined);
  for (const k of certificateKinds) assert.equal(linkOf(k, "c1"), undefined); // resolved through the certificate's subject
});
