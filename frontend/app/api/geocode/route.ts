import { photonSuggestions } from "@/lib/geocode";

export async function GET(request: Request) {
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
