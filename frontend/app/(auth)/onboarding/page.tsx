import { Check, Clock } from "lucide-react";
import { redirect } from "next/navigation";

import { SignOutButton } from "@/app/sign-out-button";
import { Notice } from "@/components/chaintrack/notice";
import { StatusChip } from "@/components/chaintrack/status";
import { Stepper } from "@/components/chaintrack/stepper";
import { getMe, getOnboarding } from "@/lib/api";
import { documentTypes, onboardingSteps } from "@/lib/schemas/org";
import { cn } from "@/lib/utils";
import { formatDateTime } from "@/lib/verification";
import { DocumentsUpload } from "./documents-upload";

export const metadata = { title: "Verification · ChainTrack" };

// Steps 3 and 4 of onboarding (screens.md §03): upload documents, then wait for the review.
export default async function Page() {
  const me = await getMe();
  if (!me) redirect("/signin");
  if (!me.organisation || me.organisation.status === "verified") redirect("/");

  const { organisation: org, documents } = await getOnboarding(me.organisation.id);
  const missing = documentTypes.filter((t) => t.required && !documents.some((d) => d.type === t.value));
  const complete = missing.length === 0;
  const rejected = org.status === "rejected";
  const latest = documents.at(-1);

  // What happens, in order; the first unfinished line is where things stand.
  const progress = [
    { name: "Account created", done: true },
    { name: "Documents received", done: complete, note: complete ? undefined : `Still needed: ${missing.map((t) => t.label.toLowerCase()).join(", ")}` },
    { name: "Registry check", done: false, note: complete ? "We are checking your documents against the national registry." : undefined },
    { name: "Approval", done: false, note: "This page shows the outcome as soon as the review is done." },
  ];
  const now = progress.findIndex((p) => !p.done);

  return (
    <main className="mx-auto grid w-full max-w-xl gap-6 px-4 py-8">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div className="grid gap-1">
          <p className="text-title">ChainTrack</p>
          <h1 className="text-page">{complete && !rejected ? "Pending verification" : "Upload your documents"}</h1>
          <p className="text-ink-muted">
            {complete && !rejected
              ? `We're checking ${org.name}. Most organisations are verified within 2 working days.`
              : "Clear photos or PDFs. We check them against the national registry."}
          </p>
        </div>
        <SignOutButton />
      </header>

      <Stepper steps={onboardingSteps} current={complete ? 3 : 2} error={rejected} />

      {rejected ? (
        <Notice kind="blocking">
          <strong>{org.name} was not verified.</strong> {org.review_reason ?? "The reviewer gave no reason."} Upload
          a corrected document below and it goes back to the reviewer.
        </Notice>
      ) : (
        <Notice kind="warning">
          <strong>You can&apos;t issue or receive passports until verification is done.</strong>
        </Notice>
      )}

      {complete && !rejected && (
        <section className="grid gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
          <div className="flex flex-wrap items-center gap-3">
            <span className="grid size-14 place-items-center rounded-full bg-pending-soft text-pending">
              <Clock className="size-7" aria-hidden />
            </span>
            {latest && <StatusChip tone="pending" icon={Clock}>Submitted {formatDateTime(latest.created_at)}</StatusChip>}
          </div>
          <ol className="grid gap-3">
            {progress.map((p, i) => (
              <li key={p.name} className="grid grid-cols-[24px_1fr] gap-x-3">
                <span className={cn("grid size-6 place-items-center rounded-full border-2",
                  p.done ? "border-verified bg-verified text-on-verified" : i === now ? "border-primary" : "border-border-strong")}>
                  {p.done && <Check className="size-4" aria-hidden />}
                </span>
                <div>
                  <p className={cn(i === now ? "font-semibold" : !p.done && "text-ink-muted")}>{p.name}</p>
                  {p.note && i >= now && <p className="text-small text-ink-muted">{p.note}</p>}
                </div>
              </li>
            ))}
          </ol>
        </section>
      )}

      <section className="rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
        <DocumentsUpload orgId={org.id} documents={documents} />
        <p className="mt-6 text-small text-ink-muted">
          Facility licences are added later, when you register each farm plot or site.
        </p>
      </section>
    </main>
  );
}
