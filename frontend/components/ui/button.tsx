import { Button as ButtonPrimitive } from "@base-ui/react/button"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "cn"

// ChainTrack's buttons (chaintrack-ui → components.md → Button). One primary
// per screen; danger only for irreversible confirmations.
const buttonVariants = cva(
  "inline-flex min-h-tap shrink-0 items-center justify-center gap-2 rounded-md border px-4 text-[15px] font-semibold whitespace-nowrap transition-colors outline-none select-none focus-visible:ring-3 focus-visible:ring-ring/50 aria-disabled:cursor-not-allowed aria-disabled:border-border aria-disabled:bg-surface-sunken aria-disabled:text-ink-subtle disabled:cursor-not-allowed disabled:border-border disabled:bg-surface-sunken disabled:text-ink-subtle [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-5",
  {
    variants: {
      variant: {
        primary: "border-primary bg-primary text-on-primary hover:bg-primary-hover",
        secondary: "border-border-strong bg-surface text-ink hover:bg-surface-sunken",
        ghost: "border-transparent bg-transparent text-primary hover:bg-primary-soft",
        "danger-outline": "border-danger bg-surface text-danger",
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
