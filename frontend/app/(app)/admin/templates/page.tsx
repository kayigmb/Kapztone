import { PageSuspense } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { Templates } from "./templates";

export const metadata = { title: copy.meta.templates };

export default function Page() {
  return (
    <PageSuspense>
      <Templates />
    </PageSuspense>
  );
}
