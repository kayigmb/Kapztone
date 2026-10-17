"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { ArrowLeft, CircleCheck } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";

import { ChoiceCards } from "@/components/chaintrack/choice-cards";
import { Notice } from "@/components/chaintrack/notice";
import { Stepper } from "@/components/chaintrack/stepper";
import { TextField } from "@/components/chaintrack/text-field";
import { Button } from "@/components/ui/button";
import {
  accountSchema, kinds, onboardingSteps, passwordLength, registrationSchema, suggestSlug, type Registration,
} from "@/lib/schemas/org";

const accountFields = Object.keys(accountSchema.shape) as (keyof Registration)[];

const problems = {
  tooMany: "Too many sign-ups from this connection. Wait an hour, then try again.",
  unreachable: "We couldn't create the account just now. Nothing was saved. Try again in a moment.",
};

export function SignUpForm() {
  const [step, setStep] = useState(0);
  const [problem, setProblem] = useState<string>();
  const [created, setCreated] = useState<{ slug: string; email: string }>();
  const { register, handleSubmit, trigger, setError, setValue, getFieldState, formState: { errors, isSubmitting } } =
    useForm<Registration>({
      resolver: zodResolver(registrationSchema),
      defaultValues: { user_name: "", email: "", password: "", name: "", slug: "" },
    });

  async function next() {
    if (await trigger(accountFields, { shouldFocus: true })) setStep(1);
  }

  async function create(values: Registration) {
    setProblem(undefined);
    let res: Response;
    try {
      res = await fetch("/api/orgs", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify(values),
      });
    } catch {
      return setProblem(problems.unreachable);
    }
    if (res.status === 201) return setCreated({ slug: values.slug, email: values.email });
    const { error = "" } = await res.json().catch(() => ({}));
    // The two things only the API can know: whether the address or the email is already in use.
    if (res.status === 409 && error.includes("address")) {
      return setError("slug", { message: "Another organisation already uses this address. Choose a different one." }, { shouldFocus: true });
    }
    if (res.status === 409) {
      setStep(0);
      return setError("email", { message: "An account with this email already exists. Sign in instead, or use another email." });
    }
    setProblem(res.status === 429 ? problems.tooMany : problems.unreachable);
  }

  if (created) {
    // ponytail: the organisation's address is built from the address this page is on, which is right when
    // sign-up is opened on the main site. There is no domain setting to build it from otherwise.
    const { protocol, host } = window.location;
    const site = `${created.slug}.${host}`;
    return (
      <div className="grid justify-items-center gap-3 text-center">
        <span className="grid size-14 place-items-center rounded-full bg-verified text-on-verified">
          <CircleCheck className="size-8" aria-hidden />
        </span>
        <h2 className="text-section">Account created</h2>
        <p className="text-ink-muted">
          Your organisation has its own address. Sign in there to upload your documents.
        </p>
        <p className="code break-all">{site}</p>
        <Button block render={<a href={`${protocol}//${site}/signin?email=${encodeURIComponent(created.email)}`} />}>
          Sign in and upload documents
        </Button>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit(create)} noValidate className="grid gap-4">
      <Stepper steps={onboardingSteps} current={step} error={Object.keys(errors).length > 0} />
      {problem && (
        <Notice kind="blocking">
          <strong>Your account was not created.</strong> {problem}
        </Notice>
      )}

      {/* Both steps stay mounted so the form keeps every value; only one is shown. */}
      <div className={step === 0 ? "grid gap-4" : "hidden"}>
        <TextField label="Full name" autoComplete="name" error={errors.user_name?.message} {...register("user_name")} />
        <TextField label="Work email" type="email" inputMode="email" autoComplete="email"
          error={errors.email?.message} {...register("email")} />
        <TextField label="Password" type="password" autoComplete="new-password"
          help={`At least ${passwordLength.min} characters.`} error={errors.password?.message} {...register("password")} />
        <div className="grid gap-1.5">
          <label className="flex min-h-tap items-start gap-3 text-sm">
            <input type="checkbox" className="mt-0.5 size-[22px] shrink-0 accent-primary" {...register("consent")} />
            I am authorised to register this organisation, and I agree that its documents will be checked.
          </label>
          {errors.consent && <p className="text-small font-semibold text-danger">{errors.consent.message}</p>}
        </div>
        <Button type="button" block onClick={next}>Continue</Button>
      </div>

      <div className={step === 1 ? "grid gap-4" : "hidden"}>
        <TextField label="Legal name" help="Exactly as on your registration certificate." autoComplete="organization"
          error={errors.name?.message}
          {...register("name", {
            // Suggest the address from the name until the person edits the address themselves.
            onChange: (e) => !getFieldState("slug").isDirty && setValue("slug", suggestSlug(e.target.value)),
          })} />
        <TextField label="Organisation address" mono autoCapitalize="none" spellCheck={false}
          help="Your team signs in at this name, in front of this site's address. Lowercase letters, digits and hyphens."
          error={errors.slug?.message} {...register("slug")} />
        <ChoiceCards legend="What does your organisation do?" options={kinds} error={errors.kind?.message} {...register("kind")} />
        <div className="grid grid-cols-[auto_1fr] gap-2">
          <Button type="button" variant="secondary" onClick={() => setStep(0)}>
            <ArrowLeft aria-hidden />
            Back
          </Button>
          <Button type="submit" disabled={isSubmitting}>{isSubmitting ? "Creating account…" : "Create account"}</Button>
        </div>
      </div>
    </form>
  );
}
