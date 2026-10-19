// What a scanned QR label or a typed lot code points to.
// Pure: `node --test` runs it.
import { z } from "zod";

import { copy } from "../consts.ts";

/**
 * A QR label holds the passport's public address (…/p/{code}); a typed value is
 * the lot code (or a passport id) itself. Either way: the part /p/ is followed by.
 */
export function passportRef(text: string): string {
  const value = text.trim();
  try {
    const match = new URL(value).pathname.match(/\/p\/([^/?#]+)/);
    if (match) return decodeURIComponent(match[1]);
  } catch {
    // not a URL: a code typed or printed as plain text
  }
  return value;
}

export const lotCodeSchema = z.object({
  code: z
    .string()
    .trim()
    .min(1, copy.validation.lotCode)
    .transform(passportRef),
});
export type LotCodeForm = z.input<typeof lotCodeSchema>;
