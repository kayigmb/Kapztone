import { copy } from "@/lib/consts";
import { ResetForm } from "./reset-form";

export const metadata = { title: copy.meta.reset };

export default async function Page({ params }: PageProps<"/reset/[token]">) {
  const { token } = await params;
  return <ResetForm token={token} />;
}
