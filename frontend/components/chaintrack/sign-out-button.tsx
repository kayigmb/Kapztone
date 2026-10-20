"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { LogOut } from "lucide-react";
import { useRouter } from "next/navigation";

import { useConfirm } from "@/components/chaintrack/confirm";
import { Button } from "@/components/ui/button";
import { endpoints } from "@/lib/api";
import { confirmCopy, copy, routes } from "@/lib/consts";

// The page cannot delete the session cookie itself: only the API can clear it.
export function SignOutButton({ onNavy }: { onNavy?: boolean }) {
  const router = useRouter();
  const client = useQueryClient();
  const signOut = useMutation({
    mutationFn: endpoints.logout,
    onSuccess: () => {
      client.clear(); // nothing cached for this person may outlive their session
      router.replace(routes.signin);
    },
  });
  const confirm = useConfirm();
  const c = confirmCopy.signOut;
  const ask = async () => {
    if (
      await confirm({
        title: c.title,
        body: c.body,
        confirmLabel: c.label,
        icon: LogOut,
      })
    )
      signOut.mutate();
  };
  return (
    <Button
      variant={onNavy ? "ghost" : "secondary"}
      onClick={ask}
      loading={signOut.isPending}
      className={
        onNavy
          ? "justify-start text-sidebar-muted hover:bg-sidebar-hover hover:text-sidebar-fg"
          : undefined
      }
    >
      <LogOut aria-hidden />
      {copy.signOut}
    </Button>
  );
}
