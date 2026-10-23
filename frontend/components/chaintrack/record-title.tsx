"use client";

import { createContext, useContext, useEffect } from "react";

/** Lets a detail page name its record for the top bar's breadcrumb ("Organisations › acme"). */
export const RecordTitleContext = createContext<(title?: string) => void>(
  () => {},
);

export function useRecordTitle(title?: string) {
  const set = useContext(RecordTitleContext);
  useEffect(() => {
    set(title);
    return () => set(undefined);
  }, [set, title]);
}
