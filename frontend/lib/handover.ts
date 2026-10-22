export const defaultTolerancePct = 1;

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

type StageLike = {
  key: string;
  events?: string[];
  handoverTolerance?: number | null;
};

/** The stages a lot can be handed over into: those that allow HANDOVER. */
export const handoverStages = <S extends StageLike>(stages: S[]) =>
  stages.filter((s) => s.events?.includes("HANDOVER"));

export const toleranceOf = (stages: StageLike[], key: string) =>
  stages.find((s) => s.key === key)?.handoverTolerance ?? defaultTolerancePct;

/** Sending less than the whole passport splits it: the rest stays with the sender. */
export const isPartial = (quantity: number, remaining: number) =>
  quantity > 0 && quantity < remaining;
