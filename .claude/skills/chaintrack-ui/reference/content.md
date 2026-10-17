# Content and copy

All copy lives in `frontend/lib/consts.ts` (`copy`). Write it there, following this file.

## Voice

Calm, precise, plain. Like a good customs officer or bank: says exactly what happened and what to do next, without drama.

- **Say what was checked, not how.** "Verified", "Signed by both parties", "Quantities balance at every step", "Records unchanged since signing". Not "hash validated on-chain".
- **Forbidden outside Technical details:** blockchain, crypto, Web3, wallet, token, gas, smart contract, on-chain, hash, block, node, mint.
- **Real numbers, with units.** "6,000 kg issued against 7,400 kg declared for Season A 2026." Not "Capacity OK".
- **Problems end with the next step.** "Ask the holder for a current report before you rely on it." / "Lower the quantity to 1,200 kg or less." / "Do not accept, pay for or ship this lot."
- Address the user as **you**; on the public page ChainTrack speaks as **we** only in "What we checked".
- No exclamation marks, no emoji, no jokes, no "Oops". Short sentences.
- Sentence case everywhere (titles, buttons, labels), except `label` keys, which are uppercased by CSS (write them in sentence case in code).

## Words we use

| Use | Not |
| --- | --- |
| Passport (digital product passport) | NFT, token, asset |
| Lot, lot code | Batch ID, token ID |
| Hand over / handover | Transfer, send token |
| Signed by both parties | Confirmed on-chain, 2/2 signatures |
| Awaiting receiver signature | Pending |
| Records unchanged since signing | Hash matches |
| Record altered after signing | Tampered, hacked |
| Transform / split / merge | Mint, burn |
| Current holder | Owner, wallet |
| Organisation | Account, entity |
| Flag | Ticket, incident |
| Verified / Warning / Failed | Valid, OK, Error, Invalid |

## Formatting

- **Dates:** `3 Jun 2026`. With time: `2 Oct 2026, 09:14 CAT` (always the zone). Relative only for recent items: "today", "2 days", "day 49" (investigation age).
- **Quantities:** thousands separator, space, unit: `1,200 kg`, `12,500 kg · 200 bags`, `62.4% Sn`. Transformations with an arrow: `6,000 kg cherry → 1,200 kg parchment`. Use `Intl.NumberFormat('en', { maximumFractionDigits: 1 })`.
- **Percentages:** one decimal for yields and tolerances: `Yield 20.0% (max 21%)`, `40 kg short (3.3%)`.
- **Lot codes:** always in full, mono, never truncated: `CT-RW-COF-26-004817` (CT-country-commodity-year-sequence). Flags `FLG-2026-0147`. Certificates `COO-RW-2026-18822`.
- **Organisations:** full legal name the first time ("Kitabi Washing Station", "Akagera Coffee Exports Ltd"), shorter after.
- **Places:** "Port of Mombasa, Kenya", "Nyamagabe, Rwanda". Coordinates in mono: `−2.4521, 29.5486`.
- **Handovers:** "Sender → **Receiver**", receiver bold.

## Buttons

Verbs naming the result. One primary per screen.
Good: Issue passport · Scan QR · Hand over · Accept and sign · Reject / raise flag · Sign and send · Sign and issue passport · Sign and create 2 passports · Print labels · Register facility · Submit for verification · Approve organisation · Accredit issuer · Revoke certificate · Confirm fraud and revoke 3 lots · Try again.
Never: Submit · OK · Yes · Confirm (alone) · Click here.

## Patterns of copy

- **Verdict summary (one sentence):** where it's from, custody, integrity. Warning/Failed add what to do.
- **Reason lines:** short facts. "Origin facility verified" · "All 6 handovers signed by both parties" · "Quantities balance at every step" · "1 certificate expired".
- **Field errors:** what's wrong + what to enter. "More than the 960 kg this passport holds. Enter 960 kg or less."
- **Blocking notices:** bold first sentence stating the block, then the fix. "**This passport can't be issued.** Lower the quantity to 1,200 kg or less. If the harvest really is larger, ask the regulator to review the declared capacity."
- **Disabled reasons:** next to the disabled control. "Continue is blocked until the quantity fits." · "Licence expired 30 Jun 2026. Renew to issue." · "Not verified yet: can't receive".
- **Empty states:** what would be here + how it fills. "No handovers waiting for you. When someone hands a lot to your organisation, it appears here for you to inspect and sign."
- **Errors:** what failed, whether anything was lost, what happens next. "Can't reach the ledger. Your signature is saved on this phone and will be sent when you're back online. Nothing has been lost."
- **What a signature means:** one sentence above the PIN. "You confirm these goods left your custody. {Receiver} must inspect and sign before the handover is complete."

## Languages

Plan for English, French and Kinyarwanda from the start (`next-intl` or similar). Keep strings out of components, allow ~35% longer text in buttons and chips, and never put text in images.
