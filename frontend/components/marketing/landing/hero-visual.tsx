import {
  Check,
  CircleCheck,
  MapPin,
  Package,
  Route,
  ShieldCheck,
} from "lucide-react";
import Image from "next/image";
import { StatusChip } from "@/components/chaintrack/status";
import { marketing as m } from "@/lib/consts";
import { cn } from "@/lib/utils";
import { heroPhoto } from "./shared";

export function HeroVisual() {
  const h = m.hero;
  return (
    <div className="relative mx-auto w-full max-w-lg pb-20 lg:pb-0">
      <div className="relative aspect-4/5 overflow-hidden rounded-2xl border border-hero-line">
        <Image
          src={heroPhoto}
          alt={h.photoAlt}
          fill
          sizes="(min-width: 1024px) 32rem, 90vw"
          className="object-cover"
          priority
        />
        <div className="absolute inset-x-4 bottom-4 flex items-center gap-3 rounded-xl bg-surface p-3 text-ink shadow-overlay">
          <span className="grid size-9 place-items-center rounded-md bg-primary-soft text-primary">
            <MapPin className="size-4" aria-hidden />
          </span>
          <div className="grid min-w-0 flex-1">
            <small className="eyebrow">
              {h.example} · {h.originLabel}
            </small>
            <strong className="truncate text-sm">{h.originPlace}</strong>
          </div>
          <StatusChip tone="verified" icon={CircleCheck}>
            {h.active}
          </StatusChip>
        </div>
      </div>
      <div className="absolute -bottom-2 -left-2 w-72 rounded-xl border border-border bg-surface p-4 text-ink shadow-overlay sm:-left-10 lg:top-10 lg:bottom-auto">
        <div className="flex items-center gap-3">
          <span className="grid size-9 place-items-center rounded-md bg-primary-soft text-primary">
            <Package className="size-4" aria-hidden />
          </span>
          <div className="grid min-w-0">
            <small className="eyebrow">
              {h.example} · {h.passportLabel}
            </small>
            <strong className="truncate text-sm">{h.passportName}</strong>
            <code className="code text-xs">{h.passportCode}</code>
          </div>
        </div>
        <ol className="mt-4 grid grid-cols-4 gap-1 text-center text-xs text-ink-muted">
          {h.stops.map((s, i) => (
            <li key={s} className="grid justify-items-center gap-1">
              <span
                className={cn(
                  "grid size-6 place-items-center rounded-full border-2",
                  i < 2
                    ? "border-verified bg-verified text-on-verified"
                    : i === 2
                      ? "border-primary text-primary"
                      : "border-border",
                )}
              >
                {i < 2 ? (
                  <Check className="size-3" aria-hidden />
                ) : i === 2 ? (
                  <Route className="size-3" aria-hidden />
                ) : null}
              </span>
              {s}
            </li>
          ))}
        </ol>
        <div className="mt-4 flex items-center justify-between gap-2 border-t border-border pt-3 text-xs text-ink-muted">
          <StatusChip tone="verified" icon={ShieldCheck}>
            {h.verified}
          </StatusChip>
          {h.checks}
        </div>
      </div>
      <div className="absolute right-0 bottom-6 hidden items-center gap-3 rounded-xl bg-surface p-3 text-ink shadow-overlay sm:flex lg:-right-6">
        <span className="grid size-8 place-items-center rounded-full bg-verified text-on-verified">
          <Check className="size-4" aria-hidden />
        </span>
        <div className="grid">
          <strong className="text-sm">{h.signedTitle}</strong>
          <small className="text-xs text-ink-muted">{h.signedSub}</small>
        </div>
      </div>
    </div>
  );
}
