"use client";

import {
  ArrowRight,
  Check,
  CircleCheck,
  Clock,
  FileText,
  Info,
  Layers,
  Link2,
  MapPin,
  OctagonX,
  Package,
  QrCode,
  Route,
  ShieldCheck,
  TriangleAlert,
  type LucideIcon,
} from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Brand } from "@/components/chaintrack/brand";
import { Segmented } from "@/components/chaintrack/segmented";
import { StatusChip } from "@/components/chaintrack/status";
import { ThemeToggle } from "@/components/chaintrack/theme-toggle";
import { Button, buttonVariants } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  marketing as m,
  marketingRoles,
  routes,
  type MarketingRole,
} from "@/lib/consts";
import { useMeGuard } from "@/lib/api";
import { cn } from "@/lib/utils";

const icons: Record<string, LucideIcon> = {
  map: MapPin,
  route: Route,
  layers: Layers,
  shield: ShieldCheck,
  link: Link2,
  file: FileText,
  clock: Clock,
  check: Check,
};
const verdictIcon = {
  verified: CircleCheck,
  warning: TriangleAlert,
  danger: OctagonX,
} as const;
const roleIcon: Record<MarketingRole, LucideIcon> = {
  producer: MapPin,
  logistics: Route,
  certifier: FileText,
  regulator: ShieldCheck,
};
const heroPhoto =
  "https://images.unsplash.com/photo-1756245994771-f06fcaee3a36?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&q=82&w=1080";
const anchor = (id: string) => `#${id}`;

/** The public home: what ChainTrack is, how it works per role, and a way to verify a label. */
export function Landing() {
  const me = useMeGuard((who) => (who ? routes.home : undefined)); // signed in: straight to the dashboard
  if (me.data) return null;
  return (
    <div className="min-h-dvh bg-bg text-ink">
      <Nav />
      <main>
        <Hero />
        <VerifyStrip />
        <HowItWorks />
        <RoleGuide />
        <Showcase />
        <TrustVerdicts />
        <Security />
        <Faq />
        <FinalCta />
      </main>
      <Footer />
    </div>
  );
}

function Container({
  className,
  children,
}: {
  className?: string;
  children: React.ReactNode;
}) {
  return (
    <div className={cn("mx-auto w-full max-w-7xl px-4 lg:px-8", className)}>
      {children}
    </div>
  );
}

function SectionHeading({
  eyebrow,
  title,
  body,
  center,
}: {
  eyebrow: string;
  title: React.ReactNode;
  body?: string;
  center?: boolean;
}) {
  return (
    <div
      className={cn("grid max-w-2xl gap-3", center && "mx-auto text-center")}
    >
      <span className="eyebrow text-primary">{eyebrow}</span>
      <h2 className="text-3xl font-semibold tracking-tight lg:text-4xl">
        {title}
      </h2>
      {body && <p className="text-ink-muted">{body}</p>}
    </div>
  );
}

function Nav() {
  return (
    <header className="sticky top-0 z-30 border-b border-border bg-surface/90 backdrop-blur">
      <Container className="flex h-16 items-center justify-between gap-4">
        <Link href={routes.landing} className="rounded-md">
          <Brand />
        </Link>
        <nav className="hidden items-center gap-6 text-sm font-medium text-ink-muted md:flex">
          <a href={anchor(m.anchors.how)} className="hover:text-ink">
            {m.nav.how}
          </a>
          <a href={anchor(m.anchors.solutions)} className="hover:text-ink">
            {m.nav.solutions}
          </a>
          <a href={anchor(m.anchors.trust)} className="hover:text-ink">
            {m.nav.trust}
          </a>
        </nav>
        <div className="flex items-center gap-2">
          <ThemeToggle />
          <Link
            href={routes.signin}
            className={cn(
              buttonVariants({ variant: "ghost" }),
              "hidden sm:inline-flex",
            )}
          >
            {m.nav.signIn}
          </Link>
          <Link href={routes.signup} className={buttonVariants()}>
            {m.nav.getStarted}
            <ArrowRight aria-hidden />
          </Link>
        </div>
      </Container>
    </header>
  );
}

function Hero() {
  const h = m.hero;
  return (
    <section className="relative overflow-hidden bg-navy-900 text-hero-ink">
      <div
        className="hero-grid pointer-events-none absolute inset-0 opacity-60"
        aria-hidden
      />
      <Container className="relative grid gap-12 py-16 lg:grid-cols-2 lg:items-center lg:py-24">
        <div className="grid content-start gap-6">
          <span className="inline-flex w-fit items-center gap-2 rounded-full border border-hero-line px-3 py-1 text-xs font-semibold text-hero-accent">
            <ShieldCheck className="size-3.5" aria-hidden />
            {h.kicker}
          </span>
          <h1 className="text-4xl leading-tight font-semibold tracking-tight lg:text-6xl">
            {h.title}
            <br />
            <em className="text-hero-accent not-italic">{h.titleEm}</em>
          </h1>
          <p className="max-w-xl text-lg text-hero-muted">{h.body}</p>
          <div className="flex flex-wrap gap-3">
            <Link href={routes.signup} className={buttonVariants()}>
              {h.primary}
              <ArrowRight aria-hidden />
            </Link>
            <a
              href={anchor(m.anchors.verify)}
              className={cn(
                buttonVariants({ variant: "secondary" }),
                "border-hero-line bg-transparent text-hero-ink hover:text-hero-accent",
              )}
            >
              <QrCode aria-hidden />
              {h.secondary}
            </a>
          </div>
          <ul className="flex flex-wrap gap-x-5 gap-y-2 text-sm text-hero-muted">
            {h.proof.map((p) => (
              <li key={p} className="flex items-center gap-1.5">
                <Check className="size-4 text-hero-accent" aria-hidden />
                {p}
              </li>
            ))}
          </ul>
        </div>
        <HeroVisual />
      </Container>
      <Container className="relative flex flex-wrap items-center gap-x-8 gap-y-3 border-t border-hero-line py-6 text-xs">
        <span className="text-hero-subtle">{m.trustStrip.label}</span>
        <ul className="flex flex-wrap gap-x-8 gap-y-2 font-bold tracking-widest text-hero-muted uppercase">
          {m.trustStrip.items.map((x) => (
            <li key={x}>{x}</li>
          ))}
        </ul>
      </Container>
    </section>
  );
}

/** An example passport over a photo: decoration, labelled as an example. */
function HeroVisual() {
  const h = m.hero;
  return (
    <div className="relative mx-auto w-full max-w-lg pb-20 lg:pb-0">
      <div className="relative aspect-4/5 overflow-hidden rounded-2xl border border-hero-line">
        <Image
          src={heroPhoto}
          alt={h.photoAlt}
          fill
          sizes="(min-width: 1024px) 32rem, 90vw"
          className="object-cover"
          priority
        />
        <div className="absolute inset-x-4 bottom-4 flex items-center gap-3 rounded-xl bg-surface p-3 text-ink shadow-overlay">
          <span className="grid size-9 place-items-center rounded-md bg-primary-soft text-primary">
            <MapPin className="size-4" aria-hidden />
          </span>
          <div className="grid min-w-0 flex-1">
            <small className="eyebrow">
              {h.example} · {h.originLabel}
            </small>
            <strong className="truncate text-sm">{h.originPlace}</strong>
          </div>
          <StatusChip tone="verified" icon={CircleCheck}>
            {h.active}
          </StatusChip>
        </div>
      </div>
      <div className="absolute -bottom-2 -left-2 w-72 rounded-xl border border-border bg-surface p-4 text-ink shadow-overlay sm:-left-10 lg:top-10 lg:bottom-auto">
        <div className="flex items-center gap-3">
          <span className="grid size-9 place-items-center rounded-md bg-primary-soft text-primary">
            <Package className="size-4" aria-hidden />
          </span>
          <div className="grid min-w-0">
            <small className="eyebrow">
              {h.example} · {h.passportLabel}
            </small>
            <strong className="truncate text-sm">{h.passportName}</strong>
            <code className="code text-xs">{h.passportCode}</code>
          </div>
        </div>
        <ol className="mt-4 grid grid-cols-4 gap-1 text-center text-xs text-ink-muted">
          {h.stops.map((s, i) => (
            <li key={s} className="grid justify-items-center gap-1">
              <span
                className={cn(
                  "grid size-6 place-items-center rounded-full border-2",
                  i < 2
                    ? "border-verified bg-verified text-on-verified"
                    : i === 2
                      ? "border-primary text-primary"
                      : "border-border",
                )}
              >
                {i < 2 ? (
                  <Check className="size-3" aria-hidden />
                ) : i === 2 ? (
                  <Route className="size-3" aria-hidden />
                ) : null}
              </span>
              {s}
            </li>
          ))}
        </ol>
        <div className="mt-4 flex items-center justify-between gap-2 border-t border-border pt-3 text-xs text-ink-muted">
          <StatusChip tone="verified" icon={ShieldCheck}>
            {h.verified}
          </StatusChip>
          {h.checks}
        </div>
      </div>
      <div className="absolute right-0 bottom-6 hidden items-center gap-3 rounded-xl bg-surface p-3 text-ink shadow-overlay sm:flex lg:-right-6">
        <span className="grid size-8 place-items-center rounded-full bg-verified text-on-verified">
          <Check className="size-4" aria-hidden />
        </span>
        <div className="grid">
          <strong className="text-sm">{h.signedTitle}</strong>
          <small className="text-xs text-ink-muted">{h.signedSub}</small>
        </div>
      </div>
    </div>
  );
}

/** Type a lot code and open its public verification. */
function VerifyStrip() {
  const router = useRouter();
  const [code, setCode] = useState("");
  const v = m.verify;
  return (
    <section
      id={m.anchors.verify}
      className="border-b border-border bg-surface"
    >
      <Container className="grid gap-6 py-10 lg:grid-cols-2 lg:items-center">
        <div className="flex items-start gap-4">
          <span className="grid size-12 shrink-0 place-items-center rounded-lg bg-primary-soft text-primary">
            <QrCode className="size-6" aria-hidden />
          </span>
          <div className="grid gap-1">
            <h2 className="text-xl font-semibold">{v.title}</h2>
            <p className="text-sm text-ink-muted">{v.body}</p>
          </div>
        </div>
        <form
          className="flex gap-2"
          onSubmit={(e) => {
            e.preventDefault();
            if (code.trim())
              router.push(routes.passport(encodeURIComponent(code.trim())));
          }}
        >
          <label className="relative flex-1">
            <span className="sr-only">{v.label}</span>
            <Package
              className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-ink-subtle"
              aria-hidden
            />
            <Input
              value={code}
              onChange={(e) => setCode(e.target.value)}
              placeholder={v.placeholder}
              className="pl-9 font-mono"
              autoCapitalize="characters"
            />
          </label>
          <Button type="submit">
            {v.submit}
            <ArrowRight aria-hidden />
          </Button>
        </form>
      </Container>
    </section>
  );
}

function HowItWorks() {
  const h = m.how;
  return (
    <section id={m.anchors.how} className="py-20">
      <Container className="grid gap-12">
        <SectionHeading
          eyebrow={h.eyebrow}
          title={h.title}
          body={h.body}
          center
        />
        <ol className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {h.steps.map((s, i) => {
            const Icon = icons[s.icon];
            return (
              <li
                key={s.title}
                className="grid content-start gap-3 rounded-xl border border-border bg-surface p-6 shadow-card"
              >
                <span className="font-mono text-xs text-ink-subtle">
                  {String(i + 1).padStart(2, "0")}
                </span>
                <span className="grid size-10 place-items-center rounded-lg bg-primary-soft text-primary">
                  <Icon className="size-5" aria-hidden />
                </span>
                <h3 className="text-lg font-semibold">{s.title}</h3>
                <p className="text-sm text-ink-muted">{s.body}</p>
              </li>
            );
          })}
        </ol>
      </Container>
    </section>
  );
}

function RoleGuide() {
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

function Showcase() {
  const s = m.showcase;
  return (
    <section id={m.anchors.solutions} className="py-20">
      <Container className="grid gap-12 lg:grid-cols-2 lg:items-center">
        <div className="grid content-start gap-6">
          <SectionHeading
            eyebrow={s.eyebrow}
            title={
              <>
                {s.title}
                <br />
                {s.titleLine}
              </>
            }
            body={s.body}
          />
          <ul className="grid gap-4">
            {s.points.map(([title, body]) => (
              <li key={title} className="flex gap-3">
                <span className="mt-0.5 grid size-6 shrink-0 place-items-center rounded-full bg-verified-soft text-verified">
                  <Check className="size-3.5" aria-hidden />
                </span>
                <div className="grid gap-0.5">
                  <strong className="text-sm">{title}</strong>
                  <p className="text-sm text-ink-muted">{body}</p>
                </div>
              </li>
            ))}
          </ul>
          <Link
            href={routes.signin}
            className={cn(buttonVariants({ variant: "ghost" }), "w-fit")}
          >
            {s.cta}
            <ArrowRight aria-hidden />
          </Link>
        </div>
        <div className="overflow-hidden rounded-2xl border border-border bg-surface shadow-overlay">
          <div className="flex items-center gap-1.5 border-b border-border bg-surface-sunken px-4 py-3 text-xs text-ink-muted">
            {[0, 1, 2].map((i) => (
              <span
                key={i}
                className="size-2.5 rounded-full bg-border-strong"
                aria-hidden
              />
            ))}
            <b className="ml-2 font-semibold">{s.window}</b>
          </div>
          <div className="flex items-center gap-4 bg-navy-900 p-6 text-hero-ink">
            <span className="grid size-12 place-items-center rounded-full bg-hero-verified text-navy-900">
              <ShieldCheck className="size-6" aria-hidden />
            </span>
            <div className="grid flex-1">
              <small className="eyebrow text-hero-subtle">
                {m.hero.example} · {s.result}
              </small>
              <h3 className="text-2xl font-semibold">{s.verdict}</h3>
              <p className="text-sm text-hero-muted">{s.verdictSub}</p>
            </div>
            <code className="hidden font-mono text-xs text-hero-muted sm:block">
              {m.hero.passportCode}
            </code>
          </div>
          <ol className="grid gap-4 p-6">
            {s.timeline.map((x) => (
              <li key={x} className="flex items-center gap-3">
                <span className="grid size-6 place-items-center rounded-full bg-verified text-on-verified">
                  <Check className="size-3" aria-hidden />
                </span>
                <div className="grid">
                  <strong className="text-sm">{x}</strong>
                  <small className="text-xs text-ink-muted">{s.signed}</small>
                </div>
              </li>
            ))}
          </ol>
        </div>
      </Container>
    </section>
  );
}

function TrustVerdicts() {
  const t = m.trust;
  return (
    <section id={m.anchors.trust} className="pb-20">
      <Container>
        <div className="grid gap-8 rounded-2xl bg-navy-900 p-8 text-hero-ink lg:grid-cols-2 lg:p-12">
          <div className="grid content-start gap-4">
            <span className="eyebrow text-hero-accent">{t.eyebrow}</span>
            <h2 className="text-3xl font-semibold tracking-tight">{t.title}</h2>
            <p className="text-hero-muted">{t.body}</p>
            <a
              href={anchor(m.anchors.verify)}
              className={cn(buttonVariants(), "w-fit")}
            >
              {t.cta}
              <ArrowRight aria-hidden />
            </a>
          </div>
          <ul className="grid gap-3">
            {t.verdicts.map((v) => (
              <li
                key={v.label}
                className="grid gap-1 rounded-xl bg-surface p-4 text-ink"
              >
                <StatusChip tone={v.tone} icon={verdictIcon[v.tone]}>
                  {v.label}
                </StatusChip>
                <strong className="text-sm">{v.body}</strong>
                <span className="text-xs text-ink-muted">{t.note}</span>
              </li>
            ))}
          </ul>
        </div>
      </Container>
    </section>
  );
}

function Security() {
  const s = m.security;
  return (
    <section
      id={m.anchors.security}
      className="border-y border-border bg-surface py-20"
    >
      <Container className="grid gap-12 lg:grid-cols-5">
        <div className="grid content-start gap-6 lg:col-span-2">
          <SectionHeading
            eyebrow={s.eyebrow}
            title={
              <>
                {s.title}
                <br />
                {s.titleLine}
              </>
            }
            body={s.body}
          />
          <div className="flex gap-3 rounded-xl border border-border bg-surface-sunken p-4">
            <ShieldCheck
              className="mt-0.5 size-5 shrink-0 text-primary"
              aria-hidden
            />
            <div className="grid gap-1">
              <strong className="text-sm">{s.assuranceTitle}</strong>
              <p className="text-sm text-ink-muted">{s.assuranceBody}</p>
            </div>
          </div>
        </div>
        <ul className="grid gap-4 sm:grid-cols-2 lg:col-span-3">
          {s.cards.map((c) => {
            const Icon = icons[c.icon];
            return (
              <li
                key={c.title}
                className="flex gap-3 rounded-xl border border-border p-5"
              >
                <span className="grid size-10 shrink-0 place-items-center rounded-lg bg-primary-soft text-primary">
                  <Icon className="size-5" aria-hidden />
                </span>
                <div className="grid gap-1">
                  <h3 className="text-sm font-semibold">{c.title}</h3>
                  <p className="text-sm text-ink-muted">{c.body}</p>
                </div>
              </li>
            );
          })}
        </ul>
      </Container>
    </section>
  );
}

function Faq() {
  const f = m.faq;
  return (
    <section className="py-20">
      <Container className="grid gap-10">
        <SectionHeading eyebrow={f.eyebrow} title={f.title} body={f.body} />
        <div className="grid gap-3 md:grid-cols-2">
          {f.items.map(([q, a], i) => (
            <details
              key={q}
              open={i === 0}
              className="group rounded-xl border border-border bg-surface p-5 shadow-card"
            >
              <summary className="flex cursor-pointer list-none items-center justify-between gap-4 font-semibold">
                {q}
                <span
                  className="text-xl text-primary transition-transform group-open:rotate-45"
                  aria-hidden
                >
                  +
                </span>
              </summary>
              <p className="mt-3 text-sm text-ink-muted">{a}</p>
            </details>
          ))}
        </div>
      </Container>
    </section>
  );
}

function FinalCta() {
  const c = m.cta;
  return (
    <section className="relative overflow-hidden bg-navy-900 py-20 text-center text-hero-ink">
      <div
        className="hero-grid pointer-events-none absolute inset-0 opacity-40"
        aria-hidden
      />
      <Container className="relative grid justify-items-center gap-5">
        <span className="grid size-14 place-items-center rounded-2xl bg-primary text-on-primary">
          <Link2 className="size-7" aria-hidden />
        </span>
        <h2 className="text-3xl font-semibold tracking-tight lg:text-4xl">
          {c.title}
        </h2>
        <p className="text-hero-muted">{c.body}</p>
        <div className="flex flex-wrap justify-center gap-3">
          <Link href={routes.signup} className={buttonVariants()}>
            {c.primary}
            <ArrowRight aria-hidden />
          </Link>
          <a
            href={anchor(m.anchors.verify)}
            className={cn(
              buttonVariants({ variant: "secondary" }),
              "border-hero-line bg-transparent text-hero-ink hover:text-hero-accent",
            )}
          >
            {c.secondary}
          </a>
        </div>
      </Container>
    </section>
  );
}

function Footer() {
  const f = m.footer;
  const col = (title: string, links: [string, string][]) => (
    <div className="grid content-start gap-2 text-sm">
      <strong>{title}</strong>
      {links.map(([label, href]) => (
        <a key={label} href={href} className="text-ink-muted hover:text-ink">
          {label}
        </a>
      ))}
    </div>
  );
  return (
    <footer className="border-t border-border bg-surface">
      <Container className="grid gap-8 py-12 sm:grid-cols-3">
        <div className="grid content-start gap-3">
          <Brand />
          <p className="text-sm text-ink-muted">{f.tagline}</p>
        </div>
        {col(f.product, [
          [f.verify, anchor(m.anchors.verify)],
          [f.workspace, routes.signin],
          [f.how, anchor(m.anchors.how)],
        ])}
        {col(f.trustTitle, [
          [f.security, anchor(m.anchors.security)],
          [f.guide, anchor(m.anchors.trust)],
        ])}
      </Container>
      <Container className="border-t border-border py-6 text-xs text-ink-muted">
        {f.rights(new Date().getFullYear())}
      </Container>
    </footer>
  );
}
