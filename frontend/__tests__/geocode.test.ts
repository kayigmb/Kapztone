import assert from "node:assert/strict";
import { test } from "node:test";

import { photonSuggestions } from "../lib/geocode.ts";

test("Photon features become concise place labels and coordinates", () => {
  assert.deepEqual(
    photonSuggestions({
      features: [
        {
          type: "Feature",
          geometry: { type: "Point", coordinates: [29.36, -1.94] },
          properties: {
            name: "Kigali",
            city: "Kigali",
            state: "Kigali City",
            country: "Rwanda",
            osm_id: 123,
          },
        },
        {
          type: "Feature",
          geometry: { type: "Point", coordinates: [29.4, -2.1] },
          properties: {
            street: "Market Road",
            housenumber: "12",
            country: "Rwanda",
          },
        },
      ],
    }),
    [
      { label: "Kigali · Kigali City, Rwanda", coordinates: [29.36, -1.94] },
      { label: "12 Market Road · Rwanda", coordinates: [29.4, -2.1] },
    ],
  );
});

test("Photon features without any useful label are omitted", () => {
  assert.deepEqual(
    photonSuggestions({
      features: [
        {
          type: "Feature",
          geometry: { type: "Point", coordinates: [29, -2] },
          properties: {},
        },
      ],
    }),
    [],
  );
});
