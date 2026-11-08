// Tool inventory: deterministic owner lookup and seam evidence for the
// determinism reviewer (determinism pack DE-001 and DE-007).
//
// Inventory mode lists every registered plugin tool (name, plugin,
// factory, source location, description) and every repo-local CLI
// script, so a reviewer names an existing owning tool for a mechanical
// step instead of searching for one by reading.
//
// Seam mode lists the plugin tools added, modified, or removed since
// merge-base(REF, HEAD) and, for each, the test files that exercise it,
// grouped by the three DE-007 seam layers (registration surface,
// documented output shape, write->read round trip). It reports
// evidence only; whether a missing layer is a finding is judgment.
//
// Run from the repository root:
//   node .opencode/lib/tool-inventory.ts [--json] [--match WORDS]
//   node .opencode/lib/tool-inventory.ts --seams --base REF [--json]
//
// Discovery is static (no plugin is loaded), so it runs under plain
// node; test/tool-inventory.test.ts pins it to the runtime registry.

import { existsSync, readdirSync, readFileSync, statSync } from "node:fs"
import { dirname, join, resolve } from "node:path"
// Plain-node CLI (type stripping): import the shared module by its real
// .ts extension; node does not remap .js specifiers to .ts files.
import { escapeControlChars, findRoot, git, isMainModule, runMain, spawnDetail, type CliIO } from "./cli-common.ts"

export interface ToolEntry {
  name: string
  plugin: string
  factory: string
  // repo-relative file and 1-based line range of the factory function
  file: string | null
  line: number | null
  endLine: number | null
  // 1-based line of the registry entry in the plugin's index.ts
  registryLine: number
  description: string | null
}

export interface CliEntry {
  name: string
  file: string
  summary: string
}

export interface Inventory {
  tools: ToolEntry[]
  clis: CliEntry[]
}

type Reader = (rel: string) => string | null

const fsReader =
  (root: string): Reader =>
  (rel) => {
    const abs = join(root, rel)
    return existsSync(abs) && statSync(abs).isFile() ? readFileSync(abs, "utf8") : null
  }

const lineOf = (text: string, offset: number): number => {
  let n = 1
  for (let i = 0; i < offset; i++) if (text.charCodeAt(i) === 10) n++
  return n
}

/** Escape a literal so it interpolates safely into a RegExp. */
const escapeRegExp = (s: string): string => s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")

/** Identifier boundary: names may contain $, which \b mishandles. */
const ident = (name: string): string => `(?<![\\w$])${escapeRegExp(name)}(?![\\w$])`

// Skip a string literal or comment starting at i; returns the index
// after it, or i when nothing is skipped.
function skipNonCode(text: string, i: number): number {
  const c = text[i]
  if (c === "/" && text[i + 1] === "/") {
    const nl = text.indexOf("\n", i)
    return nl === -1 ? text.length : nl
  }
  if (c === "/" && text[i + 1] === "*") {
    const end = text.indexOf("*/", i + 2)
    return end === -1 ? text.length : end + 2
  }
  if (c === '"' || c === "'" || c === "`") {
    let j = i + 1
    while (j < text.length && text[j] !== c) j += text[j] === "\\" ? 2 : 1
    return j + 1
  }
  return i
}

/** Comment-free text: comment spans blanked with spaces (offsets and
 * line breaks preserved) so entry-point detection never matches
 * explanatory prose. String literals stay as code: regex literals
 * with embedded quotes would derail string-skipping, and a `//` inside
 * a string can only hide an entry guard on its own line, never fake
 * one. */
function stripComments(text: string): string {
  let out = ""
  for (let i = 0; i < text.length; i++) {
    if (text[i] === "/" && text[i + 1] === "/") {
      const nl = text.indexOf("\n", i)
      const end = nl === -1 ? text.length : nl
      out += " ".repeat(end - i)
      i = end - 1
    } else if (text[i] === "/" && text[i + 1] === "*") {
      const close = text.indexOf("*/", i + 2)
      const end = close === -1 ? text.length : close + 2
      out += " ".repeat(end - i)
      i = end - 1
    } else {
      out += text[i]
    }
  }
  return out
}

// A lib module counts as a CLI when an entry guard calls runMain:
// either the import.meta.main idiom or the shared isMainModule helper,
// which also works where import.meta.main is undefined (Node < 24.2).
// The guard shape -- an if statement wrapping the runMain call -- keeps
// the shared definitions in cli-common.ts from matching, and matching
// code only keeps explanatory comments from listing libraries.
const ENTRY_BLOCK = /\bif\s*\(\s*(?:import\.meta\.main\b|isMainModule\s*\([^)]*\))\s*\)\s*runMain\s*\(/

/** Index of the registry's opening brace, or -1 when absent. Only code
 * positions count: `tool: {` inside a comment or string must not hijack
 * registry parsing. */
function findRegistryStart(text: string): number {
  let i = 0
  while (i < text.length) {
    const skipped = skipNonCode(text, i)
    if (skipped !== i) {
      i = skipped
      continue
    }
    if (text.startsWith("tool:", i) && (i === 0 || !/[\w$]/.test(text[i - 1]))) {
      let j = i + 5
      while (j < text.length && /\s/.test(text[j])) j++
      if (text[j] === "{") return j
    }
    i++
  }
  return -1
}

/** Parse the `tool: { name: factory(...), ... }` registry of a plugin
 * module. Only entries directly inside the registry object count.
 * @param text Plugin module source text.
 * @returns Tool name, factory function name, and 1-based line per entry. */
export function parseRegistry(text: string): Array<{ name: string; factory: string; line: number }> {
  const start = findRegistryStart(text)
  if (start === -1) return []
  const out: Array<{ name: string; factory: string; line: number }> = []
  const entry = /^\s*([A-Za-z_$][\w$]*)\s*:\s*([A-Za-z_$][\w$]*)\s*\(/
  let depth = 0
  let atEntry = false
  for (let i = start; i < text.length; ) {
    const skipped = skipNonCode(text, i)
    if (skipped !== i) {
      i = skipped
      continue
    }
    const c = text[i]
    if (c === "{" || c === "(" || c === "[") {
      depth++
      atEntry = depth === 1
    } else if (c === "}" || c === ")" || c === "]") {
      depth--
      if (depth === 0) break
    } else if (c === "," && depth === 1) {
      atEntry = true
    } else if (atEntry && !/\s/.test(c)) {
      atEntry = false
      const m = entry.exec(text.slice(i, i + 200))
      if (m) out.push({ name: m[1], factory: m[2], line: lineOf(text, i) })
    }
    i++
  }
  return out
}

/** Read a string-literal concatenation ("a" + 'b' + `c`) at offset i.
 * @param text Source text to read from.
 * @param i Offset of the opening quote.
 * @returns The unescaped string value, or null when `i` is not a quote. */
export function readStringConcat(text: string, i: number): string | null {
  let out = ""
  let any = false
  for (;;) {
    while (/\s/.test(text[i] ?? "")) i++
    const q = text[i]
    if (q !== '"' && q !== "'" && q !== "`") return any ? out : null
    const end = skipNonCode(text, i)
    const raw = text.slice(i + 1, end - 1)
    if (q === '"') {
      try {
        out += JSON.parse(`"${raw}"`) as string
      } catch {
        // TypeScript-only escapes (\x27, \u{2026}) are invalid JSON but
        // valid TS; unescape leniently instead of crashing.
        out += raw.replace(/\\(.)/g, "$1")
      }
    } else {
      out += raw.replace(/\\(.)/g, "$1")
    }
    any = true
    i = end
    while (/\s/.test(text[i] ?? "")) i++
    if (text[i] !== "+") return out
    i++
  }
}

interface FactoryLocation {
  file: string
  line: number
  endLine: number
  description: string | null
}

function locateIn(file: string, text: string, factory: string): FactoryLocation | null {
  const decl = new RegExp(`^(?:export\\s+)?(?:async\\s+)?(?:function\\s+${escapeRegExp(factory)}(?![\\w$])|const\\s+${escapeRegExp(factory)}\\s*=)`, "m").exec(text)
  if (!decl) return null
  const line = lineOf(text, decl.index)
  const lines = text.split("\n")
  // The declaration closes on the first line after which bracket depth
  // (comments and strings skipped) returns to zero, so both `export
  // function f() { ... }` and `const f = (...) => tool({ ... })` forms
  // end correctly.
  let endLine = lines.length
  let depth = 0
  let opened = false
  let pendingArrow = false
  let lineNo = line
  for (let i = decl.index; i < text.length; ) {
    const skipped = skipNonCode(text, i)
    if (skipped !== i) {
      i = skipped
      continue
    }
    const c = text[i]
    if (c === "{" || c === "(" || c === "[") {
      depth++
      opened = true
    } else if (c === "}" || c === ")" || c === "]") {
      depth--
    } else if (c === "=" && text[i + 1] === ">" && depth === 0) {
      // A dangling arrow (`=>` then end of line): the body starts on a
      // later line, so the parameter brackets closing at depth 0 above
      // do not end the declaration.
      let j = i + 2
      while (j < text.length && (text[j] === " " || text[j] === "\t")) j++
      if (text[j] === "\n" || j === text.length) pendingArrow = true
      i += 2
      continue
    } else if (c === "\n") {
      if (opened && depth <= 0 && !pendingArrow) {
        endLine = lineNo
        break
      }
      pendingArrow = false
      lineNo++
    }
    i++
  }
  if (endLine === lines.length && opened && depth <= 0) endLine = lineNo
  const body = lines.slice(line - 1, endLine).join("\n")
  const d = /\bdescription\s*:/.exec(body)
  return { file, line, endLine, description: d ? readStringConcat(body, d.index + d[0].length) : null }
}

/** Find a factory in the plugin file or in a relative module it imports.
 * @param pluginFile Repo-relative plugin index path.
 * @param factory Factory function name from the registry.
 * @param read File reader, injectable for tests.
 * @returns The factory's location and description, or null. */
export function locateFactory(pluginFile: string, factory: string, read: Reader): FactoryLocation | null {
  const text = read(pluginFile)
  if (text === null) return null
  const local = locateIn(pluginFile, text, factory)
  if (local) return local
  for (const m of text.matchAll(/import\s*(?:type\s*)?\{([^}]*)\}\s*from\s*["'](\.[^"']+)["']/g)) {
    if (!new RegExp(ident(factory)).test(m[1])) continue
    const rel = join(dirname(pluginFile), m[2]).replace(/\.js$/, ".ts").split("\\").join("/")
    const target = read(rel)
    if (target !== null) return locateIn(rel, target, factory)
  }
  return null
}

const listDir = (root: string, rel: string): string[] => {
  const abs = join(root, rel)
  return existsSync(abs) && statSync(abs).isDirectory() ? readdirSync(abs).sort() : []
}

/** List the repo-relative index paths of the installed plugins.
 * @param root Repository root.
 * @returns One `.opencode/plugins/<name>/index.ts` per installed plugin. */
export function pluginFiles(root: string): string[] {
  return listDir(root, ".opencode/plugins")
    .map((d) => `.opencode/plugins/${d}/index.ts`)
    .filter((f) => existsSync(join(root, f)))
}

/** Collect every registered plugin tool across the plugin files.
 * @param plugins Repo-relative plugin index paths.
 * @param read File reader, injectable for tests.
 * @returns Tool entries sorted by name, with resolved factory locations. */
export function collectTools(plugins: string[], read: Reader): ToolEntry[] {
  const out: ToolEntry[] = []
  for (const file of plugins) {
    const text = read(file)
    if (text === null) continue
    const plugin = file.split("/")[2]
    for (const r of parseRegistry(text)) {
      const loc = locateFactory(file, r.factory, read)
      out.push({
        name: r.name,
        plugin,
        factory: r.factory,
        file: loc?.file ?? null,
        line: loc?.line ?? null,
        endLine: loc?.endLine ?? null,
        registryLine: r.line,
        description: loc?.description ?? null,
      })
    }
  }
  return out.sort((x, y) => x.name.localeCompare(y.name))
}

/** Collect the repo-local runnable CLIs: lib modules with a runMain
 * entry guard and `.uf/scripts/` members.
 * @param root Repository root.
 * @returns CLI entries with file and first-comment summary. */
export function collectClis(root: string): CliEntry[] {
  const out: CliEntry[] = []
  for (const f of listDir(root, ".opencode/lib")) {
    if (!f.endsWith(".ts")) continue
    const text = readFileSync(join(root, ".opencode/lib", f), "utf8")
    if (!ENTRY_BLOCK.test(stripComments(text))) continue
    const first = /^\/\/\s?(.*)$/m.exec(text)
    out.push({ name: f.slice(0, -3), file: `.opencode/lib/${f}`, summary: first ? first[1].trim() : "" })
  }
  for (const f of listDir(root, ".uf/scripts")) {
    const text = readFileSync(join(root, ".uf/scripts", f), "utf8")
    const first = /^#(?!!)[ \t]*(\S.*)$/m.exec(text)
    out.push({ name: f, file: `.uf/scripts/${f}`, summary: first ? first[1].trim() : "" })
  }
  return out
}

/** Build the full tool and CLI inventory for a repository.
 * @param root Repository root.
 * @returns Every registered plugin tool and repo-local CLI. */
export function inventory(root: string): Inventory {
  return { tools: collectTools(pluginFiles(root), fsReader(root)), clis: collectClis(root) }
}

/**
 * Rank inventory entries by how many of the query words appear in the
 * name, summary, or description (case-insensitive). Zero-score entries
 * are dropped; ties sort by name.
 * @param inv The full inventory to filter.
 * @param words Query words.
 * @returns Matching entries ranked by score.
 */
export function match(inv: Inventory, words: string[]): Inventory {
  const ws = words.map((w) => w.toLowerCase()).filter(Boolean)
  const score = (hay: string): number => ws.filter((w) => hay.toLowerCase().includes(w)).length
  const rank = <T extends { name: string }>(xs: T[], hay: (x: T) => string): T[] =>
    xs
      .map((x) => ({ x, s: score(hay(x)) }))
      .filter((e) => e.s > 0)
      .sort((a, b) => b.s - a.s || a.x.name.localeCompare(b.x.name))
      .map((e) => e.x)
  return {
    tools: rank(inv.tools, (t) => `${t.name} ${t.description ?? ""}`),
    clis: rank(inv.clis, (c) => `${c.name} ${c.summary}`),
  }
}

// ---------------------------------------------------------------------
// Seam evidence (DE-007)
// ---------------------------------------------------------------------

export type SeamLayer = "registration" | "output-shape" | "round-trip" | "other"

export interface TestRef {
  file: string
  layer: SeamLayer
  // the test file changed since the merge base
  changed: boolean
}

export interface SeamEntry {
  tool: string
  plugin: string
  status: "added" | "modified" | "removed"
  // the plugin is covered by the generic registration-surface suite
  genericRegistration: boolean
  tests: TestRef[]
  missing: SeamLayer[]
}

export interface SeamReport {
  base: string
  mergeBase: string
  seams: SeamEntry[]
}

type Git = (args: string[]) => string

const realGit = (root: string): Git => (args) => git(root, args).toString("utf8")

/** Plugin ids credited by the generic registration-surface suite: the
 * `{ id: "..." }` entries of its PLUGINS array. Exported so a
 * real-repo discoverability test can pin this extraction against the
 * live test file instead of self-authored fixtures (DE-007).
 * @param surface Source text of registration-surface.test.ts.
 * @returns The plugin ids the suite registers. */
export function genericRegistrationPlugins(surface: string): Set<string> {
  return new Set([...surface.matchAll(/\{\s*id:\s*"([^"]+)"/g)].map((m) => m[1]))
}

/** Map a unified diff to its per-file changed head-revision lines.
 * @param diff Unified diff text (`git diff -U0` output).
 * @returns File path to the set of changed line numbers. */
export function parseChangedLines(diff: string): Map<string, Set<number>> {
  const out = new Map<string, Set<number>>()
  let cur: Set<number> | null = null
  for (const line of diff.split("\n")) {
    const f = /^\+\+\+ (?:b\/(.*)|\/dev\/null)$/.exec(line)
    if (f) {
      cur = f[1] ? (out.get(f[1]) ?? new Set()) : null
      if (f[1] && cur) out.set(f[1], cur)
      continue
    }
    const h = /^@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@/.exec(line)
    if (h && cur) {
      const start = Number(h[1])
      const count = h[2] === undefined ? 1 : Number(h[2])
      // A pure deletion (count 0) marks the line the deletion sits at.
      for (let n = start; n < start + Math.max(count, 1); n++) cur.add(n)
    }
  }
  return out
}

const LAYER_FILE: Array<[RegExp, SeamLayer]> = [
  [/registration-surface/, "registration"],
  [/golden|output-shape/, "output-shape"],
  [/round-?trip/, "round-trip"],
]

/** Classify a test file's seam layer for one tool.
 * @param file Test file path; layer suites match on the file name first.
 * @param text Test file source.
 * @param tool Tool name to look for in round-trip test titles.
 * @returns The seam layer the test covers. */
export function classifyTest(file: string, text: string, tool: string): SeamLayer {
  for (const [re, layer] of LAYER_FILE) if (re.test(file)) return layer
  const titled = new RegExp(`\\b(?:describe|it|test)\\([^\\n]*${ident(tool)}[^\\n]*round[- ]?trip|\\b(?:describe|it|test)\\([^\\n]*round[- ]?trip[^\\n]*${ident(tool)}`, "i")
  return titled.test(text) ? "round-trip" : "other"
}

/** Build the DE-007 seam evidence: plugin tools added, modified, or
 * removed since merge-base(ref, HEAD), with the test files that
 * reference each and the layers left uncovered.
 * @param root Repository root.
 * @param ref Base ref to diff against.
 * @param git Git runner, injectable for tests.
 * @returns The seam report for the change. */
export function seamReport(root: string, ref: string, git: Git = realGit(root)): SeamReport {
  const mergeBase = git(["merge-base", ref, "HEAD"]).trim()
  const untracked = git(["ls-files", "--others", "--exclude-standard", "--", ".opencode"]).split("\n").filter(Boolean)
  const changed = parseChangedLines(git(["diff", "-U0", "--no-color", "--no-ext-diff", mergeBase, "--", ".opencode"]))
  const read = fsReader(root)
  const isChanged = (file: string, line: number | null): boolean =>
    untracked.includes(file) || (line !== null && (changed.get(file)?.has(line) ?? false))
  const rangeChanged = (t: ToolEntry): boolean => {
    if (t.file === null || t.line === null || t.endLine === null) return false
    if (untracked.includes(t.file)) return true
    const lines = changed.get(t.file)
    if (!lines) return false
    for (const n of lines) if (n >= t.line && n <= t.endLine) return true
    return false
  }
  const readBase: Reader = (rel) => {
    try {
      return git(["show", `${mergeBase}:${rel}`])
    } catch {
      return null
    }
  }

  const headPlugins = pluginFiles(root)
  const basePlugins = git(["ls-tree", "--name-only", "-r", mergeBase, "--", ".opencode/plugins"])
    .split("\n")
    .filter((f) => /^\.opencode\/plugins\/[^/]+\/index\.ts$/.test(f))
  const head = collectTools(headPlugins, read)
  const baseNames = new Map<string, Set<string>>()
  for (const file of basePlugins) {
    const text = readBase(file)
    baseNames.set(file.split("/")[2], new Set(text === null ? [] : parseRegistry(text).map((r) => r.name)))
  }
  const headNames = new Map<string, Set<string>>()
  for (const t of head) headNames.set(t.plugin, (headNames.get(t.plugin) ?? new Set()).add(t.name))

  const testFiles = listDir(root, ".opencode/test")
    .filter((f) => f.endsWith(".test.ts"))
    .map((f) => `.opencode/test/${f}`)
  const testText = new Map(testFiles.map((f) => [f, readFileSync(join(root, f), "utf8")]))
  const changedTests = new Set([...changed.keys(), ...untracked].filter((f) => f.startsWith(".opencode/test/")))
  const genericPlugins = genericRegistrationPlugins(testText.get(".opencode/test/registration-surface.test.ts") ?? "")

  const entry = (tool: string, plugin: string, factory: string | null, status: SeamEntry["status"]): SeamEntry => {
    const names = [tool, ...(factory ? [factory] : [])]
    const tests: TestRef[] = []
    for (const [file, text] of testText) {
      if (!names.some((n) => new RegExp(ident(n)).test(text))) continue
      tests.push({ file, layer: classifyTest(file, text, tool), changed: changedTests.has(file) })
    }
    const genericRegistration = status !== "removed" && genericPlugins.has(plugin)
    const have = new Set(tests.map((t) => t.layer))
    if (genericRegistration) have.add("registration")
    const missing = status === "removed" ? [] : (["registration", "output-shape", "round-trip"] as SeamLayer[]).filter((l) => !have.has(l))
    return { tool, plugin, status, genericRegistration, tests, missing }
  }

  const seams: SeamEntry[] = []
  for (const t of head) {
    const added = !(baseNames.get(t.plugin)?.has(t.name) ?? false)
    const pluginFile = `.opencode/plugins/${t.plugin}/index.ts`
    if (added) seams.push(entry(t.name, t.plugin, t.factory, "added"))
    else if (rangeChanged(t) || isChanged(pluginFile, t.registryLine)) seams.push(entry(t.name, t.plugin, t.factory, "modified"))
  }
  for (const [plugin, names] of baseNames) {
    for (const name of names) if (!(headNames.get(plugin)?.has(name) ?? false)) seams.push(entry(name, plugin, null, "removed"))
  }
  seams.sort((x, y) => x.tool.localeCompare(y.tool))
  return { base: ref, mergeBase, seams }
}

// ---------------------------------------------------------------------
// CLI
// ---------------------------------------------------------------------

const USAGE = `usage: tool-inventory [--root DIR] [--json] [--match WORDS]
       tool-inventory --seams --base REF [--root DIR] [--json]

  --match WORDS  rank tools and scripts by how many of these words their name or
                 description contains (case-insensitive); drops non-matches
  --seams        list plugin tools added, modified, or removed since
                 merge-base(REF, HEAD) with their tests per DE-007 seam layer
  --base REF     base ref for --seams
  --json         machine-readable output`

const TOOL_VERSION: string | null = (() => {
  try {
    const pkg = JSON.parse(readFileSync(new URL("../package.json", import.meta.url), "utf8")) as { version?: string }
    return pkg.version ?? null
  } catch {
    return null
  }
})()

const one = (s: string | null): string => (s ?? "").replace(/\s+/g, " ").trim()

/** CLI entry point.
 * @param argv Arguments after the program name.
 * @param io Output writers and working directory.
 * @param git Git runner, injectable for tests.
 * @returns Process exit code (0 ok, 1 findings, 2 usage or IO error). */
export function main(argv: string[], io: CliIO, git?: Git): number {
  const flags = new Set<string>()
  let root: string | null = null
  let base: string | null = null
  const words: string[] = []
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i]
    if (a === "--root" || a === "--base" || a === "--match") {
      const v = argv[++i]
      if (!v) {
        io.err(`${a} needs a value\n${USAGE}`)
        return 2
      }
      if (a === "--root") root = resolve(io.cwd, v)
      else if (a === "--base") base = v
      else words.push(...v.split(/\s+/))
    } else if (["--json", "--seams", "--help"].includes(a)) {
      flags.add(a)
    } else {
      io.err(`unknown argument: ${a}\n${USAGE}`)
      return 2
    }
  }
  if (flags.has("--help")) {
    io.out(USAGE)
    return 0
  }
  root = root ?? findRoot(io.cwd)
  if (!root) {
    io.err("no .opencode/ directory found; pass --root")
    return 2
  }
  const json = flags.has("--json")
  // Text-mode output carries file-derived content: escape control
  // characters at the print sites; JSON branches stay raw (Constitution
  // III provenance on --json, V input handling on text paths).
  const print = (s: string): void => io.out(escapeControlChars(s))
  const provenance = { tool: "tool-inventory", version: TOOL_VERSION, generatedAt: new Date().toISOString() }

  if (flags.has("--seams")) {
    if (!base) {
      io.err(`--seams needs --base REF\n${USAGE}`)
      return 2
    }
    let report: SeamReport
    try {
      report = seamReport(root, base, git)
    } catch (e) {
      const detail = spawnDetail(e)
      io.err(`cannot read base ${base}: ${detail}`)
      return 2
    }
    if (json) {
      io.out(JSON.stringify({ provenance, ...report }, null, 2))
      return 0
    }
    print(`seams since ${report.base} (merge-base ${report.mergeBase.slice(0, 12)}): ${report.seams.length} tool(s)`)
    for (const s of report.seams) {
      print(`\n${s.tool}  [${s.plugin}, ${s.status}]`)
      if (s.genericRegistration) print("  registration  generic: .opencode/test/registration-surface.test.ts")
      for (const t of s.tests) print(`  ${t.layer.padEnd(12)}  ${t.file}${t.changed ? "  (changed)" : ""}`)
      if (s.missing.length > 0) print(`  missing: ${s.missing.join(", ")}`)
    }
    return 0
  }

  let inv = inventory(root)
  if (words.length > 0) inv = match(inv, words)
  if (json) {
    io.out(JSON.stringify({ provenance, ...inv }, null, 2))
    return 0
  }
  for (const t of inv.tools) {
    const where = t.file ? `${t.file}:${t.line}` : "(factory not found)"
    print(`${t.name}  [${t.plugin}]  ${where}\n  ${one(t.description)}`)
  }
  for (const c of inv.clis) print(`${c.file}  [cli]\n  ${one(c.summary)}`)
  print(`\n${inv.tools.length} tools, ${inv.clis.length} scripts`)
  return 0
}

if (isMainModule(import.meta.url)) runMain(main, "tool-inventory")
