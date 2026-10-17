// The browser's access to the API. Every call goes through this site's
// pass-through (app/api/[...path]/route.ts), so the session cookie and the
// organisation in the address travel with it. Every answer is checked
// against a zod schema; every failure becomes one ApiError.
import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { z } from "zod";

import { apiBase, apiPaths, copy, limits, queryKeys } from "@/lib/consts";
import type { ApiVerification } from "@/lib/verification";

// ---- Errors ----

const errorBody = z.object({
  error: z.string().catch(""),
  two_factor_required: z.boolean().optional(),
}).catch({ error: "" });

/** One kind of failure for every call. `status` 0 means the API could not be reached. */
export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
    readonly twoFactorRequired = false,
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
    throw new ApiError(res.status, body.error, body.two_factor_required);
  }
  return res;
}

/** Calls the API and checks the answer's shape. */
export async function api<T>(path: string, schema: z.ZodType<T>, init?: RequestInit): Promise<T> {
  const res = await call(path, init);
  return schema.parse(res.status === 204 ? undefined : await res.json());
}

const send = (method: string, body: unknown): RequestInit => ({
  method,
  headers: { "content-type": "application/json" },
  body: JSON.stringify(body),
});

// ---- Shapes (backend JSON) ----

const orgSummary = z.object({ id: z.string(), name: z.string(), slug: z.string(), kind: z.string(), status: z.string() });

export const meSchema = z.object({
  user: z.object({ id: z.string(), email: z.string(), name: z.string() }),
  organisation: orgSummary.nullable(),
  role: z.string(),
  permissions: z.array(z.string()).nullable().transform((p) => p ?? []),
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

const templateSchema = z.object({
  id: z.string(),
  stages: z.array(z.object({ key: z.string(), requiredEvidence: z.array(z.string()).optional() })),
});

export const accreditationSchema = z.object({ type: z.string(), valid_from: z.string(), valid_to: z.string() });
export type AccreditationRow = z.infer<typeof accreditationSchema>;
// A bare list, or the usual paged envelope: either is read the same way.
const accreditationList = z.union([z.array(accreditationSchema), page(accreditationSchema).transform((p) => p.data)]);

const verdict = z.enum(["verified", "warning", "failed"]);
const named = z.object({ id: z.string(), name: z.string() });
const passport = z.object({
  id: z.string(), organisation_id: z.string(), facility_id: z.string().nullable(), transform_id: z.string().nullable(),
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

const nothing = z.unknown();
const all = `?size=${limits.pageSize}`;

// ---- Endpoints ----

export const endpoints = {
  /** Who is signed in at this address, or null if nobody is. */
  me: () => api(apiPaths.me, meSchema).catch((err) => {
    if (err instanceof ApiError && err.status === 401) return null;
    throw err;
  }),
  login: (body: { email: string; password: string; code: string }) => api(apiPaths.login, nothing, send("POST", body)),
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

  /** The public verification of one passport, or null if there is no such passport. */
  verify: (id: string) => api(apiPaths.verify(id), verificationSchema).catch((err) => {
    if (err instanceof ApiError && err.status === 404) return null;
    throw err;
  }),
};

export const documentFileURL = (id: string) => apiBase + apiPaths.documentFile(id);

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
