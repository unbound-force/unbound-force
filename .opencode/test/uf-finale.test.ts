import { describe, it, expect, vi, beforeEach } from "vitest"
import { execSync } from "node:child_process"
import {
  readFileSync,
  existsSync,
  lstatSync,
  mkdirSync,
  mkdtempSync,
  writeFileSync,
  rmSync,
} from "node:fs"
import { resolve } from "node:path"

import type { PluginInput } from "@opencode-ai/plugin"
import type { ToolResult } from "../lib/uf-workflow-types.js"
import {
  createUfFinaleStartupTool,
  createUfFinaleStageTool,
  createUfFinaleCommitContextTool,
  createUfFinaleBaseStateTool,
  createUfFinalePushStateTool,
  createUfFinalePrContextTool,
  createUfFinaleCreatePrTool,
  createUfFinaleCiPollTool,
  createUfFinaleReturnToMainTool,
  type GitExec,
} from "../plugins/uf-finale/index.js"
import UfFinalePlugin from "../plugins/uf-finale/index.js"
import {
  createCheckpointTool,
  createUnleashRouteDecisionTool,
  createWriteUnleashSummaryTool,
  type ProgressState,
  type UnleashSummary,
} from "../plugins/uf-workflow/index.js"
import {
  createMockSpawn,
  mockProcess,
  patchBunSpawn,
  type MockResponse,
  type MockSpawnProcess,
  type MockSpawnResponses,
} from "./helpers/gh-spawn-mock.js"

// ---------------------------------------------------------------------------
// Module-level mocks (hoisted by vitest)
// ---------------------------------------------------------------------------

vi.mock("node:child_process", () => ({
  execSync: vi.fn(),
}))

vi.mock("node:fs", async (importOriginal) => {
  const actual = await importOriginal<typeof import("node:fs")>()
  return {
    ...actual,
    readFileSync: vi.fn(),
    writeFileSync: vi.fn(),
    mkdtempSync: vi.fn(),
    rmSync: vi.fn(),
    existsSync: vi.fn(),
    lstatSync: vi.fn(),
    mkdirSync: vi.fn(),
  }
})

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

/** Invoke a tool factory result with typed args and parse JSON output. */
async function invokeTool<T = unknown>(
  toolInstance: unknown,
  args: Record<string, unknown>,
): Promise<ToolResult<T>> {
  const result = await (
    toolInstance as {
      execute: (
        args: Record<string, unknown>,
        ctx: { sessionID: string },
      ) => Promise<{ output: string }>
    }
  ).execute(args, { sessionID: "test-session" })
  return JSON.parse(result.output) as ToolResult<T>
}

/**
 * Install a routed Bun.spawn mock for the duration of `run` and restore the
 * original global afterwards. The uf-finale poller reaches `gh pr checks`
 * through Bun.spawn (FCP-FR-001), so its fixtures live here rather than on
 * execSync — the merge-conflict pre-check still uses execSync.
 */
async function withGhSpawn<T>(
  responses: MockSpawnResponses,
  run: (spawn: ReturnType<typeof createMockSpawn>) => Promise<T>,
): Promise<T> {
  const spawn = createMockSpawn(responses)
  const restore = patchBunSpawn(spawn)
  try {
    return await run(spawn)
  } finally {
    restore()
  }
}

/** Install a spawn mock that serves canned responses in order, repeating the last. */
async function withSequentialSpawn<T>(
  responses: readonly MockResponse[],
  run: () => Promise<T>,
): Promise<T> {
  let call = 0
  const spawn = vi.fn((): MockSpawnProcess =>
    mockProcess(responses[Math.min(call++, responses.length - 1)]),
  )
  const restore = patchBunSpawn(spawn)
  try {
    return await run()
  } finally {
    restore()
  }
}

beforeEach(() => {
  vi.resetAllMocks()
})

// ---------------------------------------------------------------------------
// Plugin smoke test
// ---------------------------------------------------------------------------

describe("UfFinalePlugin", () => {
  it("exports a PluginModule with id and server", () => {
    expect(UfFinalePlugin.id).toBe("uf-finale")
    expect(typeof UfFinalePlugin.server).toBe("function")
  })

  it("server registers all 9 finale tools", async () => {
    const fakeInput = {
      directory: "/tmp/fake",
      worktree: "/tmp/fake",
      client: {} as PluginInput["client"],
    } as PluginInput

    const result = await UfFinalePlugin.server(fakeInput)
    expect(result).toBeDefined()
    expect(result.tool).toBeDefined()
    expect(Object.keys(result.tool).sort()).toEqual([
      "uf_finale_base_state",
      "uf_finale_ci_poll",
      "uf_finale_commit_context",
      "uf_finale_create_pr",
      "uf_finale_pr_context",
      "uf_finale_push_state",
      "uf_finale_return_to_main",
      "uf_finale_stage",
      "uf_finale_startup",
    ])
  })
})

// ---------------------------------------------------------------------------
// uf_finale_startup tests
// ---------------------------------------------------------------------------

describe("uf_finale_startup", () => {
  it("returns branch context for a feature branch with no changes", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "feature/my-branch"
      if (cmd.includes("--porcelain")) return ""
      if (cmd.includes("@{u}..HEAD")) throw new Error("no upstream")
      return ""
    }
    const result = await invokeTool<{
      branch: string
      isMain: boolean
      branchType: string
      changeName: string
      hasChanges: boolean
      secretCandidates: string[]
      unpushedCommits: Array<{ sha: string; subject: string }>
    }>(createUfFinaleStartupTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.branch).toBe("feature/my-branch")
      expect(result.data?.isMain).toBe(false)
      expect(result.data?.branchType).toBe("other")
      expect(result.data?.changeName).toBe("feature/my-branch")
      expect(result.data?.hasChanges).toBe(false)
      expect(result.data?.secretCandidates).toStrictEqual([])
      expect(result.data?.unpushedCommits).toStrictEqual([])
    }
  })

  it("detects main branch", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "main"
      if (cmd.includes("--porcelain")) return ""
      throw new Error("no upstream")
    }
    const result = await invokeTool<{ branch: string; isMain: boolean }>(
      createUfFinaleStartupTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.branch).toBe("main")
      expect(result.data?.isMain).toBe(true)
    }
  })

  it("detects master branch as main", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "master"
      if (cmd.includes("--porcelain")) return ""
      throw new Error("no upstream")
    }
    const result = await invokeTool<{ isMain: boolean }>(
      createUfFinaleStartupTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.isMain).toBe(true)
    }
  })

  it("reports branchType other with changeName for main branch", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "main"
      if (cmd.includes("--porcelain")) return ""
      throw new Error("no upstream")
    }
    const result = await invokeTool<{
      branchType: string
      changeName: string
      isMain: boolean
    }>(createUfFinaleStartupTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.branchType).toBe("other")
      expect(result.data?.isMain).toBe(true)
      expect(result.data?.changeName).toBe("main")
    }
  })

  it("reports branchType other with changeName for master branch", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "master"
      if (cmd.includes("--porcelain")) return ""
      throw new Error("no upstream")
    }
    const result = await invokeTool<{
      branchType: string
      changeName: string
      isMain: boolean
    }>(createUfFinaleStartupTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.branchType).toBe("other")
      expect(result.data?.isMain).toBe(true)
      expect(result.data?.changeName).toBe("master")
    }
  })

  it("classifies quick branches with the quick workflow tier (unleash-tier-routing)", async () => {
    // The shared classifyBranch classifier gained the quick branch type
    // in unleash-tier-routing; finale startup must surface it so /uf.finale
    // can finalize quick/* branches like any other feature branch.
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "quick/fix-typo"
      if (cmd.includes("--porcelain")) return ""
      throw new Error("no upstream")
    }
    const result = await invokeTool<{
      branch: string
      branchType: string
      changeName: string
      isMain: boolean
    }>(createUfFinaleStartupTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.branch).toBe("quick/fix-typo")
      expect(result.data?.branchType).toBe("quick")
      expect(result.data?.changeName).toBe("fix-typo")
      expect(result.data?.isMain).toBe(false)
    }
  })

  it("detects secret candidates in changed files", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "feature/x"
      if (cmd.includes("--porcelain"))
        return " M config.ts\n?? .env\n?? secrets/key.pem\nA  certs/server.key"
      throw new Error("no upstream")
    }
    const result = await invokeTool<{
      secretCandidates: string[]
      hasChanges: boolean
    }>(createUfFinaleStartupTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.hasChanges).toBe(true)
      expect(result.data?.secretCandidates).toContain(".env")
      expect(result.data?.secretCandidates).toContain("secrets/key.pem")
      expect(result.data?.secretCandidates).toContain("certs/server.key")
      // config.ts should NOT be flagged
      expect(result.data?.secretCandidates).not.toContain("config.ts")
    }
  })

  it("detects credentials.json as secret candidate", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "feature/x"
      if (cmd.includes("--porcelain")) return "?? credentials.json"
      throw new Error("no upstream")
    }
    const result = await invokeTool<{ secretCandidates: string[] }>(
      createUfFinaleStartupTool(mockExec, "/fake"),
      {},
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.secretCandidates).toContain("credentials.json")
    }
  })

  it("detects secrets.json as secret candidate", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "feature/x"
      if (cmd.includes("--porcelain")) return "?? secrets.json"
      throw new Error("no upstream")
    }
    const result = await invokeTool<{ secretCandidates: string[] }>(
      createUfFinaleStartupTool(mockExec, "/fake"),
      {},
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.secretCandidates).toContain("secrets.json")
    }
  })

  it("detects opsx branch type", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "opsx/add-feature"
      if (cmd.includes("--porcelain")) return ""
      throw new Error("no upstream")
    }
    const result = await invokeTool<{
      branchType: string
      changeName: string
    }>(createUfFinaleStartupTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.branchType).toBe("opsx")
      expect(result.data?.changeName).toBe("add-feature")
    }
  })

  it("detects speckit branch type (NNN-name)", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "035-hero-interface"
      if (cmd.includes("--porcelain")) return ""
      throw new Error("no upstream")
    }
    const result = await invokeTool<{
      branchType: string
      changeName: string
    }>(createUfFinaleStartupTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.branchType).toBe("speckit")
      expect(result.data?.changeName).toBe("hero-interface")
    }
  })

  it("classifies other branch types", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "chore/cleanup"
      if (cmd.includes("--porcelain")) return ""
      throw new Error("no upstream")
    }
    const result = await invokeTool<{
      branchType: string
      changeName: string
    }>(createUfFinaleStartupTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.branchType).toBe("other")
      expect(result.data?.changeName).toBe("chore/cleanup")
    }
  })

  it("includes unpushed commits when upstream exists", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "feature/x"
      if (cmd.includes("--porcelain")) return ""
      if (cmd.includes("@{u}..HEAD"))
        return "abc1234 first commit\ndef5678 second commit"
      return ""
    }
    const result = await invokeTool<{
      unpushedCommits: Array<{ sha: string; subject: string }>
    }>(createUfFinaleStartupTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.unpushedCommits).toStrictEqual([
        { sha: "abc1234", subject: "first commit" },
        { sha: "def5678", subject: "second commit" },
      ])
    }
  })

  it("returns failure on fatal error", async () => {
    const mockExec: GitExec = () => {
      throw new Error("git not found")
    }
    const result = await invokeTool(
      createUfFinaleStartupTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("STARTUP_ERROR")
      expect(result.message).toContain("git not found")
    }
  })
})

// ---------------------------------------------------------------------------
// uf_finale_stage tests
// ---------------------------------------------------------------------------

describe("uf_finale_stage", () => {
  it("stages files and returns count", async () => {
    const commands: string[] = []
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      commands.push(cmd)
      if (cmd === "add .") return ""
      if (cmd.includes("--cached --name-only"))
        return "file1.ts\nfile2.ts\nfile3.ts"
      return ""
    }
    const result = await invokeTool<{ fileCount: number; files: string[] }>(
      createUfFinaleStageTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.fileCount).toBe(3)
      expect(result.data?.files).toStrictEqual([
        "file1.ts",
        "file2.ts",
        "file3.ts",
      ])
    }
    expect(commands).toContain("add .")
    expect(commands).toContain("diff --cached --name-only")
  })

  it("returns zero count when no files staged", async () => {
    const mockExec: GitExec = () => ""
    const result = await invokeTool<{ fileCount: number; files: string[] }>(
      createUfFinaleStageTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.fileCount).toBe(0)
      expect(result.data?.files).toStrictEqual([])
    }
  })

  it("returns failure when git add fails", async () => {
    const mockExec: GitExec = () => {
      throw new Error("permission denied")
    }
    const result = await invokeTool(
      createUfFinaleStageTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("STAGE_ERROR")
      expect(result.message).toContain("permission denied")
    }
  })
})

// ---------------------------------------------------------------------------
// uf_finale_commit_context tests
// ---------------------------------------------------------------------------

describe("uf_finale_commit_context", () => {
  it("returns diffStat, recentLog, and normalized model", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--cached --stat"))
        return " file1.ts | 10 ++++++++++\n 1 file changed, 10 insertions(+)"
      if (cmd.includes("log -5"))
        return "abc1234 feat: add stuff\ndef5678 fix: thing"
      return ""
    }
    const result = await invokeTool<{
      diffStat: string
      recentLog: string
      normalizedModel: string
    }>(createUfFinaleCommitContextTool(mockExec, "/fake"), {
      rawModelId:
        "google-vertex-anthropic/claude-sonnet-4-20250514@default",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.diffStat).toContain("file1.ts")
      expect(result.data?.recentLog).toContain("abc1234")
      expect(result.data?.normalizedModel).toBe("claude-sonnet-4")
    }
  })

  it("normalizes model with chained sub-versions", async () => {
    const mockExec: GitExec = () => ""
    const result = await invokeTool<{ normalizedModel: string }>(
      createUfFinaleCommitContextTool(mockExec, "/fake"),
      { rawModelId: "anthropic/claude-opus-4-6-20260101@default" },
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.normalizedModel).toBe("claude-opus-4")
    }
  })

  it("returns unknown-model for empty rawModelId", async () => {
    const mockExec: GitExec = () => ""
    const result = await invokeTool<{ normalizedModel: string }>(
      createUfFinaleCommitContextTool(mockExec, "/fake"),
      { rawModelId: "" },
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.normalizedModel).toBe("unknown-model")
    }
  })

  it("handles errors from git commands gracefully", async () => {
    const mockExec: GitExec = () => {
      throw new Error("fail")
    }
    const result = await invokeTool<{
      diffStat: string
      recentLog: string
      normalizedModel: string
    }>(createUfFinaleCommitContextTool(mockExec, "/fake"), {
      rawModelId: "gpt-4o",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.diffStat).toBe("")
      expect(result.data?.recentLog).toBe("")
      expect(result.data?.normalizedModel).toBe("gpt-4o")
    }
  })

  it("passes through already-normalized model names", async () => {
    const mockExec: GitExec = () => ""
    const result = await invokeTool<{ normalizedModel: string }>(
      createUfFinaleCommitContextTool(mockExec, "/fake"),
      { rawModelId: "claude-sonnet-4" },
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.normalizedModel).toBe("claude-sonnet-4")
    }
  })
})

// ---------------------------------------------------------------------------
// uf_finale_base_state tests
// ---------------------------------------------------------------------------

describe("uf_finale_base_state", () => {
  it("prefers the upstream remote and resolves the base branch for forks", async () => {
    const commands: string[] = []
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      commands.push(cmd)
      if (cmd === "remote") return "origin\nupstream\n"
      if (cmd.includes("symbolic-ref refs/remotes/upstream/HEAD"))
        return "refs/remotes/upstream/main\n"
      if (cmd.includes("fetch upstream main")) return ""
      if (cmd.includes("upstream/main..HEAD --count")) return "4"
      if (cmd.includes("HEAD..upstream/main --count")) return "1"
      return ""
    }
    const result = await invokeTool<{
      remote: string
      baseBranch: string
      baseRef: string
      aheadBy: number
      behindBy: number
    }>(createUfFinaleBaseStateTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.remote).toBe("upstream")
      expect(result.data?.baseBranch).toBe("main")
      expect(result.data?.baseRef).toBe("upstream/main")
      expect(result.data?.aheadBy).toBe(4)
      expect(result.data?.behindBy).toBe(1)
    }
    expect(commands).toContain("fetch upstream main")
  })

  it("resolves origin and reports ahead/behind counts in non-fork repositories", async () => {
    const commands: string[] = []
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      commands.push(cmd)
      if (cmd === "remote") return "origin\n"
      if (cmd.includes("symbolic-ref refs/remotes/origin/HEAD"))
        return "refs/remotes/origin/main\n"
      if (cmd.includes("fetch origin main")) return ""
      if (cmd.includes("origin/main..HEAD --count")) return "3"
      if (cmd.includes("HEAD..origin/main --count")) return "0"
      return ""
    }
    const result = await invokeTool<{
      remote: string
      baseBranch: string
      baseRef: string
      aheadBy: number
      behindBy: number
    }>(createUfFinaleBaseStateTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.remote).toBe("origin")
      expect(result.data?.baseBranch).toBe("main")
      expect(result.data?.baseRef).toBe("origin/main")
      expect(result.data?.aheadBy).toBe(3)
      expect(result.data?.behindBy).toBe(0)
    }
    expect(commands).toContain("fetch origin main")
  })

  it("falls back to base branch main when the remote HEAD symbolic ref is unset", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd === "remote") return "origin\n"
      if (cmd.includes("symbolic-ref refs/remotes/origin/HEAD"))
        throw new Error("ref is not a symbolic ref")
      if (cmd.includes("fetch origin main")) return ""
      if (cmd.includes("origin/main..HEAD --count")) return "1"
      if (cmd.includes("HEAD..origin/main --count")) return "2"
      return ""
    }
    const result = await invokeTool<{
      remote: string
      baseBranch: string
      baseRef: string
      aheadBy: number
      behindBy: number
    }>(createUfFinaleBaseStateTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.remote).toBe("origin")
      expect(result.data?.baseBranch).toBe("main")
      expect(result.data?.baseRef).toBe("origin/main")
      expect(result.data?.aheadBy).toBe(1)
      expect(result.data?.behindBy).toBe(2)
    }
  })

  it("returns a retryable BASE_SYNC_FETCH failure when the fetch fails offline", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd === "remote") return "origin\nupstream\n"
      if (cmd.includes("symbolic-ref refs/remotes/upstream/HEAD"))
        return "refs/remotes/upstream/main\n"
      if (cmd.includes("fetch upstream main"))
        throw new Error("could not read from remote repository")
      return ""
    }
    const result = await invokeTool(
      createUfFinaleBaseStateTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("BASE_SYNC_FETCH")
      expect(result.retryable).toBe(true)
      expect(result.message).toContain("upstream main")
      expect(result.message).toContain("could not read from remote repository")
    }
  })

  it("returns a retryable BASE_SYNC_COUNT failure when rev-list counting fails", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd === "remote") return "origin\n"
      if (cmd.includes("symbolic-ref refs/remotes/origin/HEAD"))
        return "refs/remotes/origin/main\n"
      if (cmd.includes("fetch origin main")) return ""
      if (cmd.includes("rev-list")) throw new Error("bad revision")
      return ""
    }
    const result = await invokeTool(
      createUfFinaleBaseStateTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("BASE_SYNC_COUNT")
      expect(result.retryable).toBe(true)
      expect(result.message).toContain("origin/main")
      expect(result.message).toContain("bad revision")
    }
  })

  it("returns a retryable BASE_SYNC_COUNT failure when rev-list output is unparseable", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd === "remote") return "origin\n"
      if (cmd.includes("symbolic-ref refs/remotes/origin/HEAD"))
        return "refs/remotes/origin/main\n"
      if (cmd.includes("fetch origin main")) return ""
      // Successful exit but garbage stdout — the old parseInt || 0
      // coercion turned this into the 0 sentinel.
      if (cmd.includes("origin/main..HEAD --count")) return "2"
      if (cmd.includes("HEAD..origin/main --count")) return "garbage"
      return ""
    }
    const result = await invokeTool(
      createUfFinaleBaseStateTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("BASE_SYNC_COUNT")
      expect(result.retryable).toBe(true)
      expect(result.message).toContain("unparseable rev-list --count output")
      expect(result.message).toContain("garbage")
    }
  })

  it("reports behind-base counts without mutating the working tree", async () => {
    const commands: string[] = []
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      commands.push(cmd)
      if (cmd === "remote") return "origin\n"
      if (cmd.includes("symbolic-ref refs/remotes/origin/HEAD"))
        return "refs/remotes/origin/main\n"
      if (cmd.includes("fetch origin main")) return ""
      if (cmd.includes("origin/main..HEAD --count")) return "2"
      if (cmd.includes("HEAD..origin/main --count")) return "7"
      return ""
    }
    const result = await invokeTool<{
      remote: string
      baseBranch: string
      baseRef: string
      aheadBy: number
      behindBy: number
    }>(createUfFinaleBaseStateTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.remote).toBe("origin")
      expect(result.data?.baseBranch).toBe("main")
      expect(result.data?.baseRef).toBe("origin/main")
      expect(result.data?.aheadBy).toBe(2)
      expect(result.data?.behindBy).toBe(7)
    }
    // Read-only commands plus exactly one fetch — no checkout, rebase,
    // reset, or merge that would modify the working tree.
    expect(commands).toEqual([
      "remote",
      "symbolic-ref refs/remotes/origin/HEAD",
      "fetch origin main",
      "rev-list origin/main..HEAD --count",
      "rev-list HEAD..origin/main --count",
    ])
  })

  it("returns BASE_STATE_ERROR when git remote listing fails", async () => {
    const mockExec: GitExec = () => {
      throw new Error("git not found")
    }
    const result = await invokeTool(
      createUfFinaleBaseStateTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("BASE_STATE_ERROR")
      expect(result.message).toContain("git not found")
    }
  })
})

// ---------------------------------------------------------------------------
// uf_finale_push_state tests
// ---------------------------------------------------------------------------

describe("uf_finale_push_state", () => {
  it("reports no upstream when tracking not configured", async () => {
    const mockExec: GitExec = () => {
      throw new Error("no upstream")
    }
    const result = await invokeTool<{
      upstreamExists: boolean
      diverged: boolean
      aheadBy: number
      behindBy: number
    }>(createUfFinalePushStateTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.upstreamExists).toBe(false)
      expect(result.data?.aheadBy).toBe(0)
      expect(result.data?.behindBy).toBe(0)
      expect(result.data?.diverged).toBe(false)
    }
  })

  it("reports ahead-by count when upstream exists", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref @{u}")) return "origin/feature"
      if (cmd.includes("@{u}..HEAD --count")) return "3"
      if (cmd.includes("HEAD..@{u} --count")) return "0"
      return ""
    }
    const result = await invokeTool<{
      upstreamExists: boolean
      aheadBy: number
      behindBy: number
      diverged: boolean
    }>(createUfFinalePushStateTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.upstreamExists).toBe(true)
      expect(result.data?.aheadBy).toBe(3)
      expect(result.data?.behindBy).toBe(0)
      expect(result.data?.diverged).toBe(false)
    }
  })

  it("detects diverged branches", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref @{u}")) return "origin/feature"
      if (cmd.includes("@{u}..HEAD --count")) return "2"
      if (cmd.includes("HEAD..@{u} --count")) return "5"
      return ""
    }
    const result = await invokeTool<{
      upstreamExists: boolean
      aheadBy: number
      behindBy: number
      diverged: boolean
    }>(createUfFinalePushStateTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.upstreamExists).toBe(true)
      expect(result.data?.aheadBy).toBe(2)
      expect(result.data?.behindBy).toBe(5)
      expect(result.data?.diverged).toBe(true)
    }
  })

  it("reports behind-by count without divergence", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref @{u}")) return "origin/feature"
      if (cmd.includes("@{u}..HEAD --count")) return "0"
      if (cmd.includes("HEAD..@{u} --count")) return "4"
      return ""
    }
    const result = await invokeTool<{
      upstreamExists: boolean
      aheadBy: number
      behindBy: number
      diverged: boolean
    }>(createUfFinalePushStateTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.upstreamExists).toBe(true)
      expect(result.data?.aheadBy).toBe(0)
      expect(result.data?.behindBy).toBe(4)
      expect(result.data?.diverged).toBe(false)
    }
  })

  it("returns a retryable PUSH_SYNC_COUNT failure when rev-list counting fails", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref @{u}")) return "origin/feature"
      if (cmd.includes("rev-list")) throw new Error("bad revision")
      return ""
    }
    const result = await invokeTool(
      createUfFinalePushStateTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("PUSH_SYNC_COUNT")
      expect(result.retryable).toBe(true)
      expect(result.message).toContain("@{u}")
      expect(result.message).toContain("bad revision")
    }
  })

  it("returns a retryable PUSH_SYNC_COUNT failure when rev-list output is unparseable", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref @{u}")) return "origin/feature"
      // Successful exit but unparseable stdout — the old parseInt
      // || 0 coercion read behindBy as 0 and diverged as false.
      if (cmd.includes("@{u}..HEAD --count")) return "2"
      if (cmd.includes("HEAD..@{u} --count")) return ""
      return ""
    }
    const result = await invokeTool(
      createUfFinalePushStateTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("PUSH_SYNC_COUNT")
      expect(result.retryable).toBe(true)
      expect(result.message).toContain("unparseable rev-list --count output")
    }
  })

  it("returns a retryable PUSH_SYNC_COUNT failure when rev-list output is partially numeric", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref @{u}")) return "origin/feature"
      // Successful exit but prefix-numeric stdout — parseInt("2x")
      // parses as 2, so the full-token digits check must reject it.
      if (cmd.includes("@{u}..HEAD --count")) return "2x"
      if (cmd.includes("HEAD..@{u} --count")) return "1"
      return ""
    }
    const result = await invokeTool(
      createUfFinalePushStateTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("PUSH_SYNC_COUNT")
      expect(result.retryable).toBe(true)
      expect(result.message).toContain("unparseable rev-list --count output")
      expect(result.message).toContain("2x")
    }
  })
})

// ---------------------------------------------------------------------------
// uf_finale_pr_context tests
// ---------------------------------------------------------------------------

describe("uf_finale_pr_context", () => {
  it("returns existing PR when found", async () => {
    const mockExec: GitExec = () => {
      throw new Error("not needed")
    }
    vi.mocked(execSync).mockImplementation((cmd) => {
      const cmdStr = String(cmd)
      if (cmdStr.includes("gh pr view --json number,url,title")) {
        return JSON.stringify({
          number: 42,
          url: "https://github.com/owner/repo/pull/42",
          title: "My PR",
        })
      }
      throw new Error("not called")
    })
    vi.mocked(readFileSync).mockImplementation(() => {
      throw new Error("ENOENT")
    })

    const result = await invokeTool<{
      existingPr: { number: number; url: string; title: string } | null
    }>(createUfFinalePrContextTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.existingPr).toStrictEqual({
        number: 42,
        url: "https://github.com/owner/repo/pull/42",
        title: "My PR",
      })
    }
  })

  it("returns null existingPr when no PR exists", async () => {
    const mockExec: GitExec = () => {
      throw new Error("not needed")
    }
    vi.mocked(execSync).mockImplementation(() => {
      throw new Error("no PR for this branch")
    })
    vi.mocked(readFileSync).mockImplementation(() => {
      throw new Error("ENOENT")
    })

    const result = await invokeTool<{
      existingPr: null
      isFork: boolean
    }>(createUfFinalePrContextTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.existingPr).toBeNull()
      expect(result.data?.isFork).toBe(false)
    }
  })

  it("detects fork with parent info", async () => {
    const mockExec: GitExec = () => {
      throw new Error("skip git commands")
    }
    vi.mocked(execSync).mockImplementation((cmd) => {
      const cmdStr = String(cmd)
      if (cmdStr.includes("gh pr view")) {
        throw new Error("no PR")
      }
      if (cmdStr.includes("--json isFork --jq")) {
        return "true"
      }
      if (cmdStr.includes("--json parent")) {
        return JSON.stringify({
          parent: { owner: { login: "upstream-org" }, name: "upstream-repo" },
        })
      }
      throw new Error("unexpected")
    })
    vi.mocked(readFileSync).mockImplementation(() => {
      throw new Error("ENOENT")
    })

    const result = await invokeTool<{
      isFork: boolean
      forkParent: { owner: string; repo: string } | null
    }>(createUfFinalePrContextTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.isFork).toBe(true)
      expect(result.data?.forkParent).toStrictEqual({
        owner: "upstream-org",
        repo: "upstream-repo",
      })
    }
  })

  it("finds PR template sections", async () => {
    const mockExec: GitExec = () => {
      throw new Error("skip")
    }
    vi.mocked(execSync).mockImplementation(() => {
      throw new Error("skip gh")
    })
    vi.mocked(readFileSync).mockImplementation((pathArg) => {
      const p = String(pathArg)
      if (p.includes("PULL_REQUEST_TEMPLATE.md")) {
        return "## Summary\nDescribe changes.\n\n## Changes\nList changes.\n\n## Testing\nHow tested."
      }
      throw new Error("ENOENT")
    })

    const result = await invokeTool<{ templateSections: string[] }>(
      createUfFinalePrContextTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.templateSections).toStrictEqual([
        "Summary",
        "Changes",
        "Testing",
      ])
    }
  })

  it("parses originating issue from .openspec.yaml", async () => {
    const mockExec: GitExec = () => {
      throw new Error("skip")
    }
    vi.mocked(execSync).mockImplementation(() => {
      throw new Error("skip gh")
    })
    vi.mocked(readFileSync).mockImplementation((pathArg) => {
      const p = String(pathArg)
      if (p.includes(".openspec.yaml")) {
        return "name: my-change\noriginating_issue: 123\n"
      }
      throw new Error("ENOENT")
    })

    const result = await invokeTool<{ originatingIssue: number | null }>(
      createUfFinalePrContextTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.originatingIssue).toBe(123)
    }
    // Verify the exact path passed to readFileSync for .openspec.yaml
    const readCalls = vi.mocked(readFileSync).mock.calls
    const openspecCall = readCalls.find((call) =>
      String(call[0]).includes(".openspec.yaml"),
    )
    expect(openspecCall).toBeDefined()
    expect(openspecCall![0]).toBe(resolve("/fake", ".openspec.yaml"))
  })

  it("threads the issue number from the tier-routing route decision (unleash-tier-routing)", async () => {
    // Quick changes carry no .openspec.yaml; the branch-scoped route
    // decision checkpoint is the only issue record. The reader must
    // find it at the exact uf_unleash_route_decision write path so
    // /uf.finale can thread `Closes #<N>` into the PR body.
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "quick/fix-typo"
      throw new Error("skip")
    }
    vi.mocked(execSync).mockImplementation(() => {
      throw new Error("skip gh")
    })
    vi.mocked(readFileSync).mockImplementation((pathArg) => {
      const p = String(pathArg)
      if (p.includes(".uf/progress/quick-fix-typo/uf.unleash.json")) {
        return JSON.stringify({
          command: "uf.unleash",
          branch: "quick-fix-typo",
          key: "quick-fix-typo",
          current_step: "implement",
          completed_steps: ["tier-routing", "implement"],
          data: {
            tier: "quick",
            change_name: "fix-typo",
            issue_number: 42,
            rationale: "Trivial typo fix",
          },
        })
      }
      throw new Error("ENOENT")
    })

    const result = await invokeTool<{ originatingIssue: number | null }>(
      createUfFinalePrContextTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.originatingIssue).toBe(42)
    }
    // Verify the exact branch-scoped state path was read.
    const readCalls = vi.mocked(readFileSync).mock.calls
    const stateCall = readCalls.find((call) =>
      String(call[0]).includes(".uf/progress/"),
    )
    expect(stateCall).toBeDefined()
    expect(stateCall![0]).toBe(
      resolve("/fake", ".uf/progress/quick-fix-typo/uf.unleash.json"),
    )
  })

  it("re-keys an openspec route decision under the feature branch so the finale reader threads the issue (unleash-tier-routing 9.2)", async () => {
    // End-to-end handoff for the speckit/openspec routes: the route
    // decision (Step 1a.4) records tier-routing under the pre-route
    // branch key (main), the pipeline re-keys it under the feature
    // branch the delegated entry command created (Step 1a.7, checkpoint
    // key override), and uf_finale_pr_context reads the branch-scoped
    // state — the field /uf.finale renders as `Closes #<N>` in the PR
    // body. The three real tools share one in-memory state store: if
    // either side's branch-slug derivation drifted, the reader would
    // miss and originatingIssue would stay null.
    const files = new Map<string, string>()
    const enoent = Object.assign(new Error("ENOENT: no such file or directory"), {
      code: "ENOENT",
    })
    vi.mocked(writeFileSync).mockImplementation((pathArg, dataArg) => {
      files.set(String(pathArg), String(dataArg))
    })
    vi.mocked(readFileSync).mockImplementation((pathArg) => {
      const p = String(pathArg)
      if (files.has(p)) return files.get(p) as string
      throw enoent
    })
    vi.mocked(lstatSync).mockImplementation(() => {
      throw enoent
    })
    vi.mocked(mkdirSync).mockImplementation(() => {})
    vi.mocked(execSync).mockImplementation(() => {
      throw new Error("skip gh")
    })

    // Branch state: routing ran on main; the delegated /opsx-propose
    // created and switched to opsx/fix-token-validation before the
    // re-key and the finale read.
    let currentBranch = "main"
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd === "rev-parse --abbrev-ref HEAD") return currentBranch
      if (cmd === "status --porcelain") return ""
      throw new Error(`unexpected git command: ${cmd}`)
    }

    // Phase 1 — the route decision: a security-sensitive fix routed
    // from issue #743 escalates to openspec (risk floor) and records
    // tier-routing under the pre-route branch key.
    const routeResult = await invokeTool<{
      tier: string
      change_name: string
      branch: string
      rationale: string
      description: string
      issue_number: number | null
    }>(createUnleashRouteDecisionTool(mockExec, "/fake", () => true), {
      tier: "openspec",
      change_name: "fix-token-validation",
      rationale: "Security-sensitive fix — the risk floor routes it to openspec at minimum.",
      description: "Fix the token validation bypass reported in the issue.",
      issue_number: 743,
      risk_signals: ["security-sensitive"],
    })
    expect(routeResult.status).toBe("ok")
    if (routeResult.status === "ok") {
      expect(routeResult.data?.issue_number).toBe(743)
    }

    const preRoutePath = resolve("/fake", ".uf/progress/main/uf.unleash.json")
    const featurePath = resolve(
      "/fake",
      ".uf/progress/opsx-fix-token-validation/uf.unleash.json",
    )
    expect(files.has(preRoutePath)).toBe(true)
    expect(files.has(featurePath)).toBe(false)
    const preRouteState = JSON.parse(
      files.get(preRoutePath) as string,
    ) as ProgressState
    expect(preRouteState.completed_steps).toEqual(["tier-routing"])
    expect(preRouteState.data.issue_number).toBe(743)

    // The delegated /opsx-propose created the feature branch.
    currentBranch = "opsx/fix-token-validation"

    // Phase 2 — the re-key (Step 1a.7): the checkpoint tool's key
    // override re-records the route decision under the verified feature
    // branch, data passed through from the decision unchanged.
    if (routeResult.status !== "ok") {
      throw new Error("route decision failed before the re-key")
    }
    const decision = routeResult.data
    const rekeyResult = await invokeTool(
      createCheckpointTool(mockExec, "/fake"),
      {
        command: "uf.unleash",
        step: "tier-routing",
        key: "opsx/fix-token-validation",
        data: JSON.stringify({
          tier: decision?.tier,
          change_name: decision?.change_name,
          description: decision?.description,
          issue_number: decision?.issue_number,
          rationale: decision?.rationale,
        }),
      },
    )
    expect(rekeyResult.status).toBe("ok")
    expect(files.has(featurePath)).toBe(true)
    const featureState = JSON.parse(
      files.get(featurePath) as string,
    ) as ProgressState
    expect(featureState.completed_steps).toEqual(["tier-routing"])
    expect(featureState.key).toBe("opsx-fix-token-validation")
    expect(featureState.data.issue_number).toBe(743)

    // Phase 3 — /uf.finale's branch-scoped reader finds the re-keyed
    // decision and reports the issue the PR body closes.
    const result = await invokeTool<{ originatingIssue: number | null }>(
      createUfFinalePrContextTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.originatingIssue).toBe(743)
    }
    // The reader consumed the feature-branch state, not the pre-route one.
    const stateReads = vi.mocked(readFileSync).mock.calls
      .map((call) => String(call[0]))
      .filter((p) => p.includes(".uf/progress/"))
    expect(stateReads).toContain(featurePath)
  })

  it("prefers the route-decision issue number over .openspec.yaml", async () => {
    // The route decision is branch-scoped; the repo-root .openspec.yaml
    // can go stale across changes. The branch-scoped record wins.
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "quick/fix-typo"
      throw new Error("skip")
    }
    vi.mocked(execSync).mockImplementation(() => {
      throw new Error("skip gh")
    })
    vi.mocked(readFileSync).mockImplementation((pathArg) => {
      const p = String(pathArg)
      if (p.includes(".openspec.yaml")) {
        return "name: stale-change\noriginating_issue: 123\n"
      }
      if (p.includes(".uf/progress/quick-fix-typo/uf.unleash.json")) {
        return JSON.stringify({
          completed_steps: ["tier-routing"],
          data: { issue_number: 42 },
        })
      }
      throw new Error("ENOENT")
    })

    const result = await invokeTool<{ originatingIssue: number | null }>(
      createUfFinalePrContextTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.originatingIssue).toBe(42)
    }
  })

  it("keeps the .openspec.yaml issue when no tier-routing step was recorded", async () => {
    // Fail closed: state data mentioning an issue number without the
    // recorded tier-routing step must not override .openspec.yaml.
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "opsx/some-change"
      throw new Error("skip")
    }
    vi.mocked(execSync).mockImplementation(() => {
      throw new Error("skip gh")
    })
    vi.mocked(readFileSync).mockImplementation((pathArg) => {
      const p = String(pathArg)
      if (p.includes(".openspec.yaml")) {
        return "name: some-change\noriginating_issue: 123\n"
      }
      if (p.includes(".uf/progress/opsx-some-change/uf.unleash.json")) {
        return JSON.stringify({
          completed_steps: ["branch-safety-gate", "resumability-detection"],
          data: { issue_number: 42 },
        })
      }
      throw new Error("ENOENT")
    })

    const result = await invokeTool<{ originatingIssue: number | null }>(
      createUfFinalePrContextTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.originatingIssue).toBe(123)
    }
  })

  it("keeps the .openspec.yaml issue when the route-decision state carries a non-integer issue number", async () => {
    // Fail closed: the tier-routing step is recorded, but the state's
    // issue_number is a string rather than a positive integer. The
    // guard must reject it and let .openspec.yaml stand.
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "opsx/some-change"
      throw new Error("skip")
    }
    vi.mocked(execSync).mockImplementation(() => {
      throw new Error("skip gh")
    })
    vi.mocked(readFileSync).mockImplementation((pathArg) => {
      const p = String(pathArg)
      if (p.includes(".openspec.yaml")) {
        return "name: some-change\noriginating_issue: 123\n"
      }
      if (p.includes(".uf/progress/opsx-some-change/uf.unleash.json")) {
        return JSON.stringify({
          completed_steps: ["tier-routing"],
          data: { issue_number: "42" },
        })
      }
      throw new Error("ENOENT")
    })

    const result = await invokeTool<{ originatingIssue: number | null }>(
      createUfFinalePrContextTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.originatingIssue).toBe(123)
    }
  })

  it("keeps the .openspec.yaml issue when the route-decision state JSON is corrupt", async () => {
    // Fail closed: an unparseable route-decision state must not crash
    // the reader or leak a partial value — the guard's try/catch
    // falls back to the .openspec.yaml originating issue.
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "opsx/some-change"
      throw new Error("skip")
    }
    vi.mocked(execSync).mockImplementation(() => {
      throw new Error("skip gh")
    })
    vi.mocked(readFileSync).mockImplementation((pathArg) => {
      const p = String(pathArg)
      if (p.includes(".openspec.yaml")) {
        return "name: some-change\noriginating_issue: 123\n"
      }
      if (p.includes(".uf/progress/opsx-some-change/uf.unleash.json")) {
        return "{corrupt"
      }
      throw new Error("ENOENT")
    })

    const result = await invokeTool<{ originatingIssue: number | null }>(
      createUfFinalePrContextTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.originatingIssue).toBe(123)
    }
  })

  it("serves a write_unleash_summary artifact through uf_finale_pr_context (#707 writer-reader contract)", async () => {
    // Connecting test for the cross-plugin artifact contract (#707): the
    // artifact write_unleash_summary persists is the artifact this reader
    // returns. A hand-written fixture could drift from the writer's shape;
    // running the real writer against the same mocked filesystem and branch
    // cannot.
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("merge-base")) throw new Error("skip")
      if (cmd.includes("--abbrev-ref HEAD")) return "opsx/my-change"
      throw new Error("skip")
    }
    vi.mocked(execSync).mockImplementation(() => {
      throw new Error("skip gh")
    })

    // Phase 1 — write. lstat reports both artifact paths absent (first
    // write), the directory creation is a no-op, and the write itself is
    // captured for phase 2.
    const enoent = Object.assign(
      new Error("ENOENT: no such file or directory"),
      { code: "ENOENT" },
    )
    vi.mocked(lstatSync).mockImplementation(() => {
      throw enoent
    })
    const writerInput = {
      change_name: "my-change",
      tier: "openspec",
      feature_dir: "openspec/changes/my-change/",
      review_verdicts: {
        spec_review: { verdict: "APPROVE", iterations: 1 },
        code_review: { verdict: "APPROVE WITH ADVISORIES", iterations: 2 },
      },
      learnings_stored: [
        ".uf/dewey/learnings/uf-workflow-20261016T135617-jay-flowers.md",
      ],
      whatWasBuilt:
        "Deterministic /uf.unleash tools replace the command's LLM-judgment mechanics.",
      howToVerify: "cd .opencode && npx vitest run, then make check from the repository root.",
      testResults: "make plugin-test green: every vitest suite passing, 0 failed.",
    }
    const writeResult = await invokeTool<{ path: string; bytes: number }>(
      createWriteUnleashSummaryTool(mockExec, "/fake"),
      writerInput,
    )
    expect(writeResult.status).toBe("ok")
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [writePath, writeRaw] = vi.mocked(writeFileSync).mock.calls[0]
    if (writeResult.status === "ok") {
      expect(writeResult.data.path).toBe(String(writePath))
    }

    // Phase 2 — read. The reader is served only at the exact path the
    // writer wrote; if either side's branch-slug derivation drifted, the
    // read would miss and unleashSummary would stay null.
    vi.mocked(readFileSync).mockImplementation((pathArg) => {
      if (String(pathArg) === String(writePath)) return writeRaw as string
      throw new Error("ENOENT")
    })

    const result = await invokeTool<{
      unleashSummary: UnleashSummary | null
    }>(createUfFinalePrContextTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      // The three finale contract fields populate from the writer's
      // artifact (#707), with the writer's snake_case-to-camelCase mapping.
      expect(result.data?.unleashSummary?.whatWasBuilt).toBe(writerInput.whatWasBuilt)
      expect(result.data?.unleashSummary?.howToVerify).toBe(writerInput.howToVerify)
      expect(result.data?.unleashSummary?.testResults).toBe(writerInput.testResults)
      expect(result.data?.unleashSummary?.changeName).toBe("my-change")
      expect(result.data?.unleashSummary?.tier).toBe("openspec")
      expect(result.data?.unleashSummary?.reviewVerdicts.codeReview?.verdict).toBe(
        "APPROVE WITH ADVISORIES",
      )
      // The remaining writer parameters (seam-verification 3.1): feature_dir
      // renames to featureDir, and the review-gate/learnings structures
      // survive the JSON round-trip intact — both nested verdicts keep
      // their {verdict, iterations} shape with the snake_case gate keys
      // renamed to specReview/codeReview, and the learnings path list
      // arrives element-for-element.
      expect(result.data?.unleashSummary?.featureDir).toBe(writerInput.feature_dir)
      expect(result.data?.unleashSummary?.reviewVerdicts.specReview).toStrictEqual(
        writerInput.review_verdicts.spec_review,
      )
      expect(result.data?.unleashSummary?.reviewVerdicts.codeReview).toStrictEqual(
        writerInput.review_verdicts.code_review,
      )
      expect(result.data?.unleashSummary?.learningsStored).toStrictEqual(
        writerInput.learnings_stored,
      )
    }
  })

  it("returns all fields as null/empty when nothing found", async () => {
    const mockExec: GitExec = () => {
      throw new Error("skip")
    }
    vi.mocked(execSync).mockImplementation(() => {
      throw new Error("skip")
    })
    vi.mocked(readFileSync).mockImplementation(() => {
      throw new Error("ENOENT")
    })

    const result = await invokeTool<{
      existingPr: null
      isFork: boolean
      forkParent: null
      templateSections: string[]
      commitLog: unknown[]
      diffStat: string
      originatingIssue: null
      unleashSummary: null
    }>(createUfFinalePrContextTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.existingPr).toBeNull()
      expect(result.data?.isFork).toBe(false)
      expect(result.data?.forkParent).toBeNull()
      expect(result.data?.templateSections).toStrictEqual([])
      expect(result.data?.commitLog).toStrictEqual([])
      expect(result.data?.diffStat).toBe("")
      expect(result.data?.originatingIssue).toBeNull()
      expect(result.data?.unleashSummary).toBeNull()
    }
  })

  it("includes commit log and diff stat from merge-base", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("merge-base")) return "abc123def456abc123def456abc123def456abc1"
      if (cmd.includes("log --oneline"))
        return "def456 feat: stuff\nghi789 fix: bug"
      if (cmd.includes("diff --stat"))
        return " file.ts | 5 +++++\n 1 file changed"
      if (cmd.includes("--abbrev-ref HEAD")) return "feature/x"
      throw new Error("skip")
    }
    vi.mocked(execSync).mockImplementation(() => {
      throw new Error("skip gh")
    })
    vi.mocked(readFileSync).mockImplementation(() => {
      throw new Error("ENOENT")
    })

    const result = await invokeTool<{
      commitLog: Array<{ sha: string; subject: string }>
      diffStat: string
    }>(createUfFinalePrContextTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.commitLog).toStrictEqual([
        { sha: "def456", subject: "feat: stuff" },
        { sha: "ghi789", subject: "fix: bug" },
      ])
      expect(result.data?.diffStat).toContain("file.ts")
    }
  })
})

// ---------------------------------------------------------------------------
// uf_finale_create_pr tests
// ---------------------------------------------------------------------------

describe("uf_finale_create_pr", () => {
  it("creates PR and returns number and URL", async () => {
    const mockExec: GitExec = () => ""
    vi.mocked(mkdtempSync).mockReturnValue("/tmp/uf-finale-test123")
    vi.mocked(writeFileSync).mockImplementation(() => {})
    vi.mocked(rmSync).mockImplementation(() => {})
    vi.mocked(execSync).mockReturnValue(
      "https://github.com/owner/repo/pull/42\n",
    )

    const result = await invokeTool<{ number: number; url: string }>(
      createUfFinaleCreatePrTool(mockExec, "/fake"),
      { title: "My PR title", body: "PR body content" },
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.number).toBe(42)
      expect(result.data?.url).toBe(
        "https://github.com/owner/repo/pull/42",
      )
    }
    // Verify temp file was written with restricted permissions
    expect(writeFileSync).toHaveBeenCalledWith(
      "/tmp/uf-finale-test123/pr-body.md",
      "PR body content",
      { mode: 0o600 },
    )
    // Verify cleanup
    expect(rmSync).toHaveBeenCalledWith("/tmp/uf-finale-test123", {
      recursive: true,
      force: true,
    })
  })

  it("passes targetRepo flag for fork PRs", async () => {
    const mockExec: GitExec = () => ""
    vi.mocked(mkdtempSync).mockReturnValue("/tmp/uf-finale-fork")
    vi.mocked(writeFileSync).mockImplementation(() => {})
    vi.mocked(rmSync).mockImplementation(() => {})
    vi.mocked(execSync).mockReturnValue(
      "https://github.com/upstream/repo/pull/99\n",
    )

    const result = await invokeTool<{ number: number; url: string }>(
      createUfFinaleCreatePrTool(mockExec, "/fake"),
      {
        title: "Fork PR",
        body: "body",
        targetRepo: "upstream/repo",
      },
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.number).toBe(99)
    }
    // Verify the command included --repo flag
    const cmdArg = vi.mocked(execSync).mock.calls[0][0] as string
    expect(cmdArg).toContain("--repo upstream/repo")
  })

  it("returns INVALID_TARGET_REPO for malformed targetRepo", async () => {
    const mockExec: GitExec = () => ""
    const result = await invokeTool(
      createUfFinaleCreatePrTool(mockExec, "/fake"),
      { title: "PR", body: "body", targetRepo: "evil; rm -rf /" },
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_TARGET_REPO")
      expect(result.message).toContain("Invalid targetRepo format")
    }
    // Verify gh was never called
    expect(execSync).not.toHaveBeenCalled()
  })

  it("returns failure when gh pr create fails", async () => {
    const mockExec: GitExec = () => ""
    vi.mocked(mkdtempSync).mockReturnValue("/tmp/uf-finale-fail")
    vi.mocked(writeFileSync).mockImplementation(() => {})
    vi.mocked(rmSync).mockImplementation(() => {})
    vi.mocked(execSync).mockImplementation(() => {
      throw new Error("gh: not authenticated")
    })

    const result = await invokeTool(
      createUfFinaleCreatePrTool(mockExec, "/fake"),
      { title: "Fail PR", body: "body" },
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("PR_CREATE_ERROR")
      expect(result.message).toContain("gh: not authenticated")
    }
    // Verify cleanup still happens even on error
    expect(rmSync).toHaveBeenCalled()
  })

  it("returns failure when temp dir creation fails", async () => {
    const mockExec: GitExec = () => ""
    vi.mocked(mkdtempSync).mockImplementation(() => {
      throw new Error("ENOSPC: no space left")
    })

    const result = await invokeTool(
      createUfFinaleCreatePrTool(mockExec, "/fake"),
      { title: "PR", body: "body" },
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("TEMP_FILE_ERROR")
      expect(result.message).toContain("no space left")
    }
  })

  it("returns failure when writeFileSync fails", async () => {
    const mockExec: GitExec = () => ""
    vi.mocked(mkdtempSync).mockReturnValue("/tmp/uf-finale-write-fail")
    vi.mocked(writeFileSync).mockImplementation(() => {
      throw new Error("EACCES: permission denied")
    })
    vi.mocked(rmSync).mockImplementation(() => {})

    const result = await invokeTool(
      createUfFinaleCreatePrTool(mockExec, "/fake"),
      { title: "PR", body: "body" },
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("TEMP_FILE_ERROR")
      expect(result.message).toContain("Could not write PR body")
    }
  })

  it("handles unrecognized gh output format gracefully", async () => {
    const mockExec: GitExec = () => ""
    vi.mocked(mkdtempSync).mockReturnValue("/tmp/uf-finale-nourl")
    vi.mocked(writeFileSync).mockImplementation(() => {})
    vi.mocked(rmSync).mockImplementation(() => {})
    vi.mocked(execSync).mockReturnValue("PR created successfully\n")

    const result = await invokeTool<{ number: number; url: string }>(
      createUfFinaleCreatePrTool(mockExec, "/fake"),
      { title: "PR", body: "body" },
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.number).toBe(0)
      expect(result.data?.url).toBe("PR created successfully")
    }
  })
})

// ---------------------------------------------------------------------------
// uf_finale_ci_poll tests
// ---------------------------------------------------------------------------

describe("uf_finale_ci_poll", () => {
  /** execSync answers the merge-conflict pre-check only; checks use Bun.spawn. */
  function mockMergeablePrView(): void {
    vi.mocked(execSync).mockImplementation((cmd) => {
      const cmdStr = String(cmd)
      if (cmdStr.includes("gh pr view")) {
        return JSON.stringify({
          mergeable: "MERGEABLE",
          mergeStateStatus: "CLEAN",
        })
      }
      throw new Error(`unexpected shell invocation: ${cmdStr}`)
    })
  }

  it("returns pass when all checks bucket as pass", async () => {
    const mockExec: GitExec = () => ""
    mockMergeablePrView()
    const result = await withGhSpawn(
      {
        "pr checks": {
          stdout: JSON.stringify([
            { name: "CI", state: "SUCCESS", bucket: "pass", link: "https://ci/1" },
            { name: "Lint", state: "SUCCESS", bucket: "pass", link: "https://ci/2" },
          ]),
          stderr: "",
          exitCode: 0,
        },
      },
      () =>
        invokeTool<{
          status: string
          failedChecks: unknown[]
          mergeable: boolean
          workflowsConfigured: boolean
        }>(createUfFinaleCiPollTool(mockExec, "/fake"), { prNumber: 42 }),
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.status).toBe("pass")
      expect(result.data?.failedChecks).toStrictEqual([])
      expect(result.data?.mergeable).toBe(true)
      expect(result.data?.workflowsConfigured).toBe(true)
    }
  })

  it("pins the verified gh field list and argument-array invocation", async () => {
    const mockExec: GitExec = () => ""
    mockMergeablePrView()
    let argv: string[] | undefined
    let options: { signal?: AbortSignal; cwd?: string } | undefined
    await withGhSpawn(
      {
        "pr checks": {
          stdout: JSON.stringify([
            { name: "CI", state: "SUCCESS", bucket: "pass", link: "https://ci/1" },
          ]),
          stderr: "",
          exitCode: 0,
        },
      },
      async (spawn) => {
        await invokeTool(createUfFinaleCiPollTool(mockExec, "/fake"), {
          prNumber: 42,
        })
        const checksCall = spawn.mock.calls.find(
          (call) =>
            (call[0] as string[])[1] === "pr" &&
            (call[0] as string[])[2] === "checks",
        )
        argv = checksCall?.[0] as string[] | undefined
        options = checksCall?.[1] as
          | { signal?: AbortSignal; cwd?: string }
          | undefined
      },
    )

    expect(argv).toStrictEqual([
      "gh",
      "pr",
      "checks",
      "42",
      "--json",
      "name,state,bucket,link",
    ])
    // FCP-FR-001: each invocation is bounded by a real AbortSignal timeout.
    expect(options?.signal).toBeInstanceOf(AbortSignal)
    // FCP-FR-001: the checks poll runs in the tool's registered working
    // directory so repository resolution matches the gh pr view pre-check
    // and the .github/workflows probe — never the plugin host process cwd.
    expect(options?.cwd).toBe("/fake")
  })

  it("returns fail when checks bucket as fail", async () => {
    const mockExec: GitExec = () => ""
    mockMergeablePrView()
    const result = await withGhSpawn(
      {
        "pr checks": {
          stdout: JSON.stringify([
            { name: "CI", state: "SUCCESS", bucket: "pass", link: "https://ci/1" },
            { name: "Lint", state: "FAILURE", bucket: "fail", link: "https://ci/lint" },
          ]),
          stderr: "",
          exitCode: 1,
        },
      },
      () =>
        invokeTool<{
          status: string
          failedChecks: Array<{
            name: string
            conclusion: string
            url: string
          }>
          mergeable: boolean
        }>(createUfFinaleCiPollTool(mockExec, "/fake"), { prNumber: 42 }),
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.status).toBe("fail")
      expect(result.data?.mergeable).toBe(false)
      expect(result.data?.failedChecks).toHaveLength(1)
      expect(result.data?.failedChecks[0].name).toBe("Lint")
      expect(result.data?.failedChecks[0].conclusion).toBe("FAILURE")
      expect(result.data?.failedChecks[0].url).toBe("https://ci/lint")
    }
  })

  it("counts a cancel bucket as failed", async () => {
    const mockExec: GitExec = () => ""
    mockMergeablePrView()
    const result = await withGhSpawn(
      {
        "pr checks": {
          stdout: JSON.stringify([
            { name: "CI", state: "CANCELLED", bucket: "cancel", link: "https://ci/1" },
          ]),
          stderr: "",
          exitCode: 1,
        },
      },
      () =>
        invokeTool<{
          status: string
          failedChecks: Array<{ name: string; conclusion: string }>
        }>(createUfFinaleCiPollTool(mockExec, "/fake"), { prNumber: 42 }),
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.status).toBe("fail")
      expect(result.data?.failedChecks[0].conclusion).toBe("CANCELLED")
    }
  })

  it("treats pass and skipping buckets as passing", async () => {
    const mockExec: GitExec = () => ""
    mockMergeablePrView()
    const result = await withGhSpawn(
      {
        "pr checks": {
          stdout: JSON.stringify([
            { name: "CI", state: "SUCCESS", bucket: "pass", link: "https://ci/1" },
            { name: "Optional", state: "SKIPPED", bucket: "skipping", link: "https://ci/2" },
          ]),
          stderr: "",
          exitCode: 0,
        },
      },
      () =>
        invokeTool<{
          status: string
          failedChecks: unknown[]
        }>(createUfFinaleCiPollTool(mockExec, "/fake"), { prNumber: 42 }),
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.status).toBe("pass")
      expect(result.data?.failedChecks).toStrictEqual([])
    }
  })

  it("classifies an unknown bucket fail-closed with its state surfaced", async () => {
    const mockExec: GitExec = () => ""
    mockMergeablePrView()
    const result = await withGhSpawn(
      {
        "pr checks": {
          stdout: JSON.stringify([
            { name: "CI", state: "NEUTRAL", bucket: "blocked", link: "https://ci/1" },
          ]),
          stderr: "",
          exitCode: 0,
        },
      },
      () =>
        invokeTool<{
          status: string
          failedChecks: Array<{ name: string; conclusion: string }>
        }>(createUfFinaleCiPollTool(mockExec, "/fake"), { prNumber: 42 }),
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.status).toBe("fail")
      expect(result.data?.failedChecks[0].name).toBe("CI")
      expect(result.data?.failedChecks[0].conclusion).toBe("NEUTRAL")
    }
  })

  it("returns conflict when PR has merge conflicts", async () => {
    const mockExec: GitExec = () => ""
    vi.mocked(execSync).mockImplementation((cmd) => {
      const cmdStr = String(cmd)
      if (cmdStr.includes("gh pr view")) {
        return JSON.stringify({
          mergeable: "CONFLICTING",
          mergeStateStatus: "DIRTY",
        })
      }
      throw new Error("should not reach checks")
    })

    const result = await invokeTool<{
      status: string
      mergeable: boolean
    }>(createUfFinaleCiPollTool(mockExec, "/fake"), { prNumber: 42 })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.status).toBe("conflict")
      expect(result.data?.mergeable).toBe(false)
    }
  })

  it("returns GH_CLI_ERROR when stdout is not a JSON array and stderr has no marker", async () => {
    const mockExec: GitExec = () => ""
    mockMergeablePrView()
    const result = await withGhSpawn(
      {
        "pr checks": {
          stdout: "not a json array",
          stderr: "fatal: HTTP 503 Service Unavailable",
          exitCode: 1,
        },
      },
      () =>
        invokeTool(createUfFinaleCiPollTool(mockExec, "/fake"), { prNumber: 42 }),
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("GH_CLI_ERROR")
      expect(result.retryable).toBe(true)
      expect(result.message).toContain("HTTP 503 Service Unavailable")
    }
  })

  it("returns PR_NOT_FOUND when stderr reports an unresolvable PR", async () => {
    const mockExec: GitExec = () => ""
    mockMergeablePrView()
    const result = await withGhSpawn(
      {
        "pr checks": {
          stdout: "",
          stderr:
            "Could not resolve to a PullRequest with the number of 999999",
          exitCode: 1,
        },
      },
      () =>
        invokeTool(createUfFinaleCiPollTool(mockExec, "/fake"), {
          prNumber: 42,
        }),
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("PR_NOT_FOUND")
      expect(result.retryable).toBe(false)
      expect(result.message).toContain("42")
      // FCP-FR-002: the matched gh stderr is echoed for diagnosability.
      expect(result.message).toContain("Could not resolve to a PullRequest")
    }
  })

  it("classifies an unrelated not-found stderr as retryable GH_CLI_ERROR, not PR_NOT_FOUND", async () => {
    const mockExec: GitExec = () => ""
    mockMergeablePrView()
    const result = await withGhSpawn(
      {
        "pr checks": {
          stdout: "",
          stderr: "HTTP 404: resource not found",
          exitCode: 1,
        },
      },
      () =>
        invokeTool(createUfFinaleCiPollTool(mockExec, "/fake"), {
          prNumber: 42,
        }),
    )

    // FCP-FR-002: only gh 2.91.0's verified unresolvable-PR shape
    // classifies as PR_NOT_FOUND; looser "not found" phrasing falls
    // through to the retryable error path so transient failures keep
    // their retry semantics.
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("GH_CLI_ERROR")
      expect(result.retryable).toBe(true)
      expect(result.message).toContain("HTTP 404: resource not found")
    }
  })

  it("detects no-checks with workflows existing", async () => {
    const mockExec: GitExec = () => ""
    mockMergeablePrView()
    vi.mocked(existsSync).mockReturnValue(true)
    const result = await withGhSpawn(
      {
        "pr checks": {
          stdout: "",
          stderr: "no checks reported on the 'feature/x' branch",
          exitCode: 1,
        },
      },
      () =>
        invokeTool<{
          status: string
          workflowsConfigured: boolean
          skipReason: string
        }>(createUfFinaleCiPollTool(mockExec, "/fake"), { prNumber: 42 }),
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.status).toBe("no-checks-workflows-exist")
      expect(result.data?.workflowsConfigured).toBe(true)
      expect(result.data?.skipReason).toBeDefined()
    }
  })

  it("detects no-checks with no workflows directory", async () => {
    const mockExec: GitExec = () => ""
    mockMergeablePrView()
    vi.mocked(existsSync).mockReturnValue(false)
    const result = await withGhSpawn(
      {
        "pr checks": {
          stdout: "",
          stderr: "no checks reported on the 'feature/x' branch",
          exitCode: 1,
        },
      },
      () =>
        invokeTool<{
          status: string
          workflowsConfigured: boolean
        }>(createUfFinaleCiPollTool(mockExec, "/fake"), { prNumber: 42 }),
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.status).toBe("no-checks-no-workflows")
      expect(result.data?.workflowsConfigured).toBe(false)
    }
  })

  it("honors a parsed empty array as zero checks", async () => {
    const mockExec: GitExec = () => ""
    mockMergeablePrView()
    vi.mocked(existsSync).mockReturnValue(true)
    const result = await withGhSpawn(
      {
        "pr checks": { stdout: "[]", stderr: "", exitCode: 0 },
      },
      () =>
        invokeTool<{ status: string }>(
          createUfFinaleCiPollTool(mockExec, "/fake"),
          { prNumber: 42 },
        ),
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.status).toBe("no-checks-workflows-exist")
    }
  })

  it("keeps polling while a check buckets as pending, then resolves", async () => {
    vi.useFakeTimers()
    try {
      const mockExec: GitExec = () => ""
      mockMergeablePrView()
      const result = await withSequentialSpawn(
        [
          {
            stdout: JSON.stringify([
              { name: "CI", state: "PENDING", bucket: "pending", link: "https://ci/1" },
            ]),
            stderr: "",
            exitCode: 8,
          },
          {
            stdout: JSON.stringify([
              { name: "CI", state: "SUCCESS", bucket: "pass", link: "https://ci/1" },
            ]),
            stderr: "",
            exitCode: 0,
          },
        ],
        async () => {
          const promise = invokeTool<{
            status: string
            failedChecks: unknown[]
            mergeable: boolean
          }>(createUfFinaleCiPollTool(mockExec, "/fake"), { prNumber: 42 })
          await vi.advanceTimersByTimeAsync(30_000)
          const settled = await promise
          return settled
        },
      )

      expect(result.status).toBe("ok")
      if (result.status === "ok") {
        expect(result.data?.status).toBe("pass")
        expect(result.data?.mergeable).toBe(true)
      }
    } finally {
      vi.useRealTimers()
    }
  })

  it("returns CI_POLL_TIMEOUT when checks remain pending past timeout", async () => {
    vi.useFakeTimers()
    try {
      const mockExec: GitExec = () => ""
      mockMergeablePrView()
      const result = await withSequentialSpawn(
        [
          {
            stdout: JSON.stringify([
              { name: "CI", state: "PENDING", bucket: "pending", link: "https://ci/1" },
            ]),
            stderr: "",
            exitCode: 8,
          },
        ],
        async () => {
          const promise = invokeTool(
            createUfFinaleCiPollTool(mockExec, "/fake"),
            { prNumber: 42, timeoutMinutes: 1 },
          )
          await vi.advanceTimersByTimeAsync(30_000)
          await vi.advanceTimersByTimeAsync(30_000)
          return await promise
        },
      )

      expect(result.status).toBe("error")
      if (result.status === "error") {
        expect(result.code).toBe("CI_POLL_TIMEOUT")
        expect(result.retryable).toBe(true)
        expect(result.message).toContain("did not complete within 1 minutes")
      }
    } finally {
      vi.useRealTimers()
    }
  })

  it("classifies an aborted/hung invocation as GH_CLI_ERROR", async () => {
    const mockExec: GitExec = () => ""
    mockMergeablePrView()
    // Simulate the per-invocation AbortSignal firing without waiting 30s:
    // a real timeout rejects the exited wait with an AbortError.
    const spawn = vi.fn(
      (): MockSpawnProcess => ({
        ...mockProcess({ stdout: "", stderr: "", exitCode: 0 }),
        exited: Promise.reject(
          new DOMException("The operation was aborted", "AbortError"),
        ),
      }),
    )
    const restore = patchBunSpawn(spawn)
    try {
      const result = await invokeTool(
        createUfFinaleCiPollTool(mockExec, "/fake"),
        { prNumber: 42 },
      )

      expect(result.status).toBe("error")
      if (result.status === "error") {
        expect(result.code).toBe("GH_CLI_ERROR")
        expect(result.retryable).toBe(true)
        expect(result.message).toContain("aborted")
      }
    } finally {
      restore()
    }
  })

  it("returns CI_POLL_ERROR on unexpected error in outer try", async () => {
    const mockExec: GitExec = () => ""
    // Date.now() is called at startTime (before try) and at
    // while condition (inside outer try). Making the second
    // call throw exercises the outer catch.
    const originalDateNow = Date.now
    let callCount = 0
    Date.now = () => {
      callCount++
      if (callCount <= 1) return 1000
      throw new Error("unexpected clock failure")
    }

    try {
      vi.mocked(execSync).mockImplementation(() => {
        throw new Error("skip")
      })

      const result = await invokeTool(
        createUfFinaleCiPollTool(mockExec, "/fake"),
        { prNumber: 42 },
      )

      expect(result.status).toBe("error")
      if (result.status === "error") {
        expect(result.code).toBe("CI_POLL_ERROR")
        expect(result.retryable).toBe(false)
        expect(result.message).toContain("unexpected clock failure")
      }
    } finally {
      Date.now = originalDateNow
    }
  })

  it("returns CI_POLL_ERROR with stringified non-Error in outer catch", async () => {
    const mockExec: GitExec = () => ""
    const originalDateNow = Date.now
    let callCount = 0
    Date.now = () => {
      callCount++
      if (callCount <= 1) return 1000
      // Throw a non-Error to exercise the String(err) branch
      throw "clock string error" // eslint-disable-line no-throw-literal
    }

    try {
      vi.mocked(execSync).mockImplementation(() => {
        throw new Error("skip")
      })

      const result = await invokeTool(
        createUfFinaleCiPollTool(mockExec, "/fake"),
        { prNumber: 42 },
      )

      expect(result.status).toBe("error")
      if (result.status === "error") {
        expect(result.code).toBe("CI_POLL_ERROR")
        expect(result.retryable).toBe(false)
        expect(result.message).toBe("clock string error")
      }
    } finally {
      Date.now = originalDateNow
    }
  })

  it("returns GH_CLI_ERROR with stringified non-Error poll error", async () => {
    const mockExec: GitExec = () => ""
    vi.mocked(execSync).mockImplementation((cmd) => {
      if (String(cmd).includes("gh pr view")) {
        throw new Error("skip merge check")
      }
      throw new Error(`unexpected shell invocation: ${String(cmd)}`)
    })
    const spawn = vi.fn((): never => {
      // Throw a non-Error to exercise the String(pollErr) branch
      throw "non-error poll failure" // eslint-disable-line no-throw-literal
    })
    const restore = patchBunSpawn(spawn)
    try {
      const result = await invokeTool(
        createUfFinaleCiPollTool(mockExec, "/fake"),
        { prNumber: 42 },
      )

      expect(result.status).toBe("error")
      if (result.status === "error") {
        expect(result.code).toBe("GH_CLI_ERROR")
        expect(result.retryable).toBe(true)
        expect(result.message).toContain("non-error poll failure")
      }
    } finally {
      restore()
    }
  })

  it("returns INVALID_PR_NUMBER for non-finite prNumber", async () => {
    const mockExec: GitExec = () => ""
    const result = await invokeTool(
      createUfFinaleCiPollTool(mockExec, "/fake"),
      { prNumber: NaN },
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_PR_NUMBER")
      expect(result.message).toBe("Invalid PR number")
    }
  })

  it("returns INVALID_PR_NUMBER for prNumber exceeding max", async () => {
    const mockExec: GitExec = () => ""
    const result = await invokeTool(
      createUfFinaleCiPollTool(mockExec, "/fake"),
      { prNumber: 1000000 },
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_PR_NUMBER")
    }
  })

  it("returns INVALID_PR_NUMBER for prNumber below 1", async () => {
    const mockExec: GitExec = () => ""
    const result = await invokeTool(
      createUfFinaleCiPollTool(mockExec, "/fake"),
      { prNumber: 0 },
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_PR_NUMBER")
    }
  })
})

// ---------------------------------------------------------------------------
// uf_finale_return_to_main tests
// ---------------------------------------------------------------------------

describe("uf_finale_return_to_main", () => {
  it("checks out main and pulls successfully", async () => {
    const commands: string[] = []
    const mockExec: GitExec = (args: string[]) => {
      commands.push(args.join(" "))
      return ""
    }
    const result = await invokeTool<{ success: boolean }>(
      createUfFinaleReturnToMainTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.success).toBe(true)
    }
    expect(commands).toContain("checkout main")
    expect(commands).toContain("pull")
  })

  it("returns failure when checkout fails", async () => {
    const mockExec: GitExec = () => {
      throw new Error("error: pathspec 'main' did not match")
    }
    const result = await invokeTool(
      createUfFinaleReturnToMainTool(mockExec, "/fake"),
      {},
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("RETURN_TO_MAIN_ERROR")
      expect(result.message).toContain("pathspec")
    }
  })

  it("returns failure with stringified non-Error value", async () => {
    const mockExec: GitExec = () => {
      throw "string error from git" // eslint-disable-line no-throw-literal
    }
    const result = await invokeTool(
      createUfFinaleReturnToMainTool(mockExec, "/fake"),
      {},
    )
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("RETURN_TO_MAIN_ERROR")
      expect(result.message).toBe("string error from git")
    }
  })
})

// ---------------------------------------------------------------------------
// Terminal-write regression tests (Requirement 1, TC-006)
// ---------------------------------------------------------------------------

describe("terminal-write regression (Requirement 1)", () => {
  const throwingGitExec: GitExec = () => {
    throw new Error("simulated subprocess failure")
  }

  /**
   * Run the given invocations with console methods and stdio streams
   * spied and assert that no terminal writes occurred.
   */
  async function expectNoTerminalWrites(
    run: () => Promise<void>,
  ): Promise<void> {
    const spies = [
      vi.spyOn(console, "log").mockImplementation(() => {}),
      vi.spyOn(console, "error").mockImplementation(() => {}),
      vi.spyOn(console, "info").mockImplementation(() => {}),
      vi.spyOn(console, "warn").mockImplementation(() => {}),
      vi.spyOn(console, "debug").mockImplementation(() => {}),
      vi.spyOn(console, "trace").mockImplementation(() => {}),
      vi.spyOn(process.stdout, "write").mockImplementation(() => true),
      vi.spyOn(process.stderr, "write").mockImplementation(() => true),
    ]
    try {
      await run()
      for (const spy of spies) {
        expect(spy).not.toHaveBeenCalled()
      }
    } finally {
      for (const spy of spies) {
        spy.mockRestore()
      }
    }
  }

  it("uf_finale_startup writes nothing to the terminal", async () => {
    const successExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref HEAD")) return "feature/x"
      if (cmd.includes("--porcelain")) return "?? .env\n M src/app.ts"
      if (cmd.includes("@{u}..HEAD")) return "abc1234 first commit"
      return ""
    }
    await expectNoTerminalWrites(async () => {
      await invokeTool(createUfFinaleStartupTool(successExec, "/fake"), {})
      await invokeTool(createUfFinaleStartupTool(throwingGitExec, "/fake"), {})
    })
  })

  it("uf_finale_stage writes nothing to the terminal", async () => {
    const successExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--cached --name-only")) return "file1.ts\nfile2.ts"
      return ""
    }
    await expectNoTerminalWrites(async () => {
      await invokeTool(createUfFinaleStageTool(successExec, "/fake"), {})
      await invokeTool(createUfFinaleStageTool(throwingGitExec, "/fake"), {})
    })
  })

  it("uf_finale_commit_context writes nothing to the terminal", async () => {
    const successExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--cached --stat"))
        return " file1.ts | 10 ++++++++++\n 1 file changed, 10 insertions(+)"
      if (cmd.includes("log -5")) return "abc1234 feat: add stuff"
      return ""
    }
    await expectNoTerminalWrites(async () => {
      await invokeTool(createUfFinaleCommitContextTool(successExec, "/fake"), {
        rawModelId: "claude-sonnet-4",
      })
      await invokeTool(
        createUfFinaleCommitContextTool(throwingGitExec, "/fake"),
        { rawModelId: "claude-sonnet-4" },
      )
    })
  })

  it("uf_finale_push_state writes nothing to the terminal", async () => {
    const successExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref @{u}")) return "origin/feature"
      if (cmd.includes("@{u}..HEAD --count")) return "2"
      if (cmd.includes("HEAD..@{u} --count")) return "1"
      return ""
    }
    await expectNoTerminalWrites(async () => {
      await invokeTool(createUfFinalePushStateTool(successExec, "/fake"), {})
      await invokeTool(
        createUfFinalePushStateTool(throwingGitExec, "/fake"),
        {},
      )
    })
  })

  it("uf_finale_pr_context writes nothing to the terminal", async () => {
    const richExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("merge-base"))
        return "abc123def456abc123def456abc123def456abc1"
      if (cmd.includes("log --oneline")) return "def456 feat: stuff"
      if (cmd.includes("diff --stat")) return " file.ts | 5 +++++\n 1 file changed"
      if (cmd.includes("--abbrev-ref HEAD")) return "feature/x"
      throw new Error("skip")
    }
    await expectNoTerminalWrites(async () => {
      vi.mocked(execSync).mockImplementation((cmd) => {
        const cmdStr = String(cmd)
        if (cmdStr.includes("gh pr view --json number,url,title")) {
          return JSON.stringify({
            number: 42,
            url: "https://github.com/owner/repo/pull/42",
            title: "My PR",
          })
        }
        if (cmdStr.includes("--json isFork --jq")) return "true"
        if (cmdStr.includes("--json parent")) {
          return JSON.stringify({
            parent: { owner: { login: "upstream-org" }, name: "upstream-repo" },
          })
        }
        throw new Error(`unexpected: ${cmdStr}`)
      })
      vi.mocked(readFileSync).mockImplementation((pathArg) => {
        const p = String(pathArg)
        if (p.includes("PULL_REQUEST_TEMPLATE.md")) {
          return "## Summary\nDescribe changes.\n\n## Testing\nHow tested."
        }
        if (p.includes(".openspec.yaml")) {
          return "name: my-change\noriginating_issue: 123\n"
        }
        if (p.includes("unleash-summary")) {
          return JSON.stringify({ steps: ["s1"], verdict: "APPROVE" })
        }
        throw new Error("ENOENT")
      })
      await invokeTool(createUfFinalePrContextTool(richExec, "/fake"), {})
      await invokeTool(createUfFinalePrContextTool(throwingGitExec, "/fake"), {})
    })
  })

  it("uf_finale_create_pr writes nothing to the terminal", async () => {
    const successExec: GitExec = () => ""
    await expectNoTerminalWrites(async () => {
      vi.mocked(mkdtempSync).mockReturnValue("/tmp/uf-finale-regression")
      vi.mocked(writeFileSync).mockImplementation(() => {})
      vi.mocked(rmSync).mockImplementation(() => {})
      vi.mocked(execSync)
        .mockReturnValueOnce("https://github.com/owner/repo/pull/42\n")
        .mockImplementationOnce(() => {
          throw new Error("gh: not authenticated")
        })
      await invokeTool(createUfFinaleCreatePrTool(successExec, "/fake"), {
        title: "My PR title",
        body: "PR body content",
      })
      await invokeTool(createUfFinaleCreatePrTool(successExec, "/fake"), {
        title: "My PR title",
        body: "PR body content",
      })
    })
  })

  it("uf_finale_ci_poll writes nothing to the terminal (pass/fail/error)", async () => {
    const mockExec: GitExec = () => ""
    vi.mocked(execSync).mockImplementation((cmd) => {
      if (String(cmd).includes("gh pr view")) {
        return JSON.stringify({
          mergeable: "MERGEABLE",
          mergeStateStatus: "CLEAN",
        })
      }
      throw new Error(`unexpected shell invocation: ${String(cmd)}`)
    })

    await expectNoTerminalWrites(async () => {
      await withGhSpawn(
        {
          "pr checks": {
            stdout: JSON.stringify([
              { name: "CI", state: "SUCCESS", bucket: "pass", link: "https://ci/1" },
            ]),
            stderr: "",
            exitCode: 0,
          },
        },
        () =>
          invokeTool(createUfFinaleCiPollTool(mockExec, "/fake"), {
            prNumber: 42,
          }),
      )

      await withGhSpawn(
        {
          "pr checks": {
            stdout: JSON.stringify([
              { name: "Lint", state: "FAILURE", bucket: "fail", link: "https://ci/lint" },
            ]),
            stderr: "",
            exitCode: 1,
          },
        },
        () =>
          invokeTool(createUfFinaleCiPollTool(mockExec, "/fake"), {
            prNumber: 42,
          }),
      )

      await withGhSpawn(
        {
          "pr checks": {
            stdout: "not a json array",
            stderr: "fatal: HTTP 503 Service Unavailable",
            exitCode: 1,
          },
        },
        () =>
          invokeTool(createUfFinaleCiPollTool(mockExec, "/fake"), {
            prNumber: 42,
          }),
      )
    })
  })

  it("uf_finale_ci_poll writes nothing to the terminal (pending/timeout)", async () => {
    const mockExec: GitExec = () => ""
    vi.mocked(execSync).mockImplementation((cmd) => {
      if (String(cmd).includes("gh pr view")) {
        return JSON.stringify({
          mergeable: "MERGEABLE",
          mergeStateStatus: "CLEAN",
        })
      }
      throw new Error(`unexpected shell invocation: ${String(cmd)}`)
    })

    await expectNoTerminalWrites(async () => {
      vi.useFakeTimers()
      try {
        await withSequentialSpawn(
          [
            {
              stdout: JSON.stringify([
                { name: "CI", state: "PENDING", bucket: "pending", link: "https://ci/1" },
              ]),
              stderr: "",
              exitCode: 8,
            },
            {
              stdout: JSON.stringify([
                { name: "CI", state: "SUCCESS", bucket: "pass", link: "https://ci/1" },
              ]),
              stderr: "",
              exitCode: 0,
            },
          ],
          async () => {
            const pendingPromise = invokeTool(
              createUfFinaleCiPollTool(mockExec, "/fake"),
              { prNumber: 42 },
            )
            await vi.advanceTimersByTimeAsync(30_000)
            await pendingPromise
          },
        )

        // Timeout path: checks remain pending past the configured limit
        await withSequentialSpawn(
          [
            {
              stdout: JSON.stringify([
                { name: "CI", state: "PENDING", bucket: "pending", link: "https://ci/1" },
              ]),
              stderr: "",
              exitCode: 8,
            },
          ],
          async () => {
            const timeoutPromise = invokeTool(
              createUfFinaleCiPollTool(mockExec, "/fake"),
              { prNumber: 42, timeoutMinutes: 1 },
            )
            await vi.advanceTimersByTimeAsync(30_000)
            await vi.advanceTimersByTimeAsync(30_000)
            await timeoutPromise
          },
        )
      } finally {
        vi.useRealTimers()
      }
    })
  })

  it("uf_finale_return_to_main writes nothing to the terminal", async () => {
    const successExec: GitExec = () => ""
    await expectNoTerminalWrites(async () => {
      await invokeTool(createUfFinaleReturnToMainTool(successExec, "/fake"), {})
      await invokeTool(
        createUfFinaleReturnToMainTool(throwingGitExec, "/fake"),
        {},
      )
    })
  })

  it("uf_finale_base_state writes nothing to the terminal", async () => {
    const successExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd === "remote") return "origin\nupstream\n"
      if (cmd.includes("symbolic-ref refs/remotes/upstream/HEAD"))
        return "refs/remotes/upstream/main\n"
      if (cmd.includes("fetch upstream main")) return ""
      if (cmd.includes("upstream/main..HEAD --count")) return "4"
      if (cmd.includes("HEAD..upstream/main --count")) return "1"
      return ""
    }
    await expectNoTerminalWrites(async () => {
      await invokeTool(createUfFinaleBaseStateTool(successExec, "/fake"), {})
      await invokeTool(
        createUfFinaleBaseStateTool(throwingGitExec, "/fake"),
        {},
      )
    })
  })
})
