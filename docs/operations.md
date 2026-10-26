# Operations notes

## Reset a user's two-factor sign-in (lost device)

Enabling 2FA returns 10 recovery codes (shown once); each signs in once in place of the app code. `POST /auth/2fa/recovery` (with a current app code) replaces them. If the codes are also lost, a platform operator resets 2FA by hand, after confirming the person's identity out of band:

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
| `JWT_TTL_REMEMBER` | how long a sign-in lasts with "Keep me signed in" | whole minutes: `43200` is 30 days |
| `ENCRYPTION_KEY` | seals 2FA secrets in the database | exactly 64 hex characters (`openssl rand -hex 32`); if it is lost or changed, every user must set 2FA up again |
| `SUPERADMIN_EMAIL`, `SUPERADMIN_PASSWORD` | first platform admin | used only when no superadmin exists; password 12 to 72 characters in prod, up to 72 in dev |
| `APP_ENV` | `dev` or `prod` | `dev` also accepts `?pslug=` to choose the acting organisation |
| `APP_URL` | the web app's bare address, no path | e-mailed links (reset, verify, lock, invite) start with it, never with a request header; an invite puts the organisation's slug in front of the host (`https://acme.chaintrack.app`) |
| `RATE_LIMIT` | `true` or `false`: per-address and per-account rate limits (sign-in, PIN, 2FA, public verify) | empty: on in prod, off in dev |
| `WEB_PORT` | Host port of the web app in production, bound to 127.0.0.1 (default 3000) | `make prod` only; must match the port in the server's Caddyfile |
| `CHAIN_RPC_URL`, `CHAIN_ID` | the chain the relayer writes to | local: `http://127.0.0.1:8545` and `31337` |
| `CHAIN_RELAYER_KEY` | sends every transaction; relayer + operator on-chain | must hold funds on a real network; the value in `.env.example` is a public test key, local only |
| `CHAIN_ADMIN_KEY` | deploys, wires the contracts, grants roles | read only by `make deploy`; outside a local chain it differs from the relayer key and stays offline afterwards |
| `*_ADDRESS` (five) | the deployed contracts | printed by `make deploy` |
| `STORAGE_ENDPOINT`, `STORAGE_REGION`, `STORAGE_BUCKET`, `STORAGE_ACCESS_KEY`, `STORAGE_SECRET_KEY` | the private bucket holding uploaded files | see "File storage" below |
| `STORAGE_PORT` | RustFS's port on this machine (dev) | production ignores it but compose still needs a value |
| `PINATA_JWT` | copies certificate files to IPFS | `none` turns it off. A pinned file is public and cannot be withdrawn |
| `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_USE_TLS`, `MAIL_FROM` | sends notification e-mail | `SMTP_HOST=none` turns it off (the rest are then not needed). `SMTP_USER` is only the login; every recipient sees `MAIL_FROM`, e.g. `ChainTrack <noreply@chaintrack.digital>` (name and address, validated at start-up). With Gmail, add that address under "Send mail as" or Gmail shows the login address; a custom domain needs SPF (and ideally DKIM). The logo travels inside every e-mail as an inline PNG (`cid:logo`, from `backend/internal/mailtmpl/logo.png`), so nothing is hosted. `SMTP_USE_TLS=true` refuses a server without STARTTLS (port 587); required and `true` in prod, optional in dev; port 465 is not supported. Quote a value with spaces |

The API has no domain setting beyond `APP_URL`. The API takes the frontend's address from each request (the browser's `Origin` header, or the request's own host) and treats its first label as the organisation to act for, if the signed-in user belongs to it.

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
- `make dev` — a local chain with the contracts deployed, API, relayer and infra in containers (run `make web` beside it); `make prod` — everything in containers.

## Deploying to a VPS

`make prod` runs everything in containers; only the web app is published, on `127.0.0.1:WEB_PORT`. TLS is the server's own Caddy, shared with other apps: paste `infra/Caddyfile` into its Caddyfile (domain and port filled in) and reload it. Organisations live on their own subdomains. DNS needs only two A records, `DOMAIN` and `*.DOMAIN`, at any provider. Caddy gets a Let's Encrypt certificate for each host on its first visit (HTTP challenge) and renews it; no certbot, no DNS keys. Before each one it asks the API (`GET /tls/allowed?domain=`), which allows `DOMAIN`, the subdomains of registered organisations, and up to 3 well-formed unknown subdomains a day (a Redis counter; Let's Encrypt allows 50 new certificates per domain a week, so random names can't use them up). An unknown subdomain within that allowance opens over https and shows the web app's "no such organisation" page: the web app asks the API (`GET /address`) on each page request and caches the answer per host for a minute. Past the allowance an unknown name gets no certificate and the browser shows a connection error. Certificates live in Caddy's data directory; keep it, or every restart re-issues them and runs into Let's Encrypt's limits.

1. A Linux server with Docker and the Compose plugin; clone the repository there.
2. DNS: two A records to the server, `DOMAIN` and `*.DOMAIN`.
3. `.env` from `.env.example`, with production values: `APP_ENV=prod`, `APP_URL=https://DOMAIN`, new `JWT_SECRET` and `ENCRYPTION_KEY`, a managed bucket (File storage), `SMTP_USE_TLS=true`, and the Amoy addresses (Chain: Polygon Amoy). Prod refuses Hardhat's test keys.
4. `make prod`, then add `infra/Caddyfile` to the server's Caddy and reload it. Check `https://DOMAIN` and `https://DOMAIN/api/healthz` (shows `"mode":"prod"`). The first visit to a new organisation's subdomain takes a few seconds while its certificate is issued.
5. Firewall: open 22, 80 and 443 only (80 carries the certificate challenge and the redirect to https). The database and Redis bind to 127.0.0.1; the API and web app publish nothing.

## Chain: local set-up

1. `make chain` — a local chain (leave it running). Restarting it wipes it: deploy again and clear `chain_outbox`.
2. `make deploy` — deploys the five contracts; paste the five printed lines into `.env`.
3. `make infra`, `make api`, `make relayer` — each in its own terminal.

4. Optional: `make sim` — the evaluation dataset and fraud cases (`tools/README.md`).

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

- **Templates: draft, publish, version.** `PUT /templates/{id}` saves the next version as a draft (`draft` in `GET`); every check keeps using the published rules until `POST /templates/{id}/publish`, which makes the draft live as `version + 1` (409 without a draft). Transforms and handovers record the version they were judged by (`template_version`, filled in by a database trigger).
- **Yields.** Creating or publishing a template queues each stage's `maxYield` for the chain, plus the reserved stage `split` at yield 1 (no template may name a stage `split`). A transform is refused on-chain (`YieldNotSet`) until that item is anchored, so run the relayer after creating or publishing a template. A template's commodity and unit cannot change once passports use it; its yields can, and apply to transforms made from then on. Each transform records the yield it was judged by (`transforms.max_yield_ppm`). A template published before `split` existed gets its split yield queued at its first partial handover.
- **Freestyle moves.** Lots move any way: a template only names the product (commodity, unit, origin facility kind, capacity rule) and the `requiredEvidence` the origin facility must hold as valid certificates before issuing. Templates have no stages (migration 000028 dropped them, keeping the ISSUE stage's evidence). A handover goes to any verified organisation with an optional free `stage` label (transport, storage, sale…); a completed handover labelled `export` counts as exported in the volume report. A transform takes an optional label (default `processing`) and always runs at the `split` yield of 1: outputs never exceed inputs, on-chain too.
- **Partial handover.** `POST /passports/{id}/handovers` takes `quantity` (default: all that remains). Less than the remainder first records a split transform (one input, one child, yield 1, stage `split`) in the same transaction; the child is handed over (`child_passport_id`, `split_from_passport_id`) and the parent keeps the rest. If the handover is rejected, cancelled or lapses, the child simply stays with the sender. Certificates stay on the parent: the child starts with none.
- **Handover photos and place.** The sender's start and the receiver's accept take JSON, or a multipart form with up to 6 PNG / JPEG `photos` (10 MB each) in the bucket. `lat` / `lon` are optional, together. The sender's place and photo hashes are inside the signed record hash; the receiver's are kept beside it. `GET /handovers/{id}/photos/{photoId}`: the two parties only.
- **Weight tolerance.** Accepting needs `received_quantity`. If it differs from what was sent by more than 1% (`template.DefaultHandoverTolerance`), the answer is 409 `{error, difference, tolerance, sent, received}`; reject or flag instead. Accepted with a difference above half the tolerance, the rules raise a flag.
- **Rejecting** takes `reasons` (one or more of `weight_short`, `damaged`, `wrong_product`, `seal_broken`, `not_expected`), an optional `note`, and `also_flag`, which raises a flag about the passport in the same transaction.
- **Facilities.** A new facility needs `licence_valid_to` and waits as `pending_approval`; it is put on the ledger only when approved (`POST /facilities/{id}/approve`, permission `facility:suspend`). It cannot originate goods while pending or after its licence's last day. The licence document: `POST /facilities/{id}/licence` (multipart `file`, optional new `licence_valid_to`), `GET` to download. A boundary that touches another live facility is refused with 409: the caller's own facility is named (`{error: "overlaps <name>", facility_id}`), another organisation's is not (`facility_id: null`).
- **Rules flags.** The rules raise a flag (`raised_by: "rules"`) on an accepted handover whose difference exceeds half the tolerance. They are notified like any flag. Investigators set who looks into one with `PUT /flags/{id}/assignee {"user_id": "…" | null}`; the assignee must be a platform admin or a member of a verified regulator.
- **Accreditation.** `GET/PUT /orgs/{id}/accreditations` (full per-type set, permission `org:verify`) lets a verified certifier or regulator issue certificates of that type inside its window.
- **Revocation.** Resolving a flag as `confirmed` revokes the passport and every passport made from it, each with its own chain transaction. It cannot be undone.
- **Certificates on IPFS.** With `PINATA_JWT` set, a failed pin does not stop the certificate being issued; it is retried by the API with exponential back-off (1 min to 6 h, 10 tries).
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

`GET /verify/{passportId}` is open to anyone. The API reads the chain for it, so the API now needs `CHAIN_RPC_URL` reachable as well as the relayer does; while the chain is down every answer is a warning, never "verified". Answers are cached for 60 seconds, so a revocation can take that long to show: `redis-cli --scan --pattern 'verify:*'` then `DEL` clears them. Each address may ask 60 times a minute (the visitor's address as Caddy reports it; see "Visitor addresses").

## Audit log

Every successful change made through the API (any non-GET request answered 2xx) adds one row to `audit_log`: who acted (none for sign-in and registration), for which organisation, the method, the route pattern, the URL's `{id}`, the status and the time. Request bodies and other URL parts (passwords, PINs, invite tokens) are never recorded. Refused requests and reads are not logged.

- `GET /audit` (`audit:read`): superadmin sees every row, an organisation admin their organisation's. Takes the usual list parameters plus `actor` (user id) and `route` (e.g. `/passports/{id}/handovers`). `?format=csv` downloads every matching row.
- `GET /audit/verify` (superadmin): recomputes the chain and answers `{ok, checked, broken_at}`. Each row's `hash` is SHA-256 of its fields and the previous row's hash, so editing or deleting a row shows as `broken_at` = the first row that no longer matches. Removing only the newest row is not detected. Never edit or delete rows, also not to tidy up: the chain cannot be repaired except by restoring the original values.
- Rows are written after the change commits; a crash between the two loses that one row. A failure to write the row is logged (`audit ...`) and does not fail the request.

## Reports

`GET /reports/volume?start_date=&end_date=` (`report:read`; superadmin all, organisation admin their own): quantity issued at origin per month (UTC), commodity and unit, revoked lots included. Transform outputs are left out, since they are made from lots already counted. `exported` is the quantity of completed handovers, sent by the organisation, at a stage that allows `EXPORT`, by the month they were accepted. `?format=csv` downloads it.

`GET /reports/blocked?start_date=&end_date=` (`report:read`): `{data: [{rule, count}]}` for every rule, zero included: `capacity`, `mass_balance`, `evidence`, `tolerance`, `handover_rejected`. The first four are counted in `rule_events` when a rule refuses a request (written outside the refused request's rolled-back transaction); rejected handovers are counted from the handovers, for both parties. `?format=csv` downloads it.

## Password reset, sign-out everywhere, lock

- **Token version:** every user has a token version. Each sign-in token (and trusted-device cookie) carries it, and a token whose version is out of date is refused. Changing or resetting a password, "Sign out other devices" (`POST /auth/sessions/revoke`) and "Lock my account" all bump it, so every earlier session ends at once. The caller of a change or a sign-out gets a new token and stays signed in.
- **Forgot password:** `POST /auth/password/forgot` always answers 202, whether or not the email has an account. An account gets an emailed link `<site>/reset/<token>` that works once, for 30 minutes. In dev the link is also written to the API log. With `SMTP_HOST=none`, the log is the only place it appears, and only in dev.
- **Reset:** `POST /auth/password/reset`. A new password must be 12–72 characters and not one of the last 5. After any password change the user gets an email with a "Lock my account" link (`<site>/lock/<token>`, once, 7 days). That link disables the account; re-enable it by setting `users.status` back to `active`.
- **Keep me signed in:** that sign-in lasts `JWT_TTL_REMEMBER` minutes instead of `JWT_TTL`.
- **Trust this device:** after a 2FA sign-in with an app code, the `ct_trust` cookie (HttpOnly, per address) skips the code for 30 days. It ends early on any token-version bump.
- Only hashes of the reset and lock tokens are stored, in `password_resets`.

## Modes

`DEV=true` is development: no rate limits at all (sign-in, sign-up, verification, signing PIN), `?pslug=` names the organisation, dev links are logged. Anything else is production, with every limit on. The API's start-up line says `mode=dev|prod rate_limits=false|true`, and `GET /healthz` answers `{"status":"ok","mode":"dev"}` in its `data`.

## Switching organisation (handoff)

Sessions are host-only cookies, one per address. Switching organisation hands the session over instead of asking for the password again: the current address calls `POST /auth/handoff {slug}` (the person must belong to that organisation; `""` is the bare address), gets a code valid for one minute and one use, and opens `<slug>.<host>/handoff?code=…`; that page spends it with `POST /auth/handoff/accept`. Only the hash of code + target slug is stored (in `password_resets`, kind `handoff`), so a code works only at the address it was made for.

## Sign-in lockout

Production only. Five failed sign-ins for one email from one address lock that pair for 15 minutes. A failed sign-in says how many attempts are left (also `attempts_left` in `data`). To clear it early, delete the key in Redis: `redis-cli --scan --pattern 'auth:rate:login:*'`, then `DEL` the key.

## Visitor addresses and the API port

Rate limits count per visitor address. In production Caddy sets `X-Forwarded-For` to the real visitor address, dropping what the visitor sent; the web app passes on the last entry, and the API trusts that header. A visitor's own `X-Forwarded-For` therefore changes nothing. That holds only while the API and the web app are reachable through Caddy alone: `compose.prod.yml` publishes the web app on 127.0.0.1 only and the API not at all; neither must ever be opened to the outside.

In dev nothing sits in front of the web app, so a request may name its own address. The e2e tests and the sim use that to stay under the per-address limits; behind Caddy it no longer works, which is the point.

## Signing PIN and recovery codes

Every signing action (issue, handover start / accept / reject, transform, retire, certificate issue / revoke, flag resolution) needs the user's 6-digit signing PIN in `X-Signing-PIN`, set with `POST /auth/pin` (needs the password). Five wrong PINs lock signing for 15 minutes; to clear it early, `DEL` the `auth:rate:pin:<user id>` key in Redis.

Enabling 2FA returns 10 recovery codes, shown once. Each signs in once in place of the app code. `POST /auth/2fa/recovery` (with a current app code) replaces them; turning 2FA off removes them.

## Invitations

`POST /invites` e-mails a link valid for 24 hours and usable once. With e-mail off nobody receives it; in dev only, the response and the API log carry the link.

## Logs

The API and the relayer write one line per event to stderr, as JSON (`LOG_FORMAT=json`, the default, also in `.env.example`) or readable text (`LOG_FORMAT=text`); `LOG_LEVEL` is `debug`, `info` (default), `warn` or `error`. On a terminal the level is coloured in either format: INFO green, WARN yellow, ERROR red (`NO_COLOR=1` turns it off); piped or collected output stays plain.

- Every request gets an id (`X-Request-ID`): the web app makes it, the API logs it on every line of that request and sends it back. A person who sees "quote reference …" on a server error gives you the id to search for.
- One access line per request, field names fixed: `request_id`, `method`, `route` (the pattern, e.g. `/passports/{id}/handovers`), `path`, `status`, `bytes`, `duration_ms`, `ip`, `user_id`, `org`, and `error` + `error_at` when the handler failed. 5xx lines are `ERROR`, 4xx `WARN`, the rest `INFO`. A panic logs its stack with the request id.
- The relayer logs each sealed record (`outbox_id`, `kind`, `subject_id`, `tx_hash`) and each refusal or retry. The mailer, cache and IPFS retries log their failures with the ids involved.
- Find one request: `jq 'select(.request_id=="…")'`; slow ones: `jq 'select(.duration_ms > 1000)'`; server errors: `jq 'select(.status >= 500)'`.

## IPFS copies of certificates

With `PINATA_JWT` set, each issued certificate's file is also pinned to IPFS. A failed pin doesn't block the certificate (its hash on the chain is the proof); it is retried by the API every minute once due, waiting 1 min, 2, 4… up to 6 h between tries, 10 tries at most (`attestations.ipfs_attempts`, `ipfs_next_at`, `ipfs_last_error`).

## Report downloads

`/reports/volume`, `/reports/blocked` and `/audit` take `format=csv` or `format=pdf` (landscape A4, up to 5,000 rows; past that the PDF says how many were left out and the CSV has them all).
