"use client";

import { ArrowRight, Info } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { Segmented } from "@/components/chaintrack/segmented";
import { buttonVariants } from "@/components/ui/button";
import { marketing as m } from "@/lib/consts";
import { type MarketingRole, marketingRoles } from "@/lib/consts";
import { routes } from "@/lib/consts";
import { cn } from "@/lib/utils";
import { Container, SectionHeading, roleIcon } from "./shared";

export function RoleGuide() {
  const r = m.roles;
  const [role, setRole] = useState<MarketingRole>(marketingRoles[0]);
  const guide = r.guides[role];
  const RoleIcon = roleIcon[role];
  return (
    <section className="border-y border-border bg-surface-sunken py-20">
      <Container className="grid gap-8">
        <div className="flex flex-wrap items-end justify-between gap-6">
          <SectionHeading eyebrow={r.eyebrow} title={r.title} />
          <p className="max-w-md text-sm text-ink-muted">{r.body}</p>
        </div>
        <Segmented
          label={r.title}
          value={role}
          onChange={setRole}
          options={marketingRoles.map((x) => ({ value: x, label: r.tabs[x] }))}
        />
        <div className="grid gap-6 rounded-2xl border border-border bg-surface p-6 shadow-card lg:grid-cols-3 lg:p-8">
          <div className="grid content-start gap-3">
            <span className="grid size-11 place-items-center rounded-lg bg-primary-soft text-primary">
              <RoleIcon className="size-5" aria-hidden />
            </span>
            <h3 className="text-2xl font-semibold">{guide.title}</h3>
            <p className="text-sm text-ink-muted">{guide.body}</p>
            <Link
              href={routes.signup}
              className={cn(buttonVariants({ variant: "secondary" }), "w-fit")}
            >
              {r.cta}
              <ArrowRight aria-hidden />
            </Link>
          </div>
          <ol className="grid content-start gap-4">
            {guide.steps.map(([title, body], i) => (
              <li key={title} className="flex gap-3">
                <span className="grid size-8 shrink-0 place-items-center rounded-full bg-primary text-sm font-semibold text-on-primary">
                  {i + 1}
                </span>
                <div className="grid gap-0.5">
                  <strong className="text-sm">{title}</strong>
                  <p className="text-sm text-ink-muted">{body}</p>
                </div>
              </li>
            ))}
          </ol>
          <aside className="flex h-fit gap-3 rounded-xl bg-primary-soft p-4 text-sm">
            <Info className="mt-0.5 size-5 shrink-0 text-primary" aria-hidden />
            <div className="grid gap-1">
              <strong>{r.tipTitle}</strong>
              <p className="text-ink-muted">{r.tipBody}</p>
            </div>
          </aside>
        </div>
      </Container>
    </section>
  );
}

