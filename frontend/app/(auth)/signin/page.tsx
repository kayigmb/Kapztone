import Link from "next/link";

import { copy, routes } from "@/lib/consts";
import { SignInForm } from "./signin-form";

const t = copy.signin;

export const metadata = { title: copy.meta.signin };

// Phone-first, no app shell (screens.md §03).
export default async function Page({ searchParams }: PageProps<"/signin">) {
  const { email } = await searchParams;
  return (
    <main className="mx-auto grid min-h-dvh w-full max-w-md content-center gap-6 px-4 py-8">
      <header className="grid gap-1">
        <p className="text-title">{copy.brand}</p>
        <h1 className="text-page">{t.title}</h1>
        <p className="text-ink-muted">{t.sub}</p>
      </header>
      <section className="rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
        <SignInForm email={typeof email === "string" ? email : ""} />
      </section>
      <p className="text-sm text-ink-muted">
        {t.newHere} <Link href={routes.signup} className="font-semibold text-primary underline">{t.createAccount}</Link>
      </p>
    </main>
  );
}
