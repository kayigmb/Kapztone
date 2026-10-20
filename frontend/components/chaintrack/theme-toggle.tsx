"use client";

import { Monitor, Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";
import { useState } from "react";

import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { copy, themes, type ThemeChoice } from "@/lib/consts";
import { cn } from "@/lib/utils";

const icon = { system: Monitor, light: Sun, dark: Moon };
const t = copy.shell;

/** Light / Dark / System menu. The shell only renders after sign-in has loaded, so the stored choice is known. */
export function ThemeToggle() {
  const { theme, setTheme } = useTheme();
  const [open, setOpen] = useState(false);
  const now = (themes as readonly string[]).includes(theme ?? "")
    ? (theme as ThemeChoice)
    : "light";
  const Icon = icon[now];
  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        aria-label={`${t.themeMenu}: ${t.theme[now]}`}
        className="grid size-9 place-items-center rounded-md border border-border text-ink-muted hover:bg-surface-sunken hover:text-ink"
      >
        <Icon className="size-4.5" aria-hidden />
      </PopoverTrigger>
      <PopoverContent align="end" className="w-40 gap-0.5 p-1">
        {themes.map((choice) => {
          const ChoiceIcon = icon[choice];
          const on = choice === now;
          return (
            <button
              key={choice}
              type="button"
              aria-pressed={on}
              onClick={() => {
                setTheme(choice);
                setOpen(false);
              }}
              className={cn(
                "grid h-9 grid-cols-[auto_--spacing(4)_1fr] items-center gap-2 rounded-md px-2 text-left text-sm hover:bg-surface-sunken",
                on
                  ? "bg-surface-sunken font-semibold text-primary"
                  : "text-ink",
              )}
            >
              <span
                className={cn("size-1.5 rounded-full", on && "bg-primary")}
                aria-hidden
              />
              <ChoiceIcon className="size-4" aria-hidden />
              {t.theme[choice]}
            </button>
          );
        })}
      </PopoverContent>
    </Popover>
  );
}
