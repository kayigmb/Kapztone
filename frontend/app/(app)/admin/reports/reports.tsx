"use client";

import { useQuery } from "@tanstack/react-query";
import { CheckCircle2, Download, Lock, Search, ShieldX, WifiOff } from "lucide-react";
import { useState } from "react";

import { Loading, PageState } from "@/components/chaintrack/page-state";
import { StatusChip } from "@/components/chaintrack/status";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { endpoints, problemOf, useMeGuard } from "@/lib/api";
import {
  apiPaths,
  copy,
  permissions,
  queryKeys,
  routes,
} from "@/lib/consts";
import { formatDate, formatDateTime } from "@/lib/verification";

const t = {
  title: "Reports and audit",
  sub: "Volume, audit trail and chain verification.",
  notAllowedTitle: "You can't view reports.",
  notAllowedBody: "You don't have permission to view reports or audit logs.",
  volume: "Volume issued",
  startDate: "Start date",
  endDate: "End date",
  downloadCsv: "Download CSV",
  month: "Month",
  commodity: "Commodity",
  unit: "Unit",
  quantity: "Quantity",
  noData: "No data",
  auditTrail: "Audit trail",
  route: "Route",
  allRoutes: "All routes",
  seq: "Seq",
  created: "When",
  actor: "Actor",
  methodRoute: "Method / Route",
  subject: "Subject",
  status: "Status",
  hash: "Hash",
  chainCheck: "Chain check",
  runCheck: "Run chain check",
  running: "Running…",
  verified: "Every entry's hash matches",
  checked: (n: number) => `${n} entries checked`,
  brokenAt: (n: number | null) => (n ? `Broken at seq ${n}` : "Broken"),
};

const routeOptions = [
  { value: "", label: t.allRoutes },
  { value: "/passports", label: "/passports" },
  { value: "/handovers", label: "/handovers" },
  { value: "/flags", label: "/flags" },
  { value: "/attestations", label: "/attestations" },
  { value: "/orgs", label: "/orgs" },
];

export function Reports() {
  const me = useMeGuard((me) => (me ? undefined : routes.signin));
  const allowed = me.ready && (me.data!.permissions.includes(permissions.auditRead) || me.data!.permissions.includes(permissions.reportRead));

  const [volumeStart, setVolumeStart] = useState("");
  const [volumeEnd, setVolumeEnd] = useState("");
  const [auditStart, setAuditStart] = useState("");
  const [auditEnd, setAuditEnd] = useState("");
  const [auditRoute, setAuditRoute] = useState("");

  const volume = useQuery({
    queryKey: [...queryKeys.volume, volumeStart, volumeEnd],
    queryFn: () => endpoints.volumeReport({ startDate: volumeStart || undefined, endDate: volumeEnd || undefined }),
    enabled: allowed,
  });

  const audit = useQuery({
    queryKey: [...queryKeys.audit, auditStart, auditEnd, auditRoute],
    queryFn: () => endpoints.auditList({ startDate: auditStart || undefined, endDate: auditEnd || undefined, route: auditRoute || undefined }),
    enabled: allowed,
  });

  const chain = useQuery({
    queryKey: ["audit", "verify"],
    queryFn: () => endpoints.auditVerify(),
    enabled: false,
  });

  if (me.isError) {
    return (
      <PageState icon={WifiOff} tone="danger" title={copy.errors.loadTitle} retry={() => me.refetch()}>
        {problemOf(me.error)}
      </PageState>
    );
  }
  if (!me.ready) return <Loading className="px-4 py-8 lg:px-8" />;
  if (!allowed) {
    return (
      <PageState icon={Lock} title={t.notAllowedTitle}>
        {t.notAllowedBody}
      </PageState>
    );
  }



  return (
    <main className="grid content-start gap-6 px-4 py-6 lg:px-8 lg:py-8">
      <header className="grid gap-1">
        <h1 className="text-[22px] font-semibold lg:text-page">{t.title}</h1>
        <p className="text-sm text-ink-muted">{t.sub}</p>
      </header>

      <section className="grid gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
        <h2 className="text-section">{t.volume}</h2>
        <div className="flex flex-wrap gap-3">
          <div className="grid gap-1.5">
            <span className="text-sm font-semibold">{t.startDate}</span>
            <Input type="date" value={volumeStart} onChange={(e) => setVolumeStart(e.target.value)} className="min-h-tap w-48" />
          </div>
          <div className="grid gap-1.5">
            <span className="text-sm font-semibold">{t.endDate}</span>
            <Input type="date" value={volumeEnd} onChange={(e) => setVolumeEnd(e.target.value)} className="min-h-tap w-48" />
          </div>
          <a href={`/api/reports/volume?format=csv${volumeStart ? `&start_date=${volumeStart}` : ""}${volumeEnd ? `&end_date=${volumeEnd}` : ""}`} className="inline-flex min-h-tap items-center gap-2 rounded-md bg-primary px-4 py-2 font-semibold text-on-primary">
            <Download className="size-4" aria-hidden />
            {t.downloadCsv}
          </a>
        </div>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-border text-sm">
            <thead className="bg-surface-sunken">
              <tr>
                <th className="px-3 py-2 text-left font-semibold">{t.month}</th>
                <th className="px-3 py-2 text-left font-semibold">{t.commodity}</th>
                <th className="px-3 py-2 text-left font-semibold">{t.unit}</th>
                <th className="px-3 py-2 text-right font-semibold">{t.quantity}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {volume.isPending && (
                <tr><td colSpan={4} className="px-3 py-4"><Loading /></td></tr>
              )}
              {volume.isError && (
                <tr><td colSpan={4} className="px-3 py-4 text-danger-ink">{problemOf(volume.error)}</td></tr>
              )}
              {volume.data?.length === 0 && !volume.isPending && (
                <tr><td colSpan={4} className="px-3 py-4 text-ink-muted">{t.noData}</td></tr>
              )}
              {volume.data?.map((row, i) => (
                <tr key={i}>
                  <td className="px-3 py-2 font-mono">{row.month}</td>
                  <td className="px-3 py-2">{row.commodity}</td>
                  <td className="px-3 py-2 font-mono">{row.unit}</td>
                  <td className="px-3 py-2 text-right font-mono">{row.quantity.toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className="grid gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
        <h2 className="text-section">{t.auditTrail}</h2>
        <div className="flex flex-wrap gap-3">
          <div className="grid gap-1.5">
            <span className="text-sm font-semibold">{t.startDate}</span>
            <Input type="date" value={auditStart} onChange={(e) => setAuditStart(e.target.value)} className="min-h-tap w-48" />
          </div>
          <div className="grid gap-1.5">
            <span className="text-sm font-semibold">{t.endDate}</span>
            <Input type="date" value={auditEnd} onChange={(e) => setAuditEnd(e.target.value)} className="min-h-tap w-48" />
          </div>
          <div className="grid gap-1.5">
            <span className="text-sm font-semibold">{t.route}</span>
            <Select items={routeOptions} value={auditRoute} onValueChange={(v) => setAuditRoute(v ?? "")}>
              <SelectTrigger className="min-h-tap w-48">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {routeOptions.map((o) => (
                  <SelectItem key={o.value} value={o.value}>
                    {o.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-border text-sm">
            <thead className="bg-surface-sunken">
              <tr>
                <th className="px-2 py-2 text-left font-semibold">{t.seq}</th>
                <th className="px-2 py-2 text-left font-semibold">{t.created}</th>
                <th className="px-2 py-2 text-left font-semibold">{t.actor}</th>
                <th className="px-2 py-2 text-left font-semibold">{t.methodRoute}</th>
                <th className="px-2 py-2 text-left font-semibold">{t.subject}</th>
                <th className="px-2 py-2 text-right font-semibold">{t.status}</th>
                <th className="px-2 py-2 text-left font-semibold">{t.hash}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {audit.isPending && (
                <tr><td colSpan={7} className="px-2 py-4"><Loading /></td></tr>
              )}
              {audit.isError && (
                <tr><td colSpan={7} className="px-2 py-4 text-danger-ink">{problemOf(audit.error)}</td></tr>
              )}
              {(audit.data?.rows ?? []).length === 0 && !audit.isPending && (
                <tr><td colSpan={7} className="px-2 py-4 text-ink-muted">{t.noData}</td></tr>
              )}
              {audit.data?.rows.map((row) => (
                <tr key={row.id}>
                  <td className="px-2 py-2 font-mono">{row.seq}</td>
                  <td className="px-2 py-2 text-ink-muted">{formatDateTime(row.created_at)}</td>
                  <td className="px-2 py-2 truncate max-w-[160px]">{row.actor_email || row.actor_user_id || "-"}</td>
                  <td className="px-2 py-2">{row.method} {row.route}</td>
                  <td className="px-2 py-2 font-mono truncate max-w-[140px]" title={row.subject_id ?? ""}>{row.subject_id ? row.subject_id.slice(0, 12) + (row.subject_id.length > 12 ? "…" : "") : "-"}</td>
                  <td className="px-2 py-2 text-right font-mono">{row.status}</td>
                  <td className="px-2 py-2 font-mono text-ink-muted truncate max-w-[120px]" title={row.hash}>{row.hash.slice(0, 10)}…</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className="grid gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
        <h2 className="text-section">{t.chainCheck}</h2>
        <div className="flex flex-wrap items-center gap-3">
          <Button onClick={() => chain.refetch()} disabled={chain.isFetching}>
            <Search className="size-4" aria-hidden />
            {chain.isFetching ? t.running : t.runCheck}
          </Button>
          {chain.data && chain.data.ok && (
            <StatusChip tone="verified" icon={CheckCircle2}>{t.verified} • {t.checked(chain.data.checked)}</StatusChip>
          )}
          {chain.data && !chain.data.ok && (
            <StatusChip tone="danger" icon={ShieldX}>{t.brokenAt(chain.data.broken_at)} • {t.checked(chain.data.checked)}</StatusChip>
          )}
        </div>
      </section>
    </main>
  );
}

function apiBase() {
  return "";
}
