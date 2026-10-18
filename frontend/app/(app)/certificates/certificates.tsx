"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { FileText, Lock, OctagonX, Stamp } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { CertificateTile } from "@/components/chaintrack/certificate-tile";
import { Loading, PageState } from "@/components/chaintrack/page-state";
import { SigningPanel } from "@/components/chaintrack/signing";
import { Button, buttonVariants } from "@/components/ui/button";
import { Dialog, DialogClose, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { endpoints, problemOf, useMe, type Certificate } from "@/lib/api";
import { certificateStatus, copy, permissions, queryKeys, routes, subjectTypes } from "@/lib/consts";
import { revokeFormSchema, type RevokeForm } from "@/lib/schemas/certificate";
import { cn } from "@/lib/utils";
import { formatDate, humanize, type CertificateState, type View } from "@/lib/verification";

const t = copy.certificates;

// Copy that lib/consts.ts doesn't have yet.
const local = {
  all: "All",
  notAllowedTitle: "You can't open certificates.",
  notAllowedBody: "Ask an admin of your organisation for the issue-certificates permission.",
  emptyTitle: "No certificates issued yet.",
  revokeSign: "Sign and revoke certificate",
  cancel: "Cancel",
};

const ALL = "All";
const filters = [
  { value: ALL, label: local.all },
  { value: certificateStatus.valid, label: t.status.valid },
  { value: certificateStatus.expired, label: t.status.expired },
  { value: certificateStatus.revoked, label: t.status.revoked },
] as const;
type Filter = (typeof filters)[number]["value"];

/** 30 days, as on the public page: a valid certificate inside it reads "Expiring soon". */
function certificateState(c: Certificate): CertificateState {
  if (c.status !== "valid") return c.status;
  return (new Date(c.valid_to).getTime() - Date.now()) / 86_400_000 <= 30 ? "expiring" : "valid";
}

const subjectHref = (c: Certificate) =>
  c.subject_type === subjectTypes.passport ? routes.passportDetail(c.subject_id) : routes.facility(c.subject_id);

export function Certificates() {
  const me = useMe().data!; // the layout only renders once it is loaded
  const [filter, setFilter] = useState<Filter>(ALL);
  const [revoking, setRevoking] = useState<Certificate | null>(null);

  if (!me.organisation || !me.permissions.includes(permissions.certificateIssue)) {
    return <PageState icon={Lock} title={local.notAllowedTitle}>{local.notAllowedBody}</PageState>;
  }

  return (
    <main className="grid content-start gap-6 p-4 lg:p-8">
      <header className="flex flex-wrap items-end justify-between gap-3">
        <div className="grid gap-1">
          <h1 className="text-[22px] font-semibold lg:text-page">{t.title}</h1>
          <p className="text-sm text-ink-muted">{t.sub}</p>
        </div>
        <Link href={routes.certificateNew} className={buttonVariants()}>
          <Stamp aria-hidden />
          {t.issue}
        </Link>
      </header>

      <div className="inline-flex flex-wrap gap-0.5 rounded-md border border-border bg-surface-sunken p-[3px]" role="group" aria-label={t.headers.status}>
        {filters.map((f) => (
          <button
            key={f.value}
            type="button"
            onClick={() => setFilter(f.value)}
            aria-pressed={filter === f.value}
            className={cn("min-h-tap rounded-sm px-3 text-sm",
              filter === f.value ? "bg-surface font-semibold text-ink shadow-card" : "text-ink-muted hover:text-ink")}
          >
            {f.label}
          </button>
        ))}
      </div>

      <IssuedBy
        filter={filter} issuerId={me.organisation.id} issuerName={me.organisation.name}
        canRevoke={me.permissions.includes(permissions.certificateRevoke)}
        revoking={revoking} onRevoke={setRevoking}
      />
    </main>
  );
}

function IssuedBy({ filter, issuerId, issuerName, canRevoke, revoking, onRevoke }: {
  filter: Filter; issuerId: string; issuerName: string; canRevoke: boolean;
  revoking: Certificate | null; onRevoke: (c: Certificate | null) => void;
}) {
  const query = useQuery({ queryKey: queryKeys.certificates, queryFn: () => endpoints.attestationsList() });
  const directory = useQuery({ queryKey: queryKeys.directory, queryFn: endpoints.directory, enabled: !!query.data?.length });
  const facilities = useQuery({
    queryKey: queryKeys.facilities, queryFn: endpoints.facilitiesList,
    enabled: !!query.data?.some((c) => c.subject_type === subjectTypes.facility),
  });

  if (query.isPending) return <Loading className="py-4" />;
  if (query.isError) {
    return (
      <PageState icon={FileText} tone="danger" title={copy.errors.loadTitle} retry={() => query.refetch()}>
        {problemOf(query.error)}
      </PageState>
    );
  }

  // Only what this organisation signed; anything else on the list isn't theirs to show here.
  const mine = query.data.filter((c) => c.issuer_org_id === issuerId);
  const rows = filter === ALL ? mine : mine.filter((c) => c.status === filter);
  const orgName = (id: string) => directory.data?.get(id) ?? (id === issuerId ? issuerName : id);
  const facilityName = (id: string) => facilities.data?.find((f) => f.id === id)?.name ?? id;
  const subject = (c: Certificate) =>
    c.subject_type === subjectTypes.passport ? t.attachedLot(c.subject_id) : facilityName(c.subject_id);

  if (rows.length === 0) {
    return (
      <div className="grid justify-items-center gap-2 py-8 text-center">
        <span className="grid size-14 place-items-center rounded-full bg-surface-sunken text-ink-muted">
          <FileText className="size-7" aria-hidden />
        </span>
        <h2 className="text-title">{filter === ALL ? local.emptyTitle : t.emptyType}</h2>
        <p className="max-w-[48ch] text-ink-muted">{t.empty}</p>
      </div>
    );
  }

  return (
    <>
      <ul className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {rows.map((c) => (
          <Row
            key={c.id} certificate={c} issuer={orgName(c.issuer_org_id)} subject={subject(c)}
            canRevoke={canRevoke} onRevoke={onRevoke}
          />
        ))}
      </ul>
      {revoking && <RevokeDialog certificate={revoking} onClose={() => onRevoke(null)} />}
    </>
  );
}

function Row({ certificate: c, issuer, subject, canRevoke, onRevoke }: {
  certificate: Certificate; issuer: string; subject: string;
  canRevoke: boolean; onRevoke: (c: Certificate) => void;
}) {
  const state = certificateState(c);
  const view: View["certificates"][number] = {
    id: c.id,
    type: humanize(c.type),
    issuer,
    state,
    validFrom: formatDate(c.valid_from),
    validTo: formatDate(c.valid_to),
    note: state === "revoked"
      ? copy.verification.certificateNote.revoked(c.revoke_reason ?? null)
      : state === "expired"
        ? copy.verification.certificateNote.expired(formatDate(c.valid_to))
        : undefined,
  };
  return (
    <li className="grid gap-2">
      <Link href={subjectHref(c)} className="rounded-md focus-visible:ring-3 focus-visible:ring-ring/50">
        <CertificateTile certificate={view} />
      </Link>
      <div className="flex flex-wrap items-center justify-between gap-2 px-1">
        <span className="text-small text-ink-muted">{subject}</span>
        {canRevoke && c.status !== "revoked" && (
          <Button variant="danger-outline" onClick={() => onRevoke(c)}>
            <OctagonX aria-hidden />
            {t.revoke}
          </Button>
        )}
      </div>
    </li>
  );
}

/** Revoking is signed: a reason the holder can read, then the PIN. */
function RevokeDialog({ certificate: c, onClose }: { certificate: Certificate; onClose: () => void }) {
  const client = useQueryClient();
  const { register, handleSubmit, formState: { errors } } = useForm<RevokeForm>({
    resolver: zodResolver(revokeFormSchema),
    defaultValues: { reason: "" },
  });
  const revoke = useMutation({
    mutationFn: ({ reason, pin }: { reason: string; pin: string }) => endpoints.revokeAttestation(c.id, reason, pin),
    onSuccess: () => {
      toast.success(t.revoked);
      client.invalidateQueries({ queryKey: queryKeys.certificates });
      onClose();
    },
    onError: (err) => toast.error(problemOf(err)),
  });
  const noteId = "revoke-reason";

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent showCloseButton={false} className="max-w-[min(560px,100%)] gap-4 rounded-lg border border-border bg-surface p-6 shadow-overlay">
        <DialogTitle className="text-[19px] font-semibold">{t.revokeTitle}</DialogTitle>
        <p className="text-sm text-ink-muted">{t.revokeBody}</p>
        <div className="grid gap-1.5">
          <label htmlFor={noteId} className="text-sm font-semibold">{t.revokeReason}</label>
          <Textarea id={noteId} rows={3} aria-invalid={errors.reason ? true : undefined} aria-describedby={`${noteId}-note`}
            {...register("reason")} />
          {errors.reason ? (
            <p id={`${noteId}-note`} className="flex gap-1.5 text-small font-semibold text-danger">
              <OctagonX className="mt-px size-4 shrink-0" aria-hidden />
              {t.revokeReasonEmpty}
            </p>
          ) : (
            <p id={`${noteId}-note`} className="text-small text-ink-subtle">{t.revokeReasonHelp}</p>
          )}
        </div>
        <SigningPanel
          facts={[
            { label: t.headers.type, value: humanize(c.type) },
            { label: t.headers.attached, value: subjectLabel(c), mono: c.subject_type === subjectTypes.passport, wide: true },
          ]}
          meaning={t.revokeBody}
          button={local.revokeSign}
          busy={revoke.isPending}
          error={revoke.isError ? problemOf(revoke.error) : undefined}
          onSign={(pin) => { void handleSubmit((v) => revoke.mutate({ reason: v.reason, pin }))(); }}
        />
        <div>
          <DialogClose render={<Button variant="secondary" />}>{local.cancel}</DialogClose>
        </div>
      </DialogContent>
    </Dialog>
  );
}

// The dialog has no facility list of its own; the subject reads as the lot or the raw id.
function subjectLabel(c: Certificate): string {
  return c.subject_type === subjectTypes.passport ? t.attachedLot(c.subject_id) : t.attachedFacility(c.subject_id);
}
