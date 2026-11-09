import { execFileSync } from "node:child_process"
import { existsSync, lstatSync, mkdirSync, mkdtempSync, readFileSync, rmSync, statSync, symlinkSync, writeFileSync } from "node:fs"
import { tmpdir } from "node:os"
import { dirname, join } from "node:path"

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { z } from "zod"

import {
  FileFeedbackIssueInputSchema,
  MAX_THREAD_PAGES,
  PostFeedbackRepliesInputSchema,
  RESOLVE_THREAD_MUTATION,
  REVIEW_THREADS_QUERY,
  classifyGhError,
  defaultFeedbackGhExec,
  feedbackStatePath,
  fileFeedbackIssue,
  itemKey,
  parseIssueUrl,
  postFeedbackReplies,
  readFeedbackState,
  renderReplyBody,
  replyEndpoint,
  writeFeedbackState,
  type FeedbackGhExec,
  type FeedbackReplyDeps,
  type FileFeedbackIssueData,
  type PostFeedbackRepliesData,
  type PostFeedbackRepliesInput,
} from "../lib/feedback-replies.js"
import {
  createFileFeedbackIssueTool,
  createPostFeedbackRepliesTool,
} from "../plugins/uf-workflow/index.js"

// ---------------------------------------------------------------------------
// Module-level mock: defaultFeedbackGhExec is the only real-subprocess path.
// ---------------------------------------------------------------------------

vi.mock("node:child_process", async (importOriginal) => {
  const actual = await importOriginal<typeof import("node:child_process")>()
  return { ...actual, execFileSync: vi.fn() }
})

// ---------------------------------------------------------------------------
// Fake gh
// ---------------------------------------------------------------------------

/** One recorded gh call, with the temp-file contents gh would have read. */
interface GhCall {
  readonly args: string[]
  readonly cwd: string
  readonly fileContents: string | null
  readonly filePath: string | null
}

/** A scripted gh response: stdout text, or an error to throw. */
type GhResponse = string | { readonly error: unknown }

/** A route matches one gh invocation and yields its next scripted response. */
interface GhRoute {
  readonly match: (args: string[]) => boolean
  readonly responses: GhResponse[]
}

/**
 * Fake gh executor. Routes are tried in order; each match consumes the
 * route's next response (the last response repeats). The temp file named
 * by --input or --body-file is read during the call — after the call the
 * tool has already removed it.
 */
function fakeGh(routes: GhRoute[]): { gh: FeedbackGhExec; calls: GhCall[] } {
  const calls: GhCall[] = []
  const gh: FeedbackGhExec = (args, cwd) => {
    const flagIndex = args.findIndex((arg) => arg === "--input" || arg === "--body-file")
    const filePath = flagIndex >= 0 ? (args[flagIndex + 1] ?? null) : null
    calls.push({
      args: [...args],
      cwd,
      filePath,
      fileContents: filePath !== null ? readFileSync(filePath, "utf-8") : null,
    })
    if (filePath !== null) {
      // Permissions on the private temp file (0o600) and directory (0o700).
      expect(statSync(filePath).mode & 0o777).toBe(0o600)
      expect(statSync(dirname(filePath)).mode & 0o777).toBe(0o700)
    }
    const route = routes.find((candidate) => candidate.match(args))
    if (route === undefined) {
      throw new Error(`unexpected gh call: ${args.join(" ")}`)
    }
    const response = route.responses.length > 1 ? route.responses.shift() : route.responses[0]
    if (response === undefined) throw new Error("route has no responses")
    if (typeof response === "object") throw response.error
    return response
  }
  return { gh, calls }
}

const REPO = "acme/widgets"
const PR = 42

/** Route helpers. */
const isRepoView = (args: string[]): boolean => args[0] === "repo" && args[1] === "view"
const isIssueCreate = (args: string[]): boolean => args[0] === "issue" && args[1] === "create"
const isPost = (endpoint: string) => (args: string[]): boolean =>
  args[0] === "api" && args[1] === "--method" && args[3] === endpoint
const isThreadQuery = (args: string[]): boolean =>
  args[0] === "api" && args[1] === "graphql" && args.includes(`query=${REVIEW_THREADS_QUERY}`)
const isResolve = (threadId: string) => (args: string[]): boolean =>
  args[0] === "api" && args[1] === "graphql" && args.includes(`threadId=${threadId}`)

/** A posted-comment response. */
function postedComment(id: number): string {
  return JSON.stringify({ id, html_url: `https://github.com/${REPO}/pull/${PR}#discussion_r${id}` })
}

/** One reviewThreads page. */
function threadsPage(
  nodes: Array<{ id: string; isResolved?: boolean; comments: number[] }>,
  next: string | null = null,
): string {
  return JSON.stringify({
    data: {
      repository: {
        pullRequest: {
          reviewThreads: {
            pageInfo: { hasNextPage: next !== null, endCursor: next },
            nodes: nodes.map((node) => ({
              id: node.id,
              isResolved: node.isResolved ?? false,
              comments: { nodes: node.comments.map((databaseId) => ({ databaseId })) },
            })),
          },
        },
      },
    },
  })
}

/** A resolveReviewThread mutation response. */
function resolvedThread(id: string): string {
  return JSON.stringify({ data: { resolveReviewThread: { thread: { id, isResolved: true } } } })
}

/** A gh failure shaped like execFileSync's thrown error. */
function ghError(stderr: string, extra: Record<string, unknown> = {}): Error {
  return Object.assign(new Error("Command failed: gh"), { stderr, status: 1, ...extra })
}

// ---------------------------------------------------------------------------
// Scratch project
// ---------------------------------------------------------------------------

let root: string

beforeEach(() => {
  root = mkdtempSync(join(tmpdir(), "uf-feedback-test-"))
})

afterEach(() => {
  rmSync(root, { recursive: true, force: true })
})

/** Deps for a fake gh, a fixed clock, and a recording checkpoint. */
function deps(
  gh: FeedbackGhExec,
  checkpoints: Array<{ step: string; data: Record<string, unknown> }> = [],
): FeedbackReplyDeps {
  return {
    gh,
    cwd: root,
    now: () => new Date("2026-11-08T12:00:00.000Z"),
    recordCheckpoint: (step, data) => {
      checkpoints.push({ step, data })
    },
  }
}

/** Parse post_feedback_replies input through its schema. */
function input(raw: Record<string, unknown>): PostFeedbackRepliesInput {
  return PostFeedbackRepliesInputSchema.parse({ pr_number: PR, repo: REPO, ...raw })
}

/** The persisted cache document. */
function cache(): Record<string, unknown> {
  return JSON.parse(readFileSync(feedbackStatePath(root, PR), "utf-8")) as Record<string, unknown>
}

// ---------------------------------------------------------------------------
// Pure helpers
// ---------------------------------------------------------------------------

describe("pure helpers", () => {
  it("selects the review-thread replies endpoint or the issue-comments endpoint", () => {
    expect(replyEndpoint(REPO, PR, { comment_kind: "review", comment_id: 7 })).toBe(
      "repos/acme/widgets/pulls/42/comments/7/replies",
    )
    expect(replyEndpoint(REPO, PR, { comment_kind: "issue", comment_id: 7 })).toBe(
      "repos/acme/widgets/issues/42/comments",
    )
  })

  it("prefixes FILE replies with the filed issue and leaves others verbatim", () => {
    expect(renderReplyBody({ decision: "file", body: "tracks the retry gap. Not changed in this PR (LOW)." }, 77)).toBe(
      "Filed as #77: tracks the retry gap. Not changed in this PR (LOW).",
    )
    expect(renderReplyBody({ decision: "file", body: "x" }, null)).toBe("x")
    expect(renderReplyBody({ decision: "reject", body: "Per CS-003." }, 77)).toBe("Per CS-003.")
  })

  it("parses the issue URL gh issue create prints", () => {
    expect(parseIssueUrl("https://github.com/acme/widgets/issues/101\n")).toEqual({
      number: 101,
      url: "https://github.com/acme/widgets/issues/101",
    })
    expect(parseIssueUrl("Creating issue in acme/widgets\n\nhttps://ghe.example/acme/widgets/issues/9")).toEqual({
      number: 9,
      url: "https://ghe.example/acme/widgets/issues/9",
    })
    expect(parseIssueUrl("no url here")).toBeNull()
    expect(parseIssueUrl("https://github.com/acme/widgets/issues/0")).toBeNull()
  })

  it("keys items by comment kind and id", () => {
    expect(itemKey("review", 5)).toBe("review:5")
    expect(itemKey("issue", 5)).toBe("issue:5")
  })

  it.each([
    [Object.assign(new Error("spawn gh ENOENT"), { code: "ENOENT" }), "gh_not_found", false],
    [Object.assign(new Error("spawnSync gh ETIMEDOUT"), { code: "ETIMEDOUT" }), "timeout", true],
    [Object.assign(new Error("killed"), { signal: "SIGTERM" }), "timeout", true],
    [ghError("HTTP 403: API rate limit exceeded"), "rate_limited", true],
    [ghError("To get started with GitHub CLI, please run:  gh auth login"), "gh_auth_failed", false],
    [ghError("", { status: 4 }), "gh_auth_failed", false],
    [ghError("HTTP 422: Validation Failed"), "api_error", false],
    [Object.assign(new Error("x"), { stderr: Buffer.from("HTTP 500 rate limit") }), "rate_limited", true],
    ["plain string", "api_error", false],
  ])("classifies gh failures (%#)", (err, code, retryable) => {
    const classified = classifyGhError(err)
    expect(classified.code).toBe(code)
    expect(classified.retryable).toBe(retryable)
    expect(classified.message.length).toBeGreaterThan(0)
  })
})

// ---------------------------------------------------------------------------
// Input validation
// ---------------------------------------------------------------------------

describe("input validation", () => {
  const item = { comment_kind: "review", comment_id: 1, decision: "accept", body: "Addressed in `abc1234`." }

  it.each([
    ["missing resolve_threads", { pr_number: 1, items: [item] }],
    ["zero pr_number", { pr_number: 0, resolve_threads: true, items: [item] }],
    ["bad repo", { pr_number: 1, repo: "acme", resolve_threads: true, items: [item] }],
    ["repo injection", { pr_number: 1, repo: "acme/w; rm -rf /", resolve_threads: true, items: [item] }],
    ["empty items", { pr_number: 1, resolve_threads: true, items: [] }],
    ["blank body", { pr_number: 1, resolve_threads: true, items: [{ ...item, body: "   " }] }],
    ["file without issue", { pr_number: 1, resolve_threads: true, items: [{ ...item, decision: "file" }] }],
    [
      "issue on a non-file item",
      { pr_number: 1, resolve_threads: true, items: [{ ...item, issue: { title: "t", body: "b" } }] },
    ],
    ["duplicate items", { pr_number: 1, resolve_threads: true, items: [item, item] }],
    ["unknown decision", { pr_number: 1, resolve_threads: true, items: [{ ...item, decision: "skip" }] }],
    ["extra top-level field", { pr_number: 1, resolve_threads: true, items: [item], extra: 1 }],
    ["extra item field", { pr_number: 1, resolve_threads: true, items: [{ ...item, thread_id: "T" }] }],
  ])("rejects %s", (_name, raw) => {
    expect(PostFeedbackRepliesInputSchema.safeParse(raw).success).toBe(false)
  })

  it("accepts the same comment id under different kinds", () => {
    const parsed = PostFeedbackRepliesInputSchema.safeParse({
      pr_number: 1,
      resolve_threads: false,
      items: [item, { ...item, comment_kind: "issue" }],
    })
    expect(parsed.success).toBe(true)
  })

  it("rejects malformed file_feedback_issue input", () => {
    const valid = { pr_number: 1, comment_kind: "review", comment_id: 3, title: "t", body: "b" }
    expect(FileFeedbackIssueInputSchema.safeParse(valid).success).toBe(true)
    expect(FileFeedbackIssueInputSchema.safeParse({ ...valid, title: "" }).success).toBe(false)
    expect(FileFeedbackIssueInputSchema.safeParse({ ...valid, title: "x".repeat(257) }).success).toBe(false)
    expect(FileFeedbackIssueInputSchema.safeParse({ ...valid, comment_id: -1 }).success).toBe(false)
    expect(FileFeedbackIssueInputSchema.safeParse({ ...valid, extra: true }).success).toBe(false)
  })
})

// ---------------------------------------------------------------------------
// post_feedback_replies — happy path
// ---------------------------------------------------------------------------

describe("postFeedbackReplies", () => {
  it("files, posts, and resolves in one call, then records the checkpoint", () => {
    const { gh, calls } = fakeGh([
      { match: isIssueCreate, responses: ["https://github.com/acme/widgets/issues/301\n"] },
      { match: isPost("repos/acme/widgets/pulls/42/comments/11/replies"), responses: [postedComment(1011)] },
      { match: isPost("repos/acme/widgets/pulls/42/comments/12/replies"), responses: [postedComment(1012)] },
      { match: isPost("repos/acme/widgets/pulls/42/comments/13/replies"), responses: [postedComment(1013)] },
      { match: isPost("repos/acme/widgets/issues/42/comments"), responses: [postedComment(1014)] },
      {
        match: isThreadQuery,
        responses: [threadsPage([{ id: "T11", comments: [11, 2011] }, { id: "T12", comments: [12] }, { id: "T13", comments: [13] }])],
      },
      { match: isResolve("T11"), responses: [resolvedThread("T11")] },
      { match: isResolve("T12"), responses: [resolvedThread("T12")] },
    ])
    const checkpoints: Array<{ step: string; data: Record<string, unknown> }> = []

    const data = postFeedbackReplies(
      input({
        resolve_threads: true,
        items: [
          { comment_kind: "review", comment_id: 11, decision: "accept", body: "Addressed in `abc1234`: named the timeout." },
          {
            comment_kind: "review",
            comment_id: 12,
            decision: "file",
            body: "tracks the retry gap. Not changed in this PR (LOW).",
            issue: { title: "Retry gap in proxy", body: "Finding from PR #42 `$(rm -rf /)` \"quoted\"" },
          },
          { comment_kind: "review", comment_id: 13, decision: "reject", body: "Per CS-003, the name stays." },
          { comment_kind: "issue", comment_id: 14, decision: "reply", body: "Acknowledged; out of scope here (MEDIUM)." },
        ],
      }),
      deps(gh, checkpoints),
    )

    expect(data.complete).toBe(true)
    expect(data.counts).toEqual({
      total: 4,
      posted: 4,
      pending: 0,
      failed: 0,
      issues_filed: 1,
      threads_resolved: 2,
      threads_resolvable: 2,
    })
    expect(data.items.map((item) => [item.key, item.reply, item.issue, item.thread])).toEqual([
      ["review:11", "posted", null, "resolved"],
      ["review:12", "posted", "filed", "resolved"],
      ["review:13", "posted", null, "not_applicable"],
      ["issue:14", "posted", null, "not_applicable"],
    ])
    expect(data.items[1]?.issue_number).toBe(301)
    expect(data.items[0]?.thread_id).toBe("T11")
    expect(data.summary).toBe("Posted 4 of 4 reply comments. Resolved 2 of 2 addressed threads.")
    expect(data.checkpoint_recorded).toBe(true)
    expect(checkpoints).toEqual([
      { step: "reply-comments", data: { posted: 4, total: 4, threads_resolved: 2, threads_resolvable: 2 } },
    ])

    // Ordering: the FILE item's issue exists before its reply is posted.
    const order = calls.map((call) => call.args.slice(0, 4).join(" "))
    expect(order.indexOf("issue create --repo acme/widgets")).toBeLessThan(
      order.indexOf("api --method POST repos/acme/widgets/pulls/42/comments/12/replies"),
    )

    // Injection safety: untrusted text never reaches an argument as shell
    // text — bodies travel as temp-file JSON and the title as one argv
    // element.
    const issueCall = calls.find((call) => isIssueCreate(call.args))
    expect(issueCall?.args).toEqual([
      "issue",
      "create",
      "--repo",
      REPO,
      "--title",
      "Retry gap in proxy",
      "--body-file",
      issueCall?.filePath,
    ])
    expect(issueCall?.fileContents).toBe("Finding from PR #42 `$(rm -rf /)` \"quoted\"")
    const fileReply = calls.find((call) => isPost("repos/acme/widgets/pulls/42/comments/12/replies")(call.args))
    expect(JSON.parse(fileReply?.fileContents ?? "{}")).toEqual({
      body: "Filed as #301: tracks the retry gap. Not changed in this PR (LOW).",
    })
    for (const call of calls) {
      expect(call.cwd).toBe(root)
      if (call.filePath !== null) expect(existsSync(dirname(call.filePath))).toBe(false)
    }
  })

  it("resolves the repository through gh repo view when repo is omitted", () => {
    const { gh, calls } = fakeGh([
      { match: isRepoView, responses: ["acme/widgets\n"] },
      { match: isPost("repos/acme/widgets/issues/42/comments"), responses: [postedComment(5)] },
    ])
    const data = postFeedbackReplies(
      PostFeedbackRepliesInputSchema.parse({
        pr_number: PR,
        resolve_threads: true,
        items: [{ comment_kind: "issue", comment_id: 5, decision: "ask", body: "Could you clarify?" }],
      }),
      deps(gh),
    )
    expect(data.repo).toBe(REPO)
    expect(calls[0]?.args).toEqual(["repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner"])
    expect(data.complete).toBe(true)
  })

  it("fails the whole call when the repository cannot be resolved", () => {
    const { gh } = fakeGh([{ match: isRepoView, responses: ["not a repo"] }])
    expect(() =>
      postFeedbackReplies(
        PostFeedbackRepliesInputSchema.parse({
          pr_number: PR,
          resolve_threads: false,
          items: [{ comment_kind: "issue", comment_id: 5, decision: "ask", body: "?" }],
        }),
        deps(gh),
      ),
    ).toThrow(/could not resolve the repository/)

    const { gh: missing } = fakeGh([
      { match: isRepoView, responses: [{ error: Object.assign(new Error("ENOENT"), { code: "ENOENT" }) }] },
    ])
    expect(() =>
      postFeedbackReplies(
        PostFeedbackRepliesInputSchema.parse({
          pr_number: PR,
          resolve_threads: false,
          items: [{ comment_kind: "issue", comment_id: 5, decision: "ask", body: "?" }],
        }),
        deps(missing),
      ),
    ).toThrow(/gh CLI is not installed/)
  })

  it("leaves threads open and still checkpoints when the author declines resolution", () => {
    const { gh, calls } = fakeGh([
      { match: isPost("repos/acme/widgets/pulls/42/comments/11/replies"), responses: [postedComment(1)] },
    ])
    const checkpoints: Array<{ step: string; data: Record<string, unknown> }> = []
    const data = postFeedbackReplies(
      input({ resolve_threads: false, items: [{ comment_kind: "review", comment_id: 11, decision: "accept", body: "Done." }] }),
      deps(gh, checkpoints),
    )
    expect(data.items[0]?.thread).toBe("not_requested")
    expect(data.complete).toBe(true)
    expect(data.summary).toBe("Posted 1 of 1 reply comments. Threads left open.")
    expect(calls.some((call) => isThreadQuery(call.args))).toBe(false)
    expect(checkpoints).toHaveLength(1)
  })

  it("follows reviewThreads pagination and skips already-resolved threads", () => {
    const { gh, calls } = fakeGh([
      { match: isPost("repos/acme/widgets/pulls/42/comments/21/replies"), responses: [postedComment(1)] },
      { match: isPost("repos/acme/widgets/pulls/42/comments/22/replies"), responses: [postedComment(2)] },
      {
        match: isThreadQuery,
        responses: [
          threadsPage([{ id: "T21", comments: [21] }], "CURSOR1"),
          threadsPage([{ id: "T22", isResolved: true, comments: [22] }]),
        ],
      },
      { match: isResolve("T21"), responses: [resolvedThread("T21")] },
    ])
    const data = postFeedbackReplies(
      input({
        resolve_threads: true,
        items: [
          { comment_kind: "review", comment_id: 21, decision: "modify", body: "Addressed in `abc`." },
          { comment_kind: "review", comment_id: 22, decision: "reply", body: "Noted (LOW)." },
        ],
      }),
      deps(gh),
    )
    expect(data.items.map((item) => item.thread)).toEqual(["resolved", "already_resolved"])
    const queries = calls.filter((call) => isThreadQuery(call.args))
    expect(queries).toHaveLength(2)
    expect(queries[0]?.args).toEqual([
      "api",
      "graphql",
      "-f",
      `query=${REVIEW_THREADS_QUERY}`,
      "-f",
      "owner=acme",
      "-f",
      "name=widgets",
      "-F",
      "number=42",
    ])
    expect(queries[1]?.args.slice(-2)).toEqual(["-f", "cursor=CURSOR1"])
    const resolve = calls.find((call) => isResolve("T21")(call.args))
    expect(resolve?.args).toEqual(["api", "graphql", "-f", `query=${RESOLVE_THREAD_MUTATION}`, "-f", "threadId=T21"])
    expect(calls.some((call) => isResolve("T22")(call.args))).toBe(false)
  })
})

// ---------------------------------------------------------------------------
// post_feedback_replies — failures, ordering, and crash recovery
// ---------------------------------------------------------------------------

describe("postFeedbackReplies failure handling", () => {
  const threeReplies = [
    { comment_kind: "review", comment_id: 31, decision: "accept", body: "One." },
    { comment_kind: "review", comment_id: 32, decision: "accept", body: "Two." },
    { comment_kind: "review", comment_id: 33, decision: "accept", body: "Three." },
  ]

  it("post_feedback_replies round-trip: stops posting at the first failed post and resumes from the cache on the re-run", () => {
    const first = fakeGh([
      { match: isPost("repos/acme/widgets/pulls/42/comments/31/replies"), responses: [postedComment(1)] },
      {
        match: isPost("repos/acme/widgets/pulls/42/comments/32/replies"),
        responses: [{ error: ghError("HTTP 403: API rate limit exceeded") }],
      },
      { match: isThreadQuery, responses: [threadsPage([{ id: "T31", comments: [31] }])] },
      { match: isResolve("T31"), responses: [resolvedThread("T31")] },
    ])
    const checkpoints: Array<{ step: string; data: Record<string, unknown> }> = []
    const partial = postFeedbackReplies(input({ resolve_threads: true, items: threeReplies }), deps(first.gh, checkpoints))

    expect(partial.complete).toBe(false)
    expect(partial.items.map((item) => [item.reply, item.thread])).toEqual([
      ["posted", "resolved"],
      ["post_failed", "reply_not_posted"],
      ["pending", "reply_not_posted"],
    ])
    expect(partial.items[1]?.error).toMatchObject({ code: "rate_limited", retryable: true })
    expect(partial.counts).toMatchObject({ posted: 1, pending: 1, failed: 1, threads_resolved: 1, threads_resolvable: 3 })
    expect(partial.summary).toBe(
      "Posted 1 of 3 reply comments. Items 2, 3 not posted. Resolved 1 of 3 addressed threads. Re-run `/uf.address-feedback 42` to retry.",
    )
    expect(partial.checkpoint_recorded).toBe(false)
    expect(checkpoints).toEqual([])
    // Item 3 was never attempted (failure-stop).
    expect(first.calls.some((call) => isPost("repos/acme/widgets/pulls/42/comments/33/replies")(call.args))).toBe(false)

    // Write→read round-trip: the re-run reads the cache, skips item 1,
    // and posts and resolves the rest.
    const second = fakeGh([
      { match: isPost("repos/acme/widgets/pulls/42/comments/32/replies"), responses: [postedComment(2)] },
      { match: isPost("repos/acme/widgets/pulls/42/comments/33/replies"), responses: [postedComment(3)] },
      {
        match: isThreadQuery,
        responses: [threadsPage([{ id: "T31", isResolved: true, comments: [31] }, { id: "T32", comments: [32] }, { id: "T33", comments: [33] }])],
      },
      { match: isResolve("T32"), responses: [resolvedThread("T32")] },
      { match: isResolve("T33"), responses: [resolvedThread("T33")] },
    ])
    const resumed = postFeedbackReplies(input({ resolve_threads: true, items: threeReplies }), deps(second.gh, checkpoints))
    expect(resumed.items.map((item) => [item.reply, item.thread])).toEqual([
      ["already_posted", "already_resolved"],
      ["posted", "resolved"],
      ["posted", "resolved"],
    ])
    expect(resumed.complete).toBe(true)
    expect(second.calls.some((call) => isPost("repos/acme/widgets/pulls/42/comments/31/replies")(call.args))).toBe(false)
    expect(checkpoints).toHaveLength(1)
  })

  it("resolves threads on a re-run whose replies were all posted earlier (the old 4.5/4.6 gap)", () => {
    const first = fakeGh([{ match: isPost("repos/acme/widgets/pulls/42/comments/31/replies"), responses: [postedComment(1)] }])
    postFeedbackReplies(input({ resolve_threads: false, items: [threeReplies[0]] }), deps(first.gh))

    const second = fakeGh([
      { match: isThreadQuery, responses: [threadsPage([{ id: "T31", comments: [31] }])] },
      { match: isResolve("T31"), responses: [resolvedThread("T31")] },
    ])
    const data = postFeedbackReplies(input({ resolve_threads: true, items: [threeReplies[0]] }), deps(second.gh))
    expect(data.items[0]).toMatchObject({ reply: "already_posted", thread: "resolved", thread_id: "T31" })
    expect(second.calls.filter((call) => call.args[1] === "--method")).toHaveLength(0)
  })

  it("skips a FILE item whose issue fails twice and keeps posting the others", () => {
    const { gh, calls } = fakeGh([
      { match: isIssueCreate, responses: [{ error: ghError("HTTP 502") }, { error: ghError("HTTP 502") }] },
      { match: isPost("repos/acme/widgets/pulls/42/comments/42/replies"), responses: [postedComment(2)] },
      { match: isThreadQuery, responses: [threadsPage([{ id: "T41", comments: [41] }, { id: "T42", comments: [42] }])] },
      { match: isResolve("T42"), responses: [resolvedThread("T42")] },
    ])
    const data = postFeedbackReplies(
      input({
        resolve_threads: true,
        items: [
          { comment_kind: "review", comment_id: 41, decision: "file", body: "tracks x.", issue: { title: "x", body: "y" } },
          { comment_kind: "review", comment_id: 42, decision: "reply", body: "Noted." },
        ],
      }),
      deps(gh),
    )
    expect(calls.filter((call) => isIssueCreate(call.args))).toHaveLength(2)
    expect(data.items[0]).toMatchObject({ reply: "issue_failed", issue: "failed", thread: "reply_not_posted" })
    expect(data.items[0]?.error?.message).toMatch(/failed twice/)
    expect(data.items[1]).toMatchObject({ reply: "posted", thread: "resolved" })
    expect(calls.some((call) => isPost("repos/acme/widgets/pulls/42/comments/41/replies")(call.args))).toBe(false)
    expect(data.complete).toBe(false)
  })

  it("files on the retry when the first gh issue create fails", () => {
    const { gh } = fakeGh([
      { match: isIssueCreate, responses: [{ error: ghError("HTTP 502") }, "https://github.com/acme/widgets/issues/9\n"] },
      { match: isPost("repos/acme/widgets/pulls/42/comments/41/replies"), responses: [postedComment(2)] },
    ])
    const data = postFeedbackReplies(
      input({
        resolve_threads: false,
        items: [{ comment_kind: "review", comment_id: 41, decision: "file", body: "tracks x.", issue: { title: "x", body: "y" } }],
      }),
      deps(gh),
    )
    expect(data.items[0]).toMatchObject({ reply: "posted", issue: "filed", issue_number: 9 })
  })

  it("does not retry gh issue create when it printed no URL", () => {
    const { gh, calls } = fakeGh([{ match: isIssueCreate, responses: ["something odd"] }])
    const data = postFeedbackReplies(
      input({
        resolve_threads: false,
        items: [{ comment_kind: "review", comment_id: 41, decision: "file", body: "b", issue: { title: "x", body: "y" } }],
      }),
      deps(gh),
    )
    expect(calls.filter((call) => isIssueCreate(call.args))).toHaveLength(1)
    expect(data.items[0]?.error?.message).toMatch(/printed no issue URL/)
  })

  it("never files twice: a crash after filing reuses the cached issue", () => {
    const fileItem = {
      comment_kind: "review",
      comment_id: 51,
      decision: "file",
      body: "tracks y.",
      issue: { title: "t", body: "b" },
    }
    const first = fakeGh([
      { match: isIssueCreate, responses: ["https://github.com/acme/widgets/issues/77\n"] },
      { match: isPost("repos/acme/widgets/pulls/42/comments/51/replies"), responses: [{ error: ghError("HTTP 500") }] },
    ])
    const partial = postFeedbackReplies(input({ resolve_threads: false, items: [fileItem] }), deps(first.gh))
    expect(partial.items[0]).toMatchObject({ issue: "filed", reply: "post_failed" })

    const second = fakeGh([{ match: isPost("repos/acme/widgets/pulls/42/comments/51/replies"), responses: [postedComment(5)] }])
    const resumed = postFeedbackReplies(input({ resolve_threads: false, items: [fileItem] }), deps(second.gh))
    expect(resumed.items[0]).toMatchObject({ issue: "deduplicated", issue_number: 77, reply: "posted" })
    expect(second.calls.some((call) => isIssueCreate(call.args))).toBe(false)
    expect(JSON.parse(second.calls[0]?.fileContents ?? "{}")).toEqual({ body: "Filed as #77: tracks y." })
  })

  it("reports a missing thread and a lookup failure per item without stopping", () => {
    const missing = fakeGh([
      { match: isPost("repos/acme/widgets/pulls/42/comments/61/replies"), responses: [postedComment(1)] },
      { match: isThreadQuery, responses: [threadsPage([])] },
    ])
    const notFound = postFeedbackReplies(
      input({ resolve_threads: true, items: [{ comment_kind: "review", comment_id: 61, decision: "accept", body: "x" }] }),
      deps(missing.gh),
    )
    expect(notFound.items[0]).toMatchObject({ reply: "posted", thread: "thread_not_found" })
    expect(notFound.items[0]?.error?.code).toBe("thread_not_found")
    expect(notFound.complete).toBe(false)

    const lookup = fakeGh([{ match: isThreadQuery, responses: [{ error: ghError("HTTP 502 Bad Gateway") }] }])
    const failed = postFeedbackReplies(
      input({ resolve_threads: true, items: [{ comment_kind: "review", comment_id: 61, decision: "accept", body: "x" }] }),
      deps(lookup.gh),
    )
    expect(failed.items[0]).toMatchObject({ reply: "already_posted", thread: "failed" })
    expect(failed.items[0]?.error?.message).toMatch(/review thread lookup failed/)
  })

  it("fails closed on a malformed reviewThreads payload", () => {
    for (const payload of ["not json", JSON.stringify({ data: {} }), JSON.stringify({ data: { repository: { pullRequest: { reviewThreads: { nodes: [{ id: 1 }], pageInfo: {} } } } } })]) {
      rmSync(join(root, ".uf"), { recursive: true, force: true })
      const { gh } = fakeGh([
        { match: isPost("repos/acme/widgets/pulls/42/comments/61/replies"), responses: [postedComment(1)] },
        { match: isThreadQuery, responses: [payload] },
      ])
      const data = postFeedbackReplies(
        input({ resolve_threads: true, items: [{ comment_kind: "review", comment_id: 61, decision: "accept", body: "x" }] }),
        deps(gh),
      )
      expect(data.items[0]?.thread).toBe("failed")
      expect(data.items[0]?.error?.message).toMatch(/unexpected payload/)
    }
  })

  it("stops paging at the page bound", () => {
    const { gh, calls } = fakeGh([
      { match: isPost("repos/acme/widgets/pulls/42/comments/61/replies"), responses: [postedComment(1)] },
      { match: isThreadQuery, responses: [threadsPage([{ id: "T", comments: [1] }], "NEXT")] },
    ])
    const data = postFeedbackReplies(
      input({ resolve_threads: true, items: [{ comment_kind: "review", comment_id: 61, decision: "accept", body: "x" }] }),
      deps(gh),
    )
    expect(calls.filter((call) => isThreadQuery(call.args))).toHaveLength(MAX_THREAD_PAGES)
    expect(data.items[0]?.error?.message).toMatch(/exceeded/)
  })

  it("retries a retryable resolve failure once and reports a non-retryable one", () => {
    const retry = fakeGh([
      { match: isPost("repos/acme/widgets/pulls/42/comments/71/replies"), responses: [postedComment(1)] },
      { match: isPost("repos/acme/widgets/pulls/42/comments/72/replies"), responses: [postedComment(2)] },
      { match: isThreadQuery, responses: [threadsPage([{ id: "T71", comments: [71] }, { id: "T72", comments: [72] }])] },
      { match: isResolve("T71"), responses: [{ error: ghError("timed out") }, resolvedThread("T71")] },
      { match: isResolve("T72"), responses: [{ error: ghError("HTTP 403 Resource not accessible") }] },
    ])
    const data = postFeedbackReplies(
      input({
        resolve_threads: true,
        items: [
          { comment_kind: "review", comment_id: 71, decision: "accept", body: "x" },
          { comment_kind: "review", comment_id: 72, decision: "accept", body: "y" },
        ],
      }),
      deps(retry.gh),
    )
    expect(data.items.map((item) => item.thread)).toEqual(["resolved", "failed"])
    expect(retry.calls.filter((call) => isResolve("T71")(call.args))).toHaveLength(2)
    expect(retry.calls.filter((call) => isResolve("T72")(call.args))).toHaveLength(1)
    expect(data.items[1]?.error?.code).toBe("api_error")
  })

  it("treats a mutation that does not report the thread resolved as a failure", () => {
    const { gh } = fakeGh([
      { match: isPost("repos/acme/widgets/pulls/42/comments/71/replies"), responses: [postedComment(1)] },
      { match: isThreadQuery, responses: [threadsPage([{ id: "T71", comments: [71] }])] },
      { match: isResolve("T71"), responses: ["{}"] },
    ])
    const data = postFeedbackReplies(
      input({ resolve_threads: true, items: [{ comment_kind: "review", comment_id: 71, decision: "accept", body: "x" }] }),
      deps(gh),
    )
    expect(data.items[0]?.thread).toBe("failed")
    expect(data.items[0]?.error?.message).toMatch(/did not report the thread resolved/)
  })

  it("keeps a posted reply whose response body is not JSON", () => {
    const { gh } = fakeGh([{ match: isPost("repos/acme/widgets/issues/42/comments"), responses: ["ok"] }])
    const data = postFeedbackReplies(
      input({ resolve_threads: false, items: [{ comment_kind: "issue", comment_id: 1, decision: "ask", body: "?" }] }),
      deps(gh),
    )
    expect(data.items[0]).toMatchObject({ reply: "posted", reply_id: null, reply_url: null })
  })

  it("reports a checkpoint failure as a warning, not a failed call", () => {
    const { gh } = fakeGh([{ match: isPost("repos/acme/widgets/issues/42/comments"), responses: [postedComment(1)] }])
    const data = postFeedbackReplies(
      input({ resolve_threads: false, items: [{ comment_kind: "issue", comment_id: 1, decision: "ask", body: "?" }] }),
      {
        gh,
        cwd: root,
        recordCheckpoint: () => {
          throw new Error("Detached HEAD state")
        },
      },
    )
    expect(data.complete).toBe(true)
    expect(data.checkpoint_recorded).toBe(false)
    expect(data.warnings).toEqual(["could not record the reply-comments checkpoint: Detached HEAD state"])
  })
})

// ---------------------------------------------------------------------------
// Feedback cache
// ---------------------------------------------------------------------------

describe("feedback cache", () => {
  it("preserves the command's assessment keys and writes 0o600 under 0o700", () => {
    const path = feedbackStatePath(root, PR)
    mkdirSync(dirname(path), { recursive: true, mode: 0o700 })
    writeFileSync(path, JSON.stringify({ last_fetched: "2026-11-01T00:00:00Z", threads: { a: 1 } }), { mode: 0o600 })
    const { gh } = fakeGh([{ match: isPost("repos/acme/widgets/issues/42/comments"), responses: [postedComment(9)] }])
    postFeedbackReplies(
      input({ resolve_threads: false, items: [{ comment_kind: "issue", comment_id: 1, decision: "ask", body: "?" }] }),
      deps(gh),
    )
    const document = cache()
    expect(document.last_fetched).toBe("2026-11-01T00:00:00Z")
    expect(document.threads).toEqual({ a: 1 })
    expect(document.replies).toEqual({
      "issue:1": {
        comment_kind: "issue",
        comment_id: 1,
        decision: "ask",
        issue_number: null,
        issue_url: null,
        comment_posted: true,
        reply_id: 9,
        reply_url: "https://github.com/acme/widgets/pull/42#discussion_r9",
        thread_id: null,
        thread_resolved: false,
        updated_at: "2026-11-08T12:00:00.000Z",
      },
    })
    expect(statSync(path).mode & 0o777).toBe(0o600)
  })

  it("creates the cache directory 0o700 and round-trips state", () => {
    writeFeedbackState(root, 7, { document: { keep: true }, replies: {} })
    const path = feedbackStatePath(root, 7)
    expect(statSync(dirname(path)).mode & 0o777).toBe(0o700)
    expect(readFeedbackState(root, 7)).toEqual({ document: { keep: true, replies: {} }, replies: {} })
  })

  it.each([
    ["invalid JSON", "{not json", /not valid JSON/],
    ["a non-object", "[1]", /not a JSON object/],
    ["a non-object replies entry", JSON.stringify({ replies: [] }), /entry is not an object/],
  ])("fails closed on %s", (_name, contents, message) => {
    const path = feedbackStatePath(root, PR)
    mkdirSync(dirname(path), { recursive: true })
    writeFileSync(path, contents)
    expect(() => readFeedbackState(root, PR)).toThrow(message)
  })

  it("refuses symlinked cache files and directories", () => {
    const path = feedbackStatePath(root, PR)
    mkdirSync(dirname(path), { recursive: true })
    const target = join(root, "elsewhere.json")
    writeFileSync(target, "{}")
    symlinkSync(target, path)
    expect(lstatSync(path).isSymbolicLink()).toBe(true)
    expect(() => readFeedbackState(root, PR)).toThrow(/symlinked feedback cache/)
    expect(() => writeFeedbackState(root, PR, { document: {}, replies: {} })).toThrow(/symlinked feedback cache/)

    const realDir = join(root, "real-dir")
    mkdirSync(realDir)
    mkdirSync(join(root, ".uf", "feedback"), { recursive: true })
    symlinkSync(realDir, join(root, ".uf", "feedback", "pr-8"))
    expect(() => writeFeedbackState(root, 8, { document: {}, replies: {} })).toThrow(/symlinked feedback cache directory/)
  })

  it("surfaces unreadable caches as state_error", () => {
    const path = feedbackStatePath(root, PR)
    mkdirSync(path, { recursive: true }) // a directory where the file should be
    expect(() => readFeedbackState(root, PR)).toThrow(/could not read the feedback cache/)
    expect(() => writeFeedbackState(root, PR, { document: {}, replies: {} })).toThrow(/could not write the feedback cache/)
  })
})

// ---------------------------------------------------------------------------
// file_feedback_issue
// ---------------------------------------------------------------------------

describe("fileFeedbackIssue", () => {
  const issueInput = FileFeedbackIssueInputSchema.parse({
    pr_number: PR,
    repo: REPO,
    comment_kind: "review",
    comment_id: 81,
    title: "Name the timeout",
    body: "LOW finding from PR #42",
  })

  it("files once, records the issue, and deduplicates the next call", () => {
    const first = fakeGh([{ match: isIssueCreate, responses: ["https://github.com/acme/widgets/issues/500\n"] }])
    const filed = fileFeedbackIssue(issueInput, deps(first.gh))
    expect(filed).toEqual({
      pr_number: PR,
      repo: REPO,
      key: "review:81",
      issue_number: 500,
      issue_url: "https://github.com/acme/widgets/issues/500",
      deduplicated: false,
      attempts: 1,
    })
    expect(first.calls[0]?.fileContents).toBe("LOW finding from PR #42")

    const second = fakeGh([])
    const again = fileFeedbackIssue(issueInput, deps(second.gh))
    expect(again).toMatchObject({ issue_number: 500, deduplicated: true, attempts: 0 })
    expect(second.calls).toHaveLength(0)
  })

  it("file_feedback_issue round-trip: post_feedback_replies reads the filed issue back from the cache", () => {
    const filing = fakeGh([{ match: isIssueCreate, responses: ["https://github.com/acme/widgets/issues/501\n"] }])
    fileFeedbackIssue(issueInput, deps(filing.gh))
    const posting = fakeGh([{ match: isPost("repos/acme/widgets/pulls/42/comments/81/replies"), responses: [postedComment(1)] }])
    const data = postFeedbackReplies(
      input({
        resolve_threads: false,
        items: [{ comment_kind: "review", comment_id: 81, decision: "file", body: "tracks it.", issue: { title: "t", body: "b" } }],
      }),
      deps(posting.gh),
    )
    expect(data.items[0]).toMatchObject({ issue: "deduplicated", issue_number: 501, reply: "posted" })
    expect(posting.calls.some((call) => isIssueCreate(call.args))).toBe(false)
  })

  it("retries once and reports the second failure", () => {
    const { gh, calls } = fakeGh([{ match: isIssueCreate, responses: [{ error: ghError("HTTP 403: API rate limit exceeded") }] }])
    expect(() => fileFeedbackIssue(issueInput, deps(gh))).toThrow(/failed twice/)
    expect(calls).toHaveLength(2)
    expect(existsSync(feedbackStatePath(root, PR))).toBe(false)
  })
})

// ---------------------------------------------------------------------------
// Tool boundary (envelopes)
// ---------------------------------------------------------------------------

/** Execute a registered tool and parse its output. */
async function run(toolDef: unknown, args: Record<string, unknown>): Promise<Record<string, unknown>> {
  const result = await (
    toolDef as { execute: (args: Record<string, unknown>, ctx: unknown) => Promise<{ output: string }> }
  ).execute(args, {})
  return JSON.parse(result.output) as Record<string, unknown>
}

describe("tool envelopes", () => {
  it("exposes the validation schema shapes as the tool args (input-schema parity)", () => {
    const post = createPostFeedbackRepliesTool(deps(fakeGh([]).gh)) as unknown as { args: z.ZodRawShape }
    const file = createFileFeedbackIssueTool(deps(fakeGh([]).gh)) as unknown as { args: z.ZodRawShape }
    expect(Object.keys(post.args).sort()).toEqual(Object.keys(PostFeedbackRepliesInputSchema.shape).sort())
    expect(Object.keys(post.args).sort()).toEqual(["items", "pr_number", "repo", "resolve_threads"])
    expect(Object.keys(file.args).sort()).toEqual(["body", "comment_id", "comment_kind", "pr_number", "repo", "title"])
    expect(post.args.resolve_threads?.safeParse(undefined).success).toBe(false)
    expect(post.args.repo?.safeParse(undefined).success).toBe(true)
  })

  it("returns invalid_input with the fetch-tool error contract", async () => {
    const result = await run(createPostFeedbackRepliesTool(deps(fakeGh([]).gh)), { pr_number: 1, items: [] })
    expect(result).toMatchObject({ status: "error", error: true, code: "invalid_input", retryable: false })
    expect(String(result.message)).toMatch(/resolve_threads/)
    const fileResult = await run(createFileFeedbackIssueTool(deps(fakeGh([]).gh)), { pr_number: 1 })
    expect(fileResult).toMatchObject({ status: "error", error: true, code: "invalid_input" })
  })

  it("returns the operation error code and retryability", async () => {
    const { gh } = fakeGh([{ match: isIssueCreate, responses: [{ error: ghError("API rate limit exceeded") }] }])
    const result = await run(createFileFeedbackIssueTool(deps(gh)), {
      pr_number: PR,
      repo: REPO,
      comment_kind: "issue",
      comment_id: 1,
      title: "t",
      body: "b",
    })
    expect(result).toMatchObject({ status: "error", error: true, code: "rate_limited", retryable: true })

    const path = feedbackStatePath(root, PR)
    mkdirSync(dirname(path), { recursive: true })
    writeFileSync(path, "{bad")
    const stateResult = await run(createPostFeedbackRepliesTool(deps(fakeGh([]).gh)), {
      pr_number: PR,
      repo: REPO,
      resolve_threads: false,
      items: [{ comment_kind: "issue", comment_id: 1, decision: "ask", body: "?" }],
    })
    expect(stateResult).toMatchObject({ status: "error", error: true, code: "state_error" })
  })

  it("maps unexpected throws to internal_error", async () => {
    const result = await run(
      createPostFeedbackRepliesTool({
        gh: fakeGh([]).gh,
        get cwd(): string {
          throw new Error("boom")
        },
      } as FeedbackReplyDeps),
      { pr_number: PR, repo: REPO, resolve_threads: false, items: [{ comment_kind: "issue", comment_id: 1, decision: "ask", body: "?" }] },
    )
    expect(result).toMatchObject({ status: "error", error: true, code: "internal_error", message: "boom" })
  })

  it("returns the documented success shapes", async () => {
    const { gh } = fakeGh([
      { match: isIssueCreate, responses: ["https://github.com/acme/widgets/issues/600\n"] },
      { match: isPost("repos/acme/widgets/issues/42/comments"), responses: [postedComment(3)] },
    ])
    const filed = (await run(createFileFeedbackIssueTool(deps(gh)), {
      pr_number: PR,
      repo: REPO,
      comment_kind: "issue",
      comment_id: 2,
      title: "t",
      body: "b",
    })) as { status: string; data: FileFeedbackIssueData }
    expect(filed.status).toBe("ok")
    expect(Object.keys(filed.data).sort()).toEqual([
      "attempts",
      "deduplicated",
      "issue_number",
      "issue_url",
      "key",
      "pr_number",
      "repo",
    ])
    const posted = (await run(createPostFeedbackRepliesTool(deps(gh)), {
      pr_number: PR,
      repo: REPO,
      resolve_threads: false,
      items: [{ comment_kind: "issue", comment_id: 3, decision: "reply", body: "Noted." }],
    })) as { status: string; data: PostFeedbackRepliesData }
    expect(posted.status).toBe("ok")
    expect(Object.keys(posted.data).sort()).toEqual([
      "checkpoint_recorded",
      "complete",
      "counts",
      "items",
      "pr_number",
      "repo",
      "summary",
      "warnings",
    ])
  })
})

// ---------------------------------------------------------------------------
// Default executor
// ---------------------------------------------------------------------------

describe("defaultFeedbackGhExec", () => {
  it("runs gh with an argument array, the repository cwd, and a timeout", () => {
    vi.mocked(execFileSync).mockReturnValue("out" as never)
    expect(defaultFeedbackGhExec(["api", "x"], "/repo")).toBe("out")
    expect(vi.mocked(execFileSync)).toHaveBeenCalledWith("gh", ["api", "x"], {
      cwd: "/repo",
      encoding: "utf-8",
      stdio: ["pipe", "pipe", "pipe"],
      timeout: 30_000,
    })
  })
})
