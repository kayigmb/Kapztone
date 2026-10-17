// Shown while the records are being checked: skeleton blocks shaped like the verdict and the passport card.
export default function Loading() {
  return (
    <div aria-busy="true" className="mx-auto grid max-w-content gap-4 px-4 py-4 md:gap-6 md:py-8">
      <p className="flex items-center gap-3 text-ink-muted">
        <span className="size-5 animate-spin rounded-full border-[2.5px] border-border border-t-primary" aria-hidden />
        Checking this lot&apos;s records…
      </p>
      <div className="grid gap-4 md:grid-cols-[1.55fr_1fr] md:gap-6">
        <div className="h-56 rounded-lg bg-surface-sunken" />
        <div className="h-56 rounded-lg bg-surface-sunken" />
        <div className="h-40 rounded-lg bg-surface-sunken md:col-span-2" />
      </div>
    </div>
  );
}
