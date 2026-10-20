"use client";

import { useMutation } from "@tanstack/react-query";
import { Lock, OctagonX, TriangleAlert } from "lucide-react";
import Link from "next/link";
import { toast } from "sonner";

import { AuthResult } from "@/components/chaintrack/auth-result";
import { Button, buttonVariants } from "@/components/ui/button";
import { ApiError, endpoints, problemOf } from "@/lib/api";
import { copy, routes } from "@/lib/consts";

const t = copy.lock;

// A deliberate button, not on page load: mail scanners open links, and must not lock accounts.
export function LockConfirm({ token }: { token: string }) {
  const lock = useMutation({
    mutationFn: () => endpoints.lockAccount(token),
    meta: { inPlace: true },
    onError: (err) => {
      if (!(err instanceof ApiError && err.status === 404))
        toast.error(problemOf(err));
    },
  });
  const back = (
    <Link
      href={routes.signin}
      className={buttonVariants({ variant: "secondary", block: true })}
    >
      {t.toSignIn}
    </Link>
  );

  if (lock.isSuccess)
    return (
      <AuthResult
        icon={Lock}
        tone="verified"
        title={t.doneTitle}
        actions={back}
      >
        <p>{t.done}</p>
      </AuthResult>
    );
  if (lock.error instanceof ApiError && lock.error.status === 404) {
    return (
      <AuthResult
        icon={TriangleAlert}
        tone="warning"
        title={t.expiredTitle}
        actions={back}
      >
        <p>{t.expired}</p>
      </AuthResult>
    );
  }
  return (
    <AuthResult
      icon={OctagonX}
      tone="danger"
      title={t.title}
      actions={
        <>
          <Button
            variant="danger"
            block
            onClick={() => lock.mutate()}
            loading={lock.isPending}
            loadingText={t.submitting}
          >
            {t.submit}
          </Button>
          {back}
        </>
      }
    >
      <p>{t.body}</p>
    </AuthResult>
  );
}
