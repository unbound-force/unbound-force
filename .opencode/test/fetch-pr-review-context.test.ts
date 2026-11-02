import { mkdtemp, rm, stat } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"

import type { PluginInput } from "@opencode-ai/plugin"
import { afterEach, describe, expect, it, vi } from "vitest"

import ReviewDispatchPlugin, { _getSessionCorrelationMap, parseReviewerManifest } from "../plugins/review-dispatch/index.js"
import {
  countDiffLines,
  DIFF_THRESHOLD,
} from "../lib/review-diff-manifest.js"
import manifestFixture from "./fixtures/review-dispatch/manifest.json"
import {
  mockProcess,
  patchBunSpawn,
  setupGhMockedPlugin,
  textStream,
  writeReviewPolicyFiles,
  type MockSpawnFn,
  type MockSpawnProcess,
  type MockSpawnResponses,
} from "./helpers/gh-spawn-mock.js"

// ── Fixtures ─────────────────────────────────────────────────

const MOCK_PR_META = {
  number: 42,
  title: "Add feature X",
  body: "This PR adds feature X",
  baseRefName: "main",
  baseRefOid: "abc123def456abc123def456abc123def456abc1",
  headRefName: "feature/x",
  headRefOid: "def456abc123def456abc123def456abc123def4",
  // PR-repair fields (modified FR-003, design D2): state, mergeable,
  // mergeStateStatus, and headRepository pass through as gh returns
  // them; base_repository is derived from the base-rooted url.
  state: "OPEN",
  mergeable: "MERGEABLE",
  mergeStateStatus: "CLEAN",
  headRepository: { nameWithOwner: "unbound-force/unbound-force" },
  url: "https://github.com/unbound-force/unbound-force/pull/42",
  files: [
    { path: "src/index.ts", additions: 10, deletions: 5 },
    { path: "test/index.test.ts", additions: 20, deletions: 0 },
  ],
}

/**
 * Fork-PR fixture (modified FR-003 fork-orientation scenario): the
 * head repository is a fork of the base while the PR url stays rooted
 * at the base repository, so head_repository differs from
 * base_repository (design D2).
 */
const MOCK_PR_META_FORK = {
  ...MOCK_PR_META,
  // A conflicted fork PR: gh's mergeable enum carries CONFLICTING and
  // mergeStateStatus DIRTY, and headRepository names the fork.
  mergeable: "CONFLICTING",
  mergeStateStatus: "DIRTY",
  headRepository: { nameWithOwner: "jflowers/unbound-force" },
}

const MOCK_CI_CHECKS = [
  { name: "CI / build", state: "SUCCESS", bucket: "pass" },
  { name: "CI / test", state: "SUCCESS", bucket: "pass" },
]

const MOCK_REPO = { nameWithOwner: "unbound-force/unbound-force" }

const MOCK_REVIEWS = [
  { user: { login: "reviewer1" }, state: "APPROVED", body: "LGTM", submitted_at: "2026-01-01T00:00:00Z" },
]

const MOCK_COMMENTS = [
  { user: { login: "reviewer1" }, path: "src/index.ts", line: 5, body: "nit: rename this" },
]

const MOCK_DIFF = "diff --git a/src/index.ts b/src/index.ts\n--- a/src/index.ts\n+++ b/src/index.ts\n@@ -1,3 +1,8 @@\n+import { foo } from 'bar'\n"

// ── Mock fixture responses (harness: ./helpers/gh-spawn-mock.ts) ──

/** Suite fixture responses for every gh route fetch_pr_review_context fetches; the which/auth prologue defaults live in the harness. */
const SPAWN_FIXTURES: MockSpawnResponses = {
  "pr view": { stdout: JSON.stringify(MOCK_PR_META), stderr: "", exitCode: 0 },
  "pr checks": { stdout: JSON.stringify(MOCK_CI_CHECKS), stderr: "", exitCode: 0 },
  "pr diff": { stdout: MOCK_DIFF, stderr: "", exitCode: 0 },
  "repo view": { stdout: JSON.stringify(MOCK_REPO), stderr: "", exitCode: 0 },
  "api reviews": { stdout: JSON.stringify(MOCK_REVIEWS), stderr: "", exitCode: 0 },
  "api comments": { stdout: JSON.stringify(MOCK_COMMENTS), stderr: "", exitCode: 0 },
}

// ── Test infrastructure ──────────────────────────────────────

const scratchDirectories: string[] = []

/**
 * Sets up the plugin hooks with Bun.spawn mocked through the shared
 * gh-spawn-mock harness, selecting the fetch_pr_review_context tool.
 */
async function setupTool(spawnOverrides: MockSpawnResponses = {}) {
  const { hooks, mockSpawn, tmpDir, restore } = await setupGhMockedPlugin({
    spawnDefaults: SPAWN_FIXTURES,
    spawnOverrides,
    cleanupRegistry: scratchDirectories,
  })
  return { tool: hooks.tool.fetch_pr_review_context, mockSpawn, tmpDir, restore }
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
 * Calls the tool with NO input arguments — pr_number omitted, the
 * current-branch resolution path (modified FR-002, design D3) — and
 * returns the parsed output with the spawn mock for argv pinning.
 */
async function callToolWithoutPrNumber(
  overrides: MockSpawnResponses = {},
): Promise<{ parsed: Record<string, unknown>; mockSpawn: MockSpawnFn; tmpDir: string; restore: () => void }> {
  const { tool, mockSpawn, tmpDir, restore } = await setupTool(overrides)
  const result = await tool.execute({}, {} as never)
  const parsed = JSON.parse(result.output) as Record<string, unknown>
  return { parsed, mockSpawn, tmpDir, restore }
}

/**
 * Return the argv of the first recorded `gh <group> <verb>` spawn call
 * (e.g. group "pr", verb "view"), or undefined when the tool never
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

describe("fetch_pr_review_context", () => {
  describe("successful output", () => {
    it("returns structured output for a valid PR", async () => {
      const { parsed, restore } = await callTool()
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.pr_number).toBe(42)
        expect(parsed.title).toBe("Add feature X")
        expect(parsed.body).toBe("This PR adds feature X")
        expect(parsed.base_ref).toBe("main")
        expect(parsed.base_sha).toBe("abc123def456abc123def456abc123def456abc1")
        expect(parsed.head_ref).toBe("feature/x")
        expect(parsed.head_sha).toBe("def456abc123def456abc123def456abc123def4")
        expect(parsed.files).toEqual([
          { path: "src/index.ts", additions: 10, deletions: 5 },
          { path: "test/index.test.ts", additions: 20, deletions: 0 },
        ])
        expect(typeof parsed.diff_path).toBe("string")
        expect(typeof parsed.diff_lines).toBe("number")
        expect(typeof parsed.diff_bytes).toBe("number")
        expect(typeof parsed.exceeds_threshold).toBe("boolean")
        expect(Array.isArray(parsed.ci_checks)).toBe(true)
        expect(typeof parsed.ci_pending).toBe("boolean")
        expect(typeof parsed.input_context).toBe("string")
        expect(typeof parsed.changed_files).toBe("string")
        expect(Array.isArray(parsed.warnings)).toBe(true)
      } finally {
        restore()
      }
    })

    it("saves diff to temp file at the expected path", async () => {
      const { parsed, restore } = await callTool()
      try {
        const diffPath = parsed.diff_path as string
        expect(diffPath).toContain("pr-42-diff.patch")
        expect(diffPath).toContain(join(tmpdir(), "opencode"))

        // Verify file actually exists and contains the diff
        const { readFile } = await import("node:fs/promises")
        const content = await readFile(diffPath, "utf8")
        expect(content).toBe(MOCK_DIFF)
      } finally {
        restore()
      }
    })

    it("creates diff file with 0o600 permissions", async () => {
      const { parsed, restore } = await callTool()
      try {
        const diffPath = parsed.diff_path as string
        const fileStats = await stat(diffPath)
        // Check owner-only read/write (0o600 = 0o100600 for regular file)
        const mode = fileStats.mode & 0o777
        expect(mode).toBe(0o600)
      } finally {
        restore()
      }
    })

    it("pre-computes input_context with correct fields", async () => {
      const { parsed, restore } = await callTool()
      try {
        const inputContext = parsed.input_context as string
        expect(inputContext).toContain("kind: pr")
        expect(inputContext).toContain("pr_number: 42")
        expect(inputContext).toContain("base_ref: main")
        expect(inputContext).toContain("base_sha: abc123def456abc123def456abc123def456abc1")
        expect(inputContext).toContain("head_ref: feature/x")
        expect(inputContext).toContain("head_sha: def456abc123def456abc123def456abc123def4")
      } finally {
        restore()
      }
    })

    it("pre-computes changed_files as formatted string", async () => {
      const { parsed, restore } = await callTool()
      try {
        const changedFiles = parsed.changed_files as string
        expect(changedFiles).toContain("src/index.ts (+10/-5)")
        expect(changedFiles).toContain("test/index.test.ts (+20/-0)")
      } finally {
        restore()
      }
    })

    it("pre-computes exceeds_threshold correctly when below threshold", async () => {
      const { parsed, restore } = await callTool()
      try {
        expect(parsed.exceeds_threshold).toBe(false)
        expect(parsed.diff_lines).toBeLessThanOrEqual(DIFF_THRESHOLD)
      } finally {
        restore()
      }
    })

    it("pre-computes exceeds_threshold correctly when above threshold", async () => {
      // Generate a diff with more than DIFF_THRESHOLD lines
      const largeDiff = Array.from({ length: DIFF_THRESHOLD + 10 }, (_, i) => `+line ${i}`).join("\n")
      const { parsed, restore } = await callTool({ "pr diff": { stdout: largeDiff, stderr: "", exitCode: 0 } })
      try {
        expect(parsed.exceeds_threshold).toBe(true)
        const warnings = parsed.warnings as string[]
        expect(warnings.some((w: string) => w.includes("exceeds size advisory threshold"))).toBe(true)
      } finally {
        restore()
      }
    })

    it("pre-computes ci_pending correctly when all checks are terminal", async () => {
      const { parsed, restore } = await callTool()
      try {
        expect(parsed.ci_pending).toBe(false)
        expect(parsed.ci_checks).toEqual([
          { name: "CI / build", status: "SUCCESS", conclusion: "pass" },
          { name: "CI / test", status: "SUCCESS", conclusion: "pass" },
        ])
        expect(parsed.warnings as string[]).toHaveLength(0)
      } finally {
        restore()
      }
    })

    it("pre-computes ci_pending correctly when checks are pending", async () => {
      const pendingChecks = [
        { name: "CI / build", state: "PENDING", bucket: "pending" },
        { name: "CI / test", state: "SUCCESS", bucket: "pass" },
      ]
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: JSON.stringify(pendingChecks), stderr: "", exitCode: 8 },
      })
      try {
        expect(parsed.ci_pending).toBe(true)
        const warnings = parsed.warnings as string[]
        expect(warnings.some((w: string) => w.includes("CI checks are still in progress"))).toBe(true)
      } finally {
        restore()
      }
    })

    it("populates ci_checks for a failing PR on exit 1 without degrading", async () => {
      const failingChecks = [
        { name: "CI / build", state: "FAILURE", bucket: "fail" },
      ]
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: JSON.stringify(failingChecks), stderr: "", exitCode: 1 },
      })
      try {
        expect(parsed.ci_checks).toEqual([
          { name: "CI / build", status: "FAILURE", conclusion: "fail" },
        ])
        expect(parsed.ci_pending).toBe(false)
        const warnings = parsed.warnings as string[]
        expect(warnings.some((w: string) => w.includes("ci_checks degraded"))).toBe(false)
      } finally {
        restore()
      }
    })

    it("surfaces an unknown bucket verbatim without degrading", async () => {
      const unknownChecks = [
        { name: "CI / build", state: "NEUTRAL", bucket: "blocked" },
      ]
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: JSON.stringify(unknownChecks), stderr: "", exitCode: 0 },
      })
      try {
        expect(parsed.ci_checks).toEqual([
          { name: "CI / build", status: "NEUTRAL", conclusion: "blocked" },
        ])
        expect(parsed.ci_pending).toBe(false)
        const warnings = parsed.warnings as string[]
        expect(warnings.some((w: string) => w.includes("ci_checks degraded"))).toBe(false)
      } finally {
        restore()
      }
    })

    it("treats the checkless-PR marker as empty ci_checks without degrading", async () => {
      const { parsed, restore } = await callTool({
        "pr checks": {
          stdout: "",
          stderr: "no checks reported on the 'feature/x' branch",
          exitCode: 1,
        },
      })
      try {
        expect(parsed.ci_checks).toEqual([])
        expect(parsed.ci_pending).toBe(false)
        const warnings = parsed.warnings as string[]
        expect(warnings.some((w: string) => w.includes("ci_checks degraded"))).toBe(false)
      } finally {
        restore()
      }
    })

    it("includes existing reviews in output", async () => {
      const { parsed, restore } = await callTool()
      try {
        const reviews = parsed.existing_reviews as Array<Record<string, unknown>>
        expect(reviews).toHaveLength(1)
        expect(reviews[0].author).toBe("reviewer1")
        expect(reviews[0].state).toBe("APPROVED")
        expect(reviews[0].body).toBe("LGTM")
        expect(reviews[0].submitted_at).toBe("2026-01-01T00:00:00Z")
      } finally {
        restore()
      }
    })

    it("includes inline comments in output", async () => {
      const { parsed, restore } = await callTool()
      try {
        const comments = parsed.inline_comments as Array<Record<string, unknown>>
        expect(comments).toHaveLength(1)
        expect(comments[0].author).toBe("reviewer1")
        expect(comments[0].path).toBe("src/index.ts")
        expect(comments[0].line).toBe(5)
        expect(comments[0].body).toBe("nit: rename this")
      } finally {
        restore()
      }
    })
  })

  describe("PR-repair fields and optional pr_number (modified FR-002/FR-003)", () => {
    it("passes the five new fields through as gh returns them", async () => {
      const { parsed, mockSpawn, restore } = await callTool()
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.state).toBe("OPEN")
        expect(parsed.mergeable).toBe("MERGEABLE")
        expect(parsed.merge_state_status).toBe("CLEAN")
        expect(parsed.head_repository).toBe("unbound-force/unbound-force")
        expect(parsed.base_repository).toBe("unbound-force/unbound-force")

        // Design D2: the gh pr view --json field list grows the five
        // PR-repair sources and must not request baseRepository, which
        // gh rejects as an unknown field.
        const viewArgs = findGhArgv(mockSpawn, "pr", "view")
        expect(viewArgs).toBeDefined()
        const jsonFields = viewArgs![viewArgs!.indexOf("--json") + 1].split(",")
        expect(jsonFields).toEqual(
          expect.arrayContaining(["state", "mergeable", "mergeStateStatus", "headRepository", "url"]),
        )
        expect(jsonFields).not.toContain("baseRepository")
      } finally {
        restore()
      }
    })

    it("reports fork orientation: head_repository differs from base_repository", async () => {
      const { parsed, restore } = await callTool({
        "pr view": { stdout: JSON.stringify(MOCK_PR_META_FORK), stderr: "", exitCode: 0 },
      })
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.state).toBe("OPEN")
        expect(parsed.mergeable).toBe("CONFLICTING")
        expect(parsed.merge_state_status).toBe("DIRTY")
        // Fork orientation: head_repository names the fork while
        // base_repository is parsed from the base-rooted url.
        expect(parsed.head_repository).toBe("jflowers/unbound-force")
        expect(parsed.base_repository).toBe("unbound-force/unbound-force")
        expect(parsed.head_repository).not.toBe(parsed.base_repository)
      } finally {
        restore()
      }
    })

    it("derives base_repository from the PR url", async () => {
      // Only the url varies from the default fixture: base_repository
      // must track it even when it names a repository different from
      // both headRepository and the gh repo view slug (which the
      // fixture keeps at unbound-force/unbound-force).
      const urlVariant = {
        ...MOCK_PR_META,
        url: "https://github.com/octocat/other-repo/pull/42",
      }
      const { parsed, restore } = await callTool({
        "pr view": { stdout: JSON.stringify(urlVariant), stderr: "", exitCode: 0 },
      })
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.base_repository).toBe("octocat/other-repo")
        expect(parsed.head_repository).toBe("unbound-force/unbound-force")
      } finally {
        restore()
      }
    })

    it("degrades base_repository to empty string with a warning when the url is unparseable", async () => {
      const unparseableUrl = {
        ...MOCK_PR_META,
        url: "https://gitlab.com/unbound-force/unbound-force/-/merge_requests/42",
      }
      const { parsed, restore } = await callTool({
        "pr view": { stdout: JSON.stringify(unparseableUrl), stderr: "", exitCode: 0 },
      })
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.base_repository).toBe("")
        const warnings = parsed.warnings as string[]
        expect(
          warnings.some((w: string) => w.includes("base_repository degraded") && w.includes("unparseable PR url")),
        ).toBe(true)
      } finally {
        restore()
      }
    })

    it("resolves the current branch's open PR when pr_number is omitted", async () => {
      const { parsed, mockSpawn, restore } = await callToolWithoutPrNumber()
      try {
        expect(parsed.error).toBeUndefined()
        // gh resolved the current branch's PR; the resolved number is
        // reported in the output (design D3).
        expect(parsed.pr_number).toBe(42)

        // The gh pr view invocation carried no positional number: the
        // argument after "view" is "--json", not a PR number.
        const viewArgs = findGhArgv(mockSpawn, "pr", "view")
        expect(viewArgs).toBeDefined()
        expect(viewArgs![3]).toBe("--json")

        // Subsequent fetches anchor on the resolved number.
        const checksArgs = findGhArgv(mockSpawn, "pr", "checks")
        expect(checksArgs).toBeDefined()
        expect(checksArgs![3]).toBe("42")
      } finally {
        restore()
      }
    })

    it("returns pr_not_found when the current branch has no open PR", async () => {
      const { parsed, restore } = await callToolWithoutPrNumber({
        "pr view": { stdout: "", stderr: 'no pull requests found for branch "feature/x"', exitCode: 1 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("pr_not_found")
        expect(parsed.retryable).toBe(false)
        expect(parsed.message).toBe("No open pull request found for the current branch")
      } finally {
        restore()
      }
    })

    it("keeps the explicit pr_number path unchanged", async () => {
      const { parsed, mockSpawn, restore } = await callTool({}, 42)
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.pr_number).toBe(42)

        // The explicit number still anchors the fetch as a discrete
        // positional argument.
        const viewArgs = findGhArgv(mockSpawn, "pr", "view")
        expect(viewArgs).toBeDefined()
        expect(viewArgs![3]).toBe("42")

        // The extended surface is present on the explicit path too.
        expect(parsed.state).toBe("OPEN")
      } finally {
        restore()
      }
    })

    it("still rejects extra fields via .strict() when pr_number is omitted", async () => {
      const { tool, restore } = await setupTool()
      try {
        const result = await tool.execute({ evil: "payload" } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
      } finally {
        restore()
      }
    })
  })

  describe("warnings for ci_pending and exceeds_threshold (FR-010)", () => {
    it("populates no warnings when CI passes and diff is small", async () => {
      const { parsed, restore } = await callTool()
      try {
        const warnings = parsed.warnings as string[]
        expect(warnings).toHaveLength(0)
      } finally {
        restore()
      }
    })

    it("populates warning for ci_pending", async () => {
      const pendingChecks = [{ name: "CI / build", state: "PENDING", bucket: "pending" }]
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: JSON.stringify(pendingChecks), stderr: "", exitCode: 0 },
      })
      try {
        const warnings = parsed.warnings as string[]
        expect(warnings).toContain("CI checks are still in progress")
      } finally {
        restore()
      }
    })

    it("populates warning for exceeds_threshold", async () => {
      const largeDiff = Array.from({ length: DIFF_THRESHOLD + 10 }, (_, i) => `+line ${i}`).join("\n")
      const { parsed, restore } = await callTool({ "pr diff": { stdout: largeDiff, stderr: "", exitCode: 0 } })
      try {
        const warnings = parsed.warnings as string[]
        expect(warnings.some((w: string) => w.includes(`${DIFF_THRESHOLD} lines`))).toBe(true)
      } finally {
        restore()
      }
    })

    it("populates both warnings when both conditions apply", async () => {
      const pendingChecks = [{ name: "CI / lint", state: "PENDING", bucket: "pending" }]
      const largeDiff = Array.from({ length: DIFF_THRESHOLD + 10 }, (_, i) => `+line ${i}`).join("\n")
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: JSON.stringify(pendingChecks), stderr: "", exitCode: 0 },
        "pr diff": { stdout: largeDiff, stderr: "", exitCode: 0 },
      })
      try {
        const warnings = parsed.warnings as string[]
        expect(warnings).toHaveLength(2)
        expect(warnings.some((w: string) => w.includes("CI checks are still in progress"))).toBe(true)
        expect(warnings.some((w: string) => w.includes("exceeds size advisory threshold"))).toBe(true)
      } finally {
        restore()
      }
    })
  })

  describe("shared line-counting rule (RD-FR-032)", () => {
    it("computes diff_lines with the shared counting rule: a trailing terminator does not add a line", async () => {
      // MOCK_DIFF ends with a trailing newline; the legacy
      // split("\n").length rule counted the empty trailing segment as a
      // line. The shared countDiffLines rule must not.
      const { parsed, restore } = await callTool()
      try {
        expect(parsed.diff_lines).toBe(countDiffLines(MOCK_DIFF))
        expect(MOCK_DIFF).toMatch(/\n$/)
        expect(countDiffLines(MOCK_DIFF)).toBe(MOCK_DIFF.split("\n").length - 1)
      } finally {
        restore()
      }
    })

    it("boundary agreement: exactly 1500 lines reports exceeds_threshold false and no warning", async () => {
      const boundaryDiff = Array.from({ length: DIFF_THRESHOLD }, (_, i) => `+line ${i}`).join("\n")
      const { parsed, restore } = await callTool({ "pr diff": { stdout: boundaryDiff, stderr: "", exitCode: 0 } })
      try {
        expect(parsed.diff_lines).toBe(DIFF_THRESHOLD)
        expect(parsed.exceeds_threshold).toBe(false)
        expect(parsed.warnings).toHaveLength(0)
      } finally {
        restore()
      }
    })

    it("boundary agreement: 1501 lines reports exceeds_threshold true", async () => {
      const boundaryDiff = Array.from({ length: DIFF_THRESHOLD + 1 }, (_, i) => `+line ${i}`).join("\n")
      const { parsed, restore } = await callTool({ "pr diff": { stdout: boundaryDiff, stderr: "", exitCode: 0 } })
      try {
        expect(parsed.diff_lines).toBe(DIFF_THRESHOLD + 1)
        expect(parsed.exceeds_threshold).toBe(true)
      } finally {
        restore()
      }
    })
  })

  describe("manifest-delivery warning (FR-010 delta)", () => {
    it("states that manifest delivery is engaged while preserving the pinned threshold substring", async () => {
      const largeDiff = Array.from({ length: DIFF_THRESHOLD + 10 }, (_, i) => `+line ${i}`).join("\n")
      const { parsed, restore } = await callTool({ "pr diff": { stdout: largeDiff, stderr: "", exitCode: 0 } })
      try {
        const warnings = parsed.warnings as string[]
        const thresholdWarning = warnings.find((w: string) => w.includes("exceeds size advisory threshold"))
        expect(thresholdWarning).toBeDefined()
        expect(thresholdWarning).toContain("manifest delivery is engaged")
        expect(thresholdWarning).toContain("read_diff_chunk")
      } finally {
        restore()
      }
    })
  })

  describe("error handling", () => {
    it("returns gh_not_found error when gh CLI is not available", async () => {
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

    it("returns gh_auth_failed error when gh auth status fails", async () => {
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

    it("returns pr_not_found error when PR does not exist", async () => {
      const { parsed, restore } = await callTool({
        "pr view": { stdout: "", stderr: "Could not find pull request", exitCode: 1 },
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

    it("returns api_error for general PR view failures", async () => {
      const { parsed, restore } = await callTool({
        "pr view": { stdout: "", stderr: "server error 500", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
      } finally {
        restore()
      }
    })

    it("returns api_error when diff fetch fails", async () => {
      const { parsed, restore } = await callTool({
        "pr diff": { stdout: "", stderr: "Internal server error", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
      } finally {
        restore()
      }
    })

    it("returns timeout error when gh CLI reports timeout", async () => {
      const { parsed, restore } = await callTool({
        "pr view": { stdout: "", stderr: "request timed out", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("timeout")
        expect(parsed.retryable).toBe(true)
      } finally {
        restore()
      }
    })

    it("returns rate_limited error when GitHub rate limits", async () => {
      const { parsed, restore } = await callTool({
        "pr view": { stdout: "", stderr: "API rate limit exceeded", exitCode: 4 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("rate_limited")
        expect(parsed.retryable).toBe(true)
      } finally {
        restore()
      }
    })

    it("returns api_error when repo view fails", async () => {
      const { parsed, restore } = await callTool({
        "repo view": { stdout: "", stderr: "fatal: not a git repository", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
      } finally {
        restore()
      }
    })
  })

  describe("non-fatal degradation", () => {
    it("degrades gracefully when CI checks fail with warning", async () => {
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: "", stderr: "no checks configured", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.ci_checks).toEqual([])
        const warnings = parsed.warnings as string[]
        expect(warnings.some((w: string) => w.includes("ci_checks degraded"))).toBe(true)
      } finally {
        restore()
      }
    })

    it("degrades gracefully when reviews fetch fails with warning", async () => {
      const { parsed, restore } = await callTool({
        "api reviews": { stdout: "", stderr: "forbidden", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.existing_reviews).toEqual([])
        const warnings = parsed.warnings as string[]
        expect(warnings.some((w: string) => w.includes("existing_reviews degraded"))).toBe(true)
      } finally {
        restore()
      }
    })

    it("degrades gracefully when inline comments fetch fails with warning", async () => {
      const { parsed, restore } = await callTool({
        "api comments": { stdout: "", stderr: "not found", exitCode: 1 },
      })
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.inline_comments).toEqual([])
        const warnings = parsed.warnings as string[]
        expect(warnings.some((w: string) => w.includes("inline_comments degraded"))).toBe(true)
      } finally {
        restore()
      }
    })
  })

  describe("ci_checks sub-call gh schema alignment (FPR-FR-017)", () => {
    it("degrades the advisory sub-call when pr checks stderr matches the poller's PR-not-found phrases", async () => {
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: "", stderr: "Could not resolve to a PullRequest with the number of 42", exitCode: 1 },
      })
      try {
        // The stderr matches uf_finale_ci_poll's PR_CHECKS_NOT_FOUND_PATTERN
        // ("could not resolve to a pullrequest"), but this sub-call does not
        // own PR-not-found detection — the preceding gh pr view does. The
        // sub-call must degrade generically with the stderr carried rather
        // than surfacing pr_not_found (FPR-FR-017 divergence).
        expect(parsed.error).toBeUndefined()
        expect(parsed.ci_checks).toEqual([])
        const warnings = parsed.warnings as string[]
        expect(warnings.some((w: string) => w.includes("ci_checks degraded"))).toBe(true)
        expect(
          warnings.some((w: string) => w.includes("Could not resolve to a PullRequest with the number of 42")),
        ).toBe(true)
      } finally {
        restore()
      }
    })

    it("requests exactly the 3-field name,state,bucket json list from the sub-call", async () => {
      const { parsed, mockSpawn, restore } = await callTool()
      try {
        expect(parsed.error).toBeUndefined()
        // gh 2.91.0 rejects `conclusion` as a pr checks --json field; the
        // sub-call requests exactly the 3 fields it maps — never the
        // poller's 4-field name,state,bucket,link list (FPR-FR-017
        // divergence from uf_finale_ci_poll).
        const checksArgs = findGhArgv(mockSpawn, "pr", "checks")
        expect(checksArgs).toBeDefined()
        expect(checksArgs).toEqual(["gh", "pr", "checks", "42", "--json", "name,state,bucket"])
      } finally {
        restore()
      }
    })
  })

  describe("error code completeness", () => {
    it("exercises gh_not_found error code", async () => {
      const { parsed, restore } = await callTool({
        which: { stdout: "", stderr: "", exitCode: 1 },
      })
      try {
        expect(parsed.code).toBe("gh_not_found")
      } finally {
        restore()
      }
    })

    it("exercises gh_auth_failed error code", async () => {
      const { parsed, restore } = await callTool({
        "auth status": { stdout: "", stderr: "auth failed", exitCode: 1 },
      })
      try {
        expect(parsed.code).toBe("gh_auth_failed")
      } finally {
        restore()
      }
    })

    it("exercises pr_not_found error code", async () => {
      const { parsed, restore } = await callTool({
        "pr view": { stdout: "", stderr: "no pull request found", exitCode: 1 },
      })
      try {
        expect(parsed.code).toBe("pr_not_found")
      } finally {
        restore()
      }
    })

    it("exercises api_error error code", async () => {
      const { parsed, restore } = await callTool({
        "pr diff": { stdout: "", stderr: "unexpected error", exitCode: 1 },
      })
      try {
        expect(parsed.code).toBe("api_error")
      } finally {
        restore()
      }
    })

    it("exercises timeout error code", async () => {
      const { parsed, restore } = await callTool({
        "pr view": { stdout: "", stderr: "connection timed out", exitCode: 1 },
      })
      try {
        expect(parsed.code).toBe("timeout")
        expect(parsed.retryable).toBe(true)
      } finally {
        restore()
      }
    })

    it("exercises rate_limited error code", async () => {
      const { parsed, restore } = await callTool({
        "pr view": { stdout: "", stderr: "rate limit exceeded", exitCode: 4 },
      })
      try {
        expect(parsed.code).toBe("rate_limited")
        expect(parsed.retryable).toBe(true)
      } finally {
        restore()
      }
    })

    it("exercises partial_failure via degraded ci_checks, reviews, and comments", async () => {
      const { parsed, restore } = await callTool({
        "pr checks": { stdout: "", stderr: "checks error", exitCode: 1 },
        "api reviews": { stdout: "", stderr: "reviews error", exitCode: 1 },
        "api comments": { stdout: "", stderr: "comments error", exitCode: 1 },
      })
      try {
        // partial_failure manifests as a successful result with degradation warnings
        expect(parsed.error).toBeUndefined()
        expect(parsed.ci_checks).toEqual([])
        expect(parsed.existing_reviews).toEqual([])
        expect(parsed.inline_comments).toEqual([])
        const warnings = parsed.warnings as string[]
        expect(warnings.length).toBeGreaterThanOrEqual(3)
        expect(warnings.some((w: string) => w.includes("ci_checks degraded"))).toBe(true)
        expect(warnings.some((w: string) => w.includes("existing_reviews degraded"))).toBe(true)
        expect(warnings.some((w: string) => w.includes("inline_comments degraded"))).toBe(true)
      } finally {
        restore()
      }
    })
  })

  describe("input validation / security", () => {
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
      const { tool, restore } = await setupTool()
      try {
        const result = await tool.execute({ pr_number: "42; rm -rf /" } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
      } finally {
        restore()
      }
    })

    it("rejects object pr_number (prototype pollution prevention)", async () => {
      const { tool, restore } = await setupTool()
      try {
        const result = await tool.execute({ pr_number: { __proto__: 42 } } as never, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
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

    it("accepts valid edge-case pr_number=1", async () => {
      const { parsed, restore } = await callTool({}, 1)
      try {
        expect(parsed.error).toBeUndefined()
        expect(parsed.pr_number).toBe(42) // from MOCK_PR_META.number
      } finally {
        restore()
      }
    })

    it("accepts valid edge-case pr_number=999999", async () => {
      const { parsed, restore } = await callTool({}, 999_999)
      try {
        expect(parsed.error).toBeUndefined()
      } finally {
        restore()
      }
    })
  })

  describe("diff file write failure", () => {
    it("returns api_error when diff path is blocked by a directory", async () => {
      // Pre-create a directory at the diff path to force open("wx") to fail with EISDIR
      const { mkdir: mkdirFs, rm: rmFs } = await import("node:fs/promises")
      const diffDir = join(tmpdir(), "opencode")
      const diffPath = join(diffDir, "pr-99-diff.patch")
      await mkdirFs(diffPath, { recursive: true })

      const { parsed, restore } = await callTool({}, 99)
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
      } finally {
        restore()
        await rmFs(diffPath, { recursive: true, force: true }).catch(() => {})
      }
    })
  })

  describe("classifyError non-Error fallback branch", () => {
    it("returns api_error with fallback message when a non-Error is thrown", async () => {
      const tmpDir = await mkdtemp(join(tmpdir(), "uf-fetch-ctx-"))
      scratchDirectories.push(tmpDir)
      await writeReviewPolicyFiles(tmpDir)

      // Create a mock where the diff command returns a stream that errors with a non-Error value
      const nonErrorSpawn = vi.fn((args: string[], _options?: unknown): MockSpawnProcess => {
        if (args[0] === "which") {
          return mockProcess({ stdout: "/usr/local/bin/gh\n", stderr: "", exitCode: 0 })
        }
        const sub = args.slice(1, 3).join(" ")
        if (sub === "auth status") return mockProcess({ stdout: "ok", stderr: "", exitCode: 0 })
        if (sub === "pr view") return mockProcess({ stdout: JSON.stringify(MOCK_PR_META), stderr: "", exitCode: 0 })
        if (sub === "pr checks") return mockProcess({ stdout: JSON.stringify(MOCK_CI_CHECKS), stderr: "", exitCode: 0 })
        if (sub === "pr diff") {
          // Return a stream that errors with a non-Error value, which will be caught
          // by classifyError and hit the non-Error fallback branch (line 4149)
          return {
            stdout: new ReadableStream<Uint8Array>({
              start(controller) {
                controller.error("non-error-string-value")
              },
            }),
            stderr: textStream(""),
            exited: Promise.resolve(0),
          }
        }
        return mockProcess({ stdout: "", stderr: "", exitCode: 0 })
      })

      const restore = patchBunSpawn(nonErrorSpawn)

      try {
        const hooks = await ReviewDispatchPlugin.server({
          client: { session: {} },
          directory: tmpDir,
          worktree: tmpDir,
        } as unknown as PluginInput)

        const result = await hooks.tool.fetch_pr_review_context.execute({ pr_number: 42 }, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>

        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect(parsed.retryable).toBe(false)
      } finally {
        restore()
      }
    })
  })

  describe("classifyError DOMException/AbortError branch", () => {
    it("returns timeout for DOMException AbortError during diff fetch", async () => {
      const tmpDir = await mkdtemp(join(tmpdir(), "uf-fetch-ctx-"))
      scratchDirectories.push(tmpDir)
      await writeReviewPolicyFiles(tmpDir)

      // Create a mock where pr diff throws a DOMException AbortError
      const abortSpawn = vi.fn((args: string[], _options?: unknown): MockSpawnProcess => {
        if (args[0] === "which") {
          return mockProcess({ stdout: "/usr/local/bin/gh\n", stderr: "", exitCode: 0 })
        }
        const sub = args.slice(1, 3).join(" ")
        if (sub === "auth status") return mockProcess({ stdout: "ok", stderr: "", exitCode: 0 })
        if (sub === "pr view") return mockProcess({ stdout: JSON.stringify(MOCK_PR_META), stderr: "", exitCode: 0 })
        if (sub === "pr checks") return mockProcess({ stdout: JSON.stringify(MOCK_CI_CHECKS), stderr: "", exitCode: 0 })
        if (sub === "pr diff") {
          // Simulate AbortSignal timeout by throwing a DOMException
          throw new DOMException("The operation was aborted", "AbortError")
        }
        return mockProcess({ stdout: "", stderr: "", exitCode: 0 })
      })

      const restore = patchBunSpawn(abortSpawn)

      try {
        const hooks = await ReviewDispatchPlugin.server({
          client: { session: {} },
          directory: tmpDir,
          worktree: tmpDir,
        } as unknown as PluginInput)

        const result = await hooks.tool.fetch_pr_review_context.execute({ pr_number: 42 }, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>

        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("timeout")
        expect(parsed.retryable).toBe(true)
      } finally {
        restore()
      }
    })
  })

  describe("edge case: Bun.spawn throws during which check", () => {
    it("returns gh_not_found when Bun.spawn throws (not exit code)", async () => {
      const tmpDir = await mkdtemp(join(tmpdir(), "uf-fetch-ctx-"))
      scratchDirectories.push(tmpDir)
      await writeReviewPolicyFiles(tmpDir)

      // Create a mock that THROWS when args[0] === "which"
      const throwingSpawn = vi.fn((args: string[], _options?: unknown): MockSpawnProcess => {
        if (args[0] === "which") {
          throw new Error("spawn ENOENT")
        }
        return mockProcess({ stdout: "", stderr: "", exitCode: 0 })
      })

      const restore = patchBunSpawn(throwingSpawn)

      try {
        const hooks = await ReviewDispatchPlugin.server({
          client: { session: {} },
          directory: tmpDir,
          worktree: tmpDir,
        } as unknown as PluginInput)

        const result = await hooks.tool.fetch_pr_review_context.execute({ pr_number: 42 }, {} as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>

        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("gh_not_found")
        expect(parsed.retryable).toBe(false)
      } finally {
        restore()
      }
    })
  })

  describe("Bun.spawn call verification", () => {
    it("calls which gh as the first spawn", async () => {
      const { mockSpawn, restore } = await callTool()
      try {
        const firstCall = mockSpawn.mock.calls[0]
        expect(firstCall[0]).toEqual(["which", "gh"])
      } finally {
        restore()
      }
    })

    it("calls gh auth status before PR metadata", async () => {
      const { mockSpawn, restore } = await callTool()
      try {
        const calls = mockSpawn.mock.calls.map((c: unknown[]) => (c[0] as string[]).slice(0, 3).join(" "))
        const authIndex = calls.indexOf("gh auth status")
        const prViewIndex = calls.findIndex((c: string) => c.startsWith("gh pr view"))
        expect(authIndex).toBeGreaterThan(-1)
        expect(prViewIndex).toBeGreaterThan(authIndex)
      } finally {
        restore()
      }
    })

    it("passes pr_number as a discrete array element (no shell interpolation)", async () => {
      const { mockSpawn, restore } = await callTool({}, 42)
      try {
        // Find the pr view call and check that prNumber is a separate arg
        const prViewCall = mockSpawn.mock.calls.find(
          (c: unknown[]) => (c[0] as string[])[1] === "pr" && (c[0] as string[])[2] === "view",
        )
        expect(prViewCall).toBeDefined()
        const args = prViewCall![0] as string[]
        expect(args[3]).toBe("42") // discrete string, not embedded in another arg
      } finally {
        restore()
      }
    })
  })

  describe("classifyError ghCode undefined branch", () => {
    it("uses fallback code when error has no ghCode (e.g. JSON parse failure)", async () => {
      // Return invalid JSON from pr view — JSON.parse throws SyntaxError (no ghCode)
      // classifyError receives it, ghCode is undefined, falls back to fallbackCode
      const { parsed, restore } = await callTool({
        "pr view": { stdout: "not valid json {{{", stderr: "", exitCode: 0 },
      })
      try {
        expect(parsed.error).toBe(true)
        expect(parsed.code).toBe("api_error")
        expect(parsed.retryable).toBe(false)
      } finally {
        restore()
      }
    })
  })

  describe("reviewer manifest duplicate entry (coverage gap in lib/reviewer-manifest)", () => {
    it("rejects manifest with duplicate reviewer entries", () => {
      // Exercises the duplicate-reviewer detection branch (lines 141-144 of reviewer-manifest.ts)
      const duplicateManifest = structuredClone(manifestFixture) as { reviewers: Array<Record<string, unknown>> }
      // Duplicate the first reviewer entry
      duplicateManifest.reviewers.push({ ...duplicateManifest.reviewers[0] })
      const jsonParser = (text: string): unknown => JSON.parse(text) as unknown
      expect(() => parseReviewerManifest(JSON.stringify(duplicateManifest), jsonParser)).toThrow(/duplicate/)
    })

    it("rejects manifest with conflicting reviewer entries", () => {
      // Exercises the conflicting path where same agent has different capability/scopes
      const conflictManifest = structuredClone(manifestFixture) as { reviewers: Array<Record<string, unknown>> }
      const firstReviewer = { ...conflictManifest.reviewers[0] }
      // Change scopes to make it conflicting instead of duplicate
      firstReviewer.scopes = ["documentation"]
      conflictManifest.reviewers.push(firstReviewer)
      const jsonParser = (text: string): unknown => JSON.parse(text) as unknown
      expect(() => parseReviewerManifest(JSON.stringify(conflictManifest), jsonParser)).toThrow(/conflicting/)
    })
  })

  describe("dispatch_status readdir failure (coverage gap)", () => {
    it("returns not_found when dispatch session directory does not exist", async () => {
      const { tool: _tool, restore } = await setupTool()
      try {
        // Inject a correlation ID that maps to a non-existent directory
        const correlationMap = _getSessionCorrelationMap()
        const fakeSessionId = `test-session-${Date.now()}`
        const fakeCorrelationId = "00000000-0000-0000-0000-000000000000"
        correlationMap.set(fakeSessionId, fakeCorrelationId)

        const hooks = await ReviewDispatchPlugin.server({
          client: { session: {} },
          directory: "/tmp",
          worktree: "/tmp",
        } as unknown as PluginInput)

        const result = await hooks.tool.dispatch_status.execute({}, { sessionID: fakeSessionId } as never)
        const parsed = JSON.parse(result.output) as Record<string, unknown>

        expect(parsed.status).toBe("not_found")
        expect(parsed.correlation_id).toBe(fakeCorrelationId)
        expect(parsed.runs).toEqual([])

        // Clean up the map entry
        correlationMap.delete(fakeSessionId)
      } finally {
        restore()
      }
    })
  })
})
