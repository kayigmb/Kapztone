"use client";

// The dashboard's charts (Recharts), following the dataviz rules: thin marks,
// one axis, categorical colours in a fixed order from the theme tokens
// (--chart-1…5), a legend for two or more series, a tooltip on hover, and a
// table view beside every chart (light-mode colours rely on it, see globals.css).

import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

import { useThemeTokens } from "@/components/chaintrack/use-theme-tokens";
import { copy } from "@/lib/consts";

const chartTokens = [
  "chart-1",
  "chart-2",
  "chart-3",
  "chart-4",
  "chart-5",
  "ink-muted",
  "border",
  "surface",
  "ink",
] as const;
const seriesTokens = [
  "chart-1",
  "chart-2",
  "chart-3",
  "chart-4",
  "chart-5",
] as const;

export type Series<K extends string> = { key: K; label: string };

const axis = (c: Record<string, string>) => ({
  stroke: c["border"],
  tick: { fill: c["ink-muted"], fontSize: 12 },
  tickLine: false,
});
const tooltipStyle = (c: Record<string, string>) => ({
  contentStyle: {
    background: c["surface"],
    border: `1px solid ${c["border"]}`,
    borderRadius: 8,
    fontSize: 12,
    color: c["ink"],
  },
  labelStyle: { color: c["ink"], fontWeight: 600 },
});

/** Several counts over the same days: one line each, one axis. */
export function LinesOverTime<K extends string>({
  rows,
  x,
  series,
  label,
}: {
  rows: Record<string, string | number>[];
  x: string;
  series: Series<K>[];
  label: string;
}) {
  const c = useThemeTokens(chartTokens);
  return (
    <figure className="grid gap-3">
      <div className="h-72" role="img" aria-label={label}>
        <ResponsiveContainer width="100%" height="100%">
          <LineChart
            data={rows}
            margin={{ top: 8, right: 8, bottom: 0, left: -16 }}
          >
            <CartesianGrid
              stroke={c["border"]}
              strokeDasharray="2 4"
              vertical={false}
            />
            <XAxis
              dataKey={x}
              {...axis(c)}
              tickFormatter={(d: string) => d.slice(5)}
              minTickGap={24}
            />
            <YAxis {...axis(c)} allowDecimals={false} axisLine={false} />
            <Tooltip {...tooltipStyle(c)} />
            <Legend
              iconType="plainline"
              wrapperStyle={{ fontSize: 12, color: c["ink-muted"] }}
            />
            {series.map((s, i) => (
              <Line
                key={s.key}
                type="monotone"
                dataKey={s.key}
                name={s.label}
                stroke={c[seriesTokens[i % seriesTokens.length]]}
                strokeWidth={2}
                dot={false}
                activeDot={{ r: 4, strokeWidth: 2, stroke: c["surface"] }}
              />
            ))}
          </LineChart>
        </ResponsiveContainer>
      </div>
      <TableView
        label={label}
        head={[copy.dashboard.dateHeader, ...series.map((s) => s.label)]}
        rows={rows.map((r) => [
          String(r[x]),
          ...series.map((s) => Number(r[s.key]).toLocaleString()),
        ])}
      />
    </figure>
  );
}

/** One measure across categories, as horizontal bars (labels stay readable). */
export function Bars({
  rows,
  label,
  valueLabel,
  format = (n) => n.toLocaleString(),
}: {
  rows: { name: string; value: number }[];
  label: string;
  valueLabel: string;
  format?: (n: number) => string;
}) {
  const c = useThemeTokens(chartTokens);
  const height = Math.max(120, rows.length * 36 + 24);
  return (
    <figure className="grid gap-3">
      <div style={{ height }} role="img" aria-label={label}>
        <ResponsiveContainer width="100%" height="100%">
          <BarChart
            data={rows}
            layout="vertical"
            margin={{ top: 0, right: 16, bottom: 0, left: 0 }}
            barCategoryGap={8}
          >
            <CartesianGrid
              stroke={c["border"]}
              strokeDasharray="2 4"
              horizontal={false}
            />
            <XAxis
              type="number"
              {...axis(c)}
              allowDecimals={false}
              tickFormatter={format}
            />
            <YAxis
              type="category"
              dataKey="name"
              width={120}
              {...axis(c)}
              axisLine={false}
            />
            <Tooltip
              {...tooltipStyle(c)}
              formatter={(v) => [format(Number(v)), valueLabel]}
              cursor={{ fill: c["border"], opacity: 0.4 }}
            />
            <Bar
              dataKey="value"
              name={valueLabel}
              fill={c["chart-1"]}
              radius={[0, 4, 4, 0]}
              maxBarSize={20}
            />
          </BarChart>
        </ResponsiveContainer>
      </div>
      <TableView
        label={label}
        head={[label, valueLabel]}
        rows={rows.map((r) => [r.name, format(r.value)])}
      />
    </figure>
  );
}

/** Every chart's numbers, as a table, folded away until asked for. */
function TableView({
  label,
  head,
  rows,
}: {
  label: string;
  head: string[];
  rows: string[][];
}) {
  return (
    <details className="text-sm">
      <summary className="w-fit cursor-pointer text-xs font-semibold text-primary">
        {copy.dashboard.showTable}
      </summary>
      <div className="mt-2 max-h-64 overflow-auto rounded-md border border-border">
        <table className="w-full text-xs">
          <caption className="sr-only">{label}</caption>
          <thead className="sticky top-0 bg-surface-sunken text-left">
            <tr>
              {head.map((h) => (
                <th key={h} scope="col" className="px-3 py-2 font-semibold">
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {rows.map((r, i) => (
              <tr key={i}>
                {r.map((cell, j) => (
                  <td key={j} className="px-3 py-1.5 font-mono">
                    {cell}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </details>
  );
}
