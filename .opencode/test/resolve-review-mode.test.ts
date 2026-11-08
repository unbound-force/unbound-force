import { rm } from "node:fs/promises"

import { afterEach, describe, expect, it } from "vitest"

import {
  COMPARE_FILE_LIMIT,
  LIGHT_CHECK_MAX_CHANGED_LINES,
  decideReviewMode,
  type ReviewModeInput,
} from "../plugins/review-dispatch/index.js"
import { setupGhMockedPlugin, type MockSpawnResponses } from "./helpers/gh-spawn-mock.js"

// ── decideReviewMode (pure) ──────────────────────────────────

const HEAD = "b".repeat(40)
const PRIOR = "a".repeat(40)

function input(overrides: Partial<ReviewModeInput> = {}): ReviewModeInput {
  return {
    headSha: HEAD,
    viewer: "uf-divisor",
    reviews: [
      { id: 1, user: "uf-divisor", state: "DISMISSED", commit_id: PRIOR, submitted_at: "2026-11-08T10:00:00Z", body: "Council verdict: REQUEST CHANGES" },
    ],
    comments: [
      { id: 11, review_id: 1, path: "a.go", line: 3, body: "**HIGH — Missing check**: ..." },
      { id: 12, review_id: 1, path: "b.go", line: 9, body: "**LOW — Naming**: ..." },
      { id: 13, review_id: 1, path: "c.go", line: null, body: "No severity marker here" },
      { id: 14, review_id: 99, path: "d.go", line: 1, body: "**HIGH — another review**" },
    ],
    compare: {
      status: "ahead",
      files: [
        { filename: "a.go", status: "modified", changes: 30 },
        { filename: "a_test.go", status: "modified", changes: 18 },
      ],
    },
    checkBuckets: ["pass", "skipping"],
    ...overrides,
  }
}

describe("decideReviewMode", () => {
  it("picks a light check for a small contained push after a dismissed review", () => {
    const decision = decideReviewMode(input())
    expect(decision.mode).toBe("light")
    expect(decision.reason).toBe("2 files, 48 lines since aaaaaaa")
    expect(decision.prior_review).toEqual({ id: 1, commit_id: PRIOR, state: "DISMISSED" })
    expect(decision.delta).toEqual({ files: 2, changed_lines: 48 })
  })

  it("lists the prior review's HIGH and unlabelled findings, not LOW or other reviews'", () => {
    const decision = decideReviewMode(input())
    expect(decision.prior_findings).toEqual([
      { comment_id: 11, path: "a.go", line: 3, severity: "HIGH" },
      { comment_id: 13, path: "c.go", line: null, severity: "UNKNOWN" },
    ])
  })

  it("is full on a first review", () => {
    const decision = decideReviewMode(input({ reviews: [] }))
    expect(decision.mode).toBe("full")
    expect(decision.prior_review).toBeNull()
  })

  it("ignores reviews by other users and reviews of the head commit", () => {
    const decision = decideReviewMode(
      input({
        reviews: [
          { id: 2, user: "someone", state: "APPROVED", commit_id: PRIOR, submitted_at: "2026-11-08T11:00:00Z", body: "" },
          { id: 3, user: "UF-Divisor", state: "COMMENTED", commit_id: HEAD, submitted_at: "2026-11-08T12:00:00Z", body: "" },
        ],
      }),
    )
    expect(decision.mode).toBe("full")
    expect(decision.prior_review).toBeNull()
  })

  it("uses the latest earlier review by the viewer", () => {
    const later = { id: 5, user: "uf-divisor", state: "CHANGES_REQUESTED", commit_id: "c".repeat(40), submitted_at: "2026-11-08T12:00:00Z", body: "" }
    const decision = decideReviewMode(input({ reviews: [later, ...input().reviews] }))
    expect(decision.prior_review?.id).toBe(5)
  })

  it("is full when the prior review raised a CRITICAL finding", () => {
    expect(decideReviewMode(input({ reviews: [{ ...input().reviews[0]!, body: "- **CRITICAL — x**" }] })).mode).toBe("full")
    const comments = [{ id: 20, review_id: 1, path: "a.go", line: 1, body: "**CRITICAL — Determinism gate**" }]
    expect(decideReviewMode(input({ comments })).reason).toBe("the prior review raised a CRITICAL finding")
  })

  it("is full after a rebase", () => {
    const decision = decideReviewMode(input({ compare: { status: "diverged", files: [] } }))
    expect(decision.mode).toBe("full")
    expect(decision.reason).toContain("no longer an ancestor")
  })

  it("is full when the compare could not be read", () => {
    expect(decideReviewMode(input({ compare: null })).mode).toBe("full")
  })

  it("allows exactly the line limit and no more", () => {
    const at = { status: "ahead", files: [{ filename: "a.go", status: "modified", changes: LIGHT_CHECK_MAX_CHANGED_LINES }] }
    expect(decideReviewMode(input({ compare: at })).mode).toBe("light")
    const over = { status: "ahead", files: [{ filename: "a.go", status: "modified", changes: LIGHT_CHECK_MAX_CHANGED_LINES + 1 }] }
    expect(decideReviewMode(input({ compare: over })).mode).toBe("full")
  })

  it("is full when the compare response may be truncated", () => {
    const files = Array.from({ length: COMPARE_FILE_LIMIT }, (_, i) => ({ filename: `f${i}.go`, status: "modified", changes: 0 }))
    expect(decideReviewMode(input({ compare: { status: "ahead", files } })).reason).toBe("the compare response may be truncated")
  })

  it.each(["added", "removed", "renamed", "copied"])("is full when a file was %s", (status) => {
    const compare = { status: "ahead", files: [{ filename: "new.go", status, changes: 4 }] }
    expect(decideReviewMode(input({ compare })).mode).toBe("full")
  })

  it.each([
    ".github/workflows/ci.yml",
    "AGENTS.md",
    ".specify/memory/constitution.md",
    ".opencode/uf/packs/go.md",
    "internal/auth/session.go",
    "pkg/TokenStore.go",
    "config/secrets.yaml",
    "lib/credentials.ts",
  ])("is full when %s changed", (filename) => {
    const compare = { status: "ahead", files: [{ filename, status: "modified", changes: 2 }] }
    expect(decideReviewMode(input({ compare })).reason).toBe(`${filename} is a sensitive path`)
  })

  it("is full when a check is failing or the checks could not be read", () => {
    expect(decideReviewMode(input({ checkBuckets: ["pass", "fail"] })).mode).toBe("full")
    expect(decideReviewMode(input({ checkBuckets: null })).mode).toBe("full")
  })

  it("is light with pending checks and with no checks", () => {
    expect(decideReviewMode(input({ checkBuckets: ["pending"] })).mode).toBe("light")
    expect(decideReviewMode(input({ checkBuckets: [] })).mode).toBe("light")
  })

  it("is full when the viewer is unknown", () => {
    expect(decideReviewMode(input({ viewer: "" })).mode).toBe("full")
  })
})

// ── resolve_review_mode tool (gh mocked) ─────────────────────

const scratchDirectories: string[] = []

afterEach(async () => {
  for (const directory of scratchDirectories.splice(0)) {
    await rm(directory, { recursive: true, force: true })
  }
})

const ok = (value: unknown) => ({ stdout: typeof value === "string" ? value : JSON.stringify(value), stderr: "", exitCode: 0 })

const FIXTURES: MockSpawnResponses = {
  "pr view": ok({ headRefOid: HEAD, url: "https://github.com/acme/repo/pull/7" }),
  "api user": ok("uf-divisor\n"),
  "api reviews": ok([[{ id: 1, user: { login: "uf-divisor" }, state: "DISMISSED", commit_id: PRIOR, submitted_at: "2026-11-08T10:00:00Z", body: "" }]]),
  "api comments": ok([[{ id: 11, pull_request_review_id: 1, path: "a.go", line: 3, body: "**HIGH — x**" }]]),
  "api compare": ok({ status: "ahead", files: [{ filename: "a.go", status: "modified", changes: 12 }] }),
  "pr checks": ok([{ bucket: "pass" }]),
}

async function run(overrides: MockSpawnResponses = {}): Promise<Record<string, unknown>> {
  const { hooks, mockSpawn, restore } = await setupGhMockedPlugin({
    spawnDefaults: FIXTURES,
    spawnOverrides: overrides,
    cleanupRegistry: scratchDirectories,
  })
  try {
    const result = await hooks.tool.resolve_review_mode.execute({ pr_number: 7 }, {} as never)
    const parsed = JSON.parse(result.output) as Record<string, unknown>
    parsed.__calls = mockSpawn.mock.calls.map((call: unknown[]) => (call[0] as string[]).join(" "))
    return parsed
  } finally {
    restore()
  }
}

describe("resolve_review_mode tool", () => {
  it("returns a light decision with the prior findings", async () => {
    const result = await run()
    expect(result.mode).toBe("light")
    expect(result.head_sha).toBe(HEAD)
    expect(result.viewer).toBe("uf-divisor")
    expect(result.prior_findings).toEqual([{ comment_id: 11, path: "a.go", line: 3, severity: "HIGH" }])
    expect(result.__calls).toContain(`gh api repos/acme/repo/compare/${PRIOR}...${HEAD}`)
  })

  it("decides full without a compare call on a first review", async () => {
    const result = await run({ "api reviews": ok([[]]) })
    expect(result.mode).toBe("full")
    expect((result.__calls as string[]).some((call) => call.includes("/compare/"))).toBe(false)
  })

  it("decides full when the reviews cannot be read", async () => {
    const result = await run({ "api reviews": { stdout: "", stderr: "HTTP 403", exitCode: 1 } })
    expect(result.mode).toBe("full")
    expect(result.reason).toBe("the existing reviews could not be read")
  })

  it("decides full when the compare call fails", async () => {
    const result = await run({ "api compare": { stdout: "", stderr: "HTTP 404: Not Found", exitCode: 1 } })
    expect(result.mode).toBe("full")
  })

  it("decides full when a check fails, reading stdout on a non-zero exit", async () => {
    const result = await run({ "pr checks": { stdout: JSON.stringify([{ bucket: "fail" }]), stderr: "", exitCode: 1 } })
    expect(result.reason).toBe("a check is failing")
  })

  it("returns a structured error when the PR cannot be read", async () => {
    const result = await run({ "pr view": { stdout: "", stderr: "Could not resolve to a PullRequest", exitCode: 1 } })
    expect(result.error).toBe(true)
  })

  it("rejects an invalid PR number", async () => {
    const { hooks, restore } = await setupGhMockedPlugin({ spawnDefaults: FIXTURES, cleanupRegistry: scratchDirectories })
    try {
      const result = await hooks.tool.resolve_review_mode.execute({ pr_number: 0 }, {} as never)
      expect(JSON.parse(result.output)).toMatchObject({ error: true, code: "api_error" })
    } finally {
      restore()
    }
  })
})
