import { describe, it, expect, vi, beforeEach } from "vitest"
import { execSync } from "node:child_process"
import {
  readFileSync,
  existsSync,
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
  createUfFinalePushStateTool,
  createUfFinalePrContextTool,
  createUfFinaleCreatePrTool,
  createUfFinaleCiPollTool,
  createUfFinaleReturnToMainTool,
  type GitExec,
} from "../plugins/uf-finale/index.js"
import UfFinalePlugin from "../plugins/uf-finale/index.js"

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

  it("server registers all 8 finale tools", async () => {
    const fakeInput = {
      directory: "/tmp/fake",
      worktree: "/tmp/fake",
      client: {} as PluginInput["client"],
    } as PluginInput

    const result = await UfFinalePlugin.server(fakeInput)
    expect(result).toBeDefined()
    expect(result.tool).toBeDefined()
    expect(Object.keys(result.tool).sort()).toEqual([
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
// Tool definition args schema exposure (FR-004)
// ---------------------------------------------------------------------------

describe("uf_finale tool definition args schema exposure", () => {
  const mockExec: GitExec = () => ""

  it("exposes uf_finale_commit_context tool definition args: rawModelId", () => {
    const definition = createUfFinaleCommitContextTool(mockExec, "/fake")

    expect(Object.keys(definition.args).sort()).toEqual(["rawModelId"])
  })

  it("exposes uf_finale_create_pr tool definition args: title, body, targetRepo", () => {
    const definition = createUfFinaleCreatePrTool(mockExec, "/fake")

    expect(Object.keys(definition.args).sort()).toEqual([
      "body",
      "targetRepo",
      "title",
    ])
  })

  it("exposes uf_finale_ci_poll tool definition args: prNumber, timeoutMinutes", () => {
    const definition = createUfFinaleCiPollTool(mockExec, "/fake")

    expect(Object.keys(definition.args).sort()).toEqual([
      "prNumber",
      "timeoutMinutes",
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

  it("reads unleash summary artifact", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("merge-base")) throw new Error("skip")
      if (cmd.includes("--abbrev-ref HEAD")) return "opsx/my-change"
      throw new Error("skip")
    }
    vi.mocked(execSync).mockImplementation(() => {
      throw new Error("skip gh")
    })
    vi.mocked(readFileSync).mockImplementation((pathArg) => {
      const p = String(pathArg)
      if (p.includes("unleash-summary")) {
        return JSON.stringify({ steps: ["s1", "s2"], verdict: "APPROVE" })
      }
      throw new Error("ENOENT")
    })

    const result = await invokeTool<{
      unleashSummary: Record<string, unknown> | null
    }>(createUfFinalePrContextTool(mockExec, "/fake"), {})

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.unleashSummary).toStrictEqual({
        steps: ["s1", "s2"],
        verdict: "APPROVE",
      })
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
  it("returns pass when all checks succeed", async () => {
    const mockExec: GitExec = () => ""
    vi.mocked(execSync).mockImplementation((cmd) => {
      const cmdStr = String(cmd)
      if (cmdStr.includes("gh pr view")) {
        return JSON.stringify({
          mergeable: "MERGEABLE",
          mergeStateStatus: "CLEAN",
        })
      }
      if (cmdStr.includes("gh pr checks")) {
        return JSON.stringify([
          {
            name: "CI",
            state: "COMPLETED",
            conclusion: "SUCCESS",
            detailsUrl: "https://ci/1",
          },
          {
            name: "Lint",
            state: "COMPLETED",
            conclusion: "SUCCESS",
            detailsUrl: "https://ci/2",
          },
        ])
      }
      throw new Error(`unexpected: ${cmdStr}`)
    })

    const result = await invokeTool<{
      status: string
      failedChecks: unknown[]
      mergeable: boolean
      workflowsConfigured: boolean
    }>(createUfFinaleCiPollTool(mockExec, "/fake"), { prNumber: 42 })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.status).toBe("pass")
      expect(result.data?.failedChecks).toStrictEqual([])
      expect(result.data?.mergeable).toBe(true)
      expect(result.data?.workflowsConfigured).toBe(true)
    }
  })

  it("returns fail when checks have failures", async () => {
    const mockExec: GitExec = () => ""
    vi.mocked(execSync).mockImplementation((cmd) => {
      const cmdStr = String(cmd)
      if (cmdStr.includes("gh pr view")) {
        return JSON.stringify({
          mergeable: "MERGEABLE",
          mergeStateStatus: "CLEAN",
        })
      }
      if (cmdStr.includes("gh pr checks")) {
        return JSON.stringify([
          {
            name: "CI",
            state: "COMPLETED",
            conclusion: "SUCCESS",
            detailsUrl: "https://ci/1",
          },
          {
            name: "Lint",
            state: "COMPLETED",
            conclusion: "FAILURE",
            detailsUrl: "https://ci/lint",
          },
        ])
      }
      throw new Error("unexpected")
    })

    const result = await invokeTool<{
      status: string
      failedChecks: Array<{
        name: string
        conclusion: string
        url: string
      }>
      mergeable: boolean
    }>(createUfFinaleCiPollTool(mockExec, "/fake"), { prNumber: 42 })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.status).toBe("fail")
      expect(result.data?.mergeable).toBe(false)
      expect(result.data?.failedChecks).toHaveLength(1)
      expect(result.data?.failedChecks[0].name).toBe("Lint")
      expect(result.data?.failedChecks[0].conclusion).toBe("FAILURE")
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

  it("returns GH_CLI_ERROR when gh CLI fails during poll", async () => {
    const mockExec: GitExec = () => ""
    // Both calls throw — merge check is caught, poll check returns error
    vi.mocked(execSync).mockImplementation(() => {
      throw new Error("gh: command not found")
    })

    const result = await invokeTool(
      createUfFinaleCiPollTool(mockExec, "/fake"),
      { prNumber: 42 },
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("GH_CLI_ERROR")
      expect(result.retryable).toBe(true)
      expect(result.message).toContain("gh: command not found")
    }
  })

  it("detects no-checks with workflows existing", async () => {
    const mockExec: GitExec = () => ""
    vi.mocked(execSync).mockImplementation((cmd) => {
      const cmdStr = String(cmd)
      if (cmdStr.includes("gh pr view")) {
        return JSON.stringify({
          mergeable: "MERGEABLE",
          mergeStateStatus: "CLEAN",
        })
      }
      if (cmdStr.includes("gh pr checks")) {
        return "[]"
      }
      throw new Error("unexpected")
    })
    vi.mocked(existsSync).mockReturnValue(true)

    const result = await invokeTool<{
      status: string
      workflowsConfigured: boolean
      skipReason: string
    }>(createUfFinaleCiPollTool(mockExec, "/fake"), { prNumber: 42 })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.status).toBe("no-checks-workflows-exist")
      expect(result.data?.workflowsConfigured).toBe(true)
      expect(result.data?.skipReason).toBeDefined()
    }
  })

  it("detects no-checks with no workflows directory", async () => {
    const mockExec: GitExec = () => ""
    vi.mocked(execSync).mockImplementation((cmd) => {
      const cmdStr = String(cmd)
      if (cmdStr.includes("gh pr view")) {
        return JSON.stringify({
          mergeable: "MERGEABLE",
          mergeStateStatus: "CLEAN",
        })
      }
      if (cmdStr.includes("gh pr checks")) {
        return "[]"
      }
      throw new Error("unexpected")
    })
    vi.mocked(existsSync).mockReturnValue(false)

    const result = await invokeTool<{
      status: string
      workflowsConfigured: boolean
    }>(createUfFinaleCiPollTool(mockExec, "/fake"), { prNumber: 42 })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.status).toBe("no-checks-no-workflows")
      expect(result.data?.workflowsConfigured).toBe(false)
    }
  })

  it("treats NEUTRAL and SKIPPED conclusions as passing", async () => {
    const mockExec: GitExec = () => ""
    vi.mocked(execSync).mockImplementation((cmd) => {
      const cmdStr = String(cmd)
      if (cmdStr.includes("gh pr view")) {
        return JSON.stringify({
          mergeable: "MERGEABLE",
          mergeStateStatus: "CLEAN",
        })
      }
      if (cmdStr.includes("gh pr checks")) {
        return JSON.stringify([
          {
            name: "CI",
            state: "COMPLETED",
            conclusion: "SUCCESS",
            detailsUrl: "https://ci/1",
          },
          {
            name: "Optional",
            state: "COMPLETED",
            conclusion: "NEUTRAL",
            detailsUrl: "https://ci/2",
          },
          {
            name: "Skipped",
            state: "COMPLETED",
            conclusion: "SKIPPED",
            detailsUrl: "https://ci/3",
          },
        ])
      }
      throw new Error("unexpected")
    })

    const result = await invokeTool<{
      status: string
      failedChecks: unknown[]
    }>(createUfFinaleCiPollTool(mockExec, "/fake"), { prNumber: 42 })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.status).toBe("pass")
      expect(result.data?.failedChecks).toStrictEqual([])
    }
  })

  it("logs pending progress and resolves on next poll", async () => {
    vi.useFakeTimers()
    try {
      const mockExec: GitExec = () => ""
      let checkCallCount = 0
      vi.mocked(execSync).mockImplementation((cmd) => {
        const cmdStr = String(cmd)
        if (cmdStr.includes("gh pr view")) {
          return JSON.stringify({
            mergeable: "MERGEABLE",
            mergeStateStatus: "CLEAN",
          })
        }
        if (cmdStr.includes("gh pr checks")) {
          checkCallCount++
          if (checkCallCount === 1) {
            return JSON.stringify([
              {
                name: "CI",
                state: "IN_PROGRESS",
                conclusion: "",
                detailsUrl: "https://ci/1",
              },
            ])
          }
          return JSON.stringify([
            {
              name: "CI",
              state: "COMPLETED",
              conclusion: "SUCCESS",
              detailsUrl: "https://ci/1",
            },
          ])
        }
        throw new Error(`unexpected: ${cmdStr}`)
      })

      const promise = invokeTool<{
        status: string
        failedChecks: unknown[]
        mergeable: boolean
      }>(createUfFinaleCiPollTool(mockExec, "/fake"), { prNumber: 42 })

      // Advance past the 30s poll interval to resolve the pending wait
      await vi.advanceTimersByTimeAsync(30_000)

      const result = await promise
      expect(result.status).toBe("ok")
      if (result.status === "ok") {
        expect(result.data?.status).toBe("pass")
        expect(result.data?.mergeable).toBe(true)
      }
      expect(checkCallCount).toBe(2)
    } finally {
      vi.useRealTimers()
    }
  })

  it("returns CI_POLL_TIMEOUT when checks remain pending past timeout", async () => {
    vi.useFakeTimers()
    try {
      const mockExec: GitExec = () => ""
      vi.mocked(execSync).mockImplementation((cmd) => {
        const cmdStr = String(cmd)
        if (cmdStr.includes("gh pr view")) {
          return JSON.stringify({
            mergeable: "MERGEABLE",
            mergeStateStatus: "CLEAN",
          })
        }
        if (cmdStr.includes("gh pr checks")) {
          return JSON.stringify([
            {
              name: "CI",
              state: "PENDING",
              conclusion: "",
              detailsUrl: "https://ci/1",
            },
          ])
        }
        throw new Error(`unexpected: ${cmdStr}`)
      })

      // Use 1-minute timeout to keep test fast
      const promise = invokeTool(
        createUfFinaleCiPollTool(mockExec, "/fake"),
        { prNumber: 42, timeoutMinutes: 1 },
      )

      // Each iteration waits 30s. Two iterations (60s total) exceed the
      // 1-minute timeout, causing the while loop to exit.
      await vi.advanceTimersByTimeAsync(30_000)
      await vi.advanceTimersByTimeAsync(30_000)

      const result = await promise
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
      const cmdStr = String(cmd)
      if (cmdStr.includes("gh pr view")) {
        throw new Error("skip merge check")
      }
      if (cmdStr.includes("gh pr checks")) {
        // Throw a non-Error to exercise the String(pollErr) branch
        throw "non-error poll failure" // eslint-disable-line no-throw-literal
      }
      throw new Error("unexpected")
    })

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
    await expectNoTerminalWrites(async () => {
      vi.mocked(execSync).mockImplementation((cmd) => {
        const cmdStr = String(cmd)
        if (cmdStr.includes("gh pr view")) {
          return JSON.stringify({
            mergeable: "MERGEABLE",
            mergeStateStatus: "CLEAN",
          })
        }
        if (cmdStr.includes("gh pr checks")) {
          return JSON.stringify([
            {
              name: "CI",
              state: "COMPLETED",
              conclusion: "SUCCESS",
              detailsUrl: "https://ci/1",
            },
          ])
        }
        throw new Error(`unexpected: ${cmdStr}`)
      })
      await invokeTool(createUfFinaleCiPollTool(mockExec, "/fake"), {
        prNumber: 42,
      })

      // Fail path: a check with a FAILURE conclusion
      vi.mocked(execSync).mockImplementation((cmd) => {
        const cmdStr = String(cmd)
        if (cmdStr.includes("gh pr view")) {
          return JSON.stringify({
            mergeable: "MERGEABLE",
            mergeStateStatus: "CLEAN",
          })
        }
        if (cmdStr.includes("gh pr checks")) {
          return JSON.stringify([
            {
              name: "Lint",
              state: "COMPLETED",
              conclusion: "FAILURE",
              detailsUrl: "https://ci/lint",
            },
          ])
        }
        throw new Error(`unexpected: ${cmdStr}`)
      })
      await invokeTool(createUfFinaleCiPollTool(mockExec, "/fake"), {
        prNumber: 42,
      })

      // Error path: gh CLI failure during poll
      vi.mocked(execSync).mockImplementation(() => {
        throw new Error("gh: command not found")
      })
      await invokeTool(createUfFinaleCiPollTool(mockExec, "/fake"), {
        prNumber: 42,
      })
    })
  })

  it("uf_finale_ci_poll writes nothing to the terminal (pending/timeout)", async () => {
    const mockExec: GitExec = () => ""
    await expectNoTerminalWrites(async () => {
      vi.useFakeTimers()
      try {
        let checkCallCount = 0
        vi.mocked(execSync).mockImplementation((cmd) => {
          const cmdStr = String(cmd)
          if (cmdStr.includes("gh pr view")) {
            return JSON.stringify({
              mergeable: "MERGEABLE",
              mergeStateStatus: "CLEAN",
            })
          }
          if (cmdStr.includes("gh pr checks")) {
            checkCallCount++
            if (checkCallCount === 1) {
              return JSON.stringify([
                {
                  name: "CI",
                  state: "IN_PROGRESS",
                  conclusion: "",
                  detailsUrl: "https://ci/1",
                },
              ])
            }
            return JSON.stringify([
              {
                name: "CI",
                state: "COMPLETED",
                conclusion: "SUCCESS",
                detailsUrl: "https://ci/1",
              },
            ])
          }
          throw new Error(`unexpected: ${cmdStr}`)
        })

        const pendingPromise = invokeTool(
          createUfFinaleCiPollTool(mockExec, "/fake"),
          { prNumber: 42 },
        )
        await vi.advanceTimersByTimeAsync(30_000)
        await pendingPromise

        // Timeout path: checks remain pending past the configured limit
        vi.mocked(execSync).mockImplementation((cmd) => {
          const cmdStr = String(cmd)
          if (cmdStr.includes("gh pr view")) {
            return JSON.stringify({
              mergeable: "MERGEABLE",
              mergeStateStatus: "CLEAN",
            })
          }
          if (cmdStr.includes("gh pr checks")) {
            return JSON.stringify([
              {
                name: "CI",
                state: "PENDING",
                conclusion: "",
                detailsUrl: "https://ci/1",
              },
            ])
          }
          throw new Error(`unexpected: ${cmdStr}`)
        })

        const timeoutPromise = invokeTool(
          createUfFinaleCiPollTool(mockExec, "/fake"),
          { prNumber: 42, timeoutMinutes: 1 },
        )
        await vi.advanceTimersByTimeAsync(30_000)
        await vi.advanceTimersByTimeAsync(30_000)
        await timeoutPromise
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
})
