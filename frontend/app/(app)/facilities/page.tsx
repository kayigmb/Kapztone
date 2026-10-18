import { Suspense } from "react";

import { Loading } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { Facilities } from "./facilities";

export const metadata = { title: copy.meta.facilities };

// The selected facility lives in the address (?facility=), so the screen reads it on the client.
export default function Page() {
  return (
    <Suspense fallback={<Loading className="px-4 py-8 lg:px-8" />}>
      <Facilities />
    </Suspense>
  );
}
