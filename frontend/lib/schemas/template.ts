// A lot passport template: mirrors template.CreateInput.Validate().
// Pure: `node --test` runs it.
import { z } from "zod";

import { copy, facilityKinds, limits, templateEvents } from "../consts.ts";

const v = copy.validation;

const ruleTypes = ["yield_per_area", "facility_capacity"] as const;

export const capacityRuleSchema = z.object({
  type: z.enum(ruleTypes),
  kgPerHaPerSeason: z.coerce.number().positive(v.capacity).optional(),
}).refine((r) => r.type !== "yield_per_area" || r.kgPerHaPerSeason !== undefined, {
  path: ["kgPerHaPerSeason"], message: v.capacity,
});

export const stageFormSchema = z.object({
  key: z.string().trim().min(1, v.stage),
  events: z.array(z.enum(templateEvents.map((e) => e.value) as [string, ...string[]]))
    .min(1, v.events),
  maxYield: z.coerce.number().positive(v.capacity).optional(),
  requiredEvidence: z.array(z.string().trim().min(1)),
});

export type StageForm = z.infer<typeof stageFormSchema>;

export const templateFormSchema = z.object({
  commodity: z.string().trim().min(1, v.commodity).max(limits.name, v.tooLong(limits.name)),
  name: z.string().trim().min(1, v.templateName).max(limits.name, v.tooLong(limits.name)),
  unit: z.string().trim().min(1, v.unit).max(16, v.tooLong(16)),
  originFacilityKind: z.string().trim(),
  capacityRule: capacityRuleSchema.nullable(),
  stages: z.array(stageFormSchema).min(1, v.stages),
}).refine(
  (f) => f.originFacilityKind === "" || facilityKinds.some((k) => k.value === f.originFacilityKind),
  { path: ["originFacilityKind"], message: v.facilityKind },
);

export type TemplateForm = z.infer<typeof templateFormSchema>;
/** What the form holds before zod coerces it (numbers typed as text). */
export type TemplateFormInput = z.input<typeof templateFormSchema>;
