"use client";

import { useRecordTitle } from "@/components/chaintrack/record-title";

/** A record's title row: its name, status and actions. The top bar shows where it sits ("List › crumb"). */
export function DetailHeader({
  crumb,
  title,
  badge,
  actions,
}: {
  crumb: string;
  title: React.ReactNode;
  badge?: React.ReactNode;
  actions?: React.ReactNode;
}) {
  useRecordTitle(crumb);
  return (
    <header className="flex flex-wrap items-start justify-between gap-3">
      <div className="flex min-w-0 flex-wrap items-center gap-3">
        <h1 className="text-2xl font-semibold tracking-tight lg:text-page">
          {title}
        </h1>
        {badge}
      </div>
      {actions && (
        <div className="flex flex-wrap items-center gap-2">{actions}</div>
      )}
    </header>
  );
}
