"use client";

import { WifiOff } from "lucide-react";

import { AppShell } from "@/components/chaintrack/app-shell";
import { LoadingWithLogo, PageState } from "@/components/chaintrack/page-state";
import { problemOf, useMeGuard } from "@/lib/api";
import { copy, orgStatus, routes } from "@/lib/consts";

// Every signed-in page sits in the shell.
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
  if (!me.ready) return <LoadingWithLogo />;
  return <AppShell me={me.data!}>{children}</AppShell>;
}
