// An organisation's own address: its slug as the first label of this site's
// host (acme.example.com; acme.localhost:3000 in development). The platform
// itself is the bare host. Pure, so `node --test` runs it.

type Here = { protocol: string; host: string };

/** This site's host without the acting organisation's label in front. */
export function baseHost(here: Here, currentSlug?: string): string {
  return currentSlug && here.host.startsWith(`${currentSlug}.`) ? here.host.slice(currentSlug.length + 1) : here.host;
}

/** The address of an organisation (`slug`), or of the platform (`null`), seen from here. */
export function orgAddress(slug: string | null, here: Here, currentSlug?: string, path = ""): string {
  const base = baseHost(here, currentSlug);
  return `${here.protocol}//${slug ? `${slug}.` : ""}${base}${path}`;
}
