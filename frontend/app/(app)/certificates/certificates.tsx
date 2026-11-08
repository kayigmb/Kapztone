"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CircleCheck, Clock, FileText, OctagonX, Stamp } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { parseAsStringLiteral } from "nuqs";

import { certificateLook } from "@/components/chaintrack/certificate-tile";
import { FieldNote } from "@/components/chaintrack/field-note";
import { ListBar } from "@/components/chaintrack/list-controls";
import {
  actionsColumn,
  column,
  ListTable,
  RowAction,
  type ListColumn,
} from "@/components/chaintrack/list-table";
import { StatusChip } from "@/components/chaintrack/status";
import { PageHeader } from "@/components/chaintrack/page-header";
import { PageStats } from "@/components/chaintrack/page-stats";
import { Segmented } from "@/components/chaintrack/segmented";
import { SigningPanel } from "@/components/chaintrack/signing";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogTitle,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import {
  endpoints,
  problemOf,
  useMe,
  usePaged,
  type Certificate,
} from "@/lib/api";
import {
  certificateScopes,
  certificateStatus,
  copy,
  listSorts,
  permissions,
  queryKeys,
  routes,
  subjectTypes,
  type CertificateScope,
} from "@/lib/consts";
import { inspects } from "@/lib/dashboard";
import { liveValidation } from "@/lib/form";
import { useListParams, type ListParamsState } from "@/lib/list-params";
import { revokeFormSchema, type RevokeForm } from "@/lib/schemas/certificate";
import {
  formatDate,
  humanize,
  type CertificateState,
} from "@/lib/verification";

const t = copy.certificates;

const ALL = "";
const statusFilters = [
  { value: ALL, label: t.all },
  { value: certificateStatus.valid, label: t.status.valid },
  { value: certificateStatus.expired, label: t.status.expired },
  { value: certificateStatus.revoked, label: t.status.revoked },
] as const;
const statusValues = statusFilters
  .map((f) => f.value)
  .filter(Boolean) as Exclude<(typeof statusFilters)[number]["value"], "">[];
const scopeOptions = certificateScopes.map((s) => ({
  value: s,
  label: t.scopes[s],
}));
const listFilters = {
  status: parseAsStringLiteral(statusValues),
  scope: parseAsStringLiteral(certificateScopes),
};

/** 30 days, as on the public page: a valid certificate inside it reads "Expiring soon". */
export function certificateState(c: Certificate): CertificateState {
  if (c.status !== "valid") return c.status;
  return (new Date(c.valid_to).getTime() - Date.now()) / 86_400_000 <= 30
    ? "expiring"
    : "valid";
}

const subjectHref = (c: Certificate) =>
  c.subject_type === subjectTypes.passport
    ? routes.passportDetail(c.subject_id)
    : routes.facility(c.subject_id);

export function Certificates() {
  const me = useMe().data!; // the layout only renders once it is loaded
  const canIssue = me.permissions.includes(permissions.certificateIssue);
  const list = useListParams("-created_at", listFilters);
  const platform = !me.organisation;
  // An organisation starts on what it issued if it issues, else on what's about its lots.
  // A certifier or regulator searching looks across every organisation, to verify theirs.
  const scope: CertificateScope | undefined =
    platform || (inspects(me) && list.search)
      ? undefined
      : (list.filters.scope ??
        (canIssue ? certificateScopes[0] : certificateScopes[1]));
  const [revoking, setRevoking] = useState<Certificate | null>(null);

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <PageHeader
        title={t.title}
        sub={platform ? t.subPlatform : t.sub}
        actions={
          canIssue &&
          me.organisation && (
            <Link href={routes.certificateNew} className={buttonVariants()}>
              <Stamp aria-hidden />
              {t.issue}
            </Link>
          )
        }
      />
      <PageStats
        queryKey={queryKeys.certificates}
        fetch={endpoints.attestationsPage}
        cards={[
          {
            key: "total",
            ...copy.stats.certificates.total,
            icon: FileText,
            filters: { scope },
          },
          {
            key: "valid",
            ...copy.stats.certificates.valid,
            icon: CircleCheck,
            filters: { scope, status: certificateStatus.valid },
          },
          {
            key: "expired",
            ...copy.stats.certificates.expired,
            icon: Clock,
            whenAny: "attention",
            filters: { scope, status: certificateStatus.expired },
          },
          {
            key: "revoked",
            ...copy.stats.certificates.revoked,
            icon: OctagonX,
            filters: { scope, status: certificateStatus.revoked },
          },
        ]}
      />

      <CertificateList
        list={list}
        scope={scope}
        status={list.filters.status ?? ALL}
        myOrgId={me.organisation?.id}
        canRevoke={me.permissions.includes(permissions.certificateRevoke)}
        revoking={revoking}
        onRevoke={setRevoking}
        toolbar={
          <ListBar list={list} sorts={listSorts.certificate}>
            {scope && (
              <Segmented
                label={t.title}
                options={scopeOptions}
                value={scope}
                onChange={(v) => list.setFilters({ scope: v })}
              />
            )}
            <Segmented
              label={t.headers.status}
              options={statusFilters}
              value={list.filters.status ?? ALL}
              onChange={(v) => list.setFilters({ status: v || null })}
            />
          </ListBar>
        }
      />
    </main>
  );
}

function CertificateList({
  list,
  scope,
  status,
  myOrgId,
  canRevoke,
  revoking,
  onRevoke,
  toolbar,
}: {
  list: ListParamsState;
  scope?: CertificateScope;
  status: string;
  myOrgId?: string;
  canRevoke: boolean;
  revoking: Certificate | null;
  onRevoke: (c: Certificate | null) => void;
  toolbar: React.ReactNode;
}) {
  // The API decides what this caller may see (issued by or about its organisation; all for the platform).
  const query = usePaged(
    queryKeys.certificates,
    endpoints.attestationsPage,
    list.query({ scope, status }),
  );
  const rows = query.data?.data ?? [];
  const facilities = useQuery({
    queryKey: queryKeys.facilities,
    queryFn: endpoints.facilitiesList,
    enabled: rows.some((c) => c.subject_type === subjectTypes.facility),
  });

  const facilityName = (id: string) =>
    facilities.data?.find((f) => f.id === id)?.name ?? id;
  const subject = (c: Certificate) =>
    c.subject_type === subjectTypes.passport
      ? t.attachedLot(c.subject_id)
      : facilityName(c.subject_id);
  const mayRevoke = (c: Certificate) =>
    canRevoke &&
    c.status !== "revoked" &&
    (!myOrgId || c.issuer_org_id === myOrgId);

  const columns: ListColumn<Certificate>[] = [
    column("type", t.headers.type, (c) => (
      <Link href={subjectHref(c)} className="font-semibold hover:underline">
        {humanize(c.type)}
      </Link>
    )),
    column("issuer", t.headers.issuer, (c) =>
      // "Issuer · the site it was issued at", when it said.
      c.issued_at_name
        ? `${c.issuer_name} · ${c.issued_at_name}`
        : c.issuer_name,
    ),
    column("subject", t.headers.attached, (c) => (
      <span className="break-all">{subject(c)}</span>
    )),
    column(
      "valid_to",
      t.headers.valid,
      (c) => `${formatDate(c.valid_from)} – ${formatDate(c.valid_to)}`,
      { sortable: true },
    ),
    column("status", t.headers.status, (c) => {
      const state = certificateState(c);
      const look = certificateLook[state];
      return (
        <StatusChip tone={look.tone} icon={look.icon}>
          {copy.verification.certificate[state]}
        </StatusChip>
      );
    }),
    ...(canRevoke
      ? [
          actionsColumn<Certificate>((c) =>
            mayRevoke(c) ? (
              <RowAction
                label={t.revoke}
                icon={OctagonX}
                danger
                onClick={() => onRevoke(c)}
              />
            ) : null,
          ),
        ]
      : []),
  ];

  return (
    <>
      <ListTable
        label={t.title}
        columns={columns}
        rows={rows}
        total={query.data?.total ?? 0}
        list={list}
        rowKey={(c) => c.id}
        loading={query.isPending}
        error={query.isError ? problemOf(query.error) : undefined}
        empty={
          list.search
            ? copy.list.noMatch
            : status === ALL
              ? t.emptyTitle
              : t.emptyType
        }
        toolbar={toolbar}
      />
      {revoking && (
        <RevokeDialog certificate={revoking} onClose={() => onRevoke(null)} />
      )}
    </>
  );
}

/** Revoking is signed: a reason the holder can read, then the PIN. */
function RevokeDialog({
  certificate: c,
  onClose,
}: {
  certificate: Certificate;
  onClose: () => void;
}) {
  const client = useQueryClient();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RevokeForm>({
    ...liveValidation,
    resolver: zodResolver(revokeFormSchema),
    defaultValues: { reason: "" },
  });
  const revoke = useMutation({
    mutationFn: ({ reason, pin }: { reason: string; pin: string }) =>
      endpoints.revokeAttestation(c.id, reason, pin),
    meta: { inPlace: true },
    onSuccess: () => {
      client.invalidateQueries({ queryKey: queryKeys.certificates });
      onClose();
    },
  });
  const noteId = "revoke-reason";

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent
        showCloseButton={false}
        className="max-w-full sm:max-w-140 gap-4 rounded-lg border border-border bg-surface p-6 shadow-overlay"
      >
        <DialogTitle className="text-xl font-semibold">
          {t.revokeTitle}
        </DialogTitle>
        <p className="text-sm text-ink-muted">{t.revokeBody}</p>
        <div className="grid gap-1.5">
          <label htmlFor={noteId} className="text-sm font-semibold">
            {t.revokeReason}
          </label>
          <Textarea
            id={noteId}
            rows={3}
            aria-invalid={errors.reason ? true : undefined}
            aria-describedby={`${noteId}-note`}
            {...register("reason")}
          />
          <FieldNote
            id={`${noteId}-note`}
            error={errors.reason ? t.revokeReasonEmpty : undefined}
            help={t.revokeReasonHelp}
          />
        </div>
        <SigningPanel
          facts={[
            { label: t.headers.type, value: humanize(c.type) },
            {
              label: t.headers.attached,
              value: subjectLabel(c),
              mono: c.subject_type === subjectTypes.passport,
              wide: true,
            },
          ]}
          meaning={t.revokeBody}
          button={t.revokeSign}
          busy={revoke.isPending}
          error={revoke.isError ? problemOf(revoke.error) : undefined}
          onSign={(pin) => {
            void handleSubmit((v) =>
              revoke.mutate({ reason: v.reason, pin }),
            )();
          }}
        />
        <div>
          <DialogClose render={<Button variant="secondary" />}>
            {t.cancel}
          </DialogClose>
        </div>
      </DialogContent>
    </Dialog>
  );
}

// The dialog has no facility list of its own; the subject reads as the lot or the raw id.
function subjectLabel(c: Certificate): string {
  return c.subject_type === subjectTypes.passport
    ? t.attachedLot(c.subject_id)
    : t.attachedFacility(c.subject_id);
}
