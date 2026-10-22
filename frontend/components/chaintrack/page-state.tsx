"use client";

import { Suspense } from "react";
import { RefreshCw, type LucideIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { copy } from "@/lib/consts";
import { cn } from "@/lib/utils";
import Image from "next/image";

/** A spinner and what is happening, while a page's data loads. */
export function Loading({
  children = copy.loading,
  className,
}: {
  children?: React.ReactNode;
  className?: string;
}) {
  return (
    <p
      aria-busy="true"
      className={cn("flex items-center gap-3 text-ink-muted", className)}
    >
      <span
        className="size-5 animate-spin rounded-full border-[2.5px] border-border border-t-primary motion-reduce:animate-none"
        aria-hidden
      />
      {children}
    </p>
  );
}

/** Loading with the logo */
export function LoadingWithLogo() {
  return (
    <div className="flex min-h-screen w-full items-center justify-center">
      <div className="flex flex-col items-center">
        <Image
          src="/brand/chaintrack-mark.svg"
          alt="Logo"
          width={80}
          height={80}
        />
        <Loading className="p-4 lg:p-8" />
      </div>
    </div>
  );
}

/** The one Suspense wrapper for every `page.tsx`: pages read search params on the client inside `<Suspense>`, and this keeps the one fallback (the page frame's padding) in a single place. */
export function PageSuspense({ children }: { children: React.ReactNode }) {
  return <Suspense fallback={<LoadingWithLogo />}>{children}</Suspense>;
}

/** A whole-page outcome with no content to show: not found, unreachable, not allowed. */
export function PageState({
  icon: Icon,
  title,
  children,
  tone = "neutral",
  retry,
}: {
  icon: LucideIcon;
  title: string;
  children: React.ReactNode;
  tone?: "neutral" | "danger";
  retry?: () => void;
}) {
  return (
    <div
      role={tone === "danger" ? "alert" : undefined}
      className="mx-auto grid max-w-content justify-items-center gap-2 px-4 py-16 text-center"
    >
      <span
        className={cn(
          "grid size-14 place-items-center rounded-full",
          tone === "danger"
            ? "bg-danger-soft text-danger"
            : "bg-surface-sunken text-ink-muted",
        )}
      >
        <Icon className="size-7" aria-hidden />
      </span>
      <h1 className="text-title">{title}</h1>
      <p className="max-w-[48ch] text-ink-muted">{children}</p>
      {retry && (
        <Button onClick={retry} className="mt-2">
          <RefreshCw aria-hidden />
          {copy.errors.tryAgain}
        </Button>
      )}
    </div>
  );
}
