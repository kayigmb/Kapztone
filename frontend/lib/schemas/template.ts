// A lot passport template: mirrors template.CreateInput.Validate().
// Pure: `node --test` runs it.
import { z } from "zod";

import { copy, limits, templateEvents } from "../consts.ts";

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

export const stageFormSchema = z.object({
  key: z.string().trim().min(1, v.stage),
  events: z
    .array(z.enum(templateEvents.map((e) => e.value) as [string, ...string[]]))
    .min(1, v.events),
  maxYield: z.coerce.number().positive(v.capacity).optional(),
  requiredEvidence: z.array(z.string().trim().min(1)),
});

export type StageForm = z.infer<typeof stageFormSchema>;

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
  stages: z.array(stageFormSchema).min(1, v.stages),
});

export type TemplateForm = z.infer<typeof templateFormSchema>;
/** What the form holds before zod coerces it (numbers typed as text). */
export type TemplateFormInput = z.input<typeof templateFormSchema>;

/**
 * The API holds a stage's maximum yield as a fraction (0.2: output ≤ input × 0.2);
 * every screen shows and takes it as a percentage (20). lib/api.ts converts at the
 * boundary with these two, so nothing else has to know.
 */
export const yieldToPct = (fraction: number) =>
  Math.round(fraction * 1e6) / 1e4;
export const yieldToFraction = (pct: number) => Math.round(pct * 1e4) / 1e6;
