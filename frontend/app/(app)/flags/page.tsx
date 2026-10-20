import { PageSuspense } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { FlagQueue } from "./flags";

export const metadata = { title: copy.meta.flags };

// Draft 3k (Figma Flags): queue on the left, investigation on the right; ?tab= and ?flag= live in the address.
export default function Page() {
  return (
    <PageSuspense>
      <FlagQueue />
    </PageSuspense>
  );
}
