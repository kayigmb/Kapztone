import { copy } from "@/lib/consts";
import { Transform } from "./transform";

export const metadata = { title: copy.meta.transform };

// Turning one or more lots into new lots (process, split or merge).
export default function Page() {
  return <Transform />;
}
