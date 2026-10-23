import type { Metadata } from "next";

import { apiPaths, copy } from "@/lib/consts";
import { humanize, type ApiVerification } from "@/lib/verification";
import { VerificationView } from "./verification-view";

const t = copy.verification;

// A shared lot link previews the lot and its verdict; lots stay out of search results.
// ponytail: one more API call per view; the API caches /verify for a minute.
export async function generateMetadata({
  params,
}: PageProps<"/p/[id]">): Promise<Metadata> {
  const { id } = await params;
  const plain = {
    title: copy.meta.passport,
    robots: { index: false, follow: true },
  };
  try {
    const res = await fetch(
      `${process.env.API_URL?.replace(/\/$/, "")}${apiPaths.verify(id)}`,
    );
    if (!res.ok) return plain;
    const v = ((await res.json()) as { data: ApiVerification }).data;
    const title = `${v.passport.code ?? id} · ${humanize(v.passport.commodity)} — ${t.verdict[v.verdict]} · ${copy.brand}`;
    const origin = v.origins[0]?.organisation.name;
    const description = origin
      ? `${t.verdict[v.verdict]}: ${humanize(v.passport.commodity)} from ${origin}, checked against its sealed records.`
      : `${t.verdict[v.verdict]}: ${humanize(v.passport.commodity)}, checked against its sealed records.`;
    return { ...plain, title, description, openGraph: { title, description } };
  } catch {
    return plain;
  }
}

export default async function Page({ params }: PageProps<"/p/[id]">) {
  const { id } = await params;
  return <VerificationView id={id} />;
}
