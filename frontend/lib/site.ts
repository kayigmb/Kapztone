/** The site's bare address (APP_URL), for canonical links, previews and the sitemap. Server only. */
export function siteURL(): string {
  return (process.env.APP_URL ?? "http://localhost:3000").replace(/\/$/, "");
}

/** Search engines may list a page: the public pages, on the bare address only (robots.ts keeps other hosts out). */
export const indexed = (path: string) => ({
  robots: { index: true, follow: true },
  alternates: { canonical: path },
});
