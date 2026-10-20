import { ArrowRight } from "lucide-react";
import { StatusChip } from "@/components/chaintrack/status";
import { buttonVariants } from "@/components/ui/button";
import { marketing as m } from "@/lib/consts";
import { cn } from "@/lib/utils";
import { Container, anchor, verdictIcon } from "./shared";

export function TrustVerdicts() {
  const t = m.trust;
  return (
    <section id={m.anchors.trust} className="pb-20">
      <Container>
        <div className="grid gap-8 rounded-2xl bg-navy-900 p-8 text-hero-ink lg:grid-cols-2 lg:p-12">
          <div className="grid content-start gap-4">
            <span className="eyebrow text-hero-accent">{t.eyebrow}</span>
            <h2 className="text-3xl font-semibold tracking-tight">{t.title}</h2>
            <p className="text-hero-muted">{t.body}</p>
            <a
              href={anchor(m.anchors.verify)}
              className={cn(buttonVariants(), "w-fit")}
            >
              {t.cta}
              <ArrowRight aria-hidden />
            </a>
          </div>
          <ul className="grid gap-3">
            {t.verdicts.map((v) => (
              <li
                key={v.label}
                className="grid gap-1 rounded-xl bg-surface p-4 text-ink"
              >
                <StatusChip tone={v.tone} icon={verdictIcon[v.tone]}>
                  {v.label}
                </StatusChip>
                <strong className="text-sm">{v.body}</strong>
                <span className="text-xs text-ink-muted">{t.note}</span>
              </li>
            ))}
          </ul>
        </div>
      </Container>
    </section>
  );
}

