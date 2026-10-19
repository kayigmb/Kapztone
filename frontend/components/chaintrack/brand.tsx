import { Link2 } from "lucide-react";

import { copy } from "@/lib/consts";
import { cn } from "@/lib/utils";

/** The ChainTrack mark: a teal square with a link, and the name (Figma .brand). */
export function Brand({
  onNavy,
  small,
  nameHidden,
  className,
}: {
  onNavy?: boolean;
  small?: boolean;
  nameHidden?: boolean;
  className?: string;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-2.5 font-bold tracking-tight",
        small ? "text-base" : "text-lg",
        onNavy ? "text-hero-ink" : "text-ink",
        className,
      )}
    >
      <span
        className={cn(
          "grid shrink-0 place-items-center rounded-md bg-primary text-on-primary",
          small ? "size-8" : "size-9",
        )}
        aria-hidden
      >
        <Link2 className={small ? "size-4.5" : "size-5"} />
      </span>
      <span className={nameHidden ? "sr-only" : undefined}>{copy.brand}</span>
    </span>
  );
}
