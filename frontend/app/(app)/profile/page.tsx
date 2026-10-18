import { copy } from "@/lib/consts";
import { Profile } from "./profile";

export const metadata = { title: copy.meta.profile };

// Everyone's own account: details, password, signing PIN, two-factor sign-in.
export default function Page() {
  return <Profile />;
}
