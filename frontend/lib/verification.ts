// Turns the API's answer to GET /verify/{id} into what the public
// verification page shows. The page's copy lives in lib/consts.ts.
//
// Pure: no imports with side effects, so `node --test lib/verification.test.ts` can run it.

import { copy } from "./consts.ts";

export type Verdict = "verified" | "warning" | "failed";
export type CertificateState = "valid" | "expiring" | "expired" | "revoked";

type Named = { id: string; name: string };

export type ApiEvent = {
  id: string;
  kind: string;
  from_org: string | null;
  to_org: string | null;
  quantity: number;
  tx_hash: string | null;
  created_at: string;
};

export type ApiPassport = {
  id: string;
  organisation_id: string;
  facility_id: string | null;
  transform_id: string | null;
  commodity: string;
  unit: string;
  quantity: number;
  remaining: number;
  status: string;
  data_hash: string;
  tx_hash: string | null;
  created_at: string;
  events?: ApiEvent[];
};

export type ApiCertificate = {
  id: string;
  issuer_org_id: string;
  subject_id: string;
  type: string;
  valid_from: string;
  valid_to: string;
  doc_hash: string;
  status: "valid" | "expired" | "revoked";
  revoke_reason: string | null;
  tx_hash: string | null;
};

export type ApiVerification = {
  verdict: Verdict;
  reasons: { level: Verdict; check: string; code: string; message: string; subject_id: string }[];
  checks: { name: string; result: Verdict }[];
  passport: ApiPassport;
  origins: { facility: Named & { kind: string }; organisation: Named }[];
  organisations: Named[];
  genealogy: {
    nodes: ApiPassport[];
    edges: { parent_id: string; child_id: string; stage: string; quantity_used: number }[];
  };
  certificates: ApiCertificate[];
};

export type Result = "pass" | "warn" | "fail";

export type View = {
  verdict: Verdict;
  summary: string;
  reasons: { result: Result; text: string }[];
  meta: string;
  product: {
    name: string;
    lotCode: string;
    origin: boolean; // issued at a facility, as opposed to made from other lots
    facts: { label: string; value: string; mono?: boolean; wide?: boolean }[];
    issued: string;
  };
  origins: { id: string; name: string; kind: string; owner: string }[];
  journey: {
    id: string;
    kind: string;
    title: string;
    date: string;
    parties: { from?: string; to?: string };
    quantity: string;
    signedByBoth: boolean;
    state: "done" | "current" | "warning" | "failed";
  }[];
  madeFrom: { lotCode: string; description: string; quantity: string; origin: boolean; revoked: boolean }[];
  certificates: {
    id: string;
    type: string;
    issuer: string;
    state: CertificateState;
    validFrom: string;
    validTo: string;
    note?: string;
  }[];
  checks: { name: string; label: string; result: Verdict; explanation: string }[];
  technical: { label: string; value: string | null }[];
};

const t = copy.verification;
const checks = t.checks;
const eventTitles = t.eventTitles;

// ---- Formatting (content.md → Formatting) ----

const zone = "Africa/Kigali"; // Central Africa Time
const months = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
const partsFmt = new Intl.DateTimeFormat("en-GB", {
  day: "numeric", month: "numeric", year: "numeric", hour: "2-digit", minute: "2-digit", hour12: false, timeZone: zone,
});
const numberFmt = new Intl.NumberFormat("en", { maximumFractionDigits: 1 });

// Intl's own short months differ by locale ("Sept"), so the month name comes from the list above.
function parts(iso: string | Date) {
  const p = Object.fromEntries(partsFmt.formatToParts(new Date(iso)).map((x) => [x.type, x.value]));
  return { date: `${Number(p.day)} ${months[Number(p.month) - 1]} ${p.year}`, time: `${p.hour}:${p.minute}` };
}

export const formatDate = (iso: string | Date) => parts(iso).date;
export const formatDateTime = (iso: string | Date) => {
  const p = parts(iso);
  return `${p.date}, ${p.time} CAT`;
};
export const formatQuantity = (n: number, unit: string) => `${numberFmt.format(n)} ${unit}`;

/** "farm_plot" → "Farm plot". */
export function humanize(key: string): string {
  const words = key.replace(/_/g, " ").trim();
  return words.charAt(0).toUpperCase() + words.slice(1);
}

const expiringWithinDays = 30;

function certificateState(c: ApiCertificate, now: Date): CertificateState {
  if (c.status !== "valid") return c.status;
  const daysLeft = (new Date(c.valid_to).getTime() - now.getTime()) / 86_400_000;
  return daysLeft <= expiringWithinDays ? "expiring" : "valid";
}

const maxReasons = 4;
const order: Record<Verdict, number> = { failed: 0, warning: 1, verified: 2 };

/** At most four reasons: failures first, then warnings, then what passed. Repeats are said once. */
function topReasons(api: ApiVerification): View["reasons"] {
  const seen = new Set<string>();
  const found = [...api.reasons]
    .sort((a, b) => order[a.level] - order[b.level])
    .filter((r) => !seen.has(r.code) && seen.add(r.code))
    .map((r) => ({ result: (r.level === "failed" ? "fail" : "warn") as Result, text: r.message }));
  const passed = api.checks
    .filter((c) => c.result === "verified" && checks[c.name])
    .map((c) => ({ result: "pass" as Result, text: checks[c.name].reason }));
  return [...found, ...passed].slice(0, maxReasons);
}

function summary(api: ApiVerification, product: string): string {
  const places = api.origins.map((o) => o.facility.name).join(" and ");
  const from = places ? t.summary.from(product.toLowerCase(), places) : "";
  switch (api.verdict) {
    case "verified":
      return `${from} ${t.summary.verified}`.trim();
    case "warning":
      return `${from} ${t.summary.warning(t.summary.pointsNeed(api.reasons.length))}`.trim();
    case "failed":
      return t.summary.failed;
  }
}

export function toView(api: ApiVerification, now: Date = new Date()): View {
  const names = new Map(api.organisations.map((o) => [o.id, o.name]));
  const name = (id: string | null | undefined) => (id ? (names.get(id) ?? t.unknownOrg) : undefined);
  const p = api.passport;
  const product = humanize(p.commodity);
  const events = p.events ?? [];

  const journey: View["journey"] = events.map((e, i) => ({
    id: e.id,
    kind: e.kind,
    title: eventTitles[e.kind] ?? humanize(e.kind.toLowerCase()),
    date: formatDateTime(e.created_at),
    parties: { from: name(e.from_org), to: name(e.to_org) },
    quantity: formatQuantity(e.quantity, p.unit),
    signedByBoth: e.kind === "HANDOVER",
    state: e.kind === "REVOKE" ? "failed" : e.kind === "FLAG" ? "warning" : i === events.length - 1 ? "current" : "done",
  }));

  const explain = (check: string) => {
    const said = [...new Set(api.reasons.filter((r) => r.check === check).map((r) => r.message))];
    return said.length ? said.join(" ") : (checks[check]?.pass ?? "");
  };

  return {
    verdict: api.verdict,
    summary: summary(api, product),
    reasons: topReasons(api),
    meta: [
      t.meta.checked(formatDateTime(now)),
      t.meta.records(api.genealogy.nodes.length + api.certificates.length),
      t.meta.organisations(api.organisations.length),
    ].join(" · "),
    product: {
      name: product,
      lotCode: p.id,
      origin: p.facility_id !== null,
      facts: [
        { label: t.facts.commodity, value: product },
        { label: t.facts.quantity, value: formatQuantity(p.quantity, p.unit), mono: true },
        { label: t.facts.holder, value: name(p.organisation_id) ?? "", wide: true },
        { label: t.facts.status, value: humanize(p.status) },
        { label: t.facts.updated, value: formatDate(events.at(-1)?.created_at ?? p.created_at) },
      ],
      issued: t.issued(formatDate(p.created_at)),
    },
    origins: api.origins.map((o) => ({
      id: o.facility.id,
      name: o.facility.name,
      kind: humanize(o.facility.kind),
      owner: o.organisation.name,
    })),
    journey,
    madeFrom: api.genealogy.nodes
      .filter((n) => n.id !== p.id)
      .map((n) => ({
        lotCode: n.id,
        description: humanize(n.commodity),
        quantity: formatQuantity(n.quantity, n.unit),
        origin: n.facility_id !== null,
        revoked: n.status === "revoked",
      })),
    certificates: api.certificates.map((c) => {
      const state = certificateState(c, now);
      return {
        id: c.id,
        type: humanize(c.type),
        issuer: name(c.issuer_org_id) ?? "",
        state,
        validFrom: formatDate(c.valid_from),
        validTo: formatDate(c.valid_to),
        note:
          state === "revoked"
            ? t.certificateNote.revoked(c.revoke_reason)
            : state === "expired"
              ? t.certificateNote.expired(formatDate(c.valid_to))
              : undefined,
      };
    }),
    checks: api.checks.map((c) => ({
      name: c.name,
      label: checks[c.name]?.label ?? humanize(c.name),
      result: c.result,
      explanation: explain(c.name),
    })),
    technical: [
      { label: t.technical.fingerprint, value: p.data_hash },
      { label: t.technical.sealedIn, value: p.tx_hash },
      ...events.map((e) => ({
        label: t.technical.eventSealed(eventTitles[e.kind] ?? e.kind, formatDate(e.created_at)),
        value: e.tx_hash,
      })),
      ...api.certificates.map((c) => ({ label: t.technical.certificateFingerprint(humanize(c.type)), value: c.doc_hash })),
    ],
  };
}
