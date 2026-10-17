"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";

import { Notice } from "@/components/chaintrack/notice";
import { TextField } from "@/components/chaintrack/text-field";
import { Button } from "@/components/ui/button";
import { ApiError, endpoints, useMeGuard } from "@/lib/api";
import { copy, queryKeys, routes } from "@/lib/consts";
import { signInSchema, type SignIn } from "@/lib/schemas/auth";

const t = copy.signin;

export function SignInForm({ email }: { email: string }) {
  const router = useRouter();
  const client = useQueryClient();
  useMeGuard((me) => (me ? routes.home : undefined)); // already signed in here
  const [needsCode, setNeedsCode] = useState(false);
  const [problem, setProblem] = useState<string>();
  const { register, handleSubmit, setFocus, formState: { errors } } = useForm<SignIn>({
    resolver: zodResolver(signInSchema),
    defaultValues: { email, password: "", code: "" },
  });

  const signIn = useMutation({
    mutationFn: endpoints.login,
    onMutate: () => setProblem(undefined),
    onSuccess: async () => {
      // The API has set the session cookie; whoever was cached before is stale.
      await client.invalidateQueries({ queryKey: queryKeys.me });
      router.replace(routes.home);
    },
    onError: (err, values) => {
      if (err instanceof ApiError && err.status === 401 && err.twoFactorRequired) {
        // First time: just ask for the code. After that, the code given was wrong.
        if (needsCode && values.code) setProblem(t.problems.code);
        setNeedsCode(true);
        return setTimeout(() => setFocus("code"));
      }
      const status = err instanceof ApiError ? err.status : 0;
      setProblem(status === 401 ? t.problems.credentials : status === 429 ? t.problems.tooMany : t.problems.unreachable);
    },
  });

  return (
    <form onSubmit={handleSubmit((v) => signIn.mutate(v))} noValidate className="grid gap-4">
      {problem && (
        <Notice kind="blocking">
          <strong>{t.notSignedIn}</strong> {problem}
        </Notice>
      )}
      <TextField label={t.email} type="email" autoComplete="username" inputMode="email"
        error={errors.email?.message} {...register("email")} />
      <TextField label={t.password} type="password" autoComplete="current-password"
        error={errors.password?.message} {...register("password")} />
      {needsCode && (
        <TextField label={t.code} mono inputMode="numeric" autoComplete="one-time-code" maxLength={6}
          help={t.codeHelp} error={errors.code?.message} {...register("code")} />
      )}
      <Button type="submit" block disabled={signIn.isPending}>
        {signIn.isPending ? t.submitting : t.submit}
      </Button>
    </form>
  );
}
