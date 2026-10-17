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
});

export type SignIn = z.infer<typeof signInSchema>;
