import type { LucideIcon } from "lucide-react";
import {
  MapPin,
  Route,
  Layers,
  ShieldCheck,
  Link2,
  FileText,
  Clock,
  Check,
  CircleCheck,
  TriangleAlert,
  OctagonX,
} from "lucide-react";
import type { MarketingRole } from "@/lib/consts";
import { cn } from "@/lib/utils";

export function Container({
  className,
  children,
}: {
  className?: string;
  children: React.ReactNode;
}) {
  return (
    <div className={cn("mx-auto w-full max-w-7xl px-4 lg:px-8", className)}>
      {children}
    </div>
  );
}

/** Layout-stable placeholder while a below-fold section loads. */
export function SectionSkeleton() {
  return (
    <div aria-busy="true" className="py-20">
      <Container className="grid gap-3">
        <div className="h-4 w-32 rounded-sm bg-surface-sunken" />
        <div className="h-8 w-2/3 rounded-sm bg-surface-sunken" />
        <div className="h-4 w-1/2 rounded-sm bg-surface-sunken" />
      </Container>
    </div>
  );
}

export function SectionHeading({
  eyebrow,
  title,
  body,
  center,
}: {
  eyebrow: string;
  title: React.ReactNode;
  body?: string;
  center?: boolean;
}) {
  return (
    <div
      className={cn("grid max-w-2xl gap-3", center && "mx-auto text-center")}
    >
      <span className="eyebrow text-primary">{eyebrow}</span>
      <h2 className="text-3xl font-semibold tracking-tight lg:text-4xl">
        {title}
      </h2>
      {body && <p className="text-ink-muted">{body}</p>}
    </div>
  );
}

export const icons: Record<string, LucideIcon> = {
  map: MapPin,
  route: Route,
  layers: Layers,
  shield: ShieldCheck,
  link: Link2,
  file: FileText,
  clock: Clock,
  check: Check,
};
export const verdictIcon = {
  verified: CircleCheck,
  warning: TriangleAlert,
  danger: OctagonX,
} as const;
export const roleIcon: Record<MarketingRole, LucideIcon> = {
  producer: MapPin,
  logistics: Route,
  certifier: FileText,
  regulator: ShieldCheck,
};
export const heroPhoto =
  "https://images.unsplash.com/photo-1756245994771-f06fcaee3a36?crop=entropy&cs=tinysrgb&fit=crop&fm=jpg&q=82&w=1080";
export const anchor = (id: string) => `#${id}`;
