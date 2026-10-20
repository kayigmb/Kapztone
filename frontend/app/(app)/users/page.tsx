import { PageSuspense } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { Users } from "./users";

export const metadata = { title: copy.meta.users };

// The platform admin sees every account; an organisation admin manages their own members.
export default function Page() {
  return (
    <PageSuspense>
      <Users />
    </PageSuspense>
  );
}
