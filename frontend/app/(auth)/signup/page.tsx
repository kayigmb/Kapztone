import Link from "next/link";

import { copy, routes } from "@/lib/consts";
import { SignUpForm } from "./signup-form";

const t = copy.signup;

export const metadata = { title: copy.meta.signup };

// Phone-first, no app shell (screens.md §03, steps 1 and 2).
export default function Page() {
  return (
    <main className="mx-auto grid min-h-dvh w-full max-w-xl content-center gap-6 px-4 py-8">
      <header className="grid gap-1">
        <p className="text-title">{copy.brand}</p>
        <h1 className="text-page">{t.title}</h1>
        <p className="text-ink-muted">{t.sub}</p>
      </header>
      <section className="rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
        <SignUpForm />
      </section>
      <p className="text-sm text-ink-muted">
        {t.haveAccount} <Link href={routes.signin} className="font-semibold text-primary underline">{t.signIn}</Link>
      </p>
    </main>
  );
}
