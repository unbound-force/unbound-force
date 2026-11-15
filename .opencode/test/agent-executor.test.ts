/**
 * Tests for the agent-executor lib module's compaction-storm detector
 * (OpenSpec change compaction-storm-on-a-tight-context-mode, task 2.2;
 * design D2, DELEG-FR-002, issue #1079).
 *
 * Core rows: the threshold-reaching abort (code context_thrash, retryable,
 * timeout kind storm, resume envelope), the N-1 no-trigger boundary, the
 * substantive-progress reset, reads and failed non-read calls never
 * resetting the window (the storm fires through them), the configurable
 * threshold, the executor-side normalization of an injected threshold
 * (a bad value can neither disable the detector nor abort instantly),
 * and the probe-accounting cleanup at run end. Task 3.2 extends this
 * file with the full boundary matrix; the probe-level accounting rows
 * live in delegation-budget.test.ts.
 *
 * The final-text success guard suite (task 2.3, design D3, DELEG-FR-003)
 * covers the success-mapping reclassification: compaction-summary and
 * continuation-prompt final text never succeed (context_thrash at the
 * storm threshold, synthetic_final_text below it), a failed final probe
 * still reclassifies, legitimate prose that merely mentions compaction
 * stays success, and the preserved error text is secret-redacted and
 * length-bounded.
 *
 * Every run here omits idleTimeoutMilliseconds on purpose: the storm
 * detector is always armed, so the probe runs even when the idle
 * watchdog is disabled (the pre-change executor never probed without
 * an idle window).
 */
import type { PluginInput } from "@opencode-ai/plugin"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"

import {
  executeAgentSession,
  MAX_ERROR_LENGTH,
  MAX_PRESERVED_FINAL_TEXT_LENGTH,
  type InvocationProvenance,
} from "../lib/agent-executor.js"
import { registeredProbeCounters } from "../lib/delegation-budget.js"
import { DEFAULT_STORM_THRESHOLD } from "../lib/model-matrix.js"

type Client = PluginInput["client"]

const DIRECTORY = "/fake/project"
const POLL_MS = 1_000

const PROVENANCE: InvocationProvenance = {
  agent: "cobalt-crush-dev",
  requested_model: null,
  requested_variant: null,
  resolved_parent_model: "p/m",
  resolved_parent_variant: null,
  reported_child_model: null,
  model_mismatch: false,
  read_only: false,
}

interface ScriptOptions {
  /** Messages served to successive probes; the last repeats once the script is exhausted. */
  readonly script: readonly Record<string, unknown>[]
  /** Child prompt resolves after this many ms instead of hanging. */
  readonly resolveAfterMs?: number
  /** Final text the child prompt resolves with (default "done"). */
  readonly resolveText?: string
  /** Make every messages call from this 1-based call index on return no data (failed probe). */
  readonly failMessagesFromCall?: number
}

/** Script-driven executor fake: child creation, a hanging prompt, abort, and probe endpoints. */
function scriptClient(options: ScriptOptions) {
  const calls = { creates: 0, aborts: 0, probes: 0 }
  let index = -1
  const session = {
    create: async () => {
      calls.creates++
      return { data: { id: "storm-child" } }
    },
    prompt: async (request: { signal?: AbortSignal }) => {
      return await new Promise((resolve, reject) => {
        if (options.resolveAfterMs !== undefined) {
          setTimeout(
            () =>
              resolve({
                data: {
                  info: { providerID: "p", modelID: "m" },
                  parts: [{ type: "text", text: options.resolveText ?? "done" }],
                },
              }),
            options.resolveAfterMs,
          )
        }
        request.signal?.addEventListener("abort", () => reject(request.signal?.reason), { once: true })
      })
    },
    abort: async () => {
      calls.aborts++
      return { data: true }
    },
    children: async () => ({ data: [] }),
    messages: async () => {
      calls.probes++
      if (options.failMessagesFromCall !== undefined && calls.probes >= options.failMessagesFromCall) {
        return { data: undefined }
      }
      index = Math.min(index + 1, options.script.length - 1)
      return { data: [options.script[index]] }
    },
  }
  return { client: { session } as unknown as Client, calls }
}

function message(id: string, parts: readonly Record<string, unknown>[]) {
  return { info: { id, time: { completed: undefined } }, parts }
}

/** A compaction-marker message; distinct ids are distinct counted markers. */
const compaction = (id: string) => message(id, [{ type: "compaction", auto: true }])

/** A successfully completed non-read tool call (substantive progress). */
const completedTool = (id: string, tool: string) =>
  message(id, [{ id: `t-${id}`, type: "tool", tool, state: { status: "completed" } }])

/** A failed non-read tool call (never substantive progress). */
const failedTool = (id: string, tool: string) =>
  message(id, [{ id: `t-${id}`, type: "tool", tool, state: { status: "error" } }])

/** A user-role message carrying text (the host-injected marker shape). */
const userTextMessage = (id: string, text: string) => ({
  info: { id, role: "user", time: { completed: undefined } },
  parts: [{ type: "text", text }],
})

/** A structural compaction summary: the system-reminder wrapper. */
const SUMMARY_TEXT = "<system-reminder>\n## Objective\n- keep working\n</system-reminder>"

/** The host-injected synthetic continuation prompt. */
const CONTINUATION_TEXT =
  "Continue if you have next steps, or stop and ask for clarification if you are unsure how to proceed."

function run(
  client: Client,
  extra: { timeoutMilliseconds?: number; stormThreshold?: number; checkpointStep?: string } = {},
) {
  return executeAgentSession(
    {
      client,
      // Unique ids keep this file hermetic against the process-wide deadline
      // and probe registries other executor tests share (Symbol.for keys).
      sessionID: "storm-parent",
      directory: DIRECTORY,
      parentAbort: new AbortController().signal,
      timeoutMilliseconds: extra.timeoutMilliseconds ?? 60_000,
      idlePollMilliseconds: POLL_MS,
      stormThreshold: extra.stormThreshold,
      checkpointStep: extra.checkpointStep,
    },
    "cobalt-crush-dev",
    "prompt",
    { providerID: "p", modelID: "m", combined: "p/m" },
    undefined,
    false,
    PROVENANCE,
  )
}

describe("executeAgentSession compaction-storm detector (DELEG-FR-002, design D2)", () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })
  afterEach(() => {
    vi.useRealTimers()
  })

  it("aborts with context_thrash at the default threshold and carries the resume envelope", async () => {
    const fake = scriptClient({ script: [compaction("c1"), compaction("c2"), compaction("c3")] })
    const pending = run(fake.client)
    await vi.advanceTimersByTimeAsync(5_000)
    const result = await pending
    expect(result.status).toBe("failed")
    expect(result.error?.code).toBe("context_thrash")
    expect(result.error?.retryable).toBe(true)
    // The error names the storm window: the observed compaction count and the threshold.
    expect(result.error?.message).toContain("3 consecutive compaction events")
    expect(result.error?.message).toContain(`storm threshold ${DEFAULT_STORM_THRESHOLD}`)
    expect(result.timeout).toMatchObject({
      kind: "storm",
      idle_timeout_ms: null,
      budget_clamped: false,
      pending_question: false,
    })
    expect(result.resume).toEqual({
      project_path: DIRECTORY,
      agent_name: "cobalt-crush-dev",
      pending_checkpoint_step: null,
    })
    expect(result.childSessionID).toBe("storm-child")
    expect(fake.calls.aborts).toBe(1)
  })

  it("threads the pending checkpoint step into the resume envelope", async () => {
    const fake = scriptClient({ script: [compaction("c1"), compaction("c2"), compaction("c3")] })
    const pending = run(fake.client, { checkpointStep: "implement" })
    await vi.advanceTimersByTimeAsync(5_000)
    const result = await pending
    expect(result.error?.code).toBe("context_thrash")
    expect(result.resume).toEqual({
      project_path: DIRECTORY,
      agent_name: "cobalt-crush-dev",
      pending_checkpoint_step: "implement",
    })
  })

  it("does not trigger one compaction short of the threshold", async () => {
    // Two markers, then a quiet real message: compactionsSinceProgress
    // stays at 2 and the run ends at the wall clock, not in a storm.
    const fake = scriptClient({
      script: [compaction("c1"), compaction("c2"), message("m1", [{ id: "p1", type: "text", text: "quiet" }])],
    })
    const pending = run(fake.client, { timeoutMilliseconds: 10_000 })
    await vi.advanceTimersByTimeAsync(10_000)
    const result = await pending
    expect(result.error?.code).toBe("timeout")
    expect(result.timeout?.kind).toBe("wall_clock")
    expect(result.resume).toBeUndefined()
    expect(fake.calls.aborts).toBe(1)
  })

  it("resets the window on successful non-read progress", async () => {
    // Markers to N-1, a completed write (substantive progress resets the
    // counter), then two more markers: only 2 accrue after the reset, so
    // the storm never fires and the wall clock ends the run.
    const fake = scriptClient({
      script: [
        compaction("c1"),
        compaction("c2"),
        completedTool("w1", "write"),
        compaction("c3"),
        compaction("c4"),
      ],
    })
    const pending = run(fake.client, { timeoutMilliseconds: 12_000 })
    await vi.advanceTimersByTimeAsync(12_000)
    const result = await pending
    expect(result.error?.code).toBe("timeout")
    expect(result.timeout?.kind).toBe("wall_clock")
  })

  it("fires through read-only activity (reads never reset the window)", async () => {
    const fake = scriptClient({
      script: [compaction("c1"), compaction("c2"), completedTool("r1", "read"), compaction("c3")],
    })
    const pending = run(fake.client)
    await vi.advanceTimersByTimeAsync(5_000)
    const result = await pending
    expect(result.error?.code).toBe("context_thrash")
    expect(result.timeout?.kind).toBe("storm")
  })

  it("fires through failed non-read calls (failures never reset the window)", async () => {
    const fake = scriptClient({
      script: [compaction("c1"), compaction("c2"), failedTool("b1", "bash"), compaction("c3")],
    })
    const pending = run(fake.client)
    await vi.advanceTimersByTimeAsync(5_000)
    const result = await pending
    expect(result.error?.code).toBe("context_thrash")
    expect(result.timeout?.kind).toBe("storm")
  })

  it("honors a configured threshold below the default", async () => {
    const fake = scriptClient({ script: [compaction("c1"), compaction("c2")] })
    const pending = run(fake.client, { stormThreshold: 2 })
    await vi.advanceTimersByTimeAsync(4_000)
    const result = await pending
    expect(result.error?.code).toBe("context_thrash")
    expect(result.error?.message).toContain("storm threshold 2")
    expect(result.timeout?.kind).toBe("storm")
  })

  it("normalizes an invalid injected threshold to the documented default", async () => {
    // Zero would abort on the first probe (any accounting is >= 0) if it
    // reached the comparison unguarded; normalization keeps the detector
    // at the default 3, so the abort lands on the third marker.
    const fake = scriptClient({ script: [compaction("c1"), compaction("c2"), compaction("c3")] })
    const pending = run(fake.client, { stormThreshold: 0 })
    await vi.advanceTimersByTimeAsync(5_000)
    const result = await pending
    expect(result.error?.code).toBe("context_thrash")
    expect(result.error?.message).toContain(`storm threshold ${DEFAULT_STORM_THRESHOLD}`)
    expect(result.timeout?.elapsed_ms).toBe(3_000)
  })

  it("drops the probe accounting when the run ends", async () => {
    const fake = scriptClient({ script: [compaction("c1"), compaction("c2"), compaction("c3")] })
    const pending = run(fake.client)
    await vi.advanceTimersByTimeAsync(5_000)
    const result = await pending
    expect(result.error?.code).toBe("context_thrash")
    expect(registeredProbeCounters("storm-child")).toBeUndefined()
  })
})

describe("executeAgentSession final-text success guard (DELEG-FR-003, design D3)", () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })
  afterEach(() => {
    vi.useRealTimers()
  })

  it("reclassifies a compaction-summary final text to context_thrash when the storm threshold was met", async () => {
    // Interval polls observe two markers (below the default threshold of
    // three); the child then finishes with the injected summary as its
    // final text, and the guard's final probe folds the third user-role
    // marker the polls never sampled -- the threshold was met, so the
    // reclassification carries the storm code and resume envelope.
    const fake = scriptClient({
      script: [compaction("c1"), compaction("c2"), userTextMessage("c3", SUMMARY_TEXT)],
      resolveAfterMs: 2_500,
      resolveText: SUMMARY_TEXT,
    })
    const pending = run(fake.client)
    await vi.advanceTimersByTimeAsync(5_000)
    const result = await pending
    expect(result.status).toBe("failed")
    expect(result.error?.code).toBe("context_thrash")
    expect(result.error?.retryable).toBe(true)
    expect(result.error?.message).toContain("compaction-summary wrapper")
    expect(result.error?.message).toContain("3 consecutive compaction events")
    expect(result.error?.message).toContain(`storm threshold ${DEFAULT_STORM_THRESHOLD}`)
    expect(result.error?.message).toContain(SUMMARY_TEXT)
    expect(result.timeout).toMatchObject({ kind: "storm" })
    expect(result.resume).toEqual({
      project_path: DIRECTORY,
      agent_name: "cobalt-crush-dev",
      pending_checkpoint_step: null,
    })
    expect(result.childSessionID).toBe("storm-child")
    // The child had already finished; the guard reclassifies without aborting it again.
    expect(fake.calls.aborts).toBe(0)
  })

  it("reclassifies a compaction-summary final text to synthetic_final_text below the storm threshold", async () => {
    const fake = scriptClient({
      script: [message("m1", [{ id: "p1", type: "text", text: "quiet" }])],
      resolveAfterMs: 500,
      resolveText: SUMMARY_TEXT,
    })
    const pending = run(fake.client)
    await vi.advanceTimersByTimeAsync(2_000)
    const result = await pending
    expect(result.status).toBe("failed")
    expect(result.error?.code).toBe("synthetic_final_text")
    expect(result.error?.retryable).toBe(true)
    expect(result.error?.message).toContain("compaction-summary wrapper")
    expect(result.error?.message).toContain(SUMMARY_TEXT)
    // No limit ended the run, and the resume envelope is context_thrash-only.
    expect(result.timeout).toBeUndefined()
    expect(result.resume).toBeUndefined()
  })

  it("reclassifies the synthetic continuation prompt to synthetic_final_text", async () => {
    const fake = scriptClient({
      script: [message("m1", [{ id: "p1", type: "text", text: "quiet" }])],
      resolveAfterMs: 500,
      resolveText: CONTINUATION_TEXT,
    })
    const pending = run(fake.client)
    await vi.advanceTimersByTimeAsync(2_000)
    const result = await pending
    expect(result.status).toBe("failed")
    expect(result.error?.code).toBe("synthetic_final_text")
    expect(result.error?.retryable).toBe(true)
    expect(result.error?.message).toContain("synthetic continuation prompt")
    expect(result.error?.message).toContain(CONTINUATION_TEXT)
    expect(result.timeout).toBeUndefined()
    expect(result.resume).toBeUndefined()
  })

  it("still reclassifies when the final activity probe cannot run", async () => {
    // A failed final probe must neither crash the guard nor restore the
    // success mapping; the guard falls back to the (empty) poll-folded
    // accounting and emits the retryable synthetic code.
    const fake = scriptClient({
      script: [message("m1", [{ id: "p1", type: "text", text: "quiet" }])],
      resolveAfterMs: 500,
      resolveText: SUMMARY_TEXT,
      failMessagesFromCall: 1,
    })
    const pending = run(fake.client)
    await vi.advanceTimersByTimeAsync(2_000)
    const result = await pending
    expect(result.status).toBe("failed")
    expect(result.error?.code).toBe("synthetic_final_text")
    expect(result.error?.retryable).toBe(true)
  })

  it("keeps legitimate final text success, including prose that merely mentions compaction", async () => {
    // Negative control: the prose names compaction and even contains a
    // <system-reminder> tag mid-text, but it does not match the anchored,
    // length-bounded wrapper structure, so the success mapping survives.
    const legitimate =
      "Implementation complete. I reviewed the compaction summary the host injected earlier; " +
      "a <system-reminder> wrapper appeared mid-conversation, but this final message is real prose."
    const fake = scriptClient({
      script: [message("m1", [{ id: "p1", type: "text", text: "quiet" }])],
      resolveAfterMs: 500,
      resolveText: legitimate,
    })
    const pending = run(fake.client)
    await vi.advanceTimersByTimeAsync(2_000)
    const result = await pending
    expect(result.status).toBe("success")
    expect(result.error).toBeNull()
    expect(result.text).toBe(legitimate)
    // Legitimate text never pays the final-probe cost.
    expect(fake.calls.probes).toBe(0)
  })

  it("redacts secrets in the preserved final text", async () => {
    const fake = scriptClient({
      script: [message("m1", [{ id: "p1", type: "text", text: "quiet" }])],
      resolveAfterMs: 500,
      resolveText: `api_key: sk-live-9f8e7d6c5b ${CONTINUATION_TEXT}`,
    })
    const pending = run(fake.client)
    await vi.advanceTimersByTimeAsync(2_000)
    const result = await pending
    expect(result.error?.code).toBe("synthetic_final_text")
    expect(result.error?.message).toContain("api_key: [REDACTED]")
    expect(result.error?.message).not.toContain("sk-live-9f8e7d6c5b")
  })

  it("truncates the preserved final text to the documented maximum", async () => {
    const oversized = `<system-reminder>${"x".repeat(MAX_PRESERVED_FINAL_TEXT_LENGTH + 2_000)}TAIL_SENTINEL</system-reminder>`
    const fake = scriptClient({
      script: [message("m1", [{ id: "p1", type: "text", text: "quiet" }])],
      resolveAfterMs: 500,
      resolveText: oversized,
    })
    const pending = run(fake.client)
    await vi.advanceTimersByTimeAsync(2_000)
    const result = await pending
    expect(result.error?.code).toBe("synthetic_final_text")
    // The head of the wrapper is preserved for debugging...
    expect(result.error?.message).toContain("Preserved final text: <system-reminder>")
    expect(result.error?.message).toContain("x".repeat(64))
    // ...the tail beyond the documented bound is not, and the whole
    // message stays inside the error-sanitization bound.
    expect(result.error?.message).not.toContain("TAIL_SENTINEL")
    expect(result.error?.message.length).toBeLessThanOrEqual(MAX_ERROR_LENGTH)
  })
})
