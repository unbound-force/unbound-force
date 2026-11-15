import { rm } from "node:fs/promises"

import { afterEach, describe, expect, it } from "vitest"

import {
  assessPriorReviewStaleness,
  type StalenessInput,
} from "../plugins/review-dispatch/index.js"
import { setupGhMockedPlugin, type MockSpawnResponses } from "./helpers/gh-spawn-mock.js"

// ── assessPriorReviewStaleness (pure) ────────────────────────

const HEAD = "b".repeat(40)
const PRIOR = "a".repeat(40)

function input(overrides: Partial<StalenessInput> = {}): StalenessInput {
  return {
    headSha: HEAD,
    viewer: "uf-divisor",
    verdict: "REQUEST_CHANGES",
    reviews: [
      { id: 1, user: "uf-divisor", state: "CHANGES_REQUESTED", commit_id: PRIOR, submitted_at: "2026-11-08T10:00:00Z", body: "", html_url: "https://github.com/acme/repo/pull/7#review-1" },
    ],
    comments: [
      { id: 11, review_id: 1, path: "a.go", line: 3, body: "**HIGH — x**" },
      { id: 12, review_id: 1, path: "b.go", line: 9, body: "**LOW — y**" },
      { id: 13, review_id: 1, path: "c.go", line: null, body: "No severity marker" },
      { id: 14, review_id: 99, path: "d.go", line: 1, body: "**HIGH — another review**" },
    ],
    ...overrides,
  }
}

describe("assessPriorReviewStaleness", () => {
  it("maps the verdict, surfaces the prior review, and flags a stale commit (issue #1055)", () => {
    // The #1055 incident: a same-verdict prior review sits on an older
    // commit, so it is stale, never a true duplicate.
    const result = assessPriorReviewStaleness(input())
    expect(result.matched_state).toBe("CHANGES_REQUESTED")
    expect(result.same_verdict_review).toEqual({
      id: 1,
      html_url: "https://github.com/acme/repo/pull/7#review-1",
      state: "CHANGES_REQUESTED",
      commit_id: PRIOR,
      submitted_at: "2026-11-08T10:00:00Z",
    })
    expect(result.different_verdict_review).toBeNull()
    expect(result.commit_id_matches_head).toBe(false)
  })

  it("reports a commit match only when the same-verdict review is on the head", () => {
    const result = assessPriorReviewStaleness(input({ reviews: [{ ...input().reviews[0]!, commit_id: HEAD }] }))
    expect(result.commit_id_matches_head).toBe(true)
  })

  it("lists the same-verdict review's HIGH and unlabelled findings, not LOW or other reviews'", () => {
    const result = assessPriorReviewStaleness(input())
    expect(result.prior_findings).toEqual([
      { comment_id: 11, path: "a.go", line: 3, severity: "HIGH" },
      { comment_id: 13, path: "c.go", line: null, severity: "UNKNOWN" },
    ])
  })

  it("excludes PENDING and DISMISSED reviews and picks the latest standing same-verdict review", () => {
    const result = assessPriorReviewStaleness(
      input({
        reviews: [
          { id: 1, user: "uf-divisor", state: "DISMISSED", commit_id: HEAD, submitted_at: "2026-11-09T10:00:00Z", body: "" },
          { id: 2, user: "uf-divisor", state: "PENDING", commit_id: HEAD, submitted_at: "2026-11-09T11:00:00Z", body: "" },
          { id: 3, user: "uf-divisor", state: "CHANGES_REQUESTED", commit_id: PRIOR, submitted_at: "2026-11-09T09:00:00Z", body: "" },
          { id: 4, user: "uf-divisor", state: "CHANGES_REQUESTED", commit_id: "c".repeat(40), submitted_at: "2026-11-09T12:00:00Z", body: "" },
        ],
      }),
    )
    expect(result.same_verdict_review?.id).toBe(4)
    expect(result.different_verdict_review).toBeNull()
  })

  it("selects the latest standing different-verdict review as the fallback", () => {
    const result = assessPriorReviewStaleness(
      input({
        reviews: [
          { id: 5, user: "uf-divisor", state: "APPROVED", commit_id: PRIOR, submitted_at: "2026-11-09T08:00:00Z", body: "" },
          { id: 6, user: "uf-divisor", state: "APPROVED", commit_id: "c".repeat(40), submitted_at: "2026-11-09T13:00:00Z", body: "" },
        ],
      }),
    )
    expect(result.same_verdict_review).toBeNull()
    expect(result.different_verdict_review?.id).toBe(6)
    expect(result.prior_findings).toEqual([])
  })

  it("ignores other users' reviews", () => {
    const result = assessPriorReviewStaleness(
      input({ reviews: [{ id: 7, user: "someone", state: "CHANGES_REQUESTED", commit_id: PRIOR, submitted_at: "2026-11-09T10:00:00Z", body: "" }] }),
    )
    expect(result.same_verdict_review).toBeNull()
    expect(result.different_verdict_review).toBeNull()
  })

  it("maps APPROVE and COMMENT verdicts to their standing states", () => {
    const approve = assessPriorReviewStaleness(input({ verdict: "APPROVE" }))
    expect(approve.matched_state).toBe("APPROVED")
    expect(approve.same_verdict_review).toBeNull()
    expect(approve.different_verdict_review?.state).toBe("CHANGES_REQUESTED")
    const comment = assessPriorReviewStaleness(input({ verdict: "COMMENT" }))
    expect(comment.matched_state).toBe("COMMENTED")
  })
})

// ── assess_prior_review_staleness tool (gh mocked) ───────────

const scratchDirectories: string[] = []

afterEach(async () => {
  for (const directory of scratchDirectories.splice(0)) {
    await rm(directory, { recursive: true, force: true })
  }
})

const ok = (value: unknown) => ({ stdout: typeof value === "string" ? value : JSON.stringify(value), stderr: "", exitCode: 0 })

const FIXTURES: MockSpawnResponses = {
  "pr view": ok({ url: "https://github.com/acme/repo/pull/7" }),
  "api user": ok("uf-divisor\n"),
  "api reviews": ok([[{ id: 1, user: { login: "uf-divisor" }, state: "CHANGES_REQUESTED", commit_id: PRIOR, submitted_at: "2026-11-08T10:00:00Z", body: "", html_url: "https://github.com/acme/repo/pull/7#review-1" }]]),
  "api comments": ok([[{ id: 11, pull_request_review_id: 1, path: "a.go", line: 3, body: "**HIGH — x**" }]]),
}

async function run(overrides: MockSpawnResponses = {}): Promise<Record<string, unknown>> {
  const { hooks, mockSpawn, restore } = await setupGhMockedPlugin({
    spawnDefaults: FIXTURES,
    spawnOverrides: overrides,
    cleanupRegistry: scratchDirectories,
  })
  try {
    const result = await hooks.tool.assess_prior_review_staleness.execute({ pr_number: 7, verdict: "REQUEST_CHANGES", head_sha: HEAD }, {} as never)
    const parsed = JSON.parse(result.output) as Record<string, unknown>
    parsed.__calls = mockSpawn.mock.calls.map((call: unknown[]) => (call[0] as string[]).join(" "))
    return parsed
  } finally {
    restore()
  }
}

describe("assess_prior_review_staleness tool", () => {
  it("returns the full staleness assessment", async () => {
    const result = await run()
    expect(result.pr_number).toBe(7)
    expect(result.head_sha).toBe(HEAD)
    expect(result.verdict).toBe("REQUEST_CHANGES")
    expect(result.matched_state).toBe("CHANGES_REQUESTED")
    expect(result.same_verdict_review).toMatchObject({ id: 1, html_url: "https://github.com/acme/repo/pull/7#review-1", commit_id: PRIOR })
    expect(result.different_verdict_review).toBeNull()
    expect(result.commit_id_matches_head).toBe(false)
    expect(result.prior_findings).toEqual([{ comment_id: 11, path: "a.go", line: 3, severity: "HIGH" }])
    expect(result.__calls).toContain("gh api repos/acme/repo/pulls/7/reviews --paginate --slurp")
  })

  it("errors when the reviews cannot be read", async () => {
    const result = await run({ "api reviews": { stdout: "", stderr: "HTTP 403", exitCode: 1 } })
    expect(result.error).toBe(true)
    expect(result.code).toBe("api_error")
    expect(result.retryable).toBe(true)
  })

  it("errors when the viewer cannot be read", async () => {
    const result = await run({ "api user": { stdout: "", stderr: "HTTP 401", exitCode: 1 } })
    expect(result.error).toBe(true)
    expect(result.message).toBe("the current user could not be read")
  })

  it("returns a structured error when the PR cannot be read", async () => {
    const result = await run({ "pr view": { stdout: "", stderr: "Could not resolve to a PullRequest", exitCode: 1 } })
    expect(result.error).toBe(true)
  })

  it("rejects an invalid verdict and head SHA", async () => {
    const { hooks, restore } = await setupGhMockedPlugin({ spawnDefaults: FIXTURES, cleanupRegistry: scratchDirectories })
    try {
      const bad = await hooks.tool.assess_prior_review_staleness.execute({ pr_number: 7, verdict: "APPROVE_WITH_ADVISORIES", head_sha: HEAD }, {} as never)
      expect(JSON.parse(bad.output)).toMatchObject({ error: true, code: "api_error" })
      const badSha = await hooks.tool.assess_prior_review_staleness.execute({ pr_number: 7, verdict: "APPROVE", head_sha: "zzz" }, {} as never)
      expect(JSON.parse(badSha.output)).toMatchObject({ error: true, code: "api_error" })
    } finally {
      restore()
    }
  })
})
