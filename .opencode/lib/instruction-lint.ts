// Instruction lint: deterministic static analysis of the prose that
// drives the agent ecosystem (agents, commands, skills, convention
// packs, and always-on AGENTS.md context).
//
// Every check here is mechanical (same input, same output), per
// Constitution Principle VI and determinism pack rule DE-001. Judgment
// checks (is this objective clear? are decision criteria explicit?)
// belong in a review pack, not in this tool.
//
// Run from the repository root:
//   node .opencode/lib/instruction-lint.ts [--json] [--strict] [--all] [--report] [--base REF]
//
// Suppress a finding with an HTML comment in the instruction file:
//   <!-- instruction-lint-disable-next-line IL-023 -->
//   <!-- instruction-lint-disable IL-041 -->   (whole file)

import { execFileSync } from "node:child_process"
import { existsSync, readdirSync, readFileSync, statSync } from "node:fs"
import { join, relative, resolve, sep } from "node:path"
// `yaml` is not a declared dependency: the scaffold tests pin the plugin
// package to exactly two runtime dependencies. It is installed
// transitively (via @opencode-ai/plugin -> effect, and vitest -> vite).
import { parseDocument } from "yaml"

export type Severity = "error" | "warning" | "info"
export type ArtifactKind = "agent" | "command" | "skill" | "pack" | "context"

export interface RuleInfo {
  severity: Severity
  summary: string
}

export const RULES: Record<string, RuleInfo> = {
  "IL-001": { severity: "error", summary: "Frontmatter missing, unterminated, or not a YAML mapping" },
  "IL-002": { severity: "error", summary: "Frontmatter field missing or invalid for this artifact type" },
  "IL-003": { severity: "error", summary: "Declared name does not match the file or directory name" },
  "IL-004": { severity: "warning", summary: "Description too short or too long to route on" },
  "IL-010": { severity: "error", summary: "Reference to a slash command that does not exist" },
  "IL-011": { severity: "error", summary: "Reference to an agent that does not exist" },
  "IL-012": { severity: "error", summary: "Reference to a skill that does not exist" },
  "IL-013": { severity: "error", summary: "Reference to a .opencode/ path that does not exist" },
  "IL-014": { severity: "error", summary: "Rule ID defined more than once in a pack" },
  "IL-015": { severity: "warning", summary: "Reference to a rule ID that no pack defines" },
  "IL-020": { severity: "error", summary: "Pack rule has no [MUST]/[SHOULD]/[MAY] marker" },
  "IL-021": { severity: "warning", summary: "Pack rule marker disagrees with the RFC 2119 keywords in its body" },
  "IL-022": { severity: "warning", summary: "Lowercase must/should/shall in a pack rule has no RFC 2119 force" },
  "IL-023": { severity: "warning", summary: "Vague qualifier inside a normative (MUST/SHOULD/NEVER/ALWAYS) sentence" },
  "IL-030": { severity: "warning", summary: "File exceeds the per-file token budget" },
  "IL-031": { severity: "warning", summary: "Context loaded by a command or agent exceeds the load budget" },
  "IL-040": { severity: "info", summary: "Prescribed-output phrasing; render it from a tool instead (DE-003)" },
  "IL-041": { severity: "info", summary: "Multi-command shell procedure in prose; candidate for a tested tool (DE-001)" },
  "IL-042": { severity: "warning", summary: "Delegation call block omits task_classification or the skill contract (DE-008)" },
  "IL-050": { severity: "info", summary: "Near-duplicate paragraphs across files (instruction drift risk)" },
}

export type ReviewSeverity = "HIGH" | "MEDIUM" | "LOW"

export interface Finding {
  rule: string
  severity: Severity
  // Divisor severity (.opencode/uf/packs/severity.md) for this
  // finding, so reviewers never re-map it from prose (DE-006).
  reviewSeverity: ReviewSeverity
  file: string
  line: number
  message: string
}

const REVIEW_SEVERITY: Record<Severity, ReviewSeverity> = { error: "HIGH", warning: "MEDIUM", info: "LOW" }

export interface LintOptions {
  maxFileTokens: number
  maxLoadTokens: number
  duplicateThreshold: number
  duplicateMinWords: number
  disable: string[]
  ignorePaths: string[]
}

export const DEFAULT_OPTIONS: LintOptions = {
  maxFileTokens: 8000,
  maxLoadTokens: 20000,
  duplicateThreshold: 0.85,
  duplicateMinWords: 40,
  disable: [],
  ignorePaths: [],
}

// Built-in opencode slash commands and subagents that resolve without
// a file in this repository.
const BUILTIN_COMMANDS = new Set([
  "compact", "details", "editor", "exit", "export", "help", "init",
  "models", "new", "redo", "sessions", "share", "themes", "undo", "unshare",
])
const BUILTIN_AGENTS = new Set(["build", "explore", "general", "general-purpose", "plan"])

interface LineSuppression {
  all: boolean
  rules: Set<string>
}

export interface Artifact {
  kind: ArtifactKind
  name: string
  file: string
  lines: string[]
  hasFrontmatter: boolean
  frontmatter: Record<string, unknown> | null
  frontmatterError: string | null
  // 0-based index of the first body line (after closing ---)
  bodyStart: number
  // per 0-based line: true when the line is a fenced-code delimiter or content
  fenced: boolean[]
  tokens: number
  fileSuppression: LineSuppression
  lineSuppression: Map<number, LineSuppression>
}

// ---------------------------------------------------------------------
// Parsing
// ---------------------------------------------------------------------

export function estimateTokens(text: string): number {
  return Math.ceil(text.length / 4)
}

export function buildArtifact(file: string, rawText: string, kind: ArtifactKind, name: string): Artifact {
  const text = rawText.replace(/\r\n/g, "\n")
  const lines = text.split("\n")
  let hasFrontmatter = false
  let frontmatter: Record<string, unknown> | null = null
  let frontmatterError: string | null = null
  let bodyStart = 0

  if (lines[0] === "---") {
    hasFrontmatter = true
    let end = -1
    for (let i = 1; i < lines.length; i++) {
      if (lines[i].trim() === "---") {
        end = i
        break
      }
    }
    if (end === -1) {
      frontmatterError = "frontmatter has no closing ---"
    } else {
      bodyStart = end + 1
      const doc = parseDocument(lines.slice(1, end).join("\n"))
      if (doc.errors.length > 0) {
        frontmatterError = `invalid YAML: ${doc.errors[0].message.split("\n")[0]}`
      } else {
        const data: unknown = doc.toJS()
        if (data !== null && typeof data === "object" && !Array.isArray(data)) {
          frontmatter = data as Record<string, unknown>
        } else {
          frontmatterError = "frontmatter is not a YAML mapping"
        }
      }
    }
  }

  const fenced = new Array<boolean>(lines.length).fill(false)
  let open: { ch: string; len: number } | null = null
  for (let i = bodyStart; i < lines.length; i++) {
    const m = /^\s{0,3}(`{3,}|~{3,})/.exec(lines[i])
    if (open) {
      fenced[i] = true
      if (m && m[1][0] === open.ch && m[1].length >= open.len && lines[i].trim() === m[1]) open = null
    } else if (m) {
      fenced[i] = true
      open = { ch: m[1][0], len: m[1].length }
    }
  }

  const fileSuppression: LineSuppression = { all: false, rules: new Set() }
  const lineSuppression = new Map<number, LineSuppression>()
  const sup = /<!--\s*instruction-lint-disable(-next-line)?\b([^>]*?)-->/g
  lines.forEach((line, i) => {
    for (const m of line.matchAll(sup)) {
      const ids = m[2].split(/[\s,]+/).filter(Boolean)
      let target = fileSuppression
      if (m[1]) {
        target = lineSuppression.get(i + 2) ?? { all: false, rules: new Set() }
        lineSuppression.set(i + 2, target)
      }
      if (ids.length === 0) target.all = true
      for (const id of ids) target.rules.add(id)
    }
  })

  return {
    kind, name, file, lines, hasFrontmatter, frontmatter, frontmatterError,
    bodyStart, fenced, tokens: estimateTokens(text), fileSuppression, lineSuppression,
  }
}

function isDir(p: string): boolean {
  return existsSync(p) && statSync(p).isDirectory()
}

function toPosix(p: string): string {
  return p.split(sep).join("/")
}

// Classify a repo-relative posix path as an instruction artifact using
// the same layout discoverArtifacts walks, so base-revision trees read
// from git are discovered identically to the working tree.
export function classifyInstructionPath(rel: string): { kind: ArtifactKind; name: string } | null {
  if (rel === "AGENTS.md") return { kind: "context", name: "AGENTS" }
  const flat = /^\.opencode\/(agents|commands|uf\/packs)\/([^/]+)\.md$/.exec(rel)
  if (flat) {
    const kind: ArtifactKind = flat[1] === "agents" ? "agent" : flat[1] === "commands" ? "command" : "pack"
    return { kind, name: flat[2] }
  }
  const skill = /^\.opencode\/skills?\/([^/]+)\/SKILL\.md$/.exec(rel)
  return skill ? { kind: "skill", name: skill[1] } : null
}

export function discoverArtifacts(root: string): Artifact[] {
  const out: Artifact[] = []
  const oc = join(root, ".opencode")
  const load = (abs: string, kind: ArtifactKind, name: string): void => {
    out.push(buildArtifact(toPosix(relative(root, abs)), readFileSync(abs, "utf8"), kind, name))
  }
  const flat: Array<[string, ArtifactKind]> = [["agents", "agent"], ["commands", "command"], [join("uf", "packs"), "pack"]]
  for (const [dir, kind] of flat) {
    const abs = join(oc, dir)
    if (!isDir(abs)) continue
    for (const f of readdirSync(abs).sort()) {
      if (f.endsWith(".md")) load(join(abs, f), kind, f.slice(0, -3))
    }
  }
  for (const dir of ["skills", "skill"]) {
    const abs = join(oc, dir)
    if (!isDir(abs)) continue
    for (const d of readdirSync(abs).sort()) {
      const p = join(abs, d, "SKILL.md")
      if (existsSync(p)) load(p, "skill", d)
    }
  }
  const agentsMd = join(root, "AGENTS.md")
  if (existsSync(agentsMd)) load(agentsMd, "context", "AGENTS")
  return out
}

// Replace inline code spans and HTML comments with spaces so prose
// checks ignore them while column offsets stay stable.
function maskProse(line: string): string {
  const blank = (s: string): string => " ".repeat(s.length)
  return line.replace(/`[^`]*`/g, blank).replace(/<!--.*?-->/g, blank)
}

interface Unit {
  line: number // 1-based line of first text line
  raw: string
  masked: string
}

const LIST_ITEM = /^\s*(?:[-*+]|\d+[.)])\s+/

// Split body prose into units: paragraphs, with each list item and
// heading starting a new unit. Fenced code is excluded.
function proseUnits(a: Artifact): Unit[] {
  const units: Unit[] = []
  let cur: { line: number; raw: string[]; masked: string[] } | null = null
  const flush = (): void => {
    if (cur) units.push({ line: cur.line, raw: cur.raw.join("\n"), masked: cur.masked.join("\n") })
    cur = null
  }
  for (let i = a.bodyStart; i < a.lines.length; i++) {
    const line = a.lines[i]
    if (a.fenced[i] || line.trim() === "") {
      flush()
      continue
    }
    if (/^\s{0,3}#/.test(line) || LIST_ITEM.test(line)) flush()
    if (!cur) cur = { line: i + 1, raw: [], masked: [] }
    cur.raw.push(line)
    cur.masked.push(maskProse(line))
    if (/^\s{0,3}#/.test(line)) flush()
  }
  flush()
  return units
}

function lineAt(unit: Unit, offset: number): number {
  let n = 0
  for (let i = 0; i < offset; i++) if (unit.masked.charCodeAt(i) === 10) n++
  return unit.line + n
}

interface Sentence {
  start: number
  text: string
}

function sentences(text: string): Sentence[] {
  const out: Sentence[] = []
  let start = 0
  for (const m of text.matchAll(/[.!?]+(?=\s+[A-Z`*"(\[]|\s*$)/g)) {
    const end = (m.index ?? 0) + m[0].length
    out.push({ start, text: text.slice(start, end) })
    start = end
  }
  if (start < text.length && text.slice(start).trim() !== "") out.push({ start, text: text.slice(start) })
  return out
}

// ---------------------------------------------------------------------
// Pack rules
// ---------------------------------------------------------------------

interface PackRule {
  id: string
  marker: string | null
  line: number
  body: string
}

const RULE_DEF = /^(\s*)[-*]\s+\*\*([A-Z]{2,5}-\d{3})\*\*\s*(?:\[(MUST|SHOULD|MAY)\])?/

export function parsePackRules(a: Artifact): PackRule[] {
  const rules: PackRule[] = []
  for (let i = a.bodyStart; i < a.lines.length; i++) {
    if (a.fenced[i]) continue
    const m = RULE_DEF.exec(a.lines[i])
    if (!m) continue
    const indent = m[1].length
    const body = [a.lines[i].slice(m[0].length)]
    for (let j = i + 1; j < a.lines.length; j++) {
      const l = a.lines[j]
      if (a.fenced[j] || l.trim() === "" || /^\s{0,3}#/.test(l)) break
      const li = LIST_ITEM.exec(l)
      if (li && l.search(/\S/) <= indent) break
      body.push(l)
    }
    rules.push({ id: m[2], marker: m[3] ?? null, line: i + 1, body: body.map(maskProse).join("\n") })
  }
  return rules
}

// ---------------------------------------------------------------------
// References
// ---------------------------------------------------------------------

type RefKind = "command" | "agent" | "skill" | "path"

interface Ref {
  kind: RefKind
  name: string
  line: number
}

function collectRefs(a: Artifact): Ref[] {
  const refs: Ref[] = []
  const fm = a.frontmatter
  if (fm && a.kind === "command") {
    if (typeof fm.agent === "string") refs.push({ kind: "agent", name: fm.agent, line: 1 })
    if (Array.isArray(fm.handoffs)) {
      for (const h of fm.handoffs) {
        if (h && typeof h === "object" && typeof (h as Record<string, unknown>).agent === "string") {
          // Handoffs may target an agent or another command.
          refs.push({ kind: "agent", name: (h as Record<string, string>).agent, line: 1 })
        }
      }
    }
  }
  for (let i = a.bodyStart; i < a.lines.length; i++) {
    const line = a.lines[i]
    const ln = i + 1
    if (a.fenced[i]) {
      const m = /^\s*\/([a-z][a-z0-9_.-]*[a-z0-9])(?=\s|$)/.exec(line)
      if (m) refs.push({ kind: "command", name: m[1], line: ln })
      continue
    }
    for (const m of line.matchAll(/`\/([a-z][a-z0-9_.-]*[a-z0-9])(?=[`\s])/g)) {
      refs.push({ kind: "command", name: m[1], line: ln })
    }
    for (const m of line.matchAll(/`([a-z0-9][a-z0-9_.-]*)`\s+(?:sub)?(?:agent|persona)\b/g)) {
      refs.push({ kind: "agent", name: m[1], line: ln })
    }
    for (const m of line.matchAll(/subagent_type:\s*["'`]([a-z0-9][a-z0-9_.-]*)["'`]/g)) {
      refs.push({ kind: "agent", name: m[1], line: ln })
    }
    for (const m of line.matchAll(/`([a-z0-9][a-z0-9-]*)`\s+skill\b/g)) {
      refs.push({ kind: "skill", name: m[1], line: ln })
    }
    for (const m of line.matchAll(/skill\(\s*\{\s*name:\s*["']([a-z0-9-]+)["']/g)) {
      refs.push({ kind: "skill", name: m[1], line: ln })
    }
    for (const m of line.matchAll(/`(\.opencode\/[^`\s]+)`/g)) {
      // Only concrete files: skip globs, placeholders, and directories
      // (prose legitimately names legacy or not-yet-created dirs).
      if (/[*{}<>$]/.test(m[1]) || !/\.[A-Za-z0-9]+(?::\d+(?:-\d+)?|#.*)?$/.test(m[1])) continue
      refs.push({ kind: "path", name: m[1].replace(/(?::\d+(?:-\d+)?|#.*)$/, ""), line: ln })
    }
  }
  return refs
}

// ---------------------------------------------------------------------
// Lint
// ---------------------------------------------------------------------

const NORMATIVE = /\b(?:MUST(?: NOT)?|SHOULD(?: NOT)?|SHALL(?: NOT)?|REQUIRED|NEVER|ALWAYS)\b/
const VAGUE = new RegExp(
  "\\b(?:" + [
    "as appropriate", "appropriate(?:ly)?", "as needed", "if needed", "if necessary",
    "when necessary", "where possible", "if possible", "where applicable", "if applicable",
    "relevant", "reasonabl[ey]", "properly", "sufficient(?:ly)?", "adequate(?:ly)?",
    "etc", "and so on", "best effort", "as much as possible", "significant(?:ly)?",
  ].join("|") + ")\\b",
  "gi",
)
const PRESCRIBED_OUTPUT =
  /\b(?:exactly|verbatim)\b[^.\n]{0,40}\b(?:format|template|output)\b|\b(?:format|template|output)\b[^.\n]{0,40}\b(?:exactly|verbatim)\b|\buse (?:this|the following) (?:exact )?(?:format|template)\b/i
const SHELL_LANGS = new Set(["", "bash", "sh", "shell", "zsh", "console"])
const SHELL_CMD = /^\s*(?:\$\s+)?(?:git|gh|grep|rg|find|ls|jq|yq|sed|awk|cat|head|tail|mkdir|rm|cp|mv|test|make|go|npm|npx|node|curl|wc|sort)\b/

type Emit = (rule: string, a: Artifact, line: number, message: string) => void

function checkFrontmatter(a: Artifact, emit: Emit): void {
  if (a.kind === "context") return
  if (!a.hasFrontmatter) {
    emit("IL-001", a, 1, "file has no YAML frontmatter")
    return
  }
  if (a.frontmatterError || !a.frontmatter) {
    emit("IL-001", a, 1, a.frontmatterError ?? "frontmatter is empty")
    return
  }
  const fm = a.frontmatter
  const str = (k: string): string | null => (typeof fm[k] === "string" && (fm[k] as string).trim() !== "" ? (fm[k] as string).trim() : null)
  const desc = str("description")
  if (a.kind !== "pack" && !desc) emit("IL-002", a, 1, "`description` is required")

  if (a.kind === "agent") {
    if (fm.mode !== undefined && !["primary", "subagent", "all"].includes(fm.mode as string)) {
      emit("IL-002", a, 1, `\`mode\` must be primary, subagent, or all (got ${JSON.stringify(fm.mode)})`)
    }
    if (fm.temperature !== undefined && (typeof fm.temperature !== "number" || fm.temperature < 0 || fm.temperature > 2)) {
      emit("IL-002", a, 1, "`temperature` must be a number between 0 and 2")
    }
  }
  if (a.kind === "skill") {
    const name = str("name")
    if (!name) emit("IL-002", a, 1, "`name` is required")
    else if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(name)) emit("IL-002", a, 1, `\`name\` must be lowercase kebab-case (got "${name}")`)
    else if (name !== a.name) emit("IL-003", a, 1, `\`name\` "${name}" does not match directory "${a.name}"`)
    if (desc && desc.length > 1024) emit("IL-004", a, 1, `description is ${desc.length} chars; the skills spec limit is 1024`)
  }
  if (a.kind === "pack") {
    const id = str("pack_id")
    if (!id) {
      if (!desc) emit("IL-002", a, 1, "pack needs `pack_id` (rule pack) or `description` (shared definitions)")
    } else {
      if (id !== a.name) emit("IL-003", a, 1, `\`pack_id\` "${id}" does not match file name "${a.name}"`)
      if (!str("language")) emit("IL-002", a, 1, "`language` is required when `pack_id` is set")
      const v = fm.version === undefined ? null : String(fm.version)
      if (!v || !/^\d+\.\d+\.\d+$/.test(v)) emit("IL-002", a, 1, "`version` must be semver MAJOR.MINOR.PATCH")
    }
  }
  if (desc && a.kind !== "pack" && desc.length < 20) {
    emit("IL-004", a, 1, `description is ${desc.length} chars; too short to route on`)
  }
}

function checkPackRules(packs: Artifact[], emit: Emit): Set<string> {
  const defined = new Set<string>()
  for (const a of packs) {
    const seen = new Map<string, number>()
    for (const r of parsePackRules(a)) {
      defined.add(r.id)
      const prev = seen.get(r.id)
      if (prev !== undefined) emit("IL-014", a, r.line, `${r.id} already defined at line ${prev}`)
      else seen.set(r.id, r.line)

      if (!r.marker) {
        emit("IL-020", a, r.line, `${r.id} has no [MUST]/[SHOULD]/[MAY] marker`)
      } else {
        const strong = /\b(?:MUST|SHALL|REQUIRED)\b/.test(r.body)
        // An imperative "Never ..." carries the [MUST] marker's force.
        const prohibits = strong || /\bnever\b/i.test(r.body)
        const should = /\b(?:SHOULD|RECOMMENDED)\b/.test(r.body)
        if (r.marker === "MUST" && !prohibits && should) emit("IL-021", a, r.line, `${r.id} is [MUST] but its body only says SHOULD`)
        if (r.marker === "SHOULD" && strong && !should) emit("IL-021", a, r.line, `${r.id} is [SHOULD] but its body says MUST`)
        if (r.marker === "MAY" && strong) emit("IL-021", a, r.line, `${r.id} is [MAY] but its body says MUST`)
      }
      const lower = /\b(must|should|shall)\b/.exec(r.body)
      if (lower) {
        const off = lower.index
        const line = r.line + (r.body.slice(0, off).match(/\n/g)?.length ?? 0)
        emit("IL-022", a, line, `${r.id}: lowercase "${lower[1]}" — uppercase it for RFC 2119 force or rephrase`)
      }
    }
  }
  return defined
}

function checkRuleIdRefs(all: Artifact[], defined: Set<string>, emit: Emit): void {
  const prefixes = new Set([...defined].map((id) => id.split("-")[0]))
  if (prefixes.size === 0) return
  for (const a of all) {
    if (a.kind === "command") continue // command templates use speckit IDs (FR-001, SC-001)
    for (let i = a.bodyStart; i < a.lines.length; i++) {
      if (a.fenced[i]) continue
      for (const m of a.lines[i].matchAll(/\b([A-Z]{2,5})-(\d{3})\b/g)) {
        if (prefixes.has(m[1]) && !defined.has(m[0])) {
          emit("IL-015", a, i + 1, `${m[0]} is not defined in any pack`)
        }
      }
    }
  }
}

function checkProse(a: Artifact, emit: Emit): void {
  const procedural = a.kind === "agent" || a.kind === "command" || a.kind === "skill"
  for (const u of proseUnits(a)) {
    for (const s of sentences(u.masked)) {
      if (!NORMATIVE.test(s.text)) continue
      const seen = new Set<string>()
      for (const v of s.text.matchAll(VAGUE)) {
        const word = v[0].toLowerCase()
        if (seen.has(word)) continue
        seen.add(word)
        emit("IL-023", a, lineAt(u, s.start + (v.index ?? 0)), `"${v[0]}" in a normative sentence; state the concrete condition or criterion`)
      }
    }
    if (procedural) {
      const p = PRESCRIBED_OUTPUT.exec(u.masked)
      if (p) emit("IL-040", a, lineAt(u, p.index), `"${p[0].trim()}" — prescribed output should be tool-rendered (DE-003)`)
    }
  }
  if (!procedural) return
  // Walk opener -> closer so each block is visited once.
  for (let i = a.bodyStart; i < a.lines.length; i++) {
    const open = /^\s{0,3}(`{3,}|~{3,})\s*([A-Za-z]*)/.exec(a.lines[i])
    if (!open) continue
    const closes = (l: string): boolean => {
      const t = l.trim()
      return t.length >= open[1].length && t === open[1][0].repeat(t.length)
    }
    let n = 0
    let j = i + 1
    for (; j < a.lines.length && !closes(a.lines[j]); j++) {
      if (SHELL_CMD.test(a.lines[j])) n++
    }
    if (SHELL_LANGS.has(open[2].toLowerCase()) && n >= 3) {
      emit("IL-041", a, i + 1, `fenced block runs ${n} shell commands; extract to a tested tool (DE-001)`)
    }
    i = j
  }
}

// DE-008: delegation tools whose parameter contract binds deterministic
// work to repo tooling. `optionalWith` names the parameter that, when
// present, makes task_classification optional (uf_delegate_command
// accepts unclassified command-path delegations).
const DELEGATION_TOOLS: Record<string, { optionalWith: string | null }> = {
  uf_delegate_command: { optionalWith: "command_path" },
  replicator_forge_spawn_subtask: { optionalWith: null },
}
const DELEGATION_CALL = new RegExp(`^\\s*(${Object.keys(DELEGATION_TOOLS).join("|")})\\b\\s*(?:\\(|\\{|$)`)
const PARAM_KEY = /(?:^|[\s{,(])["']?([a-z_][a-z0-9_]*)["']?\s*:/g

function checkDelegationCall(a: Artifact, tool: string, line: number, text: string, emit: Emit): void {
  const keys = new Map<string, number>()
  for (const m of text.matchAll(PARAM_KEY)) if (!keys.has(m[1])) keys.set(m[1], m.index ?? 0)
  const { optionalWith } = DELEGATION_TOOLS[tool]
  const at = keys.get("task_classification")
  if (at === undefined) {
    if (optionalWith && keys.has(optionalWith)) return
    const unless = optionalWith ? ` (or \`${optionalWith}\`)` : ""
    emit("IL-042", a, line, `${tool} call passes no task_classification${unless}; the tool fails closed without it (DE-008)`)
    return
  }
  // The classification value runs to the next parameter key.
  const after = [...keys.values()].filter((o) => o > at).sort((x, y) => x - y)[0] ?? text.length
  const value = text.slice(at, after)
  if (/deterministic/i.test(value) && !keys.has("skill") && !keys.has("skill_absent_reason")) {
    emit("IL-042", a, line, `${tool} call can classify work deterministic but passes neither skill nor skill_absent_reason (DE-008)`)
  }
}

// Scan fenced blocks for delegation call sites: a line that starts with
// a delegation tool name begins a call, which runs to the next call or
// the end of the block. Prose mentions of the tools are not call sites.
function checkDelegations(a: Artifact, emit: Emit): void {
  if (a.kind !== "agent" && a.kind !== "command" && a.kind !== "skill") return
  let call: { tool: string; line: number; text: string[] } | null = null
  const flush = (): void => {
    if (call) checkDelegationCall(a, call.tool, call.line, call.text.join("\n"), emit)
    call = null
  }
  for (let i = a.bodyStart; i < a.lines.length; i++) {
    if (!a.fenced[i] || /^\s{0,3}(`{3,}|~{3,})/.test(a.lines[i])) {
      flush()
      continue
    }
    const m = DELEGATION_CALL.exec(a.lines[i])
    if (m) {
      flush()
      const line = a.lines[i]
      call = { tool: m[1], line: i + 1, text: [line.slice(line.indexOf(m[1]) + m[1].length)] }
    } else if (call) {
      call.text.push(a.lines[i])
    }
  }
  flush()
}

function shingles(text: string): Set<string> {
  const words = text.toLowerCase().match(/[a-z0-9']+/g) ?? []
  const out = new Set<string>()
  for (let i = 0; i + 5 <= words.length; i++) out.add(words.slice(i, i + 5).join(" "))
  return out
}

function checkDuplicates(all: Artifact[], opts: LintOptions, emit: Emit): void {
  interface Para { a: Artifact; line: number; sh: Set<string> }
  const paras: Para[] = []
  for (const a of all) {
    for (const u of proseUnits(a)) {
      if ((u.raw.match(/\S+/g)?.length ?? 0) < opts.duplicateMinWords) continue
      paras.push({ a, line: u.line, sh: shingles(u.raw) })
    }
  }
  const index = new Map<string, number[]>()
  paras.forEach((p, idx) => {
    for (const s of p.sh) {
      const list = index.get(s) ?? []
      list.push(idx)
      index.set(s, list)
    }
  })
  // Union-find paragraphs into clusters of near-duplicates, then report
  // one finding per distinct set of files sharing copied text.
  const parent = paras.map((_, i) => i)
  const find = (x: number): number => (parent[x] === x ? x : (parent[x] = find(parent[x])))
  paras.forEach((p, i) => {
    const cand = new Set<number>()
    for (const s of p.sh) for (const j of index.get(s) ?? []) if (j > i && paras[j].a !== p.a) cand.add(j)
    for (const j of cand) {
      const q = paras[j]
      let inter = 0
      for (const s of p.sh) if (q.sh.has(s)) inter++
      if (inter / (p.sh.size + q.sh.size - inter) >= opts.duplicateThreshold) parent[find(j)] = find(i)
    }
  })
  const clusters = new Map<number, Para[]>()
  paras.forEach((p, i) => {
    const r = find(i)
    clusters.set(r, [...(clusters.get(r) ?? []), p])
  })
  const groups = new Map<string, { first: Para; others: Para[]; count: number }>()
  for (const members of clusters.values()) {
    const files = [...new Set(members.map((m) => m.a.file))].sort()
    if (files.length < 2) continue
    const ordered = [...members].sort((x, y) => x.a.file.localeCompare(y.a.file) || x.line - y.line)
    const key = files.join("\0")
    const g = groups.get(key)
    if (g) g.count++
    else groups.set(key, { first: ordered[0], others: ordered.filter((m) => m.a !== ordered[0].a), count: 1 })
  }
  for (const g of groups.values()) {
    const firstPerFile = new Map<string, string>()
    for (const o of g.others) if (!firstPerFile.has(o.a.file)) firstPerFile.set(o.a.file, `${o.a.file}:${o.line}`)
    const where = [...firstPerFile.values()].join(", ")
    emit("IL-050", g.first.a, g.first.line, `${g.count} paragraph(s) also appear near-verbatim in ${where}; keep one source of truth`)
  }
}

export interface LintResult {
  findings: Finding[]
  artifacts: Artifact[]
  loadTokens: Map<string, number>
}

export function lint(
  root: string,
  artifacts: Artifact[],
  options: Partial<LintOptions> = {},
  exists: (rel: string) => boolean = (rel) => existsSync(join(root, rel)),
): LintResult {
  const opts: LintOptions = { ...DEFAULT_OPTIONS, ...options }
  const findings: Finding[] = []
  const emit: Emit = (rule, a, line, message) => {
    if (opts.disable.includes(rule)) return
    if (opts.ignorePaths.some((p) => a.file.startsWith(p))) return
    if (a.fileSuppression.all || a.fileSuppression.rules.has(rule)) return
    const ls = a.lineSuppression.get(line)
    if (ls && (ls.all || ls.rules.has(rule))) return
    const severity = RULES[rule].severity
    findings.push({ rule, severity, reviewSeverity: REVIEW_SEVERITY[severity], file: a.file, line, message })
  }

  const byKind = (k: ArtifactKind): Map<string, Artifact> =>
    new Map(artifacts.filter((a) => a.kind === k).map((a) => [a.name, a]))
  const commands = byKind("command")
  const agents = byKind("agent")
  const skills = byKind("skill")
  const byFile = new Map(artifacts.map((a) => [a.file, a]))
  const context = artifacts.filter((a) => a.kind === "context")
  const contextTokens = context.reduce((n, a) => n + a.tokens, 0)
  const loadTokens = new Map<string, number>()

  for (const a of artifacts) {
    checkFrontmatter(a, emit)
    checkProse(a, emit)
    checkDelegations(a, emit)
    if (a.tokens > opts.maxFileTokens) {
      emit("IL-030", a, 1, `~${a.tokens} tokens exceeds the ${opts.maxFileTokens}-token file budget`)
    }

    const loaded = new Map<string, Artifact>()
    for (const r of collectRefs(a)) {
      if (r.kind === "command" && !commands.has(r.name) && !BUILTIN_COMMANDS.has(r.name)) {
        emit("IL-010", a, r.line, `/${r.name} is not a command in .opencode/commands/`)
      } else if (r.kind === "agent" && !agents.has(r.name) && !commands.has(r.name) && !BUILTIN_AGENTS.has(r.name)) {
        emit("IL-011", a, r.line, `agent "${r.name}" is not defined in .opencode/agents/`)
      } else if (r.kind === "skill") {
        const s = skills.get(r.name)
        if (s) loaded.set(s.file, s)
        else emit("IL-012", a, r.line, `skill "${r.name}" is not defined in .opencode/skills/`)
      } else if (r.kind === "path") {
        if (!exists(r.name)) emit("IL-013", a, r.line, `${r.name} does not exist`)
        else {
          const target = byFile.get(r.name)
          if (target && (target.kind === "pack" || target.kind === "skill")) loaded.set(target.file, target)
        }
      }
    }
    if (a.kind === "command" || a.kind === "agent") {
      loaded.delete(a.file)
      const parts = [...loaded.values()]
      const total = contextTokens + a.tokens + parts.reduce((n, p) => n + p.tokens, 0)
      loadTokens.set(a.file, total)
      if (total > opts.maxLoadTokens) {
        const top = parts.sort((x, y) => y.tokens - x.tokens).slice(0, 3).map((p) => `${p.file} ~${p.tokens}`)
        const detail = [`self ~${a.tokens}`, ...(contextTokens ? [`AGENTS.md ~${contextTokens}`] : []), ...top].join(", ")
        emit("IL-031", a, 1, `~${total} tokens loaded exceeds the ${opts.maxLoadTokens}-token budget (${detail})`)
      }
    }
  }

  const defined = checkPackRules(artifacts.filter((a) => a.kind === "pack"), emit)
  checkRuleIdRefs(artifacts, defined, emit)
  checkDuplicates(artifacts, opts, emit)

  const rank: Record<Severity, number> = { error: 0, warning: 1, info: 2 }
  findings.sort((x, y) => x.file.localeCompare(y.file) || x.line - y.line || rank[x.severity] - rank[y.severity] || x.rule.localeCompare(y.rule))
  return { findings, artifacts, loadTokens }
}

// ---------------------------------------------------------------------
// Base-revision deltas (DE-004 evidence)
// ---------------------------------------------------------------------

export function countLines(text: string): number {
  if (text === "") return 0
  const n = text.split("\n").length
  return text.endsWith("\n") ? n - 1 : n
}

export interface BaseTree {
  ref: string
  mergeBase: string
  artifacts: Artifact[]
  // every path under .opencode/ (and AGENTS.md) at the merge base
  paths: Set<string>
}

const git = (root: string, args: string[], input?: string): Buffer =>
  execFileSync("git", args, { cwd: root, input, maxBuffer: 256 * 1024 * 1024, stdio: ["pipe", "pipe", "pipe"] })

// Read the instruction artifacts at merge-base(ref, HEAD), so a base
// that moved ahead of the branch does not count its own changes.
export function readBaseTree(root: string, ref: string): BaseTree {
  const mergeBase = git(root, ["merge-base", ref, "HEAD"]).toString("utf8").trim()
  const paths = git(root, ["ls-tree", "-r", "-z", "--name-only", mergeBase, "--", ".opencode", "AGENTS.md"])
    .toString("utf8").split("\0").filter(Boolean)
  const wanted = paths.map((p) => ({ p, c: classifyInstructionPath(p) })).filter((x) => x.c !== null)
  const artifacts: Artifact[] = []
  if (wanted.length > 0) {
    const out = git(root, ["cat-file", "--batch"], wanted.map((w) => `${mergeBase}:${w.p}\n`).join(""))
    let off = 0
    for (const w of wanted) {
      const nl = out.indexOf(10, off)
      const size = Number(out.subarray(off, nl).toString("utf8").split(" ")[2])
      const text = out.subarray(nl + 1, nl + 1 + size).toString("utf8")
      off = nl + 1 + size + 1
      artifacts.push(buildArtifact(w.p, text, w.c!.kind, w.c!.name))
    }
  }
  return { ref, mergeBase, artifacts, paths: new Set(paths) }
}

export interface Delta {
  file: string
  status: "added" | "deleted" | "modified" | "load-only"
  lines: { base: number; head: number; delta: number }
  tokens: { base: number; head: number; delta: number }
  // null for artifacts with no load set (skills, packs, AGENTS.md)
  loadTokens: { base: number; head: number; delta: number } | null
}

const pair = (base: number, head: number) => ({ base, head, delta: head - base })

export function computeDeltas(base: LintResult, head: LintResult): Delta[] {
  const b = new Map(base.artifacts.map((a) => [a.file, a]))
  const h = new Map(head.artifacts.map((a) => [a.file, a]))
  const out: Delta[] = []
  for (const file of [...new Set([...b.keys(), ...h.keys()])].sort()) {
    const ba = b.get(file)
    const ha = h.get(file)
    const text = (a: Artifact | undefined): string => (a ? a.lines.join("\n") : "")
    const bl = base.loadTokens.get(file)
    const hl = head.loadTokens.get(file)
    const loadTokens = bl === undefined && hl === undefined ? null : pair(bl ?? 0, hl ?? 0)
    let status: Delta["status"]
    if (!ba) status = "added"
    else if (!ha) status = "deleted"
    else if (text(ba) !== text(ha)) status = "modified"
    else if (loadTokens && loadTokens.delta !== 0) status = "load-only"
    else continue
    out.push({
      file,
      status,
      lines: pair(ba ? countLines(text(ba)) : 0, ha ? countLines(text(ha)) : 0),
      tokens: pair(ba?.tokens ?? 0, ha?.tokens ?? 0),
      loadTokens,
    })
  }
  return out
}

const signed = (n: number): string => (n > 0 ? `+${n}` : String(n))

// ---------------------------------------------------------------------
// CLI
// ---------------------------------------------------------------------

export interface CliIO {
  cwd: string
  out: (s: string) => void
  err: (s: string) => void
}

const USAGE = `usage: instruction-lint [--root DIR] [--config FILE] [--only PATH]... [--json] [--strict] [--all] [--report] [--base REF] [--rules]

  --root DIR     repository root (default: nearest ancestor containing .opencode/)
  --config FILE  JSON options (default: <root>/.opencode/instruction-lint.json if present)
  --only PATH    report only findings in these repo-relative files (repeatable or
                 comma-separated); the whole repo is still analyzed for references,
                 rule IDs, load sets, and duplicates
  --json         machine-readable output
  --strict       exit 1 on warnings as well as errors
  --all          show info findings (hidden by default, always counted)
  --report       print per-file token and load-set sizes
  --base REF     also report line, token, and load-set deltas for instruction files
                 changed since merge-base(REF, HEAD) (DE-004 evidence)
  --rules        list rule IDs and exit`

export function findRoot(start: string): string | null {
  let dir = resolve(start)
  for (;;) {
    if (isDir(join(dir, ".opencode"))) return dir
    const parent = resolve(dir, "..")
    if (parent === dir) return null
    dir = parent
  }
}

export function main(argv: string[], io: CliIO): number {
  const flags = new Set<string>()
  let root: string | null = null
  let config: string | null = null
  let baseRef: string | null = null
  const only = new Set<string>()
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i]
    if (a === "--root" || a === "--config" || a === "--only" || a === "--base") {
      const v = argv[++i]
      if (!v) {
        io.err(`${a} needs a value\n${USAGE}`)
        return 2
      }
      if (a === "--root") root = resolve(io.cwd, v)
      else if (a === "--config") config = resolve(io.cwd, v)
      else if (a === "--base") baseRef = v
      else for (const p of v.split(",")) if (p.trim()) only.add(p.trim().replace(/^\.\//, ""))
    } else if (["--json", "--strict", "--all", "--report", "--rules", "--help"].includes(a)) {
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
  if (flags.has("--rules")) {
    for (const [id, r] of Object.entries(RULES)) io.out(`${id}  ${r.severity.padEnd(7)}  ${r.summary}`)
    return 0
  }
  root = root ?? findRoot(io.cwd)
  if (!root) {
    io.err("no .opencode/ directory found; pass --root")
    return 2
  }
  let options: Partial<LintOptions> = {}
  const cfg = config ?? join(root, ".opencode", "instruction-lint.json")
  if (config || existsSync(cfg)) {
    try {
      options = JSON.parse(readFileSync(cfg, "utf8")) as Partial<LintOptions>
    } catch (e) {
      io.err(`cannot read config ${cfg}: ${(e as Error).message}`)
      return 2
    }
  }

  const result = lint(root, discoverArtifacts(root), options)
  let delta: { base: string; mergeBase: string; files: Delta[]; totals: { lines: number; tokens: number } } | null = null
  if (baseRef !== null) {
    let tree: BaseTree
    try {
      tree = readBaseTree(root, baseRef)
    } catch (e) {
      const detail = (e as { stderr?: Buffer }).stderr?.toString("utf8").trim() || (e as Error).message
      io.err(`cannot read base ${baseRef}: ${detail}`)
      return 2
    }
    const baseResult = lint(root, tree.artifacts, options, (rel) => tree.paths.has(rel))
    let files = computeDeltas(baseResult, result)
    if (only.size > 0) files = files.filter((d) => only.has(d.file))
    const sum = (k: "lines" | "tokens"): number => files.reduce((n, d) => n + d[k].delta, 0)
    delta = { base: baseRef, mergeBase: tree.mergeBase, files, totals: { lines: sum("lines"), tokens: sum("tokens") } }
  }
  if (only.size > 0) result.findings = result.findings.filter((f) => only.has(f.file))
  const count = (s: Severity): number => result.findings.filter((f) => f.severity === s).length
  const summary = { files: result.artifacts.length, errors: count("error"), warnings: count("warning"), info: count("info") }

  if (flags.has("--json")) {
    const loadTokens = Object.fromEntries(result.loadTokens)
    const tokens = Object.fromEntries(result.artifacts.map((a) => [a.file, a.tokens]))
    io.out(JSON.stringify({ summary, findings: result.findings, tokens, loadTokens, ...(delta ? { delta } : {}) }, null, 2))
  } else {
    for (const f of result.findings) {
      if (f.severity === "info" && !flags.has("--all")) continue
      io.out(`${f.file}:${f.line}  ${f.severity.padEnd(7)}  ${f.rule}  ${f.message}`)
    }
    if (flags.has("--report")) {
      io.out("\ntokens  load    file")
      const rows = [...result.artifacts].sort((x, y) => y.tokens - x.tokens)
      for (const a of rows) {
        const load = result.loadTokens.get(a.file)
        io.out(`${String(a.tokens).padStart(6)}  ${(load === undefined ? "-" : String(load)).padStart(6)}  ${a.file}`)
      }
    }
    if (delta) {
      io.out(`\ndeltas since ${delta.base} (merge-base ${delta.mergeBase.slice(0, 12)})`)
      io.out(`${"lines".padEnd(14)} ${"tokens".padEnd(14)} ${"load".padEnd(14)} ${"status".padEnd(10)} file`)
      const cell = (p: { base: number; head: number; delta: number } | null): string =>
        (p ? `${p.head} (${signed(p.delta)})` : "-").padEnd(14)
      for (const d of delta.files) io.out(`${cell(d.lines)} ${cell(d.tokens)} ${cell(d.loadTokens)} ${d.status.padEnd(10)} ${d.file}`)
      io.out(`total: ${signed(delta.totals.lines)} lines, ${signed(delta.totals.tokens)} tokens across ${delta.files.length} files`)
    }
    const hidden = flags.has("--all") ? "" : " (use --all to show info)"
    io.out(`\n${summary.files} files: ${summary.errors} errors, ${summary.warnings} warnings, ${summary.info} info${summary.info ? hidden : ""}`)
  }
  return summary.errors > 0 || (flags.has("--strict") && summary.warnings > 0) ? 1 : 0
}

if (import.meta.main) {
  process.exitCode = main(process.argv.slice(2), {
    cwd: process.cwd(),
    out: (s) => process.stdout.write(s + "\n"),
    err: (s) => process.stderr.write(s + "\n"),
  })
}
