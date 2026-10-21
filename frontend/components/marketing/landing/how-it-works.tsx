import { marketing as m } from "@/lib/consts";
import { Container, SectionHeading, icons } from "./shared";

export function HowItWorks() {
  const h = m.how;
  return (
    <section id={m.anchors.how} className="py-20">
      <Container className="grid gap-12">
        <SectionHeading
          eyebrow={h.eyebrow}
          title={h.title}
          body={h.body}
          center
        />
        <ol className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {h.steps.map((s, i) => {
            const Icon = icons[s.icon];
            return (
              <li
                key={s.title}
                className="grid content-start gap-3 rounded-xl border border-border bg-surface p-6 shadow-card"
              >
                <span className="font-mono text-xs text-ink-subtle">
                  {String(i + 1).padStart(2, "0")}
                </span>
                <span className="grid size-10 place-items-center rounded-lg bg-primary-soft text-primary">
                  <Icon className="size-5" aria-hidden />
                </span>
                <h3 className="text-lg font-semibold">{s.title}</h3>
                <p className="text-sm text-ink-muted">{s.body}</p>
              </li>
            );
          })}
        </ol>
      </Container>
    </section>
  );
}
