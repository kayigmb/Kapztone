import { z } from "zod";

import { copy, limits } from "../consts.ts";

const v = copy.validation;

export function isStrongPassword(password: string): boolean {
  return (
    /\p{Ll}/u.test(password) &&
    /\p{Lu}/u.test(password) &&
    /\p{N}/u.test(password) &&
    /[\p{P}\p{S}]/u.test(password)
  );
}

export function passwordLengthProblem(password: string): string | undefined {
  const bytes = new TextEncoder().encode(password).length;
  if (bytes < limits.password.min) return v.passwordShort(limits.password.min);
  if (bytes > limits.password.max) return v.tooLong(limits.password.max);
  return undefined;
}

export const passwordLengthSchema = z.string().superRefine((password, ctx) => {
  const problem = passwordLengthProblem(password);
  if (problem) ctx.addIssue({ code: "custom", message: problem });
});

export const strongPasswordSchema = passwordLengthSchema.superRefine(
  (password, ctx) => {
    if (!passwordLengthProblem(password) && !isStrongPassword(password)) {
      ctx.addIssue({ code: "custom", message: v.passwordComplexity });
    }
  },
);
