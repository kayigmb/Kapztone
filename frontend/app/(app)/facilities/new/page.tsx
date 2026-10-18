import { copy } from "@/lib/consts";
import { Register } from "../register";

export const metadata = { title: copy.meta.facilityNew };

// Registering a facility is a plain form: no PIN, nothing is signed.
export default function Page() {
  return <Register />;
}
