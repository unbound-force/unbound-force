import { Buffer } from "node:buffer"
import { readFileSync } from "node:fs"
import { resolve } from "node:path"
import { fileURLToPath } from "node:url"

import { describe, expect, it } from "vitest"
import { z } from "zod"

import type { ToolResult } from "../lib/uf-workflow-types.js"
import {
  CHECKOUT_PR_HEAD_ERROR_CODES,
  createCheckoutPrHeadTool,
  type CheckoutPrHeadData,
  type GhRunner,
  type GitExec,
} from "../plugins/uf-workflow/index.js"

// ---------------------------------------------------------------------------
// checkout_pr_head unit tests (OpenSpec change checkout-pr-head, issue #965)
//
// Every STOP condition the /uf.address-feedback Section 1.2.1 prose carried
// before the extraction maps to one case below. git and gh are injected
// fakes that answer from a mutable repository model, so a `gh pr checkout`
// call can move HEAD and the branch exactly as the real command would.
// ---------------------------------------------------------------------------

const FAKE_CWD = "/fake/project"
/** The .opencode/ directory, for reading the consuming command copies. */
const OPENCODE_ROOT = fileURLToPath(new URL("../", import.meta.url))
const HEAD_SHA = "a".repeat(40)
const CHILD_SHA = "b".repeat(40)
const OTHER_SHA = "c".repeat(40)
const PR = 42

/** Mutable repository model the fake git executor answers from. */
interface RepoModel {
  head: string
  branch: string
  dirty: string
  /** SHAs that are strict descendants of HEAD_SHA. */
  descendants: Set<string>
  config: Record<string, string>
}

/** Recorded invocations for argument-array assertions. */
interface Calls {
  git: string[][]
  gh: string[][]
}

function repo(overrides: Partial<RepoModel> = {}): RepoModel {
  return {
    head: HEAD_SHA,
    branch: "feature/x",
    dirty: "",
    descendants: new Set([CHILD_SHA]),
    config: {
      "branch.feature/x.remote": "origin",
      "branch.feature/x.merge": "refs/heads/feature/x",
    },
    ...overrides,
  }
}

/** git exits non-zero the way execFileSync surfaces it: a thrown error. */
function gitExit(status: number, stderr = ""): Error {
  return Object.assign(new Error(`Command failed with exit ${status}`), { status, stderr })
}

function fakeGit(model: RepoModel, calls: Calls): GitExec {
  return (args: string[], cwd: string) => {
    expect(cwd).toBe(FAKE_CWD)
    calls.git.push([...args])
    const key = args.join(" ")
    if (key === "rev-parse HEAD") return model.head
    if (key === "branch --show-current") return model.branch
    if (key === "status --porcelain --untracked-files=no") return model.dirty
    if (args[0] === "merge-base" && args[1] === "--is-ancestor" && args[3] === "HEAD") {
      if (args[2] === HEAD_SHA && model.descendants.has(model.head)) return ""
      throw gitExit(1)
    }
    if (args[0] === "config" && args[1] === "--get" && args.length === 3) {
      const value = model.config[args[2] as string]
      if (value === undefined) throw gitExit(1)
      return value
    }
    throw new Error(`unexpected git command: ${key}`)
  }
}

/** gh fake: `onCheckout` mutates the model or throws a checkout failure. */
function fakeGh(calls: Calls, onCheckout: () => void = () => undefined): GhRunner {
  return (args: string[], cwd: string) => {
    expect(cwd).toBe(FAKE_CWD)
    calls.gh.push([...args])
    if (args[0] === "pr" && args[1] === "checkout") {
      onCheckout()
      return ""
    }
    throw new Error(`unexpected gh command: ${args.join(" ")}`)
  }
}

interface ToolArgs {
  pr_number: number
  head_ref: string
  head_sha: string
}

const DEFAULT_ARGS: ToolArgs = { pr_number: PR, head_ref: "feature/x", head_sha: HEAD_SHA }

async function run(
  model: RepoModel,
  options: { onCheckout?: () => void; args?: Partial<ToolArgs>; git?: GitExec } = {},
): Promise<{ result: ToolResult<CheckoutPrHeadData>; calls: Calls }> {
  const calls: Calls = { git: [], gh: [] }
  const tool = createCheckoutPrHeadTool(options.git ?? fakeGit(model, calls), fakeGh(calls, options.onCheckout), FAKE_CWD)
  const output = await (
    tool as unknown as {
      execute: (args: ToolArgs, ctx: { sessionID: string }) => Promise<{ output: string }>
    }
  ).execute({ ...DEFAULT_ARGS, ...options.args }, { sessionID: "test-session" })
  return { result: JSON.parse(output.output) as ToolResult<CheckoutPrHeadData>, calls }
}

function expectError(result: ToolResult<CheckoutPrHeadData>, code: string): { message: string; retryable: boolean } {
  expect(result.status).toBe("error")
  if (result.status !== "error") throw new Error("expected error")
  expect(result.code).toBe(code)
  return { message: result.message, retryable: result.retryable }
}

function expectOk(result: ToolResult<CheckoutPrHeadData>): CheckoutPrHeadData {
  expect(result.status).toBe("ok")
  if (result.status !== "ok" || result.data === undefined) throw new Error("expected ok")
  return result.data
}

describe("checkout_pr_head — identity keyed on HEAD", () => {
  it("continues without a checkout when HEAD equals head_sha", async () => {
    const { result, calls } = await run(repo())
    expect(expectOk(result)).toEqual({
      branch: "feature/x",
      head_remote: "origin",
      merge_ref: "refs/heads/feature/x",
      ahead_of_head_sha: false,
      checked_out: false,
    })
    expect(calls.gh).toEqual([])
    // Equality short-circuits the ancestry probe and the dirty gate.
    expect(calls.git.some((args) => args[0] === "merge-base" || args[0] === "status")).toBe(false)
  })

  it("continues without a checkout when HEAD descends from head_sha (unpushed commits)", async () => {
    const { result, calls } = await run(repo({ head: CHILD_SHA, dirty: " M file.ts" }))
    const data = expectOk(result)
    expect(data.ahead_of_head_sha).toBe(true)
    expect(data.checked_out).toBe(false)
    expect(calls.gh).toEqual([])
    expect(calls.git).toContainEqual(["merge-base", "--is-ancestor", HEAD_SHA, "HEAD"])
  })

  it("reports the actual checked-out branch when it differs from head_ref (fork default branch)", async () => {
    const model = repo({
      head: OTHER_SHA,
      branch: "main",
      config: {},
    })
    const { result, calls } = await run(model, {
      args: { head_ref: "main" },
      onCheckout: () => {
        model.head = HEAD_SHA
        model.branch = "contrib/main"
        model.config = {
          "branch.contrib/main.remote": "https://github.com/contrib/repo.git",
          "branch.contrib/main.merge": "refs/heads/main",
        }
      },
    })
    expect(expectOk(result)).toEqual({
      branch: "contrib/main",
      head_remote: "https://github.com/contrib/repo.git",
      merge_ref: "refs/heads/main",
      ahead_of_head_sha: false,
      checked_out: true,
    })
    expect(calls.gh).toEqual([["pr", "checkout", "42"]])
    expect(calls.git).toContainEqual(["config", "--get", "branch.contrib/main.remote"])
    expect(calls.git).toContainEqual(["config", "--get", "branch.contrib/main.merge"])
  })
})

describe("checkout_pr_head — dirty gate and checkout", () => {
  it("fails DIRTY_TREE before any checkout when tracked files are modified", async () => {
    const { result, calls } = await run(repo({ head: OTHER_SHA, dirty: " M src/a.ts" }))
    const { message, retryable } = expectError(result, "DIRTY_TREE")
    expect(message).toBe(
      "Working tree has uncommitted changes. Commit or stash them, then rerun `/uf.address-feedback 42`.",
    )
    expect(retryable).toBe(false)
    expect(calls.gh).toEqual([])
    expect(calls.git).toContainEqual(["status", "--porcelain", "--untracked-files=no"])
  })

  it("runs gh pr checkout with an argument array and re-verifies HEAD", async () => {
    const model = repo({ head: OTHER_SHA, branch: "main" })
    const { result, calls } = await run(model, {
      onCheckout: () => {
        model.head = CHILD_SHA
        model.branch = "feature/x"
      },
    })
    const data = expectOk(result)
    expect(data.checked_out).toBe(true)
    expect(data.ahead_of_head_sha).toBe(true)
    expect(calls.gh).toEqual([["pr", "checkout", "42"]])
    expect(calls.git.filter((args) => args.join(" ") === "rev-parse HEAD")).toHaveLength(2)
  })

  it("fails CHECKOUT_FAILED with git stash -u remediation on an untracked-file collision", async () => {
    const stderr =
      "error: The following untracked working tree files would be overwritten by checkout:\n\tsrc/new.ts\nAborting"
    const { result } = await run(repo({ head: OTHER_SHA }), {
      onCheckout: () => {
        throw Object.assign(new Error("Command failed: gh pr checkout 42"), { stderr: Buffer.from(stderr) })
      },
    })
    const { message, retryable } = expectError(result, "CHECKOUT_FAILED")
    expect(message).toContain("untracked working tree files would be overwritten")
    expect(message).toContain("`git stash -u` includes untracked files; plain `git stash` skips them")
    expect(message).toContain("src/new.ts")
    expect(retryable).toBe(false)
  })

  it("fails CHECKOUT_FAILED, retryable, carrying the specific error when the fetch fails", async () => {
    const { result } = await run(repo({ head: OTHER_SHA }), {
      onCheckout: () => {
        throw Object.assign(new Error("Command failed"), {
          stderr: "fatal: unable to access 'https://github.com/o/r.git/': Could not resolve host: github.com\n",
        })
      },
    })
    const { message, retryable } = expectError(result, "CHECKOUT_FAILED")
    expect(message).toContain("Could not resolve host: github.com")
    expect(message).toContain("fetch or checkout error")
    expect(retryable).toBe(true)
  })

  it("falls back to the error message when the failure carries no stderr", async () => {
    const { result } = await run(repo({ head: OTHER_SHA }), {
      onCheckout: () => {
        throw new Error("spawn gh ENOENT")
      },
    })
    expect(expectError(result, "CHECKOUT_FAILED").message).toContain("spawn gh ENOENT")
  })

  it("stringifies a non-Error thrown value", async () => {
    const { result } = await run(repo({ head: OTHER_SHA }), {
      onCheckout: () => {
        throw "boom"
      },
    })
    expect(expectError(result, "CHECKOUT_FAILED").message).toContain("boom")
  })

  it("fails UNRELATED_BRANCH with both SHAs when the checkout lands off the PR head", async () => {
    const model = repo({ head: OTHER_SHA })
    const { result, calls } = await run(model, { onCheckout: () => undefined })
    const { message } = expectError(result, "UNRELATED_BRANCH")
    expect(message).toContain(OTHER_SHA)
    expect(message).toContain(HEAD_SHA)
    expect(message).toContain("the branch is unrelated to the PR head")
    // Never resets or discards local commits.
    expect(calls.git.some((args) => args[0] === "reset" || args[0] === "checkout")).toBe(false)
  })
})

describe("checkout_pr_head — push target", () => {
  it("fails NO_PUSH_TARGET on a detached HEAD", async () => {
    const { result, calls } = await run(repo({ branch: "" }))
    expect(expectError(result, "NO_PUSH_TARGET").message).toContain("detached")
    expect(calls.git.some((args) => args[0] === "config")).toBe(false)
  })

  it("fails NO_PUSH_TARGET when the branch remote is unset", async () => {
    const { result } = await run(repo({ config: { "branch.feature/x.merge": "refs/heads/feature/x" } }))
    expect(expectError(result, "NO_PUSH_TARGET").message).toContain("no configured push target")
  })

  it("fails NO_PUSH_TARGET when the branch merge ref is unset", async () => {
    const { result } = await run(repo({ config: { "branch.feature/x.remote": "origin" } }))
    expectError(result, "NO_PUSH_TARGET")
  })

  it("fails READ_ONLY_MERGE_REF for a refs/pull/ merge ref", async () => {
    const { result } = await run(
      repo({
        config: { "branch.feature/x.remote": "https://github.com/o/r.git", "branch.feature/x.merge": "refs/pull/42/head" },
      }),
    )
    const { message, retryable } = expectError(result, "READ_ONLY_MERGE_REF")
    expect(message).toContain("refs/pull/42/head")
    expect(message).toContain("maintainer edits disabled")
    expect(retryable).toBe(false)
  })

  it("fails UNRELATED_BRANCH when the merge ref is not refs/heads/<head_ref> (main absorbed the merged PR)", async () => {
    const { result } = await run(
      repo({
        head: CHILD_SHA,
        branch: "main",
        config: { "branch.main.remote": "origin", "branch.main.merge": "refs/heads/main" },
      }),
    )
    const { message } = expectError(result, "UNRELATED_BRANCH")
    expect(message).toContain("refs/heads/main")
    expect(message).toContain("refs/heads/feature/x")
    expect(message).toContain("Never push this PR's commits to a different branch")
  })
})

describe("checkout_pr_head — refname charset gate", () => {
  it("fails UNSAFE_PUSH_TARGET for a branch name with shell-active characters before reading config", async () => {
    const { result, calls } = await run(repo({ branch: "feat/$(touch pwned)" }))
    expectError(result, "UNSAFE_PUSH_TARGET")
    expect(calls.git.some((args) => args[0] === "config")).toBe(false)
  })

  it("fails UNSAFE_PUSH_TARGET for a merge ref with shell-active characters", async () => {
    const { result } = await run(
      repo({
        branch: "feat",
        config: { "branch.feat.remote": "origin", "branch.feat.merge": "refs/heads/a`b`" },
      }),
      { args: { head_ref: "a`b`" } },
    )
    expectError(result, "UNSAFE_PUSH_TARGET")
  })

  it.each([
    ["an https URL with userinfo", "https://user:token@github.com/o/r.git"],
    ["an ssh URL", "git@github.com:o/r.git"],
    ["the local-path remote .", "."],
    ["the local-path remote ..", ".."],
    ["a remote with a quote", 'origin"x'],
  ])("fails UNSAFE_PUSH_TARGET for %s", async (_label, remote) => {
    const { result } = await run(
      repo({ config: { "branch.feature/x.remote": remote, "branch.feature/x.merge": "refs/heads/feature/x" } }),
    )
    expectError(result, "UNSAFE_PUSH_TARGET")
  })

  it("accepts a plain https URL remote", async () => {
    const { result } = await run(
      repo({
        config: {
          "branch.feature/x.remote": "https://github.com/o/r.git",
          "branch.feature/x.merge": "refs/heads/feature/x",
        },
      }),
    )
    expect(expectOk(result).head_remote).toBe("https://github.com/o/r.git")
  })
})

describe("checkout_pr_head — unexpected failures and schema", () => {
  it("wraps an unexpected git failure as CHECKOUT_PR_HEAD_ERROR", async () => {
    const { result } = await run(repo(), {
      git: () => {
        throw new Error("fatal: not a git repository")
      },
    })
    expect(expectError(result, "CHECKOUT_PR_HEAD_ERROR").message).toContain("not a git repository")
  })

  it("declares the documented argument schema", () => {
    const tool = createCheckoutPrHeadTool(
      () => "",
      () => "",
      FAKE_CWD,
    ) as unknown as { args: z.ZodRawShape }
    expect(Object.keys(tool.args).sort()).toEqual(["head_ref", "head_sha", "pr_number"])
    const schema = z.object(tool.args)
    expect(schema.safeParse(DEFAULT_ARGS).success).toBe(true)
    for (const bad of [
      { ...DEFAULT_ARGS, pr_number: 0 },
      { ...DEFAULT_ARGS, pr_number: 1.5 },
      { ...DEFAULT_ARGS, pr_number: "42" },
      { ...DEFAULT_ARGS, head_ref: "" },
      { ...DEFAULT_ARGS, head_sha: "A".repeat(40) },
      { ...DEFAULT_ARGS, head_sha: "a".repeat(39) },
    ]) {
      expect(schema.safeParse(bad).success, JSON.stringify(bad)).toBe(false)
    }
  })

  it("documents exactly the tool's failure codes in uf.address-feedback.md Section 1.2.1 (both copies)", () => {
    // Structural contract against the tool: every exported code has a
    // handling bullet in the command, and the command handles no code
    // the tool cannot return.
    for (const rel of [
      "commands/uf.address-feedback.md",
      "../internal/scaffold/assets/opencode/commands/uf.address-feedback.md",
    ]) {
      const text = readFileSync(resolve(OPENCODE_ROOT, rel), "utf8")
      const start = text.indexOf("### 1.2.1 Check Out the PR Head")
      const end = text.indexOf("### 1.3 ", start)
      expect(start, rel).toBeGreaterThanOrEqual(0)
      expect(end, rel).toBeGreaterThan(start)
      const section = text.slice(start, end)
      expect(section).toContain("`checkout_pr_head`")
      const documented = [...section.matchAll(/^- `([A-Z_]+)`:/gm)].map((match) => match[1])
      expect(documented, rel).toEqual([...CHECKOUT_PR_HEAD_ERROR_CODES])
    }
  })

  it("exports the documented failure-code set", () => {
    expect([...CHECKOUT_PR_HEAD_ERROR_CODES]).toEqual([
      "DIRTY_TREE",
      "CHECKOUT_FAILED",
      "UNRELATED_BRANCH",
      "READ_ONLY_MERGE_REF",
      "NO_PUSH_TARGET",
      "UNSAFE_PUSH_TARGET",
    ])
  })
})
