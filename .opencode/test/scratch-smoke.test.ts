import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"

import type { PluginInput } from "@opencode-ai/plugin"
import { afterEach, describe, expect, it } from "vitest"

import InvokeAgentPlugin from "../plugins/invoke-agent/index.js"
import ReviewDispatchPlugin from "../plugins/review-dispatch/index.js"
import manifestFixture from "./fixtures/review-dispatch/manifest.json"
import matrixFixture from "./fixtures/review-dispatch/matrix.json"

const scratchDirectories: string[] = []

afterEach(async () => {
  await Promise.all(scratchDirectories.splice(0).map((directory) => rm(directory, { recursive: true, force: true })))
})

type BunRuntime = { readonly YAML: { readonly parse: (text: string) => unknown } }

function withBun<T>(run: () => T | Promise<T>): Promise<T> {
  const runtime = globalThis as typeof globalThis & { Bun?: BunRuntime }
  const original = runtime.Bun
  // The production plugin reads the narrow Bun.YAML.parse boundary. This stub
  // supplies an in-process parser so the smoke test never touches a provider,
  // package registry, GitHub, or the network.
  runtime.Bun = { YAML: { parse: (text: string): unknown => JSON.parse(text) as unknown } }
  const finish = (): void => {
    runtime.Bun = original
  }
  try {
    return Promise.resolve(run()).finally(finish)
  } catch (error) {
    finish()
    throw error
  }
}

/** Builds a target-project-shaped `.opencode/` tree with real policy files. */
async function scratchProject(): Promise<string> {
  const directory = await mkdtemp(join(tmpdir(), "uf-scratch-smoke-"))
  scratchDirectories.push(directory)
  await mkdir(join(directory, ".opencode"), { recursive: true })
  await mkdir(join(directory, ".uf"), { recursive: true })
  await writeFile(join(directory, ".uf", "review-matrix.yaml"), JSON.stringify(matrixFixture))
  await writeFile(join(directory, ".uf", "reviewer-capabilities.yaml"), JSON.stringify(manifestFixture))
  await writeFile(join(directory, ".uf", "sibling-repos.yaml"), JSON.stringify({ version: 1, siblings: [] }))
  return directory
}

describe("provider-free scratch-repository smoke test", () => {
  it("activates both plugins in a target project and registers every tool with zero network", async () => {
    await withBun(async () => {
      const project = await scratchProject()

      const invokeHooks = await InvokeAgentPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)
      const dispatchHooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      expect(Object.keys(invokeHooks.tool)).toEqual(["invoke_agent"])
      expect(Object.keys(dispatchHooks.tool).sort()).toEqual([
        "acquire_sibling_evidence",
        "finalize_review_dispatch",
        "plan_review_dispatch",
        "prepare_lesson_learning",
      ])
    })
  })

  it("produces a deterministic plan from scratch-repository policy files", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const dispatchHooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const output = await dispatchHooks.tool.plan_review_dispatch.execute(
        {
          mode: "code",
          discovered_agents: ["divisor-guard", "divisor-adversary"],
          changed_files: [{ path: ".opencode/agents/guard.md", additions: 3, deletions: 0 }],
        },
        {} as never,
      )
      const plan = JSON.parse(output) as { plan_version: number; status: string; entries: Array<{ agent: string; decision: string }> }
      expect(plan.plan_version).toBe(1)
      expect(plan.status).toBe("ready")
      const includedAgents = [...new Set(plan.entries.filter((entry) => entry.decision === "include").map((entry) => entry.agent))].sort()
      expect(includedAgents).toEqual(["divisor-adversary", "divisor-guard"])
    })
  })

  it("never emits the invoke_agent permission contract into child requests", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await InvokeAgentPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      expect(hooks.tool.invoke_agent.args).not.toHaveProperty("permission")
      expect(hooks.tool.invoke_agent.args).not.toHaveProperty("tools")
    })
  })
})
