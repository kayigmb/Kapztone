import { ArrowRight, Check, ShieldCheck } from "lucide-react";
import Link from "next/link";
import { buttonVariants } from "@/components/ui/button";
import { marketing as m } from "@/lib/consts";
import { routes } from "@/lib/consts";
import { cn } from "@/lib/utils";
import { Container, SectionHeading } from "./shared";

export function Showcase() {
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
