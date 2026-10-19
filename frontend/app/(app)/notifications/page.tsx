import { Suspense } from "react";

import { Loading } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { Notifications } from "./notifications";

export const metadata = { title: copy.meta.notifications };

// Every notification, paged; opening one marks it read and goes to what it's about.
export default function Page() {
  return (
    <Suspense fallback={<Loading className="p-4 lg:p-8" />}>
      <Notifications />
    </Suspense>
  );
}
