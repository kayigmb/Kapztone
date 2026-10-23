import {
  routeAccess,
  dashboardOfKind,
  dashboards,
  navItems,
  quickActions,
  tileTones,
  type DashboardKey,
  type TileKey,
} from "./consts.ts";

/** The part of GET /me these choices depend on. */
export type Viewer = {
  organisation: { kind: string } | null;
  permissions: string[];
};

const allowed = (v: Viewer, item: object) =>
  !("permission" in item) || v.permissions.includes(item.permission as string);

/** Acting for an organisation, or the platform; a member on the bare address acts for nobody and picks a workspace first. */
export const hasWorkspace = (v: Viewer) => dashboardFor(v) !== undefined;

const pickerNav: readonly string[] = ["workspaces", "profile"];

export const navFor = (v: Viewer) => {
  const picking = !hasWorkspace(v);
  return navItems.filter(
    (item) =>
      allowed(v, item) &&
      (picking ? pickerNav.includes(item.key) : item.key !== "workspaces"),
  );
};

/** The dashboard's quick actions, in its own order, keeping only the ones this person may use. */
export const actionsFor = (v: Viewer) => {
  const layout = layoutFor(v);
  if (!layout) return [];
  const byKey = new Map<string, (typeof quickActions)[number]>(
    quickActions.map((a) => [a.key, a]),
  );
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
export const toneOf = (tile: TileKey, count: number | undefined) =>
  count ? (tileTones[tile] ?? "neutral") : "neutral";

/** Whether this person may open the screen at pathname (routeAccess: longest prefix wins; unlisted screens are open to anyone signed in). */
export function canOpen(pathname: string, v: Pick<Viewer, "permissions">) {
  const rule = routeAccess
    .filter((r) =>
      r.prefix.endsWith("/")
        ? pathname.startsWith(r.prefix)
        : pathname === r.prefix || pathname.startsWith(`${r.prefix}/`),
    )
    .sort((a, b) => b.prefix.length - a.prefix.length)[0];
  return !rule || rule.anyOf.some((p) => v.permissions.includes(p));
}
