/** The weight difference, in percent, every handover accepts (backend DefaultHandoverTolerance). */
export const tolerancePct = 1;

export type Weighing = {
  difference: number; // received − sent; negative = short
  pct: number; // |difference| / sent, in percent, one decimal
  within: boolean; // inside the tolerance: the receiver may accept
};

/** How the weight the receiver recorded compares with what was sent. */
export function weigh(
  sent: number,
  received: number,
  tolerancePct: number,
): Weighing {
  const difference = received - sent;
  const pct =
    sent > 0 ? Math.round((Math.abs(difference) / sent) * 1000) / 10 : 0;
  return {
    difference,
    pct,
    within: Math.abs(difference) <= (sent * tolerancePct) / 100,
  };
}

/** Sending less than the whole passport splits it: the rest stays with the sender. */
export const isPartial = (quantity: number, remaining: number) =>
  quantity > 0 && quantity < remaining;
