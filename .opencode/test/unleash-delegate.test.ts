import { existsSync, lstatSync, mkdirSync, readFileSync, writeFileSync } from "node:fs"
import { resolve } from "node:path"

import type { PluginInput } from "@opencode-ai/plugin"
import { beforeEach, describe, expect, it, vi } from "vitest"

import type { ToolResult } from "../lib/uf-workflow-types.js"
import {
  createDelegateCommandTool,
  createGetProgressTool,
  type DelegateCommandData,
  type DelegateCommandDependencies,
  type FileExists,
  type GitExec,
  type ProgressState,
} from "../plugins/uf-workflow/index.js"

// ---------------------------------------------------------------------------
// Module-level mocks (hoisted by vitest)
// ---------------------------------------------------------------------------

// The tool's auto-checkpoint (D5) persists through readState/writeState,
// which hit node:fs directly. Mocking the four functions that path uses (the
// unleash-branch-gate.test.ts recipe, per Dewey learning
// uf-workflow-testing-20261015T201843) keeps every delegate test off the real
// filesystem: state I/O is observed through the recorded calls instead of
// performed. existsSync is mocked so tests can prove artifact verification
// probes existence only through the injected FileExists dependency, never
// through the default existsSync-backed checker. The command-file reader is
// injected separately (readText), so node:fs/promises — a different specifier
// used only by the plugin registration — stays untouched.
vi.mock("node:fs", async (importOriginal) => {
  const actual = await importOriginal<typeof import("node:fs")>()
  return {
    ...actual,
    existsSync: vi.fn(),
    lstatSync: vi.fn(),
    readFileSync: vi.fn(),
    mkdirSync: vi.fn(),
    writeFileSync: vi.fn(),
  }
})

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

/** Fake repository root — never a real path on any filesystem. */
const FAKE_CWD = "/fake/project"

/**
 * The zero-config model-matrix probe: every delegation reads
 * .uf/model-matrix.yaml through the shared readText seam before host
 * introspection (model-matrix-delegation D8), so successful runs record
 * it in reads.calls even when no matrix exists.
 */
const MATRIX_PROBE = resolve(FAKE_CWD, ".uf/model-matrix.yaml")

/** Host model identity the fake parent session reports. */
const HOST_MODEL = "host-provider/host-model"

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
  /** Variant reported by the parent assistant message (default "default"; explicit undefined = absent). */
  readonly hostVariant?: string
  /** Parent assistant message reports no model identity (structural no-model resolution failure). */
  readonly hostModelMissing?: boolean
  /** Error payload returned by session.message (transient host-resolution failure). */
  readonly messageError?: unknown
  /** Error payload returned by session.create (child_create_failed path). */
  readonly createError?: unknown
  /** Parts reported by the child prompt response (default one text part). */
  readonly parts?: readonly unknown[]
  /** Hang session.prompt on its abort signal (timeout/cancellation paths). */
  readonly hangPrompt?: boolean
  /** Hang session.message on its abort signal (host-resolution cancellation). */
  readonly hangMessage?: boolean
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

function finishPart(cost: number, input: number, output: number): Record<string, unknown> {
  return {
    id: `finish-${cost}-${input}-${output}`,
    sessionID: "child-session",
    messageID: "message-child",
    type: "step-finish",
    reason: "stop",
    cost,
    tokens: { input, output, reasoning: 0, cache: { read: 0, write: 0 } },
  }
}

/**
 * Fake OpenCode client (invoke-agent.test.ts precedent) driving the REAL
 * executeAgentSession executor. The child session lifecycle — host
 * introspection, create, prompt, abort cleanup — runs for real against
 * recorded fakes, so timeout and cancellation propagation is exercised
 * through the same code path production uses, not a stubbed result.
 */
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
      if (options.hangMessage === true) {
        return await new Promise<MessageResponse>((_resolve, reject) => {
          request.signal?.addEventListener("abort", () => reject(request.signal?.reason), { once: true })
        })
      }
      return fieldsResponse({
        data: {
          info: {
            // Empty provider/model identity drives reportedModel to null —
            // the structural no-model resolution failure path.
            ...assistantInfo(
              options.hostModelMissing === true ? "" : "host-provider",
              options.hostModelMissing === true ? "" : "host-model",
            ),
            variant: "hostVariant" in options ? options.hostVariant : "default",
          },
          parts: [],
        },
        error: undefined,
      } as unknown as MessageResponse)
    },
    prompt: async (request: PromptOptions): Promise<PromptResponse> => {
      calls.prompts.push(request)
      if (options.hangPrompt === true) {
        return await new Promise<PromptResponse>((_resolve, reject) => {
          request.signal?.addEventListener("abort", () => reject(request.signal?.reason), { once: true })
        })
      }
      // The child reports the model it was asked to run, so successful
      // delegations never trip the executor's model-mismatch guard.
      const bodyModel = request.body?.model
      const info = assistantInfo(bodyModel?.providerID ?? "provider", bodyModel?.modelID ?? "model")
      return fieldsResponse({
        data: { info, parts: options.parts ?? [textPart("delegated output")] },
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
 * canned branch. The delegate tool runs git only when checkpoint_step is
 * supplied on a successful run, so tests assert on the recorded calls to
 * prove exactly that.
 */
function recordingExec(
  branch = "opsx/unleash-deterministic-tools",
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
 * Fake readText keyed on absolute paths (the tool resolves command_path
 * against the cwd before reading). Records every read so tests can assert
 * both repo-relative resolution and that the reviewer manifest is never
 * consulted through this dependency.
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
 * model-matrix.test.ts / review-dispatch convention): comments-only and
 * empty documents parse to null; everything else is JSON (a superset of
 * the YAML the fixtures use).
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

/**
 * Matrix-free model-matrix selection dependencies shared by the
 * zero-config tests: the YAML stub is never reached (fakeReadText reports
 * .uf/model-matrix.yaml absent, so loadModelMatrix resolves before
 * parsing), and every classifier seam fails the test when reached — the
 * bypass predicate must keep the cascade dark when no matrix is
 * configured.
 */
const matrixFreeDeps: DelegateCommandDependencies["modelMatrix"] = {
  parseYaml: testParseYaml,
  classifierFetch: (() => {
    throw new Error("classifier fetch must not run without a configured matrix")
  }) as typeof fetch,
  classifierReadEnv: () => undefined,
  spawnClassifierAgent: async () => {
    throw new Error("classifier spawn must not run without a configured matrix")
  },
}

/**
 * Recording model-matrix selection dependencies for matrix-driven tests:
 * the spawn seam records its prompt and returns a canned response
 * (default: a classifier JSON response naming no pack, exercising the
 * fall-through), and the fetch/env seams fail the test when reached (the
 * LLM engine only runs when the agent engine fails first).
 */
function fakeModelMatrixDeps(options: {
  readonly classifierResponse?: string
}): {
  modelMatrix: DelegateCommandDependencies["modelMatrix"]
  spawnCalls: Array<{ agent: string; prompt: string }>
} {
  const spawnCalls: Array<{ agent: string; prompt: string }> = []
  const modelMatrix: DelegateCommandDependencies["modelMatrix"] = {
    parseYaml: testParseYaml,
    classifierFetch: (() => {
      throw new Error("classifier fetch must not run when the agent engine answers")
    }) as typeof fetch,
    classifierReadEnv: () => undefined,
    spawnClassifierAgent: async (_context, agent, prompt) => {
      spawnCalls.push({ agent, prompt })
      return options.classifierResponse ?? '{"pack": null, "confidence": 0.9}'
    },
  }
  return { modelMatrix, spawnCalls }
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
  abort: AbortSignal = new AbortController().signal,
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
    abort,
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
// uf_delegate_command — command-file read and prompt composition
// ---------------------------------------------------------------------------

describe("uf_delegate_command command-file read and prompt composition", () => {
  it("reads the command file repo-relative and composes command text, feature-dir context, then extra prompt", async () => {
    const commandAbs = resolve(FAKE_CWD, ".opencode/commands/speckit.plan.md")
    const fake = fakeClient()
    const reads = fakeReadText({ [commandAbs]: "Create the implementation plan." })
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      command_path: ".opencode/commands/speckit.plan.md",
      feature_dir: "specs/018-unleash-command/",
      prompt: "Focus on the parser.",
    })

    expect(result.status).toBe("ok")

    // The command file is resolved against the injected cwd before reading;
    // the zero-config matrix probe follows (task 5.2 selection wiring).
    expect(reads.calls).toEqual([commandAbs, MATRIX_PROBE])

    // Host model resolution introspects the parent session's message.
    expect(fake.calls.messages).toHaveLength(1)
    expect(fake.calls.messages[0]).toMatchObject({
      path: { id: "parent-session", messageID: "message-parent" },
      query: { directory: FAKE_CWD },
    })

    // The child session is parented, titled with agent + resolved model, and
    // prompted with the default agent (cobalt-crush-dev).
    expect(fake.calls.creates[0]?.body).toMatchObject({
      parentID: "parent-session",
      title: `cobalt-crush-dev @ ${HOST_MODEL} (default)`,
    })
    expect(fake.calls.prompts[0]?.body).toMatchObject({
      agent: "cobalt-crush-dev",
      model: { providerID: "host-provider", modelID: "host-model" },
      variant: "default",
    })

    // Prompt composition order (D2): command text, then feature-dir
    // context, then extra instructions.
    expect(deliveredPrompt(fake.calls)).toBe(
      "Create the implementation plan.\n\nFeature directory: specs/018-unleash-command/\n\nFocus on the parser.",
    )
  })

  it("delegates prompt-only calls without command file or feature-dir sections", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Implement task 2.5 exactly.",
      task_classification: "judgment",
    })

    expect(result.status).toBe("ok")
    // Prompt-only delegation (per-task implement and code-fix calls) reads
    // no command file and prepends no feature-dir section; the only read
    // is the zero-config matrix probe.
    expect(reads.calls).toEqual([MATRIX_PROBE])
    expect(deliveredPrompt(fake.calls)).toBe("Implement task 2.5 exactly.")
  })

  it("accepts a command file at exactly the 1 MiB bound", async () => {
    const atBound = "x".repeat(1024 * 1024)
    const fake = fakeClient()
    const reads = fakeReadText({ [resolve(FAKE_CWD, "cmd.md")]: atBound })
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, { command_path: "cmd.md" })

    // The bound is inclusive: exactly 1 MiB passes through to the prompt.
    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)
    expect(deliveredPrompt(fake.calls)).toBe(atBound)
  })

  it("fails with COMMAND_FILE_READ_FAILED when the command file cannot be read", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({}) // nothing mapped — every read throws ENOENT
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      command_path: ".opencode/commands/missing.md",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("COMMAND_FILE_READ_FAILED")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        `Cannot read command file: ENOENT: no such file or directory, ${resolve(FAKE_CWD, ".opencode/commands/missing.md")}`,
      )
    }
    // Failed before any session activity.
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — input validation
// ---------------------------------------------------------------------------

describe("uf_delegate_command input validation", () => {
  it("rejects command files over 1 MiB with COMMAND_FILE_TOO_LARGE", async () => {
    const over = "x".repeat(1024 * 1024 + 1)
    const fake = fakeClient()
    const reads = fakeReadText({ [resolve(FAKE_CWD, "big.md")]: over })
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, { command_path: "big.md" })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("COMMAND_FILE_TOO_LARGE")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe("Command file exceeds 1 MiB limit (1048577 bytes)")
    }
    // Rejected before any session activity.
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
  })

  it("rejects divisor-* reviewer agents with INVALID_AGENT directing to invoke_agent", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      agent: "divisor-guard",
      prompt: "Review this change.",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_AGENT")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "reviewer agents (divisor-*) are dispatched via invoke_agent; uf_delegate_command is for content delegation",
      )
    }
    // The delegation/review boundary is structural (D2): no manifest read
    // and no session activity precede the rejection.
    expect(reads.calls).toEqual([])
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("rejects inputs with neither command_path nor prompt", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {})

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      // The empty call violates both the instruction rule and the
      // prompt-only classification rule; both issues are reported.
      expect(result.message).toBe(
        "prompt: one of command_path or prompt is required; " +
          "task_classification: prompt-only delegation (no command_path) must declare task_classification -- " +
          "classify the work as deterministic or judgment before delegating",
      )
    }
    expect(fake.calls.creates).toHaveLength(0)
  })

  it.each([0, 1.5, "3"])("rejects invalid phase %p with INVALID_INPUT before any child spawn", async (phase) => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec()

    const result = await invokeDelegate(exec, deps, {
      prompt: "Implement task 5.2 of the change.",
      task_classification: "judgment",
      checkpoint_step: "implement",
      phase,
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      expect(result.message).toContain("phase")
    }
    // Rejected in zod validation before any child spawn or state I/O.
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
    expect(execCalls).toEqual([])
    expect(writeFileSync).not.toHaveBeenCalled()
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — task classification and skill contract
// ---------------------------------------------------------------------------

describe("uf_delegate_command task classification and skill contract", () => {
  it("requires task_classification on prompt-only delegation", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Implement task 1.1 of the change.",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "task_classification: prompt-only delegation (no command_path) must declare task_classification -- " +
          "classify the work as deterministic or judgment before delegating",
      )
    }
    // Rejected in zod validation before any file read or session activity.
    expect(reads.calls).toEqual([])
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("proceeds with prompt-only delegation classified judgment (judgment pass-through)", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Evaluate two design approaches for the parser.",
      task_classification: "judgment",
    })

    // Judgment work carries no skill requirement and passes through.
    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)
  })

  it("returns SKILL_REFERENCE_REQUIRED when deterministic classification supplies neither skill nor reason", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec()

    const result = await invokeDelegate(exec, deps, {
      prompt: "Implement task 3.2 of the change.",
      task_classification: "deterministic",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("SKILL_REFERENCE_REQUIRED")
      expect(result.retryable).toBe(false)
      // The envelope names every remediation the contract allows: pass a
      // skill reference, record the absence reason, or classify as judgment.
      expect(result.message).toBe(
        "Deterministic-classified delegation requires a skill reference or a recorded absence reason. " +
          "Remediation: pass a skill reference (skill), record the absence reason (skill_absent_reason), " +
          "or classify the work as judgment; retrying will not fix this.",
      )
    }
    // Fail-closed before host-model resolution and child spawn: no parent
    // introspection, no session activity, no state I/O, no command-file read.
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
    expect(fake.calls.aborts).toHaveLength(0)
    expect(reads.calls).toEqual([])
    expect(execCalls).toEqual([])
    expect(writeFileSync).not.toHaveBeenCalled()
  })

  it("returns SKILL_REFERENCE_REQUIRED for a command-path call classified deterministic without a reference", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({ [resolve(FAKE_CWD, "cmd.md")]: "Run the delegated command." })
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      command_path: "cmd.md",
      task_classification: "deterministic",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("SKILL_REFERENCE_REQUIRED")
      expect(result.retryable).toBe(false)
    }
    // The requirement binds any call declaring deterministic work --
    // command-path included -- and fires before the command file is read.
    expect(reads.calls).toEqual([])
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("proceeds with deterministic classification carrying a recorded absence reason", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Rename the variable across the module.",
      task_classification: "deterministic",
      skill_absent_reason: "no repo script exists for this one-off rename",
    })

    // The recorded reason satisfies the reference requirement, so the run
    // proceeds with no skill lookup (spec scenario: reason recorded)...
    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)
    expect(artifact.calls).toEqual([])
    // ...and no skill directive is injected: only a resolved reference
    // renders one, so the composed prompt is the bare instruction.
    expect(deliveredPrompt(fake.calls)).toBe("Rename the variable across the module.")
    // The success envelope records the classification and the reason (the
    // checkpoint-data half of the scenario is pinned by the audit-trail
    // suite's verbatim-recording test).
    if (result.status === "ok") {
      expect(result.data.task_classification).toBe("deterministic")
      expect(result.data.skill).toBeNull()
      expect(result.data.skill_absent_reason).toBe("no repo script exists for this one-off rename")
    }
  })

  it("accepts deterministic classification carrying a skill reference", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".opencode/skills/pre-flight/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      skill: "pre-flight",
    })

    // The supplied skill satisfies the deterministic reference requirement
    // (task 1.2) and resolves under the first skill root (task 1.3), so
    // the run proceeds after the existence probe.
    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)
    expect(artifact.calls).toEqual([resolve(FAKE_CWD, ".opencode/skills/pre-flight/SKILL.md")])
  })

  it.each([
    ["skill", "pre-flight"],
    ["skill_absent_reason", "no repo script exists for this one-off rename"],
  ] as const)("rejects judgment-classified %s with INVALID_INPUT", async (field, value) => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const args: Record<string, unknown> = {
      prompt: "Evaluate two design approaches for the parser.",
      task_classification: "judgment",
      [field]: value,
    }
    const result = await invokeDelegate(recordingExec().exec, deps, args)

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        `${field}: skill references apply to deterministic work only -- ` +
          "classify the work as deterministic or omit the field",
      )
    }
    // The skill directive asserts deterministic work: no session spawned.
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
  })

  it.each([
    ["skill", "pre-flight"],
    ["skill_absent_reason", "no repo script exists for this one-off rename"],
  ] as const)("rejects unclassified command-path calls supplying %s", async (field, value) => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const args: Record<string, unknown> = {
      command_path: "cmd.md",
      [field]: value,
    }
    const result = await invokeDelegate(recordingExec().exec, deps, args)

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        `${field}: skill references apply to deterministic work only -- ` +
          "classify the work as deterministic or omit the field",
      )
    }
    // Rejected before the command file was read or any session spawned.
    expect(reads.calls).toEqual([])
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("rejects skill and skill_absent_reason supplied together", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      skill: "pre-flight",
      skill_absent_reason: "no repo script exists for this one-off rename",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "skill: skill and skill_absent_reason are mutually exclusive -- " +
          "pass a skill reference or an absence reason, never both",
      )
    }
    expect(fake.calls.creates).toHaveLength(0)
  })

  it.each([
    ["skill", "x".repeat(257)],
    ["skill_absent_reason", "x".repeat(513)],
  ] as const)("rejects over-length %s values", async (field, value) => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const args: Record<string, unknown> = {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      [field]: value,
    }
    const result = await invokeDelegate(recordingExec().exec, deps, args)

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      // The path prefix names the offending field (zod's bound-message
      // wording is version-specific; the field identity is the contract).
      expect(result.message).toContain(field)
    }
    expect(fake.calls.creates).toHaveLength(0)
  })

  it.each([
    ["skill", "x".repeat(256)],
    ["skill_absent_reason", "x".repeat(512)],
  ] as const)("accepts %s at exactly its length bound", async (field, value) => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    // The at-bound skill name must also resolve (task 1.3): it exists under
    // the first skill root, so the length check and resolution both pass.
    const existing = field === "skill" ? [`.opencode/skills/${value}/SKILL.md`] : []
    const artifact = fakeFileExists(existing)
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const args: Record<string, unknown> = {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      [field]: value,
    }
    const result = await invokeDelegate(recordingExec().exec, deps, args)

    // The bounds are inclusive: exactly 256 skill / 512 reason characters
    // pass input validation and the run proceeds.
    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)
  })

  it("rejects an empty-string skill, instructing how absence is expressed", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      skill: "",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "skill: empty string is invalid -- pass a skill reference or record skill_absent_reason instead",
      )
    }
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("rejects an empty-string skill_absent_reason, instructing how absence is expressed", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      skill_absent_reason: "",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "skill_absent_reason: empty string is invalid -- absence is expressed by omitting the field",
      )
    }
    expect(fake.calls.creates).toHaveLength(0)
  })

  it.each(["Deterministic", "DETERMINISTIC", "mechanical", "deterministic "])(
    "rejects non-enum task_classification %p with INVALID_INPUT naming the accepted values",
    async (taskClassification) => {
      const fake = fakeClient()
      const reads = fakeReadText({})
      const artifact = fakeFileExists([])
      const deps: DelegateCommandDependencies = {
        client: fake.client,
        modelMatrix: matrixFreeDeps,
        readText: reads.readText,
        fileExists: artifact.fileExists,
      }

      const result = await invokeDelegate(recordingExec().exec, deps, {
        prompt: "Run the repository's pre-flight procedure.",
        task_classification: taskClassification,
      })

      expect(result.status).toBe("error")
      if (result.status === "error") {
        expect(result.code).toBe("INVALID_INPUT")
        expect(result.retryable).toBe(false)
        expect(result.message).toContain("task_classification")
        // The failure names the accepted enum values.
        expect(result.message).toContain("deterministic")
        expect(result.message).toContain("judgment")
      }
      expect(fake.calls.creates).toHaveLength(0)
    },
  )

  it("rejects a non-string task_classification value with INVALID_INPUT", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: 42,
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("INVALID_INPUT")
      expect(result.retryable).toBe(false)
      expect(result.message).toContain("task_classification")
    }
    expect(fake.calls.creates).toHaveLength(0)
  })

  it("proceeds with command-path delegation supplying no classification (command-path exemption)", async () => {
    const commandAbs = resolve(FAKE_CWD, "cmd.md")
    const fake = fakeClient()
    const reads = fakeReadText({ [commandAbs]: "Run the delegated command." })
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, { command_path: "cmd.md" })

    // The Steps 2/3 call shape is unaffected: no classification and no
    // skill fields; the run proceeds exactly as before this change (the
    // matrix probe is the one new read, zero-config D8).
    expect(result.status).toBe("ok")
    expect(reads.calls).toEqual([commandAbs, MATRIX_PROBE])
    expect(fake.calls.prompts).toHaveLength(1)
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — skill resolution (delegation-skill-enforcement D2)
// ---------------------------------------------------------------------------

describe("uf_delegate_command skill resolution", () => {
  it("resolves a skill name present only under the second root (.opencode/skill/)", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".opencode/skill/unbound-force-heroes/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Load the hero routing skill.",
      task_classification: "deterministic",
      skill: "unbound-force-heroes",
    })

    // The name misses the first root, hits the second, and the run proceeds.
    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)
    // Both roots are probed in order (D2: .opencode/skills/ then
    // .opencode/skill/) through the injected FileExists only.
    expect(artifact.calls).toEqual([
      resolve(FAKE_CWD, ".opencode/skills/unbound-force-heroes/SKILL.md"),
      resolve(FAKE_CWD, ".opencode/skill/unbound-force-heroes/SKILL.md"),
    ])
    expect(existsSync).not.toHaveBeenCalled()
  })

  it("resolves a same-name collision across the roots to the first root", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([
      ".opencode/skills/colliding-skill/SKILL.md",
      ".opencode/skill/colliding-skill/SKILL.md",
    ])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the colliding skill.",
      task_classification: "deterministic",
      skill: "colliding-skill",
    })

    // First root wins (spec scenario): probing stops at the
    // .opencode/skills/ entry — the second root is never consulted...
    expect(result.status).toBe("ok")
    expect(artifact.calls).toEqual([
      resolve(FAKE_CWD, ".opencode/skills/colliding-skill/SKILL.md"),
    ])
    // ...and that first-root entry is the resolved reference recorded in
    // the audit trail (spec scenario): the success envelope carries it.
    if (result.status === "ok") {
      expect(result.data.skill).toBe(".opencode/skills/colliding-skill/SKILL.md")
    }
  })

  it("resolves a script path without any skill-directory lookup", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".specify/scripts/bash/check-prerequisites.sh"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the prerequisites check.",
      task_classification: "deterministic",
      skill: ".specify/scripts/bash/check-prerequisites.sh",
    })

    // The path form is accepted as a traversal-protected repo-relative path
    // and the run proceeds (spec scenario: script-path reference).
    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)
    // Only the repo-relative script path is probed — no skill-directory
    // lookup occurs.
    expect(artifact.calls).toEqual([
      resolve(FAKE_CWD, ".specify/scripts/bash/check-prerequisites.sh"),
    ])
  })

  it("rejects a skill path escaping the repository root with SKILL_TRAVERSAL", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the outside script.",
      task_classification: "deterministic",
      skill: "../outside/repo-tool.sh",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("SKILL_TRAVERSAL")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "Path traversal blocked: skill resolves outside the repository root",
      )
    }
    // The traversal guard fires before any existence probe or session
    // activity (the command_path / expected_artifact precedent).
    expect(artifact.calls).toEqual([])
    expect(reads.calls).toEqual([])
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
  })

  it("returns SKILL_NOT_FOUND naming both searched skill roots and spawns no child session", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([]) // nothing exists
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec()

    const result = await invokeDelegate(exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      skill: "does-not-exist",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("SKILL_NOT_FOUND")
      expect(result.retryable).toBe(false)
      // The envelope names every searched location (both skill roots).
      expect(result.message).toBe(
        'Skill reference "does-not-exist" does not resolve. ' +
          "Searched: .opencode/skills/does-not-exist/SKILL.md, .opencode/skill/does-not-exist/SKILL.md. " +
          "Remediation: pass an existing skill name or a repo-relative script path, " +
          "or record skill_absent_reason when no repo skill applies; retrying will not fix this.",
      )
    }
    // Both roots were probed in order before failing closed.
    expect(artifact.calls).toEqual([
      resolve(FAKE_CWD, ".opencode/skills/does-not-exist/SKILL.md"),
      resolve(FAKE_CWD, ".opencode/skill/does-not-exist/SKILL.md"),
    ])
    // Fail-closed before host-model resolution and the child spawn: no
    // parent introspection, no session activity, no aborts, no state I/O,
    // no state-key derivation (spec scenario: spawns no child session).
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
    expect(fake.calls.aborts).toHaveLength(0)
    expect(reads.calls).toEqual([])
    expect(execCalls).toEqual([])
    expect(writeFileSync).not.toHaveBeenCalled()
  })

  it("returns SKILL_NOT_FOUND naming the searched path for a missing script", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the missing script.",
      task_classification: "deterministic",
      skill: "scripts/missing-tool.sh",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("SKILL_NOT_FOUND")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        'Skill reference "scripts/missing-tool.sh" does not resolve. ' +
          "Searched: scripts/missing-tool.sh (repo-relative path). " +
          "Remediation: pass an existing skill name or a repo-relative script path, " +
          "or record skill_absent_reason when no repo skill applies; retrying will not fix this.",
      )
    }
    // The repo-relative path was probed through the traversal guard first.
    expect(artifact.calls).toEqual([resolve(FAKE_CWD, "scripts/missing-tool.sh")])
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
  })

  it("resolves the skill before reading the command file on command-path delegation", async () => {
    const commandAbs = resolve(FAKE_CWD, "cmd.md")
    const fake = fakeClient()
    const reads = fakeReadText({ [commandAbs]: "Run the delegated command." })
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      command_path: "cmd.md",
      task_classification: "deterministic",
      skill: "does-not-exist",
    })

    // The reference fails closed before the command file is read or any
    // session spawns (mirrors the task 1.2 gate ordering: contract
    // validation completes before I/O).
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("SKILL_NOT_FOUND")
      expect(result.retryable).toBe(false)
    }
    expect(reads.calls).toEqual([])
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — skill directive injection (delegation-skill-enforcement D3)
// ---------------------------------------------------------------------------

describe("uf_delegate_command skill directive injection", () => {
  it("renders the directive after command text and feature-dir context, before the extra prompt", async () => {
    const commandAbs = resolve(FAKE_CWD, ".opencode/commands/speckit.plan.md")
    const fake = fakeClient()
    const reads = fakeReadText({ [commandAbs]: "Create the implementation plan." })
    const artifact = fakeFileExists([".opencode/skills/pre-flight/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      command_path: ".opencode/commands/speckit.plan.md",
      feature_dir: "specs/018-unleash-command/",
      prompt: "Focus on the parser.",
      task_classification: "deterministic",
      skill: "pre-flight",
    })

    // Golden shape (design D3): the directive is a fixed, tool-rendered
    // section — a header line naming the resolved reference, a blank
    // line, then the fixed paragraph — inserted after command text and
    // feature-dir context, immediately before the extra prompt. The
    // resolved reference headlines the directive and names the script
    // path; the `skill`-tool name is derived from the skill root.
    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)
    expect(deliveredPrompt(fake.calls)).toBe(
      "Create the implementation plan.\n\n" +
        "Feature directory: specs/018-unleash-command/\n\n" +
        "Required skill: .opencode/skills/pre-flight/SKILL.md\n\n" +
        "This task is classified as deterministic work. Load the skill (invoke the `skill` tool " +
        'with name "pre-flight") or run the script at .opencode/skills/pre-flight/SKILL.md, and ' +
        "execute its full procedure. Do not substitute an improvised procedure.\n\n" +
        "Focus on the parser.",
    )
  })

  it("renders the directive on prompt-only delegation without command text", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".opencode/skills/pre-flight/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    // Per-task implement call shape: prompt-only with feature-dir context.
    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      feature_dir: "openspec/changes/delegation-skill-enforcement/",
      task_classification: "deterministic",
      skill: "pre-flight",
    })

    // Without command text the directive holds its fixed position: after
    // feature-dir context, immediately before the extra prompt.
    expect(result.status).toBe("ok")
    expect(reads.calls).toEqual([MATRIX_PROBE])
    expect(deliveredPrompt(fake.calls)).toBe(
      "Feature directory: openspec/changes/delegation-skill-enforcement/\n\n" +
        "Required skill: .opencode/skills/pre-flight/SKILL.md\n\n" +
        "This task is classified as deterministic work. Load the skill (invoke the `skill` tool " +
        'with name "pre-flight") or run the script at .opencode/skills/pre-flight/SKILL.md, and ' +
        "execute its full procedure. Do not substitute an improvised procedure.\n\n" +
        "Run the repository's pre-flight procedure.",
    )
  })

  it("renders the directive for a script-path reference with the path in every slot", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".specify/scripts/bash/check-prerequisites.sh"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the prerequisites check.",
      task_classification: "deterministic",
      skill: ".specify/scripts/bash/check-prerequisites.sh",
    })

    // A script reference resolves to itself: the resolved path headlines
    // the directive and fills the script slot, and — carrying no skill
    // name — also fills the `skill`-tool slot, where the "or run the
    // script" clause is the operative instruction (spec scenario:
    // script-path skill).
    expect(result.status).toBe("ok")
    expect(deliveredPrompt(fake.calls)).toBe(
      "Required skill: .specify/scripts/bash/check-prerequisites.sh\n\n" +
        "This task is classified as deterministic work. Load the skill (invoke the `skill` tool " +
        'with name ".specify/scripts/bash/check-prerequisites.sh") or run the script at ' +
        ".specify/scripts/bash/check-prerequisites.sh, and execute its full procedure. " +
        "Do not substitute an improvised procedure.\n\n" +
        "Run the prerequisites check.",
    )
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — expected-artifact verification
// ---------------------------------------------------------------------------

describe("uf_delegate_command expected-artifact verification", () => {
  it("verifies the expected artifact after a successful run and returns the full success envelope", async () => {
    const fake = fakeClient({
      parts: [textPart("plan written"), finishPart(0.25, 11, 7)],
    })
    const reads = fakeReadText({})
    const artifact = fakeFileExists(["specs/018-unleash-command/plan.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Write the plan.",
      task_classification: "judgment",
      expected_artifact: "specs/018-unleash-command/plan.md",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data).toStrictEqual({
        agent: "cobalt-crush-dev",
        status: "success",
        child_session_id: "child-session",
        artifact_created: true,
        checkpoint_recorded: false,
        // Delegation-skill-enforcement D5 (task 1.5): the audit fields
        // surface in the envelope -- judgment classification, no skill
        // reference, no absence reason.
        task_classification: "judgment",
        skill: null,
        skill_absent_reason: null,
        // Model-matrix-delegation D7 (task 5.4): the selection audit
        // fields surface in every success envelope — zero-config calls
        // record the host fallback with no resolved values.
        selection_source: "host",
        selection_pack: null,
        selection_capability: null,
        selection_model: null,
        selection_confidence: null,
        selection_fallback_reason: null,
        usage: { cost_usd: 0.25, tokens: { input: 11, output: 7 } },
        provenance: {
          agent: "cobalt-crush-dev",
          requested_model: null,
          requested_variant: null,
          resolved_parent_model: HOST_MODEL,
          resolved_parent_variant: "default",
          reported_child_model: HOST_MODEL,
          model_mismatch: false,
          // Content delegation writes files, unlike review invocations.
          read_only: false,
        },
        text: "plan written",
      })
    }
    // Verification probes the artifact through the injected FileExists only.
    expect(artifact.calls).toEqual([resolve(FAKE_CWD, "specs/018-unleash-command/plan.md")])
    expect(existsSync).not.toHaveBeenCalled()
  })

  it("returns the exact plan.md STOP message and records no checkpoint when the artifact is missing", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec()

    const result = await invokeDelegate(exec, deps, {
      prompt: "Write the plan.",
      task_classification: "judgment",
      expected_artifact: "specs/018-unleash-command/plan.md",
      checkpoint_step: "plan",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("EXPECTED_ARTIFACT_MISSING")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "Plan generation failed -- plan.md was not created. Check the agent output for errors.",
      )
    }
    // The child session ran; verification failed only after it finished.
    expect(fake.calls.creates).toHaveLength(1)
    expect(fake.calls.prompts).toHaveLength(1)
    // An unverified run never checkpoints (D5: success-only) and never
    // derives a state key.
    expect(execCalls).toEqual([])
    expect(lstatSync).not.toHaveBeenCalled()
    expect(readFileSync).not.toHaveBeenCalled()
    expect(mkdirSync).not.toHaveBeenCalled()
    expect(writeFileSync).not.toHaveBeenCalled()
  })

  it("returns the exact tasks.md STOP message when tasks.md is missing", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Generate the task list.",
      task_classification: "judgment",
      expected_artifact: "openspec/changes/unleash-deterministic-tools/tasks.md",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("EXPECTED_ARTIFACT_MISSING")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "Task generation failed -- tasks.md was not created. Check the agent output for errors.",
      )
    }
  })

  it("returns the generic artifact STOP message naming the missing path", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Write the design record.",
      task_classification: "judgment",
      expected_artifact: "specs/018-unleash-command/design.md",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("EXPECTED_ARTIFACT_MISSING")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "Artifact generation failed -- specs/018-unleash-command/design.md was not created. " +
          "Check the agent output for errors.",
      )
    }
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — child-session failure propagation
// ---------------------------------------------------------------------------

describe("uf_delegate_command child-session propagation", () => {
  it("propagates a timed-out child session as DELEGATE_FAILED with a retryable error", async () => {
    const fake = fakeClient({ hangPrompt: true })
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the delegation.",
      task_classification: "judgment",
      timeout: 5,
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("DELEGATE_FAILED")
      expect(result.retryable).toBe(true)
      expect(result.message).toBe("timeout: invoke_agent run timed out")
      // The failure envelope threads the child session id so the timed-out
      // child subtree can be inspected directly from the delegation result.
      expect(result.child_session_id).toBe("child-session")
    }
    // The executor aborted the hanging child session.
    expect(fake.calls.aborts).toHaveLength(1)
  })

  it("propagates parent cancellation of the child session as DELEGATE_CANCELLED", async () => {
    const controller = new AbortController()
    const fake = fakeClient({ hangPrompt: true })
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const pending = invokeDelegate(
      recordingExec().exec,
      deps,
      { prompt: "Run the delegation.", task_classification: "judgment" },
      controller.signal,
    )
    await new Promise((resolve) => setTimeout(resolve, 0))
    controller.abort(new Error("parent stopped"))
    const result = await pending

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("DELEGATE_CANCELLED")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe("cancelled: invoke_agent run was cancelled")
      // DELEGATE_CANCELLED carries the child session id too: the child
      // existed before the parent abort reached it.
      expect(result.child_session_id).toBe("child-session")
    }
    // Cancellation reached the child session with an independent cleanup
    // signal (invoke-agent precedent).
    expect(fake.calls.aborts).toHaveLength(1)
    expect(fake.calls.aborts[0]?.path).toEqual({ id: "child-session" })
    expect(fake.calls.aborts[0]?.signal).not.toBe(controller.signal)
  })

  it("returns DELEGATE_CANCELLED before spawning when cancellation hits host resolution", async () => {
    const controller = new AbortController()
    const fake = fakeClient({ hangMessage: true })
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const pending = invokeDelegate(
      recordingExec().exec,
      deps,
      { prompt: "Run the delegation.", task_classification: "judgment" },
      controller.signal,
    )
    await new Promise((resolve) => setTimeout(resolve, 0))
    controller.abort(new Error("parent stopped"))
    const result = await pending

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("DELEGATE_CANCELLED")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe("uf_delegate_command run was cancelled")
      // No child session ever existed, so the envelope omits the field.
      expect(result.child_session_id).toBeUndefined()
    }
    // Cancelled before any child session existed.
    expect(fake.calls.messages).toHaveLength(1)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
    expect(fake.calls.aborts).toHaveLength(0)
  })

  it("propagates child creation failures as DELEGATE_FAILED with sanitized errors and no checkpoint", async () => {
    const fake = fakeClient({ createError: "credential=abc /home/alice/private" })
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec()

    const result = await invokeDelegate(exec, deps, {
      prompt: "Run the delegation.",
      task_classification: "judgment",
      checkpoint_step: "plan",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("DELEGATE_FAILED")
      expect(result.retryable).toBe(true)
      expect(result.message).toContain("child_create_failed:")
      // No child session ever existed, so the envelope omits the field
      // (the DELEGATE_FAILED counterpart of the cancelled-before-child case).
      expect(result.child_session_id).toBeUndefined()
      // The executor's sanitizer redacted the credential and home path.
      expect(result.message).not.toMatch(/abc|\/home\/alice/)
    }
    // Failed runs never checkpoint (D5) and never derive a state key.
    expect(execCalls).toEqual([])
    expect(writeFileSync).not.toHaveBeenCalled()
  })

  it("fails with HOST_MODEL_UNAVAILABLE, non-retryable, when the host variant is structurally invalid", async () => {
    const fake = fakeClient({ hostVariant: "not valid" })
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the delegation.",
      task_classification: "judgment",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("HOST_MODEL_UNAVAILABLE")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "current assistant active variant is invalid. Remediation: run the command from a provider whose active " +
          "variant is a valid slug; retrying will not fix this.",
      )
    }
    // No child session was spawned for an unresolvable host model.
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
  })

  it("fails with HOST_MODEL_UNAVAILABLE, retryable, when host introspection hits a transient error", async () => {
    const fake = fakeClient({ messageError: "service unavailable" })
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the delegation.",
      task_classification: "judgment",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("HOST_MODEL_UNAVAILABLE")
      expect(result.retryable).toBe(true)
    }
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
  })

  it("delegates without a variant override when the host assistant message reports no variant", async () => {
    const fake = fakeClient({ hostVariant: undefined })
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the delegation.",
      task_classification: "judgment",
    })

    expect(result.status).toBe("ok")
    expect(fake.calls.creates[0]?.body).toMatchObject({
      title: `cobalt-crush-dev @ ${HOST_MODEL}`,
    })
    expect(fake.calls.prompts[0]?.body).not.toHaveProperty("variant")
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — manifest isolation
// ---------------------------------------------------------------------------

describe("uf_delegate_command manifest isolation", () => {
  it("never consults .uf/reviewer-capabilities.yaml through any dependency or filesystem surface", async () => {
    const commandAbs = resolve(FAKE_CWD, ".opencode/commands/speckit.tasks.md")
    const fake = fakeClient()
    const reads = fakeReadText({ [commandAbs]: "Generate the task list." })
    const artifact = fakeFileExists(["openspec/changes/unleash-deterministic-tools/tasks.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      command_path: ".opencode/commands/speckit.tasks.md",
      feature_dir: "openspec/changes/unleash-deterministic-tools/",
      expected_artifact: "openspec/changes/unleash-deterministic-tools/tasks.md",
      checkpoint_step: "tasks",
    })

    expect(result.status).toBe("ok")

    // The only text reads are the command file and the zero-config matrix
    // probe — the reviewer manifest is not an input to content delegation
    // (D2 boundary is agent validation).
    expect(reads.calls).toEqual([commandAbs, MATRIX_PROBE])

    // No filesystem probe touches the manifest, and existence checks flow
    // exclusively through the injected FileExists.
    const probed = [
      ...vi.mocked(existsSync).mock.calls.map((call) => String(call[0])),
      ...vi.mocked(lstatSync).mock.calls.map((call) => String(call[0])),
      ...vi.mocked(readFileSync).mock.calls.map((call) => String(call[0])),
      ...vi.mocked(writeFileSync).mock.calls.map((call) => String(call[0])),
    ]
    expect(probed.length).toBeGreaterThan(0)
    for (const path of probed) {
      expect(path.includes("reviewer-capabilities")).toBe(false)
    }
    expect(existsSync).not.toHaveBeenCalled()
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — checkpoint recording semantics
// ---------------------------------------------------------------------------

describe("uf_delegate_command checkpoint recording", () => {
  it("auto-checkpoints checkpoint_step on verified success with the delegation facts as data", async () => {
    const commandAbs = resolve(FAKE_CWD, ".opencode/commands/speckit.plan.md")
    const fake = fakeClient()
    const reads = fakeReadText({ [commandAbs]: "Create the implementation plan." })
    const artifact = fakeFileExists(["specs/018-unleash-command/plan.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec("opsx/unleash-deterministic-tools")

    const result = await invokeDelegate(exec, deps, {
      command_path: ".opencode/commands/speckit.plan.md",
      feature_dir: "specs/018-unleash-command/",
      prompt: "Focus on the parser.",
      expected_artifact: "specs/018-unleash-command/plan.md",
      checkpoint_step: "plan",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.checkpoint_recorded).toBe(true)
      expect(result.data?.artifact_created).toBe(true)
    }

    // Exactly one git invocation — the state-key branch read — against the
    // injected cwd.
    expect(execCalls).toEqual([{ args: ["rev-parse", "--abbrev-ref", "HEAD"], cwd: FAKE_CWD }])

    // D5: the checkpoint persists through the writeState path —
    // .uf/progress/<sanitized-branch>/uf.unleash.json with 0o600 file mode.
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [statePathArg, stateRaw, writeOptions] = vi.mocked(writeFileSync).mock.calls[0]
    expect(statePathArg).toBe(
      resolve(FAKE_CWD, ".uf/progress/opsx-unleash-deterministic-tools/uf.unleash.json"),
    )
    expect(writeOptions).toMatchObject({ mode: 0o600 })
    const state = JSON.parse(stateRaw as string) as ProgressState
    expect(state.command).toBe("uf.unleash")
    expect(state.branch).toBe("opsx-unleash-deterministic-tools")
    expect(state.key).toBe("opsx-unleash-deterministic-tools")
    expect(state.current_step).toBe("plan")
    expect(state.completed_steps).toEqual(["plan"])
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      command_path: ".opencode/commands/speckit.plan.md",
      feature_dir: "specs/018-unleash-command/",
      expected_artifact: "specs/018-unleash-command/plan.md",
      selection_source: "host",
    })

    expect(mkdirSync).toHaveBeenCalledTimes(1)
    const [mkdirPathArg, mkdirOptions] = vi.mocked(mkdirSync).mock.calls[0]
    expect(mkdirPathArg).toBe(resolve(FAKE_CWD, ".uf/progress/opsx-unleash-deterministic-tools"))
    expect(mkdirOptions).toMatchObject({ recursive: true, mode: 0o700 })
  })

  it("records implement checkpoint data with the phase for per-task delegation", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec("opsx/unleash-deterministic-tools")

    // Per-task implement call shape (command Step 5): prompt-only, no
    // command_path or expected_artifact, with the 1-based phase number.
    const result = await invokeDelegate(exec, deps, {
      prompt: "Implement task 5.2 of the change.",
      task_classification: "judgment",
      feature_dir: "openspec/changes/unleash-deterministic-tools/",
      checkpoint_step: "implement",
      phase: 3,
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.checkpoint_recorded).toBe(true)
    }
    expect(execCalls).toEqual([{ args: ["rev-parse", "--abbrev-ref", "HEAD"], cwd: FAKE_CWD }])

    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    const state = JSON.parse(stateRaw as string) as ProgressState
    expect(state.current_step).toBe("implement")
    expect(state.completed_steps).toEqual(["implement"])
    // D5: "implement" with phase data — the per-task call merges the
    // phase into the checkpoint data alongside the delegation facts; the
    // declared classification is recorded and judgment calls add no skill
    // keys (task 1.5 audit trail).
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      feature_dir: "openspec/changes/unleash-deterministic-tools/",
      phase: 3,
      task_classification: "judgment",
      selection_source: "host",
    })
  })

  it("records no phase key in checkpoint data when phase is omitted", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Implement task 5.2 of the change.",
      task_classification: "judgment",
      feature_dir: "openspec/changes/unleash-deterministic-tools/",
      checkpoint_step: "implement",
    })

    expect(result.status).toBe("ok")
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    const state = JSON.parse(stateRaw as string) as ProgressState
    // Existing checkpoint shapes are preserved: absent phase adds no key;
    // the declared classification is the one audit key a judgment call
    // records (task 1.5).
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      feature_dir: "openspec/changes/unleash-deterministic-tools/",
      task_classification: "judgment",
      selection_source: "host",
    })
    expect(state.data).not.toHaveProperty("phase")
  })

  it("appends checkpoint_step idempotently to existing checkpoint state", async () => {
    // Simulate a re-run: prior state exists with the branch gate recorded,
    // a later current step, and stale feature-dir data.
    const existing: ProgressState = {
      command: "uf.unleash",
      branch: "opsx-unleash-deterministic-tools",
      key: "opsx-unleash-deterministic-tools",
      current_step: "branch-safety-gate",
      completed_steps: ["branch-safety-gate"],
      data: { feature_dir: "specs/018-stale-dir/" },
      created_at: "2026-10-14T00:00:00.000Z",
      updated_at: "2026-10-14T00:00:00.000Z",
    }
    vi.mocked(lstatSync).mockImplementation(
      () => ({ isSymbolicLink: () => false }) as unknown as ReturnType<typeof lstatSync>,
    )
    vi.mocked(readFileSync).mockImplementation(() => JSON.stringify(existing))

    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Write the plan.",
      task_classification: "judgment",
      checkpoint_step: "plan",
    })

    expect(result.status).toBe("ok")

    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    const state = JSON.parse(stateRaw as string) as ProgressState
    // The new step appends without duplicating earlier ones (FR-001)...
    expect(state.completed_steps).toEqual(["branch-safety-gate", "plan"])
    expect(state.current_step).toBe("plan")
    // ...shallow data merge adds this run's delegation facts while the stale
    // feature_dir survives (this run supplied none), and the judgment
    // classification is recorded (task 1.5)...
    expect(state.data).toEqual({
      feature_dir: "specs/018-stale-dir/",
      agent: "cobalt-crush-dev",
      task_classification: "judgment",
      selection_source: "host",
    })
    // ...and created_at survives while updated_at is refreshed.
    expect(state.created_at).toBe("2026-10-14T00:00:00.000Z")
    expect(state.updated_at).not.toBe("2026-10-14T00:00:00.000Z")
  })

  it("records no checkpoint when checkpoint_step is omitted", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec()

    const result = await invokeDelegate(exec, deps, {
      prompt: "Run the delegation.",
      task_classification: "judgment",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data?.checkpoint_recorded).toBe(false)
    }
    // No state-key derivation and no state I/O happened.
    expect(execCalls).toEqual([])
    expect(lstatSync).not.toHaveBeenCalled()
    expect(readFileSync).not.toHaveBeenCalled()
    expect(mkdirSync).not.toHaveBeenCalled()
    expect(writeFileSync).not.toHaveBeenCalled()
  })

  it("fails closed with DELEGATE_ERROR when the checkpoint cannot be written", async () => {
    vi.mocked(writeFileSync).mockImplementation(() => {
      throw new Error("EACCES: permission denied writing state file")
    })
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the delegation.",
      task_classification: "judgment",
      checkpoint_step: "plan",
    })

    // recordAutoCheckpoint throws on state I/O failure so a delegation never
    // reports success without a durable checkpoint (fail closed, D5).
    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("DELEGATE_ERROR")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe("EACCES: permission denied writing state file")
    }
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — audit trail (delegation-skill-enforcement D5)
// ---------------------------------------------------------------------------

describe("uf_delegate_command audit trail", () => {
  it("records task_classification and the resolved skill reference in checkpoint data", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".opencode/skills/pre-flight/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec("opsx/delegation-skill-enforcement")

    // Per-task implement call shape: prompt-only, classified deterministic
    // with a skill reference, checkpointed with the phase number.
    const result = await invokeDelegate(exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      feature_dir: "openspec/changes/delegation-skill-enforcement/",
      task_classification: "deterministic",
      skill: "pre-flight",
      checkpoint_step: "implement",
      phase: 1,
    })

    expect(result.status).toBe("ok")
    expect(execCalls).toEqual([{ args: ["rev-parse", "--abbrev-ref", "HEAD"], cwd: FAKE_CWD }])
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [statePathArg, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    expect(statePathArg).toBe(
      resolve(FAKE_CWD, ".uf/progress/opsx-delegation-skill-enforcement/uf.unleash.json"),
    )
    const state = JSON.parse(stateRaw as string) as ProgressState
    // D5 (task 1.5): the checkpoint records the declared classification and
    // the RESOLVED reference (.opencode/skills/pre-flight/SKILL.md, not the
    // raw "pre-flight" input); a skill-carrying call records no reason.
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      feature_dir: "openspec/changes/delegation-skill-enforcement/",
      phase: 1,
      task_classification: "deterministic",
      skill: ".opencode/skills/pre-flight/SKILL.md",
      selection_source: "host",
    })
    expect(state.data).not.toHaveProperty("skill_absent_reason")
  })

  it("records task_classification and skill_absent_reason verbatim in checkpoint data", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const reason =
      'searched .opencode/skills/ and .opencode/skill/ for a "rename" skill; none exists for this one-off'

    const result = await invokeDelegate(recordingExec("opsx/delegation-skill-enforcement").exec, deps, {
      prompt: "Rename the variable across the module.",
      task_classification: "deterministic",
      skill_absent_reason: reason,
      checkpoint_step: "implement",
    })

    expect(result.status).toBe("ok")
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    const state = JSON.parse(stateRaw as string) as ProgressState
    // D5 (task 1.5): the reason is recorded verbatim -- the checkpoint
    // store's JSON serialization is the sanitization boundary, no content
    // filtering -- and a reason-carrying call records no skill reference.
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      task_classification: "deterministic",
      skill_absent_reason: reason,
      selection_source: "host",
    })
    expect(state.data).not.toHaveProperty("skill")
  })

  it("round-trips the audit fields through the progress store", async () => {
    const commandAbs = resolve(FAKE_CWD, "cmd.md")
    const fake = fakeClient()
    const reads = fakeReadText({ [commandAbs]: "Run the delegated command." })
    const artifact = fakeFileExists([".opencode/skills/pre-flight/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const exec = recordingExec("opsx/delegation-skill-enforcement").exec

    // Run 1: a deterministic delegation writes the audit fields.
    const first = await invokeDelegate(exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      skill: "pre-flight",
      checkpoint_step: "implement",
    })
    expect(first.status).toBe("ok")
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    const written = JSON.parse(stateRaw as string) as ProgressState
    expect(written.data).toMatchObject({
      task_classification: "deterministic",
      skill: ".opencode/skills/pre-flight/SKILL.md",
    })

    // Round-trip: the store's read path (the same readState get_progress
    // reads through) parses run 1's serialization...
    vi.mocked(lstatSync).mockImplementation(
      () => ({ isSymbolicLink: () => false }) as unknown as ReturnType<typeof lstatSync>,
    )
    vi.mocked(readFileSync).mockImplementation(() => stateRaw)

    // ...and run 2, a command-path call supplying no classification, merges
    // over the parsed state: run 1's audit fields survive the round-trip.
    const second = await invokeDelegate(exec, deps, {
      command_path: "cmd.md",
      checkpoint_step: "plan",
    })
    expect(second.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(2)
    expect(writeFileSync).toHaveBeenCalledTimes(2)
    const [, mergedRaw] = vi.mocked(writeFileSync).mock.calls[1]
    const merged = JSON.parse(mergedRaw as string) as ProgressState
    expect(merged.completed_steps).toEqual(["implement", "plan"])
    expect(merged.current_step).toBe("plan")
    expect(merged.data).toEqual({
      task_classification: "deterministic",
      skill: ".opencode/skills/pre-flight/SKILL.md",
      agent: "cobalt-crush-dev",
      command_path: "cmd.md",
      selection_source: "host",
    })
  })

  it("surfaces the classification and resolved skill reference in the success envelope", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".opencode/skills/pre-flight/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      skill: "pre-flight",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      // D5 (task 1.5): machine-parseable envelope fields -- the declared
      // classification, the RESOLVED reference, and a null absence reason.
      expect(result.data.task_classification).toBe("deterministic")
      expect(result.data.skill).toBe(".opencode/skills/pre-flight/SKILL.md")
      expect(result.data.skill_absent_reason).toBeNull()
    }
  })

  it("surfaces the recorded absence reason in the success envelope", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Rename the variable across the module.",
      task_classification: "deterministic",
      skill_absent_reason: "no repo script exists for this one-off rename",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data.task_classification).toBe("deterministic")
      expect(result.data.skill).toBeNull()
      expect(result.data.skill_absent_reason).toBe("no repo script exists for this one-off rename")
    }
  })

  it("surfaces judgment classification with no skill reference or absence reason in the success envelope", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Evaluate two design approaches for the parser.",
      task_classification: "judgment",
    })

    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      // Spec: judgment-classified calls MUST NOT carry a skill reference or
      // absence reason -- the envelope reports the classification with
      // null skill fields.
      expect(result.data.task_classification).toBe("judgment")
      expect(result.data.skill).toBeNull()
      expect(result.data.skill_absent_reason).toBeNull()
    }
  })

  it("leaves checkpoint data and envelope audit fields absent for unclassified command-path calls", async () => {
    const commandAbs = resolve(FAKE_CWD, "cmd.md")
    const fake = fakeClient()
    const reads = fakeReadText({ [commandAbs]: "Run the delegated command." })
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      command_path: "cmd.md",
      checkpoint_step: "plan",
    })

    // Backward compatibility (D5): a call omitting every new field behaves
    // exactly as before this change -- the run proceeds, the envelope
    // surfaces the additive fields as null...
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data.task_classification).toBeNull()
      expect(result.data.skill).toBeNull()
      expect(result.data.skill_absent_reason).toBeNull()
    }
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    const state = JSON.parse(stateRaw as string) as ProgressState
    // ...and the checkpoint data keeps the pre-change shape plus the one
    // always-present model-matrix audit field: selection_source records
    // "host" for every zero-config call (task 5.4); no other new keys are
    // added when the call supplied none.
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      command_path: "cmd.md",
      selection_source: "host",
    })
    expect(state.data).not.toHaveProperty("task_classification")
    expect(state.data).not.toHaveProperty("skill")
    expect(state.data).not.toHaveProperty("skill_absent_reason")
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — seam verification (DE-007, delegation-skill-enforcement
// task 4.2)
//
// This change binds a prose procedure to a tool contract, so DE-007 requires
// the increment to verify the seams it creates. Layer 1 (registration
// surface: the registered args match the documented contract) lives in
// uf-workflow.test.ts and scratch-smoke.test.ts, which pin the registered
// argument set including task_classification, skill, and skill_absent_reason.
// The tests below cover the remaining two layers: the documented output
// shape (golden assertions on the DelegateCommandData envelope and the
// checkpoint fields) and write-then-read round-trips (checkpoint data a
// delegated run records, read back through the progress store's get_progress
// read surface — complementing the audit-trail round-trip test, which
// verifies the delegate's own readState merge path).
// ---------------------------------------------------------------------------

describe("uf_delegate_command seam verification", () => {
  it("golden output shape: deterministic delegation with a skill (valued audit fields in envelope and checkpoint)", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".opencode/skills/pre-flight/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec, calls: execCalls } = recordingExec("opsx/delegation-skill-enforcement")

    const result = await invokeDelegate(exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      feature_dir: "openspec/changes/delegation-skill-enforcement/",
      task_classification: "deterministic",
      skill: "pre-flight",
      checkpoint_step: "implement",
      phase: 2,
    })

    // Layer 2 (DE-007): the documented DelegateCommandData output shape,
    // golden end to end — every field, exact values, no extras or missing
    // keys. The audit triple takes its valued shape: the declared
    // classification, the RESOLVED reference, a null absence reason.
    // (The judgment-shaped full golden lives in the expected-artifact
    // suite; this is the valued-shape counterpart.)
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data).toStrictEqual({
        agent: "cobalt-crush-dev",
        status: "success",
        child_session_id: "child-session",
        artifact_created: false,
        checkpoint_recorded: true,
        task_classification: "deterministic",
        skill: ".opencode/skills/pre-flight/SKILL.md",
        skill_absent_reason: null,
        // Model-matrix-delegation D7 (task 5.4): zero-config selection
        // audit — host fallback, no resolved values.
        selection_source: "host",
        selection_pack: null,
        selection_capability: null,
        selection_model: null,
        selection_confidence: null,
        selection_fallback_reason: null,
        usage: null,
        provenance: {
          agent: "cobalt-crush-dev",
          requested_model: null,
          requested_variant: null,
          resolved_parent_model: HOST_MODEL,
          resolved_parent_variant: "default",
          reported_child_model: HOST_MODEL,
          model_mismatch: false,
          read_only: false,
        },
        text: "delegated output",
      })
    }

    // Layer 2 (DE-007): the documented checkpoint shape — the full
    // ProgressState field set a delegated run persists (phase-scoped
    // records belong to the checkpoint tool's surface, not the
    // delegate's auto-checkpoint)...
    expect(execCalls).toEqual([{ args: ["rev-parse", "--abbrev-ref", "HEAD"], cwd: FAKE_CWD }])
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [statePathArg, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    expect(statePathArg).toBe(
      resolve(FAKE_CWD, ".uf/progress/opsx-delegation-skill-enforcement/uf.unleash.json"),
    )
    const state = JSON.parse(stateRaw as string) as ProgressState
    expect(Object.keys(state).sort()).toEqual([
      "branch",
      "command",
      "completed_steps",
      "created_at",
      "current_step",
      "data",
      "key",
      "updated_at",
    ])
    // ...with the audit fields recorded in the checkpoint data's valued
    // shape: classification plus resolved reference, no reason key. The
    // model-matrix audit adds selection_source only — zero-config calls
    // keep the pre-change checkpoint shape plus that one field (task 5.4).
    expect(state.current_step).toBe("implement")
    expect(state.completed_steps).toEqual(["implement"])
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      feature_dir: "openspec/changes/delegation-skill-enforcement/",
      phase: 2,
      task_classification: "deterministic",
      skill: ".opencode/skills/pre-flight/SKILL.md",
      selection_source: "host",
    })
  })

  it("golden output shape: deterministic delegation with a recorded absence reason (valued audit fields in envelope)", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const reason =
      'searched .opencode/skills/ and .opencode/skill/ for a rename skill; none exists for this one-off'

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Rename the variable across the module.",
      task_classification: "deterministic",
      skill_absent_reason: reason,
    })

    // Layer 2 (DE-007): the same documented envelope shape with the audit
    // triple's other valued form — a null skill, the verbatim absence
    // reason — and no checkpoint (checkpoint_step omitted), so the golden
    // pins that shape's documented output too.
    expect(result.status).toBe("ok")
    if (result.status === "ok") {
      expect(result.data).toStrictEqual({
        agent: "cobalt-crush-dev",
        status: "success",
        child_session_id: "child-session",
        artifact_created: false,
        checkpoint_recorded: false,
        task_classification: "deterministic",
        skill: null,
        skill_absent_reason: reason,
        // Model-matrix-delegation D7 (task 5.4): zero-config selection
        // audit — host fallback, no resolved values.
        selection_source: "host",
        selection_pack: null,
        selection_capability: null,
        selection_model: null,
        selection_confidence: null,
        selection_fallback_reason: null,
        usage: null,
        provenance: {
          agent: "cobalt-crush-dev",
          requested_model: null,
          requested_variant: null,
          resolved_parent_model: HOST_MODEL,
          resolved_parent_variant: "default",
          reported_child_model: HOST_MODEL,
          model_mismatch: false,
          read_only: false,
        },
        text: "delegated output",
      })
    }
  })

  it("reads a delegated run's recorded audit fields back through the get_progress progress store", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".opencode/skills/pre-flight/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }
    const { exec } = recordingExec("opsx/delegation-skill-enforcement")

    // Write: a delegated deterministic run auto-checkpoints the audit
    // fields through writeState's JSON serialization.
    const delegated = await invokeDelegate(exec, deps, {
      prompt: "Run the repository's pre-flight procedure.",
      task_classification: "deterministic",
      skill: "pre-flight",
      checkpoint_step: "implement",
    })
    expect(delegated.status).toBe("ok")
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [statePathArg, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    expect(statePathArg).toBe(
      resolve(FAKE_CWD, ".uf/progress/opsx-delegation-skill-enforcement/uf.unleash.json"),
    )
    const written = JSON.parse(stateRaw as string) as ProgressState

    // Read: the progress store's dedicated read surface (get_progress)
    // resolves the same state key from the same branch and parses the
    // same bytes the delegated run wrote. The audit-trail round-trip
    // test above exercises the delegate's own readState merge path;
    // this is the store's read side of the seam.
    vi.mocked(lstatSync).mockImplementation(
      () => ({ isSymbolicLink: () => false }) as unknown as ReturnType<typeof lstatSync>,
    )
    vi.mocked(readFileSync).mockImplementation(() => stateRaw)

    const progress = createGetProgressTool(exec, FAKE_CWD)
    const readResult = await (
      progress as {
        execute: (
          args: Record<string, unknown>,
          ctx: { sessionID: string },
        ) => Promise<{ output: string }>
      }
    ).execute({ command: "uf.unleash" }, { sessionID: "parent-session" })
    const readback = JSON.parse(readResult.output) as ToolResult<ProgressState>

    expect(readback.status).toBe("ok")
    if (readback.status === "ok") {
      // Exactly what the delegated run wrote is what the store returns.
      expect(readback.data).toStrictEqual(written)
    }
    // The read went through the store's state path for the same key...
    expect(vi.mocked(readFileSync).mock.calls[0][0]).toBe(
      resolve(FAKE_CWD, ".uf/progress/opsx-delegation-skill-enforcement/uf.unleash.json"),
    )
    // ...and the audit fields survived the round-trip: the declared
    // classification, the RESOLVED reference, and the recorded step.
    expect(written.current_step).toBe("implement")
    expect(written.completed_steps).toEqual(["implement"])
    expect(written.data).toEqual({
      agent: "cobalt-crush-dev",
      task_classification: "deterministic",
      skill: ".opencode/skills/pre-flight/SKILL.md",
      selection_source: "host",
    })
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — spec scenario coverage (delegation-skill-enforcement)
// ---------------------------------------------------------------------------

// End-to-end mirrors of the change's spec-delta scenarios. Tasks 1.1-1.5
// landed the rule-level tests (schema validation, failure envelopes,
// resolution, directive, audit fields); this block consolidates each spec
// scenario whose THEN clause those tests split across separate flows, and
// names the scenarios no earlier test covered: deterministic-with-skill in
// one resolved+injected+recorded flow, the quick-tier whole-change
// classification pass-through, and the structural no-model resolution
// failure. Scenarios already pinned end to end by a single named test
// (SKILL_REFERENCE_REQUIRED, SKILL_NOT_FOUND, judgment pass-through, the
// prompt-only classification requirement, over-length / empty-string /
// non-enum rejection, script-path skill, command-path exemption, and the
// no-variant spawn) are not duplicated here; the deterministic-with-reason
// scenario's checkpoint half lives in the audit-trail suite.
describe("uf_delegate_command spec scenario coverage", () => {
  it("deterministic delegation carries a skill end to end: resolved, injected, recorded, run proceeds", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([".opencode/skills/pre-flight/SKILL.md"])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    // Spec scenario: a per-task delegation classified deterministic with
    // skill "pre-flight", which exists at .opencode/skills/pre-flight/SKILL.md.
    const result = await invokeDelegate(recordingExec("opsx/delegation-skill-enforcement").exec, deps, {
      prompt: "Run the repository's pre-flight procedure for this task.",
      feature_dir: "openspec/changes/delegation-skill-enforcement/",
      task_classification: "deterministic",
      skill: "pre-flight",
      checkpoint_step: "implement",
      phase: 4,
    })

    // The run proceeds: exactly one child session spawned and prompted.
    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)

    // Resolved: the reference resolved under the first skill root through
    // the injected existence probe.
    expect(artifact.calls).toEqual([resolve(FAKE_CWD, ".opencode/skills/pre-flight/SKILL.md")])

    // Injected: the composed prompt carries the tool-rendered directive
    // naming the resolved skill in its fixed position — after feature-dir
    // context, immediately before the extra prompt.
    expect(deliveredPrompt(fake.calls)).toBe(
      "Feature directory: openspec/changes/delegation-skill-enforcement/\n\n" +
        "Required skill: .opencode/skills/pre-flight/SKILL.md\n\n" +
        "This task is classified as deterministic work. Load the skill (invoke the `skill` tool " +
        'with name "pre-flight") or run the script at .opencode/skills/pre-flight/SKILL.md, and ' +
        "execute its full procedure. Do not substitute an improvised procedure.\n\n" +
        "Run the repository's pre-flight procedure for this task.",
    )

    // Recorded: the checkpoint data and the success envelope carry the
    // classification and the resolved reference — never the raw input.
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [statePathArg, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    expect(statePathArg).toBe(
      resolve(FAKE_CWD, ".uf/progress/opsx-delegation-skill-enforcement/uf.unleash.json"),
    )
    const state = JSON.parse(stateRaw as string) as ProgressState
    expect(state.current_step).toBe("implement")
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      feature_dir: "openspec/changes/delegation-skill-enforcement/",
      phase: 4,
      task_classification: "deterministic",
      skill: ".opencode/skills/pre-flight/SKILL.md",
      selection_source: "host",
    })
    if (result.status === "ok") {
      expect(result.data.task_classification).toBe("deterministic")
      expect(result.data.skill).toBe(".opencode/skills/pre-flight/SKILL.md")
      expect(result.data.skill_absent_reason).toBeNull()
    }
  })

  it("classifies the quick-tier whole-change implement delegation; an unclassified call fails closed", async () => {
    // Quick-tier call shape (command Step 5): one prompt-only delegation
    // for the whole routed change — no command file, no feature directory
    // (the quick tier has none), no phase — auto-checkpointing "implement"
    // on success. The declared answer's deterministic branch (skill or
    // reason) is enforced identically to the per-task shape and is covered
    // by the end-to-end test above; this scenario pins the quick-tier
    // shape itself.
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    // A docs-authorship change is judgment work — no repo skill or script
    // performs it end-to-end — so the declaration is "judgment" and the
    // run proceeds with the classification recorded.
    const result = await invokeDelegate(recordingExec("quick/add-docs-section").exec, deps, {
      prompt: "Implement the routed change completely on the current branch: add the usage docs section.",
      task_classification: "judgment",
      checkpoint_step: "implement",
    })

    expect(result.status).toBe("ok")
    expect(fake.calls.prompts).toHaveLength(1)
    if (result.status === "ok") {
      expect(result.data.task_classification).toBe("judgment")
      expect(result.data.skill).toBeNull()
      expect(result.data.skill_absent_reason).toBeNull()
    }
    expect(writeFileSync).toHaveBeenCalledTimes(1)
    const [statePathArg, stateRaw] = vi.mocked(writeFileSync).mock.calls[0]
    expect(statePathArg).toBe(resolve(FAKE_CWD, ".uf/progress/quick-add-docs-section/uf.unleash.json"))
    const state = JSON.parse(stateRaw as string) as ProgressState
    expect(state.current_step).toBe("implement")
    expect(state.data).toEqual({
      agent: "cobalt-crush-dev",
      task_classification: "judgment",
      selection_source: "host",
    })

    // An unclassified quick-tier delegation fails closed like any other
    // prompt-only call — the tier earns no exemption from the rule.
    const unclassified = await invokeDelegate(recordingExec("quick/add-docs-section").exec, deps, {
      prompt: "Implement the routed change completely on the current branch: add the usage docs section.",
    })
    expect(unclassified.status).toBe("error")
    if (unclassified.status === "error") {
      expect(unclassified.code).toBe("INVALID_INPUT")
      expect(unclassified.retryable).toBe(false)
      expect(unclassified.message).toBe(
        "task_classification: prompt-only delegation (no command_path) must declare task_classification -- " +
          "classify the work as deterministic or judgment before delegating",
      )
    }
    // The rejected call spawned no child session and wrote no state.
    expect(fake.calls.prompts).toHaveLength(1)
    expect(writeFileSync).toHaveBeenCalledTimes(1)
  })

  it("fails with HOST_MODEL_UNAVAILABLE, non-retryable, when no model can be resolved from the parent session at all", async () => {
    // Spec scenario (carried forward from pipeline hardening): structural
    // host-model resolution failure. The parent assistant message reports
    // no model identity, so no child session can replay it — the sibling
    // tests cover the invalid-variant and transient-failure causes.
    const fake = fakeClient({ hostModelMissing: true })
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the delegation.",
      task_classification: "judgment",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("HOST_MODEL_UNAVAILABLE")
      // Structural, not transient: non-retryable, naming the remediation.
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "current assistant model identity is unavailable. Remediation: run the command from a provider " +
          "that reports the active model identity; retrying will not fix this.",
      )
    }
    // Failed before the child spawn: host introspection ran, nothing else.
    expect(fake.calls.messages).toHaveLength(1)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
  })
})

// ---------------------------------------------------------------------------
// uf_delegate_command — path traversal
// ---------------------------------------------------------------------------

describe("uf_delegate_command path traversal", () => {
  it.each(["../secrets.md", "/etc/passwd"])(
    "rejects command_path %s escaping the repository root with COMMAND_PATH_TRAVERSAL",
    async (commandPath) => {
      const fake = fakeClient()
      const reads = fakeReadText({})
      const artifact = fakeFileExists([])
      const deps: DelegateCommandDependencies = {
        client: fake.client,
        modelMatrix: matrixFreeDeps,
        readText: reads.readText,
        fileExists: artifact.fileExists,
      }

      const result = await invokeDelegate(recordingExec().exec, deps, { command_path: commandPath })

      expect(result.status).toBe("error")
      if (result.status === "error") {
        expect(result.code).toBe("COMMAND_PATH_TRAVERSAL")
        expect(result.retryable).toBe(false)
        expect(result.message).toBe(
          "Path traversal blocked: command_path resolves outside the repository root",
        )
      }
      // Rejected before any file read, existence probe, or session activity.
      expect(reads.calls).toEqual([])
      expect(artifact.calls).toEqual([])
      expect(fake.calls.messages).toHaveLength(0)
      expect(fake.calls.creates).toHaveLength(0)
    },
  )

  it("rejects expected_artifact escaping the repository root with EXPECTED_ARTIFACT_TRAVERSAL before spawning", async () => {
    const fake = fakeClient()
    const reads = fakeReadText({})
    const artifact = fakeFileExists([])
    const deps: DelegateCommandDependencies = {
      client: fake.client,
      modelMatrix: matrixFreeDeps,
      readText: reads.readText,
      fileExists: artifact.fileExists,
    }

    const result = await invokeDelegate(recordingExec().exec, deps, {
      prompt: "Run the delegation.",
      task_classification: "judgment",
      expected_artifact: "../../outside/plan.md",
    })

    expect(result.status).toBe("error")
    if (result.status === "error") {
      expect(result.code).toBe("EXPECTED_ARTIFACT_TRAVERSAL")
      expect(result.retryable).toBe(false)
      expect(result.message).toBe(
        "Path traversal blocked: expected_artifact resolves outside the repository root",
      )
    }
    // Pre-resolved before spawning: no host introspection, no child session,
    // no artifact existence probe.
    expect(fake.calls.messages).toHaveLength(0)
    expect(fake.calls.creates).toHaveLength(0)
    expect(fake.calls.prompts).toHaveLength(0)
    expect(artifact.calls).toEqual([])
  })
})
