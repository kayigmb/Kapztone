"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useParams } from "next/navigation";
import {
  ArrowLeftRight,
  CheckCheck,
  CircleCheck,
  Clock,
  Coffee,
  ExternalLink,
  Flag as FlagIcon,
  Gem,
  Layers,
  Leaf,
  OctagonX,
  Package,
  QrCode,
  Truck,
  WifiOff,
  type LucideIcon,
} from "lucide-react";
import { useState } from "react";

import { CertificateTile } from "@/components/chaintrack/certificate-tile";
import { JourneyTimeline } from "@/components/chaintrack/journey-timeline";
import { Notice } from "@/components/chaintrack/notice";
import { DataWidget } from "@/components/chaintrack/data-widget";
import { LoadingWithLogo, PageState } from "@/components/chaintrack/page-state";
import { PassportCard, type Fact } from "@/components/chaintrack/passport-card";
import { SigningPanel } from "@/components/chaintrack/signing";
import { Status, StatusChip } from "@/components/chaintrack/status";
import { type Tone } from "@/components/chaintrack/tone";
import { TechnicalDetails } from "@/components/chaintrack/technical-details";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogTitle,
} from "@/components/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  ApiError,
  endpoints,
  problemOf,
  useMe,
  type Certificate,
  type Flag,
  type PassportDetail,
} from "@/lib/api";
import {
  copy,
  flagStatus,
  passportStatus,
  queryKeys,
  routes,
  screenCopy,
} from "@/lib/consts";
import { cn } from "@/lib/utils";
import {
  formatDate,
  formatDateTime,
  formatQuantity,
  humanize,
  type CertificateState,
  type View,
} from "@/lib/verification";

const t = copy.passportView;

// Copy this screen needs that consts.ts doesn't hold yet.
const local = screenCopy.passportDetail;

/** Raising a flag starts from the lot: the flags screen reads ?passport=. */
const flagRoute = (passportId: string) => `/flags?passport=${passportId}`;

// Icons by the first word of a commodity; anything else is a generic lot.
const commodityIcons: Record<string, LucideIcon> = {
  coffee: Coffee,
  cocoa: Leaf,
  tin: Gem,
  gold: Gem,
  mineral: Gem,
  transport: Truck,
};
const firstWord = (s: string) => s.split(/[\s_]/)[0];

export const statusLook: Record<
  string,
  { tone: Tone; icon: LucideIcon; label: string }
> = {
  active: { tone: "verified", icon: CircleCheck, label: t.status.active },
  consumed: { tone: "neutral", icon: CheckCheck, label: t.status.consumed },
  retired: { tone: "neutral", icon: Package, label: t.status.retired },
  revoked: { tone: "danger", icon: OctagonX, label: t.status.revoked },
};

const flagLook: Record<
  string,
  { tone: Tone; icon: LucideIcon; label: string }
> = {
  open: { tone: "pending", icon: Clock, label: copy.flags.status.open },
  cleared: {
    tone: "verified",
    icon: CircleCheck,
    label: copy.flags.status.cleared,
  },
  confirmed: {
    tone: "danger",
    icon: OctagonX,
    label: copy.flags.status.confirmed,
  },
};

function Card({
  title,
  note,
  className,
  children,
}: {
  title: string;
  note?: string;
  className?: string;
  children: React.ReactNode;
}) {
  return (
    <section
      className={cn(
        "min-w-0 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6",
        className,
      )}
    >
      <header className="mb-4 flex flex-wrap items-baseline justify-between gap-x-3">
        <h2 className="text-section">{title}</h2>
        {note && <p className="text-small text-ink-muted">{note}</p>}
      </header>
      {children}
    </section>
  );
}

function eventState(
  kind: string,
  last: boolean,
): "done" | "current" | "warning" | "failed" {
  if (kind === "REVOKE") return "failed";
  if (kind === "FLAG") return "warning";
  return last ? "current" : "done";
}

/** 30 days, as on the public page: a valid certificate inside it reads "Expiring soon". */
function certificateState(c: Certificate): CertificateState {
  if (c.status !== "valid") return c.status;
  return (new Date(c.valid_to).getTime() - Date.now()) / 86_400_000 <= 30
    ? "expiring"
    : "valid";
}

export function PassportDetail() {
  const { id } = useParams<{ id: string }>();
  const query = useQuery({
    queryKey: queryKeys.passport(id),
    queryFn: () => endpoints.passport(id),
  });

  if (query.isPending) return <LoadingWithLogo />;
  if (query.isError) {
    if (query.error instanceof ApiError && query.error.status === 404) {
      return (
        <PageState icon={QrCode} title={local.notFoundTitle}>
          {local.notFoundBody}
        </PageState>
      );
    }
    return (
      <PageState
        icon={WifiOff}
        tone="danger"
        title={copy.errors.loadTitle}
        retry={() => query.refetch()}
      >
        {problemOf(query.error)}
      </PageState>
    );
  }
  return <Detail passport={query.data} />;
}

function Detail({ passport: p }: { passport: PassportDetail }) {
  const me = useMe().data!;
  const holder = me.organisation?.id === p.organisation_id;
  const [retiring, setRetiring] = useState(false);

  const directory = useQuery({
    queryKey: queryKeys.directory,
    queryFn: endpoints.directory,
  });
  const orgName = (orgId: string) => directory.data?.get(orgId) ?? orgId;

  const facility = useQuery({
    queryKey: queryKeys.facility(p.facility_id ?? ""),
    queryFn: () => endpoints.facility(p.facility_id as string),
    enabled: !!p.facility_id,
  });

  const flagsQuery = useQuery({
    queryKey: queryKeys.flags,
    queryFn: () => endpoints.flagsList(),
  });
  const flags = (flagsQuery.data ?? []).filter((f) => f.passport_id === p.id);
  const openFlag = flags.some((f) => f.status === flagStatus.open);

  const certificatesQuery = useQuery({
    queryKey: [...queryKeys.certificates, "subject", p.id],
    queryFn: () => endpoints.attestationsList(p.id),
  });
  const certificates: View["certificates"] = (certificatesQuery.data ?? []).map(
    (c) => {
      const state = certificateState(c);
      return {
        id: c.id,
        type: humanize(c.type),
        issuer: orgName(c.issuer_org_id),
        state,
        validFrom: formatDate(c.valid_from),
        validTo: formatDate(c.valid_to),
        note:
          state === "revoked"
            ? copy.verification.certificateNote.revoked(c.revoke_reason ?? null)
            : state === "expired"
              ? copy.verification.certificateNote.expired(
                  formatDate(c.valid_to),
                )
              : undefined,
      };
    },
  );

  const events = p.events ?? [];
  const steps: View["journey"] = events.map((e, i) => ({
    id: e.id,
    kind: e.kind,
    title: t.eventKinds[e.kind] ?? humanize(e.kind.toLowerCase()),
    date: formatDateTime(e.created_at),
    parties: { from: e.from_org ?? undefined, to: e.to_org ?? undefined },
    quantity: formatQuantity(e.quantity, p.unit),
    signedByBoth: e.kind === "HANDOVER",
    state: eventState(e.kind, i === events.length - 1),
  }));

  const canRetire =
    holder && p.status === passportStatus.active && p.remaining > 0;
  const status = statusLook[p.status] ?? statusLook.active;
  const facts: Fact[] = [
    {
      label: t.facts.remaining,
      value: local.held(p.remaining, p.quantity, p.unit),
      mono: true,
    },
    {
      label: t.facts.status,
      value: (
        <Status tone={status.tone} icon={status.icon}>
          {status.label}
        </Status>
      ),
    },
    { label: t.facts.holder, value: orgName(p.organisation_id), wide: true },
    ...(p.facility_id
      ? [
          {
            label: t.facts.facility,
            value: facility.data?.name ?? p.facility_id,
          },
        ]
      : []),
    { label: t.facts.issued, value: formatDate(p.created_at) },
    {
      label: local.stagesLabel,
      value: `${events.length} ${events.length === 1 ? "stage" : "stages"}`,
    },
  ];

  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <header className="grid gap-1">
        <h1 className="font-mono text-2xl font-semibold break-all lg:text-page">
          {p.code}
        </h1>
        <p className="text-sm text-ink-muted">{humanize(p.commodity)}</p>
      </header>

      <div className="grid gap-4 lg:grid-cols-[1.6fr_1fr] lg:gap-6">
        <PassportCard
          product={humanize(p.commodity)}
          lotCode={p.code}
          icon={commodityIcons[firstWord(p.commodity).toLowerCase()] ?? Package}
          facts={facts}
          footer={{
            left: copy.verification.issued(formatDate(p.created_at)),
            verdict: openFlag
              ? "warning"
              : p.status === passportStatus.revoked
                ? "failed"
                : "verified",
          }}
          flagged={openFlag || p.status === passportStatus.revoked}
        />

        <div className="grid content-start gap-3">
          {holder ? (
            <>
              {openFlag && (
                <Notice kind="warning">{copy.handover.openFlag}</Notice>
              )}
              <div className="grid grid-cols-2 gap-2 lg:grid-cols-1">
                {!openFlag && (
                  <Link
                    href={routes.handover(p.id)}
                    className={cn(
                      buttonVariants({ variant: "primary" }),
                      "col-span-2 lg:col-span-1",
                    )}
                  >
                    <ArrowLeftRight aria-hidden />
                    {t.actions.handOver}
                  </Link>
                )}
                <Link
                  href={flagRoute(p.id)}
                  className={buttonVariants({ variant: "danger-outline" })}
                >
                  <FlagIcon aria-hidden />
                  {copy.flags.raise}
                </Link>
                {canRetire && (
                  <Button variant="secondary" onClick={() => setRetiring(true)}>
                    <Package aria-hidden />
                    {t.actions.retire}
                  </Button>
                )}
                <Link
                  href={routes.passport(p.code)}
                  className={buttonVariants({ variant: "secondary" })}
                >
                  <ExternalLink aria-hidden />
                  {t.actions.publicPage}
                </Link>
              </div>
            </>
          ) : (
            <Notice kind="info">{t.notHolder}</Notice>
          )}
        </div>
      </div>

      <Tabs defaultValue="journey" className="grid min-w-0 gap-4">
        <TabsList aria-label={p.code}>
          <TabsTrigger value="journey">{local.journeyTab}</TabsTrigger>
          <TabsTrigger value="evidence">{local.evidenceTab}</TabsTrigger>
          <TabsTrigger value="record">{local.recordTab}</TabsTrigger>
        </TabsList>
        <TabsContent value="journey" className="grid min-w-0 gap-4">
          <Card title={t.journey} note={copy.verification.steps(events.length)}>
            {steps.length === 0 ? (
              <p className="text-sm text-ink-muted">{t.journeyEmpty}</p>
            ) : (
              <JourneyTimeline steps={steps} />
            )}
          </Card>
          <Genealogy id={p.id} unit={p.unit} />
        </TabsContent>
        <TabsContent value="evidence" className="grid min-w-0 gap-4">
          <Card title={t.certificates}>
            <DataWidget query={certificatesQuery}>
              {() =>
                certificates.length === 0 ? (
                  <p className="text-sm text-ink-muted">
                    {t.certificatesEmpty}
                  </p>
                ) : (
                  <div className="grid gap-4 md:grid-cols-3">
                    {certificates.map((c) => (
                      <CertificateTile key={c.id} certificate={c} />
                    ))}
                  </div>
                )
              }
            </DataWidget>
          </Card>
          <Card title={local.flagsTitle}>
            <DataWidget query={flagsQuery}>
              {(rows) => {
                const mine = rows.filter((f) => f.passport_id === p.id);
                return mine.length === 0 ? (
                  <p className="text-sm text-ink-muted">{local.flagsEmpty}</p>
                ) : (
                  <ul className="grid divide-y divide-border rounded-lg border border-border bg-surface shadow-card">
                    {mine.map((f: Flag) => {
                      const look = flagLook[f.status] ?? flagLook.open;
                      return (
                        <li key={f.id} className="grid gap-2 p-4">
                          <div className="flex flex-wrap items-center justify-between gap-2">
                            <StatusChip tone={look.tone} icon={look.icon}>
                              {look.label}
                            </StatusChip>
                            <span className="text-small text-ink-subtle">
                              {copy.flags.raised(formatDateTime(f.created_at))}
                            </span>
                          </div>
                          <p className="text-sm">{f.reason}</p>
                          {f.resolution_note && (
                            <p className="text-small text-ink-muted">
                              <span className="font-semibold">
                                {local.resolution}:
                              </span>{" "}
                              {f.resolution_note}
                            </p>
                          )}
                        </li>
                      );
                    })}
                  </ul>
                );
              }}
            </DataWidget>
          </Card>
        </TabsContent>
        <TabsContent value="record">
          <TechnicalDetails
            rows={[
              { label: local.technical.fingerprint, value: p.data_hash },
              { label: local.technical.sealed, value: p.tx_hash ?? null },
              ...(p.template_id
                ? [{ label: local.technical.template, value: p.template_id }]
                : []),
              ...(p.facility_id
                ? [{ label: local.technical.facility, value: p.facility_id }]
                : []),
              { label: local.technical.created, value: p.created_at },
              ...(p.updated_at
                ? [{ label: local.technical.updated, value: p.updated_at }]
                : []),
            ]}
          />
        </TabsContent>
      </Tabs>

      {canRetire && (
        <RetireDialog
          passport={p}
          open={retiring}
          onOpenChange={setRetiring}
          held={local.held(p.remaining, p.quantity, p.unit)}
        />
      )}
    </main>
  );
}

function RetireDialog({
  passport: p,
  open,
  onOpenChange,
  held,
}: {
  passport: PassportDetail;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  held: string;
}) {
  const client = useQueryClient();
  const retire = useMutation({
    mutationFn: (pin: string) => endpoints.retirePassport(p.id, pin),
    meta: { inPlace: true },
    onSuccess: () => {
      client.invalidateQueries({ queryKey: queryKeys.passport(p.id) });
      client.invalidateQueries({ queryKey: queryKeys.passports });
      onOpenChange(false);
    },
  });
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        showCloseButton={false}
        className="max-w-full sm:max-w-140 gap-4 rounded-lg border border-border bg-surface p-6 shadow-overlay"
      >
        <DialogTitle className="text-xl font-semibold">
          {t.retireTitle}
        </DialogTitle>
        <p className="text-sm text-ink-muted">{t.retireBody}</p>
        <SigningPanel
          facts={[
            { label: copy.handover.facts.lot, value: p.code, mono: true },
            { label: t.facts.product, value: humanize(p.commodity) },
            { label: t.facts.remaining, value: held, mono: true },
          ]}
          meaning={local.retireMeaning}
          button={local.retireSign}
          onSign={(pin) => retire.mutate(pin)}
          busy={retire.isPending}
          error={retire.isError ? problemOf(retire.error) : undefined}
        />
        <div>
          <DialogClose render={<Button variant="secondary" />}>
            {local.cancel}
          </DialogClose>
        </div>
      </DialogContent>
    </Dialog>
  );
}

type Direction = "upstream" | "downstream";
const directions: Direction[] = ["upstream", "downstream"];

function Genealogy({ id, unit }: { id: string; unit: string }) {
  const [direction, setDirection] = useState<Direction>("upstream");
  const query = useQuery({
    queryKey: [...queryKeys.passport(id), "genealogy", direction],
    queryFn: () => endpoints.genealogy(id, direction),
  });

  return (
    <Card title={direction === "upstream" ? t.madeFrom : t.madeInto}>
      <div
        className="mb-4 inline-flex gap-0.5 rounded-md border border-border bg-surface-sunken p-0.75"
        role="group"
        aria-label={t.madeFrom}
      >
        {directions.map((d) => (
          <button
            key={d}
            type="button"
            onClick={() => setDirection(d)}
            aria-pressed={direction === d}
            className={cn(
              "min-h-tap rounded-sm px-3 text-sm",
              direction === d
                ? "bg-surface font-semibold text-ink shadow-card"
                : "text-ink-muted hover:text-ink",
            )}
          >
            {d === "upstream" ? local.upstream : local.downstream}
          </button>
        ))}
      </div>

      <DataWidget query={query}>
        {(g) => {
          const ns = g.nodes ?? [];
          const es = g.edges ?? [];
          const byId = new Map(ns.map((n) => [n.id, n]));
          const isEmpty = ns.every((n) => n.id === id) && es.length === 0;
          return isEmpty ? (
            <p className="text-sm text-ink-muted">{t.noGenealogy}</p>
          ) : (
            <div className="grid gap-4">
              <ul className="grid gap-2">
                {ns.map((n) => {
                  const look = statusLook[n.status] ?? statusLook.active;
                  const self = n.id === id;
                  return (
                    <li
                      key={n.id}
                      className={cn(
                        "grid min-w-0 grid-cols-[auto_1fr] items-center gap-3 rounded-md px-3 py-2 md:grid-cols-[auto_1fr_auto]",
                        self
                          ? "border-2 border-primary bg-primary-soft"
                          : "border border-border-strong bg-surface",
                      )}
                    >
                      <span
                        className={cn(
                          "grid size-8 place-items-center rounded-sm",
                          self
                            ? "bg-primary text-on-primary"
                            : "bg-surface-sunken text-ink-muted",
                        )}
                      >
                        <Layers className="size-4" aria-hidden />
                      </span>
                      <div className="min-w-0">
                        <Link
                          href={routes.passportDetail(n.id)}
                          className={cn(
                            "font-mono text-sm font-medium break-all",
                            self ? "text-primary" : "hover:text-primary",
                          )}
                        >
                          {n.code}
                        </Link>
                        <p className="text-small text-ink-muted">
                          {humanize(n.commodity)}
                        </p>
                      </div>
                      <div className="col-start-2 flex flex-wrap items-center gap-3 md:col-start-3">
                        <span className="fig text-sm">
                          {formatQuantity(n.quantity, n.unit)}
                        </span>
                        <StatusChip tone={look.tone} icon={look.icon}>
                          {look.label}
                        </StatusChip>
                      </div>
                    </li>
                  );
                })}
              </ul>
              {es.length > 0 && (
                <ul className="grid gap-2">
                  {es.map((e, i) => (
                    <li
                      key={i}
                      className="flex flex-wrap items-center gap-2 rounded-md bg-surface-sunken px-3 py-2 text-small"
                    >
                      <span className="font-mono text-sm break-all">
                        {byId.get(e.parent_id)?.code ?? e.parent_id}
                      </span>
                      <ArrowLeftRight
                        className="size-4 shrink-0 text-ink-subtle"
                        aria-hidden
                      />
                      <span className="font-mono text-sm break-all">
                        {byId.get(e.child_id)?.code ?? e.child_id}
                      </span>
                      <span className="text-ink-muted">
                        {local.edge(
                          humanize(e.stage),
                          formatQuantity(
                            e.quantity_used,
                            byId.get(e.child_id)?.unit ?? unit,
                          ),
                        )}
                      </span>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          );
        }}
      </DataWidget>
    </Card>
  );
}
