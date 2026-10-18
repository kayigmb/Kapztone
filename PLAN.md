# ChainTrack — Implementation Plan

**ChainTrack: A Blockchain-Powered Logistics Tracking Application for Transparent and Tamper-Evident Shipments**

This plan turns the research proposal into a working system. It starts with the
visual design in **Claude Design**, then builds the system in five two-week
sprints and prepares the evaluation. Dates follow the proposal timeline.

---

## 1. What we are building (MVP)

A commodity-agnostic **digital product passport** for business shipments:

- Every lot (coffee, cocoa, minerals, medicines, or any other product) gets a passport at a **verified, licensed, geolocated origin facility**.
- The passport follows the lot through **processing, logistics, export and import** to the final buyer.
- **Both businesses sign** every custody handover.
- Smart contracts enforce **mass balance** (outputs ≤ inputs × max yield) and **facility capacity**, which blocks illegal goods from being "mixed in".
- **Accredited certifiers and regulators** attach hash-anchored, revocable certificates.
- **Anyone** can scan the QR code and verify the full history back to the origin.
  **Out of scope for the MVP (future work):** SMS/USSD, payments, IoT sensors, live customs integration, mainnet deployment.

### Tech stack

| Layer        | Choice                                                                                   |
| ------------ | ---------------------------------------------------------------------------------------- |
| Front end    | Next.js (App Router), React, TypeScript, Tailwind CSS, shadcn/ui (Base UI), react-hook-form + zod, TanStack Table, MapLibre GL (react-map-gl), qrcode, @zxing/browser, Sonner; React Flow for genealogy (undecided) |
| Back end     | Go (chi router), REST + OpenAPI, JWT, pgx, sqlc, golang-migrate                          |
| Database     | PostgreSQL 16 + PostGIS                                                                  |
| Cache        | Redis 7 (access lookups now; public verification results later)                          |
| Blockchain   | Solidity, OpenZeppelin, Hardhat, Polygon Amoy testnet                                    |
| Chain client | go-ethereum (abigen bindings, EIP-712 signing, relayer)                                  |
| Files        | Private S3-compatible bucket (AWS S3 or GCS; RustFS in dev) + optional IPFS via Pinata   |
| Infra        | Docker Compose, Nginx, Linux VPS, GitHub Actions                                         |
| Testing      | Go test + testify, Hardhat (Mocha/Chai), Slither, Playwright, k6                         |

---

## 2. Phase 0 — Design in Claude Design (start now → finish by 15 Nov 2026)

Design comes first so that requirements, API and data model are shaped around real screens. All screens are designed in **Claude Design** and reviewed with experts during the October interviews.

### 2.1 Design system (do this first)

- **Name and tone:** ChainTrack. Serious, trustworthy and calm, like a customs or banking tool, not a crypto app. No coins, no wallet language.
- **Colours:** deep navy (primary), teal (verified / success), amber (warning), red (failed / flagged), neutral greys. Every verification state must also use an **icon and a label**, never colour alone.
- **Type:** one clean sans-serif for the UI (e.g., Inter), with a monospace font for IDs and hashes.
- **Core components:** verification badge (Verified / Warning / Failed), passport card, lot genealogy tree node, timeline event row, map pin for facility, signature status chip ("Awaiting receiver signature"), quantity / mass-balance meter, certificate tile (valid / expired / revoked), data table, stepper form, QR scanner frame.
- **Light and dark mode, mobile first**, because many users will scan on a phone in a warehouse or at a port.

### 2.2 Screens to design

| #   | Screen                                        | Who                  | Key content                                                                                                                                                                     |
| --- | --------------------------------------------- | -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Public verification page** (most important) | Anyone, no login     | Big verdict badge with reasons; product and lot; origin facility on a map; journey timeline; genealogy tree back to origin; certificates with validity; "what was checked" list |
| 2   | Sign-in and business onboarding (KYB)         | New organisation     | Organisation details, role, upload registration and licence documents, status "Pending review"                                                                                  |
| 3   | Business dashboard                            | All business roles   | Passports I hold, pending handovers (to sign / to receive), flags, quick actions                                                                                                |
| 4   | Facility registry                             | Producer             | List and map of facilities; register facility (draw polygon or drop pin, licence, capacity per period); capacity used this period                                               |
| 5   | Issue origin passport                         | Producer             | Choose facility and commodity template; quantity; required evidence checklist from template; capacity check result; print QR label                                              |
| 6   | Passport detail (business view)               | Holder and parties   | Timeline, genealogy, certificates, current holder, actions (hand over, transform, attach, flag)                                                                                 |
| 7   | Hand over custody (sender)                    | Any holder           | Scan QR, quantity, location, photos, choose receiving organisation, sign                                                                                                        |
| 8   | Accept custody (receiver)                     | Receiver             | Scan, inspect, compare quantity, accept and sign, or reject / flag with evidence                                                                                                |
| 9   | Transform / split / merge                     | Processor            | Select input passports, enter outputs; live mass-balance meter (output ≤ input × max yield); child passports created                                                            |
| 10  | Attach / issue certificate                    | Certifier, regulator | Choose subject (passport or facility), type, validity, upload file; revoke                                                                                                      |
| 11  | Flags and investigations                      | Regulator, admin     | Flag queue, evidence, affected descendants, resolve as cleared or confirmed                                                                                                     |
| 12  | Commodity template builder                    | Admin                | Units, stages, allowed events, max yield, capacity rule, required evidence per stage; preview                                                                                   |
| 13  | Admin: organisation verification              | Admin                | KYB queue, approve or reject, accredit issuers, assign roles                                                                                                                    |
| 14  | Reports and audit trail                       | Admin, regulator     | Filters by commodity, facility, period; export CSV/PDF                                                                                                                          |

### 2.3 Key flows to prototype (click-through)

1. **Coffee:** farm plot → washing station (transform) → exporter → importer → roaster → consumer scans.
2. **Mineral:** mine site → trader → processor (transform with yield) → exporter → refiner.
3. **Fraud moments:** capacity exceeded at issuance; mass-balance breach at processing; receiver rejects a handover; public page shows **Failed** because a certificate was altered.

### 2.4 Brief to paste into Claude Design

> Design a web app called **ChainTrack**, a digital product passport for business shipments that proves where goods come from. Users are exporters, processors, logistics providers, certifiers and regulators; consumers only use a public verification page. The tone is trustworthy and calm, like a customs or banking tool. Avoid crypto styling and wallet language. Use a deep navy primary colour, teal for verified, amber for warning and red for failed, always with an icon and a label. Mobile-first, with light and dark mode. Start with the **public verification page**: a large verdict badge (Verified / Warning / Failed) with reasons, the product and lot, the origin facility on a map, a journey timeline, a genealogy tree back to origin, and certificates with validity. Then design the business dashboard, issue-passport form (with a capacity check), dual-signature handover (sender and receiver screens), and the transform screen with a live mass-balance meter.

### 2.5 Definition of done (Phase 0)

- [ ] Design system with tokens and components, in light and dark mode
- [ ] All 14 screens at mobile and desktop widths
- [ ] Three clickable flows (coffee, mineral, fraud moments)
- [ ] Reviewed with at least 3 interviewed experts; changes logged
- [ ] Screens exported and linked from `docs/design.md` as the reference for Sprint 4

---

## 3. Repository structure (monorepo)

```
chaintrack/
├── frontend/                 # Next.js front end (business app + public verify portal)
├── backend/                  # Go REST API
│   ├── cmd/api/              # main
│   ├── cmd/relayer/          # relayer worker
│   ├── internal/{auth,org,facility,template,passport,custody,attestation,rules,verify,chain,notify,audit}/
│   ├── db/migrations/        # golang-migrate SQL
│   ├── db/queries/           # sqlc queries
│   └── api/openapi.yaml
├── contracts/                # Hardhat project (Solidity + tests)
├── tools/sim/                # synthetic dataset + fraud injection (Go or Python)
├── tools/baseline/           # document-based baseline system for the experiment
├── infra/                    # compose.infra.yml (db + migrate), compose.dev.yml, compose.prod.yml; nginx, deploy scripts
├── docs/                     # design.md, architecture, ADRs, evaluation protocol
├── Makefile                  # infra / dev / prod / down / prod-down / logs / test
└── .github/workflows/        # CI: lint, test, contract tests, build
```

---

## 4. Core design details

### 4.1 Data model (PostgreSQL)

`organisations` (`slug`, `kind`, `status`, `signing_address`, sealed `signing_key`) · `users` · `roles` · `permissions` (`resource:action`) · `role_permissions` · `userscope` (user + role + organisation) · `documents` (KYB) · `facilities` (PostGIS `geom` point or polygon, `capacity` + `capacity_unit` + `capacity_period`, `status`) · `commodity_templates` (`capacity_rule` JSONB, `stages` JSONB holding events, `maxYield`, `requiredEvidence`) · `passports` (`quantity`, `remaining`, `period`, `status`, `data_hash`, `handover_count`, `tx_hash`; `facility_id` + `period` for an origin passport, `transform_id` for one made by a transform) · `handovers` (`sig_from`, `sig_to`, `status`, `deadline`) · `trace_events` (`kind`, `data_hash`, `tx_hash`) · `transforms` (`stage`, `max_yield_ppm` as applied, `data_hash`) · `transform_inputs` (transform + input passport + `quantity_used`: the parent → child link) · `attestations` (`issuer_org_id`, `subject_type` + `subject_id`, `type`, `valid_from/to`, `doc_hash`, `ipfs_cid`, `status`) · `flags` (`reason`, `status` open / cleared / confirmed) · `notifications` (`kind`, `message`, `read_at`, `emailed_at`) · `audit_log` (not built) · `chain_outbox` (queued transactions).

**Conventions (all tables, all lists):**

- Every table has `id` (uuid), `created_at`, `updated_at`, `is_deleted`. Deleting means setting `is_deleted`; unique rules and every query ignore deleted rows.
- Migrations create tables only. Roles, permissions and allowed values (kinds, statuses, document types) are declared once in Go (`internal/auth/permissions.go`, `values.go` files) and written or validated from there. Nothing reused is restated.
- Every list endpoint takes `page`, `size`, `search`, `sort`, `start_date`, `end_date` through `internal/listq` and returns `{data, page, size, total}`.

Every provenance row stores `data_hash` (SHA-256 of a canonical JSON form) and `tx_hash`.

### 4.2 Commodity template (example)

```json
{
  "commodity": "coffee",
  "unit": "kg",
  "originFacilityKind": "farm_plot",
  "capacityRule": { "type": "yield_per_area", "kgPerHaPerSeason": 1200 },
  "stages": [
    {
      "key": "origin",
      "events": ["ISSUE"],
      "requiredEvidence": ["plot_geolocation"]
    },
    {
      "key": "washing_station",
      "events": ["HANDOVER", "TRANSFORM"],
      "maxYield": 0.2
    },
    {
      "key": "export",
      "events": ["HANDOVER", "EXPORT"],
      "requiredEvidence": ["export_permit", "certificate_of_origin"]
    },
    { "key": "import", "events": ["HANDOVER", "IMPORT"] }
  ]
}
```

### 4.3 Smart contracts (Solidity)

| Contract              | Responsibilities                           | Key rules                                                                                                       |
| --------------------- | ------------------------------------------ | --------------------------------------------------------------------------------------------------------------- |
| `OrgRegistry`         | Verified orgs, roles, accreditation per certificate type with validity window | Only admin can verify/accredit                                                                                  |
| `FacilityRegistry`    | Facility hash, capacity per period, status | Track `originatedInPeriod`; suspended facilities cannot originate                                               |
| `PassportRegistry`    | `issue`, `transform`, `retire`, `revoke`   | Capacity cap at issue; `sum(outputs) ≤ sum(inputsUsed) × maxYield`; consumed/revoked passports cannot be reused |
| `CustodyManager`      | `recordHandover`, `recordEvent`            | Sender must be current holder; **both EIP-712 signatures** must verify; nonces prevent replay                   |
| `AttestationRegistry` | `issue`, `revoke`                          | Only issuers accredited for that certificate type, inside the window; revocation is permanent                                                                |

Events: `PassportIssued`, `CustodyTransferred`, `LotTransformed`, `AttestationIssued`, `Flagged`, `Revoked`. Only IDs, hashes, addresses, quantities and status go on-chain, never personal or price data. Quantities are stored as integers in the template's smallest unit (e.g., grams) to avoid decimals.

### 4.4 Signing and relayer (wallet-free)

- Each organisation gets a **platform-managed signing key**, encrypted at rest (AES-GCM with a master key from env or secret store).
- A user approves an action after login plus a second factor; the API builds the **EIP-712 typed message**, and the org key signs it.
- The **relayer** (Go worker) reads `chain_outbox`, submits transactions with a funded relayer account, manages nonces, retries on failure, and writes `tx_hash` back.
- If the chain is down, events stay queued with original timestamps and are anchored later.

**Handover flow as built.** The sender (current holder) creates the handover; the API signs the typed message with the sender organisation's key and the handover is `awaiting_receiver`, holder unchanged. The receiver accepts: the API signs the same message with the receiver's key, the passport changes holder at once and the chain transaction is queued in the same database transaction. The receiver may instead reject (with a reason), the sender may cancel, and an unanswered handover expires after 7 days; in all three the holder does not change. The relayer later writes `tx_hash` and `anchored_at`.

**Provenance as built (Sprint 3).** No template versions: a template's commodity and unit freeze once passports use it, yields may change, and every transform stores the yield it was judged by. A transform at a stage with `maxYield` 1 is a split or merge. Confirming a flag revokes the passport and all its descendants, including a descendant only partly made from it. Certificates keep their file on the server and their SHA-256 on-chain; pinning to IPFS (public) and e-mail are switched on in `.env`. Since built: required evidence per stage (handovers name their stage), `retire`, audit log, regulator organisations resolving flags, readable lot codes, a signing PIN per action.

**Keys.** `CHAIN_ADMIN_KEY` deploys and holds `DEFAULT_ADMIN_ROLE`, which only grants roles; keep it offline after deploying. `CHAIN_RELAYER_KEY` holds the relayer and operator roles: it sends every transaction, including the six routine admin calls (verify, signer, accreditation, yield, suspend, revoke).

### 4.5 Rules engine (Go, before anything is written)

1. Organisation verified and has the right role.
2. Facility active, licensed, and within capacity for the period.
3. Template stage allows this event; required evidence is present.
4. Quantity ≤ `remaining_qty`; mass balance holds for transforms.
5. The same rules are enforced again on-chain (defence in depth).

### 4.6 Verification algorithm (public page)

1. Resolve QR → passport ID (GS1 Digital Link style URL `/p/{passportId}`).
2. Load the passport and its **full genealogy** (recursive CTE over `passport_link`).
3. For every passport, event and attestation, recompute the hash and compare it with the on-chain value.
4. Check that every leaf traces to a **verified, active origin facility**.
5. Check mass balance along every transform, and capacity at every origin.
6. Check certificates: accredited issuer, within validity, not revoked, hash matches.
7. Return **Verified** (all pass), **Warning** (e.g., expired certificate, open flag) or **Failed** (hash mismatch, breach, revoked), with a list of reasons.

**As built.** `GET /verify/{passportId}` needs no sign-in, is limited to 60 requests a minute per address and cached for 60 seconds. It returns `verdict`, `reasons` (level, check, code, message, subject), `checks` (seven, each verified / warning / failed), the passport with its journey, `origins` (facility name, kind, location; owning organisation), `genealogy` and `certificates`. Failed: a record changed after it was written, a record differing from the chain, no origin, capacity or mass balance broken, a certificate file or fingerprint altered, a revoked passport anywhere in the tree. Warning: not anchored yet, chain unreachable, facility suspended, owner not verified, certificate expired or revoked, issuer not accredited, an open flag. Handovers are checked through the chain's holder and quantity rather than by re-hashing each one.

### 4.7 API outline (REST, documented in OpenAPI)

No `/admin/...` endpoints. Every endpoint is a regular resource endpoint gated by one permission, and what a caller sees is gated by the organisation they act for (subdomain, or `pslug` in dev): users see only their own organisation's data, the superadmin sees all. The full route table with each route's permission lives in `backend/internal/server/server.go`.

```
POST   /auth/login {remember, trust_device}   GET/PUT /me   GET /me/organisations
POST   /auth/password                  POST /auth/password/forgot | /reset   POST /auth/lock
POST   /auth/sessions/revoke
POST   /orgs (public registration)     GET  /orgs            GET /orgs/{id}
POST   /orgs/{id}/verify               GET/POST /orgs/{id}/documents
GET    /documents/{id}/file
POST   /auth/2fa/setup | /enable | /disable
POST   /facilities                     GET  /facilities?bbox=&status=    GET/DELETE /facilities/{id}
POST   /facilities/{id}/status | /approve   GET/POST /facilities/{id}/licence
POST   /templates                      GET  /templates       GET/PUT/DELETE /templates/{id}   (PUT saves a draft)
POST   /templates/{id}/publish
POST   /passports                      GET  /passports       GET /passports/{id} (with journey)
GET    /passports/{id}/genealogy?direction=ancestors|descendants
POST   /passports/{id}/handovers       GET  /handovers       GET /handovers/{id}   (JSON or multipart with photos)
POST   /handovers/{id}/accept | /reject | /cancel   GET /handovers/{id}/photos/{photoId}
POST   /transforms                     (process, split or merge: inputs, outputs, stage)
GET/PUT /orgs/{id}/accreditations     (per certificate type: [{type, valid_from, valid_to}], full set)
POST   /attestations (multipart)       GET  /attestations?subject_id=&status=    GET /attestations/{id}
GET    /attestations/{id}/file         POST /attestations/{id}/revoke
POST   /flags                          GET  /flags           GET /flags/{id}
POST   /flags/{id}/resolve             (confirmed = revoke the passport and its descendants)
PUT    /flags/{id}/assignee
GET    /notifications?unread=true      POST /notifications/{id}/read
GET    /verify/{passportId}            (public, no auth, cached)
POST   /passports/{id}/retire          GET  /orgs/directory
GET    /users                          POST /invites    POST /invites/{token}/accept
PUT    /users/{id}/role                POST /users/{id}/suspend | /reinstate
POST   /auth/pin                       POST /auth/2fa/recovery
GET    /reports/volume | /blocked      GET  /audit   GET /audit/verify   (?format=csv)
```

---

## 5. Sprint plan

| Sprint | Dates             | Goal                                                                                                                                                                                             | Done when                                                                                                                                  |
| ------ | ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------ |
| **0**  | Oct – 15 Nov 2026 | Claude Design screens, design system, expert review; repo, CI, Docker skeleton; ERD and OpenAPI draft                                                                                            | Phase 0 checklist complete; `make dev` runs empty API + web                                                                       |
| **1**  | 16 – 29 Nov       | Go API foundation: auth (JWT + 2FA), RBAC, organisations, KYB flow, facilities (PostGIS), templates                                                                                              | Admin can verify an org; producer can register a facility on a map; template CRUD; ≥ 80% Go test coverage on these packages                |
| **2**  | 30 Nov – 13 Dec   | Smart contracts: all 5 contracts, EIP-712 signatures, mass balance, capacity, nonces; Go bindings; relayer                                                                                       | Contracts deployed to Amoy; ≥ 90% contract coverage; Slither clean of high findings; issue + dual-signed handover works end-to-end via API |
| **3**  | 14 – 27 Dec       | Attestations + IPFS; transform / split / merge with genealogy; flags and revocation cascade; notifications (in-app, e-mail)                                                                      | Coffee and mineral chains can be run end-to-end through the API; revoking a passport revokes its descendants                               |
| **4**  | 28 Dec – 10 Jan   | Next.js app from the Claude Design screens: dashboard, issue, handover (sender + receiver), transform, certificates, flags, admin; **public verification page** with map, timeline and genealogy | All 14 screens implemented; Playwright happy-path tests pass; verification of a 50-ancestor lot < 5 s                                      |
| **5**  | 11 – 24 Jan 2027  | Hardening and evaluation tooling: fraud simulation harness, baseline system, load tests, security review, medicine template added without code change                                            | Simulation generates 300 lots + 90 fraud cases reproducibly; k6 shows 1,000 events/hour; release tagged **v1.0-eval**                      |

Buffer: one week is held in reserve (use the Christmas–New Year period carefully). Must-have scope if time runs short: FR1–FR13.

---

## 6. Evaluation preparation (built in Sprint 5)

### 6.1 Simulation experiment (25 Jan – 7 Feb 2027)

- `tools/sim`: generates 3 chains × 100 lots (coffee/cocoa, mineral, medicine), ~3,000 events, ~1,200 attestations, 60 facilities, with a **fixed random seed**.
- Injects **90 fraud cases**: 15 each of F1 record alteration, F2 unilateral/forged handover, F3 unverified origin, F4 quantity inflation, F5 cloned/reused passport, F6 invalid certificate (5 per chain per type).
- Runs every case through **ChainTrack** and through **`tools/baseline`** (same Go/Postgres stack, no chain, PDF certificates, document-review checklist). Two blinded reviewers apply the checklist.
- Outputs a CSV of detections and false alarms for analysis (McNemar's test, Cohen's kappa, 95% CIs).

### 6.2 User evaluation sessions (8 – 28 Feb 2027)

- 36 participants (6 per stakeholder group), each ~45 minutes.
- Tasks: issue a passport, complete a dual-signed handover, record a transform, attach a certificate, and **verify two shipments** (one with a document pack, one with ChainTrack, counterbalanced; one of the two contains fraud).
- Collect: task time, success, legitimate/fraud decision, trust scale, TAM, SUS, short debrief.
- Prepare: demo accounts per role, seeded data, printed QR labels, consent forms, questionnaire in Google Forms.

---

## 7. Quality, security and privacy checklist

- [ ] Canonical JSON + SHA-256 hashing library shared by API, relayer and verifier (with test vectors)
- [ ] No personal or price data on-chain; only IDs, hashes, addresses, quantities, status
- [ ] Keys encrypted at rest; master key never in the repo; rotate the relayer key before the evaluation
- [ ] Replay protection (nonces, deadlines) on all EIP-712 messages
- [ ] Rate limiting and caching on the public `/verify` endpoint
- [ ] OWASP top-10 review of the API; input validation on all endpoints
- [ ] Slither + unit tests for every contract rule, including failure paths
- [ ] Backups of PostgreSQL during the evaluation; uptime monitoring
- [ ] Data handled per Rwanda's Law No. 058/2021; delete or anonymise personal data after the study
