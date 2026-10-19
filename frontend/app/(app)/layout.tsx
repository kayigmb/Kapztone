"use client";

import { WifiOff } from "lucide-react";

import { AppShell } from "@/components/chaintrack/app-shell";
import { Loading, PageState } from "@/components/chaintrack/page-state";
import { problemOf, useMeGuard } from "@/lib/api";
import { copy, orgStatus, routes } from "@/lib/consts";

// Every signed-in page sits in the shell. Nobody signed in → sign-in; an
// organisation that isn't verified yet → onboarding (it can't issue or receive).
export default function Layout({ children }: LayoutProps<"/">) {
  const me = useMeGuard((me) =>
    !me
      ? routes.signin
      : me.organisation && me.organisation.status !== orgStatus.verified
        ? routes.onboarding
        : undefined,
  );

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
  if (!me.ready) return <Loading className="px-4 py-8 lg:p-8" />;
  return <AppShell me={me.data!}>{children}</AppShell>;
}
