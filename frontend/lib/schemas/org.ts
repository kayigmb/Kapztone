import { z } from "zod";

// What a new organisation submits. Each rule mirrors the API's own check
// (backend/internal/org/validate.go, auth/values.go), so the form refuses
// what POST /orgs would refuse; the API stays the authority.

export const kinds = [
  { value: "producer", label: "Producer", sub: "Farms, cooperatives, mines" },
  { value: "processor", label: "Processor", sub: "Washing stations, smelters, factories" },
  { value: "transporter", label: "Logistics", sub: "Transporters, warehouses, forwarders" },
  { value: "exporter", label: "Exporter", sub: "Ships goods out of the country" },
  { value: "importer", label: "Importer", sub: "Receives goods from abroad" },
  { value: "certifier", label: "Certifier", sub: "Issues certificates; needs accreditation" },
  { value: "regulator", label: "Regulator", sub: "Oversees a sector; needs accreditation" },
] as const;

// The onboarding flow, in order (screens.md §03).
export const onboardingSteps = ["Account", "Organisation", "Documents", "Review"];

export const maxNameLength = 200;
export const passwordLength = { min: 12, max: 72 };
// The address is a hostname label, and these names are taken by the system.
const slugPattern = /^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$/;
export const reservedSlugs = ["www", "api", "admin", "app"];

export const accountSchema = z.object({
  user_name: z.string().trim().min(1, "Enter your full name.").max(maxNameLength, `Use ${maxNameLength} characters or fewer.`),
  email: z.string().trim().toLowerCase().max(254, "Use a shorter email address.").pipe(z.email("Enter your work email, like name@company.com.")),
  password: z.string()
    .min(passwordLength.min, `Use at least ${passwordLength.min} characters.`)
    .max(passwordLength.max, `Use ${passwordLength.max} characters or fewer.`),
  consent: z.literal(true, "Tick the box to confirm you may register this organisation."),
});

export const organisationSchema = z.object({
  name: z.string().trim().min(1, "Enter the legal name, exactly as on your registration certificate.")
    .max(maxNameLength, `Use ${maxNameLength} characters or fewer.`),
  slug: z.string().trim().toLowerCase()
    .regex(slugPattern, "Use lowercase letters, digits and hyphens, like kitabi-washing.")
    .refine((s) => !reservedSlugs.includes(s), "That address is reserved. Choose another."),
  kind: z.enum(kinds.map((k) => k.value), "Choose what your organisation does."),
});

export const registrationSchema = accountSchema.extend(organisationSchema.shape);
export type Registration = z.infer<typeof registrationSchema>;

/** "Kitabi Washing Station Ltd." → "kitabi-washing-station-ltd": a first guess at the address. */
export function suggestSlug(name: string): string {
  return name.toLowerCase().normalize("NFKD").replace(/\p{M}/gu, "").replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "").slice(0, 63).replace(/-+$/, "");
}

// Documents (backend/internal/upload): PDF, PNG or JPEG, 10 MB at most.
export const documentTypes = [
  { value: "registration", label: "Certificate of registration", required: true },
  { value: "licence", label: "Operating licence", required: true },
  { value: "other", label: "Other supporting document", required: false },
] as const;

export const maxUploadBytes = 10 * 1024 * 1024;
const allowedTypes = ["application/pdf", "image/png", "image/jpeg"];

/** The problem with a file, said the way a field error is, or undefined if it can be uploaded. */
export function fileProblem(file: { size: number; type: string }): string | undefined {
  if (!allowedTypes.includes(file.type)) return "This file type can't be used. Choose a PDF, PNG or JPEG.";
  if (file.size > maxUploadBytes) return "This file is larger than 10 MB. Choose a smaller file or a lower-quality scan.";
  return undefined;
}
