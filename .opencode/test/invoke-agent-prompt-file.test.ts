import type { PluginInput, ToolContext } from "@opencode-ai/plugin"
import { describe, expect, it, vi } from "vitest"

import { invokeAgent, type InvokeAgentDependencies } from "../plugins/invoke-agent/index.js"
import manifestFixture from "./fixtures/review-dispatch/manifest.json"

type Client = PluginInput["client"]

function jsonParser(text: string): unknown {
  return JSON.parse(text) as unknown
}

/**
 * Minimal fake client that succeeds for explicit-model runs.
 * Prompt response echoes the requested model identity so
 * model-mismatch checks pass.
 */
function fakeClient(): Client {
  const session = {
    create: async (): Promise<unknown> => ({
      data: { id: "child-session" },
      error: undefined,
    }),
    message: async (): Promise<unknown> => ({
      data: {
        info: {
          id: "m",
          sessionID: "s",
          role: "assistant",
          providerID: "p",
          modelID: "m",
          variant: "active",
        },
        parts: [],
      },
      error: undefined,
    }),
    prompt: async (): Promise<unknown> => ({
      data: {
        info: { role: "assistant", providerID: "p", modelID: "m" },
        parts: [
          { id: "t", sessionID: "child-session", messageID: "m", type: "text", text: "review output", ignored: false },
        ],
      },
      error: undefined,
    }),
    abort: async (): Promise<unknown> => ({ data: true, error: undefined }),
  }
  return { session } as unknown as Client
}

function toolContext(): ToolContext {
  return {
    sessionID: "parent-session",
    messageID: "message-parent",
    agent: "build",
    directory: "/workspace",
    worktree: "/workspace",
    abort: new AbortController().signal,
    metadata: () => undefined,
    ask: (() => {
      throw new Error("permission prompts are not used")
    }) as ToolContext["ask"],
  }
}

describe("invoke_agent promptFile", () => {
  it("rejects when both prompt and promptFile are provided", async () => {
    const readText = vi.fn(async () => JSON.stringify(manifestFixture))
    const deps: InvokeAgentDependencies = {
      client: fakeClient(),
      readText,
      parseYaml: jsonParser,
      timeoutMilliseconds: 1_000,
    }
    const result = await invokeAgent(
      { agent: "divisor-guard", prompt: "inline text", promptFile: "/tmp/prompt.txt", model: "p/m" },
      toolContext(),
      deps,
    )

    expect(result.status).toBe("failed")
    expect(result.error?.code).toBe("invalid_input")
    expect(result.error?.message).toContain("prompt and promptFile are mutually exclusive")
  })

  it("rejects when neither prompt nor promptFile is provided", async () => {
    const readText = vi.fn(async () => JSON.stringify(manifestFixture))
    const deps: InvokeAgentDependencies = {
      client: fakeClient(),
      readText,
      parseYaml: jsonParser,
      timeoutMilliseconds: 1_000,
    }
    const result = await invokeAgent({ agent: "divisor-guard", model: "p/m" }, toolContext(), deps)

    expect(result.status).toBe("failed")
    expect(result.error?.code).toBe("invalid_input")
    expect(result.error?.message).toContain("one of prompt or promptFile is required")
  })

  it("rejects promptFile paths exceeding 1024 characters", async () => {
    const readText = vi.fn(async () => JSON.stringify(manifestFixture))
    const deps: InvokeAgentDependencies = {
      client: fakeClient(),
      readText,
      parseYaml: jsonParser,
      timeoutMilliseconds: 1_000,
    }
    const result = await invokeAgent(
      { agent: "divisor-guard", promptFile: "/tmp/" + "a".repeat(1020), model: "p/m" },
      toolContext(),
      deps,
    )

    expect(result.status).toBe("failed")
    expect(result.error?.code).toBe("invalid_input")
  })

  it("reads prompt text from promptFile and executes the session successfully", async () => {
    const readText = vi.fn(async (path: string): Promise<string> => {
      if (path === "/tmp/prompt.txt") return "File-based prompt content."
      if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
      throw new Error(`unexpected read: ${path}`)
    })
    const deps: InvokeAgentDependencies = {
      client: fakeClient(),
      readText,
      parseYaml: jsonParser,
      timeoutMilliseconds: 1_000,
    }
    const result = await invokeAgent(
      { agent: "divisor-guard", promptFile: "/tmp/prompt.txt", model: "p/m" },
      toolContext(),
      deps,
    )

    expect(result.status).toBe("success")
    expect(result.text).toBe("review output")
    expect(readText).toHaveBeenCalledWith("/tmp/prompt.txt")
    expect(readText).toHaveBeenCalledWith(".uf/reviewer-capabilities.yaml")
    expect(readText).toHaveBeenCalledTimes(2)
  })

  it("returns a failed result with sanitized error when promptFile read fails", async () => {
    const readText = vi.fn(async (path: string): Promise<string> => {
      if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
      throw new Error("ENOENT: no such file /Users/alice/secret/prompt.txt")
    })
    const deps: InvokeAgentDependencies = {
      client: fakeClient(),
      readText,
      parseYaml: jsonParser,
      timeoutMilliseconds: 1_000,
    }
    const result = await invokeAgent(
      { agent: "divisor-guard", promptFile: "/tmp/missing.txt", model: "p/m" },
      toolContext(),
      deps,
    )

    expect(result.status).toBe("failed")
    expect(result.error?.code).toBe("prompt_file_read_failed")
    expect(result.error?.retryable).toBe(false)
    // Verify path sanitization — /Users/alice should be redacted.
    expect(result.error?.message).not.toContain("/Users/alice")
    expect(result.error?.message).toContain("[REDACTED]")
  })

  it("rejects promptFile content exceeding 4 MiB", async () => {
    const oversizedContent = "x".repeat(4 * 1024 * 1024 + 1)
    const readText = vi.fn(async (path: string): Promise<string> => {
      if (path === "/tmp/big-prompt.txt") return oversizedContent
      if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
      throw new Error(`unexpected read: ${path}`)
    })
    const deps: InvokeAgentDependencies = {
      client: fakeClient(),
      readText,
      parseYaml: jsonParser,
      timeoutMilliseconds: 1_000,
    }
    const result = await invokeAgent(
      { agent: "divisor-guard", promptFile: "/tmp/big-prompt.txt", model: "p/m" },
      toolContext(),
      deps,
    )

    expect(result.status).toBe("failed")
    expect(result.error?.code).toBe("invalid_input")
    expect(result.error?.message).toContain("promptFile content exceeds 4 MiB")
    expect(result.error?.retryable).toBe(false)
  })

  it("accepts an inline prompt for backward compatibility", async () => {
    const readText = vi.fn(async (path: string): Promise<string> => {
      if (path === ".uf/reviewer-capabilities.yaml") return JSON.stringify(manifestFixture)
      throw new Error(`unexpected read: ${path}`)
    })
    const deps: InvokeAgentDependencies = {
      client: fakeClient(),
      readText,
      parseYaml: jsonParser,
      timeoutMilliseconds: 1_000,
    }
    const result = await invokeAgent(
      { agent: "divisor-guard", prompt: "Inline prompt text.", model: "p/m" },
      toolContext(),
      deps,
    )

    expect(result.status).toBe("success")
    expect(result.text).toBe("review output")
    // readText should only be called for the manifest, not for any prompt file.
    expect(readText).toHaveBeenCalledTimes(1)
    expect(readText).toHaveBeenCalledWith(".uf/reviewer-capabilities.yaml")
  })
})
