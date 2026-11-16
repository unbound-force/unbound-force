import { rm } from "node:fs/promises"

import { afterEach, describe, expect, it } from "vitest"

import { setupGhMockedPlugin, type MockSpawnFn, type MockSpawnResponses } from "./helpers/gh-spawn-mock.js"

// ── Fixtures ─────────────────────────────────────────────────

/**
 * Base green checks. `gh pr checks <N> --json name,state,bucket,link`
 * returns links in both verified forms — /actions/runs/<id>/job/<id> and
 * the bare /runs/<id> shape (change design Context, field-validity note).
 */
const PASSING_BUILD_CHECK = {
  name: "CI / build",
  state: "SUCCESS",
  bucket: "pass",
  link: "https://github.com/unbound-force/unbound-force/actions/runs/40908825770/job/112233445566",
}

const PASSING_TEST_CHECK = {
  name: "CI / test",
  state: "SUCCESS",
  bucket: "pass",
  link: "https://github.com/unbound-force/unbound-force/runs/40908825771",
}

/** Green PR (FR-013 passing scenario): every check buckets as pass. */
const MOCK_CHECKS_GREEN = [PASSING_BUILD_CHECK, PASSING_TEST_CHECK]

const FAILING_TEST_CHECK = {
  name: "CI / test",
  state: "FAILURE",
  bucket: "fail",
  link: "https://github.com/unbound-force/unbound-force/actions/runs/40908825771/job/112233445577",
}

/** Red PR (FR-013 failing scenario): gh exits non-zero with the JSON array still on stdout. */
const MOCK_CHECKS_FAILING = [PASSING_BUILD_CHECK, FAILING_TEST_CHECK]

const PENDING_LINT_CHECK = {
  name: "CI / lint",
  state: "PENDING",
  bucket: "pending",
  link: "https://github.com/unbound-force/unbound-force/actions/runs/40908825772/job/112233445588",
}

/** Pending PR (FR-013 pending scenario): none fail, one still running (gh documents exit 8). */
const MOCK_CHECKS_PENDING = [PASSING_BUILD_CHECK, PENDING_LINT_CHECK]

const SKIPPED_MACOS_CHECK = {
  name: "CI / macos",
  state: "SKIPPED",
  bucket: "skipping",
  link: "https://github.com/unbound-force/unbound-force/runs/40908825773",
}

/** Green with skips: pass and skipping buckets both classify passing (FR-013 green scenario). */
const MOCK_CHECKS_PASS_AND_SKIPPING = [PASSING_BUILD_CHECK, SKIPPED_MACOS_CHECK]

const CANCELLED_NIGHTLY_CHECK = {
  name: "CI / nightly",
  state: "CANCELLED",
  bucket: "cancel",
  link: "https://github.com/unbound-force/unbound-force/runs/40908825774",
}

/** Cancelled check: no failures, but a cancel bucket classifies pending. */
const MOCK_CHECKS_CANCELLED = [PASSING_BUILD_CHECK, CANCELLED_NIGHTLY_CHECK]

const UNKNOWN_BUCKET_CHECK = {
  name: "CI / exotic",
  state: "NEUTRAL",
  bucket: "blocked",
  link: "https://github.com/unbound-force/unbound-force/runs/40908825775",
}

/**
 * Unknown bucket (risk R5, vocabulary drift): a value outside the verified
 * pass/fail/pending/skipping/cancel set must classify fail-closed.
 */
const MOCK_CHECKS_UNKNOWN_BUCKET = [PASSING_BUILD_CHECK, UNKNOWN_BUCKET_CHECK]

/** Fail dominates pending when both appear (FR-013: failing when ANY check buckets as fail). */
const MOCK_CHECKS_FAIL_AND_PENDING = [FAILING_TEST_CHECK, PENDING_LINT_CHECK]

// ── Mock fixture responses (harness: ./helpers/gh-spawn-mock.ts) ──

/** Suite fixture responses for the route fetch_pr_checks owns; the which/auth prologue defaults live in the harness. */
const SPAWN_FIXTURES: MockSpawnResponses = {
  "pr checks": { stdout: JSON.stringify(MOCK_CHECKS_GREEN), stderr: "", exitCode: 0 },
}

// ── Test infrastructure ──────────────────────────────────────

const scratchDirectories: string[] = []

/**
 * Sets up the plugin hooks with Bun.spawn mocked through the shared
 * gh-spawn-mock harness, selecting the fetch_pr_checks tool.
 */
async function setupTool(spawnOverrides: MockSpawnResponses = {}) {
  const { hooks, mockSpawn, tmpDir, restore } = await setupGhMockedPlugin({
    spawnDefaults: SPAWN_FIXTURES,
    spawnOverrides,
    cleanupRegistry: scratchDirectories,
  })
  return { tool: hooks.tool.fetch_pr_checks, mockSpawn, tmpDir, restore }
}

async function callTool(
  overrides: MockSpawnResponses = {},
  prNumber = 42,
): Promise<{ parsed: Record<string, unknown>; mockSpawn: MockSpawnFn; tmpDir: string; restore: () => void }> {
  const { tool, mockSpawn, tmpDir, restore } = await setupTool(overrides)
  const result = await tool.execute({ pr_number: prNumber }, {} as never)
  const parsed = JSON.parse(result.output) as Record<string, unknown>
  return { parsed, mockSpawn, tmpDir, restore }
}

/**
 * Return the argv of the first recorded `gh <group> <verb>` spawn call
 * (e.g. group "pr", verb "checks"), or undefined when the tool never
 * issued that call.
 */
function findGhArgv(mockSpawn: MockSpawnFn, group: string, verb: string): string[] | undefined {
  const call = mockSpawn.mock.calls.find(
    (c: unknown[]) => (c[0] as string[])[1] === group && (c[0] as string[])[2] === verb,
  )
  return call === undefined ? undefined : (call[0] as string[])
}

afterEach(async () => {
  await Promise.all(scratchDirectories.splice(0).map((dir) => rm(dir, { recursive: true, force: true })))
})

// ── Tests ────────────────────────────────────────────────────

describe("fetch_pr_checks", () => {
  describe("successful output", () => {
    it("returns each check's name, state, bucket, and link", async () => {
      const { parsed, restore } = await callTool()
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.pr_number).toBe(42)
        expect(parsed.eligibility).toBe("passing")
        expect(parsed.warnings).toEqual([])
        // The full four-field mapping, exactly as gh returned it.
        expect(parsed.checks).toEqual(MOCK_CHECKS_GREEN)
        const first = (parsed.checks as Array<Record<string, unknown>>)[0]
        expect(first.name).toBe("CI / build")
        expect(first.state).toBe("SUCCESS")
        expect(first.bucket).toBe("pass")
        expect(first.link).toBe("https://github.com/unbound-force/unbound-force/actions/runs/40908825770/job/112233445566")
      } finally {
        restore()
      }
    })
  })

  describe("exit-code-tolerant matrix (design D4, risk R1)", () => {
    it("returns check data on exit 0 with green checks", async () => {
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: JSON.stringify(MOCK_CHECKS_GREEN), stderr: "", exitCode: 0 },
      })
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.checks).toEqual(MOCK_CHECKS_GREEN)
        expect(parsed.eligibility).toBe("passing")
      } finally {
        restore()
      }
    })

    it("returns check data on exit 8 with pending checks", async () => {
      // gh documents exit 8 for pending checks while still printing the
      // authoritative JSON array to stdout — a non-zero exit must not be
      // misread as a fetch failure.
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: JSON.stringify(MOCK_CHECKS_PENDING), stderr: "", exitCode: 8 },
      })
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.checks).toEqual(MOCK_CHECKS_PENDING)
        expect(parsed.eligibility).toBe("pending")
      } finally {
        restore()
      }
    })

    it("returns check data on a non-zero exit with failing checks", async () => {
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: JSON.stringify(MOCK_CHECKS_FAILING), stderr: "", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.checks).toEqual(MOCK_CHECKS_FAILING)
        expect(parsed.eligibility).toBe("failing")
      } finally {
        restore()
      }
    })
  })

  describe("eligibility derivation (FR-013)", () => {
    it("classifies failing when any check buckets as fail", async () => {
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: JSON.stringify(MOCK_CHECKS_FAILING), stderr: "", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.eligibility).toBe("failing")
      } finally {
        restore()
      }
    })

    it("classifies pending when none fail but one buckets as pending", async () => {
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: JSON.stringify(MOCK_CHECKS_PENDING), stderr: "", exitCode: 8 },
      })
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.eligibility).toBe("pending")
      } finally {
        restore()
      }
    })

    it("classifies pending when a check buckets as cancel", async () => {
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: JSON.stringify(MOCK_CHECKS_CANCELLED), stderr: "", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.eligibility).toBe("pending")
      } finally {
        restore()
      }
    })

    it("classifies passing when all checks bucket as pass or skipping", async () => {
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: JSON.stringify(MOCK_CHECKS_PASS_AND_SKIPPING), stderr: "", exitCode: 0 },
      })
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.eligibility).toBe("passing")
      } finally {
        restore()
      }
    })

    it("classifies failing when fail and pending buckets appear together (fail dominates)", async () => {
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: JSON.stringify(MOCK_CHECKS_FAIL_AND_PENDING), stderr: "", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.eligibility).toBe("failing")
      } finally {
        restore()
      }
    })

    it("classifies unknown buckets fail-closed as pending with a warning naming the value", async () => {
      // Risk R5: a bucket value outside the verified vocabulary degrades to
      // "wait" (pending), never to a silent wrong path, and the warning
      // names the unrecognized value.
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: JSON.stringify(MOCK_CHECKS_UNKNOWN_BUCKET), stderr: "", exitCode: 0 },
      })
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.eligibility).toBe("pending")
        const warnings = parsed.warnings as string[]
        expect(warnings.some((w: string) => w.includes("unknown check bucket") && w.includes('"blocked"'))).toBe(true)
      } finally {
        restore()
      }
    })
  })

  describe("zero checks (verified gh shape)", () => {
    it("reports passing eligibility with empty checks when gh reports no checks", async () => {
      // Verified shape (gh 2.91.0): a PR with zero checks exits 1 with EMPTY
      // stdout and the "no checks reported" stderr marker — gh does not emit
      // `[]` — so the marker, not the exit code, gates the zero-checks data
      // classification.
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: "", stderr: "no checks reported", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.pr_number).toBe(42)
        expect(parsed.checks).toEqual([])
        expect(parsed.eligibility).toBe("passing")
        expect(parsed.warnings).toEqual([])
      } finally {
        restore()
      }
    })

    it("treats empty stdout with unrecognized stderr as a fetch failure, not zero checks (fail-closed)", async () => {
      // Without the marker, empty stdout must NOT silently report "passing".
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: "", stderr: "server error 500", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect(parsed.checks).toBeUndefined()
        expect(parsed.eligibility).toBeUndefined()
      } finally {
        restore()
      }
    })
  })

  describe("unparseable stdout", () => {
    it("returns a structured error for unparseable stdout on exit 0", async () => {
      // gh exit codes are not a reliable indicator of stdout parseability —
      // exit 0 with unparseable stdout is still a fetch failure (FR-013).
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: "not json at all", stderr: "", exitCode: 0 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect(parsed.retryable).toBe(false)
        expect((parsed.message as string)).toContain("not a JSON array")
        expect((parsed.message as string)).toContain("exit 0")
      } finally {
        restore()
      }
    })

    it("returns a structured error for unparseable stdout on a non-zero exit", async () => {
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: "gh: something went wrong", stderr: "", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect((parsed.message as string)).toContain("not a JSON array")
        expect((parsed.message as string)).toContain("exit 1")
      } finally {
        restore()
      }
    })

    it("returns a structured error when stdout parses as JSON but is not an array", async () => {
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: JSON.stringify({ name: "CI / build" }), stderr: "", exitCode: 0 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect(parsed.checks).toBeUndefined()
      } finally {
        restore()
      }
    })

    it("bounds the stdout excerpt in the error message", async () => {
      // Untrusted stdout must not flood the structured error message — the
      // excerpt is bounded (~200 chars), so the message stays small even
      // for megabyte garbage.
      const longStdout = "x".repeat(500)
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: longStdout, stderr: "", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect((parsed.message as string).length).toBeLessThan(300)
      } finally {
        restore()
      }
    })
  })

  describe("error handling", () => {
    it("returns gh_not_found when the gh CLI is not available", async () => {
      const { parsed, restore } = await callTool({
        which: { stdout: "", stderr: "which: no gh in PATH", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("gh_not_found")
        expect(parsed.retryable).toBe(false)
        expect(typeof parsed.message).toBe("string")
      } finally {
        restore()
      }
    })

    it("returns gh_auth_failed when gh auth status fails", async () => {
      const { parsed, restore } = await callTool({
        "auth status": { stdout: "", stderr: "not logged in", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("gh_auth_failed")
        expect(parsed.retryable).toBe(false)
      } finally {
        restore()
      }
    })

    it("returns pr_not_found when the PR number does not resolve", async () => {
      // gh pr checks reports unresolved PR numbers through GraphQL phrasing
      // the sibling pr-view pattern does not match ("Could not resolve to a
      // PullRequest with the number of N").
      const { parsed, restore } = await callTool({
        "pr checks": {
          stdout: "",
          stderr: "GraphQL: Could not resolve to a PullRequest with the number of 42",
          exitCode: 1,
        },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("pr_not_found")
        expect(parsed.retryable).toBe(false)
        expect((parsed.message as string)).toContain("not found")
      } finally {
        restore()
      }
    })

    it("returns api_error when stdout carries no check-status data", async () => {
      // Whitespace-only stdout trims to empty with no marker on stderr — a
      // fetch failure on any exit code, including this exit 0.
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: "   \n  ", stderr: "", exitCode: 0 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect(parsed.retryable).toBe(false)
        expect(typeof parsed.message).toBe("string")
      } finally {
        restore()
      }
    })

    it("returns timeout when gh reports a timeout", async () => {
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: "", stderr: "request timed out", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("timeout")
        expect(parsed.retryable).toBe(true)
      } finally {
        restore()
      }
    })

    it("returns rate_limited when GitHub rate limits", async () => {
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: "", stderr: "API rate limit exceeded", exitCode: 4 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("rate_limited")
        expect(parsed.retryable).toBe(true)
      } finally {
        restore()
      }
    })

    it("classifies exit 4 with no matching stderr as gh_auth_failed, not rate_limited (FRR-FR-001)", async () => {
      // gh documents exit code 4 as "if a command requires
      // authentication" (gh help exit-codes), not as a rate-limit
      // code. The stderr below deliberately matches no text key and
      // no fetch-pr-checks pattern, so the exit code is the only
      // signal — before the shared classifier this input reported
      // rate_limited/retryable (issue #1023 MEDIUM 3).
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: "", stderr: "HTTP 422: Validation Failed", exitCode: 4 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("gh_auth_failed")
        expect(parsed.retryable).toBe(false)
        expect(parsed.message as string).toContain("gh CLI is not authenticated")
      } finally {
        restore()
      }
    })
  })

  describe("input validation", () => {
    it("rejects non-integer pr_number via schema validation", async () => {
      const { tool, restore } = await setupTool()
      try {
        const result = await tool.execute({ pr_number: 3.14 } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect((parsed.message as string).toLowerCase()).toContain("invalid")
      } finally {
        restore()
      }
    })

    it("rejects zero pr_number via schema validation", async () => {
      const { tool, restore } = await setupTool()
      try {
        const result = await tool.execute({ pr_number: 0 } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
      } finally {
        restore()
      }
    })

    it("rejects negative pr_number via schema validation", async () => {
      const { tool, restore } = await setupTool()
      try {
        const result = await tool.execute({ pr_number: -1 } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
      } finally {
        restore()
      }
    })

    it("rejects pr_number above 999999 via schema validation", async () => {
      const { tool, restore } = await setupTool()
      try {
        const result = await tool.execute({ pr_number: 1_000_000 } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
      } finally {
        restore()
      }
    })

    it("rejects string pr_number (command-injection prevention)", async () => {
      const { tool, mockSpawn, restore } = await setupTool()
      try {
        const result = await tool.execute({ pr_number: "42; rm -rf /" } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        // The schema rejects the value before any gh invocation.
        expect(mockSpawn.mock.calls).toHaveLength(0)
      } finally {
        restore()
      }
    })

    it("rejects extra fields via .strict() schema", async () => {
      const { tool, restore } = await setupTool()
      try {
        const result = await tool.execute({ pr_number: 42, evil: "payload" } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
      } finally {
        restore()
      }
    })

    it("requires pr_number (omitted is a validation error, unlike fetch_pr_review_context)", async () => {
      const { tool, restore } = await setupTool()
      try {
        const result = await tool.execute({} as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect((parsed.message as string)).toContain("pr_number")
      } finally {
        restore()
      }
    })

    it("accepts valid edge-case pr_number=1", async () => {
      const { parsed, restore } = await callTool({}, 1)
      try {
        expect(parsed.error).toBeUndefined()
        // fetch_pr_checks echoes the validated input number.
        expect(parsed.pr_number).toBe(1)
        expect(parsed.eligibility).toBe("passing")
      } finally {
        restore()
      }
    })

    it("accepts valid edge-case pr_number=999999", async () => {
      const { parsed, restore } = await callTool({}, 999_999)
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.pr_number).toBe(999_999)
      } finally {
        restore()
      }
    })
  })

  describe("Bun.spawn invocation", () => {
    it("calls which gh as the first spawn", async () => {
      const { mockSpawn, restore } = await callTool()
      try {
        const firstCall = mockSpawn.mock.calls[0]
        expect(firstCall[0]).toEqual(["which", "gh"])
      } finally {
        restore()
      }
    })

    it("calls gh auth status before gh pr checks", async () => {
      const { mockSpawn, restore } = await callTool()
      try {
        const calls = mockSpawn.mock.calls.map((c: unknown[]) => (c[0] as string[]).slice(0, 3).join(" "))
        const authIndex = calls.indexOf("gh auth status")
        const checksIndex = calls.findIndex((c: string) => c.startsWith("gh pr checks"))
        expect(authIndex).toBeGreaterThan(-1)
        expect(checksIndex).toBeGreaterThan(authIndex)
      } finally {
        restore()
      }
    })

    it("invokes gh pr checks with the pinned argument vector", async () => {
      const { mockSpawn, restore } = await callTool()
      try {
        const checksArgs = findGhArgv(mockSpawn, "pr", "checks")
        expect(checksArgs).toBeDefined()
        expect(checksArgs).toEqual(["gh", "pr", "checks", "42", "--json", "name,state,bucket,link"])
      } finally {
        restore()
      }
    })

    it("passes pr_number as a discrete array element (no shell interpolation)", async () => {
      const { mockSpawn, restore } = await callTool({}, 7)
      try {
        const checksArgs = findGhArgv(mockSpawn, "pr", "checks")
        expect(checksArgs).toBeDefined()
        expect(checksArgs![3]).toBe("7") // discrete string, not embedded in another arg
      } finally {
        restore()
      }
    })
  })
})
