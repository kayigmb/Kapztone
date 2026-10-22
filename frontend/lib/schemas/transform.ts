import { z } from "zod";

import { copy, limits } from "../consts.ts";

const v = copy.validation;

const quantity = z.coerce
  .number({ message: v.quantity })
  .int(v.quantity)
  .positive(v.quantityPositive);

export const inputRowSchema = z.object({
  passport_id: z.string().trim().min(1, v.subjectId),
  quantity,
});

export const outputRowSchema = z.object({ quantity });

export const transformFormSchema = z.object({
  stage: z.string().trim().min(1, v.stage),
  inputs: z
    .array(inputRowSchema)
    .min(1, v.quantity)
    .max(limits.maxItems, v.tooLong(limits.maxItems)),
  outputs: z
    .array(outputRowSchema)
    .min(1, v.quantity)
    .max(limits.maxItems, v.tooLong(limits.maxItems)),
});

export type TransformForm = z.infer<typeof transformFormSchema>;
/** What the form holds before zod coerces it (numbers typed as text). */
export type TransformFormInput = z.input<typeof transformFormSchema>;

/** The yield of a transform: output over input, as a percentage. */
export function yieldOf(
  inputs: number[],
  outputs: number[],
  maxPct: number | null | undefined,
): {
  totalIn: number;
  totalOut: number;
  pct: number;
  max: number | null;
  over: boolean;
} {
  const totalIn = inputs.reduce((a, b) => a + b, 0);
  const totalOut = outputs.reduce((a, b) => a + b, 0);
  const pct = totalIn === 0 ? 0 : (totalOut / totalIn) * 100;
  const max = maxPct ?? null;
  return { totalIn, totalOut, pct, max, over: max !== null && pct > max };
}
