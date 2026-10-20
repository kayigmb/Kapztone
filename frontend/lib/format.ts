/**
 * Fill a copy template from JSON (`lib/copy/*.json`): every `{name}` becomes
 * `vars.name`. Templates come from the extractor, so an unknown placeholder
 * is a bug — it stays visible instead of rendering blank.
 */
export function fill(
  template: string,
  vars: Record<string, string | number> = {},
): string {
  return template.replace(/\{(\w+)\}/g, (m, name) =>
    vars[name] === undefined ? m : String(vars[name]),
  );
}

export function formatDate(
  date: Date | string | number | undefined,
  opts: Intl.DateTimeFormatOptions = {},
) {
  if (!date) return "";

  try {
    return new Intl.DateTimeFormat("en-US", {
      month: opts.month ?? "long",
      day: opts.day ?? "numeric",
      year: opts.year ?? "numeric",
      ...opts,
    }).format(new Date(date));
  } catch {
    return "";
  }
}
