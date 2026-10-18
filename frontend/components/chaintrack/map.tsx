"use client";

import dynamic from "next/dynamic";

import { Loading } from "@/components/chaintrack/page-state";

/** The map, loaded only in the browser and only on pages that show one (MapLibre is large). */
export const MapView = dynamic(() => import("./map-view"), {
  ssr: false,
  loading: () => <Loading className="p-5" />,
});
export type { Shape } from "./map-view";
