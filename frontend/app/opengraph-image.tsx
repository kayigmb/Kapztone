import { ImageResponse } from "next/og";

import { copy } from "@/lib/consts";

// The card shown when a link to the site is shared (every page without its own).
export const alt = copy.meta.landing;
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default function Image() {
  return new ImageResponse(
    <div
      style={{
        width: "100%",
        height: "100%",
        display: "flex",
        flexDirection: "column",
        justifyContent: "center",
        gap: 36,
        padding: 96,
        background: "#0b2639",
        color: "#ffffff",
      }}
    >
      <div style={{ display: "flex", alignItems: "center", gap: 28 }}>
        <svg width="112" height="112" viewBox="0 0 64 64" fill="none">
          <rect width="64" height="64" rx="16" fill="#087566" />
          <path
            d="M27 21H23.5a11 11 0 0 0 0 22H27M37 21h3.5a11 11 0 0 1 0 22H37M23 32h18"
            stroke="#fff"
            strokeWidth="4.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
        <span style={{ fontSize: 72, fontWeight: 700 }}>{copy.brand}</span>
      </div>
      <span style={{ fontSize: 56, fontWeight: 700, color: "#5fd0bd" }}>
        Know where it came from
      </span>
      <span style={{ fontSize: 32, color: "#c7d3dc", maxWidth: 960 }}>
        {copy.meta.landingDescription}
      </span>
    </div>,
    size,
  );
}
