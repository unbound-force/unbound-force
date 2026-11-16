import type { PluginInput } from "@opencode-ai/plugin"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"

import { executeAgentSession, type InvocationProvenance } from "../lib/agent-executor.js"
import {
  BUDGET_RESERVE_MS,
  IDLE_DEFERRAL_CAP_MS,
  IDLE_TOOL_GRACE_MS,
  MAX_ANCESTRY_DEPTH,
  MAX_CONTINUATION_PROMPT_LENGTH,
  MAX_PROBE_SESSIONS,
  MAX_SYNTHETIC_WRAPPER_LENGTH,
  MIN_NESTED_BUDGET_MS,
  READ_ONLY_TOOLS,
  clampToBudget,
  classifySyntheticText,
  collectRunningTools,
  idleDeferralExpired,
  isCompactionMarker,
  isSubstantiveProgress,
  messageFingerprint,
  probeSubtreeActivity,
  registerDeadline,
  registeredDeadline,
  registeredProbeCounters,
  resolveInheritedDeadline,
  selectRunningTool,
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
    ).toBe("s:m:::1:t:tool:running::0~t:running:")
  })

  it("folds the sorted running triples into the hash (DT-FR-013)", () => {
    // A quiet running bash -- constant metadata, no new messages -- has a
    // stable fingerprint; the running set itself is part of the hash, so
    // set changes still read as progress.
    const quiet = () =>
      messageFingerprint("s", message("m", [{ id: "p2", type: "tool", tool: "bash", state: { status: "running", metadata: { output: "same" } } }]))
    expect(messageFingerprint("s", quiet())).toBe(messageFingerprint("s", quiet()))
    // A tool swap on the same partId changes the triple set.
    const swap = messageFingerprint("s", message("m", [{ id: "p2", type: "tool", tool: "read", state: { status: "running" } }]))
    expect(messageFingerprint("s", quiet())).not.toBe(swap)
    // A new bash part (the old one completed, partId changed) is progress.
    const next = messageFingerprint("s", message("m", [{ id: "p3", type: "tool", tool: "bash", state: { status: "running" } }]))
    expect(messageFingerprint("s", quiet())).not.toBe(next)
    // Completing the part drops the suffix, which is also a change.
    const done = messageFingerprint("s", message("m", [{ id: "p2", type: "tool", tool: "bash", state: { status: "completed" } }]))
    expect(messageFingerprint("s", quiet())).not.toBe(done)
  })

  it("sorts multiple running parts by partId and omits the suffix when none run", () => {
    const two = messageFingerprint(
      "s",
      message("m", [
        { id: "done", type: "tool", tool: "bash", state: { status: "completed" } },
        { id: "b2", type: "tool", tool: "bash", state: { status: "running" } },
        { id: "a1", type: "tool", tool: "read", state: { status: "running" } },
      ]),
    )
    expect(two).toContain("~a1:running:read,b2:running:bash")
    expect(two).not.toContain("done")
    expect(messageFingerprint("s", message("m", [{ id: "done", type: "tool", tool: "bash", state: { status: "completed" } }]))).not.toContain("~")
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

  it("keeps running parts out of marker fingerprints (DT-FR-002 + DT-FR-013)", () => {
    // A marker message carrying a running part still collapses to the
    // synthetic constant: markers never count as fingerprint progress,
    // whatever parts ride them.
    const marker = message("m", [{ id: "p1", type: "tool", tool: "bash", state: { status: "running" } }], { mode: "compaction" })
    expect(messageFingerprint("s", marker)).toBe("s:synthetic")
  })
})

describe("collectRunningTools", () => {
  it("collects running parts sorted by partId with tool default and deadline contingency", () => {
    const parts = [
      { id: "b", type: "tool", tool: "bash", state: { status: "running", metadata: { deadline_ms: 123 } } },
      { id: "a", type: "tool", state: { status: "running", metadata: { deadline_ms: "not-a-number" } } },
      { id: "c", type: "tool", tool: "read", state: { status: "completed" } },
      { id: "d", type: "text", text: "idle" },
    ]
    expect(collectRunningTools(message("m", parts))).toEqual([
      { partId: "a", status: "running", tool: "", deadlineMs: null },
      { partId: "b", status: "running", tool: "bash", deadlineMs: 123 },
    ])
    expect(collectRunningTools(undefined)).toEqual([])
  })
})

describe("selectRunningTool (DT-FR-015)", () => {
  const part = (partId: string, tool: string) => ({ partId, status: "running" as const, tool, deadlineMs: null })

  it("returns null when no part is running", () => {
    expect(selectRunningTool([])).toBeNull()
  })

  it("selects the running part with the lexicographically smallest partId", () => {
    expect(selectRunningTool([part("z9", "bash"), part("a1", "gaze"), part("m5", "read")])).toBe("gaze")
  })

  it("breaks partId ties deterministically by tool name", () => {
    expect(selectRunningTool([part("p1", "read"), part("p1", "bash")])).toBe("bash")
  })

  it("reports null when the selected name fails validation, without falling through", () => {
    // The smallest-partId part's name is adversarial model input: it
    // must report null rather than skipping to the next part, and the
    // raw value never leaves this call to be logged.
    expect(selectRunningTool([part("z9", "bash"), part("a1", "bash; rm -rf /")])).toBeNull()
    expect(selectRunningTool([part("a1", "tool with spaces")])).toBeNull()
    expect(selectRunningTool([part("a1", "")])).toBeNull()
    expect(selectRunningTool([part("a1", `x`.repeat(129))])).toBeNull()
    expect(selectRunningTool([part("a1", "mcp__server.tool:v2")])).toBe("mcp__server.tool:v2")
  })

  it("enforces the DT-FR-015 name length boundary", () => {
    // ^[A-Za-z0-9._:-]{1,128}$: a one-char and a 128-char name pass at the
    // boundaries; a 129-char name and the empty string map to null.
    expect(selectRunningTool([part("a1", "b")])).toBe("b")
    expect(selectRunningTool([part("a1", "x".repeat(128))])).toBe("x".repeat(128))
    expect(selectRunningTool([part("a1", "x".repeat(129))])).toBeNull()
    expect(selectRunningTool([part("a1", "")])).toBeNull()
  })
})

describe("idleDeferralExpired (DT-FR-014)", () => {
  const part = (partId: string, deadlineMs: number | null) => ({ partId, status: "running" as const, tool: "bash", deadlineMs })
  const anchor = 1_000
  const idle = 600_000

  it("treats the deferral as vacuous with no running parts", () => {
    expect(idleDeferralExpired([], idle, anchor, anchor + 10 * idle)).toBe(true)
  })

  it("never reports expiry while the idle watchdog is disabled", () => {
    expect(idleDeferralExpired([part("p", null)], undefined, anchor, anchor + 10 * idle)).toBe(false)
  })

  it("holds until idle_timeout plus the cap since the last fingerprint change", () => {
    const bound = anchor + idle + IDLE_DEFERRAL_CAP_MS
    expect(idleDeferralExpired([part("p", null)], idle, anchor, bound - 1)).toBe(false)
    expect(idleDeferralExpired([part("p", null)], idle, anchor, bound)).toBe(true)
  })

  it("holds until the earliest observable deadline plus grace", () => {
    const deadline = anchor + 720_000
    expect(idleDeferralExpired([part("p", deadline)], idle, anchor, deadline + IDLE_TOOL_GRACE_MS - 1)).toBe(false)
    expect(idleDeferralExpired([part("p", deadline)], idle, anchor, deadline + IDLE_TOOL_GRACE_MS)).toBe(true)
  })

  it("covers the earliest of multiple deadlines", () => {
    const early = anchor + 60_000
    const late = anchor + 3_600_000
    expect(idleDeferralExpired([part("late", late), part("early", early)], idle, anchor, early + IDLE_TOOL_GRACE_MS)).toBe(true)
    expect(idleDeferralExpired([part("late", late)], idle, anchor, early + IDLE_TOOL_GRACE_MS)).toBe(false)
  })

  it("uses the deadline path when any running part exposes a deadline", () => {
    // A mixed set (one deadline-bearing, one not) follows the earliest
    // observable deadline, per DT-FR-014, not the cap.
    const deadline = anchor + 3_600_000
    expect(idleDeferralExpired([part("no-deadline", null), part("deadline", deadline)], idle, anchor, anchor + idle + IDLE_DEFERRAL_CAP_MS)).toBe(false)
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

describe("probeSubtreeActivity running tools (DT-FR-013)", () => {
  const roots: string[] = []
  /** Register a probe root id that afterEach cleans up. */
  const root = (id: string): string => {
    roots.push(id)
    return id
  }

  afterEach(() => {
    for (const id of roots.splice(0)) {
      unregisterProbeState(id)
    }
  })

  it("carries a quiet running bash on the sample with a stable fingerprint", async () => {
    const probeRoot = root("r-running")
    const sessions: Record<string, FakeSession> = {
      [probeRoot]: { children: ["kid"] },
      kid: {
        created: 1,
        latest: message("m", [{ id: "p1", type: "tool", tool: "bash", state: { status: "running", metadata: { output: "steady" } } }]),
      },
    }
    const fake = treeClient({ sessions })
    const first = await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    const second = await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    // Quiet bash, no new messages: the fingerprint is stable (the old
    // behavior that misread this as idle), but the running part rides
    // the sample so the watcher can defer within the DT-FR-014 bound.
    expect(second?.fingerprint).toBe(first?.fingerprint)
    expect(first?.runningTools).toEqual([{ partId: "p1", status: "running", tool: "bash", deadlineMs: null }])
    expect(second?.runningTools).toEqual(first?.runningTools)
  })

  it("unions and sorts running parts across sessions by partId", async () => {
    const probeRoot = root("r-union")
    const sessions: Record<string, FakeSession> = {
      [probeRoot]: { children: ["zKid", "aKid"] },
      zKid: { created: 1, latest: message("mz", [{ id: "z1", type: "tool", tool: "bash", state: { status: "running" } }]) },
      aKid: { created: 2, latest: message("ma", [{ id: "a1", type: "tool", tool: "gaze", state: { status: "running" } }]) },
    }
    const fake = treeClient({ sessions })
    const sample = await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    expect(sample?.runningTools.map((part) => part.partId)).toEqual(["a1", "z1"])
  })

  it("collects running parts from a marker latest message without fingerprint progress", async () => {
    const probeRoot = root("r-marker-running")
    const sessions: Record<string, FakeSession> = {
      [probeRoot]: { latest: message("m1", [{ id: "p1", type: "text", text: "working" }]) },
    }
    const fake = treeClient({ sessions })
    const first = await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    // A marker arrives carrying a running part: the print stays the
    // remembered pre-marker print (no fingerprint progress, DT-FR-002)
    // while the running part still rides the sample because DT-FR-014's
    // deferral keys off any running tool in the subtree.
    sessions[probeRoot] = {
      latest: message("c1", [{ id: "r1", type: "tool", tool: "bash", state: { status: "running" } }], { mode: "compaction" }),
    }
    const second = await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    expect(second?.fingerprint).toBe(first?.fingerprint)
    expect(second?.runningTools).toEqual([{ partId: "r1", status: "running", tool: "bash", deadlineMs: null }])
  })

  it("combines a synthetic marker session with a running part on another session (DT-FR-002 + DT-FR-013)", async () => {
    const probeRoot = root("r-marker-plus-running")
    const sessions: Record<string, FakeSession> = {
      [probeRoot]: { children: ["markerKid", "runKid"] },
      markerKid: { created: 1, latest: message("m1", [{ id: "p1", type: "text", text: "working" }]) },
      runKid: { created: 2, latest: message("m2", [{ id: "b1", type: "tool", tool: "bash", state: { status: "running" } }]) },
    }
    const fake = treeClient({ sessions })
    const first = await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    // runKid's running triple rides its print inside the combined
    // fingerprint while markerKid still contributes a real print.
    expect(first?.fingerprint).toContain("~b1:running:bash")
    expect(first?.runningTools).toEqual([{ partId: "b1", status: "running", tool: "bash", deadlineMs: null }])
    // A marker arrives on markerKid: the synthetic stays excluded from the
    // fingerprint (the remembered pre-marker print substitutes in) while
    // the running triple from runKid stays included -- both rules hold in
    // one sample, not just in isolation.
    sessions.markerKid = { created: 1, latest: message("c1", [{ type: "compaction", auto: true }]) }
    const second = await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    expect(second?.fingerprint).toBe(first?.fingerprint)
    expect(second?.runningTools).toEqual(first?.runningTools)
    expect(second?.maxSessionCompactionsSinceProgress).toBe(1)
    expect(second?.stormSessionID).toBe("markerKid")
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

// ---------------------------------------------------------------------------
// Probe compaction accounting, per session
// (compaction-storm-detector-sums-compactio design D1/D2; DT-FR-002 /
// DELEG-FR-002, issue #1125)
// ---------------------------------------------------------------------------

describe("probeSubtreeActivity compaction accounting", () => {
  const roots: string[] = []
  /** Register a probe root id that afterEach cleans up. */
  const root = (id: string): string => {
    roots.push(id)
    return id
  }
  /** A compaction-marker message with the given message id. */
  const compactionMessage = (id: string) => message(id, [{ type: "compaction", auto: true }])
  /** A completed read-only tool call: activity that never resets a session's window. */
  const readMessage = (id: string) => message(id, [{ id: `t-${id}`, type: "tool", tool: "read", state: { status: "completed" } }])

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
    expect(first?.maxSessionCompactionsSinceProgress).toBe(0)
    expect(first?.stormSessionID).toBeNull()

    // Neither a marker's arrival nor marker-to-marker churn changes the
    // fingerprint (DT-FR-002 compaction-only-child scenario).
    sessions[probeRoot] = { latest: compactionMessage("c1") }
    const second = await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    expect(second?.fingerprint).toBe(first?.fingerprint)
    expect(second?.maxSessionCompactionsSinceProgress).toBe(1)
    expect(second?.stormSessionID).toBe(probeRoot)
    expect(second?.compactionCount).toBe(1)

    sessions[probeRoot] = { latest: compactionMessage("c2") }
    const third = await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    expect(third?.fingerprint).toBe(first?.fingerprint)
    expect(third?.maxSessionCompactionsSinceProgress).toBe(2)
    expect(third?.stormSessionID).toBe(probeRoot)

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
      maxSessionCompactionsSinceProgress: 1,
      stormSessionID: probeRoot,
      nonReadProgressCount: 0,
      compactionCount: 1,
    })

    sessions[probeRoot] = { latest: compactionMessage("c2") }
    await probe()
    expect(registeredProbeCounters(probeRoot)).toMatchObject({
      maxSessionCompactionsSinceProgress: 2,
      stormSessionID: probeRoot,
    })

    // A completed write resets the storm window and counts as progress.
    sessions[probeRoot] = { latest: message("w1", [{ id: "t1", type: "tool", tool: "write", state: { status: "completed" } }]) }
    await probe()
    expect(registeredProbeCounters(probeRoot)).toEqual({
      maxSessionCompactionsSinceProgress: 0,
      stormSessionID: null,
      nonReadProgressCount: 1,
      compactionCount: 2,
    })

    // Read-only completions between compactions do not reset the window
    // (the storm fires through read-only activity; DELEG-FR-002).
    sessions[probeRoot] = { latest: readMessage("r1") }
    await probe()
    sessions[probeRoot] = { latest: compactionMessage("c3") }
    await probe()
    expect(registeredProbeCounters(probeRoot)).toMatchObject({
      maxSessionCompactionsSinceProgress: 1,
      stormSessionID: probeRoot,
      compactionCount: 3,
    })
    expect(registeredProbeCounters(probeRoot)?.nonReadProgressCount).toBe(1)
  })

  it("keeps a quiet session's completed tool from masking a storm elsewhere", async () => {
    // quiet's stale completed write stays its latest message; without
    // identity-gated resets it would count as fresh progress on every
    // probe. Per-session accounting (issue #1125) makes the masking
    // impossible outright: a reset lands only in the session that
    // produced it, so stormy's marker survives quiet's write even
    // though the same walk observes the write after the marker.
    const probeRoot = root("r-masking")
    const sessions: Record<string, FakeSession> = {
      [probeRoot]: { children: ["quiet", "stormy"] },
      quiet: { created: 1, latest: message("w1", [{ id: "t1", type: "tool", tool: "write", state: { status: "completed" } }]) },
      stormy: { created: 2, latest: compactionMessage("c1") },
    }
    const fake = treeClient({ sessions })
    // The first walk observes stormy's marker and then quiet's write
    // (newest child first): the write resets only quiet's window, the
    // marker keeps stormy's run open, and the write is counted once.
    await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    expect(registeredProbeCounters(probeRoot)).toEqual({
      maxSessionCompactionsSinceProgress: 1,
      stormSessionID: "stormy",
      nonReadProgressCount: 1,
      compactionCount: 1,
    })
    // Each further compaction in stormy climbs without further resets:
    // quiet's write is the same identity, and another session's
    // progress could not reset stormy's window anyway. One session
    // with N consecutive compactions reaches max N, and the signal
    // names it (issue #1125 acceptance criterion 2).
    for (const id of ["c2", "c3", "c4"]) {
      sessions.stormy = { created: 2, latest: compactionMessage(id) }
      await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    }
    expect(registeredProbeCounters(probeRoot)).toEqual({
      maxSessionCompactionsSinceProgress: 4,
      stormSessionID: "stormy",
      nonReadProgressCount: 1,
      compactionCount: 4,
    })
  })

  it("does not sum compactions across sibling sessions", async () => {
    // Issue #1125 acceptance criterion 1: three reviewer-like siblings,
    // one compaction each, only read-only calls in between. The removed
    // subtree-wide sum reached the default threshold of 3 and aborted
    // healthy dispatches; the per-session max stays at one, so the
    // storm signal never reaches any threshold (the normalized range
    // starts at 2) and the executor's detector cannot fire.
    const probeRoot = root("r-siblings")
    const sessions: Record<string, FakeSession> = {
      [probeRoot]: { children: ["slow", "middle", "fast"] },
      slow: { created: 1, latest: readMessage("slow-read-1") },
      middle: { created: 2, latest: readMessage("mid-read-1") },
      fast: { created: 3, latest: readMessage("fast-read-1") },
    }
    const fake = treeClient({ sessions })
    const probe = () => probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)

    expect((await probe())?.maxSessionCompactionsSinceProgress).toBe(0)
    sessions.fast = { created: 3, latest: compactionMessage("fast-c1") }
    expect((await probe())?.maxSessionCompactionsSinceProgress).toBe(1)
    // Reads between the markers never reset a session's window...
    sessions.fast = { created: 3, latest: readMessage("fast-read-2") }
    sessions.middle = { created: 2, latest: compactionMessage("mid-c1") }
    expect((await probe())?.maxSessionCompactionsSinceProgress).toBe(1)
    sessions.middle = { created: 2, latest: readMessage("mid-read-2") }
    sessions.slow = { created: 1, latest: compactionMessage("slow-c1") }
    expect((await probe())?.maxSessionCompactionsSinceProgress).toBe(1)
    // ...and the markers never sum: three sessions at one compaction
    // each are a maximum of one, not a total of three. The newest-first
    // walk observes fast before its siblings, so the tie names fast.
    expect(registeredProbeCounters(probeRoot)).toEqual({
      maxSessionCompactionsSinceProgress: 1,
      stormSessionID: "fast",
      nonReadProgressCount: 0,
      compactionCount: 3,
    })
  })

  it("resets only the session that produced substantive progress", async () => {
    // Issue #1125: a reset is scoped to the session that earned it.
    // looper's completed write clears looper's window while sibling's
    // untouched run survives -- under the old shared counter one
    // session's progress masked another session's storm.
    const probeRoot = root("r-reset-scope")
    const sessions: Record<string, FakeSession> = {
      [probeRoot]: { children: ["sibling", "looper"] },
      sibling: { created: 1, latest: compactionMessage("s-c1") },
      looper: { created: 2, latest: compactionMessage("l-c1") },
    }
    const fake = treeClient({ sessions })
    const probe = () => probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)

    await probe()
    expect(registeredProbeCounters(probeRoot)).toEqual({
      maxSessionCompactionsSinceProgress: 1,
      stormSessionID: "looper",
      nonReadProgressCount: 0,
      compactionCount: 2,
    })
    sessions.looper = { created: 2, latest: compactionMessage("l-c2") }
    await probe()
    // Both sessions hold a run; looper's is the deeper one.
    expect(registeredProbeCounters(probeRoot)).toEqual({
      maxSessionCompactionsSinceProgress: 2,
      stormSessionID: "looper",
      nonReadProgressCount: 0,
      compactionCount: 3,
    })

    // Substantive progress in looper resets only looper: the maximum
    // falls to sibling's surviving run of one.
    sessions.looper = { created: 2, latest: message("w1", [{ id: "t1", type: "tool", tool: "write", state: { status: "completed" } }]) }
    await probe()
    expect(registeredProbeCounters(probeRoot)).toEqual({
      maxSessionCompactionsSinceProgress: 1,
      stormSessionID: "sibling",
      nonReadProgressCount: 1,
      compactionCount: 3,
    })
  })

  it("excludes an executor-managed descendant from the max while totaling its markers", async () => {
    // Design D2 / DELEG-FR-002: a descendant another executor
    // registered a run deadline for carries its own always-armed
    // storm detector, so its markers never accrue toward this probe's
    // per-session storm signal -- while the subtree-wide
    // observability total still counts every one of them.
    const probeRoot = root("r-managed")
    const sessions: Record<string, FakeSession> = {
      [probeRoot]: { children: ["managed", "unmanaged"] },
      managed: { created: 1 },
      unmanaged: { created: 2, latest: compactionMessage("u-c1") },
    }
    register("managed", Date.now() + 60_000)
    const fake = treeClient({ sessions })
    const probe = () => probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)

    // The managed descendant storms past the default threshold of
    // three; only the unmanaged sibling's single marker moves the max.
    for (const id of ["m-c1", "m-c2", "m-c3", "m-c4"]) {
      sessions.managed = { created: 1, latest: compactionMessage(id) }
      await probe()
    }
    expect(registeredProbeCounters(probeRoot)).toEqual({
      maxSessionCompactionsSinceProgress: 1,
      stormSessionID: "unmanaged",
      nonReadProgressCount: 0,
      compactionCount: 5,
    })
  })

  it("never counts a managed marker retroactively after its run's deadline drops", async () => {
    // Design D2: marker keys are recorded for every observed session,
    // managed or not. A finished reviewer's deadline is unregistered
    // when its run ends while its last marker stays its latest
    // message; the seen key keeps that marker out of the count even
    // though the session is no longer recognized as executor-managed.
    // A marker first observed with no deadline visible counts per
    // session as an unrecognized descendant, failing toward detection.
    const probeRoot = root("r-retroactive")
    const sessions: Record<string, FakeSession> = {
      [probeRoot]: { children: ["reviewer"] },
      reviewer: { created: 1, latest: compactionMessage("rv-c1") },
    }
    register("reviewer", Date.now() + 60_000)
    const fake = treeClient({ sessions })
    const probe = () => probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)

    await probe()
    expect(registeredProbeCounters(probeRoot)).toEqual({
      maxSessionCompactionsSinceProgress: 0,
      stormSessionID: null,
      nonReadProgressCount: 0,
      compactionCount: 1,
    })

    // The run ends and its deadline drops; the next probe re-observes
    // the same marker still sitting as the session's latest message.
    unregisterDeadline("reviewer")
    await probe()
    expect(registeredProbeCounters(probeRoot)).toEqual({
      maxSessionCompactionsSinceProgress: 0,
      stormSessionID: null,
      nonReadProgressCount: 0,
      compactionCount: 1,
    })

    // A new marker first observed after the run ended does count.
    sessions.reviewer = { created: 1, latest: compactionMessage("rv-c2") }
    await probe()
    expect(registeredProbeCounters(probeRoot)).toEqual({
      maxSessionCompactionsSinceProgress: 1,
      stormSessionID: "reviewer",
      nonReadProgressCount: 0,
      compactionCount: 2,
    })
  })

  it("counts the probe root even though its own executor registered its deadline", async () => {
    // The #1079 root-storm protection (dewey#203): the probed root is
    // this executor's direct child, and this executor registered its
    // deadline at creation -- yet the root stays counted
    // unconditionally, so a root compaction loop still trips this
    // detector. The managed child in the same tree stays excluded.
    const probeRoot = root("r-root-counted")
    const sessions: Record<string, FakeSession> = {
      [probeRoot]: { children: ["managed"], latest: compactionMessage("root-c1") },
      managed: { created: 1, latest: compactionMessage("m-c1") },
    }
    register(probeRoot, Date.now() + 60_000)
    register("managed", Date.now() + 60_000)
    const fake = treeClient({ sessions })

    await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    expect(registeredProbeCounters(probeRoot)).toEqual({
      maxSessionCompactionsSinceProgress: 1,
      stormSessionID: probeRoot,
      nonReadProgressCount: 0,
      compactionCount: 2,
    })

    sessions[probeRoot] = { children: ["managed"], latest: compactionMessage("root-c2") }
    sessions.managed = { created: 1, latest: compactionMessage("m-c2") }
    await probeSubtreeActivity(fake.client, probeRoot, DIRECTORY, signal)
    expect(registeredProbeCounters(probeRoot)).toEqual({
      maxSessionCompactionsSinceProgress: 2,
      stormSessionID: probeRoot,
      nonReadProgressCount: 0,
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
    expect(sample?.maxSessionCompactionsSinceProgress).toBe(1)
    expect(sample?.stormSessionID).toBe("child")
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
    // DT-FR-014/015: a true stall has no running parts, so the deferral
    // is vacuous and the envelope discriminates it with running_tool null.
    expect(result.timeout).toMatchObject({ kind: "idle", child_active: false, idle_timeout_ms: 10_000, running_tool: null })
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

  it("defers idle while a quiet bash runs and reports it on the wall-clock abort (DT-FR-014/015)", async () => {
    // No fingerprint change after the baseline: under the old rule the
    // quiet running bash would idle-abort at 2 s; the deferral holds
    // idle until the wall clock fires at 5 s, and the envelope names
    // the tool that was running.
    const fake = executorClient({
      latest: () => message("m", [{ id: "b1", type: "tool", tool: "bash", state: { status: "running", metadata: { output: "steady" } } }]),
    })
    const pending = run(fake.client, { timeoutMilliseconds: 5_000, idleTimeoutMilliseconds: 2_000 })
    await vi.advanceTimersByTimeAsync(5_000)
    const result = await pending
    expect(result.error?.code).toBe("timeout")
    expect(result.timeout).toMatchObject({ kind: "wall_clock", running_tool: "bash" })
    // DT-FR-013: the running part counts as activity within the bound,
    // so progress reads at the polling interval, not the idle window.
    expect(result.timeout?.last_progress_ms_ago).toBe(1_000)
    expect(result.timeout?.child_active).toBe(true)
  })

  it("does not promote a tool timeout error to an idle abort (DT-FR-014)", async () => {
    // The quiet bash survives the 3 s idle window under the deferral;
    // at 5 s the part's status flips to error -- the tool's own
    // timeout returned -- which is a fingerprint change that resets
    // the idle clock, so the model gets a full idle window to handle
    // the error instead of an immediate idle kill.
    let phase: "running" | "error" = "running"
    const fake = executorClient({
      latest: () =>
        phase === "running"
          ? message("m", [{ id: "b1", type: "tool", tool: "bash", state: { status: "running", metadata: { output: "steady" } } }])
          : message("m", [{ id: "b1", type: "tool", tool: "bash", state: { status: "error", error: "command timed out" } }]),
    })
    const pending = run(fake.client, { timeoutMilliseconds: 60_000, idleTimeoutMilliseconds: 3_000 })
    await vi.advanceTimersByTimeAsync(4_000)
    phase = "error"
    await vi.advanceTimersByTimeAsync(10_000)
    const result = await pending
    expect(result.error?.code).toBe("idle_timeout")
    // Baseline at 1 s; the flip lands on the 5 s probe, so the idle
    // clock restarts at 5 s and fires at 8 s with no running part.
    expect(result.timeout).toMatchObject({ kind: "idle", running_tool: null, elapsed_ms: 8_000 })
  })

  it("aborts a never-returned dispatch_review_runs wait as idle with no running parts (DT-FR-014)", async () => {
    // The #1117 evidence: the model called dispatch_review_runs, it
    // never returned, and no part in the subtree carries a running
    // state -- stalling between tool calls stays subject to the plain
    // idle window.
    const fake = executorClient({
      latest: () => message("m", [{ id: "dr", type: "tool", tool: "dispatch_review_runs" }]),
    })
    const pending = run(fake.client, { timeoutMilliseconds: 60_000, idleTimeoutMilliseconds: 10_000 })
    await vi.advanceTimersByTimeAsync(15_000)
    const result = await pending
    expect(result.error?.code).toBe("idle_timeout")
    expect(result.timeout).toMatchObject({ kind: "idle", running_tool: null, idle_timeout_ms: 10_000 })
  })

  it("aborts a stuck running part at idle_timeout plus IDLE_DEFERRAL_CAP_MS (DT-FR-014)", async () => {
    // A running bash with no observable deadline and no fingerprint
    // change holds idle exactly until idle_timeout plus the cap,
    // measured from the baseline probe's fingerprint change.
    const fake = executorClient({
      latest: () => message("m", [{ id: "b1", type: "tool", tool: "bash", state: { status: "running", metadata: { output: "steady" } } }]),
    })
    const pending = run(fake.client, { timeoutMilliseconds: 2_000_000, idleTimeoutMilliseconds: 600_000 })
    await vi.advanceTimersByTimeAsync(1_501_000)
    const result = await pending
    expect(result.error?.code).toBe("idle_timeout")
    expect(result.timeout).toMatchObject({ kind: "idle", running_tool: "bash", idle_timeout_ms: 600_000 })
    // Baseline at 1 s: 1 s + 600 s idle + 900 s cap = 1,501 s.
    expect(result.timeout?.elapsed_ms).toBe(1_501_000)
  })

  it("reports running_tool null for an adversarial tool name without echoing it (DT-FR-015)", async () => {
    // The smallest-partId running part carries a name that fails
    // validation: the envelope reports null and the raw value never
    // reaches the result (Constitution V: untrusted input is never
    // echoed).
    const fake = executorClient({
      latest: () => message("m", [{ id: "evil", type: "tool", tool: "bash; rm -rf /", state: { status: "running" } }]),
    })
    const pending = run(fake.client, { timeoutMilliseconds: 5_000, idleTimeoutMilliseconds: 2_000 })
    await vi.advanceTimersByTimeAsync(5_000)
    const result = await pending
    expect(result.error?.code).toBe("timeout")
    expect(result.timeout?.running_tool).toBeNull()
    expect(JSON.stringify(result.timeout)).not.toContain("rm -rf")
    expect(JSON.stringify(result)).not.toContain("rm -rf")
  })

  it("keeps a 15-minute quiet bash alive under a 10-minute idle timeout (DT-FR-013 scenario 1)", async () => {
    // Timer-controlled #1117 scenario, no real 15-minute sleep: the
    // quiet running bash holds the idle abort past the 10-minute idle
    // deadline; the wall clock (16 minutes) ends the run, and the
    // envelope names the tool with progress at the polling interval.
    const fake = executorClient({
      latest: () => message("m", [{ id: "b1", type: "tool", tool: "bash", state: { status: "running", metadata: { output: "steady" } } }]),
    })
    const pending = run(fake.client, { timeoutMilliseconds: 960_000, idleTimeoutMilliseconds: 600_000 })
    await vi.advanceTimersByTimeAsync(900_000)
    let settled = false
    void pending.then(
      () => {
        settled = true
      },
      () => {
        settled = true
      },
    )
    await vi.advanceTimersByTimeAsync(0)
    // 15 minutes in -- five minutes past the idle deadline -- the run
    // is still alive.
    expect(settled).toBe(false)
    await vi.advanceTimersByTimeAsync(60_000)
    const result = await pending
    expect(result.error?.code).toBe("timeout")
    expect(result.timeout).toMatchObject({ kind: "wall_clock", running_tool: "bash", child_active: true })
    expect(result.timeout?.last_progress_ms_ago).toBe(1_000)
  })

  it("aborts the same 10-minute idle window with no running parts (DT-FR-014 true stall)", async () => {
    const fake = executorClient({ latest: () => message("m", [{ id: "p", type: "text", text: "same" }]) })
    const pending = run(fake.client, { timeoutMilliseconds: 960_000, idleTimeoutMilliseconds: 600_000 })
    await vi.advanceTimersByTimeAsync(700_000)
    const result = await pending
    expect(result.error?.code).toBe("idle_timeout")
    expect(result.timeout).toMatchObject({ kind: "idle", running_tool: null, idle_timeout_ms: 600_000 })
    // Baseline probe at 1 s: the idle window completes at 601 s.
    expect(result.timeout?.elapsed_ms).toBe(601_000)
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
