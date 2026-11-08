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
    // Optional: which of the issuer's own facilities it was issued at.
    issued_at_facility_id: z.string().optional(),
    valid_from: date,
    valid_to: date,
    // Optional: left out, ChainTrack writes the certificate PDF.
    file: certificateFileSchema.optional(),
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
  file: File | undefined,
  organisationId: string,
): FormData {
  const body = new FormData();
  body.set("issuer_org_id", organisationId);
  for (const [key, value] of Object.entries(form))
    if (value !== undefined && value !== "") body.set(key, String(value));
  if (file) body.set("file", file);
  return body;
}
