import { mkdtemp, rm } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"

import { afterEach, describe, expect, it } from "vitest"

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

  it("rejects a successful host run without resolved parent provenance", async () => {
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
    data.runs[0]!.resolved_parent_variant = null
    expect((await errorsFor(data)).join(" ")).toContain("requires resolved parent model and variant")
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
