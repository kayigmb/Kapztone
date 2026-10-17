// Server-side access to the API. The browser never calls the API directly:
// pages fetch here while rendering, and browser requests go through
// app/api/[...path]/route.ts. Both tell the API which site the browser is on.
import { cookies, headers } from "next/headers";

import type { ApiVerification } from "@/lib/verification";

// The API's address as the Next.js server reaches it. Required: there is no default.
export function apiURL(): string {
  const url = process.env.API_URL;
  if (!url) {
    throw new Error("API_URL is not set: add it to .env (see .env.example)");
  }
  return url.replace(/\/$/, "");
}

/** Calls the API as the person whose request is being rendered: their session cookie, their site. */
export async function apiFetch(path: string, init?: RequestInit): Promise<Response> {
  const [cookieStore, incoming] = await Promise.all([cookies(), headers()]);
  return fetch(apiURL() + path, {
    ...init,
    cache: "no-store",
    headers: {
      ...init?.headers,
      cookie: cookieStore.toString(),
      // The first label of this host names the organisation being acted for.
      "x-forwarded-host": incoming.get("host") ?? "",
    },
  });
}

export type Me = {
  user: { id: string; email: string; name: string };
  organisation: { id: string; name: string; slug: string; kind: string; status: string } | null;
  role: string;
  permissions: string[];
};

/** Who is signed in and what they may do at this address, or null if nobody is. */
export async function getMe(): Promise<Me | null> {
  const res = await apiFetch("/me");
  if (res.status === 401) return null;
  if (!res.ok) throw new Error(`GET /me failed: ${res.status}`);
  return res.json();
}

export type Organisation = NonNullable<Me["organisation"]> & { review_reason: string | null };
export type OrgDocument = { id: string; type: string; file_name: string; size_bytes: number; created_at: string };

/** The acting organisation with its review outcome, and the documents it has uploaded. */
export async function getOnboarding(orgId: string): Promise<{ organisation: Organisation; documents: OrgDocument[] }> {
  const [org, docs] = await Promise.all([apiFetch(`/orgs/${orgId}`), apiFetch(`/orgs/${orgId}/documents?size=100&sort=created_at`)]);
  if (!org.ok || !docs.ok) throw new Error(`onboarding failed: ${org.status} ${docs.status}`);
  return { organisation: await org.json(), documents: (await docs.json()).data };
}

/** The public verification of one passport, or null if there is no such passport. */
export async function getVerification(id: string): Promise<ApiVerification | null> {
  // Public and not cached here: the API already holds each answer for a minute.
  const res = await fetch(`${apiURL()}/verify/${encodeURIComponent(id)}`, { cache: "no-store" });
  if (res.status === 404) return null;
  if (!res.ok) throw new Error(`verification failed: ${res.status}`);
  return res.json();
}
