import { copy } from "@/lib/consts";
import { VerificationView } from "./verification-view";

export const metadata = { title: copy.meta.passport };

export default async function Page({ params }: PageProps<"/p/[id]">) {
  const { id } = await params;
  return <VerificationView id={id} />;
}
