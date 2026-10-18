import { Suspense } from "react";

import { Loading } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { TemplateDetail } from "./template-detail";

export const metadata = { title: copy.meta.adminTemplate };

export default function Page({ params }: { params: Promise<{ id: string }> }) {
  return (
    <Suspense fallback={<Loading className="px-4 py-8 lg:px-8" />}>
      <TemplateDetail params={params} />
    </Suspense>
  );
}
