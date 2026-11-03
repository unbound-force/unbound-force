/**
 * Hierarchical delegation budgets and child-session activity probing
 * shared by every executeAgentSession caller (OpenSpec change
 * delegation-timeout-budgets, design D1/D2, issue #930).
 *
 * A child session created by the executor records its absolute
 * deadline here; a nested executor call made from inside that child
 * (or from a session the host created beneath it) resolves the
 * deadline and clamps its own wall clock to the remaining budget, so
 * nested work never outlives its parent. The activity probe
 * fingerprints a child's session subtree so the executor can tell a
 * working child from a stalled one.
 */

import { type PluginInput } from "@opencode-ai/plugin"

// ── Constants ────────────────────────────────────────────────

/** Time reserved for the parent to receive and act on a nested failure. */
export const BUDGET_RESERVE_MS = 60_000

/** Smallest clamped budget worth spawning a nested child for. */
export const MIN_NESTED_BUDGET_MS = 30_000

/** Maximum parentID hops walked when resolving an inherited deadline. */
export const MAX_ANCESTRY_DEPTH = 8

/** Maximum sessions the activity probe fingerprints per poll. */
export const MAX_PROBE_SESSIONS = 32

/** Window within which observed progress marks a child as active. */
export const ACTIVE_WINDOW_MS = 120_000

/** Default interval between activity probes. */
export const DEFAULT_IDLE_POLL_MS = 30_000

const REGISTRY_KEY = Symbol.for("unbound-force.delegation-budget")

type Client = PluginInput["client"]

// ── Deadline registry ────────────────────────────────────────

/**
 * Process-wide deadline registry (design D1). Stored on globalThis
 * under a registered symbol so the separate plugin modules that import
 * this library share one registry even if the host loads the library
 * more than once. Holds only session ids and epoch milliseconds.
 */
function registry(): Map<string, number> {
  const holder = globalThis as unknown as Record<symbol, Map<string, number> | undefined>
  let map = holder[REGISTRY_KEY]
  if (map === undefined) {
    map = new Map<string, number>()
    holder[REGISTRY_KEY] = map
  }
  return map
}

/** Record a child session's absolute deadline (epoch ms). */
export function registerDeadline(sessionID: string, deadlineMs: number): void {
  registry().set(sessionID, deadlineMs)
}

/** Remove a child session's deadline once its run has ended. */
export function unregisterDeadline(sessionID: string): void {
  registry().delete(sessionID)
}

/** The registered deadline for a session, or undefined. */
export function registeredDeadline(sessionID: string): number | undefined {
  return registry().get(sessionID)
}

/**
 * Resolve the deadline a nested call inherits: the registered deadline
 * of `sessionID`, else of the nearest registered ancestor found by
 * walking `parentID` links up to MAX_ANCESTRY_DEPTH hops. Any lookup
 * failure (or a client without `session.get`) yields null, which
 * leaves top-level behavior unchanged.
 */
export async function resolveInheritedDeadline(
  client: Client,
  sessionID: string,
  directory: string,
  signal: AbortSignal,
): Promise<number | null> {
  let current: string | undefined = sessionID
  for (let depth = 0; current !== undefined && depth <= MAX_ANCESTRY_DEPTH; depth++) {
    const deadline = registeredDeadline(current)
    if (deadline !== undefined) {
      return deadline
    }
    if (depth === MAX_ANCESTRY_DEPTH || typeof client.session?.get !== "function") {
      return null
    }
    try {
      const response = await client.session.get({ path: { id: current }, query: { directory }, signal })
      current = response.data?.parentID
    } catch {
      return null
    }
  }
  return null
}

/** Outcome of clamping a requested wall clock to an inherited budget. */
export interface BudgetClamp {
  readonly effectiveMs: number
  readonly clamped: boolean
  readonly exhausted: boolean
}

/**
 * Clamp a requested wall clock to the inherited deadline minus
 * BUDGET_RESERVE_MS. Exhausted when less than MIN_NESTED_BUDGET_MS
 * would remain; a null deadline passes the request through.
 */
export function clampToBudget(requestedMs: number, inheritedDeadline: number | null, now: number): BudgetClamp {
  if (inheritedDeadline === null) {
    return { effectiveMs: requestedMs, clamped: false, exhausted: false }
  }
  const available = inheritedDeadline - now - BUDGET_RESERVE_MS
  if (available < MIN_NESTED_BUDGET_MS) {
    return { effectiveMs: Math.max(0, available), clamped: true, exhausted: true }
  }
  return available < requestedMs
    ? { effectiveMs: available, clamped: true, exhausted: false }
    : { effectiveMs: requestedMs, clamped: false, exhausted: false }
}

// ── Activity probe ───────────────────────────────────────────

/** One activity sample of a child's session subtree. */
export interface ActivitySample {
  /** Opaque fingerprint; any change between samples is progress. */
  readonly fingerprint: string
  /** The latest part of some subtree session is a running question tool call. */
  readonly pendingQuestion: boolean
}

interface PartLike {
  readonly id?: unknown
  readonly type?: unknown
  readonly tool?: unknown
  readonly text?: unknown
  readonly state?: { readonly status?: unknown; readonly output?: unknown }
}

interface MessageLike {
  readonly info?: {
    readonly id?: unknown
    readonly time?: { readonly completed?: unknown }
    readonly error?: unknown
  }
  readonly parts?: readonly PartLike[]
}

/**
 * Fingerprint one session's latest message: message id, completion,
 * error presence, part count, and the last part's id, type, tool
 * status, and text/output length (streaming grows the length).
 */
export function messageFingerprint(sessionID: string, message: MessageLike | undefined): string {
  if (message === undefined) {
    return `${sessionID}:-`
  }
  const parts = message.parts ?? []
  const last = parts[parts.length - 1]
  const textLength = typeof last?.text === "string" ? last.text.length : 0
  const outputLength = typeof last?.state?.output === "string" ? last.state.output.length : 0
  return [
    sessionID,
    String(message.info?.id ?? ""),
    String(message.info?.time?.completed ?? ""),
    message.info?.error === undefined ? "" : "e",
    String(parts.length),
    String(last?.id ?? ""),
    String(last?.type ?? ""),
    String(last?.state?.status ?? ""),
    String(textLength + outputLength),
  ].join(":")
}

function isPendingQuestion(message: MessageLike | undefined): boolean {
  const parts = message?.parts ?? []
  const last = parts[parts.length - 1]
  return last?.type === "tool" && last.tool === "question" && last.state?.status === "running"
}

/**
 * Sample the activity of `rootID` and its descendant sessions
 * (breadth-first over `session.children`, newest first, capped at
 * MAX_PROBE_SESSIONS). Returns null when the client lacks the probe
 * endpoints or any request fails, so callers never mistake an
 * unreachable API for a stalled child.
 */
export async function probeSubtreeActivity(
  client: Client,
  rootID: string,
  directory: string,
  signal: AbortSignal,
): Promise<ActivitySample | null> {
  if (typeof client.session?.children !== "function" || typeof client.session?.messages !== "function") {
    return null
  }
  try {
    const queue = [rootID]
    const seen = new Set<string>()
    const prints: string[] = []
    let pendingQuestion = false
    while (queue.length > 0 && seen.size < MAX_PROBE_SESSIONS) {
      const id = queue.shift() as string
      if (seen.has(id)) {
        continue
      }
      seen.add(id)
      const messages = await client.session.messages({ path: { id }, query: { directory, limit: 1 }, signal })
      if (messages.data === undefined) {
        return null
      }
      const latest = messages.data[messages.data.length - 1] as MessageLike | undefined
      prints.push(messageFingerprint(id, latest))
      pendingQuestion ||= isPendingQuestion(latest)
      const children = await client.session.children({ path: { id }, query: { directory }, signal })
      if (children.data === undefined) {
        return null
      }
      // The child count rides in the fingerprint so a newly spawned
      // session registers as progress even past the session cap, and
      // newest children are probed first: long-lived children accumulate
      // finished sessions, and the active one is the most recent.
      prints.push(`${id}#${children.data.length}`)
      const newestFirst = [...children.data].sort((a, b) => (b.time?.created ?? 0) - (a.time?.created ?? 0))
      for (const child of newestFirst) {
        queue.push(child.id)
      }
    }
    return { fingerprint: prints.join("|"), pendingQuestion }
  } catch {
    return null
  }
}
