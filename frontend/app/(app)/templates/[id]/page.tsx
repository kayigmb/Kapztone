import { PageSuspense } from "@/components/chaintrack/page-state";
import { copy } from "@/lib/consts";
import { TemplateDetail } from "./template-detail";

export const metadata = { title: copy.meta.adminTemplate };

export default function Page({ params }: { params: Promise<{ id: string }> }) {
  return (
    <PageSuspense>
      <TemplateDetail params={params} />
    </PageSuspense>
  );
}
