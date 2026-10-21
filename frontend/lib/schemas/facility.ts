// Registering a facility: mirrors facility.RegisterInput.Validate().
// Pure: `node --test` runs it.
import { z } from "zod";

import { copy, limits } from "../consts.ts";

const v = copy.validation;

const lonLat = z.tuple([
  z.number().min(-180).max(180),
  z.number().min(-90).max(90),
]);
/** Drawn on the map (or set by GPS): a pin, or a boundary as a closed ring of at least 4 positions. */
const location = z.union(
  [
    z.object({ type: z.literal("Point"), coordinates: lonLat }),
    z.object({
      type: z.literal("Polygon"),
      coordinates: z.array(z.array(lonLat).min(4)).length(1),
    }),
  ],
  { message: v.facilityLocation },
);

const periods = ["season", "year", "month"] as const;

export const facilityFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, v.facilityName)
    .max(limits.name, v.tooLong(limits.name)),
  kind: z
    .string()
    .trim()
    .min(1, v.facilityKind)
    .regex(/^[a-z][a-z0-9_]{0,49}$/, v.facilityKind),
  licence_number: z
    .string()
    .trim()
    .min(1, v.licence)
    .max(limits.name, v.tooLong(limits.name)),
  capacity: z.coerce
    .number({ message: v.capacity })
    .int(v.capacity)
    .positive(v.capacity),
  capacity_unit: z
    .string()
    .trim()
    .min(1, v.capacityUnit)
    .max(16, v.tooLong(16)),
  capacity_period: z.enum(periods, { message: v.capacityPeriod }),
  licence_valid_to: z
    .string()
    .trim()
    .min(1, v.date)
    .refine((d) => d > new Date().toISOString().slice(0, 10), v.licenceExpired),
  location,
});

export type FacilityForm = z.infer<typeof facilityFormSchema>;
/** What the form holds before zod coerces it (numbers typed as text). */
export type FacilityFormInput = z.input<typeof facilityFormSchema>;
