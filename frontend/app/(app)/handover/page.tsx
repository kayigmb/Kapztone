import { Suspense } from "react";

import { Loading } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { HandoverFlow } from "./handover";

export const metadata = { title: copy.meta.handover };

// The lot can come from ?passport= (a scan, or "Hand over" on the lot), so
// the flow reads the search params on the client.
export default function Page() {
  return (
    <Suspense fallback={<Loading className="px-4 py-8 lg:px-8" />}>
      <HandoverFlow />
    </Suspense>
  );
}
