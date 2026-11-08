"use client";

import { useQuery } from "@tanstack/react-query";
import QRCode from "qrcode";

/** A QR code of text, drawn in the browser as a PNG data URL; nothing leaves the page. */
export const useQr = (text: string, size = 192) =>
  useQuery({
    queryKey: ["qr", text, size],
    queryFn: () => QRCode.toDataURL(text, { margin: 1, width: size }),
    staleTime: Infinity,
  });

export function QrImage({ text, size = 192 }: { text: string; size?: number }) {
  const qr = useQr(text, size);
  return qr.data ? (
    // eslint-disable-next-line @next/next/no-img-element -- a data: URL, nothing for next/image to optimise
    <img
      src={qr.data}
      alt=""
      width={size}
      height={size}
      className="rounded-md bg-white p-2 ring-1 ring-border"
    />
  ) : (
    <span
      className="animate-pulse rounded-md bg-surface-sunken"
      style={{ width: size, height: size }}
    />
  );
}
