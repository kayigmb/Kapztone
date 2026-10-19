"use client";

// The last step before a signed record exists (patterns.md → Signing):
// what is being signed, what the signature means, then the PIN, then a
// button that names the result.

import { useRef, useState } from "react";
import { LoaderCircle, PenLine } from "lucide-react";

import { Button } from "@/components/ui/button";
import { copy } from "@/lib/consts";
import { cn } from "@/lib/utils";
import { Notice } from "./notice";
import { KeyValue, type Fact } from "./passport-card";

const t = copy.signing;

/** Six boxes for the 6-digit signing PIN; typing, arrows and paste all work. */
export function PinBoxes({
  value,
  onChange,
  disabled,
  id = "signing-pin",
}: {
  value: string;
  onChange: (pin: string) => void;
  disabled?: boolean;
  id?: string;
}) {
  const refs = useRef<(HTMLInputElement | null)[]>([]);
  const digits = value.padEnd(6, " ").slice(0, 6).split("");

  const put = (i: number, char: string) => {
    const next = value.padEnd(6, " ").split("");
    next[i] = char;
    onChange(next.join("").replace(/\s+$/, ""));
    if (char && i < 5) refs.current[i + 1]?.focus();
  };

  return (
    <div className="flex gap-2" role="group" aria-label={t.label}>
      {digits.map((d, i) => (
        <input
          key={i}
          id={i === 0 ? id : undefined}
          ref={(el) => {
            refs.current[i] = el;
          }}
          value={d.trim()}
          disabled={disabled}
          inputMode="numeric"
          autoComplete={i === 0 ? "one-time-code" : "off"}
          maxLength={1}
          aria-label={`${t.label} ${i + 1}`}
          className="h-13 w-11 rounded-md border border-border bg-surface text-center font-mono text-2xl font-semibold tabular-nums outline-none focus-visible:border-primary focus-visible:ring-3 focus-visible:ring-ring/50 disabled:bg-surface-sunken"
          onChange={(e) => {
            const char = e.target.value.replace(/\D/g, "").slice(-1);
            if (char) put(i, char);
            else if (d !== " ") put(i, "");
          }}
          onKeyDown={(e) => {
            if (e.key === "Backspace" && !digits[i].trim() && i > 0) {
              e.preventDefault();
              put(i - 1, "");
              refs.current[i - 1]?.focus();
            }
            if (e.key === "ArrowLeft" && i > 0) refs.current[i - 1]?.focus();
            if (e.key === "ArrowRight" && i < 5) refs.current[i + 1]?.focus();
          }}
          onPaste={(e) => {
            e.preventDefault();
            const pasted = e.clipboardData
              .getData("text")
              .replace(/\D/g, "")
              .slice(0, 6);
            if (pasted) {
              onChange(pasted);
              refs.current[Math.min(pasted.length, 5)]?.focus();
            }
          }}
        />
      ))}
    </div>
  );
}

/**
 * What is being signed, one sentence on what the signature means, then the
 * PIN panel and the button that names the result. The panel owns the PIN;
 * the page owns the mutation and passes `busy` and `error` back in.
 */
export function SigningPanel({
  facts,
  meaning,
  button,
  onSign,
  busy,
  error,
  blocked,
  className,
}: {
  facts?: Fact[];
  meaning?: string;
  button: string;
  onSign: (pin: string) => void;
  busy?: boolean;
  error?: string;
  blocked?: string;
  className?: string;
}) {
  const [pin, setPin] = useState("");
  const ready = /^\d{6}$/.test(pin);
  return (
    <section className={cn("grid gap-4", className)}>
      {facts?.length ? <KeyValue facts={facts} /> : null}
      {meaning && <p className="text-sm text-ink-muted">{meaning}</p>}
      <div className="grid gap-3 rounded-lg border-2 border-primary bg-primary-soft p-4">
        <h2 className="flex items-center gap-2 text-title">
          <PenLine className="size-5 text-primary" aria-hidden />
          {t.title}
        </h2>
        <PinBoxes value={pin} onChange={setPin} disabled={busy} />
        <p className="text-small text-ink-muted">{t.help}</p>
      </div>
      {blocked && <Notice kind="blocking">{blocked}</Notice>}
      {error && <Notice kind="blocking">{error}</Notice>}
      <div>
        <Button
          onClick={() => onSign(pin)}
          disabled={!ready || busy || !!blocked}
        >
          {busy && <LoaderCircle className="animate-spin" aria-hidden />}
          {busy ? t.signing : button}
        </Button>
      </div>
    </section>
  );
}
