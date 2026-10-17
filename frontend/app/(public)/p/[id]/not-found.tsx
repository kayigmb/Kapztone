import { QrCode } from "lucide-react";

export default function NotFound() {
  return (
    <main className="mx-auto grid max-w-content justify-items-center gap-2 px-4 py-16 text-center">
      <span className="grid size-14 place-items-center rounded-full bg-surface-sunken text-ink-muted">
        <QrCode className="size-7" aria-hidden />
      </span>
      <h1 className="text-title">We can&apos;t find this lot code</h1>
      <p className="max-w-[48ch] text-ink-muted">
        Check the code on the label and scan it again. If the label is damaged, ask the holder for the lot code.
      </p>
    </main>
  );
}
