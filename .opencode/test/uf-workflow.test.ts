import {
  existsSync,
  mkdirSync,
  mkdtempSync,
  readdirSync,
  readFileSync,
  rmSync,
  statSync,
  symlinkSync,
  writeFileSync,
} from "node:fs"
import { tmpdir } from "node:os"
import { join } from "node:path"

import type { PluginInput } from "@opencode-ai/plugin"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"

import {
  failure,
  success,
  type ToolFailure,
  type ToolResult,
  type ToolSuccess,
} from "../lib/uf-workflow-types.js"
import UfWorkflowPlugin from "../plugins/uf-workflow/index.js"
import { withBun } from "./helpers/bun-runtime.js"
import {
  createCheckpointTool,
  createClearProgressTool,
  createGitCommitNoEditTool,
  createGitConflictFilesTool,
  createGitFetchTool,
  createGitMergeAbortTool,
  createGitMergeTool,
  createGitPushForceWithLeaseTool,
  createGitPushTool,
  createGitRebaseAbortTool,
  createGitRebaseTool,
  createGetProgressTool,
  createNormalizeModelNameTool,
  createPersistUnleashSummaryTool,
  createResolveBaseRefTool,
  type GitExec,
  type ProgressState,
  type ResolveBaseRefData,
  resolveKey,
  sanitizeSlug,
  type UnleashSummary,
} from "../plugins/uf-workflow/index.js"

// ---------------------------------------------------------------------------
// Envelope helper tests
// ---------------------------------------------------------------------------

describe("success()", () => {
  it("returns { status: 'ok' } without data", () => {
    const result: ToolSuccess = success()
    expect(result).toStrictEqual({ status: "ok" })
    expect(result.data).toBeUndefined()
  })

  it("includes data when provided", () => {
    const result: ToolSuccess<{ count: number }> = success({ count: 42 })
    expect(result).toStrictEqual({ status: "ok", data: { count: 42 } })
  })

  it("carries typed payload through ToolResult union", () => {
    const result: ToolResult<string> = success("hello")
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data).toBe("hello")
    }
  })
})

describe("failure()", () => {
  it("returns error envelope with retryable defaulting to false", () => {
    const result: ToolFailure = failure("something went wrong")
    expect(result).toStrictEqual({
      status: "error",
      message: "something went wrong",
      retryable: false,
    })
  })

  it("includes retryable and code when provided", () => {
    const result: ToolFailure = failure("timeout", {
      retryable: true,
      code: "ETIMEOUT",
    })
    expect(result).toStrictEqual({
      status: "error",
      message: "timeout",
      retryable: true,
      code: "ETIMEOUT",
    })
  })

  it("omits code when not provided", () => {
    const result = failure("oops", { retryable: true })
    expect(result.code).toBeUndefined()
    expect(result.retryable).toBe(true)
  })

  it("discriminates via status in ToolResult union", () => {
    const result: ToolResult = failure("bad")
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.message).toBe("bad")
      expect(result.retryable).toBe(false)
    }
  })
})

// ---------------------------------------------------------------------------
// Plugin smoke test
// ---------------------------------------------------------------------------

describe("UfWorkflowPlugin", () => {
  it("exports a PluginModule with id and server", () => {
    expect(UfWorkflowPlugin.id).toBe("uf-workflow")
    expect(typeof UfWorkflowPlugin.server).toBe("function")
  })

  it("server resolves and registers resolve_base_ref tool", async () => {
    const fakeInput = {
      directory: "/tmp/fake",
      client: {} as PluginInput["client"],
    } as PluginInput

    // withBun: the model-matrix selection wiring (model-matrix-delegation
    // task 5.2) constructs a Bun-backed YAML parser at registration, the
    // invoke-agent/review-dispatch server() convention.
    const result = await withBun(() => UfWorkflowPlugin.server(fakeInput))
    expect(result).toBeDefined()
    expect(result.tool).toBeDefined()
    expect(Object.keys(result.tool).sort()).toEqual([
      "checkpoint",
      "clear_progress",
      "format_unleash_demo",
      "format_unleash_exit",
      "get_progress",
      "git_commit_no_edit",
      "git_conflict_files",
      "git_fetch",
      "git_merge",
      "git_merge_abort",
      "git_push",
      "git_push_force_with_lease",
      "git_rebase",
      "git_rebase_abort",
      "mark_task_complete",
      "normalize_model_name",
      "parse_tasks",
      "resolve_base_ref",
      "resolve_clarification",
      "scan_clarifications",
      "scan_conflict_markers",
      "uf_delegate_command",
      "uf_unleash_branch_gate",
      "uf_unleash_detect_resume",
      "uf_unleash_resolve_issue",
      "uf_unleash_route_decision",
      "write_unleash_summary",
    ])
  })

  it("registers uf_delegate_command with the documented argument contract", async () => {
    const fakeInput = {
      directory: "/tmp/fake",
      client: {} as PluginInput["client"],
    } as PluginInput

    const result = await withBun(() => UfWorkflowPlugin.server(fakeInput))

    // DE-007 Layer 1 (delegation-skill-enforcement task 4.2): the
    // registered args must match the documented contract — the eight
    // pre-change parameters plus task_classification, skill, and
    // skill_absent_reason (task 1.1), plus the four model-matrix selection
    // parameters model, variant, profile, and pack (model-matrix-
    // delegation task 5.1, UF-FR-002). The registration-surface suite
    // proves args is the zod raw shape the host validates input through;
    // this pin holds the key set itself to the documented contract, so a
    // parameter added to the input schema but not the registration
    // surface (or vice versa) fails here.
    expect(Object.keys(result.tool.uf_delegate_command.args).sort()).toEqual([
      "agent",
      "checkpoint_step",
      "command_path",
      "expected_artifact",
      "feature_dir",
      "model",
      "pack",
      "phase",
      "profile",
      "prompt",
      "skill",
      "skill_absent_reason",
      "task_classification",
      "timeout",
      "variant",
    ])
  })
})

// ---------------------------------------------------------------------------
// resolve_base_ref tool tests
// ---------------------------------------------------------------------------

/** Helper to invoke the tool and parse its JSON output. */
async function invokeResolveBaseRef(exec: GitExec): Promise<ToolResult<ResolveBaseRefData>> {
  const resolveBaseRefTool = createResolveBaseRefTool(exec, "/fake/project")
  // The tool() return type has an execute method; invoke it with
  // empty args and a minimal context.
  const result = await (resolveBaseRefTool as { execute: (args: Record<string, never>, ctx: { sessionID: string }) => Promise<{ output: string }> })
    .execute({}, { sessionID: "test-session" })
  return JSON.parse(result.output) as ToolResult<ResolveBaseRefData>
}

describe("resolve_base_ref", () => {
  it("returns upstream/main when it resolves", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.startsWith("merge-base")) return ""
      if (cmd.includes("rev-parse") && cmd.includes("HEAD")) return "head456def789"
      if (cmd.includes("upstream/main")) return "abc123def456"
      throw new Error("not found")
    }

    const result = await invokeResolveBaseRef(mockExec)
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data).toStrictEqual({
        ref: "upstream/main",
        sha: "abc123def456",
        source: "upstream/main",
        head_sha: "head456def789",
        diverged: false,
      })
    }
  })

  it("falls back to origin/main when upstream/main fails", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.startsWith("merge-base")) return ""
      if (cmd.includes("rev-parse") && cmd.includes("HEAD")) return "head456def789"
      if (cmd.includes("upstream/main")) throw new Error("not found")
      if (cmd.includes("origin/main")) return "def789abc012"
      throw new Error("not found")
    }

    const result = await invokeResolveBaseRef(mockExec)
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data).toStrictEqual({
        ref: "origin/main",
        sha: "def789abc012",
        source: "origin/main",
        head_sha: "head456def789",
        diverged: false,
      })
    }
  })

  it("falls back to main when remote-tracking refs fail", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.startsWith("merge-base")) return ""
      if (cmd.includes("rev-parse") && cmd.includes("HEAD")) return "head456def789"
      if (cmd.includes("upstream/main")) throw new Error("not found")
      if (cmd.includes("origin/main")) throw new Error("not found")
      if (cmd.includes("main")) return "789012abc345"
      throw new Error("not found")
    }

    const result = await invokeResolveBaseRef(mockExec)
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data).toStrictEqual({
        ref: "main",
        sha: "789012abc345",
        source: "main",
        head_sha: "head456def789",
        diverged: false,
      })
    }
  })

  it("flags divergence when the base is not an ancestor of HEAD", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.startsWith("merge-base")) throw new Error("not an ancestor")
      if (cmd.includes("rev-parse") && cmd.includes("HEAD")) return "head456def789"
      if (cmd.includes("upstream/main")) return "abc123def456"
      throw new Error("not found")
    }

    const result = await invokeResolveBaseRef(mockExec)
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data).toStrictEqual({
        ref: "upstream/main",
        sha: "abc123def456",
        source: "upstream/main",
        head_sha: "head456def789",
        diverged: true,
      })
    }
  })

  it("returns failure when no ref resolves", async () => {
    const mockExec: GitExec = () => {
      throw new Error("not found")
    }

    const result = await invokeResolveBaseRef(mockExec)
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.retryable).toBe(false)
      expect(result.code).toBe("NO_BASE_REF")
      expect(result.message).toContain("No resolvable base ref found")
      expect(result.message).toContain("upstream/main")
      expect(result.message).toContain("origin/main")
      expect(result.message).toContain("main")
    }
  })

  it("skips candidates that return empty string", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("upstream/main")) return ""
      if (cmd.includes("origin/main")) return "nonempty123"
      throw new Error("not found")
    }

    const result = await invokeResolveBaseRef(mockExec)
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.source).toBe("origin/main")
    }
  })

  it("passes cwd to the executor", async () => {
    const capturedCwds: string[] = []
    const mockExec: GitExec = (args: string[], cwd: string) => {
      capturedCwds.push(cwd)
      const cmd = args.join(" ")
      if (cmd.includes("upstream/main")) return "sha123"
      throw new Error("not found")
    }

    const resolveBaseRefTool = createResolveBaseRefTool(mockExec, "/my/project")
    await (resolveBaseRefTool as { execute: (args: Record<string, never>, ctx: { sessionID: string }) => Promise<{ output: string }> })
      .execute({}, { sessionID: "test-session" })

    expect(capturedCwds.length).toBeGreaterThan(0)
    expect(capturedCwds[0]).toBe("/my/project")
  })
})

// ---------------------------------------------------------------------------
// sanitizeSlug tests (FR-004)
// ---------------------------------------------------------------------------

describe("sanitizeSlug()", () => {
  it("replaces / with -", () => {
    expect(sanitizeSlug("opsx/my-branch")).toBe("opsx-my-branch")
  })

  it("strips non-allowed characters", () => {
    expect(sanitizeSlug("feat/add @stuff!")).toBe("feat-addstuff")
  })

  it("preserves dots and underscores", () => {
    expect(sanitizeSlug("v1.2_beta")).toBe("v1.2_beta")
  })

  it("truncates to 100 characters", () => {
    const long = "a".repeat(150)
    expect(sanitizeSlug(long).length).toBe(100)
  })

  it("throws on empty result after sanitization", () => {
    expect(() => sanitizeSlug("@@@")).toThrow("empty slug")
  })

  it("collapses dot runs", () => {
    expect(sanitizeSlug("a..b...c")).toBe("a.b.c")
  })
})

// ---------------------------------------------------------------------------
// resolveKey tests (FR-004)
// ---------------------------------------------------------------------------

describe("resolveKey()", () => {
  it("returns sanitized explicit key when provided", () => {
    const exec: GitExec = () => {
      throw new Error("should not be called")
    }
    expect(resolveKey(exec, "/tmp", "my/custom-key")).toBe("my-custom-key")
  })

  it("ignores empty string explicit key and auto-detects branch", () => {
    const exec: GitExec = () => "main"
    expect(resolveKey(exec, "/tmp", "")).toBe("main")
  })

  it("auto-detects branch from git when no explicit key", () => {
    const exec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("--abbrev-ref")) return "feature/my-branch"
      throw new Error("unexpected command")
    }
    expect(resolveKey(exec, "/tmp")).toBe("feature-my-branch")
  })

  it("throws on detached HEAD without explicit key", () => {
    const exec: GitExec = () => "HEAD"
    expect(() => resolveKey(exec, "/tmp")).toThrow("Detached HEAD")
  })
})

// ---------------------------------------------------------------------------
// Checkpoint tool helpers
// ---------------------------------------------------------------------------

/** Mock git exec that always returns a fixed branch name. */
const branchExec: GitExec = () => "opsx-test-branch"

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

// ---------------------------------------------------------------------------
// checkpoint tool tests (FR-001)
// ---------------------------------------------------------------------------

describe("checkpoint tool (FR-001)", () => {
  let tmpDir: string

  beforeEach(() => {
    tmpDir = mkdtempSync(join(tmpdir(), "uf-checkpoint-"))
  })

  afterEach(() => {
    rmSync(tmpDir, { recursive: true, force: true })
  })

  it("creates new state on first call", async () => {
    const result = await invokeTool<ProgressState>(
      createCheckpointTool(branchExec, tmpDir),
      { command: "uf.unleash", step: "step-0" },
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      const s = result.data as ProgressState
      expect(s.command).toBe("uf.unleash")
      expect(s.current_step).toBe("step-0")
      expect(s.completed_steps).toEqual(["step-0"])
      expect(s.data).toEqual({})
      expect(s.created_at).toBeTruthy()
      expect(s.updated_at).toBeTruthy()
    }
  })

  it("appends steps on subsequent calls", async () => {
    const cp = createCheckpointTool(branchExec, tmpDir)
    await invokeTool(cp, { command: "uf.unleash", step: "step-0" })
    const result = await invokeTool<ProgressState>(cp, {
      command: "uf.unleash",
      step: "step-1",
    })
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      const s = result.data as ProgressState
      expect(s.current_step).toBe("step-1")
      expect(s.completed_steps).toEqual(["step-0", "step-1"])
    }
  })

  it("duplicate step is idempotent (FR-001)", async () => {
    const cp = createCheckpointTool(branchExec, tmpDir)
    await invokeTool(cp, { command: "uf.unleash", step: "step-0" })
    const result = await invokeTool<ProgressState>(cp, {
      command: "uf.unleash",
      step: "step-0",
    })
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect((result.data as ProgressState).completed_steps).toEqual([
        "step-0",
      ])
    }
  })

  it("shallow merges data via Object.assign (D3)", async () => {
    const cp = createCheckpointTool(branchExec, tmpDir)
    await invokeTool(cp, {
      command: "uf.unleash",
      step: "step-0",
      data: JSON.stringify({ a: 1 }),
    })
    const result = await invokeTool<ProgressState>(cp, {
      command: "uf.unleash",
      step: "step-1",
      data: JSON.stringify({ b: 2 }),
    })
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect((result.data as ProgressState).data).toEqual({ a: 1, b: 2 })
    }
  })

  it("creates files with 0o600 permissions (FR-005)", async () => {
    await invokeTool(createCheckpointTool(branchExec, tmpDir), {
      command: "uf.unleash",
      step: "step-0",
    })
    const fp = join(
      tmpDir,
      ".uf",
      "progress",
      "opsx-test-branch",
      "uf.unleash.json",
    )
    const stats = statSync(fp)
    expect(stats.mode & 0o777).toBe(0o600)
  })

  it("creates directories with 0o700 permissions (FR-005)", async () => {
    await invokeTool(createCheckpointTool(branchExec, tmpDir), {
      command: "uf.unleash",
      step: "step-0",
    })
    const dir = join(tmpDir, ".uf", "progress", "opsx-test-branch")
    const stats = statSync(dir)
    expect(stats.mode & 0o777).toBe(0o700)
  })

  it("uses explicit key when provided", async () => {
    const result = await invokeTool<ProgressState>(
      createCheckpointTool(branchExec, tmpDir),
      { command: "uf.unleash", step: "step-0", key: "custom-key" },
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect((result.data as ProgressState).key).toBe("custom-key")
    }
  })

  it("returns failure on invalid JSON data", async () => {
    const result = await invokeTool(
      createCheckpointTool(branchExec, tmpDir),
      { command: "uf.unleash", step: "step-0", data: "not-json" },
    )
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("CHECKPOINT_ERROR")
    }
  })

  it("records a phase-scoped checkpoint with verdict and inventory", async () => {
    const result = await invokeTool<ProgressState>(
      createCheckpointTool(branchExec, tmpDir),
      {
        command: "uf.unleash",
        step: "implement-phase-4",
        phase: {
          number: 4,
          verdict: "green",
          inventory: [{ name: "TestX", classification: "pre-existing" }],
        },
      },
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      const s = result.data as ProgressState
      expect(s.completed_steps).toEqual(["implement-phase-4"])
      expect(s.phases?.["4"]?.verdict).toBe("green")
      expect(s.phases?.["4"]?.inventory).toEqual([
        { name: "TestX", classification: "pre-existing" },
      ])
      expect(s.phases?.["4"]?.recorded_at).toBeTruthy()
    }
  })

  it("round-trips a phase-scoped record through get_progress", async () => {
    const cp = createCheckpointTool(branchExec, tmpDir)
    await invokeTool(cp, {
      command: "uf.unleash",
      step: "implement-phase-2",
      phase: {
        number: 2,
        verdict: "green",
        inventory: [
          { name: "TestTLS", classification: "environmental" },
          { name: "TestPath", classification: "pre-existing" },
        ],
      },
    })
    const result = await invokeTool<ProgressState>(
      createGetProgressTool(branchExec, tmpDir),
      { command: "uf.unleash" },
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      const s = result.data as ProgressState
      expect(s.current_step).toBe("implement-phase-2")
      expect(s.completed_steps).toEqual(["implement-phase-2"])
      expect(s.phases?.["2"]?.verdict).toBe("green")
      expect(s.phases?.["2"]?.inventory).toEqual([
        { name: "TestTLS", classification: "environmental" },
        { name: "TestPath", classification: "pre-existing" },
      ])
    }
  })

  it("re-recording a phase replaces its record and stays step-idempotent", async () => {
    const cp = createCheckpointTool(branchExec, tmpDir)
    await invokeTool(cp, {
      command: "uf.unleash",
      step: "implement-phase-4",
      phase: {
        number: 4,
        verdict: "failed",
        inventory: [{ name: "TestA", classification: "change-caused" }],
      },
    })
    const result = await invokeTool<ProgressState>(cp, {
      command: "uf.unleash",
      step: "implement-phase-4",
      phase: { number: 4, verdict: "green", inventory: [] },
    })
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      const s = result.data as ProgressState
      expect(s.completed_steps).toEqual(["implement-phase-4"])
      expect(s.phases?.["4"]?.verdict).toBe("green")
      expect(s.phases?.["4"]?.inventory).toEqual([])
    }
  })

  it("defaults an omitted inventory to an empty list", async () => {
    const result = await invokeTool<ProgressState>(
      createCheckpointTool(branchExec, tmpDir),
      {
        command: "uf.unleash",
        step: "implement-phase-1",
        phase: { number: 1, verdict: "green" },
      },
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect((result.data as ProgressState).phases?.["1"]?.inventory).toEqual([])
    }
  })

  it("rejects a phase record whose step is not implement-phase-<n>", async () => {
    const result = await invokeTool(
      createCheckpointTool(branchExec, tmpDir),
      {
        command: "uf.unleash",
        step: "step-0",
        phase: { number: 4, verdict: "green", inventory: [] },
      },
    )
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("CHECKPOINT_ERROR")
      expect(result.message).toContain("implement-phase-4")
    }
  })
})

// ---------------------------------------------------------------------------
// get_progress tool tests (FR-002)
// ---------------------------------------------------------------------------

describe("get_progress tool (FR-002)", () => {
  let tmpDir: string

  beforeEach(() => {
    tmpDir = mkdtempSync(join(tmpdir(), "uf-getprogress-"))
  })

  afterEach(() => {
    rmSync(tmpDir, { recursive: true, force: true })
  })

  it("returns null when no state exists", async () => {
    const result = await invokeTool(
      createGetProgressTool(branchExec, tmpDir),
      { command: "uf.unleash" },
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data).toBeNull()
    }
  })

  it("returns full state after checkpoint", async () => {
    await invokeTool(createCheckpointTool(branchExec, tmpDir), {
      command: "uf.unleash",
      step: "step-0",
    })
    const result = await invokeTool<ProgressState>(
      createGetProgressTool(branchExec, tmpDir),
      { command: "uf.unleash" },
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      const s = result.data as ProgressState
      expect(s.command).toBe("uf.unleash")
      expect(s.completed_steps).toEqual(["step-0"])
    }
  })
})

// ---------------------------------------------------------------------------
// clear_progress tool tests (FR-003)
// ---------------------------------------------------------------------------

describe("clear_progress tool (FR-003)", () => {
  let tmpDir: string

  beforeEach(() => {
    tmpDir = mkdtempSync(join(tmpdir(), "uf-clearprogress-"))
  })

  afterEach(() => {
    rmSync(tmpDir, { recursive: true, force: true })
  })

  it("deletes existing state file", async () => {
    await invokeTool(createCheckpointTool(branchExec, tmpDir), {
      command: "uf.unleash",
      step: "step-0",
    })
    const clearResult = await invokeTool(
      createClearProgressTool(branchExec, tmpDir),
      { command: "uf.unleash" },
    )
    expect(clearResult.status).toBe("ok")
    // Verify file is gone by reading progress
    const getResult = await invokeTool(
      createGetProgressTool(branchExec, tmpDir),
      { command: "uf.unleash" },
    )
    expect(getResult.status).toBe("ok")
    if (getResult.status === "ok") {
      expect(getResult.data).toBeNull()
    }
  })

  it("is idempotent when no state file exists", async () => {
    const result = await invokeTool(
      createClearProgressTool(branchExec, tmpDir),
      { command: "uf.unleash" },
    )
    expect(result.status).toBe("ok")
  })

  it("returns failure on error during clear", async () => {
    // Use a failing exec to trigger error in resolveKey
    const failExec: GitExec = () => {
      throw new Error("git not found")
    }
    const result = await invokeTool(
      createClearProgressTool(failExec, tmpDir),
      { command: "uf.unleash" },
    )
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("CLEAR_PROGRESS_ERROR")
      expect(result.message).toContain("git not found")
    }
  })
})

// ---------------------------------------------------------------------------
// get_progress error path tests
// ---------------------------------------------------------------------------

describe("get_progress error paths", () => {
  let tmpDir: string

  beforeEach(() => {
    tmpDir = mkdtempSync(join(tmpdir(), "uf-getprogress-err-"))
  })

  afterEach(() => {
    rmSync(tmpDir, { recursive: true, force: true })
  })

  it("returns failure when resolveKey throws", async () => {
    const failExec: GitExec = () => {
      throw new Error("git not found")
    }
    const result = await invokeTool(
      createGetProgressTool(failExec, tmpDir),
      { command: "uf.unleash" },
    )
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("GET_PROGRESS_ERROR")
      expect(result.message).toContain("git not found")
    }
  })
})

// ---------------------------------------------------------------------------
// Symlink defense tests (FR-005)
// ---------------------------------------------------------------------------

describe("symlink defense (FR-005)", () => {
  let tmpDir: string

  beforeEach(() => {
    tmpDir = mkdtempSync(join(tmpdir(), "uf-symlink-"))
  })

  afterEach(() => {
    rmSync(tmpDir, { recursive: true, force: true })
  })

  it("checkpoint refuses to write into symlinked directory", async () => {
    // Create a real target directory and a symlink to it
    const realDir = join(tmpDir, "real-progress")
    const symlinkDir = join(tmpDir, ".uf", "progress", "opsx-test-branch")
    mkdirSync(realDir, { recursive: true })
    mkdirSync(join(tmpDir, ".uf", "progress"), { recursive: true })
    symlinkSync(realDir, symlinkDir)

    const result = await invokeTool(
      createCheckpointTool(branchExec, tmpDir),
      { command: "uf.unleash", step: "step-0" },
    )
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.message).toContain("symlinked directory")
    }
  })

  it("clear_progress refuses to delete symlinked state file", async () => {
    const progressDir = join(tmpDir, ".uf", "progress", "opsx-test-branch")
    mkdirSync(progressDir, { recursive: true, mode: 0o700 })
    const realFile = join(tmpDir, "real-state.json")
    writeFileSync(realFile, "{}")
    symlinkSync(realFile, join(progressDir, "uf.unleash.json"))

    const result = await invokeTool(
      createClearProgressTool(branchExec, tmpDir),
      { command: "uf.unleash" },
    )
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.message).toContain("symlinked state file")
    }
  })

  it("get_progress refuses to read symlinked state file", async () => {
    // First create a real state file
    const progressDir = join(tmpDir, ".uf", "progress", "opsx-test-branch")
    mkdirSync(progressDir, { recursive: true, mode: 0o700 })
    const realFile = join(tmpDir, "real-state.json")
    writeFileSync(
      realFile,
      JSON.stringify({
        command: "test",
        branch: "opsx-test-branch",
        key: "opsx-test-branch",
        current_step: "step-0",
        completed_steps: ["step-0"],
        data: {},
        created_at: "2026-01-01T00:00:00.000Z",
        updated_at: "2026-01-01T00:00:00.000Z",
      }),
    )
    // Create symlink where the state file would be
    symlinkSync(realFile, join(progressDir, "uf.unleash.json"))

    const result = await invokeTool(
      createGetProgressTool(branchExec, tmpDir),
      { command: "uf.unleash" },
    )
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.message).toContain("symlinked state file")
    }
  })
})

// ---------------------------------------------------------------------------
// write_unleash_summary tool tests (FR-001 through FR-005 of the
// persist-unleash-summary change, adjusted to the merged contract).
//
// The merge of persist-unleash-summary (origin/main) and
// unleash-deterministic-tools (this branch) deduplicated the two Step 10
// summary writers into write_unleash_summary as the mechanism of record;
// the unleash-pipeline-hardening change then removed the redundant
// persist_unleash_summary registration so the tool map lists exactly one
// summary tool. createPersistUnleashSummaryTool remains the exported
// compatibility factory delegating to that single implementation, and
// these tests exercise the merged contract through it (the full
// eight-parameter payload) and the merged failure taxonomy (DETACHED_HEAD,
// SUMMARY_DIR_IS_SYMLINK, SUMMARY_PATH_IS_SYMLINK, SUMMARY_WRITE_ERROR
// instead of the original single PERSIST_UNLEASH_SUMMARY_ERROR catch-all).
// ---------------------------------------------------------------------------

describe("write_unleash_summary (persist_unleash_summary compatibility factory)", () => {
  let tmpDir: string

  beforeEach(() => {
    tmpDir = mkdtempSync(join(tmpdir(), "uf-unleash-summary-"))
  })

  afterEach(() => {
    rmSync(tmpDir, { recursive: true, force: true })
  })

  /** Mock git exec reporting a slash branch to exercise slug sanitization. */
  const opsxBranchExec: GitExec = () => "opsx/my-change"

  /** Valid full payload: the three demo sections plus the merged workflow metadata. */
  const validArgs = {
    change_name: "my-change",
    tier: "openspec",
    feature_dir: "openspec/changes/my-change/",
    review_verdicts: {
      spec_review: { verdict: "APPROVE", iterations: 1 },
      code_review: { verdict: "APPROVE WITH ADVISORIES", iterations: 2 },
    },
    learnings_stored: [".uf/dewey/learnings/persist-unleash-summary-20261016T080000-jay-flowers.md"],
    whatWasBuilt: "Implemented the persist_unleash_summary tool",
    howToVerify: "Run `make plugin-test` and confirm the suite passes",
    testResults: "All prior tests still pass; new tool tests added",
  }

  /** Artifact path the `opsx/my-change` branch slug resolves to under cwd. */
  const artifactPath = (cwd: string): string =>
    join(cwd, ".uf", "artifacts", "unleash-summary", "opsx-my-change.json")

  it("writes the artifact at the sanitized branch slug with schema fields and provenance", async () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date("2026-01-01T00:00:00.000Z"))
    try {
      const captured: Array<{ args: string[]; cwd: string }> = []
      const mockExec: GitExec = (args: string[], cwd: string) => {
        captured.push({ args, cwd })
        return "opsx/my-change"
      }

      const result = await invokeTool<{ path: string; branch: string }>(
        createPersistUnleashSummaryTool(mockExec, tmpDir),
        validArgs,
      )

      expect(result.status).toBe("ok")
      if (result.status === "ok") {
        expect(result.data?.path).toBe(artifactPath(tmpDir))
        expect(result.data?.branch).toBe("opsx/my-change")
      }
      // Branch detection runs exactly one rev-parse with the injected cwd.
      expect(captured).toEqual([
        { args: ["rev-parse", "--abbrev-ref", "HEAD"], cwd: tmpDir },
      ])

      const raw = readFileSync(artifactPath(tmpDir), "utf-8")
      // Write format is pretty JSON with a trailing newline (FR-003).
      expect(raw.endsWith("\n")).toBe(true)
      const artifact = JSON.parse(raw) as UnleashSummary
      expect(artifact.whatWasBuilt).toBe(validArgs.whatWasBuilt)
      expect(artifact.howToVerify).toBe(validArgs.howToVerify)
      expect(artifact.testResults).toBe(validArgs.testResults)
      expect(artifact.artifact_type).toBe("unleash-summary")
      expect(artifact.schema_version).toBe(1)
      expect(artifact.branch).toBe("opsx/my-change")
      expect(artifact.timestamp).toBe("2026-01-01T00:00:00.000Z")
    } finally {
      vi.useRealTimers()
    }
  })

  it("exposes top-level whatWasBuilt and howToVerify strings (uf_finale_pr_context contract)", async () => {
    await invokeTool(
      createPersistUnleashSummaryTool(opsxBranchExec, tmpDir),
      validArgs,
    )

    // uf_finale_pr_context reads the file wholesale and finale prose
    // accesses `.whatWasBuilt` / `.howToVerify` directly — flat camelCase
    // content fields with no result-envelope nesting around them.
    const parsed = JSON.parse(
      readFileSync(artifactPath(tmpDir), "utf-8"),
    ) as Record<string, unknown>
    expect(typeof parsed.whatWasBuilt).toBe("string")
    expect(typeof parsed.howToVerify).toBe("string")
    expect(parsed.whatWasBuilt).toBe(validArgs.whatWasBuilt)
    expect(parsed.howToVerify).toBe(validArgs.howToVerify)
    expect(parsed).not.toHaveProperty("data")
    expect(parsed).not.toHaveProperty("status")
  })

  it("creates the artifact directory 0o700 and file 0o600 (FR-005)", async () => {
    await invokeTool(
      createPersistUnleashSummaryTool(opsxBranchExec, tmpDir),
      validArgs,
    )

    const dir = join(tmpDir, ".uf", "artifacts", "unleash-summary")
    expect(statSync(dir).mode & 0o777).toBe(0o700)
    expect(statSync(artifactPath(tmpDir)).mode & 0o777).toBe(0o600)
  })

  it("overwrites the previous artifact and refreshes the timestamp", async () => {
    vi.useFakeTimers()
    try {
      vi.setSystemTime(new Date("2026-01-01T00:00:00.000Z"))
      const tool = createPersistUnleashSummaryTool(opsxBranchExec, tmpDir)
      await invokeTool(tool, validArgs)

      vi.setSystemTime(new Date("2026-01-02T12:30:45.000Z"))
      const replacement = {
        ...validArgs,
        whatWasBuilt: "Second run supersedes the first",
        howToVerify: "Re-read the artifact and confirm only this content",
        testResults: "Overwrite semantics verified",
      }
      const result = await invokeTool<{ path: string }>(tool, replacement)
      expect(result.status).toBe("ok")

      const artifact = JSON.parse(
        readFileSync(artifactPath(tmpDir), "utf-8"),
      ) as UnleashSummary
      expect(artifact.whatWasBuilt).toBe(replacement.whatWasBuilt)
      expect(artifact.howToVerify).toBe(replacement.howToVerify)
      expect(artifact.testResults).toBe(replacement.testResults)
      expect(artifact.timestamp).toBe("2026-01-02T12:30:45.000Z")
      // Unconditional overwrite: the branch still owns exactly one file.
      expect(
        readdirSync(join(tmpDir, ".uf", "artifacts", "unleash-summary")),
      ).toEqual(["opsx-my-change.json"])
    } finally {
      vi.useRealTimers()
    }
  })

  it("rejects an empty whatWasBuilt without any filesystem side effect", async () => {
    // Validation must precede the git call — this exec fails the test if run.
    const neverCalledExec: GitExec = () => {
      throw new Error("exec must not be called for invalid parameters")
    }
    const result = await invokeTool(
      createPersistUnleashSummaryTool(neverCalledExec, tmpDir),
      { ...validArgs, whatWasBuilt: "" },
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_UNLEASH_SUMMARY")
      expect(result.retryable).toBe(false)
      expect(result.message).toContain("whatWasBuilt")
      expect(result.message).toContain("non-empty")
    }
    // FR-004: validate before any side effect — nothing is created at all.
    expect(existsSync(join(tmpDir, ".uf"))).toBe(false)
  })

  it("rejects whitespace-only howToVerify without any filesystem side effect", async () => {
    const result = await invokeTool(
      createPersistUnleashSummaryTool(opsxBranchExec, tmpDir),
      { ...validArgs, howToVerify: "   \n\t " },
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_UNLEASH_SUMMARY")
      expect(result.message).toContain("howToVerify")
    }
    expect(existsSync(join(tmpDir, ".uf"))).toBe(false)
  })

  it("rejects a missing testResults without any filesystem side effect", async () => {
    const result = await invokeTool(
      createPersistUnleashSummaryTool(opsxBranchExec, tmpDir),
      {
        whatWasBuilt: validArgs.whatWasBuilt,
        howToVerify: validArgs.howToVerify,
      },
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_UNLEASH_SUMMARY")
      expect(result.message).toContain("testResults")
    }
    expect(existsSync(join(tmpDir, ".uf"))).toBe(false)
  })

  it("fails on detached HEAD without writing", async () => {
    const detachedExec: GitExec = () => "HEAD"
    const result = await invokeTool(
      createPersistUnleashSummaryTool(detachedExec, tmpDir),
      validArgs,
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      // Merged taxonomy: the branch-derived file name gets a dedicated
      // DETACHED_HEAD code rather than the original catch-all.
      expect(result.code).toBe("DETACHED_HEAD")
      expect(result.message).toContain("Detached HEAD")
    }
    expect(existsSync(join(tmpDir, ".uf"))).toBe(false)
  })

  it("wraps git exec failures as SUMMARY_WRITE_ERROR", async () => {
    const failExec: GitExec = () => {
      throw new Error("git not found")
    }
    const result = await invokeTool(
      createPersistUnleashSummaryTool(failExec, tmpDir),
      validArgs,
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("SUMMARY_WRITE_ERROR")
      expect(result.retryable).toBe(false)
      expect(result.message).toContain("git not found")
    }
    expect(existsSync(join(tmpDir, ".uf"))).toBe(false)
  })

  it("stringifies non-Error operational failures", async () => {
    const failExec: GitExec = () => {
      throw "disk exploded" // eslint-disable-line no-throw-literal
    }
    const result = await invokeTool(
      createPersistUnleashSummaryTool(failExec, tmpDir),
      validArgs,
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("SUMMARY_WRITE_ERROR")
      expect(result.message).toBe("disk exploded")
    }
  })

  it("refuses to write into a symlinked artifact directory", async () => {
    const realDir = join(tmpDir, "real-target")
    const symlinkDir = join(tmpDir, ".uf", "artifacts", "unleash-summary")
    mkdirSync(realDir, { recursive: true })
    mkdirSync(join(tmpDir, ".uf", "artifacts"), { recursive: true })
    symlinkSync(realDir, symlinkDir)

    const result = await invokeTool(
      createPersistUnleashSummaryTool(opsxBranchExec, tmpDir),
      validArgs,
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      // Merged taxonomy: the directory refusal carries its own code.
      expect(result.code).toBe("SUMMARY_DIR_IS_SYMLINK")
      expect(result.message).toContain("symlinked directory")
    }
    // Nothing was written through the link.
    expect(readdirSync(realDir)).toEqual([])
  })

  it("refuses to overwrite a symlinked artifact file", async () => {
    const dir = join(tmpDir, ".uf", "artifacts", "unleash-summary")
    mkdirSync(dir, { recursive: true })
    const realFile = join(tmpDir, "real-summary.json")
    writeFileSync(realFile, "sentinel")
    symlinkSync(realFile, join(dir, "opsx-my-change.json"))

    const result = await invokeTool(
      createPersistUnleashSummaryTool(opsxBranchExec, tmpDir),
      validArgs,
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      // Merged taxonomy: the target-file refusal carries its own code.
      expect(result.code).toBe("SUMMARY_PATH_IS_SYMLINK")
      expect(result.message).toContain("symlinked path")
    }
    // The real target behind the link was not modified.
    expect(readFileSync(realFile, "utf-8")).toBe("sentinel")
  })

  it("returns operational failure on non-ENOENT directory stat errors", async () => {
    // `.uf/artifacts` occupied by a regular file makes lstatSync of the
    // artifact directory fail with ENOTDIR instead of ENOENT.
    mkdirSync(join(tmpDir, ".uf"))
    writeFileSync(join(tmpDir, ".uf", "artifacts"), "blocker")

    const result = await invokeTool(
      createPersistUnleashSummaryTool(opsxBranchExec, tmpDir),
      validArgs,
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("SUMMARY_WRITE_ERROR")
      expect(result.message).toContain("ENOTDIR")
    }
  })

  it("returns operational failure on non-ENOENT file stat errors", async () => {
    // The artifact directory occupied by a regular file makes lstatSync of
    // the artifact file path fail with ENOTDIR instead of ENOENT.
    mkdirSync(join(tmpDir, ".uf", "artifacts"), { recursive: true })
    writeFileSync(join(tmpDir, ".uf", "artifacts", "unleash-summary"), "blocker")

    const result = await invokeTool(
      createPersistUnleashSummaryTool(opsxBranchExec, tmpDir),
      validArgs,
    )

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("SUMMARY_WRITE_ERROR")
      expect(result.message).toContain("ENOTDIR")
    }
  })
})

// ---------------------------------------------------------------------------
// normalize_model_name tool tests
// ---------------------------------------------------------------------------

describe("normalize_model_name", () => {
  it("strips provider prefix and @-suffix and date suffix", async () => {
    const result = await invokeTool<{ normalized: string }>(
      createNormalizeModelNameTool(),
      { rawModelId: "google-vertex-anthropic/claude-sonnet-4-20250514@default" },
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.normalized).toBe("claude-sonnet-4")
    }
  })

  it("passes through already-normalized names", async () => {
    const result = await invokeTool<{ normalized: string }>(
      createNormalizeModelNameTool(),
      { rawModelId: "claude-sonnet-4" },
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.normalized).toBe("claude-sonnet-4")
    }
  })

  it("returns unknown-model with warning for empty input", async () => {
    const result = await invokeTool<{ normalized: string; warning: string }>(
      createNormalizeModelNameTool(),
      { rawModelId: "" },
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.normalized).toBe("unknown-model")
      expect(result.data?.warning).toBeDefined()
    }
  })

  it("strips trailing single-digit version segments repeatedly", async () => {
    // `-1` is stripped because preceded by digit `4` (chained sub-version).
    // `-4` is preserved because preceded by `e` (part of model name).
    const result = await invokeTool<{ normalized: string }>(
      createNormalizeModelNameTool(),
      { rawModelId: "provider/model-name-4-1@v2" },
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.normalized).toBe("model-name-4")
    }
  })

  it("strips chained sub-version digits like -4-6", async () => {
    // claude-opus-4-6: `-6` stripped (preceded by `4`),
    // `-4` preserved (preceded by `s`)
    const result = await invokeTool<{ normalized: string }>(
      createNormalizeModelNameTool(),
      { rawModelId: "anthropic/claude-opus-4-6-20260101@default" },
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.normalized).toBe("claude-opus-4")
    }
  })

  it("handles name with only @-suffix and no provider", async () => {
    const result = await invokeTool<{ normalized: string }>(
      createNormalizeModelNameTool(),
      { rawModelId: "gpt-4o@latest" },
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.normalized).toBe("gpt-4o")
    }
  })

  it("returns unknown-model for whitespace-only input", async () => {
    const result = await invokeTool<{ normalized: string; warning: string }>(
      createNormalizeModelNameTool(),
      { rawModelId: "   " },
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.normalized).toBe("unknown-model")
    }
  })

  it("returns unknown-model for input that normalizes to invalid chars", async () => {
    const result = await invokeTool<{ normalized: string; warning: string }>(
      createNormalizeModelNameTool(),
      { rawModelId: "provider/!!!@suffix" },
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.normalized).toBe("unknown-model")
      expect(result.data?.warning).toContain("Could not normalize")
    }
  })
})

// ---------------------------------------------------------------------------
// Git operation tool tests
// ---------------------------------------------------------------------------

describe("git_fetch", () => {
  it("executes git fetch with remote and branch", async () => {
    const commands: string[] = []
    const mockExec: GitExec = (args: string[]) => {
      commands.push(args.join(" "))
      return ""
    }
    const result = await invokeTool<{ success: boolean }>(
      createGitFetchTool(mockExec, "/fake"),
      { remote: "origin", branch: "main" },
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.success).toBe(true)
    }
    expect(commands[0]).toBe("fetch origin main")
  })

  it("executes git fetch with remote only when branch omitted", async () => {
    const commands: string[] = []
    const mockExec: GitExec = (args: string[]) => {
      commands.push(args.join(" "))
      return ""
    }
    const result = await invokeTool<{ success: boolean }>(
      createGitFetchTool(mockExec, "/fake"),
      { remote: "upstream" },
    )
    expect(result.status).toBe("ok")
    expect(commands[0]).toBe("fetch upstream")
  })

  it("returns failure on fetch error", async () => {
    const mockExec: GitExec = () => {
      throw new Error("fatal: could not read from remote")
    }
    const result = await invokeTool(
      createGitFetchTool(mockExec, "/fake"),
      { remote: "origin" },
    )
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("GIT_FETCH_ERROR")
      expect(result.retryable).toBe(true)
    }
  })
})

describe("git_push", () => {
  it("executes plain git push when no options", async () => {
    const commands: string[] = []
    const mockExec: GitExec = (args: string[]) => {
      commands.push(args.join(" "))
      return ""
    }
    const result = await invokeTool<{ success: boolean }>(
      createGitPushTool(mockExec, "/fake"),
      {},
    )
    expect(result.status).toBe("ok")
    expect(commands[0]).toBe("push")
  })

  it("sets upstream when setUpstream and branch provided", async () => {
    const commands: string[] = []
    const mockExec: GitExec = (args: string[]) => {
      commands.push(args.join(" "))
      return ""
    }
    const result = await invokeTool<{ success: boolean }>(
      createGitPushTool(mockExec, "/fake"),
      { setUpstream: true, branch: "feat/my-branch" },
    )
    expect(result.status).toBe("ok")
    expect(commands[0]).toBe("push --set-upstream origin feat/my-branch")
  })

  it("returns failure on push error", async () => {
    const mockExec: GitExec = () => {
      throw new Error("rejected")
    }
    const result = await invokeTool(
      createGitPushTool(mockExec, "/fake"),
      {},
    )
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("GIT_PUSH_ERROR")
    }
  })
})

describe("git_push_force_with_lease", () => {
  it("executes git push --force-with-lease", async () => {
    const commands: string[] = []
    const mockExec: GitExec = (args: string[]) => {
      commands.push(args.join(" "))
      return ""
    }
    const result = await invokeTool<{ success: boolean }>(
      createGitPushForceWithLeaseTool(mockExec, "/fake"),
      {},
    )
    expect(result.status).toBe("ok")
    expect(commands[0]).toBe("push --force-with-lease")
  })

  it("returns failure on force-push error", async () => {
    const mockExec: GitExec = () => {
      throw new Error("stale info")
    }
    const result = await invokeTool(
      createGitPushForceWithLeaseTool(mockExec, "/fake"),
      {},
    )
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("GIT_PUSH_FORCE_ERROR")
    }
  })
})

describe("git_merge", () => {
  it("returns success with conflicted false on clean merge", async () => {
    const mockExec: GitExec = () => "Already up to date."
    const result = await invokeTool<{
      success: boolean
      conflicted: boolean
    }>(createGitMergeTool(mockExec, "/fake"), { ref: "main" })
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.success).toBe(true)
      expect(result.data?.conflicted).toBe(false)
    }
  })

  it("detects conflicts and lists conflicting files", async () => {
    let callCount = 0
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      callCount++
      if (cmd.startsWith("merge")) {
        throw new Error("CONFLICT (content): Merge conflict in file.txt")
      }
      if (cmd.includes("--diff-filter=U")) {
        return "file.txt\nsrc/other.ts"
      }
      return ""
    }
    const result = await invokeTool<{
      success: boolean
      conflicted: boolean
      conflictingFiles: string[]
    }>(createGitMergeTool(mockExec, "/fake"), { ref: "feature" })
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.success).toBe(false)
      expect(result.data?.conflicted).toBe(true)
      expect(result.data?.conflictingFiles).toStrictEqual([
        "file.txt",
        "src/other.ts",
      ])
    }
  })

  it("returns empty conflictingFiles when diff command also fails", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.startsWith("merge")) {
        throw new Error("CONFLICT in some file")
      }
      // Listing conflicts also fails
      throw new Error("not a git repository")
    }
    const result = await invokeTool<{
      success: boolean
      conflicted: boolean
      conflictingFiles: string[]
    }>(createGitMergeTool(mockExec, "/fake"), { ref: "feature" })
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.conflicted).toBe(true)
      expect(result.data?.conflictingFiles).toStrictEqual([])
    }
  })

  it("returns failure for non-conflict errors", async () => {
    const mockExec: GitExec = () => {
      throw new Error("fatal: not a git repository")
    }
    const result = await invokeTool(
      createGitMergeTool(mockExec, "/fake"),
      { ref: "main" },
    )
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("GIT_MERGE_ERROR")
    }
  })
})

describe("git_merge_abort", () => {
  it("executes git merge --abort", async () => {
    const commands: string[] = []
    const mockExec: GitExec = (args: string[]) => {
      commands.push(args.join(" "))
      return ""
    }
    const result = await invokeTool<{ success: boolean }>(
      createGitMergeAbortTool(mockExec, "/fake"),
      {},
    )
    expect(result.status).toBe("ok")
    expect(commands[0]).toBe("merge --abort")
  })

  it("returns failure when no merge in progress", async () => {
    const mockExec: GitExec = () => {
      throw new Error("fatal: There is no merge to abort")
    }
    const result = await invokeTool(
      createGitMergeAbortTool(mockExec, "/fake"),
      {},
    )
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("GIT_MERGE_ABORT_ERROR")
    }
  })
})

describe("git_rebase", () => {
  it("returns success with conflicted false on clean rebase", async () => {
    const mockExec: GitExec = () => "Successfully rebased"
    const result = await invokeTool<{
      success: boolean
      conflicted: boolean
    }>(createGitRebaseTool(mockExec, "/fake"), { ref: "main" })
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.success).toBe(true)
      expect(result.data?.conflicted).toBe(false)
    }
  })

  it("detects conflicts and lists conflicting files", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.startsWith("rebase")) {
        throw new Error("CONFLICT (content): Merge conflict in src/app.ts")
      }
      if (cmd.includes("--diff-filter=U")) {
        return "src/app.ts"
      }
      return ""
    }
    const result = await invokeTool<{
      success: boolean
      conflicted: boolean
      conflictingFiles: string[]
    }>(createGitRebaseTool(mockExec, "/fake"), { ref: "main" })
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.success).toBe(false)
      expect(result.data?.conflicted).toBe(true)
      expect(result.data?.conflictingFiles).toStrictEqual(["src/app.ts"])
    }
  })

  it("returns empty conflictingFiles when diff command also fails", async () => {
    const mockExec: GitExec = () => {
      throw new Error("conflict detected")
    }
    const result = await invokeTool<{
      success: boolean
      conflicted: boolean
      conflictingFiles: string[]
    }>(createGitRebaseTool(mockExec, "/fake"), { ref: "main" })
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.conflicted).toBe(true)
      expect(result.data?.conflictingFiles).toStrictEqual([])
    }
  })

  it("returns failure for non-conflict errors", async () => {
    const mockExec: GitExec = () => {
      throw new Error("fatal: not a git repository")
    }
    const result = await invokeTool(
      createGitRebaseTool(mockExec, "/fake"),
      { ref: "main" },
    )
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("GIT_REBASE_ERROR")
    }
  })
})

describe("git_rebase_abort", () => {
  it("executes git rebase --abort", async () => {
    const commands: string[] = []
    const mockExec: GitExec = (args: string[]) => {
      commands.push(args.join(" "))
      return ""
    }
    const result = await invokeTool<{ success: boolean }>(
      createGitRebaseAbortTool(mockExec, "/fake"),
      {},
    )
    expect(result.status).toBe("ok")
    expect(commands[0]).toBe("rebase --abort")
  })

  it("returns failure when no rebase in progress", async () => {
    const mockExec: GitExec = () => {
      throw new Error("fatal: No rebase in progress")
    }
    const result = await invokeTool(
      createGitRebaseAbortTool(mockExec, "/fake"),
      {},
    )
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("GIT_REBASE_ABORT_ERROR")
    }
  })
})

describe("git_conflict_files", () => {
  it("lists conflicting files", async () => {
    const mockExec: GitExec = () => "file1.ts\nfile2.ts\n"
    const result = await invokeTool<{ files: string[] }>(
      createGitConflictFilesTool(mockExec, "/fake"),
      {},
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.files).toStrictEqual(["file1.ts", "file2.ts"])
    }
  })

  it("returns empty array when no conflicts", async () => {
    const mockExec: GitExec = () => ""
    const result = await invokeTool<{ files: string[] }>(
      createGitConflictFilesTool(mockExec, "/fake"),
      {},
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.files).toStrictEqual([])
    }
  })

  it("returns failure on error", async () => {
    const mockExec: GitExec = () => {
      throw new Error("not a git repo")
    }
    const result = await invokeTool(
      createGitConflictFilesTool(mockExec, "/fake"),
      {},
    )
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("GIT_CONFLICT_FILES_ERROR")
    }
  })
})

describe("git_commit_no_edit", () => {
  it("commits and returns short SHA", async () => {
    const mockExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd === "commit --no-edit") return "merge commit created"
      if (cmd.includes("rev-parse --short")) return "abc1234"
      return ""
    }
    const result = await invokeTool<{ success: boolean; sha: string }>(
      createGitCommitNoEditTool(mockExec, "/fake"),
      {},
    )
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.success).toBe(true)
      expect(result.data?.sha).toBe("abc1234")
    }
  })

  it("returns failure when nothing to commit", async () => {
    const mockExec: GitExec = () => {
      throw new Error("nothing to commit")
    }
    const result = await invokeTool(
      createGitCommitNoEditTool(mockExec, "/fake"),
      {},
    )
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("GIT_COMMIT_NO_EDIT_ERROR")
    }
  })
})

// ---------------------------------------------------------------------------
// Terminal-write regression tests (Requirement 1, TC-006)
// ---------------------------------------------------------------------------

describe("terminal-write regression (Requirement 1)", () => {
  const throwingExec: GitExec = () => {
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

  it("git_fetch writes nothing to the terminal", async () => {
    const successExec: GitExec = () => ""
    await expectNoTerminalWrites(async () => {
      await invokeTool(createGitFetchTool(successExec, "/fake"), {
        remote: "origin",
        branch: "main",
      })
      await invokeTool(createGitFetchTool(throwingExec, "/fake"), {
        remote: "origin",
      })
    })
  })

  it("git_push writes nothing to the terminal", async () => {
    const successExec: GitExec = () => ""
    await expectNoTerminalWrites(async () => {
      await invokeTool(createGitPushTool(successExec, "/fake"), {})
      await invokeTool(createGitPushTool(throwingExec, "/fake"), {})
    })
  })

  it("git_push_force_with_lease writes nothing to the terminal", async () => {
    const successExec: GitExec = () => ""
    await expectNoTerminalWrites(async () => {
      await invokeTool(createGitPushForceWithLeaseTool(successExec, "/fake"), {})
      await invokeTool(
        createGitPushForceWithLeaseTool(throwingExec, "/fake"),
        {},
      )
    })
  })

  it("git_merge writes nothing to the terminal", async () => {
    const successExec: GitExec = () => "Already up to date."
    await expectNoTerminalWrites(async () => {
      await invokeTool(createGitMergeTool(successExec, "/fake"), {
        ref: "main",
      })
      await invokeTool(createGitMergeTool(throwingExec, "/fake"), {
        ref: "main",
      })
    })
  })

  it("git_merge_abort writes nothing to the terminal", async () => {
    const successExec: GitExec = () => ""
    await expectNoTerminalWrites(async () => {
      await invokeTool(createGitMergeAbortTool(successExec, "/fake"), {})
      await invokeTool(createGitMergeAbortTool(throwingExec, "/fake"), {})
    })
  })

  it("git_rebase writes nothing to the terminal", async () => {
    const successExec: GitExec = () => "Successfully rebased"
    await expectNoTerminalWrites(async () => {
      await invokeTool(createGitRebaseTool(successExec, "/fake"), {
        ref: "main",
      })
      await invokeTool(createGitRebaseTool(throwingExec, "/fake"), {
        ref: "main",
      })
    })
  })

  it("git_rebase_abort writes nothing to the terminal", async () => {
    const successExec: GitExec = () => ""
    await expectNoTerminalWrites(async () => {
      await invokeTool(createGitRebaseAbortTool(successExec, "/fake"), {})
      await invokeTool(createGitRebaseAbortTool(throwingExec, "/fake"), {})
    })
  })

  it("git_conflict_files writes nothing to the terminal", async () => {
    const successExec: GitExec = () => "file1.ts\nfile2.ts"
    await expectNoTerminalWrites(async () => {
      await invokeTool(createGitConflictFilesTool(successExec, "/fake"), {})
      await invokeTool(createGitConflictFilesTool(throwingExec, "/fake"), {})
    })
  })

  it("git_commit_no_edit writes nothing to the terminal", async () => {
    const successExec: GitExec = (args: string[]) => {
      const cmd = args.join(" ")
      if (cmd.includes("rev-parse --short")) return "abc1234"
      return ""
    }
    await expectNoTerminalWrites(async () => {
      await invokeTool(createGitCommitNoEditTool(successExec, "/fake"), {})
      await invokeTool(createGitCommitNoEditTool(throwingExec, "/fake"), {})
    })
  })
})

// ---------------------------------------------------------------------------
// Tool args schema exposure regression tests (fix-tool-args-schema-mismatch)
// ---------------------------------------------------------------------------

describe("tool args schema exposure", () => {
  // The plugin SDK's tool() contract only recognizes `args` holding a
  // zod raw shape; a legacy `parameters` key is silently ignored, which
  // leaves `args` undefined — Object.keys() then throws and these tests
  // fail. A dropped or renamed parameter fails the sorted comparison.
  // Constructing a definition never invokes execute, so the injected
  // exec stays unused.

  it("exposes checkpoint as a zod-args tool definition", () => {
    const definition = createCheckpointTool(branchExec, "/fake")

    expect(Object.keys(definition.args).sort()).toEqual([
      "command",
      "data",
      "key",
      "phase",
      "step",
    ])
  })

  it("exposes get_progress as a zod-args tool definition", () => {
    const definition = createGetProgressTool(branchExec, "/fake")

    expect(Object.keys(definition.args).sort()).toEqual(["command", "key"])
  })

  it("exposes clear_progress as a zod-args tool definition", () => {
    const definition = createClearProgressTool(branchExec, "/fake")

    expect(Object.keys(definition.args).sort()).toEqual(["command", "key"])
  })

  it("exposes normalize_model_name as a zod-args tool definition", () => {
    const definition = createNormalizeModelNameTool()

    expect(Object.keys(definition.args).sort()).toEqual(["rawModelId"])
  })

  it("exposes git_fetch as a zod-args tool definition", () => {
    const definition = createGitFetchTool(branchExec, "/fake")

    expect(Object.keys(definition.args).sort()).toEqual(["branch", "remote"])
  })

  it("exposes git_push as a zod-args tool definition", () => {
    const definition = createGitPushTool(branchExec, "/fake")

    expect(Object.keys(definition.args).sort()).toEqual(["branch", "setUpstream"])
  })

  it("exposes git_merge as a zod-args tool definition", () => {
    const definition = createGitMergeTool(branchExec, "/fake")

    expect(Object.keys(definition.args).sort()).toEqual(["ref"])
  })

  it("exposes git_rebase as a zod-args tool definition", () => {
    const definition = createGitRebaseTool(branchExec, "/fake")

    expect(Object.keys(definition.args).sort()).toEqual(["ref"])
  })

  // Declared via zod before this change — kept as a guard so this tool
  // cannot regress to the legacy `parameters` key either.
  it("exposes the summary tool (persist compatibility factory) as a zod-args tool definition (future-regression guard)", () => {
    const definition = createPersistUnleashSummaryTool(branchExec, "/fake")

    expect(Object.keys(definition.args).sort()).toEqual([
      "change_name",
      "feature_dir",
      "howToVerify",
      "learnings_stored",
      "review_verdicts",
      "testResults",
      "tier",
      "whatWasBuilt",
    ])
  })
})
