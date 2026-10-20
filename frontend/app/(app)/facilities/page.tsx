import { PageSuspense } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { Facilities } from "./facilities";

export const metadata = { title: copy.meta.facilities };

// The selected facility lives in the address (?facility=), so the screen reads it on the client.
export default function Page() {
  return (
    <PageSuspense>
      <Facilities />
    </PageSuspense>
  );
}
