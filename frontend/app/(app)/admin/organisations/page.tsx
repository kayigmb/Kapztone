import { PageSuspense } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { Queue } from "./queue";

export const metadata = { title: copy.meta.adminOrgs };

export default function Page() {
  return (
    <PageSuspense>
      <Queue />
    </PageSuspense>
  );
}
