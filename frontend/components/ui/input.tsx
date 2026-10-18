import * as React from "react"
import { Input as InputPrimitive } from "@base-ui/react/input"
import { cn } from "cn"

function Input({ className, type, ...props }: React.ComponentProps<"input">) {
  return (
    <InputPrimitive
      type={type}
      data-slot="input"
      className={cn(
        // Figma .field-group > input: 44 px, a quiet border, a teal focus ring.
        "min-h-tap w-full min-w-0 rounded-md border border-border-strong/60 bg-surface px-3 py-2.5 text-sm text-ink transition-shadow outline-none placeholder:text-ink-subtle focus-visible:border-primary focus-visible:ring-3 focus-visible:ring-primary/15 disabled:cursor-not-allowed disabled:bg-surface-sunken disabled:text-ink-subtle aria-invalid:border-2 aria-invalid:border-danger aria-invalid:focus-visible:ring-danger/15",
        className
      )}
      {...props}
    />
  )
}

export { Input }
