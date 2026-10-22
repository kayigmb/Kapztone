import type { Verdict } from "@/lib/verification";

export type Tone = "verified" | "warning" | "danger" | "pending" | "neutral";

// Tone → token set.
export const tone = {
  verified: {
    solid: "bg-verified text-on-verified",
    soft: "bg-verified-soft text-verified-ink",
    fg: "text-verified",
    ink: "text-verified-ink",
    border: "border-verified",
    chip: "border-verified/30",
  },
  warning: {
    solid: "bg-warning text-on-warning",
    soft: "bg-warning-soft text-warning-ink",
    fg: "text-warning",
    ink: "text-warning-ink",
    border: "border-warning",
    chip: "border-warning/30",
  },
  danger: {
    solid: "bg-danger text-on-danger",
    soft: "bg-danger-soft text-danger-ink",
    fg: "text-danger",
    ink: "text-danger-ink",
    border: "border-danger",
    chip: "border-danger/30",
  },
  pending: {
    solid: "bg-pending text-on-primary",
    soft: "bg-pending-soft text-pending-ink",
    fg: "text-pending",
    ink: "text-pending-ink",
    border: "border-pending",
    chip: "border-pending/30",
  },
  neutral: {
    solid: "bg-surface-sunken text-ink",
    soft: "bg-surface-sunken text-ink-muted",
    fg: "text-ink-muted",
    ink: "text-ink-muted",
    border: "border-border",
    chip: "border-border",
  },
} as const satisfies Record<Tone, Record<string, string>>;

export const verdictTone: Record<Verdict, Tone> = {
  verified: "verified",
  warning: "warning",
  failed: "danger",
};
