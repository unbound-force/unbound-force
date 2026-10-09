import type { PluginInput, ToolContext } from "@opencode-ai/plugin"
import { randomUUID } from "node:crypto"
import { mkdir, writeFile, rm } from "node:fs/promises"
import { join } from "node:path"
import { tmpdir } from "node:os"
import { describe, expect, it, afterEach } from "vitest"

import { ReviewDispatchPlugin, createConsolidateDispatchTool, createFinalizationDependencies, _getSessionCorrelationMap } from "../plugins/review-dispatch/index.js"
import { withBun, scratchProject } from "./helpers.js"

/** Build a valid PersistedRunData fixture for writing to disk. */
function runFixture(overrides: {
  agent: string
  run_id?: string
  status?: string
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
    run_id: overrides.run_id ?? randomUUID(),
    agent: overrides.agent,
    source: "explicit",
    sequence: 1,
    status: overrides.status ?? "success",
    started_at: "2026-10-08T10:00:00.000Z",
    finished_at: "2026-10-08T10:01:00.000Z",
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

function noopContext(): ToolContext {
  return {
    sessionID: "consolidate-session",
    messageID: "msg-consolidate",
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

function validInputContext() {
  return {
    kind: "pr" as const,
    pr_number: 686,
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

function validPlan() {
  return [
    {
      agent: "divisor-guard",
      decision: "include",
      reason_code: "advisor",
      reason: "Advisor selected",
      source: "advisor",
      sequence: 1,
      read_only: true,
      tier: "standard",
      model: "provider/standard",
      variant: null,
      validation_errors: [],
    },
  ]
}

function validCoverage() {
  return { preflight_verdict: "PASS" as const, checks_total: 5, checks_passed: 5 }
}

/** Default session metadata written to session-metadata.json for consolidation. */
function defaultSessionMetadata(overrides?: Record<string, unknown>) {
  return {
    command: "review-council" as const,
    mode: "code" as const,
    full: false,
    input_context: validInputContext(),
    change_profile: validChangeProfile(),
    plan: validPlan(),
    coverage: validCoverage(),
    ...overrides,
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

/** Write run fixture files + session metadata and return the correlation_id. */
async function setupDispatchSession(
  runs: ReturnType<typeof runFixture>[],
  metadata?: Record<string, unknown>,
): Promise<string> {
  const correlationId = randomUUID()
  const dir = join(tmpdir(), "opencode", `dispatch-${correlationId}`)
  await mkdir(dir, { recursive: true })
  cleanupDirs.push(dir)
  _getSessionCorrelationMap().set("consolidate-session", correlationId)

  await writeFile(join(dir, "session-metadata.json"), JSON.stringify(defaultSessionMetadata(metadata), null, 2), "utf8")

  for (const run of runs) {
    await writeFile(join(dir, `run-${run.agent}.json`), JSON.stringify(run, null, 2), "utf8")
  }

  return correlationId
}

describe("consolidate_dispatch", () => {
  it("consolidates successful runs with no findings into APPROVE", async () => {
    const correlationId = await setupDispatchSession([
      runFixture({ agent: "divisor-guard" }),
      runFixture({ agent: "divisor-adversary" }),
    ])

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(tmpdir()))
    const result = await tool.execute({}, noopContext())

    const payload = JSON.parse(result.output)
    expect(payload.verdict).toBe("APPROVE")
    expect(payload.verdict_reason).toBe("No findings requiring changes")
    expect(payload.findings).toHaveLength(0)
    expect(payload.advisories).toHaveLength(0)
    expect(payload.run_counts.total).toBe(2)
    expect(payload.run_counts.success).toBe(2)
    expect(payload.workflow_result).toEqual({ kind: "council", value: "APPROVE" })
    expect(payload.correlation_id).toBe(correlationId)
    expect(payload.proposals).toBeUndefined()
  })

  it("includes lesson proposals from child runs in output", async () => {
    await setupDispatchSession([
      runFixture({
        agent: "divisor-architect",
        proposals: [
          { information: "When extracting a shared lib, keep re-exports for backward compat.", tag: "shared-lib", category: "pattern" },
        ],
      }),
      runFixture({ agent: "divisor-guard" }),
    ])

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(tmpdir()))
    const result = await tool.execute({}, noopContext())

    const payload = JSON.parse(result.output)
    expect(payload.proposals).toHaveLength(1)
    expect(payload.proposals[0]).toEqual({
      agent: "divisor-architect",
      information: "When extracting a shared lib, keep re-exports for backward compat.",
      tag: "shared-lib",
      category: "pattern",
    })
  })

  it("escalates to REQUEST CHANGES when HIGH findings exist", async () => {
    const runId1 = randomUUID()
    const correlationId = await setupDispatchSession([
      runFixture({
        agent: "divisor-guard",
        run_id: runId1,
        findings: [
          { severity: "HIGH", category: "security", description: "Missing auth", root_cause: "No middleware", file: "api.ts", line: 10 },
        ],
      }),
      runFixture({ agent: "divisor-adversary" }),
    ])

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(tmpdir()))
    const result = await tool.execute({}, noopContext())

    const payload = JSON.parse(result.output)
    expect(payload.verdict).toBe("REQUEST CHANGES")
    expect(payload.verdict_reason).toContain("HIGH")
    expect(payload.findings).toHaveLength(1)
    expect(payload.findings[0].run_ids).toContain(runId1)
    expect(payload.advisories).toHaveLength(0)
  })

  it("maps MEDIUM findings to APPROVE WITH ADVISORIES", async () => {
    const correlationId = await setupDispatchSession([
      runFixture({
        agent: "divisor-guard",
        findings: [
          { severity: "MEDIUM", category: "style", description: "Missing docs", root_cause: "No jsdoc", file: "lib.ts", line: 5 },
        ],
      }),
    ])

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(tmpdir()))
    const result = await tool.execute({}, noopContext())

    const payload = JSON.parse(result.output)
    expect(payload.verdict).toBe("APPROVE WITH ADVISORIES")
    expect(payload.advisories).toHaveLength(1)
    expect(payload.advisories[0].severity).toBe("MEDIUM")
  })

  it("maps LOW findings to APPROVE WITH ADVISORIES", async () => {
    const correlationId = await setupDispatchSession([
      runFixture({
        agent: "divisor-testing",
        findings: [
          { severity: "LOW", category: "style", description: "Naming convention", root_cause: "camelCase", file: "util.ts", line: 1 },
        ],
      }),
    ])

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(tmpdir()))
    const result = await tool.execute({}, noopContext())

    const payload = JSON.parse(result.output)
    expect(payload.verdict).toBe("APPROVE WITH ADVISORIES")
    expect(payload.advisories).toHaveLength(1)
    expect(payload.advisories[0].severity).toBe("LOW")
  })

  it("deduplicates findings by normalized file + root_cause and keeps highest severity", async () => {
    const runId1 = randomUUID()
    const runId2 = randomUUID()
    const correlationId = await setupDispatchSession([
      runFixture({
        agent: "divisor-guard",
        run_id: runId1,
        findings: [
          { severity: "MEDIUM", category: "security", description: "Possible issue", root_cause: "Missing validation", file: "src/handler.ts", line: 10 },
        ],
      }),
      runFixture({
        agent: "divisor-adversary",
        run_id: runId2,
        findings: [
          // Same file (normalized) + same root_cause but higher severity.
          { severity: "HIGH", category: "security", description: "Definite issue", root_cause: "Missing validation", file: "SRC/HANDLER.TS", line: 10 },
        ],
      }),
    ])

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(tmpdir()))
    const result = await tool.execute({}, noopContext())

    const payload = JSON.parse(result.output)
    // Deduplicated to one finding with highest severity.
    expect(payload.findings).toHaveLength(1)
    expect(payload.findings[0].severity).toBe("HIGH")
    expect(payload.findings[0].run_ids).toContain(runId1)
    expect(payload.findings[0].run_ids).toContain(runId2)
    expect(payload.verdict).toBe("REQUEST CHANGES")
  })

  it("counts run statuses correctly", async () => {
    const correlationId = await setupDispatchSession([
      runFixture({ agent: "divisor-guard", status: "success" }),
      runFixture({ agent: "divisor-adversary", status: "failed" }),
      runFixture({ agent: "divisor-testing", status: "cancelled" }),
    ])

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(tmpdir()))
    const result = await tool.execute({}, noopContext())

    const payload = JSON.parse(result.output)
    expect(payload.run_counts.total).toBe(3)
    expect(payload.run_counts.success).toBe(1)
    expect(payload.run_counts.failed).toBe(1)
    expect(payload.run_counts.cancelled).toBe(1)
    expect(payload.run_counts.skipped).toBe(0)
  })

  it("maps triage-issue command to triage workflow result", async () => {
    const correlationId = await setupDispatchSession(
      [runFixture({ agent: "divisor-guard" })],
      {
        command: "triage-issue",
        mode: "triage",
        input_context: {
          kind: "issue" as const,
          issue_number: 42,
          issue_url: "https://github.com/unbound-force/unbound-force/issues/42",
          content_sha256: "a".repeat(64),
        },
        change_profile: {
          kind: "issue" as const,
          text_bytes: 500,
          comment_count: 3,
          content_sha256: "a".repeat(64),
          matched_rules: [],
          security_sensitive: false,
          user_facing: true,
          categories: ["standard" as const],
          tier: "standard" as const,
        },
      },
    )

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(tmpdir()))
    const result = await tool.execute({}, noopContext())

    const payload = JSON.parse(result.output)
    expect(payload.workflow_result.kind).toBe("triage")
    expect(payload.workflow_result.value).toBe("VALID")
  })

  it("maps address-feedback APPROVE WITH ADVISORIES to AUTHOR-DECIDES", async () => {
    const correlationId = await setupDispatchSession(
      [
        runFixture({
          agent: "divisor-guard",
          findings: [
            { severity: "MEDIUM", category: "style", description: "Needs improvement", root_cause: "Convention gap", file: "a.ts", line: 1 },
          ],
        }),
      ],
      { command: "address-feedback", mode: "feedback" },
    )

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(tmpdir()))
    const result = await tool.execute({}, noopContext())

    const payload = JSON.parse(result.output)
    expect(payload.workflow_result.kind).toBe("feedback")
    expect(payload.workflow_result.value).toBe("AUTHOR-DECIDES")
  })

  it("maps address-feedback APPROVE (no findings) to ACCEPT", async () => {
    const correlationId = await setupDispatchSession(
      [runFixture({ agent: "divisor-guard" })],
      { command: "address-feedback", mode: "feedback" },
    )

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(tmpdir()))
    const result = await tool.execute({}, noopContext())

    const payload = JSON.parse(result.output)
    expect(payload.workflow_result.kind).toBe("feedback")
    expect(payload.workflow_result.value).toBe("ACCEPT")
  })

  it("maps address-feedback REQUEST CHANGES to REQUEST CHANGES", async () => {
    const correlationId = await setupDispatchSession(
      [
        runFixture({
          agent: "divisor-guard",
          findings: [
            { severity: "HIGH", category: "security", description: "Blocker", root_cause: "Missing auth", file: "a.ts", line: 1 },
          ],
        }),
      ],
      { command: "address-feedback", mode: "feedback" },
    )

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(tmpdir()))
    const result = await tool.execute({}, noopContext())

    const payload = JSON.parse(result.output)
    expect(payload.workflow_result.kind).toBe("feedback")
    // REQUEST CHANGES falls to default in feedback → returns verdict as-is.
    expect(payload.workflow_result.value).toBe("REQUEST CHANGES")
  })

  it("returns error when no dispatch session directory exists", async () => {
    const tool = createConsolidateDispatchTool(createFinalizationDependencies(tmpdir()))
    const result = await tool.execute({}, noopContext())

    expect(result.output).toContain("no dispatch session found")
  })

  it("returns error when dispatch directory exists but has no run files", async () => {
    const correlationId = randomUUID()
    const dir = join(tmpdir(), "opencode", `dispatch-${correlationId}`)
    await mkdir(dir, { recursive: true })
    cleanupDirs.push(dir)
    _getSessionCorrelationMap().set("consolidate-session", correlationId)
    await writeFile(join(dir, "session-metadata.json"), JSON.stringify(defaultSessionMetadata()), "utf8")

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(tmpdir()))
    const result = await tool.execute({}, noopContext())

    expect(result.output).toContain("no run files found")
  })

  it("returns error when a run file contains invalid JSON", async () => {
    const correlationId = randomUUID()
    const dir = join(tmpdir(), "opencode", `dispatch-${correlationId}`)
    await mkdir(dir, { recursive: true })
    cleanupDirs.push(dir)
    _getSessionCorrelationMap().set("consolidate-session", correlationId)
    await writeFile(join(dir, "session-metadata.json"), JSON.stringify(defaultSessionMetadata()), "utf8")
    await writeFile(join(dir, "run-divisor-guard.json"), "not valid json", "utf8")

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(tmpdir()))
    const result = await tool.execute({}, noopContext())

    expect(result.output).toContain("failed to read run-divisor-guard.json")
  })

  it("preserves plan_version: 1 in the output payload", async () => {
    const correlationId = await setupDispatchSession([
      runFixture({ agent: "divisor-guard" }),
    ])

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(tmpdir()))
    const result = await tool.execute({}, noopContext())

    const payload = JSON.parse(result.output)
    expect(payload.plan_version).toBe(1)
    expect(payload.plan).toEqual(validPlan())
  })

  it("finalizes internally when provenance is provided", async () => {
    const projectRoot = join(tmpdir(), `consolidate-provenance-test-${randomUUID()}`)
    await mkdir(projectRoot, { recursive: true })
    cleanupDirs.push(projectRoot)

    // The run's run_id is auto-generated by runFixture; capture it
    // so we can verify it appears in the finalization result.
    const guardRun = runFixture({
      agent: "divisor-guard",
      proposals: [{ information: "Test lesson from provenance path", tag: "test-tag" }],
    })
    // Provenance commit MUST match input_context.head_sha for semantic validation.
    const headSha = "b".repeat(40)
    const correlationId = await setupDispatchSession([guardRun], {
      plan: [
        {
          agent: "divisor-guard",
          decision: "include",
          reason_code: "explicit",
          reason: "Explicitly included",
          source: "explicit",
          sequence: 1,
          read_only: true,
          tier: "standard",
          model: "provider/standard",
          variant: null,
          validation_errors: [],
        },
      ],
    })

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(projectRoot))
    const result = await tool.execute(
      {
        provenance: {
          branch: "opsx/test-branch",
          commit: headSha, // Must match input_context.head_sha
          workflow_id: "uf.review-council",
        },
      },
      noopContext(),
    )

    const finalizationResult = JSON.parse(result.output)
    // When provenance is provided, the result is a finalization result, not a raw payload.
    expect(finalizationResult.status).toBe("success")
    expect(finalizationResult.operation_verdict).toBe("APPROVE")
    expect(finalizationResult.correlation_id).toBe(correlationId)
    expect(finalizationResult.artifact_path).toContain(".uf/artifacts/dispatch/")
    expect(finalizationResult.errors).toHaveLength(0)
    expect(finalizationResult.proposals).toHaveLength(1)
    expect(finalizationResult.proposals[0].agent).toBe("divisor-guard")
    expect(finalizationResult.proposals[0].information).toBe("Test lesson from provenance path")
  })

  it("returns finalization payload JSON when provenance is omitted", async () => {
    const correlationId = await setupDispatchSession([
      runFixture({ agent: "divisor-guard" }),
    ])

    const tool = createConsolidateDispatchTool(createFinalizationDependencies(tmpdir()))
    // No provenance — should return raw payload.
    const result = await tool.execute({}, noopContext())

    const payload = JSON.parse(result.output)
    // When provenance is omitted, we get the raw payload (not a finalization result).
    expect(payload.verdict).toBe("APPROVE")
    expect(payload.correlation_id).toBe(correlationId)
    // Raw payload has no "status" or "artifact_path" fields.
    expect(payload).not.toHaveProperty("status")
    expect(payload).not.toHaveProperty("artifact_path")
  })
})
