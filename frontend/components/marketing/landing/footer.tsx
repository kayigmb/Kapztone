import { Brand } from "@/components/chaintrack/brand";
import { marketing as m } from "@/lib/consts";
import { routes } from "@/lib/consts";
import { Container, anchor } from "./shared";

export function Footer() {
  const f = m.footer;
  const col = (title: string, links: [string, string][]) => (
    <div className="grid content-start gap-2 text-sm">
      <strong>{title}</strong>
      {links.map(([label, href]) => (
        <a key={label} href={href} className="text-ink-muted hover:text-ink">
          {label}
        </a>
      ))}
    </div>
  );
  return (
    <footer className="border-t border-border bg-surface">
      <Container className="grid gap-8 py-12 sm:grid-cols-3">
        <div className="grid content-start gap-3">
          <Brand />
          <p className="text-sm text-ink-muted">{f.tagline}</p>
        </div>
        {col(f.product, [
          [f.verify, anchor(m.anchors.verify)],
          [f.workspace, routes.signin],
          [f.how, anchor(m.anchors.how)],
        ])}
        {col(f.trustTitle, [
          [f.security, anchor(m.anchors.security)],
          [f.guide, anchor(m.anchors.trust)],
        ])}
      </Container>
      <Container className="border-t border-border py-6 text-xs text-ink-muted">
        {f.rights(new Date().getFullYear())}
      </Container>
    </footer>
  );
}
