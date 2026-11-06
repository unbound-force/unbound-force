import type { PluginInput, PluginModule } from "@opencode-ai/plugin"
import { execFileSync } from "node:child_process"
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs"
import { tmpdir } from "node:os"
import { dirname, join } from "node:path"
import { fileURLToPath } from "node:url"
import { afterEach, beforeAll, describe, expect, it } from "vitest"

import {
  classifyTest,
  collectClis,
  inventory,
  locateFactory,
  main,
  match,
  parseChangedLines,
  parseRegistry,
  readStringConcat,
  type Inventory,
  type SeamReport,
} from "../lib/tool-inventory.js"
import InvokeAgentPlugin from "../plugins/invoke-agent/index.js"
import ReviewDispatchPlugin from "../plugins/review-dispatch/index.js"
import UfFinalePlugin from "../plugins/uf-finale/index.js"
import UfWorkflowPlugin from "../plugins/uf-workflow/index.js"
import { withBun } from "./helpers/bun-runtime.js"

const REPO = fileURLToPath(new URL("../..", import.meta.url))
const roots: string[] = []

function repo(files: Record<string, string>): string {
  const root = mkdtempSync(join(tmpdir(), "tool-inventory-"))
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

function cli(args: string[], cwd: string) {
  const out: string[] = []
  const err: string[] = []
  const code = main(args, { cwd, out: (s) => out.push(s), err: (s) => err.push(s) })
  return { code, out: out.join("\n"), err: err.join("\n") }
}

const memReader = (files: Record<string, string>) => (rel: string) => files[rel] ?? null

// ---------------------------------------------------------------------------
// Registration-surface seam: the static inventory MUST equal what the
// OpenCode host registers, so a reviewer citing it cites real tools.
// ---------------------------------------------------------------------------

describe("inventory matches the runtime registry", () => {
  const PLUGINS: Array<[string, PluginModule]> = [
    ["invoke-agent", InvokeAgentPlugin],
    ["review-dispatch", ReviewDispatchPlugin],
    ["uf-finale", UfFinalePlugin],
    ["uf-workflow", UfWorkflowPlugin],
  ]
  const runtime = new Map<string, Record<string, { description: string }>>()
  let inv: Inventory

  beforeAll(async () => {
    const input = { directory: "/tmp/fake", worktree: "/tmp/fake", client: {} } as unknown as PluginInput
    await withBun(async () => {
      for (const [id, mod] of PLUGINS) {
        const hooks = await mod.server(input)
        runtime.set(id, (hooks.tool ?? {}) as Record<string, { description: string }>)
      }
    })
    inv = inventory(REPO)
  })

  it("lists exactly the registered tool names per plugin", () => {
    for (const [id] of PLUGINS) {
      const statics = inv.tools.filter((t) => t.plugin === id).map((t) => t.name).sort()
      expect(statics).toEqual(Object.keys(runtime.get(id) ?? {}).sort())
    }
  })

  it("reads each tool's model-facing description and factory location", () => {
    for (const t of inv.tools) {
      expect(t.description, t.name).toBe(runtime.get(t.plugin)?.[t.name]?.description)
      expect(t.file, t.name).not.toBeNull()
      expect(t.endLine! > t.line!, t.name).toBe(true)
    }
  })

  it("lists repo-local CLI scripts", () => {
    const files = inv.clis.map((c) => c.file)
    expect(files).toContain(".opencode/lib/instruction-lint.ts")
    expect(files).toContain(".opencode/lib/tool-inventory.ts")
    expect(files).toContain(".uf/scripts/instruction-lint")
  })
})

describe("parsing", () => {
  it("parses registry entries at depth 1 only, skipping comments and strings", () => {
    const text = [
      "const x = { tool: 1 }",
      "return {",
      "  tool: {",
      "    // note: fake: comment(",
      "    alpha: createAlpha(a, { nested: createNested() }),",
      '    beta: createBeta("}", [1, 2]), /* gamma: createGamma() */',
      "    delta: createDelta(`${x}`),",
      "  },",
      "}",
      "other: createOther()",
    ].join("\n")
    expect(parseRegistry(text)).toEqual([
      { name: "alpha", factory: "createAlpha", line: 5 },
      { name: "beta", factory: "createBeta", line: 6 },
      { name: "delta", factory: "createDelta", line: 7 },
    ])
    expect(parseRegistry("return { tool: { only: createOnly(d) } }")).toEqual([{ name: "only", factory: "createOnly", line: 1 }])
    expect(parseRegistry("no registry here")).toEqual([])
    expect(parseRegistry("tool: { a: createA(")).toEqual([{ name: "a", factory: "createA", line: 1 }])
  })

  it("reads string-literal concatenations", () => {
    expect(readStringConcat(` "a \\"q\\"" +\n 'b\\'c' + \`d\``, 0)).toBe('a "q"b\'cd')
    expect(readStringConcat(` "only", x`, 0)).toBe("only")
    expect(readStringConcat(" SOME_CONST", 0)).toBeNull()
    expect(readStringConcat(` "a" + SOME_CONST`, 0)).toBe("a")
  })

  it("locates factories locally, through relative imports, and as const declarations", () => {
    const files = {
      ".opencode/plugins/p/index.ts": [
        'import { other } from "./x.js"',
        'import { createRemote, type T } from "../../lib/remote.js"',
        "export function createLocal(\n  a: string,\n): Tool {",
        '  return tool({ description: "local tool" })',
        "}",
      ].join("\n"),
      ".opencode/lib/remote.ts": 'export const createRemote = () =>\n  tool({\n    description: "remote " + "tool",\n  })\n}\n',
    }
    const read = memReader(files)
    expect(locateFactory(".opencode/plugins/p/index.ts", "createLocal", read)).toEqual({
      file: ".opencode/plugins/p/index.ts", line: 3, endLine: 7, description: "local tool",
    })
    expect(locateFactory(".opencode/plugins/p/index.ts", "createRemote", read)).toMatchObject({
      file: ".opencode/lib/remote.ts", line: 1, description: "remote tool",
    })
    expect(locateFactory(".opencode/plugins/p/index.ts", "createMissing", read)).toBeNull()
    expect(locateFactory(".opencode/plugins/none/index.ts", "createLocal", read)).toBeNull()
    expect(locateFactory(".opencode/plugins/p/index.ts", "other", read)).toBeNull()
  })

  it("parses head-side changed lines from a zero-context diff", () => {
    const diff = [
      "diff --git a/f.ts b/f.ts",
      "--- a/f.ts",
      "+++ b/f.ts",
      "@@ -1,0 +2,2 @@",
      "@@ -9 +11 @@",
      "@@ -20,3 +21,0 @@",
      "--- a/gone.ts",
      "+++ /dev/null",
      "@@ -1,2 +0,0 @@",
    ].join("\n")
    const m = parseChangedLines(diff)
    expect([...(m.get("f.ts") ?? [])].sort((a, b) => a - b)).toEqual([2, 3, 11, 21])
    expect(m.has("gone.ts")).toBe(false)
  })

  it("classifies test files by seam layer", () => {
    expect(classifyTest(".opencode/test/registration-surface.test.ts", "", "t")).toBe("registration")
    expect(classifyTest(".opencode/test/golden-output-shape.test.ts", "", "t")).toBe("output-shape")
    expect(classifyTest(".opencode/test/x-roundtrip.test.ts", "", "t")).toBe("round-trip")
    expect(classifyTest(".opencode/test/x.test.ts", 'describe("my_tool round trip", () => {})', "my_tool")).toBe("round-trip")
    expect(classifyTest(".opencode/test/x.test.ts", 'it("round-trips through my_tool")', "my_tool")).toBe("round-trip")
    expect(classifyTest(".opencode/test/x.test.ts", 'it("round trip of other_tool")', "my_tool")).toBe("other")
  })

  it("finds the registry only in code, ignoring comment and string decoys", () => {
    const text = [
      "// tool: { fake: createFake() }",
      'const s = "tool: { in_string: createS() }"',
      "export default {",
      "  tool: {",
      "    real: createReal(),",
      "  },",
      "}",
    ].join("\n")
    expect(parseRegistry(text)).toEqual([{ name: "real", factory: "createReal", line: 5 }])
  })

  it("closes const-arrow factories on balanced brackets, not a column-0 brace", () => {
    const files = {
      ".opencode/plugins/p/index.ts": [
        'import { createA } from "../../lib/a.js"',
        "export default { server() { return { tool: { a_tool: createA() } } } }",
      ].join("\n"),
      ".opencode/lib/a.ts": [
        "export const createA = (x: string) =>",
        "  tool({",
        '    description: "a tool",',
        "  })",
        "const later = 1",
      ].join("\n"),
    }
    expect(locateFactory(".opencode/plugins/p/index.ts", "createA", memReader(files))).toEqual({
      file: ".opencode/lib/a.ts", line: 1, endLine: 4, description: "a tool",
    })
  })

  it("locates factories and classifies tests for names containing $", () => {
    const files = {
      ".opencode/plugins/p/index.ts": [
        'import { create$Tool } from "../../lib/d.js"',
        "export default { server() { return { tool: { t: create$Tool() } } } }",
      ].join("\n"),
      ".opencode/lib/d.ts": 'export function create$Tool() {\n  return tool({ description: "dollar" })\n}\n',
    }
    expect(locateFactory(".opencode/plugins/p/index.ts", "create$Tool", memReader(files))).toEqual({
      file: ".opencode/lib/d.ts", line: 1, endLine: 3, description: "dollar",
    })
    expect(classifyTest(".opencode/test/x.test.ts", 'it("round-trips through my$tool")', "my$tool")).toBe("round-trip")
  })

  it("collects lib CLIs and .uf scripts with their first comment line", () => {
    const root = repo({
      ".opencode/lib/cli.ts": "// Does the thing.\nif (import.meta.main) {}\n",
      ".opencode/lib/bare.ts": "if (import.meta.main) {}\n",
      ".opencode/lib/lib.ts": "// Library only.\nexport const x = 1\n",
      ".opencode/lib/notes.md": "x",
      ".uf/scripts/run": "#!/bin/sh\n#\n# Runs it.\nexit 0\n",
      ".uf/scripts/plain": "exit 0\n",
    })
    expect(collectClis(root)).toEqual([
      { name: "bare", file: ".opencode/lib/bare.ts", summary: "" },
      { name: "cli", file: ".opencode/lib/cli.ts", summary: "Does the thing." },
      { name: "plain", file: ".uf/scripts/plain", summary: "" },
      { name: "run", file: ".uf/scripts/run", summary: "Runs it." },
    ])
  })

  it("ranks matches by word count, then name, and drops non-matches", () => {
    const tool = (name: string, description: string | null) => ({
      name, plugin: "p", factory: "f", file: null, line: null, endLine: null, registryLine: 1, description,
    })
    const inv: Inventory = {
      tools: [tool("b_tool", "resolve the base ref"), tool("a_tool", "base only"), tool("c_tool", null)],
      clis: [{ name: "lint", file: "l", summary: "Lint the BASE" }, { name: "x", file: "x", summary: "" }],
    }
    const m = match(inv, ["Base", "ref", ""])
    expect(m.tools.map((t) => t.name)).toEqual(["b_tool", "a_tool"])
    expect(m.clis.map((c) => c.name)).toEqual(["lint"])
  })
})

// ---------------------------------------------------------------------------
// Seam mode against a real scratch git repository
// ---------------------------------------------------------------------------

describe("seam report", () => {
  const sh = (cwd: string, ...args: string[]) => execFileSync("git", args, { cwd, stdio: "pipe" }).toString("utf8")
  const commit = (root: string, msg: string) => {
    sh(root, "add", "-A")
    sh(root, "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-q", "-m", msg)
  }
  const plugin = (entries: string[], factories: string) =>
    `export default {\n  server() {\n    return {\n      tool: {\n${entries.map((e) => `        ${e},`).join("\n")}\n      },\n    }\n  },\n}\n${factories}`
  const factory = (name: string, desc: string) => `\nexport function ${name}() {\n  return tool({\n    description: "${desc}",\n  })\n}\n`

  function scratch(): string {
    const root = repo({
      ".opencode/plugins/alpha/index.ts": plugin(
        ["keep_tool: createKeep()", "edit_tool: createEdit()", "drop_tool: createDrop()"],
        factory("createKeep", "keep") + factory("createEdit", "edit v1") + factory("createDrop", "drop"),
      ),
      ".opencode/test/registration-surface.test.ts": 'const PLUGINS = [{ id: "alpha", module: A }]\n',
      ".opencode/test/golden-output-shape.test.ts": "edit_tool golden\n",
      ".opencode/test/unit.test.ts": "createKeep\n",
    })
    sh(root, "init", "-q", "-b", "main")
    commit(root, "base")
    return root
  }

  it("reports added, modified, and removed tools with tests per layer", { timeout: 30000 }, () => {
    const root = scratch()
    writeFileSync(
      join(root, ".opencode/plugins/alpha/index.ts"),
      plugin(
        ["keep_tool: createKeep()", "edit_tool: createEdit()", "new_tool: createNew()"],
        factory("createKeep", "keep") + factory("createEdit", "edit v2") + factory("createNew", "new"),
      ),
    )
    writeFileSync(join(root, ".opencode/test/new-roundtrip.test.ts"), "new_tool write then read\n")
    writeFileSync(join(root, ".opencode/test/drop.test.ts"), "drop_tool\n")

    const parsed = JSON.parse(cli(["--seams", "--base", "main", "--json"], root).out) as SeamReport
    expect(parsed.base).toBe("main")
    const by = Object.fromEntries(parsed.seams.map((s) => [s.tool, s]))
    expect(Object.keys(by)).toEqual(["drop_tool", "edit_tool", "new_tool"])
    expect(by.edit_tool).toMatchObject({
      status: "modified",
      genericRegistration: true,
      tests: [{ file: ".opencode/test/golden-output-shape.test.ts", layer: "output-shape", changed: false }],
      missing: ["round-trip"],
    })
    expect(by.new_tool).toMatchObject({
      status: "added",
      tests: [{ file: ".opencode/test/new-roundtrip.test.ts", layer: "round-trip", changed: true }],
      missing: ["output-shape"],
    })
    expect(by.drop_tool).toMatchObject({ status: "removed", genericRegistration: false, missing: [] })
    expect(by.drop_tool.tests.map((t) => t.file)).toEqual([".opencode/test/drop.test.ts"])

    const text = cli(["--seams", "--base", "main"], root).out
    expect(text).toMatch(/seams since main \(merge-base [0-9a-f]{12}\): 3 tool\(s\)/)
    expect(text).toContain("edit_tool  [alpha, modified]")
    expect(text).toContain("registration  generic: .opencode/test/registration-surface.test.ts")
    expect(text).toMatch(/round-trip\s+\.opencode\/test\/new-roundtrip\.test\.ts {2}\(changed\)/)
    expect(text).toContain("missing: round-trip")
  })

  it("treats an untracked plugin as all-new and reports nothing for an unchanged tree", { timeout: 30000 }, () => {
    const root = scratch()
    expect(JSON.parse(cli(["--seams", "--base", "main", "--json"], root).out).seams).toEqual([])
    writeFileSync(join(root, ".opencode/test/registration-surface.test.ts"), "no generic list\n")
    commit(root, "drop generic registration list")
    mkdirSync(join(root, ".opencode/plugins/beta"))
    writeFileSync(join(root, ".opencode/plugins/beta/index.ts"), plugin(["beta_tool: createBeta()"], factory("createBeta", "b")))
    const seams = (JSON.parse(cli(["--seams", "--base", "main", "--json"], root).out) as SeamReport).seams
    expect(seams).toEqual([
      { tool: "beta_tool", plugin: "beta", status: "added", genericRegistration: false, tests: [], missing: ["registration", "output-shape", "round-trip"] },
    ])
  })

  it("flags modified factories in untracked lib files and registry-line edits", { timeout: 30000 }, () => {
    const root = scratch()
    const p = join(root, ".opencode/plugins/alpha/index.ts")
    // Re-point keep_tool at a factory in a new, untracked lib module.
    writeFileSync(
      p,
      'import { createKeep2 } from "../../lib/k.js"\n' +
        plugin(["keep_tool: createKeep2()", "edit_tool: createEdit()", "drop_tool: createDrop()"], factory("createEdit", "edit v1") + factory("createDrop", "drop")),
    )
    mkdirSync(join(root, ".opencode/lib"))
    writeFileSync(join(root, ".opencode/lib/k.ts"), factory("createKeep2", "keep"))
    const seams = (JSON.parse(cli(["--seams", "--base", "main", "--json"], root).out) as SeamReport).seams
    expect(seams.map((s) => [s.tool, s.status])).toEqual([["keep_tool", "modified"]])
  })
})

describe("cli", () => {
  it("prints help and rejects bad arguments", () => {
    expect(cli(["--help"], "/").out).toMatch(/usage: tool-inventory/)
    expect(cli(["--bogus"], "/").code).toBe(2)
    expect(cli(["--match"], "/").err).toMatch(/needs a value/)
  })

  it("fails when no .opencode root is found", () => {
    const dir = mkdtempSync(join(tmpdir(), "tool-inventory-none-"))
    roots.push(dir)
    expect(cli([], dir).code).toBe(2)
  })

  it("requires --base for --seams and reports an unreadable base", () => {
    const root = repo({})
    expect(cli(["--seams"], root).err).toMatch(/--seams needs --base/)
    const r = cli(["--seams", "--base", "nope", "--root", root], "/")
    expect(r.code).toBe(2)
    expect(r.err).toMatch(/cannot read base nope/)
    const thrown = main(["--seams", "--base", "x"], { cwd: root, out: () => {}, err: (s) => expect(s).toMatch(/boom/) }, () => {
      throw new Error("boom")
    })
    expect(thrown).toBe(2)
  })

  it("prints the inventory as text and JSON, filtered by --match", () => {
    const root = repo({
      ".opencode/plugins/p/index.ts":
        'export default { server() { return { tool: { a_tool: createA(), b_tool: createB() } } } }\n' +
        'export function createA() {\n  return tool({ description: "resolve\\n  things" })\n}\n',
      ".uf/scripts/s": "# Script summary.\n",
    })
    const text = cli([], root).out
    expect(text).toContain("a_tool  [p]  .opencode/plugins/p/index.ts:2\n  resolve things")
    expect(text).toContain("b_tool  [p]  (factory not found)")
    expect(text).toContain(".uf/scripts/s  [cli]\n  Script summary.")
    expect(text).toMatch(/2 tools, 1 scripts$/)
    const j = JSON.parse(cli(["--json", "--match", "resolve"], root).out) as Inventory
    expect(j.tools.map((t) => t.name)).toEqual(["a_tool"])
    expect(j.clis).toEqual([])
  })
})
