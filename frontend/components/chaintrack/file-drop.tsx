"use client";

import { FileText, Upload, X } from "lucide-react";
import { useId, useState } from "react";

import { FieldNote } from "@/components/chaintrack/field-note";
import { cn } from "@/lib/utils";

/** One file, dropped on the zone or picked by clicking it (the label opens the picker, keyboard included). */
export function FileDrop({
  label,
  help,
  prompt,
  accept,
  file,
  onFile,
  error,
}: {
  label: string;
  help?: string;
  /** What the empty zone says, like "Drop the licence here, or choose a file". */
  prompt: string;
  accept?: string;
  file?: File;
  onFile: (f: File | undefined) => void;
  error?: string;
}) {
  const id = useId();
  const [over, setOver] = useState(false);
  return (
    <div className="grid gap-1.5">
      <span className="text-sm font-semibold">{label}</span>
      {file ? (
        <div className="flex min-h-tap items-center gap-3 rounded-md border border-border-strong bg-surface px-3 py-2">
          <FileText className="size-5 shrink-0 text-primary" aria-hidden />
          <span className="min-w-0 flex-1 truncate text-sm">{file.name}</span>
          <button
            type="button"
            onClick={() => onFile(undefined)}
            aria-label={`Remove ${file.name}`}
            className="grid size-8 place-items-center rounded-md text-ink-muted hover:bg-surface-sunken"
          >
            <X className="size-4" aria-hidden />
          </button>
        </div>
      ) : (
        <label
          htmlFor={id}
          onDragOver={(e) => {
            e.preventDefault();
            setOver(true);
          }}
          onDragLeave={() => setOver(false)}
          onDrop={(e) => {
            e.preventDefault();
            setOver(false);
            onFile(e.dataTransfer.files[0]);
          }}
          className={cn(
            "grid cursor-pointer justify-items-center gap-2 rounded-md border-2 border-dashed px-4 py-6 text-center text-sm text-ink-muted has-focus-visible:ring-3 has-focus-visible:ring-ring/50",
            over
              ? "border-primary bg-primary-soft"
              : error
                ? "border-danger"
                : "border-border-strong",
          )}
        >
          <Upload className="size-6" aria-hidden />
          {prompt}
          <input
            id={id}
            type="file"
            accept={accept}
            className="sr-only"
            aria-label={label}
            aria-invalid={error ? true : undefined}
            onChange={(e) => onFile(e.target.files?.[0])}
          />
        </label>
      )}
      <FieldNote error={error} help={help} />
    </div>
  );
}
