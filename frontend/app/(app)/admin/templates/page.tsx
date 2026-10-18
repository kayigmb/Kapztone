import { Suspense } from "react";

import { Loading } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { Templates } from "./templates";

export const metadata = { title: copy.meta.templates };

export default function Page() {
  return (
    <Suspense fallback={<Loading className="px-4 py-8 lg:px-8" />}>
      <Templates />
    </Suspense>
  );
}
