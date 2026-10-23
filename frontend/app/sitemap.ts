import type { MetadataRoute } from "next";

import { routes } from "@/lib/consts";
import { siteURL } from "@/lib/site";

// The public pages only; lots and organisations are not for search.
export default function sitemap(): MetadataRoute.Sitemap {
  const site = siteURL();
  return [
    { url: site, changeFrequency: "monthly", priority: 1 },
    { url: site + routes.signup, changeFrequency: "yearly", priority: 0.6 },
    { url: site + routes.signin, changeFrequency: "yearly", priority: 0.4 },
  ];
}
