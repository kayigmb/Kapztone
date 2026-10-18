import { Suspense } from "react";

import { Loading } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { Reports } from "./reports";

export const metadata = { title: copy.meta.reports };

export default function Page() {
  return (
    <Suspense fallback={<Loading className="px-4 py-8 lg:px-8" />}>
      <Reports />
    </Suspense>
  );
}
