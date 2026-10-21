import { ArrowRight, Link2 } from "lucide-react";
import Link from "next/link";
import { buttonVariants } from "@/components/ui/button";
import { marketing as m } from "@/lib/consts";
import { routes } from "@/lib/consts";
import { cn } from "@/lib/utils";
import { Container, anchor } from "./shared";

export function FinalCta() {
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
