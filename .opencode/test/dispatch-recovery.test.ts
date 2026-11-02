// Permanent behavioral suite for the fix-review-dispatch-recovery
// recovery tools (task 7.4): discard_orphaned_submissions (DR-FR-001)
// and redispatch_failed_runs (DR-FR-002), plus the DR-FR-003 safety
// invariants that bind them (audit records, untouched unaffected runs,
// RD-FR-007 fail-closed preserved).
//
// This file absorbs and replaces the transient task-6.2 smoke checks
// (scratch-redispatch-smoke.test.ts, a transient scratch file that was
// never committed to the branch). Every
// scenario the smoke file pinned — input validation, not_found,
// fail-closed missing session metadata, replacement-not-duplication,
// refusal of named non-failed runs, missing_prompt, no_plan_entry,
// unreadable run files, failed-again RD-FR-025 exclusion, zero-success
// fail-closed, and the early-exit outcome — lives on here permanently.
import { chmod, mkdir, mkdtemp, readFile, readdir, rm, utimes, writeFile } from "node:fs/promises"
import { randomUUID } from "node:crypto"
import { tmpdir } from "node:os"
import { join } from "node:path"

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

type Client = PluginInput["client"]

/** Fake client: succeeds, or fails the child prompt when failPrompt is set. */
function fakeClient(options?: { failPrompt?: boolean }): Client {
  const session = {
    create: async (): Promise<unknown> => ({
      data: { id: "child-session" },
      error: undefined,
    }),
    message: async (): Promise<unknown> => ({
      data: {
        info: {
          id: "m",
          sessionID: "s",
          role: "assistant",
          providerID: "p",
          modelID: "m",
          variant: "active",
        },
        parts: [],
      },
      error: undefined,
    }),
    prompt: async (request: unknown): Promise<unknown> => {
      if (options?.failPrompt === true) {
        return { data: undefined, error: { message: "child prompt exploded" } }
      }
      const req = request as { body?: { model?: { providerID: string; modelID: string } } }
      const providerID = req.body?.model?.providerID ?? "provider"
      const modelID = req.body?.model?.modelID ?? "standard"
      return {
        data: {
          info: { role: "assistant", providerID, modelID },
          parts: [
            { id: "t", sessionID: "child-session", messageID: "m", type: "text", text: "review output", ignored: false },
          ],
        },
        error: undefined,
      }
    },
    abort: async (): Promise<unknown> => ({ data: true, error: undefined }),
  }
  return { session } as unknown as Client
}

function noopContext(): ToolContext {
  return {
    sessionID: "recovery-session",
    messageID: "msg-recovery",
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

const scratchDirectories: string[] = []
const cleanupDirs: string[] = []

async function scratchProject(matrix?: unknown): Promise<string> {
  const directory = await mkdtemp(join(tmpdir(), "uf-dispatch-recovery-"))
  scratchDirectories.push(directory)
  await mkdir(join(directory, ".opencode"), { recursive: true })
  await mkdir(join(directory, ".uf"), { recursive: true })
  await writeFile(join(directory, ".uf", "review-matrix.yaml"), JSON.stringify(matrix ?? matrixFixture))
  await writeFile(join(directory, ".uf", "reviewer-capabilities.yaml"), JSON.stringify(manifestFixture))
  await writeFile(join(directory, ".uf", "sibling-repos.yaml"), JSON.stringify({ version: 1, siblings: [] }))
  return directory
}

type RunErrorShape = { code: string; message: string; retryable: boolean }

/** Build a PersistedRunData fixture for writing to a dispatch session dir. */
function makeRun(agent: string, overrides?: {
  status?: string
  sequence?: number
  prompt?: string | null
  error?: RunErrorShape | null
  findings?: number
  proposals?: number
  postSubmissionAbort?: Record<string, unknown>
  manifestMode?: boolean
}) {
  const findingsArr = Array.from({ length: overrides?.findings ?? 0 }, (_, i) => ({
    severity: "LOW",
    category: "test",
    description: `orphaned finding ${i}`,
    root_cause: `orphan cause ${i}`,
    file: `orphan${i}.ts`,
    line: i + 1,
  }))
  const proposalsArr = Array.from({ length: overrides?.proposals ?? 0 }, (_, i) => ({
    information: `orphaned lesson ${i}`,
    tag: "orphaned",
  }))
  return {
    run_id: randomUUID(),
    agent,
    source: "explicit",
    sequence: overrides?.sequence ?? 1,
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
    // Failed/cancelled fixtures need a non-null error for the finalizer's
    // validateRunState; success fixtures keep null.
    error:
      overrides?.error ??
      (overrides?.status === "failed" || overrides?.status === "cancelled"
        ? { code: "invocation_failed", message: "original failure", retryable: true }
        : null),
    workflow_verdict: null,
    findings: findingsArr,
    proposals: proposalsArr,
    text: "original output",
    read_only: true,
    ...(overrides?.postSubmissionAbort ? { post_submission_abort: overrides.postSubmissionAbort } : {}),
    // prompt: null omits the key entirely — run files that predate prompt
    // persistence carry no prompt field at all.
    ...(overrides?.prompt === null ? {} : { prompt: overrides?.prompt ?? `Original prompt for ${agent}` }),
    // Delivery-mode work signal (RD-FR-036): omitted unless explicitly
    // requested — run files that predate the field carry no manifest_mode
    // key at all, which is the pre-field fixture shape.
    ...(overrides?.manifestMode === undefined ? {} : { manifest_mode: overrides.manifestMode }),
  }
}

/** Plan entry fixture; identity (agent + source + sequence) matches makeRun runs. */
function planEntry(agent: string, overrides?: {
  sequence?: number
  model?: string | null
  tier?: string | null
  decision?: "include" | "skip"
}) {
  return {
    agent,
    decision: overrides?.decision ?? ("include" as const),
    reason_code: "explicit",
    reason: overrides?.decision === "skip" ? "Explicitly excluded" : "Explicitly included",
    source: "explicit" as const,
    sequence: overrides?.sequence ?? 1,
    read_only: true,
    tier: overrides?.tier !== undefined ? overrides.tier : "standard",
    model: overrides?.model === undefined ? "provider/standard" : overrides.model,
    variant: null,
    validation_errors: [],
  }
}

/** Session metadata matching the shape dispatch_agent_run persists (design D5). */
function sessionMetadata(plan: unknown[]) {
  return {
    command: "review-council" as const,
    mode: "code" as const,
    full: false,
    input_context: {
      kind: "pr" as const,
      pr_number: 751,
      base_ref: "main",
      base_sha: "a".repeat(40),
      head_ref: "feature",
      head_sha: "b".repeat(40),
    },
    change_profile: {
      kind: "diff" as const,
      lines: 100,
      files: 5,
      components: 2,
      security_sensitive: false,
      user_facing: true,
      categories: ["standard" as const],
      tier: "standard" as const,
    },
    plan,
    coverage: { preflight_verdict: "PASS" as const, checks_total: 5, checks_passed: 5 },
  }
}

/** Write run files + session metadata; returns correlation_id and the dir. */
async function setupSession(
  runs: ReturnType<typeof makeRun>[],
  plan: unknown[],
  options?: { writeMetadata?: boolean; metadataJson?: string },
): Promise<{ correlationId: string; dir: string }> {
  const correlationId = randomUUID()
  const dir = join(tmpdir(), "opencode", `dispatch-${correlationId}`)
  await mkdir(dir, { recursive: true })
  cleanupDirs.push(dir)
  if (options?.writeMetadata !== false) {
    const metadata = options?.metadataJson ?? JSON.stringify(sessionMetadata(plan), null, 2)
    await writeFile(join(dir, "session-metadata.json"), metadata, "utf8")
  }
  for (const run of runs) {
    await writeFile(join(dir, `run-${run.agent}.json`), JSON.stringify(run, null, 2), "utf8")
  }
  return { correlationId, dir }
}

/** consolidate_dispatch args that internally finalize; commit matches head_sha. */
function finalizeArgs(correlationId: string) {
  return {
    correlation_id: correlationId,
    provenance: {
      branch: "opsx/fix-review-dispatch-recovery",
      commit: "b".repeat(40), // Must match input_context.head_sha
      workflow_id: "uf.review-council",
    },
  }
}

async function activate(client: Client, project: string) {
  return ReviewDispatchPlugin.server({
    client,
    directory: project,
    worktree: project,
  } as unknown as PluginInput)
}

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

describe("discard_orphaned_submissions", () => {
  it("rejects invalid input", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)
      const emptyAgents = await hooks.tool.discard_orphaned_submissions.execute(
        { correlation_id: randomUUID(), agents: [] },
        noopContext(),
      )
      expect(emptyAgents.output).toContain("discard_orphaned_submissions validation failed")
      const badId = await hooks.tool.discard_orphaned_submissions.execute({ correlation_id: "not-a-uuid" }, noopContext())
      expect(badId.output).toContain("discard_orphaned_submissions validation failed")
    })
  })

  it("discards a failed run's orphaned submissions, records the audit block, and finalizes from the successful runs (DR-FR-001 Discard then finalize)", async () => {
    // Issue #751 shape: three successful reviewers plus one failed run
    // (post-submission abort) whose harvested findings and lesson
    // proposal made consolidation warn and blocked finalization.
    const failedRun = makeRun("divisor-scribe", {
      status: "failed",
      findings: 2,
      proposals: 1,
      error: { code: "child_message_error", message: "session aborted during final response generation", retryable: false },
    })
    const guardRun = makeRun("divisor-guard", { findings: 1 })
    const { correlationId, dir } = await setupSession(
      [guardRun, makeRun("divisor-adversary"), makeRun("divisor-architect"), failedRun],
      [
        planEntry("divisor-guard"),
        planEntry("divisor-adversary", { sequence: 1 }),
        planEntry("divisor-architect", { sequence: 1 }),
        planEntry("divisor-scribe", { sequence: 1 }),
      ],
    )
    const successFiles = ["divisor-guard", "divisor-adversary", "divisor-architect"].map((agent) =>
      readFile(join(dir, `run-${agent}.json`), "utf8"),
    )
    const successBefore = await Promise.all(successFiles)

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const result = await hooks.tool.discard_orphaned_submissions.execute({ correlation_id: correlationId }, noopContext())
      const data = JSON.parse(result.output)
      expect(data.status).toBe("ok")
      expect(data.correlation_id).toBe(correlationId)
      expect(data.total_runs).toBe(1)
      expect(data.discarded_runs).toBe(1)
      expect(data.refused_runs).toBe(0)
      expect(data.runs).toHaveLength(1)
      expect(data.runs[0].agent).toBe("divisor-scribe")
      expect(data.runs[0].run_id).toBe(failedRun.run_id)
      expect(data.runs[0].status).toBe("discarded")
      expect(data.runs[0].discarded_findings).toBe(2)
      expect(data.runs[0].discarded_proposals).toBe(1)

      // The failed run's persisted record no longer carries findings or
      // proposals; everything else is preserved verbatim (DR-FR-003).
      const discarded = JSON.parse(await readFile(join(dir, "run-divisor-scribe.json"), "utf8"))
      expect(discarded.findings).toEqual([])
      expect(discarded.proposals).toEqual([])
      expect(discarded.run_id).toBe(failedRun.run_id)
      expect(discarded.status).toBe("failed")
      expect(discarded.error).toEqual(failedRun.error)
      expect(discarded.text).toBe("original output")
      expect(discarded.prompt).toBe("Original prompt for divisor-scribe")

      // The discard is recorded with the run's agent, run id, and counts
      // (DR-FR-001 "Recovery is auditable").
      const audit = discarded.discarded_submissions
      expect(audit.agent).toBe("divisor-scribe")
      expect(audit.run_id).toBe(failedRun.run_id)
      expect(audit.discarded_findings).toBe(2)
      expect(audit.discarded_proposals).toBe(1)
      expect(audit.discarded_at).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$/)
      expect(audit.note).toContain("DR-FR-001")

      // Successful runs are untouched — byte-identical files.
      const successAfter = await Promise.all(successFiles)
      expect(successAfter).toEqual(successBefore)

      // Subsequent consolidation finalizes from the three successful runs
      // without an orphaned-submission warning for the discarded run.
      const payload = JSON.parse(
        (await hooks.tool.consolidate_dispatch.execute({ correlation_id: correlationId }, noopContext())).output,
      )
      expect(payload.verdict).toBe("APPROVE WITH ADVISORIES")
      expect(payload.run_counts.total).toBe(4)
      expect(payload.run_counts.success).toBe(3)
      expect(payload.findings).toHaveLength(1)
      expect(payload.findings[0].run_ids).toEqual([guardRun.run_id])
      expect(payload.warnings).toBeUndefined()

      const finalized = JSON.parse(
        (await hooks.tool.consolidate_dispatch.execute(finalizeArgs(correlationId), noopContext())).output,
      )
      expect(finalized.status).toBe("success")
      expect(finalized.operation_verdict).toBe("APPROVE WITH ADVISORIES")
      expect(finalized.errors).toEqual([])
      expect(finalized.warnings).toBeUndefined()
    })
  })

  it("refuses a named successful run and changes nothing (DR-FR-001 Discard refuses successful runs)", async () => {
    const successRun = makeRun("divisor-guard")
    const failedRun = makeRun("divisor-scribe", { status: "failed", findings: 1, proposals: 1 })
    const { correlationId, dir } = await setupSession(
      [successRun, failedRun],
      [planEntry("divisor-guard"), planEntry("divisor-scribe", { sequence: 1 })],
    )
    const successBefore = await readFile(join(dir, "run-divisor-guard.json"), "utf8")
    const failedBefore = await readFile(join(dir, "run-divisor-scribe.json"), "utf8")

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const result = await hooks.tool.discard_orphaned_submissions.execute(
        { correlation_id: correlationId, agents: ["divisor-guard"] },
        noopContext(),
      )
      const data = JSON.parse(result.output)
      expect(data.status).toBe("ok")
      expect(data.discarded_runs).toBe(0)
      expect(data.refused_runs).toBe(1)
      expect(data.total_runs).toBe(1)
      expect(data.runs[0].agent).toBe("divisor-guard")
      expect(data.runs[0].run_id).toBe(successRun.run_id)
      expect(data.runs[0].status).toBe("refused")
      expect(data.runs[0].discarded_findings).toBe(0)
      expect(data.runs[0].reason).toContain("successful runs are never discard targets")

      // The request is rejected and nothing changes: both files stay
      // byte-identical (the unnamed failed run is out of scope entirely).
      expect(await readFile(join(dir, "run-divisor-guard.json"), "utf8")).toBe(successBefore)
      expect(await readFile(join(dir, "run-divisor-scribe.json"), "utf8")).toBe(failedBefore)
    })
  })

  it("is idempotent: a repeated discard reports no_orphaned_submissions and preserves the audit block", async () => {
    const failedRun = makeRun("divisor-scribe", { status: "failed", findings: 1, proposals: 1 })
    const { correlationId, dir } = await setupSession([failedRun], [planEntry("divisor-scribe", { sequence: 1 })])

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const first = JSON.parse(
        (await hooks.tool.discard_orphaned_submissions.execute({ correlation_id: correlationId }, noopContext())).output,
      )
      expect(first.runs[0].status).toBe("discarded")

      const afterFirst = await readFile(join(dir, "run-divisor-scribe.json"), "utf8")

      const second = JSON.parse(
        (await hooks.tool.discard_orphaned_submissions.execute({ correlation_id: correlationId }, noopContext())).output,
      )
      expect(second.status).toBe("ok")
      expect(second.runs[0].status).toBe("no_orphaned_submissions")
      expect(second.runs[0].run_id).toBe(failedRun.run_id)
      expect(second.discarded_runs).toBe(0)
      expect(second.refused_runs).toBe(0)

      // The file is not rewritten: the first call's audit block survives
      // verbatim rather than being overwritten or duplicated.
      expect(await readFile(join(dir, "run-divisor-scribe.json"), "utf8")).toBe(afterFirst)
      const discarded = JSON.parse(afterFirst)
      expect(discarded.discarded_submissions.run_id).toBe(failedRun.run_id)
      expect(discarded.discarded_submissions.discarded_findings).toBe(1)
    })
  })

  it("targets every non-success run when agents is omitted", async () => {
    const guardRun = makeRun("divisor-guard")
    const scribeRun = makeRun("divisor-scribe", { status: "failed", findings: 2, proposals: 1 })
    const adversaryRun = makeRun("divisor-adversary", { status: "failed" })
    const architectRun = makeRun("divisor-architect", { status: "cancelled", findings: 1 })
    const { correlationId, dir } = await setupSession(
      [guardRun, scribeRun, adversaryRun, architectRun],
      [
        planEntry("divisor-guard"),
        planEntry("divisor-scribe", { sequence: 1 }),
        planEntry("divisor-adversary", { sequence: 1 }),
        planEntry("divisor-architect", { sequence: 1 }),
      ],
    )
    const guardBefore = await readFile(join(dir, "run-divisor-guard.json"), "utf8")
    const adversaryBefore = await readFile(join(dir, "run-divisor-adversary.json"), "utf8")

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const result = await hooks.tool.discard_orphaned_submissions.execute({ correlation_id: correlationId }, noopContext())
      const data = JSON.parse(result.output)
      // The successful run is out of scope entirely; every non-success
      // run is examined (failed, failed-empty, and cancelled alike).
      expect(data.total_runs).toBe(3)
      expect(data.discarded_runs).toBe(2)
      expect(data.runs.map((r: { agent: string }) => r.agent).sort()).toEqual([
        "divisor-adversary",
        "divisor-architect",
        "divisor-scribe",
      ])

      const scribe = data.runs.find((r: { agent: string }) => r.agent === "divisor-scribe")
      expect(scribe.status).toBe("discarded")
      expect(scribe.discarded_findings).toBe(2)
      expect(scribe.discarded_proposals).toBe(1)

      const adversary = data.runs.find((r: { agent: string }) => r.agent === "divisor-adversary")
      expect(adversary.status).toBe("no_orphaned_submissions")

      const architect = data.runs.find((r: { agent: string }) => r.agent === "divisor-architect")
      expect(architect.status).toBe("discarded")
      expect(architect.discarded_findings).toBe(1)

      // A non-success run without submissions is left untouched (it emits
      // no orphaned-submission warning), and the successful run never was
      // a target — both files stay byte-identical.
      expect(await readFile(join(dir, "run-divisor-guard.json"), "utf8")).toBe(guardBefore)
      expect(await readFile(join(dir, "run-divisor-adversary.json"), "utf8")).toBe(adversaryBefore)

      // The cancelled run's submissions were discarded on disk.
      const cancelled = JSON.parse(await readFile(join(dir, "run-divisor-architect.json"), "utf8"))
      expect(cancelled.findings).toEqual([])
      expect(cancelled.discarded_submissions.discarded_findings).toBe(1)
    })
  })

  it("returns not_found for an unknown correlation_id and for a session with no run files", async () => {
    await setupSession([], [])
    const unknownId = randomUUID()
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const unknown = JSON.parse(
        (await hooks.tool.discard_orphaned_submissions.execute({ correlation_id: unknownId }, noopContext())).output,
      )
      expect(unknown.status).toBe("not_found")
      expect(unknown.correlation_id).toBe(unknownId)
      expect(unknown.runs).toEqual([])
      expect(unknown.total_runs).toBe(0)
      expect(unknown.discarded_runs).toBe(0)
      expect(unknown.refused_runs).toBe(0)

      // A session directory with metadata but no run files is equally
      // not_found: there is nothing to discard.
      const { correlationId } = await setupSession([], [])
      const empty = JSON.parse(
        (await hooks.tool.discard_orphaned_submissions.execute({ correlation_id: correlationId }, noopContext())).output,
      )
      expect(empty.status).toBe("not_found")
      expect(empty.runs).toEqual([])
    })
  })

  it("reports unreadable run files in scope and never rewrites them", async () => {
    // Session A: one unreadable run file, addressed by name.
    const correlationA = randomUUID()
    const dirA = join(tmpdir(), "opencode", `dispatch-${correlationA}`)
    await mkdir(dirA, { recursive: true })
    cleanupDirs.push(dirA)
    await writeFile(join(dirA, "session-metadata.json"), JSON.stringify(sessionMetadata([]), null, 2), "utf8")
    await writeFile(join(dirA, "run-divisor-broken.json"), "{not json", "utf8")

    // Session B: a valid failed run plus an unreadable run file.
    const { correlationId: correlationB } = await setupSession(
      [makeRun("divisor-scribe", { status: "failed", findings: 1, proposals: 1 })],
      [planEntry("divisor-scribe", { sequence: 1 })],
    )
    const dirB = join(tmpdir(), "opencode", `dispatch-${correlationB}`)
    await writeFile(join(dirB, "run-divisor-broken.json"), "{not json", "utf8")

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      // Named: the unreadable file is reported and left untouched.
      const named = JSON.parse(
        (
          await hooks.tool.discard_orphaned_submissions.execute(
            { correlation_id: correlationA, agents: ["divisor-broken"] },
            noopContext(),
          )
        ).output,
      )
      expect(named.status).toBe("ok")
      expect(named.runs).toHaveLength(1)
      expect(named.runs[0].agent).toBe("divisor-broken")
      expect(named.runs[0].run_id).toBeNull()
      expect(named.runs[0].status).toBe("unreadable")
      expect(named.runs[0].reason).toContain("could not be read or parsed")
      expect(await readFile(join(dirA, "run-divisor-broken.json"), "utf8")).toBe("{not json")

      // Out-of-filter unreadable files are skipped entirely, and a named
      // agent with no run file reports not_found.
      const filtered = JSON.parse(
        (
          await hooks.tool.discard_orphaned_submissions.execute(
            { correlation_id: correlationA, agents: ["divisor-scribe"] },
            noopContext(),
          )
        ).output,
      )
      expect(filtered.runs).toHaveLength(1)
      expect(filtered.runs[0].agent).toBe("divisor-scribe")
      expect(filtered.runs[0].status).toBe("not_found")

      // Agents omitted: the unreadable file is reported alongside the
      // successful discard of the valid failed run.
      const unfiltered = JSON.parse(
        (await hooks.tool.discard_orphaned_submissions.execute({ correlation_id: correlationB }, noopContext())).output,
      )
      expect(unfiltered.discarded_runs).toBe(1)
      expect(unfiltered.runs.map((r: { status: string }) => r.status).sort()).toEqual(["discarded", "unreadable"])
      expect(await readFile(join(dirB, "run-divisor-broken.json"), "utf8")).toBe("{not json")
    })
  })

  it("reports a per-run error when the rewrite fails and leaves the file intact", async () => {
    const failedRun = makeRun("divisor-scribe", { status: "failed", findings: 1, proposals: 1 })
    const { correlationId, dir } = await setupSession([failedRun], [planEntry("divisor-scribe", { sequence: 1 })])
    const runPath = join(dir, "run-divisor-scribe.json")
    const before = await readFile(runPath, "utf8")
    // The rewrite is atomic (write-temp-then-rename), so a read-only run
    // file no longer blocks it: rename only needs directory write
    // permission. A read-only session directory blocks the temp write
    // itself, failing the rewrite while leaving the original intact.
    await chmod(dir, 0o555)
    try {
      await withBun(async () => {
        const project = await scratchProject()
        const hooks = await activate(fakeClient(), project)

        const result = await hooks.tool.discard_orphaned_submissions.execute(
          { correlation_id: correlationId },
          noopContext(),
        )
        const data = JSON.parse(result.output)
        expect(data.status).toBe("ok")
        expect(data.discarded_runs).toBe(0)
        expect(data.runs).toHaveLength(1)
        expect(data.runs[0].agent).toBe("divisor-scribe")
        expect(data.runs[0].status).toBe("error")
        expect(data.runs[0].discarded_findings).toBe(0)
        expect(data.runs[0].reason).toContain("failed to rewrite run-divisor-scribe.json")

        // The failed rewrite left the original record fully intact.
        expect(await readFile(runPath, "utf8")).toBe(before)
      })
    } finally {
      await chmod(dir, 0o755)
    }
  })

  it("keeps the zero-success fail-closed path after discarding (DR-FR-003 Recovery preserves fail-closed behavior)", async () => {
    // Every run failed; the operator discards the orphaned submissions.
    // Recovery MUST NOT manufacture an approving result from a
    // zero-success dispatch (RD-FR-007).
    const failedRun = makeRun("divisor-guard", {
      status: "failed",
      findings: 1,
      proposals: 1,
      error: { code: "child_message_error", message: "session aborted during final response generation", retryable: false },
    })
    const { correlationId } = await setupSession([failedRun], [planEntry("divisor-guard")])

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const discarded = JSON.parse(
        (await hooks.tool.discard_orphaned_submissions.execute({ correlation_id: correlationId }, noopContext())).output,
      )
      expect(discarded.discarded_runs).toBe(1)

      const finalized = JSON.parse(
        (await hooks.tool.consolidate_dispatch.execute(finalizeArgs(correlationId), noopContext())).output,
      )
      // fix-dispatch-verdict-derivation (issue #764): zero-success
      // finalization persists an INCONCLUSIVE artifact instead of being
      // rejected by a value mismatch — the verdict stays fail-closed.
      expect(finalized.status).toBe("success")
      expect(finalized.operation_verdict).toBe("INCONCLUSIVE")
      expect(finalized.errors).toEqual([])
      // The submissions are gone, so no orphaned-submission warning rides
      // along — but the verdict stays fail-closed.
      expect(finalized.warnings).toBeUndefined()
    })
  })
})

describe("redispatch_failed_runs", () => {
  it("rejects invalid input", async () => {
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)
      const emptyAgents = await hooks.tool.redispatch_failed_runs.execute(
        { correlation_id: randomUUID(), agents: [] },
        noopContext(),
      )
      expect(emptyAgents.output).toContain("redispatch_failed_runs validation failed")
      const badTimeout = await hooks.tool.redispatch_failed_runs.execute(
        { correlation_id: randomUUID(), timeout: 0 },
        noopContext(),
      )
      expect(badTimeout.output).toContain("redispatch_failed_runs validation failed")
    })
  })

  it("returns not_found for an unknown correlation_id and for a session with no run files", async () => {
    await setupSession([], [])
    const unknownId = randomUUID()
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const unknown = JSON.parse(
        (await hooks.tool.redispatch_failed_runs.execute({ correlation_id: unknownId }, noopContext())).output,
      )
      expect(unknown.status).toBe("not_found")
      expect(unknown.correlation_id).toBe(unknownId)
      expect(unknown.runs).toEqual([])
      expect(unknown.total_runs).toBe(0)
      expect(unknown.redispatched_runs).toBe(0)
      expect(unknown.refused_runs).toBe(0)

      // A session directory with metadata but no run files is equally
      // not_found: there is nothing to re-run.
      const { correlationId } = await setupSession([], [])
      const empty = JSON.parse(
        (await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())).output,
      )
      expect(empty.status).toBe("not_found")
      expect(empty.runs).toEqual([])
    })
  })

  it("fails closed when session metadata is missing", async () => {
    const { correlationId, dir } = await setupSession([makeRun("divisor-guard", { status: "failed" })], [planEntry("divisor-guard")], {
      writeMetadata: false,
    })
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)
      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
      expect(result.output).toContain("redispatch_failed_runs error: no session metadata found")

      // RD-FR-031: the error path still settles the operation window —
      // one record, closed with status "error" and zero-effect counts.
      const opFiles = (await readdir(dir)).filter((name) => /^operation-.*\.json$/.test(name))
      expect(opFiles).toHaveLength(1)
      const record = JSON.parse(await readFile(join(dir, opFiles[0]!), "utf8")) as Record<string, unknown>
      expect(record.kind).toBe("redispatch")
      expect(typeof record.operation_id).toBe("string")
      expect(typeof record.started_at).toBe("string")
      expect(typeof record.completed_at).toBe("string")
      expect(record.status).toBe("error")
      expect(record.run_counts).toMatchObject({ redispatched: 0, refused: 0, restored: 0 })
    })
  })

  it("fails closed when session metadata is invalid", async () => {
    const { correlationId, dir } = await setupSession([makeRun("divisor-guard", { status: "failed" })], [planEntry("divisor-guard")], {
      metadataJson: JSON.stringify({ nonsense: true }),
    })
    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)
      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
      expect(result.output).toContain("redispatch_failed_runs error: invalid session metadata")

      // RD-FR-031: the error path still settles the operation window —
      // one record, closed with status "error" and zero-effect counts.
      const opFiles = (await readdir(dir)).filter((name) => /^operation-.*\.json$/.test(name))
      expect(opFiles).toHaveLength(1)
      const record = JSON.parse(await readFile(join(dir, opFiles[0]!), "utf8")) as Record<string, unknown>
      expect(record.kind).toBe("redispatch")
      expect(typeof record.operation_id).toBe("string")
      expect(typeof record.started_at).toBe("string")
      expect(typeof record.completed_at).toBe("string")
      expect(record.status).toBe("error")
      expect(record.run_counts).toMatchObject({ redispatched: 0, refused: 0, restored: 0 })
    })
  })

  it("replaces (not duplicates) the failed run and consolidates from the replacement (DR-FR-002 Re-run a failed reviewer)", async () => {
    const guardRun = makeRun("divisor-guard", { status: "failed", sequence: 1 })
    const adversaryRun = makeRun("divisor-adversary", { status: "success", sequence: 2 })
    const { correlationId, dir } = await setupSession(
      [guardRun, adversaryRun],
      [planEntry("divisor-guard"), planEntry("divisor-adversary", { sequence: 2 })],
    )
    const adversaryBefore = await readFile(join(dir, "run-divisor-adversary.json"), "utf8")

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const result = await hooks.tool.redispatch_failed_runs.execute(
        { correlation_id: correlationId, timeout: 600_000 },
        noopContext(),
      )
      const data = JSON.parse(result.output)
      expect(data.status).toBe("ok")
      expect(data.total_runs).toBe(1)
      expect(data.redispatched_runs).toBe(1)
      expect(data.refused_runs).toBe(0)
      expect(data.runs[0].agent).toBe("divisor-guard")
      expect(data.runs[0].run_id).toBe(guardRun.run_id)
      expect(data.runs[0].outcome).toBe("redispatched")
      expect(data.runs[0].new_status).toBe("success")
      expect(data.runs[0].new_run_id).not.toBe(guardRun.run_id)
      expect(data.runs[0].superseded_run_file).toBe(`superseded/run-divisor-guard-${guardRun.run_id}.json`)

      // Replacement: exactly one active run file per agent, prompt reused.
      const activeFiles = (await readdir(dir)).filter((n) => n.startsWith("run-") && n.endsWith(".json")).sort()
      expect(activeFiles).toEqual(["run-divisor-adversary.json", "run-divisor-guard.json"])
      const replacement = JSON.parse(await readFile(join(dir, "run-divisor-guard.json"), "utf8"))
      expect(replacement.run_id).toBe(data.runs[0].new_run_id)
      expect(replacement.status).toBe("success")
      expect(replacement.prompt).toBe("Original prompt for divisor-guard")
      expect(replacement.sequence).toBe(1)
      expect(replacement.source).toBe("explicit")
      expect(replacement.requested_model).toBe("provider/standard")
      expect(replacement.findings).toEqual([])

      // Superseded archive holds the prior run verbatim — byte-identical
      // to the file the original dispatch wrote (DR-FR-003).
      const archivedBytes = await readFile(join(dir, "superseded", `run-divisor-guard-${guardRun.run_id}.json`), "utf8")
      expect(archivedBytes).toBe(JSON.stringify(guardRun, null, 2))

      // The successful run is untouched (DR-FR-003: unaffected runs
      // unchanged) — byte-identical file.
      expect(await readFile(join(dir, "run-divisor-adversary.json"), "utf8")).toBe(adversaryBefore)

      // superseded/ is outside the run-*.json glob: dispatch_status and
      // consolidate still see exactly the two active runs.
      const status = JSON.parse(
        (await hooks.tool.dispatch_status.execute({ correlation_id: correlationId }, noopContext())).output,
      )
      expect(status.status).toBe("ok")
      expect(status.total_runs).toBe(2)

      const payload = JSON.parse(
        (await hooks.tool.consolidate_dispatch.execute({ correlation_id: correlationId }, noopContext())).output,
      )
      expect(payload.run_counts.total).toBe(2)
      expect(payload.run_counts.success).toBe(2)
      expect(payload.warnings).toBeUndefined()

      // Consolidation uses the original success plus the replacement, and
      // the dispatch finalizes.
      const finalized = JSON.parse(
        (await hooks.tool.consolidate_dispatch.execute(finalizeArgs(correlationId), noopContext())).output,
      )
      expect(finalized.status).toBe("success")
      expect(finalized.operation_verdict).toBe("APPROVE")
      expect(finalized.errors).toEqual([])
    })
  })

  it("re-dispatches a failed host-source run through host resolution with host-run provenance (RD-FR-037 Re-dispatched host run reuses host resolution)", async () => {
    // A failed host-fallback run persisted with its original prompt and a
    // host-fallback plan entry (source "host", null model/tier/variant —
    // the baseRuns shape). resolvePlanEntryRuntime passes neither model nor
    // tier, so the re-run MUST resolve the parent session model through the
    // same shared-core host branch and persist host-run provenance.
    const failedHostRun = {
      ...makeRun("divisor-guard", { status: "failed" }),
      source: "host" as const,
      requested_model: null,
      provider: null,
      model_id: null,
      variant: null,
      resolved_parent_model: "p/m",
      resolved_parent_variant: "active",
      reported_model: null,
    }
    const hostPlanEntry = {
      ...planEntry("divisor-guard"),
      reason_code: "host-fallback",
      reason: "no explicit or advisor run; use the active host model",
      source: "host" as const,
      tier: null,
      model: null,
      variant: null,
    }
    const { correlationId, dir } = await setupSession([failedHostRun], [hostPlanEntry])

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const result = await hooks.tool.redispatch_failed_runs.execute(
        { correlation_id: correlationId },
        noopContext(),
      )
      const data = JSON.parse(result.output)
      expect(data.status).toBe("ok")
      expect(data.total_runs).toBe(1)
      expect(data.redispatched_runs).toBe(1)
      expect(data.runs[0].agent).toBe("divisor-guard")
      expect(data.runs[0].outcome).toBe("redispatched")
      expect(data.runs[0].new_status).toBe("success")

      // The replacement run carries host-run provenance: null requested
      // model, provider, model_id, and variant; the resolved parent model
      // from the fake client's host lookup (p/m, active). The persisted
      // prompt is reused verbatim.
      const replacement = JSON.parse(await readFile(join(dir, "run-divisor-guard.json"), "utf8"))
      expect(replacement.status).toBe("success")
      expect(replacement.source).toBe("host")
      expect(replacement.prompt).toBe("Original prompt for divisor-guard")
      expect(replacement.requested_model).toBeNull()
      expect(replacement.provider).toBeNull()
      expect(replacement.model_id).toBeNull()
      expect(replacement.variant).toBeNull()
      expect(replacement.resolved_parent_model).toBe("p/m")
      expect(replacement.resolved_parent_variant).toBe("active")
      expect(replacement.reported_model).toBe("p/m")
      expect(replacement.model_mismatch).toBe(false)

      // The recovered session finalizes cleanly — the #901 recovery path
      // that DR-FR-002 could not reach while host runs were unfinalizable.
      const finalized = JSON.parse(
        (await hooks.tool.consolidate_dispatch.execute(finalizeArgs(correlationId), noopContext())).output,
      )
      expect(finalized.status).toBe("success")
      expect(finalized.operation_verdict).toBe("APPROVE")
      expect(finalized.errors).toEqual([])
    })
  })

  it("refuses named successful runs — including reclassified successes — re-runs a named cancelled run, and reports not_found for a missing agent", async () => {
    // A post-submission-abort run reclassified to success (RD-FR-026)
    // says "success" on disk and MUST never be re-run.
    const guardRun = makeRun("divisor-guard", {
      status: "success",
      sequence: 1,
      postSubmissionAbort: {
        original_status: "failed",
        error_code: "child_message_error",
        error_message: "session aborted during final response generation",
        note: "Run reclassified from failed to success per RD-FR-026.",
      },
    })
    const adversaryRun = makeRun("divisor-adversary", { status: "success", sequence: 2 })
    const architectRun = makeRun("divisor-architect", { status: "cancelled", sequence: 3 })
    const { correlationId, dir } = await setupSession(
      [guardRun, adversaryRun, architectRun],
      [
        planEntry("divisor-guard"),
        planEntry("divisor-adversary", { sequence: 2 }),
        planEntry("divisor-architect", { sequence: 3 }),
      ],
    )
    const before = await Promise.all(
      ["divisor-guard", "divisor-adversary", "divisor-architect"].map((agent) =>
        readFile(join(dir, `run-${agent}.json`), "utf8"),
      ),
    )

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const result = await hooks.tool.redispatch_failed_runs.execute(
        { correlation_id: correlationId, agents: ["divisor-guard", "divisor-adversary", "divisor-architect", "divisor-scribe"] },
        noopContext(),
      )
      const data = JSON.parse(result.output)
      expect(data.status).toBe("ok")
      expect(data.redispatched_runs).toBe(1)
      expect(data.refused_runs).toBe(2)
      expect(data.total_runs).toBe(4)

      const guard = data.runs.find((r: { agent: string }) => r.agent === "divisor-guard")
      expect(guard.outcome).toBe("refused")
      expect(guard.run_id).toBe(guardRun.run_id)
      expect(guard.reason).toContain('run status is "success"')

      const adversary = data.runs.find((r: { agent: string }) => r.agent === "divisor-adversary")
      expect(adversary.outcome).toBe("refused")
      expect(adversary.run_id).toBe(adversaryRun.run_id)

      const architect = data.runs.find((r: { agent: string }) => r.agent === "divisor-architect")
      expect(architect.outcome).toBe("redispatched")
      expect(architect.run_id).toBe(architectRun.run_id)
      expect(architect.new_run_id).not.toBe(architectRun.run_id)

      const missing = data.runs.find((r: { agent: string }) => r.agent === "divisor-scribe")
      expect(missing.outcome).toBe("not_found")
      expect(missing.run_id).toBeNull()

      // The refused successes are untouched (byte-identical); the named
      // cancelled run was re-run, so its active file holds the replacement
      // and the prior cancelled record is archived under superseded/.
      const after = await Promise.all(
        ["divisor-guard", "divisor-adversary"].map((agent) =>
          readFile(join(dir, `run-${agent}.json`), "utf8"),
        ),
      )
      expect(after).toEqual([before[0], before[1]])

      const replacement = JSON.parse(await readFile(join(dir, "run-divisor-architect.json"), "utf8"))
      expect(replacement.run_id).not.toBe(architectRun.run_id)
      expect(replacement.status).toBe("success")
      const archived = JSON.parse(
        await readFile(join(dir, "superseded", `run-divisor-architect-${architectRun.run_id}.json`), "utf8"),
      )
      expect(archived.status).toBe("cancelled")
    })
  })

  it("reports missing_prompt for run files that predate prompt persistence", async () => {
    const guardRun = makeRun("divisor-guard", { status: "failed", prompt: null })
    const { correlationId, dir } = await setupSession([guardRun], [planEntry("divisor-guard")])

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
      const data = JSON.parse(result.output)
      expect(data.redispatched_runs).toBe(0)
      expect(data.runs[0].outcome).toBe("missing_prompt")
      expect(data.runs[0].run_id).toBe(guardRun.run_id)
      expect(data.runs[0].new_run_id).toBeNull()
      expect(data.runs[0].superseded_run_file).toBeNull()
      expect(data.runs[0].reason).toContain("predates prompt persistence")

      // No archive move: the prior file is still the active run file.
      const guard = JSON.parse(await readFile(join(dir, "run-divisor-guard.json"), "utf8"))
      expect(guard.run_id).toBe(guardRun.run_id)
      expect(await readdir(join(dir, "superseded")).catch(() => null)).toBeNull()
    })
  })

  it("reports no_plan_entry when no included plan entry matches the run identity", async () => {
    // Two failure shapes: a run whose sequence no longer matches any
    // plan entry, and a run whose matching entry flipped to "skip".
    const guardRun = makeRun("divisor-guard", { status: "failed", sequence: 9 })
    const adversaryRun = makeRun("divisor-adversary", { status: "failed", sequence: 2 })
    const { correlationId, dir } = await setupSession(
      [guardRun, adversaryRun],
      [
        planEntry("divisor-adversary", { sequence: 2, decision: "skip" }),
        planEntry("divisor-scribe", { sequence: 3 }),
      ],
    )

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
      const data = JSON.parse(result.output)
      expect(data.redispatched_runs).toBe(0)
      expect(data.runs).toHaveLength(2)

      const guard = data.runs.find((r: { agent: string }) => r.agent === "divisor-guard")
      expect(guard.outcome).toBe("no_plan_entry")
      expect(guard.run_id).toBe(guardRun.run_id)
      expect(guard.reason).toContain("no included plan entry matches divisor-guard/explicit/9")

      const adversary = data.runs.find((r: { agent: string }) => r.agent === "divisor-adversary")
      expect(adversary.outcome).toBe("no_plan_entry")
      expect(adversary.reason).toContain("no included plan entry matches divisor-adversary/explicit/2")

      // Fail-soft per run: both prior files stay active and untouched.
      const guardFile = JSON.parse(await readFile(join(dir, "run-divisor-guard.json"), "utf8"))
      expect(guardFile.run_id).toBe(guardRun.run_id)
      const adversaryFile = JSON.parse(await readFile(join(dir, "run-divisor-adversary.json"), "utf8"))
      expect(adversaryFile.run_id).toBe(adversaryRun.run_id)
      expect(await readdir(join(dir, "superseded")).catch(() => null)).toBeNull()
    })
  })

  it("reports unreadable run files named in the agents filter and leaves them untouched", async () => {
    const correlationId = randomUUID()
    const dir = join(tmpdir(), "opencode", `dispatch-${correlationId}`)
    await mkdir(dir, { recursive: true })
    cleanupDirs.push(dir)
    await writeFile(join(dir, "session-metadata.json"), JSON.stringify(sessionMetadata([planEntry("divisor-guard")]), null, 2), "utf8")
    await writeFile(join(dir, "run-divisor-guard.json"), "{not json", "utf8")

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const named = JSON.parse(
        (
          await hooks.tool.redispatch_failed_runs.execute(
            { correlation_id: correlationId, agents: ["divisor-guard"] },
            noopContext(),
          )
        ).output,
      )
      expect(named.status).toBe("ok")
      expect(named.runs).toHaveLength(1)
      expect(named.runs[0].agent).toBe("divisor-guard")
      expect(named.runs[0].outcome).toBe("unreadable")
      expect(named.runs[0].run_id).toBeNull()
      expect(named.runs[0].reason).toContain("could not be read or parsed")
      expect(await readFile(join(dir, "run-divisor-guard.json"), "utf8")).toBe("{not json")

      // An out-of-filter unreadable file is skipped entirely; a named
      // agent with no run file reports not_found.
      const filtered = JSON.parse(
        (
          await hooks.tool.redispatch_failed_runs.execute(
            { correlation_id: correlationId, agents: ["divisor-scribe"] },
            noopContext(),
          )
        ).output,
      )
      expect(filtered.runs).toHaveLength(1)
      expect(filtered.runs[0].agent).toBe("divisor-scribe")
      expect(filtered.runs[0].outcome).toBe("not_found")
    })
  })

  it("excludes a failed-again re-run's submissions under RD-FR-025 with a warning (DR-FR-002 Re-run fails again)", async () => {
    const guardRun = makeRun("divisor-guard", { status: "failed", sequence: 1 })
    const adversaryRun = makeRun("divisor-adversary", { status: "success", sequence: 2 })
    const { correlationId, dir } = await setupSession(
      [guardRun, adversaryRun],
      [planEntry("divisor-guard"), planEntry("divisor-adversary", { sequence: 2 })],
    )

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient({ failPrompt: true }), project)

      // Seed the child submission store so the failed-again re-run
      // harvests orphaned submissions (the issue #751 shape).
      _getSubmissionStore().set("child-session", {
        findings: [
          { severity: "LOW", category: "test", description: "orphaned", root_cause: "cause", file: "a.ts", line: 1 },
        ],
        proposals: [{ information: "orphaned lesson", tag: "test" }],
      })

      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
      const data = JSON.parse(result.output)
      expect(data.redispatched_runs).toBe(1)
      expect(data.runs[0].outcome).toBe("redispatched")
      expect(data.runs[0].new_status).toBe("failed")
      expect(data.runs[0].findings).toBe(1)
      expect(data.runs[0].proposals).toBe(1)
      expect(data.runs[0].error.code).toBe("child_prompt_failed")

      const replacement = JSON.parse(await readFile(join(dir, "run-divisor-guard.json"), "utf8"))
      expect(replacement.status).toBe("failed")
      expect(replacement.findings).toHaveLength(1)

      // Consolidation excludes the failed-again submissions with a
      // warning naming the NEW run id.
      const payload = JSON.parse(
        (await hooks.tool.consolidate_dispatch.execute({ correlation_id: correlationId }, noopContext())).output,
      )
      expect(payload.run_counts.success).toBe(1)
      expect(payload.findings).toHaveLength(0)
      expect(payload.proposals).toBeUndefined()
      expect(payload.warnings).toHaveLength(1)
      expect(payload.warnings[0]).toEqual({
        type: "orphaned_submissions_excluded",
        agent: "divisor-guard",
        run_id: data.runs[0].new_run_id,
        discarded_findings: 1,
        discarded_proposals: 1,
      })

      // RD-FR-025: the payload still passes finalizer validation (1
      // success) — the dispatch is finalizable, not bricked.
      const finalized = JSON.parse(
        (await hooks.tool.consolidate_dispatch.execute(finalizeArgs(correlationId), noopContext())).output,
      )
      expect(finalized.status).toBe("success")
      expect(finalized.operation_verdict).toBe("APPROVE")
      expect(finalized.errors).toEqual([])
    })
  })

  it("keeps the zero-success fail-closed path (RD-FR-007) when every re-run fails again", async () => {
    const guardRun = makeRun("divisor-guard", { status: "failed", sequence: 1 })
    const { correlationId } = await setupSession([guardRun], [planEntry("divisor-guard")])

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient({ failPrompt: true }), project)

      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
      const data = JSON.parse(result.output)
      expect(data.runs[0].outcome).toBe("redispatched")
      expect(data.runs[0].new_status).toBe("failed")

      const finalized = JSON.parse(
        (await hooks.tool.consolidate_dispatch.execute(finalizeArgs(correlationId), noopContext())).output,
      )
      // fix-dispatch-verdict-derivation (issue #764): zero-success
      // finalization persists an INCONCLUSIVE artifact instead of being
      // rejected by a value mismatch — the verdict stays fail-closed.
      expect(finalized.status).toBe("success")
      expect(finalized.operation_verdict).toBe("INCONCLUSIVE")
      expect(finalized.errors).toEqual([])
    })
  })

  it("reports the early-exit state distinctly and restores the prior run file when the re-run never executes", async () => {
    // unknown_agent is the early-exit trigger: the reviewer manifest
    // drifted between the original dispatch and the recovery (the agent
    // was removed), so executeDispatchRun exits at manifest validation —
    // before child execution, runData null, no replacement run file. The
    // plan entry keeps its model provenance (matching the run fixture)
    // so the restored session stays finalizable; a tier-based trigger
    // (tier_model_unavailable) needs a model-less plan entry, which the
    // finalizer's model-provenance check would reject for unrelated
    // reasons.
    const guardRun = makeRun("divisor-sentinel", { status: "failed", sequence: 1 })
    const adversaryRun = makeRun("divisor-adversary", { status: "success", sequence: 2 })
    const { correlationId, dir } = await setupSession(
      [guardRun, adversaryRun],
      [planEntry("divisor-sentinel"), planEntry("divisor-adversary", { sequence: 2 })],
    )

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
      const data = JSON.parse(result.output)
      expect(data.runs[0].outcome).toBe("early_exit")
      expect(data.runs[0].run_id).toBe(guardRun.run_id)
      expect(data.runs[0].new_run_id).toBeNull()
      expect(data.runs[0].new_status).toBeNull()
      expect(data.runs[0].error.code).toBe("unknown_agent")
      expect(data.runs[0].superseded_run_file).toBeNull()
      expect(data.runs[0].reason).toContain("no replacement run file was written")
      expect(data.runs[0].reason).toContain("restored to its active path run-divisor-sentinel.json")

      // The prior run file is RESTORED at the active path — byte-identical
      // to what the original dispatch wrote — and superseded/ is empty:
      // the session is back to its pre-attempt state.
      const activeFiles = (await readdir(dir)).filter((n) => n.startsWith("run-") && n.endsWith(".json")).sort()
      expect(activeFiles).toEqual(["run-divisor-adversary.json", "run-divisor-sentinel.json"])
      expect(await readFile(join(dir, "run-divisor-sentinel.json"), "utf8")).toBe(JSON.stringify(guardRun, null, 2))
      expect(await readdir(join(dir, "superseded"))).toEqual([])

      // The restored file is visible to the active-file glob again, so
      // the session stays consolidatable: dispatch_status sees both runs
      // and the finalizer finds one run per plan entry (no "has no run"
      // rejection) — the restored failed run rides RD-FR-025 exclusion.
      const status = JSON.parse(
        (await hooks.tool.dispatch_status.execute({ correlation_id: correlationId }, noopContext())).output,
      )
      expect(status.status).toBe("ok")
      expect(status.total_runs).toBe(2)

      const finalized = JSON.parse(
        (await hooks.tool.consolidate_dispatch.execute(finalizeArgs(correlationId), noopContext())).output,
      )
      expect(finalized.status).toBe("success")
      expect(finalized.operation_verdict).toBe("APPROVE")
      expect(finalized.errors).toEqual([])
    })
  })

  it("restores the prior run file when executeDispatchRun throws after the archive rename", async () => {
    const guardRun = makeRun("divisor-guard", { status: "failed", sequence: 1 })
    const adversaryRun = makeRun("divisor-adversary", { status: "success", sequence: 2 })
    const { correlationId, dir } = await setupSession(
      [guardRun, adversaryRun],
      [planEntry("divisor-guard"), planEntry("divisor-adversary", { sequence: 2 })],
    )

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      // Force the shared core to throw after the archive rename: the
      // fake client creates the child session ("child-session") and the
      // child completes, but the post-execution submission harvest reads
      // this poisoned store entry whose findings getter throws — an
      // unexpected error after executeAgentSession returned and before
      // any run file is written.
      _getSubmissionStore().set("child-session", {
        get findings(): never[] {
          throw new Error("harvest exploded")
        },
        proposals: [],
      })

      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
      const data = JSON.parse(result.output)
      expect(data.redispatched_runs).toBe(0)
      expect(data.runs[0].outcome).toBe("error")
      expect(data.runs[0].run_id).toBe(guardRun.run_id)
      expect(data.runs[0].new_run_id).toBeNull()
      expect(data.runs[0].new_status).toBeNull()
      expect(data.runs[0].superseded_run_file).toBeNull()
      expect(data.runs[0].reason).toContain("executeDispatchRun threw")
      expect(data.runs[0].reason).toContain("harvest exploded")
      expect(data.runs[0].reason).toContain("restored to its active path run-divisor-guard.json")

      // Prior run file restored byte-identical at the active path;
      // superseded/ holds nothing.
      expect(await readFile(join(dir, "run-divisor-guard.json"), "utf8")).toBe(JSON.stringify(guardRun, null, 2))
      expect(await readdir(join(dir, "superseded"))).toEqual([])

      // The session remains consolidatable from the restored prior run
      // plus the untouched success.
      const finalized = JSON.parse(
        (await hooks.tool.consolidate_dispatch.execute(finalizeArgs(correlationId), noopContext())).output,
      )
      expect(finalized.status).toBe("success")
      expect(finalized.operation_verdict).toBe("APPROVE")
      expect(finalized.errors).toEqual([])
    })
  })

  it("reports an error outcome when the replacement run file cannot be persisted", async () => {
    // DR-FR-002 persistence gate (PR #776): a replacement whose run-file
    // write never reached disk must not report redispatched. The child
    // prompt flips the session directory read-only after the archive
    // step, so the replacement write fails AND the prior file cannot be
    // restored out of superseded/ — the operator-intervention branch.
    const guardRun = makeRun("divisor-guard", { status: "failed", findings: 1, sequence: 1 })
    const { correlationId, dir } = await setupSession([guardRun], [planEntry("divisor-guard")])

    await withBun(async () => {
      const project = await scratchProject()
      const base = fakeClient()
      const client = {
        session: {
          ...base.session,
          prompt: async (...args: never[]) => {
            await chmod(dir, 0o555)
            return base.session.prompt(...args)
          },
        },
      } as unknown as Parameters<typeof activate>[0]
      const hooks = await activate(client, project)

      try {
        const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
        const data = JSON.parse(result.output)
        expect(data.status).toBe("ok")
        expect(data.redispatched_runs).toBe(0)
        expect(data.runs).toHaveLength(1)
        expect(data.runs[0].outcome).toBe("error")
        expect(data.runs[0].run_id).toBe(guardRun.run_id)
        expect(data.runs[0].new_run_id).toBeNull()
        expect(data.runs[0].new_status).toBeNull()
        expect(data.runs[0].superseded_run_file).toBe(`superseded/run-divisor-guard-${guardRun.run_id}.json`)
        expect(data.runs[0].reason).toContain("could not be persisted")
        expect(data.runs[0].reason).toContain("operator intervention required")
      } finally {
        // Cleanup needs a writable directory (afterEach rm's it).
        await chmod(dir, 0o755)
      }

      // The agent has no active run file — the prior run survives only
      // in the superseded archive, and the result says so instead of
      // claiming a redispatch that never reached disk.
      expect((await readdir(dir)).filter((f) => f.startsWith("run-") && f.endsWith(".json"))).toEqual([])
      expect(await readdir(join(dir, "superseded"))).toEqual([`run-divisor-guard-${guardRun.run_id}.json`])
    })
  })

  it("reports an error outcome when the superseded archive cannot be created", async () => {
    const guardRun = makeRun("divisor-guard", { status: "failed", findings: 1, sequence: 1 })
    const { correlationId, dir } = await setupSession([guardRun], [planEntry("divisor-guard")])
    // Block the archive: "superseded" exists as a regular file, so the
    // mkdir inside the archive step fails and the run is skipped rather
    // than overwritten without an audit copy.
    await writeFile(join(dir, "superseded"), "blocking file", "utf8")

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
      const data = JSON.parse(result.output)
      expect(data.status).toBe("ok")
      expect(data.redispatched_runs).toBe(0)
      expect(data.runs).toHaveLength(1)
      expect(data.runs[0].outcome).toBe("error")
      expect(data.runs[0].run_id).toBe(guardRun.run_id)
      expect(data.runs[0].superseded_run_file).toBeNull()
      expect(data.runs[0].new_run_id).toBeNull()
      expect(data.runs[0].new_status).toBeNull()
      expect(data.runs[0].reason).toContain("failed to archive prior run file to superseded/")

      // The prior run file was never moved or replaced, and the blocking
      // path is still a regular file.
      const guard = JSON.parse(await readFile(join(dir, "run-divisor-guard.json"), "utf8"))
      expect(guard.run_id).toBe(guardRun.run_id)
      expect(guard.findings).toHaveLength(1)
      expect(await readFile(join(dir, "superseded"), "utf8")).toBe("blocking file")
    })
  })

  it("resolves the default tier when the plan entry carries neither model nor tier", async () => {
    const guardRun = makeRun("divisor-guard", { status: "failed", sequence: 1 })
    const { correlationId, dir } = await setupSession([guardRun], [
      planEntry("divisor-guard", { model: null, tier: null }),
    ])

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
      const data = JSON.parse(result.output)
      expect(data.redispatched_runs).toBe(1)
      expect(data.runs[0].outcome).toBe("redispatched")
      expect(data.runs[0].new_status).toBe("success")

      // Neither model nor tier was passed through, so the shared core
      // resolved the matrix default ("standard" profile) for the re-run.
      const replacement = JSON.parse(await readFile(join(dir, "run-divisor-guard.json"), "utf8"))
      expect(replacement.status).toBe("success")
      expect(replacement.requested_model).toBe("provider/standard")
      expect(replacement.prompt).toBe("Original prompt for divisor-guard")
    })
  })

  it("derives the replacement timeout from persisted work signals when timeout is omitted", async () => {
    const guardRun = makeRun("divisor-guard", { status: "failed", findings: 1, sequence: 1 })
    const { correlationId, dir } = await setupSession([guardRun], [
      planEntry("divisor-guard", { model: null, tier: "standard" }),
    ])

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
      const data = JSON.parse(result.output)
      expect(data.status).toBe("ok")
      expect(data.redispatched_runs).toBe(1)

      // RD-FR-029: the replacement derives its timeout from the persisted
      // work signals — the archived prompt bytes and the plan entry's tier —
      // against the repository's current matrix. Recovery is sequential, so
      // the contention multiplier stays at 1.0.
      const replacement = JSON.parse(await readFile(join(dir, "run-divisor-guard.json"), "utf8"))
      expect(replacement.timeout).toMatchObject({
        source: "derived",
        base_ms: 600_000,
        resolved_tier: "standard",
        session_profile_tier: "standard",
        profile_term_ms: 120_000,
        prompt_term_ms: 0,
        contention_multiplier: 1,
        max_parallel_runs: 1,
        ceiling_ms: 1_800_000,
        clamped: false,
      })
      expect(replacement.timeout.prompt_bytes).toBe(Buffer.byteLength("Original prompt for divisor-guard", "utf8"))
      expect(replacement.timeout.effective_ms).toBe(720_000)
    })
  })

  it("records explicit timeout provenance when timeout is provided", async () => {
    const guardRun = makeRun("divisor-guard", { status: "failed", findings: 1, sequence: 1 })
    const { correlationId, dir } = await setupSession([guardRun], [
      planEntry("divisor-guard", { model: null, tier: "standard" }),
    ])

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId, timeout: 900_000 }, noopContext())
      const data = JSON.parse(result.output)
      expect(data.status).toBe("ok")
      expect(data.redispatched_runs).toBe(1)

      // An explicit timeout argument overrides the derivation for every
      // re-executed reviewer and is recorded with source "explicit".
      const replacement = JSON.parse(await readFile(join(dir, "run-divisor-guard.json"), "utf8"))
      expect(replacement.timeout).toMatchObject({
        source: "explicit",
        effective_ms: 900_000,
      })
    })
  })

  it("re-grants the retrieval allowance to a redispatched manifest-mode run (RD-FR-036)", async () => {
    const guardRun = makeRun("divisor-guard", { status: "failed", findings: 1, sequence: 1, manifestMode: true })
    const { correlationId, dir } = await setupSession([guardRun], [
      planEntry("divisor-guard", { model: null, tier: "standard" }),
    ])

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
      const data = JSON.parse(result.output)
      expect(data.status).toBe("ok")
      expect(data.redispatched_runs).toBe(1)

      // RD-FR-036: the persisted delivery-mode signal re-grants the
      // allowance on the recovery path without re-running prompt assembly,
      // and carries into the replacement run file so repeated recoveries
      // keep granting it. Same inputs as the work-signal derivation test
      // above (720_000 inline); the manifest-mode replacement derives
      // exactly the allowance above it (MANIFEST_RETRIEVAL_ALLOWANCE_MS,
      // a literal per TC-001), unclamped under the 1_800_000 ceiling.
      const replacement = JSON.parse(await readFile(join(dir, "run-divisor-guard.json"), "utf8"))
      expect(replacement.manifest_mode).toBe(true)
      expect(replacement.timeout).toMatchObject({
        source: "derived",
        retrieval_allowance_ms: 300_000,
        effective_ms: 1_020_000,
        clamped: false,
      })
    })
  })

  it("treats run files that predate the delivery-mode field as inline (RD-FR-036)", async () => {
    // No manifestMode override: the run file carries no manifest_mode key,
    // the shape every run file written before RD-FR-036 has on disk.
    const guardRun = makeRun("divisor-guard", { status: "failed", findings: 1, sequence: 1 })
    const { correlationId, dir } = await setupSession([guardRun], [
      planEntry("divisor-guard", { model: null, tier: "standard" }),
    ])

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)

      // The pre-field file parses, the re-run executes, and the session
      // consolidates from the replacement — undefined reads as inline, so
      // no allowance is granted and the derivation matches the pre-change
      // work-signal derivation exactly (720_000, the sibling test above).
      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
      const data = JSON.parse(result.output)
      expect(data.status).toBe("ok")
      expect(data.redispatched_runs).toBe(1)

      const replacement = JSON.parse(await readFile(join(dir, "run-divisor-guard.json"), "utf8"))
      expect(replacement.manifest_mode).toBe(false)
      expect(replacement.timeout).toMatchObject({
        source: "derived",
        effective_ms: 720_000,
      })
      expect(replacement.timeout).not.toHaveProperty("retrieval_allowance_ms")
    })
  })

  it("warns and records matrix degradation instead of silently shrinking recovery windows (RD-FR-029)", async () => {
    const guardRun = makeRun("divisor-guard", { status: "failed", findings: 1, sequence: 1 })
    const { correlationId, dir } = await setupSession([guardRun], [
      planEntry("divisor-guard", { model: null, tier: "standard" }),
    ])

    await withBun(async () => {
      const project = await scratchProject()
      // Corrupt the matrix so resolveTimeoutLimits degrades to the schema
      // defaults on the recovery path. The degradation MUST be visible as
      // a warning mirroring the dispatch path (RD-FR-029: a 1200 s matrix
      // limit silently shrinking to the 600 s default is a defect), and
      // MUST be recorded in the settled operation record. The same
      // corruption fails the per-run policy load, so the re-run exits
      // before child execution and the prior run file is restored.
      await writeFile(join(project, ".uf", "review-matrix.yaml"), "limits: [unclosed", "utf8")
      const hooks = await activate(fakeClient(), project)

      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
      const data = JSON.parse(result.output)

      expect(data.status).toBe("ok")
      expect(
        (data.warnings ?? []).some((w: string) =>
          w.includes(
            "review matrix limits unavailable; per-run timeout derivation degraded to schema defaults",
          ),
        ),
      ).toBe(true)
      expect(data.runs[0].outcome).toBe("early_exit")
      expect(data.redispatched_runs).toBe(0)
      expect(data.restored_runs).toBe(1)

      // The prior failed run is back at its active path, consolidatable
      // via RD-FR-025 exclusion.
      const restored = JSON.parse(await readFile(join(dir, "run-divisor-guard.json"), "utf8"))
      expect(restored.run_id).toBe(guardRun.run_id)

      // RD-FR-031 + RD-FR-029: the settled operation record carries the
      // warning alongside the recovery counts, so the degradation is
      // auditable from the session directory alone. The in-loop restore
      // of the archived prior run is counted in run_counts.restored.
      const opFiles = (await readdir(dir)).filter((name) => /^operation-.*\.json$/.test(name))
      expect(opFiles).toHaveLength(1)
      const record = JSON.parse(
        await readFile(join(dir, opFiles[0]), "utf8"),
      ) as Record<string, unknown>
      const recordWarnings = record.warnings as string[] | undefined
      expect(
        (recordWarnings ?? []).some((w: string) => w.includes("review matrix limits unavailable")),
      ).toBe(true)
      expect(record.run_counts).toMatchObject({ redispatched: 0, restored: 1 })
    })
  })

  it("derives redispatch windows from the session's matrix limits, not schema defaults (RD-FR-029)", async () => {
    const guardRun = makeRun("divisor-guard", { status: "failed", sequence: 1 })
    const adversaryRun = makeRun("divisor-adversary", { status: "failed", sequence: 2 })
    const { correlationId, dir } = await setupSession(
      [guardRun, adversaryRun],
      [
        planEntry("divisor-guard", { model: null, tier: "standard", sequence: 1 }),
        planEntry("divisor-adversary", { model: null, tier: "heavy", sequence: 2 }),
      ],
    )

    await withBun(async () => {
      // Non-default matrix: the scalar limit is raised to 1200 s and the
      // heavy tier carries a 1500 s override. The recovery path assembles
      // its timeout signals from the session directory rather than the
      // dispatch input, so only this test can prove the redispatch wiring
      // reads these limits — under schema defaults both windows would
      // collapse to the 600 s base with no tier override.
      const matrix = JSON.parse(JSON.stringify(matrixFixture)) as {
        limits: Record<string, unknown>
      }
      matrix.limits.per_run_timeout_seconds = 1200
      matrix.limits.per_run_timeout_tiers = { lightweight: 300, heavy: 1500 }
      const project = await scratchProject(matrix)
      const hooks = await activate(fakeClient(), project)

      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
      const data = JSON.parse(result.output)
      expect(data.status).toBe("ok")
      expect(data.redispatched_runs).toBe(2)

      // Standard entry: 1200 s matrix base + 120 s profile term, prompt
      // below the free-byte allowance, sequential multiplier 1.0.
      const standard = JSON.parse(await readFile(join(dir, "run-divisor-guard.json"), "utf8"))
      expect(standard.timeout).toMatchObject({
        source: "derived",
        base_ms: 1_200_000,
        resolved_tier: "standard",
        effective_ms: 1_320_000,
        clamped: false,
      })

      // Heavy entry: the 1500 s tier override + 300 s heavy profile term
      // lands exactly on the 1,800,000 ms ceiling under the sequential
      // 1.0 multiplier — derived equals effective, so clamped is false
      // (contrast the dispatch path, whose 1.45x contention multiplier
      // pushes the same base over the ceiling and clamps).
      const heavy = JSON.parse(await readFile(join(dir, "run-divisor-adversary.json"), "utf8"))
      expect(heavy.timeout).toMatchObject({
        source: "derived",
        base_ms: 1_500_000,
        resolved_tier: "heavy",
        ceiling_ms: 1_800_000,
        effective_ms: 1_800_000,
        clamped: false,
      })
    })
  })

  it("records its operation window with recovery counts (RD-FR-031)", async () => {
    const guardRun = makeRun("divisor-guard", { status: "failed", sequence: 1 })
    const { correlationId, dir } = await setupSession([guardRun], [planEntry("divisor-guard")])

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)
      const result = await hooks.tool.redispatch_failed_runs.execute(
        { correlation_id: correlationId },
        noopContext(),
      )

      const data = JSON.parse(result.output)
      expect(data.redispatched_runs).toBe(1)
      expect(data.refused_runs).toBe(0)

      const opFiles = (await readdir(dir)).filter((name) => /^operation-.*\.json$/.test(name))
      expect(opFiles).toHaveLength(1)
      const record = JSON.parse(
        await readFile(join(dir, opFiles[0]), "utf8"),
      ) as Record<string, unknown>
      expect(record.kind).toBe("redispatch")
      expect(typeof record.operation_id).toBe("string")
      expect(typeof record.started_at).toBe("string")
      expect(typeof record.completed_at).toBe("string")
      expect(record.status).toBe("ok")
      expect(record.run_counts).toMatchObject({
        // Pinned as literals rather than sourced from the tool's own
        // output: the record's run_counts mapping is an independent
        // contract, and self-sourcing it here could not fail unless
        // the output key itself went missing.
        redispatched: 1,
        refused: 0,
        restored: 0,
      })
    })
  })

  it("re-runs a cancelled reviewer and archives the prior cancelled record (DR-FR-002)", async () => {
    const cancelledRun = makeRun("divisor-guard", { status: "cancelled", findings: 2, proposals: 1, sequence: 1 })
    const { correlationId, dir } = await setupSession([cancelledRun], [planEntry("divisor-guard")])

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)
      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
      const data = JSON.parse(result.output)

      expect(data.runs[0].outcome).toBe("redispatched")
      expect(data.runs[0].run_id).toBe(cancelledRun.run_id)
      expect(data.redispatched_runs).toBe(1)
      expect(data.restored_runs).toBe(0)

      const replacement = JSON.parse(await readFile(join(dir, "run-divisor-guard.json"), "utf8"))
      expect(replacement.run_id).not.toBe(cancelledRun.run_id)
      expect(replacement.status).toBe("success")

      const archived = JSON.parse(
        await readFile(join(dir, "superseded", `run-divisor-guard-${cancelledRun.run_id}.json`), "utf8"),
      )
      expect(archived.status).toBe("cancelled")
      expect(archived.findings).toHaveLength(2)
      expect(archived.proposals).toHaveLength(1)
    })
  })

  it("still refuses a named successful run (DR-FR-002)", async () => {
    const successRun = makeRun("divisor-guard", { sequence: 1 })
    const cancelledRun = makeRun("divisor-scribe", { status: "cancelled", sequence: 2 })
    const { correlationId, dir } = await setupSession(
      [successRun, cancelledRun],
      [planEntry("divisor-guard"), planEntry("divisor-scribe", { sequence: 2 })],
    )

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)
      const result = await hooks.tool.redispatch_failed_runs.execute(
        { correlation_id: correlationId, agents: ["divisor-guard"] },
        noopContext(),
      )
      const data = JSON.parse(result.output)

      expect(data.runs).toHaveLength(1)
      expect(data.runs[0].outcome).toBe("refused")
      expect(data.runs[0].run_id).toBe(successRun.run_id)
      expect(data.runs[0].reason).toContain("only failed and cancelled runs are redispatch targets")
      expect(data.refused_runs).toBe(1)
      expect(data.redispatched_runs).toBe(0)

      const untouched = JSON.parse(await readFile(join(dir, "run-divisor-guard.json"), "utf8"))
      expect(untouched.run_id).toBe(successRun.run_id)
      expect(untouched.status).toBe("success")
    })
  })

  it("restores a stranded archive before target resolution (DR-FR-002)", async () => {
    const strandedRun = makeRun("divisor-guard", { status: "failed", sequence: 1 })
    const { correlationId, dir } = await setupSession([], [planEntry("divisor-guard")])
    await mkdir(join(dir, "superseded"), { recursive: true })
    await writeFile(join(dir, "superseded", `run-divisor-guard-${strandedRun.run_id}.json`), JSON.stringify(strandedRun))

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)
      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
      const data = JSON.parse(result.output)

      expect(data.status).toBe("ok")
      expect(data.restored_agents).toContain("divisor-guard")
      expect(data.restored_runs).toBe(1)
      expect(data.runs[0].outcome).toBe("redispatched")
      expect(data.runs[0].run_id).toBe(strandedRun.run_id)

      const replacement = JSON.parse(await readFile(join(dir, "run-divisor-guard.json"), "utf8"))
      expect(replacement.run_id).not.toBe(strandedRun.run_id)
    })
  })

  it("leaves archives untouched when the agent has an active run file (DR-FR-002 D8)", async () => {
    const activeRun = makeRun("divisor-guard", { sequence: 1 })
    const strandedRun = makeRun("divisor-guard", { status: "failed", sequence: 1 })
    const { correlationId, dir } = await setupSession([activeRun], [planEntry("divisor-guard")])
    await mkdir(join(dir, "superseded"), { recursive: true })
    const archivePath = join(dir, "superseded", `run-divisor-guard-${strandedRun.run_id}.json`)
    await writeFile(archivePath, JSON.stringify(strandedRun))

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)
      // Naming the agent forces target resolution through the active
      // (successful) run — without the filter it simply is not a target
      // and no refusal would be recorded.
      const result = await hooks.tool.redispatch_failed_runs.execute(
        { correlation_id: correlationId, agents: ["divisor-guard"] },
        noopContext(),
      )
      const data = JSON.parse(result.output)

      // The active (successful) run is refused, never redispatched.
      expect(data.runs).toHaveLength(1)
      expect(data.runs[0].outcome).toBe("refused")
      expect(data.runs[0].run_id).toBe(activeRun.run_id)
      expect(data.runs[0].reason).toContain("only failed and cancelled runs are redispatch targets")
      expect(data.refused_runs).toBe(1)
      expect(data.redispatched_runs).toBe(0)

      // D8 invariant: the archive behind the active replacement is never
      // restored — it stays byte-identical in superseded/.
      expect(data.restored_agents).not.toContain("divisor-guard")
      expect(data.restored_runs).toBe(0)
      expect(await readFile(archivePath, "utf8")).toBe(JSON.stringify(strandedRun))

      const untouched = JSON.parse(await readFile(join(dir, "run-divisor-guard.json"), "utf8"))
      expect(untouched.run_id).toBe(activeRun.run_id)
      expect(untouched.status).toBe("success")
    })
  })

  it("restores the most recent archive when several are stranded (DR-FR-002)", async () => {
    const olderRun = makeRun("divisor-guard", { status: "failed", sequence: 1 })
    const newerRun = makeRun("divisor-guard", { status: "failed", sequence: 1 })
    const { correlationId, dir } = await setupSession([], [planEntry("divisor-guard")])
    await mkdir(join(dir, "superseded"), { recursive: true })
    const olderPath = join(dir, "superseded", `run-divisor-guard-${olderRun.run_id}.json`)
    const newerPath = join(dir, "superseded", `run-divisor-guard-${newerRun.run_id}.json`)
    await writeFile(olderPath, JSON.stringify(olderRun))
    await writeFile(newerPath, JSON.stringify(newerRun))
    await utimes(olderPath, 1000, 1000)
    await utimes(newerPath, 2000, 2000)

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)
      const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
      const data = JSON.parse(result.output)

      expect(data.restored_agents).toEqual(["divisor-guard"])
      expect(data.restored_runs).toBe(1)
      expect(data.runs[0].outcome).toBe("redispatched")
      expect(data.runs[0].run_id).toBe(newerRun.run_id)

      const archivedNames = await readdir(join(dir, "superseded"))
      expect(archivedNames).toContain(`run-divisor-guard-${olderRun.run_id}.json`)
      expect(archivedNames).toContain(`run-divisor-guard-${newerRun.run_id}.json`)
    })
  })

  it("warns when a stranded archive cannot be restored (DR-FR-002)", async () => {
    const strandedRun = makeRun("divisor-guard", { status: "failed", sequence: 1 })
    const { correlationId, dir } = await setupSession([], [planEntry("divisor-guard")])
    const supersededDir = join(dir, "superseded")
    await mkdir(supersededDir, { recursive: true })
    const archiveName = `run-divisor-guard-${strandedRun.run_id}.json`
    await writeFile(join(supersededDir, archiveName), JSON.stringify(strandedRun))
    // Make the restore rename fail while the scan still succeeds: 0o500
    // keeps read+traverse (readdir and stat work) but removes write, so
    // restorePriorRunFile's rename out of superseded/ fails EACCES — the
    // exact failure the pre-resolution loop used to drop silently.
    await chmod(supersededDir, 0o500)

    try {
      await withBun(async () => {
        const project = await scratchProject()
        const hooks = await activate(fakeClient(), project)
        const result = await hooks.tool.redispatch_failed_runs.execute({ correlation_id: correlationId }, noopContext())
        const data = JSON.parse(result.output) as {
          status: string
          warnings?: string[]
          restored_agents: string[]
        }

        // The session holds no active run files, so target resolution
        // reports not_found — but the failed restore MUST still surface:
        // the agent is left with no active run file and only the
        // superseded archive, which is operator-intervention territory.
        expect(data.status).toBe("not_found")
        expect(data.restored_agents).toEqual([])
        const warning = (data.warnings ?? []).find((w: string) =>
          w.includes("stranded-archive restore failed for divisor-guard"),
        )
        expect(warning).toBeDefined()
        expect(warning).toContain(`superseded/${archiveName}`)
        expect(warning).toContain("operator intervention required")

        // The archive itself is preserved untouched.
        expect(await readFile(join(supersededDir, archiveName), "utf8")).toBe(JSON.stringify(strandedRun))
      })
    } finally {
      // Restore writability so the afterEach rm -rf cleanup can unlink.
      await chmod(supersededDir, 0o700)
    }
  })

  it("excludes a re-run cancelled-again reviewer's submissions at consolidation (RD-FR-025)", async () => {
    const scribeRun = makeRun("divisor-scribe", { status: "cancelled", findings: 2, proposals: 1, sequence: 2 })
    const { correlationId } = await setupSession(
      [makeRun("divisor-guard"), scribeRun],
      [planEntry("divisor-guard"), planEntry("divisor-scribe", { sequence: 2 })],
    )

    await withBun(async () => {
      const project = await scratchProject()
      const hooks = await activate(fakeClient(), project)
      const payload = JSON.parse(
        (await hooks.tool.consolidate_dispatch.execute({ correlation_id: correlationId }, noopContext())).output,
      )

      expect(payload.run_counts.total).toBe(2)
      expect(payload.run_counts.success).toBe(1)
      expect(payload.findings).toHaveLength(0)
      expect(payload.warnings).toHaveLength(1)
      expect(payload.warnings[0]).toEqual({
        type: "orphaned_submissions_excluded",
        agent: "divisor-scribe",
        run_id: scribeRun.run_id,
        discarded_findings: 2,
        discarded_proposals: 1,
      })
    })
  })
})
