// Issuing and revoking certificates: mirrors attestation.IssueInput.Validate()
// and attestation.RevokeInput.Validate().
// Pure: `node --test` runs it.
import { z } from "zod";

import { copy, limits, subjectTypes } from "../consts.ts";
import { fileProblem } from "./org.ts";

const v = copy.validation;

/** The signed document: PDF, PNG or JPEG, 10 MB at most. */
export const certificateFileSchema = z
  .instanceof(File)
  .superRefine((file, ctx) => {
    const problem = fileProblem(file);
    if (problem) ctx.addIssue({ code: "custom", message: problem });
  });

/** A form's dates arrive as YYYY-MM-DD (what the date input gives). */
const date = z.string().trim().min(1, v.date);

export const certificateFormSchema = z
  .object({
    subject_type: z.enum(Object.values(subjectTypes) as [string, ...string[]], {
      message: v.subjectId,
    }),
    subject_id: z.string().trim().min(1, v.subjectId),
    type: z.string().trim().min(1, v.certType),
    valid_from: date,
    valid_to: date,
    file: certificateFileSchema,
  })
  .refine((f) => f.valid_to > f.valid_from, {
    path: ["valid_to"],
    message: v.dateOrder,
  });

export type CertificateForm = z.infer<typeof certificateFormSchema>;

/** Revoking needs a reason (backend: reason is required, up to MaxNoteLen). */
export const revokeFormSchema = z.object({
  reason: z
    .string()
    .trim()
    .min(1, v.reasonRequired)
    .max(limits.note, v.tooLong(limits.note)),
});

export type RevokeForm = z.infer<typeof revokeFormSchema>;

/** What travels with the file (POST /attestations is multipart). */
export function certificateBody(
  form: Omit<CertificateForm, "file">,
  file: File,
  organisationId: string,
): FormData {
  const body = new FormData();
  body.set("issuer_org_id", organisationId);
  for (const [key, value] of Object.entries(form)) body.set(key, String(value));
  body.set("file", file);
  return body;
}
