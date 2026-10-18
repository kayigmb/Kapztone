import { copy } from "@/lib/consts";
import { ForgotForm } from "./forgot-form";

export const metadata = { title: copy.meta.forgot };

// Drafts 2d (forgot password) and 2e (check your email, same answer for every email).
export default function Page() {
  return <ForgotForm />;
}
