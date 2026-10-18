"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ThemeProvider } from "next-themes";
import { useState } from "react";

import { Toaster } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { ApiError } from "@/lib/api";

// Retrying can't fix a refusal (4xx); it can fix a dropped connection or a server hiccup.
const retry = (count: number, err: unknown) =>
  count < 2 && !(err instanceof ApiError && err.status >= 400 && err.status < 500);

export function Providers({ children }: { children: React.ReactNode }) {
  const [client] = useState(() => new QueryClient({
    defaultOptions: { queries: { retry, staleTime: 30_000 }, mutations: { retry: false } },
  }));
  return (
    // globals.css switches tokens on <html data-theme>; "system" leaves it to the device.
    <ThemeProvider attribute="data-theme" defaultTheme="system" enableSystem disableTransitionOnChange>
      <QueryClientProvider client={client}>
        <TooltipProvider>
          {children}
        </TooltipProvider>
        <Toaster position="top-center" />
      </QueryClientProvider>
    </ThemeProvider>
  );
}
