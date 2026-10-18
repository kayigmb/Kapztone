# Changelog

What has been built and what is left. Every finished piece of work adds an entry under Done and ticks or adds lines under Left. PLAN.md is the spec; this is the log.

## Left

### Phase 0 (due 15 Nov 2026)
- [x] Claude Design screens exported and linked in `docs/design.md` (drafts in `ui/`, ids per screen)
- [ ] Three clickable flows: coffee, mineral, fraud moments
- [ ] Review with at least 3 experts; changes logged (log in `docs/design.md` is empty)
- [ ] Decision column of the design-versus-backend gap table in `docs/design.md`
- [ ] OpenAPI draft (`backend/api/openapi.yaml` does not exist)

### Sprint 2 gaps
- [ ] Deploy contracts to Polygon Amoy (waits on key, faucet funds, RPC URL; steps in `docs/operations.md`)
- [x] Split the platform key: `CHAIN_ADMIN_KEY` (deploy only, offline) and `CHAIN_RELAYER_KEY` (relayer + operator)

### Backend not built
- [x] User-management endpoints (invite, suspend, roles; must call `InvalidateUser`)
- [x] Organisation search for choosing a handover receiver (`GET /orgs` shows only the caller's own)
- [x] Readable lot code (`RW-KTB-2026-0412`) in place of the uuid
- [x] Per-action signing step (PIN / device) for issue, handover, transform, certificates
- [x] Reports, audit log, CSV export
- [ ] PDF export, saved reports, verification-outcome report
- [x] Retire a passport
- [x] Required evidence per stage, enforced and recorded
- [x] Regulator role for resolving flags
- [x] 2FA recovery codes
- [x] Capacity used per facility per period, exposed
- [ ] Retry for failed IPFS pins
- [ ] nginx in front of the web app, appending the real visitor address (`$proxy_add_x_forwarded_for`); until then a visitor can fake their address
- [x] Real client address behind the web app (see Pending decisions)

### Sprint 4 frontend
- [x] Built: facility registry + register, issue, passport detail, scan, transform, handover (send, list, accept / reject / cancel), issue / revoke certificate, flags and investigations, certificates list, template builder, organisation verification, reports, users, profile
- [x] Maps (MapLibre + drawing) for facilities
- [ ] Public origin / route map (needs facility location in the public verification); tables (TanStack Table)
- [ ] Playwright happy-path tests
- [ ] Verification of a 50-ancestor lot in under 5 s
- [ ] Visual check of built screens at 390 px and 1280 px, light and dark

### Sprint 5
- [ ] `tools/sim`: 300 lots + 90 fraud cases, fixed seed
- [ ] `tools/baseline`: document-based comparison system
- [ ] k6 load test: 1,000 events/hour
- [ ] OWASP top-10 review of the API
- [ ] Medicine template added without code change
- [ ] nginx and VPS deployment
- [ ] Tag `v1.0-eval`
- [ ] Tick PLAN.md §7 checklist (several items already true)

## Pending decisions

- How the API learns the real client address behind the web app. Rate limits now see the web app's address for every user, so sign-in and registration limits apply to the whole site.
- The Decision column of the gap table in `docs/design.md`: build each row in the backend, or cut it from the screen.

## Done

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
