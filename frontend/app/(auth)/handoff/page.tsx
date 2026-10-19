import { copy } from "@/lib/consts";
import { HandoffAccept } from "./handoff-accept";

// The code is a sign-in for one minute: never sent on to another site.
export const metadata = { title: copy.meta.handoff, referrer: "no-referrer" };

// Reached from the organisation switcher or the workspace picker on another address.
export default async function Page({ searchParams }: PageProps<"/handoff">) {
  const { code } = await searchParams;
  return <HandoffAccept code={typeof code === "string" ? code : ""} />;
}
