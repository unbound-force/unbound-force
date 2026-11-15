/**
 * Shared progress-state persistence for the uf-workflow and review-dispatch
 * plugins (harden-spec-review-gate, HARDEN-001).
 *
 * Extracts the state-file I/O previously private to the uf-workflow plugin
 * (path derivation with traversal protection, symlink-refusing reads and
 * writes, restricted permissions — FR-005) so the review-dispatch plugin
 * can persist the spec-review gate record through the identical primitives,
 * and adds that record's durable I/O:
 *
 * - `appendSpecReviewIterationToState` — called by consolidate_dispatch in
 *   the same finalization step that persists the dispatch artifact, so the
 *   verdict reaches the progress state before control returns to the model.
 * - `readSpecReviewGateRecord` / `reconcileSpecReviewGateState` — every
 *   reader reconciles the two stores: specs-mode dispatch artifacts for the
 *   branch that are not yet iteration entries are appended (write
 *   timestamp order) and the healed record is written back, so no reader
 *   can observe iteration N-1 while artifact N is persisted on disk.
 *
 * Pure record shape and gate semantics live in uf-workflow-types.ts;
 * artifact envelope parsing lives in review-dispatch-artifact.ts. This
 * module owns only durable I/O (constitution Principle IV: the I/O seams
 * are the injectable surface; tests exercise them against temp roots).
 */

import {
  lstatSync,
  mkdirSync,
  readFileSync,
  readdirSync,
  rmSync,
  writeFileSync,
} from "node:fs"
import { resolve } from "node:path"

import {
  DISPATCH_ARTIFACT_DIRECTORY,
  parseReviewDispatchArtifactDetail,
  type ReviewDispatchArtifactDetail,
} from "./review-dispatch-artifact.js"
import {
  appendSpecReviewIteration,
  deriveSpecReviewOverrides,
  duplicateSpecReviewDecisionIds,
  isValidSpecReviewDecision,
  latestSpecReviewIteration,
  nextSpecReviewIteration,
  parseSpecReviewGateRecord,
  sanitizeSlug,
  SPEC_REVIEW_RECORD_KEY,
  SPEC_REVIEW_STATE_COMMAND,
  specReviewFindingIds,
  unknownSpecReviewDecisionIds,
  type SpecReviewDecision,
  type SpecReviewGateRecord,
  type SpecReviewIterationEntry,
} from "./uf-workflow-types.js"

// ---------------------------------------------------------------------------
// Progress state shape (moved from uf-workflow/index.ts — unchanged)
// ---------------------------------------------------------------------------

/** Failure classification for a phase pre-flight checkpoint. */
export type PhaseFailureClassification =
  | "pre-existing"
  | "environmental"
  | "change-caused"

/** One classified failure in a phase checkpoint inventory. */
export interface PhaseFailureEntry {
  name: string
  classification: PhaseFailureClassification
}

/** Phase-scoped checkpoint input accepted by the checkpoint tool. */
export interface PhaseCheckpointInput {
  number: number
  verdict: "green" | "failed"
  inventory: PhaseFailureEntry[]
}

/** Durable record of one phase's pre-flight checkpoint, keyed by phase number. */
export interface PhaseCheckpointRecord {
  verdict: "green" | "failed"
  inventory: PhaseFailureEntry[]
  recorded_at: string
}

/** Persistent workflow progress state (FR-001, FR-007). */
export interface ProgressState {
  command: string
  branch: string
  key: string
  current_step: string
  completed_steps: string[]
  data: Record<string, unknown>
  /** Phase-scoped checkpoint records keyed by phase number string. */
  phases?: Record<string, PhaseCheckpointRecord>
  created_at: string
  updated_at: string
}

// ---------------------------------------------------------------------------
// State file I/O (moved from uf-workflow/index.ts — unchanged semantics)
// ---------------------------------------------------------------------------

/** Progress state directory relative to project root. */
export const PROGRESS_DIR = ".uf/progress"

/**
 * Compute state file path with path traversal protection (FR-005).
 * Both key and sanitizedCommand MUST be pre-sanitized by the caller.
 */
function statePath(
  cwd: string,
  key: string,
  sanitizedCommand: string,
): string {
  const progressRoot = resolve(cwd, PROGRESS_DIR)
  const filePath = resolve(
    progressRoot,
    key,
    `${sanitizedCommand}.json`,
  )
  if (!filePath.startsWith(progressRoot + "/")) {
    throw new Error(
      "Path traversal blocked: resolved path escapes .uf/progress/",
    )
  }
  return filePath
}

/** Type guard for ENOENT filesystem errors. */
export function isEnoent(err: unknown): boolean {
  return (
    err instanceof Error &&
    "code" in err &&
    (err as NodeJS.ErrnoException).code === "ENOENT"
  )
}

/**
 * Read state from disk. Returns null if file does not exist.
 * FR-005: Does not follow symlinks; reports I/O errors.
 */
export function readState(
  cwd: string,
  key: string,
  sanitizedCommand: string,
): ProgressState | null {
  const fp = statePath(cwd, key, sanitizedCommand)
  try {
    const stats = lstatSync(fp)
    if (stats.isSymbolicLink()) {
      throw new Error("Refusing to read symlinked state file")
    }
    const raw = readFileSync(fp, "utf-8")
    return JSON.parse(raw) as ProgressState
  } catch (err: unknown) {
    if (isEnoent(err)) return null
    throw err
  }
}

/**
 * Write state to disk with restricted permissions (FR-005).
 * Creates directories with 0o700, files with 0o600.
 * Does not follow symlinks.
 */
export function writeState(
  cwd: string,
  key: string,
  sanitizedCommand: string,
  state: ProgressState,
): void {
  const fp = statePath(cwd, key, sanitizedCommand)
  const dir = resolve(fp, "..")
  try {
    const dirStats = lstatSync(dir)
    if (dirStats.isSymbolicLink()) {
      throw new Error("Refusing to write into symlinked directory")
    }
  } catch (err: unknown) {
    if (!isEnoent(err)) throw err
    // Directory does not exist yet — will be created below.
  }
  mkdirSync(dir, { recursive: true, mode: 0o700 })
  writeFileSync(fp, JSON.stringify(state, null, 2) + "\n", { mode: 0o600 })
}

/**
 * Delete state file. Idempotent — succeeds silently when no file
 * exists (FR-003).
 */
export function deleteState(
  cwd: string,
  key: string,
  sanitizedCommand: string,
): void {
  const fp = statePath(cwd, key, sanitizedCommand)
  try {
    const stat = lstatSync(fp)
    if (stat.isSymbolicLink()) {
      throw new Error(`Refusing to delete symlinked state file: ${fp}`)
    }
    rmSync(fp)
  } catch (err: unknown) {
    if (isEnoent(err)) return
    throw err
  }
}

// ---------------------------------------------------------------------------
// Spec-review gate record I/O (HARDEN-001, tasks 2.1/2.2)
// ---------------------------------------------------------------------------

/**
 * Extract the gate record from a state's data. An absent key yields an
 * empty record; a present-but-malformed record throws (fail closed): a
 * corrupted gate must surface as a tool error, never silently open the
 * gate.
 */
function recordFromState(state: ProgressState): SpecReviewGateRecord {
  const value = state.data[SPEC_REVIEW_RECORD_KEY]
  if (value === undefined) return { iterations: {} }
  const parsed = parseSpecReviewGateRecord(value)
  if (parsed === null) {
    throw new Error(
      `malformed ${SPEC_REVIEW_RECORD_KEY} record in ${SPEC_REVIEW_STATE_COMMAND} state (key ${state.key})`,
    )
  }
  return parsed
}

/**
 * Derive one iteration entry from a dispatch artifact. The recorded_at
 * timestamp mirrors the artifact's envelope write time so a reconciled
 * entry carries the moment the verdict actually became durable.
 */
function iterationEntryFromArtifact(
  artifact: ReviewDispatchArtifactDetail,
): SpecReviewIterationEntry {
  return {
    verdict: artifact.verdict,
    correlation_id: artifact.correlationId,
    finding_count: artifact.findingCount,
    finding_ids: specReviewFindingIds(artifact.findingCount),
    recorded_at: new Date(
      artifact.writtenAt > 0 ? artifact.writtenAt : Date.now(),
    ).toISOString(),
  }
}

/**
 * Specs-mode dispatch artifacts for the state's branch that the record
 * does not yet carry, in write-timestamp order. Unreadable or foreign
 * envelopes are skipped: reconciliation is a healing pass over the
 * tool-owned store, and an unparseable artifact has no verdict to record.
 */
function unrecordedSpecArtifacts(
  projectRoot: string,
  key: string,
  record: SpecReviewGateRecord,
): ReviewDispatchArtifactDetail[] {
  const recordedIds = new Set(
    Object.values(record.iterations).map((entry) => entry.correlation_id),
  )
  const dir = resolve(projectRoot, DISPATCH_ARTIFACT_DIRECTORY)
  let names: string[]
  try {
    names = readdirSync(dir)
  } catch {
    return [] // No artifact store yet — nothing to reconcile.
  }
  const artifacts: ReviewDispatchArtifactDetail[] = []
  for (const name of names) {
    if (!name.endsWith(".json")) continue
    let raw: string
    try {
      raw = readFileSync(resolve(dir, name), "utf8")
    } catch {
      continue // Unreadable artifact: skip (fail-open to "not reconciled").
    }
    const detail = parseReviewDispatchArtifactDetail(raw, name)
    if (detail === null) continue
    if (detail.mode !== "specs") continue
    if (sanitizeSlug(detail.branch) !== key) continue
    if (recordedIds.has(detail.correlationId)) continue
    artifacts.push(detail)
  }
  return artifacts.sort((a, b) => a.writtenAt - b.writtenAt)
}

/**
 * Heal the gate record against the artifact store and write the healed
 * state back. Appending an iteration supersedes any decisions recorded
 * against the previous latest (appendSpecReviewIteration semantics), which
 * is correct: a reconciled artifact is a consolidation the model was never
 * told about, and its verdict — not the stale decision set — governs.
 */
function healSpecReviewGateState(
  projectRoot: string,
  key: string,
  state: ProgressState,
): { state: ProgressState; record: SpecReviewGateRecord } {
  let record = recordFromState(state)
  const unrecorded = unrecordedSpecArtifacts(projectRoot, key, record)
  if (unrecorded.length === 0) {
    return { state, record }
  }
  for (const artifact of unrecorded) {
    record = appendSpecReviewIteration(
      record,
      nextSpecReviewIteration(record),
      iterationEntryFromArtifact(artifact),
    )
  }
  const healed: ProgressState = {
    ...state,
    data: { ...state.data, [SPEC_REVIEW_RECORD_KEY]: record },
    updated_at: new Date().toISOString(),
  }
  writeState(projectRoot, key, SPEC_REVIEW_STATE_COMMAND, healed)
  return { state: healed, record }
}

/**
 * Read the spec-review gate record for a state key, reconciled against the
 * dispatch artifact store (HARDEN-001 fallback, task 2.2): get_progress and
 * every gate consumer must never observe iteration N-1 while artifact N is
 * persisted. Returns null when no state file exists — reconciliation never
 * resurrects a state clear_progress removed, so completion stays idempotent.
 * The key is sanitized here (idempotent), so callers may pass either a
 * resolved state key or the raw branch — mirroring appendSpecReviewIteration
 * ToState's branch parameter and removing the branch/key asymmetry footgun.
 */
export function readSpecReviewGateRecord(
  projectRoot: string,
  key: string,
): SpecReviewGateRecord | null {
  const stateKey = sanitizeSlug(key)
  const state = readState(projectRoot, stateKey, SPEC_REVIEW_STATE_COMMAND)
  if (state === null) return null
  const { record } = healSpecReviewGateState(projectRoot, stateKey, state)
  return Object.keys(record.iterations).length > 0 ? record : null
}

/**
 * Read the uf.unleash pipeline state with the spec-review gate record
 * reconciled against the dispatch artifact store (task 2.2): the healed
 * record is written back so the two-store window closes durably, not just
 * in the returned view. Returns null when no state exists. The key is
 * sanitized here (idempotent), so callers may pass either a resolved state
 * key or the raw branch.
 */
export function reconcileSpecReviewGateState(
  projectRoot: string,
  key: string,
): ProgressState | null {
  const stateKey = sanitizeSlug(key)
  const state = readState(projectRoot, stateKey, SPEC_REVIEW_STATE_COMMAND)
  if (state === null) return null
  return healSpecReviewGateState(projectRoot, stateKey, state).state
}

/**
 * Append one spec-review iteration entry to the gate record (HARDEN-001,
 * task 2.1). Called by consolidate_dispatch after the dispatch artifact is
 * persisted, in the same finalization step, so the verdict reaches the
 * progress state before control returns to the model.
 *
 * Semantics:
 * - No prior state: one is created on demand. A standalone specs-mode
 *   review-council run has no uf.unleash pipeline state, yet its verdict
 *   is durable gate state for the branch. A state removed by
 *   clear_progress is never resurrected from artifacts — completion resets
 *   the gate.
 * - Existing state: any crash-window artifacts (persisted but unrecorded)
 *   are healed in first, keeping iteration numbers monotone with artifact
 *   chronology.
 * - A malformed existing record throws rather than being overwritten.
 *
 * @param projectRoot Project worktree holding `.uf/progress/` and the artifact store.
 * @param branch Branch the dispatch reviewed (envelope provenance); sanitized to the state key.
 * @param entry Iteration fields; `recorded_at` is stamped here.
 * @returns The written state, the full record, and the iteration number occupied.
 *
 * Idempotent by correlation (HARDEN-001): the consolidate tool persists the
 * artifact before this state write, so the healing pass may have just
 * recorded the consolidation this entry describes. When the entry's
 * correlation is already recorded, the recorded entry stands — one
 * consolidation is never appended as two iterations.
 */
export function appendSpecReviewIterationToState(
  projectRoot: string,
  branch: string,
  entry: Omit<SpecReviewIterationEntry, "recorded_at">,
): { state: ProgressState; record: SpecReviewGateRecord; iteration: number } {
  const key = sanitizeSlug(branch)
  const now = new Date().toISOString()
  const existing = readState(projectRoot, key, SPEC_REVIEW_STATE_COMMAND)

  let record: SpecReviewGateRecord
  let state: ProgressState
  if (existing === null) {
    record = { iterations: {} }
    state = {
      command: SPEC_REVIEW_STATE_COMMAND,
      branch: key,
      key,
      current_step: "spec-review",
      completed_steps: [],
      data: {},
      created_at: now,
      updated_at: now,
    }
  } else {
    record = recordFromState(existing)
    for (const artifact of unrecordedSpecArtifacts(projectRoot, key, record)) {
      record = appendSpecReviewIteration(
        record,
        nextSpecReviewIteration(record),
        iterationEntryFromArtifact(artifact),
      )
    }
    state = existing
  }

  // Idempotent by correlation (HARDEN-001): the consolidate tool persists
  // the artifact before this state write, so the healing pass above may
  // have just recorded the very consolidation this entry describes.
  // Re-appending would record one consolidation as two iterations; the
  // healed entry (same verdict, correlation, and findings, read from the
  // durable artifact) stands instead.
  const alreadyRecorded = Object.entries(record.iterations).find(
    ([, recorded]) => recorded.correlation_id === entry.correlation_id,
  )
  if (alreadyRecorded !== undefined) {
    const iteration = Number(alreadyRecorded[0])
    const healedState: ProgressState = {
      ...state,
      data: { ...state.data, [SPEC_REVIEW_RECORD_KEY]: record },
      updated_at: now,
    }
    writeState(projectRoot, key, SPEC_REVIEW_STATE_COMMAND, healedState)
    return { state: healedState, record, iteration }
  }

  const iteration = nextSpecReviewIteration(record)
  const withEntry = appendSpecReviewIteration(record, iteration, {
    ...entry,
    recorded_at: now,
  })
  const nextState: ProgressState = {
    ...state,
    data: { ...state.data, [SPEC_REVIEW_RECORD_KEY]: withEntry },
    updated_at: now,
  }
  writeState(projectRoot, key, SPEC_REVIEW_STATE_COMMAND, nextState)
  return { state: nextState, record: withEntry, iteration }
}

/**
 * Record the per-finding decision set adjudicating the LATEST spec-review
 * iteration (HARDEN-002, task 3.1). Crash-window artifacts are healed in
 * first so decisions always adjudicate the true latest iteration. Every
 * decision must name a finding the latest iteration carries, no finding may
 * be decided twice, and each decision must carry the fields its choice
 * requires (apply ⇒ fix_reference, override ⇒ non-empty reason) —
 * violations throw (fail closed; nothing is written). The set is scoped to
 * the iteration it adjudicates, so a later iteration append leaves it inert
 * for gate purposes while the resume probe can still see the adjudication
 * (HARDEN-006). The derived `overrides` audit list is written alongside the
 * set — empty when it contains no override decisions — and is never written
 * without a decision set (HARDEN-005(d)).
 *
 * @param projectRoot Project worktree holding `.uf/progress/`.
 * @param key State key or branch — sanitized here (idempotent).
 * @param decisions Full adjudication set for the latest iteration.
 * @returns The written state, the record, and the adjudicated iteration number.
 */
export function recordSpecReviewDecisionsToState(
  projectRoot: string,
  key: string,
  decisions: readonly SpecReviewDecision[],
): { state: ProgressState; record: SpecReviewGateRecord; iteration: number } {
  const stateKey = sanitizeSlug(key)
  const existing = readState(projectRoot, stateKey, SPEC_REVIEW_STATE_COMMAND)
  if (existing === null) {
    throw new Error(
      `no ${SPEC_REVIEW_STATE_COMMAND} state for key ${stateKey}: spec-review decisions require a recorded iteration`,
    )
  }
  const healed = healSpecReviewGateState(projectRoot, stateKey, existing)
  const latest = latestSpecReviewIteration(healed.record)
  if (latest === null) {
    throw new Error(
      `no spec-review iteration recorded for key ${stateKey}: decisions require a latest iteration`,
    )
  }
  const unknown = unknownSpecReviewDecisionIds(latest.entry, decisions)
  if (unknown.length > 0) {
    throw new Error(
      `unknown finding ids in spec-review decisions for iteration ${latest.iteration}: ${unknown.join(", ")}`,
    )
  }
  const duplicates = duplicateSpecReviewDecisionIds(decisions)
  if (duplicates.length > 0) {
    throw new Error(
      `duplicate finding ids in spec-review decisions: ${duplicates.join(", ")}`,
    )
  }
  const invalid = decisions.filter((decision) => !isValidSpecReviewDecision(decision))
  if (invalid.length > 0) {
    throw new Error(
      "invalid spec-review decisions (apply requires fix_reference; " +
        `override requires a non-empty reason): findings ${invalid
          .map((decision) => decision.finding_id)
          .join(", ")}`,
    )
  }
  const withDecisions: SpecReviewGateRecord = {
    ...healed.record,
    decisions: { iteration: String(latest.iteration), decisions: [...decisions] },
    overrides: deriveSpecReviewOverrides(decisions),
  }
  const nextState: ProgressState = {
    ...healed.state,
    data: { ...healed.state.data, [SPEC_REVIEW_RECORD_KEY]: withDecisions },
    updated_at: new Date().toISOString(),
  }
  writeState(projectRoot, stateKey, SPEC_REVIEW_STATE_COMMAND, nextState)
  return { state: nextState, record: withDecisions, iteration: latest.iteration }
}
