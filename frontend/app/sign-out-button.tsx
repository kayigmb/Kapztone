"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { LogOut } from "lucide-react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { endpoints, problemOf } from "@/lib/api";
import { copy, routes } from "@/lib/consts";

// The page cannot delete the session cookie itself: only the API can clear it.
export function SignOutButton() {
  const router = useRouter();
  const client = useQueryClient();
  const signOut = useMutation({
    mutationFn: endpoints.logout,
    onSuccess: () => {
      client.clear(); // nothing cached for this person may outlive their session
      router.replace(routes.signin);
    },
    onError: (err) => toast.error(problemOf(err)),
  });
  return (
    <Button variant="secondary" onClick={() => signOut.mutate()} disabled={signOut.isPending}>
      <LogOut aria-hidden />
      {copy.signOut}
    </Button>
  );
}
