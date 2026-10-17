"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { ArrowLeft, CircleCheck } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";

import { ChoiceCards } from "@/components/chaintrack/choice-cards";
import { Notice } from "@/components/chaintrack/notice";
import { Stepper } from "@/components/chaintrack/stepper";
import { TextField } from "@/components/chaintrack/text-field";
import { Button } from "@/components/ui/button";
import { ApiError, endpoints, useMeGuard } from "@/lib/api";
import { copy, limits, orgKinds, routes } from "@/lib/consts";
import { accountSchema, registrationSchema, suggestSlug, type Registration } from "@/lib/schemas/org";

const t = copy.signup;
const accountFields = Object.keys(accountSchema.shape) as (keyof Registration)[];

export function SignUpForm() {
  useMeGuard((me) => (me ? routes.home : undefined)); // already signed in here
  const [step, setStep] = useState(0);
  const [problem, setProblem] = useState<string>();
  const { register, handleSubmit, trigger, setError, setValue, getFieldState, formState: { errors } } =
    useForm<Registration>({
      resolver: zodResolver(registrationSchema),
      defaultValues: { user_name: "", email: "", password: "", name: "", slug: "" },
    });

  async function next() {
    if (await trigger(accountFields, { shouldFocus: true })) setStep(1);
  }

  const create = useMutation({
    mutationFn: endpoints.register,
    onMutate: () => setProblem(undefined),
    onError: (err) => {
      const status = err instanceof ApiError ? err.status : 0;
      // The two things only the API can know: whether the address or the email is already in use.
      if (status === 409 && err.message.includes("address")) {
        return setError("slug", { message: t.slugTaken }, { shouldFocus: true });
      }
      if (status === 409) {
        setStep(0);
        return setError("email", { message: t.emailTaken });
      }
      setProblem(status === 429 ? t.problems.tooMany : t.problems.unreachable);
    },
  });

  if (create.isSuccess) {
    const { slug, email } = create.variables as Registration;
    // ponytail: the organisation's address is built from the address this page is on, which is right when
    // sign-up is opened on the main site. There is no domain setting to build it from otherwise.
    const { protocol, host } = window.location;
    const site = `${slug}.${host}`;
    return (
      <div className="grid justify-items-center gap-3 text-center">
        <span className="grid size-14 place-items-center rounded-full bg-verified text-on-verified">
          <CircleCheck className="size-8" aria-hidden />
        </span>
        <h2 className="text-section">{t.created}</h2>
        <p className="text-ink-muted">{t.createdSub}</p>
        <p className="code break-all">{site}</p>
        <Button block render={<a href={`${protocol}//${site}${routes.signin}?email=${encodeURIComponent(email)}`} />}>
          {t.goSignIn}
        </Button>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit((v) => create.mutate(v))} noValidate className="grid gap-4">
      <Stepper steps={t.steps} current={step} error={Object.keys(errors).length > 0} />
      {problem && (
        <Notice kind="blocking">
          <strong>{t.notCreated}</strong> {problem}
        </Notice>
      )}

      {/* Both steps stay mounted so the form keeps every value; only one is shown. */}
      <div className={step === 0 ? "grid gap-4" : "hidden"}>
        <TextField label={t.fullName} autoComplete="name" error={errors.user_name?.message} {...register("user_name")} />
        <TextField label={t.email} type="email" inputMode="email" autoComplete="email"
          error={errors.email?.message} {...register("email")} />
        <TextField label={t.password} type="password" autoComplete="new-password"
          help={t.passwordHelp(limits.password.min)} error={errors.password?.message} {...register("password")} />
        <div className="grid gap-1.5">
          <label className="flex min-h-tap items-start gap-3 text-sm">
            <input type="checkbox" className="mt-0.5 size-[22px] shrink-0 accent-primary" {...register("consent")} />
            {t.consent}
          </label>
          {errors.consent && <p className="text-small font-semibold text-danger">{errors.consent.message}</p>}
        </div>
        <Button type="button" block onClick={next}>{t.next}</Button>
      </div>

      <div className={step === 1 ? "grid gap-4" : "hidden"}>
        <TextField label={t.legalName} help={t.legalNameHelp} autoComplete="organization"
          error={errors.name?.message}
          {...register("name", {
            // Suggest the address from the name until the person edits the address themselves.
            onChange: (e) => !getFieldState("slug").isDirty && setValue("slug", suggestSlug(e.target.value)),
          })} />
        <TextField label={t.slug} mono autoCapitalize="none" spellCheck={false}
          help={t.slugHelp} error={errors.slug?.message} {...register("slug")} />
        <ChoiceCards legend={t.kind} options={orgKinds} error={errors.kind?.message} {...register("kind")} />
        <div className="grid grid-cols-[auto_1fr] gap-2">
          <Button type="button" variant="secondary" onClick={() => setStep(0)}>
            <ArrowLeft aria-hidden />
            {t.back}
          </Button>
          <Button type="submit" disabled={create.isPending}>{create.isPending ? t.submitting : t.submit}</Button>
        </div>
      </div>
    </form>
  );
}
