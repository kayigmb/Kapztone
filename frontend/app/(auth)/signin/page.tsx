import Link from "next/link";
import { redirect } from "next/navigation";

import { getMe } from "@/lib/api";
import { SignInForm } from "./signin-form";

export const metadata = { title: "Sign in · ChainTrack" };

// Phone-first, no app shell (screens.md §03).
export default async function Page({ searchParams }: PageProps<"/signin">) {
  if (await getMe()) redirect("/");
  const { email } = await searchParams;
  return (
    <main className="mx-auto grid min-h-dvh w-full max-w-md content-center gap-6 px-4 py-8">
      <header className="grid gap-1">
        <p className="text-title">ChainTrack</p>
        <h1 className="text-page">Sign in to your organisation</h1>
        <p className="text-ink-muted">Use the email and password for your ChainTrack account.</p>
      </header>
      <section className="rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
        <SignInForm email={typeof email === "string" ? email : ""} />
      </section>
      <p className="text-sm text-ink-muted">
        New to ChainTrack? <Link href="/signup" className="font-semibold text-primary underline">Create your organisation&apos;s account</Link>
      </p>
    </main>
  );
}
