import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  // The marketing home's hero photo (decoration on public pages only).
  images: {
    remotePatterns: [{ protocol: "https", hostname: "images.unsplash.com" }],
  },
};

export default nextConfig;
