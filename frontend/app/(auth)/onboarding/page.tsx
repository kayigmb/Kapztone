import { copy } from "@/lib/consts";
import { Onboarding } from "./onboarding";

export const metadata = { title: copy.meta.onboarding };

export default function Page() {
  return <Onboarding />;
}
