import { PageSuspense } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { Issue } from "./issue";

export const metadata = { title: copy.meta.issue };

export default function Page() {
  return (
    <PageSuspense>
      <Issue />
    </PageSuspense>
  );
}
