<!-- BEGIN:nextjs-agent-rules -->

# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->

# ChainTrack frontend

The web app for ChainTrack: digital product passports for business shipments. Three parts:

- **Public verification** `/p/[id]`: anyone scanning a QR label; `id` is a lot code (`ACME-2026-0412`) or a passport uuid.
- **Sign-up and onboarding** `/signup`, `/signin`, `/onboarding`, `/invite/[token]`: no shell.
- **Signed-in app**: everything under `app/(app)/`, inside the shell (sidebar, top bar, phone tab bar).

For visual design, load the `chaintrack-ui` skill before touching any page, component, style or copy. This file covers how the code is put together.

## Commands

| Command | What it does |
| --- | --- |
| `make web` (repo root) | Dev server on :3000 with the repo `.env` (needs `API_URL`) |
| `pnpm lint` | ESLint, must be clean |
| `pnpm build` | Production build, must pass |
| `node --test "lib/**/*.test.ts"` | Unit tests of the pure helpers |

The browser never calls the API directly. Every call goes to `/api/...`, and `app/api/[...path]/route.ts` passes it to `API_URL` with the session cookie, `Origin`, `X-Forwarded-Host/For` and `X-Signing-PIN`. That's what keeps the session cookie on this host only.

## Where things live

| Path | Holds |
| --- | --- |
| `app/(public)/`, `app/(auth)/` | Pages without the shell |
| `app/(app)/layout.tsx` | The shell plus the guard: no session → `/signin`; unverified organisation → `/onboarding` |
| `app/(app)/<screen>/` | One folder per screen: `page.tsx` + its client component(s) |
| `app/providers.tsx` | TanStack Query, next-themes (`data-theme`), tooltips, toaster |
| `components/ui/` | shadcn/ui primitives (Base UI). Owned here, restyled to ChainTrack |
| `components/chaintrack/` | Domain components: StatTile, CapacityMeter, PassportCard, JourneyTimeline, VerdictBadge, AppShell, ScanButton… |
| `lib/api.ts` | Every API call (`endpoints.*`), its zod schema, `ApiError`, `problemOf`, `useMe`, `useMeGuard` |
| `lib/consts.ts` | Every status, kind, permission, route, API path, query key, nav item, dashboard layout and all UI copy |
| `lib/schemas/*.ts` | Form schemas (zod), mirroring the backend's `Validate()` |
| `lib/dashboard.ts`, `lib/address.ts`, `lib/verification.ts` | Pure helpers, each with a `*.test.ts` beside it |

## Rules

1. **Data:** every call is an `endpoints.*` function in `lib/api.ts` that parses the answer with a zod schema. Components read with `useQuery` and write with `useMutation`, then invalidate the query keys the write changed. Never `fetch` in a component.
2. **Errors:** every failure is an `ApiError`. Show `problemOf(err)` in a Sonner toast, or in a `Notice` / `PageState` when the page can't go on. A 400 about one field becomes that field's error.
3. **Copy and values:** no user-facing string, API path, query key, status or permission name is typed in a component. Add it to `lib/consts.ts` and import it.
4. **Forms:** react-hook-form + zod (`zodResolver`), with fields built from `TextField` (`type="password"` gets the show / hide button). The schema lives in `lib/schemas/`, and the API stays the authority.
5. **Pages:** `page.tsx` only exports `metadata` and mounts a client component. Pages read search params on the client, inside `<Suspense>`.
6. **Access:** permissions come from `GET /me` (`permissions.*` in consts). The nav (`navItems`) lists **only built screens**, and quick actions follow the same rule. A new screen adds its line.
7. **Logic:** anything with a branch (who sees what, address building, parsing) goes in a pure `lib/*.ts` helper with a node test. Pure files import only with relative paths and `.ts` extensions, so `node --test` can load them.
8. **Signing:** actions that sign (issue, handover start / accept / reject, transform, retire, certificate issue / revoke, flag resolve) send the person's 6-digit PIN as `X-Signing-PIN`. Follow `patterns.md` → Signing: summary first, then the PIN, then a button that names the result.
9. **Data display:** lot codes are shown in full, in mono. Quantities always carry their unit. Statuses are an icon plus a word.
10. **Organisation address:** the acting organisation is the first label of the host (`acme.localhost:3000`). Sessions are one per address, so switching organisation means signing in there.

## Visual rules (summary; the skill has the detail)

- Theme tokens only (`bg-surface`, `text-ink-muted`…). No hex, no default Tailwind palette, no `dark:` colour variants: tokens switch with `data-theme`.
- Status colours mean status only. One primary button per screen.
- Text at least 12 px (`text-xs`), body 14 px. Tap targets at least 44 px (`min-h-tap`); scan and handover actions 56 px (`h-tap-scan`).
- Teal is the primary colour and also means verified, so a status is always an icon and a word. The hero glow, grid and product art are for the public page only.
- Phone first (390 px), then `lg:`. Check light and dark.
- No blockchain words (blockchain, on-chain, wallet, token, hash…) outside the collapsed Technical details.

## Adding a screen

1. Read its spec in `.claude/skills/chaintrack-ui/reference/screens.md`, and the Claude Design draft (below).
2. Check which backend endpoints exist (`backend/internal/server/server.go` route table). Anything the design shows that the API can't give goes in the gap table in `docs/design.md`; don't fake it.
3. Add schemas and `endpoints.*` to `lib/api.ts`.
4. Add paths, query keys, copy and the nav item to `lib/consts.ts`.
5. Build `app/(app)/<screen>/page.tsx` plus its client component, reusing `components/chaintrack/` and `components/ui/`.
6. Put logic in a pure helper with a test.
7. Run `pnpm lint`, `pnpm build` and the node tests.
8. Update `docs/design.md` (screens table, gap rows) and add an entry to `CHANGELOG.md`.

## Never

- Run git commands. The learner handles git.
- Add another component or UI library (MUI, Chakra, Mantine…), or a library for what a few lines do.
- Put an edit button on a signed record. Corrections are new, linked records.
- Use `localStorage` for anything but per-browser preferences (sidebar folded, theme), and always inside try/catch.
- Show "Signed by both parties" before the receiver has signed.

## Design source

Import done. Two references, both read-only (never imported by the app; their demo names and codes are never copied):

- **Content and flows:** `ui/ChainTrack Drafts.dc.html` (with `support.js`), the Claude Design drafts, ids `1a`–`3o`. Original: https://claude.ai/design/p/c9f04332-afb9-4bbe-8b97-2c8686eed87e?file=ChainTrack+Drafts.dc.html
- **Look:** `ui/figma-make/App.tsx` and `ui/figma-make/styles.css`, the Figma Make app. Its tokens, fonts and component styles are in `app/globals.css` and the `chaintrack-ui` skill.

| Draft | Route | Figma component |
| --- | --- | --- |
| 1a, 1b, 1c | `/p/[id]` | `PublicVerification` |
| 1d, 3o | `/` | `Dashboard` |
| 2a, 2b, 2c | `/signin` | `AuthFlow` (sign in) |
| 2d–2h | `/forgot`, `/reset/[token]`, `/lock/[token]` | `AuthFlow` |
| 2i | `/invite/[token]` | `AuthFlow` |
| 2j, 2k, 2l | `/signup`, `/onboarding` | `AuthFlow` (sign up, documents, pending) |
| 2m | signing sheet (`components/chaintrack/signing.tsx`), `/profile` | `Profile` |
| 3a, 3b | `/facilities`, `/facilities/new` | `Facilities`, `RegisterFacility` |
| 3c, 3d | `/issue` | `IssuePassport` |
| 3e | `/passports/[id]` | `Passports` (detail) |
| 3f | `/scan` | `Scanner` |
| 3g, 1f, 3h | `/handover`, `/handovers`, `/handovers/[id]` | `Handovers` |
| 1e | `/transform` | `Transform` |
| 3i, 3j | `/certificates`, `/certificates/new` | `Certificates` |
| 3k | `/flags` | `Flags` |
| 3l | `/admin/templates` | `Templates` |
| 3m | `/admin/organisations` | `VerificationQueue` |
| 3n | `/admin/reports` | `Reports` |

Where they differ, and what we did:

- **Sizes:** the Figma file sets text at 7–10 px and buttons at 38–40 px. Here text is at least 12 px and targets at least 44 px (56 px for scan, handover, quick actions and the phone tab bar).
- **Colour:** the drafts use navy as the primary colour; the Figma look uses teal. Teal won: it is the brand and primary colour, and verified is the same teal, so verified always carries an icon and a word.
- **Decoration:** the hero glow, grid and product art appear on the public page only.
- **Phone tabs:** the drafts have five tabs with a raised Scan button; Figma has four tabs. We use four: Home, the person's first workspace screen, Scan, More.
- **Fields with no API behind them** (language, phone, country and registration number, flag severity, the verification-outcomes report, the inspection checklist, a certificate number): hidden; listed in `docs/design.md`.
