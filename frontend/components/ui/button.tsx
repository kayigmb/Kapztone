import { Button as ButtonPrimitive } from "@base-ui/react/button";
import { cva, type VariantProps } from "class-variance-authority";
import { LoaderCircle } from "lucide-react";
import { cn } from "cn";

// ChainTrack's buttons (chaintrack-ui → components.md → Button; Figma
// .primary-button / .outline-button / .danger-button). One primary per
// screen; solid danger only for irreversible confirmations.
const buttonVariants = cva(
  "inline-flex h-10 shrink-0 items-center justify-center gap-2 rounded-md border px-3.5 text-sm lg:h-9 font-semibold whitespace-nowrap transition-colors outline-none select-none focus-visible:ring-3 focus-visible:ring-ring/40 aria-disabled:cursor-not-allowed aria-disabled:border-border aria-disabled:bg-surface-sunken aria-disabled:text-ink-subtle aria-disabled:shadow-none disabled:cursor-not-allowed disabled:border-border disabled:bg-surface-sunken disabled:text-ink-subtle disabled:shadow-none [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
  {
    variants: {
      variant: {
        primary:
          "border-primary bg-primary text-on-primary hover:border-primary-hover hover:bg-primary-hover",
        secondary:
          "border-border bg-surface text-ink hover:border-primary hover:text-primary",
        // shadcn's name for the outlined button (tablecn and other shadcn blocks use it).
        outline:
          "border-border bg-surface text-ink hover:border-primary hover:text-primary",
        ghost:
          "border-transparent bg-transparent text-primary hover:bg-primary-soft",
        // A refusal that can be taken back ("Reject", "Reject & raise flag").
        "danger-soft":
          "border-transparent bg-danger-soft text-danger-ink hover:border-danger/40",
        "danger-outline":
          "border-danger bg-surface text-danger hover:bg-danger-soft",
        danger: "border-danger bg-danger text-on-danger",
      },
      size: {
        default: "",
        icon: "size-10 px-0 lg:size-9", // square, icon only: give it an aria-label
        scan: "h-11 px-5 lg:h-9 lg:px-3.5", // bigger on phones only: scanning, handover, phone primary actions
      },
      block: { true: "w-full" },
    },
    defaultVariants: { variant: "primary", size: "default" },
  },
);

/**
 * A busy button: `loading` disables it, spins a `LoaderCircle` and shows
 * `loadingText` instead of the children. Never write
 * `disabled={m.isPending}` + `{m.isPending ? t.submitting : t.submit}` again:
 *
 *   <Button type="submit" block loading={lock.isPending} loadingText={t.submitting}>
 *     {t.submit}
 *   </Button>
 */
function Button({
  className,
  variant,
  size,
  block,
  loading,
  loadingText,
  disabled,
  children,
  ...props
}: ButtonPrimitive.Props &
  VariantProps<typeof buttonVariants> & {
    /** The mutation is running: disable, spin and show `loadingText`. */
    loading?: boolean;
    /** What to say while loading; the children when idle. */
    loadingText?: React.ReactNode;
  }) {
  return (
    <ButtonPrimitive
      data-slot="button"
      className={cn(buttonVariants({ variant, size, block, className }))}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      {...props}
    >
      {loading && <LoaderCircle className="animate-spin" aria-hidden />}
      {loading && loadingText !== undefined ? loadingText : children}
    </ButtonPrimitive>
  );
}

export { Button, buttonVariants };
