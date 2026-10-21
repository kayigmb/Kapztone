import { z } from "zod";

const photonFeatureSchema = z.object({
  geometry: z.object({
    type: z.literal("Point"),
    coordinates: z.tuple([z.number(), z.number()]),
  }),
  properties: z
    .object({
      name: z.string().optional(),
      street: z.string().optional(),
      housenumber: z.string().optional(),
      city: z.string().optional(),
      county: z.string().optional(),
      state: z.string().optional(),
      country: z.string().optional(),
    })
    .passthrough(),
});

export const photonResponseSchema = z.object({
  features: z.array(photonFeatureSchema),
});

export const placeSchema = z.object({
  label: z.string(),
  coordinates: z.tuple([z.number(), z.number()]),
});

export type PlaceSuggestion = z.infer<typeof placeSchema>;

/** Convert Photon results into compact, readable labels for the map picker. */
export function photonSuggestions(value: unknown): PlaceSuggestion[] {
  const { features } = photonResponseSchema.parse(value);
  return features
    .map(({ geometry, properties }) => {
      const street = [properties.housenumber, properties.street]
        .filter(Boolean)
        .join(" ");
      const primary =
        properties.name || street || properties.city || properties.county;
      const locality = [properties.city, properties.state, properties.country]
        .filter((part) => part && part !== primary)
        .join(", ");
      const label = [primary, locality].filter(Boolean).join(" · ");
      return { label, coordinates: geometry.coordinates };
    })
    .filter((place) => place.label.length > 0);
}
