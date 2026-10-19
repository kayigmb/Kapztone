// Where a notification leads. Pure: `node --test` runs it.

import { routes } from "./consts.ts";

/** Certificate notifications name the certificate; its page is its subject's, looked up when opened. */
export const certificateKinds: readonly string[] = [
  "certificate_issued",
  "certificate_revoked",
];

/** The screen a notification is about, from its kind and subject; undefined when there is none to open. */
export function linkOf(kind: string, subjectId: string): string | undefined {
  switch (kind) {
    case "handover_started":
    case "handover_accepted":
    case "handover_rejected":
      return routes.handoverDetail(subjectId);
    case "flag_raised":
    case "flag_resolved":
      return routes.flag(subjectId);
    case "passport_revoked":
      return routes.passportDetail(subjectId);
    case "org_registered":
    case "documents_uploaded":
      return routes.adminOrg("organisations", subjectId);
  }
  return undefined;
}
