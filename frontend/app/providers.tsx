"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState } from "react";

import { Toaster } from "@/components/ui/sonner";
import { ApiError } from "@/lib/api";

// Retrying can't fix a refusal (4xx); it can fix a dropped connection or a server hiccup.
const retry = (count: number, err: unknown) =>
  count < 2 && !(err instanceof ApiError && err.status >= 400 && err.status < 500);

export function Providers({ children }: { children: React.ReactNode }) {
  const [client] = useState(() => new QueryClient({
    defaultOptions: { queries: { retry, staleTime: 30_000 }, mutations: { retry: false } },
  }));
  return (
    <QueryClientProvider client={client}>
      {children}
      <Toaster position="top-center" />
    </QueryClientProvider>
  );
}
