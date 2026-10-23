"use client";

import {
  MutationCache,
  QueryClient,
  QueryClientProvider,
} from "@tanstack/react-query";
import { ThemeProvider } from "next-themes";
import { NuqsAdapter } from "nuqs/adapters/next/app";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import { ConfirmProvider } from "@/components/chaintrack/confirm";
import { Toaster } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { ApiError, isSent, problemOf } from "@/lib/api";
import { copy } from "@/lib/consts";

// Retrying can't fix a refusal (4xx); it can fix a dropped connection or a server hiccup.
const retry = (count: number, err: unknown) =>
  count < 2 &&
  !(err instanceof ApiError && err.status >= 400 && err.status < 500);

export function Providers({ children }: { children: React.ReactNode }) {
  const [client] = useState(
    () =>
      new QueryClient({
        // One toast path: a write's message comes from the API; a failure says what went wrong, unless the screen shows it in place (`meta: { inPlace: true }`).
        mutationCache: new MutationCache({
          onSuccess: (data) => {
            if (isSent(data) && data.message) toast.success(data.message);
          },
          onError: (err, _v, _c, m) => {
            if (!m.meta?.inPlace) toast.error(problemOf(err));
          },
        }),
        defaultOptions: {
          queries: { retry, staleTime: 30_000 },
          mutations: { retry: false },
        },
      }),
  );
  useEffect(() => {
    const offline = () =>
      toast.error(copy.offline, { id: "offline", duration: Infinity });
    const online = () => {
      toast.dismiss("offline");
      toast.success(copy.online);
    };
    if (!navigator.onLine) offline();
    window.addEventListener("offline", offline);
    window.addEventListener("online", online);
    return () => {
      window.removeEventListener("offline", offline);
      window.removeEventListener("online", online);
    };
  }, []);
  return (
    // globals.css switches tokens on <html data-theme>; "system" leaves it to the device.
    <ThemeProvider
      attribute="data-theme"
      defaultTheme="system"
      enableSystem
      disableTransitionOnChange
    >
      <QueryClientProvider client={client}>
        <NuqsAdapter>
          <TooltipProvider>
            <ConfirmProvider>{children}</ConfirmProvider>
          </TooltipProvider>
        </NuqsAdapter>
        <Toaster position="top-center" />
      </QueryClientProvider>
    </ThemeProvider>
  );
}
