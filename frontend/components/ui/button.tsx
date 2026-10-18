import { Button as ButtonPrimitive } from "@base-ui/react/button"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "cn"

// ChainTrack's buttons (chaintrack-ui → components.md → Button; Figma
// .primary-button / .outline-button / .danger-button). One primary per
// screen; solid danger only for irreversible confirmations.
const buttonVariants = cva(
  "inline-flex min-h-tap shrink-0 items-center justify-center gap-2 rounded-md border px-4 text-sm font-semibold whitespace-nowrap transition-colors outline-none select-none focus-visible:ring-3 focus-visible:ring-ring/40 aria-disabled:cursor-not-allowed aria-disabled:border-border aria-disabled:bg-surface-sunken aria-disabled:text-ink-subtle aria-disabled:shadow-none disabled:cursor-not-allowed disabled:border-border disabled:bg-surface-sunken disabled:text-ink-subtle disabled:shadow-none [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-[18px]",
  {
    variants: {
      variant: {
        primary: "border-primary bg-primary text-on-primary shadow-button hover:border-primary-hover hover:bg-primary-hover",
        secondary: "border-border bg-surface text-ink hover:border-primary hover:text-primary",
        ghost: "border-transparent bg-transparent text-primary hover:bg-primary-soft",
        // A refusal that can be taken back ("Reject", "Reject & raise flag").
        "danger-soft": "border-transparent bg-danger-soft text-danger-ink hover:border-danger/40",
        "danger-outline": "border-danger bg-surface text-danger hover:bg-danger-soft",
        danger: "border-danger bg-danger text-on-danger",
      },
      size: {
        default: "",
        scan: "h-tap-scan px-6 text-base", // scanning, handover and phone primary actions
      },
      block: { true: "w-full" },
    },
    defaultVariants: { variant: "primary", size: "default" },
  }
)

function Button({
  className,
  variant,
  size,
  block,
  ...props
}: ButtonPrimitive.Props & VariantProps<typeof buttonVariants>) {
  return (
    <ButtonPrimitive
      data-slot="button"
      className={cn(buttonVariants({ variant, size, block, className }))}
      {...props}
    />
  )
}

export { Button, buttonVariants }
