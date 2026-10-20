import { PageSuspense } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { HandoverList } from "./handover-list";

export const metadata = { title: copy.meta.handovers };

// Figma Handovers: incoming / outgoing / done; the tab lives in ?tab=.
export default function Page() {
  return (
    <PageSuspense>
      <HandoverList />
    </PageSuspense>
  );
}
