import { OctagonX, type LucideIcon } from "lucide-react";

/**
 * A radio group shown as cards (components.md → Choice cards). Spread
 * react-hook-form's `register(...)` onto it; the chosen card is styled by CSS
 * from the radio's own state, so it needs no script.
 */
export function ChoiceCards({
  legend,
  options,
  error,
  ...radio
}: React.ComponentProps<"input"> & {
  legend: string;
  options: readonly {
    value: string;
    label: string;
    sub?: string;
    icon?: LucideIcon;
  }[];
  error?: string;
}) {
  return (
    <fieldset className="grid gap-1.5" aria-invalid={error ? true : undefined}>
      <legend className="mb-1.5 text-sm font-semibold">{legend}</legend>
      <div className="grid gap-2 md:grid-cols-2">
        {options.map((o) => (
          <label
            key={o.value}
            className="grid cursor-pointer grid-cols-[1fr_auto] items-center gap-x-3 rounded-md border border-border-strong bg-surface p-3 has-checked:border-2 has-checked:border-primary has-checked:bg-primary-soft has-focus-visible:ring-3 has-focus-visible:ring-ring/50"
          >
            <span>
              <span className="block text-base font-semibold">{o.label}</span>
              {o.sub && (
                <span className="block text-small text-ink-muted">{o.sub}</span>
              )}
            </span>
            <input
              type="radio"
              value={o.value}
              className="peer sr-only"
              {...radio}
            />
            <span
              className="size-5.5 rounded-full border-2 border-border-strong peer-checked:border-7 peer-checked:border-primary"
              aria-hidden
            />
          </label>
        ))}
      </div>
      {error && (
        <p className="flex gap-1.5 text-small font-semibold text-danger">
          <OctagonX className="mt-px size-4 shrink-0" aria-hidden />
          {error}
        </p>
      )}
    </fieldset>
  );
}
