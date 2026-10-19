import type { Metadata } from "next";

import { copy } from "@/lib/consts";
import { DM_Mono, Inter } from "next/font/google";
import "./globals.css";
import { Providers } from "./providers";

// Inter for the interface, DM Mono for lot codes, IDs and quantities (ui/figma-make).
const sans = Inter({ subsets: ["latin"], variable: "--font-inter" });

const mono = DM_Mono({
  subsets: ["latin"],
  weight: ["400", "500"],
  variable: "--font-dm-mono",
});

export const metadata: Metadata = {
  title: copy.brand,
  description: "Digital product passports for business shipments.",
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
