import { LandingGate } from "@/components/marketing/landing/gate";
import { Landing } from "@/components/marketing/landing/landing";
import { copy } from "@/lib/consts";
import { indexed, siteURL } from "@/lib/site";

export const metadata = {
  title: copy.meta.landing,
  description: copy.meta.landingDescription,
  ...indexed("/"),
};

export default function Page() {
  const site = siteURL();
  // What search engines read about the site: who runs it and where it lives.
  const structured = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "Organization",
        name: copy.brand,
        url: site,
        logo: `${site}/brand/chaintrack-mark.svg`,
      },
      {
        "@type": "WebSite",
        name: copy.brand,
        url: site,
        description: copy.meta.landingDescription,
      },
    ],
  };
  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(structured) }}
      />
      <LandingGate>
        <Landing />
      </LandingGate>
    </>
  );
}
