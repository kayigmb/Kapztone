"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { CircleCheck, Circle, TriangleAlert } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { toast } from "sonner";

import { AuthResult } from "@/components/chaintrack/auth-result";
import { TextField } from "@/components/chaintrack/text-field";
import { Button, buttonVariants } from "@/components/ui/button";
import { ApiError, endpoints, problemOf } from "@/lib/api";
import { copy, limits, routes } from "@/lib/consts";
import { resetSchema, type ResetForm as Form } from "@/lib/schemas/auth";
import { cn } from "@/lib/utils";

const t = copy.reset;

export function ResetForm({ token }: { token: string }) {
  const [outcome, setOutcome] = useState<"done" | "expired">();
  const { register, handleSubmit, control, setError, formState: { errors } } = useForm<Form>({
    resolver: zodResolver(resetSchema),
    defaultValues: { new: "", confirm: "" },
  });
  const [pw, confirm] = useWatch({ control, name: ["new", "confirm"] });
  const rules = [
    { ok: pw.length >= limits.password.min, label: t.rules.length },
    { ok: pw.length > 0 && pw === confirm, label: t.rules.match },
  ];

  const save = useMutation({
    mutationFn: (v: Form) => endpoints.resetPassword(token, v.new),
    onSuccess: () => setOutcome("done"),
    onError: (err) => {
      if (err instanceof ApiError && err.status === 404) return setOutcome("expired");
      // 400: too short or used recently. The API's message names which.
      if (err instanceof ApiError && err.status === 400) return setError("new", { message: err.message || t.reused });
      toast.error(problemOf(err));
    },
  });

  if (outcome === "done") {
    return (
      <AuthResult icon={CircleCheck} tone="verified" title={t.doneTitle}
        actions={<Link href={routes.signin} className={buttonVariants({ size: "scan", block: true })}>{t.toSignIn}</Link>}>
        <p>{t.done}</p>
      </AuthResult>
    );
  }
  if (outcome === "expired") {
    return (
      <AuthResult icon={TriangleAlert} tone="warning" title={t.expiredTitle} actions={<>
        <Link href={routes.forgot} className={buttonVariants({ size: "scan", block: true })}>{t.newLink}</Link>
        <Link href={routes.signin} className={buttonVariants({ variant: "secondary", block: true })}>{copy.forgot.back}</Link>
      </>}>
        <p>{t.expired}</p>
      </AuthResult>
    );
  }

  return (
    <main className="mx-auto grid w-full max-w-[400px] gap-6">
      <h1 className="text-page">{t.title}</h1>
      <form onSubmit={handleSubmit((v) => save.mutate(v))} noValidate className="grid gap-4">
        <TextField label={t.new} type="password" autoComplete="new-password" error={errors.new?.message} {...register("new")} />
        <TextField label={t.confirm} type="password" autoComplete="new-password" error={errors.confirm?.message} {...register("confirm")} />
        <ul className="grid gap-1.5" aria-live="polite">
          {rules.map((r) => (
            <li key={r.label} className={cn("flex items-center gap-2 text-sm", r.ok ? "text-verified-ink" : "text-ink-muted")}>
              {r.ok ? <CircleCheck className="size-4 shrink-0" aria-hidden /> : <Circle className="size-4 shrink-0" aria-hidden />}
              {r.label}
            </li>
          ))}
        </ul>
        <p className="text-small text-ink-muted">{t.signOutNote}</p>
        <Button type="submit" block size="scan" disabled={save.isPending}>{save.isPending ? t.submitting : t.submit}</Button>
      </form>
    </main>
  );
}
