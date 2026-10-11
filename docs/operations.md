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
| `PINATA_JWT` | copies certificate files to IPFS | `none` turns it off. A pinned file is public and cannot be withdrawn |
| `SMTP_URL` | sends notification e-mail | `smtp://user:password@host:port?from=address`; `none` turns it off |

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

## Chain: Polygon Amoy

1. Create a key used for nothing else: `openssl rand -hex 32`. Never the test key from `.env.example`.
2. Send its address test POL from the Polygon Amoy faucet.
3. In `.env`: `CHAIN_RPC_URL` (an Amoy RPC URL), `CHAIN_ID=80002`, `CHAIN_PLATFORM_KEY=0x<the key>`.
4. `make deploy`; paste the five printed lines into `.env`.
5. Start from an empty database: organisations and queued items recorded against another chain do not exist on this one. For dev: `make down`, delete the `chaintrack-dev` database volume, `make infra`.
6. `make api`, `make relayer`. Every `tx_hash` written back opens on `https://amoy.polygonscan.com/tx/<hash>`.

Amoy refuses transactions below its minimum priority fee. A `transaction underpriced` error from the deploy or the relayer means the fee has to be set explicitly at that call.

## Contracts: static analysis

`make slither` (needs `uv`) runs Slither over the contracts and fails on a high finding; CI runs the same target. Lower findings are printed and do not fail it: read them after changing a contract.

## Chain: what to watch

```sql
-- Items the contract refused. They are not retried; each needs a person.
SELECT seq, kind, subject_id, last_error FROM chain_outbox WHERE status = 'failed' ORDER BY seq;
-- Items still waiting (the chain is unreachable, or the relayer is not running).
SELECT count(*), min(created_at) FROM chain_outbox WHERE status = 'pending';
```

After changing a contract: `make bindings` regenerates the Go bindings, and the contracts must be deployed again.

## Provenance: what to know when operating it

- **Yields.** Saving a template queues each stage's `maxYield` for the chain. A transform is refused on-chain (`YieldNotSet`) until that item is anchored, so run the relayer after creating or changing a template. A template's commodity and unit cannot change once passports use it; its yields can, and apply to transforms made from then on. Each transform records the yield it was judged by (`transforms.max_yield_ppm`).
- **Split and merge** are transforms at a stage whose `maxYield` is `1`. Give each commodity such a stage if lots are split or merged without processing.
- **Accreditation.** `POST /orgs/{id}/accredit {"accredited": true}` (permission `org:verify`) lets a verified certifier or regulator issue certificates.
- **Revocation.** Resolving a flag as `confirmed` revokes the passport and every passport made from it, each with its own chain transaction. It cannot be undone.
- **Certificates on IPFS.** With `PINATA_JWT` set, a failed pin does not stop the certificate being issued; it is logged (`pin certificate <id>`) and `ipfs_cid` stays empty. Nothing retries it.
- **E-mail.** With `SMTP_URL` set, the API sends each notification once to the organisation's active users. If the mail server is down they wait, for up to a day; older ones are shown in the app only.

```sql
-- Notifications still waiting to be e-mailed.
SELECT count(*), min(created_at) FROM notifications WHERE emailed_at IS NULL;
```

## Public verification

`GET /verify/{passportId}` is open to anyone. The API reads the chain for it, so the API now needs `CHAIN_RPC_URL` reachable as well as the relayer does; while the chain is down every answer is a warning, never "verified". Answers are cached for 60 seconds, so a revocation can take that long to show: `redis-cli --scan --pattern 'verify:*'` then `DEL` clears them. Each address may ask 60 times a minute (the direct peer address; behind nginx that is the proxy until the forwarded address is used).

## Sign-in lockout

Five failed sign-ins for one email from one address lock that pair for 15 minutes. To clear it early, delete the key in Redis: `redis-cli --scan --pattern 'auth:rate:login:*'`, then `DEL` the key.
