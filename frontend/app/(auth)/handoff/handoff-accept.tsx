"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ArrowLeftRight, TriangleAlert } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useRef } from "react";

import { AuthResult } from "@/components/chaintrack/auth-result";
import { buttonVariants } from "@/components/ui/button";
import { endpoints, problemOf } from "@/lib/api";
import { copy, routes } from "@/lib/consts";

const t = copy.handoff;

/** Spends the switch code once, on arrival, then opens the dashboard here. */
export function HandoffAccept({ code }: { code: string }) {
  const router = useRouter();
  const client = useQueryClient();
  const accept = useMutation({
    mutationFn: endpoints.acceptHandoff,
    meta: { inPlace: true },
    onSuccess: async () => {
      await client.invalidateQueries();
      router.replace(routes.home);
    },
  });
  const sent = useRef(false); // a code works once; dev's double effect must not spend it twice
  useEffect(() => {
    if (sent.current) return;
    sent.current = true;
    window.history.replaceState(null, "", window.location.pathname); // the code leaves the address bar
    accept.mutate(code);
  }, [accept, code]);

  if (accept.isError)
    return (
      <AuthResult
        icon={TriangleAlert}
        tone="warning"
        title={t.failedTitle}
        actions={
          <Link
            href={routes.signin}
            className={buttonVariants({ variant: "primary", block: true })}
          >
            {t.toSignIn}
          </Link>
        }
      >
        <p>{problemOf(accept.error)}</p>
      </AuthResult>
    );
  return (
    <AuthResult icon={ArrowLeftRight} title={t.title}>
      <p>{t.body}</p>
    </AuthResult>
  );
}
