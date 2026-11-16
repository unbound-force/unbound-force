// ---------------------------------------------------------------------------
// Escalated-advisory relay and advancement hold — SYNTH-004 coverage
// (synthetic-compaction-continue-prompt-rec, task 5.3).
//
// Spec: openspec/changes/synthetic-compaction-continue-prompt-rec/specs/
// checkpoint-gate/spec.md (SYNTH-004), scenario "Escalated advisories are
// relayed before implementation" — the issue #1118 incident: a spec review
// consolidated to APPROVE WITH ADVISORIES (correlation 8ef4f276-…, 16
// findings, 2 proposals) whose advisories were never relayed, overridden by
// a synthetic compaction continue prompt.
//
// Covers the widened relay predicate (task 5.1): relay_spec_review_gate
// treats an APPROVE WITH ADVISORIES latest — the persisted canonical
// decision ESCALATED, because the review-dispatch finalizer's
// canonicalDecision mapping sends every AWA consolidation there — through
// the same relay path and output shape as REQUEST CHANGES, with the
// advisories surfacing as artifact-sourced findings; a plain APPROVE
// (canonical decision APPROVED) reports gate_paused: false.
//
// Covers the advancement hold (task 5.2): while the latest consolidated
// spec-review result is AWA and no human-decision reference per SYNTH-001
// covers the escalation, checkpoint rejects the advancement steps
// `implement` and `implement-phase-N` (both the literal step and the
// `phase` argument path) with escalated_advisory_unrelayed naming the
// correlation and finding count — a hold over the advancement steps, not
// an allowlist of permitted ones, so every other declared step still
// records. A covered call stamps the reference as provenance.
// ---------------------------------------------------------------------------

import { randomUUID } from "node:crypto"
import { mkdir, mkdtemp, readFile, rm, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"

import { afterEach, describe, expect, it } from "vitest"

import {
  appendHumanDecisions,
  appendSpecReviewIterationToState,
  readState,
  recordSpecReviewDecisionsToState,
} from "../lib/progress-state.js"
import {
  HUMAN_DECISION_ID_PATTERN,
  isEscalatedAdvisorySpecVerdict,
  specReviewFindingIds,
  type HumanDecisionAppendInput,
} from "../lib/uf-workflow-types.js"
import {
  createCheckpointTool,
  createRelaySpecReviewGateTool,
  type GitExec,
  type ProgressState,
  type RelaySpecReviewGateData,
  type ToolResult,
} from "../plugins/uf-workflow/index.js"

// ---------------------------------------------------------------------------
// Fixtures — the issue #1118 incident shape
// ---------------------------------------------------------------------------

/** The incident's branch-shaped fixture key (sanitizes to INCIDENT_KEY). */
const INCIDENT_BRANCH = "opsx/advisory-fixture"

/** Sanitized state key every tool call in this suite resolves to. */
const INCIDENT_KEY = "opsx-advisory-fixture"

/**
 * The incident's dispatch correlation (spec SYNTH-004 scenario): an
 * `8ef4f276-…` style UUID identifying the AWA consolidation whose
 * advisories were never relayed.
 */
const INCIDENT_CORRELATION = "8ef4f276-1111-4111-8111-111111111111"

/** The incident's consolidated advisory set: 2 HIGH + 14 MEDIUM findings. */
const INCIDENT_FINDINGS = [
  ...Array.from({ length: 2 }, (_, index) => ({
    severity: "HIGH",
    category: "advisory-relay",
    description: `Escalated high advisory ${index + 1} of the incident replay`,
    root_cause: "Consolidated as an advisory without a relayed human decision",
    file: `specs/advisory-${index + 1}.md`,
    line: index + 1,
  })),
  ...Array.from({ length: 14 }, (_, index) => ({
    severity: "MEDIUM",
    category: "advisory-relay",
    description: `Escalated medium advisory ${index + 3} of the incident replay`,
    root_cause: "Consolidated as an advisory without a relayed human decision",
    file: `specs/advisory-${index + 3}.md`,
    line: index + 3,
  })),
]

/** The incident's two consolidated lesson proposals (artifact fidelity). */
const INCIDENT_PROPOSALS = [
  {
    agent: "divisor-scribe",
    information: "Document the escalated-advisory relay path in the orchestrator guide.",
    tag: "review-dispatch",
    category: "documentation",
  },
  {
    agent: "divisor-curator",
    information: "File a docs issue for the un-relayed advisory incident pattern.",
    tag: "review-dispatch",
    category: "documentation",
  },
]

/** Temp roots created during a test, removed in afterEach (isolation). */
const ROOTS: string[] = []

/** Create an isolated project root holding no `.uf/progress/` yet. */
async function newRoot(): Promise<string> {
  const root = await mkdtemp(join(tmpdir(), "uf-escalated-advisory-"))
  ROOTS.push(root)
  return root
}

afterEach(async () => {
  await Promise.all(
    ROOTS.splice(0).map((root) => rm(root, { recursive: true, force: true })),
  )
})

/** The git exec never runs: every tool call resolves the fixture branch. */
const branchExec: GitExec = () => INCIDENT_BRANCH

/** The registration dependency shape: real-fs artifact reader. */
const realReadText = (path: string): Promise<string> => readFile(path, "utf8")

/** One orchestrator-relay append input, the envelope's write surface. */
function relay(
  question: string,
  answer: string,
): HumanDecisionAppendInput {
  return { source: "orchestrator-relay", question, answer }
}

/** Write one artifact envelope into the scratch dispatch store. */
async function writeArtifact(projectRoot: string, correlationId: string, raw: string): Promise<void> {
  const storeDir = join(projectRoot, ".uf", "artifacts", "dispatch")
  await mkdir(storeDir, { recursive: true })
  await writeFile(join(storeDir, `${correlationId}.json`), raw, "utf8")
}

/** Input context shaped like the incident's spec-review dispatch. */
function incidentInputContext() {
  return {
    kind: "pr" as const,
    pr_number: 1118,
    base_ref: "main",
    base_sha: "a".repeat(40),
    head_ref: INCIDENT_BRANCH,
    head_sha: "b".repeat(40),
  }
}

/** Change profile shaped like the incident's spec-review dispatch. */
function incidentChangeProfile() {
  return {
    kind: "diff" as const,
    lines: 400,
    files: 12,
    components: 3,
    security_sensitive: false,
    user_facing: false,
    categories: ["standard" as const, "documentation" as const],
    tier: "standard" as const,
  }
}

/**
 * Hand-build the incident's review-dispatch artifact envelope: AWA verdict,
 * 16 advisory findings, and the 2 consolidated proposals, at the recorded
 * `8ef4f276-…` correlation — the consolidation whose advisories the
 * incident never relayed.
 */
function incidentArtifactEnvelope(): string {
  return JSON.stringify({
    artifact_type: "review-dispatch",
    schema_version: "1.0.0",
    timestamp: "2026-11-13T00:00:00.000Z",
    context: {
      branch: INCIDENT_BRANCH,
      commit: "b".repeat(40),
      correlation_id: INCIDENT_CORRELATION,
      workflow_id: "uf.review-council",
    },
    payload: {
      command: "review-council",
      mode: "specs",
      full: false,
      input_context: incidentInputContext(),
      change_profile: incidentChangeProfile(),
      plan_version: 1,
      plan: [],
      runs: [],
      coverage: { preflight_verdict: "PASS", checks_total: 5, checks_passed: 5 },
      findings: INCIDENT_FINDINGS,
      advisories: [],
      proposals: INCIDENT_PROPOSALS,
      verdict: "APPROVE WITH ADVISORIES",
      verdict_reason: "16 advisory findings consolidated with no blocking severity",
      run_counts: { total: 5, success: 5, failed: 0, skipped: 0, budget_skipped: 0, limit_skipped: 0, cancelled: 0 },
      workflow_result: { kind: "council", value: "APPROVE WITH ADVISORIES" },
      correlation_id: INCIDENT_CORRELATION,
    },
  })
}

/**
 * Seed the incident's gate: the dispatch artifact (AWA verdict, 16
 * findings, 2 proposals) plus iteration 1 recorded at the same correlation.
 * The append's healing pass records the artifact-derived entry; the
 * matching correlation keeps it a single iteration.
 */
async function recordIncidentGate(projectRoot: string): Promise<void> {
  await writeArtifact(projectRoot, INCIDENT_CORRELATION, incidentArtifactEnvelope())
  appendSpecReviewIterationToState(projectRoot, INCIDENT_BRANCH, {
    verdict: "APPROVE WITH ADVISORIES",
    correlation_id: INCIDENT_CORRELATION,
    finding_count: 16,
    finding_ids: specReviewFindingIds(16),
  })
}

/** Invoke the checkpoint tool and parse its JSON envelope. */
async function invokeCheckpoint(
  root: string,
  args: Record<string, unknown>,
): Promise<ToolResult<ProgressState>> {
  const toolInstance = createCheckpointTool(branchExec, root)
  const result = await (
    toolInstance as {
      execute: (
        args: Record<string, unknown>,
        ctx: { sessionID: string },
      ) => Promise<{ output: string }>
    }
  ).execute(args, { sessionID: "spec-review-escalated-advisory" })
  return JSON.parse(result.output) as ToolResult<ProgressState>
}

/** Invoke the relay tool for the fixture branch and parse its envelope. */
async function invokeRelay(root: string): Promise<ToolResult<RelaySpecReviewGateData>> {
  const toolInstance = createRelaySpecReviewGateTool(branchExec, root, { readText: realReadText })
  const result = await (
    toolInstance as {
      execute: (
        args: Record<string, unknown>,
        ctx: { sessionID: string },
      ) => Promise<{ output: string }>
    }
  ).execute({ command: "uf.unleash" }, { sessionID: "spec-review-escalated-advisory" })
  return JSON.parse(result.output) as ToolResult<RelaySpecReviewGateData>
}

/** Narrowing helper for the relay envelope's data payload. */
function okRelayData(
  result: ToolResult<RelaySpecReviewGateData>,
): RelaySpecReviewGateData {
  expect(result.status).toBe("ok")
  if (result.status !== "ok" || result.data === undefined) {
    throw new Error(`expected ok data, received ${JSON.stringify(result)}`)
  }
  return result.data
}

// ---------------------------------------------------------------------------
// Battery
// ---------------------------------------------------------------------------

describe("escalated-advisory relay and advancement hold (SYNTH-004, task 5.3)", () => {
  it("isEscalatedAdvisorySpecVerdict passes only APPROVE WITH ADVISORIES — every AWA consolidation is ESCALATED", () => {
    // The review-dispatch finalizer's canonicalDecision mapping sends every
    // APPROVE WITH ADVISORIES consolidation to ESCALATED (review-dispatch
    // index.ts: "APPROVE"→APPROVED, "APPROVE WITH ADVISORIES"→ESCALATED,
    // "REQUEST CHANGES"→CHANGES_REQUESTED), so "marked ESCALATED" is not a
    // distinguishing subset of advisory results: the verdict string alone
    // decides. The predicate mirrors that mapping exactly.
    expect(isEscalatedAdvisorySpecVerdict("APPROVE WITH ADVISORIES")).toBe(true)
    // A plain APPROVE (canonical decision APPROVED) is not escalated.
    expect(isEscalatedAdvisorySpecVerdict("APPROVE")).toBe(false)
    expect(isEscalatedAdvisorySpecVerdict("REQUEST CHANGES")).toBe(false)
    expect(isEscalatedAdvisorySpecVerdict("INCONCLUSIVE")).toBe(false)
    expect(isEscalatedAdvisorySpecVerdict("UNAVAILABLE")).toBe(false)
  })

  it("relays the ESCALATED 16-finding incident through the same relay path as REQUEST CHANGES (issue #1118)", async () => {
    const root = await newRoot()
    await recordIncidentGate(root)

    const data = okRelayData(await invokeRelay(root))

    // The AWA latest pauses the relay — the same gate_paused: true output
    // shape a holding REQUEST CHANGES produces (task 5.1's widened
    // predicate), never the advancing shape.
    expect(data.gate_paused).toBe(true)
    expect(data.iteration).toBe(1)
    expect(data.verdict).toBe("APPROVE WITH ADVISORIES")
    expect(data.correlation_id).toBe(INCIDENT_CORRELATION)
    expect(data.finding_count).toBe(16)
    // Advisories are not per-finding adjudicated: no undecided id list.
    expect(data.undecided_finding_ids).toEqual([])

    // The relay is artifact-sourced and carries the escalation-specific
    // hold line naming the incident's correlation and advisory count.
    expect(data.relay).toContain(
      `Spec-review gate holds: iteration 1 (APPROVE WITH ADVISORIES), correlation ${INCIDENT_CORRELATION}.`,
    )
    expect(data.relay).toContain("16 advisories consolidated — escalated advisory (SYNTH-004)")
    expect(data.relay).toContain(
      "every APPROVE WITH ADVISORIES consolidation persists as the canonical decision ESCALATED",
    )
    expect(data.relay).toContain("Findings from the dispatch artifact:")
    // Every advisory renders from the artifact with severity, description,
    // and file:line — the deterministic finding list the human must see.
    for (const [index, finding] of INCIDENT_FINDINGS.entries()) {
      expect(data.relay).toContain(
        `${index + 1}. [${finding.severity}] ${finding.description} (${finding.file}:${finding.line}) — undecided`,
      )
    }
  })

  it("reports gate_paused false for a plain APPROVE latest (canonical decision APPROVED)", async () => {
    const root = await newRoot()
    appendSpecReviewIterationToState(root, INCIDENT_BRANCH, {
      verdict: "APPROVE",
      correlation_id: randomUUID(),
      finding_count: 0,
      finding_ids: [],
    })

    const data = okRelayData(await invokeRelay(root))

    expect(data.gate_paused).toBe(false)
    expect(data.relay).toBeNull()
    expect(data.iteration).toBe(1)
    expect(data.verdict).toBe("APPROVE")
    expect(data.finding_count).toBe(0)
    expect(data.undecided_finding_ids).toEqual([])
  })

  it("holds advancement on the ESCALATED 16-finding incident: implement and implement-phase-N reject with escalated_advisory_unrelayed", async () => {
    const root = await newRoot()
    await recordIncidentGate(root)

    // Both writer surfaces of the advancement steps: the literal `implement`
    // step, the literal `implement-phase-N` step, and the `phase` argument
    // path that records `implement-phase-N` directly with no prerequisite
    // `implement` record.
    const advancementCalls = [
      { label: "implement (literal step)", args: { command: "uf.unleash", step: "implement" } },
      {
        label: "implement-phase-3 (literal step)",
        args: { command: "uf.unleash", step: "implement-phase-3" },
      },
      {
        label: "implement-phase-3 (phase argument)",
        args: {
          command: "uf.unleash",
          step: "implement-phase-3",
          phase: { number: 3, verdict: "green", inventory: [] },
        },
      },
    ]
    for (const call of advancementCalls) {
      const result = await invokeCheckpoint(root, call.args)
      expect(result.status, call.label).toBe("error")
      if (result.status !== "error") return
      expect(result.code, call.label).toBe("escalated_advisory_unrelayed")
      expect(result.retryable, call.label).toBe(false)
      // The refusal names the holding verdict, the incident's correlation,
      // and the finding count — the relay's identifying payload.
      expect(result.message, call.label).toContain("APPROVE WITH ADVISORIES")
      expect(result.message, call.label).toContain(`correlation ${INCIDENT_CORRELATION}`)
      expect(result.message, call.label).toContain("16 findings")
      expect(result.message, call.label).toContain("SYNTH-004")
      // The recovery path names the relay tool, not a bypass.
      expect(result.message, call.label).toContain("relay_spec_review_gate")
    }

    // Nothing persisted across every rejection: no advancement step, no
    // phase record, no provenance stamp — the seeded spec-review record is
    // the only state content.
    const state = readState(root, INCIDENT_KEY, "uf.unleash")
    expect(state?.completed_steps).toEqual([])
    expect(state?.phases).toBeUndefined()
    expect(state?.data.human_decision_ref).toBeUndefined()
  })

  it("advances without a relay or a ref when the latest is a plain APPROVE (canonical decision APPROVED)", async () => {
    const root = await newRoot()
    appendSpecReviewIterationToState(root, INCIDENT_BRANCH, {
      verdict: "APPROVE",
      correlation_id: randomUUID(),
      finding_count: 0,
      finding_ids: [],
    })

    // Spec scenario "Plain APPROVE does not require the relay": the
    // canonical decision APPROVED opens advancement directly — both the
    // literal step and the phase argument path record with no
    // human_decision_ref anywhere on the call.
    const implement = await invokeCheckpoint(root, { command: "uf.unleash", step: "implement" })
    expect(implement.status).toBe("ok")
    const phase = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "implement-phase-1",
      phase: { number: 1, verdict: "green", inventory: [] },
    })
    expect(phase.status).toBe("ok")

    const state = readState(root, INCIDENT_KEY, "uf.unleash")
    expect(state?.completed_steps).toEqual(["implement", "implement-phase-1"])
    expect(state?.phases?.["1"]?.verdict).toBe("green")
    expect(state?.data.human_decision_ref).toBeUndefined()
  })

  it("treats a decided AWA latest identically to an undecided one — no verdict-string subset exempts an escalation", async () => {
    const root = await newRoot()
    await recordIncidentGate(root)
    // A full all-override adjudication of all 16 advisories — the decision
    // state that releases a REQUEST CHANGES gate under HARDEN-002. For an
    // AWA latest it releases nothing: the finalizer maps every AWA
    // consolidation to ESCALATED, so no recorded decision set, escalation
    // marker, or verdict sub-family exempts the result from the relay or
    // the hold. The predicate is the verdict string alone.
    recordSpecReviewDecisionsToState(
      root,
      INCIDENT_BRANCH,
      Array.from({ length: 16 }, (_, index) => ({
        finding_id: String(index + 1),
        decision: "override" as const,
        reason: `accepted advisory risk ${index + 1}`,
      })),
    )

    // The relay still pauses with the identical hold line and payload —
    // the recorded decisions render as marks but never un-pause the gate.
    const data = okRelayData(await invokeRelay(root))
    expect(data.gate_paused).toBe(true)
    expect(data.verdict).toBe("APPROVE WITH ADVISORIES")
    expect(data.correlation_id).toBe(INCIDENT_CORRELATION)
    expect(data.finding_count).toBe(16)
    expect(data.relay).toContain(
      `Spec-review gate holds: iteration 1 (APPROVE WITH ADVISORIES), correlation ${INCIDENT_CORRELATION}.`,
    )
    expect(data.relay).toContain("16 advisories consolidated — escalated advisory (SYNTH-004)")
    expect(data.relay).toContain("1. [HIGH] Escalated high advisory 1 of the incident replay (specs/advisory-1.md:1) — override")

    // The hold still rejects advancement on the same decided record.
    const result = await invokeCheckpoint(root, { command: "uf.unleash", step: "implement" })
    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("escalated_advisory_unrelayed")
    expect(result.message).toContain(`correlation ${INCIDENT_CORRELATION}`)
    expect(result.message).toContain("16 findings")
    expect(readState(root, INCIDENT_KEY, "uf.unleash")?.completed_steps).toEqual([])
  })

  it("rejects a fabricated ref under the escalated-advisory hold with escalated_advisory_unrelayed, never a different code", async () => {
    const root = await newRoot()
    await recordIncidentGate(root)
    // A shape-valid UUID naming no recorded decision is not a covering
    // decision (SYNTH-001): the hold is the last gate before the state
    // write, so its refusal reports escalated_advisory_unrelayed with the
    // coverage detail — it never falls through to missing_human_decision
    // (no override data key rode this call) or any other code.
    const fabricated = "ffffffff-ffff-4fff-8fff-ffffffffffff"
    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "implement",
      human_decision_ref: fabricated,
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("escalated_advisory_unrelayed")
    expect(result.retryable).toBe(false)
    // The shape-validated ref is a bounded decision_id-shaped identifier.
    expect(result.message).toContain(fabricated)
    expect(result.message).toContain("names no decision recorded")
    expect(readState(root, INCIDENT_KEY, "uf.unleash")?.completed_steps).toEqual([])
  })

  it("records a covered advancement when a human_decision_ref resolves, stamping it as provenance", async () => {
    const root = await newRoot()
    await recordIncidentGate(root)
    // The SYNTH-001 accepted-override pattern, applied to the SYNTH-004
    // hold: the orchestrator relays the advisory, the human decision is
    // recorded (here through the envelope's append surface), and the
    // advancement checkpoint cites the stamped decision_id.
    const stamped = await appendHumanDecisions(root, INCIDENT_BRANCH, [
      relay(
        `Proceed past the escalated advisory (correlation ${INCIDENT_CORRELATION}, 16 findings)?`,
        "Yes, proceed to implement.",
      ),
    ])
    const ref = stamped[0]?.decision_id
    expect(ref).toBeDefined()
    expect(HUMAN_DECISION_ID_PATTERN.test(ref ?? "")).toBe(true)

    // Both advancement writer surfaces lift under the same covering ref.
    const implement = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "implement",
      human_decision_ref: ref,
    })
    expect(implement.status).toBe("ok")
    const phase = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "implement-phase-2",
      phase: { number: 2, verdict: "green", inventory: [] },
      human_decision_ref: ref,
    })
    expect(phase.status).toBe("ok")

    // The reference is stamped as provenance — which relayed human
    // decision authorized passing the advisory — and the phase record
    // landed alongside the step.
    const state = readState(root, INCIDENT_KEY, "uf.unleash")
    expect(state?.completed_steps).toEqual(["implement", "implement-phase-2"])
    expect(state?.phases?.["2"]?.verdict).toBe("green")
    expect(state?.data.human_decision_ref).toBe(ref)
  })

  it("still records non-advancement steps under an AWA latest without a ref (hold is over advancement steps only)", async () => {
    const root = await newRoot()
    await recordIncidentGate(root)

    // Spec SYNTH-004: the hold mirrors the HARDEN-002 shape — a hold over
    // the advancement steps, NOT an allowlist of permitted ones. Every
    // other declared step still records, with no human_decision_ref.
    for (const step of ["code-review", "spec-review"]) {
      const result = await invokeCheckpoint(root, { command: "uf.unleash", step })
      expect(result.status, step).toBe("ok")
    }

    const state = readState(root, INCIDENT_KEY, "uf.unleash")
    expect(state?.completed_steps).toEqual(["code-review", "spec-review"])
    expect(state?.data.human_decision_ref).toBeUndefined()
  })
})
