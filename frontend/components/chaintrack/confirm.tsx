"use client";

import { CircleHelp, TriangleAlert, type LucideIcon } from "lucide-react";
import {
  createContext,
  useCallback,
  useContext,
  useRef,
  useState,
} from "react";

import { TextField } from "@/components/chaintrack/text-field";
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

/** What a confirmation says. */
export type ConfirmOptions = {
  title: string;
  body: React.ReactNode;
  confirmLabel: string;
  danger?: boolean;
  icon?: LucideIcon;
  /** Ask for a written reason too; confirming then answers with it. */
  reason?: { label: string; help?: string };
};

/** false: cancelled; true: confirmed; a string: confirmed with this reason. */
type Ask = (o: ConfirmOptions) => Promise<boolean | string>;
const ConfirmContext = createContext<Ask | null>(null);

/** Every decision asks first: `if (await confirm({...})) mutate()`. */
export function useConfirm(): Ask {
  const ask = useContext(ConfirmContext);
  if (!ask)
    throw new Error("useConfirm needs <ConfirmProvider> (app/providers.tsx)");
  return ask;
}

export function ConfirmProvider({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useState(false);
  const [options, setOptions] = useState<ConfirmOptions | null>(null);
  const [reason, setReason] = useState("");
  const answer = useRef<(a: boolean | string) => void>(() => {});

  const ask = useCallback<Ask>((o) => {
    answer.current(false); // a question still open is answered "no"
    setOptions(o);
    setReason("");
    setOpen(true);
    return new Promise<boolean | string>((resolve) => {
      answer.current = resolve;
    });
  }, []);

  const close = (yes: boolean) => {
    answer.current(yes && options?.reason ? reason.trim() : yes);
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
              {options.reason && (
                <TextField
                  className="px-6 pb-6"
                  label={options.reason.label}
                  help={options.reason.help}
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                  autoFocus
                />
              )}
              <div className="flex flex-col-reverse gap-2 border-t border-border bg-surface-sunken p-4 sm:flex-row sm:justify-end">
                <DialogClose render={<Button variant="secondary" />}>
                  {confirmCopy.cancel}
                </DialogClose>
                <Button
                  variant={options.danger ? "danger" : "primary"}
                  disabled={!!options.reason && !reason.trim()}
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
