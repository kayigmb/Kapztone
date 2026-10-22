"use client";

import { useMeGuard } from "@/lib/api";
import { routes } from "@/lib/consts";

/** Signed in: straight to the dashboard. */
export function LandingGate({ children }: { children: React.ReactNode }) {
  const me = useMeGuard((who) => (who ? routes.home : undefined));
  if (me.data) return null;
  return <>{children}</>;
}
