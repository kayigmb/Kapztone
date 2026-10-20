// Every value and every piece of UI copy the frontend uses, declared once.
// Statuses, kinds and paths mirror the API (backend/internal/*/values.go);
// copy follows chaintrack-ui → content.md. A second language is a
// translation of `copy` below, not a rewrite of the pages.
//
// Pure: no imports, so `node --test` can load it.

// ---- Values shared with the API ----

export const orgStatus = {
  pending: "pending",
  verified: "verified",
  rejected: "rejected",
  changesRequested: "changes_requested",
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
  { value: "approved", statuses: [orgStatus.verified] },
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
export const facilityKinds = [
  { value: "farm_plot", label: "Farm plot", sub: "Land growing a crop" },
  { value: "mine_site", label: "Mine site", sub: "Extraction site" },
  {
    value: "washing_station",
    label: "Washing station",
    sub: "Coffee or tea processing",
  },
  {
    value: "processing_plant",
    label: "Processing plant",
    sub: "Factory, mill or smelter",
  },
  { value: "warehouse", label: "Warehouse", sub: "Storage and depot" },
] as const;
export type FacilityKind = (typeof facilityKinds)[number]["value"];
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
    key: "notifications",
    href: "/notifications",
    icon: "notifications",
    group: "workspace",
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
  handoff: (code: string) => `/handoff?code=${encodeURIComponent(code)}`,
  onboarding: "/onboarding",
  adminOrgs: "/admin/organisations",
  adminOrg: (tab: string, id?: string) =>
    `/admin/organisations?tab=${tab}${id ? `&id=${id}` : ""}`,
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
  forgotPassword: "/auth/password/forgot",
  resetPassword: "/auth/password/reset",
  lockAccount: "/auth/lock",
  handoff: "/auth/handoff",
  handoffAccept: "/auth/handoff/accept",
  orgs: "/orgs",
  org: (id: string) => `/orgs/${id}`,
  orgDocuments: (id: string) => `/orgs/${id}/documents`,
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
  facilities: ["facilities"] as const,
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
  brand: "ChainTrack",
  loading: "Loading…",

  // What went wrong, and what to do next (content.md → Errors).
  errors: {
    unreachable:
      "We couldn't reach ChainTrack just now. Nothing was changed. Try again in a moment.",
    unexpected:
      "Something went wrong on our side. Nothing was changed. Try again in a moment.",
    unexpectedRef: (id: string) =>
      `Something went wrong on our side. Try again; if it keeps happening, quote reference ${id}.`,
    badAnswer:
      "ChainTrack answered in a way this page doesn't understand. Reload the page.",
    tryAgain: "Try again",
    loadTitle: "We couldn't load this page",
  },

  signOut: "Sign out",
  password: { show: "Show password", hide: "Hide password" },
  handoverStatusLabel: { expired: "Expired" },
  removeLocation: "Remove location",
  removeFile: (name: string) => `Remove ${name}`,

  shell: {
    nav: {
      home: "Dashboard",
      workspaces: "Workspaces",
      facilities: "Facilities",
      issue: "Issue passport",
      handovers: "Handovers",
      transform: "Transform",
      certificates: "Certificates",
      flags: "Flags",
      adminOrgs: "Verify organisations",
      templates: "Templates",
      reports: "Reports",
      users: "Users",
      profile: "Profile",
      notifications: "Notifications",
    } as Record<NavKey, string>,
    groups: {
      workspace: "Workspace",
      trust: "Trust & compliance",
      manage: "Manage",
    } as Record<NavGroup, string>,
    eyebrow: "Workspace",
    menu: "Menu",
    more: "More",
    scan: "Scan",
    scanQr: "Scan QR",
    platform: "ChainTrack platform",
    back: "Back",
    collapse: "Fold the sidebar",
    expand: "Unfold the sidebar",
    theme: { system: "System", light: "Light", dark: "Dark" } as Record<
      ThemeChoice,
      string
    >,
    themeMenu: "Theme",
    yourProfile: (name: string) => `${name}: your profile`,
    switchOrg: "Switch organisation",
    switchNote: "Opens that organisation's address; you stay signed in.",
    current: "Current",
  },

  list: {
    search: "Search",
    sort: "Sort by",
    perPage: "Rows",
    viewAs: "Show as",
    actions: "Actions",
    views: { table: "Table", cards: "Cards" },
    filter: "Show",
    all: "All",
    prev: "Previous page",
    next: "Next page",
    range: (from: number, to: number, total: number) =>
      `${from.toLocaleString()}–${to.toLocaleString()} of ${total.toLocaleString()}`,
    page: (page: number, pages: number) => `Page ${page} of ${pages}`,
    noMatch: "Nothing matches. Try another search or filter.",
    newest: "Newest first",
    oldest: "Oldest first",
    nameAZ: "Name A–Z",
    nameZA: "Name Z–A",
    endingSoonest: "Ending soonest",
    typeAZ: "Type A–Z",
    statusAZ: "Status",
  },

  notifications: {
    title: "Notifications",
    open: (n: number) => (n ? `Notifications, ${n} unread` : "Notifications"),
    empty: "Nothing new.",
    markRead: "Mark as read",
    markAllRead: "Mark all as read",
    read: "Read",
    unread: "Unread",
    viewAll: "View all notifications",
    sub: "Everything that happened to your lots, handovers, certificates and flags. Open one to go to it.",
    tabs: { all: "All", unread: "Unread" },
    emptyUnread: "You're all caught up.",
    // Short title per kind; the API's message is the body. An unknown kind shows the message alone.
    kinds: {
      org_registered: "New organisation registered",
      documents_uploaded: "Documents ready for review",
      flag_raised: "Flag raised",
      flag_resolved: "Flag resolved",
      handover_started: "Handover waiting for you",
      handover_accepted: "Handover accepted",
      handover_rejected: "Handover rejected",
      certificate_issued: "Certificate issued",
      certificate_revoked: "Certificate revoked",
      passport_revoked: "Passport revoked",
    } as Record<string, string>,
  },

  profile: {
    title: "Your profile",
    sub: "Your account and how you sign in and sign records.",
    details: "Details",
    name: "Full name",
    email: "Email",
    organisation: "Organisation",
    role: "Role",
    saveName: "Save name",
    password: "Password",
    currentPassword: "Current password",
    newPassword: "New password",
    newPasswordHelp: `${limits.password.min} to ${limits.password.max} characters.`,
    confirmPassword: "Confirm new password",
    changePassword: "Change password",
    pin: "Signing PIN",
    pinHelpSet: "Your PIN is set. You enter it whenever you sign a record.",
    pinHelpUnset:
      "Set a 6-digit PIN. You'll need it to sign passports, handovers and certificates.",
    pinLabel: "New 6-digit PIN",
    pinPassword: "Your password",
    setPin: "Set PIN",
    changePin: "Change PIN",
    pinStatusSet: "Set",
    pinStatusUnset: "Not set",
    twoFactor: "Two-factor sign-in",
    twoFactorHelp:
      "Ask for a code from an authenticator app each time you sign in.",
    startSetup: "Set up two-factor sign-in",
    scanQr: "Scan this with your authenticator app, or type the key below.",
    secret: "Key",
    code: "6-digit code from the app",
    turnOn: "Turn on",
    turnOff: "Turn off two-factor sign-in",
    turnOffHelp: "Enter a current code from your app to turn it off.",
    recoveryTitle: "Save your recovery codes",
    recoveryBody:
      "Each code signs you in once if you lose your phone. They won't be shown again.",
    newRecovery: "Get new recovery codes",
    newRecoveryHelp:
      "Enter a current code from your app. Your old recovery codes stop working.",
    done: "I've saved them",
  },

  users: {
    title: "Users",
    subOrg: "People who can act for your organisation.",
    subPlatform: "Everyone with a ChainTrack account.",
    search: "Search by name or email",
    empty: "No users match.",
    since: (d: string) => `Since ${d}`,
    memberships: "Organisations",
    columns: {
      person: "Person",
      role: "Role",
      status: "Status",
      joined: "Joined",
      twoFactor: "Two-factor",
      actions: "Actions",
    },
    noMemberships: "No organisation",
    twoFactorOn: "Two-factor on",
    twoFactorOff: "Two-factor off",
    status: {
      active: "Active",
      suspended: "Suspended",
      disabled: "Disabled",
    } as Record<string, string>,
    invite: "Invite person",
    inviteEmail: "Email",
    inviteRole: "Role",
    sendInvite: "Send invitation",
    notEmailed:
      "E-mail is off, so nothing was sent. Share this link with them (development only):",
    role: "Role",
    changeRole: "Change role",
    suspend: "Suspend",
    reinstate: "Reinstate",
    you: "You",
    notAllowedTitle: "You can't manage users.",
    notAllowedBody:
      "Only organisation admins and the platform admin can open this page.",
  },

  invite: {
    eyebrow: "Invitation",
    title: "Join an organisation",
    sub: "You've been invited to act for an organisation on ChainTrack.",
    name: "Full name",
    nameHelp: "Leave empty if you already have a ChainTrack account.",
    password: "Password",
    passwordHelp:
      "New account: choose one, 12 to 72 characters. Existing account: your current password.",
    submit: "Accept invitation",
    submitting: "Accepting…",
    done: "You're in.",
    signInAt: "Sign in at your organisation's address",
  },

  dashboard: {
    greeting: (name: string) => {
      const h = new Date().getHours();
      return `${h < 12 ? "Good morning" : h < 18 ? "Good afternoon" : "Good evening"}, ${name.split(" ")[0]}`;
    },
    platformLine: "Platform administration",
    dateHeader: "Date",
    showTable: "Show as a table",
    tiles: {
      passportsHeld: { label: "Passports you hold", note: "Active lots" },
      incoming: {
        label: "Incoming to accept",
        note: "Waiting for your signature",
      },
      outgoing: { label: "Sent, awaiting receiver", note: "Signed by you" },
      openFlags: { label: "Open flags", note: "Under review" },
      certValid: { label: "Certificates valid", note: "Issued by you" },
      certExpired: {
        label: "Certificates expired",
        note: "Past their end date",
      },
      certRevoked: { label: "Certificates revoked", note: "Withdrawn" },
      orgsPending: {
        label: "Organisations to verify",
        note: "Waiting for review",
      },
      orgsChanges: {
        label: "Changes requested",
        note: "Waiting for the organisation",
      },
      orgsRejected: { label: "Rejected", note: "Not verified" },
    } as Record<TileKey, { label: string; note: string }>,
    actions: {
      scan: "Scan QR",
      issue: "Issue passport",
      registerFacility: "Register facility",
      handover: "Hand over",
      transform: "Transform",
      reviewFlags: "Review flags",
      issueCertificate: "Issue certificate",
      reviewOrgs: "Review organisations",
      templates: "Templates",
      users: "Users",
      reports: "Run report",
    } as Record<ActionKey, string>,
    panels: {
      incoming: {
        title: "Incoming handovers",
        sub: "Lots on their way to you that need your answer",
        empty: "Nothing on its way to you.",
      },
      passports: {
        title: "Passports you hold",
        sub: "Recently updated lots",
        empty: "You don't hold any active passports.",
      },
      capacity: {
        title: "Facility capacity",
        sub: "Current period",
        empty:
          "No facilities yet. Register a facility before issuing passports.",
      },
      certificates: {
        title: "Latest certificates",
        sub: "Issued by your organisation",
        empty: "No certificates issued yet.",
      },
      flags: {
        title: "Open flags",
        sub: "Concerns being looked into",
        empty: "No open flags.",
      },
      orgsPending: {
        title: "Waiting for verification",
        sub: "Oldest first",
        empty: "No organisations are waiting.",
      },
    } as Record<PanelKey, { title: string; sub: string; empty: string }>,
    columns: {
      lot: "Lot",
      product: "Product",
      quantity: "Quantity",
      status: "Status",
    },
    quickActions: "Quick actions",
    all: "View all",
    from: (name: string) => `From ${name}`,
    to: (name: string) => `To ${name}`,
    someOrg: "another organisation",
    stage: (stage: string) => `Stage: ${stage}`,
    awaiting: "Awaiting receiver",
    capacityOf: (used: string, max: string) => `${used} of ${max}`,
    capacityLeft: (left: string) => `${left} left this period`,
    capacityNear: "Close to the limit",
    capacityOver: "Over the declared capacity",
    waitingDays: (d: number) =>
      d === 0 ? "Today" : d === 1 ? "1 day" : `${d} days`,
    workspaces: {
      title: "Workspaces",
      sub: (n: number) =>
        n === 1
          ? "You are a member of 1 organisation. Open it to start working."
          : `You are a member of ${n} organisations. Pick one to start working.`,
      open: "Open workspace",
      noneTitle: "You don't belong to an organisation yet.",
      noneBody:
        "Ask your organisation's admin for an invitation, or create your organisation's account.",
    },
    loadFailed: "Couldn't load",
  },

  scan: {
    title: "Scan a passport",
    hint: "Point the camera at the QR label on the bag, box or document.",
    noCamera: "The camera isn't available here. Type the lot code instead.",
    codeLabel: "Lot code",
    codeHint: "Printed under the QR code, e.g. ACME-2026-0412",
    open: "Open passport",
    close: "Close",
  },
  documentTypeLabel: (type: string) =>
    documentTypes.find((d) => d.value === type)?.label ?? type,
  stepOf: (n: number, total: number) => `Step ${n} of ${total}`,
  keyLabel: (key: string) =>
    key.replace(/_/g, " ").replace(/^./, (c) => c.toUpperCase()),

  // ---- Signing (patterns.md → Signing) ----

  signing: {
    title: "Sign with your PIN",
    label: "6-digit signing PIN",
    help: "This signature is yours alone; your PIN is checked once for this action.",
    noPinTitle: "You don't have a signing PIN yet",
    noPinBody:
      "Set a 6-digit PIN in your profile, then come back to sign this record.",
    setPin: "Set a PIN in your profile",
    signing: "Signing…",
  },

  // ---- 04 · Facility registry ----

  facilities: {
    title: "Facilities",
    sub: (n: number, org: string) =>
      `${n} ${n === 1 ? "facility" : "facilities"} registered · ${org}`,
    register: "Register facility",
    filterLabel: "Type",
    filterAll: "All types",
    empty:
      "No facilities yet. A facility is where your goods originate or are handled: a plot, a site, a station or a warehouse. Register one before issuing passports.",
    emptyType: "No facilities of this type.",
    active: "Active",
    suspended: "Suspended",
    suspendedNote: "Suspended facilities can't issue passports.",
    suspend: "Suspend",
    reinstate: "Reinstate",
    suspendTitle: (name: string) => `Suspend ${name}?`,
    suspendBody:
      "It can't issue passports while suspended. Lots already issued here keep their records.",
    suspendConfirm: "Suspend facility",
    pending: "Awaiting approval",
    pendingNote:
      "The platform checks the licence and location before this facility can issue passports.",
    expired: "Licence expired",
    expiredNote:
      "The licence has ended. Upload the renewed licence to keep issuing from here.",
    approve: "Approve facility",
    mapLabel: "Map of your facilities",
    legend: {
      active: "Active",
      pending: "Awaiting approval",
      inactive: "Suspended or licence expired",
    },
    detail: {
      licence: "Licence number",
      licenceValid: "Licence valid to",
      licenceUnknown: "Not given",
      licenceDoc: "Licence document",
      licenceOpen: "Open licence",
      licenceNone: "No licence document uploaded.",
      licenceUpload: "Upload licence",
      mapLabel: (name: string) => `Map of ${name}`,
      area: (ha: number) =>
        `${new Intl.NumberFormat("en", { maximumFractionDigits: 2 }).format(ha)} ha`,
      point: "Point location",
      declared: "Declared capacity",
      declaredOf: (n: number, unit: string, period: string) =>
        `${n.toLocaleString()} ${unit} per ${period}`,
      issued: (used: number, unit: string, period: string) =>
        `${used.toLocaleString()} ${unit} issued this ${period}`,
      coordinates: "Coordinates",
      issueHere: "Issue passport here",
      delete: "Delete facility",
      deleteTitle: (name: string) => `Delete ${name}?`,
      deleteBody:
        "The facility is removed from your list. Passports already issued here keep their records and their journey.",
      deleteConfirm: "Delete facility",
      selected: "Selected facility",
    },
    new: {
      title: "Register facility",
      sub: "A place where your goods originate or are handled. You need one before issuing passports.",
      steps: ["Location", "Details", "Capacity"],
      name: "Facility name",
      nameHelp: "What your team calls it, like Plot 14 · Kigoma Hill.",
      kind: "Type",
      licence: "Licence number",
      licenceHelp: "Printed on the facility's licence.",
      capacity: "Declared capacity",
      capacityHelp:
        "The most that may originate here in one period. Capacity is checked every time a passport is issued from here.",
      unit: "Unit",
      unitHelp: "The unit your commodity template uses, like kg.",
      period: "Per period",
      location: "Location",
      modes: { pin: "Drop pin", area: "Draw boundary" },
      hint: {
        pin: "Tap the map where the facility is.",
        area: "Tap the map to add corners, at least 3. The shape closes itself.",
      },
      mapLabel: "Map: tap to place the facility",
      corners: (n: number) => `${n} ${n === 1 ? "corner" : "corners"}`,
      centre: "Centre",
      undo: "Undo",
      clear: "Clear",
      gps: "Use my GPS position",
      gpsHelp:
        "Your browser asks for permission. Put yourself inside the facility.",
      gpsFail:
        "We couldn't get your position. Allow location access and try again, or tap the map.",
      licenceValid: "Licence valid to",
      licenceDoc: "Licence document",
      licenceDocHelp:
        "A PDF or clear photo of the licence. You can add it later from the facility.",
      approvalNote:
        "The platform approves each facility before it can issue passports. Capacity is checked every time a passport is issued from here.",
      save: "Submit for approval",
      saving: "Submitting…",
      licenceFailed:
        "The facility is saved, but the licence didn't upload. Upload it from the facility.",
    },
  },

  // ---- 05 · Issue origin passport ----

  issue: {
    title: "Issue passport",
    steps: ["Facility", "Quantity", "Sign"],
    where: "Where was it produced?",
    template: "Product template",
    templateHelp:
      "Sets the unit, the stages your lot moves through and the evidence each stage needs.",
    templateEmpty:
      "No templates yet. A platform admin creates the commodity templates.",
    facility: "Facility",
    facilityEmpty:
      "No facilities yet. Register a farm plot, site or warehouse first.",
    facilityLeft: (n: number, unit: string, period: string) =>
      `${n.toLocaleString()} ${unit} left this ${period}`,
    facilityNone: "No capacity left this period.",
    facilitySuspended: "Suspended: can't issue from here",
    facilityPending: "Awaiting approval: can't issue yet",
    howMuch: "How much?",
    quantity: "Quantity",
    after: (left: number, unit: string, period: string) =>
      `You have ${left.toLocaleString()} ${unit} left this ${period}. After this passport: ${Math.max(left, 0).toLocaleString()} ${unit}.`,
    evidence: "Evidence on file",
    evidenceHelp:
      "The template's first stage asks for a valid certificate of each of these, attached to this facility.",
    evidenceNone: "This template asks for no evidence before issuing.",
    evidenceMissing: (keys: string) =>
      `Missing: ${keys}. Ask a certifier to issue these certificates for this facility, then issue the passport.`,
    evidenceOnFile: "On file",
    evidenceMissingChip: "Missing",
    review: "Check and sign",
    meaning:
      "Your signature records this harvest on ChainTrack. It can't be edited afterwards; mistakes are corrected with a new, linked record.",
    sign: "Sign and issue passport",
    facts: {
      product: "Product",
      template: "Template",
      facility: "Facility",
      quantity: "Quantity",
      evidence: "Evidence",
    },
    blockedField: (limit: number, unit: string, period: string) =>
      `More than the ${limit.toLocaleString()} ${unit} left for this facility this ${period}.`,
    blockedNotice: (limit: number, unit: string) =>
      `This passport can't be issued. Lower the quantity to ${limit.toLocaleString()} ${unit} or less. If the harvest really is larger, ask the regulator to review the declared capacity.`,
    blockedReason: "Continue is blocked until the quantity fits.",
    createdTitle: "Passport issued",
    createdBody: (quantity: string, facility: string) =>
      `${quantity} from ${facility}.`,
    printLabels: "Print labels",
    labelsNote:
      "Print one label per bag or crate. Labels print at 50 × 30 mm on standard thermal printers.",
    handOver: "Hand over",
    viewPassport: "View passport",
    another: "Issue another passport",
  },

  // ---- 06 · Dual-signature handover ----

  handover: {
    title: "Hand over a lot",
    steps: ["Lot", "Details", "Receiver", "Sign"],
    pickTitle: "Which lot is leaving?",
    pickHint: "Scan the QR label on the bags, or choose a lot you hold.",
    pickEmpty:
      "You don't hold any active passports. Issue a passport before handing a lot over.",
    isRightLot: "Is this the right lot?",
    canHandOver:
      "You hold this lot and nothing is pending on it. It can be handed over.",
    notHolder:
      "Another organisation holds this lot. Only its holder can hand it over.",
    openFlag:
      "A flag is open on this lot. It can't be handed over until the flag is resolved.",
    openHandover:
      "A handover of this lot is still waiting for the receiver. Cancel it or wait for their answer.",
    details: "What's leaving?",
    allOfIt: "All of it",
    partOfIt: "Part of it",
    quantity: "Quantity",
    holds: (n: number, unit: string) =>
      `This passport holds ${n.toLocaleString()} ${unit}.`,
    partial: (rest: string) =>
      `Sending less keeps the rest (${rest}) on this passport; the part you send gets its own lot code.`,
    stage: "Stage it goes into",
    stageHelp: "Only stages of this product that allow a handover.",
    stagePlaceholder: "Choose a stage",
    noStages:
      "This product's template has no stage that allows a handover. Ask your platform admin to add one.",
    location: "Add my location",
    locationAdded: (lat: string, lon: string) =>
      `Location added · ${lat}, ${lon}`,
    locationFailed: "Your location isn't available. You can send without it.",
    photos: "Photos",
    photosHelp: (max: number) => `Up to ${max}: bags, seal, the truck.`,
    tooManyPhotos: (max: number) => `Choose at most ${max} photos.`,
    next: "Continue",
    back: "Back",
    stayHolder:
      "You stay the holder until the receiver signs. Unanswered handovers expire in 7 days.",
    youHold: "You hold this",
    note: "Where and how",
    noteHelp:
      "A place and how the goods were handed over, like Nyamagabe depot, sealed bags on truck KTB-409.",
    receiver: "Who is receiving it?",
    receiverSearch: "Search verified organisations",
    receiverEmpty: "No verified organisations match that name.",
    receiverEmptyAll:
      "No other verified organisations yet. A receiver must be verified by the platform before it can receive lots.",
    verified: "Verified",
    review: "Check and sign",
    facts: {
      lot: "Lot",
      product: "Product",
      quantity: "Quantity",
      from: "From",
      to: "Receiver",
      where: "Where",
    },
    meaning: (receiver: string) =>
      `You confirm these goods left your custody. ${receiver} must inspect and sign before the handover is complete.`,
    sign: "Sign and send",
    sentTitle: "Sent for signature",
    sentBody: (quantity: string, receiver: string) =>
      `${quantity} is on its way to ${receiver}.`,
    stepsNext: [
      "You signed",
      "Receiver inspects and signs",
      "Custody moves to them",
    ],
    signedBySender: "Signed by sender",
    awaitingReceiver: "Awaiting receiver signature",
    rejectedByReceiver: "Rejected by receiver",
    signedByBoth: "Signed by both parties",
    cancelNote:
      "If they reject the handover, you're notified and the lot stays with you while the flag is open.",
    cancel: "Cancel handover",
    cancelTitle: "Cancel this handover?",
    cancelBody:
      "The receiver is told the handover was cancelled. You keep the lot and can hand it over again later.",
    cancelConfirm: "Cancel handover",
    notSender:
      "Only the organisation that started this handover can cancel it.",
    cancelled: "Handover cancelled.",
    expired:
      "This handover expired before the receiver answered. Start it again if the lot is still leaving.",
    another: "Hand over another lot",
  },

  receiver: {
    title: "Handovers",
    sub: "Lots arriving from other organisations, and lots you've sent.",
    tabs: { incoming: "Incoming", outgoing: "Outgoing", done: "Done" },
    emptyIncoming:
      "No handovers are waiting for you. When another organisation hands a lot to you, it appears here to inspect and sign.",
    emptyOutgoing: "Nothing is on its way from you.",
    emptyDone: "Nothing has been completed yet.",
    open: "Inspect",
    from: (name: string) => `From ${name}`,
    to: (name: string) => `To ${name}`,
    expected: "Sent",
    received: "Weight received",
    receivedHelp:
      "What you weighed on arrival. It is recorded with your signature.",
    matches: "Matches what was sent",
    within: (n: string, unit: string, pct: string, tol: string) =>
      `Difference ${n} ${unit} (${pct}%) — within the ${tol}% tolerance.`,
    outside: (n: string, unit: string, pct: string, tol: string) =>
      `Difference ${n} ${unit} (${pct}%) — outside the ${tol}% tolerance. You can't accept this; reject it or raise a flag.`,
    location: "Add my location",
    locationAdded: (lat: string, lon: string) =>
      `Location added · ${lat}, ${lon}`,
    locationFailed: "Your location isn't available. You can sign without it.",
    photos: "Inspection photos",
    photosHelp: (max: number) => `Up to ${max} photos: bags, seal, scale.`,
    inspect: "Inspect on arrival",
    checklist: {
      seal: "Seal number matches",
      count: "Count and labels match",
      noDamage: "No damage",
    },
    checklistNote:
      "Answer each point from what you can see. The sender and a regulator read your answers if you reject.",
    accept: "Accept and sign",
    meaning: (sender: string) =>
      `You confirm you received this lot from ${sender}. The handover is complete once you sign.`,
    completeTitle: "Handover complete",
    completeBody: (quantity: string) => `You now hold ${quantity}.`,
    nextTitle: "What happens next",
    next: [
      "The lot is yours",
      "It appears in your passports",
      "You can hand it over or process it",
    ],
    viewPassport: "View passport",
    reject: "Reject / raise flag",
    rejectTitle: "Reject this handover?",
    rejectChips: {
      weight_short: "Weight short",
      damaged: "Damaged",
      wrong_product: "Wrong product",
      seal_broken: "Seal broken",
      not_expected: "Not expected",
    } as Record<RejectReason, string>,
    rejectPick: "Pick at least one reason.",
    rejectStays: (sender: string) =>
      `${sender} stays the holder. They can send the lot again with the right quantity.`,
    sentQty: "Sent",
    sentOn: "Sent on",
    partOf: (code: string) =>
      `Split from ${code}: the sender keeps the rest of that lot.`,
    rejectReason: "What's wrong?",
    rejectReasonHelp:
      "Pick one, or write your own. The sender and a regulator read it.",
    rejectNote: "What you saw",
    rejectNoteHelp: "Optional detail for the sender and the regulator.",
    rejectFlag: "Also raise a flag so a regulator looks at it",
    rejectConfirm: "Reject and flag",
    rejectOnly: "Reject handover",
    rejectedTitle: "Handover rejected",
    rejectedBody: (sender: string) =>
      `The lot stays with ${sender} until it's resolved.`,
    flagNote:
      "A flag is open on this lot. It stays with its holder until a regulator clears it or confirms it as fraud.",
    backToList: "Back to handovers",
  },

  // ---- 07 · Transform / split / merge ----

  transform: {
    title: "Transform lots",
    sub: "Processing that turns lots into new lots: washing, drying, milling, concentrating.",
    stage: "Stage",
    stageHelp:
      "The stage the new lots enter. It comes from the template of the lots you select.",
    inputs: "Inputs",
    addInput: "Scan or add another lot",
    inputEmpty: "Select the lots that go into this operation.",
    noneHeld: "You don't hold any active passports to process.",
    selected: (n: number, quantity: string) =>
      `${n} ${n === 1 ? "lot" : "lots"} selected · ${quantity}`,
    useAll: "Use all of it",
    usePart: "Part: ",
    outputs: "New output lots",
    addOutput: "Add output lot",
    output: (n: number) => `Output ${n}`,
    quantity: "Quantity",
    remove: "Remove",
    meterWithin: (below: string) =>
      `Within limit · ${below} below the allowed maximum`,
    meterOver: (over: string) =>
      `Over the allowed maximum by ${over}. Lower the outputs to continue.`,
    yieldLine: (yieldPct: string, maxPct: string) =>
      `Yield ${yieldPct}% of the ${maxPct}% allowed`,
    inOut: (into: string, out: string) => `${into} in → ${out} out`,
    blocked: "Fix the mass balance to continue",
    noYield:
      "This stage has no maximum yield recorded; check the outputs by hand before signing.",
    meaning:
      "Your signature records this processing step. The input passports are closed and linked to the new ones.",
    sign: (n: number) =>
      `Sign and create ${n} ${n === 1 ? "passport" : "passports"}`,
    createdTitle: (n: number) =>
      `${n} new ${n === 1 ? "passport" : "passports"} created`,
    createdBody: (
      into: string,
      out: string,
      yieldPct: string,
      maxPct: string,
    ) => `${into} in → ${out} out · yield ${yieldPct}% of ${maxPct}% allowed`,
    inputsClosed:
      "The input passports are closed. The new ones are linked to them.",
    printLabels: "Print labels",
    handOver: "Hand over",
    again: "Process another lot",
  },

  // ---- Passport detail ----

  passportView: {
    facts: {
      product: "Product",
      quantity: "Quantity",
      remaining: "Remaining",
      holder: "Current holder",
      status: "Status",
      issued: "Issued",
      facility: "Facility",
      template: "Template",
    },
    status: {
      active: "Active",
      consumed: "Consumed",
      retired: "Retired",
      revoked: "Revoked",
    } as Record<PassportStatus, string>,
    journey: "Journey",
    journeyEmpty: "No journey events yet.",
    certificates: "Certificates",
    certificatesEmpty: "No certificates are attached to this lot.",
    madeFrom: "Made from",
    madeInto: "Made into",
    noGenealogy:
      "This lot wasn't made from other lots, and hasn't been processed into any.",
    notHolder:
      "Another organisation holds this lot. You can read its records but can't act on them.",
    openHandover:
      "A handover of this lot is waiting for the receiver. It can be handed over again once that's answered.",
    actions: {
      handOver: "Hand over",
      transform: "Use in a transform",
      retire: "Retire lot",
      publicPage: "Open public page",
    },
    retireTitle: "Retire this lot?",
    retireBody:
      "Retiring closes the lot: it can't be handed over or transformed again. Retire it when it reached its final buyer.",
    retireConfirm: "Retire lot",
    eventKinds: {
      ISSUE: "Issued",
      HANDOVER: "Handed over",
      TRANSFORM: "Processed",
      FLAG: "Flag raised",
      REVOKE: "Revoked",
      RETIRE: "Retired",
    } as Record<string, string>,
    eventFrom: (name: string) => `from ${name}`,
    eventTo: (name: string) => `to ${name}`,
    sealed: "Sealed",
    notSealed: "Not sealed yet",
  },

  // ---- 08 · Certificates ----

  certificates: {
    title: "Certificates",
    sub: "Signed documents you've issued, and those about your lots and facilities. Public verification pages show them.",
    subPlatform: "Every certificate on the platform.",
    scopes: { issued: "Issued by us", about: "About our lots" },
    all: "All",
    notAllowedTitle: "You can't open certificates.",
    notAllowedBody:
      "Ask an admin of your organisation for access to passports.",
    emptyTitle: "No certificates yet.",
    revokeSign: "Sign and revoke certificate",
    cancel: "Cancel",
    issue: "Issue certificate",
    headers: {
      type: "Certificate",
      attached: "Attached to",
      valid: "Validity",
      status: "Status",
    },
    empty:
      "No certificates issued yet. Issue one to attach a signed document to a lot or a facility.",
    emptyType: "No certificates of this type.",
    status: {
      valid: "Valid",
      expiring: "Expiring soon",
      expired: "Expired",
      revoked: "Revoked",
    },
    expiringIn: (days: number) =>
      `Expiring in ${days} ${days === 1 ? "day" : "days"}`,
    open: "Open document",
    attachedLot: (code: string) => `Lot ${code}`,
    attachedFacility: (name: string) => name,
    unknownSubject: "Attached record",
    revoke: "Revoke",
    revokeTitle: "Revoke this certificate?",
    revokeBody:
      "It shows as revoked on every public page that mentions it. The holder and their buyers can read your reason.",
    revokeReason: "Reason",
    revokeReasonHelp: "Shown to the holder and on the public page.",
    revokeReasonEmpty: "Say why the certificate is being revoked.",
    revokeConfirm: "Revoke certificate",
    new: {
      title: "Issue certificate",
      sub: "Attach a signed document to a lot or a facility. It appears on public verification pages for that record.",
      attach: "Attach to",
      attachPassport: "Passport (lot)",
      attachFacility: "Facility",
      passport: "Passport",
      passportHelp: "Choose a lot you hold.",
      passportEmpty: "You don't hold any active passports.",
      facility: "Facility",
      facilityEmpty: "No facilities yet.",
      type: "Certificate type",
      typeHelp: "Your organisation may issue the types it is accredited for.",
      typePlaceholder: "Choose a type",
      typeEmpty:
        "No certificate types yet. A platform admin accredits your organisation for the types you may issue.",
      validFrom: "Valid from",
      validTo: "Valid until",
      validityHelp: "The end date must be after the start date.",
      document: "Signed document",
      documentHelp: "PDF, PNG or JPEG · up to 10 MB",
      documentMissing: "Attach the signed document.",
      preview: "Preview",
      checks: "Checks before issuing",
      checkVerdict: (verdict: string) =>
        `Public verdict of the lot: ${verdict}`,
      checkAccredited: "Your organisation is accredited for this type",
      checkDates: "The validity window is in order",
      checkFile: "A signed document is attached",
      checkMissing: "Not ready yet",
      cancel: "Cancel",
      sign: "Sign and issue",
      meaning:
        "Your signature issues this certificate. Its document and dates can't be edited afterwards; a mistake is corrected by revoking it and issuing a new one.",
    },
  },

  // ---- 08 · Flags and investigations ----

  flags: {
    title: "Flags and investigations",
    sub: (n: number) =>
      n === 0 ? "Nothing open" : `${n} ${n === 1 ? "flag" : "flags"} open`,
    queue: "Flags",
    empty:
      "No flags. When something looks wrong with a lot — a record that changed, quantities that don't balance, a handover the receiver refused — a flag appears here.",
    emptyResolved: "Nothing resolved yet.",
    tabs: { open: "Open", resolved: "Resolved" },
    selectOne: "Choose a flag to read it.",
    raised: (when: string) => `Raised ${when}`,
    passport: "Lot",
    raisedBy: "Raised by",
    reasonLabel: "What happened",
    status: {
      open: "Under review",
      cleared: "Cleared",
      confirmed: "Confirmed as fraud",
    },
    openChip: "Under review",
    resolution: "Resolution note",
    clear: "Clear flag",
    clearTitle: "Clear this flag?",
    clearBody:
      "The lot can be handed over and processed again. Your note is shown to its holder.",
    clearNote: "Note",
    clearNoteHelp: "Why it's cleared. Required.",
    clearNoteEmpty: "Say why the flag is cleared.",
    confirmFraud: "Confirm fraud",
    fraudTitle: "Confirm fraud on this flag?",
    fraudBody: (n: number) =>
      `${n} ${n === 1 ? "passport" : "passports"} will be revoked, including everything made from them.`,
    fraudBodyNone:
      "The flagged passport will be revoked, and everything made from it.",
    fraudReason: "Finding",
    fraudReasonHelp:
      "What was confirmed. It is shown on the public pages of the affected lots.",
    fraudReasonEmpty: "Say what was confirmed.",
    fraudWarning:
      "Public verification pages for these lots will show Failed immediately. Holders and their buyers are notified.",
    fraudConfirm: (n: number) =>
      `Confirm fraud and revoke ${n} ${n === 1 ? "lot" : "lots"}`,
    resolveNotYours: "Only a regulator or the platform can resolve flags.",
    raise: "Raise a flag",
    raiseTitle: "Raise a flag on this lot",
    raiseBody:
      "The lot stays with its current holder while the flag is open. A regulator reads your reason.",
    raiseReason: "What's wrong?",
    raiseReasonHelp: "Say what you saw, so a regulator knows where to look.",
    raiseConfirm: "Raise flag",
    raisedFlag: "Flag raised.",
    source: {
      user: "A business",
      rules: "The rules engine",
      public: "A public verifier",
    } as Record<string, string>,
    assignee: "Assigned to",
    unassigned: "Nobody yet",
    assignMe: "Assign to me",
    unassign: "Unassign",
    resolveIntro:
      "Read the lot's history before deciding. Clearing frees the lot; confirming fraud revokes it and everything made from it.",
    affected: "Affected lots",
    affectedEmpty: "Nothing else was made from this lot.",
  },

  // ---- 09 · Product templates ----

  templates: {
    title: "Product templates",
    sub: "The platform's rules for one commodity: its unit, the stages a lot moves through, yields and evidence.",
    new: "New template",
    empty: "No templates yet.",
    headers: {
      name: "Template",
      commodity: "Commodity",
      unit: "Unit",
      stages: "Stages",
      updated: "Updated",
    },
    open: "Open",
    create: "New template",
    detail: {
      title: (name: string) => `Product template · ${name}`,
      untitled: "New template",
      general: "General",
      name: "Template name",
      nameHelp:
        "What this template is called, like Arabica coffee · cherry (Rwanda).",
      commodity: "Commodity",
      commodityHelp:
        "Lowercase letters, digits and underscores, like arabica_coffee_cherry.",
      unit: "Base unit",
      unitHelp: "The unit every quantity in this template uses, like kg.",
      capacityRule: "Capacity rule",
      capacityFacility: "Each facility's own declared capacity",
      capacityPerArea: "Facility area × a rate per hectare per season",
      kgPerHa: "Kilograms per hectare per season",
      originKind: "Origin facility type",
      originKindHelp: "The kind of facility a lot may be issued at.",
      stages: "Stages",
      addStage: "Add stage",
      removeStage: "Remove stage",
      stageKey: "Stage key",
      stageKeyHelp:
        "Lowercase letters, digits and underscores, like washing_station.",
      events: "Allowed events",
      maxYield: "Maximum yield (%)",
      maxYieldHelp:
        "On a transform, the outputs may be at most this share of the inputs. Leave empty for no limit.",
      evidence: "Required evidence",
      evidenceHelp:
        "A valid certificate of each key must be attached before a lot may enter this stage.",
      evidenceKey: "Evidence key",
      evidenceAdd: "Add evidence",
      preview: "Live preview",
      previewNote:
        "A lot enters at the first stage and moves through them in order.",
      save: "Save template",
      saving: "Saving…",
      create: "Create template",
      cancel: "Cancel",
      notAllowed: "You can't edit templates.",
      notAllowedBody:
        "Only the platform team edits product templates. Ask a platform admin if a template needs changing.",
      eventChip: (events: string) => events,
      ruleChip: "Rules",
      evidenceChip: "Evidence",
    },
  },

  // ---- 09 · Reports and audit ----

  reports: {
    title: "Reports and audit",
    sub: "Every action on ChainTrack is written to a log that can't be edited.",
    audit: "Audit trail",
    auditEmpty: "No events in this period.",
    filters: { from: "From", to: "To", clear: "Clear filters" },
    headers: {
      when: "When",
      event: "Event",
      organisation: "Organisation",
      by: "By",
      reference: "Reference",
    },
    event: (method: string, route: string) => `${method} ${route}`,
    accepted: "Accepted",
    refused: "Refused",
    someone: "Someone",
    signedOut: "Not signed in",
    platformOrg: "ChainTrack platform",
    footer: (shown: number, total: number) =>
      `${shown} of ${total.toLocaleString()} events · Every row is sealed; this log can't be edited`,
    exportCsv: "Export CSV",
    exporting: "Preparing the file…",
    volume: "Volume by commodity",
    volumeEmpty: "No volume recorded in this period.",
    volumeHeaders: {
      month: "Month",
      commodity: "Commodity",
      quantity: "Quantity",
    },
    verifyChain: "Verify the whole log",
    verifyChainHelp:
      "Walks every recorded action and checks each row against the one before it. Platform only.",
    verifyOk: "Every row matches the one before it. The log is intact.",
    verifyBad: (seq: number) =>
      `The log doesn't match at row ${seq}. Contact the platform team.`,
    reference: "Reference",
  },

  // ---- Scan page ----

  scanPage: {
    title: "Scan a passport",
    hint: "Point the camera at the QR label. The lot opens as soon as the code is read.",
    orType: "Or type the lot code",
  },

  meta: {
    profile: "Your profile · ChainTrack",
    users: "Users · ChainTrack",
    notifications: "Notifications · ChainTrack",
    landing: "ChainTrack · Know where it came from",
    landingDescription:
      "Digital product passports for business shipments: verified origin, dual-signed handovers and a public verification anyone can scan.",
    invite: "Join an organisation · ChainTrack",
    signin: "Sign in · ChainTrack",
    forgot: "Reset password · ChainTrack",
    reset: "New password · ChainTrack",
    lock: "Lock account · ChainTrack",
    handoff: "Switching organisation · ChainTrack",
    signup: "Create account · ChainTrack",
    onboarding: "Verification · ChainTrack",
    adminOrgs: "Verification queue · ChainTrack",
    passport: "Product passport · ChainTrack",
    facilities: "Facilities · ChainTrack",
    facilityNew: "Register facility · ChainTrack",
    issue: "Issue passport · ChainTrack",
    passportDetail: "Lot · ChainTrack",
    handover: "Hand over a lot · ChainTrack",
    handovers: "Handovers · ChainTrack",
    handoverDetail: "Inspect handover · ChainTrack",
    transform: "Transform lots · ChainTrack",
    scan: "Scan a passport · ChainTrack",
    certificates: "Certificates · ChainTrack",
    certificateNew: "Issue certificate · ChainTrack",
    flags: "Flags and investigations · ChainTrack",
    templates: "Product templates · ChainTrack",
    adminTemplate: "Product template · ChainTrack",
    reports: "Reports and audit · ChainTrack",
  },

  // The navy brand panel beside every access screen (Figma AuthFlow).
  auth: {
    eyebrow: "Trusted trade infrastructure",
    title: "Every product has a story. Make it verifiable.",
    body: "Create trusted passports, sign custody handovers and protect your supply chain.",
    foot: "Secure by design · Built for regulated trade",
  },

  signin: {
    eyebrow: "Welcome back",
    remember: "Keep me signed in",
    title: "Sign in to your organisation",
    sub: "Use the email and password for your ChainTrack account.",
    email: "Email",
    password: "Password",
    code: "Authenticator code",
    codeHelp: "The 6 digits your authenticator app shows for ChainTrack.",
    submit: "Sign in",
    submitting: "Signing in…",
    notSignedIn: "You are not signed in.",
    newHere: "New to ChainTrack?",
    createAccount: "Create your organisation's account",
    forgot: "Forgot password?",
    trust: "Trust this device for 30 days",
    trustHelp:
      "Skip the code on this device. Don't tick this on a shared computer.",
    capsLock: "Caps Lock is on",
  },

  // Drafts 2d–2h: forgot, check your email, new password, expired link, done.
  forgot: {
    back: "Back to sign in",
    title: "Reset your password",
    sub: "Enter your Email. We'll send a link that works for 30 minutes.",
    email: "Email",
    submit: "Send reset link",
    submitting: "Sending…",
    note: "Resetting your password does not change your signing PIN or two-factor setup.",
    sentTitle: "Check your email",
    sent: (email: string) =>
      `If an account exists for ${email}, a reset link is on its way. It expires in 30 minutes.`,
    nothingTitle: "Nothing arrived?",
    nothing: [
      "Check spam or quarantine folders.",
      "Use a different email.",
      "Ask your organisation admin to send an invite.",
    ],
    again: "Use a different email",
  },
  reset: {
    title: "Choose a new password",
    new: "New password",
    confirm: "Confirm password",
    rules: {
      length: `At least ${limits.password.min} characters`,
      match: "Passwords match",
    },
    signOutNote: "Every device signed in to your account will be signed out.",
    submit: "Update password",
    submitting: "Updating…",
    expiredTitle: "This link has expired",
    expired:
      "Reset links work once, for 30 minutes. Your password has not changed.",
    newLink: "Send a new link",
    doneTitle: "Password updated",
    done: "You were signed out of every device. We've emailed you a confirmation with a link to lock your account if this wasn't you.",
    toSignIn: "Continue to sign in",
  },
  handoff: {
    title: "Switching organisation…",
    body: "Opening your workspace at this address. You stay signed in.",
    failedTitle: "You're not signed in here yet",
    toSignIn: "Sign in",
  },
  lock: {
    title: "Lock your account?",
    body: "If you didn't change your password, lock the account now. Everyone signed in is signed out, and nobody can sign in until your organisation admin or ChainTrack support unlocks it.",
    submit: "Lock my account",
    submitting: "Locking…",
    doneTitle: "Your account is locked",
    done: "Nobody can sign in to it now. Contact your organisation admin or ChainTrack support to unlock it.",
    expiredTitle: "This link no longer works",
    expired:
      "Lock links work once, for 7 days. Contact your organisation admin if you still need to lock the account.",
    toSignIn: "Back to sign in",
  },

  signup: {
    eyebrow: "Create account",
    title: "Create your organisation's account",
    sub: "You'll need your business registration and any operating licences. Verification takes about 2 working days.",
    haveAccount: "Already have an account?",
    signIn: "Sign in",
    steps: ["Account", "Organisation", "Documents", "Review"],
    fullName: "Full name",
    email: "Email",
    password: "Password",
    passwordHelp: (min: number) => `At least ${min} characters.`,
    confirmPassword: "Confirm password",
    consent:
      "I am authorised to register this organisation, and I agree that its documents will be checked.",
    next: "Continue",
    back: "Back",
    legalName: "Legal name",
    legalNameHelp: "Exactly as on your registration certificate.",
    slug: "Organisation address",
    slugHelp:
      "Your team signs in at this name, in front of this site's address. Lowercase letters, digits and hyphens.",
    kind: "What does your organisation do?",
    submit: "Create account",
    submitting: "Creating account…",
    notCreated: "Your account was not created.",
    created: "Account created",
    createdSub:
      "Your organisation has its own address. Sign in there to upload your documents.",
    goSignIn: "Sign in and upload documents",
  },

  // Form rules, said the way a field error is: what to enter.
  validation: {
    pin: "Enter exactly 6 digits.",
    samePassword: "Choose a password different from your current one.",
    passwordConfirm: "Type the password again.",
    passwordsDiffer:
      "The two passwords don't match. Type the same password twice.",
    lotCode: "Enter the lot code printed under the QR code.",
    email: "Enter your Email, like name@company.com.",
    emailLong: "Use a shorter email address.",
    passwordEmpty: "Enter your password.",
    code: "Enter the 6 digits shown in your authenticator app.",
    fullName: "Enter your full name.",
    tooLong: (n: number) => `Use ${n} characters or fewer.`,
    passwordShort: (n: number) => `Use at least ${n} characters.`,
    consent: "Tick the box to confirm you may register this organisation.",
    legalName:
      "Enter the legal name, exactly as on your registration certificate.",
    slug: "Use lowercase letters, digits and hyphens, like kitabi-washing.",
    slugReserved: "That address is reserved. Choose another.",
    kind: "Choose what your organisation does.",
    fileType: "This file type can't be used. Choose a PDF, PNG or JPEG.",
    fileSize:
      "This file is larger than 10 MB. Choose a smaller file or a lower-quality scan.",
    reason: "Say what is wrong, so the organisation knows what to fix.",
    accreditationTypes: "Choose at least one certificate type.",
    date: "Enter a date.",
    dateOrder: "The end date must be after the start date.",
    quantity: "Enter how much.",
    quantityPositive: "Enter a whole number above zero.",
    note: "Say something, so the next person knows what happened.",
    stage: "Choose the event that happens.",
    handoverStage: "Choose the stage the lot goes into.",
    facilityLocation: "Drop a pin or draw the boundary on the map.",
    licenceExpired:
      "The licence must still be valid: choose a date after today.",
    facility: "Choose where this comes from.",
    template: "Choose a template.",
    facilityName: "Enter the facility's name.",
    facilityKind: "Choose what kind of facility this is.",
    licence: "Enter the licence number, as printed on the certificate.",
    capacity: "Enter the capacity, as a whole number above zero.",
    capacityUnit: "Enter the unit, like kg or bags.",
    capacityPeriod: "Choose the period the capacity is for.",
    latitude: "Enter the latitude, between -90 and 90.",
    longitude: "Enter the longitude, between -180 and 180.",
    subjectId: "Enter the lot code this certificate is for.",
    certType: "Choose what this certificate attests.",
    outcome: "Choose what happens to this flag.",
    reasonRequired: "Say why, so the record is complete.",
    templateName: "Enter the template's name.",
    commodity: "Enter the commodity, like coffee or cassava.",
    unit: "Enter the unit, like kg or bags.",
    stages: "Each template needs at least one stage.",
    events: "Choose at least one event.",
  },

  onboarding: {
    eyebrow: "Organisation verification",
    pendingTitle: "Pending verification",
    uploadTitle: "Upload your documents",
    pendingSub: (name: string) =>
      `We're checking ${name}. Most organisations are verified within 2 working days.`,
    uploadSub:
      "Clear photos or PDFs. We check them against the national registry.",
    notVerifiedYet:
      "You can't issue or receive passports until verification is done.",
    rejected: (name: string) => `${name} was not verified.`,
    changesRequested: "Changes requested.",
    noReason: "The reviewer gave no reason.",
    uploadAgain:
      "Upload a corrected document below and it goes back to the reviewer.",
    submitted: (when: string) => `Submitted ${when}`,
    progress: {
      account: "Account created",
      documents: "Documents received",
      stillNeeded: (list: string) => `Still needed: ${list}`,
      registry: "Registry check",
      registryNote:
        "We are checking your documents against the national registry.",
      approval: "Approval",
      approvalNote:
        "This page shows the outcome as soon as the review is done.",
    },
    licencesLater:
      "Facility licences are added later, when you register each farm plot or site.",
  },

  documents: {
    required: "Required",
    optional: "Optional",
    uploaded: "Uploaded",
    uploading: "Uploading…",
    addAnother: "Add another file",
    choose: "Tap to take a photo or choose a file",
    hint: "PDF, PNG or JPEG · up to 10 MB",
    unreadable:
      "This file can't be read as a PDF, PNG or JPEG. Choose another file.",
    failed: "The upload didn't finish. Nothing was saved. Try again.",
    offline:
      "The upload didn't finish. Nothing was saved. Check your connection and try again.",
    open: "Open",
  },

  orgStatus: {
    pending: "Pending verification",
    verified: "Verified",
    rejected: "Rejected",
    changes_requested: "Changes requested",
  } as Record<OrgStatus, string>,

  admin: {
    title: "Verification queue",
    sub: (count: number, oldestDays: number | undefined) =>
      [
        `${count} ${count === 1 ? "organisation" : "organisations"}`,
        oldestDays !== undefined &&
          `oldest waiting ${oldestDays} ${oldestDays === 1 ? "day" : "days"}`,
        "target under 2 working days",
      ]
        .filter(Boolean)
        .join(" · "),
    notAllowedTitle: "You can't review organisations.",
    notAllowedBody:
      "Only the platform's reviewers can open this page. Ask a platform admin if you need access.",
    tabs: {
      organisations: "Organisations",
      issuers: "Certificate issuers",
      approved: "Approved",
      rejected: "Rejected",
    },
    empty: {
      organisations: "No organisations are waiting for review.",
      issuers: "No verified certifiers or regulators yet.",
      approved: "No organisations have been approved yet.",
      rejected: "No organisations have been rejected or asked for changes.",
    },
    selectOne: "Choose an organisation from the list to review it.",
    tabCount: (label: string, n: number) => `${label} (${n})`,
    queueLabel: (label: string) => `${label}: organisations`,
    age: (days: number) =>
      days === 0 ? "Today" : `${days} ${days === 1 ? "day" : "days"}`,
    submitted: (when: string) => `Submitted ${when}`,
    facts: {
      name: "Legal name",
      slug: "Organisation address",
      kind: "Type",
      submitted: "Submitted",
      waiting: "Waiting",
      reason: "Reviewer's note",
    },
    documents: "Documents",
    noDocuments:
      "No documents uploaded yet. It can't be approved until it has at least one.",
    review: {
      title: "Decision",
      note: "Decision note",
      noteHelp:
        "Required to reject or ask for changes. The organisation sees it.",
      reject: "Reject",
      changes: "Ask for changes",
      approve: "Approve organisation",
      saving: "Saving…",
    },
    accreditation: {
      title: "Accreditation",
      sub: "Choose which certificates this organisation may issue, and for how long.",
      types: "Certificate types they may issue",
      noTypes:
        "No templates ask for any evidence yet, so there are no certificate types to choose.",
      from: "Accredited from",
      until: "Accredited until",
      notice:
        "Accreditation is shown with every certificate they issue. Ending it doesn't cancel certificates they already issued.",
      decline: "Decline",
      accredit: "Accredit issuer",
      saving: "Saving…",
      current: (n: number) =>
        n === 0
          ? "Not accredited"
          : `Accredited for ${n} ${n === 1 ? "type" : "types"}`,
    },
  },

  // Public verification page (screens.md §01).
  verification: {
    kicker: "Product passport check",
    publicLabel: "Public verification",
    resultEyebrow: "Verification result",
    checkedOn: "Checked on",
    recordsLabel: "Records checked",
    scanAnother: "Scan another",
    businessSignIn: "Business sign in",
    lotCode: "Lot code",
    copyCode: "Copy lot code",
    copied: "Lot code copied",
    tabs: { journey: "Journey", genealogy: "Genealogy" },
    genealogyTitle: "Product genealogy",
    genealogyBody: (n: number) =>
      n === 0
        ? "This lot was issued at its origin. It was not made from other lots."
        : `This lot traces back through ${n} ${n === 1 ? "earlier lot" : "earlier lots"} to its origin.`,
    currentLot: "This lot",
    originLot: "Origin lot",
    inputLot: "Earlier lot",
    via: (stage: string, used: string) => `${stage} · ${used} used`,
    integrity: "Integrity summary",
    documentation: "Documentation",
    passedOf: (n: number, total: number) => `${n} of ${total} passed`,
    originEyebrow: "Verified origin",
    originMap: (name: string) => `Map of ${name}`,
    operator: "Operator",
    capacityCheck: "Capacity check",
    help: {
      title: "How verification works",
      body: "Every time this page opens, ChainTrack compares each signed business record with the copy sealed when it was signed. A signed record can't be edited afterwards.",
    },
    tagline: "Transparent journeys. Trusted goods.",
    route: {
      label:
        "Drawing of the lot's route, from its origin to its current holder",
      origin: "Origin",
      now: "Now with",
    },
    passportStatus: {
      active: "Active passport",
      consumed: "Used in a new lot",
      retired: "Journey complete",
      revoked: "Revoked",
    } as Record<string, string>,
    checking: "Checking this lot's records…",
    notFoundTitle: "We can't find this lot code",
    notFoundBody:
      "Check the code on the label and scan it again. If the label is damaged, ask the holder for the lot code.",
    errorTitle: "We couldn't check this lot just now",
    errorBody:
      "The records could not be reached. Nothing about the lot has changed. Try again in a moment.",
    verdict: { verified: "Verified", warning: "Warning", failed: "Failed" },
    alert: {
      lead: "Do not accept, pay for or ship this lot.",
      rest: "One or more of its records failed verification. The reasons are listed below.",
    },
    sections: {
      origin: "Origin",
      journey: "Journey",
      madeFrom: "Made from",
      certificates: "Certificates",
      checks: "What we checked",
      technical: "Technical details",
    },
    facts: {
      facility: "Facility",
      type: "Type",
      operatedBy: "Operated by",
      commodity: "Commodity",
      quantity: "Quantity",
      holder: "Current holder",
      status: "Status",
      updated: "Last updated",
    },
    steps: (n: number) => `${n} ${n === 1 ? "step" : "steps"}`,
    lotsToOrigin: (n: number) =>
      `${n} ${n === 1 ? "lot" : "lots"}, back to origin`,
    technicalIntro:
      "When a record is signed, its fingerprint is sealed on a public ledger (a blockchain), where it cannot be changed. We compare each record with its sealed fingerprint every time this page is opened.",
    noCertificates: "No certificates are attached to this lot or its origin.",
    notSealed: "Not sealed yet",
    signedByBoth: "Signed by both parties",
    originMark: "Origin",
    revoked: "Revoked",
    unknownOrg: "Unknown organisation",
    footer:
      "ChainTrack product passport · The page reflects records sealed up to the time of your scan.",
    certificate: {
      valid: "Valid",
      expiring: "Expiring soon",
      expired: "Expired",
      revoked: "Revoked",
    },
    validTo: "to",
    result: { verified: "Passed", warning: "Warning", failed: "Failed" },
    // The seven checks, as the API names them: the rule shown, and what a pass means.
    checks: {
      record_intact: {
        label: "Records unchanged since signing",
        pass: "Every passport in this lot's history reads exactly as it did when it was signed.",
        reason: "Records unchanged since signing",
      },
      anchored: {
        label: "Records match their sealed copies",
        pass: "Each record, its holder and its quantity match the copy sealed when it was signed.",
        reason: "Records match their sealed copies",
      },
      origin: {
        label: "Origin is registered and licensed",
        pass: "Every branch of this lot's history ends at a registered facility owned by a verified organisation.",
        reason: "Origin facility verified",
      },
      capacity: {
        label: "Within declared capacity",
        pass: "No more was issued from the origin than its declared capacity for the period.",
        reason: "Origin within declared capacity",
      },
      mass_balance: {
        label: "Quantities balance",
        pass: "At every processing step, what came out is no more than the inputs could yield.",
        reason: "Quantities balance at every step",
      },
      certificates: {
        label: "Certificates valid",
        pass: "Every certificate is in date, unrevoked, from an accredited issuer, and unchanged since it was issued.",
        reason: "Certificates valid",
      },
      status: {
        label: "No open concerns or revocations",
        pass: "Nothing in this lot's history has been revoked or is being looked into.",
        reason: "No open concerns",
      },
    } as Record<string, { label: string; pass: string; reason: string }>,
    eventTitles: {
      ISSUE: "Issued at origin",
      HANDOVER: "Handed over",
      TRANSFORM: "Processed",
      FLAG: "Concern raised",
      REVOKE: "Revoked",
    } as Record<string, string>,
    summary: {
      from: (product: string, places: string) =>
        `This ${product} comes from ${places}.`,
      verified:
        "Every record in its history matches the copy sealed when it was signed, and nothing has changed since.",
      warning: (points: string) =>
        `Its records are intact, but ${points} attention. Read them before you rely on this passport.`,
      pointsNeed: (n: number) =>
        `${n} ${n === 1 ? "point needs" : "points need"}`,
      failed:
        "This passport failed verification. Do not rely on it until the points below are resolved.",
    },
    meta: {
      checked: (when: string) => `Checked ${when}`,
      records: (n: number) => `${n} ${n === 1 ? "record" : "records"}`,
      organisations: (n: number) =>
        `${n} ${n === 1 ? "organisation" : "organisations"}`,
    },
    issued: (when: string) => `Passport issued ${when}`,
    certificateNote: {
      revoked: (reason: string | null) =>
        `Revoked by its issuer${reason ? `: ${reason}` : "."} Ask the holder for a replacement.`,
      expired: (when: string) =>
        `Expired ${when}. Ask the holder for a current one before you rely on it.`,
    },
    technical: {
      fingerprint: "Passport record · fingerprint",
      sealedIn: "Passport record · sealed in transaction",
      eventSealed: (title: string, when: string) =>
        `${title} · ${when} · sealed in transaction`,
      certificateFingerprint: (type: string) => `${type} · fingerprint`,
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
  nav: {
    how: "How it works",
    solutions: "Solutions",
    trust: "Trust & compliance",
    signIn: "Sign in",
    getStarted: "Get started",
  },
  anchors: {
    how: "how",
    solutions: "solutions",
    trust: "trust",
    security: "security",
    verify: "verify",
  },
  hero: {
    kicker: "Trusted product journeys, from origin to buyer",
    title: "Know where it came from.",
    titleEm: "Prove every step.",
    body: "ChainTrack gives every commodity lot a trusted digital passport, so businesses can hand over custody with confidence and anyone can verify the full journey in seconds.",
    primary: "Create your first passport",
    secondary: "Verify a product",
    proof: [
      "Verified organisations",
      "Dual-signed handovers",
      "Tamper-evident history",
    ],
    photoAlt: "A farmer harvesting crops in a green field",
    example: "Example",
    originLabel: "Verified origin",
    originPlace: "Nyamagabe, Rwanda",
    active: "Active",
    passportLabel: "Digital product passport",
    passportName: "Example coffee lot",
    passportCode: "EXAMPLE-2026-0001",
    stops: ["Origin", "Process", "Export", "Buyer"],
    verified: "Verified",
    checks: "4 of 4 checks passed",
    signedTitle: "Custody accepted",
    signedSub: "Both parties signed",
  },
  trustStrip: {
    label: "Built for trusted trade in",
    items: ["Coffee", "Minerals", "Medicines", "Agriculture", "Logistics"],
  },
  verify: {
    title: "Have a ChainTrack label?",
    body: "Scan the QR code, or type the lot code printed under it, to see its verified journey.",
    label: "Lot code",
    placeholder: "Lot code, e.g. ACME-2026-0412",
    submit: "Verify now",
  },
  how: {
    eyebrow: "One trusted record",
    title: "From first mile to final buyer",
    body: "Every participant adds to the same product story, without changing what came before.",
    steps: [
      {
        title: "Issue at origin",
        body: "A verified facility creates the passport, records the quantity and prints a QR label.",
        icon: "map",
      },
      {
        title: "Sign every handover",
        body: "Sender and receiver both confirm the quantity, place and condition.",
        icon: "route",
      },
      {
        title: "Transform with control",
        body: "Capacity and yield rules keep unexplained goods out of the chain.",
        icon: "layers",
      },
      {
        title: "Verify anywhere",
        body: "One scan shows origin, journey, certificates and any open concerns.",
        icon: "shield",
      },
    ],
  },
  roles: {
    eyebrow: "A clear first step for every role",
    title: "See how ChainTrack fits your work",
    body: "Choose your role for a short introduction. The workspace only shows what your organisation is allowed to do.",
    tabs: {
      producer: "Producer",
      logistics: "Logistics & trade",
      certifier: "Certifier",
      regulator: "Regulator",
    },
    cta: "Set up this workspace",
    tipTitle: "Checked before you sign",
    tipBody:
      "Capacity, licence, evidence and quantity checks run before you can sign. When something is wrong, ChainTrack says what to fix.",
    guides: {
      producer: {
        title: "Start at the verified origin",
        body: "Register your licensed facility, declare its capacity for the period and issue a passport when a real lot is ready.",
        steps: [
          [
            "Register your facility",
            "Add the licensed location and its capacity.",
          ],
          [
            "Issue a passport",
            "Record the quantity and evidence, then print the QR label.",
          ],
          [
            "Hand over securely",
            "Choose a verified receiver and sign the transfer.",
          ],
        ],
      },
      logistics: {
        title: "Move goods with shared accountability",
        body: "Scan before collection, record the quantity and condition, then ask the receiver to weigh, inspect and countersign.",
        steps: [
          ["Scan the label", "Confirm the lot before it enters your custody."],
          [
            "Record the handover",
            "Add the quantity, place and evidence photos.",
          ],
          [
            "Get both signatures",
            "The holder changes only once the receiver signs.",
          ],
        ],
      },
      certifier: {
        title: "Attach evidence buyers can trust",
        body: "Find the right lot or facility, choose a certificate type you're accredited for and issue it with its validity dates.",
        steps: [
          ["Choose the subject", "Find the exact lot or licensed facility."],
          ["Upload evidence", "Add the validity dates and the signed file."],
          ["Issue or revoke", "Publish valid evidence, or revoke it for good."],
        ],
      },
      regulator: {
        title: "Investigate risk across the full journey",
        body: "Work through open flags, raised by businesses or by the rules, compare the signed records and see every lot made from a flagged one.",
        steps: [
          ["Review the queue", "Start with open flags and who raised them."],
          [
            "Inspect evidence",
            "Compare quantities, documents and affected lots.",
          ],
          [
            "Record a decision",
            "Clear the concern, or confirm it with a reason.",
          ],
        ],
      },
    } satisfies Record<
      MarketingRole,
      { title: string; body: string; steps: [string, string][] }
    >,
  },
  showcase: {
    eyebrow: "For every participant",
    title: "One view of the truth.",
    titleLine: "Built for real operations.",
    body: "From farms and mines to ports, processors and regulators, ChainTrack makes trusted records part of the work, not extra work.",
    points: [
      [
        "Businesses",
        "Issue passports, manage handovers and keep compliant records.",
      ],
      [
        "Certifiers",
        "Attach valid, revocable evidence to the right lot or facility.",
      ],
      [
        "Regulators",
        "Investigate exceptions and trace every affected product.",
      ],
      [
        "Buyers & consumers",
        "Scan once and see what was independently checked.",
      ],
    ],
    cta: "Open the business workspace",
    window: "ChainTrack verification",
    result: "Verification result",
    verdict: "Verified",
    verdictSub: "Origin and journey confirmed",
    timeline: [
      "Origin facility verified",
      "Washed and processed",
      "Custody accepted",
      "Export cleared",
    ],
    signed: "Signed",
  },
  trust: {
    eyebrow: "Trust by design",
    title: "Evidence you can act on.",
    body: "ChainTrack checks records, quantities, facility capacity and certificate validity, then gives a clear verdict with its reasons.",
    cta: "Verify a product",
    verdicts: [
      {
        tone: "verified",
        label: "Verified",
        body: "Every required check passed",
      },
      { tone: "warning", label: "Warning", body: "Review before accepting" },
      { tone: "danger", label: "Failed", body: "Do not accept this product" },
    ],
    note: "A clear label, icon and explanation",
  },
  security: {
    eyebrow: "Security & privacy",
    title: "Protected at every handover.",
    titleLine: "Private by default.",
    body: "Built for regulated trade: commercial details stay private, every important action is signed with the person's own PIN, and recorded history can't be changed without it showing.",
    assuranceTitle: "No personal or price data is published",
    assuranceBody:
      "Public verification shows where a product came from and whether it checks out, not contracts, people's details or prices.",
    cards: [
      {
        icon: "shield",
        title: "Strong account protection",
        body: "Two-factor sign-in, signing PINs and recovery codes protect every user.",
      },
      {
        icon: "link",
        title: "Two-party confirmation",
        body: "Custody changes only after sender and receiver both sign the same details.",
      },
      {
        icon: "file",
        title: "Change detection",
        body: "Recorded fingerprints expose altered records, quantities or certificates.",
      },
      {
        icon: "layers",
        title: "Least-privilege access",
        body: "Roles decide who may issue, sign, certify, investigate or administer.",
      },
      {
        icon: "clock",
        title: "Complete audit trail",
        body: "Every change records who did it, for which organisation, and when.",
      },
      {
        icon: "check",
        title: "Rules checked before signing",
        body: "Capacity, yield, evidence and licence rules are checked before anything is accepted.",
      },
    ],
  },
  faq: {
    eyebrow: "Questions",
    title: "Questions before you begin?",
    body: "Short answers to what new teams ask most.",
    items: [
      [
        "Do people need an account to verify a product?",
        "No. Anyone can scan a ChainTrack QR label and see the verification without signing in.",
      ],
      [
        "Who can create a product passport?",
        "Only verified organisations, from a licensed facility the platform has approved.",
      ],
      [
        "Can a handover happen without the receiver?",
        "No. The sender starts it, but custody moves only once the verified receiver inspects and signs.",
      ],
      [
        "What if a certificate is no longer valid?",
        "The verification shows expired or revoked evidence and says whether the product needs a closer look.",
      ],
      [
        "Can we add our own commodity?",
        "Yes. A platform admin adds a commodity template (unit, stages, yield limits, required evidence) with no change to the software.",
      ],
      [
        "What does a buyer see?",
        "Origin, journey, certificates and a verdict with its reasons. Never prices, contracts or personal details.",
      ],
    ],
  },
  cta: {
    title: "Make every journey accountable.",
    body: "Start with one facility, one product and one trusted passport.",
    primary: "Get started with ChainTrack",
    secondary: "Verify a product",
  },
  footer: {
    tagline: "Transparent journeys. Trusted goods.",
    product: "Product",
    verify: "Verify a passport",
    workspace: "Business workspace",
    how: "How it works",
    trustTitle: "Trust",
    security: "Security",
    guide: "How verification works",
    rights: (year: number) => `© ${year} ChainTrack`,
  },
} as const;

// ---- Platform dashboard (/dashboard for the platform; GET /stats) ----

export const platformCopy = {
  title: "Platform overview",
  updated: (when: string) => `Updated ${when}`,
  loadFailed: "The platform figures didn't load.",
  tiles: {
    organisations: {
      label: "Organisations",
      note: (pending: number) =>
        `${pending.toLocaleString()} waiting for review`,
    },
    users: { label: "People", note: "Accounts on the platform" },
    passports: {
      label: "Passports",
      note: (active: number) => `${active.toLocaleString()} active`,
    },
    flags: { label: "Open flags", note: "Waiting for an investigator" },
  },
  activity: { title: "Activity, last 30 days", sub: "Per day, UTC" },
  series: {
    passports: "Passports issued",
    handovers: "Handovers completed",
    transforms: "Transforms",
    flags: "Flags raised",
    signups: "Sign-ups",
  },
  volume: {
    title: "Lots issued at origin",
    sub: "By commodity, all time (units differ, so lots are compared)",
    lots: "Lots",
  },
  breakdowns: {
    organisations: "Organisations by status",
    facilities: "Facilities by status",
    certificates: "Certificates by status",
    passports: "Passports by status",
  },
  topOrgs: {
    title: "Top organisations by volume",
    name: "Organisation",
    quantity: "Quantity",
    lots: "Lots",
    open: "Open organisation",
  },
  outbox: {
    title: "Ledger sealing",
    pending: "Waiting to be sealed",
    failed: "Failed",
    oldest: (minutes: number) =>
      `Oldest waiting ${minutes.toLocaleString()} min`,
    clear: "Everything is sealed.",
  },
  value: "Count",
  none: "Nothing yet.",
};

// ---- Reports and audit (/admin/reports) ----

/** Download formats the report endpoints take (?format=). */
export const exportFormats = ["csv", "pdf"] as const;

export const reportsCopy = {
  columns: {
    month: "Month",
    commodity: "Commodity",
    unit: "Unit",
    quantity: "Quantity",
    seq: "Seq",
    created: "When",
    actor: "Who",
    action: "Action",
    subject: "Subject",
    status: "Status",
    hash: "Fingerprint",
  },
  title: "Reports and audit",
  sub: "Volume, audit trail and chain verification.",
  notAllowedTitle: "You can't view reports.",
  notAllowedBody: "You don't have permission to view reports or audit logs.",
  volume: "Volume issued",
  outcomes: "Verification outcomes",
  saved: "Saved reports",
  save: "Save report",
  run: "Run and keep in bucket",
  savedName: "Report name",
  savedKind: "Report",
  files: "Files",
  startDate: "Start date",
  endDate: "End date",
  downloadCsv: "Download CSV",
  downloadPdf: "Download PDF",
  month: "Month",
  commodity: "Commodity",
  unit: "Unit",
  quantity: "Quantity",
  noData: "No data",
  auditTrail: "Audit trail",
  route: "Route",
  allRoutes: "All routes",
  seq: "Seq",
  created: "When",
  actor: "Actor",
  methodRoute: "Method / Route",
  subject: "Subject",
  status: "Status",
  hash: "Hash",
  chainCheck: "Chain check",
  runCheck: "Run chain check",
  running: "Running…",
  verified: "Every entry's hash matches",
  checked: (n: number) => `${n} entries checked`,
  brokenAt: (n: number | null) => (n ? `Broken at seq ${n}` : "Broken"),
};

// ---- Screen copy that grew inside the screens; kept per screen ----

export const screenCopy = {
  templateDetail: {
    back: "Back to templates",
    title: "Product template",
    sub: "Edit the template and its stages.",
    notAllowedTitle: "You can't manage templates.",
    notAllowedBody: "Only the platform admin can manage product templates.",
    name: "Name",
    commodity: "Commodity",
    unit: "Unit",
    originFacilityKind: "Origin facility kind",
    any: "Any",
    capacityRule: "Capacity rule",
    none: "None",
    perArea: "Per area (yield per hectare per season)",
    facilityCapacity: "Facility capacity",
    kgPerHaPerSeason: "Kg per ha per season",
    stages: "Stages",
    addStage: "Add stage",
    stageKey: "Stage key",
    events: "Events",
    maxYield: "Max yield (%)",
    requiredEvidence: "Required evidence (comma separated)",
    removeStage: "Remove",
    save: "Save",
    saving: "Saving…",
    saved: "Template saved",
    createdAt: "Created",
    updatedAt: "Updated",
    id: "ID",
  },
  templates: {
    title: "Product templates",
    sub: "Define how a commodity moves: stages, events, evidence and capacity rule.",
    new: "New template",
    name: "Name",
    commodity: "Commodity",
    unit: "Unit",
    originFacilityKind: "Origin facility kind",
    any: "Any",
    capacityRule: "Capacity rule",
    none: "None",
    perArea: "Per area (yield per hectare per season)",
    facilityCapacity: "Facility capacity",
    kgPerHaPerSeason: "Kg per ha per season",
    stages: "Stages",
    addStage: "Add stage",
    stageKey: "Stage key",
    events: "Events",
    maxYield: "Max yield (%)",
    requiredEvidence: "Required evidence (comma separated)",
    removeStage: "Remove",
    atLeastOneEvent: "At least one event",
    listName: "Name",
    listCommodity: "Commodity",
    listUnit: "Unit",
    listOrigin: "Origin facility",
    listStageCount: "Stage count",
    listUpdated: "Updated",
    noTemplates: "No templates yet.",
    notAllowedTitle: "You can't manage templates.",
    notAllowedBody: "Only the platform admin can manage product templates.",
    created: "Template created",
    open: "Open template",
  },
  issue: {
    back: "Back",
    next: "Continue",
    whatNext: "What happens next",
    notAllowedTitle: "You can't issue passports.",
    notAllowedBody:
      "Ask an admin of your organisation for the issue-passports permission.",
    afterPassport: (used: number, unit: string) =>
      `With this passport: ${used.toLocaleString()} ${unit} issued.`,
    evidenceCount: (have: number, total: number) =>
      `${have} of ${total} on file`,
    evidenceRowMissing: "Not on file yet.",
    nextSteps: [
      "Print a QR label for each bag or crate.",
      "Collectors and inspectors add their records against this lot.",
      "Hand the lot over when it leaves this site.",
    ],
  },
  facilities: {
    cancel: "Cancel",
    closeDetail: "Close facility details",
    emptyTitle: "No facilities yet",
    selectPrompt: "Choose a facility to see its details.",
    noPoint: "No point recorded yet.",
    areaLabel: "Area",
    registeredLabel: "Registered",
    issuedLabel: "Issued",
  },
  scan: {
    check: "Check this lot",
    lotCode: "Lot code",
    created: "Created",
    openRecord: "Open the full record",
    again: "Scan another",
    emptyTitle: "No lot checked yet",
    emptyBody:
      "Scan a QR label or type a lot code above. The lot's records are checked as soon as the code is read.",
  },
  passportDetail: {
    held: (held: number, total: number, unit: string) =>
      `${held.toLocaleString()} of ${total.toLocaleString()} ${unit}`,
    stagesLabel: "Stages",
    upstream: "Upstream",
    downstream: "Downstream",
    flagsTitle: "Flags on this lot",
    flagsEmpty: "No flags on this lot.",
    resolution: "Resolution note",
    edge: (stage: string, used: string) => `${stage} · ${used} used`, // stage already humanized
    technical: {
      fingerprint: "Record fingerprint",
      sealed: "Sealed in transaction",
      template: "Template",
      facility: "Facility",
      created: "Created",
      updated: "Updated",
    },
    retireSign: "Sign and retire",
    retireMeaning:
      "Retiring closes this lot for good. It can't be handed over or transformed again, and this record can't be undone.",
    cancel: "Cancel",
    notFoundTitle: "Lot not found",
    notFoundBody:
      "Check the lot code, or go back to the passports your organisation holds.",
  },
  transform: {
    notAllowedTitle: "You can't transform lots.",
    notAllowedBody:
      "Turning lots into new lots is a processing action. Ask your organisation's admin for the transform permission.",
    addLot: "Add lot",
    massBalance: "Mass balance",
    pickStage: "Choose a stage to check the outputs against its allowed yield.",
    templateLine: (name: string) => `Template: ${name}`,
    noTemplate:
      "This lot has no product template, so its stages can't be listed.",
    noYieldShort: "No maximum yield recorded",
    allowedYield: (max: number) => `Allowed yield up to ${max}%`,
    totalTooHigh: "Total too high",
    overRemaining: (max: number, unit: string) =>
      `More than the ${max.toLocaleString()} ${unit} this lot has left. Enter ${max.toLocaleString()} ${unit} or less.`,
    alreadyAdded: "That lot is already in the inputs.",
    tooManyRows: (n: number) =>
      `A transform takes up to ${n} rows on each side.`,
    readyBlocked:
      "Fill in every input, the stage and the outputs before signing.",
    yieldLabel: "Yield",
    newLots: "New lots",
    overSaid: (out: string, max: string) =>
      `Outputs total ${out}; these inputs can yield at most ${max} at this stage.`,
    formula: (into: string, pct: string, max: string) =>
      `Input ${into} × allowed yield ${pct}% = ${max} maximum output`,
  },
};

// ---- Confirmations (components/chaintrack/confirm.tsx) ----

export const confirmCopy = {
  cancel: "Cancel",
  signOut: {
    title: "Sign out of ChainTrack?",
    body: "You'll need your password to sign in again on this device.",
    label: "Sign out",
  },
  review: {
    verified: {
      title: "Approve this organisation?",
      body: "It can issue, hand over and receive passports straight away, and its signing key is created. The decision and your note are recorded.",
      label: "Approve organisation",
    },
    changes_requested: {
      title: "Ask for changes?",
      body: "The organisation sees your note and can upload new documents. It stays unable to trade until it is approved.",
      label: "Ask for changes",
    },
    rejected: {
      title: "Reject this organisation?",
      body: "The organisation sees your note. It can't trade; it may upload new documents to be reviewed again.",
      label: "Reject organisation",
    },
  },
  accredit: {
    title: "Save these accreditations?",
    body: "The organisation may issue exactly the certificate types ticked, within the dates given. The ledger records the change.",
    label: "Save accreditations",
  },
  decline: {
    title: "Remove every accreditation?",
    body: "The organisation can no longer issue certificates. Certificates it already issued stay valid.",
    label: "Remove accreditations",
  },
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
  userRole: (name: string, role: string) => ({
    title: `Make ${name} ${role}?`,
    body: "Their permissions here change at once.",
    label: "Change role",
  }),
  flagAssign: {
    title: "Take this flag?",
    body: "You become its investigator; others see it's yours.",
    label: "Assign to me",
  },
  flagUnassign: {
    title: "Let go of this flag?",
    body: "It goes back to the queue without an investigator.",
    label: "Unassign",
  },
};
