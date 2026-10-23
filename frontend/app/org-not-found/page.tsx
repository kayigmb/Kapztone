import { Building2 } from "lucide-react";

import { AuthResult } from "@/components/chaintrack/auth-result";
import { Brand } from "@/components/chaintrack/brand";
import { buttonVariants } from "@/components/ui/button";
import { copy } from "@/lib/consts";

const t = copy.orgNotFound;

export const metadata = { title: `${t.title} · ${copy.brand}` };

/** Shown (by proxy.ts) at an address that names no organisation. */
export default async function Page({
  searchParams,
}: {
  searchParams: Promise<{ host?: string; home?: string }>;
}) {
  const { host = "", home = "/" } = await searchParams;
  return (
    <div className="grid min-h-dvh place-items-center bg-surface px-5 py-12">
      <div className="grid w-full max-w-100 gap-10">
        <Brand />
        <AuthResult
          icon={Building2}
          tone="warning"
          title={t.title}
          actions={
            <a href={home} className={buttonVariants({ block: true })}>
              {t.action}
            </a>
          }
        >
          <p>{t.body(host)}</p>
        </AuthResult>
      </div>
    </div>
  );
}
