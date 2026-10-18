# Tokens

Source of truth for colour, type, spacing, radius and size. All values live in `globals.css` (a copy is beside this
file); this page says **what each one is for**. The look comes from the Figma Make app in `ui/figma-make/` (reference
only). Never hard-code a hex value, px size or shadow in a component.

## Colour

Utilities follow the token name: `bg-<name>`, `text-<name>`, `border-<name>`, `fill-<name>`, `stroke-<name>`, `ring-<name>`.
Colours switch between light and dark through CSS variables (`data-theme` on `<html>`, or the OS setting), so **don't
write `dark:` colour variants**. Opacity modifiers (`border-danger/30`) are fine.

| Token | Utility example | Light | Dark | Use for |
| --- | --- | --- | --- | --- |
| `bg` | `bg-bg` | `#f5f7f7` | `#081c29` | Page background behind cards. |
| `surface` | `bg-surface` | `#ffffff` | `#102b3a` | Cards, panels, table rows, the top bar, the form side of access screens. |
| `surface-sunken` | `bg-surface-sunken` | `#eef2f3` | `#183542` | Insets: key-value wells, table headers, skeletons, neutral pills (Figma `--soft`). |
| `border` | `border-border` | `#dfe6e8` | `#263f4b` | Hairlines around cards and between rows (Figma `--line`). |
| `border-strong` | `border-border-strong` | `#7e8d94` | `#5d7782` | Control edges that must read at 3:1; inputs use it at 60%. |
| `ink` | `text-ink` | `#14242e` | `#e7eff1` | Primary text. |
| `ink-muted` | `text-ink-muted` | `#5a6971` | `#9caeb5` | Secondary text, labels, captions (Figma `--muted`, darkened to pass 4.5:1). |
| `ink-subtle` | `text-ink-subtle` | `#63727b` | `#8ea1a9` | Metadata and helper text. Not for anything a person must act on. |
| `primary` | `bg-primary` | `#087566` | `#5fd0bd` | **Teal**: the brand, primary buttons, active tab, links, selected nav. |
| `primary-hover` | `bg-primary-hover` | `#065f53` | `#7fdccb` | Hover and pressed fill of primary buttons. |
| `on-primary` | `text-on-primary` | `#ffffff` | `#04211c` | Text and icons on a primary fill. |
| `primary-soft` | `bg-primary-soft` | `#eefaf7` | `#113b3c` | Icon squares, selected rows, done steps (Figma `--teal-50`). |
| `primary-line` | `border-primary-line` | `#d9f3ed` | `#144b49` | Border of a soft teal box (Figma `--teal-100`). |
| `verified` (+ `on-` `-soft` `-ink`) | `text-verified` | `#087566` | `#5fd0bd` | Verified, signed, valid, within limit. Same teal as primary, so **always with an icon and a word**. |
| `warning` (+ `on-` `-soft` `-ink`) | `text-warning` | `#b86d07` | `#f0a43c` | Warning, expiring, near a limit. |
| `danger` (+ `on-` `-soft` `-ink`) | `text-danger` | `#c53737` | `#f47b70` | Failed, altered, rejected, revoked, over a limit, blocking errors. |
| `pending` (+ `-soft` `-ink`) | `text-pending` | `#4a6476` | `#8fb0c6` | Awaiting a signature, in transit, information. |
| `focus` | `ring-focus` | `#0a8f7c` | `#5fd0bd` | Keyboard focus ring: 2px solid, 2px offset (base style). |
| `map-land` `map-water` `map-contour` `map-road` | `fill-map-land` | Figma `.land` / `.route-map` | Figma dark | Map and route art. |
| `scanner-bg` `scanner-ink` | `bg-scanner-bg` | `#071c2c` | `#05121c` | Camera viewport, dark in both themes. |

### Navy, sidebar and hero (the same in both themes)

| Token | Value | Use for |
| --- | --- | --- |
| `navy-950` `navy-900` `navy-800` | `#071c2c` `#0b2639` `#12364c` | The public hero, the access brand panel, the help card, dark banners. |
| `sidebar-bg` | `#092535` (dark `#061823`) | The signed-in sidebar and the phone "More" sheet. |
| `sidebar-fg` · `sidebar-muted` · `sidebar-label` | `#dce8eb` · `#a9bec6` · `#8fa7b2` | Text on the sidebar: names, nav items, group labels and captions. |
| `sidebar-line` · `sidebar-hover` · `sidebar-active` | white 9% · 6% · teal 20% | Dividers, hovered and active nav items. |
| `sidebar-accent-fg` | `#5bd3c0` | The active nav item's icon, the switcher link. |
| `hero-ink` · `hero-muted` · `hero-subtle` | `#ffffff` · `#b9cbd1` · `#9fb4bc` | Text on navy. |
| `hero-accent` | `#7fe0cf` | Eyebrows on navy. |
| `hero-verified` · `hero-warning` · `hero-danger` | `#7fe0cf` · `#f6c06e` · `#f9a39b` | Status on navy, always with an icon and a word. |
| `hero-line` · `hero-glow` | white 12% · teal 22% | Lines and rings on navy; the hero's glow. |
| `art-leaf` `art-leaf-deep` `art-bean` `art-bean-deep` | greens and browns | The product art on the public page only (`product-art`, `bean` utilities). |

### Status colour sets

Each status has four tokens. Use them together, never mix sets. `tone` in `components/chaintrack/tone.ts` holds them.

| Meaning | Icon/text on surface | Ground (soft fill) | Text on ground | Text on solid fill |
| --- | --- | --- | --- | --- |
| Verified · signed · valid · within limit | `verified` | `verified-soft` | `verified-ink` | `on-verified` |
| Warning · expiring · near a limit | `warning` | `warning-soft` | `warning-ink` | `on-warning` |
| Failed · altered · rejected · revoked · over a limit | `danger` | `danger-soft` | `danger-ink` | `on-danger` |
| Awaiting a signature · in transit · info | `pending` | `pending-soft` | `pending-ink` | — |
| Brand · primary action · selected | `primary` | `primary-soft` | `primary` / `ink` | `on-primary` |

## Type

Font: **Inter** (`font-sans`, `--font-inter`) for the interface, **DM Mono** (`font-mono`, `--font-dm-mono`) for lot codes,
IDs, quantities and hashes. Both load with `next/font/google` in `app/layout.tsx`.

**Minimums:** nothing below **12 px** (`text-xs`). Body text 14 px (`text-sm`); captions and labels 12 px. The Figma file
uses 7–10 px in places; those become 12 px here.

| Utility | Size / line | Weight | Use for |
| --- | --- | --- | --- |
| hero verdict `text-[34px] lg:text-[44px]` | 34 / 44 | 600, −0.03em | The verdict word on the public hero. Once per screen. |
| `text-page` (phones `text-[22px]`) | 24/30 | 600 | Screen titles; dashboard greeting `lg:text-[26px]` |
| `text-section` | 18/24 | 600 | Large section headings |
| `text-title` / `text-[15px] font-semibold` | 16/22 · 15/20 | 600 | Card titles (Figma `.app-card-head h2`) |
| `text-body` | 15/22 | 400 | Default body (set on `<body>`) |
| `text-sm` | 14/20 | 400–600 | Form text, rows, buttons |
| `text-small` | 13/18 | 400 | Metadata in dense lists |
| `text-xs` | 12/16 | 400–700 | Captions, pills, eyebrows: the floor |
| `eyebrow` (helper) | 12 caps, +0.1em | 700 | "Verification result", "Workspace", "Welcome back" |
| `label` (helper) | 12 caps, +0.04em | 600 | Key-value keys |
| `code` (helper) | mono 14/20 | 500 | Lot codes `ACME-2026-0412`, IDs, licence numbers |
| `fig` (helper) | mono | 500 | Quantities `1,200 kg`, percentages |
| `text-hash` + `font-mono` | 12/18 | 400 | Fingerprints, **only** inside Technical details |

## Spacing (Tailwind's default 4 px scale)

Phone gutter 12–16 px (`px-3 sm:px-4`), desktop 28 px (`lg:px-7`). Card padding 20 px (`p-5`), 24 px on desktop
forms. Gaps between cards 16 px (`gap-4`), 20 px on desktop (`lg:gap-5`).

## Radius

`rounded-sm` (4) chips' inner parts · `rounded-md` (8) buttons, inputs, icon squares · `rounded-lg` (12) cards, sheets,
maps · `rounded-full` pills, avatars, meters.

## Size

`min-h-tap` / `size-tap` (44) every button, row, input and tab · `h-tap-scan` / `min-h-tap-scan` (56) scanning and
handover actions, quick actions, the phone tab bar and the primary action on phones · `size-8` (32) icon squares ·
`size-[18px]` icons in buttons and nav · `max-w-content` (1180) page content.

## Elevation

`shadow-card` (Figma `--shadow`): every card. `shadow-overlay`: sheets, dialogs, menus, the phone tab bar, floating map
labels. `shadow-button`: the primary button's soft teal shadow (none in dark). The hero glow, grid and product art are
the only gradients, and only on the public page.

## Breakpoints

Phone first (390 px). `sm` (640) two-column bits inside cards. `md` (768) journey + route art side by side. `lg` (1024)
the sidebar replaces the phone tab bar; the public page and access screens go two-column.
