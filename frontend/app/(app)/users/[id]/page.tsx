import { PageSuspense } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { UserDetail } from "./user-detail";

export const metadata = { title: copy.meta.userDetail };

// The platform admin's view of one account: memberships, activity, and switching it off.
export default function Page({ params }: { params: Promise<{ id: string }> }) {
  return (
    <PageSuspense>
      <UserDetail params={params} />
    </PageSuspense>
  );
}
