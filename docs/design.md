# Design reference

The design is the reference for Sprint 4 (PLAN.md §2).

- **Content and flows:** 35 drafts with ids `1a`–`3o`. The "Drafts" column below names them per screen.
- **Look:** Inter + DM Mono, teal primary, navy sidebar and hero. Adopted with readable minimums (text ≥ 12 px, targets ≥ 44 px, 56 px on scan / handover).

## Phase 0 checklist (PLAN.md §2.5)

| Item | State | Where |
| --- | --- | --- |
| Design system with tokens and components, light and dark | Done | `frontend/app/globals.css` |
| All 14 screens at mobile and desktop widths | Drafts (content) and look | "Drafts" column below |
| Three clickable flows (coffee, mineral, fraud) | Drafts cover the moments; a click-through prototype is not linked | — |
| Reviewed with at least 3 experts; changes logged | Not started | "Expert review log" below |
| Screens exported and linked here | Done: draft ids per screen | "Drafts" column below |

## Frontend stack

| Concern | Choice | Rule |
| --- | --- | --- |
| Components | shadcn/ui (Base UI), in `frontend/components/ui/` | Owned and edited here. No other component library. Domain components go in `frontend/components/chaintrack/`, built on these |
| Forms | react-hook-form | Every form. Fields are built with `components/ui/field.tsx` |
| Validation | zod, through `@hookform/resolvers/zod` | One schema per form. It mirrors the backend's `Validate()` for the same input, so the browser refuses what the API would refuse. The API stays the authority |
| Data | TanStack Query (`@tanstack/react-query`), provider in `app/providers.tsx` | Every read is a `useQuery`, every write a `useMutation` that invalidates the keys it changed. Calls go through `lib/api.ts` → `/api/...` pass-through; no page fetches on the server |
| API answers and errors | zod, in `lib/api.ts` | Every response is parsed by a schema. Every failure is one `ApiError` (`status`, `message`); `problemOf()` turns it into what to tell the person (toast or blocking notice), a 400 becomes a field error |
| Values and copy | `lib/consts.ts` | The one place for statuses, kinds, document types, permission names, routes, API paths, query keys and all UI copy. Nothing is written out twice; schemas and pages import from it |

Installed primitives: button, input, label, textarea, select, checkbox, tabs, dialog, drawer, tooltip, sonner, table, field, separator.

Changed from shadcn's defaults: every `dark:` colour override removed (ChainTrack tokens switch with `data-theme` and the system theme), Button (primary teal / secondary outline / ghost / danger-soft / danger-outline / danger), Input (44 px, teal focus ring) and Tabs (underlined teal, or `pill`) restyled to the app's look.

## Screens

Numbers are PLAN.md §2.2. "API" is what the backend offers today.

| # | Screen | Route | API | Drafts |
| --- | --- | --- | --- | --- |
| 1 | Public verification | `/p/[id]` (built, new look; lot code or uuid) | `GET /verify/{id}` | 1a, 1b, 1c |
| 2 | Sign-in and onboarding (KYB) | `/signup`, `/signin`, `/onboarding` (built, new look; email availability checked before continuing) | `POST /orgs`, `GET /auth/email/availability`, `POST /auth/login` (`remember`, `trust_device`), `GET /me`, `POST /orgs/{id}/documents`, `/auth/2fa/*`, `/auth/password/forgot`, `/reset`, `/auth/lock` | 2a–2l |
| 3 | Dashboard | `/` (built, new look: navy shell, metric cards, quick actions, per-kind config in `lib/consts.ts` → `dashboards`) | `GET /passports`, `/handovers`, `/flags`, `/facilities`, `/notifications` | 1d, 3o |
| 4 | Facility registry | `/facilities` (list, map, approve, licence), `/facilities/new` (Photon place search, pin or boundary on a map; two-column desktop form) | `POST /facilities`, `GET /facilities?bbox=&status=`, `GET /facilities/{id}`, `/approve`, `/licence`, `GET /facility-kinds` | 3a, 3b |
| 5 | Issue origin passport | `/issue` (built) | `GET /templates`, `GET /facilities`, `POST /passports` | 3c, 3d |
| 6 | Passport detail (business view) | `/passports/[id]` (built with Journey, Certificates and flags, and Record details tabs) | `GET /passports/{id}`, `/passports/{id}/genealogy`, `GET /attestations?subject_id=` | 3e |
| 7 | Hand over custody (sender) | `/handover` (built, seal + checklist) | `POST /passports/{id}/handovers` (quantity, stage, place, seal_number, checklist, photos), `POST /handovers/{id}/cancel` | 3g |
| 8 | Accept custody (receiver) | `/handovers/[id]` (built) | `GET /handovers`, `POST /handovers/{id}/accept` (`received_quantity`, photos), `/reject` (`reasons`, `also_flag`), `GET /handovers/{id}/photos/{photoId}` | 1f, 3h |
| 9 | Transform / split / merge | `/transform` (built) | `POST /transforms` | 1e |
| 10 | Attach / issue certificate | `/certificates`, `/certificates/new` (built, number + notes) | `POST /attestations`, `GET /attestations?scope=issued\|about`, `POST /attestations/{id}/revoke` | 3i, 3j |
| 11 | Flags and investigations | `/flags` (built, code + severity) | `GET /flags`, `GET /flags/{id}`, `POST /flags/{id}/resolve`, `PUT /flags/{id}/assignee` | 3k |
| 12 | Commodity template builder | `/admin/templates`, `/admin/templates/[id]` (built; includes facility-kind management) | `/templates` (create, list, get, draft, publish, delete), `/facility-kinds` (create, list, update, delete) | 3l |
| 13 | Admin: organisation verification | `/admin/organisations` (built; `?tab=&id=`) | `GET /orgs?status=`, `POST /orgs/{id}/verify` (incl. `changes_requested`), `GET/PUT /orgs/{id}/accreditations`, `GET /orgs/{id}/documents`, `GET /documents/{id}/file`, `GET /templates` (certificate types) | 3m |
| 14 | Reports and audit trail | `/admin/reports` (built with Reports and Audit log tabs; audit includes Organisation) | `GET /audit`, `/audit/verify`, `/reports/volume`, `/reports/blocked` | 3n |
| — | Profile (everyone) | `/profile` (built with Account and Security and signing tabs: details, name, password, signing PIN, 2FA with recovery codes) | `PUT /me`, `POST /auth/password`, `/auth/sessions/revoke`, `POST /auth/pin`, `/auth/2fa/*` | 2m |
| — | Users | `/users` (built: platform admin sees all accounts and deactivates / activates them; org admin invites, changes role, suspends) | `GET /users`, `POST /invites`, `PUT /users/{id}/role`, `POST /users/{id}/suspend`, `/reinstate`, `POST /users/{id}/deactivate`, `/activate` | — |
| — | User detail (platform admin) | `/users/[id]` (built) | `GET /users/{id}`, role / suspend / reinstate with `?organisation_id=`, `GET /audit?actor=` | — |
| — | Organisation detail (platform admin) | `/admin/organisations/[id]` (built) | `GET /orgs/{id}`, `POST /orgs/{id}/suspend`, `/reactivate`, `GET /orgs/{id}/members`, `GET /facilities` + `/passports` with `?organisation_id=`, `GET /attestations?issuer_id=` | — |
| — | Confirm email | `/verify-email/[token]` (built) | `POST /auth/email/verify` | — |
| — | Accept invitation | `/invite/[token]` (built, new look) | `POST /invites/{token}/accept` | 2i |

## Where the design and the backend disagree

The screens were specified before the backend was built, and they ask for more than it does. Each row needs a decision: build it in the backend, or cut it from the screen. "Decision" is left empty on purpose.

**Blocks a screen as specified**

| Screen | The design shows | The backend has | Decision |
| --- | --- | --- | --- |
| 1, 5, 6, all | A readable lot code (`RW-KTB-2026-0412`), never truncated, in the QR URL | Only a uuid || Built: `SLUG-YEAR-NNNN`, usable in the URL |
| 5, 7, 8, 9, 10 | Signing as a last step with a PIN (or device biometrics) | No per-action signing step; 2FA only at sign-in || Built: 6-digit signing PIN per action (no device biometrics) |
| 7 | Search for the receiving organisation | `GET /orgs` shows a caller only their own organisation || Built: `GET /orgs/directory` |
| 14 | Saved reports, audit table, CSV / PDF export | Nothing || Partly built: audit table + CSV, volume report; no saved reports or PDF |
| 13, 9 admin | Users and roles: invite, suspend, permission matrix | No user endpoints || Partly built: invite, role change, suspend; no permission-matrix editing |

**The screen works without it, but shows less**

| Screen | The design shows | The backend has | Decision |
| --- | --- | --- | --- |
| 1, 5, 6 | Product name, packaging ("16 bags"), grade, crop year | Commodity, unit, quantity | Cut: hidden, never faked |
| 1 | Route map with a location per journey step | Journey events carry no location | Build M3: facility points only, no fake legs |
| 1 | Signed value against stored value ("signed for 2,100 kg, now shows 2,850 kg") | A reason code and message per finding, no field-level difference | Build M4: evidence field-diff on flag |
| 1 | "Report a concern" from the public page | Flags need sign-in | Build M4: signed-in Report concern → flag |
| 4 | Capacity used this period, per facility | Computed only when a passport is issued; not exposed || Built: `capacity_used` |
| 4 | Licence expiry and document, approval before first use, edit boundary, GPS walk, overlap and forest-loss checks | Licence number; boundary validity and area | Built: licence end date + document, approval, overlap check. Cut: edit boundary, GPS walk, forest-loss check |
| 5 | Required-evidence checklist, inspector countersignature | Template lists evidence; nothing enforces or records it || Built: evidence = certificate, enforced; no inspector countersignature |
| 7, 8 | Part of a lot, location, seal number, photos; weight received, tolerance, inspection checklist | Whole passport, one note | Build M4: partial, place, photos, tolerance built; add seal number + checklist |
| 8 | Rejecting a handover raises a flag and freezes the lot | Reject and flag are separate calls; no "on hold" status | Build M4: reject + `also_flag` built; add on-hold on confirm |
| 10 | Certificate number, notes, maximum validity per type, "issued by you" list | Type, validity, file; list has no issuer filter | Build M4: add number, notes, max validity, `?scope=issued\|about` already in API |
| 11 | Flag code (`FLG-2026-0147`), severity, assignee, flags raised automatically, evidence diff, "request information", suspend organisation | Reason, status, resolution note | Build M4: add code, severity, evidence diff, request information; Cut: suspend organisation |
| 12 | Template versions with "publish", who may act per stage, handover tolerance, evidence types | One live template per commodity; stages with events, yield, evidence keys | Build M4: drafts/publish/versions/tolerance built; add who-may-act per stage |
| 13 | "Changes requested" outcome; accreditation per certificate type with dates | Built: `changes_requested` (reopens on upload, like rejected); accreditation per type with a validity window, checked by API and contract | Built |
| 13 | Automatic checks (registry match, name match, licence readable and in date, duplicate, sanctions screen) | Nothing | Cut for now. Decision: learner |
| 13 | Country, registration number, contact in the detail | Name, address, kind | Cut for now. Decision: learner |
| 13 | Commodity tags, accreditation document, signing-key count on the issuers tab | Types and dates only | Cut for now. Decision: learner |
| 13 | A different validity window per certificate type in the form | API takes one per type; the form sends one window for all ticked types | Cut for now. Decision: learner |
| 3 | Counts for tiles, certificates expiring, volume by commodity | List totals only | Built from list totals. Cut: waiting-for-signature, lots verified/month, verified volume, expiring certs, active users, issuers-to-accredit |
| all | SMS; EXPORT, IMPORT and deliver steps on the journey | E-mail and in-app; issue, handover, transform, flag, revoke | Cut SMS/deliver; EXPORT/IMPORT stay as template events where defined |

**Fields the designs show that the API does not have.** Hidden in the UI until the backend has them; never faked.

| Screen | Design shows | The backend has | Decision |
| --- | --- | --- | --- |
| Profile, sign-up | Language (English / Kinyarwanda / Français) | No language setting | Hidden |
| Profile, users, sign-up | Phone number ("Email or phone") | Email only | Hidden |
| Sign-up, organisation verification | Country, registration number | Name, address, kind | Hidden |
| Flags | Severity (High / Medium / Low), severity filter | Reason, status, source, assignee | Build M4: severity + filter |
| Reports | Verification outcomes report ("96.8% verified") | Volume, blocked, audit | Build M1/M2: saved reports in bucket + outcomes tile |
| Receive handover | Inspection checklist (seal number, packaging intact) | Reject reasons, photos, received weight | Build M4: seal + checklist |
| Certificates | Certificate number (`COO-RW-2026-8401`) | Type, validity, file, document fingerprint | Build M4: number + notes |
| Public page | Country and region tag, coordinates, harvest period, grade, "Download verification report" | Facility kind, operator, capacity check | Build M3 points-only map; Cut country/region/grade; Build download via saved-report PDF |

## Expert review log

One row per piece of feedback (PLAN.md §2.5 asks for at least three experts).

| Date | Expert (role, organisation) | Screen | Feedback | Change made |
| --- | --- | --- | --- | --- |
| | | | | |
