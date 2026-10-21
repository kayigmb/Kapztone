"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { CircleCheck } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";

import { Notice } from "@/components/chaintrack/notice";
import { TextField } from "@/components/chaintrack/text-field";
import { Button, buttonVariants } from "@/components/ui/button";
import { orgAddress } from "@/lib/address";
import { ApiError, endpoints, problemOf } from "@/lib/api";
import { copy, routes } from "@/lib/consts";
import { liveValidation } from "@/lib/form";
import { acceptSchema, type AcceptForm as Values } from "@/lib/schemas/profile";

const t = copy.invite;

export function AcceptForm({ token }: { token: string }) {
  const [problem, setProblem] = useState<string>();
  const [joined, setJoined] = useState<string>(); // the organisation's address (slug)
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<Values>({
    ...liveValidation,
    resolver: zodResolver(acceptSchema),
    defaultValues: { name: "", password: "" },
  });
  const accept = useMutation({
    mutationFn: (v: Values) => endpoints.acceptInvite(token, v),
    onMutate: () => setProblem(undefined),
    meta: { inPlace: true },
    onSuccess: (r) => setJoined(r.data.organisation),
    onError: (err) => {
      const status = err instanceof ApiError ? err.status : 0;
      if (status === 403) return setError("password", { message: err.message });
      // 400: the name or password is the problem; the API says which.
      if (status === 400 && err instanceof ApiError)
        return setError(/name/i.test(err.message) ? "name" : "password", {
          message: err.message,
        });
      setProblem(problemOf(err));
    },
  });

  if (joined) {
    // The invitation was made at the organisation's own address; sign in there.
    const signIn = orgAddress(joined, window.location, joined, routes.signin);
    return (
      <div className="grid gap-4">
        <p className="flex items-center gap-2 text-title text-verified">
          <CircleCheck className="size-6" aria-hidden />
          {t.done}
        </p>
        <a href={signIn} className={buttonVariants({ block: true })}>
          {t.signInAt}
        </a>
      </div>
    );
  }

  return (
    <form
      onSubmit={handleSubmit((v) => accept.mutate(v))}
      noValidate
      className="grid gap-4"
    >
      {problem && <Notice kind="blocking">{problem}</Notice>}
      <TextField
        label={t.name}
        help={t.nameHelp}
        autoComplete="name"
        error={errors.name?.message}
        {...register("name")}
      />
      <TextField
        label={t.password}
        help={t.passwordHelp}
        type="password"
        autoComplete="new-password"
        error={errors.password?.message}
        {...register("password")}
      />
      <Button
        type="submit"
        block
        loading={accept.isPending}
        loadingText={t.submitting}
      >
        {t.submit}
      </Button>
    </form>
  );
}
