import { copy } from "@/lib/consts";
import { HandoverView } from "./handover-view";

export const metadata = { title: copy.meta.handoverDetail };

export default async function Page({ params }: PageProps<"/handovers/[id]">) {
  const { id } = await params;
  return <HandoverView id={id} />;
}
