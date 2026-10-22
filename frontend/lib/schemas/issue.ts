import { z } from "zod";

import { copy } from "../consts.ts";
import type { Facility, Template } from "../api.ts";

const v = copy.validation;

export const issueFormSchema = z.object({
  facility_id: z.string().trim().min(1, v.facility),
  template_id: z.string().trim().min(1, v.template),
  quantity: z.coerce
    .number({ message: v.quantity })
    .int(v.quantity)
    .positive(v.quantityPositive),
});

export type IssueForm = z.infer<typeof issueFormSchema>;
/** What the form holds before zod coerces it (numbers typed as text). */
export type IssueFormInput = z.input<typeof issueFormSchema>;

/** How much of a facility's capacity is still free this period. */
export function capacityLeft(
  facility: Pick<Facility, "capacity" | "capacity_used">,
): number {
  return Math.max(0, facility.capacity - facility.capacity_used);
}

/** Issuing more than the facility's remaining capacity is refused by the API with the numbers in the message; this mirrors that limit so the form can block the button and say why before the PIN is asked for. */
export function issueBlock(
  facility: Pick<Facility, "capacity" | "capacity_used">,
  quantity: number,
): { over: true; left: number; capacity: number; capacityUsed: number } | null {
  const left = capacityLeft(facility);
  if (quantity <= left) return null;
  return {
    over: true,
    left,
    capacity: facility.capacity,
    capacityUsed: facility.capacity_used,
  };
}

/** The stages a template supports, in its own order (for the event chips on issue). */
export function templateStages(template: Pick<Template, "stages">): string[] {
  return template.stages.map((s) => s.key);
}
