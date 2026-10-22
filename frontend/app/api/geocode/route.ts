import { photonSuggestions } from "@/lib/geocode";

// Signed-in only: otherwise anyone could run searches through this server.
async function signedIn(request: Request): Promise<boolean> {
  const cookie = request.headers.get("cookie");
  if (!cookie || !process.env.API_URL) return false;
  try {
    const me = await fetch(`${process.env.API_URL.replace(/\/$/, "")}/me`, {
      headers: { cookie },
      signal: AbortSignal.timeout(5_000),
    });
    return me.ok;
  } catch {
    return false;
  }
}

export async function GET(request: Request) {
  if (!(await signedIn(request))) {
    return Response.json(
      { status: 401, message: "Sign in to search places.", data: null },
      { status: 401 },
    );
  }
  const query = new URL(request.url).searchParams.get("q")?.trim() ?? "";
  if (query.length < 3 || query.length > 120) {
    return Response.json(
      {
        status: 400,
        message: "Search must be 3 to 120 characters.",
        data: null,
      },
      { status: 400 },
    );
  }

  const params = new URLSearchParams({ q: query, limit: "6", lang: "en" });
  try {
    const response = await fetch(`https://photon.komoot.io/api/?${params}`, {
      headers: { accept: "application/json" },
      signal: AbortSignal.timeout(8_000),
    });
    if (!response.ok) throw new Error("Photon search failed");
    const places = photonSuggestions(await response.json());
    return Response.json({ status: 200, message: "", data: places });
  } catch {
    return Response.json(
      { status: 502, message: "", data: null },
      { status: 502 },
    );
  }
}
