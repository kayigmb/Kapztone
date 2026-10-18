"use client";

import { Monitor, Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";

import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { copy, themes, type ThemeChoice } from "@/lib/consts";

const icon = { system: Monitor, light: Sun, dark: Moon };

/** Cycles device → light → dark. The shell only renders after sign-in has loaded, so the stored choice is known. */
export function ThemeToggle() {
  const { theme, setTheme } = useTheme();
  const now = (themes as readonly string[]).includes(theme ?? "") ? (theme as ThemeChoice) : "system";
  const next = themes[(themes.indexOf(now) + 1) % themes.length];
  const Icon = icon[now];
  return (
    <Tooltip>
      <TooltipTrigger
        render={<button type="button" onClick={() => setTheme(next)} aria-label={copy.shell.theme[now]} />}
        className="grid size-tap place-items-center rounded-md border border-border text-ink-muted hover:bg-surface-sunken hover:text-ink"
      >
        <Icon className="size-5" aria-hidden />
      </TooltipTrigger>
      <TooltipContent>{copy.shell.theme[now]}</TooltipContent>
    </Tooltip>
  );
}
