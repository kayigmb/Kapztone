// Forms on the profile, users and invitation pages. Each mirrors the API's
// own check for the same input; the API stays the authority.
import { z } from "zod";

import { copy, limits, memberRoles } from "../consts.ts";

const v = copy.validation;

const fullName = z
  .string()
  .trim()
  .min(1, v.fullName)
  .max(limits.name, v.tooLong(limits.name));
const newPassword = z
  .string()
  .min(limits.password.min, v.passwordShort(limits.password.min))
  .max(limits.password.max, v.tooLong(limits.password.max));

export const nameSchema = z.object({ name: fullName });

export const passwordChangeSchema = z
  .object({
    current: z.string().min(1, v.passwordEmpty),
    new: newPassword,
    confirm: z.string().min(1, v.passwordConfirm),
  })
  .refine((p) => p.new !== p.current, {
    path: ["new"],
    message: v.samePassword,
  })
  .refine((p) => p.new === p.confirm, {
    path: ["confirm"],
    message: v.passwordsDiffer,
  });

export const pinSchema = z.object({
  pin: z
    .string()
    .trim()
    .regex(/^\d{6}$/, v.pin),
  password: z.string().min(1, v.passwordEmpty),
});

/** A current code from the authenticator app. */
export const appCodeSchema = z.object({
  code: z
    .string()
    .trim()
    .regex(/^\d{6}$/, v.code),
});

export const inviteSchema = z.object({
  email: z
    .string()
    .trim()
    .toLowerCase()
    .max(limits.email, v.emailLong)
    .pipe(z.email(v.email)),
  role: z.enum(memberRoles.map((r) => r.value) as [string, ...string[]]),
});

/** Accepting an invitation: a name only for a new account; the password is new or the existing one. */
export const acceptSchema = z.object({
  name: z.string().trim().max(limits.name, v.tooLong(limits.name)),
  password: z
    .string()
    .min(1, v.passwordEmpty)
    .max(limits.password.max, v.tooLong(limits.password.max)),
});

export type NameForm = z.infer<typeof nameSchema>;
export type PasswordChange = z.infer<typeof passwordChangeSchema>;
export type PinForm = z.infer<typeof pinSchema>;
export type AppCode = z.infer<typeof appCodeSchema>;
export type InviteForm = z.infer<typeof inviteSchema>;
export type AcceptForm = z.infer<typeof acceptSchema>;
