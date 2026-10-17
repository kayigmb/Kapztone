import { Building2 } from "lucide-react";
import { redirect } from "next/navigation";

import { Notice } from "@/components/chaintrack/notice";
import { KeyValue } from "@/components/chaintrack/passport-card";
import { getMe } from "@/lib/api";
import { humanize } from "@/lib/verification";
import { SignOutButton } from "./sign-out-button";

// A stand-in for the dashboard (screens.md §02): it shows who is signed in
// and for which organisation, which is all the session has to prove so far.
export default async function Home() {
  const me = await getMe();
  if (!me) redirect("/signin");
  // An organisation cannot issue or receive passports until it is verified.
  if (me.organisation && me.organisation.status !== "verified") redirect("/onboarding");

  return (
    <main className="mx-auto grid max-w-content gap-4 px-4 py-8">
      <header className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-[22px] font-semibold lg:text-page">Signed in as {me.user.name}</h1>
          <p className="text-sm text-ink-muted">{me.user.email}</p>
        </div>
        <SignOutButton />
      </header>

      <section className="rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
        {me.organisation ? (
          <div className="grid grid-cols-[44px_1fr] gap-3">
            <span className="grid size-11 place-items-center rounded-md bg-primary-soft text-primary">
              <Building2 className="size-5" aria-hidden />
            </span>
            <KeyValue facts={[
              { label: "Organisation", value: me.organisation.name, wide: true },
              { label: "Type", value: humanize(me.organisation.kind) },
              { label: "Verification", value: humanize(me.organisation.status) },
              { label: "Your role", value: humanize(me.role) },
            ]} />
          </div>
        ) : me.role ? (
          <KeyValue facts={[{ label: "Your role", value: humanize(me.role), wide: true }]} />
        ) : (
          <Notice kind="info">
            <strong>No organisation at this address.</strong> Open your organisation&apos;s own address
            (its name comes first, like <span className="font-mono">acme.</span> before this site&apos;s name) to act for it.
          </Notice>
        )}
      </section>
    </main>
  );
}
