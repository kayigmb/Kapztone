import { PageSuspense } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { OrgPage } from "./org-page";

export const metadata = { title: copy.meta.adminOrg };

// The platform admin's full view of one organisation: facts, people, records, and switching it off.
export default function Page({ params }: { params: Promise<{ id: string }> }) {
  return (
    <PageSuspense>
      <OrgPage params={params} />
    </PageSuspense>
  );
}
