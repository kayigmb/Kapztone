"use client";

import { useQuery } from "@tanstack/react-query";
import {
  ArrowLeft,
  ChevronRight,
  ArrowLeftRight,
  ChartColumn,
  ClipboardCheck,
  Ellipsis,
  Factory,
  FileText,
  Flag,
  House,
  Layers,
  MapPin,
  PackagePlus,
  PanelLeftClose,
  PanelLeftOpen,
  QrCode,
  RefreshCcw,
  UserRound,
  Building2,
  Users,
  type LucideIcon,
} from "lucide-react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useState } from "react";

import { SignOutButton } from "@/components/chaintrack/sign-out-button";
import { Brand } from "@/components/chaintrack/brand";
import { Notice } from "@/components/chaintrack/notice";
import { RecordTitleContext } from "@/components/chaintrack/record-title";
import { NotificationsBell } from "@/components/chaintrack/notifications-bell";
import { OrgSwitcher } from "@/components/chaintrack/org-switcher";
import { ScanButton } from "@/components/chaintrack/scan-button";
import { OrgStatusChip } from "@/components/chaintrack/status";
import { ThemeToggle } from "@/components/chaintrack/theme-toggle";
import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { endpoints, useMe, type Me } from "@/lib/api";
import {
  copy,
  navGroups,
  queryKeys,
  routes,
  sidebarStorageKey,
  superadminRole,
  type TileKey,
} from "@/lib/consts";
import { navFor } from "@/lib/dashboard";
import { cn } from "@/lib/utils";
import { humanize } from "@/lib/verification";

const t = copy.shell;
const icons: Record<string, LucideIcon> = {
  home: House,
  workspaces: Building2,
  orgs: ClipboardCheck,
  users: Users,
  profile: UserRound,
  facilities: Factory,
  issue: PackagePlus,
  handovers: ArrowLeftRight,
  transform: RefreshCcw,
  certificates: FileText,
  flags: Flag,
  templates: Layers,
  reports: ChartColumn,
  scan: QrCode,
  registerFacility: MapPin,
};
const isActive = (path: string, href: string) =>
  path === href || path.startsWith(`${href}/`);
type NavItem = ReturnType<typeof navFor>[number];

/** Up to two capitals for an avatar: "Imena Growers" → "IG". */
const initials = (name: string) =>
  name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((w) => w[0]!.toUpperCase())
    .join("") || "?";

// localStorage can be missing or refuse (private window); folding then just isn't remembered.
const readFolded = () => {
  try {
    return localStorage.getItem(sidebarStorageKey) === "1";
  } catch {
    return false;
  }
};
const saveFolded = (v: boolean) => {
  try {
    localStorage.setItem(sidebarStorageKey, v ? "1" : "0");
  } catch {
    /* not remembered */
  }
};

/** The signed-in frame: a navy sidebar with grouped nav on desktop, folding to an icon rail; a top bar with the page title, bell, theme and Scan; four tabs on phones. */
export function AppShell({
  me,
  children,
}: {
  me: Me;
  children: React.ReactNode;
}) {
  const path = usePathname();
  const router = useRouter();
  const all = navFor(me);
  const current = all.find((n) => isActive(path, n.href));
  const [more, setMore] = useState(false);
  const [folded, setFolded] = useState(readFolded); // the shell renders only in the browser, after sign-in loads
  const fold = (v: boolean) => {
    setFolded(v);
    saveFolded(v);
  };
  const orgName = me.organisation?.name ?? t.platform;
  // A record under a list (/organisations/123): the top bar reads "List › Name" and back goes to the list.
  const detail = current && path !== current.href ? current : undefined;
  const [record, setRecord] = useState<string>();

  const back = () =>
    detail
      ? router.push(detail.href)
      : window.history.length > 1
        ? router.back()
        : router.push(routes.home);

  const groups = (rail: boolean, onPick?: () => void) =>
    navGroups.map((g) => {
      const items = all.filter((n) => "group" in n && n.group === g);
      if (items.length === 0) return null;
      return (
        <div key={g} className="grid gap-1">
          {rail ? (
            <span
              className="mx-auto my-1 h-px w-6 bg-sidebar-line"
              aria-hidden
            />
          ) : (
            <span className="px-3 pt-1 text-xs font-bold tracking-[0.08em] text-sidebar-label uppercase">
              {t.groups[g]}
            </span>
          )}
          <ul className="grid gap-0.5">
            {items.map((n) => (
              <NavLink
                key={n.key}
                item={n}
                active={isActive(path, n.href)}
                rail={rail}
                onPick={onPick}
              />
            ))}
          </ul>
        </div>
      );
    });

  const orgCard = (rail: boolean) => (
    <div
      className={cn(
        "grid items-center gap-3 rounded-lg border border-sidebar-line bg-sidebar-hover",
        rail ? "justify-items-center p-1.5" : "grid-cols-[auto_1fr] p-2.5",
      )}
    >
      <span
        className="grid size-9 place-items-center rounded-md bg-primary text-xs font-bold text-on-primary"
        title={rail ? orgName : undefined}
        aria-hidden={!rail}
      >
        {initials(orgName)}
      </span>
      {!rail && (
        <div className="grid min-w-0 gap-1.5">
          <span className="truncate text-sm font-semibold text-sidebar-fg">
            {orgName}
          </span>
          <span className="truncate text-xs text-sidebar-label">
            {humanize(me.role)}
          </span>
          {me.organisation && <OrgStatusChip status={me.organisation.status} />}
        </div>
      )}
    </div>
  );

  return (
    <RecordTitleContext.Provider value={setRecord}>
      <div
        className={cn(
          "min-h-dvh lg:grid",
          folded
            ? "lg:grid-cols-[--spacing(19)_1fr]"
            : "lg:grid-cols-[--spacing(64)_1fr]",
        )}
      >
        <aside className="sticky top-0 hidden h-dvh flex-col bg-sidebar-bg text-sidebar-fg lg:flex">
          <div
            className={cn(
              "flex h-16 shrink-0 items-center gap-2 border-b border-sidebar-line",
              folded ? "justify-center px-2" : "justify-between pr-2 pl-5",
            )}
          >
            {!folded && (
              <Link href={routes.home} className="rounded-md">
                <Brand onNavy />
              </Link>
            )}
            <button
              type="button"
              onClick={() => fold(!folded)}
              aria-label={folded ? t.expand : t.collapse}
              aria-expanded={!folded}
              className="grid size-9 place-items-center rounded-md text-sidebar-muted hover:bg-sidebar-hover hover:text-sidebar-fg"
            >
              {folded ? (
                <PanelLeftOpen className="size-5" aria-hidden />
              ) : (
                <PanelLeftClose className="size-5" aria-hidden />
              )}
            </button>
          </div>
          <nav
            aria-label={t.menu}
            className={cn(
              "grid flex-1 content-start gap-4 overflow-y-auto pt-4 pb-2",
              folded ? "px-2" : "px-3",
            )}
          >
            {groups(folded)}
          </nav>
          <div
            className={cn(
              "grid gap-1 border-t border-sidebar-line",
              folded ? "p-2" : "p-3",
            )}
          >
            {orgCard(folded)}
            {!folded && <OrgSwitcher me={me} onNavy />}
            {!folded && <SignOutButton onNavy />}
          </div>
        </aside>

        <div className="flex min-w-0 flex-col pb-24 lg:pb-0">
          <header className="sticky top-0 z-30 flex h-14 items-center justify-between gap-2 border-b border-border bg-surface/95 px-3 backdrop-blur lg:h-16 lg:px-8">
            <div className="flex min-w-0 items-center gap-2">
              <Link
                href={routes.home}
                className="rounded-md lg:hidden"
                aria-label={copy.brand}
              >
                <Brand small nameHidden />
              </Link>
              {path !== routes.home && (
                <button
                  type="button"
                  onClick={back}
                  aria-label={t.back}
                  className="grid size-9 shrink-0 place-items-center rounded-md text-ink-muted hover:bg-surface-sunken hover:text-ink"
                >
                  <ArrowLeft className="size-5" aria-hidden />
                </button>
              )}
              {detail ? (
                <nav aria-label={t.breadcrumb} className="min-w-0">
                  <ol className="flex min-w-0 items-center gap-2 text-base">
                    <li className="shrink-0">
                      <Link
                        href={detail.href}
                        className="text-ink-muted hover:text-ink"
                      >
                        {t.nav[detail.key]}
                      </Link>
                    </li>
                    <li aria-hidden>
                      <ChevronRight className="size-4 text-ink-subtle" />
                    </li>
                    <li
                      aria-current="page"
                      className="min-w-0 truncate font-semibold"
                    >
                      {record ?? copy.loading}
                    </li>
                  </ol>
                </nav>
              ) : (
                <div className="grid min-w-0">
                  <span className="hidden text-xs font-semibold tracking-[0.08em] text-ink-muted uppercase lg:block">
                    {t.eyebrow}
                  </span>
                  <span className="truncate text-base font-semibold">
                    {current ? t.nav[current.key] : copy.brand}
                  </span>
                </div>
              )}
            </div>
            <div className="flex items-center gap-2">
              <NotificationsBell />
              <ThemeToggle />
              <ScanButton
                trigger={(open) => (
                  <button
                    type="button"
                    onClick={open}
                    className="hidden h-9 items-center gap-2 rounded-md border border-border bg-surface px-3 text-sm font-semibold hover:border-primary hover:text-primary lg:inline-flex"
                  >
                    <QrCode className="size-4.5" aria-hidden />
                    {t.scanQr}
                  </button>
                )}
              />
              <Link
                href={routes.profile}
                aria-label={t.yourProfile(me.user.name)}
                title={me.user.name}
                className="grid size-9 place-items-center rounded-full bg-primary-soft text-xs font-bold text-primary hover:ring-2 hover:ring-primary/40"
              >
                {initials(me.user.name)}
              </Link>
            </div>
          </header>
          <div className="min-w-0 flex-1">
            {me.role === superadminRole && me.organisation && (
              <div className="px-4 pt-4 lg:px-8">
                <Notice kind="info">
                  {copy.shell.actingAs(me.organisation.name)}
                </Notice>
              </div>
            )}
            {children}
          </div>
        </div>

        <PhoneTabs me={me} path={path} onMore={() => setMore(true)} />

        <Drawer open={more} onOpenChange={setMore}>
          <DrawerContent className="bg-sidebar-bg text-sidebar-fg">
            <DrawerHeader>
              <DrawerTitle className="text-sidebar-fg">{t.more}</DrawerTitle>
            </DrawerHeader>
            <div className="grid max-h-[75dvh] gap-4 overflow-y-auto px-4 pb-6">
              {orgCard(false)}
              <OrgSwitcher me={me} onNavy />
              {groups(false, () => setMore(false))}
              <SignOutButton onNavy />
            </div>
          </DrawerContent>
        </Drawer>
      </div>
    </RecordTitleContext.Provider>
  );
}

/** Phone tab bar: Home, the person's first workspace screen, Scan and More, each 56 px tall; a live count rides on its tab. */
function PhoneTabs({
  me,
  path,
  onMore,
}: {
  me: Me;
  path: string;
  onMore: () => void;
}) {
  const [home, ...rest] = navFor(me); // the dashboard, or the workspace picker
  const work = rest.find((n) => "group" in n);
  const tab =
    "relative grid min-h-tap-scan content-center justify-items-center gap-1 text-xs";
  const item = (n: NavItem) => {
    const Icon = icons[n.icon];
    const on = isActive(path, n.href);
    return (
      <Link
        key={n.key}
        href={n.href}
        aria-current={on ? "page" : undefined}
        className={cn(tab, on ? "font-bold text-primary" : "text-ink-muted")}
      >
        <Icon className="size-6" aria-hidden />
        <span className="max-w-full truncate px-1">{t.nav[n.key]}</span>
        <TabCount tile={"count" in n ? n.count : undefined} />
      </Link>
    );
  };
  return (
    <nav
      aria-label={t.menu}
      className="fixed inset-x-0 bottom-0 z-30 grid grid-cols-4 border-t border-border bg-surface px-1 pb-[env(safe-area-inset-bottom)] shadow-overlay lg:hidden"
    >
      {home && item(home)}
      {work ? item(work) : <span />}
      <ScanButton
        trigger={(open) => (
          <button
            type="button"
            onClick={open}
            className={cn(tab, "text-ink-muted")}
          >
            <QrCode className="size-6" aria-hidden />
            {t.scan}
          </button>
        )}
      />
      <button
        type="button"
        onClick={onMore}
        className={cn(tab, "text-ink-muted")}
      >
        <Ellipsis className="size-6" aria-hidden />
        {t.more}
      </button>
    </nav>
  );
}

function TabCount({ tile }: { tile: TileKey | undefined }) {
  const count = useNavCount(tile);
  if (!count) return null;
  return (
    <span className="absolute top-1 right-[18%] grid min-w-5 place-items-center rounded-full bg-danger px-1 font-mono text-xs font-semibold text-on-danger">
      {count}
    </span>
  );
}

/** One nav entry: icon + label, or icon with a tooltip on the folded rail; a live count when the item has one. */
function NavLink({
  item,
  active,
  rail,
  onPick,
}: {
  item: NavItem;
  active: boolean;
  rail: boolean;
  onPick?: () => void;
}) {
  const Icon = icons[item.icon];
  const text = t.nav[item.key];
  const count = useNavCount("count" in item ? item.count : undefined);
  const link = (
    <Link
      href={item.href}
      onClick={onPick}
      aria-current={active ? "page" : undefined}
      aria-label={rail ? text : undefined}
      className={cn(
        "relative flex min-h-tap items-center gap-3 rounded-md text-sm transition-colors",
        rail ? "justify-center" : "px-3",
        active
          ? "bg-sidebar-active font-semibold text-hero-ink"
          : "text-sidebar-muted hover:bg-sidebar-hover hover:text-hero-ink",
      )}
    >
      <Icon
        className={cn("size-4.5 shrink-0", active && "text-sidebar-accent-fg")}
        aria-hidden
      />
      {!rail && <span className="min-w-0 flex-1 truncate">{text}</span>}
      {!!count && (
        <span
          className={cn(
            "grid min-w-6 place-items-center rounded-full bg-sidebar-hover px-1.5 font-mono text-xs text-sidebar-fg",
            rail && "absolute -top-1 -right-1 bg-danger text-on-danger",
          )}
        >
          {count}
        </span>
      )}
    </Link>
  );
  return (
    <li>
      {rail ? (
        <Tooltip>
          <TooltipTrigger render={link} />
          <TooltipContent side="right">{text}</TooltipContent>
        </Tooltip>
      ) : (
        link
      )}
    </li>
  );
}

// Shares the dashboard tile's query, so the badge costs no extra call when both are on screen.
function useNavCount(tile: TileKey | undefined) {
  const handover = tile === "incoming" || tile === "outgoing";
  const mine = useMe().data?.organisation?.id;
  const q = useQuery({
    queryKey: queryKeys.count(tile ?? "none"),
    queryFn: () =>
      endpoints.count(tile as Exclude<TileKey, "incoming" | "outgoing">),
    enabled: !!tile && !handover,
  });
  // Handovers waiting on either side come from the dashboard's one list (same query key, no extra call).
  const awaiting = useQuery({
    queryKey: queryKeys.panel("handovers"),
    queryFn: endpoints.awaitingHandovers,
    enabled: handover,
  });
  if (!tile) return undefined;
  if (!handover) return q.data;
  return awaiting.data?.filter(
    (h) => (tile === "incoming" ? h.to_org : h.from_org) === mine,
  ).length;
}
