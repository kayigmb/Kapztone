import { copy } from "@/lib/consts";
import { AcceptForm } from "./accept-form";

const t = copy.invite;

export const metadata = { title: copy.meta.invite };

// Opened from the invitation e-mail.
export default async function Page({ params }: PageProps<"/invite/[token]">) {
  const { token } = await params;
  return (
    <main className="mx-auto grid w-full max-w-100 gap-6">
      <header className="grid gap-2">
        <span className="eyebrow text-primary">{t.eyebrow}</span>
        <h1 className="text-page">{t.title}</h1>
        <p className="text-sm text-ink-muted">{t.sub}</p>
      </header>
      <AcceptForm token={token} />
    </main>
  );
}
