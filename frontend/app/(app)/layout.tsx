"use client";

import { Lock, WifiOff } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useRef } from "react";

import { AppShell } from "@/components/chaintrack/app-shell";
import { LoadingWithLogo, PageState } from "@/components/chaintrack/page-state";
import { buttonVariants } from "@/components/ui/button";
import { problemOf, switchOrganisation, useMeGuard } from "@/lib/api";
import { copy, orgStatus, routes, superadminRole } from "@/lib/consts";
import { canOpen } from "@/lib/dashboard";

// Every signed-in page sits in the shell.
export default function Layout({ children }: LayoutProps<"/">) {
  const pathname = usePathname();
  const me = useMeGuard((me) =>
    !me
      ? routes.signin
      : me.organisation &&
          me.organisation.status !== orgStatus.verified &&
          me.role !== superadminRole
        ? routes.onboarding
        : undefined,
  );
  // An organisation's address this person can't act for: back to the main address to pick one of their own.
  const elsewhere =
    me.data?.address_slug && !me.data.organisation
      ? me.data.address_slug
      : undefined;
  const leaving = useRef(false);
  useEffect(() => {
    if (!elsewhere || leaving.current) return;
    leaving.current = true;
    switchOrganisation(null, elsewhere).catch(() => {
      leaving.current = false;
    });
  }, [elsewhere]);

  if (me.isError) {
    return (
      <PageState
        icon={WifiOff}
        tone="danger"
        title={copy.errors.loadTitle}
        retry={() => me.refetch()}
      >
        {problemOf(me.error)}
      </PageState>
    );
  }
  if (!me.ready || elsewhere) return <LoadingWithLogo />;
  return (
    <AppShell me={me.data!}>
      {canOpen(pathname, me.data!) ? (
        children
      ) : (
        <PageState icon={Lock} title={copy.errors.noAccessTitle}>
          <span className="grid justify-items-center gap-4">
            {copy.errors.noAccessBody}
            <Link
              href={routes.home}
              className={buttonVariants({ variant: "secondary" })}
            >
              {copy.errors.backToDashboard}
            </Link>
          </span>
        </PageState>
      )}
    </AppShell>
  );
}
