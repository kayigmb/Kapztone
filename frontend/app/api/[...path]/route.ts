// Passes a browser's request through to the API, so the browser only ever
// talks to this site. That is what lets the API's session cookie belong to
// this host alone, with no domain configured anywhere.
import type { NextRequest } from "next/server";

const requestIdHeader = "x-request-id";

// The API's address as this server reaches it. Required: there is no default.
function apiURL(): string {
  const url = process.env.API_URL;
  if (!url) {
    throw new Error("API_URL is not set: add it to .env (see .env.example)");
  }
  return url.replace(/\/$/, "");
}

// Only what the API needs travels on; in particular the browser's Origin,
// which the API checks before letting a cookie change anything.
const forwarded = [
  "cookie",
  "content-type",
  "origin",
  "authorization",
  "accept",
  "x-signing-pin",
];
const returned = [
  "content-type",
  "content-disposition",
  "x-content-type-options",
  requestIdHeader,
];

async function pass(req: NextRequest, ctx: RouteContext<"/api/[...path]">) {
  const { path } = await ctx.params;
  const url = `${apiURL()}/${path.map(encodeURIComponent).join("/")}${req.nextUrl.search}`;

  const headers = new Headers();
  for (const name of forwarded) {
    const value = req.headers.get(name);
    if (value) headers.set(name, value);
  }
  // One id per request, on the API's log lines and back to the page, so a
  // reported failure can be found in the logs.
  headers.set(requestIdHeader, crypto.randomUUID());
  // The site the browser is on: its first label names the organisation.
  headers.set("x-forwarded-host", req.headers.get("host") ?? "");
  // The visitor's address, for the API's rate limits. Next fills this from the
  // socket when no proxy set it; behind nginx the last entry is the one nginx
  // appended. Only that one is passed on.
  // ponytail: without a proxy in front, a visitor can send their own header.
  const visitor = req.headers.get("x-forwarded-for")?.split(",").at(-1)?.trim();
  if (visitor) headers.set("x-forwarded-for", visitor);

  const hasBody = req.method !== "GET" && req.method !== "HEAD";
  let res: Response;
  try {
    res = await fetch(url, {
      method: req.method,
      headers,
      body: hasBody ? req.body : undefined,
      duplex: "half", // stream the body through: uploads are not held in memory
      redirect: "manual",
      cache: "no-store",
    } as RequestInit);
  } catch {
    // The API is down or unreachable: say so, so the page can tell "try again" from a refusal.
    return Response.json(
      { status: 502, message: "", data: null },
      { status: 502 },
    ); // problemOf says "unreachable"
  }

  const out = new Headers();
  for (const name of returned) {
    const value = res.headers.get(name);
    if (value) out.set(name, value);
  }
  for (const cookie of res.headers.getSetCookie())
    out.append("set-cookie", cookie);
  return new Response(res.body, { status: res.status, headers: out });
}

export { pass as GET, pass as POST, pass as PUT, pass as DELETE };
