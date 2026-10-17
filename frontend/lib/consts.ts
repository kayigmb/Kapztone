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
export const reviewDecisions = [orgStatus.rejected, orgStatus.changesRequested, orgStatus.verified] as const;
export type ReviewDecision = (typeof reviewDecisions)[number];
/** Decisions that send the organisation back to upload again; each needs a reason. */
export const reopeningDecisions: readonly ReviewDecision[] = [orgStatus.rejected, orgStatus.changesRequested];

export const orgKinds = [
  { value: "producer", label: "Producer", sub: "Farms, cooperatives, mines" },
  { value: "processor", label: "Processor", sub: "Washing stations, smelters, factories" },
  { value: "transporter", label: "Logistics", sub: "Transporters, warehouses, forwarders" },
  { value: "exporter", label: "Exporter", sub: "Ships goods out of the country" },
  { value: "importer", label: "Importer", sub: "Receives goods from abroad" },
  { value: "certifier", label: "Certifier", sub: "Issues certificates; needs accreditation" },
  { value: "regulator", label: "Regulator", sub: "Oversees a sector; needs accreditation" },
] as const;
/** Kinds of organisation that can be accredited to issue certificates. */
export const issuerKinds: readonly string[] = ["certifier", "regulator"];

export const documentTypes = [
  { value: "registration", label: "Certificate of registration", required: true },
  { value: "licence", label: "Operating licence", required: true },
  { value: "other", label: "Other supporting document", required: false },
] as const;

/** The admin verification queue's tabs: which organisations each one lists. */
export const adminTabs = [
  { value: "organisations", statuses: [orgStatus.pending] },
  { value: "issuers", statuses: [orgStatus.verified], kinds: issuerKinds },
  { value: "approved", statuses: [orgStatus.verified] },
  { value: "rejected", statuses: [orgStatus.rejected, orgStatus.changesRequested] },
] as const satisfies readonly { value: string; statuses: readonly OrgStatus[]; kinds?: readonly string[] }[];
export type AdminTab = (typeof adminTabs)[number]["value"];

export const permissions = {
  orgVerify: "org:verify",
} as const;

// Documents (backend/internal/upload): PDF, PNG or JPEG, 10 MB at most.
export const upload = {
  maxBytes: 10 * 1024 * 1024,
  types: ["application/pdf", "image/png", "image/jpeg"],
  imageTypes: ["image/png", "image/jpeg"],
};

export const limits = {
  name: 200,
  email: 254,
  password: { min: 12, max: 72 },
  // An organisation's address is a hostname label; these names are taken by the system.
  slugPattern: /^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$/,
  reservedSlugs: ["www", "api", "admin", "app"],
  codePattern: /^(\d{6})?$/,
  // ponytail: lists ask for one full page (the API's maximum); paging UI when a queue outgrows it.
  pageSize: 100,
};

// ---- Where things live ----

export const routes = {
  home: "/",
  signin: "/signin",
  signup: "/signup",
  onboarding: "/onboarding",
  adminOrgs: "/admin/organisations",
  adminOrg: (tab: string, id?: string) => `/admin/organisations?tab=${tab}${id ? `&id=${id}` : ""}`,
  passport: (id: string) => `/p/${id}`,
};

/** The browser's way to the API: the pass-through in app/api/[...path]/route.ts. */
export const apiBase = "/api";

export const apiPaths = {
  me: "/me",
  login: "/auth/login",
  logout: "/auth/logout",
  orgs: "/orgs",
  org: (id: string) => `/orgs/${id}`,
  orgDocuments: (id: string) => `/orgs/${id}/documents`,
  orgReview: (id: string) => `/orgs/${id}/verify`,
  orgAccreditations: (id: string) => `/orgs/${id}/accreditations`,
  documentFile: (id: string) => `/documents/${id}/file`,
  templates: "/templates",
  verify: (id: string) => `/verify/${encodeURIComponent(id)}`,
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
};

// ---- Copy ----

export const copy = {
  brand: "ChainTrack",
  loading: "Loading…",

  // What went wrong, and what to do next (content.md → Errors).
  errors: {
    unreachable: "We couldn't reach ChainTrack just now. Nothing was changed. Try again in a moment.",
    unexpected: "Something went wrong on our side. Nothing was changed. Try again in a moment.",
    badAnswer: "ChainTrack answered in a way this page doesn't understand. Reload the page.",
    tryAgain: "Try again",
    loadTitle: "We couldn't load this page",
  },

  signOut: "Sign out",
  documentTypeLabel: (type: string) => documentTypes.find((d) => d.value === type)?.label ?? type,
  stepOf: (n: number, total: number) => `Step ${n} of ${total}`,

  meta: {
    signin: "Sign in · ChainTrack",
    signup: "Create account · ChainTrack",
    onboarding: "Verification · ChainTrack",
    adminOrgs: "Verification queue · ChainTrack",
    passport: "Product passport · ChainTrack",
  },

  signin: {
    title: "Sign in to your organisation",
    sub: "Use the email and password for your ChainTrack account.",
    email: "Work email",
    password: "Password",
    code: "Authenticator code",
    codeHelp: "The 6 digits your authenticator app shows for ChainTrack.",
    submit: "Sign in",
    submitting: "Signing in…",
    notSignedIn: "You are not signed in.",
    newHere: "New to ChainTrack?",
    createAccount: "Create your organisation's account",
    problems: {
      credentials: "Email or password is incorrect. Check both and try again.",
      code: "That code is incorrect or already used. Enter the code your app shows now.",
      tooMany: "Too many attempts. Wait 15 minutes, then try again.",
      unreachable: "We couldn't sign you in just now. Nothing was changed. Try again in a moment.",
    },
  },

  signup: {
    title: "Create your organisation's account",
    sub: "You'll need your business registration and any operating licences. Verification takes about 2 working days.",
    haveAccount: "Already have an account?",
    signIn: "Sign in",
    steps: ["Account", "Organisation", "Documents", "Review"],
    fullName: "Full name",
    email: "Work email",
    password: "Password",
    passwordHelp: (min: number) => `At least ${min} characters.`,
    consent: "I am authorised to register this organisation, and I agree that its documents will be checked.",
    next: "Continue",
    back: "Back",
    legalName: "Legal name",
    legalNameHelp: "Exactly as on your registration certificate.",
    slug: "Organisation address",
    slugHelp: "Your team signs in at this name, in front of this site's address. Lowercase letters, digits and hyphens.",
    kind: "What does your organisation do?",
    submit: "Create account",
    submitting: "Creating account…",
    notCreated: "Your account was not created.",
    created: "Account created",
    createdSub: "Your organisation has its own address. Sign in there to upload your documents.",
    goSignIn: "Sign in and upload documents",
    slugTaken: "Another organisation already uses this address. Choose a different one.",
    emailTaken: "An account with this email already exists. Sign in instead, or use another email.",
    problems: {
      tooMany: "Too many sign-ups from this connection. Wait an hour, then try again.",
      unreachable: "We couldn't create the account just now. Nothing was saved. Try again in a moment.",
    },
  },

  // Form rules, said the way a field error is: what to enter.
  validation: {
    email: "Enter your work email, like name@company.com.",
    emailLong: "Use a shorter email address.",
    passwordEmpty: "Enter your password.",
    code: "Enter the 6 digits shown in your authenticator app.",
    fullName: "Enter your full name.",
    tooLong: (n: number) => `Use ${n} characters or fewer.`,
    passwordShort: (n: number) => `Use at least ${n} characters.`,
    consent: "Tick the box to confirm you may register this organisation.",
    legalName: "Enter the legal name, exactly as on your registration certificate.",
    slug: "Use lowercase letters, digits and hyphens, like kitabi-washing.",
    slugReserved: "That address is reserved. Choose another.",
    kind: "Choose what your organisation does.",
    fileType: "This file type can't be used. Choose a PDF, PNG or JPEG.",
    fileSize: "This file is larger than 10 MB. Choose a smaller file or a lower-quality scan.",
    reason: "Say what is wrong, so the organisation knows what to fix.",
    accreditationTypes: "Choose at least one certificate type.",
    date: "Enter a date.",
    dateOrder: "The end date must be after the start date.",
  },

  onboarding: {
    pendingTitle: "Pending verification",
    uploadTitle: "Upload your documents",
    pendingSub: (name: string) => `We're checking ${name}. Most organisations are verified within 2 working days.`,
    uploadSub: "Clear photos or PDFs. We check them against the national registry.",
    notVerifiedYet: "You can't issue or receive passports until verification is done.",
    rejected: (name: string) => `${name} was not verified.`,
    changesRequested: "Changes requested.",
    noReason: "The reviewer gave no reason.",
    uploadAgain: "Upload a corrected document below and it goes back to the reviewer.",
    submitted: (when: string) => `Submitted ${when}`,
    progress: {
      account: "Account created",
      documents: "Documents received",
      stillNeeded: (list: string) => `Still needed: ${list}`,
      registry: "Registry check",
      registryNote: "We are checking your documents against the national registry.",
      approval: "Approval",
      approvalNote: "This page shows the outcome as soon as the review is done.",
    },
    licencesLater: "Facility licences are added later, when you register each farm plot or site.",
  },

  documents: {
    required: "Required",
    optional: "Optional",
    uploaded: "Uploaded",
    uploading: "Uploading…",
    addAnother: "Add another file",
    choose: "Tap to take a photo or choose a file",
    hint: "PDF, PNG or JPEG · up to 10 MB",
    unreadable: "This file can't be read as a PDF, PNG or JPEG. Choose another file.",
    failed: "The upload didn't finish. Nothing was saved. Try again.",
    offline: "The upload didn't finish. Nothing was saved. Check your connection and try again.",
    open: "Open",
  },

  home: {
    signedInAs: (name: string) => `Signed in as ${name}`,
    organisation: "Organisation",
    type: "Type",
    verification: "Verification",
    role: "Your role",
    noOrgTitle: "No organisation at this address.",
    noOrgBody: "Open your organisation's own address (its name comes first, like",
    noOrgExample: "acme.",
    noOrgBodyEnd: "before this site's name) to act for it.",
    adminOrgs: "Review organisations",
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
      [`${count} ${count === 1 ? "organisation" : "organisations"}`,
        oldestDays !== undefined && `oldest waiting ${oldestDays} ${oldestDays === 1 ? "day" : "days"}`,
        "target under 2 working days"].filter(Boolean).join(" · "),
    notAllowedTitle: "You can't review organisations.",
    notAllowedBody: "Only the platform's reviewers can open this page. Ask a platform admin if you need access.",
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
    back: "Back to ChainTrack",
    tabCount: (label: string, n: number) => `${label} (${n})`,
    queueLabel: (label: string) => `${label}: organisations`,
    age: (days: number) => (days === 0 ? "Today" : `${days} ${days === 1 ? "day" : "days"}`),
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
    noDocuments: "No documents uploaded yet. It can't be approved until it has at least one.",
    review: {
      title: "Decision",
      note: "Decision note",
      noteHelp: "Required to reject or ask for changes. The organisation sees it.",
      reject: "Reject",
      changes: "Ask for changes",
      approve: "Approve organisation",
      saving: "Saving…",
      done: {
        verified: "Organisation approved.",
        rejected: "Organisation rejected. They can upload again.",
        changes_requested: "Changes requested. They can upload again.",
      } as Record<string, string>,
      noDocuments: "This organisation has no documents to review. It can't be approved yet.",
    },
    accreditation: {
      title: "Accreditation",
      sub: "Choose which certificates this organisation may issue, and for how long.",
      types: "Certificate types they may issue",
      noTypes: "No templates ask for any evidence yet, so there are no certificate types to choose.",
      from: "Accredited from",
      until: "Accredited until",
      notice: "Accreditation is shown with every certificate they issue. Ending it doesn't cancel certificates they already issued.",
      decline: "Decline",
      accredit: "Accredit issuer",
      saving: "Saving…",
      saved: "Accreditation saved.",
      declined: "Accreditation removed.",
      notIssuer: "Only a verified certifier or regulator can be accredited.",
      current: (n: number) => (n === 0 ? "Not accredited" : `Accredited for ${n} ${n === 1 ? "type" : "types"}`),
    },
  },

  // Public verification page (screens.md §01).
  verification: {
    kicker: "Product passport check",
    publicLabel: "Public verification",
    checking: "Checking this lot's records…",
    notFoundTitle: "We can't find this lot code",
    notFoundBody: "Check the code on the label and scan it again. If the label is damaged, ask the holder for the lot code.",
    errorTitle: "We couldn't check this lot just now",
    errorBody: "The records could not be reached. Nothing about the lot has changed. Try again in a moment.",
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
    lotsToOrigin: (n: number) => `${n} ${n === 1 ? "lot" : "lots"}, back to origin`,
    technicalIntro:
      "When a record is signed, its fingerprint is sealed on a public ledger (a blockchain), where it cannot be changed. We compare each record with its sealed fingerprint every time this page is opened.",
    noCertificates: "No certificates are attached to this lot or its origin.",
    notSealed: "Not sealed yet",
    signedByBoth: "Signed by both parties",
    originMark: "Origin",
    revoked: "Revoked",
    unknownOrg: "Unknown organisation",
    footer: "ChainTrack product passport · The page reflects records sealed up to the time of your scan.",
    certificate: { valid: "Valid", expiring: "Expiring soon", expired: "Expired", revoked: "Revoked" },
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
      from: (product: string, places: string) => `This ${product} comes from ${places}.`,
      verified: "Every record in its history matches the copy sealed when it was signed, and nothing has changed since.",
      warning: (points: string) => `Its records are intact, but ${points} attention. Read them before you rely on this passport.`,
      pointsNeed: (n: number) => `${n} ${n === 1 ? "point needs" : "points need"}`,
      failed: "This passport failed verification. Do not rely on it until the points below are resolved.",
    },
    meta: {
      checked: (when: string) => `Checked ${when}`,
      records: (n: number) => `${n} ${n === 1 ? "record" : "records"}`,
      organisations: (n: number) => `${n} ${n === 1 ? "organisation" : "organisations"}`,
    },
    issued: (when: string) => `Passport issued ${when}`,
    certificateNote: {
      revoked: (reason: string | null) => `Revoked by its issuer${reason ? `: ${reason}` : "."} Ask the holder for a replacement.`,
      expired: (when: string) => `Expired ${when}. Ask the holder for a current one before you rely on it.`,
    },
    technical: {
      fingerprint: "Passport record · fingerprint",
      sealedIn: "Passport record · sealed in transaction",
      eventSealed: (title: string, when: string) => `${title} · ${when} · sealed in transaction`,
      certificateFingerprint: (type: string) => `${type} · fingerprint`,
    },
  },
};
