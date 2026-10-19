# Screens

**Visual source:** the Claude Design drafts (`ui/ChainTrack Drafts.dc.html`) for content and flows, and the Figma Make app (`ui/figma-make/`) for the look. Both are reference only. Each screen below names its draft ids; where the Figma file shows a field the API does not have, it is listed in the gap table in `docs/design.md` and left out of the screen.

| Screen | Drafts | Figma component |
| --- | --- | --- |
| 01 Public verification | 1a, 1b, 1c | `PublicVerification` |
| 02 Dashboard | 1d, 3o | `Dashboard` |
| 03 Access and onboarding | 2a–2m | `AuthFlow` |
| 04 Facility registry | 3a, 3b | `Facilities`, `RegisterFacility` |
| 05 Issue origin passport | 3c, 3d | `IssuePassport` |
| 06 Handover (and passport detail) | 1f, 3e, 3g, 3h | `Handovers`, `Passports` |
| 07 Transform / scan | 1e, 3f | `Transform`, `Scanner` |
| 08 Certificates and flags | 3i, 3j, 3k | `Certificates`, `Flags` |
| 09 Admin (organisations, templates, users) | 3l, 3m | `VerificationQueue`, `Templates`, `Users` |
| 10 Reports and audit | 3n | `Reports` |
| Profile | 2m | `Profile` |

Every screen from the design, in build order. For each: who uses it, the sections in order, the data shown, and the states to build. Sample names and numbers are illustrative; use real data. Layout rules are in `patterns.md`; component specs in `components.md`.

---

## 01 · Public verification page — `/p/[lotCode]`

No login; opened by scanning the QR label. Full order and desktop grid are in `patterns.md` → Public verification page.

**Sections**
1. **Verdict** (VerdictBadge, full): Verified / Warning / Failed, one-sentence summary, ≤ 4 reasons, meta line "Checked {date, time CAT} · {n} records · {n} organisations".
2. **Product summary** (PassportCard): product name, lot code, commodity, quantity + packaging ("960 kg · 16 bags"), harvest/crop year, current holder (name + city, country), last updated, grade. Footer: "Passport issued {date}" + verdict.
3. **Origin** card (icon `Sprout` for farms, `Pickaxe` for mines, `Factory`): facility name + operator, type ("Farm plots (2)", "Mine site"), area, country · region · district, licence number (mono) + Status ("Valid to 31 Mar 2027"), forest check (farms) or capacity used (mines). FacilityMap with the boundary.
4. **Journey** card, note "9 steps · 6 handovers": JourneyTimeline, origin → current holder. Last step state `current`.
5. **Route** card, note "Schematic, not to navigate": RouteMap + numbered stop list.
6. **Made from** card, note "Merged from 2 origin lots": GenealogyTree rooted at this lot.
7. **Certificates**: certificate of origin, export permit, lab/assay report (CertificateTiles).
8. **What we checked**: ChecksList.
9. **Technical details**: collapsed disclosure with ledger explanation and transaction rows.
10. Links: "Report a concern" · "How verification works". Footer line.

**Three outcomes to build**
- **Verified** (e.g. coffee): all reasons pass. "This coffee comes from two registered farm plots in Rwanda. Every handover since harvest was signed by both sides, and no record has changed since it was signed."
- **Warning** (e.g. cocoa with an expired lab report): reasons pass except "1 certificate expired · Lab report expired 8 Sep 2026, 13 days before import". The expired certificate tile has the amber border and a note; the last journey step shows a warning chip "Lab report expired 8 Sep"; the check "Certificates valid at each step" is Warning. Summary ends: "Ask the holder for a current report before you rely on it."
- **Failed** (e.g. tin concentrate whose handover was altered after signing): red alert bar at the very top ("**Do not accept, pay for or ship this lot.** It is under investigation by the Mines Regulator (case FLG-2026-0147, opened 14 Aug 2026)."). Reasons: "A record was altered after it was recorded" (detail "Handover of 12 Aug: signed for 2,100 kg, now shows 2,850 kg"), "Quantities do not balance" ("750 kg has no origin passport"), "1 certificate revoked", plus the passing origin check. PassportCard flagged; quantity in red with "Signed origin 2,100 kg". The altered journey step is `failed` with the signed-vs-now compare; every later step is `untrusted`. Genealogy shows a failed node "No passport · Unaccounted material · 750 kg". Export permit tile Revoked. Technical details open, showing "sealed fingerprint" vs. "record as stored now · does not match" in red. "Report a concern" link in the verdict meta.

**States:** loading (skeleton of verdict + card, "Checking {n} records…"), lot not found ("We can't find this lot code. Check the code on the label or type it in." + input), offline (show the last cached result with the time it was checked).

---

## 02 · Dashboard — `/` (role-based)

Shell with sidebar/tab bar. See `patterns.md` → Role-based dashboards for the table of tiles, actions and columns.

**Producer (e.g. cooperative manager).** Greeting + "✓ Verified · {org} · verified producer". Tiles: Passports you hold "14 · 18,640 kg cherry"; Waiting for your signature "2 · Handovers to Kitabi WS" (attention); Incoming to accept "0 · Nothing on its way"; Open flags "0 · None in 90 days". Actions: **Issue passport**, Scan QR, Hand over, Register facility. Main: "Waiting for your signature" list (rows: "Hand over to Kitabi Washing Station", lot code, "Started by Olivier K. · 2 Oct, 07:55", quantity, chip Awaiting your signature), "Passports you hold" list ("All 14" link). Side: "Capacity this season" (one compact capacity meter per facility, near-limit amber, expired licence noted), "Certificates" (licence expired row red, registration valid row).

**Exporter.** Tiles: Passports you hold "9 · 24,860 kg"; Waiting for your signature "1"; Incoming to accept "3 · 2,640 kg on the way" (attention); Open flags "1 · Raised by you · 40 kg short" (alert). Actions: **Scan QR**, Hand over, Transform, Prepare export. Main: Incoming handovers list (Awaiting your signature / Signed by sender chips), Passports you hold table (lot, product, quantity, stage, status; flagged row red). Side: Open flag card (chip Under review, FLG code, one-paragraph summary, "Open flag →"), Ready to export checklist (certificate of origin ✓, export permit ✓, "2 lots · 2,280 kg green, all checks passing").

**Regulator.** Greeting + "Mines Regulator, Rwanda · oversight of 214 organisations and 389 mine sites". Tiles: Open flags "4 · 1 high · 2 medium · 1 low" (alert); Organisations to verify "6 · Oldest waiting 3 days" (attention); Certificates expiring "11 · In the next 30 days"; Lots verified this month "1,284 · 99.2% passed all checks". Actions: **Review flags**, Scan QR, Issue certificate, Run report. Main: Flag queue (icon by type, title, org · commodity · date, FLG code, severity chip High/Medium/Low, "Assigned to you · day 49"), Organisations waiting (Documents to check chip). Side: certificate counts, verified volume by commodity (horizontal bars in `bg-primary` on `bg-surface-sunken` tracks, values right in mono).

**States:** first-run producer with no facilities (empty state pointing to Register facility), pending verification (banner "You can't issue or receive passports until verification is done."), loading skeleton tiles.

---

## 03 · Onboarding and business verification

Phone-first, no shell. Stepper: Account · Organisation · Documents · Review.

1. **Sign up** `/signup`: wordmark, "Create your organisation's account", explainer "You'll need your business registration and any operating licences. Verification takes about 2 working days." Fields: full name, Email or phone, password ("At least 12 characters"), language (English · Kinyarwanda · Français), consent checkbox. Action "Create account"; "Already have an account? Sign in".
2. **Organisation details**: legal name ("Exactly as on your registration certificate"), country + registration number (mono) side by side, **type** as choice cards: Producer (farms, cooperatives, mines) · Processor (washing stations, smelters, factories) · Logistics (transporters, warehouses, forwarders) · Exporter or importer · Certifier or regulator (needs accreditation). Your role there (select).
3. **Documents**: "Upload your documents · Clear photos or PDFs. We check them against the national registry." Required: certificate of registration, operating licence (file rows with progress). Optional: tax ID (drop zone). Note: "Facility licences are added later, when you register each farm plot or site." Action "Submit for verification".
4. **Pending verification**: round pending mark (`Clock`), "Pending verification", "We're checking {org}. Most organisations are verified within 2 working days.", chip "Submitted {date, time}". Vertical steps: Account created ✓ · Documents received ✓ · Registry check (now: "Matching RCA/0417/2014 with the Rwanda Cooperative Agency") · Approval ("You'll get an SMS and an email"). "While you wait": invite your team, draft your facilities, print a test label. Banner: "You can't issue or receive passports until verification is done."
   Other outcomes: **Changes requested** (warning notice listing what to fix, re-upload) and **Rejected** (danger notice with the reason, contact support).

**Admin review queue** `/admin/organisations` (desktop): page title "Verification queue", sub "6 organisations · oldest waiting 3 days · target under 2 working days". Tabs: Organisations (6) · Certificate issuers (2) · Approved · Rejected. Left: queue rows (type icon, name, "Producer · Rwanda · 2 Oct", age right, problem chips like "1 document unreadable"). Right: selected org: name, "Submitted by … · {date}", chip Pending verification, KeyValue (legal name, country, registration no., type and role, contact), document previews (click to zoom), **Automatic checks** rows (registry match, name match, licence readable and in date, not a duplicate, sanctions/adverse media screen), each with a Status. Decision note textarea. Actions: Reject (danger-outline) · Ask for changes · **Approve organisation**.
**Certificate issuers tab**: accredit an already-verified org: certificate types they may issue (checkboxes, e.g. Assay report ✓, Lab report: radioactivity ✓, Certificate of origin, Export permit disabled "Only national regulators can issue this"), commodities (tags), their accreditation document (e.g. ISO/IEC 17025), accredited from/until ("Ends with their ISO certificate"), signing keys ("2 lab officers registered"), info notice on what accreditation shows publicly and that suspension doesn't void past certificates. Actions: Decline · **Accredit issuer**.

---

## 04 · Facility registry (producers) — `/facilities`

**List + map** (desktop: `grid-cols-[1fr_1.6fr]`; phone: Map | List segmented toggle). Header "Facilities · 3 registered · {org}", action **Register facility**, type filter. Rows: icon, "Plot 14 · Kigoma Hill", "Farm plot · 1,200 kg left", licence number, Status (Active ✓ / 94% of capacity ⚠ / Licence expired ✕). Map shows all boundaries; expired ones dashed red; legend "Active · Licence expired". Selected facility card: type, area, licence + validity, declared capacity "4,000 kg cherry per season", forest check; **CapacityMeter** for the current period ("You have 1,200 kg left this season. Season ends 31 Jan 2027."); actions Edit boundary · **Issue passport here**.

**Register facility** `/facilities/new`.
- Map tools (segmented): **Drop pin** | **Draw boundary**. Hint: "Tap the map to add corners. Drag a corner to move it. Tap the first corner to close the shape." Live readout (mono): corners, area in ha, centre coordinates. Undo · **Use GPS walk** (record the boundary by walking it). On phones the tools sit above the map, readout below.
- Automatic checks under the map (success notice): inside the country, no overlap with another registered facility, no forest loss since 31 Dec 2020 (farms). Overlap or forest loss → warning/danger notice and block submit.
- Details: facility name, **type** choice cards (Farm plot, Mine site, Factory or station, Warehouse), licence number (mono), licence expiry (date), **declared capacity** (kg) + **per period** (season / month / year), licence document upload. Note: "Capacity is checked every time a passport is issued from this facility. A regulator must approve changes after the first passport."
- Phone stepper: Location · Details · Licence · Capacity. Actions: Save draft · **Submit for approval**.

---

## 05 · Issue origin passport — `/issue` (focused flow)

Stepper: Facility · Quantity · Evidence · Sign.
1. **Where was it produced?** Product template select ("Arabica coffee · cherry (Rwanda) v3", help "Sets the unit, stages and evidence required.") + facility choice cards with what's left ("2.2 ha · 1,200 kg left this season"); expired-licence facilities disabled with the reason.
2. **How much?** Quantity (mono, kg addon), help "{product} · {facility}". **Live CapacityMeter** with the entered amount as a lighter segment: "You have 1,200 kg left this season. After this passport: 300 kg." Under it the template's **required evidence** checklist (Harvest date ✓ · Two field photos with location ✓ "inside plot boundary" · Picker collection sheet ✓ · Inspector countersignature "Requested when you sign"). Action "Review and sign".
   **Blocked state:** quantity above what's left → see `patterns.md` → Blocking checks (field error "More than the 1,200 kg left for this facility this season.", stepper step red, meter over, blocking notice, Continue disabled).
3. **Check and sign**: summary (product, template, facility, quantity, evidence "3 of 3 attached", countersign by), PIN panel, "Sign and issue passport".
4. **Passport created**: verified mark, "Passport issued", "900 kg from Plot 14. Waiting for the inspector's countersignature.", chip Awaiting inspector countersignature, **printable QR label**, note "Print one label per bag or crate. Labels print at 50 × 30 mm on standard thermal printers." Actions: Print labels · **Hand over**.

---

## 06 · Dual-signature handover — `/handover` (sender) and `/handovers/[id]` (receiver)

Full behaviour in `patterns.md` → Dual-signature handover. Screens:

**Sender (stepper Lot · Details · Receiver · Sign):** 1 Scan QR ("Scan the QR label on the bags") → 2 Is this the right lot? (PassportCard: quantity, packed as, current holder "(you)", footer last signed + Verified; success notice "You hold this lot and nothing is pending on it. It can be handed over.") → 3 What's leaving? (All of it | Part of it, quantity, location select with GPS coords, truck seal number, photos ≥ 2: Load, Seal, Label) → 4 Who is receiving it? (search, choice cards with Verified status; unverified disabled; "Can't find them? Invite an organisation") → 5 Check and sign ("Sign and send") → 6 Sent for signature (pending mark `Send`, "1,200 kg is on its way to {receiver}.", chips Signed by sender + Awaiting receiver signature, next-steps list, info notice "If they reject the handover, you'll be notified with their evidence and the lot comes back to you.").

**Receiver:** 1 Notification + Handovers list (tabs Incoming (1) · Outgoing · Done) → 2 Scan on arrival (found sheet: "Matches the handover from Kitabi") → 3 **Inspect on arrival** (chips Signed by sender + Awaiting your signature; Weight received; comparison panel; inspection checklist; your photos; buttons Reject / raise flag · **Accept and sign**) → 4 Handover complete (double-check mark, "You now hold 1,200 kg of red bourbon parchment.", Signed by both parties, steps, View passport · Next step).
**Problem path:** 3b weight short / seal mismatch (comparison red "40 kg short (3.3%). Above the 0.5% tolerance.", seal check unticked with "Seal number on truck: KTB-SEAL-40977") → 5 Reject sheet → 6 Flag raised (danger mark, "Handover rejected", "Flag FLG-2026-0163 is open. The lot stays with Kitabi Washing Station until it's resolved.", chip Rejected by receiver, steps: Sender and regulator notified ✓ · Regulator reviews the evidence · Cleared or confirmed as fraud).

---

## 07 · Transform / split / merge — `/transform`

Processors. Page title = the operation in plain words ("Wash and dry cherry"), sub "Template: {template} · allowed yield up to 21%". See `patterns.md` → Transform.
- **Live**: inputs selected (2 of 3, 6,000 kg), outputs Parchment A 1,080 kg + Parchment B 120 kg, meter within ("Within limit · 60 kg below the allowed maximum (yield 20.0%)"), formula line. "Sign and create 2 passports".
- **Exceeded**: outputs 1,250 + 130 = 1,380 kg vs. max 1,260 → over state everywhere, button disabled.
- **Created**: compact verified verdict "2 new passports created", new PassportCards (quantity, holder, "Linked to parents" status), GenealogyTree. Print labels · **Hand over**.
- Split = one input, many outputs (sum ≤ input × tolerance). Merge = many inputs, one output (same product and stage only; template enforces it).

---

## 08 · Certificates and flags (certifiers, regulators)

**Issue certificate** `/certificates/new` (desktop two columns). Form: Attach to (Passport (lot) | Facility), passport picker (lot code + product + quantity, help "Holder: … · verdict: Verified"), certificate type, certificate number (generated, editable), valid from / until (help "6 months · the maximum for this type"), signed document upload, notes. Side: **live preview** (CertificateTile exactly as it appears publicly) and "Checks before issuing" (passport verdict Verified ✓, origin licence valid ✓, your accreditation covers this type ✓). Actions: Cancel · **Sign and issue**.

**Certificates list + revoke** `/certificates`: table "Issued by you" (number, type, attached to, validity, Status: Valid / Expiring in 6 days / Expired, row menu). Revoke dialog: "Revoke certificate {number}?", reason select, explanation textarea ("shown to the holder and on the public page"), "Passports that will show 'Certificate revoked'" list (this lot + every lot made from it), warning "Revoking can't be undone. You can issue a new certificate once the flag is resolved.", Cancel · **Revoke certificate** (danger).

**Flags and investigations** `/flags` (desktop `grid-cols-[380px_1fr]`): title "Flags and investigations · 4 open · 1 high". Queue rows: type icon tinted by severity, title (Record altered after signing / Mass balance breach blocked / Handover rejected / Capacity exceeded (blocked)), FLG code, org · date, severity chip. Detail:
- Header: severity chip + FLG code, title, "Raised automatically 14 Aug 2026, 06:00 · assigned to {name} · day 49", Reassign.
- KeyValue: lot, organisation, commodity, **What happened** (one plain paragraph).
- **Evidence**: field-level diff of signed vs. stored record (mono rows; removed values `bg-danger-soft text-danger-ink` with "−", added values `bg-warning-soft text-warning-ink` with "+"), who changed it and how, photos.
- **Affected downstream lots** table: lot, how it was made, quantity, holder, status chip (Altered / On hold), note "3 lots · 5,660 kg on hold".
- Actions: Request information · Clear flag · **Confirm fraud** (danger). Confirm dialog per `patterns.md` → Destructive actions (revokes the lot and everything made from it; optional revoke export permit and suspend org 90 days; finding shown on public pages).

---

## 09 · Admin

**Product template builder** `/admin/templates/[id]` (desktop, `grid-cols-[1.6fr_1fr]`). Header "Product template · {name}", Preview as producer · **Publish v4**.
- General: template name, commodity, base unit (kg), **capacity rule** ("Per facility, per season (Feb–Jun, Sep–Jan)"), version ("v4 draft" + chip "v3 live · 1,904 passports").
- Stage editor (selected stage): stage name, who may act (role), **allowed events** as toggle tags (Receive, Transform, Split, Merge, Hand over, Export, Import, Deliver, Issue), input form → output form, **maximum yield** (%, help "Output ≤ input × 21%"), handover tolerance (%), unit, **required evidence** checklist with type hints ("Moisture reading · Number, % · must be 10–12.5%"), Add evidence, Remove stage.
- **Live preview of the flow**: vertical stage cards connected by arrows; each shows number, icon, name and three tag kinds: events (`bg-primary-soft text-primary`), rules (`bg-verified-soft text-verified-ink`, `Scale` icon), evidence (`bg-warning-soft text-warning-ink`, `FileText` icon). Selected stage outlined `border-2 border-primary`.
- Publishing makes a new version; existing passports keep the version they were issued under.

**Users and roles** `/admin/users`: tabs People · Roles · Sign-in policy. People table: person (avatar initials, name, email/phone), organisation, role, sign-in method ("2FA · PIN", "hardware key"), Status (Active / Invited / Suspended); search, org and role filters, **Invite person**. "What each role can do" permission matrix (permissions as rows: view passports, issue passports, sign handovers, transform lots, issue certificates, resolve flags, manage users; roles as columns; ✓ in `text-verified` or "—"; screen-reader text Yes/No). Note: "Signing always needs the person's own PIN or key. Org admins can't sign on behalf of someone else."

**Reports and audit** `/admin/reports`: saved reports as tiles (Volume by commodity · Monthly; Verification outcomes · Weekly; Flags and resolutions · On demand). Filters row: date range, organisation, event type, commodity, Clear. Audit table: When (CAT, mono), Event (Status icon + text: verified/warn/failed tone), Organisation, By, Reference (mono). Failed rows flagged. Footer "7 of 18,406 events · Every row is sealed; this log can't be edited". Export CSV · Export PDF.

---

## 10 · Prototype journeys (end-to-end flows to support)

These sequences must work end to end; use them as acceptance tests.

**a. Coffee, farm to cup:** Producer chooses facility → enters quantity (capacity OK) → signs → passport + labels → hands over to washing station → washing station inspects and accepts → transforms cherry to parchment (mass balance OK) → child passports created → exporter notified → exporter inspects and accepts → (hulls to green, transform) → hands over for export → importer accepts on arrival → roaster accepts → consumer scans the bag → public page **Verified**.

**b. Mineral, mine to refiner:** Mine enters quantity (capacity OK) → passport + bag tags → hands over to trader → trader weighs and accepts → hands over to processor → processor concentrates two ore lots into one (mass balance within 92%) → concentrate passport linked to both ore lots and back to two mine sites → hands over to exporter → exporter accepts and exports → refiner inspects and accepts → refiner sees passport **Verified** with genealogy to both mine sites.

**c. Fraud caught:** capacity exceeded at issuing (blocked, logged) → mass-balance breach at processing (blocked) → receiver weighs a short delivery with a mismatched seal → rejects and raises a flag → flag raised, lot frozen with the sender → a lot whose signed record was altered shows **Failed** on its public page.
