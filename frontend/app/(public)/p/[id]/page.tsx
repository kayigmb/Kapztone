import { notFound } from "next/navigation";
import {
  CircleCheck, Coffee, Factory, Gem, Layers, Leaf, OctagonX, Package, Pickaxe, QrCode, Sprout, Warehouse, type LucideIcon,
} from "lucide-react";

import { CertificateTile } from "@/components/chaintrack/certificate-tile";
import { ChecksList } from "@/components/chaintrack/checks-list";
import { JourneyTimeline } from "@/components/chaintrack/journey-timeline";
import { Notice } from "@/components/chaintrack/notice";
import { KeyValue, PassportCard } from "@/components/chaintrack/passport-card";
import { Status } from "@/components/chaintrack/status";
import { TechnicalDetails } from "@/components/chaintrack/technical-details";
import { VerdictBadge } from "@/components/chaintrack/verdict-badge";
import { getVerification } from "@/lib/api";
import { cn } from "@/lib/utils";
import { copy, toView } from "@/lib/verification";

// Icons by the first word of a commodity or facility kind; anything else is a generic lot or site.
const commodityIcons: Record<string, LucideIcon> = { coffee: Coffee, cocoa: Leaf, tin: Gem, gold: Gem, mineral: Gem };
const facilityIcons: Record<string, LucideIcon> = { Farm: Sprout, Mine: Pickaxe, Warehouse: Warehouse };
const firstWord = (s: string) => s.split(/[\s_]/)[0];

const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

function Card({ title, note, className, children }: {
  title: string; note?: string; className?: string; children: React.ReactNode;
}) {
  return (
    <section className={cn("min-w-0 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6", className)}>
      <header className="mb-4 flex flex-wrap items-baseline justify-between gap-x-3">
        <h2 className="text-section">{title}</h2>
        {note && <p className="text-small text-ink-muted">{note}</p>}
      </header>
      {children}
    </section>
  );
}

// Public: no sign-in. Opened by scanning the QR label on the goods.
export default async function Page({ params }: PageProps<"/p/[id]">) {
  const { id } = await params;
  // Anything that is not a passport id is simply not found; the API is not asked.
  const api = uuid.test(id) ? await getVerification(id) : null;
  if (!api) notFound();
  const v = toView(api);
  const failed = v.verdict === "failed";

  return (
    <div className="mx-auto grid max-w-content gap-4 px-4 py-4 md:gap-6 md:py-8">
      <header className="flex items-center justify-between">
        <p className="text-title">ChainTrack</p>
        <p className="flex items-center gap-1.5 text-small text-ink-muted">
          <QrCode className="size-4" aria-hidden />
          Public verification
        </p>
      </header>

      <main className="grid gap-4 md:grid-cols-[1.55fr_1fr] md:gap-6">
        {failed && (
          <div className="md:col-span-2">
            <Notice kind="blocking">
              <strong>{copy.alert.lead}</strong> {copy.alert.rest}
            </Notice>
          </div>
        )}

        <VerdictBadge verdict={v.verdict} summary={v.summary} reasons={v.reasons} meta={v.meta} />

        <PassportCard
          product={v.product.name}
          lotCode={v.product.lotCode}
          icon={commodityIcons[firstWord(v.product.name).toLowerCase()] ?? Package}
          facts={v.product.facts}
          footer={{ left: v.product.issued, verdict: v.verdict }}
          flagged={failed}
        />

        <Card title={copy.sections.origin} className="md:col-span-2">
          <ul className="grid gap-4 md:grid-cols-2">
            {v.origins.map((o) => {
              const Icon = facilityIcons[firstWord(o.kind)] ?? Factory;
              return (
                <li key={o.id} className="grid grid-cols-[40px_1fr] gap-3">
                  <span className="grid size-10 place-items-center rounded-md bg-verified-soft text-verified">
                    <Icon className="size-5" aria-hidden />
                  </span>
                  <KeyValue facts={[
                    { label: "Facility", value: o.name, wide: true },
                    { label: "Type", value: o.kind },
                    { label: "Operated by", value: o.owner },
                  ]} />
                </li>
              );
            })}
          </ul>
        </Card>

        <Card title={copy.sections.journey} note={`${v.journey.length} ${v.journey.length === 1 ? "step" : "steps"}`} className="md:col-span-2">
          <JourneyTimeline steps={v.journey} />
        </Card>

        {v.madeFrom.length > 0 && (
          <Card title={copy.sections.madeFrom} note={`${v.madeFrom.length} ${v.madeFrom.length === 1 ? "lot" : "lots"}, back to origin`} className="md:col-span-2">
            {/* ponytail: a flat list of every ancestor. The drawn tree (components.md → GenealogyTree)
                replaces it when the passport detail screen needs parent-child lines. */}
            <ul className="grid gap-2">
              {v.madeFrom.map((lot) => (
                <li key={lot.lotCode} className="grid grid-cols-[32px_1fr] items-center gap-3 rounded-md border border-border-strong bg-surface px-3 py-2 md:grid-cols-[32px_1fr_auto]">
                  <span className={cn("grid size-8 place-items-center rounded-sm",
                    lot.revoked ? "bg-danger text-on-danger" : lot.origin ? "bg-verified-soft text-verified" : "bg-surface-sunken text-ink-muted")}>
                    <Layers className="size-4" aria-hidden />
                  </span>
                  <div className="min-w-0">
                    <p className="font-mono text-[13px] font-medium break-all">{lot.lotCode}</p>
                    <p className="text-small text-ink-muted">{lot.description}</p>
                  </div>
                  <div className="col-start-2 flex flex-wrap items-center gap-3 md:col-start-3">
                    <span className="fig text-[13px]">{lot.quantity}</span>
                    {lot.revoked ? (
                      <Status tone="danger" icon={OctagonX}>{copy.revoked}</Status>
                    ) : (
                      lot.origin && <Status tone="verified" icon={CircleCheck}>{copy.originMark}</Status>
                    )}
                  </div>
                </li>
              ))}
            </ul>
          </Card>
        )}

        <Card title={copy.sections.certificates} className="md:col-span-2">
          {v.certificates.length === 0 ? (
            <p className="text-small text-ink-muted">{copy.noCertificates}</p>
          ) : (
            <div className="grid gap-4 md:grid-cols-3">
              {v.certificates.map((c) => <CertificateTile key={c.id} certificate={c} />)}
            </div>
          )}
        </Card>

        <Card title={copy.sections.checks}>
          <ChecksList checks={v.checks} />
        </Card>

        <div className="content-start">
          <TechnicalDetails rows={v.technical} open={failed} />
        </div>
      </main>

      <footer className="border-t border-border pt-4 text-small text-ink-muted">{copy.footer}</footer>
    </div>
  );
}
