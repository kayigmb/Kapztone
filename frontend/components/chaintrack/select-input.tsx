"use client";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";

export type SelectOption = { value: string; label: string };

/** The one select of the app (shadcn Select): filters, sorts and form fields. */
export function SelectInput({
  options,
  value,
  onChange,
  label,
  placeholder,
  invalid,
  disabled,
  className,
  id,
}: {
  options: readonly SelectOption[];
  value: string;
  onChange: (v: string) => void;
  /** Spoken name when there's no visible <label for>. */
  label?: string;
  placeholder?: string;
  invalid?: boolean;
  disabled?: boolean;
  className?: string;
  id?: string;
}) {
  return (
    <Select
      items={options as SelectOption[]}
      value={options.some((o) => o.value === value) ? value : null}
      onValueChange={(v) => onChange((v as string | null) ?? "")}
      disabled={disabled}
    >
      <SelectTrigger
        id={id}
        aria-label={label}
        aria-invalid={invalid || undefined}
        className={cn("w-full min-w-40", className)}
      >
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        {options.map((o) => (
          <SelectItem key={o.value} value={o.value}>
            {o.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
