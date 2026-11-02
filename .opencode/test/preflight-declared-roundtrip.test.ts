import type { PluginInput, ToolContext } from "@opencode-ai/plugin"
import { chmod, mkdir, mkdtemp, rm, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"
import { afterEach, describe, expect, it } from "vitest"

import type { DeclaredChecksData } from "../lib/preflight-declared.js"
import type { ToolResult } from "../lib/uf-workflow-types.js"
import UfWorkflowPlugin from "../plugins/uf-workflow/index.js"
import { withBun } from "./helpers/bun-runtime.js"

// ---------------------------------------------------------------------------
// Layer 3 — write→read round trip for uf_preflight_declared_checks
// (determinism pack DE-007). A real .uf/preflight.yaml and a real
// executable on disk go through the registered tool with the production
// dependencies (fs, PATH probe, execFile without a shell); the envelope
// read back must reflect exactly what the script did.
// ---------------------------------------------------------------------------

const scratch: string[] = []

afterEach(async () => {
  await Promise.all(scratch.splice(0).map((dir) => rm(dir, { recursive: true, force: true })))
})

async function project(files: Readonly<Record<string, { body: string; mode?: number }>>): Promise<string> {
  const dir = await mkdtemp(join(tmpdir(), "uf-preflight-rt-"))
  scratch.push(dir)
  for (const [rel, { body, mode }] of Object.entries(files)) {
    const target = join(dir, rel)
    await mkdir(join(target, ".."), { recursive: true })
    await writeFile(target, body, "utf8")
    if (mode !== undefined) {
      await chmod(target, mode)
    }
  }
  return dir
}

async function runTool(dir: string): Promise<ToolResult<DeclaredChecksData>> {
  const hooks = await withBun(() =>
    UfWorkflowPlugin.server({ directory: dir, worktree: dir, client: {} } as unknown as PluginInput),
  )
  const tool = (hooks.tool ?? {})["uf_preflight_declared_checks"] as unknown as {
    execute: (args: Record<string, unknown>, ctx: ToolContext) => Promise<{ output: string }>
  }
  const result = await tool.execute({ mode: "soft-gate" }, {} as ToolContext)
  return JSON.parse(result.output) as ToolResult<DeclaredChecksData>
}

describe("uf_preflight_declared_checks round trip", () => {
  it("executes the declared script without a shell and reports its exit code and output", async () => {
    const dir = await project({
      ".uf/preflight.yaml": {
        body: JSON.stringify({
          version: 1,
          checks: [
            { name: "fails", run: ".uf/scripts/fail", args: ["$HOME", "a b"] },
            { name: "advisory", run: ".uf/scripts/fail", gate: "advisory" },
          ],
        }),
      },
      ".uf/scripts/fail": { body: '#!/bin/sh\nprintf "argc=%s first=%s\\n" "$#" "$1"\nexit 3\n', mode: 0o755 },
    })

    const envelope = await runTool(dir)
    expect(envelope.status).toBe("ok")
    const data = (envelope as { data: DeclaredChecksData }).data
    expect(data.checks.map((c) => [c.name, c.status, c.exit_code])).toEqual([
      ["fails", "FAIL", 3],
      ["advisory", "ADVISORY FAIL", 3],
    ])
    // Literal argv: "$HOME" is not expanded and "a b" stays one argument.
    expect(data.checks[0]?.output.trim()).toBe("argc=2 first=$HOME")
    expect(data.required_failures).toEqual(["fails"])
  })

  it("reports a non-executable run path as unavailable instead of failing", async () => {
    const dir = await project({
      ".uf/preflight.yaml": { body: JSON.stringify({ version: 1, checks: [{ name: "noexec", run: "check.sh" }] }) },
      "check.sh": { body: "#!/bin/sh\nexit 0\n", mode: 0o644 },
    })
    const envelope = await runTool(dir)
    const data = (envelope as { data: DeclaredChecksData }).data
    expect(data.checks[0]?.status).toBe("SKIP — not available")
    expect(data.checks[0]?.unavailable).toEqual(["check.sh"])
  })

  it("returns config_present=false when no file is declared", async () => {
    const dir = await project({})
    const envelope = await runTool(dir)
    expect((envelope as { data: DeclaredChecksData }).data.config_present).toBe(false)
  })
})
