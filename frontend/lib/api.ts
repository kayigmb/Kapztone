// The browser's access to the API. Every call goes through this site's
// pass-through (app/api/[...path]/route.ts), so the session cookie and the
// organisation in the address travel with it. Every answer is checked
// against a zod schema; every failure becomes one ApiError.
import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { z } from "zod";

import {
  apiBase, apiPaths, certificateStatus, handoverPhotoField, copy, flagStatus, handoverStatus, limits, orgStatus, panelRows, passportStatus, queryKeys,
  type TileKey,
} from "@/lib/consts";
import type { ApiVerification } from "@/lib/verification";

// ---- Errors ----

const errorBody = z.object({
  error: z.string().catch(""),
  two_factor_required: z.boolean().optional(),
  attempts_left: z.number().optional(),
}).catch({ error: "" });

/** One kind of failure for every call. `status` 0 means the API could not be reached. */
export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
    readonly twoFactorRequired = false,
    /** Failed sign-ins left before the pause (POST /auth/login 401 only). */
    readonly attemptsLeft?: number,
  ) {
    super(message);
  }
}

/** What to tell a person about a failure they can't fix by changing a field. */
export function problemOf(err: unknown): string {
  if (err instanceof ApiError) {
    if (err.status === 0 || err.status === 502) return copy.errors.unreachable; // 502: the pass-through couldn't reach the API
    return err.status >= 500 || !err.message ? copy.errors.unexpected : err.message;
  }
  if (err instanceof z.ZodError) return copy.errors.badAnswer;
  return copy.errors.unexpected;
}

// ---- Calls ----

async function call(path: string, init?: RequestInit): Promise<Response> {
  let res: Response;
  try {
    res = await fetch(apiBase + path, { ...init, cache: "no-store" });
  } catch {
    throw new ApiError(0, copy.errors.unreachable);
  }
  if (!res.ok) {
    const body = errorBody.parse(await res.json().catch(() => ({})));
    throw new ApiError(res.status, body.error, body.two_factor_required, body.attempts_left);
  }
  return res;
}

/** Calls the API and checks the answer's shape. */
export async function api<T>(path: string, schema: z.ZodType<T>, init?: RequestInit): Promise<T> {
  const res = await call(path, init);
  return schema.parse(res.status === 204 ? undefined : await res.json());
}

const send = (method: string, body: unknown, pin?: string): RequestInit => ({
  method,
  headers: pin ? { "content-type": "application/json", "x-signing-pin": pin } : { "content-type": "application/json" },
  body: JSON.stringify(body),
});

// ---- Shapes (backend JSON) ----

const orgSummary = z.object({ id: z.string(), name: z.string(), slug: z.string(), kind: z.string(), status: z.string() });

export const meSchema = z.object({
  user: z.object({ id: z.string(), email: z.string(), name: z.string() }),
  organisation: orgSummary.nullable(),
  role: z.string(),
  permissions: z.array(z.string()).nullable().transform((p) => p ?? []),
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

const page = <T extends z.ZodType>(item: T) => z.object({ data: z.array(item), total: z.number() });

const stage = z.object({
  key: z.string(),
  events: z.array(z.string()).optional(),
  maxYield: z.number().nullable().optional(),
  requiredEvidence: z.array(z.string()).optional(),
  handoverTolerance: z.number().nullable().optional(),
});
const templateSchema = z.object({
  id: z.string(),
  commodity: z.string(),
  name: z.string(),
  unit: z.string(),
  originFacilityKind: z.string().nullable().optional(),
  capacityRule: z.object({ type: z.string(), kgPerHaPerSeason: z.number().nullable().optional() }).nullable().optional(),
  stages: z.array(stage),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
});
export type Template = z.infer<typeof templateSchema>;

export const accreditationSchema = z.object({ type: z.string(), valid_from: z.string(), valid_to: z.string() });
export type AccreditationRow = z.infer<typeof accreditationSchema>;
// A bare list, or the usual paged envelope: either is read the same way.
const accreditationList = z.union([z.array(accreditationSchema), page(accreditationSchema).transform((p) => p.data)]);

const verdict = z.enum(["verified", "warning", "failed"]);
const named = z.object({ id: z.string(), name: z.string() });
const passport = z.object({
  id: z.string(), code: z.string().optional(), organisation_id: z.string(), facility_id: z.string().nullable(), transform_id: z.string().nullable(),
  commodity: z.string(), unit: z.string(), quantity: z.number(), remaining: z.number(), status: z.string(),
  data_hash: z.string(), tx_hash: z.string().nullable(), created_at: z.string(),
  events: z.array(z.object({
    id: z.string(), kind: z.string(), from_org: z.string().nullable(), to_org: z.string().nullable(),
    quantity: z.number(), tx_hash: z.string().nullable(), created_at: z.string(),
  })).optional(),
});
const verificationSchema: z.ZodType<ApiVerification> = z.object({
  verdict,
  reasons: z.array(z.object({ level: verdict, check: z.string(), code: z.string(), message: z.string(), subject_id: z.string() })),
  checks: z.array(z.object({ name: z.string(), result: verdict })),
  passport,
  origins: z.array(z.object({ facility: named.extend({ kind: z.string() }), organisation: named })),
  organisations: z.array(named),
  genealogy: z.object({
    nodes: z.array(passport),
    edges: z.array(z.object({ parent_id: z.string(), child_id: z.string(), stage: z.string(), quantity_used: z.number() })),
  }),
  certificates: z.array(z.object({
    id: z.string(), issuer_org_id: z.string(), subject_id: z.string(), type: z.string(),
    valid_from: z.string(), valid_to: z.string(), doc_hash: z.string(),
    status: z.enum(["valid", "expired", "revoked"]), revoke_reason: z.string().nullable(), tx_hash: z.string().nullable(),
  })),
});

// Dashboard lists (only the fields the dashboard reads).
const passportRow = z.object({
  id: z.string(), code: z.string(), commodity: z.string(), unit: z.string(), remaining: z.number(), status: z.string(), created_at: z.string(),
});
export type PassportRow = z.infer<typeof passportRow>;
const handoverRow = z.object({
  id: z.string(), passport_id: z.string(), from_org: z.string(), to_org: z.string(), stage: z.string(),
  quantity: z.number(), status: z.string(), created_at: z.string(),
});
export type HandoverRow = z.infer<typeof handoverRow>;
const facilityRow = z.object({
  id: z.string(), name: z.string(), capacity: z.number(), capacity_unit: z.string(), capacity_used: z.number(), status: z.string(),
});
export type FacilityRow = z.infer<typeof facilityRow>;
const flagRow = z.object({ id: z.string(), passport_id: z.string(), reason: z.string(), status: z.string(), created_at: z.string() });
export type FlagRow = z.infer<typeof flagRow>;
const certificateRow = z.object({ id: z.string(), type: z.string(), status: z.string(), valid_to: z.string(), created_at: z.string() });
export type CertificateRow = z.infer<typeof certificateRow>;
const directoryRow = z.object({ id: z.string(), name: z.string(), kind: z.string(), slug: z.string() });

// ---- Full records (the work screens) ----

const geometry = z.object({ type: z.string(), coordinates: z.unknown() }).passthrough().nullable();

export const facilitySchema = facilityRow.extend({
  organisation_id: z.string(),
  kind: z.string(),
  location: geometry,
  area_ha: z.number(),
  licence_number: z.string(),
  licence_valid_to: z.string().nullable(),
  licence_document: z.object({ file_name: z.string() }).passthrough().nullable(),
  capacity_period: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
});
export type Facility = z.infer<typeof facilitySchema>;

const traceEvent = z.object({
  id: z.string(), kind: z.string(), from_org: z.string().nullable(), to_org: z.string().nullable(),
  quantity: z.number(), data_hash: z.string(), tx_hash: z.string().nullable(), created_at: z.string(),
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

const location = z.object({ lat: z.number(), lon: z.number() }).nullable().optional();
const handoverPhoto = z.object({ id: z.string(), side: z.string(), file_name: z.string(), content_type: z.string() });
export type HandoverPhoto = z.infer<typeof handoverPhoto>;

export const handoverDetailSchema = handoverRow.extend({
  note: z.string(),
  received_quantity: z.number().nullable().optional(),
  reject_reasons: z.array(z.string()).nullable().optional(),
  child_passport_id: z.string().nullable().optional(),
  split_from_passport_id: z.string().nullable().optional(),
  location,
  receiver_location: location,
  photos: z.array(handoverPhoto).nullable().optional().transform((p) => p ?? []),
  data_hash: z.string().optional(),
  deadline: z.string().optional(),
  signed_by_sender: z.boolean().optional(),
  signed_by_receiver: z.boolean().optional(),
  reject_reason: z.string().nullable().optional(),
  responded_at: z.string().nullable().optional(),
});
export type HandoverDetail = z.infer<typeof handoverDetailSchema>;

export const certificateSchema = z.object({
  id: z.string(), issuer_org_id: z.string(), subject_type: z.string(), subject_id: z.string(), type: z.string(),
  valid_from: z.string(), valid_to: z.string(), doc_hash: z.string(), file_name: z.string(),
  content_type: z.string(), size_bytes: z.number(), ipfs_cid: z.string().nullable().optional(),
  status: z.enum(["valid", "expired", "revoked"]), revoke_reason: z.string().nullable().optional(),
  revoked_at: z.string().nullable().optional(), tx_hash: z.string().nullable().optional(),
  created_at: z.string(), updated_at: z.string().optional(),
});
export type Certificate = z.infer<typeof certificateSchema>;

export const flagSchema = z.object({
  id: z.string(), passport_id: z.string(), organisation_id: z.string(), reason: z.string(),
  status: z.string(), resolution_note: z.string().nullable().optional(), resolved_at: z.string().nullable().optional(),
  tx_hash: z.string().nullable().optional(), created_at: z.string(), updated_at: z.string().optional(),
  revoked: z.number().optional(),
  raised_by: z.string().optional(),
  assignee_user_id: z.string().nullable().optional(),
  assignee_name: z.string().nullable().optional(),
});
export type Flag = z.infer<typeof flagSchema>;

export const genealogySchema = z.object({
  nodes: z.array(passportDetailSchema),
  edges: z.array(z.object({ parent_id: z.string(), child_id: z.string(), stage: z.string(), quantity_used: z.number() })),
});
export type Genealogy = z.infer<typeof genealogySchema>;

const auditEntry = z.object({
  id: z.string(), seq: z.number(), actor_user_id: z.string().nullable(), actor_email: z.string().nullable(),
  organisation_id: z.string().nullable(), method: z.string(), route: z.string(), subject_id: z.string().nullable(),
  status: z.number(), created_at: z.string(), prev_hash: z.string(), hash: z.string(),
});
export type AuditEntry = z.infer<typeof auditEntry>;

const volumeRow = z.object({ month: z.string(), commodity: z.string(), unit: z.string(), quantity: z.number() });
export type VolumeRow = z.infer<typeof volumeRow>;

const chainCheck = z.object({ ok: z.boolean(), checked: z.number(), broken_at: z.number().nullable() });
export type ChainCheck = z.infer<typeof chainCheck>;

// Shell, profile and users.
const notificationRow = z.object({
  id: z.string(), organisation_id: z.string().nullable().optional(), kind: z.string(), subject_id: z.string(),
  message: z.string(), read_at: z.string().nullable(), created_at: z.string(),
});
export type NotificationRow = z.infer<typeof notificationRow>;
const myOrganisation = z.object({ id: z.string(), name: z.string(), slug: z.string(), kind: z.string(), role: z.string() });
export type MyOrganisation = z.infer<typeof myOrganisation>;
const memberRow = z.object({
  id: z.string(), name: z.string(), email: z.string(), role: z.string(), status: z.string(), since: z.string(),
});
export type MemberRow = z.infer<typeof memberRow>;
const platformUserRow = z.object({
  id: z.string(), name: z.string(), email: z.string(), status: z.string(), two_factor: z.boolean().optional(),
  memberships: z.array(z.object({
    organisation_id: z.string().nullable().optional(), organisation: z.string().nullable(), role: z.string(), status: z.string(),
  })).nullable().transform((m) => m ?? []),
});
export type PlatformUserRow = z.infer<typeof platformUserRow>;
const twoFactorSetup = z.object({ secret: z.string(), otpauth_url: z.string() });
const recoveryCodes = z.object({ recovery_codes: z.array(z.string()) });
const twoFactorSwitch = z.object({ two_factor_enabled: z.boolean(), recovery_codes: z.array(z.string()).optional() });
const inviteResult = z.object({
  invite: z.object({ id: z.string(), email: z.string(), role: z.string(), expires_at: z.string() }),
  emailed: z.boolean(),
  link: z.string().optional(),
});
const acceptResult = z.object({ organisation: z.string(), new_account: z.boolean() });

const nothing = z.unknown();
const all = `?size=${limits.pageSize}`;

// ---- Endpoints ----

export const endpoints = {
  /** Who is signed in at this address, or null if nobody is. */
  me: () => api(apiPaths.me, meSchema).catch((err) => {
    if (err instanceof ApiError && err.status === 401) return null;
    throw err;
  }),
  login: (body: { email: string; password: string; code: string; remember?: boolean; trust_device?: boolean }) =>
    api(apiPaths.login, nothing, send("POST", body)),
  /** Always 202: the API never says whether the email has an account. */
  forgotPassword: (email: string) => api(apiPaths.forgotPassword, nothing, send("POST", { email })),
  resetPassword: (token: string, password: string) => api(apiPaths.resetPassword, nothing, send("POST", { token, new: password })),
  lockAccount: (token: string) => api(apiPaths.lockAccount, nothing, send("POST", { token })),
  logout: () => api(apiPaths.logout, nothing, { method: "POST" }),
  register: (body: object) => api(apiPaths.orgs, organisationSchema, send("POST", body)),

  org: (id: string) => api(apiPaths.org(id), organisationSchema),
  orgs: (status: string) => api(`${apiPaths.orgs}${all}&sort=created_at&status=${status}`, page(organisationSchema)).then((p) => p.data),
  documents: (orgId: string) => api(`${apiPaths.orgDocuments(orgId)}${all}&sort=created_at`, page(documentSchema)).then((p) => p.data),
  uploadDocument: (orgId: string, type: string, file: File) => {
    const body = new FormData();
    body.set("type", type);
    body.set("file", file);
    return api(apiPaths.orgDocuments(orgId), documentSchema, { method: "POST", body });
  },
  review: (orgId: string, body: { decision: string; reason: string }) => api(apiPaths.orgReview(orgId), organisationSchema, send("POST", body)),
  accreditations: (orgId: string) => api(apiPaths.orgAccreditations(orgId), accreditationList),
  setAccreditations: (orgId: string, rows: AccreditationRow[]) => api(apiPaths.orgAccreditations(orgId), accreditationList, send("PUT", rows)),

  /** Every evidence key some template asks for: the certificate types an issuer can be accredited for. */
  certificateTypes: () => api(`${apiPaths.templates}${all}`, page(templateSchema)).then((p) =>
    [...new Set(p.data.flatMap((t) => t.stages.flatMap((s) => s.requiredEvidence ?? [])))].sort()),

  // ---- Dashboard ----

  /** How many rows a list holds: one row is asked for and its `total` read. */
  count: (tile: Exclude<TileKey, "incoming" | "outgoing">) => {
    const [path, query] = countSources[tile];
    return api(`${path}?size=1&${query}`, z.object({ total: z.number() })).then((p) => p.total);
  },
  /** Handovers waiting for a receiver, to or from the acting organisation. ponytail: one page; few are ever open at once. */
  awaitingHandovers: () =>
    api(`${apiPaths.handovers}${all}&status=${handoverStatus.awaiting}`, page(handoverRow)).then((p) => p.data),
  passports: () => api(`${apiPaths.passports}?size=${panelRows}&status=${passportStatus.active}`, page(passportRow)).then((p) => p.data),
  facilities: () => api(`${apiPaths.facilities}${all}`, page(facilityRow)).then((p) => p.data),
  openFlags: () => api(`${apiPaths.flags}?size=${panelRows}&status=${flagStatus.open}`, page(flagRow)).then((p) => p.data),
  certificates: () => api(`${apiPaths.attestations}?size=${panelRows}`, page(certificateRow)).then((p) => p.data),
  /** Names of verified organisations, to say who a handover is from or to. */
  directory: () => api(`${apiPaths.directory}${all}`, page(directoryRow)).then((p) => new Map(p.data.map((o) => [o.id, o.name]))),
  /** Verified organisations other than mine, by name (draft 3g "Receiving organisation"). */
  directorySearch: (search: string) =>
    api(`${apiPaths.directory}?size=${panelRows * 2}&sort=name&search=${encodeURIComponent(search)}`, page(directoryRow)).then((p) => p.data),

  // ---- Shell ----

  unreadCount: () => api(`${apiPaths.notifications}?size=1&unread=true`, z.object({ total: z.number() })).then((p) => p.total),
  notifications: () => api(`${apiPaths.notifications}?size=${panelRows * 4}`, page(notificationRow)).then((p) => p.data),
  markRead: (id: string) => api(apiPaths.notificationRead(id), nothing, { method: "POST" }),
  myOrganisations: () => api(apiPaths.myOrganisations, z.array(myOrganisation)),

  // ---- Profile ----

  setName: (name: string) => api(apiPaths.me, nothing, send("PUT", { name })),
  changePassword: (body: { current: string; new: string }) => api(apiPaths.password, nothing, send("POST", body)),
  setPin: (body: { pin: string; password: string }) => api(apiPaths.pin, nothing, send("POST", body)),
  twoFactorSetup: () => api(apiPaths.twoFactorSetup, twoFactorSetup, { method: "POST" }),
  twoFactorEnable: (code: string) => api(apiPaths.twoFactorEnable, twoFactorSwitch, send("POST", { code })),
  twoFactorDisable: (code: string) => api(apiPaths.twoFactorDisable, twoFactorSwitch, send("POST", { code })),
  newRecoveryCodes: (code: string) => api(apiPaths.twoFactorRecovery, recoveryCodes, send("POST", { code })),

  // ---- Users ----

  members: () => api(apiPaths.users, z.array(memberRow)),
  allUsers: () => api(apiPaths.users, z.array(platformUserRow)),
  invite: (body: { email: string; role: string }) => api(apiPaths.invites, inviteResult, send("POST", body)),
  setRole: (id: string, role: string) => api(apiPaths.userRole(id), nothing, send("PUT", { role })),
  suspend: (id: string) => api(apiPaths.userSuspend(id), nothing, { method: "POST" }),
  reinstate: (id: string) => api(apiPaths.userReinstate(id), nothing, { method: "POST" }),
  acceptInvite: (token: string, body: { name: string; password: string }) =>
    api(apiPaths.inviteAccept(token), acceptResult, send("POST", body)),

  /** The public verification of one passport, or null if there is no such passport. */
  verify: (id: string) => api(apiPaths.verify(id), verificationSchema).catch((err) => {
    if (err instanceof ApiError && err.status === 404) return null;
    throw err;
  }),

  // ---- Facilities ----

  facilitiesList: () => api(`${apiPaths.facilities}${all}&sort=name`, page(facilitySchema)).then((p) => p.data),
  facility: (id: string) => api(apiPaths.facility(id), facilitySchema),
  createFacility: (body: {
    name: string; kind: string; location: unknown; licence_number: string;
    capacity: number; capacity_unit: string; capacity_period: string; licence_valid_to: string;
  }) => api(apiPaths.facilities, facilitySchema, send("POST", body)),
  approveFacility: (id: string) => api(apiPaths.facilityApprove(id), facilitySchema, { method: "POST" }),
  uploadLicence: (id: string, file: File) => {
    const body = new FormData();
    body.set("file", file);
    return api(apiPaths.facilityLicence(id), facilitySchema, { method: "POST", body });
  },
  setFacilityStatus: (id: string, status: string) => api(apiPaths.facilityStatus(id), facilitySchema, send("POST", { status })),
  deleteFacility: (id: string) => api(apiPaths.facility(id), nothing, { method: "DELETE" }),

  // ---- Templates ----

  templatesList: () => api(`${apiPaths.templates}${all}&sort=name`, page(templateSchema)).then((p) => p.data),
  template: (id: string) => api(apiPaths.template(id), templateSchema),
  createTemplate: (body: object) => api(apiPaths.templates, templateSchema, send("POST", body)),
  updateTemplate: (id: string, body: object) => api(apiPaths.template(id), templateSchema, send("PUT", body)),

  // ---- Passports ----

  passportsList: (status?: string) =>
    api(`${apiPaths.passports}${all}&sort=-created_at${status ? `&status=${status}` : ""}`, page(passportDetailSchema)).then((p) => p.data),
  passport: (id: string) => api(apiPaths.passport(id), passportDetailSchema),
  genealogy: (id: string, direction?: string) =>
    api(`${apiPaths.passportGenealogy(id)}${direction ? `?direction=${direction}` : ""}`, genealogySchema),
  issuePassport: (body: { facility_id: string; template_id: string; quantity: number }, pin: string) =>
    api(apiPaths.passports, passportDetailSchema, send("POST", body, pin)),
  retirePassport: (id: string, pin: string) => api(apiPaths.passportRetire(id), passportDetailSchema, send("POST", {}, pin)),

  // ---- Handovers ----

  handoversList: (status?: string) =>
    api(`${apiPaths.handovers}${all}&sort=-created_at${status ? `&status=${status}` : ""}`, page(handoverDetailSchema)).then((p) => p.data),
  handover: (id: string) => api(apiPaths.handover(id), handoverDetailSchema),
  /** Starts a handover; less than the whole passport splits it first (the API does that). Photos go as multipart. */
  startHandover: (passportId: string, body: HandoverStart, photos: File[], pin: string) =>
    api(apiPaths.passportHandovers(passportId), handoverDetailSchema, multipart(body, photos, pin)),
  acceptHandover: (id: string, body: { received_quantity: number; lat?: number; lon?: number }, photos: File[], pin: string) =>
    api(apiPaths.handoverAccept(id), z.unknown(), multipart(body, photos, pin)),
  rejectHandover: (id: string, body: { reasons: string[]; note: string; also_flag: boolean }, pin: string) =>
    api(apiPaths.handoverReject(id), handoverDetailSchema, send("POST", body, pin)),
  cancelHandover: (id: string) => api(apiPaths.handoverCancel(id), handoverDetailSchema, send("POST", {})),

  // ---- Transform ----

  transform: (body: {
    stage: string;
    inputs: { passport_id: string; quantity: number }[];
    outputs: { quantity: number }[];
  }, pin: string) => api(apiPaths.transforms, z.array(passportDetailSchema), send("POST", body, pin)),

  // ---- Certificates ----

  attestationsList: (subjectId?: string, status?: string) =>
    api(`${apiPaths.attestations}${all}&sort=-created_at`
      + `${subjectId ? `&subject_id=${encodeURIComponent(subjectId)}` : ""}${status ? `&status=${status}` : ""}`,
    page(certificateSchema)).then((p) => p.data),
  attestation: (id: string) => api(apiPaths.attestation(id), certificateSchema),
  /** Issues a certificate: the signed document travels with the form (POST /attestations is multipart). */
  issueAttestation: (body: FormData, pin: string) =>
    api(apiPaths.attestations, certificateSchema, { method: "POST", body, headers: { "x-signing-pin": pin } }),
  revokeAttestation: (id: string, reason: string, pin: string) =>
    api(apiPaths.attestationRevoke(id), certificateSchema, send("POST", { reason }, pin)),

  // ---- Flags ----

  flagsList: (status?: string) =>
    api(`${apiPaths.flags}${all}&sort=-created_at${status ? `&status=${status}` : ""}`, page(flagSchema)).then((p) => p.data),
  flag: (id: string) => api(apiPaths.flag(id), flagSchema),
  raiseFlag: (body: { passport_id: string; reason: string }) => api(apiPaths.flags, flagSchema, send("POST", body)),
  /** Investigators only: give the flag to a person, or to nobody (null). */
  assignFlag: (id: string, userId: string | null) => api(apiPaths.flagAssignee(id), flagSchema, send("PUT", { user_id: userId })),
  resolveFlag: (id: string, body: { outcome: string; note: string }, pin: string) =>
    api(apiPaths.flagResolve(id), flagSchema, send("POST", body, pin)),

  // ---- Audit and reports ----

  auditList: (params: { startDate?: string; endDate?: string; route?: string } = {}) => {
    const q = new URLSearchParams({ size: String(limits.pageSize) });
    if (params.startDate) q.set("start_date", params.startDate);
    if (params.endDate) q.set("end_date", params.endDate);
    if (params.route) q.set("route", params.route);
    return api(`${apiPaths.audit}?${q}`, page(auditEntry)).then((p) => ({ rows: p.data, total: p.total }));
  },
  auditVerify: () => api(apiPaths.auditVerify, chainCheck),
  volumeReport: (params: { startDate?: string; endDate?: string } = {}) => {
    const q = new URLSearchParams();
    if (params.startDate) q.set("start_date", params.startDate);
    if (params.endDate) q.set("end_date", params.endDate);
    const suffix = q.size ? `?${q}` : "";
    return api(`${apiPaths.reportsVolume}${suffix}`, z.object({ data: z.array(volumeRow) })).then((r) => r.data);
  },
};

// Where each counted tile's number comes from: a list and its filter.
const countSources: Record<Exclude<TileKey, "incoming" | "outgoing">, [string, string]> = {
  passportsHeld: [apiPaths.passports, `status=${passportStatus.active}`],
  openFlags: [apiPaths.flags, `status=${flagStatus.open}`],
  certValid: [apiPaths.attestations, `status=${certificateStatus.valid}`],
  certExpired: [apiPaths.attestations, `status=${certificateStatus.expired}`],
  certRevoked: [apiPaths.attestations, `status=${certificateStatus.revoked}`],
  orgsPending: [apiPaths.orgs, `status=${orgStatus.pending}`],
  orgsChanges: [apiPaths.orgs, `status=${orgStatus.changesRequested}`],
  orgsRejected: [apiPaths.orgs, `status=${orgStatus.rejected}`],
};

export type HandoverStart = { to_org: string; stage: string; quantity: number; note: string; lat?: number; lon?: number };

/** Form fields plus photos, with the signing PIN header (the browser sets the multipart boundary). */
function multipart(fields: Record<string, string | number | undefined>, photos: File[], pin: string): RequestInit {
  const body = new FormData();
  for (const [k, v] of Object.entries(fields)) if (v !== undefined) body.set(k, String(v));
  for (const f of photos) body.append(handoverPhotoField, f);
  return { method: "POST", body, headers: { "x-signing-pin": pin } };
}

export const handoverPhotoURL = (handoverId: string, photoId: string) => apiBase + apiPaths.handoverPhoto(handoverId, photoId);

export const documentFileURL = (id: string) => apiBase + apiPaths.documentFile(id);

export const facilityLicenceURL = (id: string) => apiBase + apiPaths.facilityLicence(id);

// ---- Hooks ----

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
