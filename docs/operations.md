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
| `JWT_SECRET` | signs sign-in tokens | prod: at least 32 characters (dev: any); changing it signs everyone out |
| `JWT_TTL` | how long a sign-in lasts | whole minutes: `90` is 90 minutes |
| `ENCRYPTION_KEY` | seals 2FA secrets in the database | exactly 64 hex characters (`openssl rand -hex 32`); if it is lost or changed, every user must set 2FA up again |
| `SUPERADMIN_EMAIL`, `SUPERADMIN_PASSWORD` | first platform admin | used only when no superadmin exists; password 12 to 72 characters in prod, up to 72 in dev |
| `APP_ENV` | `dev` or `prod` | `dev` also accepts `?pslug=` to choose the acting organisation |
| `CHAIN_RPC_URL`, `CHAIN_ID` | the chain the relayer writes to | local: `http://127.0.0.1:8545` and `31337` |
| `CHAIN_RELAYER_KEY` | sends every transaction; relayer + operator on-chain | must hold funds on a real network; the value in `.env.example` is a public test key, local only |
| `CHAIN_ADMIN_KEY` | deploys, wires the contracts, grants roles | read only by `make deploy`; outside a local chain it differs from the relayer key and stays offline afterwards |
| `*_ADDRESS` (five) | the deployed contracts | printed by `make deploy` |
| `STORAGE_ENDPOINT`, `STORAGE_REGION`, `STORAGE_BUCKET`, `STORAGE_ACCESS_KEY`, `STORAGE_SECRET_KEY` | the private bucket holding uploaded files | see "File storage" below |
| `STORAGE_PORT` | RustFS's port on this machine (dev) | production ignores it but compose still needs a value |
| `PINATA_JWT` | copies certificate files to IPFS | `none` turns it off. A pinned file is public and cannot be withdrawn |
| `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_USE_TLS` | sends notification e-mail | `SMTP_HOST=none` turns it off (the other four are then not needed). `SMTP_USER` is also the sender. `SMTP_USE_TLS=true` refuses a server without STARTTLS (port 587); required and `true` in prod, optional in dev; port 465 is not supported. Quote a password with spaces |

There is no domain setting. The API takes the frontend's address from each request (the browser's `Origin` header, or the request's own host) and treats its first label as the organisation to act for, if the signed-in user belongs to it.

## File storage

KYB documents and certificate files live in one private S3-compatible bucket, never on the server's disk. Files reach a browser only through the API, after its permission check; the bucket allows no public access. The same code talks to all three; only `.env` changes.

| Where | `STORAGE_ENDPOINT` | `STORAGE_REGION` | Keys | Bucket |
| --- | --- | --- | --- | --- |
| Development | `http://localhost:${STORAGE_PORT}` (RustFS from `make infra`) | `us-east-1` | any pair; RustFS starts with them | created by the API at start |
| AWS S3 | `https://s3.amazonaws.com` | the bucket's region | an IAM user allowed `s3:GetObject`, `s3:PutObject`, `s3:DeleteObject` on that bucket only | create it first, Block Public Access on |
| Google Cloud Storage | `https://storage.googleapis.com` | `auto` | an HMAC key of a service account with Storage Object Admin on that bucket only | create it first, uniform access, public access prevention on |

Outside development the API refuses to start if the bucket does not exist. `make storage-reset` deletes every dev upload (the RustFS volume). The database still names the deleted files, so certificates then show "file missing" on the public page.

## Running

- `make infra` — Postgres, migrations, Redis, RustFS (the dev bucket).
- `make api` — the API on your machine, using `.env`.
- `make dev` / `make prod` — everything in containers.

## Chain: local set-up

1. `make chain` — a local chain (leave it running). Restarting it wipes it: deploy again and clear `chain_outbox`.
2. `make deploy` — deploys the five contracts; paste the five printed lines into `.env`.
3. `make infra`, `make api`, `make relayer` — each in its own terminal.

Running the API or relayer in containers: a chain on your machine is `http://host.docker.internal:8545` from inside a container, not `127.0.0.1`.

## Chain: Polygon Amoy

1. Create two keys used for nothing else (`openssl rand -hex 32` each): an admin key and a relayer key. Never the test keys from `.env.example`.
2. Send both addresses test POL from the Polygon Amoy faucet (the admin pays for the deploy, the relayer for everything after).
3. In `.env`: `CHAIN_RPC_URL` (an Amoy RPC URL), `CHAIN_ID=80002`, `CHAIN_ADMIN_KEY=0x<admin key>`, `CHAIN_RELAYER_KEY=0x<relayer key>`.
4. `make deploy`; paste the five printed lines into `.env`. Then remove `CHAIN_ADMIN_KEY` from the server's `.env` and keep it offline: it is needed again only to grant or revoke roles (`grantRole` / `revokeRole` on OrgRegistry), e.g. to rotate the relayer key.
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
- **E-mail.** With `SMTP_HOST` set, the API sends each notification once to the organisation's active users. If the mail server is down they wait, for up to a day; older ones are shown in the app only.

```sql
-- Notifications still waiting to be e-mailed.
SELECT count(*), min(created_at) FROM notifications WHERE emailed_at IS NULL;
```

## Sessions

A browser's sign-in is a cookie, `ct_session`, that the API sets at `POST /auth/login` and clears at `POST /auth/logout`. It is `HttpOnly` always, `Secure` unless `APP_ENV=dev`, and has no domain: it belongs to the one address the user signed in on, so each organisation's address has its own session. Callers that are not browsers keep using `Authorization: Bearer`.

The browser only talks to the web app, which passes `/api/*` through to the API and adds `X-Forwarded-Host`. The API takes the acting organisation from that host's first label. A request that uses the cookie to change anything must carry an `Origin` equal to that host, or it is refused with 403.

Known limit: the API counts sign-in, registration and verification attempts per calling address, and behind the web app that address is the web app's. Until the API is told to trust a forwarded client address, those limits are shared by everyone.

In dev, `http://acme.localhost:3000` acts for the organisation `acme`; `http://localhost:3000` acts for none.

## Public verification

`GET /verify/{passportId}` is open to anyone. The API reads the chain for it, so the API now needs `CHAIN_RPC_URL` reachable as well as the relayer does; while the chain is down every answer is a warning, never "verified". Answers are cached for 60 seconds, so a revocation can take that long to show: `redis-cli --scan --pattern 'verify:*'` then `DEL` clears them. Each address may ask 60 times a minute (the direct peer address; behind nginx that is the proxy until the forwarded address is used).

## Audit log

Every successful change made through the API (any non-GET request answered 2xx) adds one row to `audit_log`: who acted (none for sign-in and registration), for which organisation, the method, the route pattern, the URL's `{id}`, the status and the time. Request bodies and other URL parts (passwords, PINs, invite tokens) are never recorded. Refused requests and reads are not logged.

- `GET /audit` (`audit:read`): superadmin sees every row, an organisation admin their organisation's. Takes the usual list parameters plus `actor` (user id) and `route` (e.g. `/passports/{id}/handovers`). `?format=csv` downloads every matching row.
- `GET /audit/verify` (superadmin): recomputes the chain and answers `{ok, checked, broken_at}`. Each row's `hash` is SHA-256 of its fields and the previous row's hash, so editing or deleting a row shows as `broken_at` = the first row that no longer matches. Removing only the newest row is not detected. Never edit or delete rows, also not to tidy up: the chain cannot be repaired except by restoring the original values.
- Rows are written after the change commits; a crash between the two loses that one row. A failure to write the row is logged (`audit ...`) and does not fail the request.

## Reports

`GET /reports/volume?start_date=&end_date=` (`report:read`; superadmin all, organisation admin their own): quantity issued at origin per month (UTC), commodity and unit, revoked lots included. Transform outputs are left out, since they are made from lots already counted. `?format=csv` downloads it.

## Sign-in lockout

Five failed sign-ins for one email from one address lock that pair for 15 minutes. To clear it early, delete the key in Redis: `redis-cli --scan --pattern 'auth:rate:login:*'`, then `DEL` the key.

## Visitor addresses and the API port

Rate limits count per visitor address. The web app passes the visitor's address in `X-Forwarded-For` (the last entry it received), and the API trusts that header. That is safe only while the API is never reachable except through the web app: `compose.prod.yml` publishes no API port, and none must ever be opened. Put nginx (with `proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for`) in front of the web app in production; without a proxy, a visitor can send their own header.

## Signing PIN and recovery codes

Every signing action (issue, handover start / accept / reject, transform, retire, certificate issue / revoke, flag resolution) needs the user's 6-digit signing PIN in `X-Signing-PIN`, set with `POST /auth/pin` (needs the password). Five wrong PINs lock signing for 15 minutes; to clear it early, `DEL` the `auth:rate:pin:<user id>` key in Redis.

Enabling 2FA returns 10 recovery codes, shown once. Each signs in once in place of the app code. `POST /auth/2fa/recovery` (with a current app code) replaces them; turning 2FA off removes them.

## Invitations

`POST /invites` e-mails a link valid for 24 hours and usable once. With e-mail off nobody receives it; in dev only, the response and the API log carry the link.
