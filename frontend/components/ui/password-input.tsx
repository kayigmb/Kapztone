"use client";

import { Eye, EyeOff } from "lucide-react";
import { useState } from "react";

import { Input } from "@/components/ui/input";
import { copy } from "@/lib/consts";
import { cn } from "@/lib/utils";

/** A password input with a show / hide button. Works with react-hook-form's `register` (uncontrolled). */
function PasswordInput({ className, ...props }: React.ComponentProps<"input">) {
  const [shown, setShown] = useState(false);
  return (
    <div className="relative">
      <Input
        {...props}
        type={shown ? "text" : "password"}
        // pr-12 keeps typed text clear of the button; the browser's own reveal icon (Edge) is hidden.
        className={cn("pr-12 [&::-ms-clear]:hidden [&::-ms-reveal]:hidden", className)}
      />
      <button
        type="button"
        onClick={() => setShown((s) => !s)}
        aria-pressed={shown}
        aria-label={shown ? copy.password.hide : copy.password.show}
        disabled={props.disabled}
        className="absolute inset-y-0 right-0 grid w-tap place-items-center rounded-r-md text-ink-muted hover:text-ink disabled:opacity-50"
      >
        {shown ? <EyeOff className="size-5" aria-hidden /> : <Eye className="size-5" aria-hidden />}
      </button>
    </div>
  );
}

export { PasswordInput };
