import type { PluginInput } from "@opencode-ai/plugin"
import { describe, expect, it } from "vitest"

import {
  failure,
  success,
  type ToolFailure,
  type ToolResult,
  type ToolSuccess,
} from "../lib/uf-workflow-types.js"
import UfWorkflowPlugin from "../plugins/uf-workflow/index.js"
import {
  createResolveBaseRefTool,
  type GitExec,
  type ResolveBaseRefData,
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

    const result = await UfWorkflowPlugin.server(fakeInput)
    expect(result).toBeDefined()
    expect(result.tool).toBeDefined()
    expect(Object.keys(result.tool)).toContain("resolve_base_ref")
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
    const mockExec: GitExec = (cmd: string) => {
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
      })
    }
  })

  it("falls back to origin/main when upstream/main fails", async () => {
    const mockExec: GitExec = (cmd: string) => {
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
      })
    }
  })

  it("falls back to main when remote-tracking refs fail", async () => {
    const mockExec: GitExec = (cmd: string) => {
      if (cmd.includes("upstream/main")) throw new Error("not found")
      if (cmd.includes("origin/main")) throw new Error("not found")
      if (cmd.endsWith("main")) return "789012abc345"
      throw new Error("not found")
    }

    const result = await invokeResolveBaseRef(mockExec)
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data).toStrictEqual({
        ref: "main",
        sha: "789012abc345",
        source: "main",
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
    const mockExec: GitExec = (cmd: string) => {
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
    const mockExec: GitExec = (cmd: string, cwd: string) => {
      capturedCwds.push(cwd)
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
