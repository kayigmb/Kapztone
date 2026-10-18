"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { CircleCheck, OctagonX, Upload } from "lucide-react";
import { useState } from "react";

import { Status } from "@/components/chaintrack/status";
import { ApiError, endpoints, type OrgDocument } from "@/lib/api";
import { copy, documentTypes, queryKeys, upload as uploadRules } from "@/lib/consts";
import { fileProblem } from "@/lib/schemas/org";
import { formatDate } from "@/lib/verification";

const t = copy.documents;
const extension = (name: string) => (name.includes(".") ? name.split(".").pop()! : "file").slice(0, 4).toUpperCase();
const size = (bytes: number) => (bytes < 1024 * 1024 ? `${Math.max(1, Math.round(bytes / 1024))} KB` : `${(bytes / 1024 / 1024).toFixed(1)} MB`);

/** One drop zone per document the reviewer needs, with the files already received under each. */
export function DocumentsUpload({ orgId, documents }: { orgId: string; documents: OrgDocument[] }) {
  const client = useQueryClient();
  const [problems, setProblems] = useState<Record<string, string>>({});
  const say = (type: string, problem: string) => setProblems((p) => ({ ...p, [type]: problem }));

  const upload = useMutation({
    mutationFn: ({ type, file }: { type: string; file: File }) => endpoints.uploadDocument(orgId, type, file),
    onMutate: ({ type }) => say(type, ""),
    // A new document can send a rejected organisation back to review: refresh it, its list and who is signed in.
    onSuccess: () => Promise.all([
      client.invalidateQueries({ queryKey: queryKeys.org(orgId) }),
      client.invalidateQueries({ queryKey: queryKeys.me }),
    ]),
    onError: (err, { type, file }) => {
      const status = err instanceof ApiError ? err.status : -1;
      say(type, status === 413 ? fileProblem({ size: Infinity, type: file.type })!
        : status === 400 ? t.unreadable
        : status === 0 || status === 502 ? t.offline
        : t.failed);
    },
  });

  function choose(type: string, file: File | undefined) {
    if (!file) return;
    const problem = fileProblem(file);
    if (problem) return say(type, problem);
    upload.mutate({ type, file });
  }

  const busy = upload.isPending ? upload.variables?.type : undefined;

  return (
    <div className="grid gap-6">
      {documentTypes.map((d) => {
        const received = documents.filter((doc) => doc.type === d.value);
        const inputId = `document-${d.value}`;
        return (
          <section key={d.value} className="grid gap-2">
            <h3 className="flex flex-wrap items-baseline justify-between gap-x-3 text-title">
              {d.label}
              <span className="text-small font-normal text-ink-muted">{d.required ? t.required : t.optional}</span>
            </h3>
            {received.map((doc) => (
              <div key={doc.id} className="grid grid-cols-[40px_1fr_auto] items-center gap-3 rounded-md border border-border p-3">
                <span className="grid size-10 place-items-center rounded-sm bg-surface-sunken font-mono text-xs font-semibold text-ink-muted">
                  {extension(doc.file_name)}
                </span>
                <div className="min-w-0">
                  <p className="truncate text-sm font-semibold">{doc.file_name}</p>
                  <p className="text-xs text-ink-muted">{size(doc.size_bytes)} · {formatDate(doc.created_at)}</p>
                </div>
                <Status tone="verified" icon={CircleCheck}>{t.uploaded}</Status>
              </div>
            ))}
            <label htmlFor={inputId} className="grid cursor-pointer justify-items-center gap-1 rounded-lg border-2 border-dashed border-border-strong bg-surface px-4 py-6 text-center text-sm text-ink-muted has-focus-visible:ring-3 has-focus-visible:ring-ring/50">
              <Upload className="size-7 text-primary" aria-hidden />
              <span className="font-semibold text-ink">
                {busy === d.value ? t.uploading : received.length ? t.addAnother : t.choose}
              </span>
              <span className="text-xs">{t.hint}</span>
              <input id={inputId} type="file" accept={uploadRules.types.join(",")} className="sr-only"
                disabled={busy !== undefined} aria-describedby={problems[d.value] ? `${inputId}-problem` : undefined}
                onChange={(e) => { choose(d.value, e.target.files?.[0]); e.target.value = ""; }} />
            </label>
            {problems[d.value] && (
              <p id={`${inputId}-problem`} role="alert" className="flex gap-1.5 text-small font-semibold text-danger">
                <OctagonX className="mt-px size-4 shrink-0" aria-hidden />
                {problems[d.value]}
              </p>
            )}
          </section>
        );
      })}
    </div>
  );
}
