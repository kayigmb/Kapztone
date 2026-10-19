// Raising and resolving flags on a lot: mirrors flag.RaiseInput.Validate()
// and flag.ResolveInput.Validate().
// Pure: `node --test` runs it.
import { z } from "zod";

import { copy, flagOutcomes, limits } from "../consts.ts";

const v = copy.validation;

export const raiseFlagFormSchema = z.object({
  passport_id: z.string().trim().min(1, v.subjectId),
  reason: z
    .string()
    .trim()
    .min(1, v.reasonRequired)
    .max(limits.note, v.tooLong(limits.note)),
});

export type RaiseFlagForm = z.infer<typeof raiseFlagFormSchema>;

export const resolveFlagFormSchema = z
  .object({
    outcome: z.enum(flagOutcomes, { message: v.outcome }),
    note: z.string().trim().max(limits.note, v.tooLong(limits.note)),
  })
  .refine((f) => f.note !== "", { path: ["note"], message: v.reasonRequired });

export type ResolveFlagForm = z.infer<typeof resolveFlagFormSchema>;
