import { readFile, readdir, rm, writeFile } from "node:fs/promises"
import { join } from "node:path"

import type { PluginInput, ToolContext } from "@opencode-ai/plugin"
import { afterEach, describe, expect, it } from "vitest"

import ReviewDispatchPlugin from "../plugins/review-dispatch/index.js"
import { _getSessionCorrelationMap, _getSubmissionStore } from "../plugins/review-dispatch/index.js"
import { withBun } from "./helpers/bun-runtime.js"
import {
  UUID_RE,
  dispatchDir,
  fakeClient,
  makeBaseInput,
  makeDispatchDirTracker,
  makePlanEntry,
  makeScratchProjectFactory,
  makeSessionMetadata,
  pathExists,
} from "./helpers/aggregate-dispatch-fixtures.js"

// ── Test infrastructure ─────────────────────────────────────

const scratchDirectories: string[] = []
// Dispatch session dirs (tmpdir()/opencode/dispatch-<correlation>) minted by
// dispatch_review_runs invocations; cleaned so lifecycle tests never leak a
// session into a later test's map or filesystem assertions.
const dispatchSessionDirs: string[] = []

afterEach(async () => {
  _getSessionCorrelationMap().clear()
  _getSubmissionStore().clear()
  await Promise.all(scratchDirectories.splice(0).map((directory) => rm(directory, { recursive: true, force: true })))
  await Promise.all(dispatchSessionDirs.splice(0).map((directory) => rm(directory, { recursive: true, force: true })))
})

/**
 * Aggregate-dispatch fixtures shared with the sibling review-dispatch suites
 * (helpers/aggregate-dispatch-fixtures.ts, extracted per default.md CS-004):
 * a scratch project with review matrix, manifest, sibling-repos config, and
 * a diff file, bound to the scratch-directory registry; and a dispatch
 * session dir tracker bound to the dispatch-session registry. Both
 * registries drain in afterEach.
 */
const scratchProject = makeScratchProjectFactory("uf-dispatch-runs-", scratchDirectories)
const trackDispatchDir = makeDispatchDirTracker(dispatchSessionDirs)

/** Create a prompt response with cost data in step-finish parts. */
function promptResponseWithCost(costUsd: number, request: unknown): unknown {
  const req = request as { body?: { model?: { providerID: string; modelID: string } } }
  const providerID = req.body?.model?.providerID ?? "provider"
  const modelID = req.body?.model?.modelID ?? "standard"
  return {
    data: {
      info: { role: "assistant", providerID, modelID },
      parts: [
        { id: "t", sessionID: "child", messageID: "m", type: "text", text: "review output", ignored: false },
        { type: "step-finish", cost: costUsd, tokens: { input: 100, output: 50 } },
      ],
    },
    error: undefined,
  }
}

/** Tool context for a parent session; the sessionID keys the correlation-map lookup. */
function toolContext(sessionID = "parent-session"): ToolContext {
  return {
    sessionID,
    messageID: "message-parent",
    agent: "build",
    directory: "/workspace",
    worktree: "/workspace",
    abort: new AbortController().signal,
    metadata: () => undefined,
    ask: (() => {
      throw new Error("permission prompts are not used")
    }) as ToolContext["ask"],
  }
}

/** Valid SHA1 (40 hex chars) for test input_context. */
const VALID_SHA = "a".repeat(40)
const VALID_SHA2 = "b".repeat(40)

function makeTriageSessionMetadata(): Record<string, unknown> {
  return {
    command: "triage-issue",
    mode: "triage",
    full: false,
    input_context: {
      kind: "issue",
      issue_number: 99,
      issue_url: "https://github.com/unbound-force/unbound-force/issues/99",
      content_sha256: "a".repeat(64),
    },
    change_profile: {
      kind: "issue",
      text_bytes: 200,
      comment_count: 1,
      content_sha256: "a".repeat(64),
      matched_rules: [],
      security_sensitive: false,
      user_facing: true,
      categories: ["standard"],
      tier: "standard",
    },
    plan: [],
    coverage: { preflight_verdict: "NOT_RUN", checks_total: 0, checks_passed: 0 },
  }
}

function makeTriageInput(planEntries: Record<string, unknown>[]): Record<string, unknown> {
  return {
    mode: "triage",
    command: "triage-issue",
    issue_content: JSON.stringify({
      title: "Coverage regression",
      body: "Coverage dropped.",
      comments: [],
      labels: ["bug"],
      author_association: "MEMBER",
      created_at: "2026-10-01T09:00:00Z",
    }),
    input_context: '{"kind":"issue","issue_number":99,"issue_url":"https://github.com/unbound-force/unbound-force/issues/99","content_sha256":"' + "a".repeat(64) + '"}',
    plan_entries: planEntries,
    max_parallel_runs: 4,
    session_metadata: makeTriageSessionMetadata(),
  }
}

function makeFeedbackSessionMetadata(): Record<string, unknown> {
  return {
    command: "address-feedback",
    mode: "feedback",
    full: false,
    input_context: {
      kind: "local",
      pr_number: null,
      base_ref: "main",
      base_sha: VALID_SHA,
      head_ref: "opsx/feedback-golden",
      head_sha: VALID_SHA2,
    },
    change_profile: {
      kind: "diff",
      lines: 100,
      files: 3,
      components: 1,
      security_sensitive: false,
      user_facing: false,
      categories: ["standard"],
      tier: "standard",
    },
    plan: [],
    coverage: { preflight_verdict: "PASS", checks_total: 5, checks_passed: 5 },
  }
}

function makeFeedbackInput(diffPath: string, planEntries: Record<string, unknown>[]): Record<string, unknown> {
  return {
    mode: "feedback",
    command: "address-feedback",
    diff_path: diffPath,
    changed_files: "src/foo.ts",
    input_context: "pr 42: base main, head opsx/feedback-golden",
    plan_entries: planEntries,
    max_parallel_runs: 4,
    session_metadata: makeFeedbackSessionMetadata(),
  }
}

// ── Tests ───────────────────────────────────────────────────

describe("dispatch_review_runs", () => {
  it("registers in the plugin with correct description", async () => {
    await withBun(async () => {
      const { directory } = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: fakeClient(),
        directory,
        worktree: directory,
      } as unknown as PluginInput)

      expect(Object.keys(hooks.tool)).toContain("dispatch_review_runs")
      const toolDef = hooks.tool.dispatch_review_runs
      expect(toolDef.description).toContain("Aggregate")
    })
  })

  it("locks the dispatch_review_runs description forwarding guidance (RD-CF-002)", async () => {
    await withBun(async () => {
      const { directory } = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: fakeClient(),
        directory,
        worktree: directory,
      } as unknown as PluginInput)

      const toolDef = hooks.tool.dispatch_review_runs
      expect(toolDef.description).toContain(
        "pass it as correlation_id on every subsequent consolidate_dispatch call",
      )
    })
  })

  describe("dispatches all included plan entries and returns compact summary", () => {
    it("dispatches 2 included entries and skips 1, reports correct counts", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeBaseInput(diffPath, [
          makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
          makePlanEntry({ agent: "divisor-adversary", sequence: 2 }),
          makePlanEntry({ agent: "divisor-architect", decision: "skip", sequence: 3 }),
        ])

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        expect(response.output).toContain("dispatch_review_runs")
        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("ok")
        expect(meta.runs_completed).toBe(2)
        expect(meta.runs_skipped).toBe(1)
        expect(meta.runs_failed).toBe(0)
      })
    })

    it("returns status ok with 0 completed when all entries are skipped", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeBaseInput(diffPath, [
          makePlanEntry({ agent: "divisor-guard", decision: "skip", sequence: 1 }),
          makePlanEntry({ agent: "divisor-adversary", decision: "skip", sequence: 2 }),
        ])

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("ok")
        expect(meta.runs_completed).toBe(0)
        expect(meta.runs_skipped).toBe(2)
        const warnings = meta.warnings as string[]
        expect(warnings.some((w: string) => w.includes("no plan entries with decision 'include'"))).toBe(true)
      })
    })
  })

  describe("host-fallback panel dispatch (RD-FR-037)", () => {
    it("dispatches a host-fallback plan entry and persists host-run provenance", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        // Host-fallback shape exactly as baseRuns plans it for an agent in
        // neither runs[mode] nor advisor[mode]: null model/tier/variant with
        // the "use the active host model" reason.
        const input = makeBaseInput(diffPath, [
          makePlanEntry({
            agent: "divisor-guard",
            sequence: 1,
            source: "host",
            tier: null,
            model: null,
            variant: null,
            reason_code: "host-fallback",
            reason: "no explicit or advisor run; use the active host model",
          }),
        ])

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        expect(response.output).toContain("dispatch_review_runs")
        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("ok")
        expect(meta.runs_completed).toBe(1)
        expect(meta.runs_failed).toBe(0)
        expect(meta.runs_skipped).toBe(0)

        // RD-FR-037 "Host run inherits the parent model": the persisted run
        // file carries host-run provenance — null requested_model, provider,
        // model_id, and variant; non-null resolved_parent_model from the
        // fake client's host lookup (p/m with active variant).
        const correlationId = meta.correlation_id as string
        trackDispatchDir(correlationId)
        const runFiles = await readdir(dispatchDir(correlationId))
        expect(runFiles).toContain("run-divisor-guard.json")
        const run = JSON.parse(
          await readFile(join(dispatchDir(correlationId), "run-divisor-guard.json"), "utf8"),
        ) as Record<string, unknown>
        expect(run.source).toBe("host")
        expect(run.requested_model).toBeNull()
        expect(run.provider).toBeNull()
        expect(run.model_id).toBeNull()
        expect(run.variant).toBeNull()
        expect(run.resolved_parent_model).toBe("p/m")
        expect(run.resolved_parent_variant).toBe("active")
      })
    })
  })

  describe("respects max_parallel_runs batch size", () => {
    it("dispatches all 5 entries in batches of 2 and all complete", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const promptCalls: unknown[] = []
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient({ promptCalls }),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeBaseInput(diffPath, [
          makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
          makePlanEntry({ agent: "divisor-adversary", sequence: 2 }),
          makePlanEntry({ agent: "divisor-architect", sequence: 3 }),
          makePlanEntry({ agent: "divisor-sre", sequence: 4 }),
          makePlanEntry({ agent: "divisor-testing", sequence: 5 }),
        ])
        input.max_parallel_runs = 2

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("ok")
        // All 5 included entries should dispatch and complete.
        expect(meta.runs_completed).toBe(5)
        expect(meta.runs_failed).toBe(0)
        expect(meta.runs_skipped).toBe(0)
        // Verify the client was actually called 5 times (one prompt per entry).
        expect(promptCalls.length).toBe(5)
      })
    })
  })

  describe("isolates failures (one failed run does not cancel others)", () => {
    it("completes successful runs even when one fails", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        let callCount = 0
        const client = fakeClient({
          promptHandler: async (request) => {
            callCount++
            const req = request as { body?: { agent?: string; model?: { providerID: string; modelID: string } } }
            const providerID = req.body?.model?.providerID ?? "provider"
            const modelID = req.body?.model?.modelID ?? "standard"
            // Fail for the adversary agent to simulate one run failing.
            if (req.body?.agent === "divisor-adversary") {
              return {
                data: undefined,
                error: { message: "simulated child prompt failure" },
                response: { status: 500 },
              }
            }
            return {
              data: {
                info: { role: "assistant", providerID, modelID },
                parts: [
                  { id: "t", sessionID: "child", messageID: "m", type: "text", text: "review output", ignored: false },
                ],
              },
              error: undefined,
            }
          },
        })
        const hooks = await ReviewDispatchPlugin.server({
          client,
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeBaseInput(diffPath, [
          makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
          makePlanEntry({ agent: "divisor-adversary", sequence: 2 }),
          makePlanEntry({ agent: "divisor-architect", sequence: 3 }),
        ])

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        // Status should be "ok" because at least some runs completed.
        expect(meta.status).toBe("ok")
        expect(meta.runs_completed).toBe(2)
        expect(meta.runs_failed).toBe(1)
        // All 3 prompts were attempted (failure is isolated, not cancelling).
        expect(callCount).toBe(3)
      })
    })

    it("returns status error when ALL runs fail", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const client = fakeClient({
          promptHandler: async () => ({
            data: undefined,
            error: { message: "all fail" },
            response: { status: 500 },
          }),
        })
        const hooks = await ReviewDispatchPlugin.server({
          client,
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeBaseInput(diffPath, [
          makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
          makePlanEntry({ agent: "divisor-adversary", sequence: 2 }),
        ])

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("error")
        expect(meta.runs_completed).toBe(0)
        expect(meta.runs_failed).toBe(2)
      })
    })
  })

  describe("skips remaining runs when budget exceeded", () => {
    it("skips second batch after cumulative cost exceeds budget_usd", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const client = fakeClient({
          promptHandler: async (request) => promptResponseWithCost(0.50, request),
        })
        const hooks = await ReviewDispatchPlugin.server({
          client,
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        // 4 entries, max_parallel_runs = 2 → 2 batches of 2.
        // Each run costs $0.50 → batch 1 = $1.00 total.
        // Budget is $0.75, so after batch 1, budget is exceeded.
        // Batch 2 should be skipped.
        const input = makeBaseInput(diffPath, [
          makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
          makePlanEntry({ agent: "divisor-adversary", sequence: 2 }),
          makePlanEntry({ agent: "divisor-architect", sequence: 3 }),
          makePlanEntry({ agent: "divisor-sre", sequence: 4 }),
        ])
        input.max_parallel_runs = 2
        input.budget_usd = 0.75

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("ok")
        // First batch of 2 completes.
        expect(meta.runs_completed).toBe(2)
        // Second batch of 2 is skipped due to budget.
        expect(meta.runs_skipped).toBe(2)
        const warnings = meta.warnings as string[]
        expect(warnings.some((w: string) => w.includes("budget exceeded"))).toBe(true)
      })
    })
  })

  describe("persists session metadata under the invocation's dispatch session (FR-007)", () => {
    it("persists session metadata file from first run under the invocation's minted correlation", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const promptCalls: unknown[] = []
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient({ promptCalls }),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const sessionMeta = makeSessionMetadata()
        const input = makeBaseInput(diffPath, [
          makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
          makePlanEntry({ agent: "divisor-adversary", sequence: 2 }),
        ])
        input.session_metadata = sessionMeta

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("ok")
        expect(meta.runs_completed).toBe(2)

        // RD-CR-001 replaces the old shared-correlation assertion: the
        // invocation itself mints the correlation (before any internal
        // dispatchAgentRun call) and re-points the session map at it.
        const correlationId = meta.correlation_id as string
        expect(correlationId).toBeDefined()
        expect(correlationId).toMatch(UUID_RE)
        expect(_getSessionCorrelationMap().get("parent-session")).toBe(correlationId)
        trackDispatchDir(correlationId)

        // FR-007: the metadata from the first internal run call persisted
        // under the invocation's dispatch session directory.
        const metadataRaw = await readFile(join(dispatchDir(correlationId), "session-metadata.json"), "utf8")
        const persisted = JSON.parse(metadataRaw) as Record<string, unknown>
        expect(persisted.command).toBe("review-council")
        expect(persisted.mode).toBe("code")
      })
    })
  })

  // ── RD-CR-001: per-invocation correlation lifecycle (issue #794) ──

  // Issue #794 reproduction: a second aggregate dispatch from the same
  // parent session used to reuse — and clobber — the first dispatch's
  // session, because the correlation was minted once per conversation by
  // the first internal dispatchAgentRun call. dispatchReviewRuns now mints
  // a fresh correlation per invocation immediately after input validation
  // and re-points the session map (latest-dispatch semantics), so each
  // workflow from one conversation owns its own dispatch session.
  describe("mints a fresh dispatch session per invocation (RD-CR-001, issue #794)", () => {
    it("two invocations from one parent session mint distinct correlations with distinct session dirs", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const first = await hooks.tool.dispatch_review_runs.execute(
          makeBaseInput(diffPath, [makePlanEntry({ agent: "divisor-guard", sequence: 1 })]),
          toolContext(),
        )
        const firstMeta = first.metadata as Record<string, unknown>
        const c1 = firstMeta.correlation_id as string
        trackDispatchDir(c1)

        const second = await hooks.tool.dispatch_review_runs.execute(
          makeBaseInput(diffPath, [makePlanEntry({ agent: "divisor-adversary", sequence: 1 })]),
          toolContext(),
        )
        const secondMeta = second.metadata as Record<string, unknown>
        const c2 = secondMeta.correlation_id as string
        trackDispatchDir(c2)

        expect(firstMeta.status).toBe("ok")
        expect(secondMeta.status).toBe("ok")
        // Each invocation mints its own correlation. The shared-correlation
        // behavior this replaces was the #794 session-clobbering bug.
        expect(c1).toMatch(UUID_RE)
        expect(c2).toMatch(UUID_RE)
        expect(c1).not.toBe(c2)
        // Latest-dispatch semantics: the map resolves the latest invocation.
        expect(_getSessionCorrelationMap().get("parent-session")).toBe(c2)
        // Distinct session dirs, both on disk.
        const dir1 = dispatchDir(c1)
        const dir2 = dispatchDir(c2)
        expect(dir1).not.toBe(dir2)
        expect(await pathExists(dir1)).toBe(true)
        expect(await pathExists(dir2)).toBe(true)
        // Each output summary reports its own invocation's correlation.
        expect(first.output).toContain(`(correlation_id ${c1})`)
        expect(second.output).toContain(`(correlation_id ${c2})`)
      })
    })

    it("first session's run files and metadata survive the second invocation (no clobbering)", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        // Seed the child session the fake client creates for the first
        // invocation's single run (child-session-0) so its persisted run
        // file carries harvested findings that must survive the second
        // invocation.
        _getSubmissionStore().set("child-session-0", {
          findings: [
            {
              severity: "LOW",
              category: "documentation",
              description: "First dispatch session must survive a second invocation",
              root_cause: "Aggregate dispatch used to reuse the conversation's correlation (issue #794)",
              file: null,
              line: null,
            },
          ],
          proposals: [],
        })

        const firstInput = makeBaseInput(diffPath, [makePlanEntry({ agent: "divisor-guard", sequence: 1 })])
        firstInput.session_metadata = makeSessionMetadata()
        const first = await hooks.tool.dispatch_review_runs.execute(firstInput, toolContext())
        const firstMeta = first.metadata as Record<string, unknown>
        const c1 = firstMeta.correlation_id as string
        trackDispatchDir(c1)
        expect(firstMeta.runs_completed).toBe(1)
        // The first invocation collected its own session's findings.
        expect((firstMeta.findings as unknown[]).length).toBe(1)

        // Second, different workflow from the same parent session, with its
        // own distinguishable session metadata (head SHA c*40 vs b*40).
        const secondInput = makeBaseInput(diffPath, [makePlanEntry({ agent: "divisor-adversary", sequence: 1 })])
        const secondMetadata = makeSessionMetadata()
        const secondInputContext = secondMetadata.input_context as Record<string, unknown>
        secondInputContext.head_sha = "c".repeat(40)
        secondInput.session_metadata = secondMetadata
        const second = await hooks.tool.dispatch_review_runs.execute(secondInput, toolContext())
        const secondMeta = second.metadata as Record<string, unknown>
        const c2 = secondMeta.correlation_id as string
        trackDispatchDir(c2)
        expect(secondMeta.runs_completed).toBe(1)
        // No cross-session leakage: the second invocation collected only
        // its own (empty) session, not the first dispatch's findings.
        expect((secondMeta.findings as unknown[]).length).toBe(0)

        // The first session's run file survives under C1 with its findings.
        const runRaw = await readFile(join(dispatchDir(c1), "run-divisor-guard.json"), "utf8")
        const run = JSON.parse(runRaw) as Record<string, unknown>
        expect(run.correlation_id).toBe(c1)
        const runFindings = run.findings as Array<Record<string, unknown>>
        expect(runFindings).toHaveLength(1)
        expect(runFindings[0]?.description).toBe("First dispatch session must survive a second invocation")

        // The first session's metadata survives unchanged — still the first
        // invocation's head SHA, not the second's.
        const firstMetadataRaw = await readFile(join(dispatchDir(c1), "session-metadata.json"), "utf8")
        const persistedFirst = JSON.parse(firstMetadataRaw) as { input_context: { head_sha: string } }
        expect(persistedFirst.input_context.head_sha).toBe(VALID_SHA2)

        // The second session carries its own run file and metadata.
        const secondRunRaw = await readFile(join(dispatchDir(c2), "run-divisor-adversary.json"), "utf8")
        expect((JSON.parse(secondRunRaw) as Record<string, unknown>).correlation_id).toBe(c2)
        const secondMetadataRaw = await readFile(join(dispatchDir(c2), "session-metadata.json"), "utf8")
        const persistedSecond = JSON.parse(secondMetadataRaw) as { input_context: { head_sha: string } }
        expect(persistedSecond.input_context.head_sha).toBe("c".repeat(40))
      })
    })

    it("prompts land under the invocation's dispatch session prompts/ dir, not shared across invocations", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const first = await hooks.tool.dispatch_review_runs.execute(
          makeBaseInput(diffPath, [makePlanEntry({ agent: "divisor-guard", sequence: 1 })]),
          toolContext(),
        )
        const c1 = (first.metadata as Record<string, unknown>).correlation_id as string
        trackDispatchDir(c1)

        const second = await hooks.tool.dispatch_review_runs.execute(
          makeBaseInput(diffPath, [makePlanEntry({ agent: "divisor-adversary", sequence: 1 })]),
          toolContext(),
        )
        const c2 = (second.metadata as Record<string, unknown>).correlation_id as string
        trackDispatchDir(c2)

        // Each invocation's prompts land under its own dispatch session.
        const firstPrompts = await readdir(join(dispatchDir(c1), "prompts"))
        expect(firstPrompts).toContain("divisor-guard.md")
        const secondPrompts = await readdir(join(dispatchDir(c2), "prompts"))
        expect(secondPrompts).toContain("divisor-adversary.md")
        // No cross-contamination between the sessions' prompts dirs.
        expect(firstPrompts).not.toContain("divisor-adversary.md")
        expect(secondPrompts).not.toContain("divisor-guard.md")
        // The prompt is a real assembled prompt for its agent.
        const promptContent = await readFile(join(dispatchDir(c1), "prompts", "divisor-guard.md"), "utf8")
        expect(promptContent).toContain("## Persona: divisor-guard")
        expect(promptContent).toContain("**Command**: review-council")
      })
    })
  })

  // ── RD-CR-001: session-less invocations re-point resolution ──

  // All-skip and prompt-building-failure returns happen AFTER the
  // correlation mint, so the map is left pointing at a correlation under
  // which no dispatch session directory exists. Later no-provenance
  // retrieval then fails closed as not_found instead of returning the
  // earlier dispatch's payload (RD-CR-001 "Session-less invocation
  // re-points resolution").
  describe("session-less invocations re-point the session map (RD-CR-001)", () => {
    it("all-skip invocation re-points the map to a session-less correlation and reports it", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        // A completed dispatch under C1 that the map previously resolved.
        const first = await hooks.tool.dispatch_review_runs.execute(
          makeBaseInput(diffPath, [makePlanEntry({ agent: "divisor-guard", sequence: 1 })]),
          toolContext(),
        )
        const c1 = (first.metadata as Record<string, unknown>).correlation_id as string
        trackDispatchDir(c1)
        expect(_getSessionCorrelationMap().get("parent-session")).toBe(c1)

        // Every plan entry skipped: no prompts, no runs, no session dir —
        // but the correlation was minted before the included-entries check.
        const second = await hooks.tool.dispatch_review_runs.execute(
          makeBaseInput(diffPath, [
            makePlanEntry({ agent: "divisor-guard", decision: "skip", sequence: 1 }),
            makePlanEntry({ agent: "divisor-adversary", decision: "skip", sequence: 2 }),
          ]),
          toolContext(),
        )
        const secondMeta = second.metadata as Record<string, unknown>
        const c2 = secondMeta.correlation_id as string

        expect(secondMeta.status).toBe("ok")
        expect(secondMeta.runs_completed).toBe(0)
        expect(secondMeta.runs_skipped).toBe(2)
        // The session-less outcome still reports its correlation (RD-CR-002).
        expect(c2).toMatch(UUID_RE)
        expect(c2).not.toBe(c1)
        expect(second.output).toContain(`(correlation_id ${c2})`)
        // The map was re-pointed at the session-less correlation.
        expect(_getSessionCorrelationMap().get("parent-session")).toBe(c2)
        // No dispatch session directory exists under it.
        expect(await pathExists(dispatchDir(c2))).toBe(false)
      })
    })

    it("prompt-building failure re-points the map to a session-less correlation and reports it", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const first = await hooks.tool.dispatch_review_runs.execute(
          makeBaseInput(diffPath, [makePlanEntry({ agent: "divisor-guard", sequence: 1 })]),
          toolContext(),
        )
        const c1 = (first.metadata as Record<string, unknown>).correlation_id as string
        trackDispatchDir(c1)

        // Unreadable diff path: prompt building fails after the mint and
        // before any prompt file or session dir is created.
        const second = await hooks.tool.dispatch_review_runs.execute(
          makeBaseInput(join(directory, "missing.diff"), [makePlanEntry({ agent: "divisor-guard", sequence: 1 })]),
          toolContext(),
        )
        const secondMeta = second.metadata as Record<string, unknown>
        const c2 = secondMeta.correlation_id as string

        expect(secondMeta.status).toBe("error")
        expect(secondMeta.message as string).toContain("prompt building failed")
        expect(secondMeta.runs_completed).toBe(0)
        // The prompt-failure outcome still reports its correlation (RD-CR-002).
        expect(c2).toMatch(UUID_RE)
        expect(c2).not.toBe(c1)
        expect(second.output).toContain(`(correlation_id ${c2})`)
        expect(_getSessionCorrelationMap().get("parent-session")).toBe(c2)
        expect(await pathExists(dispatchDir(c2))).toBe(false)
      })
    })

    it("no-provenance retrieval fails closed as not_found after a session-less re-point", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        // A completed dispatch under C1 with a retrievable session on disk.
        const first = await hooks.tool.dispatch_review_runs.execute(
          makeBaseInput(diffPath, [makePlanEntry({ agent: "divisor-guard", sequence: 1 })]),
          toolContext(),
        )
        const c1 = (first.metadata as Record<string, unknown>).correlation_id as string
        trackDispatchDir(c1)

        // The all-skip invocation re-points the map at a session-less C2.
        const second = await hooks.tool.dispatch_review_runs.execute(
          makeBaseInput(diffPath, [makePlanEntry({ agent: "divisor-guard", decision: "skip", sequence: 1 })]),
          toolContext(),
        )
        const c2 = (second.metadata as Record<string, unknown>).correlation_id as string

        // No-provenance consolidation resolves C2 and fails closed as
        // not_found rather than returning the earlier dispatch's payload.
        const notFound = await hooks.tool.consolidate_dispatch.execute({}, toolContext())
        const notFoundPayload = JSON.parse(notFound.output) as Record<string, unknown>
        expect(notFoundPayload.status).toBe("not_found")
        expect(notFoundPayload.correlation_id).toBe(c2)

        // The earlier session stays addressable by its explicit correlation.
        const explicit = await hooks.tool.consolidate_dispatch.execute({ correlation_id: c1 }, toolContext())
        const explicitPayload = JSON.parse(explicit.output) as Record<string, unknown>
        expect(explicitPayload.status).not.toBe("not_found")
        expect(explicitPayload.correlation_id).toBe(c1)

        // Pin the positive payload, not just the absence of not_found: the
        // c1 session had one included guard run that completed successfully.
        const counts = explicitPayload.run_counts as Record<string, number>
        expect(counts.total).toBe(1)
        expect(counts.success).toBe(1)
        const runs = explicitPayload.runs as Array<{ agent: string; run_id: string }>
        expect(runs).toHaveLength(1)
        expect(runs[0].agent).toBe("divisor-guard")
        expect(typeof runs[0].run_id).toBe("string")
        expect(explicitPayload.warnings).toBeUndefined()
      })
    })
  })

  // ── RD-CR-002: aggregate dispatch correlation observability ──

  describe("reports the invocation's correlation_id (RD-CR-002)", () => {
    it("success output summary and metadata carry the invocation's correlation id", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const response = await hooks.tool.dispatch_review_runs.execute(
          makeBaseInput(diffPath, [makePlanEntry({ agent: "divisor-guard", sequence: 1 })]),
          toolContext(),
        )

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("ok")
        const correlationId = meta.correlation_id as string
        expect(correlationId).toMatch(UUID_RE)
        // The machine-readable field and the prose summary agree, so the
        // dispatch session is addressable from the output alone.
        expect(response.output).toContain("dispatch_review_runs: ok (correlation_id ")
        expect(response.output).toContain(`(correlation_id ${correlationId})`)
        // The reported correlation is the session the map resolves.
        expect(_getSessionCorrelationMap().get("parent-session")).toBe(correlationId)
        trackDispatchDir(correlationId)
      })
    })

    it("pre-parse validation failure renders correlation_id none and mints no map entry", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        // plan_entries.min(1) fails zod validation before any correlation
        // is minted — the one return path that carries no correlation.
        const response = await hooks.tool.dispatch_review_runs.execute(
          makeBaseInput(diffPath, []),
          toolContext(),
        )

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("error")
        expect(meta.correlation_id).toBeUndefined()
        expect(response.output).toContain("(correlation_id none)")
        // Nothing was minted: the session map has no entry for the parent.
        expect(_getSessionCorrelationMap().has("parent-session")).toBe(false)
      })
    })
  })

  // ── RD-CF-001: cross-session consolidation by forwarded ID ──

  // Issue #872: the orchestration commands chain dispatch_review_runs
  // into consolidate_dispatch but never forwarded the dispatch's
  // reported correlation_id, relying on the in-process
  // sessionCorrelationMap — which only bridges calls sharing one plugin
  // process AND one context.sessionID. These tests lock the
  // aggregate-minted cross-session path (design D7): the dispatch
  // session is minted by a real dispatch_review_runs call under fixture
  // session S1, then consolidated from a DIFFERENT ToolContext (S2)
  // holding no map entry. Distinct from dispatch-recovery.test.ts,
  // which hand-writes run files to exercise the recovery tools.
  describe("cross-session consolidation by forwarded correlation_id (RD-CF-001)", () => {
    it("forwarded ID consolidates the S1 dispatch from S2 with runs, findings, and the ID echoed", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        // Seed the child session the fake client creates for this
        // single included run (child-session-0), so the persisted run
        // file carries a finding the cross-session consolidation must
        // retrieve.
        _getSubmissionStore().set("child-session-0", {
          findings: [
            {
              severity: "MEDIUM",
              category: "documentation",
              description: "Forwarded correlation IDs must consolidate a dispatch across tool sessions",
              root_cause: "Orchestration callers relied on the in-process session map alone (issue #872)",
              file: null,
              line: null,
            },
          ],
          proposals: [],
        })

        // S1: the dispatching parent session.
        const dispatch = await hooks.tool.dispatch_review_runs.execute(
          makeBaseInput(diffPath, [makePlanEntry({ agent: "divisor-guard", sequence: 1 })]),
          toolContext(),
        )
        const dispatchMeta = dispatch.metadata as Record<string, unknown>
        expect(dispatchMeta.status).toBe("ok")
        expect(dispatchMeta.runs_completed).toBe(1)
        const correlationId = dispatchMeta.correlation_id as string
        expect(correlationId).toMatch(UUID_RE)
        trackDispatchDir(correlationId)

        // S2: a different tool session holding no map entry — the
        // plugin-process or tool-session split that strands a no-ID
        // consolidation. The forwarded ID must bridge it.
        const s2 = toolContext("other-tool-session")
        expect(_getSessionCorrelationMap().has("other-tool-session")).toBe(false)

        const consolidated = await hooks.tool.consolidate_dispatch.execute({ correlation_id: correlationId }, s2)
        const payload = JSON.parse(consolidated.output) as Record<string, unknown>

        // RD-CF-001 "Forwarded ID consolidates from a different
        // session": C resolves, is echoed in the payload, and returns
        // the dispatch's run records and deduplicated findings.
        expect(payload.status).not.toBe("not_found")
        expect(payload.correlation_id).toBe(correlationId)
        const runs = payload.runs as Array<{ agent: string; run_id: string; status: string }>
        expect(runs).toHaveLength(1)
        expect(runs[0].agent).toBe("divisor-guard")
        expect(runs[0].status).toBe("success")
        expect(typeof runs[0].run_id).toBe("string")
        const counts = payload.run_counts as Record<string, number>
        expect(counts.total).toBe(1)
        expect(counts.success).toBe(1)
        const findings = payload.findings as Array<Record<string, unknown>>
        expect(findings).toHaveLength(1)
        expect(findings[0]?.description).toBe(
          "Forwarded correlation IDs must consolidate a dispatch across tool sessions",
        )
      })
    })

    it("no-ID consolidation from S2 fails closed as not_found without reading S1's state", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const dispatch = await hooks.tool.dispatch_review_runs.execute(
          makeBaseInput(diffPath, [makePlanEntry({ agent: "divisor-guard", sequence: 1 })]),
          toolContext(),
        )
        const dispatchMeta = dispatch.metadata as Record<string, unknown>
        expect(dispatchMeta.status).toBe("ok")
        expect(dispatchMeta.runs_completed).toBe(1)
        const correlationId = dispatchMeta.correlation_id as string
        trackDispatchDir(correlationId)

        // S2 holds no map entry, so no correlation resolves.
        const s2 = toolContext("other-tool-session")
        expect(_getSessionCorrelationMap().has("other-tool-session")).toBe(false)

        const consolidated = await hooks.tool.consolidate_dispatch.execute({}, s2)
        const payload = JSON.parse(consolidated.output) as Record<string, unknown>

        // RD-CF-001 "No-ID consolidation from a different session fails
        // closed": the map-miss envelope, with no dispatch session read.
        expect(payload.status).toBe("not_found")
        expect(payload.message).toBe("no dispatch session for this session")
        expect(payload.runs).toEqual([])
        expect(payload.correlation_id).toBeUndefined()
        // No cross-session leakage: S1's correlation never appears.
        expect(consolidated.output).not.toContain(correlationId)
      })
    })

    it("same-session no-ID consolidation from S1 still resolves via the map fallback", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        _getSubmissionStore().set("child-session-0", {
          findings: [
            {
              severity: "MEDIUM",
              category: "documentation",
              description: "Same-session map fallback must keep consolidating ID-less calls",
              root_cause: "RD-CF-001 keeps the map as a same-session convenience fallback",
              file: null,
              line: null,
            },
          ],
          proposals: [],
        })

        const dispatch = await hooks.tool.dispatch_review_runs.execute(
          makeBaseInput(diffPath, [makePlanEntry({ agent: "divisor-guard", sequence: 1 })]),
          toolContext(),
        )
        const dispatchMeta = dispatch.metadata as Record<string, unknown>
        expect(dispatchMeta.status).toBe("ok")
        expect(dispatchMeta.runs_completed).toBe(1)
        const correlationId = dispatchMeta.correlation_id as string
        trackDispatchDir(correlationId)
        // The map resolves S1 to the minted correlation.
        expect(_getSessionCorrelationMap().get("parent-session")).toBe(correlationId)

        const consolidated = await hooks.tool.consolidate_dispatch.execute({}, toolContext())
        const payload = JSON.parse(consolidated.output) as Record<string, unknown>

        // RD-CF-001 "Same-session no-ID consolidation still resolves":
        // the map fallback succeeds exactly as before the requirement,
        // returning the same substance the forwarded path returns.
        expect(payload.status).not.toBe("not_found")
        expect(payload.correlation_id).toBe(correlationId)
        const runs = payload.runs as Array<{ agent: string }>
        expect(runs).toHaveLength(1)
        expect(runs[0].agent).toBe("divisor-guard")
        const counts = payload.run_counts as Record<string, number>
        expect(counts.total).toBe(1)
        expect(counts.success).toBe(1)
        const findings = payload.findings as Array<Record<string, unknown>>
        expect(findings).toHaveLength(1)
        expect(findings[0]?.description).toBe("Same-session map fallback must keep consolidating ID-less calls")
      })
    })

    it("forwarded but unknown well-formed ID fails closed as not_found with the ID echoed", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const dispatch = await hooks.tool.dispatch_review_runs.execute(
          makeBaseInput(diffPath, [makePlanEntry({ agent: "divisor-guard", sequence: 1 })]),
          toolContext(),
        )
        const dispatchMeta = dispatch.metadata as Record<string, unknown>
        expect(dispatchMeta.status).toBe("ok")
        const correlationId = dispatchMeta.correlation_id as string
        trackDispatchDir(correlationId)

        // A well-formed UUID that resolves to no persisted dispatch
        // session; the premise is pinned — it validates against the
        // tool's UUIDSchema and no session directory exists under it.
        const unknownId = "00000000-1111-4222-8333-444444444444"
        expect(unknownId).toMatch(UUID_RE)
        expect(await pathExists(dispatchDir(unknownId))).toBe(false)

        const s2 = toolContext("other-tool-session")
        const consolidated = await hooks.tool.consolidate_dispatch.execute({ correlation_id: unknownId }, s2)
        const payload = JSON.parse(consolidated.output) as Record<string, unknown>

        // RD-CF-001 "Forwarded but unknown ID fails closed": not_found
        // with the supplied ID echoed and no message key — the envelope
        // that makes a forwarded not_found permanent (design D5).
        expect(payload.status).toBe("not_found")
        expect(payload.correlation_id).toBe(unknownId)
        expect(payload.runs).toEqual([])
        expect(payload.message).toBeUndefined()
      })
    })
  })

  describe("handles evidence acquisition failure gracefully (FR-010)", () => {
    it("proceeds with runs and adds warning when sibling repos YAML is invalid", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        // Overwrite sibling-repos.yaml with invalid content.
        await writeFile(join(directory, ".uf", "sibling-repos.yaml"), "not valid json at all {{{")
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeBaseInput(diffPath, [
          makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
        ])

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        // Runs should still complete despite evidence acquisition failure.
        expect(meta.runs_completed).toBe(1)
        expect(meta.runs_failed).toBe(0)
        // A warning about evidence acquisition should be present.
        const warnings = meta.warnings as string[]
        expect(
          warnings.some((w: string) => w.includes("sibling evidence") || w.includes("evidence acquisition")),
        ).toBe(true)
      })
    })

    it("proceeds with runs when sibling-repos.yaml is missing entirely", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        // Remove sibling-repos.yaml to simulate missing file.
        await rm(join(directory, ".uf", "sibling-repos.yaml"), { force: true })
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeBaseInput(diffPath, [
          makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
        ])

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        // Runs proceed despite missing evidence source.
        expect(meta.status).toBe("ok")
        expect(meta.runs_completed).toBe(1)
      })
    })
  })

  describe("returns error for invalid inputs", () => {
    it("returns error when plan_entries is empty", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeBaseInput(diffPath, [])
        // plan_entries.min(1) will fail
        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("error")
        expect(meta.runs_completed).toBe(0)
        expect(meta.runs_failed).toBe(0)
        expect(meta.runs_skipped).toBe(0)
        expect(typeof meta.message).toBe("string")
        expect((meta.message as string).toLowerCase()).toContain("validation")
        expect(meta.retryable).toBe(false)
      })
    })

    it("returns error when diff_path is empty", async () => {
      await withBun(async () => {
        const { directory } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeBaseInput("", [
          makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
        ])

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("error")
        expect(meta.runs_completed).toBe(0)
        expect(typeof meta.message).toBe("string")
        expect((meta.message as string).toLowerCase()).toContain("validation")
      })
    })

    it("returns error when required fields are missing", async () => {
      await withBun(async () => {
        const { directory } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        // Missing mode, command, session_metadata, etc.
        const input = { plan_entries: [makePlanEntry()] }

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("error")
        expect(meta.runs_completed).toBe(0)
        expect(meta.retryable).toBe(false)
      })
    })

    it("returns error when plan_entries contains invalid entry shape", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeBaseInput(diffPath, [
          { agent: "not-a-divisor-agent", decision: "include" }, // Invalid: missing required fields
        ])

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("error")
        expect(meta.runs_completed).toBe(0)
      })
    })
  })

  // ── Triage mode tests ──────────────────────────────────────

  describe("triage mode dispatches without diff", () => {
    it("dispatches triage entries successfully without diff_path or changed_files", async () => {
      await withBun(async () => {
        const { directory } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeTriageInput([
          makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
        ])

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("ok")
        expect(meta.runs_completed).toBe(1)
        expect(meta.runs_failed).toBe(0)
      })
    })

    it("rejects triage mode when diff_path is provided", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = {
          ...makeTriageInput([makePlanEntry({ agent: "divisor-guard", sequence: 1 })]),
          diff_path: diffPath,
        }

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("error")
        expect(typeof meta.message).toBe("string")
        expect((meta.message as string).toLowerCase()).toContain("diff_path")
      })
    })

    it("rejects triage mode when issue_content is missing", async () => {
      await withBun(async () => {
        const { directory } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = {
          mode: "triage",
          command: "triage-issue",
          input_context: "issue context",
          plan_entries: [makePlanEntry({ agent: "divisor-guard", sequence: 1 })],
          max_parallel_runs: 4,
          session_metadata: makeTriageSessionMetadata(),
        }

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("error")
        expect(typeof meta.message).toBe("string")
        expect((meta.message as string).toLowerCase()).toContain("issue_content")
      })
    })
  })

  // ── #735 golden contract ───────────────────────────────────

  // #735: uf.triage-issue.md documents that dispatch_review_runs returns
  // `status`, `runs_completed`, `runs_failed`, and `runs_skipped` counters,
  // consolidated `findings`, `proposals` for lesson processing, and any
  // `warnings` — but at filing time the tool did not return the findings,
  // proposals, or warnings arrays. The tool now carries all seven fields as
  // structured metadata beside the `output` summary string. This golden test
  // pins that reconciliation: it executes the tool through the plugin tool
  // registry using the documented triage inputs (uf.triage-issue.md,
  // "Dispatch Review Runs") and asserts the full documented envelope, so
  // future divergence on either side — the prose documenting a field the
  // tool lacks, or the tool dropping a documented field — fails here. The
  // same contract sentence appears in uf.review-council.md, uf.review-pr.md,
  // and uf.address-feedback.md. The pinned reconciliation is recorded in
  // #735 when this change lands.
  describe("#735 golden contract: documented output envelope", () => {
    it("returns status, run counters, findings, proposals, and warnings for documented triage inputs", async () => {
      await withBun(async () => {
        const { directory } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        // Seed the in-process submission store for the child session the
        // fake client creates for this single included run (its first
        // session.create returns "child-session-0"), so harvested findings
        // and proposals flow through the persisted run file into the
        // returned envelope — proving the documented fields carry real data.
        _getSubmissionStore().set("child-session-0", {
          findings: [
            {
              severity: "LOW",
              category: "documentation",
              description: "Documented dispatch_review_runs output contract is not pinned by a test",
              root_cause: "uf.triage-issue.md documented fields the tool did not return (#735)",
              file: null,
              line: null,
            },
          ],
          proposals: [{ information: "Pin documented tool output contracts with golden tests", tag: "issue-triage" }],
        })

        // Documented inputs per uf.triage-issue.md: mode triage, command
        // triage-issue, issue_content, input_context, plan_entries,
        // session_metadata, max_parallel_runs, budget_usd,
        // per_run_timeout_ms, and review_context. Triage mode rejects
        // diff_path and changed_files, so neither is passed.
        const input = makeTriageInput([makePlanEntry({ agent: "divisor-guard", sequence: 1 })])
        input.budget_usd = 5
        input.per_run_timeout_ms = 30_000
        input.review_context = "Repository context: unbound-force meta repo. Duplicate candidates: none."

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        // The summary string carries the status beside the structured envelope.
        expect(response.output).toContain("dispatch_review_runs: ok")
        const meta = response.metadata as Record<string, unknown>
        // The seven documented envelope fields (uf.triage-issue.md):
        expect(meta.status).toBe("ok")
        expect(meta.runs_completed).toBe(1)
        expect(meta.runs_failed).toBe(0)
        expect(meta.runs_skipped).toBe(0)
        expect(Array.isArray(meta.findings)).toBe(true)
        expect(Array.isArray(meta.proposals)).toBe(true)
        expect(Array.isArray(meta.warnings)).toBe(true)
        // Findings and proposals flow through from the child submission.
        const findings = meta.findings as Array<Record<string, unknown>>
        expect(findings).toHaveLength(1)
        expect(findings[0]?.category).toBe("documentation")
        const proposals = meta.proposals as Array<Record<string, unknown>>
        expect(proposals).toHaveLength(1)
        expect(proposals[0]?.tag).toBe("issue-triage")
      })
    })
  })

  // ── Remaining documented input variants (task 2.3 completion) ──

  // The #735 golden above pins the triage input variant
  // (uf.triage-issue.md). The same seven-field contract sentence is
  // documented at three more sites with two more input variants: code
  // mode (uf.review-council.md:260 and uf.review-pr.md:421 — review-pr
  // also routes command review-council per its documented inputs) and
  // feedback mode (uf.address-feedback.md:363). These goldens execute
  // each remaining variant through the plugin tool registry with its
  // documented inputs and assert the same envelope, so all four
  // dispatch_review_runs sites are pinned under their documented input
  // variants (metrics.md drift baseline 47/47/0).
  describe("#735 golden contract: remaining documented input variants", () => {
    it("returns the seven documented envelope fields for documented code-mode inputs (uf.review-council.md:260, uf.review-pr.md:421)", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        // Seed the in-process submission store for the child session the
        // fake client creates for this single included run, so harvested
        // findings and proposals flow through the persisted run file
        // into the returned envelope.
        _getSubmissionStore().set("child-session-0", {
          findings: [
            {
              severity: "MEDIUM",
              category: "test-quality",
              description: "Code-mode golden coverage was missing for dispatch_review_runs",
              root_cause: "Only the triage input variant executed (metrics.md drift-baseline gap)",
              file: "src/foo.ts",
              line: 1,
            },
          ],
          proposals: [{ information: "Pin every documented input variant with golden tests", tag: "review-dispatch" }],
        })

        // Documented inputs per uf.review-council.md:260 and
        // uf.review-pr.md:421: mode code, command review-council,
        // diff_path, changed_files, input_context, plan_entries,
        // session_metadata, max_parallel_runs, budget_usd,
        // per_run_timeout_ms, and — when available — pre_flight_results,
        // walkthrough, and review_context.
        const input = makeBaseInput(diffPath, [makePlanEntry({ agent: "divisor-guard", sequence: 1 })])
        input.budget_usd = 5
        input.per_run_timeout_ms = 30_000
        input.pre_flight_results = "Pre-flight: 18 of 18 CI checks pass."
        input.walkthrough = "Adds a service layer for report generation."
        input.review_context = "Repository context: unbound-force meta repo."

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        expect(response.output).toContain("dispatch_review_runs: ok")
        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("ok")
        expect(meta.runs_completed).toBe(1)
        expect(meta.runs_failed).toBe(0)
        expect(meta.runs_skipped).toBe(0)
        expect(Array.isArray(meta.findings)).toBe(true)
        expect(Array.isArray(meta.proposals)).toBe(true)
        expect(Array.isArray(meta.warnings)).toBe(true)
        const findings = meta.findings as Array<Record<string, unknown>>
        expect(findings).toHaveLength(1)
        expect(findings[0]?.category).toBe("test-quality")
        const proposals = meta.proposals as Array<Record<string, unknown>>
        expect(proposals).toHaveLength(1)
        expect(proposals[0]?.tag).toBe("review-dispatch")
      })
    })

    it("returns the seven documented envelope fields for documented feedback-mode inputs (uf.address-feedback.md:363)", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        _getSubmissionStore().set("child-session-0", {
          findings: [
            {
              severity: "MEDIUM",
              category: "documentation",
              description: "Feedback-mode golden coverage was missing for dispatch_review_runs",
              root_cause: "Only the triage input variant executed (metrics.md drift-baseline gap)",
              file: null,
              line: null,
            },
          ],
          proposals: [{ information: "Pin every documented input variant with golden tests", tag: "feedback-triage" }],
        })

        // Documented inputs per uf.address-feedback.md:363: mode feedback,
        // command address-feedback, diff_path, changed_files, input_context,
        // plan_entries, session_metadata, max_parallel_runs, budget_usd,
        // per_run_timeout_ms, and review_context.
        const input = makeFeedbackInput(diffPath, [makePlanEntry({ agent: "divisor-guard", sequence: 1 })])
        input.budget_usd = 5
        input.per_run_timeout_ms = 30_000
        input.review_context = "Feedback item 1 of 1 — assessment panel context."

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        expect(response.output).toContain("dispatch_review_runs: ok")
        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("ok")
        expect(meta.runs_completed).toBe(1)
        expect(meta.runs_failed).toBe(0)
        expect(meta.runs_skipped).toBe(0)
        expect(Array.isArray(meta.findings)).toBe(true)
        expect(Array.isArray(meta.proposals)).toBe(true)
        expect(Array.isArray(meta.warnings)).toBe(true)
        const findings = meta.findings as Array<Record<string, unknown>>
        expect(findings).toHaveLength(1)
        expect(findings[0]?.category).toBe("documentation")
        const proposals = meta.proposals as Array<Record<string, unknown>>
        expect(proposals).toHaveLength(1)
        expect(proposals[0]?.tag).toBe("feedback-triage")
      })
    })
  })

  describe("per-run timeout derivation wiring (RD-FR-028)", () => {
    it("records derived timeout provenance in run files when per_run_timeout_ms is omitted", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeBaseInput(diffPath, [
          makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
        ])

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        expect(response.output).toContain("dispatch_review_runs")
        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("ok")
        expect(meta.runs_completed).toBe(1)

        const correlationId = meta.correlation_id as string
        trackDispatchDir(correlationId)
        const run = JSON.parse(
          await readFile(join(dispatchDir(correlationId), "run-divisor-guard.json"), "utf8"),
        ) as { prompt: string; timeout?: Record<string, unknown> }

        // RD-FR-028: the run's timeout derives from the work being
        // dispatched — fixture matrix limits (base 600 s, no ceiling or
        // per-tier override fields) resolved through the same limitsFor
        // defaults, standard-tier entry, 4-way parallel batching.
        const timeout = run.timeout
        expect(timeout).toBeDefined()
        expect(timeout!.source).toBe("derived")
        expect(timeout!.base_ms).toBe(600_000)
        expect(timeout!.resolved_tier).toBe("standard")
        expect(timeout!.session_profile_tier).toBe("standard")
        expect(timeout!.profile_term_ms).toBe(120_000)
        // TC-001: pin the prompt term's byte source, never just its type —
        // the persisted prompt (D4) is what promptSizeMap measured, so a
        // key-drifted promptSizeMap would zero the prompt term while a
        // typeof-only assertion still passes.
        expect(timeout!.prompt_bytes).toBe(Buffer.byteLength(run.prompt, "utf8"))
        expect(timeout!.contention_multiplier).toBe(1.45)
        expect(timeout!.max_parallel_runs).toBe(4)
        expect(timeout!.ceiling_ms).toBe(1_800_000)
        // Effective value follows the D1 formula over the recorded terms.
        const derived = Math.round(
          ((timeout!.base_ms as number) +
            (timeout!.prompt_term_ms as number) +
            (timeout!.profile_term_ms as number)) *
            (timeout!.contention_multiplier as number),
        )
        expect(timeout!.effective_ms).toBe(
          Math.min(
            Math.max(derived, timeout!.base_ms as number),
            timeout!.ceiling_ms as number,
          ),
        )
      })
    })

    it("derives the base window from the review matrix rather than the flat default (RD-FR-028)", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()

        // Overwrite the scratch matrix with a non-default base window so a
        // wired matrix limit (1200 s) is distinguishable from the 600 s flat
        // default the field replaced.
        const matrixPath = join(directory, ".uf", "review-matrix.yaml")
        const matrix = JSON.parse(await readFile(matrixPath, "utf8")) as {
          limits: { per_run_timeout_seconds: number }
        }
        matrix.limits.per_run_timeout_seconds = 1200
        await writeFile(matrixPath, JSON.stringify(matrix), "utf8")

        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeBaseInput(diffPath, [
          makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
        ])

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("ok")
        expect(meta.runs_completed).toBe(1)

        const correlationId = meta.correlation_id as string
        trackDispatchDir(correlationId)
        const run = JSON.parse(
          await readFile(join(dispatchDir(correlationId), "run-divisor-guard.json"), "utf8"),
        ) as { timeout?: Record<string, unknown> }

        // RD-FR-028: base_ms must reflect the matrix-configured 1200 s
        // window through loadPolicies -> limitsFor -> resolveBaseWindowMs,
        // never the 600 s flat default. The ceiling stays at the
        // 1,800,000 ms schema bound because the matrix omits
        // per_run_timeout_ceiling_seconds.
        const timeout = run.timeout
        expect(timeout).toBeDefined()
        expect(timeout!.source).toBe("derived")
        expect(timeout!.base_ms).toBe(1_200_000)
        expect(timeout!.resolved_tier).toBe("standard")
        expect(timeout!.profile_term_ms).toBe(120_000)
        expect(timeout!.max_parallel_runs).toBe(4)
        expect(timeout!.ceiling_ms).toBe(1_800_000)
        const effective = timeout!.effective_ms as number
        expect(effective).toBeGreaterThanOrEqual(1_200_000)
        expect(effective).toBeLessThanOrEqual(1_800_000)
      })
    })

    it("warns when matrix degradation shrinks derived windows to schema defaults (RD-FR-029)", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()

        // Corrupt the scratch matrix so loadPolicies throws at both
        // resolution points. resolveTimeoutLimits then degrades to the
        // schema defaults (base 600 s, RD-FR-029) — which MUST be visible
        // as a warning, never silent. The same corruption also fails every
        // per-run policy load (retryable policy_load_failed early exit), so
        // no run file is persisted and the dispatch-level degradation
        // contract is the warning plus the failed-run counts, not a run
        // file's base_ms. The redispatch companion test in
        // dispatch-recovery.test.ts covers the recovery-side warning.
        const matrixPath = join(directory, ".uf", "review-matrix.yaml")
        await writeFile(matrixPath, "limits: [unclosed", "utf8")

        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeBaseInput(diffPath, [
          makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
        ])

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        const warnings = meta.warnings as string[]
        expect(
          warnings.some((w: string) =>
            w.includes(
              "review matrix limits unavailable; per-run timeout derivation degraded to schema defaults",
            ),
          ),
        ).toBe(true)
        expect(meta.status).toBe("error")
        expect(meta.runs_completed).toBe(0)
        expect(meta.runs_failed).toBe(1)
      })
    })

    it("derives distinct base windows per tier from matrix per_run_timeout_tiers (RD-FR-029)", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()

        // Overwrite the scratch matrix with per-tier base-window overrides
        // so lightweight and heavy entries resolve different base windows
        // than the scalar limit (design D6: the per-tier override wins for
        // its tier; tiers without an override fall back to the scalar).
        const matrixPath = join(directory, ".uf", "review-matrix.yaml")
        const matrix = JSON.parse(await readFile(matrixPath, "utf8")) as {
          limits: {
            per_run_timeout_seconds: number
            per_run_timeout_tiers?: { lightweight?: number; standard?: number; heavy?: number }
          }
        }
        matrix.limits.per_run_timeout_tiers = { lightweight: 300, heavy: 1500 }
        await writeFile(matrixPath, JSON.stringify(matrix), "utf8")

        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeBaseInput(diffPath, [
          makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
          makePlanEntry({ agent: "divisor-architect", tier: "lightweight", sequence: 2 }),
          makePlanEntry({ agent: "divisor-testing", tier: "heavy", sequence: 3 }),
        ])

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("ok")
        expect(meta.runs_completed).toBe(3)

        const correlationId = meta.correlation_id as string
        trackDispatchDir(correlationId)
        const readTimeout = async (agent: string): Promise<Record<string, unknown>> => {
          const run = JSON.parse(
            await readFile(join(dispatchDir(correlationId), `run-${agent}.json`), "utf8"),
          ) as { timeout?: Record<string, unknown> }
          expect(run.timeout).toBeDefined()
          return run.timeout!
        }

        // Standard (no override): the scalar 600 s base and its 120 s
        // profile term.
        const standard = await readTimeout("divisor-guard")
        expect(standard.source).toBe("derived")
        expect(standard.base_ms).toBe(600_000)
        expect(standard.resolved_tier).toBe("standard")
        expect(standard.profile_term_ms).toBe(120_000)

        // Lightweight override wins for its tier; zero profile term.
        const lightweight = await readTimeout("divisor-architect")
        expect(lightweight.source).toBe("derived")
        expect(lightweight.base_ms).toBe(300_000)
        expect(lightweight.resolved_tier).toBe("lightweight")
        expect(lightweight.profile_term_ms).toBe(0)

        // Heavy override wins for its tier: base 1500 s plus the 300 s
        // heavy profile term, expanded by the 1.45 contention multiplier
        // (4-way batching), exceeds the 1,800,000 ms ceiling and clamps
        // there (D1 floor/ceiling composition).
        const heavy = await readTimeout("divisor-testing")
        expect(heavy.source).toBe("derived")
        expect(heavy.base_ms).toBe(1_500_000)
        expect(heavy.resolved_tier).toBe("heavy")
        expect(heavy.profile_term_ms).toBe(300_000)
        expect(heavy.ceiling_ms).toBe(1_800_000)
        expect(heavy.effective_ms).toBe(1_800_000)
        expect(heavy.clamped).toBe(true)
      })
    })

    it("records explicit timeout provenance when per_run_timeout_ms is provided", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeBaseInput(diffPath, [
          makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
        ])
        // Explicit argument wins over derivation (RD-FR-028) and the
        // 1_800_000 executor bound is inclusive (D4).
        input.per_run_timeout_ms = 1_800_000

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("ok")
        expect(meta.runs_completed).toBe(1)

        const correlationId = meta.correlation_id as string
        trackDispatchDir(correlationId)
        const run = JSON.parse(
          await readFile(join(dispatchDir(correlationId), "run-divisor-guard.json"), "utf8"),
        ) as { timeout?: Record<string, unknown> }

        expect(run.timeout).toBeDefined()
        expect(run.timeout!.source).toBe("explicit")
        expect(run.timeout!.effective_ms).toBe(1_800_000)
      })
    })

    it("rejects an explicit per_run_timeout_ms above the 1_800_000 executor bound", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeBaseInput(diffPath, [
          makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
        ])
        input.per_run_timeout_ms = 1_800_001

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("error")
        expect(meta.runs_completed).toBe(0)
        expect(meta.runs_failed).toBe(0)
        expect(meta.runs_skipped).toBe(0)
        expect(typeof meta.message).toBe("string")
        expect((meta.message as string).toLowerCase()).toContain("validation")
        expect(meta.retryable).toBe(false)
      })
    })
  })

  describe("manifest-mode retrieval allowance wiring (RD-FR-036)", () => {
    it("grants the retrieval allowance to manifest-mode runs and persists manifest_mode", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()

        // RD-FR-032: prompt assembly switches to manifest delivery
        // strictly above the 1500-line DIFF_THRESHOLD, so the scratch
        // diff must exceed it for the dispatch path to carry the
        // delivery-mode signal (RD-FR-036) into the derivation inputs.
        const largeDiff = [
          "diff --git a/src/foo.ts b/src/foo.ts",
          "--- a/src/foo.ts",
          "+++ b/src/foo.ts",
          "@@ -1 +1 @@",
          ...Array.from({ length: 1_600 }, (_, i) => `+line ${i + 1}`),
        ].join("\n")
        await writeFile(diffPath, largeDiff, "utf8")

        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeBaseInput(diffPath, [
          makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
        ])

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("ok")
        expect(meta.runs_completed).toBe(1)

        const correlationId = meta.correlation_id as string
        trackDispatchDir(correlationId)
        const run = JSON.parse(
          await readFile(join(dispatchDir(correlationId), "run-divisor-guard.json"), "utf8"),
        ) as { manifest_mode?: boolean; timeout?: Record<string, unknown> }

        // The delivery-mode signal persists as a work signal so the
        // recovery path re-grants the allowance without re-running
        // prompt assembly (design D8).
        expect(run.manifest_mode).toBe(true)

        const timeout = run.timeout
        expect(timeout).toBeDefined()
        expect(timeout!.source).toBe("derived")
        // TC-001: literal expectation (MANIFEST_RETRIEVAL_ALLOWANCE_MS),
        // never a value sourced from the production output.
        expect(timeout!.retrieval_allowance_ms).toBe(300_000)

        // The allowance is the composition's fifth term, added after the
        // contention multiplier: absent clamping, the effective window is
        // the pre-change formula plus exactly the allowance (design D8).
        const derivedWithoutAllowance = Math.round(
          ((timeout!.base_ms as number) +
            (timeout!.prompt_term_ms as number) +
            (timeout!.profile_term_ms as number)) *
            (timeout!.contention_multiplier as number),
        )
        expect(timeout!.clamped).toBe(false)
        expect(timeout!.effective_ms).toBe(derivedWithoutAllowance + 300_000)
      })
    })

    it("inline-mode runs derive without the allowance and persist manifest_mode false", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()

        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        const input = makeBaseInput(diffPath, [
          makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
        ])

        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

        const meta = response.metadata as Record<string, unknown>
        expect(meta.status).toBe("ok")
        expect(meta.runs_completed).toBe(1)

        const correlationId = meta.correlation_id as string
        trackDispatchDir(correlationId)
        const run = JSON.parse(
          await readFile(join(dispatchDir(correlationId), "run-divisor-guard.json"), "utf8"),
        ) as { manifest_mode?: boolean; timeout?: Record<string, unknown> }

        // The 4-line scratch diff stays inline (at or below the
        // threshold), so the derivation keeps the pre-change shape for
        // the same inputs: no retrieval_allowance_ms key at all.
        expect(run.manifest_mode).toBe(false)
        const timeout = run.timeout
        expect(timeout).toBeDefined()
        expect(timeout!.source).toBe("derived")
        expect(timeout).not.toHaveProperty("retrieval_allowance_ms")
        const derived = Math.round(
          ((timeout!.base_ms as number) +
            (timeout!.prompt_term_ms as number) +
            (timeout!.profile_term_ms as number)) *
            (timeout!.contention_multiplier as number),
        )
        expect(timeout!.effective_ms).toBe(
          Math.min(
            Math.max(derived, timeout!.base_ms as number),
            timeout!.ceiling_ms as number,
          ),
        )
      })
    })
  })

  it("records its operation window on successful dispatch (RD-FR-031)", async () => {
    await withBun(async () => {
      const { directory, diffPath } = await scratchProject()
      const hooks = await ReviewDispatchPlugin.server({
        client: fakeClient(),
        directory,
        worktree: directory,
      } as unknown as PluginInput)

      const input = makeBaseInput(diffPath, [
        makePlanEntry({ agent: "divisor-guard", sequence: 1 }),
        makePlanEntry({ agent: "divisor-architect", decision: "skip", sequence: 2 }),
      ])

      const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())

      const meta = response.metadata as Record<string, unknown>
      const correlationId = meta.correlation_id as string
      trackDispatchDir(correlationId)
      const sessionDir = dispatchDir(correlationId)
      const opFiles = (await readdir(sessionDir)).filter((name) => /^operation-.*\.json$/.test(name))
      expect(opFiles).toHaveLength(1)
      const record = JSON.parse(
        await readFile(join(sessionDir, opFiles[0]), "utf8"),
      ) as Record<string, unknown>
      expect(record.kind).toBe("dispatch")
      expect(typeof record.operation_id).toBe("string")
      expect(typeof record.started_at).toBe("string")
      expect(typeof record.completed_at).toBe("string")
      expect(record.status).toBe("ok")
      // TC-001: expected counts are literals, not values sourced from the
      // production output being verified — self-sourced expectations
      // verify themselves and cannot catch a counting drift.
      expect(record.run_counts).toMatchObject({
        total: 2,
        success: 1,
        failed: 0,
        skipped: 1,
        budget_skipped: 0,
        limit_skipped: 0,
        cancelled: 0,
      })
    })
  })
})
