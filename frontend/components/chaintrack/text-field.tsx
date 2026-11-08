import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { cn } from "@/lib/utils";

import { FieldNote } from "./field-note";

/** One labelled input with its help or its error. */
/** A button beside a TextField in a top-aligned row: dropped by the label's height (text-sm line + gap-1.5) so it lines up with the input, whatever note grows under the field. */
export const besideField = "mt-[1.625rem]";
/** The same, for rows that stack below md. */
export const besideFieldMd = "md:mt-[1.625rem]";

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
