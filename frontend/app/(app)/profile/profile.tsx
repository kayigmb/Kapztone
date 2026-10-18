"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CircleCheck, Clock, KeyRound, PenLine, ShieldCheck } from "lucide-react";
import QRCode from "qrcode";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { Notice } from "@/components/chaintrack/notice";
import { KeyValue } from "@/components/chaintrack/passport-card";
import { StatusChip } from "@/components/chaintrack/status";
import { TextField } from "@/components/chaintrack/text-field";
import { Button } from "@/components/ui/button";
import { endpoints, problemOf, useMe, type Me } from "@/lib/api";
import { copy, queryKeys } from "@/lib/consts";
import {
  appCodeSchema, nameSchema, passwordChangeSchema, pinSchema,
  type AppCode, type NameForm, type PasswordChange, type PinForm,
} from "@/lib/schemas/profile";
import { humanize } from "@/lib/verification";

const t = copy.profile;

export function Profile() {
  const me = useMe().data!; // the layout renders only once it is loaded
  return (
    <main className="mx-auto grid w-full max-w-content content-start gap-6 p-4 lg:p-8">
      <header className="grid gap-1">
        <h1 className="text-[22px] font-semibold lg:text-page">{t.title}</h1>
        <p className="text-sm text-ink-muted">{t.sub}</p>
      </header>
      <Details me={me} />
      <Password />
      <Pin set={!!me.pin_set} />
      <TwoFactor on={me.two_factor} />
    </main>
  );
}

function Card({ title, icon: Icon, children }: { title: string; icon: typeof KeyRound; children: React.ReactNode }) {
  return (
    <section className="grid gap-4 rounded-lg border border-border bg-surface p-4 shadow-card lg:p-6">
      <h2 className="flex items-center gap-2 text-title"><Icon className="size-5 text-ink-muted" aria-hidden />{title}</h2>
      {children}
    </section>
  );
}

function Details({ me }: { me: Me }) {
  const client = useQueryClient();
  const { register, handleSubmit, formState: { errors, isDirty } } = useForm<NameForm>({
    resolver: zodResolver(nameSchema), defaultValues: { name: me.user.name },
  });
  const save = useMutation({
    mutationFn: (v: NameForm) => endpoints.setName(v.name),
    onSuccess: () => { toast.success(t.nameSaved); client.invalidateQueries({ queryKey: queryKeys.me }); },
    onError: (err) => toast.error(problemOf(err)),
  });
  return (
    <Card title={t.details} icon={ShieldCheck}>
      <KeyValue facts={[
        { label: t.email, value: me.user.email, wide: true },
        { label: t.organisation, value: me.organisation?.name ?? copy.shell.platform },
        { label: t.role, value: humanize(me.role) },
      ]} />
      <form onSubmit={handleSubmit((v) => save.mutate(v))} noValidate className="grid gap-3 md:grid-cols-[1fr_auto] md:items-end">
        <TextField label={t.name} autoComplete="name" error={errors.name?.message} {...register("name")} />
        <Button type="submit" variant="secondary" disabled={!isDirty || save.isPending}>{t.saveName}</Button>
      </form>
    </Card>
  );
}

function Password() {
  const { register, handleSubmit, reset, formState: { errors } } = useForm<PasswordChange>({
    resolver: zodResolver(passwordChangeSchema), defaultValues: { current: "", new: "" },
  });
  const change = useMutation({
    mutationFn: endpoints.changePassword,
    onSuccess: () => { toast.success(t.passwordChanged); reset(); },
    onError: (err) => toast.error(problemOf(err)),
  });
  return (
    <Card title={t.password} icon={KeyRound}>
      <form onSubmit={handleSubmit((v) => change.mutate(v))} noValidate className="grid gap-3">
        <TextField label={t.currentPassword} type="password" autoComplete="current-password" error={errors.current?.message} {...register("current")} />
        <TextField label={t.newPassword} type="password" autoComplete="new-password" help={t.newPasswordHelp} error={errors.new?.message} {...register("new")} />
        <Button type="submit" className="justify-self-start" disabled={change.isPending}>{t.changePassword}</Button>
      </form>
    </Card>
  );
}

function Pin({ set }: { set: boolean }) {
  const client = useQueryClient();
  const { register, handleSubmit, reset, formState: { errors } } = useForm<PinForm>({
    resolver: zodResolver(pinSchema), defaultValues: { pin: "", password: "" },
  });
  const save = useMutation({
    mutationFn: endpoints.setPin,
    onSuccess: () => { toast.success(t.pinSaved); reset(); client.invalidateQueries({ queryKey: queryKeys.me }); },
    onError: (err) => toast.error(problemOf(err)),
  });
  return (
    <Card title={t.pin} icon={PenLine}>
      <div className="flex flex-wrap items-center gap-2">
        {set ? <StatusChip tone="verified" icon={CircleCheck}>{t.pinStatusSet}</StatusChip>
          : <StatusChip tone="pending" icon={Clock}>{t.pinStatusUnset}</StatusChip>}
        <p className="text-sm text-ink-muted">{set ? t.pinHelpSet : t.pinHelpUnset}</p>
      </div>
      <form onSubmit={handleSubmit((v) => save.mutate(v))} noValidate className="grid gap-3 md:grid-cols-2">
        <TextField label={t.pinLabel} type="password" inputMode="numeric" maxLength={6} mono autoComplete="off"
          error={errors.pin?.message} {...register("pin")} />
        <TextField label={t.pinPassword} type="password" autoComplete="current-password"
          error={errors.password?.message} {...register("password")} />
        <Button type="submit" className="justify-self-start" disabled={save.isPending}>{set ? t.changePin : t.setPin}</Button>
      </form>
    </Card>
  );
}

/** `on` undefined: the API didn't say, so both turning on and off are offered. */
function TwoFactor({ on }: { on: boolean | undefined }) {
  const [setup, setSetup] = useState<{ secret: string; otpauth_url: string }>();
  const [codes, setCodes] = useState<string[]>();
  const client = useQueryClient();
  const done = (message: string, recovery?: string[]) => {
    toast.success(message);
    setSetup(undefined);
    if (recovery?.length) setCodes(recovery);
    client.invalidateQueries({ queryKey: queryKeys.me });
  };
  const start = useMutation({ mutationFn: endpoints.twoFactorSetup, onSuccess: setSetup, onError: (e) => toast.error(problemOf(e)) });
  const enable = useMutation({
    mutationFn: endpoints.twoFactorEnable, onSuccess: (r) => done(t.turnedOn, r.recovery_codes), onError: (e) => toast.error(problemOf(e)),
  });
  const disable = useMutation({ mutationFn: endpoints.twoFactorDisable, onSuccess: () => done(t.turnedOff), onError: (e) => toast.error(problemOf(e)) });
  const renew = useMutation({
    mutationFn: endpoints.newRecoveryCodes, onSuccess: (r) => setCodes(r.recovery_codes), onError: (e) => toast.error(problemOf(e)),
  });

  return (
    <Card title={t.twoFactor} icon={ShieldCheck}>
      <p className="text-sm text-ink-muted">{t.twoFactorHelp}</p>
      {codes && (
        <Notice kind="warning">
          <strong>{t.recoveryTitle}</strong> {t.recoveryBody}
          <ul className="mt-2 grid grid-cols-2 gap-1 font-mono text-sm">{codes.map((c) => <li key={c}>{c}</li>)}</ul>
          <Button variant="secondary" className="mt-3" onClick={() => setCodes(undefined)}>{t.done}</Button>
        </Notice>
      )}
      {on !== true && !setup && (
        <Button className="justify-self-start" onClick={() => start.mutate()} disabled={start.isPending}>{t.startSetup}</Button>
      )}
      {setup && (
        <div className="grid gap-3">
          <p className="text-sm">{t.scanQr}</p>
          <QrImage text={setup.otpauth_url} />
          <p className="text-sm">{t.secret}: <span className="font-mono break-all">{setup.secret}</span></p>
          <CodeForm label={t.code} submit={t.turnOn} pending={enable.isPending} onCode={(c) => enable.mutate(c)} />
        </div>
      )}
      {on !== false && (
        <>
          <CodeForm label={t.code} help={t.turnOffHelp} submit={t.turnOff} variant="danger-outline" pending={disable.isPending}
            onCode={(c) => disable.mutate(c)} />
          <CodeForm label={t.code} help={t.newRecoveryHelp} submit={t.newRecovery} variant="secondary" pending={renew.isPending}
            onCode={(c) => renew.mutate(c)} />
        </>
      )}
    </Card>
  );
}

function CodeForm({ label, help, submit, onCode, pending, variant = "primary" }: {
  label: string; help?: string; submit: string; onCode: (code: string) => void; pending: boolean;
  variant?: "primary" | "secondary" | "danger-outline";
}) {
  const { register, handleSubmit, reset, formState: { errors } } = useForm<AppCode>({
    resolver: zodResolver(appCodeSchema), defaultValues: { code: "" },
  });
  return (
    <form onSubmit={handleSubmit((v) => { onCode(v.code); reset(); })} noValidate className="grid gap-3 md:grid-cols-[1fr_auto] md:items-end">
      <TextField label={label} help={help} mono inputMode="numeric" maxLength={6} autoComplete="one-time-code"
        error={errors.code?.message} {...register("code")} />
      <Button type="submit" variant={variant} disabled={pending}>{submit}</Button>
    </form>
  );
}

// The setup link as a QR code, drawn in the browser; nothing leaves the page.
function QrImage({ text }: { text: string }) {
  const qr = useQuery({ queryKey: ["qr", text], queryFn: () => QRCode.toDataURL(text, { margin: 1, width: 192 }), staleTime: Infinity });
  return qr.data
    // eslint-disable-next-line @next/next/no-img-element -- a data: URL, nothing for next/image to optimise
    ? <img src={qr.data} alt="" width={192} height={192} className="rounded-md bg-surface p-2 ring-1 ring-border" />
    : <span className="size-48 animate-pulse rounded-md bg-surface-sunken" />;
}
