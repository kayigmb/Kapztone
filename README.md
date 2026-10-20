# ChainTrack

Blockchain-powered product passports for transparent, tamper-evident shipments (dev).

## Start (dev)

1. Copy `.env.example` to `.env` and fill every value.
2. `make infra` — Postgres + PostGIS, Redis, RustFS bucket, migrations.
3. `make chain` + `make deploy` — local chain, paste 5 `*_ADDRESS` lines into `.env`.
4. `make api` + `make relayer` + `make web` — API on :8080, relayer, web on :3000.
5. `make dev` — everything in containers instead.

## Checks

- `go vet ./...` + `CGO_ENABLED=0 go test ./...` in `backend/` (DB suite: `make test-db`).
- `pnpm lint`, `pnpm build`, `node --test "lib/**/*.test.ts"` in `frontend/`.
- `pnpm test` in `contracts/`, `make slither`, `make e2e` (happy path), `make sim ARGS="-lots 5 -fraud 1"` (quick eval).

Docs: `PLAN.md` (spec), `CHANGELOG.md` (done vs left), `docs/design.md`, `docs/operations.md`, `tools/README.md`.
