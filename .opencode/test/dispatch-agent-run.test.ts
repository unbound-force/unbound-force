import type { PluginInput, ToolContext } from "@opencode-ai/plugin"
import { readFile, readdir, rm } from "node:fs/promises"
import { join } from "node:path"
import { describe, expect, it, vi, afterEach } from "vitest"

import type { InvokeAgentResult } from "../lib/agent-executor.js"
import type { TimeoutProvenanceRecord } from "../lib/run-timeout.js"
import ReviewDispatchPlugin from "../plugins/review-dispatch/index.js"
import {
  createDispatchAgentRunTool,
  dispatchAgentRun,
  executeDispatchRun,
  shouldReclassifyToSuccess,
  _getSessionCorrelationMap,
  _getSubmissionStore,
  type DispatchAgentRunDependencies,
  type PlannerDependencies,
} from "../plugins/review-dispatch/index.js"
import manifestFixture from "./fixtures/review-dispatch/manifest.json"
import matrixFixture from "./fixtures/review-dispatch/matrix.json"
import { withBun } from "./helpers/bun-runtime.js"
import {
  UUID_RE,
  dispatchDir,
  makeDispatchDirTracker,
  makePlanEntry,
  makeScratchProjectFactory,
  makeSessionMetadata,
} from "./helpers/aggregate-dispatch-fixtures.js"

// Track dispatch session dirs and scratch projects for cleanup.
const cleanupDirs: string[] = []

afterEach(async () => {
  _getSessionCorrelationMap().clear()
  // Submission-store entries are keyed by child session id; clear them so
  // tests that seed harvested submissions cannot leak into later tests.
  _getSubmissionStore().clear()
  for (const dir of cleanupDirs) {
    try {
      await rm(dir, { recursive: true, force: true })
    } catch {
      // best-effort
    }
  }
  cleanupDirs.length = 0
})

type Client = PluginInput["client"]

function jsonParser(text: string): unknown {
  return JSON.parse(text) as unknown
}

/**
 * Minimal fake client that succeeds for dispatch_agent_run calls.
 * Prompt response echoes the requested model identity so
 * model-mismatch checks pass.
 */
function fakeClient(): Client {
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
      // Echo the requested model so model-mismatch checks pass.
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

/**
 * Fake client whose host lookup (session.message) and child prompt are
 * independently shapable: prompt calls are captured so tests can assert
 * the execution seam — which model identity and variant the child
 * session actually received — while message overrides drive
 * host-resolution success, variant-less, failure, and abort shapes.
 * Shared by the RD-FR-037 host-source suite and the RD-FR-038
 * host-resolution failure-classification suite.
 */
function hostLookupClient(options?: {
  message?: () => Promise<unknown>
  promptCalls?: unknown[]
}): Client {
  const session = {
    create: async (): Promise<unknown> => ({
      data: { id: "child-session" },
      error: undefined,
    }),
    message:
      options?.message ??
      (async (): Promise<unknown> => ({
        data: {
          info: { id: "m", sessionID: "s", role: "assistant", providerID: "p", modelID: "m", variant: "active" },
          parts: [],
        },
        error: undefined,
      })),
    prompt: async (request: unknown): Promise<unknown> => {
      options?.promptCalls?.push(request)
      // Echo the requested model so model-mismatch checks pass.
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

function toolContext(): ToolContext {
  return {
    sessionID: "parent-session",
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

/**
 * Aggregate-dispatch fixtures shared with the sibling review-dispatch suites
 * (helpers/aggregate-dispatch-fixtures.ts, extracted per default.md CS-004);
 * the tracker and scratch project bind this suite's cleanup registry.
 */

/** Track a dispatch result for temp dir cleanup. */
const trackForCleanup = makeDispatchDirTracker(cleanupDirs)

/**
 * Scratch project for driving the aggregate dispatch_review_runs tool
 * through ReviewDispatchPlugin.server(): review matrix, reviewer
 * manifest, sibling-repos config, and a diff file on disk.
 */
const scratchProject = makeScratchProjectFactory("uf-agent-run-", cleanupDirs)

function makeDeps(overrides?: {
  readText?: (path: string) => Promise<string>
  client?: Client
  directory?: string
}): DispatchAgentRunDependencies {
  const readText =
    overrides?.readText ??
    vi.fn(async (path: string): Promise<string> => {
      if (path === ".uf/review-matrix.yaml") return JSON.stringify(matrixFixture)
      if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
      if (path === ".uf/review-matrix.override.yaml") {
        const err = new Error("ENOENT") as NodeJS.ErrnoException
        err.code = "ENOENT"
        throw err
      }
      throw new Error(`unexpected read: ${path}`)
    })

  return {
    plannerDependencies: {
      readText,
      parseYaml: jsonParser,
    },
    client: overrides?.client ?? fakeClient(),
    directory: overrides?.directory ?? "/workspace",
  }
}

// ── RD-FR-026 reclassification / D5 persistence fixtures ────

/** A finding that satisfies RunFindingSchema, as submitted before an abort. */
const SUBMITTED_FINDING = {
  severity: "MEDIUM",
  category: "standard",
  description: "Example finding submitted before the session abort.",
  root_cause: "Example root cause for the submitted finding.",
  file: "src/example.ts",
  line: 42,
} as const

/** A lesson proposal as submitted before an abort. */
const SUBMITTED_PROPOSAL = {
  information: "Example lesson submitted before the session abort.",
  tag: "review-dispatch",
} as const

/** Partial final-response text captured from an aborted child message. */
const PARTIAL_TEXT = "partial final review text"

/** Provenance for predicate fixtures; overrides flip one criterion at a time. */
function predicateProvenance(overrides?: Partial<InvokeAgentResult["provenance"]>): InvokeAgentResult["provenance"] {
  return {
    agent: "divisor-guard",
    requested_model: "provider/standard",
    requested_variant: "high",
    resolved_parent_model: null,
    resolved_parent_variant: null,
    reported_child_model: "provider/standard",
    model_mismatch: false,
    read_only: true,
    ...overrides,
  }
}

/**
 * Build an InvokeAgentResult shaped like the issue #751 post-submission
 * abort: failed / child_message_error with partial final-response text and
 * clean provenance. Each predicate test overrides exactly one criterion.
 */
function makeAbortResult(overrides?: Partial<InvokeAgentResult>): InvokeAgentResult {
  return {
    status: "failed",
    text: PARTIAL_TEXT,
    usage: null,
    error: { code: "child_message_error", message: "session aborted mid-message", retryable: false },
    provenance: predicateProvenance(),
    ...overrides,
  }
}

/**
 * Fake client whose child prompt response carries info.error — the
 * executor's child_message_error signature — while preserving the partial
 * final-response text. reportedModel forces a provider substitution so a
 * run can report a model mismatch alongside the mid-message abort.
 */
function abortingClient(options?: { reportedModel?: { providerID: string; modelID: string } }): Client {
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
      // Echo the requested model unless a substitution is forced, so
      // model-mismatch checks pass (or deliberately fail) per test.
      const req = request as { body?: { model?: { providerID: string; modelID: string } } }
      const providerID = options?.reportedModel?.providerID ?? req.body?.model?.providerID ?? "provider"
      const modelID = options?.reportedModel?.modelID ?? req.body?.model?.modelID ?? "standard"
      return {
        data: {
          info: {
            role: "assistant",
            providerID,
            modelID,
            error: { name: "MessageAbortedError", data: {} },
          },
          parts: [
            { id: "t", sessionID: "child-session", messageID: "m", type: "text", text: PARTIAL_TEXT, ignored: false },
          ],
        },
        error: undefined,
      }
    },
    abort: async (): Promise<unknown> => ({ data: true, error: undefined }),
  }
  return { session } as unknown as Client
}

/**
 * Fake client whose child prompt call fails outright — a plain executor
 * failure (child_prompt_failed) that still writes a run record.
 */
function promptFailureClient(): Client {
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
    prompt: async (): Promise<unknown> => ({ data: undefined, error: { message: "child prompt failed" } }),
    abort: async (): Promise<unknown> => ({ data: true, error: undefined }),
  }
  return { session } as unknown as Client
}

/** Read and parse the persisted run file for one agent in a dispatch session. */
async function readRunFile(correlationId: string, agent: string): Promise<Record<string, unknown>> {
  const raw = await readFile(join(dispatchDir(correlationId), `run-${agent}.json`), "utf8")
  return JSON.parse(raw) as Record<string, unknown>
}

/** Read and parse the persisted session-metadata.json for a dispatch session. */
async function readSessionMetadataFile(correlationId: string): Promise<Record<string, unknown>> {
  const raw = await readFile(join(dispatchDir(correlationId), "session-metadata.json"), "utf8")
  return JSON.parse(raw) as Record<string, unknown>
}

/** Names of run-*.json files in a dispatch session dir ([] when absent). */
async function listRunFiles(correlationId: string): Promise<string[]> {
  try {
    const entries = await readdir(dispatchDir(correlationId))
    return entries.filter((name) => name.startsWith("run-"))
  } catch {
    // The directory is never created when nothing was persisted for the session.
    return []
  }
}

describe("dispatch_agent_run", () => {
  describe("tier-based model resolution", () => {
    it("resolves model from review matrix for 'standard' tier", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", tier: "standard" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.text).toBe("review output")
      // Auto-generated correlation_id and run_id are valid UUIDs.
      expect(result.correlation_id).toMatch(UUID_RE)
      expect(result.run_id).toMatch(UUID_RE)
      // The matrix fixture has standard profile with model "provider/standard".
      expect(result.provenance.requested_model).toBe("provider/standard")
      expect(result.provenance.requested_variant).toBe("high")
      expect(result.provenance.agent).toBe("divisor-guard")
    })

    it("resolves model from review matrix for 'lightweight' tier", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", tier: "lightweight" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      // The matrix fixture has lightweight profile with model "provider/light".
      expect(result.provenance.requested_model).toBe("provider/light")
      expect(result.provenance.requested_variant).toBe("fast")
    })
  })

  describe("explicit model bypass", () => {
    it("uses model parameter directly without loading matrix for resolution", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", model: "custom-provider/custom-model@v2" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.provenance.requested_model).toBe("custom-provider/custom-model@v2")
      expect(result.provenance.agent).toBe("divisor-guard")
    })

    it("preserves variant when model is provided directly", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        {
          agent: "divisor-guard",
          prompt: "Review this change.",
          model: "custom-provider/custom-model",
          variant: "deep",
        },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.provenance.requested_model).toBe("custom-provider/custom-model")
      expect(result.provenance.requested_variant).toBe("deep")
    })
  })

  describe("default tier fallback", () => {
    it("defaults to 'standard' tier when neither tier nor model is provided", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change." },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.correlation_id).toMatch(UUID_RE)
      expect(result.run_id).toMatch(UUID_RE)
      // Should resolve via the "standard" profile: model "provider/standard".
      expect(result.provenance.requested_model).toBe("provider/standard")
      expect(result.provenance.requested_variant).toBe("high")
    })
  })

  describe("mutual exclusivity validation", () => {
    it("rejects when both tier and model are provided", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", tier: "standard", model: "provider/model" },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("invalid_input")
      expect(result.error?.message).toContain("tier and model are mutually exclusive")
    })

    it("rejects when both prompt and promptFile are provided", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "inline text", promptFile: "/tmp/prompt.txt" },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("invalid_input")
      expect(result.error?.message).toContain("prompt and promptFile are mutually exclusive")
    })

    it("rejects when neither prompt nor promptFile is provided", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", tier: "standard" },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("invalid_input")
      expect(result.error?.message).toContain("one of prompt or promptFile is required")
    })
  })

  describe("unknown agent rejection", () => {
    it("returns failed result for an agent not in the manifest", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-nonexistent", prompt: "Review this change." },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("unknown_agent")
      expect(result.error?.message).toContain("divisor-nonexistent")
      expect(result.error?.message).toContain("absent from the reviewer manifest")
    })
  })

  describe("missing review matrix", () => {
    it("returns a failed result when loadPolicies throws", async () => {
      const readText = vi.fn(async (_path: string): Promise<string> => {
        throw new Error("ENOENT: no such file /Users/alice/.uf/review-matrix.yaml")
      })
      const deps = makeDeps({ readText })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change." },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("policy_load_failed")
      expect(result.error?.retryable).toBe(true)
      // Verify path sanitization — /Users/alice should be redacted.
      expect(result.error?.message).not.toContain("/Users/alice")
      expect(result.error?.message).toContain("[REDACTED]")
    })
  })

  describe("tier with no model configured", () => {
    it("returns tier_model_unavailable when the profile has no model", async () => {
      const emptyProfileMatrix = {
        ...matrixFixture,
        profiles: { ...matrixFixture.profiles, standard: {} },
      }
      const readText = vi.fn(async (path: string): Promise<string> => {
        if (path === ".uf/review-matrix.yaml") return JSON.stringify(emptyProfileMatrix)
        if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
        if (path === ".uf/review-matrix.override.yaml") {
          const err = new Error("ENOENT") as NodeJS.ErrnoException
          err.code = "ENOENT"
          throw err
        }
        throw new Error(`unexpected read: ${path}`)
      })
      const deps = makeDeps({ readText })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", tier: "standard" },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("tier_model_unavailable")
      expect(result.error?.message).toContain("standard")
      expect(result.error?.message).toContain("no model configured")
      expect(result.error?.retryable).toBe(false)
    })
  })

  describe("invalid timeout", () => {
    it("rejects a timeout exceeding the schema maximum", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", timeout: 2_000_000 },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("invalid_input")
      expect(result.error?.retryable).toBe(false)
    })
  })

  describe("provenance metadata", () => {
    it("reflects tier-resolved model in requested_model", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", tier: "heavy" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.provenance).toMatchObject({
        agent: "divisor-guard",
        requested_model: "provider/heavy",
        resolved_parent_model: null,
        resolved_parent_variant: null,
        read_only: true,
        model_mismatch: false,
      })
      // Heavy profile has no variant configured, so requested_variant should be null.
      expect(result.provenance.requested_variant).toBeNull()
    })

    it("reflects direct model in requested_model", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        {
          agent: "divisor-adversary",
          prompt: "Review this change.",
          model: "google-vertex/claude-opus-4@default",
          variant: "fast",
          read_only: false,
        },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.provenance).toMatchObject({
        agent: "divisor-adversary",
        requested_model: "google-vertex/claude-opus-4@default",
        requested_variant: "fast",
        resolved_parent_model: null,
        resolved_parent_variant: null,
        read_only: false,
      })
    })

    it("sets read_only to true by default", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change." },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.provenance.read_only).toBe(true)
    })
  })

  describe("promptFile support", () => {
    it("reads prompt text from promptFile and executes successfully", async () => {
      const readText = vi.fn(async (path: string): Promise<string> => {
        if (path === "/tmp/prompt.txt") return "File-based prompt content."
        if (path === ".uf/review-matrix.yaml") return JSON.stringify(matrixFixture)
        if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
        if (path === ".uf/review-matrix.override.yaml") {
          const err = new Error("ENOENT") as NodeJS.ErrnoException
          err.code = "ENOENT"
          throw err
        }
        throw new Error(`unexpected read: ${path}`)
      })
      const deps = makeDeps({ readText })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", promptFile: "/tmp/prompt.txt" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.text).toBe("review output")
      expect(readText).toHaveBeenCalledWith("/tmp/prompt.txt")
    })

    it("returns a failed result when promptFile read fails", async () => {
      const readText = vi.fn(async (path: string): Promise<string> => {
        if (path === "/tmp/missing.txt") throw new Error("ENOENT: no such file /Users/alice/secret/prompt.txt")
        if (path === ".uf/review-matrix.yaml") return JSON.stringify(matrixFixture)
        if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
        if (path === ".uf/review-matrix.override.yaml") {
          const err = new Error("ENOENT") as NodeJS.ErrnoException
          err.code = "ENOENT"
          throw err
        }
        throw new Error(`unexpected read: ${path}`)
      })
      const deps = makeDeps({ readText })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", promptFile: "/tmp/missing.txt" },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("prompt_file_read_failed")
      expect(result.error?.retryable).toBe(false)
      // Verify path sanitization.
      expect(result.error?.message).not.toContain("/Users/alice")
      expect(result.error?.message).toContain("[REDACTED]")
    })

    it("rejects promptFile content exceeding 4 MiB", async () => {
      const oversizedContent = "x".repeat(4 * 1024 * 1024 + 1)
      const readText = vi.fn(async (path: string): Promise<string> => {
        if (path === "/tmp/big-prompt.txt") return oversizedContent
        if (path === ".uf/review-matrix.yaml") return JSON.stringify(matrixFixture)
        if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
        if (path === ".uf/review-matrix.override.yaml") {
          const err = new Error("ENOENT") as NodeJS.ErrnoException
          err.code = "ENOENT"
          throw err
        }
        throw new Error(`unexpected read: ${path}`)
      })
      const deps = makeDeps({ readText })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", promptFile: "/tmp/big-prompt.txt" },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("invalid_input")
      expect(result.error?.message).toContain("promptFile content exceeds 4 MiB")
    })
  })

  describe("session-based correlation", () => {
    it("reuses the same correlation_id for multiple calls in the same session", async () => {
      const deps = makeDeps()
      const ctx = toolContext()
      const result1 = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "First review." },
        ctx,
        deps,
      )
      trackForCleanup(result1.correlation_id)
      const result2 = await dispatchAgentRun(
        { agent: "divisor-adversary", prompt: "Second review." },
        ctx,
        deps,
      )

      expect(result1.correlation_id).toMatch(UUID_RE)
      expect(result2.correlation_id).toBe(result1.correlation_id)
      expect(result1.run_id).not.toBe(result2.run_id)
    })

    it("mints the correlation on the first direct call and reuses one session for a multi-call fan-out", async () => {
      const deps = makeDeps()
      const ctx = toolContext()
      // GIVEN a parent session with no prior aggregate dispatch
      // (RD-CR-003): the session map has no entry, so the first direct
      // call mints the correlation.
      expect(_getSessionCorrelationMap().get(ctx.sessionID)).toBeUndefined()

      const first = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Fan-out review one." },
        ctx,
        deps,
      )
      trackForCleanup(first.correlation_id)
      // The first call minted: the map now resolves the parent session.
      expect(_getSessionCorrelationMap().get(ctx.sessionID)).toBe(first.correlation_id)

      const second = await dispatchAgentRun(
        { agent: "divisor-adversary", prompt: "Fan-out review two." },
        ctx,
        deps,
      )
      const third = await dispatchAgentRun(
        { agent: "divisor-architect", prompt: "Fan-out review three." },
        ctx,
        deps,
      )

      // Reuse after mint: every call reports the first call's correlation
      // with its own run record. Per-invocation minting here would shatter
      // the fan-out into N sessions that consolidation cannot validate as
      // one plan (design D2 of fix-review-dispatch-correlation-release).
      for (const result of [first, second, third]) {
        expect(result.status).toBe("success")
        expect(result.correlation_id).toBe(first.correlation_id)
      }
      expect(new Set([first.run_id, second.run_id, third.run_id]).size).toBe(3)

      // All run files land under the single dispatch session directory.
      expect((await listRunFiles(first.correlation_id)).sort()).toEqual([
        "run-divisor-adversary.json",
        "run-divisor-architect.json",
        "run-divisor-guard.json",
      ])

      // The map still resolves the parent session to that one correlation.
      expect(_getSessionCorrelationMap().get(ctx.sessionID)).toBe(first.correlation_id)
    })
  })

  describe("direct runs after an aggregate dispatch (RD-CR-003)", () => {
    it("joins the latest aggregate dispatch session", async () => {
      await withBun(async () => {
        const { directory, diffPath } = await scratchProject()
        const hooks = await ReviewDispatchPlugin.server({
          client: fakeClient(),
          directory,
          worktree: directory,
        } as unknown as PluginInput)

        // One included plan entry: the aggregate completes a single run.
        const input = {
          mode: "code",
          command: "review-council",
          diff_path: diffPath,
          changed_files: "src/foo.ts",
          input_context: "base: main, head: feature",
          plan_entries: [makePlanEntry({ agent: "divisor-guard", sequence: 1 })],
          max_parallel_runs: 4,
          session_metadata: makeSessionMetadata(),
        }
        const response = await hooks.tool.dispatch_review_runs.execute(input, toolContext())
        const aggregate = response.metadata as Record<string, unknown>
        const aggregateCorrelation = aggregate.correlation_id as string
        trackForCleanup(aggregateCorrelation)

        // The aggregate minted a fresh correlation (RD-CR-001), completed
        // its run, reported the correlation (RD-CR-002), and re-pointed
        // the session map at it.
        expect(aggregate.status).toBe("ok")
        expect(aggregate.runs_completed).toBe(1)
        expect(aggregateCorrelation).toMatch(UUID_RE)
        expect(_getSessionCorrelationMap().get("parent-session")).toBe(aggregateCorrelation)

        // WHEN a direct single-run dispatch follows from the same parent
        // session, its mint branch stays dormant (design D2): the call
        // joins the latest aggregate session instead of creating one.
        const direct = await dispatchAgentRun(
          { agent: "divisor-adversary", prompt: "Direct follow-up review." },
          toolContext(),
          makeDeps(),
        )
        expect(direct.status).toBe("success")
        expect(direct.correlation_id).toBe(aggregateCorrelation)
        expect(_getSessionCorrelationMap().get("parent-session")).toBe(aggregateCorrelation)

        // THEN the direct run is persisted under the aggregate session's
        // dispatch directory, alongside the aggregate's own run.
        expect((await listRunFiles(aggregateCorrelation)).sort()).toEqual([
          "run-divisor-adversary.json",
          "run-divisor-guard.json",
        ])
      })
    })
  })

  describe("advisor host-model fallback", () => {
    it("inherits parent session model when advisor profile has no model", async () => {
      const emptyProfileMatrix = {
        ...matrixFixture,
        profiles: { ...matrixFixture.profiles, standard: {} },
      }
      const readText = vi.fn(async (path: string): Promise<string> => {
        if (path === ".uf/review-matrix.yaml") return JSON.stringify(emptyProfileMatrix)
        if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
        if (path === ".uf/review-matrix.override.yaml") {
          const err = new Error("ENOENT") as NodeJS.ErrnoException
          err.code = "ENOENT"
          throw err
        }
        throw new Error(`unexpected read: ${path}`)
      })
      const deps = makeDeps({ readText })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", tier: "standard", source: "advisor" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.provenance).toMatchObject({
        agent: "divisor-guard",
        requested_model: null,
        resolved_parent_model: "p/m",
        resolved_parent_variant: "active",
        read_only: true,
      })
      expect(result.provenance.requested_variant).toBeNull()
    })

    it("returns host_model_unavailable when host resolution fails", async () => {
      const emptyProfileMatrix = {
        ...matrixFixture,
        profiles: { ...matrixFixture.profiles, standard: {} },
      }
      const readText = vi.fn(async (path: string): Promise<string> => {
        if (path === ".uf/review-matrix.yaml") return JSON.stringify(emptyProfileMatrix)
        if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
        if (path === ".uf/review-matrix.override.yaml") {
          const err = new Error("ENOENT") as NodeJS.ErrnoException
          err.code = "ENOENT"
          throw err
        }
        throw new Error(`unexpected read: ${path}`)
      })
      const failingClient = {
        session: {
          create: async (): Promise<unknown> => ({
            data: { id: "child-session" },
            error: undefined,
          }),
          message: async (): Promise<unknown> => ({
            data: undefined,
            error: { message: "session lookup failed" },
          }),
          prompt: async (): Promise<unknown> => ({ data: undefined, error: undefined }),
          abort: async (): Promise<unknown> => ({ data: true, error: undefined }),
        },
      } as unknown as Client
      const deps = makeDeps({ readText, client: failingClient })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", tier: "standard", source: "advisor" },
        toolContext(),
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("host_model_unavailable")
      expect(result.error?.retryable).toBe(true)
    })

    it("returns cancelled when abort signal fires during host resolution", async () => {
      const emptyProfileMatrix = {
        ...matrixFixture,
        profiles: { ...matrixFixture.profiles, standard: {} },
      }
      const readText = vi.fn(async (path: string): Promise<string> => {
        if (path === ".uf/review-matrix.yaml") return JSON.stringify(emptyProfileMatrix)
        if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
        if (path === ".uf/review-matrix.override.yaml") {
          const err = new Error("ENOENT") as NodeJS.ErrnoException
          err.code = "ENOENT"
          throw err
        }
        throw new Error(`unexpected read: ${path}`)
      })
      const abortController = new AbortController()
      const hangingClient = {
        session: {
          create: async (): Promise<unknown> => ({
            data: { id: "child-session" },
            error: undefined,
          }),
          message: async (): Promise<unknown> => {
            abortController.abort()
            return { data: undefined, error: { message: "aborted" } }
          },
          prompt: async (): Promise<unknown> => ({ data: undefined, error: undefined }),
          abort: async (): Promise<unknown> => ({ data: true, error: undefined }),
        },
      } as unknown as Client
      const deps = makeDeps({ readText, client: hangingClient })
      const ctx = {
        ...toolContext(),
        abort: abortController.signal,
      }
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", tier: "standard", source: "advisor" },
        ctx,
        deps,
      )

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("cancelled")
      expect(result.error?.retryable).toBe(false)
    })
  })

  describe("host-source model resolution (RD-FR-037)", () => {
    it("inherits the parent session model, ignoring a configured profile model", async () => {
      // RD-FR-037 "Host run ignores a configured profile model": the default
      // matrixFixture configures provider/standard on the standard profile —
      // a host-source entry MUST execute on the resolved parent model p/m
      // instead and never consult the profile.
      const promptCalls: unknown[] = []
      const deps = makeDeps({ client: hostLookupClient({ promptCalls }) })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", source: "host" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.provenance).toMatchObject({
        agent: "divisor-guard",
        requested_model: null,
        resolved_parent_model: "p/m",
        resolved_parent_variant: "active",
        requested_variant: null,
        model_mismatch: false,
        read_only: true,
      })
      // Execution seam: the child session was invoked with the resolved host
      // model and variant, not the standard profile's provider/standard.
      expect(promptCalls).toHaveLength(1)
      const promptBody = (promptCalls[0] as { body?: { model?: { providerID: string; modelID: string }; variant?: string } }).body
      expect(promptBody?.model).toEqual({ providerID: "p", modelID: "m" })
      expect(promptBody?.variant).toBe("active")
    })

    it("returns a sanitized retryable host_model_unavailable when host resolution fails", async () => {
      // RD-FR-037 "Host resolution failure": the raw lookup error carries
      // redactable material (Bearer token, URL query string, home path) so
      // the assertion pins the branch's sanitizeInvocationError() form.
      const failingClient = hostLookupClient({
        message: async (): Promise<never> => {
          throw new Error(
            "host lookup failed with Bearer abc123def456 at https://api.internal.example.com/v1/message?token=xyz789 reading /Users/jflowers/state.json",
          )
        },
      })
      const deps = makeDeps({ client: failingClient })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", source: "host" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("host_model_unavailable")
      expect(result.error?.retryable).toBe(true)
      // The sanitizeInvocationError() form: no Bearer value, no query string,
      // no home path survives into the surfaced error message.
      expect(result.error?.message).toContain("Bearer [REDACTED]")
      expect(result.error?.message).not.toContain("abc123def456")
      expect(result.error?.message).toContain("?[REDACTED]")
      expect(result.error?.message).not.toContain("token=xyz789")
      expect(result.error?.message).toContain("~/[REDACTED]")
      expect(result.error?.message).not.toContain("/Users/jflowers")
      // Pre-execution early exit: no run file is written for the correlation.
      expect(await listRunFiles(result.correlation_id)).toEqual([])
    })

    it("returns cancelled when the abort signal fires during host resolution", async () => {
      const abortController = new AbortController()
      const hangingClient = hostLookupClient({
        message: async (): Promise<unknown> => {
          abortController.abort()
          return { data: undefined, error: { message: "aborted" } }
        },
      })
      const deps = makeDeps({ client: hangingClient })
      const ctx = {
        ...toolContext(),
        abort: abortController.signal,
      }
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", source: "host" },
        ctx,
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("cancelled")
      expect(result.error?.retryable).toBe(false)
      // Pre-execution early exit: no run file is written for the correlation.
      expect(await listRunFiles(result.correlation_id)).toEqual([])
    })

    it("executes the direct model bypass without consulting host resolution", async () => {
      // RD-FR-037 "Direct bypass overrides host resolution": the bypass
      // branch precedes source-based resolution, so a source "host" entry
      // carrying an explicit model records non-host provenance. The
      // scenario's finalization clause — fail-closed on the contradictory
      // host provenance — is pinned compositionally by
      // finalize-error-boundaries.test.ts ("rejects a host run that carries
      // requested model provenance").
      const message = vi.fn(async (): Promise<never> => {
        throw new Error("host lookup must not be consulted for a direct bypass")
      })
      const promptCalls: unknown[] = []
      const deps = makeDeps({ client: hostLookupClient({ message, promptCalls }) })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", source: "host", model: "provider/bypass" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.provenance.requested_model).toBe("provider/bypass")
      expect(result.provenance.resolved_parent_model).toBeNull()
      expect(message).not.toHaveBeenCalled()
      const promptBody = (promptCalls[0] as { body?: { model?: { providerID: string; modelID: string } } }).body
      expect(promptBody?.model).toEqual({ providerID: "provider", modelID: "bypass" })
    })

    it("records a null resolved parent variant when the host model has no active variant", async () => {
      // RD-FR-037 "Host run without an active variant": the parent
      // assistant message carries a model identity but no variant field.
      const variantlessClient = hostLookupClient({
        message: async (): Promise<unknown> => ({
          data: {
            info: { id: "m", sessionID: "s", role: "assistant", providerID: "p", modelID: "m" },
            parts: [],
          },
          error: undefined,
        }),
      })
      const deps = makeDeps({ client: variantlessClient })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", source: "host" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      expect(result.provenance.requested_model).toBeNull()
      expect(result.provenance.resolved_parent_model).toBe("p/m")
      expect(result.provenance.resolved_parent_variant).toBeNull()
      expect(result.provenance.model_mismatch).toBe(false)
    })
  })

  describe("host-resolution failure classification (RD-FR-038)", () => {
    /**
     * Deps whose review matrix configures no model on the standard profile
     * (the empty profile matrix from the advisor fallback suite), so an
     * advisor dispatch falls back to host-model resolution.
     */
    function emptyStandardProfileDeps(client: Client): DispatchAgentRunDependencies {
      const emptyProfileMatrix = {
        ...matrixFixture,
        profiles: { ...matrixFixture.profiles, standard: {} },
      }
      const readText = vi.fn(async (path: string): Promise<string> => {
        if (path === ".uf/review-matrix.yaml") return JSON.stringify(emptyProfileMatrix)
        if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
        if (path === ".uf/review-matrix.override.yaml") {
          const err = new Error("ENOENT") as NodeJS.ErrnoException
          err.code = "ENOENT"
          throw err
        }
        throw new Error(`unexpected read: ${path}`)
      })
      return makeDeps({ readText, client })
    }

    it("classifies a non-assistant parent message as non-retryable host_model_structural on the host branch", async () => {
      // RD-FR-038 "Structural failure on the host-source branch": the
      // parent session's current message is a user message, so
      // resolveHostModel() throws StructuralHostModelError. The failure is
      // non-retryable and preserves the remediation in the
      // uf_delegate_command message shape; no run file is written.
      const deps = makeDeps({
        client: hostLookupClient({
          message: async (): Promise<unknown> => ({
            data: {
              info: { id: "m", sessionID: "s", role: "user", providerID: "p", modelID: "m" },
              parts: [],
            },
            error: undefined,
          }),
        }),
      })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", source: "host" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("host_model_structural")
      expect(result.error?.retryable).toBe(false)
      expect(result.error?.message).toContain(
        "current message is not an assistant message. Remediation: re-invoke the command so host introspection reads the current assistant turn; retrying will not fix this.",
      )
      expect(await listRunFiles(result.correlation_id)).toEqual([])
    })

    it("classifies a missing model identity as host_model_structural on the advisor branch", async () => {
      // RD-FR-038 "Structural failure on the advisor fallback": the empty
      // standard profile drives the host fallback, and the parent assistant
      // message reports no model identity. The advisor branch classifies
      // through the same shared helper as the host branch — the split is
      // symmetric by construction, not by mirrored edits.
      const deps = emptyStandardProfileDeps(
        hostLookupClient({
          message: async (): Promise<unknown> => ({
            data: {
              info: { id: "m", sessionID: "s", role: "assistant", providerID: "", modelID: "" },
              parts: [],
            },
            error: undefined,
          }),
        }),
      )
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", tier: "standard", source: "advisor" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("host_model_structural")
      expect(result.error?.retryable).toBe(false)
      expect(result.error?.message).toContain(
        "current assistant model identity is unavailable. Remediation: run the command from a provider that reports the active model identity; retrying will not fix this.",
      )
      expect(await listRunFiles(result.correlation_id)).toEqual([])
    })

    it("classifies a present-but-invalid parent variant as host_model_structural", async () => {
      // RD-FR-038 "Invalid parent variant is structural": the variant is
      // present but fails the shared VariantSchema, so the classification
      // preserves the invalid-variant remediation.
      const deps = makeDeps({
        client: hostLookupClient({
          message: async (): Promise<unknown> => ({
            data: {
              info: { id: "m", sessionID: "s", role: "assistant", providerID: "p", modelID: "m", variant: "not valid" },
              parts: [],
            },
            error: undefined,
          }),
        }),
      })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", source: "host" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("host_model_structural")
      expect(result.error?.retryable).toBe(false)
      expect(result.error?.message).toContain(
        "current assistant active variant is invalid. Remediation: run the command from a provider whose active variant is a valid slug; retrying will not fix this.",
      )
      expect(await listRunFiles(result.correlation_id)).toEqual([])
    })

    it("keeps a thrown transient failure as retryable host_model_unavailable with the sanitized form", async () => {
      // RD-FR-038 "Transient failure keeps the retryable class": a plain
      // Error (not StructuralHostModelError) stays retryable with no
      // remediation suffix. Driven through the advisor branch so the shared
      // helper's transient leg — including sanitizeInvocationError()
      // redaction of a thrown error — is pinned on both branches.
      const deps = emptyStandardProfileDeps(
        hostLookupClient({
          message: async (): Promise<never> => {
            throw new Error(
              "host lookup failed with Bearer abc123def456 at https://api.internal.example.com/v1/message?token=xyz789 reading /Users/jflowers/state.json",
            )
          },
        }),
      )
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", tier: "standard", source: "advisor" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("host_model_unavailable")
      expect(result.error?.retryable).toBe(true)
      // The sanitizeInvocationError() form: no Bearer value, no query
      // string, no home path survives into the surfaced error message.
      expect(result.error?.message).toContain("Bearer [REDACTED]")
      expect(result.error?.message).not.toContain("abc123def456")
      expect(result.error?.message).toContain("?[REDACTED]")
      expect(result.error?.message).not.toContain("token=xyz789")
      expect(result.error?.message).toContain("~/[REDACTED]")
      expect(result.error?.message).not.toContain("/Users/jflowers")
      // Transient failures carry no remediation suffix.
      expect(result.error?.message).not.toContain("Remediation:")
      expect(await listRunFiles(result.correlation_id)).toEqual([])
    })

    it("keeps cancellation precedence over the structural classification", async () => {
      // RD-FR-038 "Abort takes precedence over classification": the
      // message override aborts the controller and returns a non-assistant
      // message — a shape that would otherwise classify as
      // host_model_structural — so the cancelled code wins.
      const abortController = new AbortController()
      const deps = makeDeps({
        client: hostLookupClient({
          message: async (): Promise<unknown> => {
            abortController.abort()
            return {
              data: {
                info: { id: "m", sessionID: "s", role: "user", providerID: "p", modelID: "m" },
                parts: [],
              },
              error: undefined,
            }
          },
        }),
      })
      const ctx = {
        ...toolContext(),
        abort: abortController.signal,
      }
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", source: "host" },
        ctx,
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("cancelled")
      expect(result.error?.retryable).toBe(false)
      expect(await listRunFiles(result.correlation_id)).toEqual([])
    })
  })

  describe("shouldReclassifyToSuccess predicate", () => {
    it("returns true for a post-submission child_message_error abort with findings", () => {
      // The issue #751 shape: submissions accepted, final response begun,
      // session aborted mid-message, no model substitution.
      expect(shouldReclassifyToSuccess(makeAbortResult(), [SUBMITTED_FINDING], [])).toBe(true)
    })

    it("returns true when only lesson proposals were submitted", () => {
      expect(shouldReclassifyToSuccess(makeAbortResult(), [], [SUBMITTED_PROPOSAL])).toBe(true)
    })

    it("returns false when the run did not fail (cancelled status)", () => {
      expect(shouldReclassifyToSuccess(makeAbortResult({ status: "cancelled" }), [SUBMITTED_FINDING], [])).toBe(false)
    })

    it("returns false when the error code is a timeout, not a message abort", () => {
      const timedOut = makeAbortResult({
        error: { code: "timeout", message: "invoke_agent run timed out", retryable: true },
      })
      expect(shouldReclassifyToSuccess(timedOut, [SUBMITTED_FINDING], [])).toBe(false)
    })

    it("returns false when nothing was submitted before the abort", () => {
      expect(shouldReclassifyToSuccess(makeAbortResult(), [], [])).toBe(false)
    })

    it("returns false when the final-response text is empty", () => {
      expect(shouldReclassifyToSuccess(makeAbortResult({ text: "" }), [SUBMITTED_FINDING], [])).toBe(false)
    })

    it("returns false on a model-mismatched child despite submissions and text", () => {
      // The executor captures reported-model provenance BEFORE the
      // child_message_error return, so a provider-substituted child can
      // carry both signals. Such a run must stay failed: validateRunModel
      // rejects any success run with a model mismatch, so reclassifying it
      // would brick the finalization payload.
      const substituted = makeAbortResult({
        provenance: predicateProvenance({
          reported_child_model: "provider/substituted",
          model_mismatch: true,
        }),
      })
      expect(shouldReclassifyToSuccess(substituted, [SUBMITTED_FINDING], [])).toBe(false)
    })
  })

  describe("post-submission reclassification (RD-FR-026)", () => {
    it("records a reclassified run as success with an abort audit while the caller sees the raw failure", async () => {
      _getSubmissionStore().set("child-session", { findings: [SUBMITTED_FINDING], proposals: [] })
      const deps = makeDeps({ client: abortingClient() })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change." },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      // Caller-visible contract is unchanged: the raw executor failure.
      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("child_message_error")
      expect(result.text).toBe(PARTIAL_TEXT)
      expect(result.provenance.model_mismatch).toBe(false)

      // The persisted run file is the source of truth for consolidation.
      const runFile = await readRunFile(result.correlation_id, "divisor-guard")
      expect(runFile.status).toBe("success")
      expect(runFile.error).toBeNull()
      expect(runFile.text).toBe(PARTIAL_TEXT)
      expect(runFile.findings).toEqual([SUBMITTED_FINDING])
      expect(runFile.post_submission_abort).toEqual({
        original_status: "failed",
        error_code: "child_message_error",
        error_message: expect.stringContaining("MessageAbortedError"),
        note: expect.stringContaining("RD-FR-026"),
      })
    })

    it("keeps a model-mismatched child_message_error run failed", async () => {
      _getSubmissionStore().set("child-session", { findings: [SUBMITTED_FINDING], proposals: [] })
      const deps = makeDeps({
        client: abortingClient({ reportedModel: { providerID: "provider", modelID: "substituted" } }),
      })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change." },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      // The abort carries both child_message_error and model_mismatch: the
      // executor reports substituted provenance before the abort return.
      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("child_message_error")
      expect(result.provenance.model_mismatch).toBe(true)

      const runFile = await readRunFile(result.correlation_id, "divisor-guard")
      expect(runFile.status).toBe("failed")
      expect(runFile.model_mismatch).toBe(true)
      expect(runFile.error).toMatchObject({ code: "child_message_error" })
      expect("post_submission_abort" in runFile).toBe(false)
    })

    it("keeps a child_message_error run failed when nothing was submitted", async () => {
      const deps = makeDeps({ client: abortingClient() })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change." },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("child_message_error")

      const runFile = await readRunFile(result.correlation_id, "divisor-guard")
      expect(runFile.status).toBe("failed")
      expect(runFile.findings).toEqual([])
      expect(runFile.proposals).toEqual([])
      expect(runFile.error).toMatchObject({ code: "child_message_error" })
      expect("post_submission_abort" in runFile).toBe(false)
    })
  })

  describe("post_submission_abort audit field", () => {
    it("is absent from a plain successful run file", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change." },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      const runFile = await readRunFile(result.correlation_id, "divisor-guard")
      expect(runFile.status).toBe("success")
      expect(runFile.error).toBeNull()
      expect("post_submission_abort" in runFile).toBe(false)
    })

    it("is absent from a plain failed run file", async () => {
      const deps = makeDeps({ client: promptFailureClient() })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change." },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("child_prompt_failed")
      const runFile = await readRunFile(result.correlation_id, "divisor-guard")
      expect(runFile.status).toBe("failed")
      expect(runFile.error).toMatchObject({ code: "child_prompt_failed" })
      expect("post_submission_abort" in runFile).toBe(false)
    })
  })

  describe("prompt persistence", () => {
    it("persists the inline prompt verbatim in the run record", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change line by line." },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      const runFile = await readRunFile(result.correlation_id, "divisor-guard")
      expect(runFile.prompt).toBe("Review this change line by line.")
    })

    it("persists the promptFile contents verbatim in the run record", async () => {
      const readText = vi.fn(async (path: string): Promise<string> => {
        if (path === "/tmp/prompt.txt") return "File-based prompt content."
        if (path === ".uf/review-matrix.yaml") return JSON.stringify(matrixFixture)
        if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
        if (path === ".uf/review-matrix.override.yaml") {
          const err = new Error("ENOENT") as NodeJS.ErrnoException
          err.code = "ENOENT"
          throw err
        }
        throw new Error(`unexpected read: ${path}`)
      })
      const deps = makeDeps({ readText })
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", promptFile: "/tmp/prompt.txt" },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("success")
      const runFile = await readRunFile(result.correlation_id, "divisor-guard")
      expect(runFile.prompt).toBe("File-based prompt content.")
    })

    it("writes no run record when a pre-execution early exit fires", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        {
          agent: "divisor-guard",
          promptFile: "/tmp/missing.txt",
          session_metadata: makeSessionMetadata(),
        },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      expect(result.status).toBe("failed")
      expect(result.error?.code).toBe("prompt_file_read_failed")
      // Session metadata was still persisted, but no run record exists —
      // the early exit fired before executeDispatchRun could write one.
      expect(await listRunFiles(result.correlation_id)).toEqual([])
      expect((await readSessionMetadataFile(result.correlation_id)).mode).toBe("code")
    })
  })

  describe("session metadata refresh", () => {
    it("writes session-metadata.json on the first call that supplies it", async () => {
      const deps = makeDeps()
      const result = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", session_metadata: makeSessionMetadata() },
        toolContext(),
        deps,
      )
      trackForCleanup(result.correlation_id)

      const metadata = await readSessionMetadataFile(result.correlation_id)
      expect(metadata.mode).toBe("code")
      expect(metadata.input_context).toMatchObject({ kind: "local", head_ref: "feature" })
    })

    it("replaces the file when a subsequent call supplies fresh metadata", async () => {
      const deps = makeDeps()
      const ctx = toolContext()
      const first = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", session_metadata: makeSessionMetadata() },
        ctx,
        deps,
      )
      trackForCleanup(first.correlation_id)
      const second = await dispatchAgentRun(
        {
          agent: "divisor-guard",
          prompt: "Review this change again.",
          session_metadata: makeSessionMetadata({ mode: "specs", headRef: "fix-loop" }),
        },
        ctx,
        deps,
      )

      // Same parent session resolves to the same dispatch session dir.
      expect(second.correlation_id).toBe(first.correlation_id)
      const metadata = await readSessionMetadataFile(second.correlation_id)
      expect(metadata.mode).toBe("specs")
      expect(metadata.input_context).toMatchObject({ head_ref: "fix-loop" })
    })

    it("leaves the file untouched when a subsequent call omits session_metadata", async () => {
      const deps = makeDeps()
      const ctx = toolContext()
      const first = await dispatchAgentRun(
        { agent: "divisor-guard", prompt: "Review this change.", session_metadata: makeSessionMetadata() },
        ctx,
        deps,
      )
      trackForCleanup(first.correlation_id)
      await dispatchAgentRun(
        { agent: "divisor-adversary", prompt: "Review this change again." },
        ctx,
        deps,
      )

      const metadata = await readSessionMetadataFile(first.correlation_id)
      expect(metadata.mode).toBe("code")
      expect(metadata.input_context).toMatchObject({ head_ref: "feature" })
    })
  })

  describe("createDispatchAgentRunTool factory", () => {
    it("creates a tool with description and expected args", () => {
      const deps = makeDeps()
      const dispatchTool = createDispatchAgentRunTool(deps)

      expect(dispatchTool.description).toContain("Divisor review-agent run")
      expect(Object.keys(dispatchTool.args).sort()).toEqual([
        "agent",
        "model",
        "prompt",
        "promptFile",
        "read_only",
        "sequence",
        "session_metadata",
        "source",
        "tier",
        "timeout",
        "variant",
      ])
    })
  })

  describe("run-record timeout provenance (RD-FR-030)", () => {
    it("persists timeout provenance in the run file when resolved", async () => {
      // RD-FR-030 / design D5: dispatch_review_runs and redispatch_failed_runs
      // resolve each run's timeout through derivation and hand the provenance
      // record to executeDispatchRun, which persists it in the run file.
      const derivedTimeout: TimeoutProvenanceRecord = {
        effective_ms: 1_800_000,
        source: "derived",
        base_ms: 1_200_000,
        prompt_bytes: 179_200,
        prompt_term_ms: 166_500,
        profile_term_ms: 120_000,
        resolved_tier: "standard",
        session_profile_tier: "standard",
        contention_multiplier: 1.45,
        max_parallel_runs: 4,
        ceiling_ms: 1_800_000,
        clamped: true,
      }
      const outcome = await executeDispatchRun(
        {
          agent: "divisor-guard",
          promptText: "Review this change.",
          correlationId: "1b0f6c2e-6f4a-4d21-9a77-3c9f1d2e4b5c",
          sequence: 1,
          readOnly: true,
          timeoutMilliseconds: 600_000,
          timeoutProvenance: derivedTimeout,
          source: "explicit",
          model: "provider/standard",
          tier: "standard",
          variant: null,
        },
        toolContext(),
        makeDeps(),
      )
      trackForCleanup(outcome.correlation_id)

      expect(outcome.result.status).toBe("success")
      expect(outcome.persisted).toBe(true)
      expect(outcome.runData?.timeout).toMatchObject(derivedTimeout)

      const run = await readRunFile(outcome.correlation_id, "divisor-guard")
      expect(run.timeout).toMatchObject(derivedTimeout)
    })

    it("omits the timeout key when no provenance was resolved", async () => {
      // dispatch_agent_run keeps explicit-arg-or-flat-default semantics
      // (design non-goals): when the calling tool did not resolve a timeout,
      // the run file carries no timeout key at all — un-resolved runs stay
      // byte-compatible with older run files.
      const outcome = await executeDispatchRun(
        {
          agent: "divisor-guard",
          promptText: "Review this change.",
          correlationId: "2c1f7d3f-7a5b-4e32-8b88-4d0a2e3f5c6d",
          sequence: 1,
          readOnly: true,
          timeoutMilliseconds: 600_000,
          source: "explicit",
          model: "provider/standard",
          tier: "standard",
          variant: null,
        },
        toolContext(),
        makeDeps(),
      )
      trackForCleanup(outcome.correlation_id)

      expect(outcome.persisted).toBe(true)
      expect(outcome.runData?.timeout).toBeUndefined()

      const run = await readRunFile(outcome.correlation_id, "divisor-guard")
      expect("timeout" in run).toBe(false)
    })
  })
})
