import { copy } from "@/lib/consts";
import { Dashboard } from "./dashboard";

export const metadata = { title: copy.brand };

export default function Page() {
  return <Dashboard />;
}
