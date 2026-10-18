import { Suspense } from "react";

import { Loading } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { HandoverList } from "./handover-list";

export const metadata = { title: copy.meta.handovers };

// Figma Handovers: incoming / outgoing / done; the tab lives in ?tab=.
export default function Page() {
  return (
    <Suspense fallback={<Loading className="p-4 lg:p-8" />}>
      <HandoverList />
    </Suspense>
  );
}
