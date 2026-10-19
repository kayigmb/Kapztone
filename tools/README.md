# Tools

Evaluation tooling (PLAN.md §6.1). `baseline/` (the document-based comparison system) is not built yet.

## `sim/` — dataset and fraud cases

`make sim` builds the evaluation dataset against the running stack, injects the fraud cases and writes what ChainTrack made of each one to `tools/sim/sim-results.csv`.

Needs: `make infra`, `make api`, `make chain` + `make deploy`, `make relayer` (the sim waits for the relayer to seal every record before it tampers with anything). Reads `SUPERADMIN_*`, `API_URL` and `DATABASE_URL` from `.env`.

| Flag | Default | |
| --- | --- | --- |
| `-seed` | 42 | Every random choice comes from it: same seed, same dataset and fraud targets |
| `-lots` | 100 | Lots per chain |
| `-facilities` | 20 | Facilities per chain |
| `-fraud` | 5 | Cases per chain per fraud type |
| `-out` | `sim-results.csv` | |
| `-run` | time-based | Keeps names, slugs and commodities unique across runs |
| `-seal-timeout` | 15m | How long to wait for the relayer |

Quick run: `make sim ARGS="-lots 5 -facilities 3 -fraud 1"`.

**Chains** are the template files in `sim/templates/` (coffee, tin, medicine). They are data: a new commodity is a new file and a line in `chains` in `main.go`. The medicine template was added this way, with no change to the API or the app.

**Per chain:** 4 producers owning the facilities, one business per later stage, one shared certifier accredited for every evidence key. Each lot is issued, gets each stage's required evidence from the certifier, is handed over and accepted at every stage, and transformed where the stage allows it.

**Fraud cases** (5 per chain per type):

| Type | How | Expected |
| --- | --- | --- |
| F1 record alteration | SQL: change a sealed passport's quantity | detected |
| F2 unilateral / forged handover | alternately: a third business accepts a handover not addressed to it (API); or SQL moves the holder with no handover | prevented / detected |
| F3 unverified origin | API: issue from a facility not yet approved | prevented |
| F4 quantity inflation | API: issue over the facility's capacity, or transform over the stage's maximum yield | prevented |
| F5 cloned passport | SQL: copy a sealed passport under a new id and code, seal included | detected |
| F6 invalid certificate | SQL: swap a sealed certificate's fingerprint | detected |

**CSV:** `case_id, chain, kind (legit | F1–F6), mechanism (api | db | none), passport_id, code, http_status, verdict, reasons, outcome (prevented | detected | missed | clean | false_alarm), baseline_outcome` (empty: for the baseline reviewers).

The sim sends a fresh `X-Forwarded-For` on every request to stay under the API's per-address limits; that works only without a proxy in front (dev).

Unit tests (plan reproducible, outcome rules): `cd tools/sim && go test ./...`.
