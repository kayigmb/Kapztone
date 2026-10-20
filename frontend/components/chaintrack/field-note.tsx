import { OctagonX } from "lucide-react";

/**
 * The line under a field: its error or its help, never both (patterns.md →
 * Form fields). An error wins and is announced; otherwise the help shows.
 * Every hand-written `errors.x?.message ?? help` paragraph should be this.
 */
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
      <p
        id={id}
        className="flex gap-1.5 text-small font-semibold text-danger"
      >
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
