import Link from "next/link";

import { copy, routes } from "@/lib/consts";
import { indexed } from "@/lib/site";
import { SignInForm } from "./signin-form";

const t = copy.signin;

export const metadata = { title: copy.meta.signin, ...indexed(routes.signin) };

// Phone-first, inside the access layout.
export default async function Page({ searchParams }: PageProps<"/signin">) {
  const { email } = await searchParams;
  return (
    <main className="mx-auto grid w-full max-w-100 gap-6">
      <header className="grid gap-2">
        <span className="eyebrow text-primary">{t.eyebrow}</span>
        <h1 className="text-page">{t.title}</h1>
        <p className="text-sm text-ink-muted">{t.sub}</p>
      </header>
      <SignInForm email={typeof email === "string" ? email : ""} />
      <p className="text-center text-sm text-ink-muted">
        {t.newHere}{" "}
        <Link
          href={routes.signup}
          className="font-semibold text-primary hover:underline"
        >
          {t.createAccount}
        </Link>
      </p>
    </main>
  );
}
