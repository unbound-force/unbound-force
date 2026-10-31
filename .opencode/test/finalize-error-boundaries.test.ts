import { mkdtemp, readdir, rm } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"

import { afterEach, describe, expect, it } from "vitest"

import { DISPATCH_ARTIFACT_DIRECTORY } from "../lib/review-dispatch-artifact.js"
import {
  createFinalizationDependencies,
  finalizeReviewDispatch,
  type FinalizationDependencies,
  type ReviewDispatchPayload,
} from "../plugins/review-dispatch/index.js"
import samplePayload from "../../schemas/review-dispatch/samples/sample-review-dispatch.json"

const scratchDirectories: string[] = []
const COMMIT = "2222222222222222222222222222222222222222"

afterEach(async () => {
  await Promise.all(scratchDirectories.splice(0).map((directory) => rm(directory, { recursive: true, force: true })))
})

async function scratchDirectory(): Promise<string> {
  const directory = await mkdtemp(join(tmpdir(), "uf-finalize-boundaries-"))
  scratchDirectories.push(directory)
  return directory
}

function dependencies(projectRoot: string): FinalizationDependencies {
  const base = createFinalizationDependencies(projectRoot)
  return { ...base, now: (): Date => new Date("2026-10-01T12:00:04Z"), uuid: (): string => "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb" }
}

function payload(): ReviewDispatchPayload {
  return structuredClone(samplePayload) as unknown as ReviewDispatchPayload
}

function provenance(): { branch: string; commit: string; workflow_id: string } {
  return { branch: "feature/review-fanout", commit: COMMIT, workflow_id: "boundaries" }
}

async function errorsFor(payload: ReviewDispatchPayload): Promise<string[]> {
  const result = await finalizeReviewDispatch({ payload, provenance: provenance() }, dependencies(await scratchDirectory()))
  expect(result.status).toBe("failed")
  return result.errors
}

describe("finalize run-state semantic boundaries", () => {
  it("rejects a run that finished before it started", async () => {
    const data = payload()
    data.runs[0]!.finished_at = "2026-10-01T11:00:00Z"
    expect((await errorsFor(data)).join(" ")).toContain("finished before it started")
  })

  it("rejects a successful run with inconsistent timestamps or verdict", async () => {
    const data = payload()
    data.runs[0]!.workflow_verdict = null
    expect((await errorsFor(data)).join(" ")).toContain("has inconsistent timestamps, error, or verdict")
  })

  it("rejects a successful run that carries an error", async () => {
    // Task 1.1 verdict / RD-FR-026: a run reclassified from failed to
    // success after a post-submission abort MUST reach the payload with
    // a null error — validateRunState rejects the pairing, so the
    // original executor error may only live in the run file's
    // post_submission_abort audit record, never in the payload.
    const data = payload()
    data.runs[0]!.error = {
      code: "child_message_error",
      message: "session aborted during final response generation",
      retryable: false,
    }
    expect((await errorsFor(data)).join(" ")).toContain(
      "successful run 11111111-1111-4111-8111-111111111111 has inconsistent timestamps, error, or verdict",
    )
  })

  it("rejects a successful run with null started_at or finished_at", async () => {
    const missingStart = payload()
    missingStart.runs[0]!.started_at = null
    expect((await errorsFor(missingStart)).join(" ")).toContain(
      "successful run 11111111-1111-4111-8111-111111111111 has inconsistent timestamps, error, or verdict",
    )

    const missingFinish = payload()
    missingFinish.runs[0]!.finished_at = null
    expect((await errorsFor(missingFinish)).join(" ")).toContain(
      "successful run 11111111-1111-4111-8111-111111111111 has inconsistent timestamps, error, or verdict",
    )
  })

  it("rejects a non-success run that votes or emits findings", async () => {
    const data = payload()
    data.runs[0]!.status = "failed"
    expect((await errorsFor(data)).join(" ")).toContain("must not vote or emit findings")
  })

  it("rejects a failed run without an error", async () => {
    const data = payload()
    data.runs[0]!.status = "failed"
    data.runs[0]!.workflow_verdict = null
    data.runs[0]!.findings = []
    data.findings = []
    expect((await errorsFor(data)).join(" ")).toContain("requires an error")
  })

  it("rejects a cancelled run without an error", async () => {
    const data = payload()
    data.runs[0]!.status = "cancelled"
    data.runs[0]!.workflow_verdict = null
    data.runs[0]!.findings = []
    data.runs[0]!.error = null
    data.findings = []
    data.run_counts.success = 0
    data.run_counts.cancelled = 1
    expect((await errorsFor(data)).join(" ")).toContain(
      "cancelled run 11111111-1111-4111-8111-111111111111 requires an error",
    )
  })

  it("rejects a skipped run that carries usage data", async () => {
    const data = payload()
    data.runs[0]!.status = "skipped"
    data.runs[0]!.workflow_verdict = null
    data.runs[0]!.findings = []
    data.findings = []
    data.runs[0]!.error = null
    expect((await errorsFor(data)).join(" ")).toContain("must not contain error or usage data")
  })
})

describe("finalize run-model provenance boundaries", () => {
  it("rejects a host run that carries requested model provenance", async () => {
    const data = payload()
    data.plan[0]!.source = "host"
    data.plan[0]!.model = null
    data.plan[0]!.variant = null
    data.runs[0]!.source = "host"
    data.runs[0]!.resolved_parent_model = "opencode-go/deepseek-v4-pro"
    data.runs[0]!.resolved_parent_variant = "high"
    expect((await errorsFor(data)).join(" ")).toContain("must retain null requested model provenance")
  })

  it("rejects a successful host run without a resolved parent model", async () => {
    const data = payload()
    data.plan[0]!.source = "host"
    data.plan[0]!.model = null
    data.plan[0]!.variant = null
    data.runs[0]!.source = "host"
    data.runs[0]!.requested_model = null
    data.runs[0]!.provider = null
    data.runs[0]!.model_id = null
    data.runs[0]!.variant = null
    data.runs[0]!.resolved_parent_model = null
    data.runs[0]!.resolved_parent_variant = "high"
    expect(await errorsFor(data)).toContain(
      "successful host run 11111111-1111-4111-8111-111111111111 requires resolved parent model",
    )
  })

  it("accepts a successful host run whose parent model has no variant", async () => {
    const data = payload()
    data.plan[0]!.source = "host"
    data.plan[0]!.model = null
    data.plan[0]!.variant = null
    data.runs[0]!.source = "host"
    data.runs[0]!.requested_model = null
    data.runs[0]!.provider = null
    data.runs[0]!.model_id = null
    data.runs[0]!.variant = null
    data.runs[0]!.resolved_parent_model = "enmaas/rits/zai-org/glm-5-3"
    data.runs[0]!.reported_model = "enmaas/rits/zai-org/glm-5-3"

    const result = await finalizeReviewDispatch(
      { payload: data, provenance: provenance() },
      dependencies(await scratchDirectory()),
    )

    expect(result.status).toBe("success")
    expect(result.errors).toEqual([])
  })

  it("rejects a successful host-resolved advisor run without a resolved parent model", async () => {
    const data = payload()
    data.plan[0]!.source = "advisor"
    data.plan[0]!.model = null
    data.plan[0]!.variant = null
    data.runs[0]!.source = "advisor"
    data.runs[0]!.requested_model = null
    data.runs[0]!.provider = null
    data.runs[0]!.model_id = null
    data.runs[0]!.variant = null
    data.runs[0]!.resolved_parent_model = null
    data.runs[0]!.resolved_parent_variant = "high"
    expect(await errorsFor(data)).toContain(
      "successful host-resolved advisor run 11111111-1111-4111-8111-111111111111 requires resolved parent model",
    )
  })

  it("accepts a successful host-resolved advisor run whose parent model has no variant", async () => {
    const data = payload()
    data.plan[0]!.source = "advisor"
    data.plan[0]!.model = null
    data.plan[0]!.variant = null
    data.runs[0]!.source = "advisor"
    data.runs[0]!.requested_model = null
    data.runs[0]!.provider = null
    data.runs[0]!.model_id = null
    data.runs[0]!.variant = null
    data.runs[0]!.resolved_parent_model = "enmaas/rits/zai-org/glm-5-3"
    data.runs[0]!.reported_model = "enmaas/rits/zai-org/glm-5-3"

    const result = await finalizeReviewDispatch(
      { payload: data, provenance: provenance() },
      dependencies(await scratchDirectory()),
    )

    expect(result.status).toBe("success")
    expect(result.errors).toEqual([])
  })

  it("rejects an explicit run carrying resolved parent provenance", async () => {
    const data = payload()
    data.runs[0]!.resolved_parent_model = "opencode-go/deepseek-v4-pro"
    data.runs[0]!.resolved_parent_variant = "high"
    expect((await errorsFor(data)).join(" ")).toContain("must not contain resolved parent provenance")
  })

  it("rejects a successful run without reported model provenance", async () => {
    const data = payload()
    data.runs[0]!.reported_model = null
    expect((await errorsFor(data)).join(" ")).toContain("requires reported model provenance")
  })

  it("rejects a model-mismatch run that is not failed", async () => {
    const data = payload()
    data.runs[0]!.reported_model = "opencode-go/qwen3-max"
    data.runs[0]!.model_mismatch = true
    expect((await errorsFor(data)).join(" ")).toContain(
      "model mismatch run 11111111-1111-4111-8111-111111111111 must be failed",
    )
  })
})

describe("finalize plan and run identity boundaries", () => {
  it("rejects duplicate plan entries", async () => {
    const data = payload()
    data.plan.push(structuredClone(data.plan[0]!))
    expect((await errorsFor(data)).join(" ")).toContain("duplicate plan entry")
  })

  it("rejects an included plan entry carrying validation errors", async () => {
    const data = payload()
    data.plan[0]!.validation_errors = ["broken"]
    expect((await errorsFor(data)).join(" ")).toContain("has validation errors")
  })

  it("rejects duplicate run ids", async () => {
    const data = payload()
    const extra = structuredClone(data.runs[0]!)
    extra.sequence = 2
    data.plan.push(structuredClone(data.plan[0]!) as ReviewDispatchPayload["plan"][number])
    data.plan[1]!.sequence = 2
    data.runs.push(extra)
    data.run_counts.total = 2
    data.run_counts.success = 2
    expect((await errorsFor(data)).join(" ")).toContain("duplicate run id")
  })

  it("rejects a run with no included plan entry", async () => {
    const data = payload()
    data.runs[0]!.source = "advisor"
    expect((await errorsFor(data)).join(" ")).toContain("has no included plan entry")
  })

  it("rejects a plan entry with more than one run", async () => {
    const data = payload()
    const extra = structuredClone(data.runs[0]!)
    extra.run_id = "22222222-2222-4222-8222-222222222222"
    data.runs.push(extra)
    data.run_counts.total = 2
    data.run_counts.success = 2
    expect((await errorsFor(data)).join(" ")).toContain("has more than one run")
  })

  it("rejects a run verdict outside its workflow kind", async () => {
    const data = payload()
    data.runs[0]!.workflow_verdict = "VALID" as never
    expect((await errorsFor(data)).join(" ")).toContain("has a verdict outside")
  })

  it("rejects an included plan entry without a run", async () => {
    const data = payload()
    const extra = structuredClone(data.plan[0]!)
    extra.sequence = 2
    data.plan.push(extra)
    expect((await errorsFor(data)).join(" ")).toContain("has no run")
  })
})

describe("finalize finding and context boundaries", () => {
  it("rejects a consolidated finding referencing a non-success run", async () => {
    const data = payload()
    data.runs[0]!.status = "failed"
    data.runs[0]!.workflow_verdict = null
    data.runs[0]!.findings = []
    data.runs[0]!.error = { code: "calculation_failure", message: "stop", retryable: false }
    data.run_counts.success = 0
    data.run_counts.failed = 1
    data.workflow_result = { kind: "council", value: "INCONCLUSIVE" }
    data.verdict = "INCONCLUSIVE"
    expect((await errorsFor(data)).join(" ")).toContain("references a non-successful or unknown run")
  })

  it("rejects an advisory referencing a non-success run", async () => {
    const data = payload()
    data.runs[0]!.status = "failed"
    data.runs[0]!.workflow_verdict = null
    data.runs[0]!.findings = []
    data.runs[0]!.error = { code: "calculation_failure", message: "stop", retryable: false }
    data.findings = []
    data.advisories = [{ severity: "LOW", description: "still advised", run_ids: [data.runs[0]!.run_id] }]
    data.run_counts.success = 0
    data.run_counts.failed = 1
    data.workflow_result = { kind: "council", value: "INCONCLUSIVE" }
    data.verdict = "INCONCLUSIVE"
    expect((await errorsFor(data)).join(" ")).toContain("advisory references a non-successful or unknown run")
  })

  it("rejects a run finding missing from consolidated findings", async () => {
    const data = payload()
    data.findings = []
    data.verdict = "REQUEST CHANGES"
    expect((await errorsFor(data)).join(" ")).toContain("run finding is missing from consolidated findings")
  })

  it("rejects an artifact commit that differs from the reviewed head SHA", async () => {
    const data = payload()
    const result = await finalizeReviewDispatch(
      { payload: data, provenance: { branch: "feature/review-fanout", commit: "a".repeat(40), workflow_id: "boundaries" } },
      dependencies(await scratchDirectory()),
    )
    expect(result.status).toBe("failed")
    expect(result.errors.join(" ")).toContain("artifact commit must equal the immutable reviewed head SHA")
  })
})

describe("finalize zero-success fail-closed boundary (RD-FR-007)", () => {
  it("fails closed on a zero-success payload claiming an approving verdict", async () => {
    // RD-FR-007: a workflow MUST require at least one successful
    // assessment before an approving result. With every orphaned
    // submission excluded (RD-FR-025) a naive severity-derived verdict
    // would be APPROVE — the finalizer must reject that claim, report
    // the calculated fail-closed verdict, and persist no artifact.
    // fix-review-dispatch-recovery fixes the bricked dispatch upstream
    // (reclassification and exclusion); this boundary itself is
    // unchanged.
    const data = payload()
    data.runs[0]!.status = "failed"
    data.runs[0]!.workflow_verdict = null
    data.runs[0]!.findings = []
    data.runs[0]!.error = { code: "calculation_failure", message: "consolidation arithmetic failed", retryable: false }
    data.findings = []
    data.run_counts.success = 0
    data.run_counts.failed = 1
    data.verdict = "APPROVE"
    data.verdict_reason = "No findings requiring changes"
    data.workflow_result = { kind: "council", value: "APPROVE" }

    const projectRoot = await scratchDirectory()
    const result = await finalizeReviewDispatch({ payload: data, provenance: provenance() }, dependencies(projectRoot))

    expect(result.status).toBe("failed")
    expect(result.operation_verdict).toBe("INCONCLUSIVE")
    expect(result.artifact_path).toBeNull()
    expect(result.review_dispatch).toBeNull()
    expect(result.errors).toEqual([
      "verdict is APPROVE; expected INCONCLUSIVE",
      "workflow_result.value is APPROVE; expected INCONCLUSIVE",
    ])
    // Fail-closed is fail-persisted: no artifact directory is created.
    expect(await readdir(join(projectRoot, DISPATCH_ARTIFACT_DIRECTORY)).catch(() => null)).toBeNull()
  })
})

describe("finalize strict schema boundaries", () => {
  it("rejects findings with a severity outside the CRITICAL|HIGH|MEDIUM|LOW enum", async () => {
    // Issue #751 regression: a child submitted a finding whose severity
    // fell outside the enum, the harvest persisted it, and the
    // finalizer's closed schema rejected the whole dispatch payload.
    // The fix for #751 lives upstream (post-submission reclassification
    // and success-scoped consolidation) — this rejection itself MUST
    // remain exactly as strict.
    const data = payload()
    data.runs[0]!.findings[0]!.severity = "BLOCKER" as never
    data.findings[0]!.severity = "BLOCKER" as never

    const result = await finalizeReviewDispatch(
      { payload: data, provenance: provenance() },
      dependencies(await scratchDirectory()),
    )

    expect(result.status).toBe("failed")
    expect(result.operation_verdict).toBe("INCONCLUSIVE")
    // Structural failures never surface the payload correlation id.
    expect(result.correlation_id).toBeNull()
    expect(result.artifact_path).toBeNull()
    expect(result.review_dispatch).toBeNull()
    expect(result.errors).toHaveLength(1)
    expect(result.errors[0]).toContain("validate review dispatch input: ")
    expect(result.errors[0]).toContain(
      'findings.0.severity: Invalid option: expected one of "CRITICAL"|"HIGH"|"MEDIUM"|"LOW"',
    )
    expect(result.errors[0]).toContain(
      'runs.0.findings.0.severity: Invalid option: expected one of "CRITICAL"|"HIGH"|"MEDIUM"|"LOW"',
    )
  })

  it("rejects run-file-only audit and recovery fields leaking into a payload run", async () => {
    // fix-review-dispatch-recovery adds run-file-only fields —
    // post_submission_abort (RD-FR-026), discarded_submissions
    // (DR-FR-001), and prompt (design D4) — plus result-level
    // observability shapes (proposals, warnings). DispatchRunSchema is
    // strict, so none of them can ever ride into a finalization
    // payload: each injected shape is a structural failure.
    const injections: ReadonlyArray<readonly [string, unknown]> = [
      [
        "post_submission_abort",
        {
          original_status: "failed",
          error_code: "child_message_error",
          error_message: "session aborted during final response generation",
          note: "Reclassified per RD-FR-026; run-file audit record only.",
        },
      ],
      [
        "discarded_submissions",
        {
          agent: "divisor-guard",
          run_id: "11111111-1111-4111-8111-111111111111",
          discarded_findings: 2,
          discarded_proposals: 1,
          discarded_at: "2026-10-01T12:00:02Z",
          note: "Discarded per DR-FR-001; run-file audit record only.",
        },
      ],
      ["prompt", "You are divisor-guard. Review the changed files."],
      ["proposals", [{ information: "Exclude orphaned submissions before finalization.", tag: "review-dispatch" }]],
      [
        "warnings",
        [
          {
            type: "orphaned_submissions_excluded",
            agent: "divisor-guard",
            run_id: "11111111-1111-4111-8111-111111111111",
            discarded_findings: 2,
            discarded_proposals: 1,
          },
        ],
      ],
    ]

    for (const [field, value] of injections) {
      const data = payload()
      ;(data.runs[0] as unknown as Record<string, unknown>)[field] = value
      const errors = await errorsFor(data)
      expect(errors).toHaveLength(1)
      expect(errors[0]).toContain("validate review dispatch input: ")
      expect(errors[0]).toContain(`runs.0: Unrecognized key: "${field}"`)
    }
  })

  it("rejects consolidate result-wrapper observability fields on the payload", async () => {
    // consolidate_dispatch returns { ...result, proposals?, warnings? }
    // (each key omitted when empty — PR #776): the observability added by
    // this change rides on the result
    // wrapper only. A caller that forwards those wrapper fields as
    // payload keys must be rejected — ReviewDispatchPayloadSchema is
    // strict at the root too.
    const data = payload()
    const injected = data as unknown as Record<string, unknown>
    injected.warnings = [
      {
        type: "orphaned_submissions_excluded",
        agent: "divisor-guard",
        run_id: "11111111-1111-4111-8111-111111111111",
        discarded_findings: 1,
        discarded_proposals: 0,
      },
    ]
    injected.proposals = [{ agent: "divisor-guard", information: "lesson", tag: "review-dispatch" }]

    const errors = await errorsFor(data)
    expect(errors).toHaveLength(1)
    expect(errors[0]).toContain("validate review dispatch input: ")
    expect(errors[0]).toContain('payload: Unrecognized keys:')
    expect(errors[0]).toContain('"warnings"')
    expect(errors[0]).toContain('"proposals"')
  })
})
