import { execFileSync } from "node:child_process"
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs"
import { tmpdir } from "node:os"
import { dirname, join } from "node:path"
import { afterEach, describe, expect, it } from "vitest"
import {
  buildArtifact,
  classifyInstructionPath,
  countLines,
  discoverArtifacts,
  estimateTokens,
  findRoot,
  lint,
  main,
  parsePackRules,
  RULES,
  type Delta,
  type Finding,
} from "../lib/instruction-lint.js"

const roots: string[] = []

function repo(files: Record<string, string>): string {
  const root = mkdtempSync(join(tmpdir(), "instruction-lint-"))
  roots.push(root)
  mkdirSync(join(root, ".opencode"), { recursive: true })
  for (const [rel, text] of Object.entries(files)) {
    const abs = join(root, rel)
    mkdirSync(dirname(abs), { recursive: true })
    writeFileSync(abs, text)
  }
  return root
}

afterEach(() => {
  for (const r of roots.splice(0)) rmSync(r, { recursive: true, force: true })
})

const AGENT = (body = "", fm = "description: \"Reviews code for the project council.\"\nmode: subagent") =>
  `---\n${fm}\n---\n\n${body}\n`
const COMMAND = (body = "", fm = "description: Run the thing end to end for a branch.") => `---\n${fm}\n---\n\n${body}\n`
const SKILL = (name: string, body = "") =>
  `---\nname: ${name}\ndescription: "Use when doing ${name} work in this repository."\n---\n\n${body}\n`
const PACK = (id: string, body: string) => `---\npack_id: ${id}\nlanguage: Any\nversion: 1.0.0\n---\n\n${body}\n`

function run(files: Record<string, string>, opts = {}): Finding[] {
  const root = repo(files)
  return lint(root, discoverArtifacts(root), opts).findings
}

const rules = (fs: Finding[]): string[] => fs.map((f) => f.rule)

describe("parsing", () => {
  it("estimates tokens at ~4 chars per token", () => {
    expect(estimateTokens("abcdefgh")).toBe(2)
    expect(estimateTokens("abcde")).toBe(2)
  })

  it("parses frontmatter, normalizes CRLF, and tracks fences", () => {
    const a = buildArtifact("x.md", "---\r\nname: x\r\n---\r\ntext\r\n~~~\r\nin\r\n~~~\r\n```js\r\ncode\r\n```\r\nafter", "skill", "x")
    expect(a.frontmatter).toEqual({ name: "x" })
    expect(a.bodyStart).toBe(3)
    expect(a.fenced).toEqual([false, false, false, false, true, true, true, true, true, true, false])
  })

  it("does not close a fence on a different or shorter marker", () => {
    const a = buildArtifact("x.md", "````\n```\n~~~\n````\nout", "context", "x")
    expect(a.fenced).toEqual([true, true, true, true, false])
  })

  it("reports unterminated, invalid, and non-mapping frontmatter", () => {
    expect(buildArtifact("a", "---\nname: x\n", "skill", "a").frontmatterError).toMatch(/closing/)
    expect(buildArtifact("a", "---\nname: [x\n---\n", "skill", "a").frontmatterError).toMatch(/invalid YAML/)
    expect(buildArtifact("a", "---\n- x\n---\n", "skill", "a").frontmatterError).toMatch(/not a YAML mapping/)
    expect(buildArtifact("a", "no frontmatter", "skill", "a").hasFrontmatter).toBe(false)
  })

  it("parses pack rules with markers and continuation lines", () => {
    const a = buildArtifact("p.md", PACK("p", "- **AB-001** [MUST] Do the thing\n  carefully `must` here.\n- **AB-002** No marker\n  - nested detail\n\n```\n- **AB-003** [MAY] fenced\n```"), "pack", "p")
    const r = parsePackRules(a)
    expect(r.map((x) => [x.id, x.marker])).toEqual([["AB-001", "MUST"], ["AB-002", null]])
    expect(r[0].body).toContain("carefully")
    expect(r[0].body).not.toContain("must")
    expect(r[1].body).toContain("nested detail")
  })
})

describe("frontmatter rules", () => {
  it("requires frontmatter on every artifact except AGENTS.md", () => {
    const fs = run({
      ".opencode/agents/a.md": "# no frontmatter",
      ".opencode/commands/c.md": "---\nfoo: [\n---\n",
      ".opencode/commands/d.md": "---\n---\n",
      "AGENTS.md": "# context only",
    })
    expect(fs.filter((f) => f.rule === "IL-001").map((f) => f.file)).toEqual([
      ".opencode/agents/a.md", ".opencode/commands/c.md", ".opencode/commands/d.md",
    ])
  })

  it("validates agent fields", () => {
    const fs = run({
      ".opencode/agents/a.md": AGENT("", "mode: sidekick\ntemperature: 3"),
      ".opencode/agents/b.md": AGENT("", "description: tiny\ntemperature: hot"),
    })
    const msgs = fs.map((f) => `${f.rule} ${f.message}`)
    expect(msgs).toContain("IL-002 `description` is required")
    expect(msgs.some((m) => m.includes("`mode` must be"))).toBe(true)
    expect(msgs.filter((m) => m.includes("`temperature`"))).toHaveLength(2)
    expect(msgs.some((m) => m.startsWith("IL-004") && m.includes("too short"))).toBe(true)
  })

  it("validates skill name and description", () => {
    const fs = run({
      ".opencode/skills/good/SKILL.md": SKILL("good"),
      ".opencode/skills/mismatch/SKILL.md": SKILL("other"),
      ".opencode/skills/caps/SKILL.md": SKILL("Caps_Name"),
      ".opencode/skills/noname/SKILL.md": "---\ndescription: Use when there is no name given at all.\n---\n",
      ".opencode/skill/long/SKILL.md": `---\nname: long\ndescription: "${"x".repeat(1100)}"\n---\n`,
    })
    const by = (file: string) => fs.filter((f) => f.file.includes(file)).map((f) => f.rule)
    expect(by("/good/")).toEqual([])
    expect(by("/mismatch/")).toEqual(["IL-003"])
    expect(by("/caps/")).toEqual(["IL-002"])
    expect(by("/noname/")).toEqual(["IL-002"])
    expect(by("/long/")).toEqual(["IL-004"])
  })

  it("validates pack identity and version", () => {
    const fs = run({
      ".opencode/uf/packs/ok.md": PACK("ok", "text"),
      ".opencode/uf/packs/shared.md": "---\ndescription: Shared severity definitions.\n---\n",
      ".opencode/uf/packs/empty.md": "---\nfoo: 1\n---\n",
      ".opencode/uf/packs/bad.md": "---\npack_id: other\nversion: 1.0\n---\n",
      ".opencode/uf/packs/nover.md": "---\npack_id: nover\nlanguage: Go\n---\n",
    })
    const by = (file: string) => fs.filter((f) => f.file.endsWith(file)).map((f) => f.rule)
    expect(by("ok.md")).toEqual([])
    expect(by("shared.md")).toEqual([])
    expect(by("empty.md")).toEqual(["IL-002"])
    expect(by("bad.md").sort()).toEqual(["IL-002", "IL-002", "IL-003"])
    expect(by("nover.md")).toEqual(["IL-002"])
  })
})

describe("reference rules", () => {
  it("resolves slash commands, agents, skills, and paths", () => {
    const body = [
      "Run `/known` then `/missing --flag` and `/help`.",
      "```",
      "/fenced-missing arg",
      "cd /usr/local",
      "```",
      "Delegate to the `known-agent` agent, then the `ghost` subagent and the `general` subagent.",
      'Call with subagent_type: "phantom".',
      "Load the `real-skill` skill and the `fake-skill` skill.",
      'Then skill({ name: "fake-two" }).',
      "Read `.opencode/uf/packs/p.md`, `.opencode/nope.md:12`, `.opencode/agents/*.md`, and `.opencode/legacy/`.",
    ].join("\n")
    const fs = run({
      ".opencode/commands/known.md": COMMAND(body, "description: Known command used for reference checks.\nagent: known-agent\nhandoffs:\n  - label: Next\n    agent: known\n  - label: Bad\n    agent: nobody\n  - label: Shape\n    prompt: no agent key"),
      ".opencode/agents/known-agent.md": AGENT(),
      ".opencode/skills/real-skill/SKILL.md": SKILL("real-skill"),
      ".opencode/uf/packs/p.md": PACK("p", "text"),
    })
    const msgs = fs.filter((f) => f.rule.startsWith("IL-01")).map((f) => `${f.rule} ${f.message}`)
    expect(msgs.sort()).toEqual([
      "IL-010 /fenced-missing is not a command in .opencode/commands/",
      "IL-010 /missing is not a command in .opencode/commands/",
      'IL-011 agent "ghost" is not defined in .opencode/agents/',
      'IL-011 agent "nobody" is not defined in .opencode/agents/',
      'IL-011 agent "phantom" is not defined in .opencode/agents/',
      'IL-012 skill "fake-skill" is not defined in .opencode/skills/',
      'IL-012 skill "fake-two" is not defined in .opencode/skills/',
      "IL-013 .opencode/nope.md does not exist",
    ])
  })
})

describe("pack rule checks", () => {
  it("flags duplicate IDs, missing markers, force mismatches, and lowercase keywords", () => {
    const body = [
      "- **AB-001** [MUST] Things MUST happen.",
      "- **AB-001** [MUST] Again MUST happen.",
      "- **AB-002** Unmarked rule.",
      "- **AB-003** [MUST] Things SHOULD happen.",
      "- **AB-004** [SHOULD] Things MUST happen.",
      "- **AB-005** [MAY] Things MUST happen.",
      "- **AB-006** [MUST] Never do X. [SHOULD] Prefer Y.",
      "- **AB-007** [SHOULD] Things SHOULD happen,",
      "  and you must also check.",
    ].join("\n")
    const fs = run({ ".opencode/uf/packs/p.md": PACK("p", body) })
    expect(fs.map((f) => `${f.line} ${f.rule}`)).toEqual([
      "8 IL-014", "9 IL-020", "10 IL-021", "11 IL-021", "12 IL-021", "15 IL-022",
    ])
  })

  it("flags references to undefined rule IDs with a known prefix, except in commands", () => {
    const fs = run({
      ".opencode/uf/packs/p.md": PACK("p", "- **AB-001** [MUST] Things MUST happen."),
      ".opencode/agents/a.md": AGENT("Apply AB-001 and AB-009 but not ZZ-001.\n```\nAB-008\n```"),
      ".opencode/commands/c.md": COMMAND("Template uses AB-777."),
    })
    expect(fs.filter((f) => f.rule === "IL-015").map((f) => `${f.file} ${f.message}`)).toEqual([
      ".opencode/agents/a.md AB-009 is not defined in any pack",
    ])
  })

  it("skips rule ID references when no pack defines rules", () => {
    expect(run({ ".opencode/agents/a.md": AGENT("AB-001") })).toEqual([])
  })
})

describe("prose checks", () => {
  it("flags vague qualifiers only inside normative sentences, once per word", () => {
    const body = [
      "Handle errors as appropriate.",
      "You MUST handle errors appropriately and appropriately, etc. Then relax.",
      "",
      "- NEVER skip `relevant` checks; this is relevant.",
      "<!-- you MUST be reasonable -->",
    ].join("\n")
    const fs = run({ ".opencode/agents/a.md": AGENT(body) })
    expect(fs.filter((f) => f.rule === "IL-023").map((f) => [f.line, f.message.split('"')[1]])).toEqual([
      [7, "appropriately"], [7, "etc"], [9, "relevant"],
    ])
  })

  it("flags prescribed output and shell procedures in procedural files only", () => {
    const body = [
      "Print the summary using exactly this format:",
      "",
      "```bash",
      "git status",
      "$ gh pr view",
      "jq .x f.json",
      "```",
      "",
      "```python",
      "git = 1",
      "ls = 2",
      "find = 3",
      "```",
      "",
      "```",
      "git fetch",
      "```",
    ].join("\n")
    const fs = run({ ".opencode/commands/c.md": COMMAND(body), "AGENTS.md": body })
    expect(fs.map((f) => `${f.file}:${f.line} ${f.rule}`)).toEqual([
      ".opencode/commands/c.md:5 IL-040",
      ".opencode/commands/c.md:7 IL-041",
    ])
  })

  it("enforces file and load-set token budgets", () => {
    const fs = run({
      ".opencode/commands/c.md": COMMAND("Load the `big` skill and read `.opencode/uf/packs/p.md` and `.opencode/commands/c.md`."),
      ".opencode/agents/a.md": AGENT(),
      ".opencode/skills/big/SKILL.md": SKILL("big", "x".repeat(400)),
      ".opencode/uf/packs/p.md": PACK("p", "y".repeat(200)),
      "AGENTS.md": "z".repeat(100),
    }, { maxFileTokens: 80, maxLoadTokens: 150 })
    expect(fs.filter((f) => f.rule === "IL-030").map((f) => f.file)).toEqual([".opencode/skills/big/SKILL.md"])
    const load = fs.filter((f) => f.rule === "IL-031")
    expect(load.map((f) => f.file)).toEqual([".opencode/commands/c.md"])
    expect(load[0].message).toMatch(/AGENTS\.md ~25, \.opencode\/skills\/big\/SKILL\.md ~\d+, \.opencode\/uf\/packs\/p\.md/)
  })

  it("omits AGENTS.md from the load breakdown when absent", () => {
    const fs = run({ ".opencode/commands/c.md": COMMAND("x".repeat(400)) }, { maxLoadTokens: 10 })
    expect(fs.find((f) => f.rule === "IL-031")?.message).not.toContain("AGENTS.md")
  })

  it("groups near-duplicate paragraphs across files into one finding per file set", () => {
    const para = Array.from({ length: 50 }, (_, i) => `word${i}`).join(" ")
    const other = Array.from({ length: 50 }, (_, i) => `term${i}`).join(" ")
    const fs = run({
      ".opencode/commands/a.md": COMMAND(`${para}\n\n${other}`),
      ".opencode/commands/b.md": COMMAND(`${para}\n\n${other}`),
      ".opencode/commands/c.md": COMMAND(`${para}\n\n${para} extra`),
      ".opencode/commands/d.md": COMMAND("short text only"),
    })
    const dup = fs.filter((f) => f.rule === "IL-050").map((f) => `${f.file}:${f.line} ${f.message}`)
    expect(dup).toEqual([
      ".opencode/commands/a.md:5 1 paragraph(s) also appear near-verbatim in .opencode/commands/b.md:5, .opencode/commands/c.md:5; keep one source of truth",
      ".opencode/commands/a.md:7 1 paragraph(s) also appear near-verbatim in .opencode/commands/b.md:7; keep one source of truth",
    ])
  })
})

describe("suppression and options", () => {
  const vague = "You MUST act appropriately."
  it("honors next-line, file-level, disable, and ignorePaths", () => {
    expect(rules(run({ ".opencode/agents/a.md": AGENT(`<!-- instruction-lint-disable-next-line IL-023 -->\n${vague}`) }))).toEqual([])
    expect(rules(run({ ".opencode/agents/a.md": AGENT(`<!-- instruction-lint-disable-next-line -->\n${vague}`) }))).toEqual([])
    expect(rules(run({ ".opencode/agents/a.md": AGENT(`<!-- instruction-lint-disable-next-line IL-040 -->\n${vague}`) }))).toEqual(["IL-023"])
    expect(rules(run({ ".opencode/agents/a.md": AGENT(`<!-- instruction-lint-disable IL-023, IL-040 -->\n\n${vague}`) }))).toEqual([])
    expect(rules(run({ ".opencode/agents/a.md": AGENT(`<!-- instruction-lint-disable -->\n\n${vague}`) }))).toEqual([])
    expect(rules(run({ ".opencode/agents/a.md": AGENT(vague) }, { disable: ["IL-023"] }))).toEqual([])
    expect(rules(run({ ".opencode/agents/a.md": AGENT(vague) }, { ignorePaths: [".opencode/agents/"] }))).toEqual([])
  })

  it("catalogs every emitted rule", () => {
    for (const id of Object.keys(RULES)) expect(id).toMatch(/^IL-\d{3}$/)
  })
})

describe("cli", () => {
  function cli(args: string[], cwd: string) {
    const out: string[] = []
    const err: string[] = []
    const code = main(args, { cwd, out: (s) => out.push(s), err: (s) => err.push(s) })
    return { code, out: out.join("\n"), err: err.join("\n") }
  }

  it("prints help and the rule catalog", () => {
    expect(cli(["--help"], "/").out).toMatch(/usage: instruction-lint/)
    const r = cli(["--rules"], "/")
    expect(r.code).toBe(0)
    expect(r.out).toContain("IL-050")
  })

  it("rejects bad arguments", () => {
    expect(cli(["--bogus"], "/").code).toBe(2)
    expect(cli(["--root"], "/").err).toMatch(/needs a value/)
  })

  it("fails when no .opencode root is found", () => {
    const dir = mkdtempSync(join(tmpdir(), "instruction-lint-none-"))
    roots.push(dir)
    expect(findRoot(dir)).toBeNull()
    expect(cli([], dir).code).toBe(2)
  })

  it("finds the root from a subdirectory and exits by severity", () => {
    const root = repo({
      ".opencode/agents/a.md": AGENT("You MUST act appropriately.\n\nPrint exactly this format."),
    })
    const r = cli([], join(root, ".opencode", "agents"))
    expect(r.code).toBe(0)
    expect(r.out).toContain("IL-023")
    expect(r.out).not.toContain("IL-040")
    expect(r.out).toMatch(/0 errors, 1 warnings, 1 info \(use --all/)
    expect(cli(["--all"], root).out).toContain("IL-040")
    expect(cli(["--strict"], root).code).toBe(1)
  })

  it("exits 1 on errors and emits JSON and the token report", () => {
    const root = repo({
      ".opencode/agents/a.md": "# no frontmatter",
      ".opencode/commands/c.md": COMMAND(),
    })
    const j = cli(["--json", "--root", root], "/")
    expect(j.code).toBe(1)
    const parsed = JSON.parse(j.out)
    expect(parsed.summary).toEqual({ files: 2, errors: 1, warnings: 0, info: 0 })
    expect(Object.keys(parsed.loadTokens).sort()).toEqual([".opencode/agents/a.md", ".opencode/commands/c.md"])
    expect(parsed.findings[0]).toMatchObject({ rule: "IL-001", severity: "error", reviewSeverity: "HIGH" })
    const rep = cli(["--report"], root).out
    expect(rep).toMatch(/tokens\s+load\s+file/)
    expect(rep).toContain(".opencode/commands/c.md")
  })

  it("restricts reported findings with --only while analyzing the whole repo", () => {
    const root = repo({
      ".opencode/agents/a.md": "# no frontmatter",
      ".opencode/agents/b.md": AGENT("Use the `missing` skill."),
      ".opencode/commands/c.md": "# no frontmatter",
    })
    const files = (args: string[]) => JSON.parse(cli(["--json", ...args], root).out).findings.map((f: Finding) => f.file)
    expect(files(["--only", "./.opencode/agents/b.md"])).toEqual([".opencode/agents/b.md"])
    expect(files(["--only", ".opencode/agents/a.md, .opencode/commands/c.md,"]).sort()).toEqual([
      ".opencode/agents/a.md", ".opencode/commands/c.md",
    ])
    expect(files(["--only", ".opencode/agents/b.md", "--only", ".opencode/commands/c.md"]).sort()).toEqual([
      ".opencode/agents/b.md", ".opencode/commands/c.md",
    ])
    expect(cli(["--only", ".opencode/agents/b.md"], root).code).toBe(1)
    expect(cli(["--only"], root).code).toBe(2)
  })

  it("reads options from the default and explicit config files", () => {
    const root = repo({
      ".opencode/agents/a.md": AGENT("You MUST act appropriately."),
      ".opencode/instruction-lint.json": JSON.stringify({ disable: ["IL-023"] }),
      "alt.json": "{ not json",
    })
    expect(cli([], root).out).toMatch(/0 errors, 0 warnings, 0 info$/)
    const bad = cli(["--config", "alt.json"], root)
    expect(bad.code).toBe(2)
    expect(bad.err).toMatch(/cannot read config/)
  })
})

describe("delegation contract check (IL-042, DE-008)", () => {
  const block = (lines: string) => "```\n" + lines + "\n```"
  const il042 = (body: string, kind: "agent" | "command" | "skill" = "command") => {
    const files: Record<string, string> =
      kind === "agent" ? { ".opencode/agents/a.md": AGENT(body) }
      : kind === "skill" ? { ".opencode/skills/s/SKILL.md": SKILL("s", body) }
      : { ".opencode/commands/c.md": COMMAND(body) }
    return run(files).filter((f) => f.rule === "IL-042")
  }

  it("accepts complete call blocks and command-path delegations", () => {
    expect(il042(block("uf_delegate_command\n  command_path: \"x.md\"\n  checkpoint_step: \"plan\""))).toEqual([])
    expect(il042(block("uf_delegate_command\n  prompt: [p]\n  task_classification: \"judgment\""))).toEqual([])
    expect(il042(block(
      "uf_delegate_command\n  prompt: [p]\n  task_classification: [deterministic or judgment]\n  skill: [ref]\n  skill_absent_reason: [why]",
    ))).toEqual([])
    expect(il042(block("replicator_forge_spawn_subtask({ subtask_description: d, task_classification: \"deterministic\", skill: \"s\" })"), "skill")).toEqual([])
  })

  it("flags a missing classification, honoring the command_path exemption only where the tool has one", () => {
    const missing = il042(block("uf_delegate_command\n  prompt: [p]\n  feature_dir: F"))
    expect(missing).toHaveLength(1)
    expect(missing[0]).toMatchObject({ line: 6, severity: "warning", reviewSeverity: "MEDIUM" })
    expect(missing[0].message).toMatch(/no task_classification \(or `command_path`\)/)
    const spawn = il042(block("replicator_forge_spawn_subtask\n  command_path: x\n  files: [f]"), "agent")
    expect(spawn).toHaveLength(1)
    expect(spawn[0].message).not.toMatch(/command_path/)
  })

  it("flags a possibly-deterministic classification without skill fields", () => {
    const f = il042(block("uf_delegate_command\n  prompt: [p]\n  task_classification: \"deterministic\"\n  feature_dir: F"))
    expect(f).toHaveLength(1)
    expect(f[0].message).toMatch(/neither skill nor skill_absent_reason/)
  })

  it("treats each call in a block separately and ignores prose, other tools, and packs", () => {
    const two = il042(block("uf_delegate_command\n  command_path: a\nuf_delegate_command\n  prompt: p"))
    expect(two.map((f) => f.line)).toEqual([8])
    expect(il042("Call `uf_delegate_command` with the prompt.\n\nuf_delegate_command\n")).toEqual([])
    expect(il042(block("uf_delegate_command_v2\n  prompt: p\nuf_delegate_command: true"))).toEqual([])
    // A fence delimiter ends the call even when blocks are adjacent.
    expect(il042(block("uf_delegate_command\n  prompt: p") + "\n" + block("  task_classification: judgment"))).toHaveLength(1)
    expect(run({ ".opencode/uf/packs/p.md": PACK("p", block("uf_delegate_command\n  prompt: p")) }).filter((f) => f.rule === "IL-042")).toEqual([])
  })
})

describe("base deltas (--base)", () => {
  const sh = (cwd: string, ...args: string[]) => execFileSync("git", args, { cwd, stdio: "pipe" }).toString("utf8")
  function gitRepo(files: Record<string, string>): string {
    const root = repo(files)
    sh(root, "init", "-q", "-b", "main")
    sh(root, "-c", "user.email=t@t", "-c", "user.name=t", "add", "-A")
    sh(root, "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-q", "-m", "base")
    return root
  }
  function cli(args: string[], cwd: string) {
    const out: string[] = []
    const err: string[] = []
    const code = main(args, { cwd, out: (s) => out.push(s), err: (s) => err.push(s) })
    return { code, out: out.join("\n"), err: err.join("\n") }
  }

  it("counts lines like wc -l", () => {
    expect([countLines(""), countLines("a"), countLines("a\n"), countLines("a\nb")]).toEqual([0, 1, 1, 2])
  })

  it("classifies instruction paths the same way discovery does", () => {
    expect(classifyInstructionPath("AGENTS.md")).toEqual({ kind: "context", name: "AGENTS" })
    expect(classifyInstructionPath(".opencode/agents/a.md")).toEqual({ kind: "agent", name: "a" })
    expect(classifyInstructionPath(".opencode/commands/c.md")).toEqual({ kind: "command", name: "c" })
    expect(classifyInstructionPath(".opencode/uf/packs/p.md")).toEqual({ kind: "pack", name: "p" })
    expect(classifyInstructionPath(".opencode/skill/s/SKILL.md")).toEqual({ kind: "skill", name: "s" })
    expect(classifyInstructionPath(".opencode/agents/sub/a.md")).toBeNull()
    expect(classifyInstructionPath("docs/AGENTS.md")).toBeNull()
  })

  it("reports added, deleted, modified, and load-only files against the merge base", { timeout: 30000 }, () => {
    const root = gitRepo({
      ".opencode/agents/a.md": AGENT("Load `.opencode/skills/s/SKILL.md` first."),
      ".opencode/agents/gone.md": AGENT("old"),
      ".opencode/commands/c.md": COMMAND("same"),
      ".opencode/skills/s/SKILL.md": SKILL("s", "short"),
      ".opencode/uf/packs/p.md": PACK("p", "unchanged"),
    })
    rmSync(join(root, ".opencode/agents/gone.md"))
    writeFileSync(join(root, ".opencode/skills/s/SKILL.md"), SKILL("s", "a much longer body\nwith two lines"))
    writeFileSync(join(root, ".opencode/commands/new.md"), COMMAND("fresh"))

    const parsed = JSON.parse(cli(["--json", "--base", "main"], root).out)
    expect(parsed.delta.base).toBe("main")
    expect(parsed.delta.mergeBase).toMatch(/^[0-9a-f]{40}$/)
    const byFile = Object.fromEntries(parsed.delta.files.map((d: Delta) => [d.file, d]))
    expect(Object.keys(byFile)).toEqual([
      ".opencode/agents/a.md", ".opencode/agents/gone.md", ".opencode/commands/new.md", ".opencode/skills/s/SKILL.md",
    ])
    expect(byFile[".opencode/agents/a.md"].status).toBe("load-only")
    expect(byFile[".opencode/agents/a.md"].loadTokens.delta).toBeGreaterThan(0)
    expect(byFile[".opencode/agents/gone.md"]).toMatchObject({ status: "deleted", lines: { head: 0 } })
    expect(byFile[".opencode/commands/new.md"].status).toBe("added")
    expect(byFile[".opencode/skills/s/SKILL.md"]).toMatchObject({ status: "modified", lines: { delta: 1 }, loadTokens: null })
    expect(parsed.delta.totals.lines).toBe(parsed.delta.files.reduce((n: number, d: Delta) => n + d.lines.delta, 0))

    const only = JSON.parse(cli(["--json", "--base", "main", "--only", ".opencode/commands/new.md"], root).out)
    expect(only.delta.files.map((d: Delta) => d.file)).toEqual([".opencode/commands/new.md"])

    const text = cli(["--base", "main"], root).out
    expect(text).toMatch(/deltas since main \(merge-base [0-9a-f]{12}\)/)
    expect(text).toMatch(/added\s+\.opencode\/commands\/new\.md/)
    expect(text).toMatch(/total: [+-]?\d+ lines, [+-]?\d+ tokens across 4 files/)
    expect(JSON.parse(cli(["--json"], root).out).delta).toBeUndefined()
  })

  it("fails with exit 2 on an unknown base ref or a missing value", { timeout: 30000 }, () => {
    const root = gitRepo({ ".opencode/agents/a.md": AGENT("x") })
    const r = cli(["--base", "no-such-ref"], root)
    expect(r.code).toBe(2)
    expect(r.err).toMatch(/cannot read base no-such-ref/)
    expect(cli(["--base"], root).code).toBe(2)
  })

  it("handles a base with no instruction files", { timeout: 30000 }, () => {
    const root = gitRepo({ "README.md": "x" })
    writeFileSync(join(root, ".opencode/agents.md"), "")
    mkdirSync(join(root, ".opencode/agents"))
    writeFileSync(join(root, ".opencode/agents/a.md"), AGENT("x"))
    const parsed = JSON.parse(cli(["--json", "--base", "main"], root).out)
    expect(parsed.delta.files.map((d: Delta) => d.status)).toEqual(["added"])
  })
})
