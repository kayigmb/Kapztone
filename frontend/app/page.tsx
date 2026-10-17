import { copy } from "@/lib/consts";
import { Home } from "./home";

export const metadata = { title: copy.brand };

export default function Page() {
  return <Home />;
}
