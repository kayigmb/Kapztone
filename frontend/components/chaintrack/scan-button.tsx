"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { IScannerControls } from "@zxing/browser";
import { CameraOff, QrCode } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { useForm } from "react-hook-form";

import { TextField } from "@/components/chaintrack/text-field";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { copy, routes } from "@/lib/consts";
import { liveValidation } from "@/lib/form";
import {
  lotCodeSchema,
  passportRef,
  type LotCodeForm,
} from "@/lib/schemas/scan";

const t = copy.scan;

/**
 * Opens the camera to read a passport's QR label, with a typed lot code beside it.
 * The camera needs a secure page (https, localhost); elsewhere only typing works.
 * `trigger` renders the button that opens it, so the tab bar and the dashboard can each style their own.
 */
export function ScanButton({
  trigger,
}: {
  trigger: (open: () => void) => React.ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const router = useRouter();
  const go = (ref: string) => {
    setOpen(false);
    router.push(routes.passport(encodeURIComponent(ref)));
  };

  return (
    <>
      {trigger(() => setOpen(true))}
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{t.title}</DialogTitle>
            <DialogDescription>{t.hint}</DialogDescription>
          </DialogHeader>
          {open && <Camera onCode={(text) => go(passportRef(text))} />}
          <TypedCode onCode={go} />
        </DialogContent>
      </Dialog>
    </>
  );
}

function Camera({ onCode }: { onCode: (text: string) => void }) {
  const video = useRef<HTMLVideoElement>(null);
  // Only rendered once the dialog is open, so this runs in the browser.
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
    // Loaded only when the dialog opens: the reader is large and most pages never scan.
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
      .catch(() => setUnavailable(true)); // refused permission, no camera
    return () => {
      stopped = true;
      controls?.stop();
    };
  }, [unavailable]);

  if (unavailable) {
    return (
      <p className="flex items-center gap-2 rounded-md bg-surface-sunken p-3 text-sm text-ink-muted">
        <CameraOff className="size-5 shrink-0" aria-hidden />
        {t.noCamera}
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
      onSubmit={handleSubmit((v) => onCode(v.code))}
      className="grid gap-3"
      noValidate
    >
      <TextField
        label={t.codeLabel}
        help={t.codeHint}
        error={errors.code?.message}
        mono
        autoComplete="off"
        autoCapitalize="characters"
        {...register("code")}
      />
      <Button type="submit" size="scan" block>
        <QrCode aria-hidden />
        {t.open}
      </Button>
    </form>
  );
}
