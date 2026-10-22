import { OctagonX } from "lucide-react";

/** The line under a field: its error or its help, never both. */
export function FieldNote({
  id,
  error,
  help,
}: {
  id?: string;
  error?: string;
  help?: React.ReactNode;
}) {
  if (error)
    return (
      <p id={id} className="flex gap-1.5 text-small font-semibold text-danger">
        <OctagonX className="mt-px size-4 shrink-0" aria-hidden />
        {error}
      </p>
    );
  if (help == null || help === "") return null;
  return (
    <p id={id} className="text-small text-ink-subtle">
      {help}
    </p>
  );
}
