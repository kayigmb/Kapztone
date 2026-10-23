import { NextResponse, type NextRequest } from "next/server";

// An address naming no organisation (nope.chaintrack.app) gets the "no such
// organisation" page instead of a sign-in form for nobody. The API decides
// (GET /address); answers are kept per host for a minute.
// ponytail: per-instance cache; share it only if the web app runs as several replicas.
const known = new Map<
  string,
  { exists: boolean; home: string; until: number }
>();
const keepFor = 60_000;

async function address(host: string) {
  const hit = known.get(host);
  if (hit && hit.until > Date.now()) return hit;
  const res = await fetch(
    `${process.env.API_URL?.replace(/\/$/, "")}/address`,
    {
      headers: { "x-forwarded-host": host },
    },
  );
  const { data } = (await res.json()) as {
    data: { exists: boolean; home: string };
  };
  const answer = { ...data, until: Date.now() + keepFor };
  known.set(host, answer);
  return answer;
}

export async function proxy(req: NextRequest) {
  const host = req.headers.get("host") ?? "";
  try {
    const { exists, home } = await address(host);
    if (!exists) {
      const url = new URL("/org-not-found", req.url);
      url.searchParams.set("host", host.split(":")[0]);
      url.searchParams.set("home", home);
      return NextResponse.rewrite(url, { status: 404 });
    }
  } catch {
    // The API is unreachable: let the page load and report it itself.
  }
  return NextResponse.next();
}

// Pages only: not the API pass-through, Next's own files, static files or the page itself.
export const config = {
  matcher: ["/((?!api|_next|org-not-found|.*\\..*).*)"],
};
