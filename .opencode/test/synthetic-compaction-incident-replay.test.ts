// ---------------------------------------------------------------------------
// Incident checkpoint replay — SYNTH-005 regression coverage
// (synthetic-compaction-continue-prompt-rec, task 6.1).
//
// Spec: openspec/changes/synthetic-compaction-continue-prompt-rec/specs/
// checkpoint-gate/spec.md (SYNTH-005), scenario "Regression fixture replays
// incident JSON" — the repository MUST replay the exact `checkpoint` JSON
// from session ses_e200afc75ffeSuftvVXtQnBJhR (cobalt-crush-dev on
// Muse-Spark-Open, 2026-11-15) and assert `checkpoint` rejects it. The
// incident child treated opencode's post-compaction synthetic continue
// prompt as human authorization for a checklist-gate override: a free-form
// step name `checklist-override-phase8-9` whose `data.checklist_override`
// quoted 'Continue if you have next steps', leaned on an unverified "prior
// session override", and cited an un-relayed APPROVE WITH ADVISORIES
// (ESCALATED) spec-review — all with no human_decision_ref.
//
// The payload below is historical data embedded verbatim from the change's
// proposal.md incident record — not a live file dependency. Distinct from
// the task 2.5 battery (human-decisions-store.test.ts composes shortened
// incident-shaped texts) and the task 4.3 matrix (uf-workflow.test.ts
// covers the step name without the incident data): this fixture pins the
// FULL incident replay. The task 6.2 synthetic-text-alone case
// (synthetic-compaction-prompt-alone.test.ts) and the
// task 5.3 ESCALATED relay battery (spec-review-escalated-advisory.test.ts)
// live in their own suites.
//
// Layering under test (production gate ordering, uf-workflow/index.ts):
// the SYNTH-003 declared-step predicate gates first — the incident payload
// rejects with `undeclared_step` while the same envelope names the missing
// human_decision_ref (SYNTH-001's authorization detail), so the single
// observed rejection is per SYNTH-001 and SYNTH-003 as the SYNTH-005
// scenario requires. The companion case routes the same verbatim incident
// data under a declared step so the SYNTH-001 data-key tripwire's own
// envelope (`missing_human_decision`) is observed for this text: whichever
// layer evaluates first, the incident payload dies and nothing persists.
// ---------------------------------------------------------------------------

import { mkdtemp, rm } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"

import { afterEach, describe, expect, it } from "vitest"

import {
  appendSpecReviewIterationToState,
  readHumanDecisionRecord,
  readState,
  writeState,
} from "../lib/progress-state.js"
import { isSyntheticCompactionPromptAnswer, specReviewFindingIds } from "../lib/uf-workflow-types.js"
import {
  createCheckpointTool,
  type GitExec,
  type ProgressState,
  type ToolResult,
} from "../plugins/uf-workflow/index.js"

// ---------------------------------------------------------------------------
// Fixtures — the ses_e200afc75ffeSuftvVXtQnBJhR incident
// ---------------------------------------------------------------------------

/** The incident's branch-shaped fixture key (sanitizes to INCIDENT_KEY). */
const INCIDENT_BRANCH = "opsx/synthetic-compaction-incident"

/** Sanitized state key every tool call in this suite resolves to. */
const INCIDENT_KEY = "opsx-synthetic-compaction-incident"

/**
 * The incident's un-relayed spec-review consolidation (proposal.md): the
 * `8ef4f276-…` correlation whose APPROVE WITH ADVISORIES result the
 * incident payload cites as cover. Same shape as the task 5.3 battery's
 * correlation (spec-review-escalated-advisory.test.ts).
 */
const INCIDENT_CORRELATION = "8ef4f276-1111-4111-8111-111111111111"

/**
 * The incident's `data.checklist_override` justification, verbatim from
 * the proposal's incident record (historical data, embedded — the fixture
 * is not a live file dependency). 406 characters including the U+2026
 * ellipsis abbreviating the correlation; the length is pinned by the
 * fixture-integrity test so any drift in the embedded text fails loudly.
 */
const INCIDENT_CHECKLIST_OVERRIDE =
  "User explicitly directed continuation via 'Continue if you have next steps' with spec-review APPROVE WITH ADVISORIES (ESCALATED, correlation 8ef4f276-…). Checklists automation-safety.md and requirements.md are reviewer-owned artifacts never marked during 3 spec-review iterations; prior session override for Phase 1 remains in effect. Proceeding with Phase 8 T040-T042 and Phase 9 T043-T048 implementation."

/**
 * The incident payload's `data` argument as the checkpoint tool receives
 * it: a JSON object string whose single key is the override-carrying
 * `checklist_override` (the SYNTH-001 tripwire match).
 */
const INCIDENT_DATA = JSON.stringify({
  checklist_override: INCIDENT_CHECKLIST_OVERRIDE,
})

/**
 * The exact incident checkpoint payload (proposal.md incident record):
 * command `uf.unleash`, undeclared step `checklist-override-phase8-9`, and
 * the incident `data` — with no `human_decision_ref`, because the incident
 * call carried none: the free-text justification was its only claimed
 * authorization.
 */
const INCIDENT_CHECKPOINT_ARGS: Record<string, unknown> = {
  command: "uf.unleash",
  step: "checklist-override-phase8-9",
  data: INCIDENT_DATA,
}

/**
 * The incident session's pre-enforcement progress: spec-review recorded
 * (3 iterations, the latest the un-relayed ESCALATED advisory) and
 * implementation underway through Phase 7 — the baseline the phase 8-9
 * override call purported to extend.
 */
const MID_IMPLEMENTATION_COMPLETED_STEPS: readonly string[] = [
  "clarify",
  "plan",
  "tasks",
  "spec-review",
  ...Array.from({ length: 7 }, (_, index) => `implement-phase-${index + 1}`),
]

/** Temp roots created during a test, removed in afterEach (isolation). */
const ROOTS: string[] = []

/** Create an isolated project root holding no `.uf/progress/` yet. */
async function newRoot(): Promise<string> {
  const root = await mkdtemp(join(tmpdir(), "uf-incident-replay-"))
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

/**
 * Seed the incident session's state: the mid-implementation baseline
 * (phases 1-7 recorded, no checklist override anywhere in `data`) plus
 * the incident's 3 spec-review iterations — two REQUEST CHANGES rounds
 * and the latest APPROVE WITH ADVISORIES (ESCALATED) consolidation at the
 * correlation the incident payload cites. The rejection paths under test
 * return before any gate-record evaluation, so the record is inert
 * context, not a participant; seeding it proves the ESCALATED record
 * does not launder the override. Timestamps are literals plus what the
 * iteration-append machinery stamps — no other clocks, no network.
 */
function seedIncidentState(root: string): void {
  writeState(root, INCIDENT_KEY, "uf.unleash", {
    command: "uf.unleash",
    branch: INCIDENT_BRANCH,
    key: INCIDENT_KEY,
    current_step: "implement-phase-7",
    completed_steps: [...MID_IMPLEMENTATION_COMPLETED_STEPS],
    data: {},
    created_at: "2026-11-15T00:00:00.000Z",
    updated_at: "2026-11-15T00:00:00.000Z",
  })
  appendSpecReviewIterationToState(root, INCIDENT_BRANCH, {
    verdict: "REQUEST CHANGES",
    correlation_id: "11111111-1111-4111-8111-111111111111",
    finding_count: 3,
    finding_ids: specReviewFindingIds(3),
  })
  appendSpecReviewIterationToState(root, INCIDENT_BRANCH, {
    verdict: "REQUEST CHANGES",
    correlation_id: "22222222-2222-4222-8222-222222222222",
    finding_count: 5,
    finding_ids: specReviewFindingIds(5),
  })
  appendSpecReviewIterationToState(root, INCIDENT_BRANCH, {
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
  ).execute(args, { sessionID: "synthetic-compaction-incident-replay" })
  return JSON.parse(result.output) as ToolResult<ProgressState>
}

// ---------------------------------------------------------------------------
// Battery
// ---------------------------------------------------------------------------

describe("synthetic compaction incident replay (SYNTH-005, task 6.1)", () => {
  it("pins the incident payload verbatim — a partial prompt quote the SYNTH-006 deny-list alone cannot catch", () => {
    // Fixture integrity: the embedded text is the incident's, not an
    // approximation. The distinctive markers are the quoted synthetic
    // prompt, the cited un-relayed ESCALATED correlation, the unverified
    // prior-session lean, and the phases the call purported to unlock.
    expect(INCIDENT_CHECKLIST_OVERRIDE).toHaveLength(406)
    expect(INCIDENT_CHECKLIST_OVERRIDE).toContain("'Continue if you have next steps'")
    expect(INCIDENT_CHECKLIST_OVERRIDE).toContain(
      "APPROVE WITH ADVISORIES (ESCALATED, correlation 8ef4f276-…)",
    )
    expect(INCIDENT_CHECKLIST_OVERRIDE).toContain(
      "prior session override for Phase 1 remains in effect",
    )
    expect(INCIDENT_CHECKLIST_OVERRIDE).toContain("Phase 8 T040-T042")
    expect(INCIDENT_CHECKLIST_OVERRIDE).toContain("Phase 9 T043-T048")
    // The incident call carried no human_decision_ref — its only claimed
    // authorization was the free-text justification.
    expect(INCIDENT_CHECKPOINT_ARGS.human_decision_ref).toBeUndefined()
    expect(INCIDENT_CHECKPOINT_ARGS.command).toBe("uf.unleash")
    expect(INCIDENT_CHECKPOINT_ARGS.step).toBe("checklist-override-phase8-9")

    // The incident quoted only the prompt's first clause, so the text does
    // NOT contain the full synthetic compaction prompt — the SYNTH-006
    // answer deny-list (full-prompt containment) would never refuse it.
    // That is exactly why the SYNTH-001 reference requirement and the
    // SYNTH-003 declared-step predicate are the gates that must catch
    // this payload, as the replay tests below pin.
    expect(isSyntheticCompactionPromptAnswer(INCIDENT_CHECKLIST_OVERRIDE)).toBe(false)
  })

  it("rejects the exact incident payload with undeclared_step (SYNTH-003 gates first, naming the missing human_decision_ref) and persists nothing", async () => {
    const root = await newRoot()
    seedIncidentState(root)
    // GIVEN (spec scenario SYNTH-001): no entry in the human-decision
    // record covers the override — the record is the first-use empty case.
    expect(readHumanDecisionRecord(root, INCIDENT_BRANCH)).toEqual([])
    const before = readState(root, INCIDENT_KEY, "uf.unleash")
    expect(before).not.toBeNull()

    const result = await invokeCheckpoint(root, INCIDENT_CHECKPOINT_ARGS)

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    // Production gate ordering: the SYNTH-003 declared-step predicate is
    // the authoritative gate, so the incident payload (undeclared step +
    // override data + no ref) rejects with undeclared_step. The same
    // envelope carries SYNTH-001's authorization detail, so the single
    // observed rejection is per SYNTH-001 and SYNTH-003 — what the
    // SYNTH-005 scenario requires.
    expect(result.code).toBe("undeclared_step")
    expect(result.retryable).toBe(false)
    expect(result.message).toContain("checklist-override-phase8-9")
    expect(result.message).toContain("uf.unleash")
    expect(result.message).toContain("SYNTH-003")
    expect(result.message).toContain("requires human_decision_ref")
    // The step gate is step-scoped: the override-carrying data key is not
    // named in this envelope (the companion test below observes it).
    expect(result.message).not.toContain("checklist_override")
    // Bounded identifiers only — the refusal never echoes the quoted
    // prompt or the incident's justification text.
    expect(result.message).not.toContain("Continue if you have next steps")
    expect(result.message).not.toContain("prior session override")

    // MUST NOT persist progress for that step: the state stands exactly
    // as seeded — no step appended, current_step unchanged, no
    // checklist_override merged, no provenance stamped.
    const after = readState(root, INCIDENT_KEY, "uf.unleash")
    expect(after).toStrictEqual(before)
    expect(after?.completed_steps).not.toContain("checklist-override-phase8-9")
    expect(after?.completed_steps).toEqual([...MID_IMPLEMENTATION_COMPLETED_STEPS])
    expect(after?.current_step).toBe("implement-phase-7")
    expect(after?.data.checklist_override).toBeUndefined()
    expect(after?.data.human_decision_ref).toBeUndefined()
    // The rejection recorded no human decision either.
    expect(readHumanDecisionRecord(root, INCIDENT_BRANCH)).toEqual([])
  })

  it("rejects the same incident data under the declared step implement with missing_human_decision (SYNTH-001 tripwire) and persists nothing", async () => {
    const root = await newRoot()
    seedIncidentState(root)
    const before = readState(root, INCIDENT_KEY, "uf.unleash")
    expect(before).not.toBeNull()

    // Spec scenario SYNTH-001 "Override data under a declared step name
    // is rejected": routing the verbatim incident data under the declared
    // step `implement` removes the SYNTH-003 preemption, so the SYNTH-001
    // data-key tripwire evaluates the incident text itself. This pins the
    // defense-in-depth fallback — if the step-name gate ever stops
    // preempting (or a declared step name is used), the exact incident
    // text still dies at the reference requirement. The tripwire runs
    // before the HARDEN-002/SYNTH-004 gate evaluation, so the code is
    // missing_human_decision even under the seeded ESCALATED latest —
    // never escalated_advisory_unrelayed.
    const result = await invokeCheckpoint(root, {
      command: "uf.unleash",
      step: "implement",
      data: INCIDENT_DATA,
    })

    expect(result.status).toBe("error")
    if (result.status !== "error") return
    expect(result.code).toBe("missing_human_decision")
    expect(result.retryable).toBe(false)
    // The rejection names the declared step and the override-carrying key.
    expect(result.message).toContain("implement")
    expect(result.message).toContain("checklist_override")
    expect(result.message).toContain("SYNTH-001")
    expect(result.message).toContain("requires human_decision_ref")
    // Bounded identifiers only — the quoted prompt and the incident's
    // justification text are never echoed.
    expect(result.message).not.toContain("Continue if you have next steps")
    expect(result.message).not.toContain("prior session override")

    // MUST NOT persist progress for that step: the state stands exactly
    // as seeded — `implement` is not recorded (the baseline carries only
    // the phase records), and no override data or provenance merged.
    const after = readState(root, INCIDENT_KEY, "uf.unleash")
    expect(after).toStrictEqual(before)
    expect(after?.completed_steps).not.toContain("implement")
    expect(after?.current_step).toBe("implement-phase-7")
    expect(after?.data.checklist_override).toBeUndefined()
    expect(after?.data.human_decision_ref).toBeUndefined()
  })
})
