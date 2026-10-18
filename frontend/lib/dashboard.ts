// Which nav items, dashboard and quick actions a signed-in person gets.
// Pure: no imports with side effects, so `node --test` can run it.

import {
  dashboardOfKind, dashboards, navItems, quickActions, tileTones, type DashboardKey, type TileKey,
} from "./consts.ts";

/** The part of GET /me these choices depend on. */
export type Viewer = { organisation: { kind: string } | null; permissions: string[] };

const allowed = (v: Viewer, item: object) => !("permission" in item) || v.permissions.includes(item.permission as string);

export const navFor = (v: Viewer) => navItems.filter((item) => allowed(v, item));

/** The dashboard's quick actions, in its own order, keeping only the ones this person may use. */
export const actionsFor = (v: Viewer) => {
  const layout = layoutFor(v);
  if (!layout) return [];
  const byKey = new Map<string, (typeof quickActions)[number]>(quickActions.map((a) => [a.key, a]));
  return layout.actions.flatMap((key) => {
    const a = byKey.get(key);
    return a && allowed(v, a) ? [a] : [];
  });
};

/** An organisation's kind picks its dashboard; the platform admin (no organisation) gets the platform one. */
export function dashboardFor(v: Viewer): DashboardKey | undefined {
  if (v.organisation) return dashboardOfKind[v.organisation.kind];
  return v.permissions.length > 0 ? "platform" : undefined;
}

export const layoutFor = (v: Viewer) => {
  const key = dashboardFor(v);
  return key ? dashboards[key] : undefined;
};

/** A tile with something in it may need the reader: attention or alert. */
export const toneOf = (tile: TileKey, count: number | undefined) => (count ? tileTones[tile] ?? "neutral" : "neutral");
