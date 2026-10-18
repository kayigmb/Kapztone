import { copy } from "@/lib/consts";
import { Certificates } from "./certificates";

export const metadata = { title: copy.meta.certificates };

// The certificates this organisation has issued, with revoke and the way into the issue flow.
export default function Page() {
  return <Certificates />;
}
