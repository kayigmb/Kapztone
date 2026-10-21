import { ArrowRight, Check, QrCode, ShieldCheck } from "lucide-react";
import Link from "next/link";
import { buttonVariants } from "@/components/ui/button";
import { marketing as m } from "@/lib/consts";
import { routes } from "@/lib/consts";
import { cn } from "@/lib/utils";
import { Container, anchor } from "./shared";
import { HeroVisual } from "./hero-visual";

export function Hero() {
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
