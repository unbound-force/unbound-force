import { mkdtemp, mkdir, readFile, readdir, rm, stat, writeFile } from "node:fs/promises"
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

const INITIAL_CORRELATION_ID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
const TEMPORARY_ID = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
const COLLISION_ID = "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
const SECOND_TEMPORARY_ID = "dddddddd-dddd-4ddd-8ddd-dddddddddddd"
const COMMIT = "2222222222222222222222222222222222222222"

const scratchDirectories: string[] = []

afterEach(async () => {
  await Promise.all(scratchDirectories.splice(0).map((directory) => rm(directory, { recursive: true, force: true })))
})

async function scratchDirectory(): Promise<string> {
  const directory = await mkdtemp(join(tmpdir(), "uf-review-dispatch-"))
  scratchDirectories.push(directory)
  return directory
}

function payloadFixture(): ReviewDispatchPayload {
  return structuredClone(samplePayload) as unknown as ReviewDispatchPayload
}

function emptySuccessfulPayload(): ReviewDispatchPayload {
  const payload = payloadFixture()
  const run = payload.runs[0]!
  run.workflow_verdict = "APPROVE"
  run.findings = []
  payload.findings = []
  payload.advisories = []
  payload.verdict = "APPROVE"
  payload.verdict_reason = "Successful assessment approved."
  payload.workflow_result = { kind: "council", value: "APPROVE" }
  return payload
}

function dependencies(projectRoot: string, identifiers: readonly string[] = [TEMPORARY_ID]): FinalizationDependencies {
  const base = createFinalizationDependencies(projectRoot)
  const queue = [...identifiers]
  return {
    ...base,
    now: (): Date => new Date("2026-10-01T12:00:04Z"),
    uuid: (): string => queue.shift() ?? SECOND_TEMPORARY_ID,
  }
}

function provenance(): { readonly branch: string; readonly commit: string; readonly workflow_id: string } {
  return { branch: "feature/review-fanout", commit: COMMIT, workflow_id: "review-council-pr-42" }
}

async function finalize(
  payload: ReviewDispatchPayload,
  finalizationDependencies: FinalizationDependencies,
) {
  return finalizeReviewDispatch({ payload, provenance: provenance() }, finalizationDependencies)
}

function setWorkflow(
  payload: ReviewDispatchPayload,
  kind: "council" | "triage" | "feedback" | "test-review",
  native: string,
  generic: ReviewDispatchPayload["verdict"],
): void {
  const run = payload.runs[0]!
  run.findings = []
  payload.findings = []
  payload.advisories = []
  if (kind === "triage") {
    const hash = "f".repeat(64)
    payload.command = "triage-issue"
    payload.mode = "triage"
    payload.input_context = {
      kind: "issue",
      issue_number: 42,
      issue_url: "https://github.com/unbound-force/unbound-force/issues/42",
      content_sha256: hash,
    }
    payload.change_profile = {
      kind: "issue",
      text_bytes: 10,
      comment_count: 0,
      content_sha256: hash,
      matched_rules: [],
      security_sensitive: false,
      user_facing: false,
      categories: ["standard"],
      tier: "lightweight",
    }
    payload.workflow_result = { kind, value: native as "VALID" }
  } else if (kind === "feedback") {
    payload.command = "address-feedback"
    payload.mode = "feedback"
    payload.workflow_result = { kind, value: native as "ACCEPT" }
  } else if (kind === "test-review") {
    payload.command = "speckit-testreview"
    payload.mode = "test"
    payload.workflow_result = { kind, value: native as "APPROVE" }
  } else {
    payload.workflow_result = { kind, value: native as "APPROVE" }
  }
  run.workflow_verdict = native as NonNullable<ReviewDispatchPayload["runs"][number]["workflow_verdict"]>
  payload.verdict = generic
  payload.verdict_reason = `Mapped ${native}.`
  if (generic === "APPROVE WITH ADVISORIES") {
    payload.advisories = [{ severity: "LOW", description: "Human review advised.", run_ids: [run.run_id] }]
  }
}

function failedRun(
  payload: ReviewDispatchPayload,
  code: string,
  status: "failed" | "cancelled" = "failed",
): void {
  const run = payload.runs[0]!
  run.status = status
  run.usage = null
  run.error = { code, message: "Assessment could not run.", retryable: true }
  run.workflow_verdict = null
  run.findings = []
  payload.findings = []
  payload.advisories = []
  payload.run_counts = {
    total: 1,
    success: 0,
    failed: status === "failed" ? 1 : 0,
    skipped: 0,
    budget_skipped: 0,
    limit_skipped: 0,
    cancelled: status === "cancelled" ? 1 : 0,
  }
}

describe("finalize review dispatch mappings", () => {
  it.each([
    ["council", "APPROVE", "APPROVE", "APPROVED"],
    ["council", "APPROVE WITH ADVISORIES", "APPROVE WITH ADVISORIES", "ESCALATED"],
    ["council", "REQUEST CHANGES", "REQUEST CHANGES", "CHANGES_REQUESTED"],
    ["test-review", "APPROVE", "APPROVE", "APPROVED"],
    ["test-review", "APPROVE WITH ADVISORIES", "APPROVE WITH ADVISORIES", "ESCALATED"],
    ["test-review", "REQUEST CHANGES", "REQUEST CHANGES", "CHANGES_REQUESTED"],
    ["triage", "VALID", "APPROVE", "APPROVED"],
    ["triage", "INVALID", "REQUEST CHANGES", "CHANGES_REQUESTED"],
    ["triage", "NEEDS-CLARIFICATION", "APPROVE WITH ADVISORIES", "ESCALATED"],
    ["feedback", "ACCEPT", "APPROVE", "APPROVED"],
    ["feedback", "AUTHOR-DECIDES", "APPROVE WITH ADVISORIES", "ESCALATED"],
  ] as const)("maps %s %s to %s and %s", async (kind, native, generic, canonical) => {
    const directory = await scratchDirectory()
    const payload = emptySuccessfulPayload()
    setWorkflow(payload, kind, native, generic)

    const result = await finalize(payload, dependencies(directory))

    expect(result.status).toBe("success")
    expect(result.operation_verdict).toBe(generic)
    expect(result.review_dispatch?.payload.workflow_result.value).toBe(native)
    expect(result.review_verdict?.council_decision).toBe(canonical)
  })

  it.each(["INCONCLUSIVE", "UNAVAILABLE"] as const)("maps %s identically", async (verdict) => {
    const directory = await scratchDirectory()
    const payload = emptySuccessfulPayload()
    failedRun(payload, verdict === "UNAVAILABLE" ? "provider_unavailable" : "calculation_failure")
    payload.workflow_result = { kind: "council", value: verdict }
    payload.verdict = verdict

    const result = await finalize(payload, dependencies(directory))

    expect(result.status).toBe("success")
    expect(result.review_verdict?.council_decision).toBe(verdict)
  })
})

describe("finalize review dispatch semantics", () => {
  it("rejects unknown structural input fields", async () => {
    const directory = await scratchDirectory()
    const payload = { ...payloadFixture(), unexpected: true }

    const result = await finalizeReviewDispatch(
      { payload, provenance: provenance() } as unknown as Parameters<typeof finalizeReviewDispatch>[0],
      dependencies(directory),
    )

    expect(result.status).toBe("failed")
    expect(result.errors.join(" ")).toContain("Unrecognized key")
    expect(result.artifact_path).toBeNull()
  })

  it("accepts a local input context with the uncommitted flag", async () => {
    const directory = await scratchDirectory()
    const payload = emptySuccessfulPayload()
    payload.input_context = {
      kind: "local",
      pr_number: null,
      base_ref: "main",
      base_sha: COMMIT,
      head_ref: "feature/review-fanout",
      head_sha: COMMIT,
      uncommitted: true,
    }

    const result = await finalize(payload, dependencies(directory))

    expect(result.status).toBe("success")
    expect(result.artifact_path).not.toBeNull()
  })

  it("accepts a local input context with uncommitted flag set to false", async () => {
    const directory = await scratchDirectory()
    const payload = emptySuccessfulPayload()
    payload.input_context = {
      kind: "local",
      pr_number: null,
      base_ref: "main",
      base_sha: COMMIT,
      head_ref: "feature/review-fanout",
      head_sha: COMMIT,
      uncommitted: false,
    }

    const result = await finalize(payload, dependencies(directory))

    expect(result.status).toBe("success")
    expect(result.artifact_path).not.toBeNull()
  })

  it("accepts a local input context with uncommitted flag omitted", async () => {
    const directory = await scratchDirectory()
    const payload = emptySuccessfulPayload()
    payload.input_context = {
      kind: "local",
      pr_number: null,
      base_ref: "main",
      base_sha: COMMIT,
      head_ref: "feature/review-fanout",
      head_sha: COMMIT,
      // uncommitted flag is omitted
    }

    const result = await finalize(payload, dependencies(directory))

    expect(result.status).toBe("success")
    expect(result.artifact_path).not.toBeNull()
  })

  it("accepts every terminal state when exact counts reconcile", async () => {
    const directory = await scratchDirectory()
    const payload = emptySuccessfulPayload()
    const basePlan = payload.plan[0]!
    const baseRun = payload.runs[0]!
    const statuses = ["success", "failed", "skipped", "budget_skipped", "limit_skipped", "cancelled"] as const
    payload.plan = statuses.map((status, index) => ({ ...basePlan, sequence: index + 1 }))
    payload.runs = statuses.map((status, index) => {
      const run = structuredClone(baseRun)
      run.run_id = `${String(index + 1).padStart(8, "0")}-1111-4111-8111-111111111111`
      run.sequence = index + 1
      run.status = status
      if (status !== "success") {
        run.workflow_verdict = null
        run.findings = []
        run.usage = null
      }
      run.error = status === "failed" || status === "cancelled"
        ? { code: "calculation_failure", message: "Stopped.", retryable: false }
        : null
      return run
    })
    payload.run_counts = {
      total: 6,
      success: 1,
      failed: 1,
      skipped: 1,
      budget_skipped: 1,
      limit_skipped: 1,
      cancelled: 1,
    }

    const result = await finalize(payload, dependencies(directory))

    expect(result.status).toBe("success")
    expect(result.review_dispatch?.payload.run_counts).toEqual(payload.run_counts)
  })

  it("rejects bad terminal count arithmetic", async () => {
    const directory = await scratchDirectory()
    const payload = emptySuccessfulPayload()
    payload.run_counts.success = 0

    const result = await finalize(payload, dependencies(directory))

    expect(result.status).toBe("failed")
    expect(result.errors.join(" ")).toContain("run_counts.success")
    expect(result.errors.join(" ")).toContain("terminal count sum")
  })

  it("rejects nonterminal final runs", async () => {
    const directory = await scratchDirectory()
    const payload = emptySuccessfulPayload()
    payload.runs[0]!.status = "running"
    payload.runs[0]!.finished_at = null

    const result = await finalize(payload, dependencies(directory))

    expect(result.status).toBe("failed")
    expect(result.errors.join(" ")).toContain("nonterminal status running")
  })

  it("uses UNAVAILABLE only for availability-only blockers", async () => {
    const directory = await scratchDirectory()
    const payload = emptySuccessfulPayload()
    failedRun(payload, "host_model_unavailable")
    payload.workflow_result = { kind: "council", value: "UNAVAILABLE" }
    payload.verdict = "UNAVAILABLE"

    const result = await finalize(payload, dependencies(directory))

    expect(result.status).toBe("success")
    expect(result.operation_verdict).toBe("UNAVAILABLE")
  })

  it("gives limit causes precedence over availability", async () => {
    const directory = await scratchDirectory()
    const payload = emptySuccessfulPayload()
    failedRun(payload, "provider_unavailable")
    const plan = structuredClone(payload.plan[0]!)
    plan.sequence = 2
    payload.plan.push(plan)
    const limited = structuredClone(payload.runs[0]!)
    limited.run_id = "22222222-2222-4222-8222-222222222222"
    limited.sequence = 2
    limited.status = "limit_skipped"
    limited.error = null
    payload.runs.push(limited)
    payload.run_counts.total = 2
    payload.run_counts.limit_skipped = 1
    payload.workflow_result = { kind: "council", value: "INCONCLUSIVE" }
    payload.verdict = "INCONCLUSIVE"

    const result = await finalize(payload, dependencies(directory))

    expect(result.status).toBe("success")
    expect(result.operation_verdict).toBe("INCONCLUSIVE")
  })

  it("rejects a successful model mismatch instead of rewriting provenance", async () => {
    const directory = await scratchDirectory()
    const payload = emptySuccessfulPayload()
    payload.runs[0]!.reported_model = "other/model"
    payload.runs[0]!.model_mismatch = true

    const result = await finalize(payload, dependencies(directory))

    expect(result.status).toBe("failed")
    expect(result.errors.join(" ")).toContain("model mismatch")
    expect(payload.runs[0]!.requested_model).toBe("opencode-go/deepseek-v4-pro")
  })

  it("validates successful host requested, resolved, and reported model provenance", async () => {
    const directory = await scratchDirectory()
    const payload = emptySuccessfulPayload()
    payload.plan[0]!.source = "host"
    payload.plan[0]!.model = null
    payload.plan[0]!.variant = null
    const run = payload.runs[0]!
    run.source = "host"
    run.requested_model = null
    run.provider = null
    run.model_id = null
    run.variant = null
    run.resolved_parent_model = "opencode-go/deepseek-v4-pro"
    run.resolved_parent_variant = "high"

    const result = await finalize(payload, dependencies(directory))

    expect(result.status).toBe("success")
    expect(result.review_dispatch?.payload.runs[0]).toMatchObject({
      requested_model: null,
      resolved_parent_model: "opencode-go/deepseek-v4-pro",
      resolved_parent_variant: "high",
      reported_model: "opencode-go/deepseek-v4-pro",
      model_mismatch: false,
    })
  })

  it("rejects duplicate consolidated findings and advisories", async () => {
    const directory = await scratchDirectory()
    const payload = payloadFixture()
    payload.findings.push(structuredClone(payload.findings[0]!))
    payload.advisories = [
      { severity: "LOW", description: "Same advisory", run_ids: [payload.runs[0]!.run_id] },
      { severity: "MEDIUM", description: " same   advisory ", run_ids: [payload.runs[0]!.run_id] },
    ]

    const result = await finalize(payload, dependencies(directory))

    expect(result.status).toBe("failed")
    expect(result.errors).toContain("duplicate consolidated finding for normalized file and root cause")
    expect(result.errors).toContain("duplicate advisory description")
  })
})

describe("finalize review dispatch persistence", () => {
  it("writes a restricted atomic HIC envelope with immutable provenance", async () => {
    const directory = await scratchDirectory()
    const payload = payloadFixture()

    const result = await finalize(payload, dependencies(directory))

    expect(result.status).toBe("success")
    expect(result.artifact_path).toBe(join(directory, ".uf/artifacts/dispatch", `${INITIAL_CORRELATION_ID}.json`))
    const artifact = JSON.parse(await readFile(result.artifact_path!, "utf8")) as Record<string, unknown>
    expect(artifact).toMatchObject({
      hero: "the-divisor",
      version: "1.0.0",
      artifact_type: "review-dispatch",
      schema_version: "1.0.0",
      context: {
        branch: "feature/review-fanout",
        commit: COMMIT,
        correlation_id: INITIAL_CORRELATION_ID,
        workflow_id: "review-council-pr-42",
      },
    })
    expect((await stat(join(directory, ".uf/artifacts/dispatch"))).mode & 0o777).toBe(0o750)
    expect((await stat(result.artifact_path!)).mode & 0o777).toBe(0o600)
    expect((await readdir(join(directory, ".uf/artifacts/dispatch"))).some((name) => name.endsWith(".tmp"))).toBe(false)
  })

  it("flushes, closes, and chmods before exclusive publication", async () => {
    const directory = await scratchDirectory()
    const base = dependencies(directory)
    const operations: string[] = []
    const ordered: FinalizationDependencies = {
      ...base,
      open: async (path, flags, mode) => {
        operations.push("open")
        const handle = await base.open(path, flags, mode)
        return {
          writeFile: async (data, encoding): Promise<void> => {
            operations.push("write")
            await handle.writeFile(data, encoding)
          },
          sync: async (): Promise<void> => {
            operations.push("sync")
            await handle.sync()
          },
          close: async (): Promise<void> => {
            operations.push("close")
            await handle.close()
          },
        }
      },
      chmod: async (path, mode): Promise<void> => {
        if (path.endsWith(".tmp")) {
          operations.push("chmod")
        }
        await base.chmod(path, mode)
      },
      link: async (existingPath, newPath): Promise<void> => {
        operations.push("link")
        await base.link(existingPath, newPath)
      },
      unlink: async (path): Promise<void> => {
        operations.push("unlink")
        await base.unlink(path)
      },
    }

    const result = await finalize(payloadFixture(), ordered)

    expect(result.status).toBe("success")
    expect(operations).toEqual(["open", "write", "sync", "close", "chmod", "link", "unlink"])
  })

  it("uses exclusive publication to preserve a real colliding target", async () => {
    const directory = await scratchDirectory()
    const artifactDirectory = join(directory, ".uf/artifacts/dispatch")
    await mkdir(artifactDirectory, { recursive: true })
    const existingPath = join(artifactDirectory, `${INITIAL_CORRELATION_ID}.json`)
    await writeFile(existingPath, "existing\n", "utf8")

    const payload = payloadFixture()
    const result = await finalize(
      payload,
      dependencies(directory, [TEMPORARY_ID, COLLISION_ID, SECOND_TEMPORARY_ID]),
    )

    expect(result.status).toBe("success")
    expect(result.correlation_id).toBe(COLLISION_ID)
    expect(result.review_dispatch?.context.correlation_id).toBe(COLLISION_ID)
    expect(result.review_dispatch?.payload.correlation_id).toBe(COLLISION_ID)
    expect(payload.correlation_id).toBe(INITIAL_CORRELATION_ID)
    expect(await readFile(existingPath, "utf8")).toBe("existing\n")
    const published = JSON.parse(await readFile(result.artifact_path!, "utf8")) as {
      readonly payload: { readonly correlation_id: string }
    }
    expect(published.payload.correlation_id).toBe(COLLISION_ID)
  })

  it("retries a collision detected at atomic publication", async () => {
    const directory = await scratchDirectory()
    const base = dependencies(directory, [TEMPORARY_ID, COLLISION_ID, SECOND_TEMPORARY_ID])
    let linkAttempts = 0
    const racing: FinalizationDependencies = {
      ...base,
      link: async (existingPath, newPath): Promise<void> => {
        linkAttempts += 1
        if (linkAttempts === 1) {
          throw Object.assign(new Error("late collision"), { code: "EEXIST" })
        }
        await base.link(existingPath, newPath)
      },
    }

    const result = await finalize(payloadFixture(), racing)

    expect(result.status).toBe("success")
    expect(result.correlation_id).toBe(COLLISION_ID)
    expect(linkAttempts).toBe(2)
    expect(await readdir(join(directory, ".uf/artifacts/dispatch"))).toEqual([`${COLLISION_ID}.json`])
  })

  it("fails closed on write failure and cleans temporary files", async () => {
    const directory = await scratchDirectory()
    const base = dependencies(directory)
    const failing: FinalizationDependencies = {
      ...base,
      open: async (path, flags, mode) => {
        const handle = await base.open(path, flags, mode)
        return {
          ...handle,
          writeFile: async (): Promise<void> => {
            throw new Error("write denied")
          },
        }
      },
    }

    const result = await finalize(emptySuccessfulPayload(), failing)

    expect(result.status).toBe("failed")
    expect(result.operation_verdict).toBe("INCONCLUSIVE")
    expect(result.calculated_assessment).toMatchObject({ authoritative: false, payload: { verdict: "APPROVE" } })
    expect(result.calculated_assessment?.payload.findings).toEqual([])
    expect(await readdir(join(directory, ".uf/artifacts/dispatch"))).toEqual([])
  })

  it("fails closed on link failure and removes the flushed temporary file", async () => {
    const directory = await scratchDirectory()
    const base = dependencies(directory)
    const failing: FinalizationDependencies = {
      ...base,
      link: async (): Promise<void> => {
        throw new Error("link denied")
      },
    }

    const result = await finalize(emptySuccessfulPayload(), failing)

    expect(result.status).toBe("failed")
    expect(result.errors.join(" ")).toContain("link denied")
    expect(result.calculated_assessment?.review_verdict.council_decision).toBe("APPROVED")
    expect(await readdir(join(directory, ".uf/artifacts/dispatch"))).toEqual([])
  })
})
