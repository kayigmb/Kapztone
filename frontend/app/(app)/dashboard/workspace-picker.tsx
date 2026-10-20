"use client";

import { useMutation, useQuery } from "@tanstack/react-query";
import { ArrowRight, Building2 } from "lucide-react";

import { DataWidget } from "@/components/chaintrack/data-widget";
import { Notice } from "@/components/chaintrack/notice";
import { PageHeader } from "@/components/chaintrack/page-header";
import { Button } from "@/components/ui/button";
import { endpoints, switchOrganisation } from "@/lib/api";
import { copy, orgKinds, queryKeys } from "@/lib/consts";
import { humanize } from "@/lib/verification";

const t = copy.dashboard.workspaces;

/** The bare address, for a member: one card per organisation; opening one carries the session there. */
export function WorkspacePicker() {
  const orgs = useQuery({
    queryKey: queryKeys.myOrganisations,
    queryFn: endpoints.myOrganisations,
  });
  const open = useMutation({
    mutationFn: (slug: string) => switchOrganisation(slug),
  });

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <PageHeader
        title={t.title}
        sub={
          orgs.isSuccess && (orgs.data ?? []).length > 0
            ? t.sub(orgs.data!.length)
            : undefined
        }
      />
      <DataWidget query={orgs} retryLabel={copy.errors.tryAgain}>
        {(list) =>
          list.length === 0 ? (
            <Notice kind="info">
              <strong>{t.noneTitle}</strong> {t.noneBody}
            </Notice>
          ) : (
            <ul className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
              {list.map((o) => (
            <li
              key={o.id}
              className="grid gap-4 rounded-lg border border-border bg-surface p-5 shadow-card"
            >
              <div className="flex items-center gap-3">
                <span className="grid size-10 shrink-0 place-items-center rounded-md bg-primary-soft text-primary">
                  <Building2 className="size-5" aria-hidden />
                </span>
                <div className="grid min-w-0">
                  <span className="truncate font-semibold">{o.name}</span>
                  <span className="truncate text-xs text-ink-muted">
                    {orgKinds.find((k) => k.value === o.kind)?.label ??
                      humanize(o.kind)}{" "}
                    · {humanize(o.role)}
                  </span>
                </div>
              </div>
              <Button
                variant="secondary"
                onClick={() => open.mutate(o.slug)}
                loading={open.isPending}
              >
                {t.open}
                <ArrowRight className="size-4" aria-hidden />
                </Button>
              </li>
            ))}
            </ul>
          )
        }
      </DataWidget>
    </main>
  );
}
