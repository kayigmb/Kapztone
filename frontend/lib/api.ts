// The browser's access to the API. Every call goes through this site's
// pass-through (app/api/[...path]/route.ts), so the session cookie and the
// organisation in the address travel with it. Every answer is checked
// against a zod schema; every failure becomes one ApiError.
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { z } from "zod";

import {
  apiBase,
  apiPaths,
  certificateStatus,
  handoverPhotoField,
  copy,
  flagStatus,
  handoverStatus,
  limits,
  orgStatus,
  panelRows,
  passportStatus,
  queryKeys,
  routes,
  type TileKey,
} from "@/lib/consts";
import type { ApiVerification } from "@/lib/verification";
import { yieldToFraction, yieldToPct } from "@/lib/schemas/template";
import { listQuery, type ListParams } from "@/lib/list";
import { orgAddress } from "@/lib/address";
import { placeSchema } from "@/lib/geocode";

declare module "@tanstack/react-query" {
  interface Register {
    /** inPlace: the screen shows the failure itself, so no error toast. */
    mutationMeta: { inPlace?: boolean };
  }
}

// ---- Errors ----

// Every JSON answer is {status, message, data}; on a refusal `data` may hold details.
const envelope = z.object({
  status: z.number(),
  message: z.string().catch(""),
  data: z.unknown(),
});
const errorDetails = z
  .object({
    two_factor_required: z.boolean().optional(),
    email_unverified: z.boolean().optional(),
  })
  .nullable()
  .catch(null);

/** One kind of failure for every call. `status` 0 means the API could not be reached. */
export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
    readonly twoFactorRequired = false,
    /** The request's id in the server logs (X-Request-ID). */
    readonly requestId?: string,
    /** Right password, but the email was never confirmed; the API has sent a new link. */
    readonly emailUnverified = false,
  ) {
    super(message);
  }
}

/** The API's refusals read as phrases ("signing PIN is incorrect"); a toast shows a sentence. */
const sentence = (m: string) =>
  m[0].toUpperCase() + m.slice(1) + (/[.!?]$/.test(m) ? "" : ".");

/** What to tell a person about a failure they can't fix by changing a field. */
export function problemOf(err: unknown): string {
  if (err instanceof ApiError) {
    if (err.status === 0 || err.status === 502) return copy.errors.unreachable; // 502: the pass-through couldn't reach the API
    if (err.status >= 500)
      return err.requestId
        ? copy.errors.unexpectedRef(err.requestId)
        : copy.errors.unexpected;
    return err.message ? sentence(err.message) : copy.errors.unexpected;
  }
  if (err instanceof z.ZodError) return copy.errors.badAnswer;
  return copy.errors.unexpected;
}

// ---- Calls ----

/** A write's answer: what changed, and the API's sentence about it (the toast). */
export type Sent<T> = { data: T; message: string };

async function call<T>(
  path: string,
  schema: z.ZodType<T>,
  init?: RequestInit,
): Promise<Sent<T>> {
  let res: Response;
  try {
    res = await fetch(apiBase + path, { ...init, cache: "no-store" });
  } catch {
    throw new ApiError(0, copy.errors.unreachable);
  }
  const body = envelope.safeParse(await res.json().catch(() => null));
  if (!res.ok) {
    const details = errorDetails.parse(body.data?.data ?? null);
    throw new ApiError(
      res.status,
      body.data?.message ?? "",
      details?.two_factor_required,
      res.headers.get("x-request-id") ?? undefined,
      details?.email_unverified,
    );
  }
  if (!body.success) throw body.error;
  return { data: schema.parse(body.data.data), message: body.data.message };
}

/** Reads: the answer's data, checked against its schema. */
export const api = async <T>(
  path: string,
  schema: z.ZodType<T>,
  init?: RequestInit,
): Promise<T> => (await call(path, schema, init)).data;

/** Writes: the data plus the API's message, which app/providers.tsx toasts. */
const write = call;

/** Whether a mutation's result is a write's answer (it carries a message). */
export const isSent = (v: unknown): v is Sent<unknown> =>
  typeof v === "object" && v !== null && "message" in v && "data" in v;

const send = (method: string, body: unknown, pin?: string): RequestInit => ({
  method,
  headers: pin
    ? { "content-type": "application/json", "x-signing-pin": pin }
    : { "content-type": "application/json" },
  body: JSON.stringify(body),
});

// ---- Shapes (backend JSON) ----

const orgSummary = z.object({
  id: z.string(),
  name: z.string(),
  slug: z.string(),
  kind: z.string(),
  status: z.string(),
});

export const meSchema = z.object({
  user: z.object({ id: z.string(), email: z.string(), name: z.string() }),
  organisation: orgSummary.nullable(),
  role: z.string(),
  permissions: z
    .array(z.string())
    .nullable()
    .transform((p) => p ?? []),
  pin_set: z.boolean().optional(),
  // Not every API version says it; undefined means "unknown".
  two_factor: z.boolean().optional(),
});
export type Me = z.infer<typeof meSchema>;

export const organisationSchema = orgSummary.extend({
  review_reason: z.string().nullable(),
  reviewed_at: z.string().nullable(),
  created_at: z.string(),
});
export type Organisation = z.infer<typeof organisationSchema>;

export const documentSchema = z.object({
  id: z.string(),
  type: z.string(),
  file_name: z.string(),
  content_type: z.string(),
  size_bytes: z.number(),
  created_at: z.string(),
});
export type OrgDocument = z.infer<typeof documentSchema>;

const page = <T extends z.ZodType>(item: T) =>
  z.object({ data: z.array(item), total: z.number() });
/** One page of a list, paged, searched and sorted by the API. */
export type Paged<T> = { data: T[]; total: number };

const stage = z.object({
  key: z.string(),
  events: z.array(z.string()).optional(),
  maxYield: z
    .number()
    .nullable()
    .transform((f) => (f == null ? f : yieldToPct(f)))
    .optional(), // percent on screen
  requiredEvidence: z.array(z.string()).optional(),
  handoverTolerance: z.number().nullable().optional(),
});
const templateSchema = z.object({
  id: z.string(),
  commodity: z.string(),
  name: z.string(),
  unit: z.string(),
  originFacilityKind: z.string().nullable().optional(),
  capacityRule: z
    .object({
      type: z.string(),
      kgPerHaPerSeason: z.number().nullable().optional(),
    })
    .nullable()
    .optional(),
  stages: z.array(stage),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
});
export type Template = z.infer<typeof templateSchema>;

export const accreditationSchema = z.object({
  type: z.string(),
  valid_from: z.string(),
  valid_to: z.string(),
});
export type AccreditationRow = z.infer<typeof accreditationSchema>;
// A bare list, or the usual paged envelope: either is read the same way.
const accreditationList = z.union([
  z.array(accreditationSchema),
  page(accreditationSchema).transform((p) => p.data),
]);

const verdict = z.enum(["verified", "warning", "failed"]);
const named = z.object({ id: z.string(), name: z.string() });
const passport = z.object({
  id: z.string(),
  code: z.string().optional(),
  organisation_id: z.string(),
  facility_id: z.string().nullable(),
  transform_id: z.string().nullable(),
  commodity: z.string(),
  unit: z.string(),
  quantity: z.number(),
  remaining: z.number(),
  status: z.string(),
  data_hash: z.string(),
  tx_hash: z.string().nullable(),
  created_at: z.string(),
  events: z
    .array(
      z.object({
        id: z.string(),
        kind: z.string(),
        from_org: z.string().nullable(),
        to_org: z.string().nullable(),
        quantity: z.number(),
        tx_hash: z.string().nullable(),
        created_at: z.string(),
      }),
    )
    .optional(),
});
const verificationSchema: z.ZodType<ApiVerification> = z.object({
  verdict,
  reasons: z.array(
    z.object({
      level: verdict,
      check: z.string(),
      code: z.string(),
      message: z.string(),
      subject_id: z.string(),
    }),
  ),
  checks: z.array(z.object({ name: z.string(), result: verdict })),
  passport,
  origins: z.array(
    z.object({
      facility: named.extend({
        kind: z.string(),
        location: z
          .object({ type: z.string(), coordinates: z.unknown() })
          .passthrough()
          .nullable()
          .optional(),
      }),
      organisation: named,
    }),
  ),
  route: z
    .array(
      z.object({
        facility_id: z.string(),
        name: z.string(),
        kind: z.string(),
        location: z
          .object({ type: z.string(), coordinates: z.unknown() })
          .passthrough()
          .nullable()
          .optional(),
        seq: z.number(),
      }),
    )
    .optional(),
  organisations: z.array(named),
  genealogy: z.object({
    nodes: z.array(passport),
    edges: z.array(
      z.object({
        parent_id: z.string(),
        child_id: z.string(),
        stage: z.string(),
        quantity_used: z.number(),
      }),
    ),
  }),
  certificates: z.array(
    z.object({
      id: z.string(),
      issuer_org_id: z.string(),
      subject_id: z.string(),
      type: z.string(),
      valid_from: z.string(),
      valid_to: z.string(),
      doc_hash: z.string(),
      status: z.enum(["valid", "expired", "revoked"]),
      revoke_reason: z.string().nullable(),
      tx_hash: z.string().nullable(),
    }),
  ),
});

// Dashboard lists (only the fields the dashboard reads).
const passportRow = z.object({
  id: z.string(),
  code: z.string(),
  commodity: z.string(),
  unit: z.string(),
  remaining: z.number(),
  status: z.string(),
  created_at: z.string(),
});
export type PassportRow = z.infer<typeof passportRow>;
const handoverRow = z.object({
  id: z.string(),
  passport_id: z.string(),
  from_org: z.string(),
  to_org: z.string(),
  stage: z.string(),
  quantity: z.number(),
  status: z.string(),
  created_at: z.string(),
});
export type HandoverRow = z.infer<typeof handoverRow>;
const facilityRow = z.object({
  id: z.string(),
  name: z.string(),
  capacity: z.number(),
  capacity_unit: z.string(),
  capacity_used: z.number(),
  status: z.string(),
});
export type FacilityRow = z.infer<typeof facilityRow>;
export const facilityKindSchema = z.object({
  key: z.string(),
  label: z.string(),
  sub: z.string(),
  active: z.boolean(),
});
export type FacilityKindRow = z.infer<typeof facilityKindSchema>;
const flagRow = z.object({
  id: z.string(),
  passport_id: z.string(),
  reason: z.string(),
  status: z.string(),
  created_at: z.string(),
});
export type FlagRow = z.infer<typeof flagRow>;
const certificateRow = z.object({
  id: z.string(),
  type: z.string(),
  status: z.string(),
  valid_to: z.string(),
  created_at: z.string(),
});
export type CertificateRow = z.infer<typeof certificateRow>;
const directoryRow = z.object({
  id: z.string(),
  name: z.string(),
  kind: z.string(),
  slug: z.string(),
});

// ---- Full records (the work screens) ----

const geometry = z
  .object({ type: z.string(), coordinates: z.unknown() })
  .passthrough()
  .nullable();

export const facilitySchema = facilityRow.extend({
  organisation_id: z.string(),
  kind: z.string(),
  location: geometry,
  area_ha: z.number(),
  licence_number: z.string(),
  licence_valid_to: z.string().nullable(),
  licence_document: z
    .object({ file_name: z.string() })
    .passthrough()
    .nullable(),
  capacity_period: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
});
export type Facility = z.infer<typeof facilitySchema>;

const traceEvent = z.object({
  id: z.string(),
  kind: z.string(),
  from_org: z.string().nullable(),
  to_org: z.string().nullable(),
  quantity: z.number(),
  data_hash: z.string(),
  tx_hash: z.string().nullable(),
  created_at: z.string(),
});
export type TraceEvent = z.infer<typeof traceEvent>;

export const passportDetailSchema = passportRow.extend({
  organisation_id: z.string(),
  facility_id: z.string().nullable(),
  transform_id: z.string().nullable(),
  template_id: z.string().optional(),
  quantity: z.number(),
  data_hash: z.string(),
  handover_count: z.number().optional(),
  tx_hash: z.string().nullable().optional(),
  created_at: z.string(),
  updated_at: z.string().optional(),
  events: z.array(traceEvent).optional(),
});
export type PassportDetail = z.infer<typeof passportDetailSchema>;

const location = z
  .object({ lat: z.number(), lon: z.number() })
  .nullable()
  .optional();
const handoverPhoto = z.object({
  id: z.string(),
  side: z.string(),
  file_name: z.string(),
  content_type: z.string(),
});
export type HandoverPhoto = z.infer<typeof handoverPhoto>;

export const handoverDetailSchema = handoverRow.extend({
  note: z.string(),
  received_quantity: z.number().nullable().optional(),
  reject_reasons: z.array(z.string()).nullable().optional(),
  child_passport_id: z.string().nullable().optional(),
  split_from_passport_id: z.string().nullable().optional(),
  location,
  receiver_location: location,
  photos: z
    .array(handoverPhoto)
    .nullable()
    .optional()
    .transform((p) => p ?? []),
  data_hash: z.string().optional(),
  deadline: z.string().optional(),
  signed_by_sender: z.boolean().optional(),
  signed_by_receiver: z.boolean().optional(),
  reject_reason: z.string().nullable().optional(),
  responded_at: z.string().nullable().optional(),
});
export type HandoverDetail = z.infer<typeof handoverDetailSchema>;

export const certificateSchema = z.object({
  id: z.string(),
  issuer_org_id: z.string(),
  subject_type: z.string(),
  subject_id: z.string(),
  type: z.string(),
  valid_from: z.string(),
  valid_to: z.string(),
  doc_hash: z.string(),
  file_name: z.string(),
  content_type: z.string(),
  size_bytes: z.number(),
  ipfs_cid: z.string().nullable().optional(),
  status: z.enum(["valid", "expired", "revoked"]),
  revoke_reason: z.string().nullable().optional(),
  revoked_at: z.string().nullable().optional(),
  tx_hash: z.string().nullable().optional(),
  created_at: z.string(),
  updated_at: z.string().optional(),
});
export type Certificate = z.infer<typeof certificateSchema>;

export const flagSchema = z.object({
  id: z.string(),
  passport_id: z.string(),
  organisation_id: z.string(),
  reason: z.string(),
  status: z.string(),
  resolution_note: z.string().nullable().optional(),
  resolved_at: z.string().nullable().optional(),
  tx_hash: z.string().nullable().optional(),
  created_at: z.string(),
  updated_at: z.string().optional(),
  revoked: z.number().optional(),
  raised_by: z.string().optional(),
  assignee_user_id: z.string().nullable().optional(),
  assignee_name: z.string().nullable().optional(),
});
export type Flag = z.infer<typeof flagSchema>;

export const genealogySchema = z.object({
  nodes: z.array(passportDetailSchema),
  edges: z.array(
    z.object({
      parent_id: z.string(),
      child_id: z.string(),
      stage: z.string(),
      quantity_used: z.number(),
    }),
  ),
});
export type Genealogy = z.infer<typeof genealogySchema>;

const auditEntry = z.object({
  id: z.string(),
  seq: z.number(),
  actor_user_id: z.string().nullable(),
  actor_email: z.string().nullable(),
  organisation_id: z.string().nullable(),
  method: z.string(),
  route: z.string(),
  subject_id: z.string().nullable(),
  status: z.number(),
  created_at: z.string(),
  prev_hash: z.string(),
  hash: z.string(),
});
export type AuditEntry = z.infer<typeof auditEntry>;

const volumeRow = z.object({
  month: z.string(),
  commodity: z.string(),
  unit: z.string(),
  quantity: z.number(),
  exported: z.number().optional(),
});
export type VolumeRow = z.infer<typeof volumeRow>;

const outcomeRow = z.object({
  period: z.string(),
  verified: z.number(),
  warning: z.number(),
  failed: z.number(),
  total: z.number(),
  pct_verified: z.number(),
});
export type OutcomeRow = z.infer<typeof outcomeRow>;

const savedReport = z.object({
  id: z.string(),
  name: z.string(),
  kind: z.string(),
  params: z.unknown().optional(),
  organisation_id: z.string().nullable().optional(),
  created_by: z.string().nullable().optional(),
  created_at: z.string(),
});
export type SavedReport = z.infer<typeof savedReport>;

const savedFile = z.object({
  id: z.string(),
  saved_id: z.string(),
  format: z.string(),
  bucket_key: z.string(),
  sha256: z.string(),
  size: z.number(),
  created_at: z.string(),
});
export type SavedFile = z.infer<typeof savedFile>;

const chainCheck = z.object({
  ok: z.boolean(),
  checked: z.number(),
  broken_at: z.number().nullable(),
});
export type ChainCheck = z.infer<typeof chainCheck>;

// Shell, profile and users.
const notificationRow = z.object({
  id: z.string(),
  organisation_id: z.string().nullable().optional(),
  kind: z.string(),
  subject_id: z.string(),
  message: z.string(),
  read_at: z.string().nullable(),
  created_at: z.string(),
});
export type NotificationRow = z.infer<typeof notificationRow>;
const myOrganisation = z.object({
  id: z.string(),
  name: z.string(),
  slug: z.string(),
  kind: z.string(),
  role: z.string(),
});
export type MyOrganisation = z.infer<typeof myOrganisation>;
const memberRow = z.object({
  id: z.string(),
  name: z.string(),
  email: z.string(),
  role: z.string(),
  status: z.string(),
  since: z.string(),
});
export type MemberRow = z.infer<typeof memberRow>;
const platformUserRow = z.object({
  id: z.string(),
  name: z.string(),
  email: z.string(),
  status: z.string(),
  two_factor: z.boolean().optional(),
  email_verified: z.boolean().optional(),
  created_at: z.string().optional(),
  memberships: z
    .array(
      z.object({
        organisation_id: z.string().nullable().optional(),
        organisation: z.string().nullable(),
        role: z.string(),
        status: z.string(),
      }),
    )
    .nullable()
    .transform((m) => m ?? []),
});
export type PlatformUserRow = z.infer<typeof platformUserRow>;
const twoFactorSetup = z.object({
  secret: z.string(),
  otpauth_url: z.string(),
});
const recoveryCodes = z.object({ recovery_codes: z.array(z.string()) });
const twoFactorSwitch = z.object({
  two_factor_enabled: z.boolean(),
  recovery_codes: z.array(z.string()).optional(),
});
const inviteResult = z.object({
  invite: z.object({
    id: z.string(),
    email: z.string(),
    role: z.string(),
    expires_at: z.string(),
  }),
  emailed: z.boolean(),
  link: z.string().optional(),
});
const acceptResult = z.object({
  organisation: z.string(),
  new_account: z.boolean(),
});

const nothing = z.unknown();
const all = `?size=${limits.pageSize}`;
/** A list endpoint taking page, search, sort and filters (lib/list.ts). */
/** The platform admin's ?organisation_id= (the API narrows its lists, and membership changes, to it). */
const inOrg = (orgId?: string) =>
  orgId ? `?organisation_id=${encodeURIComponent(orgId)}` : "";

const paged =
  <T extends z.ZodType>(path: string, item: T) =>
  (p: ListParams) =>
    api(path + listQuery(p), page(item));

// ---- Endpoints ----

export const endpoints = {
  /** Search map locations through this app's Photon proxy. */
  searchPlaces: (query: string, signal?: AbortSignal) =>
    api(
      `${apiPaths.geocode}?q=${encodeURIComponent(query)}`,
      placeSchema.array(),
      { signal },
    ),
  /** Who is signed in at this address, or null if nobody is. */
  me: () =>
    api(apiPaths.me, meSchema).catch((err) => {
      if (err instanceof ApiError && err.status === 401) return null;
      throw err;
    }),
  emailAvailability: (email: string) =>
    api(
      `${apiPaths.emailAvailability}?email=${encodeURIComponent(email)}`,
      z.object({ available: z.boolean() }),
    ).then((result) => result.available),
  login: (body: {
    email: string;
    password: string;
    code: string;
    remember?: boolean;
    trust_device?: boolean;
  }) => write(apiPaths.login, nothing, send("POST", body)),
  /** Always 202: the API never says whether the email has an account. */
  forgotPassword: (email: string) =>
    write(apiPaths.forgotPassword, nothing, send("POST", { email })),
  resetPassword: (token: string, password: string) =>
    write(
      apiPaths.resetPassword,
      nothing,
      send("POST", { token, new: password }),
    ),
  lockAccount: (token: string) =>
    write(apiPaths.lockAccount, nothing, send("POST", { token })),
  verifyEmail: (token: string) =>
    write(apiPaths.verifyEmail, nothing, send("POST", { token })),
  logout: () => write(apiPaths.logout, nothing, { method: "POST" }),
  /** A one-minute code that signs this person in at another organisation's address ("" = the bare one). */
  startHandoff: (slug: string) =>
    api(
      apiPaths.handoff,
      z.object({ code: z.string() }),
      send("POST", { slug }),
    ).then((r) => r.code),
  acceptHandoff: (code: string) =>
    write(apiPaths.handoffAccept, nothing, send("POST", { code })),
  register: (body: object) =>
    write(apiPaths.orgs, organisationSchema, send("POST", body)),

  org: (id: string) => api(apiPaths.org(id), organisationSchema),
  orgsPage: paged(apiPaths.orgs, organisationSchema),
  orgs: (status: string) =>
    api(
      `${apiPaths.orgs}${all}&sort=created_at&status=${status}`,
      page(organisationSchema),
    ).then((p) => p.data),
  documents: (orgId: string) =>
    api(
      `${apiPaths.orgDocuments(orgId)}${all}&sort=created_at`,
      page(documentSchema),
    ).then((p) => p.data),
  uploadDocument: (orgId: string, type: string, file: File) => {
    const body = new FormData();
    body.set("type", type);
    body.set("file", file);
    return write(apiPaths.orgDocuments(orgId), documentSchema, {
      method: "POST",
      body,
    });
  },
  review: (orgId: string, body: { decision: string; reason: string }) =>
    write(apiPaths.orgReview(orgId), organisationSchema, send("POST", body)),
  accreditations: (orgId: string) =>
    api(apiPaths.orgAccreditations(orgId), accreditationList),
  setAccreditations: (orgId: string, rows: AccreditationRow[]) =>
    write(
      apiPaths.orgAccreditations(orgId),
      accreditationList,
      send("PUT", rows),
    ),

  /** Every evidence key some template asks for: the certificate types an issuer can be accredited for. */
  certificateTypes: () =>
    api(`${apiPaths.templates}${all}`, page(templateSchema)).then((p) =>
      [
        ...new Set(
          p.data.flatMap((t) =>
            t.stages.flatMap((s) => s.requiredEvidence ?? []),
          ),
        ),
      ].sort(),
    ),

  // ---- Dashboard ----

  /** How many rows a list holds: one row is asked for and its `total` read. */
  count: (tile: Exclude<TileKey, "incoming" | "outgoing">) => {
    const [path, query] = countSources[tile];
    return api(`${path}?size=1&${query}`, z.object({ total: z.number() })).then(
      (p) => p.total,
    );
  },
  /** Handovers waiting for a receiver, to or from the acting organisation. One page; few are ever open at once. */
  awaitingHandovers: () =>
    api(
      `${apiPaths.handovers}${all}&status=${handoverStatus.awaiting}`,
      page(handoverRow),
    ).then((p) => p.data),
  passports: () =>
    api(
      `${apiPaths.passports}?size=${panelRows}&status=${passportStatus.active}`,
      page(passportRow),
    ).then((p) => p.data),
  facilities: () =>
    api(`${apiPaths.facilities}${all}`, page(facilityRow)).then((p) => p.data),
  openFlags: () =>
    api(
      `${apiPaths.flags}?size=${panelRows}&status=${flagStatus.open}`,
      page(flagRow),
    ).then((p) => p.data),
  certificates: () =>
    api(
      `${apiPaths.attestations}?size=${panelRows}`,
      page(certificateRow),
    ).then((p) => p.data),
  /** Names of verified organisations, to say who a handover is from or to. */
  directory: () =>
    api(`${apiPaths.directory}${all}`, page(directoryRow)).then(
      (p) => new Map(p.data.map((o) => [o.id, o.name])),
    ),
  /** Verified organisations other than mine, by name (draft 3g "Receiving organisation"). */
  directorySearch: (search: string) =>
    api(
      `${apiPaths.directory}?size=${panelRows * 2}&sort=name&search=${encodeURIComponent(search)}`,
      page(directoryRow),
    ).then((p) => p.data),

  // ---- Shell ----

  unreadCount: () =>
    api(
      `${apiPaths.notifications}?size=1&unread=true`,
      z.object({ total: z.number() }),
    ).then((p) => p.total),
  notifications: () =>
    api(
      `${apiPaths.notifications}?size=${panelRows * 4}`,
      page(notificationRow),
    ).then((p) => p.data),
  notificationsPage: paged(apiPaths.notifications, notificationRow),
  platformStats: () => api(apiPaths.stats, platformStatsSchema),
  markRead: (id: string) =>
    write(apiPaths.notificationRead(id), nothing, { method: "POST" }),
  myOrganisations: () => api(apiPaths.myOrganisations, z.array(myOrganisation)),

  // ---- Profile ----

  setName: (name: string) => write(apiPaths.me, nothing, send("PUT", { name })),
  changePassword: (body: { current: string; new: string }) =>
    write(apiPaths.password, nothing, send("POST", body)),
  setPin: (body: { pin: string; password: string }) =>
    write(apiPaths.pin, nothing, send("POST", body)),
  twoFactorSetup: () =>
    write(apiPaths.twoFactorSetup, twoFactorSetup, { method: "POST" }),
  twoFactorEnable: (code: string) =>
    write(apiPaths.twoFactorEnable, twoFactorSwitch, send("POST", { code })),
  twoFactorDisable: (code: string) =>
    write(apiPaths.twoFactorDisable, twoFactorSwitch, send("POST", { code })),
  newRecoveryCodes: (code: string) =>
    write(apiPaths.twoFactorRecovery, recoveryCodes, send("POST", { code })),

  // ---- Users ----

  members: paged(apiPaths.users, memberRow),
  allUsers: paged(apiPaths.users, platformUserRow),
  invite: (body: { email: string; role: string }) =>
    write(apiPaths.invites, inviteResult, send("POST", body)),
  // orgId: the platform admin names the organisation; an organisation admin acts for their own.
  setRole: (id: string, role: string, orgId?: string) =>
    write(apiPaths.userRole(id) + inOrg(orgId), nothing, send("PUT", { role })),
  suspend: (id: string, orgId?: string) =>
    write(apiPaths.userSuspend(id) + inOrg(orgId), nothing, { method: "POST" }),
  reinstate: (id: string, orgId?: string) =>
    write(apiPaths.userReinstate(id) + inOrg(orgId), nothing, {
      method: "POST",
    }),
  user: (id: string) => api(apiPaths.user(id), platformUserRow),
  deactivate: (id: string) =>
    write(apiPaths.userDeactivate(id), nothing, { method: "POST" }),
  activate: (id: string) =>
    write(apiPaths.userActivate(id), nothing, { method: "POST" }),
  orgMembers: (orgId: string) => paged(apiPaths.orgMembers(orgId), memberRow),
  orgSuspend: (id: string, reason: string) =>
    write(apiPaths.orgSuspend(id), nothing, send("POST", { reason })),
  orgReactivate: (id: string) =>
    write(apiPaths.orgReactivate(id), nothing, { method: "POST" }),
  acceptInvite: (token: string, body: { name: string; password: string }) =>
    write(apiPaths.inviteAccept(token), acceptResult, send("POST", body)),

  /** The public verification of one passport, or null if there is no such passport. */
  verify: (id: string) =>
    api(apiPaths.verify(id), verificationSchema).catch((err) => {
      if (err instanceof ApiError && err.status === 404) return null;
      throw err;
    }),

  // ---- Facilities ----

  facilitiesPage: paged(apiPaths.facilities, facilitySchema),
  facilityKinds: () =>
    api(
      apiPaths.facilityKinds,
      z.object({ data: z.array(facilityKindSchema) }),
    ).then((r) => r.data),
  facilityKindsAll: () =>
    api(
      apiPaths.facilityKindsAll,
      z.object({ data: z.array(facilityKindSchema) }),
    ).then((r) => r.data),
  createFacilityKind: (body: { key: string; label: string; sub: string }) =>
    write(apiPaths.facilityKinds, facilityKindSchema, send("POST", body)),
  updateFacilityKind: (
    key: string,
    body: { label: string; sub: string; active: boolean },
  ) => write(apiPaths.facilityKind(key), facilityKindSchema, send("PUT", body)),
  deleteFacilityKind: (key: string) =>
    write(apiPaths.facilityKind(key), nothing, {
      method: "DELETE",
    }),
  facilitiesList: () =>
    api(`${apiPaths.facilities}${all}&sort=name`, page(facilitySchema)).then(
      (p) => p.data,
    ),
  facility: (id: string) => api(apiPaths.facility(id), facilitySchema),
  createFacility: (body: {
    name: string;
    kind: string;
    location: unknown;
    licence_number: string;
    capacity: number;
    capacity_unit: string;
    capacity_period: string;
    licence_valid_to: string;
  }) => write(apiPaths.facilities, facilitySchema, send("POST", body)),
  approveFacility: (id: string) =>
    write(apiPaths.facilityApprove(id), facilitySchema, { method: "POST" }),
  uploadLicence: (id: string, file: File) => {
    const body = new FormData();
    body.set("file", file);
    return write(apiPaths.facilityLicence(id), facilitySchema, {
      method: "POST",
      body,
    });
  },
  setFacilityStatus: (id: string, status: string) =>
    write(
      apiPaths.facilityStatus(id),
      facilitySchema,
      send("POST", { status }),
    ),
  deleteFacility: (id: string) =>
    write(apiPaths.facility(id), nothing, { method: "DELETE" }),

  // ---- Templates ----

  templatesPage: paged(apiPaths.templates, templateSchema),
  templatesList: () =>
    api(`${apiPaths.templates}${all}&sort=name`, page(templateSchema)).then(
      (p) => p.data,
    ),
  template: (id: string) => api(apiPaths.template(id), templateSchema),
  createTemplate: (body: TemplateBody) =>
    write(
      apiPaths.templates,
      templateSchema,
      send("POST", withFractions(body)),
    ),
  updateTemplate: (id: string, body: TemplateBody) =>
    write(
      apiPaths.template(id),
      templateSchema,
      send("PUT", withFractions(body)),
    ),

  // ---- Passports ----

  passportsList: (status?: string) =>
    api(
      `${apiPaths.passports}${all}&sort=-created_at${status ? `&status=${status}` : ""}`,
      page(passportDetailSchema),
    ).then((p) => p.data),
  passportsPage: paged(apiPaths.passports, passportRow),
  passport: (id: string) => api(apiPaths.passport(id), passportDetailSchema),
  genealogy: (id: string, direction?: string) =>
    api(
      `${apiPaths.passportGenealogy(id)}${direction ? `?direction=${direction}` : ""}`,
      genealogySchema,
    ),
  issuePassport: (
    body: { facility_id: string; template_id: string; quantity: number },
    pin: string,
  ) => write(apiPaths.passports, passportDetailSchema, send("POST", body, pin)),
  retirePassport: (id: string, pin: string) =>
    write(
      apiPaths.passportRetire(id),
      passportDetailSchema,
      send("POST", {}, pin),
    ),

  // ---- Handovers ----

  handoversPage: paged(apiPaths.handovers, handoverDetailSchema),
  handoversList: (status?: string) =>
    api(
      `${apiPaths.handovers}${all}&sort=-created_at${status ? `&status=${status}` : ""}`,
      page(handoverDetailSchema),
    ).then((p) => p.data),
  handover: (id: string) => api(apiPaths.handover(id), handoverDetailSchema),
  /** Starts a handover; less than the whole passport splits it first (the API does that). Photos go as multipart. */
  startHandover: (
    passportId: string,
    body: HandoverStart,
    photos: File[],
    pin: string,
  ) =>
    write(
      apiPaths.passportHandovers(passportId),
      handoverDetailSchema,
      multipart(body, photos, pin),
    ),
  acceptHandover: (
    id: string,
    body: { received_quantity: number; lat?: number; lon?: number },
    photos: File[],
    pin: string,
  ) =>
    write(
      apiPaths.handoverAccept(id),
      z.unknown(),
      multipart(body, photos, pin),
    ),
  rejectHandover: (
    id: string,
    body: { reasons: string[]; note: string; also_flag: boolean },
    pin: string,
  ) =>
    write(
      apiPaths.handoverReject(id),
      handoverDetailSchema,
      send("POST", body, pin),
    ),
  cancelHandover: (id: string) =>
    write(apiPaths.handoverCancel(id), handoverDetailSchema, send("POST", {})),

  // ---- Transform ----

  transform: (
    body: {
      stage: string;
      inputs: { passport_id: string; quantity: number }[];
      outputs: { quantity: number }[];
    },
    pin: string,
  ) =>
    write(
      apiPaths.transforms,
      z
        .object({ outputs: z.array(passportDetailSchema) })
        .transform((t) => t.outputs),
      send("POST", body, pin),
    ),

  // ---- Certificates ----

  attestationsPage: paged(apiPaths.attestations, certificateSchema),
  attestationsList: (subjectId?: string, status?: string) =>
    api(
      `${apiPaths.attestations}${all}&sort=-created_at` +
        `${subjectId ? `&subject_id=${encodeURIComponent(subjectId)}` : ""}${status ? `&status=${status}` : ""}`,
      page(certificateSchema),
    ).then((p) => p.data),
  attestation: (id: string) => api(apiPaths.attestation(id), certificateSchema),
  /** Issues a certificate: the signed document travels with the form (POST /attestations is multipart). */
  issueAttestation: (body: FormData, pin: string) =>
    write(apiPaths.attestations, certificateSchema, {
      method: "POST",
      body,
      headers: { "x-signing-pin": pin },
    }),
  revokeAttestation: (id: string, reason: string, pin: string) =>
    write(
      apiPaths.attestationRevoke(id),
      certificateSchema,
      send("POST", { reason }, pin),
    ),

  // ---- Flags ----

  flagsPage: paged(apiPaths.flags, flagSchema),
  flagsList: (status?: string) =>
    api(
      `${apiPaths.flags}${all}&sort=-created_at${status ? `&status=${status}` : ""}`,
      page(flagSchema),
    ).then((p) => p.data),
  flag: (id: string) => api(apiPaths.flag(id), flagSchema),
  raiseFlag: (body: { passport_id: string; reason: string }) =>
    write(apiPaths.flags, flagSchema, send("POST", body)),
  /** Investigators only: give the flag to a person, or to nobody (null). */
  assignFlag: (id: string, userId: string | null) =>
    write(
      apiPaths.flagAssignee(id),
      flagSchema,
      send("PUT", { user_id: userId }),
    ),
  resolveFlag: (
    id: string,
    body: { outcome: string; note: string },
    pin: string,
  ) => write(apiPaths.flagResolve(id), flagSchema, send("POST", body, pin)),

  // ---- Audit and reports ----

  auditPage: paged(apiPaths.audit, auditEntry),
  auditVerify: () => api(apiPaths.auditVerify, chainCheck),
  volumeReport: (params: { startDate?: string; endDate?: string } = {}) => {
    const q = new URLSearchParams();
    if (params.startDate) q.set("start_date", params.startDate);
    if (params.endDate) q.set("end_date", params.endDate);
    const suffix = q.size ? `?${q}` : "";
    return api(
      `${apiPaths.reportsVolume}${suffix}`,
      z.object({ data: z.array(volumeRow) }),
    ).then((r) => r.data);
  },
  outcomesReport: (
    params: { startDate?: string; endDate?: string; commodity?: string } = {},
  ) => {
    const q = new URLSearchParams();
    if (params.startDate) q.set("start_date", params.startDate);
    if (params.endDate) q.set("end_date", params.endDate);
    if (params.commodity) q.set("commodity", params.commodity);
    const suffix = q.size ? `?${q}` : "";
    return api(
      `${apiPaths.reportsOutcomes}${suffix}`,
      z.object({ data: z.array(outcomeRow) }),
    ).then((r) => r.data);
  },
  savedReports: () =>
    api(apiPaths.reportsSaved, z.object({ data: z.array(savedReport) })).then(
      (r) => r.data,
    ) as Promise<SavedReport[]>,
  createSaved: (body: {
    name: string;
    kind: string;
    params: Record<string, string>;
  }) => write(apiPaths.reportsSaved, savedReport, send("POST", body)),
  runSaved: (id: string, format: string = "csv") =>
    write(`${apiPaths.reportRun(id)}?format=${format}`, savedFile, {
      method: "POST",
    }),
  savedFiles: (id: string) =>
    api(apiPaths.reportFiles(id), z.object({ data: z.array(savedFile) })).then(
      (r) => r.data,
    ),
};

// Where each counted tile's number comes from: a list and its filter.
const countSources: Record<
  Exclude<TileKey, "incoming" | "outgoing">,
  [string, string]
> = {
  passportsHeld: [apiPaths.passports, `status=${passportStatus.active}`],
  openFlags: [apiPaths.flags, `status=${flagStatus.open}`],
  certValid: [apiPaths.attestations, `status=${certificateStatus.valid}`],
  certExpired: [apiPaths.attestations, `status=${certificateStatus.expired}`],
  certRevoked: [apiPaths.attestations, `status=${certificateStatus.revoked}`],
  orgsPending: [apiPaths.orgs, `status=${orgStatus.pending}`],
  orgsChanges: [apiPaths.orgs, `status=${orgStatus.changesRequested}`],
  orgsRejected: [apiPaths.orgs, `status=${orgStatus.rejected}`],
};

/** A template as the forms send it: maximum yields in percent. */
type TemplateBody = { stages: { maxYield?: number | null }[] };
const withFractions = <B extends TemplateBody>(b: B): B => ({
  ...b,
  stages: b.stages.map((st) => ({
    ...st,
    maxYield: st.maxYield == null ? st.maxYield : yieldToFraction(st.maxYield),
  })),
});

const countRow = z.object({ key: z.string(), count: z.number() });
export const platformStatsSchema = z.object({
  organisations: z.array(countRow),
  organisation_kinds: z.array(countRow),
  users: z.number(),
  facilities: z.array(countRow),
  passports: z.array(countRow),
  certificates: z.array(countRow),
  flags: z.array(countRow),
  activity: z.array(
    z.object({
      date: z.string(),
      passports: z.number(),
      handovers: z.number(),
      transforms: z.number(),
      flags: z.number(),
      signups: z.number(),
    }),
  ),
  volume: z.array(
    z.object({
      commodity: z.string(),
      unit: z.string(),
      quantity: z.number(),
      lots: z.number(),
    }),
  ),
  top_organisations: z.array(
    z.object({
      id: z.string(),
      name: z.string(),
      quantity: z.number(),
      lots: z.number(),
    }),
  ),
  outbox: z.object({
    pending: z.number(),
    failed: z.number(),
    oldest_pending_seconds: z.number().nullable(),
  }),
  generated_at: z.string(),
});
export type PlatformStats = z.infer<typeof platformStatsSchema>;

export type HandoverStart = {
  to_org: string;
  stage: string;
  quantity: number;
  note: string;
  lat?: number;
  lon?: number;
};

/** Form fields plus photos, with the signing PIN header (the browser sets the multipart boundary). */
function multipart(
  fields: Record<string, string | number | undefined>,
  photos: File[],
  pin: string,
): RequestInit {
  const body = new FormData();
  for (const [k, v] of Object.entries(fields))
    if (v !== undefined) body.set(k, String(v));
  for (const f of photos) body.append(handoverPhotoField, f);
  return { method: "POST", body, headers: { "x-signing-pin": pin } };
}

export const handoverPhotoURL = (handoverId: string, photoId: string) =>
  apiBase + apiPaths.handoverPhoto(handoverId, photoId);

/** A report download through the pass-through: format plus the filters that are set. */
export function downloadURL(
  path: string,
  format: string,
  params: Record<string, string>,
): string {
  const q = new URLSearchParams({ format });
  for (const [k, v] of Object.entries(params)) if (v) q.set(k, v);
  return `${apiBase}${path}?${q}`;
}

export const documentFileURL = (id: string) =>
  apiBase + apiPaths.documentFile(id);

export const facilityLicenceURL = (id: string) =>
  apiBase + apiPaths.facilityLicence(id);

/**
 * Opens an organisation's address (null: the bare one, the workspace picker)
 * with this session carried over: a one-time code made here, spent there.
 */
export async function switchOrganisation(
  slug: string | null,
  currentSlug?: string,
): Promise<void> {
  const code = await endpoints.startHandoff(slug ?? "");
  window.location.assign(
    orgAddress(slug, window.location, currentSlug, routes.handoff(code)),
  );
}

// ---- Hooks ----

/** One page of a list; the previous page stays on screen while the next loads. */
export function usePaged<T>(
  key: readonly unknown[],
  fetch: (p: ListParams) => Promise<Paged<T>>,
  params: ListParams,
) {
  return useQuery({
    queryKey: [...key, "page", params],
    queryFn: () => fetch(params),
    placeholderData: keepPreviousData,
  });
}

export function useMe() {
  return useQuery({ queryKey: queryKeys.me, queryFn: endpoints.me });
}

/**
 * Who is signed in, sending them elsewhere when this page isn't for them.
 * `where` gets the signed-in person (or null) and returns the page they
 * belong on, or undefined to stay. `ready` is false until they may see this page.
 */
export function useMeGuard(where: (me: Me | null) => string | undefined) {
  const router = useRouter();
  const me = useMe();
  const target = me.isSuccess ? where(me.data) : undefined;
  useEffect(() => {
    if (target) router.replace(target);
  }, [router, target]);
  return { ...me, ready: me.isSuccess && !target };
}
