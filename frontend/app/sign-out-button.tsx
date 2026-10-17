"use client";

import { LogOut } from "lucide-react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";

// The page cannot delete the session cookie itself: only the API can clear it.
export function SignOutButton() {
  const router = useRouter();
  async function signOut() {
    await fetch("/api/auth/logout", { method: "POST" });
    router.replace("/signin");
    router.refresh();
  }
  return (
    <Button variant="secondary" onClick={signOut}>
      <LogOut aria-hidden />
      Sign out
    </Button>
  );
}
