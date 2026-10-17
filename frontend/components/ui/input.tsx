import * as React from "react"
import { Input as InputPrimitive } from "@base-ui/react/input"
import { cn } from "cn"

function Input({ className, type, ...props }: React.ComponentProps<"input">) {
  return (
    <InputPrimitive
      type={type}
      data-slot="input"
      className={cn(
        "min-h-tap w-full min-w-0 rounded-md border border-border-strong bg-surface px-3 py-2.5 text-base text-ink outline-none placeholder:text-ink-subtle focus-visible:ring-3 focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:bg-surface-sunken disabled:text-ink-subtle aria-invalid:border-2 aria-invalid:border-danger",
        className
      )}
      {...props}
    />
  )
}

export { Input }
