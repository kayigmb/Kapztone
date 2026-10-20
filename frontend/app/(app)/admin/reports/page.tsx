import { PageSuspense } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { Reports } from "./reports";

export const metadata = { title: copy.meta.reports };

export default function Page() {
  return (
    <PageSuspense>
      <Reports />
    </PageSuspense>
  );
}
