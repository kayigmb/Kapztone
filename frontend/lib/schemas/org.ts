import { z } from "zod";

import {
  copy,
  documentTypes,
  limits,
  orgKinds,
  reopeningDecisions,
  reviewDecisions,
  upload,
} from "../consts.ts";

// What a new organisation submits, and what a reviewer decides. Each rule
// mirrors the API's own check (backend/internal/org/validate.go,
// auth/values.go), so a form refuses what the API would refuse; the API
// stays the authority.

const v = copy.validation;

export const accountSchema = z.object({
  user_name: z
    .string()
    .trim()
    .min(1, v.fullName)
    .max(limits.name, v.tooLong(limits.name)),
  email: z
    .string()
    .trim()
    .toLowerCase()
    .max(limits.email, v.emailLong)
    .pipe(z.email(v.email)),
  password: z
    .string()
    .min(limits.password.min, v.passwordShort(limits.password.min))
    .max(limits.password.max, v.tooLong(limits.password.max)),
  confirm_password: z.string().min(1, v.passwordConfirm),
  consent: z.literal(true, v.consent),
});

export const organisationSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, v.legalName)
    .max(limits.name, v.tooLong(limits.name)),
  slug: z
    .string()
    .trim()
    .toLowerCase()
    .regex(limits.slugPattern, v.slug)
    .refine((s) => !limits.reservedSlugs.includes(s), v.slugReserved),
  kind: z.enum(
    orgKinds.map((k) => k.value),
    v.kind,
  ),
});

/** The whole sign-up; the password is typed twice (only the first is sent). */
export const registrationSchema = accountSchema
  .extend(organisationSchema.shape)
  .refine((r) => r.password === r.confirm_password, {
    path: ["confirm_password"],
    message: v.passwordsDiffer,
    // zod skips an object refine while any field is invalid; the organisation
    // fields are still empty on step 1, so check whenever both passwords are there.
    when: ({ value }) => {
      const r = value as Partial<
        Record<"password" | "confirm_password", unknown>
      >;
      return (
        typeof r.password === "string" && typeof r.confirm_password === "string"
      );
    },
  });
export type Registration = z.infer<typeof registrationSchema>;

/** "Kitabi Washing Station Ltd." → "kitabi-washing-station-ltd": a first guess at the address. */
export function suggestSlug(name: string): string {
  return name
    .toLowerCase()
    .normalize("NFKD")
    .replace(/\p{M}/gu, "")
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 63)
    .replace(/-+$/, "");
}

/** The problem with a file, said the way a field error is, or undefined if it can be uploaded. */
export function fileProblem(file: {
  size: number;
  type: string;
}): string | undefined {
  if (!upload.types.includes(file.type)) return v.fileType;
  if (file.size > upload.maxBytes) return v.fileSize;
  return undefined;
}

/** Documents still missing before an organisation can be reviewed. */
export function missingDocuments(uploaded: { type: string }[]) {
  return documentTypes.filter(
    (t) => t.required && !uploaded.some((d) => d.type === t.value),
  );
}

// A reviewer's decision (POST /orgs/{id}/verify): a reason is required to send it back.
export const reviewSchema = z
  .object({
    decision: z.enum(reviewDecisions),
    reason: z.string().trim(),
  })
  .refine((r) => !reopeningDecisions.includes(r.decision) || r.reason !== "", {
    path: ["reason"],
    message: v.reason,
  });
export type Review = z.infer<typeof reviewSchema>;

// Accreditation (PUT /orgs/{id}/accreditations): which certificate types, from when, until when.
// The form sets one window for every type it sends; the API takes one per type.
const isoDay = /^\d{4}-\d{2}-\d{2}$/;
export const accreditationSchema = z
  .object({
    types: z.array(z.string()).min(1, v.accreditationTypes),
    valid_from: z.string().regex(isoDay, v.date),
    valid_to: z.string().regex(isoDay, v.date),
  })
  .refine((a) => a.valid_to > a.valid_from, {
    path: ["valid_to"],
    message: v.dateOrder,
  });
export type Accreditation = z.infer<typeof accreditationSchema>;

/** The form's values as the API's list: the window runs from the start of the first day to the end of the last. */
export function toAccreditations(a: Accreditation) {
  return a.types.map((type) => ({
    type,
    valid_from: `${a.valid_from}T00:00:00Z`,
    valid_to: `${a.valid_to}T23:59:59Z`,
  }));
}
