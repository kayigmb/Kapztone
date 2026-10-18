import { copy } from "@/lib/consts";
import { HandoverView } from "./handover-view";

export const metadata = { title: copy.meta.handoverDetail };

// Drafts 1f (accept: weigh, inspect, sign) and 3h (reject with reasons, maybe a flag); the sender can cancel.
export default async function Page({ params }: PageProps<"/handovers/[id]">) {
  const { id } = await params;
  return <HandoverView id={id} />;
}
