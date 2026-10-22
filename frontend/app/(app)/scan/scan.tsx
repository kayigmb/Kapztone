"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import type { IScannerControls } from "@zxing/browser";
import {
  ArrowRight,
  CameraOff,
  ExternalLink,
  QrCode,
  Search,
  SearchX,
  WifiOff,
} from "lucide-react";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { useForm } from "react-hook-form";

import { PageHeader } from "@/components/chaintrack/page-header";
import { Loading, PageState } from "@/components/chaintrack/page-state";
import { KeyValue } from "@/components/chaintrack/passport-card";
import { StatusChip, verdictIcon } from "@/components/chaintrack/status";
import { verdictTone } from "@/components/chaintrack/tone";
import { TechnicalDetails } from "@/components/chaintrack/technical-details";
import { TextField } from "@/components/chaintrack/text-field";
import { Button, buttonVariants } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { endpoints } from "@/lib/api";
import {
  copy,
  queryKeys,
  routes,
  type PassportStatus,
  screenCopy,
} from "@/lib/consts";
import { liveValidation } from "@/lib/form";
import {
  lotCodeSchema,
  passportRef,
  type LotCodeForm,
} from "@/lib/schemas/scan";
import { formatQuantity, formatDate, humanize } from "@/lib/verification";

const t = copy.scanPage;
const s = copy.scan;
const v = copy.verification;

// Copy this screen needs that consts doesn't hold yet.
const local = screenCopy.scan;

export function Scan() {
  const [lotRef, setLotRef] = useState<string>();
  return (
    <main className="grid w-full content-start gap-6 p-4 lg:p-8">
      <PageHeader title={t.title} sub={t.hint} />

      <section className="grid gap-3 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
        <LotScanner onCode={setLotRef} />
        <TypedLot onCheck={setLotRef} />
      </section>

      {lotRef === undefined ? (
        <Empty />
      ) : (
        <Result
          key={lotRef}
          lotRef={lotRef}
          onAgain={() => setLotRef(undefined)}
        />
      )}
    </main>
  );
}

export function LotScanner({
  onCode,
  disabled,
}: {
  onCode: (ref: string) => void;
  disabled?: boolean;
}) {
  const [open, setOpen] = useState(false);
  const take = (text: string) => {
    setOpen(false);
    onCode(text);
  };
  return (
    <>
      <Button
        size="scan"
        block
        onClick={() => setOpen(true)}
        disabled={disabled}
      >
        <QrCode aria-hidden />
        {s.open}
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{s.title}</DialogTitle>
            <DialogDescription>{s.hint}</DialogDescription>
          </DialogHeader>
          {open && <Camera onCode={(text) => take(passportRef(text))} />}
          <TypedCode onCode={take} />
        </DialogContent>
      </Dialog>
    </>
  );
}

function Camera({ onCode }: { onCode: (text: string) => void }) {
  const video = useRef<HTMLVideoElement>(null);
  const [unavailable, setUnavailable] = useState(
    () => !window.isSecureContext || !navigator.mediaDevices?.getUserMedia,
  );
  const done = useRef(onCode);
  useEffect(() => {
    done.current = onCode;
  }, [onCode]);

  useEffect(() => {
    if (unavailable) return;
    let controls: IScannerControls | undefined;
    let stopped = false;
    import("@zxing/browser")
      .then(({ BrowserQRCodeReader }) =>
        new BrowserQRCodeReader().decodeFromVideoDevice(
          undefined,
          video.current!,
          (result, _err, c) => {
            if (result && !stopped) {
              stopped = true;
              c.stop();
              done.current(result.getText());
            }
          },
        ),
      )
      .then((c) => {
        controls = c;
        if (stopped) c.stop();
      })
      .catch(() => setUnavailable(true));
    return () => {
      stopped = true;
      controls?.stop();
    };
  }, [unavailable]);

  if (unavailable) {
    return (
      <p className="flex items-center gap-2 rounded-md bg-surface-sunken p-3 text-sm text-ink-muted">
        <CameraOff className="size-5 shrink-0" aria-hidden />
        {s.noCamera}
      </p>
    );
  }
  return (
    <video
      ref={video}
      muted
      playsInline
      className="aspect-square w-full rounded-lg bg-surface-sunken object-cover"
    />
  );
}

function TypedCode({ onCode }: { onCode: (ref: string) => void }) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LotCodeForm, unknown, { code: string }>({
    ...liveValidation,
    resolver: zodResolver(lotCodeSchema),
    defaultValues: { code: "" },
  });
  return (
    <form
      onSubmit={handleSubmit((val) => onCode(val.code))}
      className="grid gap-3"
      noValidate
    >
      <TextField
        label={s.codeLabel}
        help={s.codeHint}
        error={errors.code?.message}
        mono
        autoComplete="off"
        autoCapitalize="characters"
        {...register("code")}
      />
      <Button type="submit" size="scan" block>
        <QrCode aria-hidden />
        {s.open}
      </Button>
    </form>
  );
}

function TypedLot({ onCheck }: { onCheck: (ref: string) => void }) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LotCodeForm, unknown, { code: string }>({
    ...liveValidation,
    resolver: zodResolver(lotCodeSchema),
    defaultValues: { code: "" },
  });
  return (
    <form
      onSubmit={handleSubmit((val) => onCheck(val.code))}
      noValidate
      className="grid gap-2"
    >
      <p className="text-small text-ink-subtle">{t.orType}</p>
      <div className="flex flex-wrap items-end gap-2">
        <TextField
          label={s.codeLabel}
          help={s.codeHint}
          error={errors.code?.message}
          mono
          className="min-w-56 flex-1"
          autoComplete="off"
          autoCapitalize="characters"
          {...register("code")}
        />
        <Button type="submit" variant="secondary">
          <Search aria-hidden />
          {local.check}
        </Button>
      </div>
    </form>
  );
}

function Empty() {
  return (
    <div className="grid justify-items-center gap-2 px-4 py-8 text-center">
      <span className="grid size-14 place-items-center rounded-full bg-surface-sunken text-ink-muted">
        <QrCode className="size-7" aria-hidden />
      </span>
      <h2 className="text-title">{local.emptyTitle}</h2>
      <p className="max-w-[48ch] text-ink-muted">{local.emptyBody}</p>
    </div>
  );
}

function Result({ lotRef, onAgain }: { lotRef: string; onAgain: () => void }) {
  const q = useQuery({
    queryKey: queryKeys.verify(lotRef),
    queryFn: () => endpoints.verify(lotRef),
  });

  if (q.isPending) return <Loading>{v.checking}</Loading>;
  if (q.isError) {
    return (
      <PageState
        icon={WifiOff}
        tone="danger"
        title={v.errorTitle}
        retry={() => q.refetch()}
      >
        {v.errorBody}
      </PageState>
    );
  }
  if (!q.data) {
    return (
      <PageState icon={SearchX} title={v.notFoundTitle}>
        {v.notFoundBody}
      </PageState>
    );
  }

  const p = q.data.passport;
  const holder =
    q.data.organisations.find((o) => o.id === p.organisation_id)?.name ??
    v.unknownOrg;
  const status =
    copy.passportView.status[p.status as PassportStatus] ?? humanize(p.status);
  return (
    <section
      role="status"
      className="grid gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6"
    >
      <header className="flex flex-wrap items-center justify-between gap-3">
        <div className="min-w-0">
          <h2 className="text-section">{humanize(p.commodity)}</h2>
          <p className="code break-all text-ink-muted">{p.id}</p>
        </div>
        <StatusChip
          tone={verdictTone[q.data.verdict]}
          icon={verdictIcon[q.data.verdict]}
        >
          {v.verdict[q.data.verdict]}
        </StatusChip>
      </header>

      <KeyValue
        facts={[
          { label: local.lotCode, value: p.id, mono: true, wide: true },
          { label: v.facts.commodity, value: humanize(p.commodity) },
          {
            label: v.facts.quantity,
            value: formatQuantity(p.quantity, p.unit),
            mono: true,
          },
          { label: v.facts.status, value: status },
          { label: v.facts.holder, value: holder, wide: true },
          { label: local.created, value: formatDate(p.created_at) },
        ]}
      />

      <TechnicalDetails
        rows={[{ label: v.technical.fingerprint, value: p.data_hash }]}
      />

      <div className="flex flex-wrap gap-2">
        <Link href={routes.passportDetail(p.id)} className={buttonVariants()}>
          {local.openRecord}
          <ArrowRight aria-hidden />
        </Link>
        <Link
          href={routes.passport(p.id)}
          className={buttonVariants({ variant: "secondary" })}
        >
          {copy.passportView.actions.publicPage}
          <ExternalLink aria-hidden />
        </Link>
        <Button variant="ghost" onClick={onAgain}>
          {local.again}
        </Button>
      </div>
    </section>
  );
}
