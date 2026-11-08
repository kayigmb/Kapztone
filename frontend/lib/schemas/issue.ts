import { z } from "zod";

import { copy } from "../consts.ts";

const v = copy.validation;

export const issueFormSchema = z.object({
  facility_id: z.string().trim().min(1, v.facility),
  // The API turns it into a key (Green coffee → green_coffee): letters first, then letters, digits, spaces or _.
  commodity: z
    .string()
    .trim()
    .min(1, v.commodity)
    .max(50, v.tooLong(50))
    .regex(/^(|[A-Za-z][A-Za-z0-9_ ]*)$/, v.commodity), // empty is min's to report
  // A common unit, or the producer's own (crates, lb…); the API takes 1 to 16 characters.
  unit: z.string().trim().min(1, v.unit).max(16, v.tooLong(16)),
  quantity: z.coerce
    .number({ message: v.quantity })
    .int(v.quantity)
    .positive(v.quantityPositive),
});

export type IssueForm = z.infer<typeof issueFormSchema>;
/** What the form holds before zod coerces it (numbers typed as text). */
export type IssueFormInput = z.input<typeof issueFormSchema>;
