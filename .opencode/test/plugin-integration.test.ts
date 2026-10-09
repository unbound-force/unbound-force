import { mkdir, mkdtemp, readFile, rm, writeFile } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"

import type { PluginInput, ToolContext } from "@opencode-ai/plugin"
import { afterEach, describe, expect, it } from "vitest"

import InvokeAgentPlugin from "../plugins/invoke-agent/index.js"
import ReviewDispatchPlugin from "../plugins/review-dispatch/index.js"
import manifestFixture from "./fixtures/review-dispatch/manifest.json"
import matrixFixture from "./fixtures/review-dispatch/matrix.json"
import samplePayload from "../../schemas/review-dispatch/samples/sample-review-dispatch.json"

type Client = PluginInput["client"]
type CreateOptions = Parameters<Client["session"]["create"]>[0]
type MessageOptions = Parameters<Client["session"]["message"]>[0]
type PromptOptions = Parameters<Client["session"]["prompt"]>[0]
type AbortOptions = Parameters<Client["session"]["abort"]>[0]

interface FakeCalls {
  readonly creates: CreateOptions[]
  readonly messages: MessageOptions[]
  readonly prompts: PromptOptions[]
  readonly aborts: AbortOptions[]
}

interface FakeClient {
  readonly client: Client
  readonly calls: FakeCalls
}

const scratchDirectories: string[] = []

afterEach(async () => {
  await Promise.all(scratchDirectories.splice(0).map((directory) => rm(directory, { recursive: true, force: true })))
})

function withBun<T>(run: () => T | Promise<T>): Promise<T> {
  const runtime = globalThis as typeof globalThis & {
    Bun?: { readonly YAML: { readonly parse: (text: string) => unknown } }
  }
  const original = runtime.Bun
  runtime.Bun = { YAML: { parse: (text: string): unknown => JSON.parse(text) as unknown } }
  const finish = (): void => {
    runtime.Bun = original
  }
  try {
    return Promise.resolve(run()).finally(finish)
  } catch (error) {
    finish()
    throw error
  }
}

function assistantInfo(providerID: string, modelID: string): Record<string, unknown> {
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
  }
}

function fakeClient(): FakeClient {
  const calls: FakeCalls = { creates: [], messages: [], prompts: [], aborts: [] }
  const session = {
    create: async (request: CreateOptions): Promise<unknown> => {
      calls.creates.push(request)
      return { data: { id: "child-session" }, error: undefined }
    },
    message: async (request: MessageOptions): Promise<unknown> => {
      calls.messages.push(request)
      return {
        data: {
          info: { ...assistantInfo("host-provider", "host/model"), variant: "active" },
          parts: [],
        },
        error: undefined,
      }
    },
    prompt: async (request: PromptOptions): Promise<unknown> => {
      calls.prompts.push(request)
      const bodyModel = request.body?.model as { providerID?: string; modelID?: string } | undefined
      return {
        data: {
          info: assistantInfo(bodyModel?.providerID ?? "provider", bodyModel?.modelID ?? "model"),
          parts: [{ id: "text-1", sessionID: "child-session", messageID: "message-child", type: "text", text: "integrated review", ignored: false }],
        },
        error: undefined,
      }
    },
    abort: async (request: AbortOptions): Promise<unknown> => {
      calls.aborts.push(request)
      return { data: true, error: undefined }
    },
  }
  return { client: { session } as unknown as Client, calls }
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

async function scratchTargetProject(): Promise<string> {
  const directory = await mkdtemp(join(tmpdir(), "uf-plugin-integration-"))
  scratchDirectories.push(directory)
  await mkdir(join(directory, ".uf"), { recursive: true })
  await writeFile(join(directory, ".uf", "review-matrix.yaml"), JSON.stringify(matrixFixture))
  await writeFile(join(directory, ".uf", "reviewer-capabilities.yaml"), JSON.stringify(manifestFixture))
  return directory
}

describe("plugin integration with a fake OpenCode client", () => {
  it("registers both plugin tool sets on activation", async () => {
    await withBun(async () => {
      const fake = fakeClient()
      const invokeHooks = await InvokeAgentPlugin.server({
        client: fake.client,
        directory: "/workspace",
        worktree: "/workspace",
      } as PluginInput)
      const dispatchHooks = await ReviewDispatchPlugin.server({
        directory: "/workspace",
        worktree: "/workspace",
      } as PluginInput)

      expect(Object.keys(invokeHooks.tool)).toEqual(["invoke_agent"])
      expect(Object.keys(dispatchHooks.tool).sort()).toEqual([
        "acquire_sibling_evidence",
        "consolidate_dispatch",
        "dispatch_agent_run",
        "dispatch_status",
        "finalize_review_dispatch",
        "plan_review_dispatch",
        "prepare_lesson_learning",
        "submit_lesson_proposal",
        "submit_review_findings",
      ])
    })
  })

  it("executes an explicit run through the registered invoke_agent tool", async () => {
    await withBun(async () => {
      const project = await scratchTargetProject()
      const fake = fakeClient()
      const hooks = await InvokeAgentPlugin.server({ client: fake.client, directory: project, worktree: project } as PluginInput)

      const response = await hooks.tool.invoke_agent.execute(
        { agent: "divisor-guard", prompt: "Review this.", model: "provider/model", variant: "high" },
        toolContext(),
      )

      expect(response.output).toBe("integrated review")
      expect(response.metadata).toMatchObject({ status: "success", provenance: { agent: "divisor-guard" } })
      expect(fake.calls.creates[0]?.body).toMatchObject({ parentID: "parent-session" })
      expect(fake.calls.prompts[0]?.body).toMatchObject({
        agent: "divisor-guard",
        model: { providerID: "provider", modelID: "model" },
        variant: "high",
      })
      expect(fake.calls.prompts[0]?.body).not.toHaveProperty("tools")
    })
  })

  it("resolves the current assistant model and replays explicit host provenance", async () => {
    await withBun(async () => {
      const project = await scratchTargetProject()
      const fake = fakeClient()
      const hooks = await InvokeAgentPlugin.server({ client: fake.client, directory: project, worktree: project } as PluginInput)

      const response = await hooks.tool.invoke_agent.execute(
        { agent: "divisor-testing", prompt: "Review tests." },
        toolContext(),
      )

      expect(response.metadata).toMatchObject({
        status: "success",
        provenance: {
          requested_model: null,
          requested_variant: null,
          resolved_parent_model: "host-provider/host/model",
          resolved_parent_variant: "active",
          reported_child_model: "host-provider/host/model",
        },
      })
      expect(fake.calls.messages[0]).toMatchObject({
        path: { id: "parent-session", messageID: "message-parent" },
      })
      expect(fake.calls.prompts[0]?.body).toMatchObject({
        model: { providerID: "host-provider", modelID: "host/model" },
        variant: "active",
      })
    })
  })

  it("plans and finalizes through the registered review-dispatch tools", async () => {
    await withBun(async () => {
      const project = await scratchTargetProject()
      const hooks = await ReviewDispatchPlugin.server({ directory: project, worktree: project } as PluginInput)

      const planOutput = await hooks.tool.plan_review_dispatch.execute(
        {
          mode: "code",
          discovered_agents: ["divisor-guard", "divisor-adversary"],
          changed_files: [{ path: "internal/feature.go", additions: 20, deletions: 0 }],
        },
        {} as never,
      )
      const plan = JSON.parse(planOutput) as { status: string; entries: unknown[] }
      expect(plan.status).toBe("ready")
      expect(plan.entries.length).toBeGreaterThan(0)

      const finalizeOutput = await hooks.tool.finalize_review_dispatch.execute(
        {
          payload: structuredClone(samplePayload),
          provenance: {
            branch: "feature/review-fanout",
            commit: "2222222222222222222222222222222222222222",
            workflow_id: "integration-1",
          },
        },
        {} as never,
      )
      const finalized = JSON.parse(finalizeOutput) as { status: string; review_verdict: { council_decision: string } }
      expect(finalized.status).toBe("success")
      expect(finalized.review_verdict.council_decision).toBe("CHANGES_REQUESTED")
      expect(await readFile(join(project, ".uf/artifacts/dispatch/aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa.json"), "utf8")).toContain(
        '"hero": "the-divisor"',
      )
    })
  })
})
