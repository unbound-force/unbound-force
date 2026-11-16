// ---------------------------------------------------------------------------
// Spec-review gate record — HARDEN-001 + HARDEN-002 coverage
// (harden-spec-review-gate, tasks 2.1–3.3).
//
// Spec: openspec/changes/harden-spec-review-gate/spec.md
//
// Covers the persistence contract: consolidate_dispatch writes the verdict
// to the uf.unleash progress state in the same finalization step that
// persists the artifact (task 2.1); get_progress reconciles the two stores
// so iteration N-1 is never returned while artifact N is persisted
// (task 2.2); REQUEST CHANGES is state-visible immediately after
// consolidation returns, the stale-store window heals, and APPROVE needs
// no decisions (task 2.3).
//
// Covers the gate contract (HARDEN-002): per-finding decisions are scoped
// to the iteration they adjudicate (task 3.1); the checkpoint tool records
// them via spec_review_decisions, refuses advancing steps while the latest
// REQUEST CHANGES iteration is undecided, pending its approving iteration,
// or paused, and permits advancement only through apply-then-APPROVE or
// audited overrides (tasks 3.2–3.3).
//
// Covers the re-dispatch guard (HARDEN-003): a specs-mode dispatch is
// refused while the latest recorded iteration stands undecided or paused
// (naming the iteration and verdict), the iteration cap is enforced before
// dispatch, and a decided latest permits the next iteration — the refusal
// precedes correlation minting, so nothing is left behind (tasks 4.1–4.2).
//
// Covers the orchestrator relay (HARDEN-004): relay_spec_review_gate
// computes the pause state from the healed record and the dispatch
// artifact — an undecided or paused gate relays the artifact-sourced
// findings (severity, description, file:line, correlation id) with
// gate_paused: true; a decided, approved, or absent gate reports false;
// a missing artifact never un-pauses (tasks 5.1/5.6). SYNTH-004
// (synthetic-compaction-continue-prompt-rec, task 5.1) widens the relay
// predicate: an escalated-advisory latest (APPROVE WITH ADVISORIES,
// persisted canonical decision ESCALATED) is treated like REQUEST
// CHANGES for relay purposes while the gate-status evaluation stays
// open, so isApprovingSpecVerdict retargets to plain APPROVE only.
//
// Covers the bypass-path regression (HARDEN-005): the incident replay —
// consolidate REQUEST CHANGES (8 findings, correlation 5355acba-…),
// three spec edits, compaction-summary return — asserts the record
// persists before the edits, checkpoint refuses advancement, get_progress
// stays fresh, and the relay renders all 8 findings from the artifact
// with no overrides:[] auto-write; edge cases pin child-text
// independence and the post-decision next-iteration re-dispatch (tasks
// 6.1–6.2). Resume adjudication (HARDEN-006) is covered by
// unleash-resume.test.ts (task 6.3).
// ---------------------------------------------------------------------------

import { randomUUID } from "node:crypto"
import { mkdir, readFile, rm, stat, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { dirname, join } from "node:path"

import type { ToolContext } from "@opencode-ai/plugin"
import { afterEach, describe, expect, it } from "vitest"

import {
  appendSpecReviewIterationToState,
  readSpecReviewGateRecord,
  reconcileSpecReviewGateState,
  recordSpecReviewDecisionsToState,
} from "../lib/progress-state.js"
import type { SiblingAcquisitionDependencies } from "../lib/review-dispatch-sibling-evidence.js"
import {
  appendSpecReviewIteration,
  decisionsForLatestIteration,
  deriveSpecReviewOverrides,
  duplicateSpecReviewDecisionIds,
  evaluateSpecReviewGate,
  hasApplyDecision,
  isApprovingSpecVerdict,
  isEscalatedAdvisorySpecVerdict,
  latestSpecReviewIteration,
  missingSpecReviewDecisions,
  nextSpecReviewIteration,
  parseSpecReviewGateRecord,
  SPEC_REVIEW_GATE_STEP_ALLOWLIST,
  SPEC_REVIEW_ITERATION_CAP,
  SPEC_REVIEW_RECORD_KEY,
  SPEC_REVIEW_STATE_COMMAND,
  specReviewFindingIds,
  unknownSpecReviewDecisionIds,
  type SpecReviewGateRecord,
  type SpecReviewIterationEntry,
} from "../lib/uf-workflow-types.js"
import {
  _getSessionCorrelationMap,
  createConsolidateDispatchTool,
  createDispatchReviewRunsTool,
  createFinalizationDependencies,
  specReviewDispatchGuard,
  type DispatchAgentRunDependencies,
} from "../plugins/review-dispatch/index.js"
import {
  createCheckpointTool,
  createGetProgressTool,
  createRelaySpecReviewGateTool,
  type GitExec,
  type ProgressState,
  type RelaySpecReviewGateData,
} from "../plugins/uf-workflow/index.js"

// ---------------------------------------------------------------------------
// Local fixtures (mirrors the consolidate-dispatch.test.ts harness shapes)
// ---------------------------------------------------------------------------

/** Build a valid PersistedRunData fixture for writing to disk. */
function runFixture(overrides: {
  agent: string
  findings?: Array<{
    severity: string
    category: string
    description: string
    root_cause: string
    file: string | null
    line: number | null
  }>
  proposals?: Array<{ information: string; tag: string; category?: string }>
}) {
  return {
    correlation_id: randomUUID(),
    run_id: randomUUID(),
    agent: overrides.agent,
    source: "explicit",
    sequence: 1,
    status: "success",
    started_at: "2026-11-14T10:00:00.000Z",
    finished_at: "2026-11-14T10:01:00.000Z",
    requested_model: "provider/standard",
    provider: "provider",
    model_id: "standard",
    variant: null,
    resolved_parent_model: null,
    resolved_parent_variant: null,
    reported_model: "provider/standard",
    model_mismatch: false,
    usage: { cost_usd: 0.5, tokens: { input: 1000, output: 500 } },
    error: null,
    workflow_verdict: null,
    findings: overrides.findings ?? [],
    proposals: overrides.proposals ?? [],
    text: "review output",
    read_only: true,
  }
}

/** Plan entry matching a runFixture run (identity = agent + source + sequence). */
function explicitRunPlan(agent: string) {
  return [
    {
      agent,
      decision: "include" as const,
      reason_code: "explicit",
      reason: "Explicitly included",
      source: "explicit" as const,
      sequence: 1,
      read_only: true,
      tier: "standard",
      model: "provider/standard",
      variant: null,
      validation_errors: [],
    },
  ]
}

function validInputContext() {
  return {
    kind: "pr" as const,
    pr_number: 1078,
    base_ref: "main",
    base_sha: "a".repeat(40),
    head_ref: "feature",
    head_sha: "b".repeat(40),
  }
}

function validChangeProfile() {
  return {
    kind: "diff" as const,
    lines: 100,
    files: 5,
    components: 2,
    security_sensitive: false,
    user_facing: true,
    categories: ["standard" as const],
    tier: "standard" as const,
  }
}

function validCoverage() {
  return { preflight_verdict: "PASS" as const, checks_total: 5, checks_passed: 5 }
}

/** Default session metadata; mode overrides to "specs" for gate fixtures. */
function defaultSessionMetadata(overrides?: Record<string, unknown>) {
  return {
    command: "review-council" as const,
    mode: "code" as const,
    full: false,
    input_context: validInputContext(),
    change_profile: validChangeProfile(),
    plan: explicitRunPlan("divisor-guard"),
    coverage: validCoverage(),
    ...overrides,
  }
}

function noopContext(): ToolContext {
  return {
    sessionID: "spec-review-gate-session",
    messageID: "msg-spec-review-gate",
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

// Track temp directories for cleanup.
const cleanupDirs: string[] = []

afterEach(async () => {
  _getSessionCorrelationMap().clear()
  for (const dir of cleanupDirs) {
    try {
      await rm(dir, { recursive: true, force: true })
    } catch {
      // best-effort
    }
  }
  cleanupDirs.length = 0
})

/** Fresh scratch project root; the label names the fixture family. */
async function scratchRoot(label: string): Promise<string> {
  const projectRoot = join(tmpdir(), `spec-review-gate-${label}-${randomUUID()}`)
  await mkdir(projectRoot, { recursive: true })
  cleanupDirs.push(projectRoot)
  return projectRoot
}

/** Write run fixtures + session metadata and register the session correlation. */
async function setupDispatchSession(
  runs: ReturnType<typeof runFixture>[],
  metadata?: Record<string, unknown>,
): Promise<string> {
  const correlationId = randomUUID()
  const dir = join(tmpdir(), "opencode", `dispatch-${correlationId}`)
  await mkdir(dir, { recursive: true })
  cleanupDirs.push(dir)
  _getSessionCorrelationMap().set("spec-review-gate-session", correlationId)

  await writeFile(
    join(dir, "session-metadata.json"),
    JSON.stringify(defaultSessionMetadata(metadata), null, 2),
    "utf8",
  )
  for (const run of runs) {
    await writeFile(join(dir, `run-${run.agent}.json`), JSON.stringify(run, null, 2), "utf8")
  }
  return correlationId
}

/**
 * Hand-build a minimal but valid review-dispatch artifact envelope — the
 * crash-window fixture: an artifact the finalizer persisted whose progress
 * state write never happened.
 */
function artifactEnvelope(overrides: {
  branch: string
  mode?: string
  verdict: string
  correlationId: string
  findingCount?: number
  timestamp?: string
}): string {
  const findingCount = overrides.findingCount ?? 0
  const findings = Array.from({ length: findingCount }, (_, i) => ({
    severity: i === 0 ? "HIGH" : "MEDIUM",
    category: "gate-semantics",
    description: `Replayed finding ${i + 1}`,
    root_cause: `Replayed cause ${i + 1}`,
    file: `file-${i + 1}.ts`,
    line: i + 1,
  }))
  return JSON.stringify({
    artifact_type: "review-dispatch",
    schema_version: "1.0.0",
    timestamp: overrides.timestamp ?? "2026-11-14T00:00:00.000Z",
    context: {
      branch: overrides.branch,
      commit: "b".repeat(40),
      correlation_id: overrides.correlationId,
      workflow_id: "uf.review-council",
    },
    payload: {
      command: "review-council",
      mode: overrides.mode ?? "specs",
      full: false,
      input_context: validInputContext(),
      change_profile: validChangeProfile(),
      plan_version: 1,
      plan: [],
      runs: [],
      coverage: validCoverage(),
      findings,
      advisories: [],
      verdict: overrides.verdict,
      verdict_reason: "replayed",
      run_counts: { total: 1, success: 1, failed: 0, skipped: 0, budget_skipped: 0, limit_skipped: 0, cancelled: 0 },
      workflow_result: { kind: "council", value: overrides.verdict },
      correlation_id: overrides.correlationId,
    },
  })
}

/** The git exec is never called — every invocation passes an explicit key. */
const unusedExec: GitExec = (): string => {
  throw new Error("unexpected git call")
}

/** Invoke a tool factory result with typed args and parse JSON output. */
async function invokeTool<T = unknown>(
  toolInstance: unknown,
  args: Record<string, unknown>,
): Promise<{ status: string; data?: T; message?: string; code?: string }> {
  const result = await (
    toolInstance as {
      execute: (args: Record<string, unknown>, ctx: { sessionID: string }) => Promise<{ output: string }>
    }
  ).execute(args, { sessionID: "spec-review-gate-session" })
  return JSON.parse(result.output) as { status: string; data?: T; message?: string; code?: string }
}

/** Invoke the dispatch tool and read its structured metadata channel. */
async function dispatchMetadata(
  toolInstance: unknown,
  args: Record<string, unknown>,
): Promise<{
  status: string
  correlation_id?: string
  message?: string
  runs_completed: number
  runs_failed: number
  runs_skipped: number
}> {
  const result = await (
    toolInstance as {
      execute: (
        args: Record<string, unknown>,
        ctx: { sessionID: string },
      ) => Promise<{
        metadata: {
          status: string
          correlation_id?: string
          message?: string
          runs_completed: number
          runs_failed: number
          runs_skipped: number
        }
      }>
    }
  ).execute(args, { sessionID: "spec-review-gate-session" })
  return result.metadata
}

/** Build one iteration entry fixture. */
function iterationEntry(overrides?: Partial<SpecReviewIterationEntry>): SpecReviewIterationEntry {
  return {
    verdict: "REQUEST CHANGES",
    correlation_id: randomUUID(),
    finding_count: 2,
    finding_ids: ["1", "2"],
    recorded_at: "2026-11-14T00:00:00.000Z",
    ...overrides,
  }
}

/** Read the uf.unleash state file directly (post-condition assertions). */
async function readStateFile(projectRoot: string, key: string): Promise<ProgressState | null> {
  try {
    const raw = await readFile(
      join(projectRoot, ".uf", "progress", key, `${SPEC_REVIEW_STATE_COMMAND}.json`),
      "utf8",
    )
    return JSON.parse(raw) as ProgressState
  } catch {
    return null
  }
}

// ---------------------------------------------------------------------------
// Pure record helpers (single normative shape, uf-workflow-types.ts)
// ---------------------------------------------------------------------------

describe("spec-review gate record helpers (HARDEN-001 shape)", () => {
  it("isApprovingSpecVerdict passes only a plain APPROVE (SYNTH-004)", () => {
    expect(isApprovingSpecVerdict("APPROVE")).toBe(true)
    // SYNTH-004 reverses the old APPROVE-family treatment: the finalizer
    // maps every APPROVE WITH ADVISORIES consolidation to the canonical
    // decision ESCALATED, so AWA is escalated (relayed), not approving.
    expect(isApprovingSpecVerdict("APPROVE WITH ADVISORIES")).toBe(false)
    expect(isApprovingSpecVerdict("REQUEST CHANGES")).toBe(false)
    expect(isApprovingSpecVerdict("INCONCLUSIVE")).toBe(false)
    expect(isApprovingSpecVerdict("UNAVAILABLE")).toBe(false)
  })

  it("isEscalatedAdvisorySpecVerdict passes only APPROVE WITH ADVISORIES (SYNTH-004)", () => {
    expect(isEscalatedAdvisorySpecVerdict("APPROVE WITH ADVISORIES")).toBe(true)
    expect(isEscalatedAdvisorySpecVerdict("APPROVE")).toBe(false)
    expect(isEscalatedAdvisorySpecVerdict("REQUEST CHANGES")).toBe(false)
    expect(isEscalatedAdvisorySpecVerdict("INCONCLUSIVE")).toBe(false)
    expect(isEscalatedAdvisorySpecVerdict("UNAVAILABLE")).toBe(false)
  })

  it("specReviewFindingIds derives 1-based identities", () => {
    expect(specReviewFindingIds(0)).toEqual([])
    expect(specReviewFindingIds(3)).toEqual(["1", "2", "3"])
  })

  it("parseSpecReviewGateRecord round-trips a valid record", () => {
    const record: SpecReviewGateRecord = {
      iterations: { "1": iterationEntry() },
      decisions: {
        iteration: "1",
        decisions: [
          { finding_id: "1", decision: "apply", fix_reference: "tasks.md 3.1" },
          { finding_id: "2", decision: "override", reason: "accepted risk" },
        ],
      },
      overrides: [{ finding_id: "2", reason: "accepted risk" }],
    }
    expect(parseSpecReviewGateRecord(JSON.parse(JSON.stringify(record)))).toEqual(record)
  })

  it("parseSpecReviewGateRecord rejects malformed shapes", () => {
    expect(parseSpecReviewGateRecord(null)).toBeNull()
    expect(parseSpecReviewGateRecord("nope")).toBeNull()
    expect(parseSpecReviewGateRecord({ iterations: "nope" })).toBeNull()
    expect(parseSpecReviewGateRecord({ iterations: { x: iterationEntry() } })).toBeNull()
    expect(parseSpecReviewGateRecord({ iterations: { "1": { verdict: "REQUEST CHANGES" } } })).toBeNull()
    expect(parseSpecReviewGateRecord({ iterations: {}, decisions: "nope" })).toBeNull()
  })

  it("a malformed decision is dropped, leaving its finding undecided (fail closed)", () => {
    // apply without fix_reference is malformed at parse time — the decision
    // disappears, so finding 1 counts as missing in the gate check.
    const parsed = parseSpecReviewGateRecord({
      iterations: { "1": iterationEntry() },
      decisions: {
        iteration: "1",
        decisions: [
          { finding_id: "1", decision: "apply" },
          { finding_id: "2", decision: "override", reason: "accepted risk" },
        ],
      },
    })
    expect(parsed).not.toBeNull()
    const missing = missingSpecReviewDecisions(
      parsed!.iterations["1"],
      parsed!.decisions!.decisions,
    )
    expect(missing).toEqual(["1"])
  })

  it("appendSpecReviewIteration appends and refuses overwrite", () => {
    const record: SpecReviewGateRecord = { iterations: { "1": iterationEntry() } }
    const appended = appendSpecReviewIteration(record, 2, iterationEntry({ verdict: "APPROVE" }))
    expect(Object.keys(appended.iterations)).toEqual(["1", "2"])
    expect(() => appendSpecReviewIteration(appended, 2, iterationEntry())).toThrow(
      /already recorded/,
    )
  })

  it("appendSpecReviewIteration preserves the decision set scoped to its iteration", () => {
    const record: SpecReviewGateRecord = {
      iterations: { "1": iterationEntry() },
      decisions: {
        iteration: "1",
        decisions: [{ finding_id: "1", decision: "apply", fix_reference: "tasks.md 3.1" }],
      },
      overrides: [],
    }
    const appended = appendSpecReviewIteration(record, 2, iterationEntry({ verdict: "REQUEST CHANGES" }))
    // The set survives the append, still scoped to iteration 1 — inert for
    // gate purposes (iteration 2's findings stay undecided) while remaining
    // visible to the HARDEN-006 resume probe.
    expect(appended.decisions).toEqual({
      iteration: "1",
      decisions: [{ finding_id: "1", decision: "apply", fix_reference: "tasks.md 3.1" }],
    })
    expect(appended.overrides).toEqual([])
    expect(decisionsForLatestIteration(appended)).toEqual([])
  })

  it("latestSpecReviewIteration and nextSpecReviewIteration are monotone", () => {
    expect(latestSpecReviewIteration({ iterations: {} })).toBeNull()
    expect(nextSpecReviewIteration({ iterations: {} })).toBe(1)
    const record: SpecReviewGateRecord = {
      iterations: {
        "1": iterationEntry(),
        "3": iterationEntry({ verdict: "APPROVE" }),
      },
    }
    expect(latestSpecReviewIteration(record)?.iteration).toBe(3)
    expect(nextSpecReviewIteration(record)).toBe(4)
  })

  it("decision validation enforces the conditional fields (HARDEN-002 shape)", () => {
    expect(hasApplyDecision([{ finding_id: "1", decision: "pause" }])).toBe(false)
    expect(hasApplyDecision([{ finding_id: "1", decision: "apply", fix_reference: "x" }])).toBe(true)
    expect(
      deriveSpecReviewOverrides([
        { finding_id: "1", decision: "apply", fix_reference: "x" },
        { finding_id: "2", decision: "override", reason: "accepted risk" },
      ]),
    ).toEqual([{ finding_id: "2", reason: "accepted risk" }])
    expect(deriveSpecReviewOverrides(undefined)).toEqual([])
  })

  it("unknown and duplicate decision ids are surfaced", () => {
    const entry = iterationEntry()
    expect(unknownSpecReviewDecisionIds(entry, [{ finding_id: "9", decision: "pause" }])).toEqual(["9"])
    expect(
      duplicateSpecReviewDecisionIds([
        { finding_id: "1", decision: "pause" },
        { finding_id: "1", decision: "pause" },
      ]),
    ).toEqual(["1"])
  })

  it("the iteration cap is three (HARDEN-003 constant)", () => {
    expect(SPEC_REVIEW_ITERATION_CAP).toBe(3)
  })
})

// ---------------------------------------------------------------------------
// appendSpecReviewIterationToState (task 2.1 — durable write semantics)
// ---------------------------------------------------------------------------

describe("appendSpecReviewIterationToState (HARDEN-001, task 2.1)", () => {
  it("creates the uf.unleash state on demand with the recorded iteration", async () => {
    const root = await scratchRoot("append-create")
    const correlationId = randomUUID()
    const { record, iteration } = appendSpecReviewIterationToState(root, "opsx/gate-fixture", {
      verdict: "REQUEST CHANGES",
      correlation_id: correlationId,
      finding_count: 2,
      finding_ids: ["1", "2"],
    })
    expect(iteration).toBe(1)
    expect(record.iterations["1"].verdict).toBe("REQUEST CHANGES")
    expect(record.iterations["1"].correlation_id).toBe(correlationId)

    const state = await readStateFile(root, "opsx-gate-fixture")
    expect(state).not.toBeNull()
    expect(state!.command).toBe(SPEC_REVIEW_STATE_COMMAND)
    expect(state!.data[SPEC_REVIEW_RECORD_KEY]).toEqual(record)
  })

  it("appends monotone iterations for the same branch", async () => {
    const root = await scratchRoot("append-monotone")
    const first = appendSpecReviewIterationToState(root, "opsx/gate-fixture", {
      verdict: "REQUEST CHANGES",
      correlation_id: randomUUID(),
      finding_count: 1,
      finding_ids: ["1"],
    })
    const second = appendSpecReviewIterationToState(root, "opsx/gate-fixture", {
      verdict: "APPROVE",
      correlation_id: randomUUID(),
      finding_count: 0,
      finding_ids: [],
    })
    expect(first.iteration).toBe(1)
    expect(second.iteration).toBe(2)
    expect(Object.keys(second.record.iterations)).toEqual(["1", "2"])
  })

  it("throws on a malformed existing record rather than overwriting it", async () => {
    const root = await scratchRoot("append-malformed")
    appendSpecReviewIterationToState(root, "opsx/gate-fixture", {
      verdict: "REQUEST CHANGES",
      correlation_id: randomUUID(),
      finding_count: 0,
      finding_ids: [],
    })
    // Tamper with the record shape directly.
    const statePath = join(root, ".uf", "progress", "opsx-gate-fixture", `${SPEC_REVIEW_STATE_COMMAND}.json`)
    const raw = JSON.parse(await readFile(statePath, "utf8")) as ProgressState
    raw.data[SPEC_REVIEW_RECORD_KEY] = { iterations: { "1": { verdict: 42 } } }
    await writeFile(statePath, JSON.stringify(raw), "utf8")

    expect(() =>
      appendSpecReviewIterationToState(root, "opsx/gate-fixture", {
        verdict: "APPROVE",
        correlation_id: randomUUID(),
        finding_count: 0,
        finding_ids: [],
      }),
    ).toThrow(/malformed spec-review record/)
  })

  it("heals a crash-window artifact before appending, keeping numbers monotone", async () => {
    const root = await scratchRoot("append-heal")
    // Iteration 1 recorded normally.
    appendSpecReviewIterationToState(root, "opsx/gate-fixture", {
      verdict: "REQUEST CHANGES",
      correlation_id: "11111111-1111-4111-8111-111111111111",
      finding_count: 1,
      finding_ids: ["1"],
    })
    // Iteration 2's artifact persisted, but its state write never happened.
    const storeDir = join(root, ".uf", "artifacts", "dispatch")
    await mkdir(storeDir, { recursive: true })
    await writeFile(
      join(storeDir, "22222222-2222-4222-8222-222222222222.json"),
      artifactEnvelope({
        branch: "opsx/gate-fixture",
        verdict: "REQUEST CHANGES",
        correlationId: "22222222-2222-4222-8222-222222222222",
        findingCount: 3,
      }),
      "utf8",
    )
    // The next consolidation appends as iteration 3 — the healed artifact
    // occupies iteration 2 first.
    const { iteration, record } = appendSpecReviewIterationToState(root, "opsx/gate-fixture", {
      verdict: "APPROVE",
      correlation_id: "33333333-3333-4333-8333-333333333333",
      finding_count: 0,
      finding_ids: [],
    })
    expect(iteration).toBe(3)
    expect(record.iterations["2"].correlation_id).toBe("22222222-2222-4222-8222-222222222222")
    expect(record.iterations["2"].finding_ids).toEqual(["1", "2", "3"])
  })
})

// ---------------------------------------------------------------------------
// Two-store reconciliation (task 2.2 — never return N-1 when N is persisted)
// ---------------------------------------------------------------------------

describe("spec-review gate reconciliation (HARDEN-001, task 2.2)", () => {
  it("replays the 5355acba crash window: get_progress never returns iteration N-1 when artifact N is persisted", async () => {
    const root = await scratchRoot("replay-5355acba")
    // Iteration 1 recorded in state; iteration 2's artifact (the
    // 5355acba REQUEST CHANGES from issue #1078's replay) persisted but
    // never recorded.
    appendSpecReviewIterationToState(root, "opsx/gate-fixture", {
      verdict: "REQUEST CHANGES",
      correlation_id: "44444444-4444-4444-8444-444444444444",
      finding_count: 1,
      finding_ids: ["1"],
    })
    const storeDir = join(root, ".uf", "artifacts", "dispatch")
    await mkdir(storeDir, { recursive: true })
    const replayedCorrelation = "5355acba-62f4-4859-8775-7294b3e8702c"
    await writeFile(
      join(storeDir, `${replayedCorrelation}.json`),
      artifactEnvelope({
        branch: "opsx/gate-fixture",
        verdict: "REQUEST CHANGES",
        correlationId: replayedCorrelation,
        findingCount: 8,
      }),
      "utf8",
    )

    // get_progress must reflect the persisted artifact immediately.
    const result = await invokeTool<ProgressState>(
      createGetProgressTool(unusedExec, root),
      { command: SPEC_REVIEW_STATE_COMMAND, key: "opsx-gate-fixture" },
    )
    expect(result.status).toBe("ok")
    const record = result.data?.data?.[SPEC_REVIEW_RECORD_KEY] as SpecReviewGateRecord
    expect(Object.keys(record.iterations)).toEqual(["1", "2"])
    expect(record.iterations["2"]).toMatchObject({
      verdict: "REQUEST CHANGES",
      correlation_id: replayedCorrelation,
      finding_count: 8,
    })
    expect(record.iterations["2"].finding_ids).toEqual([
      "1", "2", "3", "4", "5", "6", "7", "8",
    ])

    // The healed record is written back — the window closes durably, not
    // just in the returned view.
    const healedState = await readStateFile(root, "opsx-gate-fixture")
    const healedRecord = healedState?.data[SPEC_REVIEW_RECORD_KEY] as SpecReviewGateRecord
    expect(healedRecord.iterations["2"].correlation_id).toBe(replayedCorrelation)

    // The record reader agrees with the reconciled state.
    const direct = readSpecReviewGateRecord(root, "opsx-gate-fixture")
    expect(latestSpecReviewIteration(direct!)?.entry.correlation_id).toBe(replayedCorrelation)
  })

  it("ignores non-specs artifacts and other branches", async () => {
    const root = await scratchRoot("reconcile-scope")
    appendSpecReviewIterationToState(root, "opsx/gate-fixture", {
      verdict: "REQUEST CHANGES",
      correlation_id: "55555555-5555-4555-8555-555555555555",
      finding_count: 1,
      finding_ids: ["1"],
    })
    const storeDir = join(root, ".uf", "artifacts", "dispatch")
    await mkdir(storeDir, { recursive: true })
    await writeFile(
      join(storeDir, "66666666-6666-4666-8666-666666666666.json"),
      artifactEnvelope({
        branch: "opsx/gate-fixture",
        mode: "code",
        verdict: "REQUEST CHANGES",
        correlationId: "66666666-6666-4666-8666-666666666666",
        findingCount: 4,
      }),
      "utf8",
    )
    await writeFile(
      join(storeDir, "77777777-7777-4777-8777-777777777777.json"),
      artifactEnvelope({
        branch: "opsx/other-branch",
        verdict: "REQUEST CHANGES",
        correlationId: "77777777-7777-4777-8777-777777777777",
        findingCount: 5,
      }),
      "utf8",
    )

    const record = readSpecReviewGateRecord(root, "opsx-gate-fixture")
    expect(Object.keys(record!.iterations)).toEqual(["1"])
  })

  it("never resurrects a state clear_progress removed", async () => {
    const root = await scratchRoot("reconcile-cleared")
    const storeDir = join(root, ".uf", "artifacts", "dispatch")
    await mkdir(storeDir, { recursive: true })
    await writeFile(
      join(storeDir, "88888888-8888-4888-8888-888888888888.json"),
      artifactEnvelope({
        branch: "opsx/gate-fixture",
        verdict: "REQUEST CHANGES",
        correlationId: "88888888-8888-4888-8888-888888888888",
        findingCount: 1,
      }),
      "utf8",
    )
    // No uf.unleash state exists (cleared after completion): the read
    // returns null rather than resurrecting gate state from artifacts.
    expect(readSpecReviewGateRecord(root, "opsx-gate-fixture")).toBeNull()
    const state = await readStateFile(root, "opsx-gate-fixture")
    expect(state).toBeNull()
    expect(reconcileSpecReviewGateState(root, "opsx-gate-fixture")).toBeNull()
  })
})

// ---------------------------------------------------------------------------
// consolidate_dispatch persistence (task 2.3 — verdict state-visible
// before control returns to the model)
// ---------------------------------------------------------------------------

describe("consolidate_dispatch persists the spec-review verdict (HARDEN-001, task 2.3)", () => {
  it("records REQUEST CHANGES to progress state immediately after the tool returns", async () => {
    const root = await scratchRoot("consolidate-rc")
    const guardRun = runFixture({
      agent: "divisor-guard",
      findings: [
        {
          severity: "HIGH",
          category: "gate-semantics",
          description: "The gate can be bypassed",
          root_cause: "Missing per-finding decisions",
          file: "spec.md",
          line: 12,
        },
      ],
    })
    const correlationId = await setupDispatchSession([guardRun], {
      mode: "specs",
      plan: explicitRunPlan("divisor-guard"),
    })

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(root))
    const result = await tool.execute(
      {
        provenance: {
          branch: "opsx/gate-fixture",
          commit: "b".repeat(40),
          workflow_id: "uf.review-council",
        },
      },
      noopContext(),
    )
    const finalization = JSON.parse(result.output) as {
      status: string
      operation_verdict: string
      correlation_id: string
    }
    expect(finalization.status).toBe("success")
    expect(finalization.operation_verdict).toBe("REQUEST CHANGES")
    expect(finalization.correlation_id).toBe(correlationId)

    // The verdict is durable in the progress state the moment the tool
    // returned — before any edit the model might make.
    const record = readSpecReviewGateRecord(root, "opsx-gate-fixture")
    expect(record).not.toBeNull()
    const latest = latestSpecReviewIteration(record!)
    expect(latest?.iteration).toBe(1)
    expect(latest?.entry.verdict).toBe("REQUEST CHANGES")
    expect(latest?.entry.correlation_id).toBe(correlationId)
    expect(latest?.entry.finding_count).toBe(1)
    expect(latest?.entry.finding_ids).toEqual(["1"])

    // get_progress reflects the latest iteration immediately.
    const progress = await invokeTool<ProgressState>(
      createGetProgressTool(unusedExec, root),
      { command: SPEC_REVIEW_STATE_COMMAND, key: "opsx-gate-fixture" },
    )
    const viaGet = progress.data?.data?.[SPEC_REVIEW_RECORD_KEY] as SpecReviewGateRecord
    expect(latestSpecReviewIteration(viaGet)?.entry.correlation_id).toBe(correlationId)
  })

  it("records APPROVE iterations; no decisions are required to advance", async () => {
    const root = await scratchRoot("consolidate-approve")
    const guardRun = runFixture({ agent: "divisor-guard" })
    const correlationId = await setupDispatchSession([guardRun], {
      mode: "specs",
      plan: explicitRunPlan("divisor-guard"),
    })

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(root))
    const result = await tool.execute(
      {
        provenance: {
          branch: "opsx/gate-fixture",
          commit: "b".repeat(40),
          workflow_id: "uf.review-council",
        },
      },
      noopContext(),
    )
    const finalization = JSON.parse(result.output) as { operation_verdict: string }
    expect(finalization.operation_verdict).toBe("APPROVE")

    const record = readSpecReviewGateRecord(root, "opsx-gate-fixture")
    const latest = latestSpecReviewIteration(record!)
    // APPROVE passes without per-finding decisions (HARDEN-002).
    expect(latest?.entry.verdict).toBe("APPROVE")
    expect(isApprovingSpecVerdict(latest!.entry.verdict)).toBe(true)
    expect(latest?.entry.correlation_id).toBe(correlationId)
    expect(record!.decisions).toBeUndefined()
  })

  it("writes no gate record for code-mode reviews", async () => {
    const root = await scratchRoot("consolidate-code")
    const guardRun = runFixture({ agent: "divisor-guard" })
    await setupDispatchSession([guardRun])

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(root))
    const result = await tool.execute(
      {
        provenance: {
          branch: "opsx/gate-fixture",
          commit: "b".repeat(40),
          workflow_id: "uf.review-council",
        },
      },
      noopContext(),
    )
    const finalization = JSON.parse(result.output) as { status: string }
    expect(finalization.status).toBe("success")
    expect(readSpecReviewGateRecord(root, "opsx-gate-fixture")).toBeNull()
  })

  it("surfaces a warning when the state write fails after the artifact persisted", async () => {
    const root = await scratchRoot("consolidate-write-failure")
    const guardRun = runFixture({ agent: "divisor-guard" })
    await setupDispatchSession([guardRun], {
      mode: "specs",
      plan: explicitRunPlan("divisor-guard"),
    })
    // .uf exists with the artifact store, but .uf/progress is a regular
    // file: the artifact write succeeds, the state write cannot.
    await mkdir(join(root, ".uf", "artifacts", "dispatch"), { recursive: true })
    await writeFile(join(root, ".uf", "progress"), "not a directory", "utf8")

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(root))
    const result = await tool.execute(
      {
        provenance: {
          branch: "opsx/gate-fixture",
          commit: "b".repeat(40),
          workflow_id: "uf.review-council",
        },
      },
      noopContext(),
    )
    const finalization = JSON.parse(result.output) as {
      status: string
      warnings?: Array<{ kind: string; detail: string }>
    }
    // The artifact persisted; only the state write failed, and the failure
    // is reported rather than silent.
    expect(finalization.status).toBe("success")
    expect(finalization.warnings).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ kind: "spec_review_state_write_failed" }),
      ]),
    )
    // The artifact store entry exists (the verdict is durable despite the
    // state failure) — reconciliation heals the record once the state path
    // is writable again.
    const storeDir = join(root, ".uf", "artifacts", "dispatch")
    const entries = (await stat(storeDir)).isDirectory()
    expect(entries).toBe(true)
  })
})

// ---------------------------------------------------------------------------
// Spec-review gate evaluation — HARDEN-002 pure semantics (task 3.1)
// ---------------------------------------------------------------------------

describe("spec-review gate evaluation (HARDEN-002)", () => {
  it("evaluateSpecReviewGate opens on null record, empty record, and approving latest", () => {
    expect(evaluateSpecReviewGate(null)).toEqual({ kind: "open" })
    expect(evaluateSpecReviewGate({ iterations: {} })).toEqual({ kind: "open" })
    expect(
      evaluateSpecReviewGate({ iterations: { "1": iterationEntry({ verdict: "APPROVE" }) } }),
    ).toEqual({ kind: "open" })
  })

  it("evaluateSpecReviewGate stays open on an escalated-advisory latest (SYNTH-004)", () => {
    // SYNTH-004 relays APPROVE WITH ADVISORIES through
    // relay_spec_review_gate like REQUEST CHANGES, but deliberately not
    // through the HARDEN-002 per-finding machinery: the gate status stays
    // open so checkpoint's allowlist hold, the resume probe, and the
    // dispatch guard keep their pre-SYNTH-004 behavior — the advisory
    // hold is the relay predicate plus the advancement-step gate.
    expect(
      evaluateSpecReviewGate({
        iterations: { "1": iterationEntry({ verdict: "APPROVE WITH ADVISORIES" }) },
      }),
    ).toEqual({ kind: "open" })
  })

  it("non-REQUEST-CHANGES verdicts never hold the gate (no adjudicable findings)", () => {
    expect(
      evaluateSpecReviewGate({ iterations: { "1": iterationEntry({ verdict: "INCONCLUSIVE" }) } }),
    ).toEqual({ kind: "open" })
    expect(
      evaluateSpecReviewGate({
        iterations: {
          "1": iterationEntry({ verdict: "UNAVAILABLE", finding_count: 0, finding_ids: [] }),
        },
      }),
    ).toEqual({ kind: "open" })
  })

  it("missing or partial decisions hold the gate with the undecided ids", () => {
    expect(evaluateSpecReviewGate({ iterations: { "1": iterationEntry() } })).toEqual({
      kind: "missing_decisions",
      iteration: 1,
      verdict: "REQUEST CHANGES",
      undecided: ["1", "2"],
    })
    expect(
      evaluateSpecReviewGate({
        iterations: { "1": iterationEntry() },
        decisions: {
          iteration: "1",
          decisions: [{ finding_id: "1", decision: "apply", fix_reference: "tasks.md 3.1" }],
        },
      }),
    ).toEqual({
      kind: "missing_decisions",
      iteration: 1,
      verdict: "REQUEST CHANGES",
      undecided: ["2"],
    })
  })

  it("a decision set scoped to an earlier iteration does not decide the latest", () => {
    expect(
      evaluateSpecReviewGate({
        iterations: {
          "1": iterationEntry(),
          "2": iterationEntry({ verdict: "REQUEST CHANGES" }),
        },
        decisions: {
          iteration: "1",
          decisions: [
            { finding_id: "1", decision: "override", reason: "accepted risk" },
            { finding_id: "2", decision: "override", reason: "out of scope" },
          ],
        },
      }),
    ).toEqual({
      kind: "missing_decisions",
      iteration: 2,
      verdict: "REQUEST CHANGES",
      undecided: ["1", "2"],
    })
  })

  it("a pause decision outranks apply and holds the gate for the exit path", () => {
    expect(
      evaluateSpecReviewGate({
        iterations: { "1": iterationEntry() },
        decisions: {
          iteration: "1",
          decisions: [
            { finding_id: "1", decision: "apply", fix_reference: "tasks.md 3.1" },
            { finding_id: "2", decision: "pause" },
          ],
        },
      }),
    ).toEqual({
      kind: "paused",
      iteration: 1,
      verdict: "REQUEST CHANGES",
      paused: ["2"],
    })
  })

  it("complete apply decisions await the approving iteration", () => {
    expect(
      evaluateSpecReviewGate({
        iterations: { "1": iterationEntry() },
        decisions: {
          iteration: "1",
          decisions: [
            { finding_id: "1", decision: "apply", fix_reference: "tasks.md 3.1" },
            { finding_id: "2", decision: "apply", fix_reference: "tasks.md 3.2" },
          ],
        },
      }),
    ).toEqual({
      kind: "pending_approving_iteration",
      iteration: 1,
      verdict: "REQUEST CHANGES",
    })
  })

  it("an all-override set is advanceable with the derived overrides", () => {
    expect(
      evaluateSpecReviewGate({
        iterations: { "1": iterationEntry() },
        decisions: {
          iteration: "1",
          decisions: [
            { finding_id: "1", decision: "override", reason: "accepted risk" },
            { finding_id: "2", decision: "override", reason: "out of scope" },
          ],
        },
      }),
    ).toEqual({
      kind: "advanceable",
      iteration: 1,
      verdict: "REQUEST CHANGES",
      overrides: [
        { finding_id: "1", reason: "accepted risk" },
        { finding_id: "2", reason: "out of scope" },
      ],
    })
  })

  it("the step allowlist covers gate machinery and no advancing step", () => {
    expect(SPEC_REVIEW_GATE_STEP_ALLOWLIST).toEqual([
      "branch-safety-gate",
      "tier-routing",
      "resumability-detection",
      "clarify",
      "plan",
      "tasks",
      "quick-fix-budget",
    ])
    expect(SPEC_REVIEW_GATE_STEP_ALLOWLIST).not.toContain("spec-review")
    expect(SPEC_REVIEW_GATE_STEP_ALLOWLIST).not.toContain("implement")
    expect(SPEC_REVIEW_GATE_STEP_ALLOWLIST).not.toContain("code-review")
  })
})

// ---------------------------------------------------------------------------
// Spec-review decisions recording (HARDEN-002, task 3.1)
// ---------------------------------------------------------------------------

describe("recordSpecReviewDecisionsToState (HARDEN-002)", () => {
  it("records a scoped decision set with derived overrides alongside the verdict", async () => {
    const root = await scratchRoot("decisions-record")
    appendSpecReviewIterationToState(root, "opsx/gate-fixture", {
      verdict: "REQUEST CHANGES",
      correlation_id: "77777777-7777-4777-8777-777777777777",
      finding_count: 2,
      finding_ids: ["1", "2"],
    })

    const { record, iteration } = recordSpecReviewDecisionsToState(root, "opsx/gate-fixture", [
      { finding_id: "1", decision: "apply", fix_reference: "tasks.md 3.1" },
      { finding_id: "2", decision: "override", reason: "accepted risk" },
    ])

    expect(iteration).toBe(1)
    expect(record.decisions).toEqual({
      iteration: "1",
      decisions: [
        { finding_id: "1", decision: "apply", fix_reference: "tasks.md 3.1" },
        { finding_id: "2", decision: "override", reason: "accepted risk" },
      ],
    })
    expect(record.overrides).toEqual([{ finding_id: "2", reason: "accepted risk" }])

    const persisted = await readStateFile(root, "opsx-gate-fixture")
    const persistedRecord = persisted?.data[SPEC_REVIEW_RECORD_KEY] as SpecReviewGateRecord
    expect(persistedRecord.decisions?.iteration).toBe("1")
    expect(persistedRecord.overrides).toEqual([{ finding_id: "2", reason: "accepted risk" }])
  })

  it("writes an empty overrides array when the set holds no override decisions", async () => {
    const root = await scratchRoot("decisions-empty-overrides")
    appendSpecReviewIterationToState(root, "opsx/gate-fixture", {
      verdict: "REQUEST CHANGES",
      correlation_id: "88888888-8888-4888-8888-888888888888",
      finding_count: 1,
      finding_ids: ["1"],
    })

    const { record } = recordSpecReviewDecisionsToState(root, "opsx/gate-fixture", [
      { finding_id: "1", decision: "apply", fix_reference: "spec.md HARDEN-002" },
    ])

    expect(record.overrides).toEqual([])
  })

  it("rejects unknown ids, duplicates, invalid shapes, and missing state (fail closed)", async () => {
    const root = await scratchRoot("decisions-rejections")
    appendSpecReviewIterationToState(root, "opsx/gate-fixture", {
      verdict: "REQUEST CHANGES",
      correlation_id: "99999999-9999-4999-8999-999999999999",
      finding_count: 2,
      finding_ids: ["1", "2"],
    })

    expect(() =>
      recordSpecReviewDecisionsToState(root, "opsx/gate-fixture", [
        { finding_id: "3", decision: "override", reason: "nope" },
      ]),
    ).toThrow(/unknown finding ids/)
    expect(() =>
      recordSpecReviewDecisionsToState(root, "opsx/gate-fixture", [
        { finding_id: "1", decision: "override", reason: "a" },
        { finding_id: "1", decision: "override", reason: "b" },
      ]),
    ).toThrow(/duplicate finding ids/)
    expect(() =>
      recordSpecReviewDecisionsToState(root, "opsx/gate-fixture", [
        { finding_id: "1", decision: "apply" },
      ]),
    ).toThrow(/invalid spec-review decisions/)
    expect(() =>
      recordSpecReviewDecisionsToState(root, "opsx/gate-fixture", [
        { finding_id: "1", decision: "override", reason: "" },
      ]),
    ).toThrow(/invalid spec-review decisions/)
    expect(() =>
      recordSpecReviewDecisionsToState(root, "opsx/other-fixture", [
        { finding_id: "1", decision: "override", reason: "x" },
      ]),
    ).toThrow(/no uf\.unleash state/)
  })
})

// ---------------------------------------------------------------------------
// Checkpoint tool gate integration (HARDEN-002, tasks 3.2–3.3)
// ---------------------------------------------------------------------------

describe("checkpoint spec-review gate (HARDEN-002, tasks 3.2-3.3)", () => {
  /** Scratch root whose iteration 1 consolidated to REQUEST CHANGES. */
  async function gateRoot(findingCount: number): Promise<string> {
    const root = await scratchRoot("checkpoint-gate")
    appendSpecReviewIterationToState(root, "opsx/gate-fixture", {
      verdict: "REQUEST CHANGES",
      correlation_id: "5355acba-62f4-4859-8775-7294b3e8702c",
      finding_count: findingCount,
      finding_ids: specReviewFindingIds(findingCount),
    })
    return root
  }

  it("refuses to record the gated step without decisions (missing_per_finding_decisions)", async () => {
    const root = await gateRoot(8)
    const result = await invokeTool(createCheckpointTool(unusedExec, root), {
      command: SPEC_REVIEW_STATE_COMMAND,
      step: "spec-review",
      key: "opsx-gate-fixture",
    })

    expect(result.status).toBe("error")
    expect(result.code).toBe("missing_per_finding_decisions")
    expect(result.message).toContain("undecided findings [1, 2, 3, 4, 5, 6, 7, 8]")

    const state = await readStateFile(root, "opsx-gate-fixture")
    expect(state?.completed_steps).not.toContain("spec-review")
  })

  it("refuses every advancing step while the gate holds", async () => {
    const root = await gateRoot(2)
    for (const step of [
      "spec-review",
      "implement",
      "implement-phase-1",
      "quick-pre-flight",
      "code-review",
      "retrospective",
      "demo",
    ]) {
      const result = await invokeTool(createCheckpointTool(unusedExec, root), {
        command: SPEC_REVIEW_STATE_COMMAND,
        step,
        key: "opsx-gate-fixture",
      })
      expect(result.status, `step ${step}`).toBe("error")
      expect(result.code, `step ${step}`).toBe("missing_per_finding_decisions")
    }
  })

  it("records gate-machinery steps while the gate holds", async () => {
    const root = await gateRoot(2)
    for (const step of SPEC_REVIEW_GATE_STEP_ALLOWLIST) {
      const result = await invokeTool(createCheckpointTool(unusedExec, root), {
        command: SPEC_REVIEW_STATE_COMMAND,
        step,
        key: "opsx-gate-fixture",
      })
      expect(result.status, `step ${step}`).toBe("ok")
    }
  })

  it("partial decisions persist but still refuse, naming only the remainder", async () => {
    const root = await gateRoot(3)
    const result = await invokeTool(createCheckpointTool(unusedExec, root), {
      command: SPEC_REVIEW_STATE_COMMAND,
      step: "spec-review",
      key: "opsx-gate-fixture",
      spec_review_decisions: [
        { finding_id: "1", decision: "apply", fix_reference: "spec.md HARDEN-001" },
      ],
    })

    expect(result.status).toBe("error")
    expect(result.code).toBe("missing_per_finding_decisions")
    expect(result.message).toContain("undecided findings [2, 3]")

    const persisted = await readStateFile(root, "opsx-gate-fixture")
    const record = persisted?.data[SPEC_REVIEW_RECORD_KEY] as SpecReviewGateRecord
    expect(record.decisions?.decisions).toHaveLength(1)
  })

  it("complete apply decisions refuse until a subsequent iteration approves", async () => {
    const root = await gateRoot(2)
    const refused = await invokeTool(createCheckpointTool(unusedExec, root), {
      command: SPEC_REVIEW_STATE_COMMAND,
      step: "spec-review",
      key: "opsx-gate-fixture",
      spec_review_decisions: [
        { finding_id: "1", decision: "apply", fix_reference: "spec.md HARDEN-001" },
        { finding_id: "2", decision: "apply", fix_reference: "spec.md HARDEN-002" },
      ],
    })

    expect(refused.status).toBe("error")
    expect(refused.code).toBe("pending_approving_iteration")
    expect(refused.message).toContain("must consolidate to APPROVE")

    // Iteration 2 consolidates to APPROVE: the gate opens and the step lands.
    appendSpecReviewIterationToState(root, "opsx/gate-fixture", {
      verdict: "APPROVE",
      correlation_id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
      finding_count: 0,
      finding_ids: [],
    })
    const advanced = await invokeTool(createCheckpointTool(unusedExec, root), {
      command: SPEC_REVIEW_STATE_COMMAND,
      step: "spec-review",
      key: "opsx-gate-fixture",
    })

    expect(advanced.status).toBe("ok")
    const state = await readStateFile(root, "opsx-gate-fixture")
    expect(state?.completed_steps).toContain("spec-review")
  })

  it("an audited override set advances with overrides recorded alongside the verdict", async () => {
    const root = await gateRoot(2)
    const result = await invokeTool(createCheckpointTool(unusedExec, root), {
      command: SPEC_REVIEW_STATE_COMMAND,
      step: "spec-review",
      key: "opsx-gate-fixture",
      spec_review_decisions: [
        { finding_id: "1", decision: "override", reason: "accepted risk" },
        { finding_id: "2", decision: "override", reason: "out of scope" },
      ],
    })

    expect(result.status).toBe("ok")
    const state = await readStateFile(root, "opsx-gate-fixture")
    expect(state?.completed_steps).toContain("spec-review")
    const record = state?.data[SPEC_REVIEW_RECORD_KEY] as SpecReviewGateRecord
    expect(record.decisions).toEqual({
      iteration: "1",
      decisions: [
        { finding_id: "1", decision: "override", reason: "accepted risk" },
        { finding_id: "2", decision: "override", reason: "out of scope" },
      ],
    })
    expect(record.overrides).toEqual([
      { finding_id: "1", reason: "accepted risk" },
      { finding_id: "2", reason: "out of scope" },
    ])
    expect(record.iterations["1"].verdict).toBe("REQUEST CHANGES")
  })

  it("an override without a reason is rejected and nothing is recorded", async () => {
    const root = await gateRoot(1)
    const result = await invokeTool(createCheckpointTool(unusedExec, root), {
      command: SPEC_REVIEW_STATE_COMMAND,
      step: "spec-review",
      key: "opsx-gate-fixture",
      spec_review_decisions: [{ finding_id: "1", decision: "override" }],
    })

    expect(result.status).toBe("error")
    expect(result.code).toBe("invalid_spec_review_decisions")
    const state = await readStateFile(root, "opsx-gate-fixture")
    const record = state?.data[SPEC_REVIEW_RECORD_KEY] as SpecReviewGateRecord
    expect(record.decisions).toBeUndefined()
    expect(record.overrides).toBeUndefined()
    expect(state?.completed_steps).not.toContain("spec-review")
  })

  it("a pause decision holds the gate (spec_review_gate_paused)", async () => {
    const root = await gateRoot(2)
    const result = await invokeTool(createCheckpointTool(unusedExec, root), {
      command: SPEC_REVIEW_STATE_COMMAND,
      step: "spec-review",
      key: "opsx-gate-fixture",
      spec_review_decisions: [
        { finding_id: "1", decision: "apply", fix_reference: "spec.md HARDEN-001" },
        { finding_id: "2", decision: "pause" },
      ],
    })

    expect(result.status).toBe("error")
    expect(result.code).toBe("spec_review_gate_paused")
    expect(result.message).toContain("pause decision on findings [2]")

    const state = await readStateFile(root, "opsx-gate-fixture")
    expect(state?.completed_steps).not.toContain("spec-review")
  })

  it("spec_review_decisions is rejected for other commands", async () => {
    const root = await gateRoot(1)
    const result = await invokeTool(createCheckpointTool(unusedExec, root), {
      command: "uf.review-council",
      step: "review",
      key: "opsx-gate-fixture",
      spec_review_decisions: [{ finding_id: "1", decision: "override", reason: "x" }],
    })

    expect(result.status).toBe("error")
    expect(result.code).toBe("invalid_spec_review_decisions")
  })

  it("other commands bypass the gate entirely", async () => {
    const root = await gateRoot(2)
    const result = await invokeTool(createCheckpointTool(unusedExec, root), {
      command: "uf.review-council",
      step: "implement",
      key: "opsx-gate-fixture",
    })

    expect(result.status).toBe("ok")
  })
})

// ---------------------------------------------------------------------------
// Dispatch re-dispatch guard (HARDEN-003, tasks 4.1–4.2)
// ---------------------------------------------------------------------------

describe("dispatch re-dispatch guard (HARDEN-003, tasks 4.1-4.2)", () => {
  /** Scratch root whose iteration 1 consolidated to REQUEST CHANGES. */
  async function guardRoot(): Promise<string> {
    const root = await scratchRoot("dispatch-guard")
    appendSpecReviewIterationToState(root, "opsx/gate-fixture", {
      verdict: "REQUEST CHANGES",
      correlation_id: "5355acba-62f4-4859-8775-7294b3e8702c",
      finding_count: 2,
      finding_ids: ["1", "2"],
    })
    return root
  }

  it("refuses a re-dispatch while the latest iteration is undecided", async () => {
    const root = await guardRoot()
    const refusal = specReviewDispatchGuard(root, "specs", "opsx/gate-fixture")
    expect(refusal).not.toBeNull()
    expect(refusal!.message).toContain("iteration_already_recorded")
    expect(refusal!.message).toContain("iteration 1")
    expect(refusal!.message).toContain("REQUEST CHANGES")
    expect(refusal!.message).toContain("undecided findings [1, 2]")
  })

  it("refuses a re-dispatch while the gate is held by a pause decision", async () => {
    const root = await guardRoot()
    recordSpecReviewDecisionsToState(root, "opsx/gate-fixture", [
      { finding_id: "1", decision: "apply", fix_reference: "spec.md HARDEN-001" },
      { finding_id: "2", decision: "pause" },
    ])
    const refusal = specReviewDispatchGuard(root, "specs", "opsx/gate-fixture")
    expect(refusal).not.toBeNull()
    expect(refusal!.message).toContain("iteration_already_recorded")
    expect(refusal!.message).toContain("pause decision on findings [2]")
  })

  it("permits iteration 4 after a decided iteration 3 (scenario 2)", async () => {
    // The spec's acceptance scenario: iterations 1–3 recorded REQUEST
    // CHANGES, iteration 3's decision set complete with applied fixes —
    // the iteration 4 dispatch MUST proceed.
    const root = await guardRoot()
    for (let iteration = 2; iteration <= 3; iteration++) {
      appendSpecReviewIterationToState(root, "opsx/gate-fixture", {
        verdict: "REQUEST CHANGES",
        correlation_id: `dddddddd-dddd-4ddd-8ddd-ddddddddddd${iteration}`,
        finding_count: 1,
        finding_ids: ["1"],
      })
      recordSpecReviewDecisionsToState(root, "opsx/gate-fixture", [
        { finding_id: "1", decision: "apply", fix_reference: `spec.md fix ${iteration}` },
      ])
    }
    // Iteration 3 (latest) is decided: the next iteration is 4 — within
    // the initial-review-plus-3-fix-iterations budget.
    expect(specReviewDispatchGuard(root, "specs", "opsx/gate-fixture")).toBeNull()
  })

  it("permits an all-override adjudication and an approving latest", async () => {
    const root = await guardRoot()
    recordSpecReviewDecisionsToState(root, "opsx/gate-fixture", [
      { finding_id: "1", decision: "override", reason: "accepted risk" },
      { finding_id: "2", decision: "override", reason: "out of scope" },
    ])
    expect(specReviewDispatchGuard(root, "specs", "opsx/gate-fixture")).toBeNull()

    const approved = await scratchRoot("dispatch-guard-approved")
    appendSpecReviewIterationToState(approved, "opsx/gate-fixture", {
      verdict: "APPROVE",
      correlation_id: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
      finding_count: 0,
      finding_ids: [],
    })
    expect(specReviewDispatchGuard(approved, "specs", "opsx/gate-fixture")).toBeNull()
  })

  it("enforces the iteration cap beyond the fourth dispatch", async () => {
    const root = await guardRoot()
    for (let iteration = 2; iteration <= 4; iteration++) {
      appendSpecReviewIterationToState(root, "opsx/gate-fixture", {
        verdict: "REQUEST CHANGES",
        correlation_id: `cccccccc-cccc-4ccc-8ccc-ccccccccccc${iteration}`,
        finding_count: 1,
        finding_ids: ["1"],
      })
      recordSpecReviewDecisionsToState(root, "opsx/gate-fixture", [
        { finding_id: "1", decision: "apply", fix_reference: `spec.md fix ${iteration}` },
      ])
    }
    // Four recorded iterations, the latest decided: only the cap refuses —
    // the would-be iteration 5 exceeds initial review + 3 fix iterations.
    const refusal = specReviewDispatchGuard(root, "specs", "opsx/gate-fixture")
    expect(refusal).not.toBeNull()
    expect(refusal!.message).toContain("spec_review_iteration_cap_exceeded")
    expect(refusal!.message).toContain("iteration 5")
  })

  it("never consults the record for non-specs modes or unkeyable contexts", async () => {
    const root = await guardRoot()
    expect(specReviewDispatchGuard(root, "code", "opsx/gate-fixture")).toBeNull()
    expect(specReviewDispatchGuard(root, "triage", "opsx/gate-fixture")).toBeNull()
    expect(specReviewDispatchGuard(root, "specs", undefined)).toBeNull()
    expect(specReviewDispatchGuard(root, "specs", "opsx/other-fixture")).toBeNull()
  })

  it("the dispatch tool refuses before minting a correlation (no session left behind)", async () => {
    const root = await guardRoot()
    const dispatchTool = createDispatchReviewRunsTool(
      // Never reached: the guard returns before sibling evidence
      // acquisition or any run dispatch touches these dependencies.
      undefined as unknown as DispatchAgentRunDependencies,
      undefined as unknown as SiblingAcquisitionDependencies,
      root,
    )
    const metadata = await dispatchMetadata(dispatchTool, {
      mode: "specs",
      command: "review-council",
      diff_path: join(root, "review.diff"),
      changed_files: "specs/spec-review-gate/spec.md",
      input_context: "local opsx/gate-fixture",
      plan_entries: explicitRunPlan("divisor-guard"),
      session_metadata: {
        command: "review-council",
        mode: "specs",
        full: false,
        input_context: {
          kind: "local",
          pr_number: null,
          base_ref: "main",
          base_sha: "a".repeat(40),
          head_ref: "opsx/gate-fixture",
          head_sha: "b".repeat(40),
          uncommitted: false,
        },
        change_profile: validChangeProfile(),
        plan: explicitRunPlan("divisor-guard"),
        coverage: validCoverage(),
      },
    })

    expect(metadata.status).toBe("error")
    expect(metadata.message).toContain("iteration_already_recorded")
    expect(metadata.runs_completed).toBe(0)
    expect(metadata.runs_skipped).toBe(0)
    // The refusal precedes correlation minting: no correlation is reported
    // and the session correlation map keeps its prior (absent) entry.
    expect(metadata.correlation_id).toBeUndefined()
    expect(_getSessionCorrelationMap().get("spec-review-gate-session")).toBeUndefined()
    // The iteration log is untouched: no overwrite, no auto-bump.
    const record = readSpecReviewGateRecord(root, "opsx/gate-fixture")
    expect(Object.keys(record!.iterations)).toEqual(["1"])
  })
})

// ---------------------------------------------------------------------------
// relay_spec_review_gate (HARDEN-004, tasks 5.1/5.6)
// ---------------------------------------------------------------------------

describe("relay_spec_review_gate (HARDEN-004, tasks 5.1/5.6)", () => {
  const GATE_BRANCH = "opsx/gate-fixture"

  /** The registration dependency shape: real-fs artifact reader. */
  const realReadText = (path: string): Promise<string> => readFile(path, "utf8")

  /** Narrowing helper for the relay envelope's data payload. */
  function okRelayData(
    result: { status: string; data?: RelaySpecReviewGateData; message?: string; code?: string },
  ): RelaySpecReviewGateData {
    expect(result.status).toBe("ok")
    if (result.data === undefined) {
      throw new Error(`expected ok data, received ${JSON.stringify(result)}`)
    }
    return result.data
  }

  /** Write one artifact envelope into the scratch dispatch store. */
  async function writeArtifact(projectRoot: string, correlationId: string, raw: string): Promise<void> {
    const storeDir = join(projectRoot, ".uf", "artifacts", "dispatch")
    await mkdir(storeDir, { recursive: true })
    await writeFile(join(storeDir, `${correlationId}.json`), raw, "utf8")
  }

  /** Record one undecided RC iteration plus its backing artifact. */
  async function recordUndecidedIteration(projectRoot: string, correlationId: string): Promise<void> {
    await writeArtifact(
      projectRoot,
      correlationId,
      artifactEnvelope({
        branch: GATE_BRANCH,
        verdict: "REQUEST CHANGES",
        correlationId,
        findingCount: 2,
      }),
    )
    await appendSpecReviewIterationToState(projectRoot, GATE_BRANCH, {
      verdict: "REQUEST CHANGES",
      correlation_id: correlationId,
      finding_count: 2,
      finding_ids: ["1", "2"],
    })
  }

  function relayTool(projectRoot: string) {
    return createRelaySpecReviewGateTool(unusedExec, projectRoot, { readText: realReadText })
  }

  it("reports gate_paused false with null status fields when no spec-review record exists", async () => {
    const root = await scratchRoot("relay-no-record")
    const result = await invokeTool<RelaySpecReviewGateData>(relayTool(root), {
      command: "uf.unleash",
      key: GATE_BRANCH,
    })
    expect(result.status).toBe("ok")
    expect(result.data).toStrictEqual({
      gate_paused: false,
      relay: null,
      iteration: null,
      verdict: null,
      correlation_id: null,
      finding_count: null,
      undecided_finding_ids: [],
    })
  })

  it("renders the undecided gate from the dispatch artifact (severity, description, file:line, correlation id)", async () => {
    const root = await scratchRoot("relay-undecided")
    const cid = randomUUID()
    await recordUndecidedIteration(root, cid)

    const data = okRelayData(
      await invokeTool<RelaySpecReviewGateData>(relayTool(root), { command: "uf.unleash", key: GATE_BRANCH }),
    )

    expect(data.gate_paused).toBe(true)
    expect(data.iteration).toBe(1)
    expect(data.verdict).toBe("REQUEST CHANGES")
    expect(data.correlation_id).toBe(cid)
    expect(data.finding_count).toBe(2)
    expect(data.undecided_finding_ids).toEqual(["1", "2"])
    // The relay is artifact-sourced (HARDEN-004): every finding renders from
    // the dispatch artifact with severity, description, and file:line.
    expect(data.relay).toContain(`Spec-review gate holds: iteration 1 (REQUEST CHANGES), correlation ${cid}.`)
    expect(data.relay).toContain("2 findings consolidated; undecided finding ids: [1, 2].")
    expect(data.relay).toContain("Findings from the dispatch artifact:")
    expect(data.relay).toContain("1. [HIGH] Replayed finding 1 (file-1.ts:1) — undecided")
    expect(data.relay).toContain("2. [MEDIUM] Replayed finding 2 (file-2.ts:2) — undecided")
    expect(data.relay).toContain("iteration_already_recorded")
  })

  it("annotates each finding with its recorded decision and marks the rest undecided", async () => {
    const root = await scratchRoot("relay-partial")
    const cid = randomUUID()
    await recordUndecidedIteration(root, cid)
    await recordSpecReviewDecisionsToState(root, GATE_BRANCH, [
      { finding_id: "1", decision: "apply", fix_reference: "deadbeef" },
    ])

    const data = okRelayData(
      await invokeTool<RelaySpecReviewGateData>(relayTool(root), { command: "uf.unleash", key: GATE_BRANCH }),
    )

    expect(data.gate_paused).toBe(true)
    expect(data.undecided_finding_ids).toEqual(["2"])
    expect(data.relay).toContain("1. [HIGH] Replayed finding 1 (file-1.ts:1) — apply")
    expect(data.relay).toContain("2. [MEDIUM] Replayed finding 2 (file-2.ts:2) — undecided")
  })

  it("renders the paused gate naming the pause and the exit path", async () => {
    const root = await scratchRoot("relay-paused")
    const cid = randomUUID()
    await recordUndecidedIteration(root, cid)
    await recordSpecReviewDecisionsToState(root, GATE_BRANCH, [
      { finding_id: "1", decision: "pause" },
      { finding_id: "2", decision: "override", reason: "accepted risk" },
    ])

    const data = okRelayData(
      await invokeTool<RelaySpecReviewGateData>(relayTool(root), { command: "uf.unleash", key: GATE_BRANCH }),
    )

    expect(data.gate_paused).toBe(true)
    expect(data.undecided_finding_ids).toEqual([])
    expect(data.relay).toContain(`Spec-review gate paused: iteration 1 (REQUEST CHANGES), correlation ${cid}.`)
    expect(data.relay).toContain("The human paused findings [1].")
    expect(data.relay).toContain("spec-review paused exit path")
  })

  it("reports gate_paused false for an all-override decision set (decided gate)", async () => {
    const root = await scratchRoot("relay-override")
    const cid = randomUUID()
    await recordUndecidedIteration(root, cid)
    await recordSpecReviewDecisionsToState(root, GATE_BRANCH, [
      { finding_id: "1", decision: "override", reason: "accepted risk" },
      { finding_id: "2", decision: "override", reason: "accepted risk" },
    ])

    const data = okRelayData(
      await invokeTool<RelaySpecReviewGateData>(relayTool(root), { command: "uf.unleash", key: GATE_BRANCH }),
    )

    expect(data.gate_paused).toBe(false)
    expect(data.relay).toBeNull()
    expect(data.undecided_finding_ids).toEqual([])
    expect(data.iteration).toBe(1)
    expect(data.verdict).toBe("REQUEST CHANGES")
  })

  it("reports gate_paused false for apply decisions awaiting the next iteration (decided gate)", async () => {
    const root = await scratchRoot("relay-apply")
    const cid = randomUUID()
    await recordUndecidedIteration(root, cid)
    await recordSpecReviewDecisionsToState(root, GATE_BRANCH, [
      { finding_id: "1", decision: "apply", fix_reference: "deadbeef" },
      { finding_id: "2", decision: "apply", fix_reference: "feedface" },
    ])

    const data = okRelayData(
      await invokeTool<RelaySpecReviewGateData>(relayTool(root), { command: "uf.unleash", key: GATE_BRANCH }),
    )

    expect(data.gate_paused).toBe(false)
    expect(data.relay).toBeNull()
    expect(data.undecided_finding_ids).toEqual([])
  })

  it("reports gate_paused false for an approving latest iteration (approved gate)", async () => {
    const root = await scratchRoot("relay-approve")
    await appendSpecReviewIterationToState(
      root,
      GATE_BRANCH,
      iterationEntry({ verdict: "APPROVE", finding_count: 0, finding_ids: [] }),
    )

    const data = okRelayData(
      await invokeTool<RelaySpecReviewGateData>(relayTool(root), { command: "uf.unleash", key: GATE_BRANCH }),
    )

    expect(data.gate_paused).toBe(false)
    expect(data.relay).toBeNull()
    expect(data.verdict).toBe("APPROVE")
    expect(data.iteration).toBe(1)
  })

  it("keeps the gate paused when the artifact is missing (fail-safe relay from the record)", async () => {
    const root = await scratchRoot("relay-missing-artifact")
    const cid = randomUUID()
    await appendSpecReviewIterationToState(root, GATE_BRANCH, {
      verdict: "REQUEST CHANGES",
      correlation_id: cid,
      finding_count: 2,
      finding_ids: ["1", "2"],
    })

    const data = okRelayData(
      await invokeTool<RelaySpecReviewGateData>(relayTool(root), { command: "uf.unleash", key: GATE_BRANCH }),
    )

    // A missing artifact never un-pauses the gate; the relay falls back to
    // the record's metadata with the artifact-unreadable note.
    expect(data.gate_paused).toBe(true)
    expect(data.undecided_finding_ids).toEqual(["1", "2"])
    expect(data.relay).toContain(`Spec-review gate holds: iteration 1 (REQUEST CHANGES), correlation ${cid}.`)
    expect(data.relay).toContain(`.uf/artifacts/dispatch/${cid}.json could not be read`)
    expect(data.relay).toContain("finding details are unavailable")
  })

  it("treats a malformed artifact as unreadable (same fail-safe relay)", async () => {
    const root = await scratchRoot("relay-malformed-artifact")
    const cid = randomUUID()
    await writeArtifact(root, cid, "{not json")
    await appendSpecReviewIterationToState(root, GATE_BRANCH, {
      verdict: "REQUEST CHANGES",
      correlation_id: cid,
      finding_count: 2,
      finding_ids: ["1", "2"],
    })

    const data = okRelayData(
      await invokeTool<RelaySpecReviewGateData>(relayTool(root), { command: "uf.unleash", key: GATE_BRANCH }),
    )

    expect(data.gate_paused).toBe(true)
    expect(data.relay).toContain("could not be read")
  })

  it("rejects commands other than uf.unleash", async () => {
    const root = await scratchRoot("relay-command")
    const result = await invokeTool(relayTool(root), {
      command: "uf.review-council",
      key: GATE_BRANCH,
    })
    expect(result.status).toBe("error")
    expect(result.code).toBe("RELAY_SPEC_REVIEW_COMMAND")
  })

  it("round-trips the consolidated record through get_progress, checkpoint, and the relay (DE-007 layer 3)", async () => {
    const root = await scratchRoot("relay-round-trip")
    const cid = randomUUID()
    await recordUndecidedIteration(root, cid)

    // get_progress (reconciled) exposes the record the consolidation wrote.
    const progress = await invokeTool<ProgressState>(createGetProgressTool(unusedExec, root), {
      command: "uf.unleash",
      key: GATE_BRANCH,
    })
    expect(progress.status).toBe("ok")
    const record = progress.data?.data["spec-review"] as SpecReviewGateRecord | undefined
    expect(record?.iterations["1"]?.correlation_id).toBe(cid)
    expect(record?.iterations["1"]?.verdict).toBe("REQUEST CHANGES")

    // checkpoint refuses to advance while the gate holds (HARDEN-002).
    const checkpoint = await invokeTool(createCheckpointTool(unusedExec, root), {
      command: "uf.unleash",
      key: GATE_BRANCH,
      step: "implement",
      data: "{}",
    })
    expect(checkpoint.status).toBe("error")
    expect(checkpoint.code).toBe("missing_per_finding_decisions")

    // the relay pauses on the same record (HARDEN-004).
    const relay = okRelayData(
      await invokeTool<RelaySpecReviewGateData>(relayTool(root), { command: "uf.unleash", key: GATE_BRANCH }),
    )
    expect(relay.gate_paused).toBe(true)
    expect(relay.correlation_id).toBe(cid)
  })
})

// ---------------------------------------------------------------------------
// Bypass-path regression coverage (HARDEN-005, tasks 6.1–6.2)
// ---------------------------------------------------------------------------

/** The incident's dispatch correlation (spec HARDEN-001/005 scenarios). */
const INCIDENT_CORRELATION = "5355acba-62f4-4859-8775-7294b3e8702c"

/** The incident's consolidated finding set: 1 HIGH + 7 MEDIUM/LOW. */
const INCIDENT_FINDINGS = [
  {
    severity: "HIGH",
    category: "gate-semantics",
    description: "Spec edits after REQUEST CHANGES bypass the per-finding gate",
    root_cause: "No per-finding decision record gates advancement",
    file: "specs/spec.md",
    line: 42,
  },
  ...Array.from({ length: 4 }, (_, index) => ({
    severity: "MEDIUM",
    category: "gate-semantics",
    description: `Undecided medium finding ${index + 1} from the incident replay`,
    root_cause: "Consolidated without a per-finding decision",
    file: `specs/file-${index + 1}.md`,
    line: index + 1,
  })),
  ...Array.from({ length: 3 }, (_, index) => ({
    severity: "LOW",
    category: "gate-semantics",
    description: `Undecided low finding ${index + 1} from the incident replay`,
    root_cause: "Consolidated without a per-finding decision",
    file: `specs/file-${index + 5}.md`,
    line: index + 5,
  })),
]

describe("bypass-path regression coverage (HARDEN-005, tasks 6.1-6.2)", () => {
  const GATE_BRANCH = "opsx/gate-fixture"

  /** The registration dependency shape: real-fs artifact reader. */
  const realReadText = (path: string): Promise<string> => readFile(path, "utf8")

  /** Narrowing helper for the relay envelope's data payload. */
  function okRelayData(
    result: { status: string; data?: RelaySpecReviewGateData; message?: string; code?: string },
  ): RelaySpecReviewGateData {
    expect(result.status).toBe("ok")
    if (result.data === undefined) {
      throw new Error(`expected ok data, received ${JSON.stringify(result)}`)
    }
    return result.data
  }

  /** Write one artifact envelope into the scratch dispatch store. */
  async function writeArtifact(projectRoot: string, correlationId: string, raw: string): Promise<void> {
    const storeDir = join(projectRoot, ".uf", "artifacts", "dispatch")
    await mkdir(storeDir, { recursive: true })
    await writeFile(join(storeDir, `${correlationId}.json`), raw, "utf8")
  }

  function relayTool(projectRoot: string) {
    return createRelaySpecReviewGateTool(unusedExec, projectRoot, { readText: realReadText })
  }

  /**
   * Write the incident's mock dispatch store at its recorded correlation —
   * mirrors setupDispatchSession but pins the correlation id the record and
   * the relay must agree on.
   */
  async function setupIncidentStore(): Promise<void> {
    const dir = join(tmpdir(), "opencode", `dispatch-${INCIDENT_CORRELATION}`)
    await mkdir(dir, { recursive: true })
    cleanupDirs.push(dir)
    await writeFile(
      join(dir, "session-metadata.json"),
      JSON.stringify(
        defaultSessionMetadata({ mode: "specs", plan: explicitRunPlan("divisor-guard") }),
        null,
        2,
      ),
      "utf8",
    )
    await writeFile(
      join(dir, "run-divisor-guard.json"),
      JSON.stringify(runFixture({ agent: "divisor-guard", findings: INCIDENT_FINDINGS }), null, 2),
      "utf8",
    )
    _getSessionCorrelationMap().set("spec-review-gate-session", INCIDENT_CORRELATION)
  }

  /** Artifact (8 findings) + undecided iteration 1 at the incident correlation. */
  async function recordIncidentGate(projectRoot: string): Promise<void> {
    await writeArtifact(
      projectRoot,
      INCIDENT_CORRELATION,
      artifactEnvelope({
        branch: GATE_BRANCH,
        verdict: "REQUEST CHANGES",
        correlationId: INCIDENT_CORRELATION,
        findingCount: 8,
      }),
    )
    appendSpecReviewIterationToState(projectRoot, GATE_BRANCH, {
      verdict: "REQUEST CHANGES",
      correlation_id: INCIDENT_CORRELATION,
      finding_count: 8,
      finding_ids: specReviewFindingIds(8),
    })
  }

  it("replays the compaction-summary bypass: the gate holds through spec edits and a child-text return (task 6.1)", async () => {
    const root = await scratchRoot("bypass-replay")
    await setupIncidentStore()

    // The child consolidates REQUEST CHANGES — the incident's 8 findings.
    const tool = createConsolidateDispatchTool(createFinalizationDependencies(root))
    const result = await tool.execute(
      {
        provenance: {
          branch: GATE_BRANCH,
          commit: "b".repeat(40),
          workflow_id: "uf.review-council",
        },
      },
      noopContext(),
    )
    const finalization = JSON.parse(result.output) as {
      status: string
      operation_verdict: string
      correlation_id: string
    }
    expect(finalization.status).toBe("success")
    expect(finalization.operation_verdict).toBe("REQUEST CHANGES")
    expect(finalization.correlation_id).toBe(INCIDENT_CORRELATION)

    // (a) Consolidation persisted BEFORE the edits: the record is durable
    // the moment the tool returned, before any model edit can happen.
    const persisted = readSpecReviewGateRecord(root, GATE_BRANCH)
    expect(persisted).not.toBeNull()
    const latest = latestSpecReviewIteration(persisted!)
    expect(latest?.entry).toMatchObject({
      verdict: "REQUEST CHANGES",
      correlation_id: INCIDENT_CORRELATION,
      finding_count: 8,
    })
    expect(latest?.entry.finding_ids).toEqual(specReviewFindingIds(8))

    // The child edits spec.md three times — the bypass attempt — then
    // returns a compaction summary claiming the findings were addressed.
    // The record survives every edit unchanged.
    const specPath = join(root, "openspec", "changes", "bypass-replay", "specs", "spec.md")
    await mkdir(dirname(specPath), { recursive: true })
    const childReturnText =
      "[Session compacted] Continued from a compressed context. " +
      "Spec review returned REQUEST CHANGES earlier; all findings were " +
      "addressed across three spec edits."
    for (let edit = 1; edit <= 3; edit += 1) {
      await writeFile(specPath, `# incident replay — edit ${edit}\n`, "utf8")
      expect(readSpecReviewGateRecord(root, GATE_BRANCH)).toEqual(persisted)
    }

    // (b) checkpoint refuses to advance without per-finding decisions.
    const refused = await invokeTool(createCheckpointTool(unusedExec, root), {
      command: SPEC_REVIEW_STATE_COMMAND,
      step: "implement",
      key: GATE_BRANCH,
    })
    expect(refused.status).toBe("error")
    expect(refused.code).toBe("missing_per_finding_decisions")
    expect(refused.message).toContain("undecided findings [1, 2, 3, 4, 5, 6, 7, 8]")

    // (c) get_progress is not stale: the recorded iteration resurfaces.
    const progress = await invokeTool<ProgressState>(
      createGetProgressTool(unusedExec, root),
      { command: SPEC_REVIEW_STATE_COMMAND, key: GATE_BRANCH },
    )
    const viaGet = progress.data?.data?.[SPEC_REVIEW_RECORD_KEY] as SpecReviewGateRecord
    expect(latestSpecReviewIteration(viaGet)?.entry).toMatchObject({
      verdict: "REQUEST CHANGES",
      correlation_id: INCIDENT_CORRELATION,
      finding_count: 8,
    })

    // (d) The unconditional relay renders all 8 findings from the artifact
    // (deterministic, no LLM); the compaction summary and its "addressed"
    // claim never substitute for the artifact, and no `overrides: []` was
    // auto-written as if the findings were never raised.
    const relay = okRelayData(
      await invokeTool<RelaySpecReviewGateData>(relayTool(root), {
        command: SPEC_REVIEW_STATE_COMMAND,
        key: GATE_BRANCH,
      }),
    )
    expect(relay.gate_paused).toBe(true)
    expect(relay.correlation_id).toBe(INCIDENT_CORRELATION)
    expect(relay.finding_count).toBe(8)
    expect(relay.undecided_finding_ids).toEqual(specReviewFindingIds(8))
    expect(relay.relay).toContain(
      `Spec-review gate holds: iteration 1 (REQUEST CHANGES), correlation ${INCIDENT_CORRELATION}.`,
    )
    expect(relay.relay).toContain(
      "8 findings consolidated; undecided finding ids: [1, 2, 3, 4, 5, 6, 7, 8].",
    )
    for (const [index, finding] of INCIDENT_FINDINGS.entries()) {
      expect(relay.relay).toContain(
        `${index + 1}. [${finding.severity}] ${finding.description} (${finding.file}:${finding.line}) — undecided`,
      )
    }
    expect(relay.relay).not.toContain(childReturnText)
    expect(relay.relay).not.toContain("Session compacted")
    expect(relay.relay).not.toContain("all findings were addressed")

    // No empty-override auto-write cleared the gate.
    const finalRecord = readSpecReviewGateRecord(root, GATE_BRANCH)
    expect(finalRecord?.decisions).toBeUndefined()
  })

  it("renders the full relay from the artifact when the child returned empty text (task 6.2)", async () => {
    const root = await scratchRoot("edge-empty-return")
    await recordIncidentGate(root)

    // The child returned empty text (a wedged or truncated completion).
    // The relay takes no text input — the render comes from the record and
    // the artifact alone, so the gate still presents every finding.
    const relay = okRelayData(
      await invokeTool<RelaySpecReviewGateData>(relayTool(root), {
        command: SPEC_REVIEW_STATE_COMMAND,
        key: GATE_BRANCH,
      }),
    )
    expect(relay.gate_paused).toBe(true)
    expect(relay.finding_count).toBe(8)
    expect(relay.undecided_finding_ids).toEqual(specReviewFindingIds(8))
    expect(relay.relay).toContain(`correlation ${INCIDENT_CORRELATION}.`)
    for (let index = 1; index <= 8; index += 1) {
      const severity = index === 1 ? "HIGH" : "MEDIUM"
      expect(relay.relay).toContain(
        `${index}. [${severity}] Replayed finding ${index} (file-${index}.ts:${index}) — undecided`,
      )
    }
  })

  it("renders artifact-sourced findings even when the child text is a correct REQUEST CHANGES report (task 6.2)", async () => {
    const root = await scratchRoot("edge-correct-text")
    await recordIncidentGate(root)

    // HARDEN-004: accurate child prose never substitutes for the tool's
    // artifact render — the relayed list carries the correlation id and
    // the artifact's file:line locs, not the child's wording.
    const childReturnText =
      "REQUEST CHANGES: 8 findings (1 HIGH, 7 MEDIUM/LOW). " +
      "The HIGH finding sits at specs/gate.md:99; consider proceeding."

    const relay = okRelayData(
      await invokeTool<RelaySpecReviewGateData>(relayTool(root), {
        command: SPEC_REVIEW_STATE_COMMAND,
        key: GATE_BRANCH,
      }),
    )
    expect(relay.gate_paused).toBe(true)
    expect(relay.relay).toContain("1. [HIGH] Replayed finding 1 (file-1.ts:1) — undecided")
    expect(relay.relay).toContain("8. [MEDIUM] Replayed finding 8 (file-8.ts:8) — undecided")
    expect(relay.relay).toContain(`correlation ${INCIDENT_CORRELATION}.`)
    expect(relay.relay).not.toContain(childReturnText)
    expect(relay.relay).not.toContain("specs/gate.md:99")
    expect(relay.relay).not.toContain("consider proceeding")
  })

  it("re-dispatches after decisions and lands as the next iteration (task 6.2)", async () => {
    const root = await scratchRoot("edge-next-iteration")
    const firstCorrelation = "11111111-1111-4111-8111-111111111111"
    await writeArtifact(
      root,
      firstCorrelation,
      artifactEnvelope({
        branch: GATE_BRANCH,
        verdict: "REQUEST CHANGES",
        correlationId: firstCorrelation,
        findingCount: 2,
      }),
    )
    appendSpecReviewIterationToState(root, GATE_BRANCH, {
      verdict: "REQUEST CHANGES",
      correlation_id: firstCorrelation,
      finding_count: 2,
      finding_ids: ["1", "2"],
    })
    // The human adjudicates iteration 1: all-override.
    recordSpecReviewDecisionsToState(root, GATE_BRANCH, [
      { finding_id: "1", decision: "override", reason: "accepted risk" },
      { finding_id: "2", decision: "override", reason: "accepted risk" },
    ])

    // The decided gate no longer holds the dispatch guard (HARDEN-003).
    expect(specReviewDispatchGuard(root, "specs", GATE_BRANCH)).toBeNull()

    // The next dispatch consolidates as iteration 2 — appended, never an
    // overwrite of the adjudicated iteration 1.
    const secondCorrelation = await setupDispatchSession(
      [runFixture({ agent: "divisor-guard" })],
      { mode: "specs", plan: explicitRunPlan("divisor-guard") },
    )
    const tool = createConsolidateDispatchTool(createFinalizationDependencies(root))
    const result = await tool.execute(
      {
        provenance: {
          branch: GATE_BRANCH,
          commit: "b".repeat(40),
          workflow_id: "uf.review-council",
        },
      },
      noopContext(),
    )
    const finalization = JSON.parse(result.output) as { operation_verdict: string }
    expect(finalization.operation_verdict).toBe("APPROVE")

    const record = readSpecReviewGateRecord(root, GATE_BRANCH)
    expect(Object.keys(record!.iterations)).toEqual(["1", "2"])
    expect(record!.iterations["1"]?.correlation_id).toBe(firstCorrelation)
    expect(record!.iterations["2"]?.correlation_id).toBe(secondCorrelation)
    expect(record!.iterations["2"]?.verdict).toBe("APPROVE")
    // The iteration-1 adjudication stays scoped alongside the new verdict.
    expect(record!.decisions?.iteration).toBe("1")
  })
})
