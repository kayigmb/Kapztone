import { marketing as m } from "@/lib/consts";
import { Container, SectionHeading } from "./shared";

export function Faq() {
  const f = m.faq;
  return (
    <section className="py-20">
      <Container className="grid gap-10">
        <SectionHeading eyebrow={f.eyebrow} title={f.title} body={f.body} />
        <div className="grid gap-3 md:grid-cols-2">
          {f.items.map(([q, a], i) => (
            <details
              key={q}
              open={i === 0}
              className="group rounded-xl border border-border bg-surface p-5 shadow-card"
            >
              <summary className="flex cursor-pointer list-none items-center justify-between gap-4 font-semibold">
                {q}
                <span
                  className="text-xl text-primary transition-transform group-open:rotate-45"
                  aria-hidden
                >
                  +
                </span>
              </summary>
              <p className="mt-3 text-sm text-ink-muted">{a}</p>
            </details>
          ))}
        </div>
      </Container>
    </section>
  );
}

