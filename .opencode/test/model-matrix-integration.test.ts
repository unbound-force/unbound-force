/**
 * Model-matrix delegation-tool integration tests (OpenSpec change
 * model-matrix-delegation, tasks 5.5 and 5.6; designs D2/D5/D6/D7/D8).
 *
 * Task 5.5 — the AC1-AC8 acceptance coverage issue #887 names, plus the
 * UF-FR-002 input-shape boundary rows and the matrix-model
 * host-introspection skip, exercising the REAL delegation tool against
 * the REAL lib selection module (loadModelMatrix, ClassifierCascade,
 * resolveSelection) with only the environment seams faked: filesystem
 * reads (matrix + override + command files served through readText),
 * pack-file existence (fileExists), the hidden classifier spawn, and the
 * OpenCode client (fakeClient driving the real executeAgentSession
 * executor, so the child session's model identity is observable in the
 * recorded create/prompt calls).
 *
 * Task 5.6 — the DE-007 seam verification: (a) the registration surface
 * (the tool advertises model/variant/profile/pack and its input schema
 * accepts them), (b) the documented output shape (the MM-FR-007 audit
 * fields land in the checkpoint data exactly as specified), and (c)
 * write→read round-trips (a matrix written with
 * schemas/model-matrix/samples/-shaped content loads through
 * loadModelMatrix and round-trips its selected capability into
 * resolveSelection output).
 *
 * The zero-config regression coverage (AC4/AC8 host-fallback rows for
 * matrix-less repositories) lives in unleash-delegate.test.ts, which
 * threads matrixFreeDeps through every pre-existing test site; this file
 * covers the matrix-CONFIGURED behaviors that only exist when the
 * selection layer engages (Constitution IV — every behavior testable
 * without a network, a live subagent, or a real filesystem).
 */
import { lstatSync, mkdirSync, readFileSync, writeFileSync } from "node:fs"
import { resolve } from "node:path"
import { fileURLToPath } from "node:url"

import type { PluginInput } from "@opencode-ai/plugin"
import { beforeEach, describe, expect, it, vi } from "vitest"

import type { ToolResult } from "../lib/uf-workflow-types.js"
import {
  loadModelMatrix,
  resolveSelection,
  type ModelMatrixDependencies,
} from "../lib/model-matrix.js"
import {
  createDelegateCommandTool,
  type DelegateCommandData,
  type DelegateCommandDependencies,
  type FileExists,
  type GitExec,
  type ProgressState,
} from "../plugins/uf-workflow/index.js"

// ---------------------------------------------------------------------------
// Module-level mocks (hoisted by vitest) — the unleash-delegate.test.ts
// recipe: the auto-checkpoint path persists through readState/writeState,
// which hit node:fs directly, so mocking the four functions that path uses
// keeps every test off the real filesystem. The 5.6(c) round-trip test
// needs the REAL readFileSync for the samples fixture and recovers it via
// vi.importActual inside the test.
// ---------------------------------------------------------------------------
vi.mock("node:fs", async (importOriginal) => {
  const actual = await importOriginal<typeof import("node:fs")>()
  return {
    ...actual,
    readFileSync: vi.fn(),
    lstatSync: vi.fn(),
    mkdirSync: vi.fn(),
    writeFileSync: vi.fn(),
  }
})

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

/** Fake repository root — never a real path on any filesystem. */
const FAKE_CWD = "/fake/project"

/** The zero-config matrix probe and its override companion, absolute. */
const MATRIX_PROBE = resolve(FAKE_CWD, ".uf/model-matrix.yaml")
const OVERRIDE_PROBE = resolve(FAKE_CWD, ".uf/model-matrix.override.yaml")

/**
 * Tracked matrix fixture (the task 4.1 starter shape, reduced to the
 * packs under test): spec-authoring/code-writing carry models via the
 * override (D9 — identities arrive per-developer), content-writing stays
 * model-less so a profile resolution can fall back to host with the
 * capability still recorded.
 */
const TRACKED_MATRIX = JSON.stringify({
  version: 1,
  profiles: {
    "spec-authoring": {},
    "code-writing": {},
    "content-writing": {},
  },
  packs: {
    "meta-prompt": "spec-authoring",
    go: "code-writing",
    content: "content-writing",
  },
  defaults: {
    plan: "spec-authoring",
    tasks: "spec-authoring",
    implement: "code-writing",
  },
  agents: {},
  classifier: {},
})

/**
 * Per-developer override fixture (the gitignored
 * .uf/model-matrix.override.yaml path, design D4): whitelisted
 * profiles.*.model leaves only.
 */
const MODEL_OVERRIDE = JSON.stringify({
  profiles: {
    "spec-authoring": { model: "override-provider/spec-model" },
    "code-writing": { model: "override-provider/code-model" },
  },
})

/** Pack files the loader's pack-existence check (MM-FR-001) discovers. */
const PACK_FILES = [
  ".opencode/uf/packs/meta-prompt.md",
  ".opencode/uf/packs/go.md",
  ".opencode/uf/packs/content.md",
]

/** Host model identity the fake parent session reports. */
const HOST_MODEL = "host-provider/host-model"

// ---------------------------------------------------------------------------
// Test helpers (unleash-delegate.test.ts harness, duplicated per the
// per-file isolation convention — see TC-002/TC-004)
// ---------------------------------------------------------------------------

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
  /** Parent assistant message reports no model identity (structural failure). */
  readonly hostModelMissing?: boolean
}

function fieldsResponse<T>(data: T): T {
  return data
}

/** Assistant message info carrying only the fields the executor reads. */
function assistantInfo(providerID: string, modelID: string): Record<string, unknown> {
  return {
    id: "message-child",
    sessionID: "child-session",
    role: "assistant",
    time: { created: 1, completed: 2 },
    parentID: "message-parent",
    providerID,
    modelID,
    mode: "primary",
    path: { cwd: FAKE_CWD, root: FAKE_CWD },
    cost: 0,
    tokens: { input: 0, output: 0, reasoning: 0, cache: { read: 0, write: 0 } },
  }
}

function textPart(text: string): Record<string, unknown> {
  return {
    id: `text-${text}`,
    sessionID: "child-session",
    messageID: "message-child",
    type: "text",
    text,
  }
}

/**
 * Fake OpenCode client (invoke-agent.test.ts precedent) driving the REAL
 * executeAgentSession executor. The child reports the model it was asked
 * to run, so successful delegations never trip the model-mismatch guard,
 * and the recorded prompt body exposes the model the selection layer
 * chose — the observable contract AC1 asserts on.
 */
function fakeClient(options: FakeClientOptions = {}): { readonly client: Client; readonly calls: FakeCalls } {
  const calls: FakeCalls = { creates: [], messages: [], prompts: [], aborts: [] }
  const session = {
    create: async (request: CreateOptions): Promise<CreateResponse> => {
      calls.creates.push(request)
      return fieldsResponse({ data: { id: "child-session" }, error: undefined } as unknown as CreateResponse)
    },
    message: async (request: MessageOptions): Promise<MessageResponse> => {
      calls.messages.push(request)
      return fieldsResponse({
        data: {
          info: {
            // Empty provider/model identity drives reportedModel to null —
            // the structural no-model resolution failure path.
            ...assistantInfo(
              options.hostModelMissing === true ? "" : "host-provider",
              options.hostModelMissing === true ? "" : "host-model",
            ),
            variant: "default",
          },
          parts: [],
        },
        error: undefined,
      } as unknown as MessageResponse)
    },
    prompt: async (request: PromptOptions): Promise<PromptResponse> => {
      calls.prompts.push(request)
      const bodyModel = request.body?.model
      const info = assistantInfo(bodyModel?.providerID ?? "provider", bodyModel?.modelID ?? "model")
      return fieldsResponse({
        data: { info, parts: [textPart("delegated output")] },
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

/**
 * Recording GitExec answering the auto-checkpoint key derivation with a
 * canned branch (this change's branch, so checkpoint assertions read
 * naturally).
 */
function recordingExec(
  branch = "opsx/model-matrix-delegation",
): { exec: GitExec; calls: Array<{ args: string[]; cwd: string }> } {
  const calls: Array<{ args: string[]; cwd: string }> = []
  const exec: GitExec = (args, cwd) => {
    calls.push({ args, cwd })
    if (args.join(" ") === "rev-parse --abbrev-ref HEAD") return branch
    throw new Error(`unexpected git command: ${args.join(" ")}`)
  }
  return { exec, calls }
}

/**
 * Fake readText keyed on absolute paths. Absent files reject with
 * Node-shaped ENOENT so loadModelMatrix resolves them as absent — the
 * matrix/override probes and command-file reads all flow through here.
 */
function fakeReadText(files: Record<string, string>): {
  readText: DelegateCommandDependencies["readText"]
  calls: string[]
} {
  const calls: string[] = []
  const readText = async (path: string): Promise<string> => {
    calls.push(path)
    if (!(path in files)) {
      throw Object.assign(new Error(`ENOENT: no such file or directory, ${path}`), { code: "ENOENT" })
    }
    return files[path]
  }
  return { readText, calls }
}

/** Fake FileExists keyed on repo-relative paths resolved against FAKE_CWD. */
function fakeFileExists(existing: readonly string[]): { fileExists: FileExists; calls: string[] } {
  const existingPaths = new Set(existing.map((rel) => resolve(FAKE_CWD, rel)))
  const calls: string[] = []
  const fileExists: FileExists = (path: string) => {
    calls.push(path)
    return existingPaths.has(path)
  }
  return { fileExists, calls }
}

/**
 * Parser stub mirroring Bun.YAML for the shapes under test (the
 * model-matrix.test.ts convention): comments-only and empty documents
 * parse to null; everything else is JSON (a superset of the YAML the
 * fixtures use).
 */
function testParseYaml(text: string): unknown {
  const content = text
    .split(/\r?\n/)
    .filter((line) => line.trim() !== "" && !line.trim().startsWith("#"))
    .join("\n")
  if (content === "") {
    return null
  }
  return JSON.parse(text) as unknown
}

/** How the fake classifier spawn seam behaves for a test. */
type SpawnMode = "respond" | "reject" | "hang"

/**
 * Assembles the full delegate dependency set for one integration test:
 * filesystem seams serving the tracked matrix (default TRACKED_MATRIX),
 * the override (default MODEL_OVERRIDE), any command files, and the pack
 * files; plus the classifier spawn seam recording its calls. The LLM
 * fetch/env seams fail the test when reached — every test here either
 * bypasses the LLM engine (matrix fixtures carry no classifier.llm) or
 * expects it skipped, keeping the suite network-free.
 */
function integrationDeps(options: {
  /** Extra readText files (command files) beyond the matrix pair. */
  readonly files?: Record<string, string>
  /** Tracked matrix content; null serves no matrix (zero-config). */
  readonly matrix?: string | null
  /** Override content; null serves no override. */
  readonly override?: string | null
  /** Pack files reported existing (loader MM-FR-001 check). */
  readonly existing?: readonly string[]
  /** Skill files reported existing (skill resolution). */
  readonly existingSkills?: readonly string[]
  /** Fake client override (host-missing variants). */
  readonly clientOptions?: FakeClientOptions
  /** Canned classifier spawn response (spawnMode "respond"). */
  readonly classifierResponse?: string
  /** Classifier spawn behavior (default "respond"). */
  readonly spawnMode?: SpawnMode
}): {
  readonly deps: DelegateCommandDependencies
  readonly calls: FakeCalls
  readonly reads: string[]
  readonly spawnCalls: Array<{ agent: string; prompt: string }>
} {
  const fake = fakeClient(options.clientOptions)
  const files: Record<string, string> = { ...(options.files ?? {}) }
  // Explicit undefined checks (not nullish coalescing): matrix: null and
  // override: null express absence, which must not fall back to defaults.
  const matrix = options.matrix === undefined ? TRACKED_MATRIX : options.matrix
  const override = options.override === undefined ? MODEL_OVERRIDE : options.override
  if (matrix !== null) {
    files[MATRIX_PROBE] = matrix
  }
  if (override !== null) {
    files[OVERRIDE_PROBE] = override
  }
  const reads = fakeReadText(files)
  const artifact = fakeFileExists([
    ...(options.existing ?? PACK_FILES),
    ...(options.existingSkills ?? []),
  ])
  const spawnCalls: Array<{ agent: string; prompt: string }> = []
  const mode: SpawnMode = options.spawnMode ?? "respond"
  const deps: DelegateCommandDependencies = {
    client: fake.client,
    modelMatrix: {
      parseYaml: testParseYaml,
      classifierFetch: (() => {
        throw new Error("classifier fetch must not run — no test configures classifier.llm")
      }) as typeof fetch,
      classifierReadEnv: () => undefined,
      spawnClassifierAgent: async (_context, agent, prompt) => {
        spawnCalls.push({ agent, prompt })
        if (mode === "reject") {
          throw new Error("classifier agent engine unavailable")
        }
        if (mode === "hang") {
          await new Promise<never>(() => {}) // bounded by the engine's agent_timeout_ms race
        }
        return options.classifierResponse ?? '{"pack": null, "confidence": 0.9}'
      },
    },
    readText: reads.readText,
    fileExists: artifact.fileExists,
  }
  return { deps, calls: fake.calls, reads: reads.calls, spawnCalls }
}

/** Prompt text delivered to the child session (exactly one text part). */
function deliveredPrompt(calls: FakeCalls): string {
  const parts = calls.prompts[0]?.body?.parts as unknown as Array<{ type: string; text: string }>
  if (parts === undefined || parts.length !== 1 || parts[0].type !== "text") {
    throw new Error(`expected exactly one text part, got ${JSON.stringify(parts)}`)
  }
  return parts[0].text
}

/** Node-shaped ENOENT error so readState/writeState treat a path as absent. */
function enoent(what: string): NodeJS.ErrnoException {
  return Object.assign(new Error(`ENOENT: no such file or directory, ${what}`), { code: "ENOENT" })
}

/** Invoke the delegate tool with a full parent-session context and parse its JSON output. */
async function invokeDelegate(
  exec: GitExec,
  deps: DelegateCommandDependencies,
  args: Record<string, unknown>,
): Promise<ToolResult<DelegateCommandData>> {
  const delegate = createDelegateCommandTool(exec, FAKE_CWD, deps)
  const result = await (
    delegate as {
      execute: (
        args: Record<string, unknown>,
        ctx: { sessionID: string; messageID: string; directory: string; abort: AbortSignal },
      ) => Promise<{ output: string }>
    }
  ).execute(args, {
    sessionID: "parent-session",
    messageID: "message-parent",
    directory: FAKE_CWD,
    abort: new AbortController().signal,
  })
  return JSON.parse(result.output) as ToolResult<DelegateCommandData>
}

beforeEach(() => {
  vi.resetAllMocks()
  // Default state-store simulation: no prior state file or directory, so
  // readState returns null (first checkpoint) and writeState takes the
  // create-directory branch.
  vi.mocked(lstatSync).mockImplementation(() => {
    throw enoent("lstat state path")
  })
})

// ---------------------------------------------------------------------------
// AC1-AC8 — issue #887 acceptance coverage (task 5.5)
// ---------------------------------------------------------------------------

describe("uf_delegate_command model-matrix integration", () => {
  it("AC1: runs a delegated task on a non-host model configured via matrix + override", async () => {
    const harness = integrationDeps({})

    // plan is a static task kind (AC2): defaults.plan → spec-authoring,
    // whose model arrives via the gitignored override.
    const result = await invokeDelegate(recordingExec().exec, harness.deps, {
      prompt: "Draft the spec section.",
      task_classification: "judgment",
      checkpoint_step: "plan",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data.selection_source).toBe("default-kind")
      expect(result.data.selection_capability).toBe("spec-authoring")
      expect(result.data.selection_model).toBe("override-provider/spec-model")
      expect(result.data.selection_pack).toBeNull()
      expect(result.data.selection_confidence).toBeNull()
      expect(result.data.selection_fallback_reason).toBeNull()
    }

    // The child session ran on the override model, not the host model:
    // the create title and prompt body carry the matrix-resolved identity
    // (the mocked child session reports it back).
    expect(harness.calls.creates[0]?.body?.title).toBe("cobalt-crush-dev @ override-provider/spec-model")
    expect(harness.calls.prompts[0]?.body?.model).toEqual({
      providerID: "override-provider",
      modelID: "spec-model",
    })

    // Host introspection never ran — the matrix model replaced it entirely
    // (design D2: skip resolveHostModel on a resolved matrix model).
    expect(harness.calls.messages).toHaveLength(0)

    // Both config files were read through the shared readText seam.
    expect(harness.reads).toEqual([MATRIX_PROBE, OVERRIDE_PROBE])

    // The classifier stayed dark: plan resolves statically.
    expect(harness.spawnCalls).toHaveLength(0)
  })

  it("AC2: tasks resolves statically to spec-authoring like plan (classifier never invoked)", async () => {
    const harness = integrationDeps({ classifierResponse: '{"pack": "go", "confidence": 0.9}' })

    const result = await invokeDelegate(recordingExec().exec, harness.deps, {
      prompt: "Generate the task list.",
      task_classification: "judgment",
      checkpoint_step: "tasks",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data.selection_source).toBe("default-kind")
      expect(result.data.selection_capability).toBe("spec-authoring")
      expect(result.data.selection_model).toBe("override-provider/spec-model")
    }
    // The canned classifier response would have selected "go" — the
    // static resolution must win without ever consulting the cascade.
    expect(harness.spawnCalls).toHaveLength(0)
  })

  it("AC2: implement reaches the classifier cascade and a classified pack resolves through its profile", async () => {
    const harness = integrationDeps({ classifierResponse: '{"pack": "go", "confidence": 0.9}' })

    const result = await invokeDelegate(recordingExec().exec, harness.deps, {
      prompt: "Implement the parser.",
      task_classification: "judgment",
      checkpoint_step: "implement",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data.selection_source).toBe("classifier-agent")
      expect(result.data.selection_pack).toBe("go")
      expect(result.data.selection_capability).toBe("code-writing")
      expect(result.data.selection_model).toBe("override-provider/code-model")
      expect(result.data.selection_confidence).toBe(0.9)
    }

    // The hidden classifier subagent ran once with the default agent name
    // and a prompt naming the matrix's candidate packs.
    expect(harness.spawnCalls).toHaveLength(1)
    expect(harness.spawnCalls[0].agent).toBe("model-matrix-classifier")
    expect(harness.spawnCalls[0].prompt).toContain("Candidate packs: content, go, meta-prompt")
    expect(harness.spawnCalls[0].prompt).toContain("Implement the parser.")

    // The classifier-selected pack renders the convention-pack directive
    // (MM-FR-005 double duty) before the extra prompt.
    expect(deliveredPrompt(harness.calls)).toBe(
      "Required convention pack: go\n\n" +
        "Before writing, load the convention pack at .opencode/uf/packs/go.md and apply " +
        "its rules to this task. The pack file is the source of truth for its conventions — " +
        "do not reconstruct its content from memory or improvise equivalent rules.\n\n" +
        "Implement the parser.",
    )
  })

  it("AC2: an unknown task kind with a discarded classifier result falls back to host with the reason recorded", async () => {
    // No checkpoint_step: the kind is unknown, so the cascade runs and its
    // below-threshold result is discarded; unknown kinds have no static
    // default, so selection falls back to host with the reason audited.
    const harness = integrationDeps({ classifierResponse: '{"pack": "go", "confidence": 0.2}' })

    const result = await invokeDelegate(recordingExec().exec, harness.deps, {
      prompt: "Do the thing.",
      task_classification: "judgment",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data.selection_source).toBe("host")
      expect(result.data.selection_model).toBeNull()
      expect(result.data.selection_capability).toBeNull()
      expect(result.data.selection_fallback_reason).toContain("below the threshold 0.6")
      expect(result.data.selection_fallback_reason).toContain('no static default for task kind "unknown"')
    }
    // Host fallback introspects the parent session (the pre-change path).
    expect(harness.calls.messages).toHaveLength(1)
    expect(harness.calls.prompts[0]?.body?.model).toEqual({
      providerID: "host-provider",
      modelID: "host-model",
    })
  })

  it("AC3: an explicit model wins the chain, with the validated pack retained for the directive", async () => {
    const harness = integrationDeps({})

    const result = await invokeDelegate(recordingExec().exec, harness.deps, {
      prompt: "Implement the parser.",
      task_classification: "judgment",
      model: "explicit/provider-model",
      variant: "fast",
      profile: "code-writing",
      pack: "go",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      // Step 1 wins outright; the validated companions ride along rather
      // than shadowing (unconditional validation) or being dropped.
      expect(result.data.selection_source).toBe("explicit-model")
      expect(result.data.selection_model).toBe("explicit/provider-model")
      expect(result.data.selection_pack).toBe("go")
      expect(result.data.selection_capability).toBeNull()
    }
    // The child runs on the explicit identity with its variant, host
    // introspection is skipped, and the classifier is bypassed.
    expect(harness.calls.prompts[0]?.body?.model).toEqual({
      providerID: "explicit",
      modelID: "provider-model",
    })
    expect(harness.calls.prompts[0]?.body?.variant).toBe("fast")
    expect(harness.calls.messages).toHaveLength(0)
    expect(harness.spawnCalls).toHaveLength(0)
    // The retained pack still renders its directive.
    expect(deliveredPrompt(harness.calls)).toContain("Required convention pack: go")
  })

  it("AC3: an explicit profile beats the pack, and a model-less profile falls back to host with the pack retained", async () => {
    // content-writing carries no model even after the override merge, so
    // the profile resolution falls to host (step 7) — but the capability
    // and the supplied pack stay recorded, and the pack directive renders.
    const harness = integrationDeps({})

    const result = await invokeDelegate(recordingExec().exec, harness.deps, {
      prompt: "Write the announcement.",
      task_classification: "judgment",
      profile: "content-writing",
      pack: "go",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data.selection_source).toBe("host")
      expect(result.data.selection_capability).toBe("content-writing")
      expect(result.data.selection_pack).toBe("go")
      expect(result.data.selection_model).toBeNull()
      expect(result.data.selection_fallback_reason).toBe('profile "content-writing" carries no model')
    }
    // Host fallback introspected the parent and delivered the host model.
    expect(harness.calls.messages).toHaveLength(1)
    expect(harness.calls.prompts[0]?.body?.model).toEqual({
      providerID: "host-provider",
      modelID: "host-model",
    })
    // The supplied pack rides along for directive rendering.
    expect(deliveredPrompt(harness.calls)).toContain("Required convention pack: go")
  })

  it("AC3: an explicit pack resolves through its mapped profile's model", async () => {
    const harness = integrationDeps({})

    const result = await invokeDelegate(recordingExec().exec, harness.deps, {
      prompt: "Implement the parser.",
      task_classification: "judgment",
      pack: "go",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data.selection_source).toBe("explicit-pack")
      expect(result.data.selection_pack).toBe("go")
      expect(result.data.selection_capability).toBe("code-writing")
      expect(result.data.selection_model).toBe("override-provider/code-model")
    }
    expect(harness.calls.messages).toHaveLength(0)
    expect(harness.spawnCalls).toHaveLength(0)
    expect(harness.calls.prompts[0]?.body?.model).toEqual({
      providerID: "override-provider",
      modelID: "code-model",
    })
  })

  it("AC4: a missing matrix keeps the zero-config behavior — host model, no classifier, no errors", async () => {
    const harness = integrationDeps({ matrix: null, override: null })

    const result = await invokeDelegate(recordingExec().exec, harness.deps, {
      prompt: "Implement the parser.",
      task_classification: "judgment",
      checkpoint_step: "implement",
    })

    // Behavior identical to today (D8): the implement kind would reach the
    // classifier with a matrix, but zero-config resolves to host.
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data.selection_source).toBe("host")
      expect(result.data.selection_model).toBeNull()
      expect(result.data.selection_fallback_reason).toBeNull()
    }
    expect(harness.calls.messages).toHaveLength(1)
    expect(harness.calls.prompts[0]?.body?.model).toEqual({
      providerID: "host-provider",
      modelID: "host-model",
    })
    expect(harness.spawnCalls).toHaveLength(0)
    // Zero-config short-circuit: the matrix probe is the only read — the
    // override is never consulted without a tracked matrix.
    expect(harness.reads).toEqual([MATRIX_PROBE])
  })

  it("AC5: a non-whitelisted override key fails closed naming the whitelist, pre-spawn", async () => {
    const harness = integrationDeps({
      // classifier.enabled is structural — only classifier.llm.* is
      // per-developer (design D4), so the override fails the whitelist.
      override: JSON.stringify({ classifier: { enabled: false } }),
    })

    const result = await invokeDelegate(recordingExec().exec, harness.deps, {
      prompt: "Implement the parser.",
      task_classification: "judgment",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("MATRIX_INVALID")
      expect(result.retryable).toBe(false)
      expect(result.message).toContain('override key "classifier.enabled" is not permitted')
      expect(result.message).toContain("profiles.*.model, profiles.*.variant, classifier.llm.*")
    }
    // Fail-closed pre-spawn: zero child sessions, zero classifier spawns,
    // and both config files were read before the rejection.
    expect(harness.calls.creates).toHaveLength(0)
    expect(harness.calls.prompts).toHaveLength(0)
    expect(harness.spawnCalls).toHaveLength(0)
    expect(harness.reads).toEqual([MATRIX_PROBE, OVERRIDE_PROBE])
  })

  it("AC6: the pack directive holds its golden shape and fixed position with a coexisting skill directive", async () => {
    const commandAbs = resolve(FAKE_CWD, ".opencode/commands/speckit.implement.md")
    const harness = integrationDeps({
      files: { [commandAbs]: "Implement the change." },
      existingSkills: [".opencode/skills/pre-flight/SKILL.md"],
    })

    const result = await invokeDelegate(recordingExec().exec, harness.deps, {
      command_path: ".opencode/commands/speckit.implement.md",
      feature_dir: "openspec/changes/model-matrix-delegation/",
      prompt: "Focus on the parser.",
      task_classification: "deterministic",
      skill: "pre-flight",
      pack: "go",
    })

    // Golden shape (design D6): command text → feature-dir context →
    // skill directive → pack directive → extra prompt. The pack directive
    // is a fixed, tool-rendered section — a header line naming the pack,
    // a blank line, then the fixed paragraph — never inlined content,
    // never agent-reconstructed.
    expect(result.status).toBe("ok")
    expect(deliveredPrompt(harness.calls)).toBe(
      "Implement the change.\n\n" +
        "Feature directory: openspec/changes/model-matrix-delegation/\n\n" +
        "Required skill: .opencode/skills/pre-flight/SKILL.md\n\n" +
        "This task is classified as deterministic work. Load the skill (invoke the `skill` tool " +
        'with name "pre-flight") or run the script at .opencode/skills/pre-flight/SKILL.md, and ' +
        "execute its full procedure. Do not substitute an improvised procedure.\n\n" +
        "Required convention pack: go\n\n" +
        "Before writing, load the convention pack at .opencode/uf/packs/go.md and apply " +
        "its rules to this task. The pack file is the source of truth for its conventions — " +
        "do not reconstruct its content from memory or improvise equivalent rules.\n\n" +
        "Focus on the parser.",
    )
  })

  it("AC7: a failed classifier engine never blocks — static fallback resolves and the audit records the reason", async () => {
    const harness = integrationDeps({ spawnMode: "reject" })

    const result = await invokeDelegate(recordingExec().exec, harness.deps, {
      prompt: "Implement the parser.",
      task_classification: "judgment",
      checkpoint_step: "implement",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      // Static fallback for implement: defaults.implement → code-writing.
      expect(result.data.selection_source).toBe("default-kind")
      expect(result.data.selection_capability).toBe("code-writing")
      expect(result.data.selection_model).toBe("override-provider/code-model")
      // Both engines' outcomes are audited: the agent engine's failure and
      // the LLM engine's skip (unconfigured — no classifier.llm override).
      expect(result.data.selection_fallback_reason).toContain("agent engine: classifier agent engine unavailable")
      expect(result.data.selection_fallback_reason).toContain("llm engine")
    }
    // The delegation completed: one child session on the fallback model.
    expect(harness.calls.creates).toHaveLength(1)
    expect(harness.calls.prompts[0]?.body?.model).toEqual({
      providerID: "override-provider",
      modelID: "code-model",
    })
  })

  it("AC7: a timed-out classifier engine falls back statically within the configured bound", async () => {
    // agent_timeout_ms is structural matrix configuration (tracked-file
    // legal): a 1 ms bound plus a hanging spawn exercises the engine's
    // abort race without any fake timers.
    const timeoutMatrix = JSON.stringify({
      version: 1,
      profiles: { "code-writing": {} },
      packs: { go: "code-writing" },
      defaults: { implement: "code-writing" },
      agents: {},
      classifier: { agent_timeout_ms: 1 },
    })
    const harness = integrationDeps({
      matrix: timeoutMatrix,
      // The override may only name base-declared capabilities (design
      // D4), so this reduced matrix's override carries code-writing only.
      override: JSON.stringify({
        profiles: { "code-writing": { model: "override-provider/code-model" } },
      }),
      spawnMode: "hang",
    })

    const result = await invokeDelegate(recordingExec().exec, harness.deps, {
      prompt: "Implement the parser.",
      task_classification: "judgment",
      checkpoint_step: "implement",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data.selection_source).toBe("default-kind")
      expect(result.data.selection_model).toBe("override-provider/code-model")
      expect(result.data.selection_fallback_reason).toContain("exceeded agent_timeout_ms 1")
    }
  })

  it("AC7: a below-threshold classifier result falls back statically with the threshold named", async () => {
    const harness = integrationDeps({ classifierResponse: '{"pack": "go", "confidence": 0.2}' })

    const result = await invokeDelegate(recordingExec().exec, harness.deps, {
      prompt: "Implement the parser.",
      task_classification: "judgment",
      checkpoint_step: "implement",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data.selection_source).toBe("default-kind")
      expect(result.data.selection_model).toBe("override-provider/code-model")
      expect(result.data.selection_fallback_reason).toContain("below the threshold 0.6")
    }
  })

  it("AC7: a classifier result naming a matrix-invalid pack is discarded and falls back statically", async () => {
    const harness = integrationDeps({ classifierResponse: '{"pack": "nonexistent", "confidence": 0.95}' })

    const result = await invokeDelegate(recordingExec().exec, harness.deps, {
      prompt: "Implement the parser.",
      task_classification: "judgment",
      checkpoint_step: "implement",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data.selection_source).toBe("default-kind")
      expect(result.data.selection_fallback_reason).toContain(
        'pack "nonexistent" is not a packs key in the configured matrix',
      )
    }
  })

  it("a matrix-resolved model proceeds when host introspection would fail (no HOST_MODEL_UNAVAILABLE)", async () => {
    // The parent session reports no model identity — the structural
    // failure that returns HOST_MODEL_UNAVAILABLE on the host path — but
    // the matrix resolves the model first, so introspection is never
    // attempted (design D2: skip resolveHostModel on a resolved model).
    const harness = integrationDeps({ clientOptions: { hostModelMissing: true } })

    const result = await invokeDelegate(recordingExec().exec, harness.deps, {
      prompt: "Draft the spec section.",
      task_classification: "judgment",
      checkpoint_step: "plan",
    })

    expect(result.status).toBe("ok")
    expect(harness.calls.messages).toHaveLength(0)
    expect(harness.calls.creates).toHaveLength(1)
    expect(harness.calls.prompts[0]?.body?.model).toEqual({
      providerID: "override-provider",
      modelID: "spec-model",
    })
  })

  it("the host path preserves the pre-change structural-failure semantics on a matrix-less repository", async () => {
    const harness = integrationDeps({
      matrix: null,
      override: null,
      clientOptions: { hostModelMissing: true },
    })

    const result = await invokeDelegate(recordingExec().exec, harness.deps, {
      prompt: "Draft the spec section.",
      task_classification: "judgment",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("HOST_MODEL_UNAVAILABLE")
      expect(result.retryable).toBe(false)
    }
    expect(harness.calls.creates).toHaveLength(0)
  })

  it("UF-FR-002 boundary rows: profile and pack at exactly 1 and 128 chars are accepted", async () => {
    // The matrix's own capability namespace is closed at 32 chars
    // (CapabilityNameSchema) and pack names at 64 (PackNameSchema), so the
    // 128-char parameter bound is observable as input-shape acceptance: a
    // 128-char value passes the tool schema and fails later at resolution
    // (UNKNOWN_PROFILE / INVALID_PACK), never INVALID_INPUT. The 1-char
    // values resolve end-to-end against legal single-character names.
    const boundaryMatrix = JSON.stringify({
      version: 1,
      profiles: { a: { model: "boundary-provider/one" } },
      packs: { z: "a" },
      classifier: {},
    })
    const existing = [".opencode/uf/packs/z.md"]

    // 1-char profile: accepted at input shape and resolved end-to-end.
    const one = integrationDeps({ matrix: boundaryMatrix, override: null, existing })
    const oneResult = await invokeDelegate(recordingExec().exec, one.deps, {
      prompt: "Work.",
      task_classification: "judgment",
      profile: "a",
    })
    expect(oneResult.status).toBe("ok")
    if (oneResult.status === "ok") {
      expect(oneResult.data.selection_source).toBe("explicit-profile")
      expect(oneResult.data.selection_model).toBe("boundary-provider/one")
    }

    // 1-char pack: the same inclusive lower bound.
    const packOne = integrationDeps({ matrix: boundaryMatrix, override: null, existing })
    const packOneResult = await invokeDelegate(recordingExec().exec, packOne.deps, {
      prompt: "Work.",
      task_classification: "judgment",
      pack: "z",
    })
    expect(packOneResult.status).toBe("ok")
    if (packOneResult.status === "ok") {
      expect(packOneResult.data.selection_source).toBe("explicit-pack")
      expect(packOneResult.data.selection_model).toBe("boundary-provider/one")
    }

    // 128-char profile: the schema's upper bound is inclusive — the value
    // passes input validation (no INVALID_INPUT) and fails at resolution
    // naming the value and the configured capabilities.
    const max = integrationDeps({ matrix: boundaryMatrix, override: null, existing })
    const maxResult = await invokeDelegate(recordingExec().exec, max.deps, {
      prompt: "Work.",
      task_classification: "judgment",
      profile: "p".repeat(128),
    })
    expect(maxResult.status).toBe("error")
    if (maxResult.status === "error") {
      expect(maxResult.code).toBe("UNKNOWN_PROFILE")
      expect(maxResult.retryable).toBe(false)
      expect(maxResult.message).toContain("p".repeat(128))
    }

    // 128-char pack: the same inclusive upper bound, failing at the
    // packs-key membership check rather than input validation.
    const packMax = integrationDeps({ matrix: boundaryMatrix, override: null, existing })
    const packMaxResult = await invokeDelegate(recordingExec().exec, packMax.deps, {
      prompt: "Work.",
      task_classification: "judgment",
      pack: "k".repeat(128),
    })
    expect(packMaxResult.status).toBe("error")
    if (packMaxResult.status === "error") {
      expect(packMaxResult.code).toBe("INVALID_PACK")
      expect(packMaxResult.retryable).toBe(false)
      expect(packMaxResult.message).toContain("k".repeat(128))
    }

    // Every boundary call cost zero child sessions: all four probes
    // rejected or resolved before any spawn beyond the two successes,
    // and the two successes each created exactly one session.
    expect(one.calls.creates).toHaveLength(1)
    expect(packOne.calls.creates).toHaveLength(1)
    expect(max.calls.creates).toHaveLength(0)
    expect(packMax.calls.creates).toHaveLength(0)
  })

  it("UF-FR-002 boundary rows: empty-string and 129-char profile/pack values fail input validation pre-spawn", async () => {
    const harness = integrationDeps({})
    const over = "x".repeat(129)

    for (const args of [
      { profile: "" },
      { profile: over },
      { pack: "" },
      { pack: over },
      // The variant-requires-model refinement (task 5.1): a variant with
      // no model has no defined meaning in the chain.
      { variant: "v" },
    ]) {
      const result = await invokeDelegate(recordingExec().exec, harness.deps, {
        prompt: "Work.",
        task_classification: "judgment",
        ...args,
      })
      expect(result.status).toBe("error")
      if (result.status === "error") {
        expect(result.code).toBe("INVALID_INPUT")
        expect(result.retryable).toBe(false)
      }
    }
    // Every rejection cost zero sessions: the loop above made five calls
    // and none reached the child spawn.
    expect(harness.calls.creates).toHaveLength(0)
    expect(harness.calls.prompts).toHaveLength(0)
  })
})

// ---------------------------------------------------------------------------
// DE-007 seam verification (task 5.6)
// ---------------------------------------------------------------------------

describe("uf_delegate_command model-matrix seam verification", () => {
  it("(a) the registration surface advertises the selection parameters and accepts them", async () => {
    const harness = integrationDeps({})
    const delegate = createDelegateCommandTool(recordingExec().exec, FAKE_CWD, harness.deps)

    // The tool's registered args carry the full fifteen-parameter surface —
    // the ten pre-change parameters plus the four selection parameters and
    // variant — so the host-side schema and the executor's internal
    // re-validation cannot drift apart.
    const registered = delegate as unknown as {
      args: Record<string, { safeParse: (value: unknown) => { success: boolean } }>
    }
    expect(Object.keys(registered.args).sort()).toEqual(
      [
        "agent",
        "command_path",
        "prompt",
        "feature_dir",
        "expected_artifact",
        "checkpoint_step",
        "phase",
        "timeout",
        "task_classification",
        "skill",
        "skill_absent_reason",
        "model",
        "variant",
        "profile",
        "pack",
      ].sort(),
    )

    // Each registered selection-parameter schema accepts a well-formed
    // value (the boundary rows above prove the rejections).
    expect(registered.args.model.safeParse("provider/model").success).toBe(true)
    expect(registered.args.variant.safeParse("fast").success).toBe(true)
    expect(registered.args.profile.safeParse("code-writing").success).toBe(true)
    expect(registered.args.pack.safeParse("go").success).toBe(true)
  })

  it("(b) the MM-FR-007 audit fields land in the checkpoint data exactly as documented", async () => {
    const harness = integrationDeps({ classifierResponse: '{"pack": "go", "confidence": 0.85}' })

    const result = await invokeDelegate(recordingExec().exec, harness.deps, {
      prompt: "Implement the parser.",
      task_classification: "judgment",
      checkpoint_step: "implement",
      phase: 5,
    })

    expect(result.status).toBe("ok")
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    const state = JSON.parse(stateRaw as string) as ProgressState

    // MM-FR-007 "Audit fields accompany every selection" and "Classifier
    // confidence is recorded": the checkpoint data carries the documented
    // field names with the classifier's outcome, and omits the keys that
    // carry no value (the valued-shape convention the audit-trail tests
    // in unleash-delegate.test.ts pin for the zero-config rows).
    expect(state.current_step).toBe("implement")
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      phase: 5,
      task_classification: "judgment",
      selection_source: "classifier-agent",
      selection_pack: "go",
      selection_capability: "code-writing",
      selection_model: "override-provider/code-model",
      selection_confidence: 0.85,
    })
    expect(state.data).not.toHaveProperty("selection_fallback_reason")
  })

  it("(c) a samples-shaped matrix loads through loadModelMatrix and round-trips its capability into resolveSelection", async () => {
    // The registry's own starter fixture is the write side of the
    // round-trip: content shaped exactly like
    // schemas/model-matrix/samples/valid-starter-matrix.json must load
    // through the same loader the tool uses and resolve a selection whose
    // capability matches the fixture's packs mapping — the three layer
    // surfaces (schema registry, lib loader, selection resolver) stay in
    // sync (DE-007).
    const actual = await vi.importActual<typeof import("node:fs")>("node:fs")
    const samplePath = fileURLToPath(
      new URL("../../schemas/model-matrix/samples/valid-starter-matrix.json", import.meta.url),
    )
    const sample = actual.readFileSync(samplePath, "utf8")

    const starterPacks = [
      ".opencode/uf/packs/meta-prompt.md",
      ".opencode/uf/packs/go.md",
      ".opencode/uf/packs/typescript.md",
      ".opencode/uf/packs/python.md",
      ".opencode/uf/packs/content.md",
    ]
    const dependencies: ModelMatrixDependencies = {
      readText: async (path) => {
        if (path === ".uf/model-matrix.yaml") {
          return sample
        }
        throw enoent(path)
      },
      parseYaml: testParseYaml,
      fileExists: (path) => starterPacks.includes(path),
    }

    const load = await loadModelMatrix(dependencies)
    expect(load.status).toBe("ok")

    const resolution = resolveSelection(
      load.status === "ok" ? load.matrix : null,
      { pack: "go" },
      { status: "not-run" },
    )
    expect(resolution.ok).toBe(true)
    if (resolution.ok) {
      // The starter matrix's profiles carry no models (identities arrive
      // per-developer, D9), so an explicit pack resolves its capability
      // and falls to host with the capability and pack still recorded —
      // the round-tripped value is the packs mapping's "go" →
      // "code-writing".
      expect(resolution.selection.selection_source).toBe("host")
      expect(resolution.selection.pack).toBe("go")
      expect(resolution.selection.capability).toBe("code-writing")
      expect(resolution.selection.model).toBeNull()
      expect(resolution.selection.fallback_reason).toBe('profile "code-writing" carries no model')
    }
  })
})
