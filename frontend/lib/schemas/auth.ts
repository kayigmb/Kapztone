import { z } from "zod";

import { copy, limits } from "../consts.ts";

const v = copy.validation;

// What the sign-in form accepts. The API (POST /auth/login) only asks that
// the email and password are right, so this checks shape, not strength; each
// message says what to enter.
export const signInSchema = z.object({
  email: z.string().trim().toLowerCase().pipe(z.email(v.email)),
  password: z.string().min(1, v.passwordEmpty),
  // Asked for only when the account has two-factor sign-in on.
  code: z.string().trim().regex(limits.codePattern, v.code),
  // "Keep me signed in": the API then issues the longer sign-in.
  remember: z.boolean().optional(),
  // After a good 2FA code: skip the code on this device for 30 days.
  trust_device: z.boolean().optional(),
});

export const forgotSchema = z.object({
  email: z.string().trim().toLowerCase().pipe(z.email(v.email)),
});

/** A new password from a reset link: the API's length rule, typed twice. */
export const resetSchema = z
  .object({
    new: z
      .string()
      .min(limits.password.min, v.passwordShort(limits.password.min))
      .max(limits.password.max, v.tooLong(limits.password.max)),
    confirm: z.string(),
  })
  .refine((p) => p.new === p.confirm, {
    path: ["confirm"],
    message: v.passwordsDiffer,
  });

export type SignIn = z.infer<typeof signInSchema>;
export type ForgotForm = z.infer<typeof forgotSchema>;
export type ResetForm = z.infer<typeof resetSchema>;
