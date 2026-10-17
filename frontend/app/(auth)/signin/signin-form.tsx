"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";

import { Notice } from "@/components/chaintrack/notice";
import { TextField } from "@/components/chaintrack/text-field";
import { Button } from "@/components/ui/button";
import { signInSchema, type SignIn } from "@/lib/schemas/auth";

// What went wrong, and what to do next (content.md → Errors).
const problems = {
  credentials: "Email or password is incorrect. Check both and try again.",
  code: "That code is incorrect or already used. Enter the code your app shows now.",
  tooMany: "Too many attempts. Wait 15 minutes, then try again.",
  unreachable: "We couldn't sign you in just now. Nothing was changed. Try again in a moment.",
};

export function SignInForm({ email }: { email: string }) {
  const router = useRouter();
  const [needsCode, setNeedsCode] = useState(false);
  const [problem, setProblem] = useState<string>();
  const { register, handleSubmit, setFocus, formState: { errors, isSubmitting } } = useForm<SignIn>({
    resolver: zodResolver(signInSchema),
    defaultValues: { email, password: "", code: "" },
  });

  async function signIn(values: SignIn) {
    setProblem(undefined);
    let res: Response;
    try {
      res = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify(values),
      });
    } catch {
      return setProblem(problems.unreachable);
    }
    if (res.ok) {
      // The API has set the session cookie; render the app as the signed-in user.
      router.replace("/");
      return router.refresh();
    }
    const body = await res.json().catch(() => ({}));
    if (res.status === 401 && body.two_factor_required) {
      // First time: just ask for the code. After that, the code given was wrong.
      if (needsCode && values.code) setProblem(problems.code);
      setNeedsCode(true);
      return setTimeout(() => setFocus("code"));
    }
    setProblem(res.status === 401 ? problems.credentials : res.status === 429 ? problems.tooMany : problems.unreachable);
  }

  return (
    <form onSubmit={handleSubmit(signIn)} noValidate className="grid gap-4">
      {problem && (
        <Notice kind="blocking">
          <strong>You are not signed in.</strong> {problem}
        </Notice>
      )}
      <TextField label="Work email" type="email" autoComplete="username" inputMode="email"
        error={errors.email?.message} {...register("email")} />
      <TextField label="Password" type="password" autoComplete="current-password"
        error={errors.password?.message} {...register("password")} />
      {needsCode && (
        <TextField label="Authenticator code" mono inputMode="numeric" autoComplete="one-time-code" maxLength={6}
          help="The 6 digits your authenticator app shows for ChainTrack."
          error={errors.code?.message} {...register("code")} />
      )}
      <Button type="submit" block disabled={isSubmitting}>
        {isSubmitting ? "Signing in…" : "Sign in"}
      </Button>
    </form>
  );
}
