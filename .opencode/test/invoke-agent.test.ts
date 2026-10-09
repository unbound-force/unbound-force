import type { PluginInput, ToolContext } from "@opencode-ai/plugin"
import { describe, expect, it } from "vitest"

import {
  createInvokeAgentTool,
  invokeAgent,
  sanitizeInvocationError,
  type InvokeAgentDependencies,
} from "../plugins/invoke-agent/index.js"
import { parseReviewMatrix } from "../plugins/review-dispatch/index.js"
import manifestFixture from "./fixtures/review-dispatch/manifest.json"
import matrixFixture from "./fixtures/review-dispatch/matrix.json"

type Client = PluginInput["client"]
type CreateOptions = Parameters<Client["session"]["create"]>[0]
type MessageOptions = Parameters<Client["session"]["message"]>[0]
type PromptOptions = Parameters<Client["session"]["prompt"]>[0]
type AbortOptions = Parameters<Client["session"]["abort"]>[0]
type CreateResponse = Awaited<ReturnType<Client["session"]["create"]>>
type MessageResponse = Awaited<ReturnType<Client["session"]["message"]>>
type PromptResponse = Awaited<ReturnType<Client["session"]["prompt"]>>
type AbortResponse = Awaited<ReturnType<Client["session"]["abort"]>>

interface FakeCalls {
  readonly creates: CreateOptions[]
  readonly messages: MessageOptions[]
  readonly prompts: PromptOptions[]
  readonly aborts: AbortOptions[]
}

interface FakeClientOptions {
  readonly parentProvider?: string
  readonly parentModel?: string
  readonly parentVariant?: string
  readonly parentRole?: "assistant" | "user"
  readonly messageError?: unknown
  readonly childProvider?: string
  readonly childModel?: string
  readonly childError?: unknown
  readonly parts?: readonly unknown[]
  readonly createError?: unknown
  readonly promptError?: unknown
  readonly hangPrompt?: boolean
}

function fieldsResponse<T>(data: T): T {
  return data
}

function assistantInfo(providerID: string, modelID: string, error?: unknown): Record<string, unknown> {
  return {
    id: "message-child",
    sessionID: "child-session",
    role: "assistant",
    time: { created: 1, completed: 2 },
    parentID: "message-parent",
    providerID,
    modelID,
    mode: "divisor-guard",
    path: { cwd: "/workspace", root: "/workspace" },
    cost: 0,
    tokens: { input: 0, output: 0, reasoning: 0, cache: { read: 0, write: 0 } },
    ...(error === undefined ? {} : { error }),
  }
}

function textPart(text: string, ignored = false): Record<string, unknown> {
  return {
    id: `text-${text}`,
    sessionID: "child-session",
    messageID: "message-child",
    type: "text",
    text,
    ignored,
  }
}

function finishPart(cost: number, input: number, output: number): Record<string, unknown> {
  return {
    id: `finish-${cost}`,
    sessionID: "child-session",
    messageID: "message-child",
    type: "step-finish",
    reason: "stop",
    cost,
    tokens: { input, output, reasoning: 0, cache: { read: 0, write: 0 } },
  }
}

function fakeClient(options: FakeClientOptions = {}): { readonly client: Client; readonly calls: FakeCalls } {
  const calls: FakeCalls = { creates: [], messages: [], prompts: [], aborts: [] }
  const session = {
    create: async (request: CreateOptions): Promise<CreateResponse> => {
      calls.creates.push(request)
      if (options.createError !== undefined) {
        return fieldsResponse({ data: undefined, error: options.createError } as unknown as CreateResponse)
      }
      return fieldsResponse({ data: { id: "child-session" }, error: undefined } as unknown as CreateResponse)
    },
    message: async (request: MessageOptions): Promise<MessageResponse> => {
      calls.messages.push(request)
      if (options.messageError !== undefined) {
        return fieldsResponse({ data: undefined, error: options.messageError } as unknown as MessageResponse)
      }
      const role = options.parentRole ?? "assistant"
      const info =
        role === "assistant"
          ? {
              ...assistantInfo(options.parentProvider ?? "host-provider", options.parentModel ?? "host-model"),
              variant: options.parentVariant,
            }
          : {
              id: "message-parent",
              sessionID: "parent-session",
              role: "user",
              time: { created: 1 },
              agent: "build",
              model: { providerID: "configured", modelID: "default" },
            }
      return fieldsResponse({ data: { info, parts: [] }, error: undefined } as unknown as MessageResponse)
    },
    prompt: async (request: PromptOptions): Promise<PromptResponse> => {
      calls.prompts.push(request)
      if (options.hangPrompt === true) {
        return await new Promise<PromptResponse>((_resolve, reject) => {
          request.signal?.addEventListener("abort", () => reject(request.signal?.reason), { once: true })
        })
      }
      if (options.promptError !== undefined) {
        return fieldsResponse({ data: undefined, error: options.promptError } as unknown as PromptResponse)
      }
      const bodyModel = request.body?.model
      const info = assistantInfo(
        options.childProvider ?? bodyModel?.providerID ?? "provider",
        options.childModel ?? bodyModel?.modelID ?? "model",
        options.childError,
      )
      return fieldsResponse({
        data: { info, parts: options.parts ?? [textPart("review output")] },
        error: undefined,
      } as unknown as PromptResponse)
    },
    abort: async (request: AbortOptions): Promise<AbortResponse> => {
      calls.aborts.push(request)
      return fieldsResponse({ data: true, error: undefined } as unknown as AbortResponse)
    },
  }
  return { client: { session } as unknown as Client, calls }
}

function jsonParser(text: string): unknown {
  return JSON.parse(text) as unknown
}

function dependencies(
  client: Client,
  manifest: unknown = manifestFixture,
  timeoutMilliseconds = 1_000,
  reads?: string[],
): InvokeAgentDependencies {
  return {
    client,
    readText: async (relativePath: string): Promise<string> => {
      reads?.push(relativePath)
      if (relativePath !== ".uf/reviewer-capabilities.yaml") {
        throw new Error(`unexpected policy or artifact operation: ${relativePath}`)
      }
      return JSON.stringify(manifest)
    },
    parseYaml: jsonParser,
    timeoutMilliseconds,
  }
}

function toolContext(abort = new AbortController()): ToolContext {
  return {
    sessionID: "parent-session",
    messageID: "message-parent",
    agent: "build",
    directory: "/workspace",
    worktree: "/workspace",
    abort: abort.signal,
    metadata: () => undefined,
    ask: (() => {
      throw new Error("permission prompts are not used by invoke_agent")
    }) as ToolContext["ask"],
  }
}

function explicitInput(overrides: Readonly<Record<string, unknown>> = {}): Record<string, unknown> {
  return {
    agent: "divisor-guard",
    prompt: "Review this change.",
    model: "provider/model",
    variant: "high",
    ...overrides,
  }
}

describe("invoke_agent explicit and host execution", () => {
  it("executes one explicit model run with top-level variant and separated provenance", async () => {
    const fake = fakeClient()
    const result = await invokeAgent(explicitInput(), toolContext(), dependencies(fake.client))

    expect(result.status).toBe("success")
    expect(result.text).toBe("review output")
    expect(result.provenance).toEqual({
      agent: "divisor-guard",
      requested_model: "provider/model",
      requested_variant: "high",
      resolved_parent_model: null,
      resolved_parent_variant: null,
      reported_child_model: "provider/model",
      model_mismatch: false,
      read_only: true,
    })
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates[0]?.body).toMatchObject({ parentID: "parent-session" })
    expect(fake.calls.prompts[0]?.body).toMatchObject({
      agent: "divisor-guard",
      model: { providerID: "provider", modelID: "model" },
      variant: "high",
    })
    expect(fake.calls.prompts[0]?.body).not.toHaveProperty("tools")
    expect(fake.calls.prompts[0]?.body).not.toHaveProperty("permission")
  })

  it("resolves the current assistant model and active variant for a host run", async () => {
    const fake = fakeClient({ parentProvider: "actual-provider", parentModel: "actual/model", parentVariant: "active" })
    const result = await invokeAgent(
      { agent: "divisor-testing", prompt: "Review tests.", read_only: false },
      toolContext(),
      dependencies(fake.client),
    )

    expect(result.status).toBe("success")
    expect(fake.calls.messages[0]).toMatchObject({
      path: { id: "parent-session", messageID: "message-parent" },
      query: { directory: "/workspace" },
    })
    expect(fake.calls.prompts[0]?.body).toMatchObject({
      model: { providerID: "actual-provider", modelID: "actual/model" },
      variant: "active",
    })
    expect(result.provenance).toMatchObject({
      requested_model: null,
      requested_variant: null,
      resolved_parent_model: "actual-provider/actual/model",
      resolved_parent_variant: "active",
      reported_child_model: "actual-provider/actual/model",
      read_only: false,
    })
  })

  it.each([
    ["missing message", { messageError: "service unavailable token=secret" }],
    ["non-assistant message", { parentRole: "user" as const }],
    ["missing model", { parentProvider: "" }],
    ["missing variant", { parentVariant: undefined }],
    ["invalid variant", { parentVariant: "not valid" }],
  ])("fails host resolution without substituting defaults for %s", async (_name, options) => {
    const fake = fakeClient(options)
    const result = await invokeAgent(
      { agent: "divisor-guard", prompt: "Host review." },
      toolContext(),
      dependencies(fake.client),
    )

    expect(result.status).toBe("failed")
    expect(result.error?.code).toBe("host_model_unavailable")
    expect(result.error?.retryable).toBe(true)
    expect(result.error?.message).not.toContain("secret")
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
  })
})

describe("invoke_agent validation boundaries", () => {
  it("accepts the intentionally loose first-slash direct model that strict matrix grammar rejects", async () => {
    const fake = fakeClient()
    const direct = "provider/model?candidate=1"
    const result = await invokeAgent(explicitInput({ model: direct }), toolContext(), dependencies(fake.client))
    const matrix = structuredClone(matrixFixture)
    matrix.profiles.standard.model = direct

    expect(result.status).toBe("success")
    expect(fake.calls.prompts[0]?.body?.model).toEqual({ providerID: "provider", modelID: "model?candidate=1" })
    expect(() => parseReviewMatrix(JSON.stringify(matrix), jsonParser)).toThrow()
  })

  it.each([
    ["agent format", explicitInput({ agent: "guard" })],
    ["agent length", explicitInput({ agent: `divisor-${"a".repeat(129)}` })],
    ["prompt bytes", explicitInput({ prompt: "💥".repeat(32_769) })],
    ["model length", explicitInput({ model: `provider/${"m".repeat(248)}` })],
    ["leading slash", explicitInput({ model: "/model" })],
    ["trailing slash", explicitInput({ model: "provider/" })],
    ["variant length", explicitInput({ variant: "v".repeat(65) })],
    ["variant format", explicitInput({ variant: "not valid" })],
    ["host requested variant", { agent: "divisor-guard", prompt: "Review.", variant: "high" }],
  ])("rejects bounded or malformed %s before child creation", async (_name, input) => {
    const fake = fakeClient()
    const result = await invokeAgent(input, toolContext(), dependencies(fake.client))

    expect(result.status).toBe("failed")
    expect(result.error?.code).toBe("invalid_input")
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("allows only manifested review agents and performs no selection or artifact operations", async () => {
    const reads: string[] = []
    const reviewFake = fakeClient()
    const optedInFake = fakeClient()
    const contentFake = fakeClient()
    const unknownFake = fakeClient()
    const optedInManifest = {
      ...structuredClone(manifestFixture),
      reviewers: [
        ...structuredClone(manifestFixture.reviewers),
        { agent: "divisor-specialist", capability: "review", scopes: ["standard"] },
      ],
    }
    const review = await invokeAgent(
      explicitInput(),
      toolContext(),
      dependencies(reviewFake.client, manifestFixture, 1_000, reads),
    )
    const content = await invokeAgent(
      explicitInput({ agent: "divisor-envoy" }),
      toolContext(),
      dependencies(contentFake.client, manifestFixture, 1_000, reads),
    )
    const unknown = await invokeAgent(
      explicitInput({ agent: "divisor-unknown" }),
      toolContext(),
      dependencies(unknownFake.client, manifestFixture, 1_000, reads),
    )
    const optedIn = await invokeAgent(
      explicitInput({ agent: "divisor-specialist" }),
      toolContext(),
      dependencies(optedInFake.client, optedInManifest, 1_000, reads),
    )

    expect(review.status).toBe("success")
    expect(optedIn.status).toBe("success")
    expect(content.error?.message).toContain("content capability")
    expect(unknown.error?.message).toContain("absent from the reviewer manifest")
    expect(contentFake.calls.creates).toHaveLength(0)
    expect(unknownFake.calls.creates).toHaveLength(0)
    expect(reads).toEqual([
      ".uf/reviewer-capabilities.yaml",
      ".uf/reviewer-capabilities.yaml",
      ".uf/reviewer-capabilities.yaml",
      ".uf/reviewer-capabilities.yaml",
    ])
  })
})

describe("invoke_agent cancellation, timeout, and responses", () => {
  it("propagates parent cancellation to the active child with an independent cleanup signal", async () => {
    const controller = new AbortController()
    const fake = fakeClient({ hangPrompt: true })
    const pending = invokeAgent(explicitInput(), toolContext(controller), dependencies(fake.client, manifestFixture, 1_000))
    await new Promise((resolve) => setTimeout(resolve, 0))
    controller.abort(new Error("parent stopped"))
    const result = await pending

    expect(result.status).toBe("cancelled")
    expect(result.error?.code).toBe("cancelled")
    expect(fake.calls.aborts).toHaveLength(1)
    expect(fake.calls.aborts[0]?.path).toEqual({ id: "child-session" })
    expect(fake.calls.aborts[0]?.signal).not.toBe(controller.signal)
  })

  it("enforces the injected bounded per-run timeout and aborts the active child", async () => {
    const fake = fakeClient({ hangPrompt: true })
    const result = await invokeAgent(explicitInput(), toolContext(), dependencies(fake.client, manifestFixture, 5))

    expect(result.status).toBe("failed")
    expect(result.error).toMatchObject({ code: "timeout", retryable: true })
    expect(fake.calls.aborts).toHaveLength(1)
  })

  it("honors a model-supplied timeout through the rawInput boundary", async () => {
    const fake = fakeClient({ hangPrompt: true })
    const result = await invokeAgent(
      explicitInput({ timeout: 5 }),
      toolContext(),
      dependencies(fake.client, manifestFixture, 60_000),
    )

    expect(result.status).toBe("failed")
    expect(result.error).toMatchObject({ code: "timeout", retryable: true })
    expect(fake.calls.aborts).toHaveLength(1)
  })

  it("extracts ordered non-ignored text and sums all reported finish usage", async () => {
    const fake = fakeClient({
      parts: [
        textPart(" first "),
        textPart("ignored", true),
        finishPart(0.25, 10, 4),
        textPart("second"),
        finishPart(0, 3, 2),
      ],
    })
    const result = await invokeAgent(explicitInput(), toolContext(), dependencies(fake.client))

    expect(result.text).toBe("first \n\nsecond")
    expect(result.usage).toEqual({ cost_usd: 0.25, tokens: { input: 13, output: 6 } })
  })

  it("keeps usage null when OpenCode reports no step-finish part", async () => {
    const fake = fakeClient({ parts: [textPart("only text")] })
    const result = await invokeAgent(explicitInput(), toolContext(), dependencies(fake.client))

    expect(result.status).toBe("success")
    expect(result.usage).toBeNull()
  })

  it("fails and reports requested-versus-child model mismatch without rewriting provenance", async () => {
    const fake = fakeClient({ childProvider: "other", childModel: "reported" })
    const result = await invokeAgent(explicitInput(), toolContext(), dependencies(fake.client))

    expect(result.status).toBe("failed")
    expect(result.error?.code).toBe("model_mismatch")
    expect(result.provenance).toMatchObject({
      requested_model: "provider/model",
      reported_child_model: "other/reported",
      model_mismatch: true,
    })
  })

  it("retains child text and provenance while sanitizing child message errors", async () => {
    const fake = fakeClient({
      childError: {
        name: "APIError",
        data: {
          message: "Bearer top-secret at https://provider.example/run?api_key=value /Users/alice/private/key",
          isRetryable: true,
        },
      },
      parts: [textPart("partial response"), finishPart(1, 2, 3)],
    })
    const result = await invokeAgent(explicitInput(), toolContext(), dependencies(fake.client))

    expect(result.status).toBe("failed")
    expect(result.error).toMatchObject({ code: "child_message_error", retryable: true })
    expect(result.error?.message).not.toMatch(/top-secret|api_key=value|\/Users\/alice/)
    expect(result.text).toBe("partial response")
    expect(result.usage).toEqual({ cost_usd: 1, tokens: { input: 2, output: 3 } })
    expect(result.provenance.reported_child_model).toBe("provider/model")
  })

  it.each([
    ["create", { createError: "credential=abc /home/alice/private" }, "child_create_failed"],
    ["prompt", { promptError: "Bearer abc https://host/path?token=value" }, "child_prompt_failed"],
  ])("sanitizes %s response failures", async (_name, options, code) => {
    const fake = fakeClient(options)
    const result = await invokeAgent(explicitInput(), toolContext(), dependencies(fake.client))

    expect(result.error?.code).toBe(code)
    expect(result.error?.message).not.toMatch(/abc|token=value|\/home\/alice/)
  })
})

describe("invoke_agent metadata and tool surface", () => {
  it.each([true, false])("retains read_only=%s without changing permissions or invocation", async (readOnly) => {
    const fake = fakeClient()
    const result = await invokeAgent(
      explicitInput({ read_only: readOnly }),
      toolContext(),
      dependencies(fake.client),
    )

    expect(result.status).toBe("success")
    expect(result.provenance.read_only).toBe(readOnly)
    expect(fake.calls.creates).toHaveLength(1)
    expect(fake.calls.prompts).toHaveLength(1)
    expect(fake.calls.creates[0]).not.toHaveProperty("permission")
    expect(fake.calls.prompts[0]?.body).not.toHaveProperty("permission")
    expect(fake.calls.prompts[0]?.body).not.toHaveProperty("tools")
  })

  it("exposes invoke_agent as a provider-free tool definition", () => {
    const fake = fakeClient()
    const definition = createInvokeAgentTool(dependencies(fake.client))

    expect(definition.description).toContain("one already-planned")
    expect(Object.keys(definition.args).sort()).toEqual([
      "agent",
      "model",
      "prompt",
      "promptFile",
      "read_only",
      "timeout",
      "variant",
    ])
  })

  it("redacts every required sensitive error class", () => {
    const sanitized = sanitizeInvocationError(
      'password="hunter two" Bearer bearer-value https://example.test/path?secret=value /Users/alice/private C:\\Users\\bob\\secret',
    )

    expect(sanitized).toContain("[REDACTED]")
    expect(sanitized).not.toMatch(/hunter two|bearer-value|secret=value|\/Users\/alice|C:\\Users\\bob/)
  })
})
