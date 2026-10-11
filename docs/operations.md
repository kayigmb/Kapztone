# Operations notes

## Reset a user's two-factor sign-in (lost device)

There are no recovery codes yet. Until user management lands, a platform operator resets 2FA by hand, after confirming the person's identity out of band:

```sql
UPDATE users SET totp_enabled = false, totp_secret = '', totp_last_step = 0
WHERE email = 'person@example.com';
```

The user then signs in with their password and sets 2FA up again (`POST /auth/2fa/setup`, then `/auth/2fa/enable`).

## Settings

Every setting comes from `.env` (see `.env.example`); nothing has a default. A missing or invalid value stops `make` and the API with a message naming it. Setting names are declared once, in `backend/internal/config/config.go`.

| Variable | Purpose | Notes |
| --- | --- | --- |
| `JWT_SECRET` | signs sign-in tokens | at least 32 characters; changing it signs everyone out |
| `JWT_TTL` | how long a sign-in lasts | whole minutes: `90` is 90 minutes |
| `ENCRYPTION_KEY` | seals 2FA secrets in the database | exactly 64 hex characters (`openssl rand -hex 32`); if it is lost or changed, every user must set 2FA up again |
| `SUPERADMIN_EMAIL`, `SUPERADMIN_PASSWORD` | first platform admin | used only when no superadmin exists; password 12 to 72 characters |
| `APP_ENV` | `dev` or `prod` | `dev` also accepts `?pslug=` to choose the acting organisation |
| `CHAIN_RPC_URL`, `CHAIN_ID` | the chain the relayer writes to | local: `http://127.0.0.1:8545` and `31337` |
| `CHAIN_PLATFORM_KEY` | sends every transaction; admin and relayer on-chain | must hold funds on a real network; the value in `.env.example` is a public test key, local only |
| `*_ADDRESS` (five) | the deployed contracts | printed by `make deploy` |

There is no domain setting. The API takes the frontend's address from each request (the browser's `Origin` header, or the request's own host) and treats its first label as the organisation to act for, if the signed-in user belongs to it.

## Running

- `make infra` — Postgres, migrations, Redis.
- `make api` — the API on your machine, using `.env`.
- `make dev` / `make prod` — everything in containers.

## Chain: local set-up

1. `make chain` — a local chain (leave it running). Restarting it wipes it: deploy again and clear `chain_outbox`.
2. `make deploy` — deploys the five contracts; paste the five printed lines into `.env`.
3. `make infra`, `make api`, `make relayer` — each in its own terminal.

Running the API or relayer in containers: a chain on your machine is `http://host.docker.internal:8545` from inside a container, not `127.0.0.1`.

For Polygon Amoy set `CHAIN_RPC_URL`, `CHAIN_ID=80002` and a funded `CHAIN_PLATFORM_KEY`, then `make deploy`.

## Chain: what to watch

```sql
-- Items the contract refused. They are not retried; each needs a person.
SELECT seq, kind, subject_id, last_error FROM chain_outbox WHERE status = 'failed' ORDER BY seq;
-- Items still waiting (the chain is unreachable, or the relayer is not running).
SELECT count(*), min(created_at) FROM chain_outbox WHERE status = 'pending';
```

After changing a contract: `make bindings` regenerates the Go bindings, and the contracts must be deployed again.

## Sign-in lockout

Five failed sign-ins for one email from one address lock that pair for 15 minutes. To clear it early, delete the key in Redis: `redis-cli --scan --pattern 'auth:rate:login:*'`, then `DEL` the key.
