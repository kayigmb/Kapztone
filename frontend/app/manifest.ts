import type { MetadataRoute } from "next";

import { copy } from "@/lib/consts";

export default function manifest(): MetadataRoute.Manifest {
  return {
    name: copy.brand,
    short_name: copy.brand,
    description: copy.meta.landingDescription,
    start_url: "/",
    display: "standalone",
    background_color: "#ffffff",
    theme_color: "#087566",
    icons: [{ src: "/icon.svg", sizes: "any", type: "image/svg+xml" }],
  };
}
