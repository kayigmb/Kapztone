"use client";

import { useTheme } from "next-themes";
import { useMemo } from "react";

const read = (name: string) =>
  typeof window === "undefined"
    ? ""
    : getComputedStyle(document.documentElement)
        .getPropertyValue(`--${name}`)
        .trim();

/**
 * Theme colours for code that can't take a CSS class (MapLibre paint, chart
 * strokes): read from the tokens and read again when the theme flips.
 */
export function useThemeTokens<K extends string>(
  names: readonly K[],
): Record<K, string> {
  const { resolvedTheme } = useTheme();
  const key = names.join(",");
  return useMemo(
    () =>
      Object.fromEntries(names.map((n) => [n, read(n)])) as Record<K, string>,
    [resolvedTheme, key], // eslint-disable-line react-hooks/exhaustive-deps -- names are read through key
  );
}
