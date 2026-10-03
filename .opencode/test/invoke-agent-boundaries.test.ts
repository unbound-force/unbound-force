import type { PluginInput, ToolContext } from "@opencode-ai/plugin"
import { describe, expect, it } from "vitest"

import { invokeAgent, type InvokeAgentDependencies } from "../plugins/invoke-agent/index.js"
import manifestFixture from "./fixtures/review-dispatch/manifest.json"

type Client = PluginInput["client"]
type CreateOptions = Parameters<Client["session"]["create"]>[0]
type PromptOptions = Parameters<Client["session"]["prompt"]>[0]

interface FakeClientOptions {
  readonly promptThrows?: unknown
  readonly createReturnsNoID?: boolean
}

function jsonParser(text: string): unknown {
  return JSON.parse(text) as unknown
}

function fakeClient(options: FakeClientOptions = {}): Client {
  const session = {
    create: async (request: CreateOptions): Promise<unknown> => {
      void request
      return { data: options.createReturnsNoID === true ? {} : { id: "child-session" }, error: undefined }
    },
    message: async (): Promise<unknown> => {
      return {
        data: {
          info: { id: "m", sessionID: "s", role: "assistant", providerID: "p", modelID: "m", variant: "active" },
          parts: [],
        },
        error: undefined,
      }
    },
    prompt: async (request: PromptOptions): Promise<unknown> => {
      if (options.promptThrows !== undefined) {
        throw options.promptThrows
      }
      void request
      return { data: { info: { role: "assistant", providerID: "p", modelID: "m" }, parts: [] }, error: undefined }
    },
    abort: async (): Promise<unknown> => ({ data: true, error: undefined }),
  }
  return { session } as unknown as Client
}

function dependencies(client: Client, timeoutMilliseconds = 1_000): InvokeAgentDependencies {
  return {
    client,
    readText: async (): Promise<string> => JSON.stringify(manifestFixture),
    parseYaml: jsonParser,
    timeoutMilliseconds,
  }
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

describe("invoke_agent error boundaries", () => {
  it.each([0, -1, 1_800_001, Number.NaN])("rejects an invalid injected timeout %s", async (timeout) => {
    const result = await invokeAgent(
      { agent: "divisor-guard", prompt: "Review.", model: "provider/model" },
      toolContext(),
      dependencies(fakeClient(), timeout),
    )
    expect(result.status).toBe("failed")
    expect(result.error?.code).toBe("invalid_invocation")
    expect(result.error?.retryable).toBe(false)
  })

  it("sanitizes a non-string thrown invocation failure into a bounded message", async () => {
    const circular: Record<string, unknown> = {}
    circular.self = circular
    const result = await invokeAgent(
      { agent: "divisor-guard", prompt: "Review.", model: "provider/model" },
      toolContext(),
      dependencies(fakeClient({ promptThrows: circular })),
    )
    expect(result.status).toBe("failed")
    expect(result.error?.code).toBe("invocation_failed")
    expect(result.error?.message).toBe("unknown invocation failure")
  })

  it("retains a null requested variant for an explicit model without variant", async () => {
    const result = await invokeAgent(
      { agent: "divisor-guard", prompt: "Review.", model: "p/m" },
      toolContext(),
      dependencies(fakeClient()),
    )
    expect(result.status).toBe("success")
    expect(result.provenance.requested_variant).toBeNull()
  })

  it("fails child creation when the client reports no child id", async () => {
    const result = await invokeAgent(
      { agent: "divisor-guard", prompt: "Review.", model: "provider/model" },
      toolContext(),
      dependencies(fakeClient({ createReturnsNoID: true })),
    )
    expect(result.status).toBe("failed")
    expect(result.error?.code).toBe("child_create_failed")
  })
})
