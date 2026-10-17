"use client";

import { Building2, ClipboardCheck, WifiOff } from "lucide-react";
import Link from "next/link";

import { Notice } from "@/components/chaintrack/notice";
import { Loading, PageState } from "@/components/chaintrack/page-state";
import { KeyValue } from "@/components/chaintrack/passport-card";
import { buttonVariants } from "@/components/ui/button";
import { problemOf, useMeGuard } from "@/lib/api";
import { copy, orgStatus, permissions, routes } from "@/lib/consts";
import { humanize } from "@/lib/verification";
import { SignOutButton } from "./sign-out-button";

const t = copy.home;

// A stand-in for the dashboard (screens.md §02): it shows who is signed in
// and for which organisation, which is all the session has to prove so far.
export function Home() {
  // An organisation cannot issue or receive passports until it is verified.
  const me = useMeGuard((me) =>
    !me ? routes.signin : me.organisation && me.organisation.status !== orgStatus.verified ? routes.onboarding : undefined);

  if (me.isError) {
    return <PageState icon={WifiOff} tone="danger" title={copy.errors.loadTitle} retry={() => me.refetch()}>{problemOf(me.error)}</PageState>;
  }
  if (!me.ready) return <Loading className="mx-auto max-w-content px-4 py-8" />;
  const { user, organisation, role, permissions: allowed } = me.data!;

  return (
    <main className="mx-auto grid max-w-content gap-4 px-4 py-8">
      <header className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-[22px] font-semibold lg:text-page">{t.signedInAs(user.name)}</h1>
          <p className="text-sm text-ink-muted">{user.email}</p>
        </div>
        <div className="flex flex-wrap gap-2">
          {allowed.includes(permissions.orgVerify) && (
            <Link href={routes.adminOrgs} className={buttonVariants({ variant: "primary" })}>
              <ClipboardCheck aria-hidden />
              {t.adminOrgs}
            </Link>
          )}
          <SignOutButton />
        </div>
      </header>

      <section className="rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
        {organisation ? (
          <div className="grid grid-cols-[44px_1fr] gap-3">
            <span className="grid size-11 place-items-center rounded-md bg-primary-soft text-primary">
              <Building2 className="size-5" aria-hidden />
            </span>
            <KeyValue facts={[
              { label: t.organisation, value: organisation.name, wide: true },
              { label: t.type, value: humanize(organisation.kind) },
              { label: t.verification, value: humanize(organisation.status) },
              { label: t.role, value: humanize(role) },
            ]} />
          </div>
        ) : role ? (
          <KeyValue facts={[{ label: t.role, value: humanize(role), wide: true }]} />
        ) : (
          <Notice kind="info">
            <strong>{t.noOrgTitle}</strong> {t.noOrgBody} <span className="font-mono">{t.noOrgExample}</span> {t.noOrgBodyEnd}
          </Notice>
        )}
      </section>
    </main>
  );
}
