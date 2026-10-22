import { facilityStatus } from "./consts.ts";

export type FacilityState = "active" | "pending" | "suspended" | "expired";

/** Waiting for approval or suspended wins; an active facility past its licence end date is expired. */
export function facilityState(
  f: { status: string; licence_valid_to?: string | null },
  today = new Date().toISOString().slice(0, 10),
): FacilityState {
  if (f.status === facilityStatus.pending) return "pending";
  if (f.status === facilityStatus.suspended) return "suspended";
  if (f.licence_valid_to && f.licence_valid_to < today) return "expired";
  return "active";
}
