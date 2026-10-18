import { copy } from "@/lib/consts";
import { LockConfirm } from "./lock-confirm";

export const metadata = { title: copy.meta.lock };

// Opened from the "password changed" email: lock the account if it wasn't you.
export default async function Page({ params }: PageProps<"/lock/[token]">) {
  const { token } = await params;
  return <LockConfirm token={token} />;
}
