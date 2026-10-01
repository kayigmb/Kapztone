# Components

Each entry gives: what it's for, props, anatomy with Tailwind classes, states, and rules.
Domain components go in `components/chaintrack/`, primitives in `components/ui/`. Shared types:

```ts
type Verdict = 'verified' | 'warning' | 'failed';
type Tone = 'verified' | 'warning' | 'danger' | 'pending' | 'neutral';
type SignatureState = 'signed-both' | 'signed-sender' | 'awaiting-receiver' | 'awaiting-you' | 'discrepancy' | 'rejected' | 'altered' | 'draft';
type CertificateState = 'valid' | 'expiring' | 'expired' | 'revoked';
```

Tone → token set (use all four together):

```ts
const tone = {
  verified: { solid: 'bg-verified text-on-verified', soft: 'bg-verified-soft text-verified-ink', fg: 'text-verified', border: 'border-verified' },
  warning:  { solid: 'bg-warning text-on-warning',   soft: 'bg-warning-soft text-warning-ink',   fg: 'text-warning',  border: 'border-warning' },
  danger:   { solid: 'bg-danger text-on-danger',     soft: 'bg-danger-soft text-danger-ink',     fg: 'text-danger',   border: 'border-danger' },
  pending:  { solid: 'bg-pending text-on-primary',   soft: 'bg-pending-soft text-pending-ink',   fg: 'text-pending',  border: 'border-pending' },
  neutral:  { solid: 'bg-surface-sunken text-ink',   soft: 'bg-surface-sunken text-ink-muted',   fg: 'text-ink-muted', border: 'border-border' },
} as const;
```

---

## Status

### VerdictBadge

The result of a passport check. Full size once at the top of the public verification page; `compact` in dashboards, scan results and flag queues.

```ts
type VerdictBadgeProps = {
  verdict: Verdict;
  summary: string;                    // one sentence; for warning/failed it ends with what to do
  reasons: { result: 'pass' | 'warn' | 'fail'; text: string; detail?: string }[];  // ≤ 4, fails first, then warns, then passes
  meta?: React.ReactNode;             // "Checked 2 Oct 2026, 09:14 CAT · 15 records · 7 organisations"
  compact?: boolean;
};
```

- Panel: `rounded-lg border p-4 pt-6 md:p-6 md:pt-8` + soft ground of the verdict (`bg-verified-soft border-verified/35`, etc.). `role="status"`.
- Head row `flex items-center gap-4`: mark = `size-14 md:size-16 rounded-full grid place-items-center` solid fill (`bg-verified text-on-verified`) with `ShieldCheck` / `ShieldAlert` / `ShieldX` at `size-8`. Next to it: kicker `label` in the `-ink` colour ("Product passport check") and the word in `text-verdict md:text-verdict-lg` in the `-ink` colour: **Verified / Warning / Failed**.
- Summary `mt-4 text-[16px] leading-6 text-ink max-w-[60ch]`.
- Reasons `mt-4 grid gap-2 md:grid-cols-2 md:gap-x-6`: each `flex gap-2 items-start`; icon `CircleCheck` (`text-verified`), `TriangleAlert` (`text-warning`), `OctagonX` (`text-danger`) at `size-5`. Warn/fail text `font-semibold`; detail under it in `text-small text-ink-muted`.
- Meta `mt-4 pt-3 border-t border-current/20 text-small text-ink-muted flex flex-wrap gap-x-3 gap-y-1`.
- Compact: `p-3 px-4`, mark `size-9`, word `text-section`.

### SignatureChip / StatusChip

Where a handover or record stands in the dual-signature flow. Also used for "within limit" chips.

| State | Classes | Icon | Text |
| --- | --- | --- | --- |
| signed-both | `bg-verified-soft text-verified-ink border-verified/30` | `CheckCheck` | Signed by both parties |
| signed-sender | same as signed | `PenLine` | Signed by sender |
| awaiting-receiver | `bg-pending-soft text-pending-ink border-pending/30` | `Clock` | Awaiting receiver signature |
| awaiting-you | pending | `PenLine` | Awaiting your signature |
| discrepancy | `bg-warning-soft text-warning-ink border-warning/35` | `TriangleAlert` | Quantity differs by 40 kg |
| rejected | `bg-danger-soft text-danger-ink border-danger/35` | `X` | Rejected by receiver |
| altered | danger | `OctagonX` | Record altered after signing |
| draft / not trusted | `bg-surface-sunken text-ink-muted border-border` | `PenLine` / `TriangleAlert` | Draft, not yet signed / Not trusted |

Anatomy: `inline-flex items-center gap-1 min-h-[26px] pl-2 pr-2.5 py-[3px] rounded-full border text-small font-semibold whitespace-nowrap`, icon `size-4`. Accept a custom `children` label (e.g. "Yield 20.0% (max 21%)" with `Scale`).

### Status (inline)

`inline-flex items-center gap-1.5 text-sm font-semibold` + `text-verified|warning|danger|pending|ink-muted` + icon `size-4`. For table cells, card footers and key-value values ("Valid to 31 Mar 2027").

### CertificateTile

```ts
type CertificateTileProps = { type: string; issuer: string; number: string; state: CertificateState;
  validFrom?: string; validTo?: string; note?: string /* required for expired/revoked: why it matters for THIS lot */ };
```

- `grid grid-cols-[40px_1fr] gap-3 p-4 bg-surface border rounded-md`. Border: `border-border` (valid, expiring), `border-warning` (expired), `border-danger` (revoked).
- Icon box `size-10 rounded-md bg-surface-sunken text-ink-muted` (`FileText`, `Ship` for permits, `Scale` for lab/assay).
- Type `text-[15px] font-semibold`, issuer `text-small text-ink-muted`, number `font-mono text-xs text-ink-subtle`.
- Footer `mt-3 pt-3 border-t border-border flex justify-between flex-wrap gap-2`: dates (`text-small text-ink-muted`, key date `font-semibold text-ink`) and a badge pill in the state's soft colours: Valid (`CircleCheck`), Expiring soon (`Clock`, within 30 days), Expired (`TriangleAlert`, amber), Revoked (`OctagonX`, red).
- `note` spans both columns: `text-small rounded-sm px-3 py-2` in the soft colours.
- Layout in a grid: 1 column on phones, `md:grid-cols-3`.

---

## Passport

### PassportCard

```ts
type PassportCardProps = { product: string; lotCode: string; commodityIcon: LucideIcon;
  facts: { label: string; value: React.ReactNode; wide?: boolean }[];   // ≤ 6
  footer: { left: React.ReactNode; verdict: Verdict }; flagged?: boolean; href?: string };
```

- `bg-surface border rounded-lg shadow-card overflow-hidden`; `flagged` → `border-danger`, icon box red.
- Top `flex gap-3 p-4 border-b border-border`: icon box `size-11 rounded-md bg-primary-soft text-primary`; name `text-title`; lot code `code text-ink-muted`.
- Body `p-4 md:p-6`: KeyValue grid (below). Quantities in `fig`.
- Footer `flex justify-between items-center px-4 py-3 bg-surface-sunken text-small text-ink-muted`: "Passport issued 3 Jun 2026" + inline Status of the verdict.

### KeyValue

`<dl class="grid grid-cols-2 gap-4">`; each `<div>`: `<dt class="label mb-0.5">` + `<dd class="text-body break-words">`. `wide` items `col-span-2`. On desktop inside wide cards: `md:grid-cols-3`.

### JourneyTimeline

The ordered history of a lot. One row per event, origin first on the public page.

```ts
type JourneyStep = {
  event: 'issued' | 'handed-over' | 'processed' | 'exported' | 'imported' | 'delivered' | string; // display label, e.g. "Processed: washed and dried"
  icon: LucideIcon; date: string; from?: string; to?: string; org?: string; place: string;
  quantity: string;                       // "6,000 kg cherry → 1,200 kg parchment"
  chips: React.ReactNode[];               // a SignatureChip for every handover; a yield chip for every process step
  state?: 'done' | 'current' | 'pending' | 'warning' | 'failed' | 'untrusted';
  compare?: { signed: string; now: string; note: string };   // only for failed (altered) steps
};
```

- `<ol>`; each `<li class="relative grid grid-cols-[40px_1fr] gap-x-3 pb-6 last:pb-0">`. Rail: absolute `left-[19px] top-10 bottom-0 w-0.5 bg-border-strong` (hidden on last).
- Marker `size-10 rounded-full grid place-items-center border-2 bg-surface relative z-10` with event icon `size-5`:
  done `border-verified text-verified` · current `bg-primary border-primary text-on-primary` · pending `border-dashed border-pending text-pending` · warning `border-warning text-warning` · failed `bg-danger border-danger text-on-danger` · untrusted `border-dashed border-border-strong text-ink-subtle`.
- Body: row `flex justify-between flex-wrap gap-x-3`: event `text-title` + date `text-small text-ink-subtle tabular-nums`. Then org line `text-body`: "Sender → **Receiver**" (receiver `font-semibold`). Place `text-small text-ink-muted` with `MapPin size-4`. Meta row `mt-2 flex flex-wrap gap-2`: quantity pill `fig text-sm bg-surface-sunken rounded-sm px-2 py-0.5` + chips.
- Failed body: wrapped in `bg-danger-soft rounded-md p-3`, event in `text-danger-ink`, plus compare: two boxes side by side (`grid grid-cols-2 gap-2`), "Signed by both" value vs. "Record now says" value (`border-danger`, value `text-danger`), then the note in `text-small text-danger-ink`.
- Untrusted rows (after a failed record): event and org in `text-ink-muted`, chip "Not trusted: built on the altered record".

### GenealogyTree

Which lots a lot was made from (merge, split, transform), back to every origin.

```ts
type LotNode = { lotCode: string; description: string; quantity: string; icon: LucideIcon;
  kind?: 'this' | 'origin' | 'failed'; relation?: 'Made from' | 'Merged from' | 'Split from'; status?: React.ReactNode; children?: LotNode[] };
```

- Nested `<ul>`; child lists `ml-3.5 md:ml-7`, each child `relative pl-4 md:pl-7 pt-3` with connector lines: vertical `absolute left-0 top-0 bottom-0 w-0.5 bg-border-strong` (last child stops at its node's middle) and horizontal `absolute left-0 top-[37px] w-3.5 md:w-6.5 h-0.5 bg-border-strong`.
- Relation label above the first child of a group: `label text-ink-subtle mb-1`.
- Node `grid grid-cols-[32px_1fr] md:grid-cols-[32px_1fr_auto] items-center gap-3 px-3 py-2 min-h-13 bg-surface border border-border-strong rounded-md max-w-[460px]`. Icon box `size-8 rounded-sm bg-surface-sunken text-ink-muted`. Lot `font-mono text-[13px] font-medium whitespace-nowrap`; description `text-small text-ink-muted`. Side (quantity `fig text-[13px]` + status): its own row under the text on phones (`col-start-2`), right column on desktop.
- `this`: `border-2 border-primary bg-primary-soft`, icon box `bg-primary text-on-primary`. `origin`: icon box `bg-verified-soft text-verified`, status "Origin" ✓. `failed`: `border-2 border-danger bg-danger-soft`, icon box solid red. Material with no passport is shown as its own failed node ("No passport · Unaccounted material · 750 kg").

### FacilityMap

A facility's registered boundary polygon and/or pin. **Use a real map library** (MapLibre GL or Leaflet via `react-map-gl` / `react-leaflet`, client component, `dynamic(..., { ssr: false })`).

- Container `relative rounded-lg overflow-hidden border border-border`, aspect 5:3.
- Layers, styled from CSS variables (read with `getComputedStyle(document.documentElement).getPropertyValue('--verified')` and re-read on theme change):
  registered boundary: fill `--verified` at 18% opacity, outline `--verified` 2.5 px · drawing: fill `--primary` 12%, outline `--primary` dashed 6/4, vertex handles white with `--primary` 2 px ring · flagged / licence expired: outline `--danger` dashed, fill 18%.
- Pin: `MapPin`-shaped marker in `--primary`.
- Legend bar bottom-left: `bg-surface/90 border border-border rounded-sm px-2.5 py-1.5 text-xs text-ink-muted` naming each shape ("Registered boundary"). Bottom-right: coordinates in mono and scale.
- Prefer a muted basemap (light grey land, pale blue water) for light and a dark navy basemap for dark.

### RouteMap

Where a lot travelled between journey steps.

- Origin: teal dot. Intermediate stops: white circles with the **step number** matching the JourneyTimeline. Current position: navy dot. Stop with a failed record: red dot.
- Road legs solid `--primary` 2.5 px; sea legs dotted; legs that depend on a failed record dashed `--ink-subtle`.
- Caption "Schematic, not to navigate" in the card head; list the numbered stops in text under the map (`ol`, number in `font-mono font-semibold text-primary`).

### ChecksList ("What we checked")

One row per rule: `grid grid-cols-[24px_1fr_auto] gap-x-3 py-3 border-t border-border first:border-0 first:pt-0`. Icon (pass/warn/fail colours), rule name `font-semibold`, result Status (Passed / Warning / Failed), then explanation `col-span-2 col-start-2 text-small text-ink-muted` with the real numbers.

Standard rules, in this order: Origin is registered and licensed · Within declared capacity · No forest loss on the plots (agricultural only) · Every handover signed by both parties · Quantities balance · Records unchanged since signing · Certificates valid · Organisations verified.

### TechnicalDetails

`<details>` collapsed by default; **open by default when the verdict is Failed**. Summary row `min-h-tap px-4 py-3 flex items-center gap-2 font-semibold` with `Lock` icon and `ChevronDown` rotating 180° when open. Body: one sentence explaining the ledger in plain words, then rows `grid grid-cols-[1fr_auto]`: event name, "View ↗" link, hash in `font-mono text-hash text-ink-muted break-all`. A mismatched hash row uses `text-danger`. This is the only place hashes appear.

---

## Measures

### Meter (MassBalanceMeter, CapacityMeter)

```ts
type MeterProps = { title: string; value: number; max: number;      // max = allowed maximum (input × yield, or declared capacity)
  pending?: number;                                                  // capacity: quantity being entered now (lighter segment)
  unit: 'kg' | string; scaleMax?: number; message: React.ReactNode; formula?: React.ReactNode };
// state: within (< 90% of max for capacity, < 98% for mass balance) | near | over (value > max)
```

- Head `flex justify-between flex-wrap`: title `font-semibold` + value `fig` ("**1,200** / 1,260 kg").
- Track `relative h-3.5 mt-8 mb-2 rounded-sm bg-surface-sunken ring-1 ring-inset ring-border-strong`. Fill `absolute inset-y-0 left-0 rounded-l-sm` in `bg-verified` / `bg-warning` (near) / `bg-danger` (over). Pending segment same colour at 45% opacity. Overshoot beyond max: hatched `repeating-linear-gradient(135deg, var(--danger) 0 4px, var(--danger-soft) 4px 7px)`. Limit tick `absolute -inset-y-1.5 w-0.5 bg-ink` with label "Max 1,260" above it (`text-[11px] font-semibold`; right-aligned when near the end).
- Scale row `flex justify-between font-mono text-xs text-ink-subtle`.
- Message `mt-3 flex gap-2 text-sm font-semibold` in the state colour with its icon, plain text after it: "Within limit · 60 kg below the allowed maximum" / "Close to the limit · 15 kg left" / **over**: a `bg-danger-soft text-danger-ink rounded-md p-3` block, `role="alert"`: "**Outputs exceed the maximum by 120 kg.** You can't create these lots. Lower the outputs or add another input passport."
- Formula `mt-2 text-small text-ink-muted`: "Input `6,000 kg` × allowed yield `21%` = `1,260 kg` maximum output".
- `role="meter"` with `aria-valuenow`, `aria-valuemax`, `aria-label`.
- Capacity variant message: "You have **1,200 kg** left this season." Over: "**300 kg over the declared capacity.** You can issue at most 1,200 kg from this facility in Season B 2026."

---

## Data and forms

### DataTable

- Wrapper `bg-surface border border-border rounded-lg overflow-hidden` (+ `overflow-x-auto` for wide admin tables).
- Optional toolbar `flex justify-between flex-wrap gap-2 px-4 py-3 border-b border-border`: title, filters, export.
- `th`: `label bg-surface-sunken px-4 py-3 text-left whitespace-nowrap border-b border-border`. `td`: `px-4 py-3 border-b border-border text-sm`. Row hover `hover:bg-primary-soft/55`.
- Lot code cells `font-mono font-medium whitespace-nowrap`; number cells `text-right fig whitespace-nowrap`.
- Flagged row: `bg-danger-soft` + `shadow-[inset_3px_0_0_var(--danger)]` on the first cell, plus its chip.
- Footer `flex justify-between px-4 py-3 border-t text-small text-ink-muted`: "4 of 23" · sort order.
- **Below `md`:** don't scroll sideways. Render each row as a two-column card (`grid grid-cols-[1fr_auto] gap-x-3 px-4 py-3 border-b`): lot + product left, quantity + status right, secondary cells labelled inline.

### ListRow

The workhorse for queues (handovers, flags, passports) on every screen size.

- `grid grid-cols-[40px_1fr_auto] gap-x-3 gap-y-1 items-center px-4 py-3 min-h-16 border-t border-border first:border-0`, link rows `hover:bg-primary-soft/70`; selected `bg-primary-soft/70 shadow-[inset_3px_0_0_var(--primary)]`.
- Icon box `size-10 rounded-md` tinted by tone (`bg-pending-soft text-pending`, …). Title `text-[15px] font-semibold truncate`, lot code `font-mono text-xs text-ink-muted`, sub `text-small text-ink-muted`. Right: quantity `fig text-sm` and/or a chip. Extra chips row `col-start-2 col-span-2 flex flex-wrap gap-2`.
- In a list card: `bg-surface border rounded-lg overflow-hidden` with header `flex justify-between px-4 py-3 border-b` (icon + `text-title`, "All 14" link).

### StatTile

`bg-surface border border-border rounded-lg px-4 py-3 lg:px-6 lg:py-4 min-h-24 grid gap-0.5`, link. Label `text-small text-ink-muted font-medium` with icon `size-4`; value `text-[28px] leading-[34px] font-semibold tabular-nums`; note `text-xs text-ink-subtle`.
Attention variant (something for you to do): `bg-pending-soft border-pending/45`, all text `text-pending-ink`. Alert variant (open flags > 0): `bg-danger-soft border-danger`, text `text-danger-ink`. Grid: `grid-cols-2 gap-3 lg:grid-cols-4 lg:gap-4`.

### Button

| Variant | Classes | Use |
| --- | --- | --- |
| primary | `bg-primary text-on-primary border-primary hover:bg-primary-hover` | The one main action on a screen |
| secondary | `bg-surface text-ink border-border-strong hover:bg-surface-sunken` | Everything else |
| ghost | `bg-transparent text-primary border-transparent hover:bg-primary-soft` | Low-emphasis navigation |
| danger-outline | `bg-surface text-danger border-danger` | Reject / raise flag |
| danger | `bg-danger text-on-danger border-danger` | Irreversible confirmations only |
| disabled | `bg-surface-sunken text-ink-subtle border-border cursor-not-allowed` | Use `aria-disabled` + a visible reason nearby |

Base: `inline-flex items-center justify-center gap-2 min-h-tap px-4 rounded-md border font-semibold text-[15px]`. Size `scan`: `h-tap-scan px-6 text-base`. `block`: `w-full`. Icons `size-5`. Labels are verbs naming the result ("Accept and sign", "Issue passport"), never "Submit"/"OK".

### Form fields

- Field `grid gap-1.5`: label `text-sm font-semibold` (`htmlFor`), control, then help `text-small text-ink-subtle` **or** error.
- Input `w-full min-h-tap px-3 py-2.5 rounded-md border border-border-strong bg-surface text-ink text-base` (16 px stops iOS zoom). Mono inputs (quantities, codes) `font-mono text-[15px]`.
- Unit addon: input `rounded-r-none` + `px-3 grid place-items-center border border-l-0 border-border-strong rounded-r-md bg-surface-sunken text-ink-muted font-mono text-sm` ("kg").
- Error: input `border-2 border-danger` + `aria-invalid` + `aria-describedby`; message `flex gap-1.5 text-small font-semibold text-danger` with `OctagonX size-4`, saying what's wrong **and** what to enter: "More than the 960 kg this passport holds. Enter 960 kg or less."
- Select: same box as input with `ChevronDown` right. Date fields show `Calendar` left.
- Choice cards (radio-like, for facility, organisation type, receiver): `grid grid-cols-[auto_1fr_auto] gap-x-3 p-3 min-h-tap-scan rounded-md border border-border-strong bg-surface`; icon box `size-9`; title `text-[15px] font-semibold`; sub `text-small text-ink-muted`; radio mark `size-[22px] rounded-full border-2`. Selected: `border-2 border-primary bg-primary-soft`, icon box `bg-primary text-on-primary`, mark `border-[7px] border-primary`. Disabled: `bg-surface-sunken border-dashed`, title muted, sub says why ("Licence expired 30 Jun 2026. Renew to issue.").
- Checkbox: `size-[22px] rounded-sm border-2 border-border-strong`; checked `bg-primary border-primary` with white `Check`; "done/attached" (evidence) `bg-verified border-verified`.
- Segmented control: `inline-flex p-[3px] bg-surface-sunken border rounded-md gap-0.5`; items `min-h-[38px] px-3 rounded-sm text-sm`; active `bg-surface text-ink font-semibold shadow-card`.

### Stepper

`flex gap-1`; each step `flex-1 grid gap-1.5`: bar `h-1 rounded-sm` + name `text-xs truncate`. Done `bg-verified` + `Check` + `text-ink-muted`; current `bg-primary` + `text-ink font-semibold`; error `bg-danger` + `text-danger font-semibold`; upcoming `bg-border` + `text-ink-subtle`. `aria-label="Step 3 of 5"`.

### FileUpload

- Drop zone `grid justify-items-center text-center gap-1 px-4 py-6 border-2 border-dashed border-border-strong rounded-lg bg-surface text-sm text-ink-muted`, `Upload` icon `size-7 text-primary`, title `font-semibold text-ink`, "Tap to take a photo or choose a file", limits `text-xs`. On phones, open the camera (`capture="environment"` for photos).
- File row `grid grid-cols-[40px_1fr_auto] gap-3 p-3 border rounded-md`: extension box (`font-mono text-[11px] font-semibold`), name `truncate font-semibold text-sm`, meta `text-xs text-ink-muted`, right status (Uploaded ✓ / 64% / Unreadable ✕); progress bar `h-1.5 rounded bg-surface-sunken` with `bg-primary` fill.
- Photo grid `grid grid-cols-4 gap-2`, square thumbs `rounded-md border`, tag bottom-left (`text-[10px] font-semibold bg-black/75 text-white px-1.5 rounded-[3px]`), "Add photo" tile dashed with `Camera`.

### QRScannerFrame

Client component using the camera (e.g. `@zxing/browser` or `html5-qrcode`).

- Viewport `relative rounded-lg overflow-hidden bg-scanner-bg text-scanner-ink` (dark in **both** themes), video fills it.
- Instruction at top `text-center font-semibold text-[15px]`: "Point at the QR label on the bag or crate".
- Window 64–70% wide, square, four corner brackets `border-4 border-scanner-ink` (only two sides each, rounded outer corner 12 px), thin scan line. Found: brackets turn `border-verified`, vibrate once (`navigator.vibrate?.(30)`).
- Tools at the bottom, `h-tap-scan` buttons on a translucent dark fill: **Light** (`Zap`, torch) and **Type code** (`Keyboard`, manual lot-code entry). Never hide manual entry.
- Result sheet slides up inside the viewport: white card with commodity icon, product name, lot code (mono, no wrap) and verdict Status.
- Camera blocked → ErrorState with "Type code" as the primary action.

### QR label (printable)

Print layout, always light: white background, `#0f1a2c` ink, 2 px border. Grid `[120px_1fr]`: QR (real QR encoding the public passport URL), then "ChainTrack · Product passport" (bold 13 px), product name, lot code (mono 14 px semibold), meta ("900 kg · Plot 14 · Issued 2 Oct 2026"); dashed footer with short URL and "Scan to verify". Print size 50 × 30 mm; provide a `@media print` stylesheet.

### Empty, loading and error states

- Empty: centered `grid justify-items-center text-center gap-2 py-8 px-4`: icon in `size-14 rounded-full bg-surface-sunken text-ink-muted`, title `text-title` that says what would be here, one sentence on how it gets filled, one action. "No handovers waiting for you · When someone hands a lot to your organisation, it appears here for you to inspect and sign · [Scan a QR label]".
- Loading: say what is being checked ("Checking 15 records…") with a spinner (`size-5 rounded-full border-[2.5px] border-border border-t-primary animate-spin`) plus skeleton blocks shaped like the content (`bg-surface-sunken rounded-sm`). `aria-busy="true"`.
- Error: icon in `bg-danger-soft text-danger`; say what failed, whether anything was lost, and what happens next. Offline is common at mines and ports: queue signatures locally and say "Your signature is saved on this phone and will be sent when you're back online. Nothing has been lost." Action: "Try again" (`RefreshCw`).
- No illustrations, no jokes.

### Notice (inline)

`flex gap-3 items-start px-4 py-3 rounded-md text-sm`: info `bg-pending-soft text-pending-ink` (`Info`), success `bg-verified-soft text-verified-ink` (`CircleCheck`), warning `bg-warning-soft text-warning-ink` (`TriangleAlert`), blocking `bg-danger-soft text-danger-ink border border-danger` (`OctagonX`, `role="alert"`). Lead with a bold sentence when blocking.

### Dialog and Sheet

- Dialog (desktop, destructive confirmations): scrim `bg-black/55`; panel `w-[min(560px,100%)] bg-surface rounded-lg border shadow-overlay`. Head `flex gap-3 p-6 pb-0`: round icon `size-10 bg-danger-soft text-danger` + title `text-[19px] font-semibold` (a question naming the object: "Revoke certificate COO-RW-2026-17320?") + sub. Body `p-6 grid gap-4`: what will be affected (list of lot codes), required reason, public consequence notice. Footer `flex justify-end gap-2 px-6 py-4 border-t bg-surface-sunken`: Cancel + solid danger button naming the exact action ("Confirm fraud and revoke 3 lots").
- Sheet (phones): `fixed inset-x-0 bottom-0 rounded-t-2xl bg-surface shadow-overlay p-4 pb-6 grid gap-3` with a 40×4 grab handle. Same content rules.
- Use Radix Dialog (or shadcn/ui Dialog/Drawer) for focus trapping, restyled with these tokens.

### Toast / push

Push notification text: "**Incoming handover** · Kitabi Washing Station sent you 1,200 kg of red bourbon parchment. Inspect and sign on arrival." In-app toast: `bg-surface border shadow-overlay rounded-lg p-3` with app tile `size-9 rounded-[9px] bg-primary text-on-primary`.

---

## App shell

See `patterns.md` → App shell for layout. Pieces:

- **Sidebar** (`lg:` and up, 248 px): wordmark "ChainTrack" (`font-bold text-[17px]`), organisation card (avatar square + legal name + "Role · Job title"), nav items `flex items-center gap-3 min-h-10 px-2 rounded-md text-sm font-medium text-ink-muted hover:bg-surface-sunken`, active `bg-primary-soft text-primary font-semibold`; counts on the right `font-mono text-xs bg-surface-sunken rounded-full px-2` (red `bg-danger-soft text-danger-ink` when it needs action); signed-in person at the bottom.
- **Top bar** (desktop 64 px / phone 56 px): back arrow when nested, title (`text-[17px] font-semibold`), search (desktop only, 360 px), notifications bell with red count badge.
- **Bottom tab bar** (below `lg`): 5 slots: Home, Passports, **Scan** (raised 60 px round `bg-primary` button in the centre), Handovers, More. Regulators: Home, Flags, Scan, Certificates, More. Active tab `text-primary font-bold`. Sticky, `pb-[env(safe-area-inset-bottom)]`.
- **Action bar** (focused flows on phones): sticky bottom `bg-surface border-t shadow-overlay px-4 pt-3 pb-4 grid gap-2`, one `h-tap-scan` full-width primary button (or two side by side). On desktop it becomes a right-aligned button row in the page.
- **Wordmark:** there is no logo yet. Set "ChainTrack" in IBM Plex Sans Bold. Don't draw a mark.
