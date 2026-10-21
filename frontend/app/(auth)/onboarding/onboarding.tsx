"use client";

import { useQuery } from "@tanstack/react-query";
import { Check, Clock, WifiOff } from "lucide-react";

import { SignOutButton } from "@/components/chaintrack/sign-out-button";
import { Notice } from "@/components/chaintrack/notice";
import { LoadingWithLogo, PageState } from "@/components/chaintrack/page-state";
import { StatusChip } from "@/components/chaintrack/status";
import { Stepper } from "@/components/chaintrack/stepper";
import { endpoints, problemOf, useMeGuard } from "@/lib/api";
import { copy, orgStatus, queryKeys, routes } from "@/lib/consts";
import { missingDocuments } from "@/lib/schemas/org";
import { cn } from "@/lib/utils";
import { formatDateTime } from "@/lib/verification";
import { DocumentsUpload } from "./documents-upload";

const t = copy.onboarding;

// Steps 3 and 4 of onboarding (screens.md §03): upload documents, then wait for the review.
export function Onboarding() {
  const me = useMeGuard((me) =>
    !me
      ? routes.signin
      : !me.organisation || me.organisation.status === orgStatus.verified
        ? routes.home
        : undefined,
  );
  const orgId = me.ready ? me.data!.organisation!.id : "";
  const org = useQuery({
    queryKey: queryKeys.org(orgId),
    queryFn: () => endpoints.org(orgId),
    enabled: !!orgId,
  });
  const docs = useQuery({
    queryKey: queryKeys.orgDocuments(orgId),
    queryFn: () => endpoints.documents(orgId),
    enabled: !!orgId,
  });

  const failed = me.error ?? org.error ?? docs.error;
  if (failed) {
    return (
      <PageState
        icon={WifiOff}
        tone="danger"
        title={copy.errors.loadTitle}
        retry={() => {
          me.refetch();
          org.refetch();
          docs.refetch();
        }}
      >
        {problemOf(failed)}
      </PageState>
    );
  }
  if (!org.data || !docs.data)
    return <LoadingWithLogo />;

  const o = org.data;
  const documents = docs.data;
  const missing = missingDocuments(documents);
  const complete = missing.length === 0;
  const sentBack =
    o.status === orgStatus.rejected || o.status === orgStatus.changesRequested;
  const latest = documents.at(-1);

  // What happens, in order; the first unfinished line is where things stand.
  const progress = [
    { name: t.progress.account, done: true },
    {
      name: t.progress.documents,
      done: complete,
      note: complete
        ? undefined
        : t.progress.stillNeeded(
            missing.map((m) => m.label.toLowerCase()).join(", "),
          ),
    },
    {
      name: t.progress.registry,
      done: false,
      note: complete ? t.progress.registryNote : undefined,
    },
    { name: t.progress.approval, done: false, note: t.progress.approvalNote },
  ];
  const now = progress.findIndex((p) => !p.done);
  const waiting = complete && !sentBack;

  return (
    <main className="mx-auto grid w-full max-w-135 gap-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div className="grid gap-2">
          <span className="eyebrow text-primary">{t.eyebrow}</span>
          <h1 className="text-page">
            {waiting ? t.pendingTitle : t.uploadTitle}
          </h1>
          <p className="text-sm text-ink-muted">
            {waiting ? t.pendingSub(o.name) : t.uploadSub}
          </p>
        </div>
        <SignOutButton />
      </header>

      <Stepper
        steps={copy.signup.steps}
        current={complete ? 3 : 2}
        error={o.status === orgStatus.rejected}
      />

      {o.status === orgStatus.rejected ? (
        <Notice kind="blocking">
          <strong>{t.rejected(o.name)}</strong> {o.review_reason ?? t.noReason}{" "}
          {t.uploadAgain}
        </Notice>
      ) : o.status === orgStatus.changesRequested ? (
        <Notice kind="warning">
          <strong>{t.changesRequested}</strong> {o.review_reason ?? t.noReason}{" "}
          {t.uploadAgain}
        </Notice>
      ) : (
        <Notice kind="warning">
          <strong>{t.notVerifiedYet}</strong>
        </Notice>
      )}

      {waiting && (
        <section className="grid gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
          <div className="flex flex-wrap items-center gap-3">
            <span className="grid size-14 place-items-center rounded-full bg-pending-soft text-pending">
              <Clock className="size-7" aria-hidden />
            </span>
            {latest && (
              <StatusChip tone="pending" icon={Clock}>
                {t.submitted(formatDateTime(latest.created_at))}
              </StatusChip>
            )}
          </div>
          <ol className="grid gap-3">
            {progress.map((p, i) => (
              <li key={p.name} className="grid grid-cols-[auto_1fr] gap-x-3">
                <span
                  className={cn(
                    "grid size-6 place-items-center rounded-full border-2",
                    p.done
                      ? "border-verified bg-verified text-on-verified"
                      : i === now
                        ? "border-primary"
                        : "border-border-strong",
                  )}
                >
                  {p.done && <Check className="size-4" aria-hidden />}
                </span>
                <div>
                  <p
                    className={cn(
                      i === now ? "font-semibold" : !p.done && "text-ink-muted",
                    )}
                  >
                    {p.name}
                  </p>
                  {p.note && i >= now && (
                    <p className="text-small text-ink-muted">{p.note}</p>
                  )}
                </div>
              </li>
            ))}
          </ol>
        </section>
      )}

      <section className="rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
        <DocumentsUpload orgId={o.id} documents={documents} />
        <p className="mt-6 text-small text-ink-muted">{t.licencesLater}</p>
      </section>
    </main>
  );
}
