"use client";

import { CircleCheck, OctagonX, Upload } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { Status } from "@/components/chaintrack/status";
import type { OrgDocument } from "@/lib/api";
import { documentTypes, fileProblem } from "@/lib/schemas/org";
import { formatDate } from "@/lib/verification";

const extension = (name: string) => (name.includes(".") ? name.split(".").pop()! : "file").slice(0, 4).toUpperCase();
const size = (bytes: number) => (bytes < 1024 * 1024 ? `${Math.max(1, Math.round(bytes / 1024))} KB` : `${(bytes / 1024 / 1024).toFixed(1)} MB`);

/** One drop zone per document the reviewer needs, with the files already received under each. */
export function DocumentsUpload({ orgId, documents }: { orgId: string; documents: OrgDocument[] }) {
  const router = useRouter();
  const [busy, setBusy] = useState<string>();
  const [problems, setProblems] = useState<Record<string, string>>({});

  async function upload(type: string, file: File | undefined) {
    if (!file) return;
    const problem = fileProblem(file);
    if (problem) return setProblems((p) => ({ ...p, [type]: problem }));
    setProblems((p) => ({ ...p, [type]: "" }));
    setBusy(type);
    try {
      const body = new FormData();
      body.set("type", type);
      body.set("file", file);
      const res = await fetch(`/api/orgs/${orgId}/documents`, { method: "POST", body });
      if (res.status === 201) return router.refresh(); // the list below comes from the API
      const said = res.status === 413 ? fileProblem({ size: Infinity, type: file.type })
        : res.status === 400 ? "This file can't be read as a PDF, PNG or JPEG. Choose another file."
        : "The upload didn't finish. Nothing was saved. Try again.";
      setProblems((p) => ({ ...p, [type]: said! }));
    } catch {
      setProblems((p) => ({ ...p, [type]: "The upload didn't finish. Nothing was saved. Check your connection and try again." }));
    } finally {
      setBusy(undefined);
    }
  }

  return (
    <div className="grid gap-6">
      {documentTypes.map((t) => {
        const received = documents.filter((d) => d.type === t.value);
        const inputId = `document-${t.value}`;
        return (
          <section key={t.value} className="grid gap-2">
            <h3 className="flex flex-wrap items-baseline justify-between gap-x-3 text-title">
              {t.label}
              <span className="text-small font-normal text-ink-muted">{t.required ? "Required" : "Optional"}</span>
            </h3>
            {received.map((d) => (
              <div key={d.id} className="grid grid-cols-[40px_1fr_auto] items-center gap-3 rounded-md border border-border p-3">
                <span className="grid size-10 place-items-center rounded-sm bg-surface-sunken font-mono text-[11px] font-semibold text-ink-muted">
                  {extension(d.file_name)}
                </span>
                <div className="min-w-0">
                  <p className="truncate text-sm font-semibold">{d.file_name}</p>
                  <p className="text-xs text-ink-muted">{size(d.size_bytes)} · {formatDate(d.created_at)}</p>
                </div>
                <Status tone="verified" icon={CircleCheck}>Uploaded</Status>
              </div>
            ))}
            <label htmlFor={inputId} className="grid cursor-pointer justify-items-center gap-1 rounded-lg border-2 border-dashed border-border-strong bg-surface px-4 py-6 text-center text-sm text-ink-muted has-focus-visible:ring-3 has-focus-visible:ring-ring/50">
              <Upload className="size-7 text-primary" aria-hidden />
              <span className="font-semibold text-ink">
                {busy === t.value ? "Uploading…" : received.length ? "Add another file" : "Tap to take a photo or choose a file"}
              </span>
              <span className="text-xs">PDF, PNG or JPEG · up to 10 MB</span>
              <input id={inputId} type="file" accept="application/pdf,image/png,image/jpeg" className="sr-only"
                disabled={busy !== undefined} aria-describedby={problems[t.value] ? `${inputId}-problem` : undefined}
                onChange={(e) => { upload(t.value, e.target.files?.[0]); e.target.value = ""; }} />
            </label>
            {problems[t.value] && (
              <p id={`${inputId}-problem`} role="alert" className="flex gap-1.5 text-small font-semibold text-danger">
                <OctagonX className="mt-px size-4 shrink-0" aria-hidden />
                {problems[t.value]}
              </p>
            )}
          </section>
        );
      })}
    </div>
  );
}
