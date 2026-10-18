import { copy } from "@/lib/consts";
import { IssueCertificate } from "./issue";

export const metadata = { title: copy.meta.certificateNew };

// Attaching a signed document to a lot or a facility: form on the left, signing on the right.
export default function Page() {
  return <IssueCertificate />;
}
