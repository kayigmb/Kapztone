import dynamic from "next/dynamic";

import { Hero } from "./hero";
import { Nav } from "./nav";
import { SectionSkeleton } from "./shared";
import { VerifyStrip } from "./verify-strip";

// Below the fold: each section ships in its own chunk and streams in behind a skeleton.
const HowItWorks = dynamic(
  () => import("./how-it-works").then((m) => m.HowItWorks),
  {
    loading: () => <SectionSkeleton />,
  },
);
const RoleGuide = dynamic(
  () => import("./role-guide").then((m) => m.RoleGuide),
  {
    loading: () => <SectionSkeleton />,
  },
);
const Showcase = dynamic(() => import("./showcase").then((m) => m.Showcase), {
  loading: () => <SectionSkeleton />,
});
const TrustVerdicts = dynamic(
  () => import("./trust").then((m) => m.TrustVerdicts),
  { loading: () => <SectionSkeleton /> },
);
const Security = dynamic(() => import("./security").then((m) => m.Security), {
  loading: () => <SectionSkeleton />,
});
const Faq = dynamic(() => import("./faq").then((m) => m.Faq), {
  loading: () => <SectionSkeleton />,
});
const FinalCta = dynamic(() => import("./final-cta").then((m) => m.FinalCta), {
  loading: () => <SectionSkeleton />,
});
const Footer = dynamic(() => import("./footer").then((m) => m.Footer), {
  loading: () => <SectionSkeleton />,
});

export function Landing() {
  return (
    <div className="min-h-dvh bg-bg text-ink">
      <Nav />
      <main>
        <Hero />
        <VerifyStrip />
        <HowItWorks />
        <RoleGuide />
        <Showcase />
        <TrustVerdicts />
        <Security />
        <Faq />
        <FinalCta />
      </main>
      <Footer />
    </div>
  );
}
