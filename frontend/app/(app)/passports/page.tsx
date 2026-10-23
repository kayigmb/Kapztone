import { PageSuspense } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { Passports } from "./passports";

export const metadata = { title: copy.meta.passports };

export default function Page() {
  return (
    <PageSuspense>
      <Passports />
    </PageSuspense>
  );
}
