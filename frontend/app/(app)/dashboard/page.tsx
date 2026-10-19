import { copy } from "@/lib/consts";
import { Dashboard } from "./dashboard";

export const metadata = { title: copy.brand };

// Screen 3 (screens.md → Dashboard): one page, laid out per kind of organisation (lib/consts.ts → dashboards).
export default function Page() {
  return <Dashboard />;
}
