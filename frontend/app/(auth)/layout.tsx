import Link from "next/link";

import { Brand } from "@/components/chaintrack/brand";
import { ThemeToggle } from "@/components/chaintrack/theme-toggle";
import { copy, routes } from "@/lib/consts";

const t = copy.auth;

// Every access screen: a navy brand panel on the left, the form on the right; on phones only the form, under the logo.
export default function AuthLayout({ children }: LayoutProps<"/">) {
  return (
    <div className="min-h-dvh bg-surface lg:grid lg:grid-cols-[minmax(--spacing(87),0.8fr)_1.4fr]">
      <aside className="relative hidden overflow-hidden bg-navy-900 p-10 text-hero-ink lg:flex lg:flex-col lg:justify-between">
        <Link href={routes.landing} className="self-start rounded-md">
          <Brand onNavy />
        </Link>
        <div className="relative z-10 grid gap-4">
          <span className="text-xs font-bold tracking-[0.1em] text-hero-accent uppercase">
            {t.eyebrow}
          </span>
          <h1 className="max-w-[16ch] text-5xl leading-[1.1] font-semibold tracking-[-0.04em]">
            {t.title}
          </h1>
          <p className="max-w-105 text-base leading-6 text-hero-muted">
            {t.body}
          </p>
        </div>
        <small className="relative z-10 text-xs text-hero-subtle">
          {t.foot}
        </small>
        <span
          className="pointer-events-none absolute -right-52.5 -bottom-25 size-105 rounded-full border border-hero-line shadow-[0_0_0_60px_var(--hero-line),0_0_0_120px_var(--hero-line)] opacity-40"
          aria-hidden
        />
      </aside>
      <div className="relative flex min-h-dvh flex-col px-5 pt-7 pb-12 lg:items-center lg:justify-center lg:p-12">
        <div className="absolute top-5 right-5">
          <ThemeToggle />
        </div>
        <Link
          href={routes.landing}
          className="mb-12 self-start rounded-md lg:hidden"
        >
          <Brand />
        </Link>
        {children}
      </div>
    </div>
  );
}
