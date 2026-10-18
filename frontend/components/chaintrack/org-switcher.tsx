"use client";

import { useQuery } from "@tanstack/react-query";
import { ArrowLeftRight, Building2, Check } from "lucide-react";

import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { endpoints, type Me } from "@/lib/api";
import { orgAddress } from "@/lib/address";
import { copy, queryKeys, routes } from "@/lib/consts";
import { humanize } from "@/lib/verification";

const t = copy.shell;

/**
 * The organisations this person belongs to; picking one opens that
 * organisation's own address. Sessions are one per address, so they sign in there.
 */
export function OrgSwitcher({ me, onNavy }: { me: Me; onNavy?: boolean }) {
  const orgs = useQuery({ queryKey: queryKeys.myOrganisations, queryFn: endpoints.myOrganisations });
  const others = (orgs.data ?? []).filter((o) => o.id !== me.organisation?.id);
  if (others.length === 0) return null;
  const here = () => window.location;
  const current = me.organisation?.slug;

  return (
    <Popover>
      <PopoverTrigger className={onNavy
        ? "flex min-h-tap w-full items-center gap-2 rounded-md px-2 text-sm font-semibold text-sidebar-accent-fg hover:bg-sidebar-hover"
        : "flex min-h-tap w-full items-center gap-2 rounded-md px-2 text-sm font-semibold text-primary hover:bg-surface-sunken"}>
        <ArrowLeftRight className="size-4 shrink-0" aria-hidden />
        {t.switchOrg}
      </PopoverTrigger>
      <PopoverContent align="start" className="w-72 gap-1 p-2">
        <p className="px-2 pt-1 text-xs text-ink-muted">{t.switchNote}</p>
        <ul className="grid gap-1">
          {me.organisation && (
            <li className="flex min-h-tap items-center gap-2 rounded-md bg-surface-sunken px-2 text-sm">
              <Check className="size-4 shrink-0 text-verified" aria-hidden />
              <span className="min-w-0 flex-1 truncate font-semibold">{me.organisation.name}</span>
              <span className="text-xs text-ink-muted">{t.current}</span>
            </li>
          )}
          {others.map((o) => (
            <li key={o.id}>
              <a href={orgAddress(o.slug, here(), current, routes.signin)}
                className="flex min-h-tap items-center gap-2 rounded-md px-2 text-sm hover:bg-surface-sunken">
                <Building2 className="size-4 shrink-0 text-ink-muted" aria-hidden />
                <span className="min-w-0 flex-1 truncate">{o.name}</span>
                <span className="text-xs text-ink-muted">{humanize(o.role)}</span>
              </a>
            </li>
          ))}
        </ul>
      </PopoverContent>
    </Popover>
  );
}
