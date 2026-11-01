/**
 * Tests for the model-matrix classifier cascade (OpenSpec change
 * model-matrix-delegation — MM-FR-005/MM-FR-006 scenarios, design D5,
 * task 3.3).
 *
 * Every cascade behavior is exercised through the injected seams —
 * spawnAgent, fetch, readEnv, and the result cache — so no test touches
 * a live subagent session or a network endpoint (Constitution IV:
 * testability). Engine-bound tests use short real timers (20ms) inside
 * vitest's 5s per-test budget. The drift test pins the hidden subagent
 * definition on disk to the default agent constant, so the definition
 * and the code that names it cannot drift apart silently.
 */
import { readFileSync } from "node:fs"
import { fileURLToPath } from "node:url"
import { describe, expect, it } from "vitest"

import {
  CLASSIFIER_CACHE_CAPACITY,
  DEFAULT_AGENT_TIMEOUT_MS,
  DEFAULT_CLASSIFIER_AGENT,
  DEFAULT_LLM_REQUEST_TIMEOUT_MS,
  EMPTY_MODEL_MATRIX,
  MAX_LLM_RESPONSE_BYTES,
  ClassifierCascade,
  classifierCacheKey,
  createClassifierCache,
  parseModelMatrix,
  resolveSelection,
  type ClassifierCache,
  type ClassifierCacheEntry,
  type ClassifierRequest,
} from "../lib/model-matrix.js"

/**
 * Parser stub mirroring Bun.YAML for the shapes under test: comments-only
 * and empty documents parse to null; everything else is JSON (matching the
 * phase-2 model-matrix test convention).
 */
function testParser(text: string): unknown {
  const content = text
    .split(/\r?\n/)
    .filter((line) => line.trim() !== "" && !line.trim().startsWith("#"))
    .join("\n")
  if (content === "") {
    return null
  }
  return JSON.parse(text) as unknown
}

/** Parses a matrix document through the throwing parse boundary. */
function parse(document: unknown): ReturnType<typeof parseModelMatrix> {
  return parseModelMatrix(JSON.stringify(document), testParser)
}

/** Profiles for the classifier-test matrix: three modeled, one model-less. */
const BASE_PROFILES = {
  "spec-authoring": { model: "provider/spec-model" },
  "code-writing": { model: "provider/code-model" },
  "content-writing": { model: "provider/content-model" },
  mechanical: {},
}

/** Pack-to-profile mapping shared with the starter matrix (design D9). */
const PACKS = {
  "meta-prompt": "spec-authoring",
  go: "code-writing",
  typescript: "code-writing",
  python: "code-writing",
  content: "content-writing",
}

/**
 * Builds a configured matrix for cascade tests. `agents` defaults to empty
 * so the delegated agent name in IMPLEMENT_REQUEST is never pin-bypassed —
 * the pin-bypass row supplies its own map.
 * @param classifier The `classifier` stanza (defaults to empty: enabled).
 * @param agents The `agents` pin map (defaults to empty).
 * @returns The parsed, schema-valid matrix.
 */
function classifierMatrix(
  classifier: Record<string, unknown> = {},
  agents: Record<string, string> = {},
): ReturnType<typeof parseModelMatrix> {
  return parse({
    version: 1,
    profiles: BASE_PROFILES,
    packs: PACKS,
    defaults: { plan: "spec-authoring", tasks: "spec-authoring", implement: "code-writing" },
    agents,
    classifier,
  })
}

/** A local chat-completions endpoint config (design D4: gitignored override only). */
const LLM_CONFIG = {
  endpoint: "http://localhost:11434/v1/chat/completions",
  model: "granite-embedding:30m",
}

/** The delegation under classification — an implement call naming an agent. */
const IMPLEMENT_REQUEST: ClassifierRequest = {
  inputs: { checkpoint_step: "implement", agent: "cobalt-crush-dev" },
  feature_dir: "openspec/changes/model-matrix-delegation/",
  prompt: "Implement the classifier cascade for the model matrix delegation change.",
}

/** A fast agent-engine failure: prose that is not the JSON contract. */
const AGENT_MALFORMED = "The best pack here is go, I am fairly sure."

/** One logged agent-engine spawn. */
interface AgentCall {
  readonly agent: string
  readonly prompt: string
  readonly signal: AbortSignal
}

/** One logged LLM-engine fetch. */
interface FetchCall {
  readonly url: string
  readonly init: RequestInit | undefined
}

/** Scripted agent-engine behavior: a response text, a rejection, or a never-settling spawn. */
type AgentScriptEntry = string | Error | "pending"

/** The buildCascade return: the cascade plus every seam's call log. */
interface CascadeHarness {
  readonly cascade: ClassifierCascade
  readonly agentCalls: AgentCall[]
  readonly fetchCalls: FetchCall[]
  readonly envReads: string[]
}

/**
 * Builds a ClassifierCascade over logged, scripted seams. The agent engine
 * consumes `agentScript` entries in spawn order; an exhausted script
 * rejects loudly so an unexpected extra spawn cannot pass silently. The
 * default fetch impl rejects any request — tests that expect one provide
 * their own.
 */
function buildCascade(options: {
  readonly agentScript?: readonly AgentScriptEntry[]
  readonly fetchImpl?: typeof fetch
  readonly env?: Readonly<Record<string, string>>
  readonly cache?: ClassifierCache
} = {}): CascadeHarness {
  const agentCalls: AgentCall[] = []
  const fetchCalls: FetchCall[] = []
  const envReads: string[] = []
  const script = options.agentScript ?? []
  const spawnAgent = (agent: string, prompt: string, signal: AbortSignal): Promise<string> => {
    agentCalls.push({ agent, prompt, signal })
    const entry = script[agentCalls.length - 1]
    if (entry === undefined) {
      return Promise.reject(new Error("unexpected agent spawn — no scripted result remains"))
    }
    if (entry === "pending") {
      return new Promise<string>(() => {})
    }
    if (entry instanceof Error) {
      return Promise.reject(entry)
    }
    return Promise.resolve(entry)
  }
  const innerFetch =
    options.fetchImpl ??
    (async () => {
      throw new Error("unexpected fetch — the LLM engine was not expected to issue a request")
    })
  const loggedFetch: typeof fetch = async (input, init) => {
    fetchCalls.push({ url: String(input), init })
    return innerFetch(input, init)
  }
  const readEnv = (name: string): string | undefined => {
    envReads.push(name)
    return (options.env ?? {})[name]
  }
  const cascade = new ClassifierCascade({ spawnAgent, fetch: loggedFetch, readEnv, cache: options.cache })
  return { cascade, agentCalls, fetchCalls, envReads }
}

/** Builds a 200 chat-completions response whose single choice carries `content`. */
function chatResponse(content: string, status = 200): Response {
  return new Response(JSON.stringify({ choices: [{ message: { role: "assistant", content } }] }), {
    status,
    headers: { "content-type": "application/json" },
  })
}

describe("classifier cascade contract constants (MM-FR-005/MM-FR-006)", () => {
  it("pins the engine defaults, bounds, and cache capacity", () => {
    expect(DEFAULT_CLASSIFIER_AGENT).toBe("model-matrix-classifier")
    expect(DEFAULT_AGENT_TIMEOUT_MS).toBe(45_000)
    expect(DEFAULT_LLM_REQUEST_TIMEOUT_MS).toBe(10_000)
    expect(MAX_LLM_RESPONSE_BYTES).toBe(52_428_800)
    expect(CLASSIFIER_CACHE_CAPACITY).toBe(128)
  })

  it("keeps the subagent definition in sync with the default agent constant", () => {
    // Drift guard: the hidden agent file on disk must keep declaring the
    // name, subagent mode, hidden flag, read-only deny permissions, and the
    // strict JSON response contract the cascade parses.
    const definition = readFileSync(
      fileURLToPath(new URL("../agents/model-matrix-classifier.md", import.meta.url)),
      "utf8",
    )
    expect(definition).toContain(`name: ${DEFAULT_CLASSIFIER_AGENT}`)
    expect(definition).toContain("mode: subagent")
    expect(definition).toContain("hidden: true")
    expect(definition).toContain("  edit: deny")
    expect(definition).toContain("  bash: deny")
    expect(definition).toContain("  webfetch: deny")
    expect(definition).toContain('{"pack": "<name>", "confidence": <0.0-1.0>, "reasoning": "<optional>"}')
    expect(definition).toContain("No markdown fences")
  })
})

describe("cascade bypass predicate (MM-FR-003 — chains that never reach step 6)", () => {
  it("returns not-run without engines for explicit model, profile, and pack parameters", async () => {
    const { cascade, agentCalls, fetchCalls } = buildCascade()
    const matrix = classifierMatrix()
    for (const inputs of [
      { model: "provider/explicit-model" },
      { profile: "code-writing" },
      { pack: "go" },
    ]) {
      const outcome = await cascade.classify(matrix, { ...IMPLEMENT_REQUEST, inputs })
      expect(outcome.status).toBe("not-run")
      if (outcome.status === "not-run") {
        expect(outcome.reason).toContain("bypasses the classifier")
      }
    }
    expect(agentCalls.length).toBe(0)
    expect(fetchCalls.length).toBe(0)
  })

  it("returns not-run without engines for static plan and tasks kinds", async () => {
    const { cascade, agentCalls } = buildCascade()
    for (const checkpoint_step of ["plan", "tasks"] as const) {
      const outcome = await cascade.classify(classifierMatrix(), {
        ...IMPLEMENT_REQUEST,
        inputs: { checkpoint_step },
      })
      expect(outcome.status).toBe("not-run")
      if (outcome.status === "not-run") {
        expect(outcome.reason).toContain("resolves statically")
      }
    }
    expect(agentCalls.length).toBe(0)
  })

  it("returns not-run for a missing or no-op matrix", async () => {
    const { cascade, agentCalls } = buildCascade()
    for (const matrix of [null, EMPTY_MODEL_MATRIX]) {
      const outcome = await cascade.classify(matrix, IMPLEMENT_REQUEST)
      expect(outcome.status).toBe("not-run")
      if (outcome.status === "not-run") {
        expect(outcome.reason).toContain("no model matrix is configured")
      }
    }
    expect(agentCalls.length).toBe(0)
  })

  it("returns not-run when classifier.enabled is false (absent defaults to enabled)", async () => {
    const { cascade, agentCalls } = buildCascade()
    const outcome = await cascade.classify(classifierMatrix({ enabled: false }), IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("not-run")
    if (outcome.status === "not-run") {
      expect(outcome.reason).toContain("classifier.enabled: false")
    }
    expect(agentCalls.length).toBe(0)
  })

  it("returns not-run for an agent pin that yields before the classifier", async () => {
    const { cascade, agentCalls } = buildCascade()
    const pinned = classifierMatrix({}, { "cobalt-crush-dev": "code-writing" })
    const outcome = await cascade.classify(pinned, IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("not-run")
    if (outcome.status === "not-run") {
      expect(outcome.reason).toContain('agents pin for "cobalt-crush-dev"')
      expect(outcome.reason).toContain("yields before the classifier")
    }
    expect(agentCalls.length).toBe(0)
  })
})

describe("agent engine (MM-FR-005)", () => {
  it("selects the first valid above-threshold response and composes the selection", async () => {
    const { cascade, agentCalls, fetchCalls } = buildCascade({
      agentScript: ['{"pack":"go","confidence":0.9,"reasoning":"a code task"}'],
    })
    const matrix = classifierMatrix()
    const outcome = await cascade.classify(matrix, IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("selected")
    if (outcome.status === "selected") {
      expect(outcome.engine).toBe("agent")
      expect(outcome.pack).toBe("go")
      expect(outcome.confidence).toBe(0.9)
    }
    expect(agentCalls.length).toBe(1)
    expect(fetchCalls.length).toBe(0)
    const selection = resolveSelection(matrix, IMPLEMENT_REQUEST.inputs, outcome)
    expect(selection.ok).toBe(true)
    if (selection.ok) {
      expect(selection.selection.selection_source).toBe("classifier-agent")
      expect(selection.selection.pack).toBe("go")
      expect(selection.selection.capability).toBe("code-writing")
      expect(selection.selection.model).toBe("provider/code-model")
      expect(selection.selection.confidence).toBe(0.9)
    }
  })

  it("spawns the default hidden subagent with the sorted candidate packs and the task in the prompt", async () => {
    const { cascade, agentCalls } = buildCascade({ agentScript: ['{"pack":"go","confidence":0.9}'] })
    await cascade.classify(classifierMatrix(), IMPLEMENT_REQUEST)
    expect(agentCalls[0].agent).toBe(DEFAULT_CLASSIFIER_AGENT)
    expect(agentCalls[0].prompt).toContain("content, go, meta-prompt, python, typescript")
    expect(agentCalls[0].prompt).toContain(IMPLEMENT_REQUEST.prompt)
    expect(agentCalls[0].prompt).toContain("Feature directory: openspec/changes/model-matrix-delegation/")
    expect(agentCalls[0].prompt).toContain("JSON")
  })

  it("spawns the configured classifier.agent name", async () => {
    const { cascade, agentCalls } = buildCascade({ agentScript: ['{"pack":"go","confidence":0.9}'] })
    await cascade.classify(classifierMatrix({ agent: "custom-classifier" }), IMPLEMENT_REQUEST)
    expect(agentCalls[0].agent).toBe("custom-classifier")
  })

  it("omits the feature directory line when the request carries none", async () => {
    const { cascade, agentCalls } = buildCascade({ agentScript: ['{"pack":"go","confidence":0.9}'] })
    await cascade.classify(classifierMatrix(), { ...IMPLEMENT_REQUEST, feature_dir: null })
    expect(agentCalls[0].prompt).not.toContain("Feature directory:")
  })

  it("falls to the LLM engine when the agent engine exceeds agent_timeout_ms, aborting the spawn signal", async () => {
    const { cascade, agentCalls, fetchCalls } = buildCascade({
      agentScript: ["pending"],
      fetchImpl: async () => chatResponse('{"pack":"typescript","confidence":0.85}'),
    })
    const matrix = classifierMatrix({ agent_timeout_ms: 20, llm: LLM_CONFIG })
    const outcome = await cascade.classify(matrix, IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("selected")
    if (outcome.status === "selected") {
      expect(outcome.engine).toBe("llm")
      expect(outcome.pack).toBe("typescript")
    }
    expect(agentCalls.length).toBe(1)
    expect(agentCalls[0].signal.aborted).toBe(true)
    expect(fetchCalls.length).toBe(1)
  })

  it("records the spawn rejection reason when the agent engine fails", async () => {
    const { cascade } = buildCascade({ agentScript: [new Error("spawn transport failure")] })
    const outcome = await cascade.classify(classifierMatrix(), IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("failed")
    if (outcome.status === "failed") {
      expect(outcome.reason).toContain("agent engine: spawn transport failure")
      expect(outcome.reason).toContain("llm engine:")
    }
  })

  it("rejects prose and markdown-fenced responses as malformed", async () => {
    const { cascade, agentCalls } = buildCascade({
      agentScript: ["Sure! The best pack is go.", '```json\n{"pack":"go","confidence":0.9}\n```'],
    })
    for (let index = 0; index < 2; index++) {
      const outcome = await cascade.classify(classifierMatrix(), IMPLEMENT_REQUEST)
      expect(outcome.status).toBe("failed")
      if (outcome.status === "failed") {
        expect(outcome.reason).toContain("malformed response")
      }
    }
    expect(agentCalls.length).toBe(2)
  })

  it("rejects a matrix-invalid pack, falling to the next engine", async () => {
    const { cascade, agentCalls } = buildCascade({
      agentScript: ['{"pack":"no-such-pack","confidence":0.99}'],
      fetchImpl: async () => chatResponse('{"pack":"python","confidence":0.8}'),
    })
    const matrix = classifierMatrix({ llm: LLM_CONFIG })
    const outcome = await cascade.classify(matrix, IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("selected")
    if (outcome.status === "selected") {
      expect(outcome.engine).toBe("llm")
      expect(outcome.pack).toBe("python")
    }
    expect(agentCalls.length).toBe(1)
  })

  it("rejects a below-threshold response at the default threshold", async () => {
    const { cascade } = buildCascade({ agentScript: ['{"pack":"go","confidence":0.4}'] })
    const outcome = await cascade.classify(classifierMatrix(), IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("failed")
    if (outcome.status === "failed") {
      expect(outcome.reason).toContain("confidence 0.4 is below the threshold 0.6")
    }
  })

  it("rejects a below-threshold response at a configured threshold", async () => {
    const { cascade } = buildCascade({ agentScript: ['{"pack":"go","confidence":0.9}'] })
    const outcome = await cascade.classify(classifierMatrix({ confidence_threshold: 0.95 }), IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("failed")
    if (outcome.status === "failed") {
      expect(outcome.reason).toContain("confidence 0.9 is below the threshold 0.95")
    }
  })

  it("rejects an out-of-range or mistyped confidence as malformed", async () => {
    const { cascade, agentCalls } = buildCascade({
      agentScript: [
        '{"pack":"go","confidence":1.5}',
        '{"pack":"go","confidence":"high"}',
        '{"pack":"","confidence":0.9}',
      ],
    })
    for (let index = 0; index < 3; index++) {
      const outcome = await cascade.classify(classifierMatrix(), IMPLEMENT_REQUEST)
      expect(outcome.status).toBe("failed")
      if (outcome.status === "failed") {
        expect(outcome.reason).toContain("malformed response")
      }
    }
    expect(agentCalls.length).toBe(3)
  })

  it("rejects non-object JSON bodies and non-string reasoning as malformed", async () => {
    const { cascade, agentCalls } = buildCascade({
      agentScript: ['["an","array"]', '{"pack":"go","confidence":0.9,"reasoning":42}'],
    })
    for (let index = 0; index < 2; index++) {
      const outcome = await cascade.classify(classifierMatrix(), IMPLEMENT_REQUEST)
      expect(outcome.status).toBe("failed")
      if (outcome.status === "failed") {
        expect(outcome.reason).toContain("malformed response")
      }
    }
    expect(agentCalls.length).toBe(2)
  })

  it("rejects an unparseable endpoint URL before any request", async () => {
    const { cascade, fetchCalls } = buildCascade({ agentScript: [AGENT_MALFORMED] })
    const matrix = classifierMatrix({ llm: { ...LLM_CONFIG, endpoint: "not a valid url" } })
    const outcome = await cascade.classify(matrix, IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("failed")
    if (outcome.status === "failed") {
      expect(outcome.reason).toContain("is not a valid URL")
    }
    expect(fetchCalls.length).toBe(0)
  })

  it("tolerates extra response keys and selects confidence exactly at the threshold", async () => {
    const { cascade } = buildCascade({
      agentScript: ['{"pack":"go","confidence":0.9,"reasoning":"code task","extra":true,"nested":{"a":1}}'],
    })
    const outcome = await cascade.classify(classifierMatrix({ confidence_threshold: 0.9 }), IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("selected")
    if (outcome.status === "selected") {
      expect(outcome.engine).toBe("agent")
      expect(outcome.pack).toBe("go")
      expect(outcome.confidence).toBe(0.9)
    }
  })
})

describe("llm engine (MM-FR-006)", () => {
  it("selects the first valid above-threshold response and composes the selection", async () => {
    const { cascade, fetchCalls, agentCalls } = buildCascade({
      agentScript: [AGENT_MALFORMED],
      fetchImpl: async () => chatResponse('{"pack":"typescript","confidence":0.85}'),
    })
    const matrix = classifierMatrix({ llm: LLM_CONFIG })
    const outcome = await cascade.classify(matrix, IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("selected")
    if (outcome.status === "selected") {
      expect(outcome.engine).toBe("llm")
      expect(outcome.pack).toBe("typescript")
      expect(outcome.confidence).toBe(0.85)
    }
    expect(agentCalls.length).toBe(1)
    expect(fetchCalls.length).toBe(1)
    const selection = resolveSelection(matrix, IMPLEMENT_REQUEST.inputs, outcome)
    expect(selection.ok).toBe(true)
    if (selection.ok) {
      expect(selection.selection.selection_source).toBe("classifier-llm")
      expect(selection.selection.pack).toBe("typescript")
      expect(selection.selection.capability).toBe("code-writing")
      expect(selection.selection.model).toBe("provider/code-model")
      expect(selection.selection.confidence).toBe(0.85)
    }
  })

  it("issues the chat-completions POST with temperature 0, response_format json_object, and the bearer apiKey", async () => {
    const { cascade, fetchCalls } = buildCascade({
      agentScript: [AGENT_MALFORMED],
      env: { LLM_TEST_API_KEY: "sk-test-123" },
      fetchImpl: async () => chatResponse('{"pack":"go","confidence":0.9}'),
    })
    await cascade.classify(
      classifierMatrix({ llm: { ...LLM_CONFIG, apiKey: "{env:LLM_TEST_API_KEY}" } }),
      IMPLEMENT_REQUEST,
    )
    expect(fetchCalls.length).toBe(1)
    expect(fetchCalls[0].url).toBe("http://localhost:11434/v1/chat/completions")
    const init = fetchCalls[0].init
    expect(init?.method).toBe("POST")
    const headers = init?.headers as Record<string, string>
    expect(headers["Content-Type"]).toBe("application/json")
    expect(headers.Authorization).toBe("Bearer sk-test-123")
    const body = JSON.parse(String(init?.body)) as Record<string, unknown>
    expect(body.model).toBe("granite-embedding:30m")
    expect(body.temperature).toBe(0)
    expect(body.response_format).toEqual({ type: "json_object" })
    const messages = body.messages as Array<{ role: string; content: string }>
    expect(messages.length).toBe(1)
    expect(messages[0].role).toBe("user")
    expect(messages[0].content).toContain(IMPLEMENT_REQUEST.prompt)
    expect(messages[0].content).toContain("content, go, meta-prompt, python, typescript")
    expect(messages[0].content).toContain("JSON")
    expect(messages[0].content).toContain("Feature directory: openspec/changes/model-matrix-delegation/")
  })

  it("sends no Authorization header when no apiKey is configured", async () => {
    const { cascade, fetchCalls } = buildCascade({
      agentScript: [AGENT_MALFORMED],
      env: {},
      fetchImpl: async () => chatResponse('{"pack":"go","confidence":0.9}'),
    })
    await cascade.classify(classifierMatrix({ llm: LLM_CONFIG }), IMPLEMENT_REQUEST)
    expect(fetchCalls.length).toBe(1)
    const headers = fetchCalls[0].init?.headers as Record<string, string>
    expect(headers.Authorization).toBeUndefined()
  })

  it("skips without a request when the apiKey is a literal key, never echoing it", async () => {
    const { cascade, fetchCalls, envReads } = buildCascade({ agentScript: [AGENT_MALFORMED] })
    const matrix = classifierMatrix({ llm: { ...LLM_CONFIG, apiKey: "sk-literal-secret-value" } })
    const outcome = await cascade.classify(matrix, IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("failed")
    if (outcome.status === "failed") {
      expect(outcome.reason).toContain("literal key")
      expect(outcome.reason).not.toContain("sk-literal-secret-value")
    }
    expect(fetchCalls.length).toBe(0)
    expect(envReads.length).toBe(0)
  })

  it("skips without a request when the referenced environment variable is unset", async () => {
    const { cascade, fetchCalls, envReads } = buildCascade({ agentScript: [AGENT_MALFORMED], env: {} })
    const matrix = classifierMatrix({ llm: { ...LLM_CONFIG, apiKey: "{env:LLM_TEST_UNSET_KEY}" } })
    const outcome = await cascade.classify(matrix, IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("failed")
    if (outcome.status === "failed") {
      expect(outcome.reason).toContain("LLM_TEST_UNSET_KEY")
      expect(outcome.reason).toContain("not set")
    }
    expect(fetchCalls.length).toBe(0)
    expect(envReads).toEqual(["LLM_TEST_UNSET_KEY"])
  })

  it("retries exactly once without response_format when the endpoint rejects it with 400", async () => {
    let served = 0
    const { cascade, fetchCalls } = buildCascade({
      agentScript: [AGENT_MALFORMED],
      fetchImpl: async () => {
        served += 1
        return served === 1 ? chatResponse("response_format unsupported", 400) : chatResponse('{"pack":"go","confidence":0.9}')
      },
    })
    const outcome = await cascade.classify(classifierMatrix({ llm: LLM_CONFIG }), IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("selected")
    expect(fetchCalls.length).toBe(2)
    const first = JSON.parse(String(fetchCalls[0].init?.body)) as Record<string, unknown>
    const second = JSON.parse(String(fetchCalls[1].init?.body)) as Record<string, unknown>
    expect(first.response_format).toEqual({ type: "json_object" })
    expect(second.response_format).toBeUndefined()
    expect(second.model).toBe("granite-embedding:30m")
    expect(second.temperature).toBe(0)
  })

  it("does not retry a second 400 or any 500", async () => {
    const twice400 = buildCascade({
      agentScript: [AGENT_MALFORMED],
      fetchImpl: async () => chatResponse("bad request", 400),
    })
    const second400 = await twice400.cascade.classify(classifierMatrix({ llm: LLM_CONFIG }), IMPLEMENT_REQUEST)
    expect(second400.status).toBe("failed")
    if (second400.status === "failed") {
      expect(second400.reason).toContain("HTTP 400")
    }
    expect(twice400.fetchCalls.length).toBe(2)

    const once500 = buildCascade({
      agentScript: [AGENT_MALFORMED],
      fetchImpl: async () => chatResponse("internal error", 500),
    })
    const serverError = await once500.cascade.classify(classifierMatrix({ llm: LLM_CONFIG }), IMPLEMENT_REQUEST)
    expect(serverError.status).toBe("failed")
    if (serverError.status === "failed") {
      expect(serverError.reason).toContain("HTTP 500")
    }
    expect(once500.fetchCalls.length).toBe(1)
  })

  it("skips without a request when no endpoint is configured", async () => {
    const { cascade, fetchCalls } = buildCascade({ agentScript: [AGENT_MALFORMED] })
    const outcome = await cascade.classify(classifierMatrix(), IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("failed")
    if (outcome.status === "failed") {
      expect(outcome.reason).toContain("no endpoint configuration")
    }
    expect(fetchCalls.length).toBe(0)
  })

  it("rejects a non-loopback http endpoint before any request", async () => {
    const { cascade, fetchCalls } = buildCascade({ agentScript: [AGENT_MALFORMED] })
    const matrix = classifierMatrix({ llm: { ...LLM_CONFIG, endpoint: "http://192.168.1.5:8080/v1/chat/completions" } })
    const outcome = await cascade.classify(matrix, IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("failed")
    if (outcome.status === "failed") {
      expect(outcome.reason).toContain("must use the https scheme or an http loopback host")
    }
    expect(fetchCalls.length).toBe(0)
  })

  it("accepts https and http loopback endpoints", async () => {
    for (const endpoint of [
      "https://api.example.com/v1/chat/completions",
      "http://localhost:11434/v1/chat/completions",
      "http://127.0.0.1:8080/v1/chat/completions",
      "http://[::1]:9000/v1/chat/completions",
    ]) {
      const { cascade, fetchCalls } = buildCascade({
        agentScript: [AGENT_MALFORMED],
        fetchImpl: async () => chatResponse('{"pack":"go","confidence":0.9}'),
      })
      const outcome = await cascade.classify(classifierMatrix({ llm: { ...LLM_CONFIG, endpoint } }), IMPLEMENT_REQUEST)
      expect(outcome.status).toBe("selected")
      expect(fetchCalls.length).toBe(1)
      expect(fetchCalls[0].url).toBe(endpoint)
    }
  })

  it("records a non-timeout fetch failure with its error text", async () => {
    const { cascade, fetchCalls } = buildCascade({
      agentScript: [AGENT_MALFORMED],
      fetchImpl: async () => {
        throw new Error("connect ECONNREFUSED 127.0.0.1:11434")
      },
    })
    const outcome = await cascade.classify(classifierMatrix({ llm: LLM_CONFIG }), IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("failed")
    if (outcome.status === "failed") {
      expect(outcome.reason).toContain("request failed: connect ECONNREFUSED 127.0.0.1:11434")
    }
    expect(fetchCalls.length).toBe(1)
  })

  it("fails with a timeout reason when the request exceeds request_timeout_ms", async () => {
    const { cascade, fetchCalls } = buildCascade({
      agentScript: [AGENT_MALFORMED],
      fetchImpl: () =>
        new Promise<Response>(() => {
          // Never settles: the engine's timer race must bound the attempt
          // even when the fetch seam ignores the abort signal.
        }),
    })
    const matrix = classifierMatrix({ llm: { ...LLM_CONFIG, request_timeout_ms: 20 } })
    const outcome = await cascade.classify(matrix, IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("failed")
    if (outcome.status === "failed") {
      expect(outcome.reason).toContain("request timed out after 20ms")
    }
    expect(fetchCalls.length).toBe(1)
  })

  it("rejects a response whose declared content-length exceeds the cap without reading the body", async () => {
    let textCalled = false
    let cancelled = false
    const oversized = {
      ok: true,
      status: 200,
      headers: new Headers({ "content-length": String(MAX_LLM_RESPONSE_BYTES + 1) }),
      body: {
        cancel: async () => {
          cancelled = true
        },
      },
      text: () => {
        textCalled = true
        throw new Error("text() must not be called on a declared-oversized body")
      },
    } as unknown as Response
    const { cascade } = buildCascade({
      agentScript: [AGENT_MALFORMED],
      fetchImpl: async () => oversized,
    })
    const outcome = await cascade.classify(classifierMatrix({ llm: LLM_CONFIG }), IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("failed")
    if (outcome.status === "failed") {
      expect(outcome.reason).toContain("response body exceeds")
    }
    expect(textCalled).toBe(false)
    expect(cancelled).toBe(true)
  })

  it("reads a null-body response through the text() fallback", async () => {
    // Mock and empty responses can carry a null body; the reader falls
    // back to text() under the same byte cap.
    const content = '{"pack":"go","confidence":0.9}'
    const bodiless = {
      ok: true,
      status: 200,
      headers: new Headers(),
      body: null,
      text: async () => JSON.stringify({ choices: [{ message: { content } }] }),
    } as unknown as Response
    const { cascade } = buildCascade({
      agentScript: [AGENT_MALFORMED],
      fetchImpl: async () => bodiless,
    })
    const outcome = await cascade.classify(classifierMatrix({ llm: LLM_CONFIG }), IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("selected")
    if (outcome.status === "selected") {
      expect(outcome.engine).toBe("llm")
      expect(outcome.pack).toBe("go")
      expect(outcome.confidence).toBe(0.9)
    }
  })

  it("rejects a streaming body that exceeds the cap without reading it to the end", async () => {
    // Duck-typed reader whose first read carries no value (the defensive
    // no-byte continue), then 1 MiB chunks forever: the loop must stop at
    // the 52nd read (51 MiB of chunks > 50 MiB cap) instead of draining.
    const chunk = new Uint8Array(1024 * 1024)
    let reads = 0
    let cancelled = false
    const endless = {
      ok: true,
      status: 200,
      headers: new Headers(),
      body: {
        getReader: () => ({
          read: async () => {
            reads += 1
            if (reads === 1) {
              return { done: false, value: undefined }
            }
            return { done: false, value: chunk }
          },
          cancel: async () => {
            cancelled = true
          },
        }),
      },
    } as unknown as Response
    const { cascade } = buildCascade({
      agentScript: [AGENT_MALFORMED],
      fetchImpl: async () => endless,
    })
    const outcome = await cascade.classify(classifierMatrix({ llm: LLM_CONFIG }), IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("failed")
    if (outcome.status === "failed") {
      expect(outcome.reason).toContain("response body exceeds")
    }
    expect(reads).toBe(52)
    expect(cancelled).toBe(true)
  })

  it("rejects a response without a chat-completions choices[0].message.content string", async () => {
    const { cascade } = buildCascade({
      agentScript: [AGENT_MALFORMED],
      fetchImpl: async () => new Response('{"object": "list", "data": []}', { status: 200 }),
    })
    const outcome = await cascade.classify(classifierMatrix({ llm: LLM_CONFIG }), IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("failed")
    if (outcome.status === "failed") {
      expect(outcome.reason).toContain("not a chat-completions choices[0].message.content string")
    }
  })

  it("rejects each malformed chat-completions body shape", async () => {
    // Every defensive branch of the content extraction: unparseable body,
    // non-object body, empty choices, non-object choice, non-object
    // message, and non-string content.
    const malformedBodies = [
      "not json at all",
      '"a json string"',
      '{"choices": []}',
      '{"choices": ["not-an-object"]}',
      '{"choices": [{"message": "not-an-object"}]}',
      '{"choices": [{"message": {"content": 42}}]}',
    ]
    for (const body of malformedBodies) {
      const { cascade } = buildCascade({
        agentScript: [AGENT_MALFORMED],
        fetchImpl: async () => new Response(body, { status: 200 }),
      })
      const outcome = await cascade.classify(classifierMatrix({ llm: LLM_CONFIG }), IMPLEMENT_REQUEST)
      expect(outcome.status).toBe("failed")
      if (outcome.status === "failed") {
        expect(outcome.reason).toContain("not a chat-completions choices[0].message.content string")
      }
    }
  })
})

describe("cascade failure composition (MM-FR-005 — classifier failure never blocks delegation)", () => {
  it("composes both engines' failure reasons into one failed outcome without throwing", async () => {
    const { cascade } = buildCascade({ agentScript: [new Error("spawn transport failure")] })
    const outcome = await cascade.classify(classifierMatrix(), IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("failed")
    if (outcome.status === "failed") {
      expect(outcome.reason).toMatch(/^agent engine: .+; llm engine: .+$/)
      expect(outcome.reason).toContain("spawn transport failure")
      expect(outcome.reason).toContain("no endpoint configuration")
    }
  })

  it("falls back statically through resolveSelection when the cascade fails", async () => {
    const { cascade } = buildCascade({ agentScript: [AGENT_MALFORMED] })
    const matrix = classifierMatrix()
    const outcome = await cascade.classify(matrix, IMPLEMENT_REQUEST)
    expect(outcome.status).toBe("failed")
    const selection = resolveSelection(matrix, IMPLEMENT_REQUEST.inputs, outcome)
    expect(selection.ok).toBe(true)
    if (selection.ok) {
      expect(selection.selection.selection_source).toBe("default-kind")
      expect(selection.selection.model).toBe("provider/code-model")
      expect(selection.selection.fallback_reason).toContain("classifier cascade failed")
    }
  })
})

describe("result cache (design D5)", () => {
  it("returns the cached selection without spawning either engine", async () => {
    const { cascade, agentCalls, fetchCalls } = buildCascade({
      agentScript: ['{"pack":"go","confidence":0.9}'],
    })
    const matrix = classifierMatrix()
    const first = await cascade.classify(matrix, IMPLEMENT_REQUEST)
    expect(first.status).toBe("selected")
    const second = await cascade.classify(matrix, IMPLEMENT_REQUEST)
    expect(second.status).toBe("selected")
    if (second.status === "selected") {
      expect(second.engine).toBe("agent")
      expect(second.pack).toBe("go")
      expect(second.confidence).toBe(0.9)
    }
    expect(agentCalls.length).toBe(1)
    expect(fetchCalls.length).toBe(0)
  })

  it("distinguishes prompts, feature dirs, and agents in the cache key", async () => {
    const { cascade, agentCalls } = buildCascade({
      agentScript: ['{"pack":"go","confidence":0.9}', '{"pack":"go","confidence":0.9}', '{"pack":"go","confidence":0.9}'],
    })
    const matrix = classifierMatrix()
    await cascade.classify(matrix, IMPLEMENT_REQUEST)
    await cascade.classify(matrix, { ...IMPLEMENT_REQUEST, prompt: `${IMPLEMENT_REQUEST.prompt} (variant)` })
    await cascade.classify(matrix, { ...IMPLEMENT_REQUEST, feature_dir: "openspec/changes/other/" })
    expect(agentCalls.length).toBe(3)
  })

  it("evicts the oldest entry at capacity and re-runs the evicted request", async () => {
    const cache = createClassifierCache(2)
    const { cascade, agentCalls } = buildCascade({
      agentScript: Array.from({ length: 4 }, () => '{"pack":"go","confidence":0.9}'),
      cache,
    })
    const matrix = classifierMatrix()
    await cascade.classify(matrix, { ...IMPLEMENT_REQUEST, prompt: "task one" })
    await cascade.classify(matrix, { ...IMPLEMENT_REQUEST, prompt: "task two" })
    await cascade.classify(matrix, { ...IMPLEMENT_REQUEST, prompt: "task three" })
    expect(agentCalls.length).toBe(3)
    // "task one" was evicted admitting "task three" — the re-run spawns again.
    await cascade.classify(matrix, { ...IMPLEMENT_REQUEST, prompt: "task one" })
    expect(agentCalls.length).toBe(4)
    // "task three" survived the eviction that admitted the re-run — a hit.
    await cascade.classify(matrix, { ...IMPLEMENT_REQUEST, prompt: "task three" })
    expect(agentCalls.length).toBe(4)
  })

  it("defaults createClassifierCache to the pinned capacity", () => {
    const cache = createClassifierCache()
    const entry: ClassifierCacheEntry = { engine: "agent", pack: "go", confidence: 0.9 }
    for (let index = 0; index < CLASSIFIER_CACHE_CAPACITY; index++) {
      cache.set(`key-${index}`, entry)
    }
    expect(cache.get("key-0")).toEqual(entry)
    cache.set("key-new", entry)
    expect(cache.get("key-0")).toBeUndefined()
    expect(cache.get("key-1")).toEqual(entry)
    expect(cache.get("key-new")).toEqual(entry)
  })

  it("keeps an updated key's original eviction position", () => {
    const cache = createClassifierCache(2)
    const entry: ClassifierCacheEntry = { engine: "llm", pack: "go", confidence: 0.7 }
    cache.set("a", entry)
    cache.set("b", entry)
    const updated: ClassifierCacheEntry = { engine: "agent", pack: "typescript", confidence: 0.8 }
    cache.set("a", updated)
    expect(cache.get("a")).toEqual(updated)
    cache.set("c", entry)
    expect(cache.get("a")).toBeUndefined() // still the oldest — evicted by "c"
    expect(cache.get("b")).toEqual(entry)
    expect(cache.get("c")).toEqual(entry)
  })

  it("hashes the (feature_dir, prompt, agent) triple deterministically", () => {
    const first = classifierCacheKey("openspec/changes/a/", "prompt", "agent-one")
    expect(classifierCacheKey("openspec/changes/a/", "prompt", "agent-one")).toBe(first)
    expect(first).toMatch(/^[0-9a-f]{64}$/)
    expect(classifierCacheKey("openspec/changes/b/", "prompt", "agent-one")).not.toBe(first)
    expect(classifierCacheKey("openspec/changes/a/", "other prompt", "agent-one")).not.toBe(first)
    expect(classifierCacheKey("openspec/changes/a/", "prompt", "agent-two")).not.toBe(first)
    expect(classifierCacheKey(null, "prompt", undefined)).toBe(classifierCacheKey(null, "prompt", undefined))
  })

  it("never shares cache state between cascade instances", async () => {
    const first = buildCascade({ agentScript: ['{"pack":"go","confidence":0.9}'] })
    const second = buildCascade({ agentScript: ['{"pack":"go","confidence":0.9}'] })
    await first.cascade.classify(classifierMatrix(), IMPLEMENT_REQUEST)
    expect(first.agentCalls.length).toBe(1)
    await second.cascade.classify(classifierMatrix(), IMPLEMENT_REQUEST)
    expect(second.agentCalls.length).toBe(1)
  })
})
