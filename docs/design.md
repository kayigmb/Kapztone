# Design reference

The design is the reference for Sprint 4 (PLAN.md §2). It lives in two places:

- **Visuals:** Claude Design. Links to the exported screens go in the "Export" column below.
- **Rules in words:** `.claude/skills/chaintrack-ui/` — tokens, components, patterns, copy, icons and a spec for every screen. Anything built in `frontend/` follows it.

## Phase 0 checklist (PLAN.md §2.5)

| Item | State | Where |
| --- | --- | --- |
| Design system with tokens and components, light and dark | Done | `chaintrack-ui/reference/tokens.md`, `components.md`; theme in `frontend/app/globals.css` |
| All 14 screens at mobile and desktop widths | Specified in words; visuals not linked yet | `chaintrack-ui/reference/screens.md` |
| Three clickable flows (coffee, mineral, fraud) | Specified as sequences; prototype not linked yet | `screens.md` §10 |
| Reviewed with at least 3 experts; changes logged | Not started | "Expert review log" below |
| Screens exported and linked here | Not done | "Export" column below |

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

Changed from shadcn's defaults so far: every `dark:` colour override removed (ChainTrack tokens already switch with the system theme), and button, input and select are at least 44 px tall (`min-h-tap`). Still to do when the first screen uses them: Button variants renamed to primary / secondary / ghost / danger-outline / danger, and the heights and radii in `components.md`.

## Screens

Numbers are PLAN.md §2.2. "Spec" is the section of `screens.md`. "API" is what the backend offers today.

| # | Screen | Spec | Route | API | Export |
| --- | --- | --- | --- | --- | --- |
| 1 | Public verification | 01 | `/p/[id]` (built; uuid until lot codes exist) | `GET /verify/{id}` | |
| 2 | Sign-in and onboarding (KYB) | 03 | `/signup`, `/signin`, `/onboarding` (built) | `POST /orgs`, `POST /auth/login`, `GET /me`, `POST /orgs/{id}/documents`, `/auth/2fa/*` | |
| 3 | Dashboard | 02 | `/` | `GET /passports`, `/handovers`, `/flags`, `/facilities`, `/notifications` | |
| 4 | Facility registry | 04 | `/facilities`, `/facilities/new` | `POST /facilities`, `GET /facilities?bbox=&status=`, `GET /facilities/{id}` | |
| 5 | Issue origin passport | 05 | `/issue` | `GET /templates`, `GET /facilities`, `POST /passports` | |
| 6 | Passport detail | routes in `patterns.md` | `/passports/[id]` | `GET /passports/{id}`, `/passports/{id}/genealogy`, `GET /attestations?subject_id=` | |
| 7 | Hand over custody (sender) | 06 | `/handover` | `POST /passports/{id}/handovers`, `POST /handovers/{id}/cancel` | |
| 8 | Accept custody (receiver) | 06 | `/handovers/[id]` | `GET /handovers`, `POST /handovers/{id}/accept`, `/reject`, `POST /flags` | |
| 9 | Transform / split / merge | 07 | `/transform` | `POST /transforms` | |
| 10 | Attach / issue certificate | 08 | `/certificates`, `/certificates/new` | `POST /attestations`, `GET /attestations`, `POST /attestations/{id}/revoke` | |
| 11 | Flags and investigations | 08 | `/flags` | `GET /flags`, `GET /flags/{id}`, `POST /flags/{id}/resolve`, `GET /passports/{id}/genealogy?direction=descendants` | |
| 12 | Commodity template builder | 09 | `/admin/templates/[id]` | `/templates` (create, list, get, replace, delete) | |
| 13 | Admin: organisation verification | 03 | `/admin/organisations` (built; `?tab=&id=`) | `GET /orgs?status=`, `POST /orgs/{id}/verify` (incl. `changes_requested`), `GET/PUT /orgs/{id}/accreditations`, `GET /orgs/{id}/documents`, `GET /documents/{id}/file`, `GET /templates` (certificate types) | |
| 14 | Reports and audit trail | 09 | `/admin/reports` | none | |

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
| 1, 5, 6 | Product name, packaging ("16 bags"), grade, crop year | Commodity, unit, quantity | |
| 1 | Route map with a location per journey step | Journey events carry no location | |
| 1 | Signed value against stored value ("signed for 2,100 kg, now shows 2,850 kg") | A reason code and message per finding, no field-level difference | |
| 1 | "Report a concern" from the public page | Flags need sign-in | |
| 4 | Capacity used this period, per facility | Computed only when a passport is issued; not exposed || Built: `capacity_used` |
| 4 | Licence expiry and document, approval before first use, edit boundary, GPS walk, overlap and forest-loss checks | Licence number; boundary validity and area | |
| 5 | Required-evidence checklist, inspector countersignature | Template lists evidence; nothing enforces or records it || Built: evidence = certificate, enforced; no inspector countersignature |
| 7, 8 | Part of a lot, location, seal number, photos; weight received, tolerance, inspection checklist | Whole passport, one note | |
| 8 | Rejecting a handover raises a flag and freezes the lot | Reject and flag are separate calls; no "on hold" status | |
| 10 | Certificate number, notes, maximum validity per type, "issued by you" list | Type, validity, file; list has no issuer filter | |
| 11 | Flag code (`FLG-2026-0147`), severity, assignee, flags raised automatically, evidence diff, "request information", suspend organisation | Reason, status, resolution note | |
| 12 | Template versions with "publish", who may act per stage, handover tolerance, evidence types | One live template per commodity; stages with events, yield, evidence keys | |
| 13 | "Changes requested" outcome; accreditation per certificate type with dates | Built: `changes_requested` (reopens on upload, like rejected); accreditation per type with a validity window, checked by API and contract | Built |
| 13 | Automatic checks (registry match, name match, licence readable and in date, duplicate, sanctions screen) | Nothing | Cut for now. Decision: learner |
| 13 | Country, registration number, contact in the detail | Name, address, kind | Cut for now. Decision: learner |
| 13 | Commodity tags, accreditation document, signing-key count on the issuers tab | Types and dates only | Cut for now. Decision: learner |
| 13 | A different validity window per certificate type in the form | API takes one per type; the form sends one window for all ticked types | Cut for now. Decision: learner |
| 3 | Counts for tiles, certificates expiring, volume by commodity | List totals only | |
| all | SMS; EXPORT, IMPORT and deliver steps on the journey | E-mail and in-app; issue, handover, transform, flag, revoke | |

## Expert review log

One row per piece of feedback (PLAN.md §2.5 asks for at least three experts).

| Date | Expert (role, organisation) | Screen | Feedback | Change made |
| --- | --- | --- | --- | --- |
| | | | | |
