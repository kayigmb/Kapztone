import { PageSuspense } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { Queue } from "./queue";

export const metadata = { title: copy.meta.adminOrgs };

// Screen 13 (screens.md → Admin review queue). The tab and the selected
// organisation live in the address (?tab=&id=), so the queue reads them on the client.
export default function Page() {
  return (
    <PageSuspense>
      <Queue />
    </PageSuspense>
  );
}
