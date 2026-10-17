import { z } from "zod";

// What the sign-in form accepts. The API (POST /auth/login) only asks that
// the email and password are right, so this checks shape, not strength; each
// message says what to enter.
export const signInSchema = z.object({
  email: z.string().trim().toLowerCase().pipe(z.email("Enter your work email, like name@company.com.")),
  password: z.string().min(1, "Enter your password."),
  // Asked for only when the account has two-factor sign-in on.
  code: z.string().trim().regex(/^(\d{6})?$/, "Enter the 6 digits shown in your authenticator app."),
});

export type SignIn = z.infer<typeof signInSchema>;
