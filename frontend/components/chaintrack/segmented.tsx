import { cn } from "@/lib/utils";

/** A small set of mutually exclusive choices shown as one pill row (filters, modes). */
export function Segmented<V extends string>({
  label,
  options,
  value,
  onChange,
}: {
  label: string;
  options: readonly { value: V; label: string }[];
  value: V;
  onChange: (v: V) => void;
}) {
  return (
    <div
      role="group"
      aria-label={label}
      className="inline-flex flex-wrap gap-0.5 rounded-md border border-border bg-surface-sunken p-0.5"
    >
      {options.map((o) => (
        <button
          key={o.value}
          type="button"
          aria-pressed={value === o.value}
          onClick={() => onChange(o.value)}
          className={cn(
            "h-8 rounded-sm px-3 text-sm",
            value === o.value
              ? "bg-surface font-semibold text-ink shadow-card"
              : "text-ink-muted hover:text-ink",
          )}
        >
          {o.label}
        </button>
      ))}
    </div>
  );
}
