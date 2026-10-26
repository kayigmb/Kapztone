import { z } from "zod";

import { copy, limits, rejectReasons } from "../consts.ts";

const v = copy.validation;
const quantity = z.coerce
  .number({ message: v.quantity })
  .int(v.quantity)
  .positive(v.quantityPositive);

/** The sender: how much leaves (`all` = the whole passport), an optional label for the move, and where / how. */
export const handoverFormSchema = z
  .object({
    all: z.boolean(),
    quantity: quantity.optional(),
    stage: z.string().trim().max(limits.stage, v.tooLong(limits.stage)),
    note: z.string().trim().max(limits.note, v.tooLong(limits.note)),
  })
  .refine((f) => f.all || f.quantity !== undefined, {
    path: ["quantity"],
    message: v.quantityPositive,
  });

export type HandoverForm = z.infer<typeof handoverFormSchema>;
export type HandoverFormInput = z.input<typeof handoverFormSchema>;

/** What leaves, ready for the API; more than the passport holds is the API's refusal to make. */
export function handoverQuantity(
  form: HandoverForm,
  remaining: number,
): number {
  return form.all ? remaining : (form.quantity ?? 0);
}

/** The receiver records what they weighed. */
export const receiveFormSchema = z.object({ received_quantity: quantity });
export type ReceiveForm = z.infer<typeof receiveFormSchema>;
export type ReceiveFormInput = z.input<typeof receiveFormSchema>;

/** Rejecting needs at least one reason from the backend's list. */
export const rejectFormSchema = z.object({
  reasons: z.array(z.enum(rejectReasons)).min(1, copy.receiver.rejectPick),
  note: z.string().trim().max(limits.note, v.tooLong(limits.note)),
  also_flag: z.boolean(),
});
export type RejectForm = z.infer<typeof rejectFormSchema>;
