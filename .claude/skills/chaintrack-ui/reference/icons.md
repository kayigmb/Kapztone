# Icons

Use `lucide-react` (names checked against v1.x). One stroke style everywhere: default `strokeWidth={2}`, colour from `currentColor`.
Sizes: `size-4` in chips, tables and inline status · `size-5` beside text and in buttons · `size-7`/`size-8` in markers and verdict marks.
Add `aria-hidden` when the icon sits next to a word (it always should for statuses). Icon-only buttons need `aria-label`.

## Status (never use these for anything else)

| Meaning | Icon |
| --- | --- |
| Pass / valid / verified (inline) | `CircleCheck` |
| Warning / expiring / near limit | `TriangleAlert` |
| Fail / revoked / over limit / blocking error | `OctagonX` |
| Awaiting / pending | `Clock` |
| Signed by both parties | `CheckCheck` |
| Signature needed / signed by one side | `PenLine` |
| Rejected | `X` |
| Verdict marks | `ShieldCheck` (Verified) · `ShieldAlert` (Warning) · `ShieldX` (Failed) |
| Info notice | `Info` |
| Flag / investigation | `Flag` |

## Facilities, events and commodities

| Thing | Icon |
| --- | --- |
| Farm plot | `Sprout` |
| Mine site | `Pickaxe` (`Mountain` for a concession area) |
| Factory, washing station, smelter | `Factory` |
| Warehouse, depot, import | `Warehouse` |
| Transport, handover for transport | `Truck` |
| Export, sea leg, export permit | `Ship` |
| Handover (generic) | `ArrowLeftRight` |
| Transform | `RefreshCcw` |
| Split | `Split` |
| Merge | `Merge` |
| Genealogy, "Made from", templates | `Layers` |
| Generic lot, delivered | `Package` |
| Coffee | `Coffee` (or `Bean`) |
| Cocoa, other crops | `Leaf` |
| Minerals, ore | `Gem` |
| Medicines | `Pill` |
| Certificate | `FileText` |
| Certifier / issue certificate | `Stamp` |
| Lab or assay report, mass balance | `Scale` |
| Capacity | `Gauge` |
| Route | `Route` |
| Map location | `MapPin` |
| Boundary drawing | `Pentagon` |
| Organisation | `Building2` |

## App and actions

| Action | Icon |
| --- | --- |
| Scan QR, QR label | `QrCode` |
| Camera / add photo | `Camera` · `Image` |
| Torch (scanner) | `Zap` |
| Type code manually | `Keyboard` |
| Home | `House` |
| Notifications | `Bell` |
| Users | `Users` · invite `UserPlus` |
| Settings | `Settings` |
| Reports | `ChartColumn` · audit `Activity` |
| Search · Filter | `Search` · `Funnel` |
| More | `Ellipsis` |
| Back · Forward | `ArrowLeft` · `ArrowRight` |
| Disclosure | `ChevronDown` (rotates 180° open) |
| External link | `ExternalLink` |
| Technical details | `Lock` |
| Upload · Download/export | `Upload` · `Download` |
| Print | `Printer` |
| Send (handover sent) | `Send` |
| Add | `Plus` |
| Edit · Delete | `Pencil` · `Trash2` |
| Retry | `RefreshCw` |
| Offline | `WifiOff` |
| Empty inbox | `Inbox` |
| Calendar field | `Calendar` |
| Inspect / preview | `Eye` |

No emoji. No coin, wallet, chain-link or cube imagery.
