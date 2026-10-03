import { Buffer } from "node:buffer"
import { readFile } from "node:fs/promises"
import { resolve } from "node:path"

import { tool, type PluginInput, type PluginModule, type ToolContext } from "@opencode-ai/plugin"
import { z } from "zod"

import {
  createBunYamlParser,
  parseReviewerManifest,
  type YamlParser,
} from "../../lib/reviewer-manifest.js"

const MANIFEST_PATH = ".uf/reviewer-capabilities.yaml"
const MAX_PROMPT_BYTES = 128 * 1024
const DEFAULT_TIMEOUT_MILLISECONDS = 600_000
const MAX_ERROR_LENGTH = 4_096
const CLEANUP_ABORT_TIMEOUT_MS = 5_000

const AgentSchema = z.string().max(128).regex(/^divisor-[a-z0-9-]{1,63}$/)
const PromptSchema = z
  .string()
  .refine((value) => Buffer.byteLength(value, "utf8") <= MAX_PROMPT_BYTES, "prompt exceeds 128 KiB UTF-8")
const ModelSchema = z
  .string()
  .max(256)
  .refine((value) => {
    const separator = value.indexOf("/")
    return separator > 0 && separator < value.length - 1
  }, "model must contain non-empty provider and model-id values separated by the first slash")
const VariantSchema = z.string().max(64).regex(/^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$/)
const TimeoutSchema = z.number().int().positive().max(1_800_000)
const InvokeAgentInputSchema = z
  .object({
    agent: AgentSchema,
    prompt: PromptSchema,
    model: ModelSchema.optional(),
    variant: VariantSchema.optional(),
    read_only: z.boolean().optional(),
    timeout: TimeoutSchema.optional(),
  })
  .strict()
  .superRefine((input, context) => {
    if (input.model === undefined && input.variant !== undefined) {
      context.addIssue({
        code: "custom",
        path: ["variant"],
        message: "variant requires an explicit model; host runs replay the active parent variant",
      })
    }
  })

type PromptOptions = Parameters<PluginInput["client"]["session"]["prompt"]>[0]
type PromptBodyWithVariant = NonNullable<PromptOptions["body"]> & {
  variant?: string
}
type PromptResponse = Awaited<ReturnType<PluginInput["client"]["session"]["prompt"]>>
type PromptPart = NonNullable<PromptResponse["data"]>["parts"][number]

type CurrentMessageResponse = Awaited<ReturnType<PluginInput["client"]["session"]["message"]>>
type CurrentMessageInfo = NonNullable<CurrentMessageResponse["data"]>["info"]
type CurrentAssistant = Extract<CurrentMessageInfo, { role: "assistant" }>
type CurrentAssistantWithVariant<T> = T & {
  variant?: string
}

type ValidatedInput = z.output<typeof InvokeAgentInputSchema>

interface ModelIdentity {
  readonly providerID: string
  readonly modelID: string
  readonly combined: string
}

/** Token counts reported by OpenCode step-finish response parts. */
export interface InvocationTokens {
  readonly input: number
  readonly output: number
}

/** Nullable invocation usage assembled only from reported step-finish parts. */
export interface InvocationUsage {
  readonly cost_usd: number | null
  readonly tokens: InvocationTokens | null
}

/** Sanitized failure details for one invoke_agent run. */
export interface InvocationError {
  readonly code: string
  readonly message: string
  readonly retryable: boolean
}

/** Request, host-resolution, and child-report provenance for one invocation. */
export interface InvocationProvenance {
  readonly agent: string
  readonly requested_model: string | null
  readonly requested_variant: string | null
  readonly resolved_parent_model: string | null
  readonly resolved_parent_variant: string | null
  readonly reported_child_model: string | null
  readonly model_mismatch: boolean
  readonly read_only: boolean
}

/** Observable result from exactly one policy-free agent invocation. */
export interface InvokeAgentResult {
  readonly status: "success" | "failed" | "cancelled"
  readonly text: string
  readonly usage: InvocationUsage | null
  readonly error: InvocationError | null
  readonly provenance: InvocationProvenance
}

/** Injectable I/O, client, parser, and timeout boundary used by invoke_agent. */
export interface InvokeAgentDependencies {
  readonly client: PluginInput["client"]
  readonly readText: (relativePath: string) => Promise<string>
  readonly parseYaml: YamlParser
  readonly timeoutMilliseconds: number
}

interface ResponseErrorLike {
  readonly error?: unknown
  readonly response?: { readonly status: number }
}

function errorText(error: unknown): string {
  if (error instanceof Error) {
    return error.message
  }
  if (typeof error === "string") {
    return error
  }
  try {
    return JSON.stringify(error) ?? "unknown invocation failure"
  } catch {
    return "unknown invocation failure"
  }
}

/**
 * Redacts credentials, bearer values, URL queries, and absolute home paths from an error.
 * @param message Untrusted provider, runtime, or validation error text.
 * @returns A bounded message safe to retain in invocation provenance.
 */
export function sanitizeInvocationError(message: string): string {
  const sanitized = message
    .replace(/\bBearer\s+[^\s,;"']+/gi, "Bearer [REDACTED]")
    .replace(
      /\b(api[-_ ]?key|access[-_ ]?token|refresh[-_ ]?token|token|secret|password|credential)s?\b(\s*[:=]\s*)(?:"[^"]*"|'[^']*'|[^\s,;"']+)/gi,
      "$1$2[REDACTED]",
    )
    .replace(/(https?:\/\/[^\s?#]+)\?[^\s#]*/gi, "$1?[REDACTED]")
    .replace(/(?:\/Users|\/home)\/[^/\s]+(?:\/[^\s,;:"']*)?/g, "~\/[REDACTED]")
    .replace(/[A-Za-z]:\\+Users\\+[^\\\s]+(?:\\+[^\s,;:"']*)?/g, "~\\[REDACTED]")
  return sanitized.slice(0, MAX_ERROR_LENGTH)
}

function responseFailure(label: string, response: ResponseErrorLike): string {
  const detail = response.error ?? response.response?.status ?? "missing response data"
  return sanitizeInvocationError(`${label}: ${errorText(detail)}`)
}

function directModelIdentity(model: string): ModelIdentity {
  const separator = model.indexOf("/")
  return {
    providerID: model.slice(0, separator),
    modelID: model.slice(separator + 1),
    combined: model,
  }
}

function reportedModel(providerID: string, modelID: string): string | null {
  return providerID.length > 0 && modelID.length > 0 ? `${providerID}/${modelID}` : null
}

function initialProvenance(input: ValidatedInput): InvocationProvenance {
  return {
    agent: input.agent,
    requested_model: input.model ?? null,
    requested_variant: input.model === undefined ? null : (input.variant ?? null),
    resolved_parent_model: null,
    resolved_parent_variant: null,
    reported_child_model: null,
    model_mismatch: false,
    read_only: input.read_only ?? true,
  }
}

function failedResult(
  provenance: InvocationProvenance,
  code: string,
  message: string,
  retryable: boolean,
  status: "failed" | "cancelled" = "failed",
  text = "",
  usage: InvocationUsage | null = null,
): InvokeAgentResult {
  return {
    status,
    text,
    usage,
    error: { code, message: sanitizeInvocationError(message), retryable },
    provenance,
  }
}

async function validateReviewAgent(input: ValidatedInput, dependencies: InvokeAgentDependencies): Promise<void> {
  const text = await dependencies.readText(MANIFEST_PATH)
  const manifest = parseReviewerManifest(text, dependencies.parseYaml)
  const reviewer = manifest.reviewers.find((candidate) => candidate.agent === input.agent)
  if (reviewer === undefined) {
    throw new Error(`agent ${input.agent} is absent from the reviewer manifest`)
  }
  if (reviewer.capability !== "review") {
    throw new Error(`agent ${input.agent} has content capability and cannot perform a review invocation`)
  }
}

async function resolveHostModel(
  context: ToolContext,
  dependencies: InvokeAgentDependencies,
  signal: AbortSignal,
): Promise<{ readonly identity: ModelIdentity; readonly variant: string }> {
  const response = await dependencies.client.session.message({
    path: { id: context.sessionID, messageID: context.messageID },
    query: { directory: context.directory },
    signal,
  })
  if (response.data === undefined) {
    throw new Error(responseFailure("current assistant message is unavailable", response))
  }
  if (response.data.info.role !== "assistant") {
    throw new Error("current message is not an assistant message")
  }

  const assistant = response.data.info as CurrentAssistantWithVariant<CurrentAssistant>
  const combined = reportedModel(assistant.providerID, assistant.modelID)
  if (combined === null || combined.length > 256) {
    throw new Error("current assistant model identity is unavailable")
  }
  const variant = VariantSchema.safeParse(assistant.variant)
  if (!variant.success) {
    throw new Error("current assistant active variant is unavailable or invalid")
  }
  return {
    identity: { providerID: assistant.providerID, modelID: assistant.modelID, combined },
    variant: variant.data,
  }
}

function invocationUsage(parts: readonly PromptPart[]): InvocationUsage | null {
  const finishes = parts.filter(
    (part): part is Extract<PromptPart, { type: "step-finish" }> => part.type === "step-finish",
  )
  if (finishes.length === 0) {
    return null
  }
  return finishes.reduce<InvocationUsage>(
    (usage, finish) => ({
      cost_usd: (usage.cost_usd ?? 0) + finish.cost,
      tokens: {
        input: (usage.tokens?.input ?? 0) + finish.tokens.input,
        output: (usage.tokens?.output ?? 0) + finish.tokens.output,
      },
    }),
    { cost_usd: 0, tokens: { input: 0, output: 0 } },
  )
}

function responseText(parts: readonly PromptPart[]): string {
  return parts
    .filter(
      (part): part is Extract<PromptPart, { type: "text" }> => part.type === "text" && part.ignored !== true,
    )
    .map((part) => part.text)
    .join("\n\n")
    .trim()
}

function validateTimeout(timeoutMilliseconds: number): void {
  if (!Number.isSafeInteger(timeoutMilliseconds) || timeoutMilliseconds <= 0 || timeoutMilliseconds > 1_800_000) {
    throw new Error("invoke_agent timeout must be a positive integer no greater than 1800000 milliseconds")
  }
}

/**
 * Executes exactly one manifested Divisor review-agent run.
 * @param rawInput Untrusted invocation arguments from the OpenCode tool boundary.
 * @param context Calling tool context used for parenting, host resolution, and cancellation.
 * @param dependencies Injected client, manifest reader/parser, and bounded timeout.
 * @returns Child text, nullable reported usage, and separated request/runtime provenance.
 */
export async function invokeAgent(
  rawInput: unknown,
  context: ToolContext,
  dependencies: InvokeAgentDependencies,
): Promise<InvokeAgentResult> {
  const parsed = InvokeAgentInputSchema.safeParse(rawInput)
  if (!parsed.success) {
    const placeholder: InvocationProvenance = {
      agent: "invalid",
      requested_model: null,
      requested_variant: null,
      resolved_parent_model: null,
      resolved_parent_variant: null,
      reported_child_model: null,
      model_mismatch: false,
      read_only: true,
    }
    return failedResult(placeholder, "invalid_input", parsed.error.message, false)
  }

  const input = parsed.data
  const timeoutMilliseconds = input.timeout ?? dependencies.timeoutMilliseconds
  let provenance = initialProvenance(input)
  try {
    validateTimeout(timeoutMilliseconds)
    await validateReviewAgent(input, dependencies)
  } catch (error: unknown) {
    return failedResult(provenance, "invalid_invocation", errorText(error), false)
  }

  const linkedAbort = new AbortController()
  let abortCause: "cancelled" | "timeout" | null = null
  let childID: string | null = null
  let cleanupAbort = Promise.resolve()

  const abortChild = (): void => {
    if (childID === null) {
      return
    }
    const id = childID
    cleanupAbort = cleanupAbort.then(async () => {
      const cleanupController = new AbortController()
      const cleanupTimeout = setTimeout(() => cleanupController.abort(), CLEANUP_ABORT_TIMEOUT_MS)
      try {
        await dependencies.client.session.abort({
          path: { id },
          query: { directory: context.directory },
          signal: cleanupController.signal,
        })
      } catch {
        // Cleanup is best-effort because the primary cancellation result remains authoritative.
      } finally {
        clearTimeout(cleanupTimeout)
      }
    })
  }
  const cancelFromParent = (): void => {
    abortCause = "cancelled"
    linkedAbort.abort(context.abort.reason)
    abortChild()
  }
  const timeout = setTimeout(() => {
    abortCause = "timeout"
    linkedAbort.abort(new Error("invoke_agent run timed out"))
    abortChild()
  }, timeoutMilliseconds)

  if (context.abort.aborted) {
    cancelFromParent()
  } else {
    context.abort.addEventListener("abort", cancelFromParent, { once: true })
  }

  try {
    let selectedModel: ModelIdentity
    let selectedVariant: string | undefined
    if (input.model === undefined) {
      let host
      try {
        host = await resolveHostModel(context, dependencies, linkedAbort.signal)
      } catch (error: unknown) {
        if (abortCause !== null) {
          throw error
        }
        return failedResult(provenance, "host_model_unavailable", errorText(error), true)
      }
      selectedModel = host.identity
      selectedVariant = host.variant
      provenance = {
        ...provenance,
        resolved_parent_model: host.identity.combined,
        resolved_parent_variant: host.variant,
      }
    } else {
      selectedModel = directModelIdentity(input.model)
      selectedVariant = input.variant
    }

    const created = await dependencies.client.session.create({
      body: {
        parentID: context.sessionID,
        title: `${input.agent} @ ${selectedModel.combined}${selectedVariant === undefined ? "" : ` (${selectedVariant})`}`,
      },
      query: { directory: context.directory },
      signal: linkedAbort.signal,
    })
    childID = created.data?.id ?? null
    if (childID === null) {
      return failedResult(provenance, "child_create_failed", responseFailure("create child session", created), true)
    }
    if (linkedAbort.signal.aborted) {
      abortChild()
      throw linkedAbort.signal.reason
    }

    const body: PromptBodyWithVariant = {
      agent: input.agent,
      model: { providerID: selectedModel.providerID, modelID: selectedModel.modelID },
      ...(selectedVariant === undefined ? {} : { variant: selectedVariant }),
      parts: [{ type: "text", text: input.prompt }],
    }
    const prompted = await dependencies.client.session.prompt({
      path: { id: childID },
      query: { directory: context.directory },
      body,
      signal: linkedAbort.signal,
    })
    if (prompted.data === undefined) {
      return failedResult(provenance, "child_prompt_failed", responseFailure("prompt child session", prompted), true)
    }

    const childModel = reportedModel(prompted.data.info.providerID, prompted.data.info.modelID)
    const mismatch = childModel !== selectedModel.combined
    provenance = { ...provenance, reported_child_model: childModel, model_mismatch: mismatch }
    const text = responseText(prompted.data.parts)
    const usage = invocationUsage(prompted.data.parts)

    if (prompted.data.info.error !== undefined) {
      return failedResult(
        provenance,
        "child_message_error",
        errorText(prompted.data.info.error),
        "isRetryable" in prompted.data.info.error.data
          ? prompted.data.info.error.data.isRetryable
          : prompted.data.info.error.name !== "MessageAbortedError",
        "failed",
        text,
        usage,
      )
    }
    if (mismatch) {
      return failedResult(
        provenance,
        "model_mismatch",
        `child reported ${childModel ?? "an unavailable model"}; expected ${selectedModel.combined}`,
        false,
        "failed",
        text,
        usage,
      )
    }

    return { status: "success", text, usage, error: null, provenance }
  } catch (error: unknown) {
    if (abortCause === "timeout") {
      return failedResult(provenance, "timeout", "invoke_agent run timed out", true)
    }
    if (abortCause === "cancelled") {
      return failedResult(provenance, "cancelled", "invoke_agent run was cancelled", false, "cancelled")
    }
    return failedResult(provenance, "invocation_failed", errorText(error), true)
  } finally {
    clearTimeout(timeout)
    context.abort.removeEventListener("abort", cancelFromParent)
    await cleanupAbort
  }
}

/**
 * Creates the invoke_agent OpenCode tool with isolated runtime dependencies.
 * @param dependencies Client, manifest, parser, and timeout dependencies.
 * @returns The policy-free OpenCode tool definition.
 */
export function createInvokeAgentTool(dependencies: InvokeAgentDependencies): ReturnType<typeof tool> {
  return tool({
    description:
      "Execute one already-planned Divisor review-agent run in a parented child session with explicit model provenance.",
    args: {
      agent: AgentSchema.describe("Manifest-declared divisor-* review agent."),
      prompt: PromptSchema.describe("Bounded full prompt for this single planned run."),
      model: ModelSchema.optional().describe("Optional direct provider/model-id selected by the dispatch plan."),
      variant: VariantSchema.optional().describe("Optional direct-model runtime variant."),
      read_only: z.boolean().optional().describe("Invocation provenance only; does not alter permissions."),
      timeout: TimeoutSchema.optional().describe("Optional run timeout in milliseconds (bounded to 1_800_000)."),
    },
    async execute(args, context): Promise<{ readonly output: string; readonly metadata: Record<string, unknown> }> {
      const result = await invokeAgent(args, context, dependencies)
      return {
        output:
          result.error === null
            ? result.text || "(no text produced)"
            : `invoke_agent error: ${result.error.message}`,
        metadata: { ...result },
      }
    },
  })
}

/** Auto-discovered OpenCode plugin exposing the policy-free invoke_agent tool. */
export const InvokeAgentPlugin = {
  id: "invoke-agent",
  server: async (input) => {
    const projectRoot = input.worktree || input.directory
    const dependencies: InvokeAgentDependencies = {
      client: input.client,
      readText: async (relativePath: string): Promise<string> =>
        readFile(resolve(projectRoot, relativePath), "utf8"),
      parseYaml: createBunYamlParser(),
      timeoutMilliseconds: DEFAULT_TIMEOUT_MILLISECONDS,
    }
    return { tool: { invoke_agent: createInvokeAgentTool(dependencies) } }
  },
} satisfies PluginModule

export default InvokeAgentPlugin
