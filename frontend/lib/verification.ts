// Turns the API's answer to GET /verify/{id} into what the public
// verification page shows. All of that page's copy lives in this file, so a
// second language is a translation of this file and not a rewrite of the page.
//
// Pure: no imports with side effects, so `node --test lib/verification.test.ts` can run it.

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

// ---- Copy ----

export const copy = {
  kicker: "Product passport check",
  verdict: { verified: "Verified", warning: "Warning", failed: "Failed" } as Record<Verdict, string>,
  alert: {
    lead: "Do not accept, pay for or ship this lot.",
    rest: "One or more of its records failed verification. The reasons are listed below.",
  },
  sections: {
    origin: "Origin",
    journey: "Journey",
    madeFrom: "Made from",
    certificates: "Certificates",
    checks: "What we checked",
    technical: "Technical details",
  },
  technicalIntro:
    "When a record is signed, its fingerprint is sealed on a public ledger (a blockchain), where it cannot be changed. We compare each record with its sealed fingerprint every time this page is opened.",
  noCertificates: "No certificates are attached to this lot or its origin.",
  notSealed: "Not sealed yet",
  signedByBoth: "Signed by both parties",
  originMark: "Origin",
  revoked: "Revoked",
  footer: "ChainTrack product passport · The page reflects records sealed up to the time of your scan.",
  certificate: {
    valid: "Valid",
    expiring: "Expiring soon",
    expired: "Expired",
    revoked: "Revoked",
  } as Record<CertificateState, string>,
  result: { verified: "Passed", warning: "Warning", failed: "Failed" } as Record<Verdict, string>,
};

// The seven checks, as the API names them: the rule shown, and what a pass means.
const checks: Record<string, { label: string; pass: string; reason: string }> = {
  record_intact: {
    label: "Records unchanged since signing",
    pass: "Every passport in this lot's history reads exactly as it did when it was signed.",
    reason: "Records unchanged since signing",
  },
  anchored: {
    label: "Records match their sealed copies",
    pass: "Each record, its holder and its quantity match the copy sealed when it was signed.",
    reason: "Records match their sealed copies",
  },
  origin: {
    label: "Origin is registered and licensed",
    pass: "Every branch of this lot's history ends at a registered facility owned by a verified organisation.",
    reason: "Origin facility verified",
  },
  capacity: {
    label: "Within declared capacity",
    pass: "No more was issued from the origin than its declared capacity for the period.",
    reason: "Origin within declared capacity",
  },
  mass_balance: {
    label: "Quantities balance",
    pass: "At every processing step, what came out is no more than the inputs could yield.",
    reason: "Quantities balance at every step",
  },
  certificates: {
    label: "Certificates valid",
    pass: "Every certificate is in date, unrevoked, from an accredited issuer, and unchanged since it was issued.",
    reason: "Certificates valid",
  },
  status: {
    label: "No open concerns or revocations",
    pass: "Nothing in this lot's history has been revoked or is being looked into.",
    reason: "No open concerns",
  },
};

const eventTitles: Record<string, string> = {
  ISSUE: "Issued at origin",
  HANDOVER: "Handed over",
  TRANSFORM: "Processed",
  FLAG: "Concern raised",
  REVOKE: "Revoked",
};

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

const plural = (n: number, one: string, many: string) => `${n} ${n === 1 ? one : many}`;

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
  const from = places ? `This ${product.toLowerCase()} comes from ${places}.` : "";
  switch (api.verdict) {
    case "verified":
      return `${from} Every record in its history matches the copy sealed when it was signed, and nothing has changed since.`.trim();
    case "warning":
      return `${from} Its records are intact, but ${plural(api.reasons.length, "point needs", "points need")} attention. Read them before you rely on this passport.`.trim();
    case "failed":
      return "This passport failed verification. Do not rely on it until the points below are resolved.";
  }
}

export function toView(api: ApiVerification, now: Date = new Date()): View {
  const names = new Map(api.organisations.map((o) => [o.id, o.name]));
  const name = (id: string | null | undefined) => (id ? (names.get(id) ?? "Unknown organisation") : undefined);
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
      `Checked ${formatDateTime(now)}`,
      plural(api.genealogy.nodes.length + api.certificates.length, "record", "records"),
      plural(api.organisations.length, "organisation", "organisations"),
    ].join(" · "),
    product: {
      name: product,
      lotCode: p.id,
      origin: p.facility_id !== null,
      facts: [
        { label: "Commodity", value: product },
        { label: "Quantity", value: formatQuantity(p.quantity, p.unit), mono: true },
        { label: "Current holder", value: name(p.organisation_id) ?? "", wide: true },
        { label: "Status", value: humanize(p.status) },
        { label: "Last updated", value: formatDate(events.at(-1)?.created_at ?? p.created_at) },
      ],
      issued: `Passport issued ${formatDate(p.created_at)}`,
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
            ? `Revoked by its issuer${c.revoke_reason ? `: ${c.revoke_reason}` : "."} Ask the holder for a replacement.`
            : state === "expired"
              ? `Expired ${formatDate(c.valid_to)}. Ask the holder for a current one before you rely on it.`
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
      { label: "Passport record · fingerprint", value: p.data_hash },
      { label: "Passport record · sealed in transaction", value: p.tx_hash },
      ...events.map((e) => ({
        label: `${eventTitles[e.kind] ?? e.kind} · ${formatDate(e.created_at)} · sealed in transaction`,
        value: e.tx_hash,
      })),
      ...api.certificates.map((c) => ({ label: `${humanize(c.type)} · fingerprint`, value: c.doc_hash })),
    ],
  };
}
