import Link from "next/link";

import { copy, routes } from "@/lib/consts";
import { SignUpForm } from "./signup-form";

const t = copy.signup;

export const metadata = { title: copy.meta.signup };

// Phone-first, inside the access layout (screens.md §03, steps 1 and 2; Figma AuthFlow → sign up).
export default function Page() {
  return (
    <main className="mx-auto grid w-full max-w-[540px] gap-6">
      <header className="grid gap-2">
        <span className="eyebrow text-primary">{t.eyebrow}</span>
        <h1 className="text-page">{t.title}</h1>
        <p className="text-sm text-ink-muted">{t.sub}</p>
      </header>
      <SignUpForm />
      <p className="text-center text-sm text-ink-muted">
        {t.haveAccount} <Link href={routes.signin} className="font-semibold text-primary hover:underline">{t.signIn}</Link>
      </p>
    </main>
  );
}
