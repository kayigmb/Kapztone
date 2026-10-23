import type { MetadataRoute } from "next";
import { headers } from "next/headers";

import { siteURL } from "@/lib/site";

// Only the bare address is for search engines; organisations' own addresses are not.
export default async function robots(): Promise<MetadataRoute.Robots> {
  const host = (await headers()).get("host");
  const site = siteURL();
  if (host !== new URL(site).host) {
    return { rules: { userAgent: "*", disallow: "/" } };
  }
  return {
    rules: {
      userAgent: "*",
      allow: "/",
      disallow: [
        "/api/",
        "/p/",
        "/dashboard",
        "/handoff",
        "/onboarding",
        "/org-not-found",
      ],
    },
    sitemap: `${site}/sitemap.xml`,
  };
}
