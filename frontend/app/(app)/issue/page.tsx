import { Suspense } from "react";

import { Loading } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { Issue } from "./issue";

export const metadata = { title: copy.meta.issue };

export default function Page() {
  return (
    <Suspense fallback={<Loading className="p-4 lg:p-8" />}>
      <Issue />
    </Suspense>
  );
}
