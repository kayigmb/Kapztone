"use client";

import { ArrowRight, Package, QrCode } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { marketing as m } from "@/lib/consts";
import { routes } from "@/lib/consts";
import { Container } from "./shared";

export function VerifyStrip() {
  const router = useRouter();
  const [code, setCode] = useState("");
  const v = m.verify;
  return (
    <section
      id={m.anchors.verify}
      className="border-b border-border bg-surface"
    >
      <Container className="grid gap-6 py-10 lg:grid-cols-2 lg:items-center">
        <div className="flex items-start gap-4">
          <span className="grid size-12 shrink-0 place-items-center rounded-lg bg-primary-soft text-primary">
            <QrCode className="size-6" aria-hidden />
          </span>
          <div className="grid gap-1">
            <h2 className="text-xl font-semibold">{v.title}</h2>
            <p className="text-sm text-ink-muted">{v.body}</p>
          </div>
        </div>
        <form
          className="flex gap-2"
          onSubmit={(e) => {
            e.preventDefault();
            if (code.trim())
              router.push(routes.passport(encodeURIComponent(code.trim())));
          }}
        >
          <label className="relative flex-1">
            <span className="sr-only">{v.label}</span>
            <Package
              className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-ink-subtle"
              aria-hidden
            />
            <Input
              value={code}
              onChange={(e) => setCode(e.target.value)}
              placeholder={v.placeholder}
              className="pl-9 font-mono"
              autoCapitalize="characters"
            />
          </label>
          <Button type="submit">
            {v.submit}
            <ArrowRight aria-hidden />
          </Button>
        </form>
      </Container>
    </section>
  );
}
