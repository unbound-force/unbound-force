import type { PluginInput, ToolContext } from "@opencode-ai/plugin"
import { constants } from "node:fs"
import { access, copyFile, mkdir, mkdtemp, readFile, rm, symlink, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"
import { fileURLToPath } from "node:url"
import { afterEach, describe, expect, it } from "vitest"
import { parse } from "yaml"

import { parsePreflightConfig, type DeclaredCheck, type DeclaredChecksData } from "../lib/preflight-declared.js"
import type { ToolResult } from "../lib/uf-workflow-types.js"
import UfWorkflowPlugin from "../plugins/uf-workflow/index.js"
import { withBun } from "./helpers/bun-runtime.js"

// Seam verification (DE-007) for the instruction-lint pre-flight check:
// the repository declaration, the wrapper script, and a write->read round
// trip through the real uf_preflight_declared_checks tool.

const repo = (rel: string): string => fileURLToPath(new URL(`../../${rel}`, import.meta.url))
const scratch: string[] = []

afterEach(async () => {
  await Promise.all(scratch.splice(0).map((dir) => rm(dir, { recursive: true, force: true })))
})

async function declaredCheck(): Promise<DeclaredCheck> {
  const config = parsePreflightConfig(parse(await readFile(repo(".uf/preflight.yaml"), "utf8")))
  const check = config.checks.find((c) => c.name === "instruction-lint")
  if (!check) throw new Error("instruction-lint is not declared in .uf/preflight.yaml")
  return check
}

async function runTool(dir: string): Promise<DeclaredChecksData> {
  const hooks = await withBun(() =>
    UfWorkflowPlugin.server({ directory: dir, worktree: dir, client: {} } as unknown as PluginInput),
  )
  const tool = (hooks.tool ?? {})["uf_preflight_declared_checks"] as unknown as {
    execute: (args: Record<string, unknown>, ctx: ToolContext) => Promise<{ output: string }>
  }
  const result = await tool.execute({ mode: "soft-gate" }, {} as ToolContext)
  const envelope = JSON.parse(result.output) as ToolResult<DeclaredChecksData>
  if (envelope.status !== "ok") throw new Error(`tool error: ${JSON.stringify(envelope)}`)
  return envelope.data
}

describe("instruction-lint pre-flight check", () => {
  it("is declared as an advisory, instruction-scoped check with an executable wrapper", async () => {
    const check = await declaredCheck()
    expect(check).toMatchObject({ run: ".uf/scripts/instruction-lint", requires: ["node"], gate: "advisory" })
    expect(check.paths).toEqual(expect.arrayContaining([".opencode/agents/", ".opencode/commands/", "AGENTS.md"]))
    await expect(access(repo(check.run), constants.X_OK)).resolves.toBeUndefined()
  })

  it("round-trips linter findings through uf_preflight_declared_checks without failing the verdict", async () => {
    const check = await declaredCheck()
    const dir = await mkdtemp(join(tmpdir(), "uf-instruction-lint-preflight-"))
    scratch.push(dir)
    await mkdir(join(dir, ".uf", "scripts"), { recursive: true })
    await mkdir(join(dir, ".opencode", "lib"), { recursive: true })
    await mkdir(join(dir, ".opencode", "agents"), { recursive: true })
    await writeFile(join(dir, ".uf", "preflight.yaml"), JSON.stringify({ version: 1, checks: [check] }))
    await copyFile(repo(check.run), join(dir, check.run))
    await copyFile(repo(".opencode/lib/instruction-lint.ts"), join(dir, ".opencode", "lib", "instruction-lint.ts"))
    await symlink(repo(".opencode/node_modules"), join(dir, ".opencode", "node_modules"))
    const agent = join(dir, ".opencode", "agents", "a.md")

    await writeFile(agent, "# no frontmatter\n")
    const failing = await runTool(dir)
    expect(failing.checks.map((c) => [c.name, c.status, c.exit_code])).toEqual([["instruction-lint", "ADVISORY FAIL", 1]])
    expect(failing.checks[0]?.output).toContain(".opencode/agents/a.md:1  error    IL-001  file has no YAML frontmatter")
    expect(failing.required_failures).toEqual([])
    expect(failing.verdict).toBe("PASS")

    await writeFile(agent, "---\ndescription: \"Reviews instruction prose for the project.\"\nmode: subagent\n---\n\nBody.\n")
    const passing = await runTool(dir)
    expect(passing.checks.map((c) => [c.status, c.exit_code])).toEqual([["ADVISORY PASS", 0]])
    expect(passing.checks[0]?.output).toContain("1 files: 0 errors, 0 warnings, 0 info")
  }, 30_000)
})
