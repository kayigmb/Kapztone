import { copy } from "@/lib/consts";
import { ResetForm } from "./reset-form";

export const metadata = { title: copy.meta.reset };

// Drafts 2f (new password, live rules), 2g (link expired or used), 2h (updated).
export default async function Page({ params }: PageProps<"/reset/[token]">) {
  const { token } = await params;
  return <ResetForm token={token} />;
}
