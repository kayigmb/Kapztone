import { copy } from "@/lib/consts";
import { WorkspacePicker } from "../dashboard/workspace-picker";

export const metadata = {
  title: `${copy.shell.nav.workspaces} · ${copy.brand}`,
};

export default function Page() {
  return <WorkspacePicker />;
}
