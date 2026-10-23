import type { Metadata } from "next";

import { copy } from "@/lib/consts";
import { siteURL } from "@/lib/site";
import { DM_Mono, Inter } from "next/font/google";
import "./globals.css";
import { Providers } from "./providers";

// Inter for the interface, DM Mono for lot codes, IDs and quantities.
const sans = Inter({ subsets: ["latin"], variable: "--font-inter" });

const mono = DM_Mono({
  subsets: ["latin"],
  weight: ["400", "500"],
  variable: "--font-dm-mono",
});

// Pages stay out of search results unless they opt in (lib/site.ts → indexed).
export const metadata: Metadata = {
  metadataBase: new URL(siteURL()),
  title: copy.brand,
  applicationName: copy.brand,
  description: copy.meta.landingDescription,
  openGraph: { siteName: copy.brand, type: "website", locale: "en_GB" },
  twitter: { card: "summary_large_image" },
  robots: { index: false, follow: false },
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html
      lang="en"
      className={`${sans.variable} ${mono.variable}`}
      suppressHydrationWarning
    >
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
