import { ArrowRight } from "lucide-react";
import Link from "next/link";
import { Brand } from "@/components/chaintrack/brand";
import { ThemeToggle } from "@/components/chaintrack/theme-toggle";
import { buttonVariants } from "@/components/ui/button";
import { marketing as m } from "@/lib/consts";
import { routes } from "@/lib/consts";
import { cn } from "@/lib/utils";
import { Container, anchor } from "./shared";

export function Nav() {
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
