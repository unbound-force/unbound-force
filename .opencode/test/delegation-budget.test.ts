import type { PluginInput } from "@opencode-ai/plugin"
import { afterEach, describe, expect, it } from "vitest"

import { executeAgentSession, type InvocationProvenance } from "../lib/agent-executor.js"
import {
  BUDGET_RESERVE_MS,
  MAX_ANCESTRY_DEPTH,
  MAX_PROBE_SESSIONS,
  MIN_NESTED_BUDGET_MS,
  clampToBudget,
  messageFingerprint,
  probeSubtreeActivity,
  registerDeadline,
  registeredDeadline,
  resolveInheritedDeadline,
  unregisterDeadline,
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
      if (options.failMessages === true) {
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

  it("records completion and error presence", () => {
    const open = messageFingerprint("s", message("m", []))
    const failed = messageFingerprint("s", message("m", [], { error: { name: "x" } }))
    expect(open).not.toBe(failed)
    expect(messageFingerprint("s", { parts: [] })).toBe("s::::0::::0")
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
  /** Parent session id registered or walked for an inherited deadline. */
  readonly parentID?: string
  /** Returns the child's latest message on each probe (undefined = probe failure). */
  readonly latest?: () => Record<string, unknown> | undefined
  /** Child prompt resolves after this many ms instead of hanging. */
  readonly resolveAfterMs?: number
}

/** Executor-facing fake: child creation, a hanging prompt, abort, and probe endpoints. */
function executorClient(options: ExecutorFakeOptions = {}) {
  const calls = { creates: 0, aborts: 0, deadlineDuringRun: undefined as number | undefined }
  const session = {
    get: async () => ({ data: { id: "parent", parentID: options.parentID } }),
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
        request.signal?.addEventListener("abort", () => reject(request.signal?.reason), { once: true })
      })
    },
    abort: async () => {
      calls.aborts++
      return { data: true }
    },
    children: async () => ({ data: [] }),
    messages: async () => {
      const latest = options.latest?.()
      if (latest === undefined) {
        throw new Error("probe failed")
      }
      return { data: [latest] }
    },
  }
  return { client: { session } as unknown as Client, calls }
}

function run(
  client: Client,
  extra: {
    timeoutMilliseconds?: number
    idleTimeoutMilliseconds?: number
    now?: () => number
    parentAbort?: AbortSignal
  } = {},
) {
  return executeAgentSession(
    {
      client,
      sessionID: "parent",
      directory: DIRECTORY,
      parentAbort: extra.parentAbort ?? new AbortController().signal,
      timeoutMilliseconds: extra.timeoutMilliseconds ?? 60_000,
      idleTimeoutMilliseconds: extra.idleTimeoutMilliseconds,
      idlePollMilliseconds: 1,
      now: extra.now,
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
    const clock = 1_000_000
    register("parent", clock + BUDGET_RESERVE_MS + MIN_NESTED_BUDGET_MS)
    const controller = new AbortController()
    const fake = executorClient()
    const pending = run(fake.client, { timeoutMilliseconds: 1_800_000, now: () => clock, parentAbort: controller.signal })
    while (fake.calls.deadlineDuringRun === undefined) {
      await new Promise((resolve) => setTimeout(resolve, 1))
    }
    // The child inherits the clamped absolute deadline, not the request.
    expect(fake.calls.deadlineDuringRun).toBe(clock + MIN_NESTED_BUDGET_MS)
    controller.abort(new Error("stop"))
    const result = await pending
    expect(result.error?.code).toBe("cancelled")
    expect(result.timeout).toBeUndefined()
  })

  it("reports a plain wall-clock timeout with no probe data", async () => {
    const result = await run(executorClient().client, { timeoutMilliseconds: 5 })
    expect(result.error?.code).toBe("timeout")
    expect(result.timeout).toMatchObject({
      kind: "wall_clock",
      effective_ms: 5,
      requested_ms: 5,
      budget_clamped: false,
      idle_timeout_ms: null,
      last_progress_ms_ago: null,
      child_active: null,
      pending_question: false,
    })
  })

  it("aborts a stalled child with idle_timeout after an unchanged probe", async () => {
    let clock = 0
    const fake = executorClient({ latest: () => message("m", [{ id: "p", type: "text", text: "same" }]) })
    const pending = run(fake.client, {
      idleTimeoutMilliseconds: 50,
      now: () => {
        clock += 20
        return clock
      },
    })
    const result = await pending
    expect(result.error?.code).toBe("idle_timeout")
    expect(result.error?.retryable).toBe(true)
    expect(result.error?.message).toBe("invoke_agent run made no progress for 50 ms")
    expect(result.timeout?.kind).toBe("idle")
    expect(result.timeout?.child_active).toBe(false)
    expect(result.timeout?.idle_timeout_ms).toBe(50)
    expect(result.childSessionID).toBe("exec-child")
    expect(fake.calls.aborts).toBe(1)
    // The child's deadline is removed once the run ends.
    expect(registeredDeadline("exec-child")).toBeUndefined()
  })

  it("keeps an active child alive and reports it active on wall-clock timeout", async () => {
    let counter = 0
    const fake = executorClient({
      latest: () => message("m", [{ id: "p", type: "text", text: "x".repeat(++counter) }]),
    })
    const result = await run(fake.client, { timeoutMilliseconds: 40, idleTimeoutMilliseconds: 1 })
    expect(result.error?.code).toBe("timeout")
    expect(result.timeout?.kind).toBe("wall_clock")
    expect(result.timeout?.child_active).toBe(true)
    expect(result.timeout?.last_progress_ms_ago).not.toBeNull()
  })

  it("never idle-aborts when every probe fails", async () => {
    const fake = executorClient({ latest: () => undefined })
    const result = await run(fake.client, { timeoutMilliseconds: 30, idleTimeoutMilliseconds: 1 })
    expect(result.error?.code).toBe("timeout")
    expect(result.timeout?.child_active).toBeNull()
  })

  it("returns success without timeout details when the child finishes", async () => {
    const fake = executorClient({
      resolveAfterMs: 5,
      latest: () => message("m", [{ id: "p", type: "text", text: "same" }]),
    })
    const result = await run(fake.client, { idleTimeoutMilliseconds: 60_000 })
    expect(result.status).toBe("success")
    expect(result.timeout).toBeUndefined()
  })
})
