import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"
import { randomUUID } from "node:crypto"

import type { PluginInput, ToolContext } from "@opencode-ai/plugin"
import { afterEach, beforeEach, describe, expect, it } from "vitest"

import { ReviewDispatchPlugin, _getSubmissionStore, _getSessionCorrelationMap } from "../plugins/review-dispatch/index.js"
import manifestFixture from "./fixtures/review-dispatch/manifest.json"
import matrixFixture from "./fixtures/review-dispatch/matrix.json"

type BunRuntime = { readonly YAML: { readonly parse: (text: string) => unknown } }

function withBun<T>(run: () => T | Promise<T>): Promise<T> {
  const runtime = globalThis as typeof globalThis & { Bun?: BunRuntime }
  const original = runtime.Bun
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

const scratchDirectories: string[] = []

async function scratchProject(): Promise<string> {
  const directory = await mkdtemp(join(tmpdir(), "uf-dispatch-status-"))
  scratchDirectories.push(directory)
  await mkdir(join(directory, ".opencode"), { recursive: true })
  await mkdir(join(directory, ".uf"), { recursive: true })
  await writeFile(join(directory, ".uf", "review-matrix.yaml"), JSON.stringify(matrixFixture))
  await writeFile(join(directory, ".uf", "reviewer-capabilities.yaml"), JSON.stringify(manifestFixture))
  await writeFile(join(directory, ".uf", "sibling-repos.yaml"), JSON.stringify({ version: 1, siblings: [] }))
  return directory
}

function noopContext(): ToolContext {
  return {
    sessionID: "status-session",
    messageID: "msg-status",
    agent: "build",
    directory: "/workspace",
    worktree: "/workspace",
    abort: new AbortController().signal,
    metadata: () => undefined,
    ask: (() => {
      throw new Error("not used")
    }) as ToolContext["ask"],
  }
}

function runFixture(agent: string, overrides?: { status?: string; findings?: number; proposals?: number }) {
  const findingsArr = Array.from({ length: overrides?.findings ?? 0 }, (_, i) => ({
    severity: "LOW",
    category: "test",
    description: `finding ${i}`,
    root_cause: `cause ${i}`,
    file: "a.ts",
    line: i + 1,
  }))
  const proposalsArr = Array.from({ length: overrides?.proposals ?? 0 }, (_, i) => ({
    information: `lesson ${i}`,
    tag: "test",
  }))

  return {
    run_id: randomUUID(),
    agent,
    source: "explicit",
    sequence: 1,
    status: overrides?.status ?? "success",
    started_at: "2026-10-08T10:00:00.000Z",
    finished_at: "2026-10-08T10:01:00.000Z",
    requested_model: "provider/standard",
    provider: "provider",
    model_id: "standard",
    variant: null,
    resolved_parent_model: null,
    resolved_parent_variant: null,
    reported_model: "provider/standard",
    model_mismatch: false,
    usage: null,
    error: null,
    workflow_verdict: null,
    findings: findingsArr,
    proposals: proposalsArr,
    text: "output",
    read_only: true,
  }
}

const cleanupDirs: string[] = []

afterEach(async () => {
  _getSubmissionStore().clear()
  _getSessionCorrelationMap().clear()
  await Promise.all(scratchDirectories.splice(0).map((d) => rm(d, { recursive: true, force: true })))
  for (const dir of cleanupDirs) {
    try {
      await rm(dir, { recursive: true, force: true })
    } catch {
      // best-effort
    }
  }
  cleanupDirs.length = 0
})

beforeEach(() => {
  _getSubmissionStore().clear()
  _getSessionCorrelationMap().clear()
})

async function setupDispatchSession(
  runs: ReturnType<typeof runFixture>[],
): Promise<string> {
  const correlationId = randomUUID()
  const dir = join(tmpdir(), "opencode", `dispatch-${correlationId}`)
  await mkdir(dir, { recursive: true })
  cleanupDirs.push(dir)
  _getSessionCorrelationMap().set("status-session", correlationId)

  for (const run of runs) {
    await writeFile(join(dir, `run-${run.agent}.json`), JSON.stringify(run, null, 2), "utf8")
  }

  return correlationId
}

describe("dispatch_status", () => {
  it("returns not_found when no dispatch session exists", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const result = await hooks.tool.dispatch_status.execute(
        {},
        noopContext(),
      )

      const data = JSON.parse(result.output)
      expect(data.status).toBe("not_found")
      expect(data.runs).toEqual([])
    })
  })

  it("returns run summaries from persisted files", async () => {
    const correlationId = await setupDispatchSession([
      runFixture("divisor-guard", { findings: 3, proposals: 1 }),
      runFixture("divisor-adversary", { status: "failed", findings: 0 }),
    ])

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const result = await hooks.tool.dispatch_status.execute(
        {},
        noopContext(),
      )

      const data = JSON.parse(result.output)
      expect(data.status).toBe("ok")
      expect(data.total_runs).toBe(2)
      expect(data.runs).toHaveLength(2)

      const guard = data.runs.find((r: { agent: string }) => r.agent === "divisor-guard")
      expect(guard.status).toBe("success")
      expect(guard.findings).toBe(3)
      expect(guard.proposals).toBe(1)

      const adversary = data.runs.find((r: { agent: string }) => r.agent === "divisor-adversary")
      expect(adversary.status).toBe("failed")
      expect(adversary.findings).toBe(0)
    })
  })

  it("handles unreadable run files gracefully", async () => {
    const correlationId = randomUUID()
    const dir = join(tmpdir(), "opencode", `dispatch-${correlationId}`)
    await mkdir(dir, { recursive: true })
    cleanupDirs.push(dir)
    _getSessionCorrelationMap().set("status-session", correlationId)
    await writeFile(join(dir, "run-divisor-broken.json"), "not valid json", "utf8")

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const result = await hooks.tool.dispatch_status.execute(
        {},
        noopContext(),
      )

      const data = JSON.parse(result.output)
      expect(data.status).toBe("ok")
      expect(data.runs).toHaveLength(1)
      expect(data.runs[0].agent).toBe("divisor-broken")
      expect(data.runs[0].status).toBe("unreadable")
    })
  })

  it("reports pending submissions from the in-process store", async () => {
    const correlationId = await setupDispatchSession([
      runFixture("divisor-guard"),
    ])

    // Simulate an in-flight submission that hasn't been harvested yet.
    const store = _getSubmissionStore()
    store.set("pending-child", {
      findings: [
        { severity: "LOW", category: "test", description: "d", root_cause: "r", file: null, line: null },
      ],
      proposals: [],
    })

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: { session: {} },
        directory: project,
        worktree: project,
      } as unknown as PluginInput)

      const result = await hooks.tool.dispatch_status.execute(
        {},
        noopContext(),
      )

      const data = JSON.parse(result.output)
      expect(data.pending_submissions).toBe(1)
    })
  })
})
