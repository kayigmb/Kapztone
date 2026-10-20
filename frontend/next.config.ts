import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  // The marketing home's hero photo (decoration on public pages only).
  images: {
    remotePatterns: [{ protocol: "https", hostname: "images.unsplash.com" }],
  },
  // OWASP A02: deny framing and MIME sniffing, limit referrer data.
  async headers() {
    return [
      {
        source: "/:path*",
        headers: [
          { key: "X-Frame-Options", value: "DENY" },
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
        ],
      },
    ];
  },
};

export default nextConfig;
