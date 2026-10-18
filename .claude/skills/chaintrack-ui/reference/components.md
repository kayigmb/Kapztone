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

The stand-alone verdict panel (used in the signed-in app). On the public page the verdict is the navy hero instead (`patterns.md` → Public verification page): a round mark `size-12 lg:size-[58px]` in the verdict's solid tone with an 8 px `ring-hero-line`, the eyebrow "Verification result" in `hero-verified` / `hero-warning` / `hero-danger`, the word `text-[34px] lg:text-[44px] font-semibold`, the summary `text-hero-muted`, and failing or warning reasons as icon + sentence lines. Always `role="status"`.

Panel: `rounded-lg border p-4 pt-6 md:p-6` on the verdict's soft ground; mark `size-14 rounded-full` solid; kicker `label`; word `text-verdict md:text-verdict-lg`; up to four reasons, failures first.

### SignatureChip / StatusChip (Figma `.status-chip`, `.pill`)

A pill: `inline-flex min-h-[26px] w-fit items-center gap-1 rounded-full border py-[3px] pl-2 pr-2.5 text-xs font-bold`, tinted by tone (`tone[t].soft` + a 30% border), icon `size-3.5`. Always an icon and a word: "Signed by both parties" (`CheckCheck`, verified), "Awaiting receiver" (`Clock`, pending), "Rejected" (`OctagonX`, danger). A neutral pill (commodity name) is `bg-surface-sunken text-ink-muted` without an icon.

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

### JourneyTimeline (Figma `.timeline`)

`ol`; each step `grid grid-cols-[32px_1fr] gap-x-3 pb-6`, a 1 px `bg-border` rail between markers. Marker `size-8 rounded-full border` with the event icon `size-4`: done `border-primary/25 bg-primary-soft text-primary`, current `bg-primary text-on-primary`, concern `bg-warning-soft text-warning-ink`, revoked `bg-danger text-on-danger` (and the step's body on `bg-danger-soft`). Body: date eyebrow `text-xs font-semibold uppercase tracking-[0.06em] text-ink-muted`, title `text-[15px] font-semibold`, parties `text-sm text-ink-muted` ("From → **To**"), then the quantity `fig text-sm rounded-sm border bg-bg px-2` and, for a completed handover only, the "Signed by both parties" pill.

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

### ChecksList ("What we checked") (Figma `.checks-list`)

Rows `grid grid-cols-[28px_1fr] gap-x-3 border-t py-3`: a solid round icon `size-7` in the result's tone (`CircleCheck` / `TriangleAlert` / `OctagonX`), the rule `text-sm font-semibold` with the result word right (`text-xs font-bold`, tone ink), the explanation `text-xs text-ink-muted`. The card head carries "n of 7 passed" as a pill.

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

### StatTile (Figma `.metric-card`)

`grid min-h-28 content-start gap-2 rounded-lg border bg-surface p-4 shadow-card lg:p-5`, a link when it leads somewhere (`hover:border-primary`). Top row: label `text-small font-semibold text-ink-muted` and an icon square `size-8 rounded-md bg-primary-soft text-primary`. Value `text-[28px] leading-8 font-semibold tracking-tight tabular-nums`; note `text-xs font-medium text-ink-muted`.
Attention (something for you to do): border `border-pending/40`, icon square and note in pending. Alert (open flags > 0): `border-danger/40`, value `text-danger`, note `text-danger-ink`. Grid `grid-cols-2 gap-2 sm:gap-3 lg:grid-cols-4`.

### Button (Figma `.primary-button`, `.outline-button`, `.danger-button`)

| Variant | Classes | Use |
| --- | --- | --- |
| primary | `bg-primary text-on-primary border-primary shadow-button hover:bg-primary-hover` | The one main action on a screen (teal) |
| secondary | `bg-surface text-ink border-border hover:border-primary hover:text-primary` | Everything else: outlined |
| ghost | `bg-transparent text-primary border-transparent hover:bg-primary-soft` | Low-emphasis navigation |
| danger-soft | `bg-danger-soft text-danger-ink border-transparent` | Refusals that can be undone: "Reject", "Reject & raise flag" |
| danger-outline | `bg-surface text-danger border-danger` | A refusal that needs more weight |
| danger | `bg-danger text-on-danger border-danger` | Irreversible confirmations only |
| disabled | `bg-surface-sunken text-ink-subtle border-border` | Use `aria-disabled` + a visible reason nearby |

Base: `inline-flex items-center justify-center gap-2 min-h-tap px-4 rounded-md border text-sm font-semibold`, icons 18 px. Size `scan`: `h-tap-scan px-6 text-base`. `block`: `w-full`. Labels are verbs naming the result ("Accept and sign", "Issue passport"), never "Submit"/"OK".

### Form fields

- Field `grid gap-1.5`: label `text-sm font-semibold` (`htmlFor`), control, then help `text-small text-ink-subtle` **or** error.
- Input (Figma `.field-group > input`): `w-full min-h-tap px-3 py-2.5 rounded-md border border-border-strong/60 bg-surface text-ink text-sm`, focus `border-primary ring-3 ring-primary/15`. Mono inputs (quantities, codes) `font-mono text-[15px]`.
- Unit addon: input `rounded-r-none` + `px-3 grid place-items-center border border-l-0 border-border-strong rounded-r-md bg-surface-sunken text-ink-muted font-mono text-sm` ("kg").
- Error: input `border-2 border-danger` + `aria-invalid` + `aria-describedby`; message `flex gap-1.5 text-small font-semibold text-danger` with `OctagonX size-4`, saying what's wrong **and** what to enter: "More than the 960 kg this passport holds. Enter 960 kg or less."
- Select: same box as input with `ChevronDown` right. Date fields show `Calendar` left.
- Choice cards (radio-like, for facility, organisation type, receiver): `grid grid-cols-[auto_1fr_auto] gap-x-3 p-3 min-h-tap-scan rounded-md border border-border-strong bg-surface`; icon box `size-9`; title `text-[15px] font-semibold`; sub `text-small text-ink-muted`; radio mark `size-[22px] rounded-full border-2`. Selected: `border-2 border-primary bg-primary-soft`, icon box `bg-primary text-on-primary`, mark `border-[7px] border-primary`. Disabled: `bg-surface-sunken border-dashed`, title muted, sub says why ("Licence expired 30 Jun 2026. Renew to issue.").
- Checkbox: `size-[22px] rounded-sm border-2 border-border-strong`; checked `bg-primary border-primary` with white `Check`; "done/attached" (evidence) `bg-verified border-verified`.
- Segmented control: `inline-flex p-[3px] bg-surface-sunken border rounded-md gap-0.5`; items `min-h-[38px] px-3 rounded-sm text-sm`; active `bg-surface text-ink font-semibold shadow-card`.

### Stepper (Figma `.flow-stepper`)

`ol.flex.items-center.gap-2`; each step a 28 px circle (`size-7 rounded-full border text-xs font-semibold`) and its name `text-xs`, joined by a `h-px flex-1 bg-border` line. Done: `bg-primary-soft text-primary` with `Check`; current: `bg-primary text-on-primary`, name `font-semibold text-ink`; error: `bg-danger text-on-danger`, name `text-danger`; upcoming: `bg-surface text-ink-muted`. On phones only the current step keeps its name. `aria-label="Step 3 of 5"`, `aria-current="step"`.

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

### Notice (inline) (Figma `.notice`)

`flex gap-3 items-start rounded-md border px-4 py-3 text-sm`, the border in the notice's own tone: info `border-pending/25 bg-pending-soft text-pending-ink` (`Info`), success `border-verified/25 bg-verified-soft text-verified-ink` (`CircleCheck`), warning `border-warning/30 bg-warning-soft text-warning-ink` (`TriangleAlert`), blocking `border-danger bg-danger-soft text-danger-ink` (`OctagonX`, `role="alert"`). Lead with a bold sentence when blocking.

### Dialog and Sheet

- Dialog (desktop, destructive confirmations): scrim `bg-black/55`; panel `w-[min(560px,100%)] bg-surface rounded-lg border shadow-overlay`. Head `flex gap-3 p-6 pb-0`: round icon `size-10 bg-danger-soft text-danger` + title `text-[19px] font-semibold` (a question naming the object: "Revoke certificate COO-RW-2026-17320?") + sub. Body `p-6 grid gap-4`: what will be affected (list of lot codes), required reason, public consequence notice. Footer `flex justify-end gap-2 px-6 py-4 border-t bg-surface-sunken`: Cancel + solid danger button naming the exact action ("Confirm fraud and revoke 3 lots").
- Sheet (phones): `fixed inset-x-0 bottom-0 rounded-t-2xl bg-surface shadow-overlay p-4 pb-6 grid gap-3` with a 40×4 grab handle. Same content rules.
- Use Radix Dialog (or shadcn/ui Dialog/Drawer) for focus trapping, restyled with these tokens.

### Toast / push

Push notification text: "**Incoming handover** · Kitabi Washing Station sent you 1,200 kg of red bourbon parchment. Inspect and sign on arrival." In-app toast: `bg-surface border shadow-overlay rounded-lg p-3` with app tile `size-9 rounded-[9px] bg-primary text-on-primary`.

---

## App shell (Figma `BusinessApp`)

See `patterns.md` → App shell for layout. Built in `components/chaintrack/app-shell.tsx`.

- **Sidebar** (`lg:`, 256 px, folds to a 76 px rail; navy `bg-sidebar-bg text-sidebar-fg` in both themes) — Figma `.app-sidebar`:
  - logo row 72 px (`Brand onNavy`: teal square with `Link2` + "ChainTrack") with the fold button;
  - **organisation card** (`.org-switcher`): `rounded-lg border-sidebar-line bg-sidebar-hover p-2.5`, initials square `size-9 rounded-md bg-primary`, name `text-sm font-semibold`, role `text-xs text-sidebar-label`; the status pill under it; "Switch organisation" (`OrgSwitcher onNavy`) when the person has more than one;
  - **nav groups** (`.nav-group`): label `text-xs font-bold uppercase tracking-[0.08em] text-sidebar-label`; items `min-h-tap px-3 rounded-md text-sm text-sidebar-muted hover:bg-sidebar-hover`, active `bg-sidebar-active text-hero-ink font-semibold` with the icon in `text-sidebar-accent-fg`; a live count `font-mono text-xs rounded-full bg-sidebar-hover` (red on the folded rail);
  - **user row** (`.sidebar-user`) at the bottom → `/profile`, then Sign out.
- **Top bar** (`.app-topbar`, 56 / 64 px, `bg-surface/95 backdrop-blur border-b`): phone brand mark, back arrow when not on Home, eyebrow "Workspace" (desktop) over the page title `text-base font-semibold`; right: bell (red count badge), theme toggle (both `size-tap rounded-md border`), "Scan QR" outline button (desktop).
- **Phone tab bar** (`.mobile-tabs`, below `lg`): four 56 px tabs: Home, the person's first workspace screen, Scan, More (a navy sheet with everything else). Active `text-primary font-bold`; counts as red badges on their tab. `pb-[env(safe-area-inset-bottom)]`.
- **Action bar** (focused flows on phones): sticky bottom `bg-surface border-t shadow-overlay px-4 pt-3 pb-4 grid gap-2`, one `h-tap-scan` full-width primary button (or two side by side). On desktop it becomes a right-aligned button row (Figma `.form-actions`).
- **Brand** (`components/chaintrack/brand.tsx`): the mark and the name. `onNavy` on navy grounds, `small` in footers, `nameHidden` for the phone top bar.

## Figma pieces → ChainTrack components

Each Figma class in `ui/figma-make/styles.css` and the piece that builds it. Build the missing ones when their screen is built, from these specs, at the readable minimums.

| Piece | Figma class | Built as | Spec |
| --- | --- | --- | --- |
| Metric card | `.metric-card` | `StatTile` | above |
| Pill | `.pill`, `.status-chip` | `StatusChip` | above |
| Sidebar, nav group, org switcher | `.app-sidebar`, `.nav-group`, `.org-switcher` | `AppShell`, `OrgSwitcher` | above |
| Card + head | `.app-card`, `.app-card-head` | `ListCard` | `rounded-lg border bg-surface shadow-card`; head `min-h-16 px-5 py-3 border-b`, title 15 px semibold, caption `text-xs text-ink-muted`, "View all" link right |
| Quick action | `.quick-actions button` | dashboard `QuickActions` | `min-h-tap-scan grid-cols-[32px_1fr_16px] rounded-lg border px-3`, teal icon square, `ChevronRight` |
| Queue list | `.queue-list`, `.inbox-row`, `.facility-row` | to build | rows `min-h-20 grid-cols-[36px_1fr_auto] gap-3 px-4 border-b`; selected `bg-primary-soft shadow-[inset_3px_0_0_var(--primary)]`; title `text-sm font-semibold`, code `code text-xs`, caption `text-xs text-ink-muted` |
| Split queue | `.split-queue` | to build | `lg:grid-cols-[minmax(275px,0.8fr)_minmax(0,1.5fr)] gap-4`: queue list card left, investigation panel right; one column on phones (list `max-h-72 overflow-y-auto`) |
| Investigation panel | `.investigation`, `.investigation-head`, `.key-values` | to build | head `px-5 py-4 border-b` (pill, title `text-lg`, code); key values `grid sm:grid-cols-3 gap-px bg-border rounded-md border overflow-hidden`, cells `bg-surface p-3` with `text-xs text-ink-muted` key; decision row `.decision-actions` `flex justify-end gap-2 border-t pt-4` |
| Weight compare | `.weight-compare` | to build | `grid grid-cols-[1fr_auto_1fr] items-center gap-4 rounded-lg border bg-bg p-4 text-center`; each side a `text-xs` caption and `fig text-lg` value + unit; "vs" muted; then a Notice with the difference and the tolerance |
| Balance card | `.balance-card`, `.balance-visual` | to build | `grid lg:grid-cols-[1fr_260px_auto] gap-4 rounded-lg border border-primary-line bg-primary-soft p-5`; failed `border-danger/30 bg-danger-soft`; round icon `size-9 bg-primary text-on-primary` (danger when failed); bar `h-2 rounded-full bg-surface` with a teal / red fill; the sign button right |
| Upload zone | `.upload-zone`, `.upload-row` | `FileUpload` (onboarding) | dashed `rounded-md border border-dashed bg-bg min-h-24 grid place-items-center gap-1.5 text-primary`; title `text-sm font-semibold text-ink`; limits `text-xs text-ink-muted` |
| Evidence box, auto checks | `.evidence-box`, `.auto-checks` | to build | `rounded-md border bg-bg p-4`, rows `flex gap-2 text-sm text-ink-muted` with a teal `Check` |
| Quantity field | `.quantity-field` | `TextField` + unit addon | input `font-mono`, unit addon `bg-bg border-l text-xs text-ink-muted min-w-12` |
| Detail banner | `.passport-detail-banner` | passport detail | `rounded-lg bg-navy-900 text-hero-ink p-5 grid lg:grid-cols-[1.6fr_0.7fr_1fr_auto]`; captions `text-xs text-hero-subtle uppercase` |
| Public hero, product card, route art, tree | `.passport-hero`, `.product-card`, `.route-map`, `.tree` | `app/(public)/p/[id]/verification-view.tsx` | see `patterns.md` → Public verification page |
| Access shell | `.auth-shell`, `.auth-brand` | `app/(auth)/layout.tsx` | see `patterns.md` → Access screens |
