# Patterns

How screens are put together and how the core flows behave.

## Routes (Next.js App Router)

```
app/
  (public)/p/[lotCode]/page.tsx          Public verification page (no login, opened from QR). Client component reading `useQuery(["verify", id])`.
  (auth)/signup/…  (auth)/signin/…         Onboarding (phone-first, no shell)
  (app)/layout.tsx                         App shell: sidebar / tab bar
  (app)/page.tsx                           Dashboard (role-based)
  (app)/passports/[id]/…  handovers/…  facilities/…  transform/…  certificates/…  flags/…
  (app)/issue/…  (app)/handover/…          Focused flows: no tab bar, back + close, stepper, action bar
  (admin)/…                                Organisations, issuers, templates, users, reports
```

## App shell (Figma `BusinessApp`)

- **≥ lg:** `grid grid-cols-[256px_1fr] min-h-dvh` (76 px when folded): the **navy sidebar** (brand, organisation card, switcher, nav groups **Workspace / Trust & compliance / Manage** with live counts, the person at the bottom → Profile, Sign out) | top bar 64 px + content. `navItems` in `lib/consts.ts` gives each item its `group`; only built screens are listed.
- **Content:** `mx-auto w-full max-w-content px-3 py-5 sm:px-4 lg:px-7 lg:py-8 grid content-start gap-4 lg:gap-5`.
- **Top bar:** eyebrow "Workspace" over the page title on desktop; back arrow when not on Home; bell, theme toggle, "Scan QR" outline button.
- **< lg:** top bar 56 px (brand mark, back, title, bell, theme) + content + the **four-tab bar** (Home, first workspace screen, Scan, More). More opens a navy sheet with the organisation card, switcher, every nav group, the person and Sign out. No hamburger.
- **Focused flows** (issue passport, handover, register facility, onboarding): stay in the shell; a page header, then the Stepper, then one card per step (Figma `.narrow-flow` at `max-w-[650px]`, or `.form-layout` form left / live checks right on desktop). Buttons right-aligned in the card's footer row (`.form-actions`: `border-t px-6 py-4 flex justify-end gap-2`); on phones the sticky action bar.

## Page header (Figma `.app-page-header`)

`flex flex-wrap items-start justify-between gap-4`: title `text-[22px] leading-7 font-semibold tracking-tight lg:text-[26px]`, subtitle `text-sm text-ink-muted` (often a status pill + organisation name), the one primary action on the right (on phones a 44 px square with just its icon, the label as `aria-label`). Dashboards greet: "Good morning, Jeanne".

## Split queue (Figma `.split-queue`)

Queues that are read one item at a time (flags, organisation verification, incoming handovers on desktop): a list card on the left (search row on top, rows with an icon square, title, code and caption, the selected row `bg-primary-soft` with a 3 px inset teal edge) and the selected item's investigation panel on the right (`components.md` → Investigation panel), ending in a decision row. Tabs above the split for status ("Organisations 4 · Approved · Rejected"), as the underlined `Tabs`. Phones: the list, then the panel, in one column.

## Flows with a stepper

A flow that takes several decisions (register facility, issue, hand over, receive, onboarding) shows `Stepper` under the page header, names each step by its decision ("Facility · Quantity · Evidence · Sign · Done"), and keeps one primary button per step that names what it does next ("Continue to evidence"). A blocking rule turns the current step red and disables the button with the reason beside it. The last step is signing (`patterns.md` → Signing).

## Layout grids

- Two-column work layouts on desktop: main `1.6fr` + side `1fr` (`lg:grid-cols-[1.6fr_1fr]`). Queue + detail: `lg:grid-cols-[380px_1fr]` (list left, selected item right). Equal halves for input/output (transform).
- Everything stacks to one column below `lg` in the app, below `md` on the public page.
- `min-w-0` on grid children so long codes don't blow the layout.

## Role-based dashboards

Same components, different emphasis. Top: 4 StatTiles. Then quick actions (`grid grid-cols-2 lg:grid-cols-4 gap-2`, each `h-tap-scan` outlined button with icon; the first one primary). Then work queues.

| Role | Stat tiles | Quick actions (first = primary) | Main column | Side column |
| --- | --- | --- | --- | --- |
| Producer | Passports you hold · Waiting for your signature (attention) · Incoming to accept · Open flags | Issue passport · Scan QR · Hand over · Register facility | Waiting for your signature · Passports you hold | Capacity this season (meter per facility) · Certificates |
| Processor | same | Scan QR · Transform · Hand over · Issue (if allowed) | Incoming handovers · Lots to process | Mass balance this week · Certificates |
| Exporter / importer | Passports you hold · Waiting for your signature · Incoming to accept (attention) · Open flags (alert if > 0) | Scan QR · Hand over · Transform · Prepare export | Incoming handovers · Passports you hold (table) | Open flag summary · Ready to export checklist |
| Regulator / certifier | Open flags (alert) · Organisations to verify (attention) · Certificates expiring · Lots verified this month | Review flags · Scan QR · Issue certificate · Run report | Flag queue (by severity) · Organisations waiting | Certificates expiring/revoked · Verified volume by commodity (horizontal bars) |
| Admin | Organisations to verify · Issuers to accredit · Active users · Open flags | Review organisations · Templates · Users · Reports | Verification queue | Audit highlights |

## Signing

Every record that changes custody or creates material is signed, and signing is a deliberate last step:

1. **Summary** of exactly what is being signed (a `dl` summary card: Lot, Quantity, From, To, Where, Evidence).
2. One sentence on what the signature means: "You confirm these goods left your custody. Akagera Coffee Exports Ltd must inspect and sign before the handover is complete." / "Your signature records this harvest on ChainTrack. It can't be edited afterwards; mistakes are corrected with a new, linked record."
3. **PIN entry**: a panel `border-2 border-primary bg-primary-soft rounded-lg p-4` titled "Sign with your PIN" (`PenLine`), 6 boxes `w-11 h-13 font-mono text-[22px]`. Use device biometrics (WebAuthn) when available, PIN as fallback.
4. Button names the result: "Sign and issue passport", "Sign and send", "Accept and sign", "Sign and create 2 passports".
5. Result screen: big round mark (`size-18 rounded-full`), title, one-line outcome, the status chips, a vertical "what happens next" step list, then next actions.

Never let someone sign on another person's behalf; org admins can't either.

## Dual-signature handover

States and chips, in order:

| Moment | Sender sees | Receiver sees | Custody |
| --- | --- | --- | --- |
| Sender signed | Signed by sender · Awaiting receiver signature | Awaiting your signature | Sender ("in transit") |
| Receiver accepted | Signed by both parties | Signed by both parties | Receiver |
| Receiver rejected | Rejected by receiver + flag ID | Rejected by receiver | Stays with sender, frozen until the flag is resolved |

- **Sender flow (phone):** Scan QR → Confirm lot (passport card + "You hold this lot and nothing is pending on it") → Details (all/part, quantity, location from GPS within 2 km of a registered facility, seal number, ≥ 2 photos) → Receiver (search; only verified organisations selectable; unverified shown disabled with the reason) → Sign → Sent (vertical steps: You signed ✓ · Receiver inspects and signs ⏱ · Custody moves to them).
- **Receiver flow (phone):** Notification → Incoming list (row selected, "Awaiting your signature") → Scan on arrival (label found → "Use this lot") → **Inspect**: weight received input, **expected vs. received** comparison panel (two big mono numbers; footer green "Matches what was sent" / red "40 kg short (3.3%). Above the 0.5% tolerance."), inspection checklist (seal number matches, count and labels match, no damage), photos → Accept and sign **or** Reject / raise flag.
- **Reject sheet:** reason chips (Quantity short, Seal doesn't match, Wrong product, Damaged, multi-select), notes textarea for sender and regulator, evidence summary ("3 photos and the scale reading are attached"), Cancel + "Reject and flag" (solid danger). Result: "Handover rejected · Flag FLG-2026-0163 is open. The lot stays with Kitabi Washing Station until it's resolved."
- When the difference is within tolerance, the comparison is green and Accept is primary. When outside, Reject is primary-placed (right) as `danger-outline` and "Accept anyway" is secondary.

## Blocking checks (fraud prevention in the UI)

| Check | Where | What the UI does |
| --- | --- | --- |
| Capacity | Issue passport, quantity step | Live CapacityMeter as you type, including the new quantity as a lighter segment. Over: input error, stepper step red, meter hatched, blocking notice "This passport can't be issued. Lower the quantity to 1,200 kg or less. If the harvest really is larger, ask the regulator to review the declared capacity.", Continue disabled with "Continue is blocked until the quantity fits." The attempt is logged. |
| Mass balance | Transform / split / merge | Live MassBalanceMeter: total outputs vs. inputs × allowed yield. Over: meter card `border-2 border-danger`, first output field error "Total too high", sign button disabled with "Fix the mass balance to continue". |
| Licence | Facility pickers | Facilities with expired licences are disabled choice cards: "Licence expired 30 Jun 2026. Renew to issue." |
| Verified receivers | Handover receiver picker | Unverified orgs disabled: "Not verified yet: can't receive". |
| Holder | Handover confirm lot | Only the current holder can hand over; lots with an open flag show the flag instead of the button. |
| Tolerance | Receiver inspect | Comparison panel red above tolerance; reject is the expected path. |

Messages always give the number, the limit and the next step.

## Transform / split / merge

Desktop: two columns. Left: **Inputs** list (checkbox, lot code mono, description, quantity) with "2 selected · 6,000 kg" in the header and "Scan or add another input". Right: **New output lots** card with a segmented control (Transform | Split | Merge), output rows (name + quantity with kg addon + remove), "Add output lot"; then the meter card. Bottom right: "Sign and create 2 passports".
Success: compact verified VerdictBadge "2 new passports created · 6,000 kg in → 1,200 kg out · yield 20.0% of 21% allowed · input passports closed", the new PassportCards, and a GenealogyTree showing them linked to their parents. Actions: Print labels, Hand over.

## Destructive actions (revoke certificate, confirm fraud, suspend organisation)

Always a dialog (desktop) or sheet (phone) that:
1. Asks a question naming the object ("Confirm fraud on FLG-2026-0147?").
2. Lists every lot or record that will be affected, by code ("3 passports will be revoked").
3. Requires a reason; shows optional linked actions as checkboxes (revoke export permit, suspend organisation 90 days).
4. States the public consequence in a blocking notice ("Public verification pages for these lots will show **Failed** immediately. Holders and their buyers are notified.").
5. Ends with Cancel + a solid danger button naming the action and count.

Clearing a flag is non-destructive (secondary button), "Request information" is secondary, "Confirm fraud" is the only danger button.

## Public verification page (Figma `PublicVerification`; drafts 1a, 1b, 1c)

Built in `app/(public)/p/[id]/verification-view.tsx`; `id` is a lot code or a passport id.

1. **Top bar** (white, 64 px): Brand; theme toggle, "Scan another" (outline, from `sm`), "Business sign in" link (from `md`).
2. **Navy hero** (`bg-navy-900`, the faint `hero-grid` and a teal glow, both decoration): breadcrumb "Public verification › {lot code}", the verdict (round mark + "Verification result" + Verified / Warning / Failed + summary + failing reasons), and on the right "Checked on" and "Records checked". Failed and Warning use the same layout in their own tone.
3. **Failed only:** the blocking Notice "Do not accept, pay for or ship this lot." under the hero.
4. **Two columns** on `lg` (`minmax(0,1.75fr) minmax(300px,0.85fr)`), one on phones:
   - left: **product card** (product art panel with beans for coffee or the commodity icon, origin-kind tag; commodity pill + passport status pill; name; lot code with copy; quantity, holder, last updated) → **Journey / Genealogy** tabs (timeline beside the route art, which is a drawing that names origin and current holder, never a located map; the tree of earlier lots with "stage · quantity used" on each connector) → **certificates** list;
   - right: **What we checked** (n of 7 passed) → **origin** card(s) (facility, kind, operator, capacity result) → navy **How verification works** card → **Technical details** (collapsed; open when Failed).
5. **Footer:** Brand, "Transparent journeys. Trusted goods.", the records note.

No blockchain words outside Technical details. Not shown until the API has them (`docs/design.md` gaps): country and region, coordinates, harvest period, grade, packaging, a certificate number, a downloadable verification report, report-a-concern.

## Access screens (Figma `AuthFlow`; drafts 2a–2m)

`app/(auth)/layout.tsx`: on `lg` a navy brand panel on the left (`minmax(350px,0.8fr)`: Brand, eyebrow "Trusted trade infrastructure", "Every product has a story. Make it verifiable.", the foot line, rings in the corner) and the form on white on the right (`max-w-[400px]`, or `max-w-[540px]` for sign-up and onboarding); on phones the brand mark above the form, no panel. Each page: eyebrow (teal), `text-page` title, `text-sm` muted line, the form, the switch link centred under it. Sign in has "Keep me signed in" (`remember`).

## Offline and weak connections

Mines, farms and ports often have poor signal:
- Cache the shell and the user's current lots (service worker / PWA).
- Queue signatures and photos locally; show a pending chip "Saved on this phone · will send when online" (pending style) and an error state that says nothing was lost.
- Compress photos client-side before upload; show upload progress per file.
- Never block reading a passport on a live connection if it was loaded before.

## Notifications

Bell with red count in the top bar. Events: incoming handover, handover accepted/rejected, flag raised on your lot, certificate expiring (30 / 7 days), organisation verified. Push copy is one sentence with the quantity and the counterparty, ending with what to do.
