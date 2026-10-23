import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  // The marketing home's hero photo (decoration on public pages only).
  images: {
    remotePatterns: [{ protocol: "https", hostname: "images.unsplash.com" }],
  },
  // The admin screens used to live under /admin; old links still land.
  async redirects() {
    return [
      { source: "/admin/:path*", destination: "/:path*", permanent: true },
    ];
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
          // No script-src: Next's inline scripts and MapLibre's workers stay allowed.
          {
            key: "Content-Security-Policy",
            value:
              "frame-ancestors 'none'; base-uri 'self'; object-src 'none'; form-action 'self'",
          },
          // Production is served over https only; dev runs on plain http.
          ...(process.env.NODE_ENV === "production"
            ? [
                {
                  key: "Strict-Transport-Security",
                  value: "max-age=63072000; includeSubDomains",
                },
              ]
            : []),
        ],
      },
    ];
  },
};

export default nextConfig;
