# ChainTrack

Blockchain-powered product passports for transparent, tamper-evident shipments (dev).

## Start (dev)

1. Copy `.env.example` to `.env` and fill every value (`APP_URL=http://localhost:3000`; organisations open at `<slug>.localhost:3000`).
2. `make infra` — Postgres + PostGIS, Redis, RustFS bucket, migrations.
3. `make chain` + `make deploy` — local chain, paste 5 `*_ADDRESS` lines into `.env`.
4. `make api` + `make relayer` + `make web` — API on :8080, relayer, web on :3000.
5. `make dev` — everything in containers instead.

## Checks

- `go vet ./...` + `CGO_ENABLED=0 go test ./...` in `backend/` (DB suite: `make test-db`; stop the relayer first, it shares the dev database).
- `pnpm lint`, `pnpm build`, `node --test "lib/**/*.test.ts"` in `frontend/`.
- `pnpm test` in `contracts/`, `make slither`, `make e2e` (happy path + reject, cancel, certificate, flag), `make sim ARGS="-lots 5 -fraud 1"` (quick eval).
- Screenshots: `SCREENS=1 pnpm exec playwright test e2e/screens.spec.ts` in `frontend/` (stack running) → `frontend/screens/`.

## Production

1. DNS: A records for `DOMAIN` and `*.DOMAIN` to the server.
2. `.env` with `APP_ENV=prod`, `APP_URL=https://DOMAIN`, `DOMAIN`, the Amoy addresses and real secrets.
3. `make prod` — Caddy on 80/443 is the only open port and gets every certificate itself.

Full steps: `docs/operations.md` → Deploying to a VPS.

Docs: `PLAN.md` (spec), `CHANGELOG.md` (done vs left), `docs/design.md`, `docs/operations.md`, `tools/README.md`.
