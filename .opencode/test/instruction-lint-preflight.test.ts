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
    expect(check.paths).toEqual(
      expect.arrayContaining([
        ".opencode/agents/",
        ".opencode/commands/",
        "AGENTS.md",
        check.run,
        ".opencode/package.json",
      ]),
    )
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
    await copyFile(repo(".opencode/lib/cli-common.ts"), join(dir, ".opencode", "lib", "cli-common.ts"))
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

  it("exits 2 with an actionable message when dependencies are not installed", async () => {
    const dir = await mkdtemp(join(tmpdir(), "uf-instruction-lint-nodeps-"))
    scratch.push(dir)
    await mkdir(join(dir, ".opencode", "lib"), { recursive: true })
    await copyFile(repo(".opencode/lib/instruction-lint.ts"), join(dir, ".opencode", "lib", "instruction-lint.ts"))
    const { execFile } = await import("node:child_process")
    const result = await new Promise<{ code: number | undefined; stderr: string }>((resolve) => {
      execFile(
        repo(".uf/scripts/instruction-lint"),
        [],
        { env: { ...process.env, UF_PREFLIGHT_PROJECT_ROOT: dir }, cwd: dir },
        (err, _stdout, stderr) => resolve({ code: err?.code, stderr: String(stderr) }),
      )
    })
    expect(result.code).toBe(2)
    expect(result.stderr).toContain("dependencies not installed")
    expect(result.stderr).toContain("npm install")
  })

  it("exits 2 when node is absent or cannot run TypeScript", async () => {
    const dir = await mkdtemp(join(tmpdir(), "uf-instruction-lint-nonode-"))
    scratch.push(dir)
    await mkdir(join(dir, ".opencode", "lib"), { recursive: true })
    await mkdir(join(dir, "absent"))
    await mkdir(join(dir, "old"))
    await copyFile(repo(".opencode/lib/instruction-lint.ts"), join(dir, ".opencode", "lib", "instruction-lint.ts"))
    await symlink(repo(".opencode/node_modules"), join(dir, ".opencode", "node_modules"))
    await writeFile(join(dir, "old", "node"), "#!/bin/sh\nexit 1\n", { mode: 0o755 })
    const { execFile } = await import("node:child_process")
    for (const bin of ["absent", "old"]) {
      const result = await new Promise<{ code: number | undefined; stderr: string }>((resolve) => {
        execFile(
          repo(".uf/scripts/instruction-lint"),
          [],
          { env: { ...process.env, PATH: join(dir, bin), UF_PREFLIGHT_PROJECT_ROOT: dir }, cwd: dir },
          (err, _stdout, stderr) => resolve({ code: err?.code as number | undefined, stderr: String(stderr) }),
        )
      })
      expect(result.code).toBe(2)
      expect(result.stderr).toContain("cannot run TypeScript")
    }
  })

  it("routes the make target and the persona through the wrapper", async () => {
    const makefile = await readFile(repo("Makefile"), "utf8")
    expect(makefile).toMatch(/^lint-instructions:.*\n\t\.uf\/scripts\/instruction-lint \$\(ARGS\)$/m)
    const persona = await readFile(repo(".opencode/agents/divisor-linguist.md"), "utf8")
    expect(persona).toContain(".uf/scripts/instruction-lint --json --only")
    expect(makefile + persona).not.toMatch(/node\s+\S*instruction-lint\.ts/)
  })

  it("keeps the persona's metacharacter denies after every bash allow", async () => {
    // The permission engine is last-match-wins: an allow after these
    // denies would reopen redirection and chaining on its prefix.
    const persona = await readFile(repo(".opencode/agents/divisor-linguist.md"), "utf8")
    const frontmatter = persona.split(/^---$/m)[1] ?? ""
    const bash = Object.entries((parse(frontmatter) as { permission: { bash: Record<string, string> } }).permission.bash)
    const terminal = ["*>*", "*<*", "*|*", "*;*", "*&*", "*$(*", "*`*"]
    expect(bash.slice(-terminal.length)).toEqual(terminal.map((pattern) => [pattern, "deny"]))

    const decide = (command: string): string | undefined =>
      bash.findLast(([pattern]) => {
        const source = pattern.split("*").map((part) => part.replace(/[.+?^${}()|[\]\\]/g, "\\$&")).join(".*")
        return new RegExp(`^${source}$`, "s").test(command)
      })?.[1]
    expect(decide("git diff --stat")).toBe("allow")
    expect(decide("git status --short")).toBe("allow")
    expect(decide(".uf/scripts/instruction-lint --json")).toBe("allow")
    expect(decide("git diff --output=/tmp/x")).toBe("deny")
    expect(decide("git difftool --extcmd=sh")).toBe("deny")
    expect(decide("git status > file")).toBe("deny")
    expect(decide("git diff | sh")).toBe("deny")
  })

  it("checks every package the linter imports before starting it", async () => {
    const source = await readFile(repo(".opencode/lib/instruction-lint.ts"), "utf8")
    const imported = [...source.matchAll(/^import .* from "([^"]+)"$/gm)]
      .map((match) => match[1] as string)
      .filter((specifier) => !specifier.startsWith("node:") && !specifier.startsWith("."))
    const wrapper = await readFile(repo(".uf/scripts/instruction-lint"), "utf8")
    const checked = /^for pkg in (.+); do$/m.exec(wrapper)?.[1]?.split(" ")
    expect(checked?.sort()).toEqual(imported.sort())
  })

  it("exits 2 when the linter entry is missing or a dependency package is incomplete", async () => {
    const { execFile } = await import("node:child_process")
    const run = (cwd: string) =>
      new Promise<{ code: number | undefined; stderr: string }>((resolve) => {
        execFile(
          repo(".uf/scripts/instruction-lint"),
          [],
          { env: { ...process.env, UF_PREFLIGHT_PROJECT_ROOT: cwd }, cwd },
          (err, _stdout, stderr) => resolve({ code: err?.code, stderr: String(stderr) }),
        )
      })

    const noEntry = await mkdtemp(join(tmpdir(), "uf-instruction-lint-noentry-"))
    scratch.push(noEntry)
    await mkdir(join(noEntry, ".opencode"), { recursive: true })
    await symlink(repo(".opencode/node_modules"), join(noEntry, ".opencode", "node_modules"))
    const entryMissing = await run(noEntry)
    expect(entryMissing.code).toBe(2)
    expect(entryMissing.stderr).toContain("linter entry not found")

    const partial = await mkdtemp(join(tmpdir(), "uf-instruction-lint-partial-"))
    scratch.push(partial)
    await mkdir(join(partial, ".opencode", "lib"), { recursive: true })
    await copyFile(repo(".opencode/lib/instruction-lint.ts"), join(partial, ".opencode", "lib", "instruction-lint.ts"))
    await mkdir(join(partial, ".opencode", "node_modules"), { recursive: true })
    await symlink(repo(".opencode/node_modules/yaml"), join(partial, ".opencode", "node_modules", "yaml"))
    const noZod = await run(partial)
    expect(noZod.code).toBe(2)
    expect(noZod.stderr).toContain("dependencies not installed")
    expect(noZod.stderr).toContain("node_modules/zod")
  }, 30_000)

  it("lints the working directory using the primary checkout's linter (split roots)", async () => {
    const primary = await mkdtemp(join(tmpdir(), "uf-instruction-lint-primary-"))
    const worktree = await mkdtemp(join(tmpdir(), "uf-instruction-lint-worktree-"))
    scratch.push(primary, worktree)
    await mkdir(join(primary, ".opencode", "lib"), { recursive: true })
    await mkdir(join(primary, ".opencode", "agents"), { recursive: true })
    await copyFile(repo(".opencode/lib/instruction-lint.ts"), join(primary, ".opencode", "lib", "instruction-lint.ts"))
    await copyFile(repo(".opencode/lib/cli-common.ts"), join(primary, ".opencode", "lib", "cli-common.ts"))
    await symlink(repo(".opencode/node_modules"), join(primary, ".opencode", "node_modules"))
    // The primary checkout is dirty and the worktree clean: exit 0 with
    // the worktree's file count proves the working directory was
    // linted (linting the primary instead would report its dirty agent
    // as an IL-001 error), using the primary's linter and deps.
    await writeFile(join(primary, ".opencode", "agents", "dirty.md"), "# no frontmatter\n")
    await mkdir(join(worktree, ".opencode", "agents"), { recursive: true })
    await writeFile(
      join(worktree, ".opencode", "agents", "clean.md"),
      "---\ndescription: \"Reviews instruction prose for the project.\"\nmode: subagent\n---\n\nBody.\n",
    )
    const { execFile } = await import("node:child_process")
    const result = await new Promise<{ code: number; stdout: string }>((resolve) => {
      execFile(
        repo(".uf/scripts/instruction-lint"),
        [],
        { env: { ...process.env, UF_PREFLIGHT_PROJECT_ROOT: primary }, cwd: worktree },
        (err, stdout, _stderr) => resolve({ code: err?.code ?? 0, stdout: String(stdout) }),
      )
    })
    expect(result.code).toBe(0)
    expect(result.stdout).toContain("1 files: 0 errors, 0 warnings, 0 info")
  }, 30_000)
})
