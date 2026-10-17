import Link from "next/link";
import { redirect } from "next/navigation";

import { getMe } from "@/lib/api";
import { SignUpForm } from "./signup-form";

export const metadata = { title: "Create account · ChainTrack" };

// Phone-first, no app shell (screens.md §03, steps 1 and 2).
export default async function Page() {
  if (await getMe()) redirect("/");
  return (
    <main className="mx-auto grid min-h-dvh w-full max-w-xl content-center gap-6 px-4 py-8">
      <header className="grid gap-1">
        <p className="text-title">ChainTrack</p>
        <h1 className="text-page">Create your organisation&apos;s account</h1>
        <p className="text-ink-muted">
          You&apos;ll need your business registration and any operating licences. Verification takes about 2 working days.
        </p>
      </header>
      <section className="rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
        <SignUpForm />
      </section>
      <p className="text-sm text-ink-muted">
        Already have an account? <Link href="/signin" className="font-semibold text-primary underline">Sign in</Link>
      </p>
    </main>
  );
}
