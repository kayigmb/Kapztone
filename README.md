# ChainTrack

Blockchain-backed digital product passports for transparent, tamper-evident shipments.

Every lot (coffee, cocoa, minerals, medicines, any commodity) gets a passport at a verified, licensed, geolocated origin facility. The passport follows the lot through processing, transport, export and import. Each change of hands is signed by both businesses, the rules of the commodity are enforced on every step, and the record is anchored on a blockchain. Anyone can scan the passport's QR code and check its whole history back to the origin.

## What it does

- **Passports at the origin.** A passport can only be issued from an approved facility, within the facility's declared capacity for the period, and with any evidence the commodity's first stage requires.
- **Custody handovers signed by both sides.** The sender signs the handover (EIP-712 typed data) and the receiver accepts or rejects it after weighing the goods. A handover can carry photos, a seal number, a checklist and a location. Part of a lot can be handed over, which splits it into its own passport.
- **Transforms with mass balance.** Processing turns input passports into output passports. Outputs may never exceed inputs × the stage's maximum yield; the contract checks the same rule.
- **Commodity templates.** The platform defines each commodity's stages: which events each allows, its maximum yield, its handover weight tolerance and the certificates it requires.
- **Certificates.** Accredited certifiers and regulators attach certificates to facilities and passports. The file's hash goes on-chain; a copy can go to IPFS. Certificates can be revoked.
- **Rules and flags.** The API flags actions it allows but finds close to a limit, such as a yield of 98% or more of the maximum, or a weight near the tolerance. Anyone in the chain can raise a flag, and a regulator's admin clears or upholds it. A revoked passport revokes everything made from it.
- **Public verification.** `/p/<code>` shows the lot's journey, origins, certificates and a verdict (verified, warning, failed). The verdict comes from re-checking the database against the chain, the stored files and the mass balance.
- **Audit and reports.** Every successful change is written to a hash-chained audit log. Volume, blocked-action and verification-outcome reports export to CSV or PDF.

## Who uses it

Organisations register, upload KYB documents and are verified by the platform. Their kinds are:

- **Producer**, processor, transporter, exporter and importer: hold and move lots.
- **Certifier**: issues certificates for the types it is accredited for.
- **Regulator**: reviews and resolves flags, and can also be accredited to certify.
- **Platform admin** (superadmin): approves organisations and facilities, publishes templates and runs reports.

Each organisation works on its own subdomain (`acme.chaintrack.app`). Inside an organisation, users have roles; an org admin invites colleagues. Sign-in supports two-factor codes, and signing an action on the chain needs a separate signing PIN.

## Architecture

```
browser ──▶ Caddy (TLS, prod only) ──▶ Next.js web app ──/api──▶ Go API ──▶ PostgreSQL + PostGIS
                                                                    │  ├──▶ Redis (cache, rate limits)
                                                                    │  ├──▶ S3-compatible bucket (RustFS in dev)
                                                                    │  └──▶ IPFS via Pinata (optional)
                                                                    ▼
                                                     chain_outbox table
                                                                    │
                                              relayer ──────────────┴──▶ contracts (Hardhat local / Polygon Amoy)
```

- The **web app** (Next.js, App Router) passes `/api/*` through to the Go API with the browser's session cookie.
- The **API** never talks to the chain when writing. Each change is written to Postgres together with an outbox item in the same transaction.
- The **relayer** sends outbox items to the contracts one at a time, in order, and records the transaction hash.
- **Verification** reads the chain directly. While the chain is down it answers with a warning, never "verified".

## Repository layout

| Path | What is there |
| --- | --- |
| `backend/cmd/api`, `backend/cmd/relayer` | The two Go programs |
| `backend/internal/` | One package per area: `auth` (sessions, roles, 2FA, rate limits), `org`, `user`, `facility`, `template`, `passport` (issue, handovers, transforms, rules), `attestation` (certificates), `flag`, `verify`, `audit`, `notify`, `stats`, and the shared `chain`, `config`, `httpx`, `listq`, `logx`, `upload`, `cache`, `secret`, `dbx`, `mailtmpl`, `server` |
| `backend/internal/chain/bindings/` | Generated Go bindings for the contracts (see below) |
| `backend/db/migrations/` | SQL migrations |
| `backend/api/openapi.yaml` | The API contract; a test fails if a route is missing from it |
| `backend/tests/` | Go tests, including the DB suite against real Postgres and Redis |
| `frontend/` | Next.js app; `e2e/` holds Playwright specs and `__tests__/` holds unit tests |
| `contracts/` | Solidity: `OrgRegistry`, `FacilityRegistry`, `PassportRegistry`, `CustodyManager`, `AttestationRegistry` (Hardhat) |
| `infra/` | Docker Compose files (infra, dev, prod) and the production `Caddyfile` |
| `tools/sim`, `tools/baseline` | Evaluation dataset with fraud cases, and the baseline comparison |
| `docs/` | `design.md` (screens and flows), `operations.md` (settings, deployment, running it) |

## Prerequisites

- Go 1.26
- Node 22 and pnpm 10 (`corepack enable`)
- Docker with the Compose plugin
- uv, only for `make slither`

## Start (dev)

1. Copy `.env.example` to `.env` and fill every value. Use `APP_URL=http://localhost:3000`; organisations then open at `<slug>.localhost:3000`.
2. `make infra` starts Postgres + PostGIS, Redis and the RustFS bucket, and runs the migrations.
3. `make chain` starts the local chain. Then `make deploy` and paste the 5 `*_ADDRESS` lines it prints into `.env`.
4. `make api`, `make relayer` and `make web` run the API on :8080, the relayer, and the web app on :3000.
5. Or `make dev` runs everything in containers instead.

Sign in with `SUPERADMIN_EMAIL` / `SUPERADMIN_PASSWORD` from `.env`; the first superadmin is created on start-up.

## Make targets

| Target | What it does |
| --- | --- |
| `make infra` | Database, migrations, Redis, RustFS (file bucket) |
| `make storage-reset` | Delete every dev upload; the API recreates the bucket on start |
| `make api` / `make web` / `make relayer` | Run each part on this machine |
| `make chain` | A local chain on http://127.0.0.1:8545 (chain id 31337) |
| `make deploy` | Deploy the contracts to the chain in `.env`; prints the address lines |
| `make bindings` | Compile the contracts and regenerate the Go bindings |
| `make dev` / `make prod` | Everything in containers (dev reloads on save; prod uses built images) |
| `make down` / `make prod-down` / `make logs` | Stop the containers, or follow their logs |
| `make test` | Go vet + tests, frontend lint + build + unit tests, contract tests |
| `make test-db` | Backend tests against the real Postgres and Redis, with coverage |
| `make e2e` | Playwright against the running stack |
| `make sim` | Evaluation dataset + fraud cases (`ARGS="-lots 5 -fraud 1"` for a quick run) |
| `make slither` | Static analysis of the contracts; fails on a high finding |

## Checks

- `go vet ./...` and `CGO_ENABLED=0 go test ./...` in `backend/`. For the DB suite run `make test-db`; stop the relayer first, because it shares the dev database.
- `pnpm lint`, `pnpm build`, `node --test "__tests__/**/*.test.ts"` and `pnpm audit --prod` in `frontend/`.
- `pnpm test` in `contracts/`, and `make slither`.
- `make e2e` covers the happy path plus the custody side paths: reject, cancel, certificate issue and revoke, and a flag cleared by a regulator.
- Screenshots: `SCREENS=1 pnpm exec playwright test e2e/screens.spec.ts` in `frontend/`, with the stack running. The images land in `frontend/screens/`.
- `make sim ARGS="-lots 5 -fraud 1"` for a quick evaluation run (see `tools/README.md`).

CI (`.github/workflows/ci.yml`) runs the Go, frontend (including the dependency audit) and contract checks on pushes to `main` and on pull requests.

## Generated code

`backend/internal/chain/bindings/*.go` is generated by `make bindings`, which compiles the contracts with Hardhat and runs go-ethereum's `abigen` on each ABI. Do not edit these files by hand: the next `make bindings` overwrites them. To change a contract, edit its `.sol` file in `contracts/contracts/`, run `make bindings`, and redeploy with `make deploy`.

## Production

1. DNS: point A records for `DOMAIN` and `*.DOMAIN` at the server.
2. Write a `.env` with `APP_ENV=prod`, `APP_URL=https://DOMAIN`, `DOMAIN`, the Polygon Amoy contract addresses, a managed bucket and fresh secrets. Prod refuses Hardhat's test keys.
3. Run `make prod`. Caddy on ports 80 and 443 is the only open entry point, and it gets every certificate itself.

Caddy asks the API (`GET /tls/allowed`) before getting a certificate, so only the platform's own domain and registered organisations' subdomains get one. Full steps are in `docs/operations.md` under "Deploying to a VPS".

## Security notes

- **Sessions:** the session is an HttpOnly cookie (a signed JWT with a token version, so "sign out everywhere" works). Every change must come from the site's own origin.
- **Rate limits:** sign-in is limited per address and per email, with a lockout. Codes, sign-ups and resets are limited too.
- **Two-factor sign-in and signing PIN:** 2FA has recovery codes. The signing PIN is required for every action that goes to the chain.
- **Organisation keys:** each organisation's signing key is encrypted at rest (`ENCRYPTION_KEY`).
- **Database:** every query takes values as bound parameters. List sorting is limited to an allow-list of columns.
- **Visitor addresses:** in production the API and web app publish no ports. Caddy sets the real visitor address, so a visitor cannot fake their own.

See `docs/operations.md` for settings, key handling, lockouts and the audit log.
