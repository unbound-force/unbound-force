/**
 * Hierarchical delegation budgets and child-session activity probing
 * shared by every executeAgentSession caller (OpenSpec change
 * delegation-timeout-budgets, design D1/D2, issue #930). Compaction
 * accounting per compaction-storm-on-a-tight-context-mode design D2
 * (DT-FR-002 / DELEG-FR-002, issue #1079).
 *
 * A child session created by the executor records its absolute
 * deadline here; a nested executor call made from inside that child
 * (or from a session the host created beneath it) resolves the
 * deadline and clamps its own wall clock to the remaining budget, so
 * nested work never outlives its parent. The activity probe
 * fingerprints a child's session subtree so the executor can tell a
 * working child from a stalled one. Compaction markers (context
 * compression summaries and synthetic continuation prompts) are
 * structurally classified and excluded from progress, and the probe
 * exposes the compaction-since-progress accounting shared by the
 * idle watchdog and the compaction-storm detector.
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

/**
 * Maximum length of a text that may still match the system-reminder
 * compaction-summary wrapper. Real compaction summaries measure in the
 * tens of thousands of characters, so the bound only exists to keep a
 * pathological legitimate text that merely starts and ends with the
 * wrapper tags from matching (design D3: matchers are anchored and
 * length-bounded, never bare substrings).
 */
export const MAX_SYNTHETIC_WRAPPER_LENGTH = 65_536

/**
 * Maximum length of a text that may still match the synthetic
 * continuation prompt. The prompt is one sentence, optionally prefixed
 * by opencode's documented large-media overflow preamble (about 460
 * characters total); anything longer cannot be the injected prompt.
 */
export const MAX_CONTINUATION_PROMPT_LENGTH = 4_096

/**
 * Tool names whose completion is read-only inspection and therefore
 * never counts as substantive progress (DT-FR-002: file reads,
 * searches, and other inspection MUST NOT count). Every other tool --
 * including project MCP tools -- defaults to non-read: the spec's
 * exclusion list is explicit, and an unknown tool that completes has
 * usually performed durable work worth resetting the storm window
 * for. Read tools are the exception, not the rule.
 */
export const READ_ONLY_TOOLS: ReadonlySet<string> = new Set(["read", "grep", "glob", "list", "ls", "webfetch"])

const REGISTRY_KEY = Symbol.for("unbound-force.delegation-budget")
const PROBE_REGISTRY_KEY = Symbol.for("unbound-force.delegation-budget.probe")

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

/**
 * Compaction-accounting counters shared by the idle watchdog and the
 * compaction-storm detector (compaction-storm change design D2;
 * DT-FR-002 requires one accounting for both consumers).
 */
export interface ProbeCounters {
  /** Compaction markers observed since the last substantive progress. */
  readonly compactionsSinceProgress: number
  /** Substantive non-read progress observations across all probes. */
  readonly nonReadProgressCount: number
  /** Total compaction markers observed across all probes. */
  readonly compactionCount: number
}

/** One activity sample of a child's session subtree. */
export interface ActivitySample extends ProbeCounters {
  /** Opaque fingerprint; any change between samples is progress. */
  readonly fingerprint: string
  /** The latest part of some subtree session is a running question tool call. */
  readonly pendingQuestion: boolean
}

/** Structural kind of a host-injected synthetic text shape (design D3). */
export type SyntheticTextKind = "compaction-summary" | "continuation-prompt"

interface PartLike {
  readonly id?: unknown
  readonly type?: unknown
  readonly tool?: unknown
  readonly text?: unknown
  readonly metadata?: unknown
  readonly state?: {
    readonly status?: unknown
    readonly output?: unknown
    readonly title?: unknown
    readonly metadata?: unknown
  }
}

interface MessageLike {
  readonly info?: {
    readonly id?: unknown
    readonly role?: unknown
    readonly mode?: unknown
    readonly agent?: unknown
    readonly time?: { readonly completed?: unknown }
    readonly error?: unknown
  }
  readonly parts?: readonly PartLike[]
}

/** Serialized length of a running tool's metadata; 0 when absent or unserializable. */
function metadataLength(metadata: unknown): number {
  if (metadata === undefined || metadata === null) {
    return 0
  }
  try {
    return JSON.stringify(metadata)?.length ?? 0
  } catch {
    return 0
  }
}

const SYSTEM_REMINDER_OPEN = "<system-reminder>"
const SYSTEM_REMINDER_CLOSE = "</system-reminder>"
// Both injected variants of opencode's continuation prompt share this
// closing sentence: the bare prompt and the large-media overflow
// preamble prefixed to it. Anchoring on the shared suffix (plus the
// length bound) matches both without hardcoding the volatile preamble.
const CONTINUATION_PROMPT_SUFFIX =
  "continue if you have next steps, or stop and ask for clarification if you are unsure how to proceed."

/**
 * Classify a text against the known host-injected synthetic shapes: a
 * system-reminder-wrapped compaction summary or the synthetic
 * continuation prompt. Matchers are structural per design D3 --
 * anchored at the wrapper edges or the prompt's closing sentence,
 * case-insensitive, and length-bounded -- so legitimate prose that
 * merely mentions compaction or reminders never matches.
 *
 * @param text - The text to classify.
 * @returns The matched shape, or null when the text is not a known synthetic wrapper.
 */
export function classifySyntheticText(text: string): SyntheticTextKind | null {
  const trimmed = text.trim()
  if (trimmed.length === 0) {
    return null
  }
  const lowered = trimmed.toLowerCase()
  if (
    trimmed.length <= MAX_SYNTHETIC_WRAPPER_LENGTH &&
    lowered.startsWith(SYSTEM_REMINDER_OPEN) &&
    lowered.endsWith(SYSTEM_REMINDER_CLOSE)
  ) {
    return "compaction-summary"
  }
  if (trimmed.length <= MAX_CONTINUATION_PROMPT_LENGTH && lowered.endsWith(CONTINUATION_PROMPT_SUFFIX)) {
    return "continuation-prompt"
  }
  return null
}

/**
 * Whether a part-level metadata object carries a boolean flag set to true.
 * Narrowing helper for the untyped part metadata the probe observes.
 */
function partMetadataFlag(metadata: unknown, flag: string): boolean {
  if (typeof metadata !== "object" || metadata === null) {
    return false
  }
  return (metadata as Record<string, unknown>)[flag] === true
}

/**
 * Whether a message structurally matches a compaction marker: a
 * host-injected context-compression artifact. Matches the opencode
 * storage shapes (a `compaction` part, a message generated by the
 * `compaction` agent, a part flagged `compaction_continue`) and the
 * text shapes (a user-role system-reminder wrapper or the synthetic
 * continuation prompt on the last part, mirroring the fingerprint's
 * view), so the classification survives host-side representation
 * drift. Never matches bare words like "compaction" appearing in
 * legitimate prose (design D3).
 *
 * @param message - The latest message of a subtree session, if any.
 * @returns True when the message is a compaction/synthetic marker.
 */
export function isCompactionMarker(message: MessageLike | undefined): boolean {
  if (message === undefined) {
    return false
  }
  if (message.info?.mode === "compaction" || message.info?.agent === "compaction") {
    return true
  }
  const parts = message.parts ?? []
  for (const part of parts) {
    if (part?.type === "compaction") {
      return true
    }
    if (partMetadataFlag(part?.metadata, "compaction_continue")) {
      return true
    }
  }
  if (message.info?.role === "user") {
    const last = parts[parts.length - 1]
    if (typeof last?.text === "string" && classifySyntheticText(last.text) !== null) {
      return true
    }
  }
  return false
}

/**
 * Whether a message's latest part is the successful completion of a
 * non-read tool call -- the observable form of substantive progress
 * (file writes, edits, shell commands, and checkpoint records are all
 * completed tool calls). Read-only tools never count whatever their
 * outcome, and a non-read tool call that failed or has not completed
 * does not count (DT-FR-002 / DELEG-FR-002: only successful
 * completions reset the storm window).
 *
 * @param message - The latest message of a subtree session, if any.
 * @returns True when the message records completed non-read work.
 */
export function isSubstantiveProgress(message: MessageLike | undefined): boolean {
  if (message === undefined) {
    return false
  }
  const parts = message.parts ?? []
  const last = parts[parts.length - 1]
  if (last?.type !== "tool" || typeof last.tool !== "string" || READ_ONLY_TOOLS.has(last.tool)) {
    return false
  }
  return last.state?.status === "completed"
}

/**
 * Stable identity of a marker message: message id, part count, and
 * last part id. Deliberately excludes streamed text length so a
 * compaction summary that grows while it generates stays one counted
 * marker, not one per poll; message ids are append-only in practice,
 * so per-session last-key tracking cannot double count an event.
 */
function markerKey(message: MessageLike): string {
  const parts = message.parts ?? []
  const last = parts[parts.length - 1]
  return `${String(message.info?.id ?? "")}:${parts.length}:${String(last?.id ?? "")}`
}

/**
 * Fingerprint one session's latest message: message id, completion,
 * error presence, part count, and the last part's id, type, tool
 * status, title, and text/output/metadata length. Streaming grows the
 * text length, and a running tool streams through its state metadata
 * (bash keeps live output in `state.metadata.output`; the running
 * state has no `output` field), so a long command that keeps printing
 * counts as progress. Compaction markers collapse to one constant per
 * session so marker-to-marker churn leaves the fingerprint unchanged
 * (DT-FR-002); probeSubtreeActivity additionally substitutes the
 * session's remembered pre-marker print so a marker's arrival after a
 * real message also reads as no change.
 */
export function messageFingerprint(sessionID: string, message: MessageLike | undefined): string {
  if (message === undefined) {
    return `${sessionID}:-`
  }
  if (isCompactionMarker(message)) {
    return `${sessionID}:synthetic`
  }
  const parts = message.parts ?? []
  const last = parts[parts.length - 1]
  const textLength = typeof last?.text === "string" ? last.text.length : 0
  const outputLength = typeof last?.state?.output === "string" ? last.state.output.length : 0
  const streamedLength = textLength + outputLength + metadataLength(last?.state?.metadata)
  return [
    sessionID,
    String(message.info?.id ?? ""),
    String(message.info?.time?.completed ?? ""),
    message.info?.error === undefined ? "" : "e",
    String(parts.length),
    String(last?.id ?? ""),
    String(last?.type ?? ""),
    String(last?.state?.status ?? ""),
    typeof last?.state?.title === "string" ? last.state.title : "",
    String(streamedLength),
  ].join(":")
}

function isPendingQuestion(message: MessageLike | undefined): boolean {
  const parts = message?.parts ?? []
  const last = parts[parts.length - 1]
  return last?.type === "tool" && last.tool === "question" && last.state?.status === "running"
}

/** One session's latest message as observed during a probe walk. */
interface ProbeObservation {
  readonly id: string
  readonly latest: MessageLike | undefined
}

/** Per-session probe memory: the last real print and the last counted marker and progress identities. */
interface ProbeSessionState {
  lastPrint: string | undefined
  lastMarkerKey: string | undefined
  lastProgressKey: string | undefined
}

/** Persistent compaction accounting for one probed subtree root (design D2). */
interface ProbeState {
  readonly sessions: Map<string, ProbeSessionState>
  compactionsSinceProgress: number
  nonReadProgressCount: number
  compactionCount: number
}

/**
 * Process-wide probe-state registry. Stored on globalThis under a
 * registered symbol for the same reason as the deadline registry:
 * separately loaded plugin modules share one accounting per probed
 * child. Keyed by the probed root session id.
 */
function probeRegistry(): Map<string, ProbeState> {
  const holder = globalThis as unknown as Record<symbol, Map<string, ProbeState> | undefined>
  let map = holder[PROBE_REGISTRY_KEY]
  if (map === undefined) {
    map = new Map<string, ProbeState>()
    holder[PROBE_REGISTRY_KEY] = map
  }
  return map
}

/**
 * The recorded compaction accounting for a probed subtree root.
 *
 * @param rootID - The root session id to look up.
 * @returns A snapshot of the counters, or undefined when the root has no probe state.
 */
export function registeredProbeCounters(rootID: string): ProbeCounters | undefined {
  const state = probeRegistry().get(rootID)
  if (state === undefined) {
    return undefined
  }
  return {
    compactionsSinceProgress: state.compactionsSinceProgress,
    nonReadProgressCount: state.nonReadProgressCount,
    compactionCount: state.compactionCount,
  }
}

/**
 * Remove a subtree root's probe accounting once its run has ended.
 * Mirrors unregisterDeadline; the executor drops both when a child
 * finishes so state for finished children cannot accumulate.
 *
 * @param rootID - The root session id whose probe state should be dropped.
 */
export function unregisterProbeState(rootID: string): void {
  probeRegistry().delete(rootID)
}

/**
 * Fingerprint contribution of one session's latest message. A marker
 * contributes the session's remembered pre-marker print (or the
 * absent-message placeholder when no real message was ever observed),
 * so neither a marker's arrival nor marker-to-marker churn reads as
 * subtree progress (DT-FR-002: compaction events MUST NOT count as
 * activity changes for the idle clock).
 */
function probeContribution(state: ProbeState | undefined, sessionID: string, latest: MessageLike | undefined): string {
  if (!isCompactionMarker(latest)) {
    return messageFingerprint(sessionID, latest)
  }
  return state?.sessions.get(sessionID)?.lastPrint ?? `${sessionID}:-`
}

/**
 * Fold one completed walk's observations into the root's persistent
 * accounting, in walk order (newest session first). Applied only
 * after the walk succeeds so a failed probe advances nothing
 * (DT-FR-002: a failed probe changes no clock). Marker increments
 * and substantive resets are identity-gated per session: re-observing
 * the same marker or the same completed tool is a no-op, which keeps
 * a quiet session's stale latest message from masking a storm
 * elsewhere in the subtree. Substantive progress is a completed
 * non-read tool call (reads and failed calls never reset; DELEG-FR-002).
 */
function foldProbeObservations(rootID: string, observations: readonly ProbeObservation[]): ProbeCounters {
  const registry = probeRegistry()
  let state = registry.get(rootID)
  if (state === undefined) {
    state = { sessions: new Map(), compactionsSinceProgress: 0, nonReadProgressCount: 0, compactionCount: 0 }
    registry.set(rootID, state)
  }
  for (const observation of observations) {
    let session = state.sessions.get(observation.id)
    if (session === undefined) {
      session = { lastPrint: undefined, lastMarkerKey: undefined, lastProgressKey: undefined }
      state.sessions.set(observation.id, session)
    }
    if (isCompactionMarker(observation.latest)) {
      const key = markerKey(observation.latest)
      if (session.lastMarkerKey !== key) {
        session.lastMarkerKey = key
        state.compactionsSinceProgress += 1
        state.compactionCount += 1
      }
      continue
    }
    const print = messageFingerprint(observation.id, observation.latest)
    session.lastPrint = print
    if (isSubstantiveProgress(observation.latest) && session.lastProgressKey !== print) {
      session.lastProgressKey = print
      state.compactionsSinceProgress = 0
      state.nonReadProgressCount += 1
    }
  }
  return {
    compactionsSinceProgress: state.compactionsSinceProgress,
    nonReadProgressCount: state.nonReadProgressCount,
    compactionCount: state.compactionCount,
  }
}

/**
 * Sample the activity of `rootID` and its descendant sessions
 * (depth-first over `session.children`, newest child first, capped at
 * MAX_PROBE_SESSIONS; two requests per session). Depth-first matters:
 * a long-lived child accumulates finished direct children (classifier
 * and reviewer sessions), and a breadth-first walk would spend the cap
 * on them before reaching the active grandchild under the newest one.
 * Returns null when the client lacks the probe endpoints or any
 * request fails, so callers never mistake an unreachable API for a
 * stalled child; a failed probe also leaves the compaction accounting
 * untouched because the fold runs only after a successful walk.
 *
 * The sample carries the compaction accounting (design D2): compaction
 * markers never change the fingerprint, and the counters expose how
 * many markers accrued since the last substantive progress for the
 * idle watchdog and the storm detector to share (DT-FR-002 /
 * DELEG-FR-002). Accounting persists across probes per root id;
 * callers drop it with unregisterProbeState when the run ends.
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
    const stack = [rootID]
    const seen = new Set<string>()
    const prints: string[] = []
    const observations: ProbeObservation[] = []
    let pendingQuestion = false
    // Read-only during the walk: foldProbeObservations is the only
    // writer, and it runs only after the walk succeeds.
    const state = probeRegistry().get(rootID)
    while (stack.length > 0 && seen.size < MAX_PROBE_SESSIONS) {
      const id = stack.pop() as string
      if (seen.has(id)) {
        continue
      }
      seen.add(id)
      const messages = await client.session.messages({ path: { id }, query: { directory, limit: 1 }, signal })
      if (messages.data === undefined) {
        return null
      }
      const latest = messages.data[messages.data.length - 1] as MessageLike | undefined
      observations.push({ id, latest })
      prints.push(probeContribution(state, id, latest))
      pendingQuestion ||= isPendingQuestion(latest)
      const children = await client.session.children({ path: { id }, query: { directory }, signal })
      if (children.data === undefined) {
        return null
      }
      // The child count rides in the fingerprint so a newly spawned
      // session registers as progress even past the session cap. Children
      // are pushed oldest first so the newest (the active one) pops next.
      prints.push(`${id}#${children.data.length}`)
      const oldestFirst = [...children.data].sort((a, b) => (a.time?.created ?? 0) - (b.time?.created ?? 0))
      for (const child of oldestFirst) {
        stack.push(child.id)
      }
    }
    return { fingerprint: prints.join("|"), pendingQuestion, ...foldProbeObservations(rootID, observations) }
  } catch {
    return null
  }
}
