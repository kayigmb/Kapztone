import { copy } from "@/lib/consts";
import { PassportDetail } from "./passport-detail";

export const metadata = { title: copy.meta.passportDetail };

// The signed-in business view of one lot: record, actions, journey,
// evidence, flags, genealogy and technical details.
export default function Page() {
  return <PassportDetail />;
}
