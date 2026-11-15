import type { PluginInput } from "@opencode-ai/plugin"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"

import { executeAgentSession, type InvocationProvenance } from "../lib/agent-executor.js"
import {
  BUDGET_RESERVE_MS,
  MAX_ANCESTRY_DEPTH,
  MAX_CONTINUATION_PROMPT_LENGTH,
  MAX_PROBE_SESSIONS,
  MAX_SYNTHETIC_WRAPPER_LENGTH,
  MIN_NESTED_BUDGET_MS,
  READ_ONLY_TOOLS,
  clampToBudget,
  classifySyntheticText,
  isCompactionMarker,
  isSubstantiveProgress,
  messageFingerprint,
  probeSubtreeActivity,
  registerDeadline,
  registeredDeadline,
  registeredProbeCounters,
  resolveInheritedDeadline,
  unregisterDeadline,
  unregisterProbeState,
} from "../lib/delegation-budget.js"

type Client = PluginInput["client"]

const DIRECTORY = "/fake/project"
const signal = new AbortController().signal
const registered: string[] = []

/** Register a deadline that afterEach removes, keeping the shared registry clean. */
function register(sessionID: string, deadline: number): void {
  registered.push(sessionID)
  registerDeadline(sessionID, deadline)
}

afterEach(() => {
  for (const id of registered.splice(0)) {
    unregisterDeadline(id)
  }
})

interface FakeSession {
  readonly parentID?: string
  readonly created?: number
  readonly children?: readonly string[]
  readonly latest?: Record<string, unknown>
}

interface TreeOptions {
  readonly sessions?: Record<string, FakeSession>
  readonly failGet?: boolean
  readonly failMessages?: boolean
  readonly failMessagesFor?: readonly string[]
  readonly undefinedChildren?: boolean
  readonly undefinedMessages?: boolean
}

/** Fake client exposing only the session endpoints the budget library reads. */
function treeClient(options: TreeOptions): { readonly client: Client; readonly gets: string[]; readonly probed: string[] } {
  const gets: string[] = []
  const probed: string[] = []
  const sessions = options.sessions ?? {}
  const session = {
    get: async (request: { path: { id: string } }) => {
      gets.push(request.path.id)
      if (options.failGet === true) {
        throw new Error("get failed")
      }
      return { data: { id: request.path.id, parentID: sessions[request.path.id]?.parentID } }
    },
    messages: async (request: { path: { id: string } }) => {
      probed.push(request.path.id)
      if (options.failMessages === true || options.failMessagesFor?.includes(request.path.id) === true) {
        throw new Error("messages failed")
      }
      if (options.undefinedMessages === true) {
        return { data: undefined }
      }
      const latest = sessions[request.path.id]?.latest
      return { data: latest === undefined ? [] : [latest] }
    },
    children: async (request: { path: { id: string } }) => {
      if (options.undefinedChildren === true) {
        return { data: undefined }
      }
      const ids = sessions[request.path.id]?.children ?? []
      return { data: ids.map((id) => ({ id, time: { created: sessions[id]?.created ?? 0 } })) }
    },
  }
  return { client: { session } as unknown as Client, gets, probed }
}

function message(id: string, parts: readonly Record<string, unknown>[], extra: Record<string, unknown> = {}) {
  return { info: { id, time: { completed: undefined }, ...extra }, parts }
}

// ---------------------------------------------------------------------------
// Registry and inheritance (design D1)
// ---------------------------------------------------------------------------

describe("deadline registry", () => {
  it("registers, reads, and unregisters a deadline", () => {
    register("s1", 1_000)
    expect(registeredDeadline("s1")).toBe(1_000)
    unregisterDeadline("s1")
    expect(registeredDeadline("s1")).toBeUndefined()
  })

  it("shares one registry through the global symbol", () => {
    register("s-global", 42)
    const holder = globalThis as unknown as Record<symbol, Map<string, number>>
    expect(holder[Symbol.for("unbound-force.delegation-budget")]?.get("s-global")).toBe(42)
  })
})

describe("resolveInheritedDeadline", () => {
  it("returns a direct registry hit without any API call", async () => {
    register("child", 5_000)
    const fake = treeClient({})
    expect(await resolveInheritedDeadline(fake.client, "child", DIRECTORY, signal)).toBe(5_000)
    expect(fake.gets).toEqual([])
  })

  it("walks parentID links to the nearest registered ancestor", async () => {
    register("budgeted", 9_000)
    const fake = treeClient({ sessions: { task: { parentID: "mid" }, mid: { parentID: "budgeted" } } })
    expect(await resolveInheritedDeadline(fake.client, "task", DIRECTORY, signal)).toBe(9_000)
    expect(fake.gets).toEqual(["task", "mid"])
  })

  it("returns null for a top-level session", async () => {
    const fake = treeClient({ sessions: { root: {} } })
    expect(await resolveInheritedDeadline(fake.client, "root", DIRECTORY, signal)).toBeNull()
  })

  it("stops after MAX_ANCESTRY_DEPTH hops", async () => {
    const sessions: Record<string, FakeSession> = {}
    for (let index = 0; index < MAX_ANCESTRY_DEPTH + 2; index++) {
      sessions[`s${index}`] = { parentID: `s${index + 1}` }
    }
    register(`s${MAX_ANCESTRY_DEPTH + 1}`, 1)
    const fake = treeClient({ sessions })
    expect(await resolveInheritedDeadline(fake.client, "s0", DIRECTORY, signal)).toBeNull()
    expect(fake.gets).toHaveLength(MAX_ANCESTRY_DEPTH)
  })

  it("treats lookup errors and clients without session.get as no budget", async () => {
    expect(await resolveInheritedDeadline(treeClient({ failGet: true }).client, "x", DIRECTORY, signal)).toBeNull()
    expect(await resolveInheritedDeadline({ session: {} } as unknown as Client, "x", DIRECTORY, signal)).toBeNull()
  })
})

describe("clampToBudget", () => {
  it("passes the request through without an inherited deadline", () => {
    expect(clampToBudget(600_000, null, 0)).toEqual({ effectiveMs: 600_000, clamped: false, exhausted: false })
  })

  it("keeps a request that fits inside the remaining budget", () => {
    const deadline = 2_000_000 + BUDGET_RESERVE_MS
    expect(clampToBudget(600_000, deadline, 0)).toEqual({ effectiveMs: 600_000, clamped: false, exhausted: false })
  })

  it("clamps to the remaining budget minus the reserve", () => {
    const deadline = 20 * 60_000
    expect(clampToBudget(1_800_000, deadline, 0)).toEqual({
      effectiveMs: deadline - BUDGET_RESERVE_MS,
      clamped: true,
      exhausted: false,
    })
  })

  it("reports exhaustion below the minimum nested budget", () => {
    const deadline = BUDGET_RESERVE_MS + MIN_NESTED_BUDGET_MS - 1
    expect(clampToBudget(1_800_000, deadline, 0)).toEqual({
      effectiveMs: MIN_NESTED_BUDGET_MS - 1,
      clamped: true,
      exhausted: true,
    })
    expect(clampToBudget(1_800_000, 0, 10_000).effectiveMs).toBe(0)
  })
})

// ---------------------------------------------------------------------------
// Activity probe (design D2)
// ---------------------------------------------------------------------------

describe("messageFingerprint", () => {
  it("distinguishes an absent message", () => {
    expect(messageFingerprint("s", undefined)).toBe("s:-")
  })

  it("changes when streamed text or tool output grows", () => {
    const short = messageFingerprint("s", message("m", [{ id: "p", type: "text", text: "ab" }]))
    const long = messageFingerprint("s", message("m", [{ id: "p", type: "text", text: "abc" }]))
    expect(short).not.toBe(long)
    const running = messageFingerprint("s", message("m", [{ id: "t", type: "tool", state: { status: "running" } }]))
    const output = messageFingerprint(
      "s",
      message("m", [{ id: "t", type: "tool", state: { status: "running", output: "x" } }]),
    )
    expect(running).not.toBe(output)
  })

  it("changes when a running tool's metadata output or title changes", () => {
    const tool = (title: string, output: string) =>
      messageFingerprint(
        "s",
        message("m", [{ id: "t", type: "tool", tool: "bash", state: { status: "running", title, metadata: { output } } }]),
      )
    expect(tool("go test", "a")).not.toBe(tool("go test", "ab"))
    expect(tool("go test", "a")).not.toBe(tool("go vet", "a"))
    const circular: Record<string, unknown> = {}
    circular.self = circular
    expect(
      messageFingerprint("s", message("m", [{ id: "t", type: "tool", state: { status: "running", metadata: circular } }])),
    ).toBe("s:m:::1:t:tool:running::0")
  })

  it("records completion and error presence", () => {
    const open = messageFingerprint("s", message("m", []))
    const failed = messageFingerprint("s", message("m", [], { error: { name: "x" } }))
    expect(open).not.toBe(failed)
    expect(messageFingerprint("s", { parts: [] })).toBe("s::::0:::::0")
  })

  it("collapses compaction markers to one constant fingerprint", () => {
    // DT-FR-002: compaction-to-compaction churn must not read as
    // progress, so every marker shape fingerprints identically.
    const trigger = messageFingerprint("s", message("c1", [{ type: "compaction", auto: true }]))
    const summary = messageFingerprint(
      "s",
      message("c2", [{ id: "p", type: "text", text: "a different, longer summary body" }], { mode: "compaction" }),
    )
    const continuation = messageFingerprint(
      "s",
      message("c3", [
        { type: "text", metadata: { compaction_continue: true }, text: "Continue if you have next steps, or stop and ask for clarification if you are unsure how to proceed." },
      ]),
    )
    expect(trigger).toBe("s:synthetic")
    expect(summary).toBe("s:synthetic")
    expect(continuation).toBe("s:synthetic")
  })
})

describe("probeSubtreeActivity", () => {
  it("fingerprints the child and its descendants and flags a pending question", async () => {
    const fake = treeClient({
      sessions: {
        root: { children: ["old", "new"], latest: message("m1", [{ id: "p1", type: "text", text: "hi" }]) },
        old: { created: 1 },
        new: {
          created: 2,
          latest: message("m2", [{ id: "q", type: "tool", tool: "question", state: { status: "running" } }]),
        },
      },
    })
    const sample = await probeSubtreeActivity(fake.client, "root", DIRECTORY, signal)
    expect(sample?.pendingQuestion).toBe(true)
    expect(sample?.fingerprint).toContain("root#2")
    // Newest children are probed first.
    expect(fake.probed).toEqual(["root", "new", "old"])
  })

  it("does not flag a finished question call", async () => {
    const fake = treeClient({
      sessions: { root: { latest: message("m", [{ type: "tool", tool: "question", state: { status: "completed" } }]) } },
    })
    expect((await probeSubtreeActivity(fake.client, "root", DIRECTORY, signal))?.pendingQuestion).toBe(false)
  })

  it("reaches the active grandchild under the newest of many finished children", async () => {
    const sessions: Record<string, FakeSession> = {}
    const ids = Array.from({ length: MAX_PROBE_SESSIONS + 8 }, (_unused, index) => `c${index}`)
    ids.forEach((id, index) => {
      sessions[id] = { created: index }
    })
    const newest = ids[ids.length - 1]
    sessions[newest] = { created: ids.length, children: ["grandchild"] }
    sessions.root = { children: ids }
    let text = "a"
    sessions.grandchild = { latest: message("g", [{ id: "p", type: "text", text }]) }
    const fake = treeClient({ sessions })
    const first = await probeSubtreeActivity(fake.client, "root", DIRECTORY, signal)
    expect(fake.probed.slice(0, 3)).toEqual(["root", newest, "grandchild"])
    text = "ab"
    sessions.grandchild = { latest: message("g", [{ id: "p", type: "text", text }]) }
    const second = await probeSubtreeActivity(fake.client, "root", DIRECTORY, signal)
    expect(second?.fingerprint).not.toBe(first?.fingerprint)
  })

  it("caps the number of probed sessions", async () => {
    const ids = Array.from({ length: MAX_PROBE_SESSIONS + 5 }, (_unused, index) => `c${index}`)
    const sessions: Record<string, FakeSession> = { root: { children: [...ids, "root"] } }
    const fake = treeClient({ sessions })
    await probeSubtreeActivity(fake.client, "root", DIRECTORY, signal)
    expect(fake.probed).toHaveLength(MAX_PROBE_SESSIONS)
  })

  it("returns null when endpoints are missing, fail, or return no data", async () => {
    expect(await probeSubtreeActivity({ session: {} } as unknown as Client, "r", DIRECTORY, signal)).toBeNull()
    expect(await probeSubtreeActivity(treeClient({ failMessages: true }).client, "r", DIRECTORY, signal)).toBeNull()
    expect(await probeSubtreeActivity(treeClient({ undefinedMessages: true }).client, "r", DIRECTORY, signal)).toBeNull()
    expect(await probeSubtreeActivity(treeClient({ undefinedChildren: true }).client, "r", DIRECTORY, signal)).toBeNull()
  })
})

// ---------------------------------------------------------------------------
// Compaction-marker classification (compaction-storm change D2, DT-FR-002)
// ---------------------------------------------------------------------------

const CONTINUATION_PROMPT =
  "Continue if you have next steps, or stop and ask for clarification if you are unsure how to proceed."

describe("classifySyntheticText", () => {
  it("classifies a system-reminder-wrapped summary", () => {
    expect(classifySyntheticText("<system-reminder>\n## Objective\n- keep working\n</system-reminder>\n")).toBe(
      "compaction-summary",
    )
    expect(classifySyntheticText("<SYSTEM-REMINDER>summary body</SYSTEM-REMINDER>")).toBe("compaction-summary")
  })

  it("classifies the continuation prompt with and without the overflow preamble", () => {
    expect(classifySyntheticText(CONTINUATION_PROMPT)).toBe("continuation-prompt")
    expect(
      classifySyntheticText(
        "The previous request exceeded the provider's size limit due to large media attachments. " +
          "The conversation was compacted and media files were removed from context.  " +
          CONTINUATION_PROMPT,
      ),
    ).toBe("continuation-prompt")
  })

  it("rejects prose that merely mentions compaction or reminders", () => {
    expect(classifySyntheticText("I compacted the conversation and kept working on the task.")).toBeNull()
    expect(classifySyntheticText("The system-reminder tags contain useful information, so I followed them.")).toBeNull()
    expect(classifySyntheticText("")).toBeNull()
    expect(classifySyntheticText("   ")).toBeNull()
  })

  it("rejects over-long and unanchored text", () => {
    expect(classifySyntheticText(`<system-reminder>${"x".repeat(MAX_SYNTHETIC_WRAPPER_LENGTH)}</system-reminder>`)).toBeNull()
    expect(classifySyntheticText("noise before <system-reminder>summary</system-reminder>")).toBeNull()
    expect(classifySyntheticText("<system-reminder>summary</system-reminder> trailing")).toBeNull()
    // A continuation-prompt match at exactly the documented bound still matches.
    const atBound = `${"y".repeat(MAX_CONTINUATION_PROMPT_LENGTH - CONTINUATION_PROMPT.length)}${CONTINUATION_PROMPT}`
    expect(atBound.length).toBe(MAX_CONTINUATION_PROMPT_LENGTH)
    expect(classifySyntheticText(atBound)).toBe("continuation-prompt")
    expect(classifySyntheticText(`${atBound} and one more word`)).toBeNull()
  })
})

describe("isCompactionMarker", () => {
  it("recognizes the opencode storage shapes", () => {
    expect(isCompactionMarker(message("m", [{ type: "compaction", auto: true }]))).toBe(true)
    expect(isCompactionMarker(message("m", [{ type: "text", text: "## Objective" }], { mode: "compaction" }))).toBe(true)
    expect(isCompactionMarker(message("m", [], { agent: "compaction" }))).toBe(true)
    expect(
      isCompactionMarker(
        message("m", [{ type: "text", metadata: { compaction_continue: true }, text: CONTINUATION_PROMPT }]),
      ),
    ).toBe(true)
  })

  it("recognizes user-role wrapped and continuation texts", () => {
    expect(
      isCompactionMarker(message("m", [{ type: "text", text: "<system-reminder>summary body</system-reminder>" }], { role: "user" })),
    ).toBe(true)
    expect(isCompactionMarker(message("m", [{ type: "text", text: CONTINUATION_PROMPT }], { role: "user" }))).toBe(true)
  })

  it("rejects normal messages, non-user wrapped text, and absent messages", () => {
    expect(isCompactionMarker(undefined)).toBe(false)
    expect(isCompactionMarker(message("m", [{ type: "text", text: "working" }]))).toBe(false)
    // Legitimate prose mentioning compaction stays legitimate (D3 negative control).
    expect(isCompactionMarker(message("m", [{ type: "text", text: "I will now compact the conversation." }], { role: "user" }))).toBe(false)
    // A wrapped shape on a non-user message is not a host injection.
    expect(
      isCompactionMarker(message("m", [{ type: "text", text: "<system-reminder>summary body</system-reminder>" }], { role: "assistant" })),
    ).toBe(false)
    // Non-flag metadata is ignored.
    expect(isCompactionMarker(message("m", [{ type: "text", metadata: "not-an-object", text: "working" }]))).toBe(false)
    expect(isCompactionMarker(message("m", [{ type: "text", metadata: { other: true }, text: "working" }]))).toBe(false)
  })
})

describe("isSubstantiveProgress", () => {
  it("counts a completed non-read tool call", () => {
    expect(isSubstantiveProgress(message("m", [{ type: "tool", tool: "write", state: { status: "completed" } }]))).toBe(true)
    expect(isSubstantiveProgress(message("m", [{ type: "tool", tool: "bash", state: { status: "completed" } }]))).toBe(true)
  })

  it("rejects read-only tools, running and failed tools, and non-tool messages", () => {
    for (const tool of READ_ONLY_TOOLS) {
      expect(isSubstantiveProgress(message("m", [{ type: "tool", tool, state: { status: "completed" } }]))).toBe(false)
    }
    expect(isSubstantiveProgress(message("m", [{ type: "tool", tool: "bash", state: { status: "running" } }]))).toBe(false)
    expect(isSubstantiveProgress(message("m", [{ type: "tool", tool: "bash", state: { status: "error" } }]))).toBe(false)
    expect(isSubstantiveProgress(message("m", [{ type: "text", text: "done" }]))).toBe(false)
    expect(isSubstantiveProgress(undefined)).toBe(false)
  })
})

describe("probeSubtreeActivity compaction accounting", () => {
  const roots: string[] = []
  /** Register a probe root id that afterEach cleans up. */
  const root = (id: string): string => {
    roots.push(id)
    return id
  }
  /** A compaction-marker message with the given message id. */
  const compactionMessage = (id: string) => message(id, [{ type: "compaction", auto: true }])

  afterEach(() => {
    for (const id of roots.splice(0)) {
      unregisterProbeState(id)
    }
  })

  it("excludes compaction markers from fingerprint progress", async () => {
    const probeRoot = root("r-fingerprint")
    const sessions: Record<string, FakeSession> = {
      [probeRoot]: { latest: message("m1", [{ id: "p1", type: "text", text: "working" }]) },
    }
    const fake = treeClient({ sessions })
    const first = await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    expect(first?.fingerprint).toContain(probeRoot)
    expect(first?.compactionsSinceProgress).toBe(0)

    // Neither a marker's arrival nor marker-to-marker churn changes the
    // fingerprint (DT-FR-002 compaction-only-child scenario).
    sessions[probeRoot] = { latest: compactionMessage("c1") }
    const second = await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    expect(second?.fingerprint).toBe(first?.fingerprint)
    expect(second?.compactionsSinceProgress).toBe(1)
    expect(second?.compactionCount).toBe(1)

    sessions[probeRoot] = { latest: compactionMessage("c2") }
    const third = await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    expect(third?.fingerprint).toBe(first?.fingerprint)
    expect(third?.compactionsSinceProgress).toBe(2)

    // A real message after the markers is progress again.
    sessions[probeRoot] = { latest: message("m2", [{ id: "p2", type: "text", text: "resumed" }]) }
    const fourth = await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    expect(fourth?.fingerprint).not.toBe(first?.fingerprint)
  })

  it("increments per new marker and resets only on substantive non-read progress", async () => {
    const probeRoot = root("r-reset")
    const sessions: Record<string, FakeSession> = { [probeRoot]: {} }
    const fake = treeClient({ sessions })
    const probe = () => probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)

    sessions[probeRoot] = { latest: compactionMessage("c1") }
    await probe()
    // Re-observing the same marker is not a new compaction.
    await probe()
    expect(registeredProbeCounters(probeRoot)).toEqual({
      compactionsSinceProgress: 1,
      nonReadProgressCount: 0,
      compactionCount: 1,
    })

    sessions[probeRoot] = { latest: compactionMessage("c2") }
    await probe()
    expect(registeredProbeCounters(probeRoot)?.compactionsSinceProgress).toBe(2)

    // A completed write resets the storm window and counts as progress.
    sessions[probeRoot] = { latest: message("w1", [{ id: "t1", type: "tool", tool: "write", state: { status: "completed" } }]) }
    await probe()
    expect(registeredProbeCounters(probeRoot)).toEqual({
      compactionsSinceProgress: 0,
      nonReadProgressCount: 1,
      compactionCount: 2,
    })

    // Read-only completions between compactions do not reset the window
    // (the storm fires through read-only activity; DELEG-FR-002).
    sessions[probeRoot] = { latest: message("r1", [{ id: "t2", type: "tool", tool: "read", state: { status: "completed" } }]) }
    await probe()
    sessions[probeRoot] = { latest: compactionMessage("c3") }
    await probe()
    expect(registeredProbeCounters(probeRoot)).toMatchObject({ compactionsSinceProgress: 1, compactionCount: 3 })
    expect(registeredProbeCounters(probeRoot)?.nonReadProgressCount).toBe(1)
  })

  it("keeps a quiet session's completed tool from masking a storm elsewhere", async () => {
    // quiet's stale completed write stays its latest message; without
    // identity-gated resets it would reset the window on every probe
    // and hide stormy's compaction loop.
    const probeRoot = root("r-masking")
    const sessions: Record<string, FakeSession> = {
      [probeRoot]: { children: ["quiet", "stormy"] },
      quiet: { created: 1, latest: message("w1", [{ id: "t1", type: "tool", tool: "write", state: { status: "completed" } }]) },
      stormy: { created: 2, latest: compactionMessage("c1") },
    }
    const fake = treeClient({ sessions })
    // The first walk observes stormy's marker and then quiet's write
    // (newest child first), so the write's reset lands last and the
    // window opens at zero -- but the write is counted exactly once.
    await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    expect(registeredProbeCounters(probeRoot)).toEqual({
      compactionsSinceProgress: 0,
      nonReadProgressCount: 1,
      compactionCount: 1,
    })
    // Each further compaction in stormy climbs without further resets:
    // quiet's write is the same identity, so it cannot mask the storm.
    for (const id of ["c2", "c3", "c4"]) {
      sessions.stormy = { created: 2, latest: compactionMessage(id) }
      await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    }
    expect(registeredProbeCounters(probeRoot)).toEqual({
      compactionsSinceProgress: 3,
      nonReadProgressCount: 1,
      compactionCount: 4,
    })
  })

  it("keeps accounting untouched when a probe fails mid-walk", async () => {
    const probeRoot = root("r-failed")
    const failFor: string[] = ["child"]
    const sessions: Record<string, FakeSession> = {
      [probeRoot]: { children: ["child"] },
      child: { created: 1, latest: compactionMessage("c1") },
    }
    const fake = treeClient({ sessions, failMessagesFor: failFor })
    expect(await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)).toBeNull()
    // The fold never ran: no state exists for the root.
    expect(registeredProbeCounters(probeRoot)).toBeUndefined()
    failFor.splice(0)
    const sample = await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    expect(sample?.compactionsSinceProgress).toBe(1)
  })

  it("unregisters a root's probe accounting", async () => {
    const probeRoot = root("r-unregister")
    const sessions: Record<string, FakeSession> = { [probeRoot]: { latest: compactionMessage("c1") } }
    const fake = treeClient({ sessions })
    await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    expect(registeredProbeCounters(probeRoot)?.compactionCount).toBe(1)
    unregisterProbeState(probeRoot)
    expect(registeredProbeCounters(probeRoot)).toBeUndefined()
  })
})

// ---------------------------------------------------------------------------
// Executor integration (design D1/D2/D4)
// ---------------------------------------------------------------------------

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

interface ExecutorFakeOptions {
  /** Returns the child's latest message on each probe (undefined = probe failure). */
  readonly latest?: () => Record<string, unknown> | undefined
  /** Delay before the Nth (1-based) probe answers, in ms (default 0). */
  readonly probeDelayMs?: (call: number) => number
  /** Child prompt resolves after this many ms instead of hanging. */
  readonly resolveAfterMs?: number
  /** Child prompt rejects this many ms after the abort signal fires (default immediate). */
  readonly rejectAfterAbortMs?: number
}

/** Executor-facing fake: child creation, a hanging prompt, abort, and probe endpoints. */
function executorClient(options: ExecutorFakeOptions = {}) {
  const calls = { creates: 0, aborts: 0, probes: 0, deadlineDuringRun: undefined as number | undefined }
  const session = {
    get: async () => ({ data: { id: "parent" } }),
    create: async () => {
      calls.creates++
      return { data: { id: "exec-child" } }
    },
    prompt: async (request: { signal?: AbortSignal }) => {
      calls.deadlineDuringRun = registeredDeadline("exec-child")
      return await new Promise((resolve, reject) => {
        if (options.resolveAfterMs !== undefined) {
          setTimeout(
            () =>
              resolve({
                data: {
                  info: { providerID: "p", modelID: "m" },
                  parts: [{ type: "text", text: "done" }],
                },
              }),
            options.resolveAfterMs,
          )
        }
        request.signal?.addEventListener(
          "abort",
          () => {
            const reason = request.signal?.reason
            // A deferred rejection lets a late wall-clock timer fire
            // before the prompt's rejection settles, exercising the
            // abort-cause classification race.
            if (options.rejectAfterAbortMs !== undefined) {
              setTimeout(() => reject(reason), options.rejectAfterAbortMs)
            } else {
              reject(reason)
            }
          },
          { once: true },
        )
      })
    },
    abort: async () => {
      calls.aborts++
      return { data: true }
    },
    children: async () => ({ data: [] }),
    messages: async () => {
      calls.probes++
      const delay = options.probeDelayMs?.(calls.probes) ?? 0
      if (delay > 0) {
        await new Promise((resolve) => setTimeout(resolve, delay))
      }
      const latest = options.latest?.()
      if (latest === undefined) {
        throw new Error("probe failed")
      }
      return { data: [latest] }
    },
  }
  return { client: { session } as unknown as Client, calls }
}

const POLL_MS = 1_000

function run(
  client: Client,
  extra: { timeoutMilliseconds?: number; idleTimeoutMilliseconds?: number; parentAbort?: AbortSignal } = {},
) {
  return executeAgentSession(
    {
      client,
      sessionID: "parent",
      directory: DIRECTORY,
      parentAbort: extra.parentAbort ?? new AbortController().signal,
      timeoutMilliseconds: extra.timeoutMilliseconds ?? 60_000,
      idleTimeoutMilliseconds: extra.idleTimeoutMilliseconds,
      idlePollMilliseconds: POLL_MS,
    },
    "cobalt-crush-dev",
    "prompt",
    { providerID: "p", modelID: "m", combined: "p/m" },
    undefined,
    false,
    PROVENANCE,
  )
}

describe("executeAgentSession budgets and idle watchdog", () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })
  afterEach(() => {
    vi.useRealTimers()
  })

  it("fails fast with budget_exhausted when the parent budget is spent", async () => {
    register("parent", Date.now() + BUDGET_RESERVE_MS + 1_000)
    const fake = executorClient()
    const result = await run(fake.client)
    expect(result.error?.code).toBe("budget_exhausted")
    expect(result.error?.retryable).toBe(false)
    expect(result.timeout?.kind).toBe("budget")
    expect(result.timeout?.budget_clamped).toBe(true)
    expect(result.timeout?.child_active).toBeNull()
    expect(fake.calls.creates).toBe(0)
  })

  it("clamps a nested run to the parent budget and registers the child deadline", async () => {
    const start = Date.now()
    register("parent", start + BUDGET_RESERVE_MS + MIN_NESTED_BUDGET_MS)
    const controller = new AbortController()
    const fake = executorClient()
    const pending = run(fake.client, { timeoutMilliseconds: 1_800_000, parentAbort: controller.signal })
    await vi.advanceTimersByTimeAsync(0)
    // The child inherits the clamped absolute deadline, not the request.
    expect(fake.calls.deadlineDuringRun).toBe(start + MIN_NESTED_BUDGET_MS)
    controller.abort(new Error("stop"))
    const result = await pending
    expect(result.error?.code).toBe("cancelled")
    expect(result.timeout).toBeUndefined()
    expect(registeredDeadline("exec-child")).toBeUndefined()
  })

  it("reports a plain wall-clock timeout with no probe data", async () => {
    const pending = run(executorClient().client, { timeoutMilliseconds: 5_000 })
    await vi.advanceTimersByTimeAsync(5_000)
    const result = await pending
    expect(result.error?.code).toBe("timeout")
    expect(result.timeout).toMatchObject({
      kind: "wall_clock",
      effective_ms: 5_000,
      requested_ms: 5_000,
      budget_clamped: false,
      idle_timeout_ms: null,
      elapsed_ms: 5_000,
      last_progress_ms_ago: null,
      child_active: null,
      pending_question: false,
    })
  })

  it("aborts a stalled child with idle_timeout after an unchanged probe", async () => {
    const fake = executorClient({ latest: () => message("m", [{ id: "p", type: "text", text: "same" }]) })
    const pending = run(fake.client, { idleTimeoutMilliseconds: 10_000 })
    await vi.advanceTimersByTimeAsync(15_000)
    const result = await pending
    expect(result.error?.code).toBe("idle_timeout")
    expect(result.error?.retryable).toBe(true)
    expect(result.error?.message).toBe("invoke_agent run made no progress for 10000 ms")
    expect(result.timeout).toMatchObject({ kind: "idle", child_active: false, idle_timeout_ms: 10_000 })
    // First probe at 1 s records the baseline; the unchanged probe at 11 s ends the run.
    expect(result.timeout?.elapsed_ms).toBe(11_000)
    expect(result.childSessionID).toBe("exec-child")
    expect(fake.calls.aborts).toBe(1)
    // The child's deadline is removed once the run ends.
    expect(registeredDeadline("exec-child")).toBeUndefined()
  })

  it("keeps a child with streaming text alive and reports it active on wall-clock timeout", async () => {
    let counter = 0
    const fake = executorClient({
      latest: () => message("m", [{ id: "p", type: "text", text: "x".repeat(++counter) }]),
    })
    const pending = run(fake.client, { timeoutMilliseconds: 20_000, idleTimeoutMilliseconds: 5_000 })
    await vi.advanceTimersByTimeAsync(20_000)
    const result = await pending
    expect(result.error?.code).toBe("timeout")
    expect(result.timeout?.kind).toBe("wall_clock")
    expect(result.timeout?.child_active).toBe(true)
    expect(result.timeout?.last_progress_ms_ago).toBe(1_000)
  })

  it("counts a running tool's streamed metadata output as progress", async () => {
    // Bash keeps live output in state.metadata.output; the running tool
    // state carries no top-level output field.
    let output = ""
    const fake = executorClient({
      latest: () => {
        output += "ok  pkg\n"
        return message("m", [
          { id: "bash", type: "tool", tool: "bash", state: { status: "running", title: "go test", metadata: { output } } },
        ])
      },
    })
    const pending = run(fake.client, { timeoutMilliseconds: 20_000, idleTimeoutMilliseconds: 5_000 })
    await vi.advanceTimersByTimeAsync(20_000)
    const result = await pending
    expect(result.error?.code).toBe("timeout")
    expect(result.timeout?.child_active).toBe(true)
  })

  it("never idle-aborts when every probe fails", async () => {
    const pending = run(executorClient({ latest: () => undefined }).client, {
      timeoutMilliseconds: 10_000,
      idleTimeoutMilliseconds: 2_000,
    })
    await vi.advanceTimersByTimeAsync(10_000)
    const result = await pending
    expect(result.error?.code).toBe("timeout")
    expect(result.timeout?.child_active).toBeNull()
  })

  it("returns success without timeout details when the child finishes", async () => {
    const fake = executorClient({
      resolveAfterMs: 5_000,
      latest: () => message("m", [{ id: "p", type: "text", text: "same" }]),
    })
    const pending = run(fake.client, { idleTimeoutMilliseconds: 60_000 })
    await vi.advanceTimersByTimeAsync(5_000)
    const result = await pending
    expect(result.status).toBe("success")
    expect(result.timeout).toBeUndefined()
  })

  it("ignores a probe that answers after the child finished", async () => {
    // Probes at 1 s and 2 s answer at once; the 3 s probe answers at 5 s,
    // after the prompt resolved at 3.5 s, unchanged and past the idle window.
    const fake = executorClient({
      resolveAfterMs: 3_500,
      probeDelayMs: (call) => (call === 3 ? 2_000 : 0),
      latest: () => message("m", [{ id: "p", type: "text", text: "same" }]),
    })
    const pending = run(fake.client, { idleTimeoutMilliseconds: 2_000 })
    await vi.advanceTimersByTimeAsync(3_500)
    const result = await pending
    await vi.advanceTimersByTimeAsync(5_000)
    expect(result.status).toBe("success")
    expect(result.timeout).toBeUndefined()
    expect(fake.calls.probes).toBe(3)
    expect(fake.calls.aborts).toBe(0)
  })

  it("fires the budget-clamped wall clock at the effective deadline", async () => {
    // Regression guard for #930: the wall clock must use the clamped
    // effective timeout. A regression that passes the requested
    // timeout to setTimeout would not fire at the clamp deadline.
    const start = Date.now()
    register("parent", start + BUDGET_RESERVE_MS + MIN_NESTED_BUDGET_MS)
    const pending = run(executorClient().client, { timeoutMilliseconds: 1_800_000 })
    await vi.advanceTimersByTimeAsync(MIN_NESTED_BUDGET_MS)
    const result = await pending
    expect(result.error?.code).toBe("timeout")
    expect(result.timeout).toMatchObject({
      kind: "wall_clock",
      effective_ms: MIN_NESTED_BUDGET_MS,
      requested_ms: 1_800_000,
      budget_clamped: true,
    })
  })

  it("does not reset the idle clock on failed probes", async () => {
    // Probe 1 at 1 s succeeds ("a", baseline and last progress);
    // probes 2-6 fail; probe 7 at 7 s succeeds unchanged, 6 s after
    // the last real progress. If failures reset the idle clock the
    // run would reach the wall clock instead of aborting.
    let call = 0
    const fake = executorClient({
      latest: () => {
        call++
        return call === 1 || call === 7
          ? message("m", [{ id: "p", type: "text", text: "a" }])
          : undefined
      },
    })
    const pending = run(fake.client, { timeoutMilliseconds: 60_000, idleTimeoutMilliseconds: 5_000 })
    await vi.advanceTimersByTimeAsync(7_000)
    const result = await pending
    expect(result.error?.code).toBe("idle_timeout")
    expect(result.timeout).toMatchObject({ kind: "idle", elapsed_ms: 7_000 })
    expect(result.timeout?.last_progress_ms_ago).toBe(6_000)
  })

  it("keeps the idle classification when the wall-clock timer fires before a deferred rejection", async () => {
    // The idle watchdog aborts at 11 s (the probe at 1 s is the last
    // progress; the 11 s probe is unchanged, 10 s later). The child's
    // prompt rejects 5 s after the abort, so the wall-clock timer at
    // 12 s fires while the rejection is still pending. Without the
    // first-cause-wins guard the late timer overwrites abortCause and
    // the run is misclassified as wall_clock.
    const fake = executorClient({
      rejectAfterAbortMs: 5_000,
      latest: () => message("m", [{ id: "p", type: "text", text: "same" }]),
    })
    const pending = run(fake.client, { timeoutMilliseconds: 12_000, idleTimeoutMilliseconds: 10_000 })
    await vi.advanceTimersByTimeAsync(20_000)
    const result = await pending
    expect(result.error?.code).toBe("idle_timeout")
    expect(result.timeout?.kind).toBe("idle")
  })

  it("reports child_active false on a wall-clock timeout with stale progress", async () => {
    // The only progress is the first probe at 1 s; 129 s later the
    // 120 s activity window has lapsed, so the wall-clock timeout
    // reports the child as inactive.
    const fake = executorClient({ latest: () => message("m", [{ id: "p", type: "text", text: "same" }]) })
    const pending = run(fake.client, { timeoutMilliseconds: 130_000, idleTimeoutMilliseconds: 300_000 })
    await vi.advanceTimersByTimeAsync(130_000)
    const result = await pending
    expect(result.error?.code).toBe("timeout")
    expect(result.timeout).toMatchObject({ kind: "wall_clock", child_active: false })
    expect(result.timeout?.last_progress_ms_ago).toBe(129_000)
  })
})
