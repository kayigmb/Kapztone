<!-- BEGIN:nextjs-agent-rules -->

# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->

# ChainTrack frontend

The web app for ChainTrack: digital product passports for business shipments. Three parts:

- **Public verification** `/p/[id]`: anyone scanning a QR label; `id` is a lot code (`ACME-2026-0412`) or a passport uuid.
- **Sign-up and onboarding** `/signup`, `/signin`, `/onboarding`, `/invite/[token]`: no shell.
- **Signed-in app**: everything under `app/(app)/`, inside the shell (sidebar, top bar, phone tab bar). Home is `/dashboard`; the platform's dashboard adds charts from `GET /stats` (Recharts, `components/chaintrack/charts.tsx`).
- **Marketing home** `/` (`components/marketing/landing.tsx`, copy in `lib/consts.ts` → `marketing`): only claims the product backs.

For visual design, load the `chaintrack-ui` skill before touching any page, component, style or copy. This file covers how the code is put together.

## Commands

| Command                                | What it does                                                                                                                                                          |
| -------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `make web` (repo root)                 | Dev server on :3000 with the repo `.env` (needs `API_URL`)                                                                                                            |
| `pnpm lint`                            | ESLint, must be clean                                                                                                                                                 |
| `pnpm build`                           | Production build, must pass                                                                                                                                           |
| `node --test "__tests__/**/*.test.ts"` | Unit tests of the pure helpers (`__tests__/`, mirrored names; schema tests prefixed `schemas-`)                                                                       |
| `make e2e` (repo root)                 | Playwright happy path (`e2e/coffee.spec.ts`) against the running API and web app; needs `SUPERADMIN_*` in `.env`. First time: `pnpm exec playwright install chromium` |

The browser never calls the API directly. Every call goes to `/api/...`, and `app/api/[...path]/route.ts` passes it to `API_URL` with the session cookie, `Origin`, `X-Forwarded-Host/For` and `X-Signing-PIN`. That's what keeps the session cookie on this host only.

## Where things live

| Path                                                        | Holds                                                                                                          |
| ----------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| `app/(public)/`, `app/(auth)/`                              | Pages without the shell                                                                                        |
| `app/(app)/layout.tsx`                                      | The shell plus the guard: no session → `/signin`; unverified organisation → `/onboarding`                      |
| `app/(app)/<screen>/`                                       | One folder per screen: `page.tsx` + its client component(s)                                                    |
| `app/providers.tsx`                                         | TanStack Query, next-themes (`data-theme`), tooltips, toaster                                                  |
| `components/ui/`                                            | shadcn/ui primitives (Base UI). Owned here, restyled to ChainTrack                                             |
| `components/chaintrack/`                                    | Domain components: StatTile, CapacityMeter, PassportCard, JourneyTimeline, VerdictBadge, AppShell, ScanButton… |
| `components/chaintrack/data-widget.tsx`                     | `DataWidget`: the one pending / error / data branch for every section (`query` + render fn)                    |
| `components/chaintrack/field-note.tsx`                      | `FieldNote`: the one error-or-help line under every field; used inside `TextField`                             |
| `components/chaintrack/page-state.tsx`                      | `Loading`, `LoadingWithLogo`, `PageState`, `PageSuspense`: whole-page wait → `LoadingWithLogo`; section → `DataWidget` |
| `lib/api.ts`                                                | Every API call (`endpoints.*`), its zod schema, `ApiError`, `problemOf`, `useMe`, `useMeGuard`                 |
| `lib/consts.ts`                                             | Every status, kind, permission, route, API path, query key, nav item, dashboard layout; copy re-exported       |
| `lib/copy/*.json`                                           | All UI copy (second-language-ready JSON; `{placeholders}` filled by `fill()` in `lib/format.ts`)               |
| `lib/format.ts`                                             | `fill()` for copy templates, `formatDate`; pure, with `format.test.ts`                                         |
| `lib/schemas/*.ts`                                          | Form schemas (zod), mirroring the backend's `Validate()`                                                       |
| `lib/dashboard.ts`, `lib/address.ts`, `lib/verification.ts` | Pure helpers, each with a test in `__tests__/` (`schemas-` prefix for `lib/schemas/`)                          |

## Rules

0. **Pages and lists:** every `<main>` is `grid w-full content-start gap-6 p-4 lg:p-8` and starts with `PageHeader`. Lists are paged, searched and sorted by the API, with one address scheme everywhere: `page`, `perPage` (10 / 20 / 50), `sort=name.asc`, `q`, `view`, plus each screen's filters (`useListParams(defaultSort, filters)` in `lib/list-params.ts`, nuqs). **Tables** are tablecn (`components/data-table/`, vendored from tablecn's Base UI registry) through `ListTable` (one API page) or `ClientTable` (a short list on the page) in `components/chaintrack/list-table.tsx`: columns from `column(id, label, cell, { sortable })` (a sortable id is the API's sort column), row actions from `actionsColumn(...)` with icon `RowAction`s, always pinned to the end; a Table | Cards toggle (cards by default below `md`) and a View menu (columns). Card lists use `ListBar` + `ListPager`. Every select is `SelectInput` (shadcn Select), never a native `<select>`. Tabs, filters and sorts are `as const` in `lib/consts.ts`.

1. **Data:** every call is an `endpoints.*` function in `lib/api.ts` that parses the answer with a zod schema. Components read with `useQuery` and write with `useMutation`, then invalidate the query keys the write changed. Never `fetch` in a component.
2. **Errors:** every failure is an `ApiError` carrying the API's message. Mutations get the toast for free (rule 11). A section waiting on a query renders `<DataWidget query={q}>{(data) => …}</DataWidget>` — pending → `Loading`, failure → the API's message in a blocking `Notice` with retry; empty stays the table's own `empty` state. `PageState` is for whole pages that can't go on. A 400 about one field becomes that field's error.
3. **Copy and values:** no user-facing string, API path, query key, status or permission name is typed in a component. Static copy lives in `lib/copy/*.json` (translated as whole files, never edited per page); plurals, conditions and locale formatting stay as functions in `lib/consts.ts`, and `{placeholder}` templates are filled by `fill()` in `lib/format.ts`. Paths, keys, statuses and permission names stay in `lib/consts.ts` — import them from there, never from the JSON directly.
4. **Forms:** react-hook-form + zod (`zodResolver`) with `...liveValidation` (`lib/form.ts`: errors show on leaving a field, then update while typing); new passwords are typed twice. Every decision (approve, reject, suspend, delete, role change…) asks first with `useConfirm()`. with fields built from `TextField` (`type="password"` gets the show / hide button); a bare help-or-error line is `<FieldNote error={…} help={…}>` — never a hand-written `errors.x?.message` paragraph. The schema lives in `lib/schemas/`, and the API stays the authority.
5. **Pages:** `page.tsx` only exports `metadata` and mounts a client component inside `<PageSuspense>` (the one fallback, in `page-state.tsx`). Pages read search params on the client.
6. **Access:** permissions come from `GET /me` (`permissions.*` in consts). The nav (`navItems`) lists **only built screens**, and quick actions follow the same rule. A new screen adds its line.
   6b. **Server errors** carry the request id (`X-Request-ID`, made by the pass-through, logged on the API's access line); `problemOf` shows it as a reference on 5xx.
7. **Logic:** anything with a branch (who sees what, address building, parsing) goes in a pure `lib/*.ts` helper with a node test in `__tests__/` (same name; `lib/schemas/x.ts` → `__tests__/schemas-x.test.ts`). Pure files and tests import only with relative paths and `.ts` extensions, so `node --test` can load them.
8. **Signing:** actions that sign (issue, handover start / accept / reject, transform, retire, certificate issue / revoke, flag resolve) send the person's 6-digit PIN as `X-Signing-PIN`. Follow `patterns.md` → Signing: summary first, then the PIN, then a button that names the result.
9. **Data display:** lot codes are shown in full, in mono. Quantities always carry their unit. Statuses are an icon plus a word.
10. **Organisation address:** the acting organisation is the first label of the host (`acme.localhost:3000`). Sessions are one per address; switching organisation hands the session over (`switchOrganisation` in `lib/api.ts` → `/handoff?code=`), no new sign-in. On the bare address a member acts for nobody and gets the workspace picker (`hasWorkspace`, `lib/dashboard.ts`).
11. **Messages:** every API answer is `{status, message, data}`. Success and refusal sentences come from the API and are toasted once, by the `MutationCache` in `app/providers.tsx`; never `toast.success` / `toast.error` an API outcome in a screen. A screen that shows the failure itself sets `meta: { inPlace: true }`. Field-rule messages (zod) stay in `lib/schemas/`.

## Visual rules (summary; the skill has the detail)

- Theme tokens only (`bg-surface`, `text-ink-muted`…). No hex, no default Tailwind palette, no `dark:` colour variants: tokens switch with `data-theme`.
- No arbitrary pixel values (`text-[22px]`, `size-[18px]`): the Tailwind scale (`text-2xl`, `size-4.5`, `max-w-140`) and theme tokens (`shadow-selected`). Only the public page art keeps arbitrary positions.
- Status colours mean status only. One primary button per screen.
- Text at least 12 px (`text-xs`), body 14 px. Controls 40 px on phones, 36 px from `lg` (`h-10 lg:h-9`; `min-h-tap` = 40 px, `h-tap-scan` = 48 px for scan, handover and phone tabs). Buttons carry no shadow.
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
- Use a native `<select>`, or a hand-written `<table>`: `SelectInput`, `ListTable` / `ClientTable`.
- Write a new `isPending ? <Loading/> : isError ? …` ternary, a new `Suspense fallback={<Loading…>}`, or a hand-written error/help `<p>` with `OctagonX`: `DataWidget`, `PageSuspense`, `FieldNote` / `TextField` already do it.
- Put a new user-facing string in a component or in `lib/consts.ts`: static copy goes in `lib/copy/*.json`, functions stay in `lib/consts.ts`.
- Add another component or UI library (MUI, Chakra, Mantine…), or a library for what a few lines do.
- Add another component or UI library (MUI, Chakra, Mantine…), or a library for what a few lines do.
- Put an edit button on a signed record. Corrections are new, linked records.
- Use `localStorage` for anything but per-browser preferences (sidebar folded, theme), and always inside try/catch.
- Show "Signed by both parties" before the receiver has signed.

## Design source

Import done. Two references, both read-only (never imported by the app; their demo names and codes are never copied):

- **Content and flows:** `ui/ChainTrack Drafts.dc.html` (with `support.js`), the Claude Design drafts, ids `1a`–`3o`. Original: https://claude.ai/design/p/c9f04332-afb9-4bbe-8b97-2c8686eed87e?file=ChainTrack+Drafts.dc.html
- **Look:** `ui/figma-make/App.tsx` and `ui/figma-make/styles.css`, the Figma Make app. Its tokens, fonts and component styles are in `app/globals.css` and the `chaintrack-ui` skill.

| Draft      | Route                                                           | Figma component                          |
| ---------- | --------------------------------------------------------------- | ---------------------------------------- |
| 1a, 1b, 1c | `/p/[id]`                                                       | `PublicVerification`                     |
| 1d, 3o     | `/`                                                             | `Dashboard`                              |
| 2a, 2b, 2c | `/signin`                                                       | `AuthFlow` (sign in)                     |
| 2d–2h      | `/forgot`, `/reset/[token]`, `/lock/[token]`                    | `AuthFlow`                               |
| 2i         | `/invite/[token]`                                               | `AuthFlow`                               |
| 2j, 2k, 2l | `/signup`, `/onboarding`                                        | `AuthFlow` (sign up, documents, pending) |
| 2m         | signing sheet (`components/chaintrack/signing.tsx`), `/profile` | `Profile`                                |
| 3a, 3b     | `/facilities`, `/facilities/new`                                | `Facilities`, `RegisterFacility`         |
| 3c, 3d     | `/issue`                                                        | `IssuePassport`                          |
| 3e         | `/passports/[id]`                                               | `Passports` (detail)                     |
| 3f         | `/scan`                                                         | `Scanner`                                |
| 3g, 1f, 3h | `/handover`, `/handovers`, `/handovers/[id]`                    | `Handovers`                              |
| 1e         | `/transform`                                                    | `Transform`                              |
| 3i, 3j     | `/certificates`, `/certificates/new`                            | `Certificates`                           |
| 3k         | `/flags`                                                        | `Flags`                                  |
| —          | `/notifications`                                                | (bell popover, full list)                |
| 3l         | `/admin/templates`                                              | `Templates`                              |
| 3m         | `/admin/organisations`                                          | `VerificationQueue`                      |
| 3n         | `/admin/reports`                                                | `Reports`                                |

Where they differ, and what we did:

- **Sizes:** the Figma file sets text at 7–10 px and buttons at 38–40 px. Here text is at least 12 px and controls 36–40 px (48 px for scan, handover, quick actions and the phone tab bar).
- **Colour:** the drafts use navy as the primary colour; the Figma look uses teal. Teal won: it is the brand and primary colour, and verified is the same teal, so verified always carries an icon and a word.
- **Decoration:** the hero glow, grid and product art appear on the public page only.
- **Phone tabs:** the drafts have five tabs with a raised Scan button; Figma has four tabs. We use four: Home, the person's first workspace screen, Scan, More.
- **Fields with no API behind them** (language, phone, country and registration number, flag severity, the verification-outcomes report, the inspection checklist, a certificate number): hidden; listed in `docs/design.md`.
