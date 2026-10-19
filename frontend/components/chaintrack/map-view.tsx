"use client";

// A MapLibre map of facility shapes, optionally clickable for drawing a pin or
// a boundary. Loaded only in the browser (see map.tsx). Colours come from the
// theme tokens at runtime, so the map follows light / dark like everything else.

import "maplibre-gl/dist/maplibre-gl.css";

import Map, {
  Layer,
  Marker,
  NavigationControl,
  Source,
} from "react-map-gl/maplibre";

import { useThemeTokens } from "@/components/chaintrack/use-theme-tokens";
import { mapStyleURL } from "@/lib/consts";
import { centreOf, type Geometry, type LonLat } from "@/lib/geo";

const tokenNames = ["primary", "pending", "warning", "danger"] as const;

export type Shape = {
  id: string;
  geometry: Geometry;
  tone?: "primary" | "pending" | "warning" | "danger";
};

export default function MapView({
  shapes = [],
  draft,
  onPick,
  selected,
  label,
  className,
}: {
  shapes?: Shape[];
  draft?: { points: LonLat[]; geometry?: Geometry };
  onPick?: (p: LonLat) => void;
  selected?: string;
  label: string;
  className?: string;
}) {
  const colours = useThemeTokens(tokenNames);
  const all = [
    ...shapes.map((s) => s.geometry),
    ...(draft?.geometry ? [draft.geometry] : []),
  ];
  const [lon, lat] = centreOf(all) ?? [29.87, -1.94]; // no shapes yet: Rwanda (the pilot)

  const features = (list: Shape[]) => ({
    type: "FeatureCollection" as const,
    features: list.map((s) => ({
      type: "Feature" as const,
      id: s.id,
      geometry: s.geometry,
      properties: {
        colour: colours[s.tone ?? "primary"],
        selected: s.id === selected,
      },
    })),
  });
  const areas = features(shapes.filter((s) => s.geometry.type === "Polygon"));
  const pins = shapes.filter((s) => s.geometry.type === "Point");

  return (
    <div className={className} role="region" aria-label={label}>
      <Map
        initialViewState={{
          longitude: lon,
          latitude: lat,
          zoom: all.length ? 13 : 7,
        }}
        mapStyle={mapStyleURL}
        style={{ width: "100%", height: "100%" }}
        cursor={onPick ? "crosshair" : "grab"}
        onClick={
          onPick
            ? (e) =>
                onPick([
                  Number(e.lngLat.lng.toFixed(6)),
                  Number(e.lngLat.lat.toFixed(6)),
                ])
            : undefined
        }
      >
        <NavigationControl position="top-right" showCompass={false} />
        <Source id="areas" type="geojson" data={areas}>
          <Layer
            id="areas-fill"
            type="fill"
            paint={{ "fill-color": ["get", "colour"], "fill-opacity": 0.18 }}
          />
          <Layer
            id="areas-line"
            type="line"
            paint={{
              "line-color": ["get", "colour"],
              "line-width": ["case", ["get", "selected"], 4, 2],
            }}
          />
        </Source>
        {pins.map((s) => (
          <Marker
            key={s.id}
            longitude={(s.geometry as { coordinates: LonLat }).coordinates[0]}
            latitude={(s.geometry as { coordinates: LonLat }).coordinates[1]}
            color={colours[s.tone ?? "primary"]}
          />
        ))}
        {draft && draft.points.length > 0 && (
          <Source
            id="draft"
            type="geojson"
            data={{
              type: "Feature",
              properties: {},
              geometry:
                draft.geometry?.type === "Polygon"
                  ? draft.geometry
                  : draft.points.length > 1
                    ? { type: "LineString", coordinates: draft.points }
                    : { type: "Point", coordinates: draft.points[0] },
            }}
          >
            <Layer
              id="draft-fill"
              type="fill"
              filter={["==", "$type", "Polygon"]}
              paint={{ "fill-color": colours.primary, "fill-opacity": 0.22 }}
            />
            <Layer
              id="draft-line"
              type="line"
              paint={{
                "line-color": colours.primary,
                "line-width": 3,
                "line-dasharray": [2, 1],
              }}
            />
          </Source>
        )}
        {draft?.points.map(([x, y], i) => (
          <Marker
            key={i}
            longitude={x}
            latitude={y}
            color={colours.primary}
            scale={0.6}
          />
        ))}
      </Map>
    </div>
  );
}
