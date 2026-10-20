import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { cn } from "@/lib/utils";

import { FieldNote } from "./field-note";

/**
 * One labelled input with its help or its error (components.md → Form fields).
 * Spread react-hook-form's `register(...)` onto it:
 *
 *   <TextField label="Email" error={errors.email?.message} {...register("email")} />
 */
export function TextField({
  label,
  help,
  error,
  mono,
  id,
  name,
  className,
  type,
  ...props
}: React.ComponentProps<"input"> & {
  label: string;
  help?: string;
  error?: string;
  mono?: boolean;
}) {
  const fieldId = id ?? name;
  const noteId = `${fieldId}-note`;
  // A password field gets the show / hide button; everything else is a plain input.
  const Field = type === "password" ? PasswordInput : Input;
  return (
    <div className={cn("grid gap-1.5", className)}>
      <label htmlFor={fieldId} className="text-sm font-semibold">
        {label}
      </label>
      <Field
        id={fieldId}
        name={name}
        aria-invalid={error ? true : undefined}
        aria-describedby={error || help ? noteId : undefined}
        className={cn(mono && "font-mono text-base")}
        type={type}
        {...props}
      />
      <FieldNote id={noteId} error={error} help={help} />
    </div>
  );
}
