import { z } from "zod";

import { copy, limits } from "../consts.ts";

const v = copy.validation;

const ruleTypes = ["yield_per_area", "facility_capacity"] as const;

export const capacityRuleSchema = z
  .object({
    type: z.enum(ruleTypes),
    kgPerHaPerSeason: z.coerce.number().positive(v.capacity).optional(),
  })
  .refine(
    (r) => r.type !== "yield_per_area" || r.kgPerHaPerSeason !== undefined,
    {
      path: ["kgPerHaPerSeason"],
      message: v.capacity,
    },
  );

export const templateFormSchema = z.object({
  commodity: z
    .string()
    .trim()
    .min(1, v.commodity)
    .max(limits.name, v.tooLong(limits.name)),
  name: z
    .string()
    .trim()
    .min(1, v.templateName)
    .max(limits.name, v.tooLong(limits.name)),
  unit: z.string().trim().min(1, v.unit).max(16, v.tooLong(16)),
  originFacilityKind: z
    .string()
    .trim()
    .min(1, v.facilityKind)
    .regex(/^[a-z][a-z0-9_]{0,49}$/, v.facilityKind),
  capacityRule: capacityRuleSchema.nullable(),
  /** Typed comma separated; sent as the list of certificate types the origin facility must hold. */
  requiredEvidence: z
    .string()
    .transform((s) =>
      s
        .split(",")
        .map((e) => e.trim())
        .filter(Boolean),
    )
    .pipe(z.array(z.string().regex(/^[a-z][a-z0-9_]{0,49}$/, v.evidence))),
});

export type TemplateForm = z.infer<typeof templateFormSchema>;
/** What the form holds before zod coerces it (numbers typed as text). */
export type TemplateFormInput = z.input<typeof templateFormSchema>;

