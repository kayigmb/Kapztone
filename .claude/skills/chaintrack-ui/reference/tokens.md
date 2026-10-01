# Tokens

Source of truth for colour, type, spacing, radius and size. All values live in `globals.css`;
this page says **what each one is for**. Never hard-code a hex value, px size or shadow in a component.

## Colour

Utilities follow the token name: `bg-<name>`, `text-<name>`, `border-<name>`, `fill-<name>`, `stroke-<name>`, `ring-<name>`.
Colours switch between light and dark through CSS variables, so **don't write `dark:` colour variants**.

| Token | Utility example | Light | Dark | Use for |
| --- | --- | --- | --- | --- |
| `bg` | `bg-bg` | `#f4f5f7` | `#0a1120` | Page background behind cards and sections. |
| `surface` | `bg-surface` | `#ffffff` | `#111b2e` | Cards, panels, table rows, sheets. |
| `surface-sunken` | `bg-surface-sunken` | `#eceef2` | `#0c1424` | Inset areas: key-value wells, code blocks, table headers, skeletons. |
| `border` | `bg-border` | `#d6dae1` | `#24314a` | Hairlines between rows and around cards. Decorative only: never the sole edge of a control. |
| `border-strong` | `bg-border-strong` | `#7f8898` | `#66748d` | Input and control borders, meter tracks' outline, timeline rail. 3:1 on bg and surface in both themes. |
| `ink` | `bg-ink` | `#0f1a2c` | `#e8edf5` | Primary text on bg, surface, surface-sunken and every *-soft fill. |
| `ink-muted` | `bg-ink-muted` | `#475367` | `#b0bccd` | Secondary text and labels on bg, surface, surface-sunken and the *-soft fills. |
| `ink-subtle` | `bg-ink-subtle` | `#5b677a` | `#97a3b7` | Metadata, timestamps, helper text on bg, surface and surface-sunken. Not for anything a person must act on. |
| `primary` | `bg-primary` | `#15294b` | `#a8c1ec` | Deep navy: primary buttons, active step, selected tab, links on surface. The brand colour. |
| `primary-hover` | `bg-primary-hover` | `#0d1d38` | `#c3d5f3` | Hover and pressed fill of primary buttons. |
| `on-primary` | `bg-on-primary` | `#ffffff` | `#0b1426` | Text and icons on a primary fill. |
| `primary-soft` | `bg-primary-soft` | `#e7ecf5` | `#1b2a47` | Selected row, current timeline step, info wells. Text on it is ink or primary. |
| `verified` | `bg-verified` | `#0a7266` | `#45c4b1` | Teal: Verified, signed, valid, within limits. Icons and text on surface and bg; fill for the verdict mark. |
| `on-verified` | `bg-on-verified` | `#ffffff` | `#06201c` | Icon or text on a solid verified fill. |
| `verified-soft` | `bg-verified-soft` | `#e2f3f0` | `#0f2f2c` | Ground of a Verified verdict, valid certificate tile or signed chip. |
| `verified-ink` | `bg-verified-ink` | `#075a51` | `#7adccc` | Text on verified-soft. |
| `warning` | `bg-warning` | `#b45309` | `#f0a43c` | Amber: Warning, expiring, near a limit, awaiting action. Icons and text on surface and bg. |
| `on-warning` | `bg-on-warning` | `#ffffff` | `#241503` | Icon or text on a solid warning fill. |
| `warning-soft` | `bg-warning-soft` | `#fcf1df` | `#33240e` | Ground of a Warning verdict, expiring or expired certificate, near-limit meter. |
| `warning-ink` | `bg-warning-ink` | `#874006` | `#f6c06e` | Text on warning-soft. |
| `danger` | `bg-danger` | `#b42318` | `#f47b70` | Red: Failed, altered, rejected, revoked, over a limit, blocking errors. Icons and text on surface and bg. |
| `on-danger` | `bg-on-danger` | `#ffffff` | `#2a0b08` | Icon or text on a solid danger fill. |
| `danger-soft` | `bg-danger-soft` | `#fdeceb` | `#3b1714` | Ground of a Failed verdict, flagged rows, blocking messages. |
| `danger-ink` | `bg-danger-ink` | `#8c1c13` | `#f9a39b` | Text on danger-soft. |
| `pending` | `bg-pending` | `#3a5a8c` | `#8fb0e6` | Slate blue: awaiting a signature, in transit, informational. Icons and text on surface and bg. |
| `pending-soft` | `bg-pending-soft` | `#e8eef8` | `#182742` | Ground of awaiting chips and info notices. |
| `pending-ink` | `bg-pending-ink` | `#2c4875` | `#a9c3ef` | Text on pending-soft. |
| `focus` | `bg-focus` | `#1d5bd6` | `#7aa6ff` | Keyboard focus ring: 2px solid, 2px offset. 3:1 on every surface in both themes. |
| `map-land` | `bg-map-land` | `#eceee6` | `#1b2a42` | Map base fill. |
| `map-water` | `bg-map-water` | `#c8dae8` | `#0b1524` | Rivers, lakes and sea on maps. |
| `map-contour` | `bg-map-contour` | `#cdd4c9` | `#22324a` | Contour lines, graticule and minor roads on maps. Decorative. |
| `map-road` | `bg-map-road` | `#ffffff` | `#2b3b57` | Roads on facility maps. |
| `scanner-bg` | `bg-scanner-bg` | `#0b111c` | `#05080f` | Camera viewport behind the QR scanner frame (dark in both themes). |
| `scanner-ink` | `bg-scanner-ink` | `#f2f5fa` | `#f2f5fa` | Text and corner brackets over scanner-bg. |

### Status colour sets

Each status has four tokens. Use them together, never mix sets.

| Meaning | Icon/text on surface | Ground (soft fill) | Text on ground | Text on solid fill |
| --- | --- | --- | --- | --- |
| Verified · signed · valid · within limit | `verified` | `verified-soft` | `verified-ink` | `on-verified` |
| Warning · expiring · near a limit | `warning` | `warning-soft` | `warning-ink` | `on-warning` |
| Failed · altered · rejected · revoked · over a limit | `danger` | `danger-soft` | `danger-ink` | `on-danger` |
| Awaiting a signature · in transit · info | `pending` | `pending-soft` | `pending-ink` | — |
| Brand · primary action · selected | `primary` | `primary-soft` | `primary` / `ink` | `on-primary` |

Contrast is checked: `ink`, `ink-muted` and `ink-subtle` pass 4.5:1 on `bg`, `surface` and `surface-sunken` in both themes;
`ink` and `ink-muted` also pass on every `*-soft` ground. `border-strong` and `focus` pass 3:1. Keep to these pairs.

## Type

Font: **IBM Plex Sans** (`font-sans`) for UI, **IBM Plex Mono** (`font-mono`) for lot codes, IDs, quantities and hashes.
Load both with `next/font/google` (see SKILL.md). Each size utility already sets line height, weight and tracking.

| Utility | Size / line | Weight | Use for |
| --- | --- | --- | --- |
| `text-verdict` (`md:text-verdict-lg`) | 32/38 (40/46) | 600 | The verdict word on the public page. Once per screen. |
| `text-page` | 24/30 (22/28 on phones: `text-[22px] leading-7 md:text-page`) | 600 | Screen titles |
| `text-section` | 18/24 | 600 | Card and section headings |
| `text-title` | 16/22 | 600 | Card titles, product names, timeline events |
| `text-body` | 15/22 | 400 | Default text, form values (inputs use 16px to stop iOS zoom) |
| `text-small` | 13/18 | 400 | Metadata, helper text, dense tables |
| `label` (helper) | 12/16 caps, +0.04em | 600 | Field labels, key-value keys |
| `code` (helper) | mono 14/20 | 500 | Lot codes `CT-RW-COF-26-004817`, IDs, licence numbers |
| `fig` (helper) | mono | 500 | Quantities `1,200 kg`, percentages |
| `text-hash` + `font-mono` | 12/18 | 400 | Transaction hashes, **only** inside Technical details |

## Spacing (Tailwind's default 4 px scale)

| Design token | Tailwind | px | Use for |
| --- | --- | --- | --- |
| space-1 | `1` | 4 | Icon-to-label gap inside chips |
| space-2 | `2` | 8 | Related items: chip rows, key and value |
| space-3 | `3` | 12 | Chip and table-cell padding |
| space-4 | `4` | 16 | Card padding on phones, phone side gutter, gap between cards and fields |
| space-5 | `6` | 24 | Card padding on desktop, gap between cards on desktop |
| space-6 | `8` | 32 | Gap between page sections, desktop gutter |
| space-7 | `12` | 48 | Top and bottom page margin on desktop |

## Radius

`rounded-sm` (4) chips' inner parts, meter tracks · `rounded-md` (6) buttons, inputs, tiles, tree nodes ·
`rounded-lg` (10) cards, sheets, maps, the verdict panel · `rounded-full` status chips and avatars only.

## Size

`min-h-tap` / `h-tap` (44) every button, row and input · `h-tap-scan` (56) buttons on scanning and handover screens and the
primary action on phones · `size-4` (16) icons in chips and tables · `size-5` (20) icons beside text · `size-7` (28) icons in
markers · `max-w-content` (1120) page content on desktop.

## Elevation

Separation comes from **borders** (`border border-border`). `shadow-card` is a whisper for cards on `bg`;
`shadow-overlay` only for sheets, dialogs, menus and sticky action bars. No gradients, glows or blur.

## Breakpoints

Phone first. `md` (768) = public verification page goes two-column. `lg` (1024) = business app shows the sidebar
instead of the bottom tab bar. Content max `max-w-content`.
