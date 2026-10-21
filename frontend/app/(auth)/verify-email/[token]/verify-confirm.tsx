"use client";

import { useMutation } from "@tanstack/react-query";
import { CircleCheck, MailCheck, TriangleAlert } from "lucide-react";
import Link from "next/link";
import { toast } from "sonner";

import { AuthResult } from "@/components/chaintrack/auth-result";
import { Button, buttonVariants } from "@/components/ui/button";
import { ApiError, endpoints, problemOf } from "@/lib/api";
import { copy, routes } from "@/lib/consts";

const t = copy.verifyEmail;

// A deliberate button, not on page load: mail scanners open links, and must not spend them.
export function VerifyConfirm({ token }: { token: string }) {
  const verify = useMutation({
    mutationFn: () => endpoints.verifyEmail(token),
    meta: { inPlace: true },
    onError: (err) => {
      if (!(err instanceof ApiError && err.status === 404))
        toast.error(problemOf(err));
    },
  });
  const signIn = (variant: "primary" | "secondary") => (
    <Link
      href={routes.signin}
      className={buttonVariants({ variant, block: true })}
    >
      {t.toSignIn}
    </Link>
  );

  if (verify.isSuccess)
    return (
      <AuthResult
        icon={CircleCheck}
        tone="verified"
        title={t.doneTitle}
        actions={signIn("primary")}
      >
        <p>{t.done}</p>
      </AuthResult>
    );
  if (verify.error instanceof ApiError && verify.error.status === 404) {
    return (
      <AuthResult
        icon={TriangleAlert}
        tone="warning"
        title={t.expiredTitle}
        actions={signIn("secondary")}
      >
        <p>{t.expired}</p>
      </AuthResult>
    );
  }
  return (
    <AuthResult
      icon={MailCheck}
      title={t.title}
      actions={
        <Button
          block
          onClick={() => verify.mutate()}
          loading={verify.isPending}
          loadingText={t.submitting}
        >
          {t.submit}
        </Button>
      }
    >
      <p>{t.body}</p>
    </AuthResult>
  );
}
