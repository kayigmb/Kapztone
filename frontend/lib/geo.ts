// Clicked points → the GeoJSON the facility API expects. Pure: `node --test` runs it.

export type LonLat = [number, number];
export type DrawMode = "pin" | "area";
export type Geometry = { type: "Point"; coordinates: LonLat } | { type: "Polygon"; coordinates: LonLat[][] };

/** A boundary needs at least three corners. */
export const minCorners = 3;

/** The pin, or the boundary as a closed ring; undefined while there isn't enough to save. */
export function toGeometry(mode: DrawMode, points: LonLat[]): Geometry | undefined {
  if (mode === "pin") return points.length ? { type: "Point", coordinates: points[points.length - 1] } : undefined;
  if (points.length < minCorners) return undefined;
  return { type: "Polygon", coordinates: [[...points, points[0]]] };
}

/** Where to centre a map on some geometries: the middle of their corners. */
export function centreOf(geoms: Geometry[]): LonLat | undefined {
  const pts = geoms.flatMap((g) => (g.type === "Point" ? [g.coordinates] : g.coordinates[0]));
  if (!pts.length) return undefined;
  const sum = pts.reduce(([x, y], [a, b]) => [x + a, y + b], [0, 0]);
  return [sum[0] / pts.length, sum[1] / pts.length];
}
