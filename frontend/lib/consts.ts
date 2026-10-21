// Every value and every piece of UI copy the frontend uses, declared once.
// Statuses, kinds and paths mirror the API (backend/internal/*/values.go);
// copy lives in `lib/copy/*.json` (a second language is a translation of
// those files, not a rewrite of the pages); only plurals, conditions and
// locale formatting stay here as functions.
//
// Pure besides JSON: format.ts has no imports, so `node --test` can load it.
import { fill } from "./format.ts";
import copyJson from "./copy/copy.json" with { type: "json" };
import marketingJson from "./copy/marketing.json" with { type: "json" };
import platformCopyJson from "./copy/platformCopy.json" with { type: "json" };
import reportsCopyJson from "./copy/reportsCopy.json" with { type: "json" };
import screenCopyJson from "./copy/screenCopy.json" with { type: "json" };
import confirmCopyJson from "./copy/confirmCopy.json" with { type: "json" };

// Copy lives in JSON (lib/copy/*.json) so translators never touch code.
// Only functions stay here: plurals, conditions and locale formatting.
// JSON leaves are untyped maps; the two aliases below are the one place that
// says so — never write `Record<string, string>` at a use site again.
/** A flat copy table from JSON: every leaf is a finished string. */
export type StrMap = Record<string, string>;
/** A copy table whose leaves are `{ label, pass, reason }` check rows. */
export type CheckMap = Record<
  string,
  { label: string; pass: string; reason: string }
>;

// ---- Values shared with the API ----

export const orgStatus = {
  pending: "pending",
  verified: "verified",
  rejected: "rejected",
  changesRequested: "changes_requested",
  suspended: "suspended",
} as const;
export type OrgStatus = (typeof orgStatus)[keyof typeof orgStatus];

/** A reviewer's decisions on an organisation (POST /orgs/{id}/verify). */
export const reviewDecisions = [
  orgStatus.rejected,
  orgStatus.changesRequested,
  orgStatus.verified,
] as const;
export type ReviewDecision = (typeof reviewDecisions)[number];
/** Decisions that send the organisation back to upload again; each needs a reason. */
export const reopeningDecisions: readonly ReviewDecision[] = [
  orgStatus.rejected,
  orgStatus.changesRequested,
];

export const orgKinds = [
  { value: "producer", label: "Producer", sub: "Farms, cooperatives, mines" },
  {
    value: "processor",
    label: "Processor",
    sub: "Washing stations, smelters, factories",
  },
  {
    value: "transporter",
    label: "Logistics",
    sub: "Transporters, warehouses, forwarders",
  },
  {
    value: "exporter",
    label: "Exporter",
    sub: "Ships goods out of the country",
  },
  { value: "importer", label: "Importer", sub: "Receives goods from abroad" },
  {
    value: "certifier",
    label: "Certifier",
    sub: "Issues certificates; needs accreditation",
  },
  {
    value: "regulator",
    label: "Regulator",
    sub: "Oversees a sector; needs accreditation",
  },
] as const;
/** Kinds of organisation that can be accredited to issue certificates. */
export const issuerKinds: readonly string[] = ["certifier", "regulator"];

export const documentTypes = [
  {
    value: "registration",
    label: "Certificate of registration",
    required: true,
  },
  { value: "licence", label: "Operating licence", required: true },
  { value: "other", label: "Other supporting document", required: false },
] as const;

/** The admin verification queue's tabs: which organisations each one lists. */
export const adminTabs = [
  { value: "organisations", statuses: [orgStatus.pending] },
  { value: "issuers", statuses: [orgStatus.verified], kinds: issuerKinds },
  { value: "approved", statuses: [orgStatus.verified, orgStatus.suspended] },
  {
    value: "rejected",
    statuses: [orgStatus.rejected, orgStatus.changesRequested],
  },
] as const satisfies readonly {
  value: string;
  statuses: readonly OrgStatus[];
  kinds?: readonly string[];
}[];
export type AdminTab = (typeof adminTabs)[number]["value"];

export const permissions = {
  orgVerify: "org:verify",
  userManage: "user:manage",
  passportIssue: "passport:issue",
  passportRead: "passport:read",
  passportRetire: "passport:retire",
  handoverCreate: "handover:create",
  handoverAccept: "handover:accept",
  transformCreate: "transform:create",
  certificateIssue: "certificate:issue",
  certificateRevoke: "certificate:revoke",
  flagCreate: "flag:create",
  flagResolve: "flag:resolve",
  templateRead: "template:read",
  templateManage: "template:manage",
  facilityRead: "facility:read",
  facilityManage: "facility:manage",
  facilitySuspend: "facility:suspend",
  auditRead: "audit:read",
  reportRead: "report:read",
} as const;

/** Roles an organisation admin can give a member (POST /invites, PUT /users/{id}/role). */
export const memberRoles = [
  { value: "org_admin", label: "Organisation admin" },
  { value: "member", label: "Member" },
] as const;
export type MemberRole = (typeof memberRoles)[number]["value"];
export const memberStatus = {
  active: "active",
  suspended: "suspended",
} as const;

export const passportStatus = {
  active: "active",
  consumed: "consumed",
  retired: "retired",
  revoked: "revoked",
} as const;
export type PassportStatus =
  (typeof passportStatus)[keyof typeof passportStatus];
/** Multipart field holding handover photos (backend passport.FormPhotos), and how many fit. */
export const handoverPhotoField = "photos";
export const maxHandoverPhotos = 6;

/** Why a receiver rejects (backend passport.RejectReasons). */
export const rejectReasons = [
  "weight_short",
  "damaged",
  "wrong_product",
  "seal_broken",
  "not_expected",
] as const;
export type RejectReason = (typeof rejectReasons)[number];

/** Every handover status; `expired` is never stored, an awaiting handover past its deadline reads as it. */
export const handoverStatus = {
  awaiting: "awaiting_receiver",
  completed: "completed",
  rejected: "rejected",
  cancelled: "cancelled",
  expired: "expired",
} as const;
export type HandoverStatus =
  (typeof handoverStatus)[keyof typeof handoverStatus];
export const flagStatus = {
  open: "open",
  cleared: "cleared",
  confirmed: "confirmed",
} as const;
/** What resolving a flag does (POST /flags/{id}/resolve). */
export const flagOutcomes = [flagStatus.cleared, flagStatus.confirmed] as const;
export type FlagOutcome = (typeof flagOutcomes)[number];
export const certificateStatus = {
  valid: "valid",
  expired: "expired",
  revoked: "revoked",
} as const;

// Facilities (backend/internal/facility/values.go).
export const facilityStatus = {
  pending: "pending_approval",
  active: "active",
  suspended: "suspended",
} as const;
export type FacilityKind = string;
/** The period a facility's declared capacity applies to. */
export const capacityPeriods = [
  { value: "season", label: "Season" },
  { value: "year", label: "Year" },
  { value: "month", label: "Month" },
] as const;
export type CapacityPeriod = (typeof capacityPeriods)[number]["value"];

// What a certificate is attached to (POST /attestations form fields).
export const subjectTypes = {
  passport: "passport",
  facility: "facility",
} as const;
export type SubjectType = (typeof subjectTypes)[keyof typeof subjectTypes];

// Commodity templates (backend/internal/template/values.go).
export const templateEvents = [
  { value: "ISSUE", label: "Issue" },
  { value: "HANDOVER", label: "Hand over" },
  { value: "TRANSFORM", label: "Transform" },
  { value: "EXPORT", label: "Export" },
  { value: "IMPORT", label: "Import" },
] as const;
export type TemplateEvent = (typeof templateEvents)[number]["value"];
/** The events the screens branch on. */
export const templateEvent = {
  issue: "ISSUE",
  handover: "HANDOVER",
  transform: "TRANSFORM",
} as const;
/** How a template's per-facility capacity is worked out. */
export const capacityRuleTypes = {
  perArea: "yield_per_area",
  facility: "facility_capacity",
} as const;

// ---- Shell and dashboard (patterns.md → App shell, Role-based dashboards) ----

/** Sidebar / More items. Only screens that exist are listed; a new screen adds a line. */
/** Sidebar groups (Figma BusinessApp), in order. */
export const navGroups = ["workspace", "trust", "manage"] as const;
export type NavGroup = (typeof navGroups)[number];

export const navItems = [
  { key: "home", href: "/dashboard", icon: "home", group: "workspace" },
  // Instead of every item above and below, for a member on the bare address (lib/dashboard.ts → hasWorkspace).
  {
    key: "workspaces",
    href: "/dashboard",
    icon: "workspaces",
    group: "workspace",
  },
  {
    key: "facilities",
    href: "/facilities",
    icon: "facilities",
    group: "workspace",
    permission: permissions.facilityRead,
  },
  {
    key: "issue",
    href: "/issue",
    icon: "issue",
    group: "workspace",
    permission: permissions.passportIssue,
  },
  {
    key: "handovers",
    href: "/handovers",
    icon: "handovers",
    group: "workspace",
    permission: permissions.passportRead,
    count: "incoming",
  },
  {
    key: "transform",
    href: "/transform",
    icon: "transform",
    group: "workspace",
    permission: permissions.transformCreate,
  },
  {
    key: "certificates",
    href: "/certificates",
    icon: "certificates",
    group: "trust",
    permission: permissions.passportRead,
  },
  {
    key: "flags",
    href: "/flags",
    icon: "flags",
    group: "trust",
    permission: permissions.flagCreate,
    count: "openFlags",
  },
  {
    key: "adminOrgs",
    href: "/admin/organisations",
    icon: "orgs",
    group: "trust",
    permission: permissions.orgVerify,
    count: "orgsPending",
  },
  {
    key: "templates",
    href: "/admin/templates",
    icon: "templates",
    group: "manage",
    permission: permissions.templateRead,
  },
  {
    key: "reports",
    href: "/admin/reports",
    icon: "reports",
    group: "manage",
    permission: permissions.auditRead,
  },
  {
    key: "users",
    href: "/users",
    icon: "users",
    group: "manage",
    permission: permissions.userManage,
  },
  { key: "profile", href: "/profile", icon: "profile", group: "manage" },
] as const satisfies readonly {
  key: string;
  href: string;
  icon: string;
  group?: NavGroup;
  permission?: string;
  count?: TileKey;
}[];

/** Where the sidebar remembers being folded (this browser only). */
export const sidebarStorageKey = "ct:sidebar-collapsed";
/** How often the notification bell asks for the unread count. */
export const notificationRefetchMs = 60_000;
/** Theme choices, in the top bar's menu. next-themes writes `data-theme` on <html>. */
export const themes = ["light", "dark", "system"] as const;
export type ThemeChoice = (typeof themes)[number];
export type NavKey = (typeof navItems)[number]["key"];

/** Counts a dashboard tile can show; each is one list call read for its `total`. */
export type TileKey =
  | "passportsHeld"
  | "incoming"
  | "outgoing"
  | "openFlags"
  | "certValid"
  | "certExpired"
  | "certRevoked"
  | "orgsPending"
  | "orgsChanges"
  | "orgsRejected";
/** Work lists a dashboard can show. */
export type PanelKey =
  | "incoming"
  | "passports"
  | "capacity"
  | "certificates"
  | "flags"
  | "orgsPending";
export type ActionKey =
  | "scan"
  | "issue"
  | "registerFacility"
  | "handover"
  | "transform"
  | "reviewFlags"
  | "issueCertificate"
  | "reviewOrgs"
  | "templates"
  | "users"
  | "reports";

/** Which dashboard each kind of organisation gets. "platform" is the platform admin (no organisation). */
export const dashboards = {
  trader: {
    tiles: ["passportsHeld", "incoming", "outgoing", "openFlags"],
    actions: ["scan", "handover", "transform"],
    main: ["incoming", "passports"],
    side: [],
  },
  producer: {
    tiles: ["passportsHeld", "incoming", "outgoing", "openFlags"],
    actions: ["issue", "scan", "handover", "registerFacility"],
    main: ["incoming", "passports"],
    side: ["capacity"],
  },
  certifier: {
    tiles: ["certValid", "certExpired", "certRevoked", "openFlags"],
    actions: ["issueCertificate", "scan"],
    main: ["certificates"],
    side: [],
  },
  regulator: {
    tiles: ["openFlags", "certValid", "passportsHeld", "incoming"],
    actions: ["scan", "issueCertificate", "reports"],
    main: ["flags"],
    side: [],
  },
  platform: {
    tiles: ["orgsPending", "orgsChanges", "orgsRejected", "openFlags"],
    actions: ["reviewOrgs", "templates", "users", "reports"],
    main: ["orgsPending"],
    side: [],
  },
} as const satisfies Record<
  string,
  {
    tiles: readonly TileKey[];
    actions: readonly ActionKey[];
    main: readonly PanelKey[];
    side: readonly PanelKey[];
  }
>;
export type DashboardKey = keyof typeof dashboards;
export const dashboardOfKind: Record<string, DashboardKey> = {
  producer: "producer",
  processor: "trader",
  transporter: "trader",
  exporter: "trader",
  importer: "trader",
  certifier: "certifier",
  regulator: "regulator",
};

/** Quick actions a dashboard shows; each is a screen that exists. `actionsFor` picks the ones the person may use. */
export const quickActions = [
  { key: "scan" },
  { key: "issue", href: "/issue", permission: permissions.passportIssue },
  {
    key: "handover",
    href: "/handover",
    permission: permissions.handoverCreate,
  },
  {
    key: "issueCertificate",
    href: "/certificates/new",
    permission: permissions.certificateIssue,
  },
  { key: "reviewFlags", href: "/flags", permission: permissions.flagCreate },
  {
    key: "registerFacility",
    href: "/facilities/new",
    permission: permissions.facilityManage,
  },
  {
    key: "transform",
    href: "/transform",
    permission: permissions.transformCreate,
  },
  {
    key: "reviewOrgs",
    href: "/admin/organisations",
    permission: permissions.orgVerify,
  },
  {
    key: "templates",
    href: "/admin/templates",
    permission: permissions.templateRead,
  },
  { key: "users", href: "/users", permission: permissions.userManage },
  { key: "reports", href: "/admin/reports", permission: permissions.auditRead },
] as const satisfies readonly {
  key: ActionKey;
  href?: string;
  permission?: string;
}[];

/** Tiles that ask for attention (something for you to do) or raise an alert, when above zero. */
export const tileTones: Partial<Record<TileKey, "attention" | "alert">> = {
  incoming: "attention",
  orgsPending: "attention",
  orgsChanges: "attention",
  openFlags: "alert",
};
/** A facility's capacity meter turns amber at this share of its declared capacity. */
export const capacityNearShare = 0.8;
/** How many rows a dashboard list shows before its "All" link. */
export const panelRows = 5;

// Documents (backend/internal/upload): PDF, PNG or JPEG, 10 MB at most.
export const upload = {
  maxBytes: 10 * 1024 * 1024,
  types: ["application/pdf", "image/png", "image/jpeg"],
  imageTypes: ["image/png", "image/jpeg"],
};

export const limits = {
  name: 200,
  email: 254,
  password: { min: 8, max: 72 },
  // An organisation's address is a hostname label; these names are taken by the system.
  slugPattern: /^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$/,
  reservedSlugs: ["www", "api", "admin", "app"],
  codePattern: /^(\d{6})?$/,
  // A note on a passport, handover or flag (backend MaxNoteLen).
  note: 1000,
  // Inputs and outputs of one transform (backend MaxTransformItems).
  maxItems: 50,
  // ponytail: lists ask for one full page (the API's maximum); paging UI when a queue outgrows it.
  pageSize: 100,
};

// ---- Where things live ----

export const routes = {
  home: "/dashboard",
  landing: "/",
  signin: "/signin",
  signup: "/signup",
  forgot: "/forgot",
  reset: (token: string) => `/reset/${token}`,
  lock: (token: string) => `/lock/${token}`,
  verifyEmail: (token: string) => `/verify-email/${token}`,
  handoff: (code: string) => `/handoff?code=${encodeURIComponent(code)}`,
  onboarding: "/onboarding",
  adminOrgs: "/admin/organisations",
  adminOrg: (tab: string, id?: string) =>
    `/admin/organisations?tab=${tab}${id ? `&id=${id}` : ""}`,
  adminOrgDetail: (id: string) => `/admin/organisations/${id}`,
  adminTemplates: "/admin/templates",
  adminTemplate: (id: string) => `/admin/templates/${id}`,
  adminReports: "/admin/reports",
  passport: (id: string) => `/p/${id}`,
  passportDetail: (id: string) => `/passports/${id}`,
  facilities: "/facilities",
  facilityNew: "/facilities/new",
  facility: (id: string) => `/facilities?facility=${id}`,
  issue: "/issue",
  handover: (passportId?: string) =>
    `/handover${passportId ? `?passport=${passportId}` : ""}`,
  handovers: "/handovers",
  handoverDetail: (id: string) => `/handovers/${id}`,
  transform: "/transform",
  scan: "/scan",
  certificates: "/certificates",
  certificateNew: "/certificates/new",
  flags: "/flags",
  flag: (id: string) => `/flags?flag=${id}`,
  profile: "/profile",
  users: "/users",
  userDetail: (id: string) => `/users/${id}`,
  notifications: "/notifications",
  invite: (token: string) => `/invite/${token}`,
};

/** Keyless street map (OpenFreeMap). ponytail: no satellite layer; add one if a keyless source with an acceptable licence turns up. */
export const mapStyleURL = "https://tiles.openfreemap.org/styles/liberty";

/** The browser's way to the API: the pass-through in app/api/[...path]/route.ts. */
export const apiBase = "/api";

export const apiPaths = {
  me: "/me",
  login: "/auth/login",
  logout: "/auth/logout",
  emailAvailability: "/auth/email/availability",
  forgotPassword: "/auth/password/forgot",
  resetPassword: "/auth/password/reset",
  lockAccount: "/auth/lock",
  verifyEmail: "/auth/email/verify",
  handoff: "/auth/handoff",
  handoffAccept: "/auth/handoff/accept",
  orgs: "/orgs",
  org: (id: string) => `/orgs/${id}`,
  orgDocuments: (id: string) => `/orgs/${id}/documents`,
  orgMembers: (id: string) => `/orgs/${id}/members`,
  orgSuspend: (id: string) => `/orgs/${id}/suspend`,
  orgReactivate: (id: string) => `/orgs/${id}/reactivate`,
  orgReview: (id: string) => `/orgs/${id}/verify`,
  orgAccreditations: (id: string) => `/orgs/${id}/accreditations`,
  documentFile: (id: string) => `/documents/${id}/file`,
  templates: "/templates",
  template: (id: string) => `/templates/${id}`,
  verify: (id: string) => `/verify/${encodeURIComponent(id)}`,
  passports: "/passports",
  passport: (id: string) => `/passports/${encodeURIComponent(id)}`,
  passportRetire: (id: string) => `/passports/${encodeURIComponent(id)}/retire`,
  passportHandovers: (id: string) =>
    `/passports/${encodeURIComponent(id)}/handovers`,
  passportGenealogy: (id: string) =>
    `/passports/${encodeURIComponent(id)}/genealogy`,
  transforms: "/transforms",
  handovers: "/handovers",
  handover: (id: string) => `/handovers/${encodeURIComponent(id)}`,
  handoverAccept: (id: string) => `/handovers/${encodeURIComponent(id)}/accept`,
  handoverReject: (id: string) => `/handovers/${encodeURIComponent(id)}/reject`,
  handoverCancel: (id: string) => `/handovers/${encodeURIComponent(id)}/cancel`,
  handoverPhoto: (id: string, photoId: string) =>
    `/handovers/${encodeURIComponent(id)}/photos/${encodeURIComponent(photoId)}`,
  flags: "/flags",
  flag: (id: string) => `/flags/${encodeURIComponent(id)}`,
  flagAssignee: (id: string) => `/flags/${encodeURIComponent(id)}/assignee`,
  flagResolve: (id: string) => `/flags/${encodeURIComponent(id)}/resolve`,
  attestations: "/attestations",
  attestation: (id: string) => `/attestations/${encodeURIComponent(id)}`,
  attestationFile: (id: string) =>
    `/attestations/${encodeURIComponent(id)}/file`,
  attestationRevoke: (id: string) =>
    `/attestations/${encodeURIComponent(id)}/revoke`,
  facilities: "/facilities",
  geocode: "/geocode",
  facilityKinds: "/facility-kinds",
  facilityKindsAll: "/facility-kinds/all",
  facilityKind: (key: string) => `/facility-kinds/${encodeURIComponent(key)}`,
  facility: (id: string) => `/facilities/${encodeURIComponent(id)}`,
  facilityStatus: (id: string) =>
    `/facilities/${encodeURIComponent(id)}/status`,
  facilityApprove: (id: string) =>
    `/facilities/${encodeURIComponent(id)}/approve`,
  facilityLicence: (id: string) =>
    `/facilities/${encodeURIComponent(id)}/licence`,
  directory: "/orgs/directory",
  audit: "/audit",
  auditVerify: "/audit/verify",
  reportsVolume: "/reports/volume",
  reportsBlocked: "/reports/blocked",
  reportsOutcomes: "/reports/verification-outcomes",
  reportsSaved: "/reports/saved",
  reportSaved: (id: string) => `/reports/saved/${encodeURIComponent(id)}`,
  reportRun: (id: string) => `/reports/saved/${encodeURIComponent(id)}/run`,
  reportFiles: (id: string) => `/reports/saved/${encodeURIComponent(id)}/files`,
  reportFile: (id: string) => `/reports/files/${encodeURIComponent(id)}`,
  myOrganisations: "/me/organisations",
  password: "/auth/password",
  pin: "/auth/pin",
  twoFactorSetup: "/auth/2fa/setup",
  twoFactorEnable: "/auth/2fa/enable",
  twoFactorDisable: "/auth/2fa/disable",
  twoFactorRecovery: "/auth/2fa/recovery",
  notifications: "/notifications",
  stats: "/stats",
  notificationRead: (id: string) => `/notifications/${id}/read`,
  users: "/users",
  userRole: (id: string) => `/users/${id}/role`,
  userSuspend: (id: string) => `/users/${id}/suspend`,
  userReinstate: (id: string) => `/users/${id}/reinstate`,
  user: (id: string) => `/users/${id}`,
  userDeactivate: (id: string) => `/users/${id}/deactivate`,
  userActivate: (id: string) => `/users/${id}/activate`,
  invites: "/invites",
  inviteAccept: (token: string) =>
    `/invites/${encodeURIComponent(token)}/accept`,
};

export const queryKeys = {
  me: ["me"] as const,
  verify: (id: string) => ["verify", id] as const,
  orgs: ["orgs"] as const,
  orgList: (status: string) => ["orgs", "list", status] as const,
  org: (id: string) => ["orgs", id] as const,
  orgDocuments: (id: string) => ["orgs", id, "documents"] as const,
  orgAccreditations: (id: string) => ["orgs", id, "accreditations"] as const,
  templates: ["templates"] as const,
  template: (id: string) => ["templates", id] as const,
  count: (tile: string) => ["dashboard", "count", tile] as const,
  panel: (panel: string) => ["dashboard", "panel", panel] as const,
  directory: ["orgs", "directory"] as const,
  myOrganisations: ["me", "organisations"] as const,
  notifications: ["notifications"] as const,
  stats: ["stats"] as const,
  unread: ["notifications", "unread"] as const,
  users: ["users"] as const,
  user: (id: string) => ["users", id] as const,
  facilities: ["facilities"] as const,
  facilityKinds: ["facility-kinds"] as const,
  geocode: (query: string) => ["geocode", query] as const,
  emailAvailability: (email: string) => ["email-availability", email] as const,
  facility: (id: string) => ["facilities", id] as const,
  passports: ["passports"] as const,
  passport: (id: string) => ["passports", id] as const,
  handovers: ["handovers"] as const,
  handover: (id: string) => ["handovers", id] as const,
  certificates: ["certificates"] as const,
  certificate: (id: string) => ["certificates", id] as const,
  flags: ["flags"] as const,
  flag: (id: string) => ["flags", id] as const,
  audit: ["audit"] as const,
  volume: ["audit", "volume"] as const,
  outcomes: ["audit", "outcomes"] as const,
  saved: ["audit", "saved"] as const,
};

// ---- Copy ----

export const copy = {
  ...copyJson,
  errors: {
    ...copyJson.errors,
    unexpectedRef: (id: string) =>
      fill(copyJson.errors.unexpectedRef.$template, { id }),
  },
  removeFile: (name: string) => fill(copyJson.removeFile.$template, { name }),
  shell: {
    ...copyJson.shell,
    nav: copyJson.shell.nav as Record<NavKey, string>,
    groups: copyJson.shell.groups as Record<NavGroup, string>,
    theme: copyJson.shell.theme as Record<ThemeChoice, string>,
    yourProfile: (name: string) =>
      fill(copyJson.shell.yourProfile.$template, { name }),
  },
  list: {
    ...copyJson.list,
    range: (from: number, to: number, total: number) =>
      `${from.toLocaleString()}–${to.toLocaleString()} of ${total.toLocaleString()}`,
    page: (page: number, pages: number) =>
      fill(copyJson.list.page.$template, { page, pages }),
  },
  notifications: {
    ...copyJson.notifications,
    kinds: copyJson.notifications.kinds as StrMap,
    open: (n: number) => (n ? `Notifications, ${n} unread` : "Notifications"),
  },
  profile: {
    ...copyJson.profile,
  },
  signin: {
    ...copyJson.signin,
    confirmBody: (email: string) =>
      fill(copyJson.signin.confirmBody.$template, { email }),
  },
  users: {
    ...copyJson.users,
    status: copyJson.users.status as StrMap,
    since: (d: string) => fill(copyJson.users.since.$template, { d }),
  },
  dashboard: {
    ...copyJson.dashboard,
    greeting: (name: string) => {
      const h = new Date().getHours();
      return `${h < 12 ? "Good morning" : h < 18 ? "Good afternoon" : "Good evening"}, ${name.split(" ")[0]}`;
    },
    from: (name: string) => fill(copyJson.dashboard.from.$template, { name }),
    to: (name: string) => fill(copyJson.dashboard.to.$template, { name }),
    stage: (stage: string) =>
      fill(copyJson.dashboard.stage.$template, { stage }),
    capacityOf: (used: string, max: string) =>
      fill(copyJson.dashboard.capacityOf.$template, { used, max }),
    capacityLeft: (left: string) =>
      fill(copyJson.dashboard.capacityLeft.$template, { left }),
    waitingDays: (d: number) =>
      d === 0 ? "Today" : d === 1 ? "1 day" : `${d} days`,
    workspaces: {
      ...copyJson.dashboard.workspaces,
      sub: (n: number) =>
        n === 1
          ? "You are a member of 1 organisation. Open it to start working."
          : `You are a member of ${n} organisations. Pick one to start working.`,
    },
    tiles: copyJson.dashboard.tiles as Record<
      TileKey,
      { label: string; note: string }
    >,
    actions: copyJson.dashboard.actions as Record<ActionKey, string>,
    panels: copyJson.dashboard.panels as Record<
      PanelKey,
      { title: string; sub: string; empty: string }
    >,
  },
  documentTypeLabel: (type: string) =>
    documentTypes.find((d) => d.value === type)?.label ?? type,
  stepOf: (n: number, total: number) =>
    fill(copyJson.stepOf.$template, { n, total }),
  keyLabel: (key: string) =>
    key.replace(/_/g, " ").replace(/^./, (c) => c.toUpperCase()),
  facilities: {
    ...copyJson.facilities,
    sub: (n: number, org: string) =>
      `${n} ${n === 1 ? "facility" : "facilities"} registered · ${org}`,
    suspendTitle: (name: string) =>
      fill(copyJson.facilities.suspendTitle.$template, { name }),
    detail: {
      ...copyJson.facilities.detail,
      mapLabel: (name: string) =>
        fill(copyJson.facilities.detail.mapLabel.$template, { name }),
      area: (ha: number) =>
        `${new Intl.NumberFormat("en", { maximumFractionDigits: 2 }).format(ha)} ha`,
      declaredOf: (n: number, unit: string, period: string) =>
        `${n.toLocaleString()} ${unit} per ${period}`,
      issued: (used: number, unit: string, period: string) =>
        `${used.toLocaleString()} ${unit} issued this ${period}`,
      deleteTitle: (name: string) =>
        fill(copyJson.facilities.detail.deleteTitle.$template, { name }),
    },
    new: {
      ...copyJson.facilities.new,
      corners: (n: number) => `${n} ${n === 1 ? "corner" : "corners"}`,
    },
  },
  issue: {
    ...copyJson.issue,
    facilityLeft: (n: number, unit: string, period: string) =>
      `${n.toLocaleString()} ${unit} left this ${period}`,
    after: (left: number, unit: string, period: string) =>
      `You have ${left.toLocaleString()} ${unit} left this ${period}. After this passport: ${Math.max(left, 0).toLocaleString()} ${unit}.`,
    evidenceMissing: (keys: string) =>
      fill(copyJson.issue.evidenceMissing.$template, { keys }),
    blockedField: (limit: number, unit: string, period: string) =>
      `More than the ${limit.toLocaleString()} ${unit} left for this facility this ${period}.`,
    blockedNotice: (limit: number, unit: string) =>
      `This passport can't be issued. Lower the quantity to ${limit.toLocaleString()} ${unit} or less. If the harvest really is larger, ask the regulator to review the declared capacity.`,
    createdBody: (quantity: string, facility: string) =>
      fill(copyJson.issue.createdBody.$template, { quantity, facility }),
  },
  handover: {
    ...copyJson.handover,
    holds: (n: number, unit: string) =>
      `This passport holds ${n.toLocaleString()} ${unit}.`,
    partial: (rest: string) =>
      fill(copyJson.handover.partial.$template, { rest }),
    locationAdded: (lat: string, lon: string) =>
      fill(copyJson.handover.locationAdded.$template, { lat, lon }),
    photosHelp: (max: number) =>
      fill(copyJson.handover.photosHelp.$template, { max }),
    tooManyPhotos: (max: number) =>
      fill(copyJson.handover.tooManyPhotos.$template, { max }),
    meaning: (receiver: string) =>
      fill(copyJson.handover.meaning.$template, { receiver }),
    sentBody: (quantity: string, receiver: string) =>
      fill(copyJson.handover.sentBody.$template, { quantity, receiver }),
  },
  receiver: {
    ...copyJson.receiver,
    rejectChips: copyJson.receiver.rejectChips as Record<RejectReason, string>,
    from: (name: string) => fill(copyJson.receiver.from.$template, { name }),
    to: (name: string) => fill(copyJson.receiver.to.$template, { name }),
    within: (n: string, unit: string, pct: string, tol: string) =>
      fill(copyJson.receiver.within.$template, { n, unit, pct, tol }),
    outside: (n: string, unit: string, pct: string, tol: string) =>
      fill(copyJson.receiver.outside.$template, { n, unit, pct, tol }),
    locationAdded: (lat: string, lon: string) =>
      fill(copyJson.receiver.locationAdded.$template, { lat, lon }),
    photosHelp: (max: number) =>
      fill(copyJson.receiver.photosHelp.$template, { max }),
    meaning: (sender: string) =>
      fill(copyJson.receiver.meaning.$template, { sender }),
    completeBody: (quantity: string) =>
      fill(copyJson.receiver.completeBody.$template, { quantity }),
    rejectStays: (sender: string) =>
      fill(copyJson.receiver.rejectStays.$template, { sender }),
    partOf: (code: string) =>
      fill(copyJson.receiver.partOf.$template, { code }),
    rejectedBody: (sender: string) =>
      fill(copyJson.receiver.rejectedBody.$template, { sender }),
  },
  transform: {
    ...copyJson.transform,
    selected: (n: number, quantity: string) =>
      `${n} ${n === 1 ? "lot" : "lots"} selected · ${quantity}`,
    output: (n: number) => fill(copyJson.transform.output.$template, { n }),
    meterWithin: (below: string) =>
      fill(copyJson.transform.meterWithin.$template, { below }),
    meterOver: (over: string) =>
      fill(copyJson.transform.meterOver.$template, { over }),
    yieldLine: (yieldPct: string, maxPct: string) =>
      fill(copyJson.transform.yieldLine.$template, { yieldPct, maxPct }),
    inOut: (into: string, out: string) =>
      fill(copyJson.transform.inOut.$template, { into, out }),
    sign: (n: number) =>
      `Sign and create ${n} ${n === 1 ? "passport" : "passports"}`,
    createdTitle: (n: number) =>
      `${n} new ${n === 1 ? "passport" : "passports"} created`,
    createdBody: (
      into: string,
      out: string,
      yieldPct: string,
      maxPct: string,
    ) =>
      fill(copyJson.transform.createdBody.$template, {
        into,
        out,
        yieldPct,
        maxPct,
      }),
  },
  passportView: {
    ...copyJson.passportView,
    status: copyJson.passportView.status as Record<PassportStatus, string>,
    eventKinds: copyJson.passportView.eventKinds as StrMap,
    eventFrom: (name: string) =>
      fill(copyJson.passportView.eventFrom.$template, { name }),
    eventTo: (name: string) =>
      fill(copyJson.passportView.eventTo.$template, { name }),
  },
  certificates: {
    ...copyJson.certificates,
    expiringIn: (days: number) =>
      `Expiring in ${days} ${days === 1 ? "day" : "days"}`,
    attachedLot: (code: string) =>
      fill(copyJson.certificates.attachedLot.$template, { code }),
    attachedFacility: (name: string) => name,
    new: {
      ...copyJson.certificates.new,
      checkVerdict: (verdict: string) =>
        fill(copyJson.certificates.new.checkVerdict.$template, { verdict }),
    },
  },
  flags: {
    ...copyJson.flags,
    source: copyJson.flags.source as StrMap,
    sub: (n: number) =>
      n === 0 ? "Nothing open" : `${n} ${n === 1 ? "flag" : "flags"} open`,
    raised: (when: string) => fill(copyJson.flags.raised.$template, { when }),
    fraudBody: (n: number) =>
      `${n} ${n === 1 ? "passport" : "passports"} will be revoked, including everything made from them.`,
    fraudConfirm: (n: number) =>
      `Confirm fraud and revoke ${n} ${n === 1 ? "lot" : "lots"}`,
  },
  templates: {
    ...copyJson.templates,
    detail: {
      ...copyJson.templates.detail,
      title: (name: string) =>
        fill(copyJson.templates.detail.title.$template, { name }),
      eventChip: (events: string) => events,
    },
  },
  reports: {
    ...copyJson.reports,
    event: (method: string, route: string) =>
      fill(copyJson.reports.event.$template, { method, route }),
    footer: (shown: number, total: number) =>
      `${shown} of ${total.toLocaleString()} events · Every row is sealed; this log can't be edited`,
    verifyBad: (seq: number) =>
      fill(copyJson.reports.verifyBad.$template, { seq }),
  },
  forgot: {
    ...copyJson.forgot,
    sent: (email: string) => fill(copyJson.forgot.sent.$template, { email }),
  },
  reset: {
    ...copyJson.reset,
    rules: copyJson.reset.rules,
  },
  signup: {
    ...copyJson.signup,
    passwordHelp: (min: number, max: number) =>
      fill(copyJson.signup.passwordHelp.$template, { min, max }),
  },
  validation: {
    ...copyJson.validation,
    tooLong: (n: number) => fill(copyJson.validation.tooLong.$template, { n }),
    passwordShort: (n: number) =>
      fill(copyJson.validation.passwordShort.$template, { n }),
  },
  onboarding: {
    ...copyJson.onboarding,
    pendingSub: (name: string) =>
      fill(copyJson.onboarding.pendingSub.$template, { name }),
    rejected: (name: string) =>
      fill(copyJson.onboarding.rejected.$template, { name }),
    submitted: (when: string) =>
      fill(copyJson.onboarding.submitted.$template, { when }),
    progress: {
      ...copyJson.onboarding.progress,
      stillNeeded: (list: string) =>
        fill(copyJson.onboarding.progress.stillNeeded.$template, { list }),
    },
  },
  admin: {
    ...copyJson.admin,
    sub: (count: number, oldestDays: number | undefined) =>
      [
        `${count} ${count === 1 ? "organisation" : "organisations"}`,
        oldestDays !== undefined &&
          `oldest waiting ${oldestDays} ${oldestDays === 1 ? "day" : "days"}`,
        "target under 2 working days",
      ]
        .filter(Boolean)
        .join(" · "),
    tabCount: (label: string, n: number) =>
      fill(copyJson.admin.tabCount.$template, { label, n }),
    queueLabel: (label: string) =>
      fill(copyJson.admin.queueLabel.$template, { label }),
    age: (days: number) =>
      days === 0 ? "Today" : `${days} ${days === 1 ? "day" : "days"}`,
    submitted: (when: string) =>
      fill(copyJson.admin.submitted.$template, { when }),
    accreditation: {
      ...copyJson.admin.accreditation,
      current: (n: number) =>
        n === 0
          ? "Not accredited"
          : `Accredited for ${n} ${n === 1 ? "type" : "types"}`,
    },
  },
  orgStatus: copyJson.orgStatus as Record<OrgStatus, string>,
  verification: {
    ...copyJson.verification,
    checks: copyJson.verification.checks as CheckMap,
    eventTitles: copyJson.verification.eventTitles as StrMap,
    passportStatus: copyJson.verification.passportStatus as StrMap,
    genealogyBody: (n: number) =>
      n === 0
        ? "This lot was issued at its origin. It was not made from other lots."
        : `This lot traces back through ${n} ${n === 1 ? "earlier lot" : "earlier lots"} to its origin.`,
    via: (stage: string, used: string) =>
      fill(copyJson.verification.via.$template, { stage, used }),
    passedOf: (n: number, total: number) =>
      fill(copyJson.verification.passedOf.$template, { n, total }),
    originMap: (name: string) =>
      fill(copyJson.verification.originMap.$template, { name }),
    steps: (n: number) => `${n} ${n === 1 ? "step" : "steps"}`,
    lotsToOrigin: (n: number) =>
      `${n} ${n === 1 ? "lot" : "lots"}, back to origin`,
    summary: {
      ...copyJson.verification.summary,
      from: (product: string, places: string) =>
        fill(copyJson.verification.summary.from.$template, { product, places }),
      warning: (points: string) =>
        fill(copyJson.verification.summary.warning.$template, { points }),
      pointsNeed: (n: number) =>
        `${n} ${n === 1 ? "point needs" : "points need"}`,
    },
    meta: {
      ...copyJson.verification.meta,
      checked: (when: string) =>
        fill(copyJson.verification.meta.checked.$template, { when }),
      records: (n: number) => `${n} ${n === 1 ? "record" : "records"}`,
      organisations: (n: number) =>
        `${n} ${n === 1 ? "organisation" : "organisations"}`,
    },
    issued: (when: string) =>
      fill(copyJson.verification.issued.$template, { when }),
    certificateNote: {
      ...copyJson.verification.certificateNote,
      revoked: (reason: string | null) =>
        `Revoked by its issuer${reason ? `: ${reason}` : "."} Ask the holder for a replacement.`,
      expired: (when: string) =>
        fill(copyJson.verification.certificateNote.expired.$template, { when }),
    },
    technical: {
      ...copyJson.verification.technical,
      eventSealed: (title: string, when: string) =>
        fill(copyJson.verification.technical.eventSealed.$template, {
          title,
          when,
        }),
      certificateFingerprint: (type: string) =>
        fill(copyJson.verification.technical.certificateFingerprint.$template, {
          type,
        }),
    },
  },
};

// ---- Lists: tabs and sorts shared by the list screens ----

/** Handover boxes (?box=), as the API sorts them for the acting organisation. */
export const handoverBoxes = ["incoming", "outgoing", "done"] as const;
export type HandoverBox = (typeof handoverBoxes)[number];
/** Flag queue tabs; "resolved" is cleared or confirmed (the API's list filter). */
export const flagTabs = ["open", "resolved"] as const;
export const notificationTabs = ["all", "unread"] as const;
/** How a table shows its rows; unset means cards on phones, a table from md up. */
export const listViews = ["table", "cards"] as const;
export type ListView = (typeof listViews)[number];
/** Certificates an organisation sees: issued by it, or about its lots and facilities. */
export const certificateScopes = ["issued", "about"] as const;
export type CertificateScope = (typeof certificateScopes)[number];

const byDate = [
  { value: "-created_at", label: copy.list.newest },
  { value: "created_at", label: copy.list.oldest },
];
/** Sort choices, by what a list can be sorted on (the API's listq Sort columns). */
export const listSorts = {
  date: byDate,
  name: [
    { value: "name", label: copy.list.nameAZ },
    { value: "-name", label: copy.list.nameZA },
    ...byDate,
  ],
  certificate: [
    ...byDate,
    { value: "valid_to", label: copy.list.endingSoonest },
    { value: "type", label: copy.list.typeAZ },
  ],
  waiting: [
    { value: "created_at", label: copy.list.oldest },
    { value: "-created_at", label: copy.list.newest },
    { value: "name", label: copy.list.nameAZ },
  ],
};

// ---- Marketing home (/) ----
// Only what the product does today: no offline sync, no other languages, no
// help centre, no seal or certificate numbers, no flag severity. Example lots
// and places are labelled as examples.

export const marketingRoles = [
  "producer",
  "logistics",
  "certifier",
  "regulator",
] as const;
export type MarketingRole = (typeof marketingRoles)[number];

export const marketing = {
  ...marketingJson,
  trust: {
    ...marketingJson.trust,
    verdicts: marketingJson.trust.verdicts as {
      label: string;
      body: string;
      tone: "verified" | "warning" | "danger";
    }[],
  },
  footer: {
    ...marketingJson.footer,
    rights: (year: number) =>
      fill(marketingJson.footer.rights.$template, { year }),
  },
};

// ---- Platform dashboard (/dashboard for the platform; GET /stats) ----

export const platformCopy = {
  ...platformCopyJson,
  updated: (when: string) => fill(platformCopyJson.updated.$template, { when }),
  tiles: {
    ...platformCopyJson.tiles,
    organisations: {
      ...platformCopyJson.tiles.organisations,
      note: (pending: number) =>
        `${pending.toLocaleString()} waiting for review`,
    },
    passports: {
      ...platformCopyJson.tiles.passports,
      note: (active: number) => `${active.toLocaleString()} active`,
    },
  },
  outbox: {
    ...platformCopyJson.outbox,
    oldest: (minutes: number) =>
      `Oldest waiting ${minutes.toLocaleString()} min`,
  },
};

// ---- Reports and audit (/admin/reports) ----

/** Download formats the report endpoints take (?format=). */
export const exportFormats = ["csv", "pdf"] as const;

export const reportsCopy = {
  ...reportsCopyJson,
  checked: (n: number) => fill(reportsCopyJson.checked.$template, { n }),
  brokenAt: (n: number | null) => (n ? `Broken at seq ${n}` : "Broken"),
};

// ---- Screen copy that grew inside the screens; kept per screen ----

export const screenCopy = {
  ...screenCopyJson,
  issue: {
    ...screenCopyJson.issue,
    afterPassport: (used: number, unit: string) =>
      `With this passport: ${used.toLocaleString()} ${unit} issued.`,
    evidenceCount: (have: number, total: number) =>
      fill(screenCopyJson.issue.evidenceCount.$template, { have, total }),
  },
  passportDetail: {
    ...screenCopyJson.passportDetail,
    held: (held: number, total: number, unit: string) =>
      `${held.toLocaleString()} of ${total.toLocaleString()} ${unit}`,
    edge: (stage: string, used: string) =>
      fill(screenCopyJson.passportDetail.edge.$template, { stage, used }),
  },
  transform: {
    ...screenCopyJson.transform,
    templateLine: (name: string) =>
      fill(screenCopyJson.transform.templateLine.$template, { name }),
    allowedYield: (max: number) =>
      fill(screenCopyJson.transform.allowedYield.$template, { max }),
    overRemaining: (max: number, unit: string) =>
      `More than the ${max.toLocaleString()} ${unit} this lot has left. Enter ${max.toLocaleString()} ${unit} or less.`,
    tooManyRows: (n: number) =>
      fill(screenCopyJson.transform.tooManyRows.$template, { n }),
    overSaid: (out: string, max: string) =>
      fill(screenCopyJson.transform.overSaid.$template, { out, max }),
    formula: (into: string, pct: string, max: string) =>
      fill(screenCopyJson.transform.formula.$template, { into, pct, max }),
  },
};

// ---- Confirmations (components/chaintrack/confirm.tsx) ----

export const confirmCopy = {
  ...confirmCopyJson,
  facilityApprove: (name: string) => ({
    title: `Approve ${name}?`,
    body: "It goes on the ledger and can issue passports from now on, up to its declared capacity.",
    label: "Approve facility",
  }),
  facilityReinstate: (name: string) => ({
    title: `Reinstate ${name}?`,
    body: "It can issue passports again.",
    label: "Reinstate facility",
  }),
  userSuspend: (name: string) => ({
    title: `Suspend ${name}?`,
    body: "They are signed out here and can't act for this organisation until reinstated. What they signed stays signed.",
    label: "Suspend",
  }),
  userReinstate: (name: string) => ({
    title: `Reinstate ${name}?`,
    body: "They can sign in and act for this organisation again, with their role.",
    label: "Reinstate",
  }),
  userDeactivate: (name: string) => ({
    title: `Deactivate ${name}'s account?`,
    body: "They are signed out everywhere and can't sign in to any organisation until the account is activated again. What they signed stays signed.",
    label: "Deactivate account",
  }),
  userActivate: (name: string) => ({
    title: `Activate ${name}'s account?`,
    body: "They can sign in again, with the memberships they had.",
    label: "Activate account",
  }),
  userScopeSuspend: (name: string, org: string) => ({
    title: `Suspend ${name} in ${org}?`,
    body: "They can't act for this organisation until reinstated. Their account and other organisations stay as they are.",
    label: "Suspend",
  }),
  userScopeReinstate: (name: string, org: string) => ({
    title: `Reinstate ${name} in ${org}?`,
    body: "They can act for this organisation again, with their role.",
    label: "Reinstate",
  }),
  orgSuspend: (name: string) => ({
    title: `Suspend ${name}?`,
    body: "Its members can't use the workspace or sign anything until it is reactivated. What it signed stays signed.",
    label: "Suspend organisation",
  }),
  orgReactivate: (name: string) => ({
    title: `Reactivate ${name}?`,
    body: "It is verified again and can act at once.",
    label: "Reactivate",
  }),
  userRole: (name: string, role: string) => ({
    title: `Make ${name} ${role}?`,
    body: "Their permissions here change at once.",
    label: "Change role",
  }),
};
