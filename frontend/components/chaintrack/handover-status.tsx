import { CircleCheck, Clock, OctagonX, TriangleAlert } from "lucide-react";

import type { HandoverDetail } from "@/lib/api";
import { copy, handoverStatus } from "@/lib/consts";
import { StatusChip } from "./status";

const t = copy.handover;

/**
 * Where a handover stands, from the reader's side. "Signed by both parties"
 * only once the receiver has signed (hard rule 8).
 */
export function HandoverStatusChip({ h, mine }: { h: Pick<HandoverDetail, "status" | "to_org">; mine?: string }) {
  switch (h.status) {
    case handoverStatus.completed:
      return <StatusChip tone="verified" icon={CircleCheck}>{t.signedByBoth}</StatusChip>;
    case handoverStatus.rejected:
      return <StatusChip tone="danger" icon={OctagonX}>{t.rejectedByReceiver}</StatusChip>;
    case handoverStatus.cancelled:
      return <StatusChip tone="neutral" icon={OctagonX}>{copy.handover.cancelled}</StatusChip>;
    case handoverStatus.expired:
      return <StatusChip tone="warning" icon={TriangleAlert}>{copy.handoverStatusLabel.expired}</StatusChip>;
    default:
      return (
        <StatusChip tone="pending" icon={Clock}>
          {h.to_org === mine ? copy.dashboard.tiles.incoming.note : t.awaitingReceiver}
        </StatusChip>
      );
  }
}
