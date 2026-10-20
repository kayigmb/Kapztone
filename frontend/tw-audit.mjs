import { createRequire } from "node:module";
import fs from "node:fs";
import path from "node:path";

const oxPath = "/Users/bk/Documents/alu/Kapztone/frontend/node_modules/.pnpm/@tailwindcss+oxide@4.3.3/node_modules/@tailwindcss/oxide/index.js";
const require = createRequire(oxPath);
const { Scanner } = require("@tailwindcss/oxide");

const s = fs.readFileSync("/Users/bk/Documents/alu/Kapztone/frontend/.next/static/chunks/28x30zsd2ek0s.css", "utf8")
  + fs.readFileSync("/Users/bk/Documents/alu/Kapztone/frontend/.next/static/chunks/0eh3hg0a099nx.css", "utf8");

function walk(dir) {
  const out = [];
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) {
      if (e.name === "node_modules" || e.name === ".next") continue;
      out.push(...walk(p));
    } else if (/\.(tsx|ts|css)$/.test(e.name)) {
      out.push(p);
    }
  }
  return out;
}

const scanner = new Scanner({});
const files = walk("/Users/bk/Documents/alu/Kapztone/frontend/app").concat(
  walk("/Users/bk/Documents/alu/Kapztone/frontend/components"),
  walk("/Users/bk/Documents/alu/Kapztone/frontend/lib"),
);
console.error("files:", files.length);
// per-file candidate extraction; report files whose distinctive classes are absent
const built = s;
let invisible = 0;
for (const f of files) {
  const content = fs.readFileSync(f, "utf8");
  const ext = f.split(".").pop();
  const raw = scanner.getCandidatesWithPositions({ content, extension: ext });
  const list = Array.isArray(raw) ? raw.map((e) => (Array.isArray(e) ? e[0] : e.candidate ?? e)) : Object.keys(raw);
  // distinctive = long class-like tokens (>=12 chars, has dash) unique-ish
  const distinctive = [...new Set(list.filter((c) => /^[A-Za-z][A-Za-z0-9/_:.[\]@-]{11,}$/.test(c) && !/^(absolute|relative|flex|grid|hidden|block|inline|truncate|shrink|grow|w-full|h-full|min-h-tap|text-sm|font|rounded|border|shadow|bg-|text-|p-|m-|gap|space|items|justify|overflow|whitespace|truncate)/.test(c)))];
  const missing = distinctive.filter((c) => !built.includes(c.split(":").pop().replace(/^\//, "")));
  if (missing.length > 0) {
    invisible++;
    console.log("FILE:", f.replace("/Users/bk/Documents/alu/Kapztone/frontend/", ""), `cand=${list.length} missingSample=${JSON.stringify(missing.slice(0, 4))}`);
  }
}
console.log("files with missing distinctive classes:", invisible, "of", files.length);
