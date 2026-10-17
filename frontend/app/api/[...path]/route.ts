// Passes a browser's request through to the API, so the browser only ever
// talks to this site. That is what lets the API's session cookie belong to
// this host alone, with no domain configured anywhere.
import type { NextRequest } from "next/server";

import { apiURL } from "@/lib/api";

// Only what the API needs travels on; in particular the browser's Origin,
// which the API checks before letting a cookie change anything.
const forwarded = ["cookie", "content-type", "origin", "authorization", "accept"];
const returned = ["content-type", "content-disposition", "x-content-type-options"];

async function pass(req: NextRequest, ctx: RouteContext<"/api/[...path]">) {
  const { path } = await ctx.params;
  const url = `${apiURL()}/${path.map(encodeURIComponent).join("/")}${req.nextUrl.search}`;

  const headers = new Headers();
  for (const name of forwarded) {
    const value = req.headers.get(name);
    if (value) headers.set(name, value);
  }
  // The site the browser is on: its first label names the organisation.
  headers.set("x-forwarded-host", req.headers.get("host") ?? "");
  headers.set("x-forwarded-proto", req.nextUrl.protocol.replace(":", ""));

  const hasBody = req.method !== "GET" && req.method !== "HEAD";
  const res = await fetch(url, {
    method: req.method,
    headers,
    body: hasBody ? req.body : undefined,
    duplex: "half", // stream the body through: uploads are not held in memory
    redirect: "manual",
    cache: "no-store",
  } as RequestInit);

  const out = new Headers();
  for (const name of returned) {
    const value = res.headers.get(name);
    if (value) out.set(name, value);
  }
  for (const cookie of res.headers.getSetCookie()) out.append("set-cookie", cookie);
  return new Response(res.body, { status: res.status, headers: out });
}

export { pass as GET, pass as POST, pass as PUT, pass as DELETE };
