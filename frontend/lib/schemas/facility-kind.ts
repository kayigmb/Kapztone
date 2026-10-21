import { z } from "zod";

import { copy } from "../consts.ts";

const keyPattern = /^[a-z][a-z0-9_]{0,49}$/;

export const facilityKindSchema = z.object({
  key: z.string().trim().regex(keyPattern, copy.validation.facilityKindKey),
  label: z.string().trim().min(1, copy.validation.facilityKindName).max(120),
  sub: z.string().trim().max(500),
  active: z.boolean(),
});

export const facilityKindCreateSchema = facilityKindSchema.omit({
  active: true,
});
export const facilityKindUpdateSchema = facilityKindSchema.omit({ key: true });

export type FacilityKindForm = z.infer<typeof facilityKindSchema>;
