import { PageSuspense } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { FlagQueue } from "./flags";

export const metadata = { title: copy.meta.flags };

export default function Page() {
  return (
    <PageSuspense>
      <FlagQueue />
    </PageSuspense>
  );
}
