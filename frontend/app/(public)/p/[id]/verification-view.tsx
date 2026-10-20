"use client";

import { useQuery } from "@tanstack/react-query";
import {
  Check,
  ChevronRight,
  CircleCheck,
  Coffee,
  Copy,
  Factory,
  FileText,
  Gem,
  Info,
  Layers,
  Leaf,
  MapPin,
  OctagonX,
  Package,
  Pickaxe,
  QrCode,
  Route,
  ShieldAlert,
  ShieldCheck,
  ShieldX,
  Sprout,
  Warehouse,
  WifiOff,
  type LucideIcon,
} from "lucide-react";
import Link from "next/link";
import { useState } from "react";

import { MapView } from "@/components/chaintrack/map";
import { Brand } from "@/components/chaintrack/brand";
import { ChecksList } from "@/components/chaintrack/checks-list";
import { JourneyTimeline } from "@/components/chaintrack/journey-timeline";
import { Notice } from "@/components/chaintrack/notice";
import { Loading, PageState } from "@/components/chaintrack/page-state";
import { ScanButton } from "@/components/chaintrack/scan-button";
import { StatusChip } from "@/components/chaintrack/status";
import { TechnicalDetails } from "@/components/chaintrack/technical-details";
import { ThemeToggle } from "@/components/chaintrack/theme-toggle";
import { tone, verdictTone, type Tone } from "@/components/chaintrack/tone";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { endpoints } from "@/lib/api";
import { copy, passportStatus, queryKeys, routes } from "@/lib/consts";
import { cn } from "@/lib/utils";
import {
  toView,
  type CertificateState,
  type Verdict,
  type View,
} from "@/lib/verification";

const t = copy.verification;

// Icons by the first word of a commodity or facility kind; anything else is a generic lot or site.
const commodityIcons: Record<string, LucideIcon> = {
  coffee: Coffee,
  cocoa: Leaf,
  tin: Gem,
  gold: Gem,
  mineral: Gem,
};
const facilityIcons: Record<string, LucideIcon> = {
  Farm: Sprout,
  Mine: Pickaxe,
  Warehouse: Warehouse,
};
const firstWord = (s: string) => s.split(/[\s_]/)[0];

// A passport id, or a lot code (ACME-2026-0412). Anything else is simply not found; the API is not asked.
const reference =
  /^([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}|[A-Za-z0-9][A-Za-z0-9-]{2,79})$/i;

const verdictMark: Record<Verdict, LucideIcon> = {
  verified: ShieldCheck,
  warning: ShieldAlert,
  failed: ShieldX,
};
const onNavy: Record<Verdict, string> = {
  verified: "text-hero-verified",
  warning: "text-hero-warning",
  failed: "text-hero-danger",
};
const passportTone: Record<string, Tone> = {
  [passportStatus.active]: "verified",
  [passportStatus.revoked]: "danger",
};
const certificateLook: Record<
  CertificateState,
  { tone: Tone; icon: LucideIcon }
> = {
  valid: { tone: "verified", icon: CircleCheck },
  expiring: { tone: "pending", icon: Info },
  expired: { tone: "warning", icon: ShieldAlert },
  revoked: { tone: "danger", icon: OctagonX },
};

function Card({
  eyebrow,
  title,
  aside,
  className,
  children,
}: {
  eyebrow?: string;
  title?: string;
  aside?: React.ReactNode;
  className?: string;
  children: React.ReactNode;
}) {
  return (
    <section
      className={cn(
        "min-w-0 overflow-hidden rounded-lg border border-border bg-surface shadow-card",
        className,
      )}
    >
      {title && (
        <header className="flex items-center justify-between gap-3 px-5 pt-5 pb-4 lg:px-6">
          <div className="grid gap-1">
            {eyebrow && <span className="eyebrow">{eyebrow}</span>}
            <h2 className="text-title">{title}</h2>
          </div>
          {aside}
        </header>
      )}
      {children}
    </section>
  );
}

// Public: no sign-in. Opened by scanning the QR label on the goods (Figma PublicVerification).
export function VerificationView({ id }: { id: string }) {
  const valid = reference.test(id);
  const query = useQuery({
    queryKey: queryKeys.verify(id),
    queryFn: () => endpoints.verify(id),
    enabled: valid,
  });

  let body: React.ReactNode;
  if (valid && query.isPending) {
    body = (
      <div className="mx-auto grid w-full max-w-content gap-4 px-4 py-6 lg:px-6">
        <Loading>{t.checking}</Loading>
        <div className="grid gap-4 lg:grid-cols-[1.75fr_1fr]">
          <div className="h-60 rounded-lg bg-surface-sunken" />
          <div className="h-60 rounded-lg bg-surface-sunken" />
        </div>
      </div>
    );
  } else if (query.isError) {
    body = (
      <PageState
        icon={WifiOff}
        tone="danger"
        title={t.errorTitle}
        retry={() => query.refetch()}
      >
        {t.errorBody}
      </PageState>
    );
  } else if (!valid || !query.data) {
    body = (
      <PageState icon={QrCode} title={t.notFoundTitle}>
        {t.notFoundBody}
      </PageState>
    );
  } else {
    body = <Passport v={toView(query.data)} />;
  }

  return (
    <div className="flex min-h-dvh flex-col bg-bg">
      <header className="border-b border-border bg-surface">
        <div className="mx-auto flex h-16 w-full max-w-content items-center justify-between gap-2 px-4 lg:h-[68px] lg:px-6">
          <Link href={routes.landing} className="rounded-md">
            <Brand />
          </Link>
          <div className="flex items-center gap-2">
            <ThemeToggle />
            <ScanButton
              trigger={(open) => (
                <button
                  type="button"
                  onClick={open}
                  className="hidden min-h-tap items-center gap-2 rounded-md border border-border bg-surface px-3.5 text-sm font-semibold hover:border-primary hover:text-primary sm:inline-flex"
                >
                  <QrCode className="size-4.5" aria-hidden />
                  {t.scanAnother}
                </button>
              )}
            />
            <Link
              href={routes.signin}
              className="hidden min-h-tap items-center gap-1 rounded-md px-2 text-sm font-semibold text-ink hover:text-primary md:inline-flex"
            >
              {t.businessSignIn}
              <ChevronRight className="size-4" aria-hidden />
            </Link>
          </div>
        </div>
      </header>
      <main className="flex-1">{body}</main>
      <footer className="border-t border-border bg-surface">
        <div className="mx-auto flex w-full max-w-content flex-wrap items-center gap-x-6 gap-y-2 px-4 py-6 lg:px-6">
          <Brand small />
          <p className="text-sm text-ink-muted">{t.tagline}</p>
          <p className="basis-full text-xs text-ink-muted lg:ml-auto lg:basis-auto">
            {t.footer}
          </p>
        </div>
      </footer>
    </div>
  );
}

function Passport({ v }: { v: View }) {
  const failed = v.verdict === "failed";
  return (
    <>
      <Hero v={v} />
      <div className="mx-auto grid w-full max-w-content gap-4 px-3 pt-4 pb-12 sm:px-4 lg:grid-cols-[minmax(0,1.75fr)_minmax(--spacing(75),0.85fr)] lg:gap-5 lg:px-6 lg:pt-6">
        {failed && (
          <div className="lg:col-span-2">
            <Notice kind="blocking">
              <strong>{t.alert.lead}</strong> {t.alert.rest}
            </Notice>
          </div>
        )}
        <div className="grid min-w-0 content-start gap-4 lg:gap-5">
          <ProductCard v={v} />
          <JourneyCard v={v} />
          <CertificatesCard v={v} />
        </div>
        <aside className="grid min-w-0 content-start gap-4 lg:gap-5">
          <Card
            eyebrow={t.integrity}
            title={t.sections.checks}
            aside={
              <span
                className={cn(
                  "rounded-full px-2.5 py-1 text-xs font-bold",
                  tone[verdictTone[v.verdict]].soft,
                )}
              >
                {t.passedOf(
                  v.checks.filter((c) => c.result === "verified").length,
                  v.checks.length,
                )}
              </span>
            }
          >
            <div className="px-5 pb-5 lg:px-6">
              <ChecksList checks={v.checks} />
            </div>
          </Card>
          {v.origins.map((o) => (
            <OriginCard key={o.id} origin={o} v={v} />
          ))}
          <section className="flex gap-3 rounded-lg bg-navy-900 p-5 text-hero-ink">
            <span className="grid size-9 shrink-0 place-items-center rounded-md bg-hero-line text-hero-accent">
              <Info className="size-5" aria-hidden />
            </span>
            <div className="grid gap-1.5">
              <h2 className="text-sm font-semibold">{t.help.title}</h2>
              <p className="text-sm leading-5 text-hero-subtle">
                {t.help.body}
              </p>
            </div>
          </section>
          <TechnicalDetails rows={v.technical} open={failed} />
        </aside>
      </div>
    </>
  );
}

/** The navy hero with the verdict (Figma .passport-hero): an icon and a word, never colour alone. */
function Hero({ v }: { v: View }) {
  const Mark = verdictMark[v.verdict];
  return (
    <section
      role="status"
      className="relative overflow-hidden bg-navy-900 text-hero-ink"
    >
      <div
        className="hero-grid pointer-events-none absolute inset-0 opacity-60"
        aria-hidden
      />
      <div
        className="pointer-events-none absolute -top-48 right-[9%] size-[440px] rounded-full bg-hero-glow blur-[50px]"
        aria-hidden
      />
      <div className="relative mx-auto grid w-full max-w-content gap-6 px-4 pt-5 pb-8 lg:px-6 lg:pt-6 lg:pb-9">
        <p className="flex min-w-0 items-center gap-1.5 text-xs text-hero-subtle">
          <span>{t.publicLabel}</span>
          <ChevronRight className="size-3.5 shrink-0" aria-hidden />
          <span className="truncate font-mono text-hero-muted">
            {v.product.lotCode}
          </span>
        </p>
        <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between lg:gap-12">
          <div className="flex max-w-162 items-start gap-4">
            <span
              className={cn(
                "grid size-12 shrink-0 place-items-center rounded-full ring-8 ring-hero-line lg:size-14.5",
                tone[verdictTone[v.verdict]].solid,
              )}
            >
              <Mark className="size-7 lg:size-8" aria-hidden />
            </span>
            <div className="grid gap-2">
              <span
                className={cn(
                  "text-xs font-bold tracking-[0.1em] uppercase",
                  onNavy[v.verdict],
                )}
              >
                {t.resultEyebrow}
              </span>
              <h1 className="text-4xl leading-none font-semibold tracking-[-0.03em] lg:text-5xl">
                {t.verdict[v.verdict]}
              </h1>
              <p className="max-w-[60ch] text-sm leading-6 text-hero-muted lg:text-base">
                {v.summary}
              </p>
              {v.reasons.some((r) => r.result !== "pass") && (
                <ul className="mt-1 grid gap-1.5">
                  {v.reasons
                    .filter((r) => r.result !== "pass")
                    .map((r) => (
                      <li
                        key={r.text}
                        className="flex items-start gap-2 text-sm font-semibold"
                      >
                        {r.result === "fail" ? (
                          <OctagonX
                            className="mt-0.5 size-4 shrink-0 text-hero-danger"
                            aria-hidden
                          />
                        ) : (
                          <ShieldAlert
                            className="mt-0.5 size-4 shrink-0 text-hero-warning"
                            aria-hidden
                          />
                        )}
                        {r.text}
                      </li>
                    ))}
                </ul>
              )}
            </div>
          </div>
          <dl className="grid grid-cols-2 gap-6 border-t border-hero-line pt-5 lg:min-w-90 lg:border-0 lg:pt-0">
            <div className="grid gap-1.5">
              <dt className="text-xs tracking-[0.07em] text-hero-subtle uppercase">
                {t.checkedOn}
              </dt>
              <dd className="text-sm font-medium">{v.checkedAt}</dd>
            </div>
            <div className="grid gap-1.5">
              <dt className="text-xs tracking-[0.07em] text-hero-subtle uppercase">
                {t.recordsLabel}
              </dt>
              <dd
                className={cn(
                  "flex items-center gap-1.5 text-sm font-medium",
                  onNavy[v.verdict],
                )}
              >
                <Check className="size-4" aria-hidden />
                {v.records}
              </dd>
            </div>
          </dl>
        </div>
      </div>
    </section>
  );
}

function ProductCard({ v }: { v: View }) {
  const [copied, setCopied] = useState(false);
  const p = v.product;
  const Icon = commodityIcons[firstWord(p.name).toLowerCase()] ?? Package;
  const coffee = firstWord(p.name).toLowerCase() === "coffee";
  const copyCode = () => {
    navigator.clipboard?.writeText(p.lotCode).then(
      () => {
        setCopied(true);
        window.setTimeout(() => setCopied(false), 1500);
      },
      () => undefined,
    );
  };
  const statusTone = passportTone[p.status] ?? "neutral";
  const facts = p.facts.filter(
    (f) => f.label !== t.facts.commodity && f.label !== t.facts.status,
  );
  return (
    <Card className="grid sm:grid-cols-[--spacing(53)_1fr]">
      {/* Product art: decoration only, allowed on this page. */}
      <div
        className="product-art relative min-h-40 overflow-hidden sm:min-h-[238px]"
        aria-hidden
      >
        {coffee ? (
          <>
            <span className="bean absolute top-8 left-[32%] h-[74px] w-14 rotate-[36deg] sm:top-[54px] sm:left-[53px]" />
            <span className="bean absolute top-14 left-[53%] h-[74px] w-14 scale-[0.8] -rotate-[28deg] sm:top-[91px] sm:left-[119px]" />
            <span className="bean absolute top-[100px] left-[33%] h-[74px] w-14 scale-[0.56] rotate-[79deg] sm:top-[143px] sm:left-[47px]" />
          </>
        ) : (
          <Icon className="absolute inset-0 m-auto size-20 text-hero-ink/30" />
        )}
        {v.origins[0] && (
          <span className="absolute bottom-4 left-4 rounded-sm border border-hero-line bg-navy-950/60 px-2 py-1.5 text-xs font-semibold tracking-[0.1em] text-hero-ink uppercase backdrop-blur-sm">
            {v.origins[0].kind}
          </span>
        )}
      </div>
      <div className="grid content-start gap-4 p-5 lg:p-7">
        <div className="flex flex-wrap gap-2">
          <span className="inline-flex min-h-6.5 items-center rounded-full bg-surface-sunken px-2.5 text-xs font-bold text-ink-muted">
            {p.name}
          </span>
          <StatusChip
            tone={statusTone}
            icon={statusTone === "danger" ? OctagonX : CircleCheck}
          >
            {t.passportStatus[p.status] ?? p.status}
          </StatusChip>
        </div>
        <div>
          <h2 className="text-2xl leading-8 font-semibold tracking-tight">
            {p.name}
          </h2>
          <p className="text-sm text-ink-muted">{p.issued}</p>
        </div>
        <dl className="grid grid-cols-2 gap-x-6 gap-y-4 border-t border-border pt-4">
          <div className="col-span-2 grid gap-1 sm:col-span-1">
            <dt className="text-xs font-semibold tracking-[0.07em] text-ink-muted uppercase">
              {t.lotCode}
            </dt>
            <dd>
              <button
                type="button"
                onClick={copyCode}
                aria-label={`${t.copyCode}: ${p.lotCode}`}
                className="flex min-h-tap items-center gap-2 rounded-md text-left text-sm font-semibold hover:text-primary"
              >
                <code className="code break-all">{p.lotCode}</code>
                {copied ? (
                  <Check
                    className="size-4 shrink-0 text-verified"
                    aria-hidden
                  />
                ) : (
                  <Copy
                    className="size-4 shrink-0 text-ink-muted"
                    aria-hidden
                  />
                )}
              </button>
              <span role="status" className="sr-only">
                {copied ? t.copied : ""}
              </span>
            </dd>
          </div>
          {facts.map((f) => (
            <div
              key={f.label}
              className={cn(
                "grid content-start gap-1",
                f.wide && "col-span-2 sm:col-span-1",
              )}
            >
              <dt className="text-xs font-semibold tracking-[0.07em] text-ink-muted uppercase">
                {f.label}
              </dt>
              <dd className={cn("text-sm font-semibold", f.mono && "fig")}>
                {f.value}
              </dd>
            </div>
          ))}
        </dl>
      </div>
    </Card>
  );
}

function JourneyCard({ v }: { v: View }) {
  return (
    <Card>
      <Tabs defaultValue="journey" className="gap-0">
        <TabsList className="px-4 lg:px-6">
          <TabsTrigger value="journey">
            <Route aria-hidden />
            {t.tabs.journey}
          </TabsTrigger>
          <TabsTrigger value="genealogy">
            <Layers aria-hidden />
            {t.tabs.genealogy}
          </TabsTrigger>
        </TabsList>
        <TabsContent
          value="journey"
          className="grid md:grid-cols-[1.05fr_0.95fr]"
        >
          <div className="p-5 lg:p-7">
            <JourneyTimeline steps={v.journey} />
          </div>
          <div className="grid gap-4">
            <RouteArt v={v} />
            <ol className="grid gap-2 p-5 pt-0 lg:px-7" aria-label="Facility points">
              {v.origins.map((o, i) => (
                <li key={o.id} className="flex items-center gap-2 text-sm">
                  <span className="font-mono text-ink-muted">{String(i + 1).padStart(2, "0")}</span>
                  <span className="font-medium">{o.name}</span>
                  <span className="text-ink-muted">· {o.kind}</span>
                  <span className="text-ink-muted">· {o.owner}</span>
                </li>
              ))}
            </ol>
          </div>
        </TabsContent>
        <TabsContent value="genealogy" className="p-5 lg:p-7">
          <Genealogy v={v} />
        </TabsContent>
      </Tabs>
    </Card>
  );
}

/**
 * Facility points only (M3): origin to current holder as named points.
 * No leg is invented: old handovers without GPS still show their facilities.
 */
function RouteArt({ v }: { v: View }) {
  const origin = v.origins[0]?.name;
  return (
    <figure className="relative min-h-75 overflow-hidden border-t border-border bg-map-water md:min-h-107.5 md:border-t-0 md:border-l">
      <svg
        viewBox="0 0 420 480"
        className="absolute inset-0 size-full"
        preserveAspectRatio="xMidYMid slice"
        role="img"
        aria-label={t.route.label}
      >
        <path
          className="fill-map-land stroke-map-contour"
          d="M19 26 148 12l41 50 87 13 20 58 70 54-8 72 43 72-77 129-152-9-62-72-53-14 9-92-47-64 43-75z"
        />
        <path
          className="fill-map-land stroke-map-contour"
          d="m300 16 99 3-5 117-75-26z"
        />
        <path
          className="fill-none stroke-map-road"
          d="m128 90 37 54-16 76 75 36 26 64 75 28M54 210l95 10M224 256l79-65"
        />
        <path
          className="fill-none stroke-primary"
          strokeWidth="3"
          strokeDasharray="5 5"
          d="M130 278 C157 261 165 251 185 238 S245 205 297 173"
        />
        {[
          [130, 278, 10],
          [185, 238, 9],
          [297, 173, 10],
        ].map(([cx, cy, r]) => (
          <g key={cx}>
            <circle
              cx={cx}
              cy={cy}
              r={r}
              className="fill-surface stroke-primary"
              strokeWidth="2"
            />
            <circle cx={cx} cy={cy} r={r / 2} className="fill-primary" />
          </g>
        ))}
      </svg>
      {origin && (
        <RouteLabel
          n={1}
          small={t.route.origin}
          name={origin}
          className="top-[57%] left-[8%]"
        />
      )}
      {v.product.holder && (
        <RouteLabel
          n={v.journey.length}
          small={t.route.now}
          name={v.product.holder}
          className="top-[28%] right-[4%]"
        />
      )}
    </figure>
  );
}

function RouteLabel({
  n,
  small,
  name,
  className,
}: {
  n: number;
  small: string;
  name: string;
  className: string;
}) {
  return (
    <figcaption
      className={cn(
        "absolute flex max-w-[60%] items-center gap-2 rounded-md border border-border bg-surface px-2.5 py-2 shadow-overlay",
        className,
      )}
    >
      <span className="grid size-6 shrink-0 place-items-center rounded-full bg-primary text-xs font-bold text-on-primary">
        {n}
      </span>
      <span className="grid min-w-0">
        <span className="text-xs font-bold text-primary uppercase">
          {small}
        </span>
        <span className="truncate text-xs font-semibold">{name}</span>
      </span>
    </figcaption>
  );
}

/** The lot and every lot it was made from, back to origin (Figma .tree). */
function Genealogy({ v }: { v: View }) {
  const p = v.product;
  const node = (
    key: string,
    Icon: LucideIcon,
    small: string,
    code: string,
    line: string,
    look: string,
    end?: React.ReactNode,
  ) => (
    <div
      key={key}
      className="flex items-center gap-3 rounded-lg border border-border bg-surface p-3.5 shadow-card"
    >
      <span
        className={cn(
          "grid size-9 shrink-0 place-items-center rounded-md border",
          look,
        )}
      >
        <Icon className="size-5" aria-hidden />
      </span>
      <div className="grid min-w-0 flex-1 gap-0.5">
        <span className="text-xs font-semibold tracking-[0.08em] text-ink-muted uppercase">
          {small}
        </span>
        <code className="code text-sm break-all">{code}</code>
        <span className="text-xs text-ink-muted">{line}</span>
      </div>
      {end}
    </div>
  );
  return (
    <div className="grid gap-8">
      <div className="flex items-center gap-3">
        <span className="grid size-10 shrink-0 place-items-center rounded-md border border-primary/25 bg-primary-soft text-primary">
          <Layers className="size-5" aria-hidden />
        </span>
        <div>
          <h3 className="text-title">{t.genealogyTitle}</h3>
          <p className="text-sm text-ink-muted">
            {t.genealogyBody(v.madeFrom.length)}
          </p>
        </div>
      </div>
      <div className="mx-auto grid w-full max-w-130">
        {node(
          "current",
          Package,
          t.currentLot,
          p.lotCode,
          p.facts.find((f) => f.label === t.facts.quantity)?.value ?? "",
          "border-primary/25 bg-primary-soft text-primary",
        )}
        {v.madeFrom.map((lot) => (
          <div key={lot.lotCode} className="grid">
            <div
              className="relative grid h-14 place-items-center"
              aria-hidden={!lot.via}
            >
              <span className="absolute inset-y-0 w-px bg-border" />
              {lot.via && (
                <span className="relative rounded-full bg-bg px-2.5 py-1 text-xs text-ink-muted">
                  {lot.via}
                </span>
              )}
            </div>
            {node(
              lot.lotCode,
              lot.origin ? MapPin : Package,
              lot.origin ? t.originLot : t.inputLot,
              lot.lotCode,
              `${lot.description} · ${lot.quantity}`,
              lot.revoked
                ? "border-danger/30 bg-danger-soft text-danger"
                : lot.origin
                  ? "border-warning/30 bg-warning-soft text-warning-ink"
                  : "border-border bg-surface-sunken text-ink-muted",
              lot.revoked && (
                <StatusChip tone="danger" icon={OctagonX}>
                  {t.revoked}
                </StatusChip>
              ),
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

function CertificatesCard({ v }: { v: View }) {
  return (
    <Card
      eyebrow={t.documentation}
      title={t.sections.certificates}
      aside={
        <span className="rounded-full bg-surface-sunken px-2.5 py-1 text-xs text-ink-muted">
          {t.meta.records(v.certificates.length)}
        </span>
      }
    >
      {v.certificates.length === 0 ? (
        <p className="px-5 pb-5 text-sm text-ink-muted lg:px-6">
          {t.noCertificates}
        </p>
      ) : (
        <ul className="px-2.5 pb-2.5">
          {v.certificates.map((c) => {
            const look = certificateLook[c.state];
            return (
              <li
                key={c.id}
                className="grid grid-cols-[auto_1fr] items-center gap-3 border-t border-border px-3 py-3.5 sm:grid-cols-[auto_1fr_auto]"
              >
                <span className="grid h-10 w-9 place-items-center rounded-md bg-primary-soft text-primary">
                  <FileText className="size-5" aria-hidden />
                </span>
                <span className="grid min-w-0 gap-1">
                  <span className="text-sm font-semibold">{c.type}</span>
                  <span className="text-xs text-ink-muted">
                    {c.issuer} · {c.validFrom} {t.validTo} {c.validTo}
                  </span>
                </span>
                <span className="col-start-2 sm:col-start-3">
                  <StatusChip tone={look.tone} icon={look.icon}>
                    {t.certificate[c.state]}
                  </StatusChip>
                </span>
                {c.note && (
                  <p
                    className={cn(
                      "col-span-full rounded-sm px-3 py-2 text-xs",
                      tone[look.tone].soft,
                    )}
                  >
                    {c.note}
                  </p>
                )}
              </li>
            );
          })}
        </ul>
      )}
    </Card>
  );
}

function OriginCard({
  origin: o,
  v,
}: {
  origin: View["origins"][number];
  v: View;
}) {
  const Icon = facilityIcons[firstWord(o.kind)] ?? Factory;
  const capacity = v.checks.find((c) => c.name === "capacity");
  return (
    <Card className="grid gap-4 p-5 lg:p-6">
      <span className="eyebrow">{t.originEyebrow}</span>
      <div className="flex gap-3">
        <span className="grid size-10 shrink-0 place-items-center rounded-md bg-primary-soft text-primary">
          <Icon className="size-5" aria-hidden />
        </span>
        <div className="min-w-0">
          <h2 className="text-base leading-5 font-semibold">{o.name}</h2>
          <p className="text-xs text-ink-muted">{o.kind}</p>
        </div>
      </div>
      {o.geometry && (
        <MapView
          label={t.originMap(o.name)}
          shapes={[{ id: o.id, geometry: o.geometry }]}
          className="h-56 overflow-hidden rounded-md border border-border"
        />
      )}
      <dl className="border-t border-border text-sm">
        <div className="grid grid-cols-[--spacing(27)_1fr] items-center gap-2 border-b border-border py-3">
          <dt className="text-ink-muted">{t.operator}</dt>
          <dd className="font-medium">{o.owner}</dd>
        </div>
        {capacity && (
          <div className="grid grid-cols-[--spacing(27)_1fr] items-center gap-2 py-3">
            <dt className="text-ink-muted">{t.capacityCheck}</dt>
            <dd
              className={cn(
                "flex items-center gap-1.5 font-semibold",
                tone[verdictTone[capacity.result]].fg,
              )}
            >
              {capacity.result === "verified" ? (
                <CircleCheck className="size-4" aria-hidden />
              ) : (
                <ShieldAlert className="size-4" aria-hidden />
              )}
              {t.result[capacity.result]}
            </dd>
          </div>
        )}
      </dl>
    </Card>
  );
}
