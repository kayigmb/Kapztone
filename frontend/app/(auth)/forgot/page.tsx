import { copy } from "@/lib/consts";
import { ForgotForm } from "./forgot-form";

export const metadata = { title: copy.meta.forgot };

export default function Page() {
  return <ForgotForm />;
}
