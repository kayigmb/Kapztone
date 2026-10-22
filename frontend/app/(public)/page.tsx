import { LandingGate } from "@/components/marketing/landing/gate";
import { Landing } from "@/components/marketing/landing/landing";
import { copy } from "@/lib/consts";

export const metadata = {
  title: copy.meta.landing,
  description: copy.meta.landingDescription,
};

export default function Page() {
  return (
    <LandingGate>
      <Landing />
    </LandingGate>
  );
}
