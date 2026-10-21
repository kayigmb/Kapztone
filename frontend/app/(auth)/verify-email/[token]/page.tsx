import { copy } from "@/lib/consts";
import { VerifyConfirm } from "./verify-confirm";

export const metadata = { title: copy.meta.verifyEmail };

// Opened from the email sent at sign-up (or on signing in before confirming).
export default async function Page({
  params,
}: PageProps<"/verify-email/[token]">) {
  const { token } = await params;
  return <VerifyConfirm token={token} />;
}
