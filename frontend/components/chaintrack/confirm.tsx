"use client";

import { CircleHelp, TriangleAlert, type LucideIcon } from "lucide-react";
import {
  createContext,
  useCallback,
  useContext,
  useRef,
  useState,
} from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from "@/components/ui/dialog";
import { confirmCopy } from "@/lib/consts";
import { cn } from "@/lib/utils";

/** What a confirmation says. `danger` for what can't be undone (it turns the button red). */
export type ConfirmOptions = {
  title: string;
  body: React.ReactNode;
  confirmLabel: string;
  danger?: boolean;
  icon?: LucideIcon;
};

type Ask = (o: ConfirmOptions) => Promise<boolean>;
const ConfirmContext = createContext<Ask | null>(null);

/**
 * Every decision asks first (patterns.md → Destructive actions):
 * `if (await confirm({...})) mutate()`. One dialog for the whole app,
 * mounted in app/providers.tsx.
 */
export function useConfirm(): Ask {
  const ask = useContext(ConfirmContext);
  if (!ask)
    throw new Error("useConfirm needs <ConfirmProvider> (app/providers.tsx)");
  return ask;
}

export function ConfirmProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useState(false);
  const [options, setOptions] = useState<ConfirmOptions | null>(null);
  const answer = useRef<(yes: boolean) => void>(() => {});

  const ask = useCallback<Ask>((o) => {
    answer.current(false); // a question still open is answered "no"
    setOptions(o);
    setOpen(true);
    return new Promise<boolean>((resolve) => {
      answer.current = resolve;
    });
  }, []);

  const close = (yes: boolean) => {
    answer.current(yes);
    answer.current = () => {};
    setOpen(false);
  };

  const Icon = options?.icon ?? (options?.danger ? TriangleAlert : CircleHelp);
  return (
    <ConfirmContext.Provider value={ask}>
      {children}
      <Dialog
        open={open}
        onOpenChange={(o) => {
          if (!o) close(false);
        }}
      >
        <DialogContent
          showCloseButton={false}
          className="gap-0 overflow-hidden rounded-lg border border-border bg-surface p-0 shadow-overlay sm:max-w-md"
        >
          {options && (
            <>
              <div className="flex gap-3 p-6 pb-0">
                <span
                  className={cn(
                    "grid size-10 shrink-0 place-items-center rounded-md",
                    options.danger
                      ? "bg-danger-soft text-danger"
                      : "bg-primary-soft text-primary",
                  )}
                >
                  <Icon className="size-5" aria-hidden />
                </span>
                <DialogTitle className="text-lg leading-snug font-semibold">
                  {options.title}
                </DialogTitle>
              </div>
              <DialogDescription
                render={<div />}
                className="p-6 text-sm text-ink-muted"
              >
                {options.body}
              </DialogDescription>
              <div className="flex flex-col-reverse gap-2 border-t border-border bg-surface-sunken p-4 sm:flex-row sm:justify-end">
                <DialogClose render={<Button variant="secondary" />}>
                  {confirmCopy.cancel}
                </DialogClose>
                <Button
                  variant={options.danger ? "danger" : "primary"}
                  onClick={() => close(true)}
                >
                  {options.confirmLabel}
                </Button>
              </div>
            </>
          )}
        </DialogContent>
      </Dialog>
    </ConfirmContext.Provider>
  );
}
