"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";

import { TriangleAlert } from "lucide-react";

import { Notice } from "@/components/chaintrack/notice";
import { TextField } from "@/components/chaintrack/text-field";
import { Button } from "@/components/ui/button";
import { ApiError, endpoints, problemOf, useMeGuard } from "@/lib/api";
import { copy, queryKeys, routes } from "@/lib/consts";
import { liveValidation } from "@/lib/form";
import { signInSchema, type SignIn } from "@/lib/schemas/auth";

const t = copy.signin;

export function SignInForm({ email }: { email: string }) {
  const router = useRouter();
  const client = useQueryClient();
  useMeGuard((me) => (me ? routes.home : undefined)); // already signed in here
  const [needsCode, setNeedsCode] = useState(false);
  const [problem, setProblem] = useState<string>();
  const [unconfirmed, setUnconfirmed] = useState<string>(); // the email a confirmation link just went to
  const [capsLock, setCapsLock] = useState(false);
  const {
    register,
    handleSubmit,
    setFocus,
    formState: { errors },
  } = useForm<SignIn>({
    ...liveValidation,
    resolver: zodResolver(signInSchema),
    defaultValues: {
      email,
      password: "",
      code: "",
      remember: false,
      trust_device: false,
    },
  });

  const signIn = useMutation({
    mutationFn: endpoints.login,
    onMutate: () => {
      setProblem(undefined);
      setUnconfirmed(undefined);
    },
    meta: { inPlace: true },
    onSuccess: async () => {
      // The API has set the session cookie; whoever was cached before is stale.
      await client.invalidateQueries({ queryKey: queryKeys.me });
      router.replace(routes.home);
    },
    onError: (err, values) => {
      if (
        err instanceof ApiError &&
        err.status === 401 &&
        err.twoFactorRequired
      ) {
        // First time: just ask for the code. After that, the code given was wrong.
        if (needsCode && values.code) setProblem(err.message);
        setNeedsCode(true);
        return setTimeout(() => setFocus("code"));
      }
      if (err instanceof ApiError && err.emailUnverified)
        return setUnconfirmed(values.email);
      setProblem(problemOf(err)); // wrong password, attempts left, paused: the API's words
    },
  });

  return (
    <form
      onSubmit={handleSubmit((v) => signIn.mutate(v))}
      noValidate
      className="grid gap-4"
    >
      {problem && (
        <Notice kind="blocking">
          <strong>{t.notSignedIn}</strong> {problem}
        </Notice>
      )}
      {unconfirmed && (
        <Notice kind="warning">
          <strong>{t.confirmTitle}</strong> {t.confirmBody(unconfirmed)}
        </Notice>
      )}
      <TextField
        label={t.email}
        type="email"
        autoComplete="username"
        inputMode="email"
        autoFocus
        error={errors.email?.message}
        {...register("email")}
      />
      <div className="grid gap-1.5">
        <TextField
          label={t.password}
          type="password"
          autoComplete="current-password"
          onKeyUp={(e) => setCapsLock(e.getModifierState("CapsLock"))}
          error={errors.password?.message}
          {...register("password")}
        />
        {capsLock && (
          <p
            className="flex items-center gap-1.5 text-small font-semibold text-warning-ink"
            role="status"
          >
            <TriangleAlert className="size-4 shrink-0" aria-hidden />
            {t.capsLock}
          </p>
        )}
        <Link
          href={routes.forgot}
          className="inline-flex min-h-tap w-fit items-center text-sm font-semibold text-primary hover:underline"
        >
          {t.forgot}
        </Link>
      </div>
      {needsCode && (
        <TextField
          label={t.code}
          mono
          inputMode="numeric"
          autoComplete="one-time-code"
          maxLength={6}
          help={t.codeHelp}
          error={errors.code?.message}
          {...register("code")}
        />
      )}
      {needsCode && (
        <label className="grid min-h-tap w-fit grid-cols-[auto_1fr] items-start gap-x-2.5 gap-y-0.5 text-sm">
          <input
            type="checkbox"
            className="mt-0.5 size-5 shrink-0 accent-primary"
            {...register("trust_device")}
          />
          <span>{t.trust}</span>
          <span className="col-start-2 text-small text-ink-muted">
            {t.trustHelp}
          </span>
        </label>
      )}
      <label className="flex min-h-tap w-fit items-center gap-2.5 text-sm text-ink-muted">
        <input
          type="checkbox"
          className="size-5 shrink-0 accent-primary"
          {...register("remember")}
        />
        {t.remember}
      </label>
      <Button
        type="submit"
        block
        loading={signIn.isPending}
        loadingText={t.submitting}
      >
        {t.submit}
      </Button>
    </form>
  );
}
