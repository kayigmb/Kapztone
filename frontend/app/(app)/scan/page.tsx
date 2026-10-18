import { copy } from "@/lib/consts";
import { Scan } from "./scan";

export const metadata = { title: copy.meta.scan };

// Scan or type a lot code: the check runs here, inside the app.
export default function Page() {
  return <Scan />;
}
