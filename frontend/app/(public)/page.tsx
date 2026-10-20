import { LandingGate } from "@/components/marketing/landing/gate";
import { Landing } from "@/components/marketing/landing/landing";
import { copy } from "@/lib/consts";

export const metadata = {
  title: copy.meta.landing,
  description: copy.meta.landingDescription,
};

// The public home: what ChainTrack is, and a way to verify a label.
export default function Page() {
  return (
    <LandingGate>
      <Landing />
    </LandingGate>
  );
}
