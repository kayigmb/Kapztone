---
name: chaintrack-ui
description: ChainTrack's design system for the Next.js + Tailwind frontend. Use for ANY frontend work — building or changing a page, component, form, layout, style, copy or icon — so the UI matches the ChainTrack design.
---

# ChainTrack UI

ChainTrack gives every shipment a digital product passport: where goods came from, who handled them, and proof that nothing changed after it was signed. Users are businesses moving goods across borders (producers, processors, transporters, exporters, importers, certifiers, regulators, a platform admin). Consumers and buyers only see the public verification page they reach by scanning a QR label.

**It must feel like a customs, banking or logistics control tool: trustworthy, calm, precise. It is NOT a crypto app.** No coins, wallets, neon, gradients or "Web3" styling. The blockchain stays invisible.

## Before you build anything

1. Find the screen you're building in `reference/screens.md` and follow its section order, content and states.
2. Use the components in `reference/components.md`. If one exists, build or reuse it; don't invent a variant.
3. For flows (forms, signing, errors, destructive actions, the app shell), follow `reference/patterns.md`.
4. For copy, follow `reference/content.md`. For icons, `reference/icons.md`. For token meaning, `reference/tokens.md`.
   Read only the files you need for the task at hand.

## Setup (once per project)

- Copy `reference/globals.css` to `app/globals.css`. It holds the full Tailwind v4 theme: colours (light + dark via CSS variables), type scale, radii, sizes and shadows. **Tailwind v3 project?** Port the same variables into `tailwind.config` `theme.extend` with `colors: { surface: 'var(--surface)', … }`.
- Fonts in `app/layout.tsx`:
  ```tsx
  import { IBM_Plex_Sans, IBM_Plex_Mono } from "next/font/google";
  const sans = IBM_Plex_Sans({
    subsets: ["latin"],
    weight: ["400", "500", "600", "700"],
    variable: "--font-plex-sans",
  });
  const mono = IBM_Plex_Mono({
    subsets: ["latin"],
    weight: ["400", "500", "600"],
    variable: "--font-plex-mono",
  });
  // <html lang="en" className={`${sans.variable} ${mono.variable}`}>
  ```
- Icons: `lucide-react`, mapped in `reference/icons.md`. Default `strokeWidth={2}`, `aria-hidden` when next to text.
- **UI primitives: shadcn/ui** (Base UI underneath). Add components with `npx shadcn add button dialog drawer select checkbox tabs tooltip sonner form table`. They go in `components/ui/` and you own and edit them.
  - When `shadcn init` touches `app/globals.css`, keep ChainTrack's theme: delete any colour variables shadcn adds (`--background`, `--primary`, … in `:root`/`.dark`) and keep the "shadcn/ui colour names → ChainTrack tokens" block. shadcn classes like `bg-background`, `text-muted-foreground` and `ring-ring` then follow ChainTrack in light and dark.
  - Restyle each added component to match `reference/components.md` (heights, radii, variants). For example, Button gets `min-h-tap`, `rounded-md`, and the variants primary / secondary / ghost / danger-outline / danger.
  - Don't use MUI, Chakra, Mantine or any other styled component library.
- **Domain components** (VerdictBadge, PassportCard, JourneyTimeline, GenealogyTree, meters, CertificateTile, SignatureChip…) are custom, in `components/chaintrack/`, built on the primitives above. Use `cn()` from shadcn for class merging.
- **Other libraries:** react-hook-form + zod (forms), TanStack Table (data tables), MapLibre GL via `react-map-gl` (maps), `qrcode` (labels), `@zxing/browser` (scanning), Sonner (toasts).
- **Data, errors and copy (project rules):**
  - **TanStack Query** for every API call: `useQuery` to read, `useMutation` to write, then invalidate the keys the write changed. The provider is `app/providers.tsx`. Calls live in `lib/api.ts` (`endpoints.*`) and go through the `/api/...` pass-through. Pages render on the client; a `page.tsx` only exports `metadata` and mounts the client component.
  - **zod for API errors and answers**: every response is parsed by a schema in `lib/api.ts`; every failure is an `ApiError`. Show `problemOf(err)` in a Sonner toast, or in a blocking `Notice` when the page can't continue. A 400 about one field becomes that field's error.
  - **`lib/consts.ts`** holds every status, kind, permission name, route, API path, query key and all UI copy. Never type a user-facing string or an API path in a component; add it there and import it.
  - Signed-in guards use `useMeGuard()` from `lib/api.ts`; permission checks use `permissions.*` from consts.

## Hard rules (never break these)

1. **Only theme tokens.** Use `bg-surface`, `text-ink-muted`, `border-border-strong`, `p-4`, `rounded-lg`, etc. No hex values, no arbitrary colours (`bg-[#…]`), no Tailwind default palette (`text-red-500`; it is disabled in the theme). No `dark:` colour variants: tokens already switch.
2. **Status = icon + word, never colour alone.** Verified → `CircleCheck` + "Verified"; Warning → `TriangleAlert` + "Warning"; Failed → `OctagonX` + "Failed"; Awaiting → `Clock` + "Awaiting …".
3. **Status colours only mean status.** Teal `verified` = verified/signed/valid/within limit. Amber `warning` = warning/expiring/near limit. Red `danger` = failed/altered/rejected/revoked/over limit. Slate `pending` = awaiting/in transit. Navy `primary` = brand and primary actions. Never use them decoratively.
4. **Hide the blockchain.** Say "Verified", "Signed by both parties", "Records unchanged since signing". Hashes, block numbers, ledger links and the words blockchain/crypto/wallet/token/on-chain/smart contract appear **only** inside the collapsed "Technical details" disclosure.
5. **Mono for data.** Lot codes, passport IDs, licence and certificate numbers, quantities and hashes use `font-mono` (`code`, `fig` helpers). Never truncate a lot code.
6. **Tap targets.** Every interactive element ≥ 44 px (`min-h-tap`). Scanning, handover and phone primary actions are 56 px (`h-tap-scan`).
7. **Phone first.** Design at 390 px, then add `md:`/`lg:`. Side gutter 16 px on phones. No horizontal page scroll; tables collapse to stacked rows.
8. **Two signatures.** A handover is complete only when sender **and** receiver have signed. Never show "Signed by both parties" before that.
9. **Signed records are never edited.** No edit buttons on signed records; corrections create a new linked record.
10. **Blocking errors block.** Capacity exceeded or mass balance exceeded: field, step and meter turn red, the message says by how much and the largest allowed value, and the primary button is disabled with a reason.
11. **Light and dark both work.** Check both before you're done.
12. **Accessible by default.** Real `<button>`/`<a>`, labels tied to inputs, `aria-invalid` + error text on bad fields, `role="status"` on verdicts, `role="alert"` on blocking messages, visible focus ring (from globals.css), `prefers-reduced-motion` respected.

## Look and feel in one glance

- Background `bg-bg`; cards `bg-surface border border-border rounded-lg shadow-card p-4 lg:p-6`; insets and table headers `bg-surface-sunken`.
- Headings semibold, never bold-black. Labels small caps-style (`label`). Generous whitespace: `gap-4` between cards on phones, `gap-6` on desktop, `gap-8` between sections.
- One primary button (navy) per screen. Secondary buttons are outlined. Red solid buttons only for irreversible confirmations ("Confirm fraud", "Revoke").
- Motion is minimal: spinners, a sheet sliding up, a chevron rotating. No bouncing, no confetti.
- No emoji, no illustrations of people, no stock photos, no cards with a coloured left border.

## Done checklist

- [ ] Matches the screen spec's section order and states (empty, loading, error included)
- [ ] Only token utilities; no hex, no default palette, no `dark:` colour overrides
- [ ] Built on shadcn/ui primitives from `components/ui/`, restyled to ChainTrack; no other component library
- [ ] Every status has an icon and a word
- [ ] Lot codes and quantities in mono; units always shown (`1,200 kg`)
- [ ] Works at 390 px and 1280 px, in light and dark
- [ ] Tap targets ≥ 44 px (56 px on scan/handover)
- [ ] Copy follows `content.md` (verbs on buttons, numbers with units, no blockchain words outside Technical details)
