"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { ArrowLeft, Mail } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { AuthResult } from "@/components/chaintrack/auth-result";
import { TextField } from "@/components/chaintrack/text-field";
import { Button } from "@/components/ui/button";
import { endpoints, problemOf } from "@/lib/api";
import { copy, routes } from "@/lib/consts";
import { forgotSchema, type ForgotForm as Form } from "@/lib/schemas/auth";

const t = copy.forgot;

export function ForgotForm() {
  // The answer is the same for every email, so the page never says whether an account exists.
  const [sentTo, setSentTo] = useState<string>();
  const { register, handleSubmit, reset, formState: { errors } } = useForm<Form>({
    resolver: zodResolver(forgotSchema),
    defaultValues: { email: "" },
  });
  const send = useMutation({
    mutationFn: (v: Form) => endpoints.forgotPassword(v.email),
    onSuccess: (_, v) => setSentTo(v.email),
    onError: (err) => toast.error(problemOf(err)),
  });

  if (sentTo) {
    return (
      <AuthResult icon={Mail} title={t.sentTitle} actions={
        <Button variant="secondary" block onClick={() => { reset(); setSentTo(undefined); }}>{t.again}</Button>
      }>
        <p>{t.sent(sentTo)}</p>
        <div>
          <p className="font-semibold text-ink">{t.nothingTitle}</p>
          <ul className="mt-1 list-disc pl-5">{t.nothing.map((line) => <li key={line}>{line}</li>)}</ul>
        </div>
      </AuthResult>
    );
  }

  return (
    <main className="mx-auto grid w-full max-w-[400px] gap-6">
      <Link href={routes.signin} className="inline-flex min-h-tap w-fit items-center gap-1.5 text-sm font-semibold text-ink-muted hover:text-ink">
        <ArrowLeft className="size-4" aria-hidden />
        {t.back}
      </Link>
      <header className="grid gap-2">
        <h1 className="text-page">{t.title}</h1>
        <p className="text-sm text-ink-muted">{t.sub}</p>
      </header>
      <form onSubmit={handleSubmit((v) => send.mutate(v))} noValidate className="grid gap-4">
        <TextField label={t.email} type="email" autoComplete="username" inputMode="email"
          error={errors.email?.message} {...register("email")} />
        <Button type="submit" block size="scan" disabled={send.isPending}>{send.isPending ? t.submitting : t.submit}</Button>
      </form>
      <p className="text-small text-ink-muted">{t.note}</p>
    </main>
  );
}
