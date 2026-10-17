"use client"; // Error boundaries must be Client Components

import { RefreshCw, WifiOff } from "lucide-react";

import { Button } from "@/components/ui/button";

export default function Error({ retry }: { error: Error & { digest?: string }; retry: () => void }) {
  return (
    <main className="mx-auto grid max-w-content justify-items-center gap-2 px-4 py-16 text-center">
      <span className="grid size-14 place-items-center rounded-full bg-danger-soft text-danger">
        <WifiOff className="size-7" aria-hidden />
      </span>
      <h1 className="text-title">We couldn&apos;t check this lot just now</h1>
      <p className="max-w-[48ch] text-ink-muted">
        The records could not be reached. Nothing about the lot has changed. Try again in a moment.
      </p>
      <Button onClick={() => retry()} className="mt-2">
        <RefreshCw aria-hidden />
        Try again
      </Button>
    </main>
  );
}
