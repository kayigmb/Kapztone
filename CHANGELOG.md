# Changelog

What has been built and what is left. Every finished piece of work adds an entry under Done and ticks or adds lines under Left. PLAN.md is the spec; this is the log.

## Left

### Phase 0 (due 15 Nov 2026)
- [x] Claude Design screens exported and linked in `docs/design.md` (drafts in `ui/`, ids per screen)
- [ ] Three clickable flows: coffee, mineral, fraud moments
- [ ] Review with at least 3 experts; changes logged (log in `docs/design.md` is empty)
- [x] Decision column of the design-versus-backend gap table in `docs/design.md` (dev scope frozen: build vs cut)
- [x] OpenAPI draft (`backend/api/openapi.yaml`; new report/route schemas added, remaining bodies still `object`)

### Sprint 2 gaps
- [ ] Deploy contracts to Polygon Amoy (waits on key, faucet funds, RPC URL; steps in `docs/operations.md`)
- [x] Split the platform key: `CHAIN_ADMIN_KEY` (deploy only, offline) and `CHAIN_RELAYER_KEY` (relayer + operator)

### Backend not built
- [x] User-management endpoints (invite, suspend, roles; must call `InvalidateUser`)
- [x] Organisation search for choosing a handover receiver (`GET /orgs` shows only the caller's own)
- [x] Readable lot code (`RW-KTB-2026-0412`) in place of the uuid
- [x] Per-action signing step (PIN / device) for issue, handover, transform, certificates
- [x] Reports, audit log, CSV export
- [x] PDF export (volume, blocked, audit)
- [x] Saved reports in bucket + verification-outcomes report (dev; eval runs later)
- [x] Retire a passport
- [x] Required evidence per stage, enforced and recorded
- [x] Regulator role for resolving flags
- [x] 2FA recovery codes
- [x] Capacity used per facility per period, exposed
- [x] Retry for failed IPFS pins
- [ ] nginx in front of the web app, appending the real visitor address (`$proxy_add_x_forwarded_for`); until then a visitor can fake their address
- [x] Real client address behind the web app (see Pending decisions)

### Sprint 4 frontend
- [x] Built: facility registry + register, issue, passport detail, scan, transform, handover (send, list, accept / reject / cancel), issue / revoke certificate, flags and investigations, certificates list, template builder, organisation verification, reports, users, profile
- [x] Maps (MapLibre + drawing) for facilities
- [x] Public origin map
- [x] Tables (TanStack Table, server-side)
- [x] Route map points-only (facility points in verify `route`, no invented legs)
- [x] Playwright happy-path test (coffee: facility → issue → handover → accept → transform → public)
- [ ] Playwright: certificates, reject / cancel, flags
- [ ] Verification of a 50-ancestor lot in under 5 s
- [ ] Visual check of built screens at 390 px and 1280 px, light and dark

### Sprint 5
- [ ] `tools/sim`: built (300 lots + 90 fraud cases, fixed seed); first run pending
- [x] `tools/baseline`: document-based comparison built (dev; first run pending)
- [ ] k6 load test: 1,000 events/hour
- [x] OWASP top-10 dev hardening (headers, caps, validation; full review + test runs later)
- [x] Medicine template added without code change (`tools/sim/templates/medicine.json`)
- [ ] nginx and VPS deployment
- [ ] Tag `v1.0-eval`
- [ ] Tick PLAN.md §7 checklist (several items already true)

### Security review leftovers (2026-10-22, low)
- [x] Origin check on public cookie POSTs (login, logout, handoff/accept): closes handoff login-CSRF
- [x] `/api/geocode` open proxy to Photon: no sign-in, no limit
- [x] Passport stays visible to the receiver of a rejected / cancelled handover (`passport/store.go` `visible`, no status filter)
- [x] Suspended organisations keep `Permission` routes (revoke certificates, invite, delete facilities)
- [x] Any regulator member, not only its admin, can confirm flags
- [x] Huge `?page=` overflows OFFSET (500); saved report with a non-uuid actor (500)
- [x] Logout leaves a bearer token valid until expiry (by design: the cookie is HttpOnly and cleared; bearer tokens are for scripts; `/auth/sessions/revoke` voids every token)
- [x] `http.Server`: add ReadTimeout / IdleTimeout
- [x] CSP and HSTS; CI `permissions:` block and SHA-pinned actions
- [x] Refuse Hardhat test keys when `APP_ENV=prod`

## Pending decisions

- How the API learns the real client address behind the web app. Rate limits now see the web app's address for every user, so sign-in and registration limits apply to the whole site.
- The Decision column of the gap table in `docs/design.md`: build each row in the backend, or cut it from the screen.

## Done

### 2026-10-22 — Security leftovers closed
- **Public POSTs:** `auth.SameOrigin` refuses a public change (sign-in, handoff, forgot, register…) sent from another site's page; callers without an `Origin` (scripts) pass. Closes the handoff login-CSRF.
- **Geocode:** `/api/geocode` asks the API's `/me` with the visitor's cookie first and answers 401 to strangers.
- **Visibility:** the receiver of a handover sees the lot only while it awaits them (not after reject, cancel or expiry).
- **Suspended organisations** can read but every non-GET answers 403 (`MsgOrgSuspended`, in `auth.Require`).
- **Flags:** assigning and resolving need the superadmin or a verified regulator's **admin**; regulator members still read every flag. The flags screen mirrors it (`orgAdminRole`).
- **Bad input:** `?page=` capped at `listq.MaxPage` (100000); a saved report with a non-uuid `actor` is refused when saved (400).
- **Server:** `ReadTimeout` and `IdleTimeout` 2 min (no WriteTimeout, for long exports). Web: CSP `frame-ancestors 'none'; base-uri 'self'; object-src 'none'; form-action 'self'`, HSTS in production. CI: `permissions: contents: read`, actions pinned by commit SHA.
- **Config:** prod refuses Hardhat's public test keys as `CHAIN_RELAYER_KEY`; `APP_URL` refuses a query or fragment.
- **Cleanup from review:** `Frontend` type and `X-Forwarded-Proto` removed (`frontendHost` only); `attempts_left` reports the tighter of the two sign-in limits.
- **Tests:** `TestRejectedReceiverLosesSight`, `TestSavedReportActorChecked`, `TestConfigHardhatKeyByEnvironment`, regulator member 403, suspended invite 403 / read 200, forgot from another site 403, page overflow 400; full `make test-db`, frontend lint / build / 77 node tests green.

### 2026-10-22 — Security review fixes
- **E-mailed links from config:** new required `APP_URL`. Reset, verify, lock and invite links started with the request's `Origin` / forwarded host, so `POST /auth/password/forgot` with a forged `Origin` mailed a victim a real reset token pointing at another site. Links now start with `APP_URL`; invites with `auth.OrgURL(APP_URL, slug)`. `auth.WithFrontend` / `FrontendOf` removed (no users left).
- **Certificates scoped:** `GET /attestations/{id}` and `/file` answered any organisation; `Store.Get` now applies the same visibility as the list (issuer, holder of the lot, owner of the facility, platform).
- **Rate limits:** failed sign-ins also count per email from any address (`LoginEmailRate`, 20 / 15 min), since the forwarded address can still be faked until nginx is in front; wrong 2FA codes from a signed-in user (enable, disable, recovery codes) are limited (`CodeRate`, 5 / 15 min).
- **Tests:** forgot-password with a forged Origin, invite link on the org address, other-org certificate 404, per-email and 2FA limits, `TestOrgURL`; full `make test-db` green.

### 2026-10-21 — Email confirmation, platform admin user and organisation pages
- **Email confirmation:** migration `000024_email_verified` adds `users.email_verified_at` (existing accounts backfilled as confirmed). Sign-up emails a one-use link (24 h, `/verify-email/<token>`, `POST /auth/email/verify`); invited and seeded accounts count as confirmed. Signing in with the right password but an unconfirmed email answers 403 and sends a fresh link (at most 5 an hour per email). Migration `000025_users_is_verified` adds `users.is_verified`, a boolean generated from `email_verified_at`; sign-in checks it.
- **Users (platform admin):** `/users` rows open `/users/[id]` and can deactivate / activate the account (`POST /users/{id}/deactivate|activate`; deactivating signs out everywhere, never yourself). The detail page shows account facts (email confirmed, 2FA), each membership with role change and suspend / reinstate per organisation (`?organisation_id=` on the existing role, suspend and reinstate routes), and recent activity from the audit log.
- **Organisations (platform admin):** new status `suspended` (`POST /orgs/{id}/suspend` with a reason, `POST /orgs/{id}/reactivate`); a suspended organisation fails `RequireVerified` and its members see a suspended notice on onboarding. `/admin/organisations/[id]` shows the existing review panel plus status, members (`GET /orgs/{id}/members`), facilities, passports and certificates issued. The queue links to it; the Approved tab lists suspended ones too.
- **Lists narrowed by organisation:** `auth.Gate` lets the platform pass `?organisation_id=` to narrow any gated list to one organisation (only narrows, never widens).
- **Verification:** backend unit tests, new DB tests (`TestEmailVerification`, `TestPlatformAdminActions`), frontend lint/build and 77 node tests pass. `TestCommodityChainsEndToEnd/mineral` still hits the shared-database mailer race noted below.

### 2026-10-21 — Facility kinds, map search, and account UX
- **Facility kinds:** CRUD table and editor in template administration; active kinds feed registration and filters, inactive kinds remain visible on existing records/templates. Added migration `000023_facility_kinds_id` to preserve existing data while meeting the table UUID-id convention; `key` remains unique.
- **Facility registration:** Photon place search with selectable results and map fly-to, plus a map/details two-column desktop layout.
- **Workspace and detail screens:** reports, profile and passport detail now use underlined tabs; notifications are opened from the bell; organisation verification status is shown in the workspace card; audit includes an Organisation column.
- **Account security:** all new-password paths enforce 8–72 bytes plus lowercase, uppercase, number and symbol on frontend and backend. Signup checks email availability before proceeding through a rate-limited endpoint; final registration remains authoritative.
- **Verification:** backend unit tests, frontend lint/build and 77 frontend tests pass. Targeted DB tests for facility kinds, email availability, reset and change-password pass. The full `make test-db` suite encountered an existing shared-database mailer race in `TestCommodityChainsEndToEnd` (`notifications are still waiting to be e-mailed`).

### 2026-10-20 — Branded sender, background invites, HTML mail
- **Sender identity:** new required `MAIL_FROM` (fail-fast `mail.ParseAddress`, e.g. `"ChainTrack <noreply@chaintrack.digital>"`); envelope keeps the bare address, the `From:` header shows the display name. `SMTP_USER` is login-only now (docs + `.env.example` updated, Gmail alias/SPF note).
- **Background invites:** `POST /invites` returns immediately; the e-mail sends in a `go` routine like `sendLink`, failures only logged (re-invite makes a new link). Response semantics (`emailed` = queued) and frontend unchanged.
- **HTML mail:** new leaf package `mailtmpl` (table layout, inline styles, navy header, teal action button, optional `MAIL_LOGO_URL` mark, else text wordmark); `compose` builds `multipart/alternative`, plain-text part byte-identical so existing body assertions hold; 4-arg `Mail` func threaded through `auth`/`user`/`mailer.SendOne`; logo file lives at `frontend/public/brand/chaintrack-mark.svg`, app icon at `frontend/app/icon.svg` (staging `frontend/images/` removed).
- **Tests:** mailtmpl unit test (escaping, button, no-image fallback), config tests for `MAIL_FROM`/`MAIL_LOGO_URL`, mailer From-header assertions, async invite poll; full `make test-db` green.

### 2026-10-20 — No self role change (UI gate + API 409)
- **Users page:** the role control is disabled on your own row (same pattern as suspend); the existing "Make X Y?" confirm dialog is unchanged for everyone else.
- **API:** `PUT /users/{id}/role` refuses your own id with 409 `you cannot change your own role` (new `MsgSelfRole`), mirroring the suspend self-check. Last-admin rule untouched.
- **Test:** self-demotion attempt with two admins present returns 409 and the role stays `org_admin` (`access_db_test.go`).

### 2026-10-20 — Dev master plan: saved reports in bucket, outcomes, route points, baseline, OWASP hardening
- **Saved reports in bucket** (migration 000020): `saved_reports` + `report_files`, `POST/GET /reports/saved`, `POST /reports/saved/{id}/run` renders CSV/PDF and keeps it via `upload.Bucket.SaveBytes`, `GET /reports/saved/{id}/files`, `GET /reports/files/{id}` streams permission-checked. Frontend tiles + save/run in `admin/reports`.
- **Verification outcomes** `GET /reports/verification-outcomes`: revoked is failed, open flag is warning, rest verified, per month; CSV/PDF + tile.
- **Route points only** in verify `route`: facility points in origin order, no invented legs; public page lists points.
- **Flags/certs/handovers:** code `FLG-YEAR-…`, severity, cert number/notes, seal + checklist (migrations 000021, validation, OpenAPI).
- **`tools/baseline` built:** paper checklist over sim CSV, writes `baseline_outcome`; not run yet.
- **OWASP dev:** security headers in `next.config.ts`, saved-report cap (50/scope, 429), tightened Validate + OpenAPI schemas; `go vet` clean, OpenAPI route test passes, `pnpm lint` 0 errors, `pnpm build` passes.
- **Docs:** gap Decisions frozen, screens table + operations + README synced.

### 2026-10-19 — One response shape, workspaces, switching without signing in, dev without rate limits
- **Every JSON answer is `{status, message, data}`** (`httpx.OK / Error / ErrorData`; `WriteJSON` removed). Each change answers with a sentence from its package's `values.go` (`Msg…`); reads send `""`. Details a screen needs ride in `data` (`two_factor_required`, `attempts_left`, the handover difference, the overlapping facility). Files and CSV / PDF downloads are unchanged. OpenAPI: `Envelope` schema.
- **Messages come from the API, one toast path:** `lib/api.ts` parses the envelope; reads return `data`, writes return `{data, message}`. The QueryClient's `MutationCache` (`app/providers.tsx`) toasts the API's message on success and `problemOf(err)` on failure, unless the screen shows the failure itself (`meta: { inPlace: true }`). Screens' own `toast.success / toast.error` and the frontend's copies of success and refusal sentences are gone; the sign-in's attempts left now come in the API's message. The sim client and the e2e test read the envelope.
- **Switch organisation without signing in again:** `POST /auth/handoff {slug}` (signed in, members only) makes a one-minute, one-time code bound to that address; the target spends it with `POST /auth/handoff/accept` and gets its own session (`internal/auth/handoff.go`, DB test). The switcher and the new `/handoff` page use it.
- **Workspace picker:** a member on the bare address (acting for nobody) sees one card per organisation on `/dashboard`, and a sidebar of Workspaces + Profile only (`hasWorkspace` in `lib/dashboard.ts`, tested).
- **Signed in → dashboard:** `/`, `/forgot` and `/reset` send a signed-in person to `/dashboard` (sign in and sign up already did). Sign out asks first.
- **Modes:** `DEV=true` runs without rate limits (sign-in, sign-up, PIN); production keeps them. The start-up line logs `mode` and `rate_limits`; `/healthz` answers `{status, mode}`.
- Fixed the mail test that failed when the dev API shared the database (a deadline in place of a round count).
- Verified: backend vet + full DB suite (twice), lint (0 errors), build, 66 node tests.

### 2026-10-19 — Confirmations, passwords, live form errors, theme on access pages
- **Every decision asks first:** one shared `useConfirm()` dialog (`components/chaintrack/confirm.tsx`, mounted in `app/providers.tsx`). Used for approving / rejecting / asking changes of an organisation, saving or removing accreditations, approving, suspending, reinstating and deleting facilities, suspending, reinstating and changing the role of members, taking or letting go of a flag. Irreversible ones use the red button. The facilities screen's own dialog was removed for it.
- **Passwords:** 8 to 72 characters (was 12; API, form rules, OpenAPI, help texts). Sign-up and the profile password change now ask for the password twice; a mismatch is shown on the second field; only the first is sent.
- **Forms validate as you go:** every react-hook-form + zod form uses `liveValidation` (`lib/form.ts`): a field's error shows once you leave it and updates on each keystroke until it's right.
- **Theme menu** on every access page (sign in, sign up, onboarding, reset, invite, lock), top right.
- Verified: lint (0 errors), build, 64 node tests, backend build and the registration / OpenAPI tests.

### 2026-10-19 — Tables on tablecn, cards on phones, pinned actions, one select
- **tablecn** installed from its Base UI registry into `components/data-table/` (data table, column header, pagination, view options, skeleton, `use-data-table`, parsers); `@tanstack/react-table` 9.2.6. Unused pieces (toolbar filters, calendar, slider; date-fns and react-day-picker) removed; its few React-compiler lint hits fixed.
- `ListTable` (one API page) and `ClientTable` (a short list on the page) wrap it with a **Table | Cards** toggle (cards by default on phones), the **View** column menu, and an **actions column always pinned to the end** with icon actions and tooltips. Users (members and platform accounts), templates, the audit log, the volume report and the dashboard's top organisations all use it; hand-written tables are gone.
- One address scheme for every list, tables and card lists alike: `page`, `perPage`, `sort=name.asc`, `q`, `view` (`lib/list-params.ts`; `apiSort` / `tableSort` in `lib/list.ts`, tested).
- **One select:** `SelectInput` (shadcn Select) replaces every native `<select>` (list filters and sorts, handover stage, certificate type); the Select trigger now matches the input height. Button gained shadcn's `outline` and `icon` names for shadcn blocks.
- Verified: lint (0 errors), build, 64 node tests; no native `<select>` left. Not checked in the browser yet (API and web app were stopped); `make e2e` updated for the new stage select, not re-run.

### 2026-10-19 — Senior pass: list kit, marketing home, platform dashboard, logging, certificates scope, PDF, IPFS retry
- **One way to build a list** (frontend): `useListParams` on nuqs (typed URL state: page, rows per page 10/20/50, search, sort, per-screen filters), `usePaged` + `endpoints.*Page`, `ListBar` / `ListPager`, and `DataTable` on TanStack Table (server-side sort by header click, stacked rows on phones). Every list uses it: users, templates, facilities, handovers, certificates, flags, notifications, verification queue (compact rows now), audit. Tabs, filters and sorts are constants in `lib/consts.ts`; `PageHeader`, `Segmented` and `useThemeTokens` are shared.
- **Buttons and sizes:** every action is `Button` / `buttonVariants` (the reports "Download CSV" was hand-styled and stayed big); `size="scan"` only on the camera sheet. No arbitrary pixel classes left outside the public page art (Tailwind scale, `shadow-selected` token).
- **Certificates** (backend + screen): an organisation lists only certificates it issued or that are about its lots and facilities (`?scope=issued|about`); the platform sees all and can open the page. Before, any organisation could list every certificate. Tests for each side.
- **Marketing home at `/`** from the supplied design, rebuilt in our tokens and components, with the claims the product doesn't back removed (offline sync, other languages, help centre, seal and certificate numbers, flag severity, implementation team); the lot-code box opens the real verification. The app's home moved to **`/dashboard`**.
- **Platform dashboard:** `GET /stats` (totals by status, 30-day activity, lots by commodity, top organisations, ledger sealing backlog; one read-only snapshot) and charts (Recharts; colours validated for colour-blind readers; a table view under every chart).
- **Logging:** `internal/logx` (slog, JSON or text). Request ids from the web app through the API and back, one access line per request with route, status, duration, user, organisation and the server error, panics with stacks. Every `log.Printf` replaced with structured fields; the relayer logs each sealed record. 5xx messages in the app quote the reference.
- **PDF export:** `format=pdf` on volume, blocked and audit (go-pdf/fpdf), buttons beside CSV on the reports screen.
- **IPFS retry:** migration 000019; failed pins retried with exponential back-off (1 min to 6 h, 10 tries), claimed with SKIP LOCKED.
- **Also fixed:** `TestFacilityRulesAndReports` no longer fails after midnight (the test wrote a local date the API reads as UTC).
- Verified: frontend lint (0 errors), build, node tests; backend vet, logx unit test, DB tests for scopes, stats, PDF, retry, users, OpenAPI drift + Redocly. Not verified in the browser against the running API (it needs a restart for all of this), and `make e2e` not re-run.

### 2026-10-19 — UI pass: one page frame, smaller controls, server-side lists, notifications page, shell
From the review of the running app.
- **One page frame:** every screen's `<main>` is the same full-width `grid w-full content-start gap-6 p-4 lg:p-8` (issue, transform, scan, profile, passport and handover pages were centred and narrower).
- **Smaller controls:** buttons and inputs 40 px on phones, 36 px from `lg`, no button shadow; `min-h-tap` 40 px, `h-tap-scan` 48 px (was 44 / 56).
- **Lists paged, searched and sorted by the API** (10 compact rows a page, search box, sort, pager; state in the address): facilities (type filter), handovers (incoming / outgoing / done counted by the API), certificates (issued by you), flags (open / resolved), templates, verification queue (per tab), users, audit log, notifications; issue passport's template and facility pickers search and page too (6 a page). Facility rows are one line with a small capacity bar; suspend / reinstate / delete moved into the facility's detail.
- **Backend for it:** `GET /handovers?box=incoming|outgoing|done`, `GET /facilities?kind=`, `GET /attestations?issuer_id=`, `GET /flags?status=resolved`, `GET /orgs?status=a,b&kind=a,b` (comma-separated), `GET /users` paged through listq for organisations and the platform. Tests for each; OpenAPI updated.
- **Notifications page** `/notifications` (all / unread, search, pages). Opening a notification, here or in the bell, marks it read and goes to what it's about (handover, flag, passport, organisation; a certificate opens its lot or facility). `lib/notifications.ts`, tested. Nav item; the bell shows the latest six and "View all".
- **Shell:** sidebar brand row and top bar both 64 px so their borders meet; the organisation card moved to the bottom of the sidebar; Profile is a nav item under Manage; the person's initials sit at the right of the top bar (to Profile); the theme button is a Light / Dark / System menu.
- **Public page:** the origin card shows the facility on a map (the API already sent its location; the page now keeps it). The design gap row was wrong and is removed.
- Verified: lint (0 errors), build, 63 node tests; backend vet, the touched DB tests and the OpenAPI test pass. Not verified in the browser: the API on :8080 was started before these changes and must be restarted, then `make e2e` (the spec now searches the issue pickers).
- Known (not from this change): `TestFacilityRulesAndReports` fails between midnight and 02:00 CAT (a "yesterday" licence date in local time is still today in UTC), and `TestCommodityChainsEndToEnd` sometimes finds notifications not yet e-mailed.

### 2026-10-19 — Simulation harness, OpenAPI draft, medicine template
- **`tools/sim`** (`make sim`): 3 chains (coffee, tin, medicine) × 100 lots, 20 facilities each, fixed seed; 90 fraud cases (F1–F6, 5 per chain per type) as API attempts (refused = prevented) or insider SQL after sealing (verification not Verified = detected); every legitimate lot verified for false alarms; CSV + summary. Mechanisms in `tools/README.md`. Unit tests: plan reproducible per seed, 5 per type, outcome rules. **Not run yet** (first run pending).
- **Medicine template** added as a data file (`tools/sim/templates/medicine.json`: manufacturer → distributor → pharmacy, unit box, GMP certificate and batch release as evidence, zero weight tolerance). No code change anywhere.
- **OpenAPI draft** `backend/api/openapi.yaml` (3.1): every route with its permission (`x-permission`), public routes, the signing PIN header, list parameters, request bodies for the main inputs; response bodies mostly `object` for now. Valid (Redocly lint, style rules aside). `backend/tests/openapi_test.go` fails when a route, its permission or its public status drifts from `server.go`.
- Verified: sim `go vet` + tests, backend `go vet` + OpenAPI test.

### 2026-10-18 — Playwright happy path, and the bugs it found
- `make e2e` → `frontend/e2e/coffee.spec.ts` (Playwright 1.63, Chromium) against the running stack. Setup through the API: a test template, a producer and a processor registered, documents uploaded, verified, PINs set. Then through the UI: register a farm plot (pin on the map) → platform approves → issue 1,000 kg → hand over to the washing station → it weighs and accepts → transform into 180 kg → the public page of the new lot shows a verdict that isn't Failed (Verified needs the chain and relayer; without them it's Warning). Runs in about 7 s.
- Bugs fixed on the way:
  - **Transform** never worked: the API answers the transform with its outputs inside it, the page expected a bare list ("answered in a way this page doesn't understand" after the lot was already made).
  - **Maximum yield** units: the API holds a fraction (0.2), every screen showed and took a percentage, so a 20% stage read as 0.2% and blocked every transform, and the template builder saved 20 as 2,000%. Converted once in `lib/api.ts` (`yieldToPct` / `yieldToFraction` in `lib/schemas/template.ts`, tested).
  - **Accepting a handover**: the "Handover complete" confirmation vanished as soon as the handover refreshed. It now shows from the handover's status.
  - **Issue**: required evidence was looked up under a stage named `ISSUE` (an event, not a stage key), so the checklist was always empty. **Transform** listed every stage except one keyed `ISSUE`; it now lists the stages that allow TRANSFORM, labelled like everywhere else ("Washing station").
  - **Template builder** started new templates with stage keys `ISSUE` / `HANDOVER`, which the API refuses; now `origin` / `transport`.
  - A leftover "Current holder:" with nothing after it on transform input rows, removed.
- Known: each run leaves 2 organisations and 1 template in the dev database (names carry the run id). Sign-ups go through a random `X-Forwarded-For` to stay under 10 per hour; that stops working once nginx sets the real address.
- Not covered yet: certificates (needs a third organisation and accreditation), reject / cancel, flags.
- Verified: e2e green 3 runs in a row, lint (0 errors), build, 60 node tests.

### 2026-10-18 — Facility maps, approval, licence (drafts 3a, 3b)
- `/facilities/new`: location first. Drop pin | Draw boundary on a MapLibre map (OpenFreeMap street style, no key), undo / clear, centre and corner readout, GPS sets the pin. Licence end date and licence document (uploaded right after the facility is created). "Submit for approval".
- `/facilities`: an overview map of the listed facilities with a legend; the selected facility shows its own map, licence end date and document (open / upload a renewed one).
- Statuses: Active, Awaiting approval, Suspended, Licence expired (`lib/facility.ts`, tested). The platform approves pending facilities from the detail; suspend / reinstate is hidden while pending.
- `/issue`: facilities awaiting approval are disabled with the reason.
- Map colours follow the theme tokens and update when the theme flips.
- Verified: lint (0 errors), build, 59 node tests; pages load. Not checked: drawing by hand, the map at 390 px in light and dark.
- Not built: public origin map (the public verification gives no facility location; gap row added), edit boundary, GPS walk, area readout while drawing (the API computes area), dashed outline for expired licences.

### 2026-10-18 — Flags and investigations (draft 3k)
- `/flags`: a split queue (open / resolved) with the investigation beside it. Shows the lot (code from its public page), who raised it (business, rules engine, public verifier), the assignee and the affected lots (descendants, where readable).
- Investigators only (platform, or a verified regulator), matching the API: assign to me / unassign; clear with a note, or confirm fraud. Confirming shows the consequences first (lots revoked, public pages turn Failed); both are signed with the PIN. Others see a read-only view.
- Nav "Flags" with the open count; "Review flags" quick action restored.
- Verified: lint (0 errors), build, 56 node tests; page loads.

### 2026-10-18 — Issue certificate (draft 3i)
- `/certificates/new`: find the lot by its code (typed or scanned) through its public verification, since issuers certify other organisations' lots. Attach the certificate to the lot or to one of its origin facilities.
- Type list = this organisation's accreditations valid today; dates; signed document.
- Live preview plus "checks before issuing" (lot verdict, accredited type, dates, file). The PIN signs the form as it stands after validation.
- "Issue certificate" quick action restored. Revoke (3j) was already built on `/certificates`.
- Verified: lint (0 errors), build, 56 node tests; page loads.

### 2026-10-18 — Handovers: send, list, accept, reject, cancel (drafts 3g, 1f, 3h)
- `/handover` (3g): pick or scan a lot → all or part of it, the stage it goes into (only stages allowing HANDOVER), where / how, GPS and up to 6 photos → verified receiver search → PIN. Sending part of a lot shows what stays.
- `/handovers`: incoming / outgoing / done tabs with counts. `/handovers/[id]`:
  - receiver accepts (1f) after entering the weighed quantity; the template tolerance is checked live and blocks the PIN step when outside;
  - or rejects (3h) with reasons from the API's list, a note and "also raise a flag";
  - the sender can cancel while it waits;
  - photos, location and the split note are shown.
- Nav "Handovers" with the incoming count; the "Hand over" quick action and buttons work again.
- New: `lib/handover.ts` (tolerance, handover stages, partial; tested), `components/chaintrack/place-and-photos.tsx`, `handover-status.tsx`. `lib/schemas/handover.ts` rewritten for the current API; its hard-coded 0.5% tolerance removed.
- Verified: lint (0 errors), build, 56 node tests. Pages load; not clicked through against the API yet (it needs restarting).

### 2026-10-18 — Forgot / reset / lock pages, sign-in extras (drafts 2b–2h)
- `/forgot` (2d, then 2e "check your email": the same answer for every email), `/reset/[token]` (2f new password with live rules, 2g expired or used link, 2h done) and `/lock/[token]` (a deliberate button, so mail scanners opening the link don't lock anyone).
- Sign in: "Forgot password?" link, attempts left after a wrong password (from `attempts_left`), a Caps Lock warning, and "Trust this device for 30 days" once the code is asked for.
- Shared `components/chaintrack/auth-result.tsx` for the outcome screens.
- Verified: lint (0 errors), build, 54 node tests (new: reset schema). Pages load; the endpoints need the restarted API (`JWT_TTL_REMEMBER` in `.env`).

### 2026-10-18 — Forgot password no longer reveals accounts
- `POST /auth/password/forgot` answers 202 before looking anything up; the lookup, link and mail run in the background. Known and unknown emails now take the same time, and a failure is logged, never returned. Access tests pass.

### 2026-10-18 — New look from Figma Make
- **Reference:** `ui/figma-make/App.tsx` and `styles.css` kept as reference only; `ui/ChainTrack Drafts.dc.html` for content and flows.
- **Tokens and fonts:** the Figma palette under the same token names (teal primary, navy, sidebar and hero sets, light and `data-theme` dark); Inter + DM Mono; radii 8 / 12; the Figma card shadow; readable minimums (text ≥ 12 px, targets ≥ 44 px, 56 px on scan, handover, quick actions and phone tabs).
- **Primitives and components:** Button (teal primary, outline secondary, `danger-soft`), Input (teal focus ring), Tabs (underlined), StatTile (metric card), StatusChip (pill), Notice, Stepper (numbered circles), ListCard, CapacityMeter, JourneyTimeline, ChecksList, Brand.
- **Shell:** navy sidebar with brand, organisation card, switcher, nav groups (Workspace / Trust & compliance / Manage) with live counts, the person → Profile; fold rail, back button, bell, theme toggle and "Scan QR" kept; four phone tabs.
- **Public verification** (`/p/[id]`, lot code or id): navy hero with the verdict, product card with art and copyable lot code, Journey / Genealogy tabs (timeline + route art, tree with stage and quantity used), certificates, What we checked, origin, help card, Technical details.
- **Access screens:** the AuthFlow layout (navy brand panel, form on white, logo only on phones) for sign in (now with "Keep me signed in"), sign up, onboarding and invite.
- **Dashboard:** Figma metric cards, quick-action cards, card heads with "View all", passports as table-like rows, capacity panel; still from the per-kind config.
- **Repairs to earlier draft screens** so lint and build pass: hooks above early returns (reports, templates), zod input / output form types (templates, facilities, issue, transform). `/handover` and `/certificates/new` had no client component: parked as `page.tsx.draft`, and their nav entry and quick actions removed. "Review flags" left the dashboards (no `/flags` screen).
- **Docs:** `chaintrack-ui` SKILL.md and reference (tokens, components with Figma class map, patterns, screens' draft map, globals.css copy), `frontend/AGENTS.md` design source, `docs/design.md` (visuals linked, draft ids, gap rows).
- **Verified:** lint (warnings only), build, 53 node tests; no hex outside `globals.css`, no text under 12 px. Not checked: the visuals at 390 / 1280 px in light and dark (the learner's eyes).

### 2026-10-18 — Backend: account safety, custody, business rules
- **Access** (migration 000016):
  - `POST /auth/password/forgot` (always 202) and `/reset` (30-minute one-use link, last 5 passwords refused);
  - "Lock my account" link mailed after every password change (`POST /auth/lock`, 7 days);
  - `POST /auth/sessions/revoke` signs out other devices, the caller keeps a fresh token;
  - every token carries the user's token version, so a password change, reset, revoke or lock ends every older session (the earlier "open sessions survive a password change" is fixed);
  - sign-in takes `remember` (`JWT_TTL_REMEMBER`, new setting) and `trust_device` (`ct_trust` cookie skips 2FA for 30 days); a failed sign-in answers `attempts_left`.
- **Custody** (migration 000017):
  - partial handovers: a split (stage `split`, yield 1) cuts the part into its own passport, no contract change;
  - handover photos (up to 6, in the bucket, hashed into the signed record) and place, on both sides;
  - `received_quantity` checked against the stage's `handoverTolerance` (409 with the difference);
  - reject with declared reasons and `also_flag`.
- **Business rules** (migration 000018):
  - facilities wait for approval (`POST /facilities/{id}/approve`), need a licence end date, keep a licence document, and may not overlap another live facility;
  - templates: drafts, `POST /templates/{id}/publish`, versions recorded on transforms and handovers;
  - flags: `raised_by` (user / rules / public), assignee (`PUT /flags/{id}/assignee`); the rules flag transforms at 98% of the maximum yield and accepted handovers off by more than half the tolerance;
  - reports: `exported` in `GET /reports/volume`, new `GET /reports/blocked` (JSON / CSV).
- **Verified:** `make test-db` 82.2% coverage, Go vet and unit tests, 37 contract tests.
- **Known:**
  - a child passport from a split starts without the parent's certificates;
  - nothing produces `raised_by: "public"` yet (no public report form).

### 2026-10-17 — Shell upgrade, platform notifications, profile, users
- **Shell:**
  - collapsible sidebar (72 px icon rail, remembered per browser);
  - back button in the top bar;
  - light / dark / device theme toggle (next-themes on `data-theme`);
  - notifications bell in the header (unread count, mark as read);
  - org switcher, which goes to the other organisation's address, where you sign in again;
  - live pending count on "Verify organisations";
  - Users and Profile in the nav.
- **Platform notifications:**
  - nullable `notifications.organisation_id` (migration 000015);
  - the superadmin is told about a new organisation, uploaded documents and raised flags, and sees only platform notices;
  - they are e-mailed to platform users.
- **Profile** (`/profile`, everyone):
  - edit name (`PUT /me`) and change password (`POST /auth/password`);
  - signing PIN;
  - 2FA on / off with a QR and recovery codes;
  - `GET /me` now returns `two_factor`.
- **Users** (`/users`): the superadmin sees every user with their memberships (`GET /users` from the platform). An org admin sees their own members, and can invite, change roles and suspend.
- **Invite accept page** at `/invite/[token]`.
- **Switcher:** `GET /me/organisations`.
- **Verified:**
  - DB tests 82.2% (backend agent), Go unit tests, 37 contract tests;
  - frontend lint, build and 21 node tests.
  - Not clicked through live (the dev API needs a restart).
- **Known:**
  - open sessions survive a password change until they expire;
  - one `documents_uploaded` notice per file;
  - with SMTP on, `make test-db` can send real e-mails if the dev API is running.

### 2026-10-17 — App shell, dashboard (screen 3), Scan QR
- **Shell** (`frontend/app/(app)/layout.tsx`, `components/chaintrack/app-shell.tsx`):
  - sidebar on desktop; app bar and a bottom tab bar (Home, Scan, More) on phones;
  - the sign-in and verification guard lives here;
  - nav lists only built screens (`navItems` in `lib/consts.ts`);
  - the admin queue now sits inside the shell.
- **Dashboard:** one page, laid out per kind of organisation (`dashboards` in `lib/consts.ts`): trader, producer, certifier, regulator, platform admin.
  - Tiles come from list `total`s, each its own query.
  - Main panels: incoming handovers, passports held, latest certificates, open flags, organisations waiting.
  - Producers also get capacity meters per facility.
  - Quick actions: only screens that exist (Scan QR, Review organisations).
- **Scan QR:** `@zxing/browser` 0.2.1 in a dialog, loaded only when it opens, with a typed lot-code field (the camera needs https or localhost). A label URL `…/p/{code}` or a typed code opens the public passport page.
- **Cut, with no data behind them:** "Waiting for your signature", lots verified this month, verified volume, certificates expiring in 30 days, flag severity, active users, issuers to accredit. Recorded in the gap table.
- **Verified:** lint, build, 16 node tests (new: dashboard config, scan parsing). Live: superadmin dashboard calls through the web app all answer 200. Not checked: a producer's dashboard with real data, the camera on a phone, and visuals at 390 / 1280 px in light and dark.

### 2026-10-17 — Backend finished: 12 items, designed by the learner
- **Org search:** `GET /orgs/directory` lists verified organisations (id, name, kind, slug), excluding the caller's own; new permission `org:search`.
- **Dashboard numbers:** the frontend reads list `total`s; facilities gain `capacity_used` for the current period.
- **Real client address:**
  - the API uses the last `X-Forwarded-For` entry, and the web pass-through forwards it;
  - `compose.prod.yml` no longer publishes the API port;
  - this is honest only with nginx in front, still in Left.
- **Regulator:** members of a verified regulator organisation list, read and resolve every flag.
- **User management:**
  - `GET /users`, `POST /invites` (an e-mailed link, 24 h, single use; the link is also returned in dev), `POST /invites/{token}/accept`, `PUT /users/{id}/role`, suspend / reinstate per organisation;
  - the last org admin and your own account are protected.
  - The frontend needs an `/invite/[token]` page.
- **2FA recovery codes:** 10 codes, shown once on enable and on `POST /auth/2fa/recovery`; each works once at sign-in.
- **Signing PIN:**
  - set with `POST /auth/pin` (6 digits, password check);
  - required in the `X-Signing-PIN` header for issue, handover start / accept / reject, transform, retire, certificate issue / revoke and flag resolve;
  - locked after 5 wrong tries in 15 min.
- **Retire:** `POST /passports/{id}/retire` (holder, any time); anchored on-chain; afterwards no handover, transform or retire; verify shows the lot as retired.
- **Lot code:** `SLUG-YEAR-NNNN` on every passport, including transform outputs; `/passports/{idOrCode}` and `/verify/{idOrCode}` take either.
- **Required evidence:**
  - a valid certificate of the evidence type, checked at issue (on the facility), handover (handovers now name their `stage`) and transform (on every input);
  - missing evidence gives 409 naming it.
- **Audit log and reports:**
  - every successful write is logged in a hash chain;
  - `GET /audit` (+ `?format=csv`), `GET /audit/verify`, and `GET /reports/volume` (+ CSV);
  - CSV cells are guarded against spreadsheet formulas.
- **Keys:**
  - a new on-chain `OPERATOR_ROLE` does the six routine admin calls;
  - `CHAIN_ADMIN_KEY` is only used by `make deploy` and can then go offline;
  - `CHAIN_PLATFORM_KEY` is renamed `CHAIN_RELAYER_KEY`.
- **Migrations:** 000012_access, 000013_passports, 000014_audit.
- **Verified:** Go + DB tests 82.5%, 37 contract tests at 100% coverage, Slither 0 high, frontend lint/build/tests. Built by four parallel agents and integrated afterwards.
- **Not done:** a live run on a local chain (needs the new `.env` keys).

### 2026-10-17 — Dev-relaxed settings (superadmin password, JWT secret, SMTP TLS), SMTP as separate settings
- `SUPERADMIN_PASSWORD`: any length up to 72 in dev; 12–72 still enforced in prod. Sign-up passwords keep their 12-character rule.
- `JWT_SECRET`: any value in dev; at least 32 characters in prod. `SMTP_USE_TLS`: optional in dev (empty = false), required and `true` in prod.
- `SMTP_URL` replaced by `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_USE_TLS`.
  - `SMTP_HOST=none` turns e-mail off.
  - `SMTP_USER` is also the sender.
  - `SMTP_USE_TLS=true` refuses a server without STARTTLS (port 587); port 465 is not supported.
- Verified: unit tests, including a fake mail server without encryption that receives nothing. `make test-db` 82.7%.

### 2026-10-17 — Uploads in a bucket (AWS S3 / GCS in production, RustFS in dev)
- KYB documents and certificate files now go to a private S3-compatible bucket instead of the local `UPLOAD_DIR`.
  - `backend/internal/upload` gains `Bucket` (minio-go v7.3.0); org, attestation (including the IPFS pin) and verify read through it.
  - Files still stream through the API after its permission check; the bucket stays private.
- Settings: `STORAGE_ENDPOINT`, `STORAGE_REGION`, `STORAGE_BUCKET`, `STORAGE_ACCESS_KEY`, `STORAGE_SECRET_KEY`, plus `STORAGE_PORT` for RustFS. `UPLOAD_DIR` and the `uploads` volumes are removed.
  - Switching between RustFS, AWS and GCS is a `.env` change only (`docs/operations.md`, "File storage").
- Infra:
  - RustFS 1.0.0 in `compose.infra.yml` under the `dev` profile, so production never starts it.
  - `make infra` loads only the infra file, so it no longer needs the contract addresses.
  - New `make storage-reset` wipes the dev uploads.
- Verified:
  - `go vet`, unit tests, and `make test-db` against RustFS (82.8% coverage), including the altered-file and missing-file checks.
  - Live run: uploads and certificates stored in the bucket; download gives the same SHA-256; anonymous bucket read 403; nothing under `backend/uploads`.
- Known, unchanged: the web pass-through does not forward `Range`, so partial downloads only work against the API directly.

### 2026-10-17 — Admin verification queue (screen 13), ask for changes, per-type accreditation, frontend rules
- **Ask for changes:** new organisation status `changes_requested`. A reason is required, as it is for `rejected`. Both statuses go back to `pending` when the organisation uploads a new document.
- **Per-type accreditation:**
  - `OrgRegistry.setAccreditation(org, certType, from, to)` and `isAccredited(org, certType)` replace the yes/no flag.
  - `AttestationRegistry.issue` takes `certType` and checks it.
  - Backend: `accreditations` table (migration 000011; `organisations.accredited` dropped), `GET`/`PUT /orgs/{id}/accreditations` (full set), allowed types = evidence keys in live templates, one outbox item per change, `tx_hash` written back.
  - Verify: issuer accreditation is checked per type; a certificate whose type differs from the chain's is Failed.
- **Frontend rules:** shadcn, zod (forms, API errors, response schemas), TanStack Query for every call, `frontend/lib/consts.ts` for every value and all copy. Existing screens refactored, and `/p/[id]` is now rendered in the browser.
- **Screen 13** `/admin/organisations`:
  - Needs `org:verify`.
  - Tabs: pending / issuers / approved / rejected.
  - Review form: Reject / Ask for changes / Approve.
  - Accreditation form: checkboxes for evidence keys, one date range.
  - Onboarding shows `changes_requested` as an amber notice.
- **Verified:** Go tests + DB tests (83.3% coverage), 36 contract tests (100% coverage), Slither 0 high, `pnpm lint`/`build`, 11 node tests. Live run on a local chain through the web app's `/api`: ask for changes → re-upload → pending → approve; accreditation refusals (409 / 400); certificate of an accredited type 201, other type 403. All outbox items confirmed.
- **Not built (cut for now; gap table):** automatic checks, country and registration number, commodity tags, signing-key count, per-type dates in the form. Admin lists fetch at most 100 per status (no paging).
- Visuals at 390 / 1280 px, light and dark: not checked yet.

### 2026-10-12 — Sign-up and onboarding (screen 2) — `343d3c6`
- Two-step sign-up (`frontend/app/(auth)/signup/`, schema in `lib/schemas/org.ts`) posting once to `POST /orgs`; seven organisation kinds as choice cards.
- `/onboarding`: document upload per type and a status view (pending, rejected with reason); `/` sends unverified organisations there.
- Verified live through the web host: register, duplicate 409, upload, rejection reason shown, verification redirects to `/`. Lint, build, 9 node checks pass.
- Not built (no backend): language, country, registration number, "changes requested", SMS / e-mail promise.

### 2026-10-12 — Cookie session and sign-in screen — `343d3c6`
- Session token in cookie `ct_session` set and cleared by the API: `HttpOnly`, `SameSite=Lax`, `Secure` outside dev, host-only. Bearer header still accepted.
- Cookie requests that change anything need `Origin` = the site's host.
- `frontend/app/api/[...path]/route.ts` passes browser requests through to the API; sign-in form in `app/(auth)/signin/` with react-hook-form + zod.
- Verified live: cookie flags, `/api/me`, 403 for foreign origins, sign-out clears. Tests in `backend/tests/session_test.go`.

### 2026-10-12 — Public verification page (screen 1) — `343d3c6`
- `frontend/app/(public)/p/[id]/`: server-rendered, no client JavaScript; `loading`, `not-found`, `error` states.
- All page copy in `frontend/lib/verification.ts`, checked by `lib/verification.test.ts`; components in `components/chaintrack/`.
- Organisation names added to the verify answer; API messages reworded so no ledger words reach the page.
- Verified live against a local chain: Verified, Warning and Failed pages. Phone / desktop and dark mode not checked yet.

### 2026-10-12 — Phase 0 design reference and frontend stack
- `docs/design.md`: Phase 0 checklist state, stack rules, 14-screen index with routes and API, design-versus-backend gap table, expert review log.
- shadcn/ui on Base UI (14 primitives in `components/ui/`), react-hook-form, zod. `dark:` overrides removed; 44 px minimum height on button, input, select.
- Verified: `pnpm lint` and `pnpm build` pass.

### 2026-10-12 — Public verification endpoint
- `GET /verify/{id}` (`backend/internal/verify/`): verdict + reasons from seven checks over the ancestor tree, database compared with chain.
- Cached 60 s, 60 requests a minute per address; an unreachable chain can never give "Verified".
- One test walks a lot through clean, unanchored, chain down, ten kinds of tampering, flags, revocation, cache and rate limit. Coverage 83.1%.

### 2026-10-12 — Sprint 3: transforms, certificates, flags, notifications — `7c0c77a`
- `POST /transforms` (process / split / merge, mass balance in parts per million); genealogy both ways by recursive CTE (`internal/passport/transform.go`).
- Certificates (`internal/attestation/`): file on the server, SHA-256 on-chain, optional Pinata pin.
- Flags (`internal/flag/`): confirming one revokes the passport and all descendants. Notifications (`internal/notify/`): in-app and e-mail.
- Verified: coffee and mineral chains end to end; live run with 21 queued items all confirmed. Coverage 82.6%.
- Not built: required evidence per stage, retire, audit log, regulator role, pin retry. No template versions; each transform stores the yield applied.

### 2026-10-11 — Sprint 2 gaps: Slither
- `make slither` (also in CI) fails on a high finding; run gave 0 high and 9 lower findings, left unannotated.
- Not done: Amoy deploy (waits on key, faucet funds, RPC URL; steps in `docs/operations.md`).

### 2026-10-02 — Sprint 2: keys, outbox, relayer, passports, handover
- Organisation signing key created at verification, sealed with `ENCRYPTION_KEY` (`internal/secret/`).
- Transactional outbox (`chain_outbox`) and relayer process (`cmd/relayer/`): oldest first, refused → failed, unreachable → wait.
- Passports: issue with rules and capacity; handover: sender signs → receiver accepts (holder changes, queued) / rejects / sender cancels / expires after 7 days.
- Verified: unit, DB and in-process chain tests (83%), plus a live run on a local Hardhat node.

### 2026-10-02 — Sprint 2: contracts
- Five contracts in `contracts/` plus shared `Managed.sol`: OrgRegistry, FacilityRegistry, PassportRegistry, CustodyManager (EIP-712, nonce per passport), AttestationRegistry.
- 33 tests at the time, 100% coverage (`test/ChainTrack.ts`). Deployed only to a local node.

### 2026-10-02 — Configuration clean-up
- `internal/config` is the only reader of the environment; everything from `.env`, no defaults, one combined error.
- No `BASE_DOMAIN`: the frontend address comes from the `Origin` header; `JWT_TTL` in whole minutes; `make api` runs the API natively.

### 2026-10-02 — Sprint 1: facilities, templates, 2FA
- Facilities with PostGIS point or polygon, area from boundary, bbox filter (`internal/facility/`).
- Commodity templates with full validation and CRUD (`internal/template/`).
- TOTP 2FA (`internal/auth/twofactor.go`) and Redis rate limits on sign-in and registration.
- Verified: Sprint 1 "done when" met; coverage 84%.

### 2026-10-02 — Platform conventions, cache, KYB
- Roles and permissions declared in code (`internal/auth/permissions.go`), synced at start; migrations are tables only.
- Every table has id, created_at, updated_at, is_deleted; every list uses `internal/listq` (page, size, search, sort, dates).
- Redis cache for access lookups with invalidation (`internal/auth/cached_store.go`); data gated by acting organisation; no `/admin` routes.
- KYB documents and review (`internal/org/`).

### 2026-10-02 — Identity, organisations, roles, permissions
- Tables users, userscope, roles, permissions, role_permissions, documents, organisations.
- bcrypt passwords, JWT with user id only, middleware Authenticate → ResolveOrg → Require; acting organisation from subdomain (`pslug` in dev).
- Superadmin seeded from `.env`.

### 2026-10-02 — Test layout
- Backend tests live in `backend/tests/`; the router moved to `internal/server` so tests can reach it.

### 2026-10-02 — Backend plumbing and compose split
- pgx pool, `/healthz` with database ping (200 / 503), air reload for Go.
- Compose split into `infra/compose.infra.yml` (db + one-shot migrate), `compose.dev.yml`, `compose.prod.yml` (requires `POSTGRES_PASSWORD`).

### 2026-10-02 — Repo scaffold — `aa8dd92`
- Top-level `frontend/`, `backend/`, `contracts/`, `infra/`, `docs/`, `tools/`, and a Makefile.
- Runnable skeletons, chi `/healthz`, PostGIS in compose, Hardhat config, CI, `.env.example`.
