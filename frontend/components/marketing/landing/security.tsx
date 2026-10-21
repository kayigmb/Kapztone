import { ShieldCheck } from "lucide-react";
import { marketing as m } from "@/lib/consts";
import { Container, SectionHeading, icons } from "./shared";

export function Security() {
  const s = m.security;
  return (
    <section
      id={m.anchors.security}
      className="border-y border-border bg-surface py-20"
    >
      <Container className="grid gap-12 lg:grid-cols-5">
        <div className="grid content-start gap-6 lg:col-span-2">
          <SectionHeading
            eyebrow={s.eyebrow}
            title={
              <>
                {s.title}
                <br />
                {s.titleLine}
              </>
            }
            body={s.body}
          />
          <div className="flex gap-3 rounded-xl border border-border bg-surface-sunken p-4">
            <ShieldCheck
              className="mt-0.5 size-5 shrink-0 text-primary"
              aria-hidden
            />
            <div className="grid gap-1">
              <strong className="text-sm">{s.assuranceTitle}</strong>
              <p className="text-sm text-ink-muted">{s.assuranceBody}</p>
            </div>
          </div>
        </div>
        <ul className="grid gap-4 sm:grid-cols-2 lg:col-span-3">
          {s.cards.map((c) => {
            const Icon = icons[c.icon];
            return (
              <li
                key={c.title}
                className="flex gap-3 rounded-xl border border-border p-5"
              >
                <span className="grid size-10 shrink-0 place-items-center rounded-lg bg-primary-soft text-primary">
                  <Icon className="size-5" aria-hidden />
                </span>
                <div className="grid gap-1">
                  <h3 className="text-sm font-semibold">{c.title}</h3>
                  <p className="text-sm text-ink-muted">{c.body}</p>
                </div>
              </li>
            );
          })}
        </ul>
      </Container>
    </section>
  );
}
